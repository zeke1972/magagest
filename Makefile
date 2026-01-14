# Makefile

.PHONY: help build run test clean docker-build docker-up docker-down docker-logs install-deps

# Variables
APP_NAME=ricambi-manager
DOCKER_COMPOSE=docker-compose
GO=go
GOFLAGS=-v

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-deps: ## Install Go dependencies
	$(GO) mod download
	$(GO) mod tidy

build: ## Build the application
	$(GO) build $(GOFLAGS) -o bin/$(APP_NAME) ./cmd/server

run: ## Run the application locally
	$(GO) run ./cmd/server/main.go

test: ## Run tests
	$(GO) test -v -cover ./...

test-coverage: ## Run tests with coverage report
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf logs/*.log
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	$(DOCKER_COMPOSE) build

docker-up: ## Start Docker containers
	$(DOCKER_COMPOSE) up -d

docker-down: ## Stop Docker containers
	$(DOCKER_COMPOSE) down

docker-logs: ## Show Docker logs
	$(DOCKER_COMPOSE) logs -f

docker-restart: ## Restart Docker containers
	$(DOCKER_COMPOSE) restart

docker-clean: ## Remove Docker containers and volumes
	$(DOCKER_COMPOSE) down -v

mongodb-shell: ## Connect to MongoDB shell
	docker exec -it ricambi-mongodb mongosh -u admin -p password ricambi_db

lint: ## Run linter
	golangci-lint run ./...

fmt: ## Format code
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

mod-update: ## Update Go modules
	$(GO) get -u ./...
	$(GO) mod tidy

seed-data: ## Seed database with sample data
	$(GO) run ./cmd/seed/main.go

backup-db: ## Backup MongoDB database
	docker exec ricambi-mongodb mongodump --uri="mongodb://admin:password@localhost:27017/ricambi_db?authSource=admin" --out=/dump
	docker cp ricambi-mongodb:/dump ./backups/$(shell date +%Y%m%d_%H%M%S)

restore-db: ## Restore MongoDB database (specify BACKUP_DIR=path)
	@test -n "$(BACKUP_DIR)" || (echo "Please specify BACKUP_DIR=path"; exit 1)
	docker cp $(BACKUP_DIR) ricambi-mongodb:/dump
	docker exec ricambi-mongodb mongorestore --uri="mongodb://admin:password@localhost:27017/ricambi_db?authSource=admin" /dump

all: clean install-deps build test ## Build and test everything
