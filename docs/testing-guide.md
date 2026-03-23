# End-to-End Testing Guide

## Scope
This guide validates the full recruitment lifecycle across authentication, candidate management, selection, approval, status tracking, notifications, rejection, expiry, and authorization guards.

- Base URL: `http://localhost:8080`
- Auth header format: `Authorization: Bearer <token>`
- Roles:
  - `ethiopian_agent`
  - `foreign_agent`

---

## Prerequisites

1. API is running and healthy:

```bash
curl -s http://localhost:8080/health
```

Expected:
- HTTP `200`
- body contains health response.

2. Test data URLs for documents are reachable (passport/photo/video).
3. You have `curl` and `jq` (or `httpie`) installed.
4. API runtime dependencies are configured:
   - Valid `DATABASE_URL` (not placeholder credentials)
   - Required AWS env vars for config bootstrap (`AWS_ACCESS_KEY`, `AWS_SECRET_KEY`, `AWS_REGION`, `AWS_S3_BUCKET`)

### Windows (PowerShell) quick start

Run native PowerShell scripts:

```powershell
pwsh -File .\scripts\seed-test-data.ps1
pwsh -File .\scripts\run-e2e-tests.ps1
```

If API is already running externally, skip local startup:

```powershell
pwsh -File .\scripts\run-e2e-tests.ps1 -SkipApiStart
```

Optional switches:

```powershell
pwsh -File .\scripts\run-e2e-tests.ps1 -SkipApiSmoke
pwsh -File .\scripts\run-e2e-tests.ps1 -SkipGoTests
pwsh -File .\scripts\seed-test-data.ps1 -IncludeCvGeneration
```

---

## Scenario 1: Happy Path - Complete Recruitment

- [ ] Ethiopian agent registers
- [ ] Ethiopian agent creates candidate
- [ ] Ethiopian agent uploads passport, photo, video
- [ ] Ethiopian agent generates CV
- [ ] Ethiopian agent publishes candidate
- [ ] Foreign agent registers
- [ ] Foreign agent views candidate list
- [ ] Foreign agent selects candidate
- [ ] Verify candidate locked
- [ ] Verify notifications sent
- [ ] Ethiopian agent approves
- [ ] Foreign agent approves
- [ ] Verify status steps created
- [ ] Ethiopian agent updates medical test to completed
- [ ] Verify foreign agent sees update
- [ ] Complete all steps
- [ ] Verify candidate status = Completed

### Manual steps and expected results

1. Register Ethiopian user via `POST /auth/register`.
   - Expected: `201`, response contains `token` and `user.role=ethiopian_agent`.
2. Register Foreign user via `POST /auth/register`.
   - Expected: `201`, response contains `token` and `user.role=foreign_agent`.
3. Ethiopian creates candidate via `POST /candidates/`.
   - Expected: `201`, response contains `candidate.id`, `status=draft`.
4. Ethiopian uploads three docs via `POST /candidates/{id}/documents` (`passport`, `photo`, `video`).
   - Expected: `201` each time; document added with matching `document_type`.
5. Ethiopian generates CV via `POST /candidates/{id}/generate-cv`.
   - Expected: `200`, response includes `cv_pdf_url`.
6. Ethiopian publishes candidate via `POST /candidates/{id}/publish`.
   - Expected: `200`; candidate status becomes `available`.
7. Foreign lists candidates via `GET /candidates/`.
   - Expected: `200`; candidate appears in list.
8. Foreign selects candidate via `POST /candidates/{id}/select`.
   - Expected: `201`; selection status `pending`; candidate status `locked`.
9. Verify lock via `GET /candidates/{id}`.
   - Expected: `status=locked`, `locked_by` and `lock_expires_at` are set.
10. Verify notifications via `GET /notifications/` for both users.
   - Expected: both receive selection-related notifications.
11. Ethiopian approves via `POST /selections/{selection_id}/approve`.
   - Expected: `200`; approval recorded, may still be pending second approval.
12. Foreign approves via same endpoint.
   - Expected: `200`; selection becomes `approved`; candidate enters progress lifecycle.
13. Verify status steps via `GET /candidates/{id}/status-steps`.
   - Expected: 5 predefined steps exist (`Medical Test`, `LMIS Approval`, `Visa Processing`, `Flight Booked`, `Deployed`).
14. Ethiopian updates `Medical Test` via `PATCH /candidates/{id}/status-steps/Medical%20Test` with `completed`.
   - Expected: `200`; step status updates.
15. Verify foreign can read progress via `GET /candidates/{id}/status-steps`.
   - Expected: updated step visible.
16. Complete all steps in sequence to `completed`.
   - Expected: all steps show `completed`; candidate terminal status should be `completed` per business expectation.

---

## Scenario 2: Rejection Flow

- [ ] Foreign agent selects candidate
- [ ] Ethiopian agent rejects
- [ ] Verify candidate available again
- [ ] Verify both notified

### Manual steps and expected results

1. Foreign selects an `available` candidate.
   - Expected: selection `pending`, candidate `locked`.
2. Ethiopian rejects via `POST /selections/{selection_id}/reject` with optional reason.
   - Expected: `200`, message confirms rejection.
3. Get candidate via `GET /candidates/{id}`.
   - Expected: candidate status reverts to `available`, lock cleared.
4. Check `GET /notifications/` for both users.
   - Expected: both receive rejection notification.

---

## Scenario 3: Expiry Flow

- [ ] Foreign agent selects candidate
- [ ] Wait 24+ hours
- [ ] Run expiry job
- [ ] Verify candidate released
- [ ] Verify both notified

### Manual steps and expected results

1. Foreign selects candidate.
   - Expected: selection `pending`, candidate `locked`.
2. Wait until `lock_expires_at` passes (24+ hours).
3. Ensure expiry scheduler runs (configured every 5 minutes in-app).
   - Expected: expiry processor executes automatically.
4. Fetch candidate and selection.
   - Expected: selection status `expired`; candidate status `available`, lock fields cleared.
5. Check notifications for both users.
   - Expected: both receive expiry notifications.

---

## Scenario 4: Authorization Tests

- [ ] Foreign agent tries to create candidate (should fail)
- [ ] Ethiopian agent tries to select candidate (should fail)
- [ ] Ethiopian agent tries to edit another's candidate (should fail)
- [ ] Foreign agent tries to update status steps (should fail)

### Manual steps and expected results

1. Foreign `POST /candidates/`.
   - Expected: `403` forbidden.
2. Ethiopian `POST /candidates/{id}/select`.
   - Expected: `403` forbidden.
3. Ethiopian tries `PUT /candidates/{other_user_candidate_id}`.
   - Expected: `403` forbidden.
4. Foreign tries `PATCH /candidates/{id}/status-steps/{step_name}`.
   - Expected: `403` forbidden.

---

## Scenario 5: Edge Cases

- [ ] Two foreign agents try to select same candidate simultaneously
- [ ] Approve already approved selection
- [ ] Update locked candidate
- [ ] Upload invalid file type
- [ ] Upload file > 50MB

### Manual steps and expected results

1. Fire two concurrent `POST /candidates/{id}/select` requests from different foreign agents.
   - Expected: exactly one success (`201`), one conflict (`409`) or not-available error.
2. Re-run approval for already approved user/selection.
   - Expected: idempotent/no duplicate approval records.
3. Attempt `PUT /candidates/{id}` while candidate is locked by another user.
   - Expected: request rejected (conflict/forbidden according to business rule path).
4. Upload document with invalid `document_type` or unsupported file metadata.
   - Expected: `400` validation error.
5. Upload file metadata > 50MB.
   - Expected: rejected by validation/business rule (confirm status + message).

---

## API Request Examples (curl)

### Register Ethiopian agent

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email":"ethiopian.e2e@example.com",
    "password":"Password123!",
    "full_name":"Ethiopian Agent",
    "role":"ethiopian_agent",
    "company_name":""
  }'
```

### Register Foreign agent

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email":"foreign.e2e@example.com",
    "password":"Password123!",
    "full_name":"Foreign Agent",
    "role":"foreign_agent",
    "company_name":"Global Agency"
  }'
```

### Login and capture token

```bash
ETH_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"ethiopian.e2e@example.com","password":"Password123!"}' | jq -r '.token')

FOR_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"foreign.e2e@example.com","password":"Password123!"}' | jq -r '.token')
```

### Create candidate

```bash
CANDIDATE_ID=$(curl -s -X POST http://localhost:8080/candidates/ \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name":"Sample Candidate",
    "age":24,
    "experience_years":3,
    "languages":["English","Amharic"],
    "skills":["Cooking","Cleaning"]
  }' | jq -r '.candidate.id')
```

### Upload passport/photo/video

```bash
curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/documents \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"document_type":"passport","file_url":"https://example.com/test/passport.pdf","file_name":"passport.pdf","file_size":1024}'

curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/documents \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"document_type":"photo","file_url":"https://example.com/test/photo.jpg","file_name":"photo.jpg","file_size":1024}'

curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/documents \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"document_type":"video","file_url":"https://example.com/test/video.mp4","file_name":"video.mp4","file_size":1024}'
```

### Generate CV and publish

```bash
curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/generate-cv \
  -H "Authorization: Bearer ${ETH_TOKEN}"

curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/publish \
  -H "Authorization: Bearer ${ETH_TOKEN}"
```

### Foreign list + select

```bash
curl -s -X GET http://localhost:8080/candidates/ \
  -H "Authorization: Bearer ${FOR_TOKEN}"

SELECTION_ID=$(curl -s -X POST http://localhost:8080/candidates/${CANDIDATE_ID}/select \
  -H "Authorization: Bearer ${FOR_TOKEN}" | jq -r '.selection.id')
```

### Approve / reject

```bash
curl -s -X POST http://localhost:8080/selections/${SELECTION_ID}/approve \
  -H "Authorization: Bearer ${ETH_TOKEN}"

curl -s -X POST http://localhost:8080/selections/${SELECTION_ID}/approve \
  -H "Authorization: Bearer ${FOR_TOKEN}"

curl -s -X POST http://localhost:8080/selections/${SELECTION_ID}/reject \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"reason":"Mismatch in requirement"}'
```

### Update status step

```bash
curl -s -X PATCH "http://localhost:8080/candidates/${CANDIDATE_ID}/status-steps/Medical%20Test" \
  -H "Authorization: Bearer ${ETH_TOKEN}" \
  -H "Content-Type: application/json" \
  -d '{"status":"completed","notes":"Medical exam passed"}'
```

### Verify notifications

```bash
curl -s -X GET http://localhost:8080/notifications/ \
  -H "Authorization: Bearer ${ETH_TOKEN}"

curl -s -X GET http://localhost:8080/notifications/ \
  -H "Authorization: Bearer ${FOR_TOKEN}"
```

---

## API Request Examples (HTTPie)

```bash
http POST :8080/auth/register email=ethiopian.httpie@example.com password=Password123! full_name='Ethiopian Agent' role=ethiopian_agent company_name=''
http POST :8080/auth/login email=ethiopian.httpie@example.com password=Password123!
http POST :8080/candidates/ Authorization:"Bearer $ETH_TOKEN" full_name='HTTPie Candidate' age:=26 experience_years:=4 languages:='["English"]' skills:='["Cooking"]'
http POST :8080/candidates/$CANDIDATE_ID/select Authorization:"Bearer $FOR_TOKEN"
http POST :8080/selections/$SELECTION_ID/approve Authorization:"Bearer $ETH_TOKEN"
http PATCH :8080/candidates/$CANDIDATE_ID/status-steps/Medical%20Test Authorization:"Bearer $ETH_TOKEN" status=completed notes='Done'
```

---

## Notes for Repeatable Runs

- Prefer unique emails per run (timestamp suffix) to avoid registration conflicts.
- If selection conflicts occur, create a new candidate before retrying.
- Expiry validation can be time-consuming due to 24-hour lock; for faster validation use a test environment with shortened expiry configuration or direct DB setup.
