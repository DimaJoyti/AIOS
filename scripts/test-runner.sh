#!/bin/bash

# AIOS Test Runner Script
# Comprehensive testing automation for the AIOS platform

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_TIMEOUT=${TEST_TIMEOUT:-"10m"}
COVERAGE_THRESHOLD=${COVERAGE_THRESHOLD:-80}
PARALLEL_JOBS=${PARALLEL_JOBS:-4}
VERBOSE=${VERBOSE:-false}

# Directories
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TEST_RESULTS_DIR="${PROJECT_ROOT}/test-results"
COVERAGE_DIR="${PROJECT_ROOT}/coverage"

# Functions
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

print_header() {
    echo -e "\n${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}\n"
}

cleanup() {
    log_info "Cleaning up test environment..."
    
    # Stop any running test services
    pkill -f "aios-ai" || true
    pkill -f "aios-mcp-enhanced" || true
    pkill -f "npm start" || true
    
    # Clean up test databases
    if command -v docker &> /dev/null; then
        docker stop aios-test-postgres aios-test-redis 2>/dev/null || true
        docker rm aios-test-postgres aios-test-redis 2>/dev/null || true
    fi
    
    log_success "Cleanup completed"
}

setup_test_environment() {
    print_header "Setting Up Test Environment"
    
    # Create test directories
    mkdir -p "${TEST_RESULTS_DIR}"
    mkdir -p "${COVERAGE_DIR}"
    
    # Start test databases with Docker
    if command -v docker &> /dev/null; then
        log_info "Starting test databases..."
        
        # PostgreSQL
        docker run -d --name aios-test-postgres \
            -e POSTGRES_PASSWORD=test \
            -e POSTGRES_DB=aios_test \
            -p 5433:5432 \
            postgres:15 || log_warning "PostgreSQL container already running"
        
        # Redis
        docker run -d --name aios-test-redis \
            -p 6380:6379 \
            redis:7 || log_warning "Redis container already running"
        
        # Wait for databases to be ready
        log_info "Waiting for databases to be ready..."
        sleep 10
        
        # Test database connections
        if ! docker exec aios-test-postgres pg_isready -U postgres; then
            log_error "PostgreSQL is not ready"
            exit 1
        fi
        
        if ! docker exec aios-test-redis redis-cli ping; then
            log_error "Redis is not ready"
            exit 1
        fi
        
        log_success "Test databases are ready"
    else
        log_warning "Docker not available, using local databases"
    fi
    
    # Set test environment variables
    export DATABASE_URL="postgres://postgres:test@localhost:5433/aios_test?sslmode=disable"
    export REDIS_URL="redis://localhost:6380"
    export AIOS_TEST_MODE="true"
    export LOG_LEVEL="error"
    
    log_success "Test environment setup completed"
}

run_code_quality_checks() {
    print_header "Running Code Quality Checks"
    
    cd "${PROJECT_ROOT}"
    
    # Go formatting check
    log_info "Checking Go code formatting..."
    if [ "$(gofmt -s -l . | wc -l)" -gt 0 ]; then
        log_error "Code is not formatted properly:"
        gofmt -s -l .
        return 1
    fi
    log_success "Go code formatting is correct"
    
    # Go vet
    log_info "Running go vet..."
    if ! go vet ./...; then
        log_error "go vet found issues"
        return 1
    fi
    log_success "go vet passed"
    
    # golangci-lint (if available)
    if command -v golangci-lint &> /dev/null; then
        log_info "Running golangci-lint..."
        if ! golangci-lint run --timeout=5m; then
            log_error "golangci-lint found issues"
            return 1
        fi
        log_success "golangci-lint passed"
    else
        log_warning "golangci-lint not available, skipping"
    fi
    
    # Frontend linting (if Node.js is available)
    if command -v npm &> /dev/null && [ -d "web" ]; then
        log_info "Running frontend linting..."
        cd web
        if ! npm run lint; then
            log_error "Frontend linting failed"
            return 1
        fi
        cd ..
        log_success "Frontend linting passed"
    fi
    
    log_success "All code quality checks passed"
}

run_unit_tests() {
    print_header "Running Unit Tests"
    
    cd "${PROJECT_ROOT}"
    
    local coverage_file="${COVERAGE_DIR}/unit_coverage.out"
    local coverage_html="${COVERAGE_DIR}/unit_coverage.html"
    
    log_info "Running Go unit tests..."
    if ! go test -v -race -timeout="${TEST_TIMEOUT}" \
        -coverprofile="${coverage_file}" \
        -covermode=atomic \
        ./tests/unit/...; then
        log_error "Unit tests failed"
        return 1
    fi
    
    # Generate coverage report
    if [ -f "${coverage_file}" ]; then
        go tool cover -html="${coverage_file}" -o "${coverage_html}"
        
        # Check coverage threshold
        local coverage_percent=$(go tool cover -func="${coverage_file}" | grep total | awk '{print $3}' | sed 's/%//')
        log_info "Unit test coverage: ${coverage_percent}%"
        
        if (( $(echo "${coverage_percent} < ${COVERAGE_THRESHOLD}" | bc -l) )); then
            log_warning "Coverage ${coverage_percent}% is below threshold ${COVERAGE_THRESHOLD}%"
        else
            log_success "Coverage ${coverage_percent}% meets threshold ${COVERAGE_THRESHOLD}%"
        fi
    fi
    
    # Frontend unit tests (if available)
    if command -v npm &> /dev/null && [ -d "web" ]; then
        log_info "Running frontend unit tests..."
        cd web
        if ! npm test; then
            log_error "Frontend unit tests failed"
            return 1
        fi
        cd ..
        log_success "Frontend unit tests passed"
    fi
    
    log_success "All unit tests passed"
}

run_integration_tests() {
    print_header "Running Integration Tests"
    
    cd "${PROJECT_ROOT}"
    
    local coverage_file="${COVERAGE_DIR}/integration_coverage.out"
    
    log_info "Running integration tests..."
    if ! go test -v -race -timeout="${TEST_TIMEOUT}" \
        -coverprofile="${coverage_file}" \
        ./tests/integration/...; then
        log_error "Integration tests failed"
        return 1
    fi
    
    log_success "Integration tests passed"
}

run_performance_tests() {
    print_header "Running Performance Tests"
    
    cd "${PROJECT_ROOT}"
    
    local benchmark_file="${TEST_RESULTS_DIR}/benchmark_results.txt"
    
    log_info "Running performance benchmarks..."
    if ! go test -bench=. -benchmem -run=^$ \
        ./tests/performance/... > "${benchmark_file}"; then
        log_error "Performance tests failed"
        return 1
    fi
    
    log_info "Benchmark results saved to ${benchmark_file}"
    
    # Display key benchmark results
    if [ -f "${benchmark_file}" ]; then
        log_info "Key benchmark results:"
        grep -E "Benchmark.*-[0-9]+" "${benchmark_file}" | head -10
    fi
    
    log_success "Performance tests completed"
}

build_services() {
    print_header "Building Services"
    
    cd "${PROJECT_ROOT}"
    
    log_info "Building AI service..."
    if ! go build -o bin/aios-ai ./cmd/aios-ai; then
        log_error "Failed to build AI service"
        return 1
    fi
    
    log_info "Building MCP service..."
    if ! go build -o bin/aios-mcp-enhanced ./cmd/aios-mcp-enhanced; then
        log_error "Failed to build MCP service"
        return 1
    fi
    
    # Build frontend
    if command -v npm &> /dev/null && [ -d "web" ]; then
        log_info "Building frontend..."
        cd web
        if ! npm run build; then
            log_error "Failed to build frontend"
            return 1
        fi
        cd ..
    fi
    
    log_success "All services built successfully"
}

run_e2e_tests() {
    print_header "Running End-to-End Tests"
    
    cd "${PROJECT_ROOT}"
    
    # Start services
    log_info "Starting services for E2E tests..."
    
    ./bin/aios-ai &
    AI_PID=$!
    
    ./bin/aios-mcp-enhanced &
    MCP_PID=$!
    
    if [ -d "web" ]; then
        cd web
        npm start &
        WEB_PID=$!
        cd ..
    fi
    
    # Wait for services to start
    log_info "Waiting for services to start..."
    sleep 30
    
    # Run E2E tests
    log_info "Running E2E tests..."
    local e2e_success=true
    
    if ! go test -v -timeout="${TEST_TIMEOUT}" ./tests/e2e/...; then
        log_error "E2E tests failed"
        e2e_success=false
    fi
    
    # Stop services
    log_info "Stopping services..."
    kill $AI_PID $MCP_PID 2>/dev/null || true
    [ ! -z "$WEB_PID" ] && kill $WEB_PID 2>/dev/null || true
    
    if [ "$e2e_success" = true ]; then
        log_success "E2E tests passed"
    else
        return 1
    fi
}

generate_test_report() {
    print_header "Generating Test Report"
    
    local report_file="${TEST_RESULTS_DIR}/test_report.html"
    
    cat > "${report_file}" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>AIOS Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background-color: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; padding: 15px; border: 1px solid #ddd; border-radius: 5px; }
        .success { background-color: #d4edda; border-color: #c3e6cb; }
        .warning { background-color: #fff3cd; border-color: #ffeaa7; }
        .error { background-color: #f8d7da; border-color: #f5c6cb; }
        pre { background-color: #f8f9fa; padding: 10px; border-radius: 3px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="header">
        <h1>AIOS Test Report</h1>
        <p>Generated on: $(date)</p>
        <p>Test run completed at: $(date)</p>
    </div>
    
    <div class="section success">
        <h2>Test Summary</h2>
        <p>All tests completed successfully!</p>
    </div>
    
    <div class="section">
        <h2>Coverage Reports</h2>
        <p>Coverage reports are available in the coverage directory:</p>
        <ul>
            <li><a href="../coverage/unit_coverage.html">Unit Test Coverage</a></li>
        </ul>
    </div>
    
    <div class="section">
        <h2>Performance Results</h2>
        <p>Benchmark results are available in: ${TEST_RESULTS_DIR}/benchmark_results.txt</p>
    </div>
</body>
</html>
EOF
    
    log_success "Test report generated: ${report_file}"
}

# Main execution
main() {
    local test_type="${1:-all}"
    
    print_header "AIOS Test Runner"
    log_info "Running test type: ${test_type}"
    log_info "Project root: ${PROJECT_ROOT}"
    
    # Set up signal handlers for cleanup
    trap cleanup EXIT INT TERM
    
    case "${test_type}" in
        "quality")
            run_code_quality_checks
            ;;
        "unit")
            setup_test_environment
            run_unit_tests
            ;;
        "integration")
            setup_test_environment
            run_integration_tests
            ;;
        "performance")
            setup_test_environment
            run_performance_tests
            ;;
        "e2e")
            setup_test_environment
            build_services
            run_e2e_tests
            ;;
        "all")
            setup_test_environment
            run_code_quality_checks
            run_unit_tests
            run_integration_tests
            run_performance_tests
            build_services
            run_e2e_tests
            generate_test_report
            ;;
        *)
            log_error "Unknown test type: ${test_type}"
            echo "Usage: $0 [quality|unit|integration|performance|e2e|all]"
            exit 1
            ;;
    esac
    
    log_success "Test run completed successfully!"
}

# Check if script is being sourced or executed
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
