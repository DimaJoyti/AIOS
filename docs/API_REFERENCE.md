# AIOS API Reference

This document provides comprehensive API reference for AIOS (AI-Powered Operating System) services.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

AIOS uses JWT-based authentication for API access.

```bash
# Login to get JWT token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'

# Use token in subsequent requests
curl -H "Authorization: Bearer <jwt-token>" \
  http://localhost:8080/api/v1/system/status
```

## System APIs

### Get System Status

Get overall system status and health information.

**Endpoint:** `GET /system/status`

**Response:**
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": "2h30m15s",
  "components": {
    "ai_services": "healthy",
    "database": "healthy",
    "cache": "healthy"
  },
  "resources": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 23.1
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get System Resources

Get detailed system resource information.

**Endpoint:** `GET /system/resources`

**Response:**
```json
{
  "cpu": {
    "cores": 8,
    "usage_percent": 45.2,
    "load_average": [1.2, 1.5, 1.8]
  },
  "memory": {
    "total_gb": 16,
    "used_gb": 10.8,
    "available_gb": 5.2,
    "usage_percent": 67.8
  },
  "disk": {
    "total_gb": 500,
    "used_gb": 115.5,
    "available_gb": 384.5,
    "usage_percent": 23.1
  },
  "network": {
    "interfaces": [
      {
        "name": "eth0",
        "bytes_sent": 1048576,
        "bytes_received": 2097152
      }
    ]
  }
}
```

### Get System Processes

Get information about running system processes.

**Endpoint:** `GET /system/processes`

**Query Parameters:**
- `limit` (optional): Number of processes to return (default: 50)
- `sort` (optional): Sort by cpu, memory, or name (default: cpu)

**Response:**
```json
{
  "processes": [
    {
      "pid": 1234,
      "name": "aios-daemon",
      "cpu_percent": 15.2,
      "memory_percent": 8.5,
      "status": "running",
      "started_at": "2024-01-15T08:00:00Z"
    }
  ],
  "total_processes": 156,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## AI Services APIs

### Chat with AI Assistant

Send a message to the AI assistant and get a response.

**Endpoint:** `POST /ai/chat`

**Request:**
```json
{
  "message": "Help me optimize my system performance",
  "conversation_id": "user-123",
  "context": {
    "system_info": true,
    "previous_messages": 5
  }
}
```

**Response:**
```json
{
  "response": "I can help you optimize your system performance. Based on your current resource usage, I recommend...",
  "conversation_id": "user-123",
  "message_id": "msg-456",
  "suggestions": [
    "Check running processes",
    "Analyze disk usage",
    "Review memory allocation"
  ],
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Generate Text

Generate text using the language model.

**Endpoint:** `POST /ai/llm/generate`

**Request:**
```json
{
  "prompt": "Write a Python function to calculate fibonacci numbers",
  "model": "llama2",
  "max_tokens": 500,
  "temperature": 0.7,
  "stop_sequences": ["```"]
}
```

**Response:**
```json
{
  "generated_text": "def fibonacci(n):\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)",
  "model": "llama2",
  "tokens_used": 45,
  "generation_time": "2.3s",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Analyze Image

Analyze an image using computer vision.

**Endpoint:** `POST /ai/cv/analyze`

**Request:**
```json
{
  "image_data": "base64_encoded_image_data",
  "analysis_type": "object_detection",
  "confidence_threshold": 0.8
}
```

**Response:**
```json
{
  "objects": [
    {
      "label": "person",
      "confidence": 0.95,
      "bounding_box": {
        "x": 100,
        "y": 150,
        "width": 200,
        "height": 300
      }
    }
  ],
  "analysis_time": "1.2s",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Get AI Models

List available AI models.

**Endpoint:** `GET /ai/models`

**Response:**
```json
{
  "models": [
    {
      "id": "llama2",
      "name": "Llama 2 7B",
      "type": "language_model",
      "size": "7B",
      "status": "loaded",
      "capabilities": ["text_generation", "chat", "code_generation"]
    },
    {
      "id": "yolo-v8",
      "name": "YOLO v8",
      "type": "computer_vision",
      "status": "available",
      "capabilities": ["object_detection", "image_classification"]
    }
  ]
}
```

## Desktop Environment APIs

### Get Desktop Status

Get desktop environment status and information.

**Endpoint:** `GET /desktop/status`

**Response:**
```json
{
  "status": "running",
  "session_id": "session-123",
  "user": "admin",
  "display": ":0",
  "resolution": "1920x1080",
  "windows": 5,
  "workspaces": 4,
  "current_workspace": 1,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### List Windows

Get list of open windows.

**Endpoint:** `GET /desktop/windows`

**Response:**
```json
{
  "windows": [
    {
      "id": "window-1",
      "title": "Terminal",
      "application": "gnome-terminal",
      "workspace": 1,
      "position": {
        "x": 100,
        "y": 100,
        "width": 800,
        "height": 600
      },
      "state": "normal",
      "focused": true
    }
  ]
}
```

### Control Window

Control window operations.

**Endpoint:** `POST /desktop/windows/{window_id}/action`

**Request:**
```json
{
  "action": "minimize",
  "parameters": {}
}
```

**Actions:**
- `minimize`: Minimize window
- `maximize`: Maximize window
- `close`: Close window
- `focus`: Focus window
- `move`: Move window (requires x, y parameters)
- `resize`: Resize window (requires width, height parameters)

## File System APIs

### List Directory

List contents of a directory.

**Endpoint:** `GET /filesystem/list`

**Query Parameters:**
- `path` (required): Directory path
- `show_hidden` (optional): Include hidden files (default: false)
- `sort` (optional): Sort by name, size, or modified (default: name)

**Response:**
```json
{
  "path": "/home/user",
  "items": [
    {
      "name": "Documents",
      "type": "directory",
      "size": 4096,
      "permissions": "drwxr-xr-x",
      "owner": "user",
      "group": "user",
      "modified": "2024-01-15T09:00:00Z"
    },
    {
      "name": "file.txt",
      "type": "file",
      "size": 1024,
      "permissions": "-rw-r--r--",
      "owner": "user",
      "group": "user",
      "modified": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### Search Files

Search for files and directories.

**Endpoint:** `GET /filesystem/search`

**Query Parameters:**
- `query` (required): Search query
- `path` (optional): Search path (default: /)
- `type` (optional): file, directory, or both (default: both)
- `limit` (optional): Maximum results (default: 100)

**Response:**
```json
{
  "query": "*.txt",
  "results": [
    {
      "path": "/home/user/document.txt",
      "type": "file",
      "size": 2048,
      "modified": "2024-01-15T10:00:00Z"
    }
  ],
  "total_results": 1,
  "search_time": "0.5s"
}
```

### File Operations

Perform file operations.

**Endpoint:** `POST /filesystem/operations`

**Request:**
```json
{
  "operation": "copy",
  "source": "/home/user/source.txt",
  "destination": "/home/user/backup.txt",
  "options": {
    "overwrite": false,
    "preserve_permissions": true
  }
}
```

**Operations:**
- `copy`: Copy file or directory
- `move`: Move file or directory
- `delete`: Delete file or directory
- `create_directory`: Create directory
- `create_file`: Create empty file

## Security APIs

### Get Security Status

Get overall security status.

**Endpoint:** `GET /security/status`

**Response:**
```json
{
  "status": "secure",
  "authentication": {
    "enabled": true,
    "method": "jwt",
    "session_timeout": "24h"
  },
  "encryption": {
    "enabled": true,
    "algorithm": "AES-256-GCM"
  },
  "firewall": {
    "enabled": true,
    "rules": 25,
    "blocked_attempts": 12
  },
  "last_scan": "2024-01-15T09:00:00Z",
  "threats_detected": 0
}
```

### Scan for Threats

Perform security threat scan.

**Endpoint:** `POST /security/scan`

**Request:**
```json
{
  "scan_type": "full",
  "targets": ["/home", "/var", "/tmp"],
  "options": {
    "deep_scan": true,
    "check_permissions": true
  }
}
```

**Response:**
```json
{
  "scan_id": "scan-123",
  "status": "completed",
  "start_time": "2024-01-15T10:00:00Z",
  "end_time": "2024-01-15T10:15:00Z",
  "duration": "15m",
  "files_scanned": 50000,
  "threats_found": 0,
  "warnings": 2,
  "results": []
}
```

## Testing APIs

### Get Testing Status

Get testing framework status.

**Endpoint:** `GET /testing/status`

**Response:**
```json
{
  "enabled": true,
  "running": false,
  "unit": {
    "enabled": true,
    "tests_run": 150,
    "tests_passed": 148,
    "tests_failed": 2,
    "coverage": 85.5
  },
  "integration": {
    "enabled": true,
    "tests_run": 25,
    "tests_passed": 24,
    "tests_failed": 1
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Run Tests

Execute test suites.

**Endpoint:** `POST /testing/run`

**Request:**
```json
{
  "type": "unit",
  "options": {
    "parallel": true,
    "verbose": false,
    "coverage": true
  }
}
```

**Response:**
```json
{
  "test_id": "test-123",
  "type": "unit",
  "status": "running",
  "started_at": "2024-01-15T10:30:00Z"
}
```

## Deployment APIs

### Get Deployment Status

Get deployment status and information.

**Endpoint:** `GET /deployment/status`

**Response:**
```json
{
  "enabled": true,
  "running": true,
  "environment": "production",
  "platform": "kubernetes",
  "container": {
    "enabled": true,
    "running": true,
    "containers": 5
  },
  "kubernetes": {
    "enabled": true,
    "connected": true,
    "namespace": "aios",
    "pods": 3
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Deploy Application

Deploy application to target environment.

**Endpoint:** `POST /deployment/deploy`

**Request:**
```json
{
  "version": "1.2.0",
  "environment": "production",
  "platform": "kubernetes",
  "config": {
    "replicas": 3,
    "resources": {
      "cpu": "500m",
      "memory": "1Gi"
    }
  },
  "dry_run": false
}
```

**Response:**
```json
{
  "deployment_id": "deploy-123",
  "version": "1.2.0",
  "environment": "production",
  "platform": "kubernetes",
  "status": "running",
  "started_at": "2024-01-15T10:30:00Z"
}
```

## Error Responses

All API endpoints return consistent error responses:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request parameters",
    "details": {
      "field": "username",
      "reason": "required field missing"
    }
  },
  "timestamp": "2024-01-15T10:30:00Z",
  "request_id": "req-123"
}
```

### Error Codes

- `INVALID_REQUEST`: Invalid request parameters
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `CONFLICT`: Resource conflict
- `INTERNAL_ERROR`: Internal server error
- `SERVICE_UNAVAILABLE`: Service temporarily unavailable

## Rate Limiting

API requests are rate limited:

- **Default**: 100 requests per minute per IP
- **Authenticated**: 1000 requests per minute per user
- **Headers**: Rate limit information in response headers

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## Webhooks

AIOS supports webhooks for real-time notifications:

### Register Webhook

**Endpoint:** `POST /webhooks`

**Request:**
```json
{
  "url": "https://your-app.com/webhook",
  "events": ["deployment.completed", "security.threat_detected"],
  "secret": "webhook_secret"
}
```

### Webhook Events

- `system.status_changed`
- `deployment.started`
- `deployment.completed`
- `deployment.failed`
- `security.threat_detected`
- `ai.model_loaded`
- `test.suite_completed`

## SDKs and Libraries

Official SDKs are available for:

- **Go**: `go get github.com/aios/aios-go-sdk`
- **Python**: `pip install aios-python-sdk`
- **JavaScript**: `npm install aios-js-sdk`
- **Rust**: `cargo add aios-rust-sdk`

### Example Usage (Go)

```go
import "github.com/aios/aios-go-sdk"

client := aios.NewClient("http://localhost:8080", "your-jwt-token")

status, err := client.System.GetStatus()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("System status: %s\n", status.Status)
```

## OpenAPI Specification

The complete OpenAPI 3.0 specification is available at:

```
GET /api/v1/openapi.json
GET /api/v1/docs
```

This provides interactive API documentation and testing capabilities.
