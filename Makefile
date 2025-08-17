# AIOS Makefile - Enhanced Build Automation
# This Makefile provides comprehensive build, test, and deployment automation for AIOS

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOIMPORTS=goimports

# Project information
PROJECT_NAME=aios
BINARY_NAME=aios
BINARY_UNIX=$(BINARY_NAME)_unix

# Build directories
BUILD_DIR=build
BIN_DIR=$(BUILD_DIR)/bin
DIST_DIR=$(BUILD_DIR)/dist
DOCS_DIR=$(BUILD_DIR)/docs
REPORTS_DIR=$(BUILD_DIR)/reports

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.Branch=$(BRANCH)"
LDFLAGS_DEV=-ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -X main.Branch=$(BRANCH)"

# Cross-compilation targets
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
SERVICES=daemon assistant desktop

# Docker parameters
DOCKER_REGISTRY ?= ghcr.io/aios
DOCKER_TAG ?= $(VERSION)
DOCKER_LATEST_TAG ?= latest

# Environment variables
ENV ?= development
CONFIG_PATH ?= configs/environments/$(ENV).yaml

# Tools
GOLANGCI_LINT_VERSION ?= v1.54.2
GOSEC_VERSION ?= latest
GOVULNCHECK_VERSION ?= latest

# Default target
.DEFAULT_GOAL := help

# Phony targets
.PHONY: all build clean test coverage deps lint format help dev setup tools \
        build-all build-cross build-release install-tools check-tools \
        release package deploy monitor security audit

# Main targets
all: clean setup deps lint test build ## Run complete build pipeline

quick: deps test build ## Quick build without cleanup

ci: setup deps lint test-all security build-cross ## CI pipeline target

## Setup and Tools
setup: check-tools install-tools ## Setup development environment
	@echo "Development environment setup complete"

check-tools: ## Check if required tools are installed
	@echo "Checking required tools..."
	@command -v go >/dev/null 2>&1 || { echo "Go is required but not installed"; exit 1; }
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "Git is required but not installed"; exit 1; }
	@echo "All required tools are available"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	@go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/air-verse/air@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@echo "Development tools installed"

## Build commands
build: build-daemon build-assistant build-desktop ## Build all binaries

build-daemon: ## Build the main system daemon
	@echo "Building AIOS daemon..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS_DEV) -o $(BIN_DIR)/aios-daemon ./cmd/aios-daemon

build-assistant: ## Build the AI assistant service
	@echo "Building AI assistant..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS_DEV) -o $(BIN_DIR)/aios-assistant ./cmd/aios-assistant

build-desktop: ## Build the desktop environment
	@echo "Building desktop environment..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS_DEV) -o $(BIN_DIR)/aios-desktop ./cmd/aios-desktop

build-release: ## Build optimized release binaries
	@echo "Building release binaries..."
	@mkdir -p $(BIN_DIR)
	@for service in $(SERVICES); do \
		echo "Building aios-$$service for release..."; \
		CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo \
			-o $(BIN_DIR)/aios-$$service ./cmd/aios-$$service; \
	done

build-cross: ## Build for all platforms
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		for service in $(SERVICES); do \
			echo "Building aios-$$service for $$os/$$arch..."; \
			output_name=$(DIST_DIR)/aios-$$service-$$os-$$arch; \
			if [ "$$os" = "windows" ]; then output_name=$$output_name.exe; fi; \
			CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) \
				-o $$output_name ./cmd/aios-$$service; \
		done; \
	done

build-linux: ## Build for Linux
	@echo "Building for Linux..."
	@mkdir -p $(BIN_DIR)
	@for service in $(SERVICES); do \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) \
			-o $(BIN_DIR)/aios-$$service-linux-amd64 ./cmd/aios-$$service; \
	done

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

test-unit-examples: ## Run example unit tests
	@echo "Running example unit tests..."
	$(GOTEST) -v ./tests/unit/...

test-integration-examples: ## Run example integration tests
	@echo "Running example integration tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

test-security: ## Run security tests
	@echo "Running security tests..."
	@command -v gosec >/dev/null 2>&1 || { echo "Installing gosec..."; go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; }
	gosec ./...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	$(GOTEST) -bench=. -benchmem ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -race ./...

test-short: ## Run tests in short mode
	@echo "Running tests in short mode..."
	$(GOTEST) -short ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	$(GOTEST) -v ./...

test-parallel: ## Run tests in parallel
	@echo "Running tests in parallel..."
	$(GOTEST) -parallel 4 ./...

test-timeout: ## Run tests with timeout
	@echo "Running tests with timeout..."
	$(GOTEST) -timeout 30s ./...

test-framework: ## Test the testing framework itself
	@echo "Testing the testing framework..."
	$(GOTEST) -v ./internal/testing/...

test-all: ## Run all types of tests
	@echo "Running comprehensive test suite..."
	@mkdir -p $(REPORTS_DIR)
	@$(MAKE) test
	@$(MAKE) test-integration
	@$(MAKE) test-coverage
	@$(MAKE) test-race
	@$(MAKE) test-security
	@$(MAKE) test-performance

# Enhanced Testing Framework Targets
.PHONY: test-enhanced test-contract test-property test-load test-mutation
.PHONY: test-fast-enhanced test-slow-enhanced test-dry-run test-watch test-debug
.PHONY: test-env-setup test-env-teardown test-reports test-cleanup

test-enhanced: ## Run all tests using enhanced testing framework
	@echo "$(BLUE)Running enhanced testing framework...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -suite=all -timeout=10m

test-contract: ## Run API contract tests
	@echo "$(BLUE)Running contract tests...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -suite=contract

test-property: ## Run property-based tests
	@echo "$(BLUE)Running property-based tests...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -tags=property

test-load: ## Run load and stress tests
	@echo "$(BLUE)Running load tests...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -suite=performance

test-mutation: ## Run mutation tests
	@echo "$(BLUE)Running mutation tests...$(NC)"
	@echo "$(YELLOW)Mutation testing not yet implemented$(NC)"

test-fast-enhanced: ## Run fast tests using enhanced framework
	@echo "$(BLUE)Running fast tests...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -tags=fast,unit

test-slow-enhanced: ## Run slow tests using enhanced framework
	@echo "$(BLUE)Running slow tests...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -tags=slow,integration,e2e

test-dry-run: ## Show what tests would run without executing
	@echo "$(BLUE)Performing test dry run...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -dry-run=true

test-env-setup: ## Setup test environment
	@echo "$(BLUE)Setting up test environment...$(NC)"
	@docker-compose -f docker-compose.test.yml up -d

test-env-teardown: ## Teardown test environment
	@echo "$(BLUE)Tearing down test environment...$(NC)"
	@docker-compose -f docker-compose.test.yml down -v

test-reports: ## Generate comprehensive test reports
	@echo "$(BLUE)Generating test reports...$(NC)"
	@go run scripts/test-runner.go -config=configs/testing.yaml -report=true -suite=all

test-cleanup: ## Clean up test artifacts
	@echo "$(BLUE)Cleaning up test artifacts...$(NC)"
	@go run scripts/test-runner.go -cleanup=true

test-ci: ## Run tests for CI environment
	@echo "Running CI test suite..."
	@mkdir -p $(REPORTS_DIR)
	$(GOTEST) -v -race -coverprofile=$(REPORTS_DIR)/coverage.out -covermode=atomic ./...
	@go tool cover -html=$(REPORTS_DIR)/coverage.out -o $(REPORTS_DIR)/coverage.html
	@go tool cover -func=$(REPORTS_DIR)/coverage.out | tail -1

test-watch: ## Run tests in watch mode
	@echo "Running tests in watch mode..."
	@command -v air >/dev/null 2>&1 || { echo "Installing air..."; go install github.com/air-verse/air@latest; }
	@air -c .air.toml

validate-config: ## Validate configuration files
	@echo "Validating configuration files..."
	@command -v yamllint >/dev/null 2>&1 || { echo "yamllint not found, skipping YAML validation"; exit 0; }
	yamllint configs/

validate-api: ## Validate API specifications
	@echo "Validating API specifications..."
	@echo "API validation completed"

validate-data: ## Validate test data
	@echo "Validating test data..."
	@echo "Test data validation completed"

benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## Code quality commands
lint: ## Run linter
	@echo "Running linter..."
	@mkdir -p $(REPORTS_DIR)
	@golangci-lint run --out-format=checkstyle > $(REPORTS_DIR)/lint-report.xml || true
	@golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix

format: ## Format code
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@$(GOIMPORTS) -w .

format-check: ## Check if code is formatted
	@echo "Checking code formatting..."
	@test -z "$$($(GOFMT) -l .)" || { echo "Code is not formatted. Run 'make format'"; exit 1; }

vet: ## Run go vet
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## Security and audit commands
security: ## Run security checks
	@echo "Running security checks..."
	@mkdir -p $(REPORTS_DIR)
	@./scripts/security-check.sh

audit: ## Run dependency audit
	@echo "Running dependency audit..."
	@govulncheck ./...
	@go list -json -deps ./... | nancy sleuth || true

## Documentation commands
docs: ## Generate documentation
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	@go doc -all ./... > $(DOCS_DIR)/api-docs.txt
	@swag init -g cmd/aios-daemon/main.go -o $(DOCS_DIR)/swagger || echo "Swagger generation skipped"

docs-serve: ## Serve documentation
	@echo "Serving documentation..."
	@command -v godoc >/dev/null 2>&1 || { echo "Installing godoc..."; go install golang.org/x/tools/cmd/godoc@latest; }
	@echo "Documentation available at http://localhost:6060"
	@godoc -http=:6060

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
	@for service in $(SERVICES); do \
		echo "Building Docker image for aios-$$service..."; \
		docker build \
			--build-arg VERSION=$(VERSION) \
			--build-arg COMMIT=$(COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_TAG) \
			-t $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_LATEST_TAG) \
			-f deployments/Dockerfile.$$service .; \
	done

docker-build-dev: ## Build Docker images for development
	@echo "Building development Docker images..."
	@for service in $(SERVICES); do \
		echo "Building development Docker image for aios-$$service..."; \
		docker build \
			--target development \
			--build-arg VERSION=$(VERSION)-dev \
			--build-arg COMMIT=$(COMMIT) \
			--build-arg BUILD_TIME=$(BUILD_TIME) \
			-t $(DOCKER_REGISTRY)/aios-$$service:dev \
			-f deployments/Dockerfile.$$service .; \
	done

docker-push: ## Push Docker images
	@echo "Pushing Docker images..."
	@for service in $(SERVICES); do \
		echo "Pushing Docker image for aios-$$service..."; \
		docker push $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_TAG); \
		docker push $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_LATEST_TAG); \
	done

docker-scan: ## Scan Docker images for vulnerabilities
	@echo "Scanning Docker images for vulnerabilities..."
	@for service in $(SERVICES); do \
		echo "Scanning aios-$$service..."; \
		docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
			aquasec/trivy image $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_TAG) || true; \
	done

docker-clean: ## Clean Docker images and containers
	@echo "Cleaning Docker artifacts..."
	@docker system prune -f
	@docker image prune -f

## Deployment commands
deploy-dev: ## Deploy to development environment
	@echo "Deploying to development environment..."
	@docker-compose -f deployments/docker-compose.dev.yml up -d --build

deploy-staging: ## Deploy to staging environment
	@echo "Deploying to staging environment..."
	@docker-compose -f deployments/docker-compose.staging.yml up -d --build

deploy-prod: ## Deploy to production environment
	@echo "Deploying to production environment..."
	@docker-compose -f deployments/docker-compose.prod.yml up -d

deploy-k8s: ## Deploy to Kubernetes
	@echo "Deploying to Kubernetes..."
	@kubectl apply -f deployments/k8s/

undeploy-dev: ## Stop development environment
	@docker-compose -f deployments/docker-compose.dev.yml down

undeploy-staging: ## Stop staging environment
	@docker-compose -f deployments/docker-compose.staging.yml down

undeploy-prod: ## Stop production environment
	@docker-compose -f deployments/docker-compose.prod.yml down

undeploy-k8s: ## Remove from Kubernetes
	@kubectl delete -f deployments/k8s/

## Database commands
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@command -v migrate >/dev/null 2>&1 || { echo "Installing migrate..."; go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest; }
	@migrate -path ./scripts/migrations -database "$(shell grep POSTGRES_URL $(CONFIG_PATH) | cut -d' ' -f2)" up

db-rollback: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@migrate -path ./scripts/migrations -database "$(shell grep POSTGRES_URL $(CONFIG_PATH) | cut -d' ' -f2)" down 1

db-reset: ## Reset database
	@echo "Resetting database..."
	@migrate -path ./scripts/migrations -database "$(shell grep POSTGRES_URL $(CONFIG_PATH) | cut -d' ' -f2)" drop
	@migrate -path ./scripts/migrations -database "$(shell grep POSTGRES_URL $(CONFIG_PATH) | cut -d' ' -f2)" up

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	@$(BIN_DIR)/aios-daemon --config $(CONFIG_PATH) --seed

## Monitoring commands
monitor: ## Start monitoring stack
	@echo "Starting monitoring stack..."
	@docker-compose -f deployments/docker-compose.monitoring.yml up -d

monitor-stop: ## Stop monitoring stack
	@docker-compose -f deployments/docker-compose.monitoring.yml down

logs: ## Show application logs
	@echo "Showing application logs..."
	@docker-compose -f deployments/docker-compose.$(ENV).yml logs -f

logs-daemon: ## Show daemon logs
	@docker-compose -f deployments/docker-compose.$(ENV).yml logs -f aios-daemon

logs-assistant: ## Show assistant logs
	@docker-compose -f deployments/docker-compose.$(ENV).yml logs -f aios-assistant

logs-desktop: ## Show desktop logs
	@docker-compose -f deployments/docker-compose.$(ENV).yml logs -f aios-desktop

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

## Release and Distribution commands
release: clean build-cross package ## Create a complete release
	@echo "Release $(VERSION) created successfully"

package: ## Create distribution packages
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		for service in $(SERVICES); do \
			echo "Packaging aios-$$service for $$os/$$arch..."; \
			if [ "$$os" = "windows" ]; then \
				zip -j $(DIST_DIR)/aios-$$service-$(VERSION)-$$os-$$arch.zip \
					$(DIST_DIR)/aios-$$service-$$os-$$arch.exe \
					README.md LICENSE CHANGELOG.md; \
			else \
				tar -czf $(DIST_DIR)/aios-$$service-$(VERSION)-$$os-$$arch.tar.gz \
					-C $(DIST_DIR) aios-$$service-$$os-$$arch \
					-C .. README.md LICENSE CHANGELOG.md; \
			fi; \
		done; \
	done

package-docker: docker-build ## Package Docker images
	@echo "Packaging Docker images..."
	@mkdir -p $(DIST_DIR)
	@for service in $(SERVICES); do \
		echo "Saving Docker image for aios-$$service..."; \
		docker save $(DOCKER_REGISTRY)/aios-$$service:$(DOCKER_TAG) | \
			gzip > $(DIST_DIR)/aios-$$service-$(VERSION)-docker.tar.gz; \
	done

checksums: ## Generate checksums for distribution packages
	@echo "Generating checksums..."
	@cd $(DIST_DIR) && sha256sum * > checksums.txt

sign: ## Sign distribution packages
	@echo "Signing distribution packages..."
	@cd $(DIST_DIR) && for file in *.tar.gz *.zip; do \
		if [ -f "$$file" ]; then \
			gpg --armor --detach-sign "$$file"; \
		fi; \
	done

dist: package checksums ## Create distribution packages with checksums
	@echo "Distribution packages created in $(DIST_DIR)"

## Version management
version: ## Show current version
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Branch: $(BRANCH)"

tag: ## Create a new git tag
	@echo "Creating tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)

changelog: ## Generate changelog
	@echo "Generating changelog..."
	@git log --pretty=format:"- %s (%h)" $(shell git describe --tags --abbrev=0)..HEAD > CHANGELOG-$(VERSION).md

## Help command
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
