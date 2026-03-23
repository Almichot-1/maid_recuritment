param(
    [string]$ApiBaseUrl = "http://localhost:8080",
    [string]$OutDir = "./reports",
    [string]$EtEmail = "ethiopian.agent@test.com",
    [string]$ForEmail = "foreign.agent@test.com",
    [string]$TestPassword = "Password123!",
    [string]$AdminEmail = "super.admin@test.com",
    [string]$AdminPassword = "AdminPassword123!",
    [string]$AdminName = "Platform Super Admin",
    [switch]$IncludeCvGeneration
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

New-Item -ItemType Directory -Path $OutDir -Force | Out-Null

function New-TestAssetFiles {
    param([string]$Directory)

    New-Item -ItemType Directory -Path $Directory -Force | Out-Null

    $photoPath = Join-Path $Directory 'seed-photo.png'
    $passportPath = Join-Path $Directory 'seed-passport.pdf'
    $videoPath = Join-Path $Directory 'seed-video.mp4'

    $pngBase64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aS2kAAAAASUVORK5CYII='
    [System.IO.File]::WriteAllBytes($photoPath, [System.Convert]::FromBase64String($pngBase64))

    $passportContent = @"
%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 200 200] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT /F1 12 Tf 50 150 Td (Seed Passport) Tj ET
endstream
endobj
trailer
<< /Root 1 0 R >>
%%EOF
"@
    Set-Content -Path $passportPath -Value $passportContent -Encoding ASCII

    [System.IO.File]::WriteAllBytes($videoPath, [System.Text.Encoding]::ASCII.GetBytes('seed-video-placeholder'))

    return @{
        PhotoPath = $photoPath
        PassportPath = $passportPath
        VideoPath = $videoPath
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
    $request.Method = $Method
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
        throw "Request failed [$Method $Url] status=$($resp.StatusCode) body=$($resp.Content)"
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

function Seed-Admin {
    $seedScript = Join-Path $PSScriptRoot 'seed-admin.ps1'
    $result = & $seedScript -email $AdminEmail -password $AdminPassword -name $AdminName -role 'super_admin'
    if ($LASTEXITCODE -ne 0) {
        throw 'Failed to seed admin account.'
    }

    return $result | ConvertFrom-Json
}

function Get-AdminToken {
    param([pscustomobject]$AdminSeed)

    $payload = @{
        email = $AdminSeed.email
        password = $AdminPassword
        otp_code = $AdminSeed.current_otp
    }

    $result = Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/admin/login" -Body $payload -ExpectedStatus @(200)
    return [string]$result.token
}

function Get-AgencyIdByEmail {
    param(
        [string]$AdminToken,
        [string]$Email
    )

    $result = Invoke-ApiRequest -Method Get -Url "$ApiBaseUrl/admin/agencies?search=$([uri]::EscapeDataString($Email))" -Token $AdminToken -Body $null -ExpectedStatus @(200)
    $match = @($result.agencies | Where-Object { $_.email -ieq $Email } | Select-Object -First 1)
    if (-not $match) {
        throw "Could not find agency by email: $Email"
    }
    return [string]$match.id
}

function Register-User {
    param(
        [string]$AdminToken,
        [string]$Email,
        [string]$FullName,
        [string]$Role,
        [string]$CompanyName
    )

    $payload = @{
        email        = $Email
        password     = $TestPassword
        full_name    = $FullName
        role         = $Role
        company_name = $CompanyName
    }

    $resp = Invoke-HttpRequest -Method Post -Url "$ApiBaseUrl/auth/register" -Headers @{ 'Content-Type' = 'application/json' } -Body ($payload | ConvertTo-Json -Depth 10)

    if ($resp.StatusCode -eq 202) {
        $body = $resp.Content | ConvertFrom-Json
        Write-Host "Registered ${Role}: ${Email}"
        return [string]$body.user.id
    }
    if ($resp.StatusCode -eq 409) {
        Write-Host "User exists, continuing: $Email"
        return Get-AgencyIdByEmail -AdminToken $AdminToken -Email $Email
    }

    throw "Registration failed for $Email status=$($resp.StatusCode) body=$($resp.Content)"
}

function Activate-Agency {
    param(
        [string]$AdminToken,
        [string]$AgencyId
    )

    Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/admin/agencies/$AgencyId/approve" -Token $AdminToken -Body $null -ExpectedStatus @(200, 409) | Out-Null
}

function Get-Token {
    param([string]$Email)

    $payload = @{ email = $Email; password = $TestPassword }
    $result = Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/auth/login" -Body $payload -ExpectedStatus @(200)
    if (-not $result.token) {
        throw "Login token missing for $Email"
    }
    return [string]$result.token
}

function New-Candidate {
    param(
        [string]$Token,
        [string]$FullName,
        [int]$Age,
        [int]$ExperienceYears
    )

    $payload = @{
        full_name        = $FullName
        age              = $Age
        experience_years = $ExperienceYears
        languages        = @('English', 'Amharic')
        skills           = @('Cooking', 'Cleaning')
    }

    $result = Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/candidates/" -Token $Token -Body $payload -ExpectedStatus @(201)
    if (-not $result.candidate.id) {
        throw "Candidate ID missing in create response"
    }
    return [string]$result.candidate.id
}

function Add-CandidateDocument {
    param(
        [string]$Token,
        [string]$CandidateId,
        [string]$DocumentType,
        [string]$FilePath,
        [string]$ContentType
    )

    $response = & curl.exe -sS -X POST "$ApiBaseUrl/candidates/$CandidateId/documents" `
        -H "Authorization: Bearer $Token" `
        -F "document_type=$DocumentType" `
        -F "file=@$FilePath;type=$ContentType"

    if ($LASTEXITCODE -ne 0) {
        throw "curl upload failed for $DocumentType on candidate $CandidateId"
    }

    try {
        $result = $response | ConvertFrom-Json
    }
    catch {
        throw "Upload response was not valid JSON for ${DocumentType}: $response"
    }

    if (-not $result.document.id) {
        throw "Upload failed for ${DocumentType}: $response"
    }
}

function Publish-Candidate {
    param([string]$Token, [string]$CandidateId)

    Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/candidates/$CandidateId/publish" -Token $Token -Body $null -ExpectedStatus @(200) | Out-Null
}

function Generate-Cv {
    param([string]$Token, [string]$CandidateId)

    Invoke-ApiRequest -Method Post -Url "$ApiBaseUrl/candidates/$CandidateId/generate-cv" -Token $Token -Body $null -ExpectedStatus @(200) | Out-Null
}

Write-Host "Seeding admin, agencies, and candidates against $ApiBaseUrl"

$adminSeed = Seed-Admin
$adminToken = Get-AdminToken -AdminSeed $adminSeed
$assets = New-TestAssetFiles -Directory (Join-Path $OutDir 'seed-assets')

$ethiopianAgencyId = Register-User -AdminToken $adminToken -Email $EtEmail -FullName 'E2E Ethiopian Agent' -Role 'ethiopian_agent' -CompanyName 'Addis Placement'
$foreignAgencyId = Register-User -AdminToken $adminToken -Email $ForEmail -FullName 'E2E Foreign Agent' -Role 'foreign_agent' -CompanyName 'Global Agency'

Activate-Agency -AdminToken $adminToken -AgencyId $ethiopianAgencyId
Activate-Agency -AdminToken $adminToken -AgencyId $foreignAgencyId

$ethToken = Get-Token -Email $EtEmail
$forToken = Get-Token -Email $ForEmail

$candidate1Id = New-Candidate -Token $ethToken -FullName 'E2E Candidate One' -Age 24 -ExperienceYears 3
$candidate2Id = New-Candidate -Token $ethToken -FullName 'E2E Candidate Two' -Age 29 -ExperienceYears 5

Add-CandidateDocument -Token $ethToken -CandidateId $candidate1Id -DocumentType 'passport' -FilePath $assets.PassportPath -ContentType 'application/pdf'
Add-CandidateDocument -Token $ethToken -CandidateId $candidate1Id -DocumentType 'photo' -FilePath $assets.PhotoPath -ContentType 'image/png'
Add-CandidateDocument -Token $ethToken -CandidateId $candidate1Id -DocumentType 'video' -FilePath $assets.VideoPath -ContentType 'video/mp4'

if ($IncludeCvGeneration) {
    Generate-Cv -Token $ethToken -CandidateId $candidate1Id
}

Publish-Candidate -Token $ethToken -CandidateId $candidate1Id
Publish-Candidate -Token $ethToken -CandidateId $candidate2Id

$seedOutput = [ordered]@{
    api_base_url         = $ApiBaseUrl
    admin_email          = $AdminEmail
    admin_password       = $AdminPassword
    admin_mfa_secret     = $adminSeed.mfa_secret
    et_email             = $EtEmail
    for_email            = $ForEmail
    et_agency_id         = $ethiopianAgencyId
    for_agency_id        = $foreignAgencyId
    admin_token          = $adminToken
    eth_token            = $ethToken
    for_token            = $forToken
    candidate_1_id       = $candidate1Id
    candidate_2_id       = $candidate2Id
}

$seedOutput | ConvertTo-Json -Depth 10 | Set-Content -Path (Join-Path $OutDir 'seed-output.json') -Encoding UTF8

@(
    "API_BASE_URL=$ApiBaseUrl"
    "ADMIN_EMAIL=$AdminEmail"
    "ADMIN_PASSWORD=$AdminPassword"
    "ET_EMAIL=$EtEmail"
    "FOR_EMAIL=$ForEmail"
    "ETH_TOKEN=$ethToken"
    "FOR_TOKEN=$forToken"
    "CANDIDATE_1_ID=$candidate1Id"
    "CANDIDATE_2_ID=$candidate2Id"
) | Set-Content -Path (Join-Path $OutDir 'seed-output.env') -Encoding UTF8

Write-Host "Seed complete. Outputs:"
Write-Host "- $(Join-Path $OutDir 'seed-output.json')"
Write-Host "- $(Join-Path $OutDir 'seed-output.env')"
