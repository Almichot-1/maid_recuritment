# Local Development Port Conflict Fix

## Problem
You have Windows PostgreSQL services (postgresql-x64-17 and postgresql-x64-18) running on port 5432, which conflicts with your Docker PostgreSQL container.

## Solution

### Option 1: Stop Windows PostgreSQL Services (Recommended)

1. **Open PowerShell as Administrator** (Right-click PowerShell → Run as Administrator)

2. **Run the stop script:**
   ```powershell
   cd "C:\Users\NOOR AL MUSABAH\Documents\PROJECT_2"
   .\scripts\stop_windows_postgres.ps1
   ```

3. **Verify port is free:**
   ```powershell
   netstat -ano | findstr :5432
   ```
   You should now see only 2 entries (Docker), not 4

4. **Run migrations:**
   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts\run_all_migrations.ps1
   ```

5. **Start your backend:**
   ```powershell
   go run cmd/api/main.go
   ```

### Option 2: Change Docker Port (Alternative)

If you need to keep Windows PostgreSQL running, you can change Docker's port:

1. **Edit `docker-compose.yml`:**
   ```yaml
   ports:
     - "15432:5432"  # Changed from 5432:5432
   ```

2. **Update `.env` and `.env.local`:**
   ```env
   DATABASE_URL=postgresql://postgres:postgres@127.0.0.1:15432/maid_recruitment?sslmode=disable
   ```

3. **Restart Docker container:**
   ```powershell
   docker-compose down
   docker-compose up -d
   ```

## Verification

After fixing, test the connection:

```powershell
docker exec maid-recruitment-db psql -U postgres -d maid_recruitment -c "SELECT version()"
```

Then run a migration:

```powershell
go run cmd/devmigrate/main.go -file migrations/000001_init.up.sql
```

## Current Configuration

- **Docker PostgreSQL**: 
  - Container: `maid-recruitment-db`
  - Image: `postgres:17-alpine`
  - Username: `postgres`
  - Password: `postgres`
  - Database: `maid_recruitment`
  - Port: `5432` (external) → `5432` (internal)

- **Connection String**:
  ```
  postgresql://postgres:postgres@127.0.0.1:5432/maid_recruitment?sslmode=disable
  ```

## Quick Commands

```powershell
# Check what's running on port 5432
netstat -ano | findstr :5432

# Check Windows PostgreSQL services
Get-Service | Where-Object {$_.DisplayName -like "*postgres*"}

# Check Docker containers
docker ps

# View Docker PostgreSQL logs
docker logs maid-recruitment-db

# Access PostgreSQL inside Docker
docker exec -it maid-recruitment-db psql -U postgres -d maid_recruitment
```
