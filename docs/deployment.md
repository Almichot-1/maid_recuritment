# Deployment Guide

This repository is prepared for:
- Frontend: Vercel
- Backend API: Render Web Service
- Expiry processing: Render Cron Job
- Database: Supabase Postgres
- File storage: S3-compatible object storage
- Email: SMTP provider

## 1. Recommended Production Architecture

- `app.yourdomain.com` -> Vercel frontend
- `api.yourdomain.com` -> Render backend API
- Supabase -> primary PostgreSQL database
- S3 / R2 bucket -> candidate photos, passports, videos, generated CVs, contract files
- Render Cron -> runs the expiry worker every 5 minutes

## 2. Important Production Notes

- The API process and the expiry worker are now separated.
- Set `RUN_EXPIRY_SCHEDULER=false` on the Render web service.
- Use the dedicated cron service in [`render.yaml`](../render.yaml) for selection expiry.
- `REDIS_URL` is currently optional in the codebase and can be left empty unless Redis-backed features are added later.

## 3. Supabase Setup

1. Create a new Supabase project.
2. Copy the pooled Postgres connection string for the running API.
3. Keep the direct Postgres connection string available for migrations and manual database work.
4. Run all SQL files in [`migrations`](../migrations) in order.

Recommended:
- use pooled connection for the running Render API
- use direct connection for migrations and admin tasks

## 4. S3-Compatible Storage Setup

Provision one bucket for:
- candidate photos
- passports
- interview videos
- generated CV PDFs
- selection contract files
- employer ID files

Supported providers:
- AWS S3
- Cloudflare R2
- MinIO

Required backend envs:
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `AWS_REGION`
- `S3_BUCKET`
- `S3_ENDPOINT` if the provider is not AWS S3
- `S3_PUBLIC_BASE_URL` when the upload API endpoint is different from the public file URL

For Cloudflare R2 specifically:
- `S3_ENDPOINT=https://<account-id>.r2.cloudflarestorage.com`
- `S3_PUBLIC_BASE_URL=https://<your-public-bucket-domain-or-custom-domain>`

## 5. Backend on Render

Use the Render blueprint in [`render.yaml`](../render.yaml), or configure manually.

### Web Service

- Root directory: repo root
- Build command:
  ```bash
  go build -o ./bin/api ./cmd/api
  ```
- Start command:
  ```bash
  ./bin/api
  ```
- Health check path:
  ```text
  /api/v1/health
  ```

### Required env vars

- `DATABASE_URL`
- `JWT_SECRET`
- `APP_BASE_URL`
- `CORS_ALLOWED_ORIGINS`
- `AWS_ACCESS_KEY`
- `AWS_SECRET_KEY`
- `AWS_REGION`
- `S3_BUCKET`
- `S3_PUBLIC_BASE_URL`
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USER`
- `SMTP_PASS`

Optional:
- `S3_ENDPOINT`
- `REDIS_URL`

Important:
- Set `RUN_EXPIRY_SCHEDULER=false` on the web service.

## 6. Expiry Worker on Render Cron

The worker lives in [`cmd/expiryworker`](../cmd/expiryworker).

- Build command:
  ```bash
  go build -o ./bin/expiryworker ./cmd/expiryworker
  ```
- Start command:
  ```bash
  ./bin/expiryworker
  ```
- Schedule:
  ```text
  every 5 minutes
  ```

Use the same backend env vars as the API except `CORS_ALLOWED_ORIGINS` and `RUN_EXPIRY_SCHEDULER`.

## 7. Frontend on Vercel

Create a Vercel project using the `frontend` directory as the project root.

### Build settings

- Framework: Next.js
- Root directory: `frontend`
- Install command:
  ```bash
  npm install
  ```
- Build command:
  ```bash
  npm run build
  ```

### Required env vars

- `NEXT_PUBLIC_APP_URL=https://app.yourdomain.com`
- `NEXT_PUBLIC_API_URL=https://api.yourdomain.com/api/v1`
- `NEXT_PUBLIC_ENABLE_MOCK_MODE=false`

## 8. CORS Configuration

Set the backend `CORS_ALLOWED_ORIGINS` variable to the exact frontend origins, for example:

```text
https://app.yourdomain.com,https://your-vercel-project.vercel.app
```

For local development:

```text
http://localhost:3000,http://localhost:3001
```

## 9. Deployment Order

1. Push code to GitHub
2. Create Supabase database
3. Run migrations
4. Provision S3-compatible storage
5. Provision SMTP credentials
6. Deploy Render web service
7. Deploy Render cron worker
8. Deploy Vercel frontend
9. Update domains and DNS
10. Seed the first admin user

## 10. Seed the First Admin

Run:

```bash
go run ./cmd/adminseed --email super.admin@example.com --password "AdminPassword123!" --name "Platform Super Admin" --role super_admin
```

This prints the MFA secret and current OTP setup details for the first admin account.

## 11. Post-Deploy Smoke Test

Verify:
1. `/api/v1/health` returns `200`
2. agency registration works
3. admin login works
4. admin approval works
5. pairing creation works
6. Ethiopian agency creates and shares a candidate
7. foreign agency sees only shared candidates in the selected workspace
8. selection and dual approval work
9. process tracking updates work
10. CV generation and download work
11. expiry worker runs and releases expired pending selections

## 12. Recommended Next Hardening Pass

- move to a shared DB connection pool instead of opening one GORM connection per repository
- add error monitoring
- add private/signed URLs for sensitive document delivery
- add CI for tests and build checks on pull requests
