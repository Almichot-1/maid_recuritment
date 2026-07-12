# Phase 2: Backend Implementation - COMPLETE

## Summary
Backend implementation for selection progress tracking system is complete. The system replaces the old `status_steps` table with a new comprehensive progress tracking feature.

## Files Created

### 1. Service Layer
**File**: `internal/service/selection_progress_service.go`
- `NewSelectionProgressService()` - Constructor with dependency injection
- `CreateProgress()` - Auto-creates progress when selection is approved
- `GetProgress()` - Retrieves progress for a selection
- `UpdateProgress()` - Updates progress fields (Ethiopian agents only)
- `UploadDocument()` - Handles document uploads for COC/Medical/Visa/Ticket
- `SendProgressNotification()` - Sends summary notification to foreign agent
- Validation functions for all status types
- Document type parsing and validation

### 2. Handler Layer
**File**: `internal/handler/selection_progress_handler.go`
- `GetProgress()` - GET /api/v1/selections/:id/progress
- `UpdateProgress()` - PUT /api/v1/selections/:id/progress
- `UploadDocument()` - POST /api/v1/selections/:id/progress/documents/:type
- Request/Response mapping
- Error handling
- Document URL signing

## Files Modified

### 1. Domain Layer
- `internal/domain/selection.go` - Added `Progress` field to Selection struct
- `internal/domain/selection_progress.go` - Already created in Phase 1

### 2. Repository Layer
- `internal/repository/selection_progress_repository.go` - Already created in Phase 1
- `internal/repository/selection_repository.go` - Added `Preload("Progress")` to queries

### 3. Service Integration
- `internal/service/approval_service.go`:
  - Added `progressService` field
  - Added `SetProgressService()` method
  - Modified `ApproveSelection()` to auto-create progress and send notification

### 4. Main Application
- `cmd/api/main.go`:
  - Initialize `selectionProgressRepository`
  - Initialize `selectionProgressService`
  - Wire up `selectionProgressHandler`
  - Connect progress service to approval service
  - Add 3 new API routes for progress tracking

### 5. Handler Updates
- `internal/handler/selection_handler.go`:
  - Added `SelectionProgressSummary` type
  - Added `Progress` field to `SelectionResponse`
  - Added `mapProgressSummary()` helper method
  - Progress now included in all selection API responses

## API Endpoints

### 1. Get Progress
```
GET /api/v1/selections/:id/progress
```
- Returns full progress details with documents
- Available to both Ethiopian and Foreign agents

### 2. Update Progress
```
PUT /api/v1/selections/:id/progress
Content-Type: application/json

{
  "coc_status": "in_progress",
  "coc_type": "online",
  "medical_status": "done",
  "visa_status": "approved",
  "ticket_status": "booked",
  "arrival_status": "arrived",
  "arrival_date": "2024-03-15T10:00:00Z",
  "arrival_city": "Dubai"
}
```
- **Restricted**: Ethiopian agents only
- All fields are optional
- Updates only provided fields

### 3. Upload Document
```
POST /api/v1/selections/:id/progress/documents/:type
Content-Type: multipart/form-data

file: <binary>
```
- **Restricted**: Ethiopian agents only
- Document types: `coc`, `medical`, `visa`, `ticket`
- Supports: PDF, JPG, PNG (max 10MB)

## Progress Fields

### COC (Certificate of Competency)
- **Status**: pending, in_progress, done, failed
- **Type**: online, offline (optional)
- **Document**: Optional PDF/Image

### Medical
- **Status**: pending, in_progress, done, failed
- **Document**: Optional PDF/Image

### Visa
- **Status**: pending, in_progress, approved, rejected
- **Document**: Optional PDF/Image

### Ticket
- **Status**: pending, booked, confirmed
- **Document**: Optional PDF/Image

### Arrival
- **Status**: not_arrived, arrived
- **Date**: Optional ISO 8601 timestamp
- **City**: Optional destination city

## Permissions

### Ethiopian Agents
- ✅ View progress
- ✅ Update all fields
- ✅ Upload documents

### Foreign Agents
- ✅ View progress
- ❌ Update fields (read-only)
- ❌ Upload documents

## Notifications

When a selection is approved:
1. Progress record is automatically created
2. Summary notification sent to foreign agent
3. Notification includes status of all 5 tracking fields

## Integration Flow

1. **Selection Approved** → `ApprovalService.ApproveSelection()`
   - Creates selection progress record
   - Sends initial notification to foreign agent

2. **Ethiopian Agent Updates Progress** → `PUT /selections/:id/progress`
   - Validates user is candidate owner
   - Updates specified fields
   - Returns updated progress

3. **Ethiopian Agent Uploads Document** → `POST /selections/:id/progress/documents/:type`
   - Validates user is candidate owner
   - Uploads to S3/MinIO
   - Updates progress record with URL
   - Deletes old document if exists

4. **Anyone Views Selection** → `GET /selections/:id`
   - Selection includes progress summary in response
   - Shows current status of all 5 fields

5. **Anyone Views Full Progress** → `GET /selections/:id/progress`
   - Returns complete progress details
   - Includes all documents with signed URLs

## Database Migration

✅ Migration 000043 - Creates `selection_progress` table
✅ Migration 000044 - Drops old `status_steps` table

## Next Steps - Phase 3: Frontend

1. **Types** (`frontend/src/types/index.ts`)
   - Define TypeScript interfaces for progress

2. **Hooks** (`frontend/src/hooks/use-selection-progress.ts`)
   - `useSelectionProgress` - Fetch progress
   - `useUpdateProgress` - Update progress fields
   - `useUploadProgressDocument` - Upload documents

3. **Components**:
   - `progress-tracking.tsx` - Main progress tab component
   - `progress-badges.tsx` - Summary badges for list view
   - `progress-field.tsx` - Individual field editor
   - `progress-document-upload.tsx` - Document uploader

4. **Pages**:
   - Update `selections/[id]/page.tsx` to add Progress tab
   - Update `selections/page.tsx` to show progress badges

## Testing Checklist

- [ ] Rebuild backend: `docker-compose up -d --build api`
- [ ] Approve a selection and verify progress is created
- [ ] Test Ethiopian agent can update progress fields
- [ ] Test Foreign agent cannot update (403 forbidden)
- [ ] Test document uploads for all 4 document types
- [ ] Verify progress appears in selection GET response
- [ ] Verify progress detail endpoint returns full data
- [ ] Test notification is sent on approval

## Build Command

```powershell
docker-compose up -d --build api
```

This will rebuild the Go backend with all the new progress tracking code.
