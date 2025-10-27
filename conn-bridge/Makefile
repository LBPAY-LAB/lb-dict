.PHONY: help build build-signer test lint run docker-build docker-push clean

# Variables
GO_VERSION := 1.24.5
BINARY_NAME := rsfn-bridge
DOCKER_IMAGE := lbpay-lab/conn-bridge
DOCKER_TAG := latest
XML_SIGNER_JAR := xml-signer.jar

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)

## Help
help: ## Show this help message
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${WHITE}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

## Build Targets
build: ## Build Go service binary
	@echo "${GREEN}Building Go service...${RESET}"
	go mod download
	go build -o bin/$(BINARY_NAME) cmd/server/main.go
	@echo "${GREEN}Build completed: bin/$(BINARY_NAME)${RESET}"

build-signer: ## Build Java XML Signer JAR
	@echo "${GREEN}Building Java XML Signer...${RESET}"
	cd xml-signer && mvn clean package
	cp xml-signer/target/xml-signer-1.0.0-jar-with-dependencies.jar $(XML_SIGNER_JAR)
	@echo "${GREEN}XML Signer JAR created: $(XML_SIGNER_JAR)${RESET}"

## Test Targets
test: ## Run all tests with coverage
	@echo "${GREEN}Running all tests...${RESET}"
	go test -v -race -coverprofile=coverage.out ./...
	@echo "${GREEN}Tests completed${RESET}"

test-unit: ## Run unit tests only
	@echo "${GREEN}Running unit tests...${RESET}"
	go test -v -short ./...
	@echo "${GREEN}Unit tests completed${RESET}"

test-integration: ## Run integration tests
	@echo "${GREEN}Running integration tests...${RESET}"
	go test -v -run Integration ./tests/integration/...
	@echo "${GREEN}Integration tests completed${RESET}"

coverage: ## Generate coverage report
	@echo "${GREEN}Generating coverage report...${RESET}"
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "${GREEN}Coverage report: coverage.html${RESET}"
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

test-coverage: coverage ## Alias for coverage target

## Lint Targets
lint: ## Run linters (golangci-lint)
	@echo "${GREEN}Running linters...${RESET}"
	golangci-lint run --timeout 5m ./...
	@echo "${GREEN}Lint completed${RESET}"

lint-fix: ## Run linters and auto-fix issues
	@echo "${GREEN}Running linters with auto-fix...${RESET}"
	golangci-lint run --fix --timeout 5m ./...
	@echo "${GREEN}Lint fix completed${RESET}"

fmt: ## Format Go code
	@echo "${GREEN}Formatting Go code...${RESET}"
	go fmt ./...
	@echo "${GREEN}Formatting completed${RESET}"

## Run Targets
run: ## Run Go service locally
	@echo "${GREEN}Starting Go service...${RESET}"
	go run cmd/server/main.go

run-signer: ## Run Java XML Signer locally
	@echo "${GREEN}Starting XML Signer service...${RESET}"
	cd xml-signer && mvn spring-boot:run

run-docker: ## Run services with Docker Compose
	@echo "${GREEN}Starting services with Docker Compose...${RESET}"
	docker-compose up -d
	@echo "${GREEN}Services started. Check logs with: docker-compose logs -f${RESET}"

## Docker Targets
docker-build: ## Build Docker images for both services
	@echo "${GREEN}Building Docker images...${RESET}"
	docker build -t $(DOCKER_IMAGE)-go:$(DOCKER_TAG) -f Dockerfile .
	docker build -t $(DOCKER_IMAGE)-signer:$(DOCKER_TAG) -f xml-signer/Dockerfile xml-signer/
	@echo "${GREEN}Docker images built successfully${RESET}"

docker-push: ## Push Docker images to registry
	@echo "${GREEN}Pushing Docker images...${RESET}"
	docker push $(DOCKER_IMAGE)-go:$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE)-signer:$(DOCKER_TAG)
	@echo "${GREEN}Docker images pushed successfully${RESET}"

docker-clean: ## Remove Docker containers and images
	@echo "${GREEN}Cleaning Docker resources...${RESET}"
	docker-compose down -v
	docker rmi $(DOCKER_IMAGE)-go:$(DOCKER_TAG) $(DOCKER_IMAGE)-signer:$(DOCKER_TAG) || true
	@echo "${GREEN}Docker cleanup completed${RESET}"

## Clean Targets
clean: ## Clean build artifacts
	@echo "${GREEN}Cleaning build artifacts...${RESET}"
	rm -rf bin/
	rm -f $(XML_SIGNER_JAR)
	rm -f coverage.out coverage.html
	cd xml-signer && mvn clean
	@echo "${GREEN}Cleanup completed${RESET}"

## Development Targets
dev: ## Run development environment (docker-compose with hot reload)
	@echo "${GREEN}Starting development environment...${RESET}"
	docker-compose -f docker-compose.dev.yml up --build

logs: ## Show logs from running services
	docker-compose logs -f

stop: ## Stop running services
	@echo "${GREEN}Stopping services...${RESET}"
	docker-compose down
	@echo "${GREEN}Services stopped${RESET}"

restart: stop run-docker ## Restart services

## Dependency Targets
deps: ## Download and verify dependencies
	@echo "${GREEN}Downloading dependencies...${RESET}"
	go mod download
	go mod verify
	@echo "${GREEN}Dependencies downloaded${RESET}"

deps-update: ## Update all dependencies
	@echo "${GREEN}Updating dependencies...${RESET}"
	go get -u ./...
	go mod tidy
	@echo "${GREEN}Dependencies updated${RESET}"

## Proto Targets
proto: ## Generate protobuf files
	@echo "${GREEN}Generating protobuf files...${RESET}"
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto
	@echo "${GREEN}Protobuf generation completed${RESET}"

## Misc Targets
check: lint test ## Run linters and tests
	@echo "${GREEN}All checks passed${RESET}"

install-tools: ## Install development tools
	@echo "${GREEN}Installing development tools...${RESET}"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	@echo "${GREEN}Tools installed${RESET}"

health: ## Check health of running services
	@echo "${GREEN}Checking service health...${RESET}"
	@curl -f http://localhost:8082/health || echo "Service is not running"

.DEFAULT_GOAL := help
