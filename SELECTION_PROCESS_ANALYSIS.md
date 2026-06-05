# Maid Selection Process - Comprehensive Analysis

## 1. Complete User Flow for Selecting/Creating a Maid Selection

### Step 1: Foreign Agent Selects Candidate
**Flow:**
- Foreign agent navigates to candidate list in the dashboard
- Clicks "Select" on an available candidate
- Opens `SelectCandidateDialog` component showing candidate details and terms confirmation
- Confirms selection with checkbox agreement
- Frontend calls: `POST /candidates/{id}/select`

**Backend Operations** (`SelectCandidateInPairing` in selection_service.go):
1. Validate foreign agent role via notification service
2. Resolve active agency pairing if in workspace mode
3. Verify candidate is `available` status
4. Check candidate hasn't been selected already (unique constraint)
5. Lock candidate with pessimistic locking (`UPDATE` strength lock)
6. Create new Selection record with:
   - Status: `pending`
   - ExpiresAt: now + lock_duration_hours (default 24h, configurable)
   - SelectedBy: foreign agent ID
   - CandidateID & PairingID
7. Update Candidate status to `locked`
8. Send notifications:
   - To Ethiopian agent (candidate owner): "Candidate selected" - review approval request
   - To Foreign agent: "Selection confirmed" - upload contract package

**Database Transaction:**
- Ensures atomicity of candidate lock + selection creation
- Prevents race conditions (N+1 selections via unique constraint)

---

### Step 2: Foreign Agent Uploads Supporting Documents
**Flow:**
- Foreign agent navigates to selection detail page: `/selections/{id}`
- Views countdown timer showing remaining lock time
- Uploads two documents:
  1. Employer contract (PDF/image)
  2. Employer ID (PDF/image)
- Progress indicators show upload status

**Frontend Call:**
- `POST /selections/{id}/documents` with multipart form data
- Document types: `contract`, `employer_id`
- File size limit: 50MB

**Backend Operations** (`UploadSelectionDocument` in selection_service.go):
1. Validate selection exists and is in `pending` status
2. Verify uploaded_by is the selecting user
3. Validate file type (PDF/JPEG/PNG)
4. Upload to secure storage via StorageService
5. Store in transaction:
   - EmployerContractURL, EmployerIDURL
   - Filenames and upload timestamps
   - Previous versions deleted
6. Return updated selection with document URLs

**Validation:**
- Required before approval: `selectionHasRequiredSupportingDocuments()` checks employer contract URL exists

---

### Step 3: Both Parties Approve Selection
**Approval Flow:**

**Ethiopian Agent (Candidate Owner) Reviews:**
- Views selection detail page
- Can see:
  - Candidate summary with photo
  - Employer contract and ID documents
  - Countdown timer
  - Foreign agent approval status
  - "Approve" or "Reject" buttons
- Opens `ApprovalDialog` when confirming
- Calls: `POST /selections/{id}/approve` or `POST /selections/{id}/reject`

**Foreign Agent (Selecting Agent):**
- Also approves from their selection detail page
- Must upload documents first before approving

**Approval Logic** (`ApproveSelection` in approval_service.go):
1. Validate user is involved party (candidate owner OR foreign agent)
2. If approver is employer: check required documents uploaded
3. Check user hasn't already approved (via approvals table)
4. Create Approval record:
   - Decision: `approved`
   - DecidedAt: now
   - UserID: approving user
   - SelectionID
5. Query all approvals for selection
6. Check if both parties approved:
   - Ethiopian agent approval required? Check CreatedBy match
   - Foreign agent approval required? Check SelectedBy match
7. **If fully approved:**
   - Update Selection status to `approved`
   - Update Candidate status to `in_progress`
   - Initialize status steps for tracking
   - Send notifications: "Selection approved"
   - Candidate moves to recruitment tracking phase
8. **If awaiting other party:**
   - Send notification to both parties about pending approval
   - Selection remains `pending`

**Rejection Logic** (`RejectSelection`):
- Either party can reject at any time while selection is pending
- Reason can be provided
- Candidate status returns to `available`
- Candidate lock removed (released for re-selection)
- Selection status: `rejected`

---

### Step 4: Expiry Handling
**Automatic Expiry** (`ProcessExpiredSelections` background job):
1. Query all pending selections where `expires_at < NOW()`
2. Lock selection and candidate with `UPDATE` strength
3. Update Selection status to `expired`
4. Unlock Candidate:
   - Status back to `available`
   - Clear locked_by, locked_at, lock_expires_at
5. Send notifications:
   - To candidate owner: "Selection expired"
   - To foreign agent: "Selection expired"
6. Candidate available for re-selection

**Manual Override:** Admin can configure `auto_expire_selections` in platform settings

---

## 2. Key Backend Operations and Database Queries

### Selection Creation Query
```sql
INSERT INTO selections (id, candidate_id, selected_by, pairing_id, status, expires_at, created_at, updated_at)
VALUES (?, ?, ?, ?, 'pending', ?, NOW(), NOW())
-- Enforced by unique constraint: idx_selections_one_pending_per_candidate
```

### Get My Selections
**Foreign Agent:**
```sql
SELECT s.*, c.* FROM selections s
JOIN candidates c ON c.id = s.candidate_id
WHERE s.selected_by = ? AND s.pairing_id = ?
ORDER BY s.created_at DESC
-- Uses index: idx_selections_selected_by_pairing_status_created_at
```

**Ethiopian Agent:**
```sql
SELECT s.*, c.* FROM selections s
JOIN candidates c ON c.id = s.candidate_id
WHERE c.created_by = ? AND s.pairing_id = ?
ORDER BY s.created_at DESC
-- Uses index: idx_selections_pairing_status_created_at
```

### Get Expired Selections
```sql
SELECT * FROM selections
WHERE status = 'pending' AND expires_at < NOW()
```

### Get Approvals Status
```sql
SELECT * FROM approvals WHERE selection_id = ?
-- Computed approval status: both parties approved?
SELECT COUNT(*) FROM approvals 
WHERE selection_id = ? AND decision = 'approved' AND user_id IN (?, ?)
```

### Candidate Lock Update
```sql
UPDATE candidates 
SET status = 'locked', locked_by = ?, locked_at = NOW(), lock_expires_at = ?
WHERE id = ?
-- Enforced by pessimistic locking during transaction
```

### Selection to In-Progress
```sql
BEGIN TRANSACTION;
  UPDATE selections SET status = 'approved' WHERE id = ?;
  UPDATE candidates SET status = 'in_progress', locked_by = NULL WHERE id = ?;
  -- Initialize status_steps
COMMIT;
```

---

## 3. Frontend Components Involved

### Pages
- **[dashboard]/selections/page.tsx**: Main selections list with tabs (active, approved, rejected, expired)
- **[dashboard]/selections/[id]/page.tsx**: Selection detail page with document upload, approval buttons, candidate info

### Components
1. **SelectionList** (`selection-list.tsx`)
   - Search by candidate name
   - Sort by: newest, expiring soon
   - Filters selections by status
   - Displays SelectionCard for each selection

2. **SelectionCard** (`selection-card.tsx`)
   - Candidate photo, name, status
   - Selection expiration time
   - Approval status badges
   - Quick action buttons

3. **SelectCandidateDialog** (`select-candidate-dialog.tsx`)
   - Confirmation dialog when selecting candidate
   - Shows candidate summary
   - Terms agreement checkbox
   - Calls `useSelectCandidate` mutation

4. **ApprovalDialog** (`approval-dialog.tsx`)
   - Separate dialogs for approve/reject
   - Reject includes optional reason textarea
   - Confirmation messages

5. **LockCountdown** (`lock-countdown.tsx`)
   - Displays time remaining before selection expires
   - Real-time countdown timer

6. **StatusTimeline** (from candidates components)
   - Shows recruitment process steps after approval

### Data Fetching Hooks
- **useMySelections()**: Fetches all user's selections (role-based)
  - Disabled until pairing is ready
  - Cache: 30 seconds
  - Query key: `['my-selections', activePairingId]`

- **useSelection(id)**: Fetch single selection with candidate details
  - Triggered when viewing detail page
  - Query key: `['selection', id, activePairingId]`

- **useSelectionApprovals(id)**: Fetch approval status
  - Shows who approved/rejected
  - Pending approval list
  - Query key: `['selection-approvals', id, activePairingId]`

- **useApproveSelection(id, candidateId)**: Mutation for approving
  - Invalidates: selection, approvals, my-selections, candidates, progress
  - Handles both party approval triggers

- **useRejectSelection(id, candidateId)**: Mutation for rejecting
  - Invalidates same query keys as approve

---

## 4. Current Data Fetching Patterns

### N+1 Query Issues Found

**GetMySelections Handler** (selection_handler.go, lines 150-175):
```go
selections, err := h.selectionService.GetSelectionsForWorkspace(userID, role, pairingID)
// Returns []*domain.Selection

for _, selection := range selections {
    candidate, err := h.candidateRepo.GetByID(selection.CandidateID)  // ← N+1 ISSUE
    response, err := h.mapSelectionResponse(selection, candidate)
}
```
**Issue:** For N selections, makes N additional database queries to fetch candidates.
**Impact:** On selections list page with 50+ selections, causes 50+ queries.

### Eager Loading Needed
In `GetBySelectedByAndPairing` and similar queries:
```go
// Current (bad)
WHERE selected_by = ? AND pairing_id = ?

// Should be (good)
Preload("Candidate").
Preload("Candidate.Documents").
WHERE selected_by = ? AND pairing_id = ?
```

### Missing Indexes
- No index on `approvals(selection_id, decision)` - used frequently
- No index on `candidates(created_by)` - used for Ethiopian agent queries

### Frontend Query Caching
- Cache staleTime: 30 seconds (reasonable)
- No selective query invalidation - invalidates all selections on any change
- Could be optimized to only invalidate affected selections

---

## 5. Data Model Overview

### Selections Table
```sql
selections {
  id UUID PRIMARY KEY
  candidate_id UUID NOT NULL (indexed, unique with pending status)
  pairing_id UUID NOT NULL (indexed)
  selected_by UUID NOT NULL (indexed, foreign key to users)
  status selection_status NOT NULL ('pending', 'approved', 'rejected', 'expired')
  
  -- Document URLs (added in migration 004)
  employer_contract_url TEXT
  employer_contract_file_name TEXT
  employer_contract_uploaded_at TIMESTAMPTZ
  employer_id_url TEXT
  employer_id_file_name TEXT
  employer_id_uploaded_at TIMESTAMPTZ
  
  -- Expiry
  expires_at TIMESTAMPTZ NOT NULL
  
  -- Audit
  warning_sent_flags INT (for expiry warnings)
  created_at TIMESTAMPTZ NOT NULL
  updated_at TIMESTAMPTZ NOT NULL (with trigger)
}
```

### Approvals Table
```sql
approvals {
  id UUID PRIMARY KEY
  selection_id UUID NOT NULL (foreign key to selections)
  user_id UUID NOT NULL (foreign key to users)
  decision approval_decision NOT NULL ('approved', 'rejected')
  decided_at TIMESTAMPTZ NOT NULL
}
```

### Key Relationships
```
User (ethiopian_agent) 1 --→ ∞ Candidate (created_by)
User (foreign_agent) 1 --→ ∞ Selection (selected_by)
Candidate 1 --→ ∞ Selection (candidate_id)
Selection 1 --→ ∞ Approval (selection_id)
Selection 1 --→ ∞ Document (via candidate)

AgencyPairing 1 --→ ∞ Selection (pairing_id)
```

---

## 6. Performance Issues and Bottlenecks

### Critical Issues

#### 1. N+1 Queries on Selection List
- **Location:** `selection_handler.go` GetMySelections endpoint
- **Severity:** HIGH
- **Impact:** With 50 selections, makes 50+1 queries
- **Fix:** Implement GORM Preload for candidates in query

#### 2. Missing Approval Indexes
- **Location:** Approvals table in migrations
- **Severity:** MEDIUM
- **Impact:** Approval status queries may be slow
- **Fix:** Add `CREATE INDEX idx_approvals_selection_id_decision ON approvals(selection_id, decision);`

#### 3. Candidate Owner Query Uses JOIN
- **Location:** `GetByCandidateOwner` in repository
- **Query:** 
  ```sql
  SELECT s.* FROM selections s
  JOIN candidates c ON c.id = s.candidate_id
  WHERE c.created_by = ?
  ```
- **Severity:** MEDIUM
- **Issue:** Forces join on candidates table for every query
- **Fix:** Denormalize by storing `candidate_created_by` in selections, or add index on candidates(created_by)

#### 4. Frontend Polling
- **Location:** `selections/page.tsx` lines 66-73
- **Issue:** Refetches all selections every 30 seconds if any are pending
  ```typescript
  React.useEffect(() => {
    const interval = setInterval(() => {
      if (activeSelections.length > 0) {
        refetch()  // Refetches ALL selections
      }
    }, 30000)
  }, [activeSelections.length, refetch])
  ```
- **Severity:** MEDIUM
- **Impact:** Unnecessary server load
- **Fix:** Use WebSocket or server-sent events for real-time updates

### Medium Issues

#### 5. Candidate Document Fetching
- **Location:** `mapSelectionResponse` in handler
- **Issue:** Loops through all candidate documents to find photo
  ```go
  for _, document := range candidate.Documents {
      if document.DocumentType == string(domain.Photo) {
  ```
- **Severity:** LOW-MEDIUM
- **Fix:** Index documents or add photo_url denormalization to candidates

#### 6. Platform Settings Queried Per-Request
- **Location:** `SelectCandidateInPairing` service
- **Issue:** Fetches platform settings for every selection creation
- **Fix:** Cache platform settings with TTL or subscribe to changes

#### 7. Query Result Sorting Happens in Memory
- **Location:** Frontend `selection-list.tsx` lines 44-71
- **Issue:** Frontend sorts in JavaScript instead of database
- **Fix:** Add sortBy parameter to API, let database handle sorting

---

## 7. Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                    FOREIGN AGENT                            │
└─────────────────────────────────────────────────────────────┘
           │
           ▼
[1] POST /candidates/{id}/select
    └─→ SelectCandidateInPairing()
        ├─ Load candidate (lock with UPDATE)
        ├─ Check no pending selection exists
        ├─ Create Selection (pending, expires_at = now + 24h)
        ├─ Lock candidate (status = locked)
        ├─ Notify Ethiopian agent
        └─ Return Selection
           │
           ▼
[2] POST /selections/{id}/documents (contract + employer_id)
    └─→ UploadSelectionDocument()
        ├─ Load selection
        ├─ Validate documents
        ├─ Upload to storage
        └─ Update selection with URLs
           │
           ▼
[3] POST /selections/{id}/approve
    └─→ ApproveSelection()
        ├─ Lock selection & candidate
        ├─ Record approval decision
        ├─ Check if both approved
        ├─ If yes: update status to approved, unlock candidate
        └─ Send notifications

┌─────────────────────────────────────────────────────────────┐
│               ETHIOPIAN AGENT (Candidate Owner)              │
└─────────────────────────────────────────────────────────────┘
           │
           ▼ (Notified when candidate selected)
GET /selections/my
    └─→ GetSelectionsForWorkspace()
        ├─ GetByCandidateOwner (JOIN candidates)
        └─ Return selections list
           │
           ▼
GET /selections/{id}
    └─→ GetSelection()
        └─ Return selection with candidate details
           │
           ▼
GET /selections/{id}/approvals
    └─→ GetApprovals()
        └─ Query approvals for selection
           │
           ▼
[Optional: Reject] POST /selections/{id}/reject
    └─→ RejectSelection()
        ├─ Create approval with 'rejected' decision
        ├─ Update selection to 'rejected'
        ├─ Unlock candidate (status = available)
        └─ Send notifications

[Or: Approve] POST /selections/{id}/approve
    └─→ Same as Foreign Agent approval
```

---

## 8. Configuration Points

### Platform Settings (domain/platform_settings.go)
- `SelectionLockDurationHours` (default: 24) - How long selection is locked
- `RequireBothApprovals` (default: true) - Require both parties to approve
- `AutoExpireSelections` (default: true) - Auto-expire pending selections
- `AutoApproveAgencies` (default: false) - Auto-approve agency selections

### Admin Configuration UI (frontend/src/app/admin/portal/settings/page.tsx)
- Selection lock duration dropdown (3-72 hours)
- Auto-expire selections toggle
- Selection notification email template

---

## 9. Security & Access Control

### Role-Based Authorization
- **Foreign Agent:** Can only select candidates, approve selections they created
- **Ethiopian Agent:** Can only approve/reject selections for candidates they created
- **Admin:** Full visibility and control

### Function: `isInvolvedParty()`
```go
func isInvolvedParty(role, userID string, selection *domain.Selection, candidate *domain.Candidate) bool {
    if role == "foreign_agent" && selection.SelectedBy == userID {
        return true
    }
    if role == "ethiopian_agent" && candidate.CreatedBy == userID {
        return true
    }
    return false
}
```

### Pessimistic Locking
- Uses `GORM clause.Locking{Strength: "UPDATE"}` to prevent race conditions
- Applied when:
  - Checking candidate availability during selection
  - Processing expired selections
  - Recording approvals

---

## 10. Summary of Key Queries and Methods

| Operation | Method | Query Type | Performance |
|-----------|--------|-----------|-------------|
| Select candidate | SelectCandidateInPairing | Write (TX) | O(1) with locking |
| Get my selections | GetSelectionsForWorkspace | Read (JOIN) | O(N) + N×GetByID |
| Get single selection | GetSelection | Read | O(1) indexed |
| Approve selection | ApproveSelection | Write (TX) | O(1) |
| Upload document | UploadSelectionDocument | Write (TX) | O(1) |
| Expire selections | ProcessExpiredSelections | Write (TX) | O(N) batch |
| Get approvals | GetBySelectionID | Read | O(1) indexed |
| Check approval status | Computed in handler | Read | O(N) for N approvals |

---

## 11. Recommendations for Improvement

### Immediate (High Priority)
1. **Fix N+1 queries** - Add GORM Preload to eager load candidates
2. **Add missing indexes** - `approvals(selection_id, decision)`, `candidates(created_by)`
3. **Implement WebSocket** - Replace polling with real-time updates

### Short-term (Medium Priority)
1. **Cache platform settings** - Don't fetch on every request
2. **Denormalize candidate data** - Add cached fields to selections table
3. **Optimize sorting** - Move to database level, add composite indexes

### Long-term (Lower Priority)
1. **Implement event sourcing** - Better audit trail
2. **Add query optimization** - Use database views for complex queries
3. **Implement caching layer** - Redis for frequently accessed data
