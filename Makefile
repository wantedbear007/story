# ─────────────────────────────────────────────────────────
# Story — Makefile
# Targets for development, testing, building, and deployment
# ─────────────────────────────────────────────────────────

.PHONY: help build run test test-cover lint tidy fmt \
        migrate-up migrate-down migrate-create migrate-status \
        docker-build docker-run docker-stop clean

APP_NAME   := story
BUILD_DIR  := ./build
MAIN_PKG   := ./cmd/$(APP_NAME)/
MIGRATIONS := ./migrations
CONFIG_PATH := configs/config.yaml

# ── Database DSN ─────────────────────────────────────────
# Uses STORY_DATABASE_DSN if set, otherwise constructs from config values.
# Override when running against non-local databases:
#   export STORY_DATABASE_DSN="postgres://user:pass@host:5432/db?sslmode=require"
DB_DSN ?= $(STORY_DATABASE_DSN)

help:
	@echo 'Story - CLI-first developer second brain'
	@echo ''
	@echo 'Development:'
	@echo '  make build           Build the binary into ./build/'
	@echo '  make run             Run the application'
	@echo '  make test            Run all unit tests (race detector on)'
	@echo '  make test-cover      Run tests with coverage report'
	@echo '  make lint            Run golangci-lint'
	@echo '  make tidy            Tidy and verify Go modules'
	@echo '  make fmt             Format all Go source files'
	@echo ''
	@echo 'Database:'
	@echo '  make migrate-up      Apply all pending migrations'
	@echo '  make migrate-down    Rollback the last migration'
	@echo '  make migrate-status  Show migration status'
	@echo '  make migrate-create  Create a new migration file'
	@echo ''
	@echo 'Docker:'
	@echo '  make docker-build    Build the Docker image'
	@echo '  make docker-run      Start all services (docker compose)'
	@echo '  make docker-stop     Stop all services'
	@echo ''
	@echo 'Cleanup:'
	@echo '  make clean           Remove build artifacts'

build:
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PKG)
	@echo "Binary built: $(BUILD_DIR)/$(APP_NAME)"

run:
	go run $(MAIN_PKG)

test:
	go test ./internal/... -v -count=1 -race

test-cover:
	go test ./internal/... -v -count=1 -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	golangci-lint run ./...

tidy:
	go mod tidy
	go mod verify

fmt:
	go fmt ./...

# ── Database Migrations (goose) ──────────────────────────
# Requires a DSN. Export STORY_DATABASE_DSN or pass DB_DSN:
#   make migrate-up DB_DSN="postgres://..."

migrate-up:
	@[ -n "$(DB_DSN)" ] || (echo "Error: DB_DSN is not set. Export STORY_DATABASE_DSN or pass DB_DSN=..." ; exit 1)
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" up

migrate-down:
	@[ -n "$(DB_DSN)" ] || (echo "Error: DB_DSN is not set." ; exit 1)
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" down

migrate-status:
	@[ -n "$(DB_DSN)" ] || (echo "Error: DB_DSN is not set." ; exit 1)
	goose -dir $(MIGRATIONS) postgres "$(DB_DSN)" status

migrate-create:
	@read -p "Migration name: " name; \
	goose -dir $(MIGRATIONS) create $$name sql

# ── Docker ───────────────────────────────────────────────
docker-build:
	docker build -t $(APP_NAME) .

docker-run:
	docker compose up -d --build

docker-stop:
	docker compose down

# ── Cleanup ──────────────────────────────────────────────
clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html
