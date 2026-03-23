.PHONY: run run-api run-expiry-worker build-api build-expiry-worker migrate-up migrate-down test

run:
	go run ./cmd/api

run-api:
	go run ./cmd/api

run-expiry-worker:
	go run ./cmd/expiryworker

build-api:
	go build -o ./bin/api ./cmd/api

build-expiry-worker:
	go build -o ./bin/expiryworker ./cmd/expiryworker

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

test:
	go test ./...
