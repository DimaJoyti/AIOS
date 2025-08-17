#!/bin/bash

# AIOS Monitoring Setup Script
# Sets up comprehensive monitoring infrastructure

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

# Configuration
ENVIRONMENT=${ENVIRONMENT:-"development"}
MONITORING_NAMESPACE=${MONITORING_NAMESPACE:-"aios-monitoring"}

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

# Check dependencies
check_dependencies() {
    log_info "Checking monitoring dependencies..."
    
    local deps=("docker" "docker-compose")
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
    
    log_success "Dependencies check completed"
}

# Create monitoring configuration
create_monitoring_config() {
    log_info "Creating monitoring configuration..."
    
    local monitoring_dir="$PROJECT_ROOT/deployments/monitoring"
    mkdir -p "$monitoring_dir"
    
    # Create Docker Compose file for monitoring stack
    cat > "$monitoring_dir/docker-compose.yml" << 'EOF'
version: '3.8'

services:
  # Prometheus
  prometheus:
    image: prom/prometheus:latest
    container_name: aios-prometheus
    restart: unless-stopped
    ports:
      - "9090:9090"
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
      - '--web.enable-admin-api'
    volumes:
      - ../../configs/prometheus.${ENVIRONMENT:-dev}.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus-data:/prometheus
    networks:
      - monitoring

  # Grafana
  grafana:
    image: grafana/grafana:latest
    container_name: aios-grafana
    restart: unless-stopped
    ports:
      - "3001:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD:-admin}
      - GF_USERS_ALLOW_SIGN_UP=false
      - GF_INSTALL_PLUGINS=grafana-piechart-panel,grafana-clock-panel
    volumes:
      - grafana-data:/var/lib/grafana
      - ../../configs/grafana/provisioning:/etc/grafana/provisioning:ro
    networks:
      - monitoring
    depends_on:
      - prometheus

  # AlertManager
  alertmanager:
    image: prom/alertmanager:latest
    container_name: aios-alertmanager
    restart: unless-stopped
    ports:
      - "9093:9093"
    volumes:
      - ../../configs/alertmanager.yml:/etc/alertmanager/alertmanager.yml:ro
      - alertmanager-data:/alertmanager
    networks:
      - monitoring

  # Node Exporter
  node-exporter:
    image: prom/node-exporter:latest
    container_name: aios-node-exporter
    restart: unless-stopped
    ports:
      - "9100:9100"
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    networks:
      - monitoring

  # cAdvisor
  cadvisor:
    image: gcr.io/cadvisor/cadvisor:latest
    container_name: aios-cadvisor
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - /:/rootfs:ro
      - /var/run:/var/run:ro
      - /sys:/sys:ro
      - /var/lib/docker/:/var/lib/docker:ro
      - /dev/disk/:/dev/disk:ro
    privileged: true
    devices:
      - /dev/kmsg
    networks:
      - monitoring

  # Jaeger (for tracing)
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: aios-jaeger
    restart: unless-stopped
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # HTTP collector
      - "14250:14250"  # gRPC collector
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    volumes:
      - jaeger-data:/tmp
    networks:
      - monitoring

  # Loki (for log aggregation)
  loki:
    image: grafana/loki:latest
    container_name: aios-loki
    restart: unless-stopped
    ports:
      - "3100:3100"
    command: -config.file=/etc/loki/local-config.yaml
    volumes:
      - loki-data:/loki
    networks:
      - monitoring

  # Promtail (log collector)
  promtail:
    image: grafana/promtail:latest
    container_name: aios-promtail
    restart: unless-stopped
    volumes:
      - /var/log:/var/log:ro
      - /var/lib/docker/containers:/var/lib/docker/containers:ro
      - ../../configs/promtail.yml:/etc/promtail/config.yml:ro
    command: -config.file=/etc/promtail/config.yml
    networks:
      - monitoring
    depends_on:
      - loki

volumes:
  prometheus-data:
  grafana-data:
  alertmanager-data:
  jaeger-data:
  loki-data:

networks:
  monitoring:
    driver: bridge
EOF
    
    log_success "Monitoring configuration created"
}

# Create AlertManager configuration
create_alertmanager_config() {
    log_info "Creating AlertManager configuration..."
    
    cat > "$PROJECT_ROOT/configs/alertmanager.yml" << 'EOF'
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alerts@aios.dev'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://localhost:5001/'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
EOF
    
    log_success "AlertManager configuration created"
}

# Create Promtail configuration
create_promtail_config() {
    log_info "Creating Promtail configuration..."
    
    cat > "$PROJECT_ROOT/configs/promtail.yml" << 'EOF'
server:
  http_listen_port: 9080
  grpc_listen_port: 0

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: containers
    static_configs:
      - targets:
          - localhost
        labels:
          job: containerlogs
          __path__: /var/lib/docker/containers/*/*log

    pipeline_stages:
      - json:
          expressions:
            output: log
            stream: stream
            attrs:
      - json:
          expressions:
            tag:
          source: attrs
      - regex:
          expression: (?P<container_name>(?:[^|]*))\|
          source: tag
      - timestamp:
          format: RFC3339Nano
          source: time
      - labels:
          stream:
          container_name:
      - output:
          source: output

  - job_name: syslog
    static_configs:
      - targets:
          - localhost
        labels:
          job: syslog
          __path__: /var/log/syslog
EOF
    
    log_success "Promtail configuration created"
}

# Setup monitoring stack
setup_monitoring() {
    log_info "Setting up monitoring stack..."
    
    cd "$PROJECT_ROOT/deployments/monitoring"
    
    # Start monitoring services
    ENVIRONMENT="$ENVIRONMENT" docker-compose up -d
    
    # Wait for services to be ready
    log_info "Waiting for monitoring services to be ready..."
    sleep 30
    
    # Check if services are running
    local services=("prometheus" "grafana" "alertmanager" "jaeger")
    for service in "${services[@]}"; do
        if docker-compose ps "$service" | grep -q "Up"; then
            log_success "$service is running"
        else
            log_error "$service failed to start"
        fi
    done
    
    log_success "Monitoring stack setup completed"
}

# Show monitoring URLs
show_monitoring_urls() {
    log_info "Monitoring Services URLs:"
    echo "  Prometheus:   http://localhost:9090"
    echo "  Grafana:      http://localhost:3001 (admin/admin)"
    echo "  AlertManager: http://localhost:9093"
    echo "  Jaeger:       http://localhost:16686"
    echo "  Node Exporter: http://localhost:9100"
    echo "  cAdvisor:     http://localhost:8080"
}

# Create monitoring dashboards
create_dashboards() {
    log_info "Creating Grafana dashboards..."
    
    local dashboards_dir="$PROJECT_ROOT/configs/grafana/provisioning/dashboards"
    mkdir -p "$dashboards_dir"
    
    # Create dashboard provisioning config
    cat > "$dashboards_dir/dashboards.yml" << 'EOF'
apiVersion: 1

providers:
  - name: 'AIOS Dashboards'
    orgId: 1
    folder: 'AIOS'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF
    
    log_success "Dashboard configuration created"
}

# Main function
main() {
    local action=${1:-"setup"}
    
    log_info "AIOS Monitoring Setup"
    log_info "Environment: $ENVIRONMENT"
    log_info "Action: $action"
    
    case "$action" in
        "setup")
            check_dependencies
            create_monitoring_config
            create_alertmanager_config
            create_promtail_config
            create_dashboards
            setup_monitoring
            show_monitoring_urls
            ;;
        "start")
            cd "$PROJECT_ROOT/deployments/monitoring"
            ENVIRONMENT="$ENVIRONMENT" docker-compose up -d
            show_monitoring_urls
            ;;
        "stop")
            cd "$PROJECT_ROOT/deployments/monitoring"
            docker-compose down
            ;;
        "restart")
            cd "$PROJECT_ROOT/deployments/monitoring"
            docker-compose restart
            ;;
        "logs")
            cd "$PROJECT_ROOT/deployments/monitoring"
            docker-compose logs -f
            ;;
        "status")
            cd "$PROJECT_ROOT/deployments/monitoring"
            docker-compose ps
            ;;
        "clean")
            cd "$PROJECT_ROOT/deployments/monitoring"
            docker-compose down -v
            docker system prune -f
            ;;
        *)
            log_error "Unknown action: $action"
            echo "Usage: $0 [setup|start|stop|restart|logs|status|clean]"
            exit 1
            ;;
    esac
    
    log_success "Monitoring operation completed!"
}

# Run main function
main "$@"
