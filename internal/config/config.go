package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port               string
	DatabaseURL        string
	JWTSecret          string
	RedisURL           string
	AWSAccessKey       string
	AWSSecretKey       string
	AWSRegion          string
	S3Bucket           string
	S3Endpoint         string
	S3PublicBaseURL    string
	SMTPHost           string
	SMTPPort           string
	SMTPUser           string
	SMTPPass           string
	SMTPFromEmail      string
	SMTPFromName       string
	AppBaseURL         string
	CORSAllowedOrigins []string
	RunExpiryScheduler bool
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		JWTSecret:          os.Getenv("JWT_SECRET"),
		RedisURL:           os.Getenv("REDIS_URL"),
		AWSAccessKey:       os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretKey:       os.Getenv("AWS_SECRET_KEY"),
		AWSRegion:          os.Getenv("AWS_REGION"),
		S3Bucket:           firstNonEmpty(os.Getenv("S3_BUCKET"), os.Getenv("AWS_S3_BUCKET")),
		S3Endpoint:         os.Getenv("S3_ENDPOINT"),
		S3PublicBaseURL:    os.Getenv("S3_PUBLIC_BASE_URL"),
		SMTPHost:           os.Getenv("SMTP_HOST"),
		SMTPPort:           os.Getenv("SMTP_PORT"),
		SMTPUser:           os.Getenv("SMTP_USER"),
		SMTPPass:           os.Getenv("SMTP_PASS"),
		SMTPFromEmail:      os.Getenv("SMTP_FROM_EMAIL"),
		SMTPFromName:       os.Getenv("SMTP_FROM_NAME"),
		AppBaseURL:         os.Getenv("APP_BASE_URL"),
		CORSAllowedOrigins: splitCSV(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:3001")),
		RunExpiryScheduler: getEnvAsBool("RUN_EXPIRY_SCHEDULER", true),
	}

	missing := make([]string, 0)
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if cfg.AWSAccessKey == "" {
		missing = append(missing, "AWS_ACCESS_KEY")
	}
	if cfg.AWSSecretKey == "" {
		missing = append(missing, "AWS_SECRET_KEY")
	}
	if cfg.AWSRegion == "" {
		missing = append(missing, "AWS_REGION")
	}
	if cfg.S3Bucket == "" {
		missing = append(missing, "S3_BUCKET")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func getEnvAsBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}

	value, err := strconv.ParseBool(raw)
	if err != nil {
		return fallback
	}
	return value
}
