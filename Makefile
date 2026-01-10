# Makefile for Xiaohongshu Image Generation System

.PHONY: help build run-api run-worker test clean docker-up docker-down docker-logs

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all binaries
	@echo "Building API..."
	@go build -o bin/api ./cmd/api
	@echo "Building Worker..."
	@go build -o bin/worker ./cmd/worker
	@echo "Build complete!"

run-api: ## Run API server
	@go run ./cmd/api/main.go

run-worker: ## Run worker
	@go run ./cmd/worker/main.go

test: ## Run all tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	@rm -rf bin/
	@rm -f coverage.out coverage.html

docker-up: ## Start all services with Docker Compose
	@docker-compose up -d

docker-down: ## Stop all services
	@docker-compose down

docker-logs: ## Show logs from all services
	@docker-compose logs -f

docker-rebuild: ## Rebuild and restart services
	@docker-compose down
	@docker-compose build
	@docker-compose up -d

deps: ## Download dependencies
	@go mod download
	@go mod tidy

web-dev: ## Run frontend in development mode
	@cd web && npm run dev

web-build: ## Build frontend for production
	@cd web && npm run build
