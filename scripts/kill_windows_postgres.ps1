# Run as Administrator
# Force kill Windows PostgreSQL processes that won't stop

Write-Host "Force killing Windows PostgreSQL processes..." -ForegroundColor Yellow

# Get all postgres processes not running in Docker
$postgresProcesses = Get-Process -Name postgres -ErrorAction SilentlyContinue | Where-Object {
    $_.Path -notlike "*docker*"
}

if ($postgresProcesses) {
    foreach ($proc in $postgresProcesses) {
        Write-Host "Killing process $($proc.Id): $($proc.Path)" -ForegroundColor Yellow
        Stop-Process -Id $proc.Id -Force -ErrorAction SilentlyContinue
    }
    Write-Host "Done!" -ForegroundColor Green
} else {
    Write-Host "No Windows PostgreSQL processes found." -ForegroundColor Gray
}

Write-Host ""
Write-Host "Checking port 5432..." -ForegroundColor Cyan
$port5432 = netstat -ano | findstr ":5432"
Write-Host $port5432

Write-Host ""
Write-Host "If you still see multiple processes, restart your computer." -ForegroundColor Yellow
