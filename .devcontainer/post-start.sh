#!/bin/bash

# AIOS Development Environment Post-Start Script
# This script runs every time the dev container starts

set -e

echo "🔄 Starting AIOS development environment..."

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Wait for services to be ready
print_status "Waiting for services to be ready..."

# Wait for PostgreSQL
print_status "Waiting for PostgreSQL..."
timeout 30 bash -c 'until nc -z postgres 5432; do sleep 1; done' || {
    print_warning "PostgreSQL not ready after 30 seconds"
}

# Wait for Redis
print_status "Waiting for Redis..."
timeout 30 bash -c 'until nc -z redis 6379; do sleep 1; done' || {
    print_warning "Redis not ready after 30 seconds"
}

# Check if services are healthy
if nc -z postgres 5432; then
    print_success "PostgreSQL is ready"
else
    print_warning "PostgreSQL is not accessible"
fi

if nc -z redis 6379; then
    print_success "Redis is ready"
else
    print_warning "Redis is not accessible"
fi

# Update Go dependencies if needed
if [ -f "go.mod" ]; then
    print_status "Checking Go dependencies..."
    go mod download
    go mod tidy
fi

# Update frontend dependencies if needed
if [ -d "web" ] && [ -f "web/package.json" ]; then
    print_status "Checking frontend dependencies..."
    cd web
    if [ ! -d "node_modules" ] || [ "package.json" -nt "node_modules" ]; then
        print_status "Installing/updating frontend dependencies..."
        npm ci
    fi
    cd ..
fi

# Run database migrations if available
if [ -f "scripts/init-db.sql" ] && nc -z postgres 5432; then
    print_status "Running database initialization..."
    PGPASSWORD=aios psql -h postgres -U aios -d aios -f scripts/init-db.sql > /dev/null 2>&1 || {
        print_warning "Database initialization failed or already completed"
    }
fi

# Start background services for development
print_status "Starting background development services..."

# Start file watcher for auto-reload (if air is available)
if command -v air &> /dev/null; then
    print_status "Air (hot reload) is available"
fi

# Display useful information
echo ""
print_success "Development environment is ready!"
echo ""
echo "📊 Service Status:"
echo "  PostgreSQL: $(nc -z postgres 5432 && echo '✅ Ready' || echo '❌ Not Ready')"
echo "  Redis:      $(nc -z redis 6379 && echo '✅ Ready' || echo '❌ Not Ready')"
echo ""
echo "🔧 Development Commands:"
echo "  make dev          - Start all services with Docker Compose"
echo "  make build        - Build all AIOS services"
echo "  make test         - Run test suite"
echo "  make lint         - Run code linting"
echo "  make format       - Format code"
echo ""
echo "🚀 Quick Start:"
echo "  ./scripts/dev-server.sh  - Start development servers"
echo "  ./scripts/run-tests.sh   - Run comprehensive tests"
echo ""
echo "🌐 Service URLs (when running):"
echo "  API:      http://localhost:8080"
echo "  Metrics:  http://localhost:9090"
echo "  Frontend: http://localhost:3000"
echo "  Grafana:  http://localhost:3001 (admin/admin)"
echo ""
echo "📝 Logs:"
echo "  docker-compose logs -f        - View all service logs"
echo "  docker-compose logs -f daemon - View daemon logs only"
echo ""

# Create helpful aliases in the shell
cat >> ~/.bashrc << 'EOF'

# AIOS Development Aliases
alias aios-logs='docker-compose -f deployments/docker-compose.dev.yml logs -f'
alias aios-status='docker-compose -f deployments/docker-compose.dev.yml ps'
alias aios-restart='docker-compose -f deployments/docker-compose.dev.yml restart'
alias aios-build='make build'
alias aios-test='make test'
alias aios-dev='make dev'

# Go development aliases
alias gob='go build'
alias got='go test'
alias gor='go run'
alias gof='go fmt'
alias gol='golangci-lint run'

# Docker aliases
alias dc='docker-compose'
alias dps='docker ps'
alias dlogs='docker logs'

# Kubernetes aliases (if using k8s)
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
alias kgd='kubectl get deployments'
EOF

# Also add to zsh if it exists
if [ -f ~/.zshrc ]; then
    cat >> ~/.zshrc << 'EOF'

# AIOS Development Aliases
alias aios-logs='docker-compose -f deployments/docker-compose.dev.yml logs -f'
alias aios-status='docker-compose -f deployments/docker-compose.dev.yml ps'
alias aios-restart='docker-compose -f deployments/docker-compose.dev.yml restart'
alias aios-build='make build'
alias aios-test='make test'
alias aios-dev='make dev'

# Go development aliases
alias gob='go build'
alias got='go test'
alias gor='go run'
alias gof='go fmt'
alias gol='golangci-lint run'

# Docker aliases
alias dc='docker-compose'
alias dps='docker ps'
alias dlogs='docker logs'

# Kubernetes aliases (if using k8s)
alias k='kubectl'
alias kgp='kubectl get pods'
alias kgs='kubectl get services'
alias kgd='kubectl get deployments'
EOF
fi

print_success "Development environment startup completed!"
echo "Happy coding! 🎉"
