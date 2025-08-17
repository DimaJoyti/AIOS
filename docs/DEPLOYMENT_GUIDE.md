# AIOS Deployment Guide

This comprehensive guide covers all aspects of deploying AIOS (AI-Powered Operating System) across different environments and platforms.

## Overview

AIOS supports multiple deployment strategies:

- **Docker Containers**: Containerized deployment for development and production
- **Kubernetes**: Orchestrated deployment for scalable production environments
- **Bare Metal**: Direct installation on physical or virtual machines
- **Cloud Platforms**: AWS, GCP, Azure deployment configurations
- **Edge Devices**: Lightweight deployment for edge computing

## Quick Start

### Prerequisites

- **Docker**: Version 20.10 or higher
- **Docker Compose**: Version 2.0 or higher
- **Kubernetes**: Version 1.24 or higher (for K8s deployments)
- **Go**: Version 1.22 or higher (for source builds)
- **Git**: For version control

### Development Deployment

```bash
# Clone the repository
git clone <repository-url> aios
cd aios

# Start development environment
make dev

# Or using Docker Compose directly
docker-compose -f deployments/docker-compose.dev.yml up -d
```

### Production Deployment

```bash
# Build production images
make build-prod

# Deploy to production
docker-compose up -d

# Or deploy to Kubernetes
kubectl apply -f deployments/k8s/
```

## Docker Deployment

### Single Container Deployment

```bash
# Build the image
docker build -f deployments/Dockerfile.daemon -t aios:latest .

# Run the container
docker run -d \
  --name aios-daemon \
  -p 8080:8080 \
  -p 9090:9090 \
  -e AIOS_ENV=production \
  -e AIOS_LOG_LEVEL=info \
  -v $(pwd)/configs:/app/configs:ro \
  -v aios_data:/app/data \
  aios:latest
```

### Multi-Container Deployment

```bash
# Use Docker Compose for full stack
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f aios-daemon
```

### Docker Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  aios-daemon:
    build:
      context: .
      dockerfile: deployments/Dockerfile.daemon
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - AIOS_ENV=production
      - AIOS_LOG_LEVEL=info
      - AIOS_DB_HOST=postgres
      - AIOS_REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./configs:/app/configs:ro
      - aios_data:/app/data
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      - POSTGRES_DB=aios
      - POSTGRES_USER=aios
      - POSTGRES_PASSWORD=aios_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  aios_data:
  postgres_data:
  redis_data:
```

## Kubernetes Deployment

### Prerequisites

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Install Helm (optional)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Basic Kubernetes Deployment

```bash
# Create namespace
kubectl create namespace aios

# Apply configurations
kubectl apply -f deployments/k8s/configmap.yaml
kubectl apply -f deployments/k8s/secret.yaml
kubectl apply -f deployments/k8s/postgres.yaml
kubectl apply -f deployments/k8s/redis.yaml
kubectl apply -f deployments/k8s/aios-daemon.yaml

# Check deployment status
kubectl get pods -n aios
kubectl get services -n aios
```

### Helm Deployment

```bash
# Add AIOS Helm repository
helm repo add aios https://charts.aios.dev
helm repo update

# Install AIOS
helm install aios aios/aios \
  --namespace aios \
  --create-namespace \
  --values deployments/helm/values.yaml

# Upgrade deployment
helm upgrade aios aios/aios \
  --namespace aios \
  --values deployments/helm/values.yaml
```

### Kubernetes Manifests

```yaml
# deployments/k8s/aios-daemon.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aios-daemon
  namespace: aios
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aios-daemon
  template:
    metadata:
      labels:
        app: aios-daemon
    spec:
      containers:
      - name: aios-daemon
        image: aios/daemon:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: AIOS_ENV
          value: "production"
        - name: AIOS_DB_HOST
          value: "postgres"
        - name: AIOS_REDIS_HOST
          value: "redis"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: aios-daemon
  namespace: aios
spec:
  selector:
    app: aios-daemon
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: metrics
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

## Cloud Platform Deployment

### AWS Deployment

#### ECS Deployment

```bash
# Create ECS cluster
aws ecs create-cluster --cluster-name aios-cluster

# Register task definition
aws ecs register-task-definition \
  --cli-input-json file://deployments/aws/task-definition.json

# Create service
aws ecs create-service \
  --cluster aios-cluster \
  --service-name aios-daemon \
  --task-definition aios-daemon:1 \
  --desired-count 2
```

#### EKS Deployment

```bash
# Create EKS cluster
eksctl create cluster \
  --name aios-cluster \
  --region us-west-2 \
  --nodegroup-name aios-nodes \
  --node-type m5.large \
  --nodes 3

# Deploy AIOS
kubectl apply -f deployments/k8s/
```

### Google Cloud Platform

#### GKE Deployment

```bash
# Create GKE cluster
gcloud container clusters create aios-cluster \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n1-standard-2

# Get credentials
gcloud container clusters get-credentials aios-cluster \
  --zone us-central1-a

# Deploy AIOS
kubectl apply -f deployments/k8s/
```

### Azure Deployment

#### AKS Deployment

```bash
# Create resource group
az group create --name aios-rg --location eastus

# Create AKS cluster
az aks create \
  --resource-group aios-rg \
  --name aios-cluster \
  --node-count 3 \
  --node-vm-size Standard_D2s_v3

# Get credentials
az aks get-credentials \
  --resource-group aios-rg \
  --name aios-cluster

# Deploy AIOS
kubectl apply -f deployments/k8s/
```

## Configuration Management

### Environment Variables

```bash
# Core configuration
AIOS_ENV=production
AIOS_LOG_LEVEL=info
AIOS_CONFIG_PATH=/app/configs/prod.yaml

# Database configuration
AIOS_DB_HOST=postgres
AIOS_DB_PORT=5432
AIOS_DB_NAME=aios
AIOS_DB_USER=aios
AIOS_DB_PASSWORD=secure_password

# Redis configuration
AIOS_REDIS_HOST=redis
AIOS_REDIS_PORT=6379
AIOS_REDIS_PASSWORD=redis_password

# AI services configuration
AIOS_OLLAMA_HOST=ollama
AIOS_OLLAMA_PORT=11434

# Observability configuration
AIOS_JAEGER_ENDPOINT=http://jaeger:14268/api/traces
AIOS_PROMETHEUS_ENDPOINT=http://prometheus:9090
```

### Configuration Files

```yaml
# configs/prod.yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

database:
  host: "${AIOS_DB_HOST}"
  port: ${AIOS_DB_PORT}
  name: "${AIOS_DB_NAME}"
  user: "${AIOS_DB_USER}"
  password: "${AIOS_DB_PASSWORD}"
  ssl_mode: "require"
  max_connections: 25
  max_idle_connections: 5

redis:
  host: "${AIOS_REDIS_HOST}"
  port: ${AIOS_REDIS_PORT}
  password: "${AIOS_REDIS_PASSWORD}"
  db: 0
  pool_size: 10

ai:
  ollama:
    host: "${AIOS_OLLAMA_HOST}"
    port: ${AIOS_OLLAMA_PORT}
    timeout: "30s"
  
  models:
    path: "/app/models"
    default_model: "llama2"
    max_tokens: 2048

logging:
  level: "${AIOS_LOG_LEVEL}"
  format: "json"
  output: "stdout"

metrics:
  enabled: true
  port: 9090
  path: "/metrics"

tracing:
  enabled: true
  endpoint: "${AIOS_JAEGER_ENDPOINT}"
  sample_rate: 0.1
```

## Security Configuration

### TLS/SSL Setup

```bash
# Generate certificates
openssl req -x509 -newkey rsa:4096 \
  -keyout configs/tls/server.key \
  -out configs/tls/server.crt \
  -days 365 -nodes \
  -subj "/CN=aios.local"

# Create Kubernetes secret
kubectl create secret tls aios-tls \
  --cert=configs/tls/server.crt \
  --key=configs/tls/server.key \
  --namespace=aios
```

### Network Security

```yaml
# Network policies for Kubernetes
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aios-network-policy
  namespace: aios
spec:
  podSelector:
    matchLabels:
      app: aios-daemon
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - podSelector:
        matchLabels:
          app: redis
    ports:
    - protocol: TCP
      port: 6379
```

## Monitoring and Observability

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'aios-daemon'
    static_configs:
      - targets: ['aios-daemon:9090']
    metrics_path: /metrics
    scrape_interval: 10s

  - job_name: 'aios-assistant'
    static_configs:
      - targets: ['aios-assistant:9091']
    metrics_path: /metrics
    scrape_interval: 10s
```

### Grafana Dashboards

```bash
# Import AIOS dashboards
curl -X POST \
  http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @deployments/monitoring/grafana-dashboard.json
```

### Log Aggregation

```yaml
# Fluentd configuration
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
  namespace: aios
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/aios-*.log
      pos_file /var/log/fluentd-aios.log.pos
      tag aios.*
      format json
    </source>
    
    <match aios.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      index_name aios-logs
    </match>
```

## Scaling and Performance

### Horizontal Pod Autoscaler

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: aios-daemon-hpa
  namespace: aios
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: aios-daemon
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Vertical Pod Autoscaler

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: aios-daemon-vpa
  namespace: aios
spec:
  targetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: aios-daemon
  updatePolicy:
    updateMode: "Auto"
  resourcePolicy:
    containerPolicies:
    - containerName: aios-daemon
      maxAllowed:
        cpu: 2
        memory: 4Gi
      minAllowed:
        cpu: 100m
        memory: 256Mi
```

## Backup and Disaster Recovery

### Database Backup

```bash
# PostgreSQL backup
kubectl exec -n aios postgres-0 -- \
  pg_dump -U aios aios > backup-$(date +%Y%m%d).sql

# Restore from backup
kubectl exec -i -n aios postgres-0 -- \
  psql -U aios aios < backup-20240101.sql
```

### Volume Backup

```bash
# Create volume snapshot
kubectl apply -f - <<EOF
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: aios-data-snapshot
  namespace: aios
spec:
  volumeSnapshotClassName: csi-hostpath-snapclass
  source:
    persistentVolumeClaimName: aios-data-pvc
EOF
```

## Troubleshooting

### Common Issues

1. **Container Won't Start**
   ```bash
   # Check logs
   docker logs aios-daemon
   kubectl logs -n aios deployment/aios-daemon
   
   # Check configuration
   docker exec -it aios-daemon cat /app/configs/prod.yaml
   ```

2. **Database Connection Issues**
   ```bash
   # Test database connectivity
   kubectl exec -n aios deployment/aios-daemon -- \
     nc -zv postgres 5432
   
   # Check database logs
   kubectl logs -n aios deployment/postgres
   ```

3. **High Memory Usage**
   ```bash
   # Check resource usage
   kubectl top pods -n aios
   
   # Adjust resource limits
   kubectl patch deployment aios-daemon -n aios -p \
     '{"spec":{"template":{"spec":{"containers":[{"name":"aios-daemon","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
   ```

### Debug Commands

```bash
# Get deployment status
kubectl get deployments -n aios

# Describe pod issues
kubectl describe pod -n aios -l app=aios-daemon

# Check events
kubectl get events -n aios --sort-by=.metadata.creationTimestamp

# Port forward for debugging
kubectl port-forward -n aios service/aios-daemon 8080:80

# Execute shell in container
kubectl exec -it -n aios deployment/aios-daemon -- /bin/sh
```

## Performance Optimization

### Resource Tuning

```yaml
# Optimized resource configuration
resources:
  requests:
    memory: "1Gi"
    cpu: "500m"
  limits:
    memory: "2Gi"
    cpu: "1000m"
```

### JVM Tuning (if applicable)

```bash
# Java heap settings
JAVA_OPTS="-Xms1g -Xmx2g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"
```

### Database Optimization

```sql
-- PostgreSQL optimization
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
SELECT pg_reload_conf();
```

This deployment guide provides comprehensive coverage of AIOS deployment across various platforms and environments, ensuring successful deployment and operation in production scenarios.
