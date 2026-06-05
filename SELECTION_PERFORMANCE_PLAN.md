# Maid Selection Process - Performance Enhancement Plan

## Executive Summary
The selection process has **7 identified performance bottlenecks** ranging from HIGH to LOW severity. The most impactful optimization (N+1 query fix) could reduce API response times by **50-70%** and database load by **60-80%** on large selections lists. This plan prioritizes fixes by impact/effort ratio and implementation complexity.

---

## Priority 1: Critical Issues (Implement First)

### 1.1 Fix N+1 Queries in GetMySelections Handler ⭐ HIGHEST IMPACT
**Severity:** HIGH  
**Current State:** For every selection returned, code queries database for candidate data  
**Problem:** 1 query for selections + N queries for each candidate's details = N+1 total queries  
**User Impact:** Selection list page with 50+ selections = 50+ database hits, slow page load  

**Current Code (selection_handler.go ~line 150):**
```go
func (h *SelectionHandler) GetMySelections(w http.ResponseWriter, r *http.Request) {
    selections, err := h.selectionService.GetSelectionsForWorkspace(userID, role, pairingID)
    
    for _, selection := range selections {
        candidate, err := h.candidateRepo.GetByID(selection.CandidateID)  // ← N+1
        response, err := h.mapSelectionResponse(selection, candidate)
    }
}
```

**Fix:** Use GORM Preload to fetch candidates in single query
```go
func (h *SelectionHandler) GetMySelections(w http.ResponseWriter, r *http.Request) {
    // Already includes candidates via Preload
    selections, err := h.selectionService.GetSelectionsForWorkspace(userID, role, pairingID)
    
    for _, selection := range selections {
        // selection.Candidate already loaded
        response, err := h.mapSelectionResponse(selection, selection.Candidate)
    }
}
```

**Backend Changes Required:**
1. **selection_repository.go**: Update query methods to use Preload
   ```go
   // GetBySelectedByAndPairing
   .Preload("Candidate").
   .Preload("Candidate.Documents").  // For photo_url
   
   // GetByCandidateOwner
   .Preload("Candidate").
   .Preload("Candidate.Documents").
   ```

2. **selection_service.go**: Ensure Preload is called in GetSelectionsForWorkspace
   ```go
   func (s *SelectionService) GetSelectionsForWorkspace(userID, role, pairingID string) {
       return repo.FindAll(
           db.Preload("Candidate").
           Preload("Candidate.Documents").
           Where("selected_by = ? AND pairing_id = ?", userID, pairingID)
       )
   }
   ```

**Expected Impact:**
- Reduce: 50+ queries → 2 queries (selections + candidates)
- Response time: ~2-3s → ~200-400ms
- Database CPU: -70%

**Implementation Time:** 30 minutes  
**Risk Level:** LOW (straightforward GORM change)

---

### 1.2 Add Missing Database Indexes
**Severity:** MEDIUM (affects query planning)  
**Problem:** Approval status queries scan entire approvals table  
**Current Query Pattern:**
```sql
SELECT * FROM approvals 
WHERE selection_id = ? AND decision = 'approved'
```

**Required Migration (000023_selection_performance_indexes.up.sql):**
```sql
-- Index for approval decision queries
CREATE INDEX IF NOT EXISTS idx_approvals_selection_id_decision 
ON approvals(selection_id, decision);

-- Index for candidate owner queries (to avoid JOIN)
CREATE INDEX IF NOT EXISTS idx_candidates_created_by_pairing 
ON candidates(created_by, pairing_id) WHERE status = 'available';

-- Composite index for selection queries by Ethiopian agent
CREATE INDEX IF NOT EXISTS idx_selections_candidate_created_by 
ON selections(candidate_id, pairing_id) 
WHERE status IN ('pending', 'approved');
```

**Rollback (000023_selection_performance_indexes.down.sql):**
```sql
DROP INDEX IF EXISTS idx_approvals_selection_id_decision;
DROP INDEX IF EXISTS idx_candidates_created_by_pairing;
DROP INDEX IF EXISTS idx_selections_candidate_created_by;
```

**Expected Impact:**
- Approval queries: 10-50ms → 1-2ms
- Ethiopian agent selection queries: 30-100ms → 5-10ms

**Implementation Time:** 15 minutes  
**Risk Level:** VERY LOW (indexes don't change data)

---

## Priority 2: High-Impact Issues (Implement Second)

### 2.1 Implement Real-Time Updates Instead of Frontend Polling
**Severity:** MEDIUM  
**Current State:** Frontend refetches ALL selections every 30 seconds  
**Problem:** Unnecessary server load + stale data (30s delay)  

**Frontend Code (selections/page.tsx ~line 66):**
```typescript
React.useEffect(() => {
    const interval = setInterval(() => {
        if (activeSelections.length > 0) {
            refetch()  // Refetches ALL selections every 30s
        }
    }, 30000)
    
    return () => clearInterval(interval)
}, [activeSelections.length, refetch])
```

**Solution A: WebSocket Approach (Recommended)**
1. Connect WebSocket when on selections page
2. Server sends selection update events only when status changes
3. Client updates React Query cache with new data
4. No periodic polling needed

**Backend WebSocket Handler (new file: internal/handler/websocket_handler.go):**
```go
type SelectionUpdateMessage struct {
    SelectionID string `json:"selection_id"`
    Status      string `json:"status"`
    UpdatedAt   time.Time `json:"updated_at"`
    Action      string `json:"action"`  // "approve", "reject", "expire"
}

func (h *WebSocketHandler) HandleSelectionUpdates(w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{}
    conn, _ := upgrader.Upgrade(w, r, nil)
    defer conn.Close()
    
    userID := extractUserIDFromContext(r)
    h.registerClient(userID, conn)
    
    for {
        select {
        case update := <-h.updates:
            if isRelevantToUser(update, userID) {
                conn.WriteJSON(update)
            }
        }
    }
}
```

**Frontend WebSocket Hook (new file: frontend/src/hooks/use-selection-updates.ts):**
```typescript
export function useSelectionUpdates(enabled: boolean) {
    const queryClient = useQueryClient()
    
    React.useEffect(() => {
        if (!enabled) return
        
        const ws = new WebSocket(`wss://api.yourdomain.com/ws/selections`)
        
        ws.onmessage = (event) => {
            const update = JSON.parse(event.data) as SelectionUpdateMessage
            
            // Update React Query cache instead of refetching all
            queryClient.setQueryData(['selection', update.selectionID], (old) => ({
                ...old,
                status: update.status,
                updatedAt: update.updatedAt,
            }))
            
            // Invalidate list to trigger UI update
            queryClient.invalidateQueries({ queryKey: ['my-selections'] })
        }
        
        return () => ws.close()
    }, [enabled, queryClient])
}
```

**Update selections page to use WebSocket:**
```typescript
// selections/page.tsx
const { data: selections } = useMySelections(activePairingId)
useSelectionUpdates(!!selections && selections.length > 0)

// Remove the old polling code completely
```

**Expected Impact:**
- Eliminate 2 requests/minute per user per browser tab
- For 100 concurrent users = 200 requests/min saved
- Latency for updates: ~500ms-2s (instant vs 30s delay)

**Implementation Time:** 3-4 hours (requires Go + TypeScript changes)  
**Risk Level:** MEDIUM (WebSocket connection management required)

---

### 2.2 Optimize Candidate Owner Query - Add Denormalization
**Severity:** MEDIUM  
**Current Problem:** Ethiopian agent query forces JOIN on candidates table
```sql
SELECT s.* FROM selections s
JOIN candidates c ON c.id = s.candidate_id
WHERE c.created_by = ?  -- Requires scanning candidates table
```

**Solution: Store candidate_created_by in selections table**

**Migration (000024_add_candidate_owner_denorm.up.sql):**
```sql
ALTER TABLE selections 
ADD COLUMN candidate_created_by UUID;

-- Backfill existing data
UPDATE selections s
SET candidate_created_by = c.created_by
FROM candidates c
WHERE s.candidate_id = c.id;

-- Add constraint
ALTER TABLE selections
ADD FOREIGN KEY (candidate_created_by) REFERENCES users(id);

-- Add index for queries
CREATE INDEX idx_selections_candidate_created_by 
ON selections(candidate_created_by, pairing_id, status);
```

**Rollback (000024_add_candidate_owner_denorm.down.sql):**
```sql
DROP INDEX idx_selections_candidate_created_by;
ALTER TABLE selections 
DROP CONSTRAINT selections_candidate_created_by_fkey;
ALTER TABLE selections 
DROP COLUMN candidate_created_by;
```

**Backend Change (selection_service.go):**
```go
// Before: Requires JOIN
func GetByCandidateOwner(candidateOwnerID, pairingID string) {
    return db.Where("c.created_by = ? AND s.pairing_id = ?", candidateOwnerID, pairingID).
            Joins("JOIN candidates c ON c.id = s.candidate_id")
}

// After: Direct query
func GetByCandidateOwner(candidateOwnerID, pairingID string) {
    return db.Where("candidate_created_by = ? AND pairing_id = ?", candidateOwnerID, pairingID).
            Preload("Candidate").
            Preload("Candidate.Documents")
}
```

**When to Update candidate_created_by:**
1. In SelectCandidateInPairing: Set when creating selection
   ```go
   selection.CandidateCreatedBy = candidate.CreatedBy
   ```
2. Create database trigger to auto-update on candidate changes (optional)

**Expected Impact:**
- Query time: 20-50ms → 2-5ms (no JOIN needed)
- Eliminates need to scan candidates table

**Implementation Time:** 1.5 hours  
**Risk Level:** LOW (denormalization with clear semantics)

---

## Priority 3: Medium-Impact Issues (Implement Third)

### 3.1 Cache Platform Settings
**Severity:** LOW-MEDIUM  
**Current Problem:** Platform settings queried on every selection creation  
**Current Code (selection_service.go):**
```go
func SelectCandidateInPairing(...) {
    settings, _ := h.settingsRepo.GetSettings()  // DB query every time
    lockDuration := settings.AutoExpireSelectionsHours
}
```

**Solution: In-Memory Cache with TTL**

**New Cache Service (internal/service/settings_cache.go):**
```go
type SettingsCacheService struct {
    repo repository.SettingsRepository
    cache *domain.PlatformSettings
    cacheTTL time.Duration
    lastFetch time.Time
}

func (s *SettingsCacheService) GetSettings() (*domain.PlatformSettings, error) {
    if time.Since(s.lastFetch) < s.cacheTTL && s.cache != nil {
        return s.cache, nil
    }
    
    settings, err := s.repo.GetSettings()
    if err == nil {
        s.cache = settings
        s.lastFetch = time.Now()
    }
    return settings, err
}
```

**Update DI Container:**
```go
// config/container.go
settingsCache := service.NewSettingsCacheService(
    settingsRepo,
    5 * time.Minute,  // Cache for 5 minutes
)
```

**Expected Impact:**
- Settings queries: ~20 per second → ~0.4 per second (12x reduction)
- Database load: -5-10%

**Implementation Time:** 1 hour  
**Risk Level:** LOW (simple caching pattern)

---

### 3.2 Move Sorting to Database
**Severity:** LOW  
**Current Problem:** Frontend sorts selections in JavaScript  

**Current Code (frontend/src/components/selections/selection-list.tsx ~line 44):**
```typescript
const sorted = selections.sort((a, b) => {
    if (sortBy === 'expiring-soon') {
        return new Date(a.expires_at).getTime() - new Date(b.expires_at).getTime()
    }
    return new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
})
```

**Solution: Add sortBy parameter to API**

**Backend (selection_handler.go):**
```go
func (h *SelectionHandler) GetMySelections(w http.ResponseWriter, r *http.Request) {
    sortBy := r.URL.Query().Get("sortBy")  // "newest" or "expiring-soon"
    pairingID := r.URL.Query().Get("pairing_id")
    
    var query = db.Where(...)
    
    if sortBy == "expiring-soon" {
        query = query.Order("expires_at ASC")
    } else {
        query = query.Order("created_at DESC")
    }
    
    selections, _ := h.selectionService.GetSelectionsWithQuery(query)
    h.WriteJSON(w, selections)
}
```

**Frontend (use-my-selections.ts):**
```typescript
export function useMySelections(pairingId: string, sortBy: string = 'newest') {
    return useQuery({
        queryKey: ['my-selections', pairingId, sortBy],
        queryFn: async () => {
            const response = await api.get('/selections', {
                params: { pairing_id: pairingId, sortBy }
            })
            return response.data  // Already sorted by database
        },
    })
}
```

**Expected Impact:**
- Shifts 50 items sorting from client (JavaScript) to database
- Client CPU: -5%
- Better for pagination later

**Implementation Time:** 30 minutes  
**Risk Level:** VERY LOW

---

### 3.3 Optimize Candidate Photo Lookup
**Severity:** LOW-MEDIUM  
**Current Problem:** Loops through all documents to find photo  

**Current Code (selection_handler.go):**
```go
func mapSelectionResponse(selection *domain.Selection, candidate *domain.Candidate) {
    var photoURL string
    for _, doc := range candidate.Documents {  // Loop through all docs
        if doc.DocumentType == string(domain.Photo) {
            photoURL = doc.FileURL
            break
        }
    }
}
```

**Solution A: Index Documents by Type**
```go
// Add to Candidate domain struct
func (c *Candidate) GetPhotoURL() string {
    for _, doc := range c.Documents {
        if doc.DocumentType == string(domain.Photo) {
            return doc.FileURL
        }
    }
    return ""
}
```

**Solution B: Denormalize (Recommended)**

**Migration (000025_candidate_photo_denorm.up.sql):**
```sql
ALTER TABLE candidates ADD COLUMN photo_url TEXT;
ALTER TABLE candidates ADD COLUMN photo_uploaded_at TIMESTAMPTZ;

-- Backfill from documents
UPDATE candidates c
SET photo_url = d.file_url, photo_uploaded_at = d.created_at
FROM documents d
WHERE d.candidate_id = c.id 
AND d.document_type = 'photo'
AND d.is_current = true;
```

**Backend Update:**
```go
// When uploading photo
candidate.PhotoURL = uploadedFileURL
candidate.PhotoUploadedAt = time.Now()

// When mapping response
response.PhotoURL = candidate.PhotoURL  // Direct access
```

**Expected Impact:**
- Eliminates document loop on every selection view
- Faster selection list rendering (50+ candidates)

**Implementation Time:** 45 minutes  
**Risk Level:** MEDIUM (denormalization requires careful sync)

---

## Priority 4: Low-Impact Optimizations (If Time Permits)

### 4.1 Add Selection Response Caching Headers
**Severity:** LOW  
**Backend (selection_handler.go):**
```go
w.Header().Set("Cache-Control", "public, max-age=30")  // Cache for 30s
w.Header().Set("ETag", calculateETag(selections))
```

### 4.2 Pagination for Large Selection Lists
**Add limit/offset parameters:**
```go
limit := r.URL.Query().Get("limit")  // Default 25
offset := r.URL.Query().Get("offset")  // Default 0
query = query.Limit(limit).Offset(offset)
```

### 4.3 GraphQL/DataLoader Pattern
**Consider if query complexity becomes issue:**
- Use DataLoader to batch candidate queries
- Allows frontend to request only needed fields

---

## Implementation Roadmap

### Phase 1: Quick Wins (1-2 days) - HIGH ROI
- [ ] Add database indexes (15 min)
- [ ] Fix N+1 queries with Preload (30 min)
- [ ] Add sorting to database (30 min)
- [ ] Cache platform settings (1 hour)
- **Total: ~2.5 hours | Impact: 60-70% performance improvement**

### Phase 2: Major Improvements (3-4 days) - HIGH ROI
- [ ] Implement denormalization for candidate_created_by (1.5 hours)
- [ ] Implement denormalization for photo_url (45 min)
- **Total: ~2.25 hours | Impact: Additional 15-20% improvement**

### Phase 3: Advanced (5-7 days) - MEDIUM ROI
- [ ] WebSocket implementation (3-4 hours)
- [ ] Response caching headers (15 min)
- [ ] Pagination support (1 hour)
- **Total: ~4.5 hours | Impact: Better UX + reduced polling load**

### Phase 4: Future (Optional)
- GraphQL/DataLoader pattern
- Redis caching for approval status
- Elasticsearch for advanced search

---

## Testing & Validation

### Before/After Metrics to Track
1. **API Response Time**
   - Before: GetMySelections with 50+ selections = 2-3s
   - Target: GetMySelections = 200-400ms
   - Tool: Chrome DevTools Network tab or curl with time

2. **Database Metrics**
   - Before: 50+ queries per selections page load
   - Target: 2-3 queries
   - Tool: PostgreSQL query logs or EXPLAIN ANALYZE

3. **Frontend Performance**
   - Before: Page load + 30s polling intervals
   - Target: Instant page load + WebSocket updates
   - Tool: Lighthouse, Web Vitals

4. **Concurrent Load Test**
   - Simulate 100 users viewing selections
   - Before: High CPU/memory spike
   - Target: Smooth performance

### Testing Steps
```bash
# 1. Enable PostgreSQL query logging
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();

# 2. Clear query logs
TRUNCATE postgres_log;

# 3. Run test load
# - Login as foreign agent
# - View selections page with 50+ selections
# - Check Vercel analytics
# - Check database logs for query count

# 4. Compare metrics
# - Count queries before/after
# - Measure response times
```

---

## Risk Mitigation

| Issue | Risk | Mitigation |
|-------|------|-----------|
| Database schema changes | Data loss | Test on staging environment first |
| Denormalization sync | Data inconsistency | Add database triggers for automatic sync |
| WebSocket scaling | Server memory | Monitor connections, implement reconnection |
| Cache invalidation | Stale data | Set reasonable TTL, test invalidation logic |

---

## Success Criteria

✅ Selection list loads in < 500ms (from current ~2-3s)  
✅ Database queries reduced by 80% on selection operations  
✅ Polling removed, replaced with real-time updates  
✅ All tests passing (unit + integration)  
✅ Vercel deployment build time unchanged  
✅ Zero data loss during migration  

---

## Estimated Timeline
- **Phase 1:** ~2.5 hours (high priority, do first)
- **Phase 2:** ~2.25 hours (implement after Phase 1)
- **Phase 3:** ~4.5 hours (if needed for real-time UX)
- **Total:** ~9-10 hours to implement all optimizations

**With these changes, expect:**
- 60-80% reduction in page load time
- 70-80% reduction in database load
- Real-time updates instead of stale 30-second polling
- Better scalability for 1000+ selections
