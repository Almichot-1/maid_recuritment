# Smart CV + One-Click Publishing — Final Plan

## Problem Summary

1. **Publish is broken** — `SetPairingService`/`SetUserRepository` are setter-injected,
   likely never wired → publish succeeds silently without sharing
2. **No smart defaults** — Foreign agent's country (`OperatingCountry`) not exposed anywhere
3. **No bulk publish** — No batch endpoint or UI
4. **Publish flow is clunky** — Dialog with setup steps is extra friction
5. **"Prepare CV" button** — No longer needed since CV auto-generates

## Design Decisions

| Topic | Decision |
|-------|----------|
| Salary resolution | Partner default pre-fills form → user changes freely → that is CV value. No chain. |
| Country resolution | Auto from foreign agent's `OperatingCountry` → falls back to partner `DefaultCountry` → falls back to candidate's field |
| Publish UX | Split-button, one click. No dialog. Partner list + "All Partners" in dropdown. |
| Mass publish | Candidates list → selection mode → dropdown → batch API |
| Batch fill | Empty fields get partner defaults → stored in `CandidatePairOverride` (not candidate record) |
| "Prepare CV" button | Removed entirely. CV auto-generates. Show download/status instead. |

---

## Phase 0 — Fix Wiring (blocker)

**Goal:** Unblock publish + auto-CV by fixing dependency injection.

**Changes:**

`internal/service/candidate_service.go`:
- Move `shareRepository`, `pairingService`, `userRepository`, `pairOverrideRepository`
  into `NewCandidateService` constructor parameters
- Remove Set* methods for these fields

`cmd/api/main.go`:
- Pass all dependencies to `NewCandidateService` at construction time
- Remove any Set* calls after construction

**Why:** Currently these are setter-injected. If a Set* call was missed (which is likely),
publish silently succeeds (status = Available) but creates no share record → no CV generated.
No error surfaced. The feature appears broken.

---

## Phase 1 — Foreign Agent's Country in Workspace

**Goal:** Every partner's country is available everywhere in the frontend.

**Changes:**

`internal/handler/pairing_handler.go`:
- Add `OperatingCountry string \`json:"operating_country"\`` to `PairingAgencySummary`
- In `mapPairingAgencySummary`, add: `OperatingCountry: *user.OperatingCountry` (defensive nil check)

`frontend/src/types/index.ts`:
- Add `operating_country?: string` to `PairingAgencySummary` interface

**Result:** `workspace.partner_agency.operating_country` is now available in every
component that has access to pairing context. The candidate form, publish button,
and CV generator all read from this single source.

---

## Phase 2 — Remove "Prepare CV" Button

**Goal:** CV is auto-generated. No manual "Prepare" step.

**Changes in `candidates/[id]/page.tsx`:**

**Location 1 — CV status card (around line 610-625):**

Before:
```tsx
{isEthiopianAgent && canGenerateCV ? (
  <Button asChild>
    <Link href={cvPageHref}>
      <Download className="h-4 w-4 mr-2" />
      Prepare CV
    </Link>
  </Button>
) : null}
```

After:
```tsx
{candidate.cv_pdf_url ? (
  <Button onClick={handleDownloadCV} disabled={isDownloadingCV}>
    <Download className="h-4 w-4 mr-2" />
    {isDownloadingCV ? "Downloading..." : "Download CV"}
  </Button>
) : (
  <p className="text-sm text-muted-foreground">
    CV will auto-generate once documents are complete.
  </p>
)}
```

**Location 2 — Action bar (around line 726-746):**

Before:
```tsx
{candidate.cv_pdf_url ? (
  <span>
    <Download className="h-4 w-4 mr-2" />
    {isDownloadingCV ? "Downloading..." : "Download CV"}
  </span>
) : (
  <Link href={cvPageHref}>
    <Download className="h-4 w-4 mr-2" />
    Prepare CV
  </Link>
)}
```

After:
```tsx
{candidate.cv_pdf_url ? (
  <Button onClick={handleDownloadCV} disabled={isDownloadingCV}>
    <Download className="h-4 w-4 mr-2" />
    {isDownloadingCV ? "Downloading..." : "Download CV"}
  </Button>
) : (
  <Button variant="outline" disabled>
    <Download className="h-4 w-4 mr-2" />
    CV Pending...
  </Button>
)}
<Button variant="ghost" size="sm" onClick={handleRegenerate}>
  Regenerate
</Button>
```

The `/candidates/[id]/cv` page stays as a power-user preview page but is no longer
the primary flow.

---

## Phase 3 — Redesign Publish: Split Button

**Goal:** Zero-click publish to default partner. One-click to any other partner.
No dialog.

**Replace `publish-dialog.tsx` with `publish-button.tsx`:**

```tsx
"use client";

import * as React from "react";
import { ChevronDown } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import type { WorkspaceSummary } from "@/types";

interface PublishButtonProps {
  workspaces: WorkspaceSummary[];
  isPublishing: boolean;
  onPublish: (pairingId?: string) => void;
}

export function PublishButton({ workspaces, isPublishing, onPublish }: PublishButtonProps) {
  const partnerName = (ws: WorkspaceSummary) =>
    ws.partner_agency?.company_name || ws.partner_agency?.full_name || ws.id;

  const needsSetup = (ws: WorkspaceSummary) =>
    !ws.default_country || !ws.default_currency;

  const canPublish = workspaces.some((ws) => !needsSetup(ws));
  const readyWorkspaces = workspaces.filter((ws) => !needsSetup(ws));

  if (readyWorkspaces.length === 0) {
    return (
      <Button disabled title="Set up partner defaults on the Partners page first">
        Publish
      </Button>
    );
  }

  if (readyWorkspaces.length === 1) {
    const ws = readyWorkspaces[0];
    return (
      <Button onClick={() => onPublish(ws.id)} disabled={isPublishing} className="gap-1">
        {isPublishing ? "Publishing..." : `Publish to ${partnerName(ws)}`}
        <ChevronDown className="h-4 w-4" />
      </Button>
    );
  }

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button disabled={isPublishing} className="gap-1">
          {isPublishing ? "Publishing..." : "Publish"}
          <ChevronDown className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        {workspaces.map((ws) => (
          <DropdownMenuItem
            key={ws.id}
            disabled={needsSetup(ws)}
            onClick={() => !needsSetup(ws) && onPublish(ws.id)}
          >
            <div className="flex items-center justify-between w-full gap-4">
              <span>{partnerName(ws)}</span>
              <span className="text-xs text-muted-foreground">
                {needsSetup(ws)
                  ? "⚠ Setup needed"
                  : `${ws.default_salary || ""} ${ws.default_currency || ""}`}
              </span>
            </div>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => onPublish()}>
          All Partners
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
```

**Key behaviors:**
- 0 partners ready → disabled button with tooltip
- 1 partner ready → main button shows "Publish to Kuwait", dropdown still shows options
- 2+ partners ready → main button shows "Publish", dropdown lists all
- Partner missing defaults → menu item shows "⚠ Setup needed", disabled
- "All Partners" → publishes to all active partners
- No dialog at all in the happy path

---

## Phase 4 — Auto-Fill on Publish

**Goal:** When publishing a candidate, empty country/salary get auto-filled from
the target partner's defaults and stored per-pairing (not on the candidate record).

**Logic in `PublishCandidate` and `BulkPublish`:**

```
For each (candidate, partner) pair:

1. Resolve country:
   a. Check existing CandidatePairOverride for this pairing → keep if exists
   b. IF candidate.country_applied is empty:
      - Use partner_agent.operating_country (auto from foreign agent profile)
      - IF empty → use partner.DefaultCountry
      - Save as CandidatePairOverride

2. Resolve salary:
   a. Check existing CandidatePairOverride → keep if exists
   b. IF candidate.salary_offered is empty:
      - Use partner.DefaultSalary + " " + partner.DefaultCurrency
      - Save as CandidatePairOverride

3. Create CandidatePairShare (if not already shared)

4. Generate per-pairing CV
```

**Where stored:** `CandidatePairOverride` table (already exists with unique constraint
`uq_pair_override` on pairing_id + candidate_id). This keeps each partner's values
independent.

---

## Phase 5 — CV Auto-Generation & Status Display

**Triggers:**

| Trigger | Action |
|---------|--------|
| Passport + Photo uploaded | Generate default CV (`candidates/{id}/cv/default.pdf`) |
| Published to partner | Generate per-pairing CV (`candidates/{id}/cv/{pairingID}.pdf`) |
| Candidate edited | Regenerate default CV + all per-pairing CVs |

**Status display on candidate detail page:**

```
Shared With:
┌──────────┬──────────┬──────────┬──────────┐
│ Partner  │ Country  │ Salary   │ CV       │
├──────────┼──────────┼──────────┼──────────┤
│ Kuwait   │ Kuwait   │ 2000 KWD │ ✅ Ready │
│ Jordan   │ Jordan   │ 1500 JOD │ ✅ Ready │
└──────────┴──────────┴──────────┴──────────┘
```

Each row has a small [Regenerate] button for manual re-gen if needed.

---

## Phase 6 — Batch Publish (Backend)

**New endpoint:** `POST /candidates/bulk-publish`

**Input:**
```json
{
  "candidate_ids": ["uuid-1", "uuid-2"],
  "pairing_ids": ["pairing-uuid-1"]
}
```

- `pairing_ids` empty → publishes to all active partners of the caller
- Max 200 candidates, max 10 partners

**Flow per (candidate, partner):**
1. Skip if already shared (idempotent)
2. Auto-fill empty country/salary → store as `CandidatePairOverride` (Phase 4 logic)
3. Create `CandidatePairShare`
4. Generate per-pairing CV

**Response:** `BatchResult` (success_count, error_count, errors)

**`internal/service/candidate_service.go`** — Add:
```go
type BulkPublishInput struct {
    CandidateIDs []string `json:"candidate_ids"`
    PairingIDs   []string `json:"pairing_ids"`
}

func (s *CandidateService) BulkPublish(userID string, input BulkPublishInput) *BatchResult {
    // For each candidate: validate ownership, check status
    // For each (candidate, partner): auto-fill → share → generate CV
    // Concurrent with semaphore (max 5)
    // Skip already-shared silently
}
```

---

## Phase 7 — Batch Publish (Frontend)

**New hook in `use-candidates.ts`:**
```typescript
export function useBulkPublish() { ... }
// POST /candidates/bulk-publish
// Invalidate candidates query on success
// Toast with result counts
```

**New component `BulkPublishDialog`:**
- Partner selection via checkboxes
- Shows candidate count + partner count
- Confirm button → calls `useBulkPublish`

**Changes in `candidates/page.tsx`:**
```
[ Cancel Selection ]  [ Bulk CV Actions ▼ ]
                      ┌──────────────────────────────┐
                      │ Regenerate CVs               │
                      │ Partner-Specific Values       │
                      │ ──────────────────────────── │
                      │ Publish to Partners           │  ← NEW
                      └──────────────────────────────┘
```

---

## File Change Summary

### Backend (Go)

| File | Phase | Change |
|------|-------|--------|
| `internal/service/candidate_service.go` | 0, 4, 6 | Constructor deps, auto-fill, BulkPublish |
| `internal/handler/pairing_handler.go` | 1 | `operating_country` in summary |
| `internal/handler/candidate_handler.go` | 6 | BulkPublish handler |
| `cmd/api/main.go` | 0, 6 | Fix DI, add route |

### Frontend (TypeScript/React)

| File | Phase | Change |
|------|-------|--------|
| `frontend/src/types/index.ts` | 1 | `operating_country` on summary |
| `frontend/src/app/(dashboard)/candidates/[id]/page.tsx` | 2, 3 | Remove "Prepare CV", wire publish button |
| `frontend/src/components/candidates/publish-dialog.tsx` | 3 | **Replace** with split button |
| `frontend/src/components/candidates/publish-button.tsx` | 3 | **New** |
| `frontend/src/components/candidates/bulk-publish-dialog.tsx` | 7 | **New** |
| `frontend/src/app/(dashboard)/candidates/page.tsx` | 7 | Add "Publish to Partners" |
| `frontend/src/hooks/use-candidates.ts` | 7 | Add useBulkPublish |

---

## Implementation Order

```
Phase 0 — Fix Wiring
Phase 1 — Foreign Agent Country
Phase 2 — Remove "Prepare CV" Button
Phase 3 — Split-Button Publish
Phase 4 — Auto-Fill on Publish
Phase 5 — CV Status Display
Phase 6 — Batch Publish (Backend)
Phase 7 — Batch Publish (Frontend)
```

Order ensures each phase builds on working foundations.

## Rollback

1. Revert Go code: revert commits for each phase
2. Revert frontend: revert commits
3. Data: `CandidatePairOverride` and `CandidatePairShare` records can be
   cleaned up if needed. No destructive migrations.
