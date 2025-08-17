#!/bin/bash

# AIOS Development Environment Post-Create Script
# This script runs after the dev container is created

set -e

echo "🚀 Setting up AIOS development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Set up Go environment
print_status "Setting up Go environment..."
go version
go env GOPATH
go env GOROOT

# Download Go dependencies
print_status "Downloading Go dependencies..."
if [ -f "go.mod" ]; then
    go mod download
    go mod tidy
    print_success "Go dependencies downloaded"
else
    print_warning "No go.mod found, skipping Go dependency download"
fi

# Install additional Go tools
print_status "Installing additional Go tools..."
go install github.com/air-verse/air@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
print_success "Additional Go tools installed"

# Set up Node.js environment
print_status "Setting up Node.js environment..."
node --version
npm --version

# Install frontend dependencies
if [ -d "web" ] && [ -f "web/package.json" ]; then
    print_status "Installing frontend dependencies..."
    cd web
    npm ci
    cd ..
    print_success "Frontend dependencies installed"
else
    print_warning "No web/package.json found, skipping frontend dependency installation"
fi

# Set up Git hooks
print_status "Setting up Git hooks..."
if [ -d ".git" ]; then
    # Create pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# AIOS pre-commit hook

echo "Running pre-commit checks..."

# Run Go formatting
if ! gofmt -l . | grep -q '^$'; then
    echo "Code is not formatted. Please run 'make format'"
    exit 1
fi

# Run Go linting
if ! golangci-lint run --timeout=5m; then
    echo "Linting failed. Please fix the issues."
    exit 1
fi

# Run Go tests
if ! go test -short ./...; then
    echo "Tests failed. Please fix the failing tests."
    exit 1
fi

echo "Pre-commit checks passed!"
EOF

    chmod +x .git/hooks/pre-commit
    print_success "Git hooks configured"
else
    print_warning "Not a Git repository, skipping Git hooks setup"
fi

# Create development configuration
print_status "Creating development configuration..."
if [ ! -f "configs/dev.yaml" ]; then
    mkdir -p configs
    cat > configs/dev.yaml << 'EOF'
# AIOS Development Configuration
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

metrics:
  host: "0.0.0.0"
  port: 9090
  path: "/metrics"

logging:
  level: "debug"
  format: "json"
  output: "stdout"

database:
  url: "postgres://aios:aios@postgres:5432/aios?sslmode=disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 300s

redis:
  url: "redis://redis:6379"
  pool_size: 10
  min_idle_conns: 5

ai:
  enabled: true
  models_path: "/tmp/aios/models"
  default_model: "llama2"
  gpu_enabled: false

security:
  jwt_secret: "dev-secret-key-change-in-production"
  encryption_key: "dev-encryption-key-32-chars-long"
  cors_origins: ["http://localhost:3000", "http://localhost:8080"]

features:
  voice_control: false
  computer_vision: true
  predictive_fs: true
  developer_tools: true
EOF
    print_success "Development configuration created"
fi

# Create database initialization script
print_status "Creating database initialization script..."
if [ ! -f "scripts/init-db.sql" ]; then
    mkdir -p scripts
    cat > scripts/init-db.sql << 'EOF'
-- AIOS Database Initialization Script

-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create schemas
CREATE SCHEMA IF NOT EXISTS aios;
CREATE SCHEMA IF NOT EXISTS ai_models;
CREATE SCHEMA IF NOT EXISTS system_metrics;

-- Create basic tables
CREATE TABLE IF NOT EXISTS aios.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS aios.sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES aios.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS system_metrics.performance_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cpu_usage DECIMAL(5,2),
    memory_usage DECIMAL(5,2),
    disk_usage DECIMAL(5,2),
    network_io JSONB,
    metadata JSONB
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON aios.users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON aios.users(email);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON aios.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON aios.sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_performance_logs_timestamp ON system_metrics.performance_logs(timestamp);

-- Insert default admin user (password: admin123)
INSERT INTO aios.users (username, email, password_hash) 
VALUES ('admin', 'admin@aios.dev', '$2a$10$rQZ8ZqNQzqNQzqNQzqNQzOeKqNQzqNQzqNQzqNQzqNQzqNQzqNQzq')
ON CONFLICT (username) DO NOTHING;
EOF
    print_success "Database initialization script created"
fi

# Create useful development scripts
print_status "Creating development scripts..."
mkdir -p scripts

# Create development server script
cat > scripts/dev-server.sh << 'EOF'
#!/bin/bash
# Start AIOS development servers

echo "Starting AIOS development environment..."

# Start backend services
echo "Starting backend services..."
go run cmd/aios-daemon/main.go --config configs/dev.yaml &
DAEMON_PID=$!

go run cmd/aios-assistant/main.go --config configs/dev.yaml &
ASSISTANT_PID=$!

go run cmd/aios-desktop/main.go --config configs/dev.yaml &
DESKTOP_PID=$!

# Start frontend development server
echo "Starting frontend development server..."
cd web && npm run dev &
FRONTEND_PID=$!

# Function to cleanup on exit
cleanup() {
    echo "Stopping development servers..."
    kill $DAEMON_PID $ASSISTANT_PID $DESKTOP_PID $FRONTEND_PID 2>/dev/null
    exit 0
}

trap cleanup SIGINT SIGTERM

echo "Development servers started!"
echo "API: http://localhost:8080"
echo "Metrics: http://localhost:9090"
echo "Frontend: http://localhost:3000"
echo "Press Ctrl+C to stop all servers"

wait
EOF

chmod +x scripts/dev-server.sh

# Create test script
cat > scripts/run-tests.sh << 'EOF'
#!/bin/bash
# Run all AIOS tests

echo "Running AIOS test suite..."

# Run Go tests
echo "Running Go tests..."
go test -v -race -coverprofile=coverage.out ./...

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html

# Run frontend tests
if [ -d "web" ]; then
    echo "Running frontend tests..."
    cd web && npm test
    cd ..
fi

echo "Test suite completed!"
echo "Coverage report: coverage.html"
EOF

chmod +x scripts/run-tests.sh

print_success "Development scripts created"

# Set up workspace settings
print_status "Setting up workspace settings..."
mkdir -p .vscode

cat > .vscode/settings.json << 'EOF'
{
    "go.toolsManagement.checkForUpdates": "local",
    "go.useLanguageServer": true,
    "go.lintTool": "golangci-lint",
    "go.lintFlags": ["--fast"],
    "go.formatTool": "goimports",
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverOnSingleTest": true,
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "files.eol": "\n",
    "files.insertFinalNewline": true,
    "files.trimTrailingWhitespace": true,
    "eslint.workingDirectories": ["web"],
    "typescript.preferences.importModuleSpecifier": "relative"
}
EOF

cat > .vscode/launch.json << 'EOF'
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch AIOS Daemon",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/aios-daemon",
            "args": ["--config", "configs/dev.yaml"],
            "env": {
                "AIOS_ENV": "development"
            }
        },
        {
            "name": "Launch AIOS Assistant",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/aios-assistant",
            "args": ["--config", "configs/dev.yaml"],
            "env": {
                "AIOS_ENV": "development"
            }
        },
        {
            "name": "Launch AIOS Desktop",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/cmd/aios-desktop",
            "args": ["--config", "configs/dev.yaml"],
            "env": {
                "AIOS_ENV": "development"
            }
        }
    ]
}
EOF

cat > .vscode/tasks.json << 'EOF'
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Build All",
            "type": "shell",
            "command": "make",
            "args": ["build"],
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Test All",
            "type": "shell",
            "command": "make",
            "args": ["test"],
            "group": "test",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        },
        {
            "label": "Run Development Server",
            "type": "shell",
            "command": "./scripts/dev-server.sh",
            "group": "build",
            "presentation": {
                "echo": true,
                "reveal": "always",
                "focus": false,
                "panel": "shared"
            }
        }
    ]
}
EOF

print_success "VS Code workspace settings configured"

# Final setup
print_status "Finalizing setup..."

# Create .env file for development
cat > .env.development << 'EOF'
# AIOS Development Environment Variables
AIOS_ENV=development
AIOS_LOG_LEVEL=debug
AIOS_CONFIG_PATH=configs/dev.yaml
POSTGRES_URL=postgres://aios:aios@postgres:5432/aios?sslmode=disable
REDIS_URL=redis://redis:6379
EOF

print_success "Development environment setup completed!"

echo ""
echo "🎉 AIOS development environment is ready!"
echo ""
echo "Quick start commands:"
echo "  make dev          - Start development environment with Docker Compose"
echo "  make build        - Build all services"
echo "  make test         - Run all tests"
echo "  make lint         - Run code linting"
echo ""
echo "Development servers:"
echo "  ./scripts/dev-server.sh  - Start all development servers"
echo "  ./scripts/run-tests.sh   - Run comprehensive test suite"
echo ""
echo "Happy coding! 🚀"
