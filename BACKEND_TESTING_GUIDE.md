# Backend Testing Guide - Selection Progress Tracking

## Prerequisites
1. Backend rebuilt and running: `docker-compose up -d --build api`
2. Migrations applied (000043 and 000044)
3. Test users logged in (Ethiopian and Foreign agents)
4. At least one approved selection

## Test Scenarios

### Scenario 1: Auto-Create Progress on Approval
**Goal**: Verify progress is automatically created when selection is approved

1. Create a new selection (Foreign agent selects a candidate)
2. Approve the selection (Ethiopian agent approves)
3. Verify progress record is created automatically
4. Verify notification is sent to foreign agent

**Expected Result**:
- Progress record created with all fields at "pending"
- Foreign agent receives notification about progress

**API Test**:
```bash
# After approval, get the selection
curl -X GET "http://localhost:8080/api/v1/selections/{selection_id}" \
  -H "Authorization: Bearer {token}"

# Response should include progress summary
{
  "selection": {
    ...
    "progress": {
      "coc_status": "pending",
      "medical_status": "pending",
      "visa_status": "pending",
      "ticket_status": "pending",
      "arrival_status": "not_arrived"
    }
  }
}
```

### Scenario 2: Get Full Progress Details
**Goal**: Verify both agents can view full progress

**API Test**:
```bash
curl -X GET "http://localhost:8080/api/v1/selections/{selection_id}/progress" \
  -H "Authorization: Bearer {token}"
```

**Expected Response**:
```json
{
  "progress": {
    "id": "uuid",
    "selection_id": "uuid",
    "updated_by": "ethiopian_agent_id",
    "coc_status": "pending",
    "coc_type": null,
    "coc_document": null,
    "medical_status": "pending",
    "medical_document": null,
    "visa_status": "pending",
    "visa_document": null,
    "ticket_status": "pending",
    "ticket_document": null,
    "arrival_status": "not_arrived",
    "arrival_date": null,
    "arrival_city": null,
    "created_at": "2024-01-15T10:00:00Z",
    "updated_at": "2024-01-15T10:00:00Z"
  }
}
```

### Scenario 3: Update Progress (Ethiopian Agent)
**Goal**: Verify Ethiopian agent can update progress fields

**API Test**:
```bash
curl -X PUT "http://localhost:8080/api/v1/selections/{selection_id}/progress" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "coc_status": "in_progress",
    "coc_type": "online",
    "medical_status": "done",
    "visa_status": "approved",
    "ticket_status": "booked",
    "arrival_status": "arrived",
    "arrival_date": "2024-03-15T10:00:00Z",
    "arrival_city": "Dubai"
  }'
```

**Expected**: 200 OK with updated progress

### Scenario 4: Update Progress Forbidden (Foreign Agent)
**Goal**: Verify foreign agent CANNOT update progress

**API Test**:
```bash
curl -X PUT "http://localhost:8080/api/v1/selections/{selection_id}/progress" \
  -H "Authorization: Bearer {foreign_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "coc_status": "done"
  }'
```

**Expected**: 403 Forbidden
```json
{
  "error": "only ethiopian agents can update progress"
}
```

### Scenario 5: Upload COC Document (Ethiopian Agent)
**Goal**: Verify document upload works

**API Test**:
```bash
curl -X POST "http://localhost:8080/api/v1/selections/{selection_id}/progress/documents/coc" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -F "file=@test-document.pdf"
```

**Expected**: 200 OK with progress including document URL

### Scenario 6: Upload Medical Document
```bash
curl -X POST "http://localhost:8080/api/v1/selections/{selection_id}/progress/documents/medical" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -F "file=@medical-report.pdf"
```

### Scenario 7: Upload Visa Document
```bash
curl -X POST "http://localhost:8080/api/v1/selections/{selection_id}/progress/documents/visa" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -F "file=@visa.jpg"
```

### Scenario 8: Upload Ticket Document
```bash
curl -X POST "http://localhost:8080/api/v1/selections/{selection_id}/progress/documents/ticket" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -F "file=@ticket.pdf"
```

### Scenario 9: Invalid Document Type
```bash
curl -X POST "http://localhost:8080/api/v1/selections/{selection_id}/progress/documents/invalid" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -F "file=@test.pdf"
```

**Expected**: 400 Bad Request - invalid document type

### Scenario 10: File Too Large
Upload a file > 10MB

**Expected**: 413 Request Entity Too Large

### Scenario 11: Invalid File Type
Upload a .txt or .exe file

**Expected**: 400 Bad Request - invalid file type

### Scenario 12: Partial Update
**Goal**: Verify only specified fields are updated

```bash
curl -X PUT "http://localhost:8080/api/v1/selections/{selection_id}/progress" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "coc_status": "done"
  }'
```

**Expected**: Only COC status updated, other fields unchanged

### Scenario 13: Invalid Status Value
```bash
curl -X PUT "http://localhost:8080/api/v1/selections/{selection_id}/progress" \
  -H "Authorization: Bearer {ethiopian_token}" \
  -H "Content-Type: application/json" \
  -d '{
    "coc_status": "invalid_status"
  }'
```

**Expected**: 400 Bad Request - invalid status

### Scenario 14: Progress in Selection List
**Goal**: Verify progress summary appears in selection list

```bash
curl -X GET "http://localhost:8080/api/v1/selections/my" \
  -H "Authorization: Bearer {token}"
```

**Expected**: Each selection includes progress summary

## Validation Checklist

### Database
- [ ] `selection_progress` table exists
- [ ] Old `status_steps` table is dropped
- [ ] Progress record created when selection approved
- [ ] Document URLs stored correctly
- [ ] Timestamps updated properly

### API Endpoints
- [ ] GET `/selections/:id/progress` returns full details
- [ ] PUT `/selections/:id/progress` updates fields (Ethiopian only)
- [ ] POST `/selections/:id/progress/documents/:type` uploads documents
- [ ] Progress summary in selection GET responses
- [ ] Progress summary in selection list responses

### Permissions
- [ ] Ethiopian agent can view progress
- [ ] Ethiopian agent can update progress
- [ ] Ethiopian agent can upload documents
- [ ] Foreign agent can view progress
- [ ] Foreign agent CANNOT update progress (403)
- [ ] Foreign agent CANNOT upload documents (403)

### Validation
- [ ] Invalid status values rejected
- [ ] Invalid document types rejected
- [ ] Files > 10MB rejected
- [ ] Non-PDF/JPG/PNG files rejected
- [ ] Invalid date format rejected

### Notifications
- [ ] Notification sent when progress created (on approval)
- [ ] Notification includes summary of all fields

### Integration
- [ ] Progress auto-created on selection approval
- [ ] Progress preloaded in selection queries
- [ ] Document replacement works (old document deleted)
- [ ] Progress updates don't affect other selections

## Docker Commands

```powershell
# Rebuild backend
docker-compose up -d --build api

# Check logs
docker-compose logs -f api

# Check API health
curl http://localhost:8080/api/v1/health

# Database check
docker-compose exec -T postgres psql -U postgres -d maid_recruitment -c "SELECT * FROM selection_progress LIMIT 5;"
```

## Troubleshooting

### Build Fails
- Check for compilation errors in logs
- Verify all imports are correct
- Ensure repository interface is defined in domain

### Progress Not Created
- Check approval service logs
- Verify migrations ran successfully
- Check if progress service is wired up in main.go

### 403 Forbidden
- Verify user token is valid
- Check if user is Ethiopian agent for the candidate
- Verify middleware is applied correctly

### Document Upload Fails
- Check S3/MinIO is configured
- Verify storage service is set on progress service
- Check file size and type
