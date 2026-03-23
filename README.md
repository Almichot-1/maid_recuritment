# Maid Recruitment Tracking Platform (Go Backend)

A Clean Architecture project skeleton for a maid recruitment tracking platform.

## Project Structure

- `cmd/api` - API entry point
- `internal/domain` - Core entities and interfaces
- `internal/repository` - Data access layer
- `internal/service` - Business logic layer
- `internal/handler` - HTTP handlers
- `internal/middleware` - Cross-cutting concerns (auth, logging)
- `internal/config` - Configuration loader
- `pkg/utils` - Shared utility functions
- `migrations` - SQL migration files

## Prerequisites

- Go 1.22+
- `make`
- Optional for migrations: [golang-migrate](https://github.com/golang-migrate/migrate)

## Setup

1. Install dependencies:
   ```bash
   go mod tidy
   ```
2. Copy/edit environment variables in `.env`.
3. Run the API:
   ```bash
   make run
   ```

## Environment Variables

- `PORT` (default: `8080`)
- `DATABASE_URL`
- `JWT_SECRET`
- `AWS_S3_BUCKET`
- `REDIS_URL`

## Available Make Commands

- `make run` - Start API server
- `make migrate-up` - Apply SQL migrations
- `make migrate-down` - Rollback one migration step
- `make test` - Run tests

## Current Endpoints

- `GET /health` - Health check endpoint
