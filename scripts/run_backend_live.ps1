function Set-DefaultEnv([string]$name, [string]$value) {
	$current = [System.Environment]::GetEnvironmentVariable($name, "Process")
	if ([string]::IsNullOrWhiteSpace($current)) {
		Set-Item -Path "Env:$name" -Value $value
	}
}

Set-DefaultEnv "PORT" "8080"
Set-DefaultEnv "DATABASE_URL" "postgres://maid_app:maid_app_pw@127.0.0.1:55432/maid_tracking?sslmode=disable"
Set-DefaultEnv "JWT_SECRET" "dev-jwt-secret"
Set-DefaultEnv "REDIS_URL" "redis://127.0.0.1:6379/0"
Set-DefaultEnv "AWS_ACCESS_KEY" "dev-access"
Set-DefaultEnv "AWS_SECRET_KEY" "dev-secret"
Set-DefaultEnv "AWS_REGION" "us-east-1"
Set-DefaultEnv "AWS_S3_BUCKET" "e2e-local-bucket"
Set-DefaultEnv "S3_ENDPOINT" "http://127.0.0.1:9000"
Set-DefaultEnv "SMTP_HOST" "127.0.0.1"
Set-DefaultEnv "SMTP_PORT" "1025"
Set-DefaultEnv "SMTP_USER" "noreply@example.test"
Set-DefaultEnv "SMTP_PASS" "dev-pass"
Set-DefaultEnv "APP_BASE_URL" "http://localhost:3001"

go run ./cmd/api
