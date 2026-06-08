# Local Dev S3 Workaround

## Issue
When creating candidates locally, the frontend tries to fetch documents and gets 500 errors because S3 credentials are incomplete.

## Workaround Options

### Option 1: Get AWS Secret Key from Render (Recommended)
```bash
# Get the secret from Render dashboard or use Render CLI
```

Then update `.env`:
```env
AWS_SECRET_KEY=<actual-secret-key-from-render>
```

### Option 2: Ignore Document Fetch Errors (Quick Fix)
The candidates ARE being created successfully. The 500 errors are only when fetching documents list.

**What works:**
✅ Create candidate
✅ Upload documents to S3 
✅ Save candidate to database

**What fails:**
❌ Fetching document list from S3 (returns 500)

**Impact:** 
- You can create candidates
- Documents ARE uploaded and stored
- Frontend just can't display the document list
- Candidate data is complete in database

### Option 3: Use Local MinIO (S3 Alternative)
Set up local S3-compatible storage:

```yaml
# Add to docker-compose.yml
minio:
  image: minio/minio
  ports:
    - "9000:9000"
    - "9001:9001"
  environment:
    MINIO_ROOT_USER: minioadmin
    MINIO_ROOT_PASSWORD: minioadmin
  command: server /data --console-address ":9001"
```

Update `.env.local`:
```env
AWS_ACCESS_KEY=minioadmin
AWS_SECRET_KEY=minioadmin
AWS_REGION=us-east-1
S3_BUCKET=maid-recruitment
S3_ENDPOINT=http://localhost:9000
S3_PUBLIC_BASE_URL=http://localhost:9000/maid-recruitment
```

## Quick Test

Check what's actually in the database:
```sql
SELECT id, full_name, status, cv_pdf_url FROM candidates ORDER BY created_at DESC LIMIT 5;
SELECT candidate_id, document_type, file_url FROM documents ORDER BY uploaded_at DESC LIMIT 10;
```

The data IS there! Just S3 fetching is failing.
