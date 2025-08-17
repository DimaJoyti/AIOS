#!/bin/bash

# AIOS Security Check Script
# This script runs comprehensive security checks on the AIOS codebase

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

# Check if required tools are installed
check_tools() {
    log_info "Checking required security tools..."
    
    local tools=(
        "gosec"
        "govulncheck"
        "trivy"
        "gitleaks"
        "npm"
        "docker"
    )
    
    local missing_tools=()
    
    for tool in "${tools[@]}"; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Please install missing tools and try again"
        exit 1
    fi
    
    log_success "All required tools are installed"
}

# Install missing Go security tools
install_go_tools() {
    log_info "Installing Go security tools..."
    
    # Install gosec if not present
    if ! command -v gosec &> /dev/null; then
        log_info "Installing gosec..."
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    fi
    
    # Install govulncheck if not present
    if ! command -v govulncheck &> /dev/null; then
        log_info "Installing govulncheck..."
        go install golang.org/x/vuln/cmd/govulncheck@latest
    fi
    
    # Install go-licenses if not present
    if ! command -v go-licenses &> /dev/null; then
        log_info "Installing go-licenses..."
        go install github.com/google/go-licenses@latest
    fi
    
    log_success "Go security tools installed"
}

# Run Go security checks
run_go_security() {
    log_info "Running Go security checks..."
    
    # Run gosec
    log_info "Running gosec security scanner..."
    if gosec -fmt json -out gosec-report.json ./...; then
        log_success "Gosec scan completed successfully"
    else
        log_warning "Gosec found security issues - check gosec-report.json"
    fi
    
    # Run govulncheck
    log_info "Running govulncheck vulnerability scanner..."
    if govulncheck ./...; then
        log_success "No known vulnerabilities found in Go dependencies"
    else
        log_warning "Vulnerabilities found in Go dependencies"
    fi
    
    # Check Go licenses
    log_info "Checking Go dependency licenses..."
    if go-licenses check ./...; then
        log_success "All Go dependencies have compatible licenses"
        go-licenses csv ./... > go-licenses-report.csv
    else
        log_warning "Some Go dependencies may have incompatible licenses"
    fi
}

# Run Node.js security checks
run_node_security() {
    if [ ! -d "web" ]; then
        log_info "No web directory found, skipping Node.js security checks"
        return
    fi
    
    log_info "Running Node.js security checks..."
    
    cd web
    
    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        log_info "Installing Node.js dependencies..."
        npm ci
    fi
    
    # Run npm audit
    log_info "Running npm audit..."
    if npm audit --audit-level=moderate; then
        log_success "No moderate or higher vulnerabilities found in Node.js dependencies"
    else
        log_warning "Vulnerabilities found in Node.js dependencies"
        npm audit --json > ../npm-audit-report.json || true
    fi
    
    # Check Node.js licenses
    log_info "Checking Node.js dependency licenses..."
    if command -v license-checker &> /dev/null; then
        license-checker --csv --production > ../node-licenses-report.csv
        log_success "Node.js license report generated"
    else
        log_warning "license-checker not found, skipping Node.js license check"
    fi
    
    cd ..
}

# Run container security checks
run_container_security() {
    log_info "Running container security checks..."
    
    local services=("daemon" "assistant" "desktop")
    
    for service in "${services[@]}"; do
        local dockerfile="deployments/Dockerfile.$service"
        
        if [ ! -f "$dockerfile" ]; then
            log_warning "Dockerfile not found: $dockerfile"
            continue
        fi
        
        log_info "Scanning $dockerfile with Trivy..."
        
        # Build image for scanning
        local image_name="aios-$service:security-scan"
        if docker build -t "$image_name" -f "$dockerfile" .; then
            # Scan with Trivy
            if trivy image --format json --output "trivy-$service-report.json" "$image_name"; then
                log_success "Trivy scan completed for $service"
            else
                log_warning "Trivy found issues in $service container"
            fi
            
            # Clean up image
            docker rmi "$image_name" || true
        else
            log_error "Failed to build image for $service"
        fi
    done
}

# Run secrets scanning
run_secrets_scan() {
    log_info "Running secrets scanning..."
    
    # Run gitleaks
    if command -v gitleaks &> /dev/null; then
        log_info "Running gitleaks secrets scanner..."
        if gitleaks detect --source . --report-format json --report-path gitleaks-report.json; then
            log_success "No secrets found by gitleaks"
        else
            log_warning "Potential secrets found - check gitleaks-report.json"
        fi
    else
        log_warning "gitleaks not found, skipping secrets scan"
    fi
}

# Check configuration security
check_config_security() {
    log_info "Checking configuration security..."
    
    # Check for hardcoded secrets in config files
    log_info "Checking for hardcoded secrets in configuration files..."
    
    local config_issues=0
    
    # Check for hardcoded passwords
    if grep -r -i "password.*=" configs/ --include="*.yaml" --include="*.yml" | grep -v "\${" | grep -v "password: \"\"" | grep -v "password: ''" > /dev/null; then
        log_warning "Potential hardcoded passwords found in configuration files"
        config_issues=$((config_issues + 1))
    fi
    
    # Check for insecure SSL settings
    if grep -r "ssl_mode.*disable" configs/ --include="*.yaml" --include="*.yml" > /dev/null; then
        log_warning "SSL disabled in some configuration files"
        config_issues=$((config_issues + 1))
    fi
    
    # Check for debug mode in production configs
    if grep -r "debug.*true" configs/environments/production.yaml > /dev/null; then
        log_warning "Debug mode enabled in production configuration"
        config_issues=$((config_issues + 1))
    fi
    
    if [ $config_issues -eq 0 ]; then
        log_success "Configuration security checks passed"
    else
        log_warning "Found $config_issues configuration security issues"
    fi
}

# Generate security report
generate_report() {
    log_info "Generating security report..."
    
    local report_file="security-report.md"
    
    cat > "$report_file" << EOF
# AIOS Security Report

Generated on: $(date)

## Summary

This report contains the results of automated security scans performed on the AIOS codebase.

## Go Security (Gosec)

EOF
    
    if [ -f "gosec-report.json" ]; then
        echo "- Gosec report: gosec-report.json" >> "$report_file"
    else
        echo "- Gosec scan: Not run or no issues found" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Vulnerability Scanning

EOF
    
    if [ -f "npm-audit-report.json" ]; then
        echo "- Node.js vulnerabilities: npm-audit-report.json" >> "$report_file"
    else
        echo "- Node.js vulnerabilities: No issues found" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Container Security

EOF
    
    local container_reports=(trivy-*-report.json)
    if [ -f "${container_reports[0]}" ]; then
        for report in "${container_reports[@]}"; do
            echo "- Container scan: $report" >> "$report_file"
        done
    else
        echo "- Container scans: No reports generated" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Secrets Scanning

EOF
    
    if [ -f "gitleaks-report.json" ]; then
        echo "- Secrets scan: gitleaks-report.json" >> "$report_file"
    else
        echo "- Secrets scan: No secrets found" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## License Compliance

EOF
    
    if [ -f "go-licenses-report.csv" ]; then
        echo "- Go licenses: go-licenses-report.csv" >> "$report_file"
    fi
    
    if [ -f "node-licenses-report.csv" ]; then
        echo "- Node.js licenses: node-licenses-report.csv" >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Recommendations

1. Review all generated reports for security issues
2. Update dependencies with known vulnerabilities
3. Fix any configuration security issues
4. Ensure secrets are properly managed
5. Regularly run security scans

EOF
    
    log_success "Security report generated: $report_file"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Add cleanup logic here if needed
}

# Main function
main() {
    log_info "Starting AIOS security check..."
    
    # Set up cleanup trap
    trap cleanup EXIT
    
    # Check and install tools
    install_go_tools
    
    # Run security checks
    run_go_security
    run_node_security
    run_container_security
    run_secrets_scan
    check_config_security
    
    # Generate report
    generate_report
    
    log_success "Security check completed!"
    log_info "Review the generated reports and address any issues found"
}

# Run main function
main "$@"
