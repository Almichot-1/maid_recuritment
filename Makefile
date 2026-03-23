.PHONY: run migrate-up migrate-down test

run:
	go run ./cmd/api

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

test:
	go test ./...
