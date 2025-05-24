# Makefile for Strava Data Importer

.PHONY: help build test test-verbose test-coverage clean run docker-up docker-down docker-build lint format deps install-tools

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=strava-data-importer
BINARY_PATH=bin/$(BINARY_NAME)
MAIN_PATH=./cmd/main.go

# Docker parameters
DOCKER_COMPOSE_FILE=docker/docker-compose.yml
DOCKER_COMPOSE=docker-compose -f $(DOCKER_COMPOSE_FILE)

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
