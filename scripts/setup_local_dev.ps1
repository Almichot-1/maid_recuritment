# Local Development Setup Script
# This will help you set up and verify your local development environment

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Local Development Environment Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check Docker
Write-Host "[1/5] Checking Docker..." -ForegroundColor Yellow
$dockerRunning = docker ps 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}
Write-Host "Docker is running" -ForegroundColor Green
Write-Host ""

# Step 2: Check if Docker PostgreSQL is running
Write-Host "[2/5] Checking Docker PostgreSQL..." -ForegroundColor Yellow
$pgContainer = docker ps --filter "name=maid-recruitment-db" --format "{{.Names}}"
if (-not $pgContainer) {
    Write-Host "Docker PostgreSQL not running. Starting..." -ForegroundColor Yellow
    docker-compose up -d postgres
    Start-Sleep -Seconds 5
}

$pgContainer = docker ps --filter "name=maid-recruitment-db" --filter "status=running" --format "{{.Names}}"
if ($pgContainer) {
    Write-Host "PostgreSQL container is running" -ForegroundColor Green
} else {
    Write-Host "ERROR: Failed to start PostgreSQL container" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 3: Check port conflicts
Write-Host "[3/5] Checking for port conflicts..." -ForegroundColor Yellow
$port5432 = netstat -ano | findstr ":5432" | Measure-Object | Select-Object -ExpandProperty Count
if ($port5432 -gt 2) {
    Write-Host "WARNING: Multiple processes listening on port 5432" -ForegroundColor Red
    Write-Host "You may have Windows PostgreSQL services running." -ForegroundColor Yellow
    Write-Host "Run scripts/stop_windows_postgres.ps1 as Administrator to stop them." -ForegroundColor Yellow
    Write-Host ""
}

# Step 4: Test database connection
Write-Host "[4/5] Testing database connection..." -ForegroundColor Yellow
$testResult = docker exec maid-recruitment-db psql -U postgres -d maid_recruitment -c "SELECT 1" 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "Database connection successful" -ForegroundColor Green
} else {
    Write-Host "WARNING: Could not connect to database" -ForegroundColor Yellow
}
Write-Host ""

# Step 5: Show next steps
Write-Host "[5/5] Setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "1. Run migrations:" -ForegroundColor White
Write-Host "   powershell -ExecutionPolicy Bypass -File scripts/run_all_migrations.ps1" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Start the backend:" -ForegroundColor White
Write-Host "   go run cmd/api/main.go" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Start the frontend (in another terminal):" -ForegroundColor White
Write-Host "   cd frontend" -ForegroundColor Gray
Write-Host "   npm run dev" -ForegroundColor Gray
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
