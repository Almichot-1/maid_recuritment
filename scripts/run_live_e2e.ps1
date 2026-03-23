$ErrorActionPreference = "Stop"

$base = "http://127.0.0.1:8080/api/v1"
$artifactDir = "C:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\.dev-artifacts\e2e"
New-Item -ItemType Directory -Force -Path $artifactDir | Out-Null

$pngBase64 = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+aS2kAAAAASUVORK5CYII="
$photoPath = Join-Path $artifactDir "candidate-photo.png"
[System.IO.File]::WriteAllBytes($photoPath, [System.Convert]::FromBase64String($pngBase64))

$passportPath = Join-Path $artifactDir "candidate-passport.pdf"
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
BT /F1 12 Tf 50 150 Td (Passport) Tj ET
endstream
endobj
trailer
<< /Root 1 0 R >>
%%EOF
"@
Set-Content -Path $passportPath -Value $passportContent -Encoding ASCII

$stamp = Get-Date -Format "yyyyMMddHHmmss"
$ethiopianEmail = "ethiopian.$stamp@example.test"
$foreignEmail = "foreign.$stamp@example.test"
$password = "Password123!"

function Post-Json {
	param(
		[string]$Url,
		[object]$Body,
		[string]$Token
	)

	$headers = @{}
	if ($Token) {
		$headers["Authorization"] = "Bearer $Token"
	}

	return Invoke-RestMethod -Method Post -Uri $Url -Headers $headers -ContentType "application/json" -Body ($Body | ConvertTo-Json -Depth 10)
}

function Get-Json {
	param(
		[string]$Url,
		[string]$Token
	)

	$headers = @{}
	if ($Token) {
		$headers["Authorization"] = "Bearer $Token"
	}

	return Invoke-RestMethod -Method Get -Uri $Url -Headers $headers
}

$ethiopianRegister = Post-Json "$base/auth/register" @{
	email = $ethiopianEmail
	password = $password
	full_name = "Ethiopian Agent Live"
	role = "ethiopian_agent"
	company_name = "Addis Placement"
} ""

$foreignRegister = Post-Json "$base/auth/register" @{
	email = $foreignEmail
	password = $password
	full_name = "Foreign Agent Live"
	role = "foreign_agent"
	company_name = "Global Households"
} ""

$ethToken = $ethiopianRegister.token
$forToken = $foreignRegister.token

$candidateCreate = Post-Json "$base/candidates" @{
	full_name = "Live Candidate Demo"
	age = 26
	experience_years = 4
	languages = @("Amharic", "English")
	skills = @("Cleaning", "Childcare")
} $ethToken
$candidateId = $candidateCreate.candidate.id

$passportJson = & curl.exe -sS -X POST "$base/candidates/$candidateId/documents" -H "Authorization: Bearer $ethToken" -F "document_type=passport" -F "file=@$passportPath;type=application/pdf"
$passportUpload = $passportJson | ConvertFrom-Json

$photoJson = & curl.exe -sS -X POST "$base/candidates/$candidateId/documents" -H "Authorization: Bearer $ethToken" -F "document_type=photo" -F "file=@$photoPath;type=image/png"
$photoUpload = $photoJson | ConvertFrom-Json

$cvGenerate = Post-Json "$base/candidates/$candidateId/generate-cv" @{} $ethToken
$publishResponse = Post-Json "$base/candidates/$candidateId/publish" @{} $ethToken

$foreignList = Get-Json "$base/candidates?page=1&page_size=20" $forToken
$foreignVisible = @($foreignList.candidates | Where-Object { $_.id -eq $candidateId }).Count -gt 0

$selectionCreate = Post-Json "$base/candidates/$candidateId/select" @{} $forToken
$selectionId = $selectionCreate.selection.id

$foreignApprove = Post-Json "$base/selections/$selectionId/approve" @{} $forToken
$ethiopianApprove = Post-Json "$base/selections/$selectionId/approve" @{} $ethToken

$selectionDetail = Get-Json "$base/selections/$selectionId" $ethToken
$approvalStatus = Get-Json "$base/selections/$selectionId/approvals" $ethToken
$candidateDetail = Get-Json "$base/candidates/$candidateId" $ethToken
$ethNotifications = Get-Json "$base/notifications" $ethToken
$forNotifications = Get-Json "$base/notifications" $forToken

Start-Sleep -Seconds 2
$smtpLog = "C:\Users\NOOR AL MUSABAH\Documents\PROJECT_2\smtp-live.log"
$smtpMessages = 0
if (Test-Path $smtpLog) {
	$smtpMessages = (Select-String -Path $smtpLog -Pattern "^--- MESSAGE ---$" | Measure-Object).Count
}

[PSCustomObject]@{
	ethiopian_email = $ethiopianEmail
	foreign_email = $foreignEmail
	candidate_id = $candidateId
	candidate_status = $candidateDetail.candidate.status
	cv_pdf_url = $candidateDetail.candidate.cv_pdf_url
	document_types = @($candidateDetail.candidate.documents | ForEach-Object { $_.document_type })
	passport_upload_id = $passportUpload.document.id
	photo_upload_id = $photoUpload.document.id
	foreign_can_see_candidate = $foreignVisible
	selection_id = $selectionId
	selection_status = $selectionDetail.selection.status
	is_fully_approved = $approvalStatus.is_fully_approved
	pending_approval_from = @($approvalStatus.pending_approval_from)
	ethiopian_notifications = @($ethNotifications.notifications).Count
	foreign_notifications = @($forNotifications.notifications).Count
	smtp_messages_logged = $smtpMessages
} | ConvertTo-Json -Depth 10
