.PHONY: build run test lint tidy migrate-up migrate-down migrate-create docker-build docker-run clean

APP_NAME := story
BUILD_DIR := ./build

help:
	@echo "Story - CLI-first developer second brain"
	@echo ""
	@echo "Usage:"
	@echo "  make build              Build the binary"
	@echo "  make run                Run the CLI"
	@echo "  make test               Run all tests"
	@echo "  make lint               Run linter"
	@echo "  make tidy               Tidy Go modules"
	@echo "  make migrate-up         Run database migrations"
	@echo "  make migrate-down       Rollback database migrations"
	@echo "  make migrate-create     Create a new migration"
	@echo "  make docker-build       Build Docker image"
	@echo "  make docker-run         Start all services with Docker Compose"
	@echo "  make clean              Clean build artifacts"

build:
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/$(APP_NAME)/

run:
	go run ./cmd/$(APP_NAME)/

test:
	go test ./internal/... -v -count=1 -race

test-cover:
	go test ./internal/... -v -count=1 -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
	go mod verify

migrate-up:
	goose -dir migrations postgres "$$STORY_DATABASE_DSN" up

migrate-down:
	goose -dir migrations postgres "$$STORY_DATABASE_DSN" down

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir migrations create $$name sql

docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker compose up -d --build

docker-stop:
	docker compose down

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

fmt:
	go fmt ./...
