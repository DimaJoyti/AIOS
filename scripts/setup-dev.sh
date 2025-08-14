#!/bin/bash

# AIOS Development Environment Setup Script
# This script sets up the complete development environment for AIOS

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on Arch Linux
check_arch_linux() {
    if [[ ! -f /etc/arch-release ]]; then
        log_error "This script is designed for Arch Linux. Please run on an Arch Linux system."
        exit 1
    fi
    log_success "Arch Linux detected"
}

# Check if running as root
check_not_root() {
    if [[ $EUID -eq 0 ]]; then
        log_error "This script should not be run as root. Please run as a regular user."
        exit 1
    fi
}

# Update system packages
update_system() {
    log_info "Updating system packages..."
    sudo pacman -Syu --noconfirm
    log_success "System packages updated"
}

# Install base development tools
install_base_tools() {
    log_info "Installing base development tools..."
    
    local packages=(
        "base-devel"
        "git"
        "curl"
        "wget"
        "unzip"
        "vim"
        "nano"
        "htop"
        "tree"
        "jq"
        "yq"
    )
    
    sudo pacman -S --needed --noconfirm "${packages[@]}"
    log_success "Base development tools installed"
}

# Install Go
install_go() {
    log_info "Installing Go..."
    
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        log_info "Go $go_version is already installed"
        
        # Check if version is 1.22 or higher
        if [[ $(echo "$go_version 1.22" | tr " " "\n" | sort -V | head -n1) == "1.22" ]]; then
            log_success "Go version is compatible"
            return
        else
            log_warning "Go version is too old, updating..."
        fi
    fi
    
    sudo pacman -S --needed --noconfirm go
    
    # Set up Go environment
    if ! grep -q "export GOPATH" ~/.bashrc; then
        echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
    fi
    
    # Source the changes
    export GOPATH=$HOME/go
    export PATH=$PATH:$GOPATH/bin
    
    log_success "Go installed and configured"
}

# Install Node.js and npm
install_nodejs() {
    log_info "Installing Node.js and npm..."
    
    if command -v node &> /dev/null; then
        local node_version=$(node --version | sed 's/v//')
        log_info "Node.js $node_version is already installed"
        
        # Check if version is 18 or higher
        if [[ $(echo "$node_version 18.0.0" | tr " " "\n" | sort -V | head -n1) == "18.0.0" ]]; then
            log_success "Node.js version is compatible"
            return
        else
            log_warning "Node.js version is too old, updating..."
        fi
    fi
    
    sudo pacman -S --needed --noconfirm nodejs npm
    log_success "Node.js and npm installed"
}

# Install Docker and Docker Compose
install_docker() {
    log_info "Installing Docker and Docker Compose..."
    
    sudo pacman -S --needed --noconfirm docker docker-compose
    
    # Enable and start Docker service
    sudo systemctl enable docker
    sudo systemctl start docker
    
    # Add user to docker group
    sudo usermod -aG docker $USER
    
    log_success "Docker and Docker Compose installed"
    log_warning "Please log out and log back in for Docker group changes to take effect"
}

# Install development tools
install_dev_tools() {
    log_info "Installing development tools..."
    
    local packages=(
        "code"  # Visual Studio Code
        "git-lfs"
        "github-cli"
        "make"
        "cmake"
        "gcc"
        "clang"
        "llvm"
        "gdb"
        "valgrind"
        "strace"
        "lsof"
        "netstat-nat"
        "tcpdump"
        "wireshark-cli"
    )
    
    # Install from official repositories
    sudo pacman -S --needed --noconfirm "${packages[@]}" || true
    
    log_success "Development tools installed"
}

# Install Go development tools
install_go_tools() {
    log_info "Installing Go development tools..."
    
    # Ensure Go is in PATH
    export PATH=$PATH:$GOPATH/bin:/usr/local/go/bin
    
    local tools=(
        "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
        "golang.org/x/tools/cmd/goimports@latest"
        "golang.org/x/tools/cmd/godoc@latest"
        "github.com/swaggo/swag/cmd/swag@latest"
        "github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        "github.com/air-verse/air@latest"
        "github.com/go-delve/delve/cmd/dlv@latest"
    )
    
    for tool in "${tools[@]}"; do
        log_info "Installing $tool..."
        go install "$tool" || log_warning "Failed to install $tool"
    done
    
    log_success "Go development tools installed"
}

# Install AI/ML tools
install_ai_tools() {
    log_info "Installing AI/ML tools..."
    
    local packages=(
        "python"
        "python-pip"
        "python-virtualenv"
        "python-numpy"
        "python-scipy"
        "python-matplotlib"
        "python-pandas"
        "python-scikit-learn"
        "python-tensorflow"
        "python-pytorch"
        "opencv"
        "python-opencv"
    )
    
    sudo pacman -S --needed --noconfirm "${packages[@]}" || true
    
    # Install Ollama for local LLM
    if ! command -v ollama &> /dev/null; then
        log_info "Installing Ollama..."
        curl -fsSL https://ollama.ai/install.sh | sh
        log_success "Ollama installed"
    else
        log_info "Ollama is already installed"
    fi
    
    log_success "AI/ML tools installed"
}

# Set up project dependencies
setup_project_deps() {
    log_info "Setting up project dependencies..."
    
    # Go dependencies
    if [[ -f "go.mod" ]]; then
        log_info "Installing Go dependencies..."
        go mod download
        go mod tidy
        log_success "Go dependencies installed"
    fi
    
    # Node.js dependencies
    if [[ -f "web/package.json" ]]; then
        log_info "Installing Node.js dependencies..."
        cd web
        npm install
        cd ..
        log_success "Node.js dependencies installed"
    fi
}

# Set up development database
setup_dev_database() {
    log_info "Setting up development database..."
    
    # Install PostgreSQL
    sudo pacman -S --needed --noconfirm postgresql
    
    # Initialize database if not already done
    if [[ ! -d "/var/lib/postgres/data" ]]; then
        sudo -u postgres initdb -D /var/lib/postgres/data
    fi
    
    # Enable and start PostgreSQL
    sudo systemctl enable postgresql
    sudo systemctl start postgresql
    
    # Create development database and user
    sudo -u postgres createuser -s $USER 2>/dev/null || true
    sudo -u postgres createdb aios_dev 2>/dev/null || true
    
    log_success "Development database set up"
}

# Set up Redis
setup_redis() {
    log_info "Setting up Redis..."
    
    sudo pacman -S --needed --noconfirm redis
    
    # Enable and start Redis
    sudo systemctl enable redis
    sudo systemctl start redis
    
    log_success "Redis set up"
}

# Create development configuration
create_dev_config() {
    log_info "Creating development configuration..."
    
    mkdir -p configs
    
    cat > configs/dev.yaml << EOF
# AIOS Development Configuration

server:
  host: "localhost"
  port: 8080
  metrics_port: 9090

database:
  host: "localhost"
  port: 5432
  name: "aios_dev"
  user: "$USER"
  password: ""
  ssl_mode: "disable"

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0

ai:
  ollama:
    host: "localhost"
    port: 11434
  models_path: "./models"

logging:
  level: "debug"
  format: "json"

tracing:
  jaeger_endpoint: "http://localhost:14268/api/traces"

security:
  jwt_secret: "dev-secret-key-change-in-production"
  session_timeout: "24h"
EOF
    
    log_success "Development configuration created"
}

# Set up Git hooks
setup_git_hooks() {
    log_info "Setting up Git hooks..."
    
    mkdir -p .git/hooks
    
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
# Pre-commit hook for AIOS

set -e

echo "Running pre-commit checks..."

# Format Go code
echo "Formatting Go code..."
gofmt -s -w .
goimports -w .

# Run Go linter
echo "Running Go linter..."
golangci-lint run

# Run Go tests
echo "Running Go tests..."
go test ./...

# Format frontend code if it exists
if [[ -d "web" ]]; then
    echo "Formatting frontend code..."
    cd web
    npm run format || true
    cd ..
fi

echo "Pre-commit checks passed!"
EOF
    
    chmod +x .git/hooks/pre-commit
    
    log_success "Git hooks set up"
}

# Main setup function
main() {
    log_info "Starting AIOS development environment setup..."
    
    check_arch_linux
    check_not_root
    
    update_system
    install_base_tools
    install_go
    install_nodejs
    install_docker
    install_dev_tools
    install_go_tools
    install_ai_tools
    setup_project_deps
    setup_dev_database
    setup_redis
    create_dev_config
    setup_git_hooks
    
    log_success "AIOS development environment setup completed!"
    log_info "Please log out and log back in for all changes to take effect."
    log_info "Run 'make dev' to start the development environment."
}

# Run main function
main "$@"
