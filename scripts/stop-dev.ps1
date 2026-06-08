#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Stop the local development environment
.DESCRIPTION
    This script stops the Docker Compose stack
#>

Write-Host "🛑 Stopping local development environment..." -ForegroundColor Yellow
docker-compose down

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Environment stopped successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Failed to stop environment" -ForegroundColor Red
    exit 1
}
