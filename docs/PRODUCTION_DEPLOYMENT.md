# AIOS Production Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying AIOS to production environments with enterprise-grade security, scalability, and reliability. It covers deployment strategies, infrastructure requirements, configuration management, monitoring, and operational best practices.

## 🏗️ Deployment Architecture

### Production Architecture

```
AIOS Production Deployment
├── Load Balancer (HAProxy/NGINX/AWS ALB)
├── API Gateway (Kong/Ambassador/AWS API Gateway)
├── AIOS Services Cluster
│   ├── Agent Service (3+ replicas)
│   ├── Collaboration Service (3+ replicas)
│   ├── Integration Service (2+ replicas)
│   ├── Data Integration Service (2+ replicas)
│   └── System Integration Hub (2+ replicas)
├── Data Layer
│   ├── Primary Database (PostgreSQL/MySQL)
│   ├── Cache Layer (Redis Cluster)
│   ├── Search Engine (Elasticsearch)
│   └── Message Queue (RabbitMQ/Apache Kafka)
├── Storage Layer
│   ├── Object Storage (AWS S3/MinIO)
│   ├── File Storage (NFS/EFS)
│   └── Backup Storage (AWS Glacier/Azure Archive)
├── Monitoring & Observability
│   ├── Metrics (Prometheus + Grafana)
│   ├── Logging (ELK Stack/Fluentd)
│   ├── Tracing (Jaeger/Zipkin)
│   └── Alerting (AlertManager/PagerDuty)
└── Security Layer
    ├── WAF (Web Application Firewall)
    ├── SSL/TLS Termination
    ├── Identity Provider (OAuth2/OIDC)
    └── Secrets Management (Vault/AWS Secrets Manager)
```

## 🚀 Deployment Strategies

### 1. Blue-Green Deployment

Zero-downtime deployment with instant rollback capability:

```yaml
# blue-green-deployment.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aios-deployment-config
data:
  deployment_strategy: "blue_green"
  health_check_path: "/health"
  readiness_timeout: "300s"
  traffic_switch_delay: "60s"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aios-blue
  labels:
    app: aios
    version: blue
spec:
  replicas: 3
  selector:
    matchLabels:
      app: aios
      version: blue
  template:
    metadata:
      labels:
        app: aios
        version: blue
    spec:
      containers:
      - name: aios
        image: aios:v1.0.0
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: aios-secrets
              key: database-url
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
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: aios-service
spec:
  selector:
    app: aios
    version: blue  # Switch to green during deployment
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
```

### 2. Canary Deployment

Gradual rollout with traffic splitting:

```yaml
# canary-deployment.yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: aios-rollout
spec:
  replicas: 5
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 2m}
      - setWeight: 20
      - pause: {duration: 2m}
      - setWeight: 50
      - pause: {duration: 2m}
      - setWeight: 100
      canaryService: aios-canary
      stableService: aios-stable
      trafficRouting:
        nginx:
          stableIngress: aios-stable
          annotationPrefix: nginx.ingress.kubernetes.io
          additionalIngressAnnotations:
            canary-by-header: X-Canary
      analysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: aios-canary
        startingStep: 2
        interval: 30s
  selector:
    matchLabels:
      app: aios
  template:
    metadata:
      labels:
        app: aios
    spec:
      containers:
      - name: aios
        image: aios:v1.1.0
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
```

### 3. Rolling Deployment

Standard Kubernetes rolling update:

```yaml
# rolling-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aios-deployment
spec:
  replicas: 6
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 2
  selector:
    matchLabels:
      app: aios
  template:
    metadata:
      labels:
        app: aios
    spec:
      containers:
      - name: aios
        image: aios:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        - name: LOG_LEVEL
          value: "info"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
```

## 🔧 Infrastructure Requirements

### Minimum Production Requirements

#### Compute Resources
- **CPU**: 8 cores minimum (16 cores recommended)
- **Memory**: 16GB minimum (32GB recommended)
- **Storage**: 100GB SSD minimum (500GB recommended)
- **Network**: 1Gbps minimum (10Gbps recommended)

#### Kubernetes Cluster
- **Nodes**: 3 minimum (5+ recommended for HA)
- **Node Size**: 4 CPU, 8GB RAM minimum per node
- **Kubernetes Version**: 1.24+ (latest stable recommended)
- **Container Runtime**: containerd or Docker
- **CNI**: Calico, Flannel, or cloud provider CNI

#### Database Requirements
- **PostgreSQL**: 13+ with replication
- **Redis**: 6+ with clustering
- **Elasticsearch**: 7+ with 3+ nodes
- **Storage**: High-performance SSD with IOPS 3000+

### Cloud Provider Configurations

#### AWS Deployment

```yaml
# aws-infrastructure.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-config
data:
  region: "us-west-2"
  availability_zones: "us-west-2a,us-west-2b,us-west-2c"
  instance_types: "m5.xlarge,m5.2xlarge"
  storage_class: "gp3"
  backup_retention: "30"
---
# EKS Cluster Configuration
apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: aios-production
  region: us-west-2
vpc:
  enableDnsHostnames: true
  enableDnsSupport: true
nodeGroups:
- name: aios-workers
  instanceType: m5.xlarge
  minSize: 3
  maxSize: 10
  desiredCapacity: 5
  volumeSize: 100
  volumeType: gp3
  ssh:
    enableSsm: true
  iam:
    withAddonPolicies:
      autoScaler: true
      cloudWatch: true
      ebs: true
      efs: true
      albIngress: true
addons:
- name: vpc-cni
- name: coredns
- name: kube-proxy
- name: aws-ebs-csi-driver
cloudWatch:
  clusterLogging:
    enableTypes: ["*"]
```

#### Azure Deployment

```yaml
# azure-infrastructure.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: azure-config
data:
  location: "East US 2"
  resource_group: "aios-production-rg"
  vm_size: "Standard_D4s_v3"
  storage_account_type: "Premium_LRS"
---
# AKS Cluster Configuration
resource "azurerm_kubernetes_cluster" "aios" {
  name                = "aios-production"
  location            = azurerm_resource_group.aios.location
  resource_group_name = azurerm_resource_group.aios.name
  dns_prefix          = "aios-prod"
  kubernetes_version  = "1.25.6"

  default_node_pool {
    name                = "default"
    node_count          = 3
    vm_size            = "Standard_D4s_v3"
    os_disk_size_gb    = 100
    os_disk_type       = "Managed"
    vnet_subnet_id     = azurerm_subnet.internal.id
    enable_auto_scaling = true
    min_count          = 3
    max_count          = 10
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
  }

  addon_profile {
    azure_policy {
      enabled = true
    }
    oms_agent {
      enabled                    = true
      log_analytics_workspace_id = azurerm_log_analytics_workspace.aios.id
    }
  }
}
```

#### Google Cloud Deployment

```yaml
# gcp-infrastructure.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gcp-config
data:
  project_id: "aios-production-project"
  region: "us-central1"
  zones: "us-central1-a,us-central1-b,us-central1-c"
  machine_type: "n1-standard-4"
  disk_type: "pd-ssd"
---
# GKE Cluster Configuration
resource "google_container_cluster" "aios" {
  name     = "aios-production"
  location = "us-central1"
  
  remove_default_node_pool = true
  initial_node_count       = 1
  
  network    = google_compute_network.vpc.name
  subnetwork = google_compute_subnetwork.subnet.name
  
  workload_identity_config {
    workload_pool = "${var.project_id}.svc.id.goog"
  }
  
  addons_config {
    horizontal_pod_autoscaling {
      disabled = false
    }
    network_policy_config {
      disabled = false
    }
  }
}

resource "google_container_node_pool" "aios_nodes" {
  name       = "aios-node-pool"
  location   = "us-central1"
  cluster    = google_container_cluster.aios.name
  node_count = 3
  
  autoscaling {
    min_node_count = 3
    max_node_count = 10
  }
  
  node_config {
    preemptible  = false
    machine_type = "n1-standard-4"
    disk_size_gb = 100
    disk_type    = "pd-ssd"
    
    oauth_scopes = [
      "https://www.googleapis.com/auth/logging.write",
      "https://www.googleapis.com/auth/monitoring",
    ]
  }
}
```

## 🔐 Security Configuration

### SSL/TLS Configuration

```yaml
# ssl-tls-config.yaml
apiVersion: v1
kind: Secret
metadata:
  name: aios-tls-secret
type: kubernetes.io/tls
data:
  tls.crt: LS0tLS1CRUdJTi... # Base64 encoded certificate
  tls.key: LS0tLS1CRUdJTi... # Base64 encoded private key
---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aios-ingress
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/ssl-protocols: "TLSv1.2 TLSv1.3"
    nginx.ingress.kubernetes.io/ssl-ciphers: "ECDHE-RSA-AES128-GCM-SHA256,ECDHE-RSA-AES256-GCM-SHA384"
spec:
  tls:
  - hosts:
    - api.aios.company.com
    - app.aios.company.com
    secretName: aios-tls-secret
  rules:
  - host: api.aios.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: aios-api-service
            port:
              number: 80
  - host: app.aios.company.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: aios-app-service
            port:
              number: 80
```

### Network Policies

```yaml
# network-policies.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: aios-network-policy
spec:
  podSelector:
    matchLabels:
      app: aios
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    - podSelector:
        matchLabels:
          app: aios-frontend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: database
    ports:
    - protocol: TCP
      port: 5432
  - to:
    - namespaceSelector:
        matchLabels:
          name: cache
    ports:
    - protocol: TCP
      port: 6379
  - to: []
    ports:
    - protocol: TCP
      port: 443
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

### RBAC Configuration

```yaml
# rbac-config.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aios-service-account
  namespace: aios-production
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  namespace: aios-production
  name: aios-role
rules:
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
- apiGroups: ["apps"]
  resources: ["deployments", "replicasets"]
  verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: aios-role-binding
  namespace: aios-production
subjects:
- kind: ServiceAccount
  name: aios-service-account
  namespace: aios-production
roleRef:
  kind: Role
  name: aios-role
  apiGroup: rbac.authorization.k8s.io
```

## 📊 Monitoring and Observability

### Prometheus Configuration

```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
      evaluation_interval: 15s
    
    rule_files:
      - "aios_rules.yml"
    
    alerting:
      alertmanagers:
        - static_configs:
            - targets:
              - alertmanager:9093
    
    scrape_configs:
      - job_name: 'aios-services'
        kubernetes_sd_configs:
          - role: pod
        relabel_configs:
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
            action: keep
            regex: true
          - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
            action: replace
            target_label: __metrics_path__
            regex: (.+)
          - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
            action: replace
            regex: ([^:]+)(?::\d+)?;(\d+)
            replacement: $1:$2
            target_label: __address__
          - action: labelmap
            regex: __meta_kubernetes_pod_label_(.+)
          - source_labels: [__meta_kubernetes_namespace]
            action: replace
            target_label: kubernetes_namespace
          - source_labels: [__meta_kubernetes_pod_name]
            action: replace
            target_label: kubernetes_pod_name
      
      - job_name: 'kubernetes-nodes'
        kubernetes_sd_configs:
          - role: node
        relabel_configs:
          - action: labelmap
            regex: __meta_kubernetes_node_label_(.+)
          - target_label: __address__
            replacement: kubernetes.default.svc:443
          - source_labels: [__meta_kubernetes_node_name]
            regex: (.+)
            target_label: __metrics_path__
            replacement: /api/v1/nodes/${1}/proxy/metrics
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: aios-alert-rules
data:
  aios_rules.yml: |
    groups:
    - name: aios.rules
      rules:
      - alert: AIOSServiceDown
        expr: up{job="aios-services"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "AIOS service is down"
          description: "AIOS service {{ $labels.instance }} has been down for more than 1 minute."
      
      - alert: AIOSHighErrorRate
        expr: rate(aios_http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High error rate in AIOS"
          description: "AIOS error rate is {{ $value }} errors per second."
      
      - alert: AIOSHighLatency
        expr: histogram_quantile(0.95, rate(aios_http_request_duration_seconds_bucket[5m])) > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency in AIOS"
          description: "AIOS 95th percentile latency is {{ $value }} seconds."
      
      - alert: AIOSHighMemoryUsage
        expr: (container_memory_usage_bytes{pod=~"aios-.*"} / container_spec_memory_limit_bytes) > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage in AIOS"
          description: "AIOS pod {{ $labels.pod }} memory usage is {{ $value | humanizePercentage }}."
```

### Grafana Dashboards

```json
{
  "dashboard": {
    "id": null,
    "title": "AIOS Production Dashboard",
    "tags": ["aios", "production"],
    "timezone": "browser",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(aios_http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ],
        "yAxes": [
          {
            "label": "Requests/sec"
          }
        ]
      },
      {
        "id": 2,
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, rate(aios_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "50th percentile"
          },
          {
            "expr": "histogram_quantile(0.95, rate(aios_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.99, rate(aios_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "99th percentile"
          }
        ],
        "yAxes": [
          {
            "label": "Seconds"
          }
        ]
      },
      {
        "id": 3,
        "title": "Error Rate",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(aios_http_requests_total{status=~\"5..\"}[5m]) / rate(aios_http_requests_total[5m])",
            "legendFormat": "Error Rate"
          }
        ],
        "valueName": "current",
        "format": "percentunit",
        "thresholds": "0.01,0.05"
      },
      {
        "id": 4,
        "title": "Active Agents",
        "type": "singlestat",
        "targets": [
          {
            "expr": "aios_active_agents_total",
            "legendFormat": "Active Agents"
          }
        ],
        "valueName": "current"
      },
      {
        "id": 5,
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "container_memory_usage_bytes{pod=~\"aios-.*\"} / 1024 / 1024",
            "legendFormat": "{{pod}}"
          }
        ],
        "yAxes": [
          {
            "label": "MB"
          }
        ]
      },
      {
        "id": 6,
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(container_cpu_usage_seconds_total{pod=~\"aios-.*\"}[5m]) * 100",
            "legendFormat": "{{pod}}"
          }
        ],
        "yAxes": [
          {
            "label": "Percent"
          }
        ]
      }
    ],
    "time": {
      "from": "now-1h",
      "to": "now"
    },
    "refresh": "30s"
  }
}
```

## 🔄 CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/production-deploy.yml
name: Production Deployment

on:
  push:
    branches: [main]
    tags: ['v*']
  pull_request:
    branches: [main]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
    
    - name: Upload coverage reports
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run security scan
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: 'security-scan-results.sarif'

  build:
    needs: test
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
    - uses: actions/checkout@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=semver,pattern={{version}}
          type=semver,pattern={{major}}.{{minor}}
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: staging
    steps:
    - uses: actions/checkout@v3
    
    - name: Configure kubectl
      uses: azure/k8s-set-context@v1
      with:
        method: kubeconfig
        kubeconfig: ${{ secrets.KUBE_CONFIG_STAGING }}
    
    - name: Deploy to staging
      run: |
        kubectl set image deployment/aios-deployment aios=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main
        kubectl rollout status deployment/aios-deployment --timeout=300s
    
    - name: Run integration tests
      run: |
        kubectl run integration-tests --image=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main \
          --rm -i --restart=Never -- go test -tags=integration ./tests/integration/...

  deploy-production:
    needs: [build, deploy-staging]
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    environment: production
    steps:
    - uses: actions/checkout@v3
    
    - name: Configure kubectl
      uses: azure/k8s-set-context@v1
      with:
        method: kubeconfig
        kubeconfig: ${{ secrets.KUBE_CONFIG_PRODUCTION }}
    
    - name: Deploy to production
      run: |
        # Blue-Green Deployment
        kubectl patch service aios-service -p '{"spec":{"selector":{"version":"green"}}}'
        kubectl set image deployment/aios-blue aios=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.ref_name }}
        kubectl rollout status deployment/aios-blue --timeout=600s
        
        # Health check
        kubectl wait --for=condition=ready pod -l app=aios,version=blue --timeout=300s
        
        # Switch traffic
        kubectl patch service aios-service -p '{"spec":{"selector":{"version":"blue"}}}'
        
        # Cleanup old version
        sleep 60
        kubectl delete deployment aios-green || true
    
    - name: Notify deployment
      uses: 8398a7/action-slack@v3
      with:
        status: ${{ job.status }}
        channel: '#deployments'
        webhook_url: ${{ secrets.SLACK_WEBHOOK }}
        message: |
          AIOS ${{ github.ref_name }} deployed to production
          Commit: ${{ github.sha }}
          Author: ${{ github.actor }}
```

## 🧪 Testing

Run comprehensive tests before deployment:

```bash
# Unit tests
go test -v -race -coverprofile=coverage.out ./...

# Integration tests
go test -tags=integration -v ./tests/integration/...

# Performance tests
go test -tags=performance -v ./tests/performance/...

# Security tests
go test -tags=security -v ./tests/security/...

# End-to-end tests
go test -tags=e2e -v ./tests/e2e/...
```

## 📖 Operational Procedures

### Health Checks

```bash
# Check service health
curl -f http://api.aios.company.com/health

# Check readiness
curl -f http://api.aios.company.com/ready

# Check metrics
curl http://api.aios.company.com/metrics
```

### Backup Procedures

```bash
# Database backup
kubectl exec -it postgres-0 -- pg_dump -U postgres aios > backup-$(date +%Y%m%d).sql

# Configuration backup
kubectl get configmaps,secrets -o yaml > config-backup-$(date +%Y%m%d).yaml

# Persistent volume backup
kubectl get pv,pvc -o yaml > storage-backup-$(date +%Y%m%d).yaml
```

### Rollback Procedures

```bash
# Rollback deployment
kubectl rollout undo deployment/aios-deployment

# Rollback to specific revision
kubectl rollout undo deployment/aios-deployment --to-revision=2

# Check rollout status
kubectl rollout status deployment/aios-deployment
```

## 🤝 Contributing

1. Follow deployment best practices and security guidelines
2. Test all changes in staging environment first
3. Update documentation for infrastructure changes
4. Ensure proper monitoring and alerting
5. Follow change management procedures
6. Maintain deployment automation and scripts

## 📄 License

This Production Deployment Guide is part of the AIOS project and follows the same licensing terms.
