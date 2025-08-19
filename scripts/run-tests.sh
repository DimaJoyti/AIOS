#!/bin/bash

# Quick test runner for development
# This is a simplified version of the comprehensive test-runner.sh

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Quick unit tests
run_quick_tests() {
    log_info "Running quick unit tests..."
    
    if go test -short -race ./tests/unit/...; then
        log_success "Quick tests passed"
    else
        log_error "Quick tests failed"
        exit 1
    fi
}

# Build verification
verify_build() {
    log_info "Verifying build..."
    
    if go build -o /tmp/aios-ai ./cmd/aios-ai && \
       go build -o /tmp/aios-mcp-enhanced ./cmd/aios-mcp-enhanced; then
        log_success "Build verification passed"
        rm -f /tmp/aios-ai /tmp/aios-mcp-enhanced
    else
        log_error "Build verification failed"
        exit 1
    fi
}

# Code formatting check
check_formatting() {
    log_info "Checking code formatting..."
    
    if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
        log_error "Code is not formatted properly. Run: gofmt -s -w ."
        gofmt -s -l .
        exit 1
    else
        log_success "Code formatting is correct"
    fi
}

# Main execution
main() {
    local test_type="${1:-quick}"
    
    case "${test_type}" in
        "quick")
            check_formatting
            verify_build
            run_quick_tests
            ;;
        "format")
            check_formatting
            ;;
        "build")
            verify_build
            ;;
        *)
            echo "Usage: $0 [quick|format|build]"
            exit 1
            ;;
    esac
    
    log_success "All checks passed!"
}

main "$@"
