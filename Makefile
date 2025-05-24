# Makefile for Strava Data Importer

.PHONY: help build test test-verbose test-coverage test-integration benchmark clean run run-dev setup-env init-project docker-up docker-down docker-build docker-logs lint format security deps install-tools build-all

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint
BINARY_NAME=strava-data-importer
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/main.go

# Build parameters
BUILD_FLAGS=-ldflags="-s -w"

# Docker parameters
DOCKER_COMPOSE_FILE=docker/docker-compose.yml
DOCKER_COMPOSE=docker compose -f $(DOCKER_COMPOSE_FILE) --env-file .env

# Colors for output
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Strava Data Importer Makefile$(NC)"
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	$(GOMOD) download
	$(GOMOD) tidy

build: deps ## Build the application
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) -a -installsuffix cgo -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "$(GREEN)Build completed: $(BINARY_PATH)$(NC)"

test: ## Run tests
	@echo "$(GREEN)Running tests...$(NC)"
	$(GOTEST) -v ./...

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

run-dev: ## Run the application in development mode
	@echo "$(GREEN)Running in development mode...$(NC)"
	$(GOCMD) run $(MAIN_PATH)

run: build ## Run the built application
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	./$(BINARY_PATH)

setup-env: ## Setup environment files
	@echo "$(GREEN)Setting up environment...$(NC)"
	@if [ ! -f .env ]; then \
		cp .env.example .env 2>/dev/null || echo "# Environment variables\nPORT=8080\nLOG_LEVEL=info\nSTRAVA_CLIENT_ID=\nSTRAVA_CLIENT_SECRET=\nSTRAVA_REDIRECT_URI=http://localhost:8080/auth/callback\nINFLUXDB_URL=http://localhost:8086\nINFLUXDB_TOKEN=\nINFLUXDB_ORG=strava\nINFLUXDB_BUCKET=activities\nTOKEN_REFRESH_INTERVAL=24h\nDATA_IMPORT_INTERVAL=1h" > .env; \
		echo "$(GREEN).env file created. Please edit it with your settings.$(NC)"; \
	else \
		echo "$(YELLOW).env file already exists.$(NC)"; \
	fi

init-project: deps install-tools setup-env ## Initialize the project
	@echo "$(GREEN)Initializing project...$(NC)"
	@mkdir -p bin logs conf
	@if [ ! -f conf/ftp.csv ]; then \
		printf "date,ftp\n2024-01-01,170\n2024-08-29,191\n2024-10-27,217\n2025-02-05,248\n" > conf/ftp.csv; \
		echo "$(GREEN)Created example FTP configuration file.$(NC)"; \
	else \
		echo "$(GREEN)FTP configuration file already exists.$(NC)"; \
	fi

install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@echo "$(GREEN)Development tools installed.$(NC)"

docker-up: ## Start Docker services
	@echo "$(GREEN)Starting Docker services...$(NC)"
	@if [ ! -f $(DOCKER_COMPOSE_FILE) ]; then \
		echo "$(RED)Error: Docker Compose file not found at $(DOCKER_COMPOSE_FILE)$(NC)"; \
		exit 1; \
	fi
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)Error: Docker is not installed or not available$(NC)"; exit 1; }
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker services
	@echo "$(GREEN)Stopping Docker services...$(NC)"
	@if [ ! -f $(DOCKER_COMPOSE_FILE) ]; then \
		echo "$(RED)Error: Docker Compose file not found at $(DOCKER_COMPOSE_FILE)$(NC)"; \
		exit 1; \
	fi
	@command -v docker >/dev/null 2>&1 || { echo "$(RED)Error: Docker is not installed or not available$(NC)"; exit 1; }
	$(DOCKER_COMPOSE) down

docker-build: ## Build Docker image
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -f docker/Dockerfile -t $(BINARY_NAME):latest .

docker-logs: ## Show Docker logs
	@echo "$(GREEN)Showing Docker logs...$(NC)"
	$(DOCKER_COMPOSE) logs -f

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

test-integration: docker-up ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	@sleep 5  # Wait for services to start
	$(GOTEST) -v -tags=integration ./...

benchmark: ## Run benchmark tests
	@echo "$(GREEN)Running benchmarks...$(NC)"
	$(GOTEST) -bench=. -benchmem ./...

lint: ## Run linter
	@echo "$(GREEN)Running linter...$(NC)"
	$(GOLINT) run ./...

format: ## Format code
	@echo "$(GREEN)Formatting code...$(NC)"
	$(GOFMT) -s -w .
	go mod tidy

security: ## Run security checks
	@echo "$(GREEN)Running security checks...$(NC)"
	$(GOLINT) run --enable=gosec,gas ./...

build-all: ## Build for all platforms
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p bin
	@for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64; do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name="bin/$(BINARY_NAME)-$$GOOS-$$GOARCH"; \
		if [ "$$GOOS" = "windows" ]; then output_name="$$output_name.exe"; fi; \
		echo "Building $$output_name..."; \
		env GOOS=$$GOOS GOARCH=$$GOARCH CGO_ENABLED=0 $(GOBUILD) $(BUILD_FLAGS) -o $$output_name $(MAIN_PATH); \
	done
	@echo "$(GREEN)Multi-platform build completed!$(NC)"
