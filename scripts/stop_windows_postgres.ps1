# This script must be run as Administrator
# It stops Windows PostgreSQL services so Docker PostgreSQL can use port 5432

Write-Host "Stopping Windows PostgreSQL services..." -ForegroundColor Yellow

try {
    Stop-Service postgresql-x64-17 -ErrorAction SilentlyContinue
    Write-Host "Stopped postgresql-x64-17" -ForegroundColor Green
} catch {
    Write-Host "postgresql-x64-17 not running or already stopped" -ForegroundColor Gray
}

try {
    Stop-Service postgresql-x64-18 -ErrorAction SilentlyContinue
    Write-Host "Stopped postgresql-x64-18" -ForegroundColor Green
} catch {
    Write-Host "postgresql-x64-18 not running or already stopped" -ForegroundColor Gray
}

Write-Host ""
Write-Host "Done! Now your Docker PostgreSQL on port 5432 should be accessible." -ForegroundColor Green
Write-Host "You can restart the Windows PostgreSQL services later if needed." -ForegroundColor Cyan
