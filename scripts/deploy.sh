#!/bin/bash

# AIOS Deployment Script
# This script handles deployment of AIOS to various environments

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="${VERSION:-latest}"
ENVIRONMENT="${ENVIRONMENT:-development}"
PLATFORM="${PLATFORM:-docker}"
DRY_RUN="${DRY_RUN:-false}"
FORCE="${FORCE:-false}"

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

# Help function
show_help() {
    cat << EOF
AIOS Deployment Script

Usage: $0 [OPTIONS]

Options:
    -e, --environment ENV    Target environment (development, staging, production)
    -p, --platform PLATFORM Deployment platform (docker, kubernetes, aws, gcp, azure)
    -v, --version VERSION    Version to deploy (default: latest)
    -d, --dry-run           Perform a dry run without making changes
    -f, --force             Force deployment even if validation fails
    -h, --help              Show this help message

Examples:
    $0 --environment production --platform kubernetes --version 1.2.0
    $0 --environment staging --platform docker --dry-run
    $0 --environment development --platform docker

Environment Variables:
    VERSION                 Version to deploy
    ENVIRONMENT            Target environment
    PLATFORM               Deployment platform
    DRY_RUN                Perform dry run (true/false)
    FORCE                  Force deployment (true/false)
    DOCKER_REGISTRY        Docker registry URL
    KUBECONFIG             Kubernetes config file path
    AWS_PROFILE            AWS profile for AWS deployments
    GCP_PROJECT            GCP project for GCP deployments
    AZURE_SUBSCRIPTION     Azure subscription for Azure deployments

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -p|--platform)
                PLATFORM="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dry-run)
                DRY_RUN="true"
                shift
                ;;
            -f|--force)
                FORCE="true"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Validate environment
validate_environment() {
    log_info "Validating deployment environment..."
    
    case $ENVIRONMENT in
        development|staging|production)
            log_success "Environment '$ENVIRONMENT' is valid"
            ;;
        *)
            log_error "Invalid environment: $ENVIRONMENT"
            log_error "Valid environments: development, staging, production"
            exit 1
            ;;
    esac
    
    case $PLATFORM in
        docker|kubernetes|aws|gcp|azure)
            log_success "Platform '$PLATFORM' is valid"
            ;;
        *)
            log_error "Invalid platform: $PLATFORM"
            log_error "Valid platforms: docker, kubernetes, aws, gcp, azure"
            exit 1
            ;;
    esac
}

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites for $PLATFORM deployment..."
    
    case $PLATFORM in
        docker)
            if ! command -v docker &> /dev/null; then
                log_error "Docker is not installed or not in PATH"
                exit 1
            fi
            
            if ! command -v docker-compose &> /dev/null; then
                log_error "Docker Compose is not installed or not in PATH"
                exit 1
            fi
            
            if ! docker info &> /dev/null; then
                log_error "Docker daemon is not running"
                exit 1
            fi
            ;;
            
        kubernetes)
            if ! command -v kubectl &> /dev/null; then
                log_error "kubectl is not installed or not in PATH"
                exit 1
            fi
            
            if ! kubectl cluster-info &> /dev/null; then
                log_error "Cannot connect to Kubernetes cluster"
                exit 1
            fi
            ;;
            
        aws)
            if ! command -v aws &> /dev/null; then
                log_error "AWS CLI is not installed or not in PATH"
                exit 1
            fi
            
            if ! aws sts get-caller-identity &> /dev/null; then
                log_error "AWS credentials not configured"
                exit 1
            fi
            ;;
            
        gcp)
            if ! command -v gcloud &> /dev/null; then
                log_error "Google Cloud SDK is not installed or not in PATH"
                exit 1
            fi
            
            if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -n1 &> /dev/null; then
                log_error "GCP credentials not configured"
                exit 1
            fi
            ;;
            
        azure)
            if ! command -v az &> /dev/null; then
                log_error "Azure CLI is not installed or not in PATH"
                exit 1
            fi
            
            if ! az account show &> /dev/null; then
                log_error "Azure credentials not configured"
                exit 1
            fi
            ;;
    esac
    
    log_success "Prerequisites check passed"
}

# Build application
build_application() {
    log_info "Building AIOS application..."
    
    cd "$PROJECT_ROOT"
    
    # Build Go binaries
    log_info "Building Go binaries..."
    make build
    
    # Build Docker images if needed
    if [[ "$PLATFORM" == "docker" || "$PLATFORM" == "kubernetes" ]]; then
        log_info "Building Docker images..."
        make build-images VERSION="$VERSION"
    fi
    
    log_success "Application build completed"
}

# Deploy to Docker
deploy_docker() {
    log_info "Deploying to Docker..."
    
    cd "$PROJECT_ROOT"
    
    local compose_file="docker-compose.yml"
    if [[ "$ENVIRONMENT" == "development" ]]; then
        compose_file="deployments/docker-compose.dev.yml"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Dry run: Would execute: docker-compose -f $compose_file up -d"
        return 0
    fi
    
    # Stop existing containers
    docker-compose -f "$compose_file" down || true
    
    # Start new containers
    docker-compose -f "$compose_file" up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    sleep 10
    
    # Health check
    if curl -f http://localhost:8080/health &> /dev/null; then
        log_success "AIOS is running and healthy"
    else
        log_error "AIOS health check failed"
        exit 1
    fi
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    cd "$PROJECT_ROOT"
    
    local namespace="aios"
    if [[ "$ENVIRONMENT" == "development" ]]; then
        namespace="aios-dev"
    elif [[ "$ENVIRONMENT" == "staging" ]]; then
        namespace="aios-staging"
    fi
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "Dry run: Would deploy to namespace $namespace"
        kubectl apply --dry-run=client -f deployments/k8s/ -n "$namespace"
        return 0
    fi
    
    # Create namespace if it doesn't exist
    kubectl create namespace "$namespace" --dry-run=client -o yaml | kubectl apply -f -
    
    # Apply configurations
    kubectl apply -f deployments/k8s/configmap.yaml -n "$namespace"
    kubectl apply -f deployments/k8s/secret.yaml -n "$namespace"
    
    # Deploy infrastructure
    kubectl apply -f deployments/k8s/postgres.yaml -n "$namespace"
    kubectl apply -f deployments/k8s/redis.yaml -n "$namespace"
    
    # Wait for infrastructure to be ready
    kubectl wait --for=condition=ready pod -l app=postgres -n "$namespace" --timeout=300s
    kubectl wait --for=condition=ready pod -l app=redis -n "$namespace" --timeout=300s
    
    # Deploy AIOS services
    kubectl apply -f deployments/k8s/aios-daemon.yaml -n "$namespace"
    kubectl apply -f deployments/k8s/aios-assistant.yaml -n "$namespace"
    
    # Wait for deployment to be ready
    kubectl wait --for=condition=available deployment/aios-daemon -n "$namespace" --timeout=600s
    kubectl wait --for=condition=available deployment/aios-assistant -n "$namespace" --timeout=600s
    
    log_success "Kubernetes deployment completed"
}

# Deploy to AWS
deploy_aws() {
    log_info "Deploying to AWS..."
    
    # TODO: Implement AWS deployment (ECS, EKS, etc.)
    log_warning "AWS deployment not yet implemented"
}

# Deploy to GCP
deploy_gcp() {
    log_info "Deploying to GCP..."
    
    # TODO: Implement GCP deployment (GKE, Cloud Run, etc.)
    log_warning "GCP deployment not yet implemented"
}

# Deploy to Azure
deploy_azure() {
    log_info "Deploying to Azure..."
    
    # TODO: Implement Azure deployment (AKS, Container Instances, etc.)
    log_warning "Azure deployment not yet implemented"
}

# Main deployment function
deploy() {
    log_info "Starting AIOS deployment..."
    log_info "Environment: $ENVIRONMENT"
    log_info "Platform: $PLATFORM"
    log_info "Version: $VERSION"
    log_info "Dry Run: $DRY_RUN"
    
    case $PLATFORM in
        docker)
            deploy_docker
            ;;
        kubernetes)
            deploy_kubernetes
            ;;
        aws)
            deploy_aws
            ;;
        gcp)
            deploy_gcp
            ;;
        azure)
            deploy_azure
            ;;
    esac
    
    log_success "AIOS deployment completed successfully!"
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Add any cleanup logic here
}

# Trap cleanup on exit
trap cleanup EXIT

# Main execution
main() {
    parse_args "$@"
    validate_environment
    check_prerequisites
    build_application
    deploy
}

# Run main function if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
