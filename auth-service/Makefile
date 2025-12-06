.PHONY: help build run test clean docker-build docker-up docker-down migrate-up migrate-down migrate-create sqlc fmt lint vet tidy

# Variables
APP_NAME=auth-service
DOCKER_IMAGE=$(APP_NAME):latest
DOCKER_COMPOSE=docker compose -f deploy/docker/docker-compose.yaml
GOBIN=$(shell go env GOPATH)/bin
GOOSE=$(GOBIN)/goose
SQLC=$(GOBIN)/sqlc
MIGRATION_DIR=sql/schema
DATABASE_URL ?= postgres://authuser:authpass@localhost:5436/authdb?sslmode=disable

help: ## Display this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the application binary
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) cmd/server/main.go

run: ## Run the application locally
	@echo "Running $(APP_NAME)..."
	@go run cmd/server/main.go

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race -short ./internal/service/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -race -run Integration ./...

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

test-watch: ## Run tests in watch mode (requires entr)
	@echo "Watching for changes..."
	@find . -name '*.go' | entr -c go test -v ./...

test-clean: ## Clean test cache
	@go clean -testcache

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/ dist/ coverage.out coverage.html

# Docker commands
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) -f deploy/docker/Dockerfile .

docker-up: ## Start all services with docker-compose
	@echo "Starting services..."
	@$(DOCKER_COMPOSE) up -d

docker-down: ## Stop all services
	@echo "Stopping services..."
	@$(DOCKER_COMPOSE) down

docker-logs: ## View docker-compose logs
	@$(DOCKER_COMPOSE) logs -f

docker-ps: ## List running containers
	@$(DOCKER_COMPOSE) ps

# Database migration commands
migrate-up: ## Run all up migrations
	@echo "Running migrations up..."
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	@echo "Rolling back migration..."
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" down

migrate-status: ## Show migration status
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" status

migrate-create: ## Create a new migration file (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required. Usage: make migrate-create NAME=your_migration_name"; \
		exit 1; \
	fi
	@echo "Creating migration: $(NAME)"
	@$(GOOSE) -dir $(MIGRATION_DIR) create $(NAME) sql

migrate-reset: ## Reset database (down all, then up all)
	@echo "Resetting database..."
	@$(GOOSE) -dir $(MIGRATION_DIR) postgres "$(DATABASE_URL)" reset

# SQLC commands
sqlc: ## Generate Go code from SQL queries
	@echo "Generating SQLC code..."
	@$(SQLC) generate

sqlc-verify: ## Verify SQLC configuration
	@$(SQLC) verify

# Code quality commands
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

lint: ## Run golangci-lint
	@echo "Running linter..."
	@golangci-lint run ./...

vet: ## Run go vet
	@echo "Running go vet..."
	@go vet ./...

tidy: ## Tidy go modules
	@echo "Tidying modules..."
	@go mod tidy

# Development helpers
dev-setup: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/pressly/goose/v3/cmd/goose@latest
	@go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

dev-init: dev-setup docker-up migrate-up sqlc ## Initialize development environment
	@echo "Development environment ready!"

seed: ## Seed database with sample data
	@echo "Seeding database..."
	@go run scripts/local_dev/seed.go

all: clean tidy fmt vet test build ## Run all checks and build

swagger-gen: ## Generate swagger documentation
	@echo "Generating Swagger documentation..."
	@$(GOBIN)/swag init -g cmd/server/main.go -o docs/swagger --parseDependency --parseInternal

swagger-serve: ## Serve swagger UI locally  
	@echo "Swagger UI available at http://localhost:6969/swagger/index.html"
