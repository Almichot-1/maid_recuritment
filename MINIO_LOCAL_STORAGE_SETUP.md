# MinIO Local Storage Setup - Complete ✅

## What is MinIO?
MinIO is a local S3-compatible object storage that runs on your machine. It's like having AWS S3 locally - same API, no cloud costs, no internet needed.

## Setup Complete

### Services Running
- ✅ **MinIO API**: `http://localhost:9000`
- ✅ **MinIO Console UI**: `http://localhost:9001`
- ✅ **Bucket Created**: `maid-recruitment`
- ✅ **Public Download**: Enabled

### Credentials
```env
AWS_ACCESS_KEY=minioadmin
AWS_SECRET_KEY=minioadmin
S3_BUCKET=maid-recruitment
S3_ENDPOINT=http://localhost:9000
S3_PUBLIC_BASE_URL=http://localhost:9000/maid-recruitment
```

## Access MinIO Console

**URL**: http://localhost:9001

**Login:**
- Username: `minioadmin`
- Password: `minioadmin`

From the console you can:
- Browse uploaded files
- View bucket contents
- Download files manually
- Monitor storage usage

## How It Works

**Upload Flow:**
1. User uploads passport/photo in frontend
2. Backend saves file to MinIO at `http://localhost:9000/maid-recruitment/passports/...`
3. URL saved to database
4. File is now accessible locally

**Download Flow:**
1. Frontend requests document list
2. Backend fetches from database (includes MinIO URLs)
3. Images display using `http://localhost:9000/maid-recruitment/...`
4. No more 500 errors! 🎉

## File Structure in MinIO

```
maid-recruitment/
├── passports/
│   └── candidate-id-timestamp.jpg
├── photos/
│   └── candidate-id-timestamp.jpg
├── videos/
│   └── candidate-id-timestamp.mp4
└── cvs/
    └── candidate-id-timestamp.pdf
```

## Verify It's Working

1. **Create a new candidate** in the UI
2. **Upload passport and photo**
3. **Check MinIO Console**: http://localhost:9001 → Browse → `maid-recruitment` bucket
4. **You should see your files!**

## Advantages vs AWS S3

✅ **Free** - no AWS charges
✅ **Fast** - local network speed
✅ **Offline** - works without internet
✅ **Private** - files never leave your machine
✅ **Same API** - code works with both MinIO and AWS S3

## Production Deployment

When deploying to production:
1. Keep using AWS S3 (already configured in Render)
2. MinIO is ONLY for local development
3. The same backend code works with both!

## Troubleshooting

### Can't access MinIO Console
```bash
docker ps | grep minio
# Check if maid-recruitment-minio is running
```

### Files not uploading
```bash
docker logs maid-recruitment-minio
# Check MinIO logs for errors
```

### Bucket not found
```bash
docker exec maid-recruitment-minio mc ls local/
# List all buckets
```

### Reset MinIO (delete all files)
```bash
docker-compose down -v
docker-compose up -d minio
# Recreate bucket after
docker exec maid-recruitment-minio mc mb local/maid-recruitment
docker exec maid-recruitment-minio mc anonymous set download local/maid-recruitment
```

## Storage Location

MinIO data persists in Docker volume: `project_2_minio_data`

To view storage usage:
```bash
docker exec maid-recruitment-minio mc du local/maid-recruitment
```

---

**Your local development environment now has complete file storage! 🚀**
