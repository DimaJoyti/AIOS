# AIOS Makefile

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=aios
BINARY_UNIX=$(BINARY_NAME)_unix

# Build directories
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin
DIST_DIR=$(BUILD_DIR)/dist

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT ?= $(shell git rev-parse HEAD)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Docker parameters
DOCKER_REGISTRY ?= localhost:5000
DOCKER_TAG ?= latest

.PHONY: all build clean test coverage deps lint format help dev

all: clean deps test build

## Build commands
build: build-daemon build-assistant build-desktop ## Build all binaries

build-daemon: ## Build the main system daemon
	@echo "Building AIOS daemon..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/aios-daemon ./cmd/aios-daemon

build-assistant: ## Build the AI assistant service
	@echo "Building AI assistant..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/aios-assistant ./cmd/aios-assistant

build-desktop: ## Build the desktop environment
	@echo "Building desktop environment..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/aios-desktop ./cmd/aios-desktop

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_UNIX) ./cmd/aios-daemon

## Development commands
dev: ## Run in development mode
	@echo "Starting AIOS in development mode..."
	@docker-compose -f deployments/docker-compose.dev.yml up --build

dev-stop: ## Stop development environment
	@docker-compose -f deployments/docker-compose.dev.yml down

dev-logs: ## Show development logs
	@docker-compose -f deployments/docker-compose.dev.yml logs -f

## Testing commands
test: ## Run tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

test-e2e: ## Run end-to-end tests
	@echo "Running end-to-end tests..."
	$(GOTEST) -v -tags=e2e ./tests/e2e/...

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Code quality commands
lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run

format: ## Format code
	@echo "Formatting code..."
	@gofmt -s -w .
	@goimports -w .

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## Dependency commands
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy

## Docker commands
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	@docker build -t $(DOCKER_REGISTRY)/aios-daemon:$(DOCKER_TAG) -f deployments/Dockerfile.daemon .
	@docker build -t $(DOCKER_REGISTRY)/aios-assistant:$(DOCKER_TAG) -f deployments/Dockerfile.assistant .
	@docker build -t $(DOCKER_REGISTRY)/aios-desktop:$(DOCKER_TAG) -f deployments/Dockerfile.desktop .

docker-push: ## Push Docker images
	@echo "Pushing Docker images..."
	@docker push $(DOCKER_REGISTRY)/aios-daemon:$(DOCKER_TAG)
	@docker push $(DOCKER_REGISTRY)/aios-assistant:$(DOCKER_TAG)
	@docker push $(DOCKER_REGISTRY)/aios-desktop:$(DOCKER_TAG)

## Database commands
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@migrate -path ./migrations -database "postgres://localhost/aios?sslmode=disable" up

db-rollback: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@migrate -path ./migrations -database "postgres://localhost/aios?sslmode=disable" down 1

db-reset: ## Reset database
	@echo "Resetting database..."
	@migrate -path ./migrations -database "postgres://localhost/aios?sslmode=disable" drop
	@migrate -path ./migrations -database "postgres://localhost/aios?sslmode=disable" up

## Frontend commands
web-install: ## Install frontend dependencies
	@echo "Installing frontend dependencies..."
	@cd web && npm install

web-build: ## Build frontend applications
	@echo "Building frontend applications..."
	@cd web && npm run build

web-dev: ## Run frontend in development mode
	@echo "Starting frontend development server..."
	@cd web && npm run dev

## Cleanup commands
clean: ## Clean build artifacts
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

clean-docker: ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	@docker system prune -f

## Installation commands
install: build ## Install binaries to system
	@echo "Installing AIOS..."
	@sudo cp $(BIN_DIR)/aios-daemon /usr/local/bin/
	@sudo cp $(BIN_DIR)/aios-assistant /usr/local/bin/
	@sudo cp $(BIN_DIR)/aios-desktop /usr/local/bin/
	@sudo chmod +x /usr/local/bin/aios-*

uninstall: ## Uninstall binaries from system
	@echo "Uninstalling AIOS..."
	@sudo rm -f /usr/local/bin/aios-daemon
	@sudo rm -f /usr/local/bin/aios-assistant
	@sudo rm -f /usr/local/bin/aios-desktop

## Distribution commands
dist: build ## Create distribution packages
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	@tar -czf $(DIST_DIR)/aios-$(VERSION)-linux-amd64.tar.gz -C $(BIN_DIR) .

## Help command
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
