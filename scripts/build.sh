#!/bin/bash

# AIOS Build Script
# Comprehensive build automation for AIOS project

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build"
BIN_DIR="$BUILD_DIR/bin"
DIST_DIR="$BUILD_DIR/dist"
REPORTS_DIR="$BUILD_DIR/reports"

# Build configuration
SERVICES=("daemon" "assistant" "desktop")
PLATFORMS=("linux/amd64" "linux/arm64" "darwin/amd64" "darwin/arm64" "windows/amd64")

# Version information
VERSION=${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo "dev")}
COMMIT=${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo "unknown")}
BUILD_TIME=${BUILD_TIME:-$(date -u '+%Y-%m-%d_%H:%M:%S')}
BRANCH=${BRANCH:-$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")}

# Build flags
LDFLAGS="-s -w -X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME -X main.Branch=$BRANCH"
LDFLAGS_DEV="-X main.Version=$VERSION -X main.Commit=$COMMIT -X main.BuildTime=$BUILD_TIME -X main.Branch=$BRANCH"

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

# Check if required tools are installed
check_dependencies() {
    log_info "Checking build dependencies..."
    
    local deps=("go" "git")
    local missing_deps=()
    
    for dep in "${deps[@]}"; do
        if ! command -v "$dep" &> /dev/null; then
            missing_deps+=("$dep")
        fi
    done
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Setup build directories
setup_directories() {
    log_info "Setting up build directories..."
    mkdir -p "$BIN_DIR" "$DIST_DIR" "$REPORTS_DIR"
    log_success "Build directories created"
}

# Clean previous builds
clean_build() {
    log_info "Cleaning previous builds..."
    rm -rf "$BUILD_DIR"
    go clean -cache -modcache -testcache
    log_success "Build artifacts cleaned"
}

# Download and verify dependencies
setup_dependencies() {
    log_info "Setting up Go dependencies..."
    cd "$PROJECT_ROOT"
    go mod download
    go mod verify
    go mod tidy
    log_success "Dependencies setup complete"
}

# Build single service
build_service() {
    local service=$1
    local output_dir=$2
    local ldflags=$3
    local goos=${4:-}
    local goarch=${5:-}
    
    local output_name="aios-$service"
    if [ -n "$goos" ] && [ -n "$goarch" ]; then
        output_name="aios-$service-$goos-$goarch"
        if [ "$goos" = "windows" ]; then
            output_name="$output_name.exe"
        fi
    fi
    
    local output_path="$output_dir/$output_name"
    
    log_info "Building $service for ${goos:-native}/${goarch:-native}..."
    
    # Set environment variables for cross-compilation
    local env_vars=""
    if [ -n "$goos" ]; then
        env_vars="GOOS=$goos"
    fi
    if [ -n "$goarch" ]; then
        env_vars="$env_vars GOARCH=$goarch"
    fi
    
    # Build the service
    eval "$env_vars CGO_ENABLED=0 go build -ldflags=\"$ldflags\" -o \"$output_path\" ./cmd/aios-$service"
    
    # Verify the build
    if [ -f "$output_path" ]; then
        log_success "Built $service successfully: $output_path"
        
        # Show binary info
        local size=$(du -h "$output_path" | cut -f1)
        log_info "Binary size: $size"
    else
        log_error "Failed to build $service"
        return 1
    fi
}

# Build all services for native platform
build_native() {
    log_info "Building all services for native platform..."
    
    for service in "${SERVICES[@]}"; do
        build_service "$service" "$BIN_DIR" "$LDFLAGS_DEV"
    done
    
    log_success "Native build completed"
}

# Build all services for all platforms
build_cross() {
    log_info "Cross-compiling for all platforms..."
    
    for platform in "${PLATFORMS[@]}"; do
        local goos=$(echo "$platform" | cut -d'/' -f1)
        local goarch=$(echo "$platform" | cut -d'/' -f2)
        
        log_info "Building for $goos/$goarch..."
        
        for service in "${SERVICES[@]}"; do
            build_service "$service" "$DIST_DIR" "$LDFLAGS" "$goos" "$goarch"
        done
    done
    
    log_success "Cross-compilation completed"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    cd "$PROJECT_ROOT"
    
    # Run unit tests with coverage
    go test -v -race -coverprofile="$REPORTS_DIR/coverage.out" -covermode=atomic ./...
    
    # Generate coverage report
    go tool cover -html="$REPORTS_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"
    
    # Show coverage summary
    local coverage=$(go tool cover -func="$REPORTS_DIR/coverage.out" | tail -1 | awk '{print $3}')
    log_info "Test coverage: $coverage"
    
    log_success "Tests completed"
}

# Run linting
run_lint() {
    log_info "Running code linting..."
    cd "$PROJECT_ROOT"
    
    # Check if golangci-lint is installed
    if ! command -v golangci-lint &> /dev/null; then
        log_warning "golangci-lint not found, installing..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    # Run linter
    golangci-lint run --out-format=checkstyle > "$REPORTS_DIR/lint-report.xml" || true
    golangci-lint run
    
    log_success "Linting completed"
}

# Generate build info
generate_build_info() {
    log_info "Generating build information..."
    
    cat > "$BUILD_DIR/build-info.json" << EOF
{
  "version": "$VERSION",
  "commit": "$COMMIT",
  "branch": "$BRANCH",
  "build_time": "$BUILD_TIME",
  "go_version": "$(go version | cut -d' ' -f3)",
  "platform": "$(go env GOOS)/$(go env GOARCH)",
  "services": [$(printf '"%s",' "${SERVICES[@]}" | sed 's/,$//')]
}
EOF
    
    log_success "Build information generated"
}

# Create checksums
create_checksums() {
    log_info "Creating checksums..."
    cd "$DIST_DIR"
    
    if [ -n "$(ls -A . 2>/dev/null)" ]; then
        sha256sum * > checksums.txt
        log_success "Checksums created"
    else
        log_warning "No files to checksum in $DIST_DIR"
    fi
}

# Show build summary
show_summary() {
    log_info "Build Summary:"
    echo "  Version: $VERSION"
    echo "  Commit: $COMMIT"
    echo "  Branch: $BRANCH"
    echo "  Build Time: $BUILD_TIME"
    echo "  Build Directory: $BUILD_DIR"
    
    if [ -d "$BIN_DIR" ] && [ -n "$(ls -A "$BIN_DIR" 2>/dev/null)" ]; then
        echo "  Native Binaries:"
        ls -la "$BIN_DIR" | grep -E "aios-" | awk '{print "    " $9 " (" $5 " bytes)"}'
    fi
    
    if [ -d "$DIST_DIR" ] && [ -n "$(ls -A "$DIST_DIR" 2>/dev/null)" ]; then
        echo "  Distribution Packages:"
        ls -la "$DIST_DIR" | grep -E "\.(tar\.gz|zip|exe)$" | wc -l | awk '{print "    " $1 " packages created"}'
    fi
}

# Main build function
main() {
    local build_type=${1:-"native"}
    
    log_info "Starting AIOS build process..."
    log_info "Build type: $build_type"
    
    # Setup
    check_dependencies
    setup_directories
    setup_dependencies
    
    # Build based on type
    case "$build_type" in
        "clean")
            clean_build
            ;;
        "native")
            build_native
            ;;
        "cross")
            build_cross
            create_checksums
            ;;
        "release")
            run_lint
            run_tests
            build_cross
            create_checksums
            ;;
        "ci")
            run_lint
            run_tests
            build_native
            ;;
        "test")
            run_tests
            ;;
        "lint")
            run_lint
            ;;
        *)
            log_error "Unknown build type: $build_type"
            echo "Usage: $0 [clean|native|cross|release|ci|test|lint]"
            exit 1
            ;;
    esac
    
    # Generate build info
    generate_build_info
    
    # Show summary
    show_summary
    
    log_success "Build process completed successfully!"
}

# Run main function with all arguments
main "$@"
