package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	RedisURL     string
	AWSAccessKey string
	AWSSecretKey string
	AWSRegion    string
	S3Bucket     string
	S3Endpoint   string
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPass     string
	AppBaseURL   string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		JWTSecret:    os.Getenv("JWT_SECRET"),
		RedisURL:     os.Getenv("REDIS_URL"),
		AWSAccessKey: os.Getenv("AWS_ACCESS_KEY"),
		AWSSecretKey: os.Getenv("AWS_SECRET_KEY"),
		AWSRegion:    os.Getenv("AWS_REGION"),
		S3Bucket:     firstNonEmpty(os.Getenv("S3_BUCKET"), os.Getenv("AWS_S3_BUCKET")),
		S3Endpoint:   os.Getenv("S3_ENDPOINT"),
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPUser:     os.Getenv("SMTP_USER"),
		SMTPPass:     os.Getenv("SMTP_PASS"),
		AppBaseURL:   os.Getenv("APP_BASE_URL"),
	}

	missing := make([]string, 0)
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if cfg.RedisURL == "" {
		missing = append(missing, "REDIS_URL")
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
