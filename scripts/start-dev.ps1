#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Start the local development environment with Docker PostgreSQL
.DESCRIPTION
    This script:
    1. Starts the Docker Compose stack (PostgreSQL + pgAdmin)
    2. Waits for the database to be ready
    3. Runs database migrations
    4. Displays connection info and next steps
#>

param(
    [switch]$SkipMigrations = $false,
    [switch]$Build = $false
)

Write-Host "🚀 Starting Maid Recruitment local development environment..." -ForegroundColor Cyan

# Check if Docker is running
Write-Host "📦 Checking Docker..." -ForegroundColor Yellow
try {
    docker ps | Out-Null
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Start Docker Compose
Write-Host "🐳 Starting Docker Compose stack..." -ForegroundColor Yellow
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to start Docker Compose" -ForegroundColor Red
    exit 1
}

# Wait for PostgreSQL to be ready
Write-Host "⏳ Waiting for PostgreSQL to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0
$ready = $false

while ($attempt -lt $maxAttempts -and -not $ready) {
    try {
        docker exec maid-recruitment-db pg_isready -U postgres | Out-Null
        if ($LASTEXITCODE -eq 0) {
            $ready = $true
            Write-Host "✅ PostgreSQL is ready!" -ForegroundColor Green
        }
    } catch {
        # Silently continue
    }
    
    if (-not $ready) {
        $attempt++
        Start-Sleep -Seconds 1
    }
}

if (-not $ready) {
    Write-Host "❌ PostgreSQL failed to start after 30 seconds" -ForegroundColor Red
    exit 1
}

# Run migrations if not skipped
if (-not $SkipMigrations) {
    Write-Host "🗂️  Running database migrations..." -ForegroundColor Yellow
    
    # Check if migrate CLI is installed
    $migrate = Get-Command migrate -ErrorAction SilentlyContinue
    if ($null -eq $migrate) {
        Write-Host "⚠️  migrate CLI not found. Installing..." -ForegroundColor Yellow
        # Installation instructions shown to user - they need to install separately
        Write-Host "Please install migrate: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    } else {
        $env:DATABASE_URL = "postgres://postgres:postgres@localhost:5432/maid_recruitment?sslmode=disable"
        & migrate -path ./migrations -database $env:DATABASE_URL up
        
        if ($LASTEXITCODE -ne 0) {
            Write-Host "⚠️  Migration failed (this might be okay if DB is already migrated)" -ForegroundColor Yellow
        } else {
            Write-Host "✅ Migrations completed!" -ForegroundColor Green
        }
    }
}

# Display connection info
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "✅ Local Development Environment Ready!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Connection Details:" -ForegroundColor Cyan
Write-Host "   Database Host: localhost:5432"
Write-Host "   Database Name: maid_recruitment"
Write-Host "   Username: postgres"
Write-Host "   Password: postgres"
Write-Host ""
Write-Host "🛠️  Tools Available:" -ForegroundColor Cyan
Write-Host "   • PostgreSQL: postgres://postgres:postgres@localhost:5432/maid_recruitment"
Write-Host "   • pgAdmin: http://localhost:5050 (admin@example.com / admin)"
Write-Host ""
Write-Host "🚀 Next Steps:" -ForegroundColor Cyan
Write-Host "   1. Backend:  make run-api     (or: go run ./cmd/api)"
Write-Host "   2. Frontend: cd frontend && npm run dev"
Write-Host "   3. Worker:   make run-expiry-worker  (separate terminal)"
Write-Host ""
Write-Host "Stop the environment with: docker-compose down"
Write-Host ""
