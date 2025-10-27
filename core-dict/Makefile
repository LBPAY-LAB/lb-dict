# ============================================
# Core DICT - Makefile
# ============================================

# Variables
APP_NAME := core-dict
VERSION := 1.0.0
GO_VERSION := 1.24.5
DOCKER_IMAGE := lbpay/core-dict
DOCKER_TAG := latest

# Paths
BIN_DIR := ./bin
CMD_DIR := ./cmd/server
MIGRATIONS_DIR := ./migrations
PROTO_DIR := ./api/proto

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GORUN := $(GOCMD) run
GOCLEAN := $(GOCMD) clean

# Tools
GOLANGCI_LINT := golangci-lint
PROTOC := protoc
GOOSE := goose
AIR := air

# Build flags
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

.PHONY: help
help: ## Show this help message
	@echo "$(COLOR_BOLD)Core DICT - Makefile Commands$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""

# ============================================
# Development
# ============================================

.PHONY: install
install: ## Install all dependencies
	@echo "$(COLOR_BLUE)Installing dependencies...$(COLOR_RESET)"
	$(GOGET) -v ./...
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Dependencies installed!$(COLOR_RESET)"

.PHONY: run
run: ## Run the application in development mode
	@echo "$(COLOR_BLUE)Running application in development mode...$(COLOR_RESET)"
	$(GORUN) $(CMD_DIR)/main.go

.PHONY: dev
dev: ## Run the application with hot-reload (requires air)
	@echo "$(COLOR_BLUE)Running application with hot-reload...$(COLOR_RESET)"
	$(AIR) -c .air.toml

.PHONY: build
build: ## Build the application binary
	@echo "$(COLOR_BLUE)Building application...$(COLOR_RESET)"
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(APP_NAME) $(CMD_DIR)/main.go
	@echo "$(COLOR_GREEN)Build complete: $(BIN_DIR)/$(APP_NAME)$(COLOR_RESET)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning build artifacts...$(COLOR_RESET)"
	$(GOCLEAN)
	@rm -rf $(BIN_DIR)
	@echo "$(COLOR_GREEN)Clean complete!$(COLOR_RESET)"

# ============================================
# Testing
# ============================================

.PHONY: test
test: ## Run unit tests
	@echo "$(COLOR_BLUE)Running unit tests...$(COLOR_RESET)"
	$(GOTEST) -v -race -timeout 30s ./...
	@echo "$(COLOR_GREEN)Tests passed!$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Running tests with coverage...$(COLOR_RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)Coverage report generated: coverage.html$(COLOR_RESET)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	$(GOTEST) -v -race -tags=integration -timeout 5m ./test/integration/...
	@echo "$(COLOR_GREEN)Integration tests passed!$(COLOR_RESET)"

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(COLOR_BLUE)Running e2e tests...$(COLOR_RESET)"
	$(GOTEST) -v -race -tags=e2e -timeout 10m ./test/e2e/...
	@echo "$(COLOR_GREEN)E2E tests passed!$(COLOR_RESET)"

# ============================================
# Code Quality
# ============================================

.PHONY: lint
lint: ## Run linter (golangci-lint)
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	$(GOLANGCI_LINT) run --config .golangci.yml ./...
	@echo "$(COLOR_GREEN)Linting complete!$(COLOR_RESET)"

.PHONY: fmt
fmt: ## Format Go code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	$(GOCMD) fmt ./...
	@echo "$(COLOR_GREEN)Code formatted!$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	$(GOCMD) vet ./...
	@echo "$(COLOR_GREEN)Vet complete!$(COLOR_RESET)"

.PHONY: check
check: fmt vet lint ## Run all code quality checks

# ============================================
# Database
# ============================================

.PHONY: migrate-up
migrate-up: ## Apply all database migrations
	@echo "$(COLOR_BLUE)Applying database migrations...$(COLOR_RESET)"
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$${DATABASE_URL}" up
	@echo "$(COLOR_GREEN)Migrations applied!$(COLOR_RESET)"

.PHONY: migrate-down
migrate-down: ## Rollback last database migration
	@echo "$(COLOR_YELLOW)Rolling back last migration...$(COLOR_RESET)"
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$${DATABASE_URL}" down
	@echo "$(COLOR_GREEN)Migration rolled back!$(COLOR_RESET)"

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "$(COLOR_BLUE)Migration status:$(COLOR_RESET)"
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$${DATABASE_URL}" status

.PHONY: migrate-create
migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "$(COLOR_BLUE)Creating new migration...$(COLOR_RESET)"
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $(NAME) sql
	@echo "$(COLOR_GREEN)Migration created!$(COLOR_RESET)"

# ============================================
# Docker
# ============================================

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) -t $(DOCKER_IMAGE):$(VERSION) .
	@echo "$(COLOR_GREEN)Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)$(COLOR_RESET)"

.PHONY: docker-run
docker-run: ## Run application in Docker container
	@echo "$(COLOR_BLUE)Running Docker container...$(COLOR_RESET)"
	docker run --rm -it \
		--env-file .env \
		-p 8080:8080 \
		-p 9090:9090 \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image to registry
	@echo "$(COLOR_BLUE)Pushing Docker image...$(COLOR_RESET)"
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):$(VERSION)
	@echo "$(COLOR_GREEN)Docker image pushed!$(COLOR_RESET)"

.PHONY: docker-compose-up
docker-compose-up: ## Start all services with docker-compose
	@echo "$(COLOR_BLUE)Starting services with docker-compose...$(COLOR_RESET)"
	docker-compose up -d
	@echo "$(COLOR_GREEN)Services started!$(COLOR_RESET)"

.PHONY: docker-compose-down
docker-compose-down: ## Stop all services with docker-compose
	@echo "$(COLOR_BLUE)Stopping services with docker-compose...$(COLOR_RESET)"
	docker-compose down
	@echo "$(COLOR_GREEN)Services stopped!$(COLOR_RESET)"

.PHONY: docker-compose-logs
docker-compose-logs: ## Show docker-compose logs
	docker-compose logs -f

# ============================================
# Protobuf / gRPC
# ============================================

.PHONY: proto-generate
proto-generate: ## Generate Go code from protobuf files
	@echo "$(COLOR_BLUE)Generating Go code from protobuf...$(COLOR_RESET)"
	$(PROTOC) --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/dict/v1/*.proto
	@echo "$(COLOR_GREEN)Protobuf code generated!$(COLOR_RESET)"

.PHONY: proto-lint
proto-lint: ## Lint protobuf files
	@echo "$(COLOR_BLUE)Linting protobuf files...$(COLOR_RESET)"
	buf lint $(PROTO_DIR)
	@echo "$(COLOR_GREEN)Protobuf linting complete!$(COLOR_RESET)"

# ============================================
# Tools Installation
# ============================================

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/pressly/goose/v3/cmd/goose@latest
	go install github.com/cosmtrek/air@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "$(COLOR_GREEN)Tools installed!$(COLOR_RESET)"

# ============================================
# Miscellaneous
# ============================================

.PHONY: version
version: ## Show application version
	@echo "$(COLOR_BOLD)Core DICT$(COLOR_RESET)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"

.PHONY: deps-update
deps-update: ## Update all dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "$(COLOR_GREEN)Dependencies updated!$(COLOR_RESET)"

.PHONY: security-scan
security-scan: ## Run security scan (gosec)
	@echo "$(COLOR_BLUE)Running security scan...$(COLOR_RESET)"
	gosec ./...
	@echo "$(COLOR_GREEN)Security scan complete!$(COLOR_RESET)"

.PHONY: all
all: clean install check build test ## Run all common tasks (clean, install, check, build, test)

# Default target
.DEFAULT_GOAL := help
