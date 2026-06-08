# Local Development Setup

This guide helps you set up the Maid Recruitment application for local development using Docker.

## Prerequisites

### Required
- **Docker Desktop** - Download from https://www.docker.com/products/docker-desktop
- **Go 1.25** - Download from https://golang.org/dl/
- **Node.js 18+** - Download from https://nodejs.org/
- **Git** - Download from https://git-scm.com/

### Optional but Recommended
- **migrate CLI** - For running database migrations manually
  ```bash
  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
  ```

## Quick Start

### 1. Start the Docker Database

Run this in PowerShell from the project root:

```powershell
./scripts/start-dev.ps1
```

This will:
- Start PostgreSQL in Docker
- Start pgAdmin for database management
- Run migrations automatically
- Display connection details

**First time only?** The database will be initialized automatically.

### 2. Configure Environment Variables

Both `.env.local` and `frontend/.env.local` are already configured for local development:

**Backend (.env.local)**:
```
DATABASE_URL=postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable
PORT=8080
API_URL=http://localhost:8080
```

**Frontend (frontend/.env.local)**:
```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

### 3. Start the Backend API

In a new PowerShell terminal:

```powershell
# Using Makefile
make run-api

# OR directly with Go
go run ./cmd/api
```

The API will start on `http://localhost:8080`

### 4. Start the Frontend

In another PowerShell terminal:

```powershell
cd frontend
npm install  # First time only
npm run dev
```

The frontend will start on `http://localhost:3000`

### 5. Start the Expiry Worker (Optional)

In a separate terminal:

```powershell
make run-expiry-worker
```

## Running on Different Machines/Environments

### Switch Between Local and Production APIs

**Local Development** (uses Docker):
```powershell
# Already set in frontend/.env.local
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

**Production** (uses Render API):
```powershell
# Edit frontend/.env.production.local
NEXT_PUBLIC_API_URL=https://maid-recruitment-api-frankfurt.onrender.com/api/v1
```

## Database Management

### Access Database

**Via Command Line**:
```powershell
docker exec -it maid-recruitment-db psql -U postgres -d maid_recruitment
```

**Via pgAdmin UI**:
- URL: http://localhost:5050
- Email: admin@example.com
- Password: admin

### Run Migrations

#### Automatically (on startup)
```powershell
./scripts/start-dev.ps1
```

#### Manually
```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable"
migrate -path ./migrations -database $env:DATABASE_URL up
```

#### Rollback Last Migration
```powershell
$env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable"
migrate -path ./migrations -database $env:DATABASE_URL down 1
```

### View Current Migration Version
```powershell
docker exec -it maid-recruitment-db psql -U postgres -d maid_recruitment -c "SELECT * FROM schema_migrations;"
```

## Troubleshooting

### "Port 5432 already in use"
Another PostgreSQL instance is running. Either:
- Stop the other instance
- Change the port in `docker-compose.yml` (e.g., `5433:5432`)

### "Database connection refused"
Make sure Docker is running and PostgreSQL is ready:
```powershell
docker ps  # Should show maid-recruitment-db
docker logs maid-recruitment-db  # View logs
```

### "API can't connect to database"
Check that:
1. `.env.local` has the correct `DATABASE_URL`
2. PostgreSQL is running: `docker-compose ps`
3. Database name is correct: `maid_recruitment`

### Frontend can't reach API
Check that:
1. Backend is running on `http://localhost:8080`
2. `frontend/.env.local` has correct `NEXT_PUBLIC_API_URL`
3. CORS is configured in backend

## Stopping Development Environment

```powershell
./scripts/stop-dev.ps1
# OR manually
docker-compose down
```

## Working on the Development Branch

You're currently on the `development` branch. This keeps your local work separate from `main`.

### Commit and Push Changes
```bash
git add .
git commit -m "Your commit message"
git push origin development
```

### Merge Back to Main (when ready)
```bash
git checkout main
git pull origin main
git merge development
git push origin main
```

## Useful Commands

| Command | Purpose |
|---------|---------|
| `make run-api` | Start backend API |
| `make run-expiry-worker` | Start expiry worker |
| `make test` | Run Go tests |
| `docker-compose ps` | View running containers |
| `docker-compose logs postgres` | View database logs |
| `docker-compose down` | Stop all containers |
| `docker-compose down -v` | Stop containers AND delete data |

## Important Notes

⚠️ **Never commit `.env.local` files** - They contain secrets
✅ **Always use `docker-compose.yml`** - It's version controlled and shared
🔄 **Database data persists** - Even after `docker-compose down`
🗑️ **To reset database**: `docker-compose down -v && docker-compose up -d`

## Need Help?

- Check logs: `docker-compose logs`
- Restart everything: `./scripts/stop-dev.ps1 && ./scripts/start-dev.ps1`
- Reset database: `docker-compose down -v`
