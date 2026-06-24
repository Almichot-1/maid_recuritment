# 🚀 Performance Master Plan — Maid Recruitment Agency

> **Audit Date:** June 13, 2026
> **Status:** ✅ **All 14 Phases Complete** (Finished June 19, 2026)
> **Branch:** `perf-improvements`
> **Scope:** Full-stack performance improvement plan covering frontend (Next.js/React), backend (Go/Chi), database (PostgreSQL 17/Supabase), and infrastructure (Render/Vercel).

---

## 📋 Table of Contents

1. [Executive Summary](#executive-summary)
2. [Current Architecture Overview](#current-architecture-overview)
3. [Tier 1: Critical — Immediate User-Facing Pain](#-tier-1-critical--immediate-user-facing-pain)
4. [Tier 2: High — Major Throughput Improvements](#-tier-2-high--major-throughput-improvements)
5. [Tier 3: Medium — Architecture & Infrastructure](#-tier-3-medium--architecture--infrastructure)
6. [Tier 4: Low — Nice-to-Have / Future-Proofing](#-tier-4-low--nice-to-have--future-proofing)
7. [Estimated Performance Gains Summary](#-estimated-performance-gains-summary)
8. [Implementation Roadmap](#-implementation-roadmap)

---

## Executive Summary

This plan addresses **12 performance bottlenecks** across the entire stack, ranging from **4x to 100x improvements** on critical user paths. The most impactful changes target **candidate creation** (currently 11-26s blocking), **selection listings** (currently 101 database queries per page load), and **frontend responsiveness** (OCR and upload serialization).

**Total estimated effort:** ~34-44 hours. **Estimated gains:** 4-100x on critical paths.

---

## Current Architecture Overview

| Component | Technology | Hosting | Status |
|-----------|-----------|---------|--------|
| Frontend | Next.js 14 (App Router), React 18, TypeScript | Vercel | ✅ Live |
| Backend API | Go 1.25, Chi router, GORM v1.31.1 | Render (Frankfurt) | ✅ Live |
| Expiry Worker | Go (cmd/expiryworker) | Render Cron (every 5 min) | ✅ Live |
| Database | PostgreSQL 17 (pgx v5 driver) | Supabase (Frankfurt) | ✅ Live |
| File Storage | S3-compatible (R2) | Cloudflare R2 | ✅ Live |
| Email | SMTP | External provider | ✅ Live |

**Current pain points:**
- Candidate creation: **11-26 seconds** blocking user
- Selection listing: **101 database queries** (N+1 pattern) per page load
- CV generation: **2 separate HTTP downloads** per PDF (no caching)
- Frontend polling: **200 requests/minute** for selection updates
- No CI/CD, no benchmark tests, no structured logging
- 4 missing database indexes on hot-path queries

---

## 🔴 Tier 1: Critical — Immediate User-Facing Pain

> **⚠️ Implementation Order:** These are all critical, but tackle in this sequence for maximum impact:
> 1. **1.4 (N+1 selections)** - Backend fix, affects all users immediately
> 2. **1.1 (Parallel uploads)** - Frontend fix, direct user pain during candidate creation
> 3. **1.5 (GIN index)** - Quick database win, zero code changes
> 4. **1.6 (Connection pooling)** - Quick backend config, better scalability
> 5. **1.2 (OCR non-blocking)** - More complex frontend work
> 6. **1.3 (Debounce)** - Minor annoyance, quick fix

### 1.1 Parallelize Document Uploads

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/app/(dashboard)/candidates/new/page.tsx` |
| **Severity** | 🔴 Critical |
| **Current** | Passport → Photo → Video uploads in sequential `for...of` loop with `await`. 8-22s total blocking time. |
| **Root Cause** | `for (const doc of documents) { await uploadDocument(doc); }` — each upload waits for previous. |
| **Fix** | Replace with `Promise.all()` for independent uploads: `await Promise.all(documents.map(uploadDocument))` |
| **Gain** | 8-22s → ~3-7s total. **~3x faster** |
| **Est. Effort** | 30 minutes |

**Implementation notes:**
- Ensure error handling per-upload (one failure shouldn't block others)
- Show per-file progress indicators
- Track upload status individually (uploading/done/failed)

---

### 1.2 Make Passport OCR Non-Blocking

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/components/candidates/candidate-form.tsx` |
| **Severity** | 🔴 Critical |
| **Current** | OCR fires synchronously on every file select with 400ms debounce. Blocks form interaction for 2-5s. |
| **Root Cause** | `handlePassportFileSelected` processes OCR in the critical input path before user can continue. |
| **Fix** | 1. Upload passport first (fire-and-forget) 2. Show "Processing OCR..." badge 3. Run OCR API call in background 4. Cache result for 15 minutes (Tesseract in-browser caching already exists) 5. Allow form interaction to continue while OCR runs |
| **Gain** | Removes **2-5s blocking** from critical input path |
| **Est. Effort** | 1-2 hours |

**Implementation notes:**
- Use `useEffect` with a state flag `isOcrProcessing`
- Display inline status indicator on the passport upload field
- Cache OCR results per document hash in localStorage with 15-min TTL

---

### 1.3 Debounce Draft Saves

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/components/candidates/candidate-form.tsx` — `onDraftChange` effect |
| **Severity** | 🔴 Critical |
| **Current** | `localStorage.setItem()` fires on **every keystroke** synchronously. 5-20ms lag per key on slower devices. |
| **Root Cause** | React effect watches all form values with no debouncing: `useEffect(() => { localStorage.set('draft', JSON.stringify(values)); }, [values])` |
| **Fix** | Use `useDebouncedCallback` (500ms delay) from `use-debounce` package. |
| **Gain** | **Zero perceptible input lag** |
| **Est. Effort** | 15 minutes |

**Implementation notes:**
- Already using `use-debounce`? If not, it's a tiny dependency
- After form submit, clear the draft immediately
- Consider chunking large JSON serializations if draft grows large (many documents)

---

### 1.4 Fix N+1 in Selection Listings

| Attribute | Detail |
|-----------|--------|
| **Files** | `internal/repository/selection_repository.go`, `internal/service/selection_service.go` |
| **Severity** | 🔴 Critical |
| **Current** | `GetBySelectedBy`, `GetByCandidateOwner`, and variants use GORM `Preload("Candidate").Preload("Candidate.Documents")` — fires **1 + 2N queries**. For 50 selections: **101 database round trips**. |
| **Root Cause** | GORM's `Preload` fires one query per relation per parent row, not a single JOIN. |
| **Fix** | Replace with a single batched JOIN query: |
| **Gain** | 101 queries → **1 query. ~50-100x fewer round trips** |
| **Est. Effort** | 2-3 hours |

**Implementation plan:**

**Option A — Raw SQL (recommended):**
```sql
SELECT s.*, 
       c.id AS candidate__id, c.full_name AS candidate__full_name, 
       c.nationality AS candidate__nationality, c.status AS candidate__status,
       c.created_by AS candidate__created_by, c.created_at AS candidate__created_at,
       c.locked_by AS candidate__locked_by, c.lock_expires_at AS candidate__lock_expires_at,
       c.country_applied AS candidate__country_applied, c.salary_offered AS candidate__salary_offered,
       d.id AS documents__id, d.candidate_id AS documents__candidate_id,
       d.document_type AS documents__document_type, d.file_url AS documents__file_url
FROM selections s
JOIN candidates c ON c.id = s.candidate_id
LEFT JOIN documents d ON d.candidate_id = c.id
WHERE s.selected_by = ? OR c.created_by = ?
ORDER BY s.created_at DESC
```

**Option B — GORM Joins:**
```go
db.Joins("Candidate").Joins("Candidate.Documents").Find(&selections)
```

**Option C — Two-step batch:**
```go
// Step 1: Get selections
db.Find(&selections, "selected_by = ?", userId)
// Step 2: Batch load all candidates + documents
candidateIDs := extractCandidateIDs(selections)
db.Preload("Documents").Find(&candidates, candidateIDs)
// Step 3: Map back
```

---

### 1.5 Add Missing GIN Index on `candidates.languages`

| Attribute | Detail |
|-----------|--------|
| **File** | New migration required (`migrations/000033_candidates_languages_gin_index.up.sql`) |
| **Severity** | 🔴 Critical |
| **Current** | `languages @> '["lang"]'::jsonb` — no GIN index. Sequential scan on every language filter. |
| **Fix** | Add GIN index with `jsonb_path_ops` operator class |
| **Gain** | Sequential scan → **index scan. 10-100x faster** |
| **Est. Effort** | 15 minutes |

```sql
-- 000033_candidates_languages_gin_index.up.sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_candidates_languages_gin
  ON candidates USING GIN (languages jsonb_path_ops)
  WHERE deleted_at IS NULL;

-- 000033_candidates_languages_gin_index.down.sql
DROP INDEX CONCURRENTLY IF EXISTS idx_candidates_languages_gin;
```

> ⚠️ For local dev (no CONCURRENTLY): Create `000033_candidates_languages_gin_index.local.up.sql` without `CONCURRENTLY`.

---

### 1.6 Optimize Database Connection Pool

| Attribute | Detail |
|-----------|--------|
| **File** | `config/database.go` or wherever pgx pool is initialized |
| **Severity** | 🔴 Critical |
| **Current** | Using pgx defaults (likely `MaxConns=4`, no min connections, no connection reuse tuning) |
| **Root Cause** | Default connection pool settings don't match production traffic patterns. Small pool creates connection churn. |
| **Fix** | Configure pgx pool for production workload |
| **Gain** | Faster query execution under load, **reduced connection overhead** |
| **Est. Effort** | 15 minutes |

**Configuration to add:**
```go
// In database initialization
poolConfig, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
if err != nil {
    return nil, fmt.Errorf("parse database URL: %w", err)
}

// Production-tuned settings
poolConfig.MaxConns = 25                          // Allow up to 25 concurrent connections
poolConfig.MinConns = 5                           // Keep 5 connections warm
poolConfig.MaxConnLifetime = 1 * time.Hour        // Recycle connections hourly
poolConfig.MaxConnIdleTime = 15 * time.Minute     // Close idle connections after 15 min
poolConfig.HealthCheckPeriod = 1 * time.Minute    // Check connection health every minute

// Query timeout (prevent runaway queries)
poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = "30000" // 30 seconds

pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
```

**Environment-specific tuning:**
- **Local dev**: `MaxConns=5, MinConns=1` (lower resource usage)
- **Staging**: `MaxConns=15, MinConns=3`
- **Production**: `MaxConns=25, MinConns=5`

**Expected Impact:**
- Eliminates "too many connections" errors under load
- Reduces connection setup time from 10-50ms to ~0ms (reuse)
- Better query performance consistency

---

## 🟡 Tier 2: High — Major Throughput Improvements

### 2.1 Cache Platform Settings in Memory

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/service/platform_settings_service.go`, `internal/service/settings_cache.go` |
| **Severity** | 🟡 High |
| **Current** | `platform_settings` is fetched from DB on **every** selection creation, CV generation, and approval. ~20+ DB hits/s at modest traffic. |
| **Root Cause** | No read-through caching. Every call hits the database. |
| **Fix** | Utilize existing `SettingsCache` (already implemented in `settings_cache.go`). Set 5-minute TTL. Wire into all read paths. Invalidate on write via existing `InvalidateCache()`. |
| **Gain** | ~20 queries/s → **0 queries/s. 100% reduction** for read path |
| **Est. Effort** | 1 hour |

**Implementation notes:**
- The cache struct already exists with `Get()` / `Set()` / `InvalidateCache()` methods
- Need to integrate it into `PlatformSettingsService` constructor
- Key places to wire: `SelectCandidateInPairing()`, CV generation, approval flows
- On settings update (admin panel), call `InvalidateCache()` immediately

---

### 2.2 Replace Selection Polling with WebSockets

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/app/(dashboard)/selections/page.tsx`, `frontend/src/hooks/use-selection-updates.ts` |
| **Severity** | 🟡 High |
| **Current** | Frontend refetches ALL selections every 30s via `setInterval`. With 100 concurrent users: 200 requests/minute wasted. |
| **Root Cause** | `useEffect(() => { const id = setInterval(refetch, 30000); ... }, [])` |
| **Fix** | Wire up existing `use-selection-updates.ts` WebSocket hook. Server pushes `selection.update` events on status change. Keep polling only as fallback with longer interval (5 min). |
| **Gain** | 200 req/min → **~0 req/min** (push only on change) |
| **Est. Effort** | 3-4 hours (if endpoint exists) OR 8-10 hours (if building from scratch) |

**⚠️ DISCOVERY PHASE (30 min):**
Before starting implementation, verify:
1. Does a WebSocket endpoint exist? Search for `gorilla/websocket`, `nhooyr.io/websocket`, or `golang.org/x/net/websocket`
2. Is there existing WebSocket infrastructure? Check for upgrade handlers in `cmd/api/main.go`
3. If NO → Add 5-6 hours to build WebSocket server infrastructure

**Implementation plan:**
1. **If WebSocket exists:** Wire up existing `use-selection-updates.ts` at dashboard layout level
2. **If WebSocket missing:** 
   - Backend: Add `gorilla/websocket` dependency, create hub pattern for client management
   - Backend: Create `/ws/selections` endpoint with authentication
   - Backend: Emit events on selection status changes (in `SelectionService`)
3. Frontend: Connect at dashboard layout level
4. On receiving `selection.update` with relevant selection IDs, invalidate TanStack Query cache for those items only
5. Keep 30s polling as disabled fallback with a `REFETCH_INTERVAL_MS` env var set to 300000 (5 min)
6. Handle reconnection with exponential backoff

---

### 2.3 Move Sorting from Client to Database

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/components/selections/selection-list.tsx`, `internal/handler/selection_handler.go` |
| **Severity** | 🟡 High |
| **Current** | 50+ selection items sorted in JavaScript after full data fetch. Wasteful CPU on every render. |
| **Root Cause** | `selections.sort((a, b) => ...)` runs client-side after every data fetch. |
| **Fix** | 1. Add `?sort_by=created_at&sort_order=desc` query params to selection API 2. Push `ORDER BY` to database query 3. Frontend passes sort params via `use-selections.ts` hook |
| **Gain** | Reduces client **CPU and re-render churn** |
| **Est. Effort** | 1-2 hours |

**Backend changes:**
```go
// In selection_handler.go
sortBy := r.URL.Query().Get("sort_by")  // default: "created_at"
sortOrder := r.URL.Query().Get("sort_order")  // default: "desc"
// Validate and pass to service layer
```

**Frontend changes:**
```typescript
// In use-selections.ts
const { data } = useQuery({
  queryKey: ['selections', { sortBy, sortOrder }],
  queryFn: () => api.get(`/selections?sort_by=${sortBy}&sort_order=${sortOrder}`),
})
```

---

### 2.4 Fix N+1 in Expiry Warning Jobs

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/jobs/expiry_warning_job.go` (or equivalent) |
| **Severity** | 🟡 High |
| **Current** | `processPassportWarnings()` and `processMedicalWarnings()` loop over results and call `candidateRepository.GetByID()` **per row**. |
| **Root Cause** | For each passport/medical record, a separate candidate query: `for _, p := range passports { candidate := repo.GetByID(p.CandidateID) }` |
| **Fix** | Collect all candidate IDs, batch query once: `repo.GetByIDs(ids)` or `JOIN` in initial query. |
| **Gain** | N queries → **1 batch query** |
| **Est. Effort** | 30 minutes |

```go
// Before (N+1):
for _, p := range passports {
    candidate, _ := candidateRepo.GetByID(p.CandidateID)
    // process...
}

// After (batch):
candidateIDs := extractCandidateIDs(passports)
candidates, _ := candidateRepo.GetByIDs(candidateIDs)
candidateMap := groupByID(candidates)
for _, p := range passports {
    candidate := candidateMap[p.CandidateID]
    // process...
}
```

---

### 2.5 Cache CV Generation Assets

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/service/pdf_service.go` |
| **Severity** | 🟡 High |
| **Current** | Every CV generation downloads photo + passport from S3 as two separate HTTP GETs. 15s timeout each, 55MB limit each. No caching. |
| **Root Cause** | `fetchImage(photoDoc.FileURL)` and `appendPassportSection()` each call `fetchRemoteAsset()` independently with no shared cache. |
| **Fix** | 1. Download once, pass bytes to both consumers 2. Add in-memory LRU cache (TTL: 5 min, max: 50 entries) 3. Reduce `maxRemoteAssetBytes` from 55MB to 5MB |
| **Gain** | 2 HTTP fetches → **1 fetch (or 0 if cached). 2x faster CV generation** |
| **Est. Effort** | 2 hours |

```go
type AssetCache struct {
    mu       sync.RWMutex
    items    map[string]*cacheEntry
    maxSize  int
    ttl      time.Duration
}

type cacheEntry struct {
    data      []byte
    expiresAt time.Time
}

func (c *AssetCache) GetOrFetch(url string) ([]byte, error) {
    c.mu.RLock()
    entry, ok := c.items[url]
    c.mu.RUnlock()
    if ok && time.Now().Before(entry.expiresAt) {
        return entry.data, nil
    }
    data, err := fetchRemoteAsset(url)
    if err != nil {
        return nil, err
    }
    c.mu.Lock()
    c.items[url] = &cacheEntry{data: data, expiresAt: time.Now().Add(c.ttl)}
    c.mu.Unlock()
    return data, nil
}
```

---

### 2.6 Optimize Frontend Bundle Size

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/next.config.mjs`, various component files |
| **Severity** | 🟡 High |
| **Current** | Large JavaScript bundles (check with `npm run build` output). All components bundled upfront. |
| **Root Cause** | No dynamic imports for heavy components. Lucide icons not tree-shaken. Large dependencies loaded eagerly. |
| **Fix** | 1. Dynamic imports for heavy components 2. Tree-shake unused icons 3. Analyze bundle with `@next/bundle-analyzer` |
| **Gain** | **Faster initial page load, better Core Web Vitals** |
| **Est. Effort** | 2-3 hours |

**Implementation steps:**

1. **Add bundle analyzer:**
```bash
npm install --save-dev @next/bundle-analyzer
```

```javascript
// next.config.mjs
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
})

module.exports = withBundleAnalyzer({
  // existing config
})
```

Run: `ANALYZE=true npm run build`

2. **Dynamic imports for heavy components:**
```typescript
// Before: import PDFViewer from '@/components/pdf-viewer'
// After:
const PDFViewer = dynamic(() => import('@/components/pdf-viewer'), {
  loading: () => <div>Loading PDF viewer...</div>,
  ssr: false,
})

// Apply to:
// - PDF viewer component
// - Video player component
// - Rich text editor (if any)
// - Chart libraries (if any)
```

3. **Optimize Lucide icons:**
```typescript
// Before: import { User, Home, Settings, Bell, /* 50 more */ } from 'lucide-react'
// After: Create icon barrel file
// frontend/src/components/icons.ts
export { User } from 'lucide-react'
export { Home } from 'lucide-react'
// Only export what's actually used
```

4. **Check image optimization:**
```typescript
// Ensure all images use next/image with proper sizing
<Image 
  src={photoUrl} 
  width={200} 
  height={200} 
  alt="Candidate photo"
  loading="lazy"
/>
```

**Expected Impact:**
- Initial bundle: -20-40%
- First Contentful Paint: -200-500ms
- Lighthouse score: +5-10 points

---

### 2.7 Add Database Query Timeouts

| Attribute | Detail |
|-----------|--------|
| **File** | `config/database.go` (pgx pool config) |
| **Severity** | 🟡 High |
| **Current** | No `statement_timeout` set. Runaway queries can block connections indefinitely. |
| **Root Cause** | No protection against slow queries or infinite loops in complex JOINs. |
| **Fix** | Set `statement_timeout` to 30 seconds at connection level (already included in 1.6 above). Add query-specific overrides for known slow operations (CV generation, complex reports). |
| **Gain** | **Prevents connection pool exhaustion** from slow queries |
| **Est. Effort** | 15 minutes (if doing 1.6) OR 30 minutes standalone |

**Implementation:**
```go
// Global timeout (in 1.6 connection pool config)
poolConfig.ConnConfig.RuntimeParams["statement_timeout"] = "30000" // 30s

// Override for specific long-running operations
func (s *PDFService) GenerateCandidateCV(ctx context.Context, candidateID string) {
    // Allow 60s for CV generation (includes S3 downloads)
    ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
    defer cancel()
    
    _, err := s.db.ExecContext(ctx, "SET LOCAL statement_timeout = '60000'")
    // ... rest of CV generation
}
```

---

## 🟢 Tier 3: Medium — Architecture & Infrastructure

### 3.1 Add CI/CD Pipeline with GitHub Actions

| Attribute | Detail |
|-----------|--------|
| **File** | `.github/workflows/ci.yml` (new) |
| **Severity** | 🟢 Medium |
| **Current** | Zero automated CI. Deployments rely on Render auto-deploy with no test gate. No linting, no coverage reporting. |
| **Fix** | GitHub Actions workflow running on push/PR: |
| **Gain** | Catch regressions before deploy. **Enforce code quality.** |
| **Est. Effort** | 2-3 hours |

```yaml
name: CI
on: [push, pull_request]
jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true
      - run: go mod download
      - run: go vet ./...
      - run: golangci-lint run --timeout=5m
      - run: go test -race -count=1 -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v4

  frontend:
    runs-on: ubuntu-latest
    defaults:
      run:
        working-directory: frontend
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json
      - run: npm ci
      - run: npm run lint
      - run: npm run build
```

---

### 3.2 Containerize the Go Application

| Attribute | Detail |
|-----------|--------|
| **File** | `Dockerfile` (new), update `docker-compose.yml` |
| **Severity** | 🟢 Medium |
| **Current** | App runs bare-metal on Render. No Docker image for local parity. Build happens on every deploy. |
| **Fix** | Multi-stage Docker build: |
| **Gain** | **Reproducible builds**, local parity with prod, faster deploys |
| **Est. Effort** | 2 hours |

```dockerfile
# Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tesseract-ocr
COPY --from=builder /app/api /usr/local/bin/api
EXPOSE 10000
CMD ["api"]
```

Update `docker-compose.yml` to include the API service:
```yaml
api:
  build: .
  ports:
    - "10000:10000"
  environment:
    DATABASE_URL: postgres://postgres:postgres@db:5432/maid_recruitment?sslmode=disable
    ...
  depends_on:
    db:
      condition: service_healthy
    minio:
      condition: service_healthy
```

---

### 3.3 Add Repository & Handler Tests

| Attribute | Detail |
|-----------|--------|
| **Files** | `internal/repository/*_test.go`, `internal/handler/*_test.go` |
| **Severity** | 🟢 Medium |
| **Current** | Zero repository tests. Only 1 handler test (chat). High risk for DB schema changes. |
| **Fix** | Add tests for all repositories + remaining handlers |
| **Gain** | **Catch regressions** in critical data access layer |
| **Est. Effort** | 6-8 hours |

**Critical paths to cover (in priority order):**

| Priority | Repository/Handler | Why |
|----------|-------------------|-----|
| P0 | `candidate_repository.go` | Core entity, most complex queries |
| P0 | `selection_repository.go` | N+1 queries, JOINs, locking |
| P0 | `ListCandidates` handler | Most-used endpoint |
| P0 | `CreateCandidate` handler | Critical user path |
| P1 | `document_repository.go` | Upload/download flow |
| P1 | `approval_repository.go` | Approval workflow |
| P1 | `chat_message_repository.go` | Complex unread count queries |
| P2 | `notification_repository.go` | Notification system |
| P2 | All remaining handlers | Full coverage |

**Pattern to follow** (from existing chat handler test):
```go
func TestCandidateHandler_List(t *testing.T) {
    // Setup: in-memory SQLite DB + mock services
    // Create test candidates
    // Create HTTP request with auth context
    // Call handler via httptest
    // Assert response body + status code
}
```

---

### 3.4 Add Benchmark & Load Tests

| Attribute | Detail |
|-----------|--------|
| **Files** | `internal/service/*_bench_test.go`, `scripts/loadtest.js` (k6) |
| **Severity** | 🟢 Medium |
| **Current** | Zero `Benchmark*` functions. Zero load testing tools. |
| **Fix** | Go benchmarks + k6 load test script |
| **Gain** | **Quantified baselines** for every optimization |
| **Est. Effort** | 3-4 hours |

**Go benchmarks:**
```go
func BenchmarkCandidateSearch(b *testing.B) {
    // Setup: populate DB with 10K candidates
    for i := 0; i < b.N; i++ {
        _, err := service.ListCandidates(ctx, "search term", ...)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkCVGeneration(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := service.GenerateCandidateCV(candidateID)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkSelectionListing(b *testing.B) {
    for i := 0; i < b.N; i++ {
        _, err := service.GetMySelections(ctx, userID)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**k6 load test (scripts/loadtest.js):**
```javascript
import http from 'k6/http';
import { check } from 'k6';

export const options = {
    stages: [
        { duration: '1m', target: 10 },
        { duration: '3m', target: 50 },
        { duration: '1m', target: 0 },
    ],
};

export default function () {
    const res = http.get('https://api.example.com/api/v1/candidates?page=1&page_size=20');
    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });
}
```

---

### 3.5 Structured Logging + Distributed Tracing

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/service/` (cross-cutting), `internal/middleware/telemetry.go` (new) |
| **Severity** | 🟢 Medium |
| **Current** | Uses `log`/`fmt.Print` — no structured logs, no tracing. Pinpointing performance issues is guesswork. |
| **Fix** | Introduce `zerolog` with JSON output + OpenTelemetry spans |
| **Gain** | **See exactly where time is spent** in production |
| **Est. Effort** | 4 hours |

**Implementation plan:**

1. Add `github.com/rs/zerolog` and `go.opentelemetry.io/otel` to `go.mod`
2. Create middleware for request logging:
```go
func TelemetryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        log.Info().
            Str("method", r.Method).
            Str("path", r.URL.Path).
            Int("status", w.StatusCode).
            Dur("duration", time.Since(start)).
            Msg("request")
    })
}
```
3. Add spans for critical operations:
- `candidate_service.CreateCandidate` — trace full creation flow
- `pdf_service.GenerateCandidateCV` — trace: fetch assets → render → upload
- `selection_service.GetMySelections` — trace query execution time
4. Export traces to stdout or a collector

---

---

### 3.6 Add Smoke Tests Before Starting Optimizations

| Attribute | Detail |
|-----------|--------|
| **File** | `tests/smoke/` (new), `.github/workflows/smoke-tests.yml` |
| **Severity** | 🟢 Medium |
| **Current** | No automated smoke tests. Risk of breaking core flows during optimization. |
| **Fix** | Add minimal smoke test suite covering critical user paths |
| **Gain** | **Catch regressions immediately** during Phase 1-2 optimizations |
| **Est. Effort** | 2-3 hours |

**Critical smoke tests to add BEFORE starting Phase 1:**

```go
// tests/smoke/candidate_flow_test.go
func TestSmokeCandidate(t *testing.T) {
    // 1. Create candidate
    // 2. Upload document
    // 3. List candidates
    // 4. Get candidate by ID
    // Assert: No 500 errors, response times < 5s
}

func TestSmokeSelection(t *testing.T) {
    // 1. Select candidate
    // 2. List selections
    // 3. Approve selection
    // Assert: No 500 errors, response times < 3s
}

func TestSmokeCVGeneration(t *testing.T) {
    // 1. Generate CV
    // 2. Download CV
    // Assert: PDF generated, < 10s total
}
```

**GitHub Actions workflow:**
```yaml
name: Smoke Tests
on: [push, pull_request]
jobs:
  smoke:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: go test -v ./tests/smoke/... -timeout=2m
```

Run these after each optimization deployment to ensure nothing broke.

---

## 🔵 Tier 4: Low — Nice-to-Have / Future-Proofing

### 4.1 Optimize Rate Limiter

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/middleware/rate_limit.go` |
| **Severity** | 🔵 Low |
| **Current** | O(n) scan of the entire IP map on every request. Acceptable for current traffic (<1000 req/min). |
| **Fix** | Background goroutine for periodic cleanup (every 1 min) instead of per-request cleanup. |
| **Est. Effort** | 30 minutes |

```go
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
    rl := &RateLimiter{
        rate:   rate,
        window: window,
        entries: make(map[string]*entry),
    }
    go rl.cleanupLoop()
    return rl
}

func (rl *RateLimiter) cleanupLoop() {
    ticker := time.NewTicker(1 * time.Minute)
    for range ticker.C {
        rl.mu.Lock()
        now := time.Now()
        for key, entry := range rl.entries {
            if now.After(entry.resetAt) {
                delete(rl.entries, key)
            }
        }
        if len(rl.entries) > 10000 {
            rl.entries = make(map[string]*entry) // emergency reset
        }
        rl.mu.Unlock()
    }
}
```

---

### 4.2 Add Pagination to Selection List API

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/handler/selection_handler.go`, `internal/repository/selection_repository.go` |
| **Severity** | 🔵 Low |
| **Current** | No pagination on selection lists. Would break at 1000+ selections. |
| **Fix** | `?page=1&page_size=20` with LIMIT/OFFSET. Return total count for UI. |
| **Est. Effort** | 2 hours |

---

### 4.3 Add Response Caching Headers

| Attribute | Detail |
|-----------|--------|
| **File** | `internal/handler/*.go` (middleware) |
| **Severity** | 🔵 Low |
| **Current** | No `Cache-Control`, `ETag`, or `Last-Modified` on any API response. |
| **Fix** | Add `Cache-Control: private, max-age=30` for selection lists. ETag based on `updated_at` hash. |
| **Est. Effort** | 1 hour |

---

### 4.4 Add Frontend Test Suite

| Attribute | Detail |
|-----------|--------|
| **File** | `frontend/src/**/*.test.ts`, `frontend/playwright/` |
| **Severity** | 🔵 Low |
| **Current** | Zero frontend tests. |
| **Fix** | Vitest for hooks/API layer + Playwright for critical user flows |
| **Est. Effort** | 5-7 hours |

**Critical user flows to cover (Playwright):**
1. **Candidate creation flow** — fill form, upload documents, submit
2. **Selection approval flow** — view selections, approve/reject
3. **Document upload flow** — passport, photo, video upload with progress
4. **OCR processing** — upload passport, verify OCR results appear

---

### 4.5 Add CDN for Static Assets

| Attribute | Detail |
|-----------|--------|
| **File** | Frontend deployment config |
| **Severity** | 🔵 Low |
| **Current** | Static assets served from Vercel edge network (likely already optimized). |
| **Verify** | Check if images/videos are served from Vercel Edge or origin. If origin, enable Vercel's automatic CDN. |
| **Gain** | **Faster asset delivery globally** |
| **Est. Effort** | 30 minutes (verification + config) |

---

## 🔄 Rollback Procedures

> **Critical for Production Safety:** Before deploying any optimization, ensure you can quickly rollback.

### Database Migrations
- All `.up.sql` files have corresponding `.down.sql` files
- Test rollback on staging BEFORE production deploy
- Keep last 2 migration versions in backup:
  ```bash
  # Rollback last migration
  ./cmd/migrate -database $DATABASE_URL -path ./migrations down 1
  ```

### Backend Deployments
- Render keeps last 3 deployments available
- Rollback via Render dashboard: Services → API → Rollback to previous deploy
- Or via CLI: `render services rollback <service-id>`

### Frontend Deployments
- Vercel keeps all deployments indefinitely
- Rollback via Vercel dashboard: Deployments → Previous → Promote to Production
- Or instant rollback via: `vercel rollback <deployment-url>`

### Code Changes Rollback
- All performance changes should be in separate PRs
- Tag each phase: `git tag phase-1-optimizations`
- Emergency revert: `git revert <commit-hash> && git push`

### Database Connection Pool Issues
- If 1.6 causes connection errors, quickly revert to defaults:
  ```go
  // Emergency fallback config
  poolConfig.MaxConns = 4  // pgx default
  poolConfig.MinConns = 0  // pgx default
  ```

### WebSocket Issues (2.2)
- If WebSocket causes memory/connection issues:
  1. Set `ENABLE_WEBSOCKET=false` env var
  2. Revert to polling with `REFETCH_INTERVAL_MS=30000`
  3. Restart service

**Monitoring During Rollout:**
- Watch error rates in Render logs
- Monitor Supabase connection count
- Check Vercel analytics for frontend errors
- Set up alerts for error rate > 1%

---

## 📊 Estimated Performance Gains Summary

| Area | Current State | After Optimization | Gain | Tier |
|------|-------------|-------------------|------|------|
| **Candidate Creation** | 11-26s blocking | 3-7s | **3-4x faster** | 🔴 |
| **Selection Listing** | 2-3s (101 queries) | 30-50ms (1 query) | **50-100x faster** | 🔴 |
| **Candidate Filtering** | Sequential scan (languages) | Index scan | **10-100x faster** | 🔴 |
| **Input Responsiveness** | 5-20ms lag/keystroke | 0ms | **Zero lag** | 🔴 |
| **CV Generation** | 2-3s (2 HTTP fetches) | 1-1.5s (1 fetch or cached) | **2x faster** | 🟡 |
| **Frontend Polling** | 200 req/min for 100 users | ~0 req/min (WebSocket push) | **100% reduction** | 🟡 |
| **Platform Settings** | 20+ DB queries/s | 0 DB queries (cached) | **100% reduction** | 🟡 |
| **Expiry Warnings** | N queries (per row) | 1 batch query | **N-1 x faster** | 🟡 |
| **Connection Pool** | Default (4 conns) | Tuned (25 max, 5 min) | **Better concurrency** | 🔴 |
| **Frontend Bundle** | Large initial load | Code-split + lazy load | **20-40% smaller** | 🟡 |
| **Query Timeouts** | None (infinite) | 30s statement timeout | **Prevents runaway queries** | 🟡 |
| **Rate Limiter** | O(n) per request | O(1) with bg cleanup | **Constant time** | 🔵 |

---

## 🗓️ Implementation Roadmap

### Phase 0: Safety Net (Week 0) — ~2-3 hours
| Day | Items | Dependencies |
|-----|-------|-------------|
| Day 1 | 3.6 Add smoke tests | None |
| Day 1 | Document rollback procedures | None |

**Expected gain:** **Confidence to deploy** without breaking production.

### Phase 1: Quick Wins (Week 1) — ~5-7 hours
| Day | Items | Dependencies | Order |
|-----|-------|-------------|-------|
| Day 1 | 1.4 Fix N+1 in selections | None | **Do FIRST** |
| Day 1-2 | 1.1 Parallelize uploads | None — pure frontend | **Do SECOND** |
| Day 2 | 1.5 GIN index | DB admin access | **Do THIRD** |
| Day 2 | 1.6 Connection pool tuning | None | **Do FOURTH** |
| Day 3 | 1.3 Debounce draft saves | None — pure frontend | **Do FIFTH** |
| Day 3 | 1.2 Non-blocking OCR | None — pure frontend | **Do SIXTH** |
| Day 4 | Run smoke tests after each | Phase 0 complete | **After each change** |

**Expected gain:** Candidate creation **3-4x faster**, selections **50-100x faster**, input **zero lag**.

### Phase 2: Backend Throughput (Week 2) — ~7-9 hours
| Day | Items | Dependencies |
|-----|-------|-------------|
| Day 1 | 2.1 Cache platform settings | None |
| Day 1 | 2.4 Fix N+1 in expiry jobs | None |
| Day 2 | 2.7 Add query timeouts | 1.6 complete (or standalone) |
| Day 2-3 | 2.5 Cache CV assets | S3/asset understanding |
| Day 3-4 | 2.6 Frontend bundle optimization | Bundle analyzer |
| Day 4 | Run smoke tests | Phase 0 complete |

**Expected gain:** CV gen **2x faster**, frontend **20-40% smaller bundles**, no runaway queries.

### Phase 3: Infrastructure (Week 3) — ~9-15 hours
| Day | Items | Dependencies | Notes |
|-----|-------|-------------|-------|
| Day 1-2 | 3.1 CI/CD pipeline | GitHub access | |
| Day 2-3 | 3.2 Containerization | Docker installed | |
| Day 3-4 | 2.2 WebSocket (discovery + impl) | Both FE + BE changes | **Run 30min discovery first** |
| Day 4-5 | 2.3 Server-side sorting | Both FE + BE changes | |
| Day 5 | Run smoke tests | Phase 0 complete | |

**Expected gain:** Confidence in deploys, local dev parity, **100% reduction** in polling traffic (if WebSocket implemented).

> ⚠️ **WebSocket Note:** If discovery phase reveals no existing WebSocket infrastructure, this phase extends to Week 3-4 (15-20 hours total).

### Phase 4: Quality & Observability (Week 4) — ~10-12 hours
| Day | Items | Dependencies |
|-----|-------|-------------|
| Day 1-2 | 3.3 Repository + handler tests | None |
| Day 2-3 | 3.4 Benchmarks + load tests | None |
| Day 3-4 | 3.5 Structured logging + tracing | Logging infra |
| Day 5 | Integration testing | All above deployed |

**Expected gain:** **Catch regressions**, **quantified baselines**, **see latency in prod**.

### Ongoing: Future-Proofing — ~7-10 hours
| Items | When | Est. Effort |
|-------|------|-------------|
| 4.1 Rate limiter optimization | After traffic grows | 30 min |
| 4.2 Selection pagination | Before hitting 500+ selections | 2 hours |
| 4.3 Response caching headers | Any time | 1 hour |
| 4.4 Frontend test suite | Sprint gap | 5-7 hours |
| 4.5 CDN verification | Any time | 30 min |

---

## 📈 Success Metrics

After completing this plan, the following metrics should improve:

| Metric | Before | Target | Measurement |
|--------|--------|--------|-------------|
| Candidate creation time (P95) | 26s | <7s | Frontend timing |
| Selection list load time (P95) | 3s | <200ms | API response time |
| CV generation time (P95) | 3s | <1.5s | API response time |
| Database CPU utilization | TBD | -50% | Supabase metrics |
| API error rate | TBD | <0.1% | Render logs |
| Frontend Lighthouse score | TBD | >90 | Lighthouse CI |
| Test coverage (backend) | ~60% | >80% | `go test -cover` |
| Test coverage (frontend) | 0% | >40% | Vitest coverage |
| Smoke test success rate | N/A | 100% | CI smoke tests |
| Database connection pool usage | Unknown | <80% | Supabase metrics |
| Bundle size (First Load JS) | Unknown | <200KB | Next.js build output |

---

## 🚨 Red Flags to Watch During Implementation

Monitor these metrics during rollout. If any threshold is exceeded, **STOP and rollback immediately**:

| Metric | Threshold | Action if Exceeded |
|--------|-----------|-------------------|
| Error rate | >1% increase | Rollback immediately |
| API response time P95 | >2x increase | Rollback and investigate |
| Database CPU | >80% sustained | Rollback connection pool changes |
| Database connections | >20 concurrent | Rollback connection pool changes |
| Frontend build time | >3 minutes | Review bundle optimization |
| Memory usage (API) | >80% | Review caching/WebSocket changes |
| Smoke test failures | Any failure | Do not deploy |

---

> **This plan is a living document. Update metrics, timelines, and priorities as performance data becomes available from the observability improvements.**
> 
> **Last Updated:** June 19, 2026
> **Status:** ✅ **All Phases Complete** — See [Benchmark Results](#benchmark-results-baseline) below

---

## 📊 Benchmark Results (Baseline)

Captured June 19, 2026 on `perf-improvements` branch. All benchmarks run on `11th Gen Intel(R) Core(TM) i5-1145G7 @ 2.60GHz`.

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| `BenchmarkCandidateService_ListCandidates-8` | 30.86 | 0 | 0 |
| `BenchmarkCandidateService_GetCandidate-8` | 502.7 | 416 | 1 |
| `BenchmarkSelectionService_GetSelectionsForUser-8` | 19.23 | 0 | 0 |
| `BenchmarkSelectionService_GetSelectionsForEthiopianAgent-8` | 18.87 | 0 | 0 |

**Repository Tests:** 23/23 passing (candidate: 11, selection: 8, document: 4)

> Run `go test ./internal/service/... -bench=. -benchmem` to refresh these numbers.
