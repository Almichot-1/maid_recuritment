# Maid Recruitment Platform

Full-stack recruitment platform for Ethiopian and foreign agencies, with:
- agency authentication and approvals
- private partner workspaces
- candidate sharing, selection, approvals, and tracking
- admin portal with audit logs and platform settings

## Tech Stack

- Backend: Go, Chi, GORM, PostgreSQL
- Frontend: Next.js 14, TypeScript, Tailwind, React Query, Zustand
- File storage: S3-compatible object storage
- Email: SMTP

## Repository Layout

- [`cmd/api`](./cmd/api) - main API server
- [`cmd/expiryworker`](./cmd/expiryworker) - scheduled expiry worker for production
- [`cmd/adminseed`](./cmd/adminseed) - admin seeding tool
- [`frontend`](./frontend) - Next.js application
- [`internal`](./internal) - backend domain, services, handlers, repositories, middleware
- [`migrations`](./migrations) - SQL migrations
- [`scripts`](./scripts) - local dev and test helpers

## Local Development

1. Copy [`.env.example`](./.env.example) to `.env` and fill values.
2. Copy [`frontend/.env.example`](./frontend/.env.example) to `frontend/.env.local`.
3. Run migrations.
4. Start the API:
   ```bash
   make run-api
   ```
5. Start the frontend:
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

## Useful Commands

- `make run-api`
- `make run-expiry-worker`
- `make build-api`
- `make build-expiry-worker`
- `make migrate-up`
- `make migrate-down`
- `make test`

Frontend:
- `npm run dev`
- `npm run build`
- `npm run lint`
- `npx tsc --noEmit`

## Health Check

- `GET /api/v1/health`

## Deployment

Deployment is prepared for:
- Frontend on Vercel
- Backend API on Render
- Expiry worker on Render Cron Jobs
- PostgreSQL on Supabase
- S3-compatible storage for documents and media

See the full deployment guide in [docs/deployment.md](./docs/deployment.md).
