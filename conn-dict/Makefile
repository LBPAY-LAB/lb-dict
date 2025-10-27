.PHONY: help build clean test lint fmt vet check docker-build docker-up docker-down docker-logs migrate-up migrate-down run-worker run-server

# Colors for output
BLUE := \033[0;34m
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

# Project variables
PROJECT_NAME := conn-dict
WORKER_BIN := bin/worker
SERVER_BIN := bin/server
GO_FILES := $(shell find . -name '*.go' -type f -not -path './vendor/*')
MIGRATION_DIR := migrations

# Database variables
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= dict_user
DB_PASSWORD ?= dict_password
DB_NAME ?= dict_db
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# Tools
GOLANGCI_LINT := golangci-lint
GOOSE := goose

## help: Show this help message
help:
	@echo '$(BLUE)═══════════════════════════════════════════════════════════$(NC)'
	@echo '$(BLUE)   $(PROJECT_NAME) - Makefile Targets$(NC)'
	@echo '$(BLUE)═══════════════════════════════════════════════════════════$(NC)'
	@echo ''
	@echo '$(GREEN)Build:$(NC)'
	@echo '  make build         - Build worker and server binaries'
	@echo '  make build-worker  - Build worker binary'
	@echo '  make build-server  - Build server binary (when implemented)'
	@echo '  make clean         - Remove build artifacts'
	@echo ''
	@echo '$(GREEN)Testing:$(NC)'
	@echo '  make test          - Run unit tests'
	@echo '  make test-coverage - Run tests with coverage report'
	@echo '  make test-verbose  - Run tests with verbose output'
	@echo ''
	@echo '$(GREEN)Code Quality:$(NC)'
	@echo '  make lint          - Run golangci-lint'
	@echo '  make fmt           - Format code with gofmt'
	@echo '  make vet           - Run go vet'
	@echo '  make check         - Run fmt + vet + lint'
	@echo ''
	@echo '$(GREEN)Database:$(NC)'
	@echo '  make migrate-up    - Run database migrations'
	@echo '  make migrate-down  - Rollback last migration'
	@echo '  make migrate-status - Show migration status'
	@echo '  make migrate-create NAME=<name> - Create new migration'
	@echo ''
	@echo '$(GREEN)Docker:$(NC)'
	@echo '  make docker-build  - Build Docker image'
	@echo '  make docker-up     - Start all services'
	@echo '  make docker-down   - Stop all services'
	@echo '  make docker-logs   - View logs'
	@echo '  make docker-clean  - Remove volumes and networks'
	@echo ''
	@echo '$(GREEN)Development:$(NC)'
	@echo '  make run-worker    - Run worker locally'
	@echo '  make deps          - Download dependencies'
	@echo '  make tidy          - Tidy dependencies'
	@echo ''

## build: Build all binaries
build: check-go build-worker
	@echo "$(GREEN)✓ Build complete$(NC)"

## build-worker: Build worker binary
build-worker:
	@echo "$(BLUE)Building worker...$(NC)"
	@mkdir -p bin
	@go build -o $(WORKER_BIN) ./cmd/worker
	@echo "$(GREEN)✓ Worker built: $(WORKER_BIN)$(NC)"

## build-server: Build server binary (placeholder)
build-server:
	@echo "$(YELLOW)Server not yet implemented$(NC)"

## clean: Remove build artifacts
clean:
	@echo "$(BLUE)Cleaning build artifacts...$(NC)"
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "$(GREEN)✓ Clean complete$(NC)"

## test: Run all tests with coverage
test: check-go
	@echo "$(BLUE)Running tests...$(NC)"
	@go test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)✓ Tests passed$(NC)"

## test-unit: Run unit tests only
test-unit: check-go
	@echo "$(BLUE)Running unit tests...$(NC)"
	@go test -v -short ./...
	@echo "$(GREEN)✓ Unit tests passed$(NC)"

## test-integration: Run integration tests
test-integration: check-go
	@echo "$(BLUE)Running integration tests...$(NC)"
	@go test -v -run Integration ./tests/integration/...
	@echo "$(GREEN)✓ Integration tests passed$(NC)"

## coverage: Generate coverage report
coverage: check-go
	@echo "$(BLUE)Generating coverage report...$(NC)"
	@go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: coverage.html$(NC)"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## test-coverage: Alias for coverage
test-coverage: coverage

## test-verbose: Run tests with verbose output
test-verbose: check-go
	@echo "$(BLUE)Running verbose tests...$(NC)"
	@go test -v -race -timeout 30s ./... 2>&1 | tee test-output.log
	@echo "$(GREEN)✓ Tests complete$(NC)"

## lint: Run golangci-lint
lint: check-golangci-lint
	@echo "$(BLUE)Running linter...$(NC)"
	@$(GOLANGCI_LINT) run --timeout 5m
	@echo "$(GREEN)✓ Lint passed$(NC)"

## fmt: Format code
fmt: check-go
	@echo "$(BLUE)Formatting code...$(NC)"
	@gofmt -w -s $(GO_FILES)
	@echo "$(GREEN)✓ Format complete$(NC)"

## vet: Run go vet
vet: check-go
	@echo "$(BLUE)Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✓ Vet passed$(NC)"

## check: Run all checks (fmt + vet + lint)
check: fmt vet lint
	@echo "$(GREEN)✓ All checks passed$(NC)"

## migrate-up: Run database migrations
migrate-up: check-goose
	@echo "$(BLUE)Running migrations...$(NC)"
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DB_URL)" up
	@echo "$(GREEN)✓ Migrations complete$(NC)"

## migrate-down: Rollback last migration
migrate-down: check-goose
	@echo "$(YELLOW)Rolling back last migration...$(NC)"
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DB_URL)" down
	@echo "$(GREEN)✓ Rollback complete$(NC)"

## migrate-status: Show migration status
migrate-status: check-goose
	@echo "$(BLUE)Migration status:$(NC)"
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DB_URL)" status

## migrate-create: Create new migration (usage: make migrate-create NAME=create_foo_table)
migrate-create: check-goose
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=create_foo_table$(NC)"; \
		exit 1; \
	fi
	@echo "$(BLUE)Creating migration: $(NAME)...$(NC)"
	@$(GOOSE) -dir $(MIGRATION_DIR) create $(NAME) sql
	@echo "$(GREEN)✓ Migration created$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(BLUE)Building Docker image...$(NC)"
	@docker build -t $(PROJECT_NAME):latest .
	@echo "$(GREEN)✓ Docker image built$(NC)"

## docker-up: Start all services
docker-up:
	@echo "$(BLUE)Starting services...$(NC)"
	@docker-compose up -d
	@echo "$(GREEN)✓ Services started$(NC)"
	@echo "Run 'make docker-logs' to view logs"

## docker-down: Stop all services
docker-down:
	@echo "$(BLUE)Stopping services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✓ Services stopped$(NC)"

## docker-logs: View logs
docker-logs:
	@docker-compose logs -f

## docker-clean: Remove volumes and networks
docker-clean:
	@echo "$(YELLOW)Removing volumes and networks...$(NC)"
	@docker-compose down -v --remove-orphans
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

## run-worker: Run worker locally
run-worker: build-worker
	@echo "$(BLUE)Starting worker...$(NC)"
	@$(WORKER_BIN)

## deps: Download dependencies
deps: check-go
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	@go mod download
	@echo "$(GREEN)✓ Dependencies downloaded$(NC)"

## tidy: Tidy dependencies
tidy: check-go
	@echo "$(BLUE)Tidying dependencies...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✓ Dependencies tidied$(NC)"

## check-go: Verify Go installation
check-go:
	@command -v go >/dev/null 2>&1 || { echo "$(RED)Error: Go is not installed$(NC)"; exit 1; }

## check-golangci-lint: Verify golangci-lint installation
check-golangci-lint:
	@command -v $(GOLANGCI_LINT) >/dev/null 2>&1 || { \
		echo "$(YELLOW)Warning: golangci-lint not found. Installing...$(NC)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}

## check-goose: Verify goose installation
check-goose:
	@command -v $(GOOSE) >/dev/null 2>&1 || { \
		echo "$(YELLOW)Warning: goose not found. Installing...$(NC)"; \
		go install github.com/pressly/goose/v3/cmd/goose@latest; \
	}

.DEFAULT_GOAL := help
