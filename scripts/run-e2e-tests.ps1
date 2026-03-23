param(
    [string]$ApiBaseUrl = 'http://localhost:8080',
    [string]$OutDir = '',
    [switch]$SkipApiStart,
    [switch]$SkipApiSmoke,
    [switch]$SkipGoTests,
    [switch]$IncludeCvGeneration
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$RootDir = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
Set-Location $RootDir

if (-not $OutDir) {
    $OutDir = Join-Path $RootDir 'reports'
}
New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$apiLog = Join-Path $OutDir "api-$timestamp.log"
$apiErrLog = Join-Path $OutDir "api-$timestamp.err.log"
$goTestLog = Join-Path $OutDir "go-test-$timestamp.log"
$apiE2eLog = Join-Path $OutDir "api-e2e-$timestamp.log"
$reportFile = Join-Path $OutDir "e2e-report-$timestamp.md"
$seedJson = Join-Path $OutDir 'seed-output.json'

$apiProcess = $null

function Import-DotEnv {
    param([string]$Path)

    if (-not (Test-Path $Path)) {
        return
    }

    Get-Content $Path | ForEach-Object {
        $line = $_.Trim()
        if (-not $line -or $line.StartsWith('#')) {
            return
        }

        $parts = $line -split '=', 2
        if ($parts.Count -ne 2) {
            return
        }

        $key = $parts[0].Trim()
        $value = $parts[1].Trim().Trim('"').Trim("'")
        if ($key) {
            [System.Environment]::SetEnvironmentVariable($key, $value, 'Process')
        }
    }
}

function Ensure-LocalDefaults {
    $defaults = @{
        'AWS_ACCESS_KEY' = 'e2e-access-key'
        'AWS_SECRET_KEY' = 'e2e-secret-key'
        'AWS_REGION'     = 'us-east-1'
        'AWS_S3_BUCKET'  = 'e2e-local-bucket'
    }

    foreach ($kv in $defaults.GetEnumerator()) {
        $current = [System.Environment]::GetEnvironmentVariable($kv.Key, 'Process')
        if ([string]::IsNullOrWhiteSpace($current)) {
            [System.Environment]::SetEnvironmentVariable($kv.Key, $kv.Value, 'Process')
        }
    }
}

function Invoke-HttpRequest {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Url,
        [hashtable]$Headers = @{},
        [AllowNull()]$Body
    )

    $request = [System.Net.HttpWebRequest]::Create($Url)
    $request.Method = $Method.ToUpperInvariant()
    $request.Accept = 'application/json'

    foreach ($header in $Headers.GetEnumerator()) {
        if ($header.Key -ieq 'Content-Type') {
            $request.ContentType = [string]$header.Value
            continue
        }
        $request.Headers[$header.Key] = [string]$header.Value
    }

    if (-not [string]::IsNullOrEmpty([string]$Body)) {
        if (-not $request.ContentType) {
            $request.ContentType = 'application/json'
        }
        $bodyBytes = [System.Text.Encoding]::UTF8.GetBytes([string]$Body)
        $request.ContentLength = $bodyBytes.Length
        $stream = $request.GetRequestStream()
        $stream.Write($bodyBytes, 0, $bodyBytes.Length)
        $stream.Dispose()
    }

    try {
        $response = $request.GetResponse()
    }
    catch [System.Net.WebException] {
        if ($_.Exception.Response -eq $null) {
            throw
        }
        $response = $_.Exception.Response
    }

    $statusCode = [int]([System.Net.HttpWebResponse]$response).StatusCode
    $responseStream = $response.GetResponseStream()
    $reader = New-Object System.IO.StreamReader($responseStream)
    $content = $reader.ReadToEnd()
    $reader.Dispose()
    $responseStream.Dispose()
    $response.Dispose()

    return [PSCustomObject]@{
        StatusCode = $statusCode
        Content = [string]$content
    }
}

function Test-Health {
    try {
        $resp = Invoke-HttpRequest -Method Get -Url "$ApiBaseUrl/health"
        return $resp.StatusCode -eq 200
    }
    catch {
        return $false
    }
}

function Start-ApiIfNeeded {
    if ($SkipApiStart) {
        Write-Host 'Skipping API start as requested.'
        return
    }

    if (Test-Health) {
        Write-Host "API already running at $ApiBaseUrl"
        return
    }

    $databaseUrl = [System.Environment]::GetEnvironmentVariable('DATABASE_URL', 'Process')
    if ([string]::IsNullOrWhiteSpace($databaseUrl) -or $databaseUrl -like '*user:password@localhost:5432/maid_tracking*') {
        throw 'DATABASE_URL is missing or uses placeholder credentials. Set a real PostgreSQL connection or start API separately and run with -SkipApiStart.'
    }

    Write-Host 'Starting API server...'
    $script:apiProcess = Start-Process -FilePath 'go' -ArgumentList @('run', './cmd/api') -RedirectStandardOutput $apiLog -RedirectStandardError $apiErrLog -PassThru

    for ($i = 0; $i -lt 60; $i++) {
        Start-Sleep -Seconds 1
        if (Test-Health) {
            Write-Host 'API is healthy.'
            return
        }
        if ($script:apiProcess.HasExited) {
            throw "API process exited early. Check $apiLog"
        }
    }

    throw "API did not become healthy in time. Check $apiLog"
}

function Invoke-ApiRequest {
    param(
        [Parameter(Mandatory = $true)][string]$Method,
        [Parameter(Mandatory = $true)][string]$Url,
        [AllowNull()][string]$Token,
        [AllowNull()]$Body,
        [int[]]$ExpectedStatus = @(200, 201)
    )

    $headers = @{}
    if ($Token) {
        $headers['Authorization'] = "Bearer $Token"
    }

    if ($Body -ne $null) {
        $headers['Content-Type'] = 'application/json'
        $jsonBody = $Body | ConvertTo-Json -Depth 10
        $resp = Invoke-HttpRequest -Method $Method -Url $Url -Headers $headers -Body $jsonBody
    }
    else {
        $resp = Invoke-HttpRequest -Method $Method -Url $Url -Headers $headers -Body $null
    }

    if ($ExpectedStatus -notcontains [int]$resp.StatusCode) {
        throw "API call failed [$Method $Url] status=$($resp.StatusCode) body=$($resp.Content)"
    }

    if ([string]::IsNullOrWhiteSpace($resp.Content)) {
        return $null
    }

    try {
        return $resp.Content | ConvertFrom-Json
    }
    catch {
        return $resp.Content
    }
}

function Run-ApiSmokeSuite {
    if ($SkipApiSmoke) {
        Write-Host 'Skipping API smoke suite as requested.'
        return $true
    }

    if (-not (Test-Path $seedJson)) {
        throw "Seed output not found: $seedJson"
    }

    $seed = Get-Content $seedJson -Raw | ConvertFrom-Json

    "Running API E2E smoke suite" | Tee-Object -FilePath $apiE2eLog -Append | Out-Null

    $settingsResp = Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/admin/settings" -Token $seed.admin_token -Body $null -ExpectedStatus @(200)
    if (-not $settingsResp.settings) {
        throw 'Missing platform settings payload.'
    }

    $settingsPayload = @{
        id                                    = [string]$settingsResp.settings.id
        selection_lock_duration_hours         = 12
        require_both_approvals                = [bool]$settingsResp.settings.require_both_approvals
        auto_approve_agencies                 = $false
        auto_expire_selections                = [bool]$settingsResp.settings.auto_expire_selections
        email_notifications_enabled           = [bool]$settingsResp.settings.email_notifications_enabled
        maintenance_mode                      = $false
        maintenance_message                   = [string]$settingsResp.settings.maintenance_message
        agency_approval_email_template        = [string]$settingsResp.settings.agency_approval_email_template
        agency_rejection_email_template       = [string]$settingsResp.settings.agency_rejection_email_template
        selection_notification_email_template = [string]$settingsResp.settings.selection_notification_email_template
        expiry_notification_email_template    = [string]$settingsResp.settings.expiry_notification_email_template
    }

    $updatedSettingsResp = Invoke-ApiRequest -Method Patch -Url "$ApiBaseUrl/admin/settings" -Token $seed.admin_token -Body $settingsPayload -ExpectedStatus @(200)
    if ([int]$updatedSettingsResp.settings.selection_lock_duration_hours -ne 12) {
        throw "Expected platform lock duration update to 12, got '$($updatedSettingsResp.settings.selection_lock_duration_hours)'"
    }

    $selectionResp = Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/candidates/$($seed.candidate_1_id)/select" -Token $seed.for_token -Body $null -ExpectedStatus @(201)
    if (-not $selectionResp.selection.id) {
        throw 'Missing selection id in select response.'
    }
    $selectionId = [string]$selectionResp.selection.id

    $candidateResp = Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/candidates/$($seed.candidate_1_id)" -Token $seed.eth_token -Body $null -ExpectedStatus @(200)
    if ([string]$candidateResp.candidate.status -ne 'locked') {
        throw "Expected candidate status 'locked', got '$($candidateResp.candidate.status)'"
    }

    Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/selections/$selectionId/approve" -Token $seed.eth_token -Body $null -ExpectedStatus @(200) | Out-Null
    $approvalResp = Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/selections/$selectionId/approve" -Token $seed.for_token -Body $null -ExpectedStatus @(200)

    if ($approvalResp.is_fully_approved -ne $true) {
        "Selection not fully approved yet for selection_id=$selectionId" | Tee-Object -FilePath $apiE2eLog -Append | Out-Null
    }

    Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/candidates/$($seed.candidate_1_id)/status-steps" -Token $seed.eth_token -Body $null -ExpectedStatus @(200) | Out-Null
    Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/notifications/" -Token $seed.eth_token -Body $null -ExpectedStatus @(200) | Out-Null
    Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/notifications/" -Token $seed.for_token -Body $null -ExpectedStatus @(200) | Out-Null

    "API E2E smoke suite passed" | Tee-Object -FilePath $apiE2eLog -Append | Out-Null
    return $true
}

function Run-GoTests {
    if ($SkipGoTests) {
        Write-Host 'Skipping Go tests as requested.'
        return $true
    }

    Write-Host 'Running go test ./...'
    $output = & go test ./... 2>&1
    $output | Set-Content -Path $goTestLog -Encoding UTF8
    return $LASTEXITCODE -eq 0
}

function Write-Report {
    param(
        [string]$ApiResult,
        [string]$GoTestResult
    )

    @"
# E2E Test Report

- Timestamp: $timestamp
- API Base URL: $ApiBaseUrl
- API smoke result: $ApiResult
- Go test suite result: $GoTestResult

## Artifacts

- API server log: $apiLog
- API server stderr log: $apiErrLog
- API smoke log: $apiE2eLog
- Go test log: $goTestLog
- Seed output: $seedJson

## Commands Executed

1. Started API server (if needed)
2. scripts/seed-test-data.ps1
3. API smoke workflow (select -> lock verify -> dual approve -> status/notifications checks)
4. go test ./...
"@ | Set-Content -Path $reportFile -Encoding UTF8

    Write-Host "Report written: $reportFile"
}

try {
    Import-DotEnv -Path (Join-Path $RootDir '.env')
    Ensure-LocalDefaults
    Start-ApiIfNeeded

    Write-Host 'Seeding test data...'
    & (Join-Path $PSScriptRoot 'seed-test-data.ps1') -ApiBaseUrl $ApiBaseUrl -OutDir $OutDir -IncludeCvGeneration:$IncludeCvGeneration

    $apiResult = 'PASS'
    $goResult = 'PASS'

    try {
        [void](Run-ApiSmokeSuite)
    }
    catch {
        $_ | Out-String | Tee-Object -FilePath $apiE2eLog -Append | Out-Null
        $apiResult = 'FAIL'
    }

    if (-not (Run-GoTests)) {
        $goResult = 'FAIL'
    }

    Write-Report -ApiResult $apiResult -GoTestResult $goResult

    if ($apiResult -eq 'FAIL' -or $goResult -eq 'FAIL') {
        throw "E2E run completed with failures. See $reportFile"
    }

    Write-Host "E2E run completed successfully. See $reportFile"
}
finally {
    if ($apiProcess -and -not $apiProcess.HasExited) {
        Stop-Process -Id $apiProcess.Id -Force
    }
}
