# Run all migrations for local Docker PostgreSQL database
$ErrorActionPreference = "Stop"

Write-Host "Starting migration process..." -ForegroundColor Cyan

# Get all .up.sql files sorted by name
$migrations = Get-ChildItem -Path "migrations\*.up.sql" | Sort-Object Name

$successCount = 0
$skipCount = 0
$failCount = 0

foreach ($migration in $migrations) {
    Write-Host ""
    Write-Host "Applying: $($migration.Name)" -ForegroundColor Yellow
    
    $result = go run cmd/devmigrate/main.go -file $migration.FullName
    $exitCode = $LASTEXITCODE
    
    if ($exitCode -eq 0) {
        Write-Host "Success" -ForegroundColor Green
        $successCount++
    }
    elseif ($exitCode -eq 2) {
        Write-Host "Already applied (skipped)" -ForegroundColor Gray
        $skipCount++
    }
    else {
        Write-Host "Failed" -ForegroundColor Red
        $failCount++
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Migration Summary:" -ForegroundColor Cyan
Write-Host "  Applied: $successCount" -ForegroundColor Green
Write-Host "  Skipped: $skipCount" -ForegroundColor Gray
Write-Host "  Failed:  $failCount" -ForegroundColor Red
Write-Host "========================================" -ForegroundColor Cyan

if ($failCount -gt 0) {
    Write-Host ""
    Write-Host "Some migrations failed. Check the output above." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "All migrations completed successfully!" -ForegroundColor Green
