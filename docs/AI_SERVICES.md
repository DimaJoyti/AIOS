# AIOS AI Services Layer

The AI Services Layer is the core intelligence component of AIOS, providing advanced AI capabilities for natural language processing, computer vision, system optimization, and more.

## Overview

The AI Services Layer consists of several specialized services orchestrated by a central AI Orchestrator:

- **Language Model Service (LLM)**: Natural language processing and generation
- **Computer Vision Service (CV)**: Image analysis and UI understanding
- **System Optimization Service**: AI-powered performance optimization
- **Voice Service**: Speech recognition and synthesis (planned)
- **Natural Language Processing Service**: Intent parsing and entity extraction (planned)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AI Orchestrator                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │              Request Router                         │   │
│  │         Workflow Manager                            │   │
│  │         Result Aggregator                           │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
┌───────▼──────┐    ┌────────▼────────┐    ┌──────▼──────┐
│ LLM Service  │    │  CV Service     │    │ Optimization│
│              │    │                 │    │ Service     │
│ • Chat       │    │ • Screen        │    │             │
│ • Code Gen   │    │   Analysis      │    │ • Performance│
│ • Text       │    │ • OCR           │    │   Analysis  │
│   Analysis   │    │ • Object        │    │ • Prediction│
│ • Translation│    │   Detection     │    │ • Resource  │
│ • Summary    │    │ • UI Analysis   │    │   Optimization│
└──────────────┘    └─────────────────┘    └─────────────┘
```

## Services

### 1. Language Model Service (LLM)

Provides natural language processing capabilities using local language models (Ollama).

**Key Features:**
- Chat conversations with context
- Code generation and analysis
- Text analysis and summarization
- Language translation
- Intent recognition

**API Endpoints:**
```
POST /api/v1/ai/chat
POST /api/v1/llm/generate
POST /api/v1/llm/analyze
POST /api/v1/llm/translate
POST /api/v1/llm/summarize
```

**Example Usage:**
```bash
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Help me optimize my system performance",
    "conversation_id": "user-123"
  }'
```

### 2. Computer Vision Service (CV)

Analyzes images and provides UI understanding capabilities.

**Key Features:**
- Screen analysis and UI element detection
- Optical Character Recognition (OCR)
- Object detection and classification
- Layout analysis and suggestions
- Image comparison and description

**API Endpoints:**
```
POST /api/v1/ai/vision
POST /api/v1/cv/analyze-screen
POST /api/v1/cv/ocr
POST /api/v1/cv/detect-objects
POST /api/v1/cv/analyze-layout
```

**Example Usage:**
```bash
curl -X POST http://localhost:8080/api/v1/ai/vision \
  -F "image=@screenshot.png" \
  -F "task=analyze_screen"
```

### 3. System Optimization Service

Provides AI-powered system performance analysis and optimization.

**Key Features:**
- Performance analysis and bottleneck detection
- Resource usage prediction
- Optimization recommendations
- Health monitoring and failure prediction
- Workload optimization

**API Endpoints:**
```
POST /api/v1/ai/optimize
POST /api/v1/optimization/analyze
POST /api/v1/optimization/predict
POST /api/v1/optimization/recommend
POST /api/v1/optimization/health
```

**Example Usage:**
```bash
curl -X POST http://localhost:8080/api/v1/ai/optimize \
  -H "Content-Type: application/json" \
  -d '{
    "task": "analyze",
    "parameters": {
      "timeframe": "1h"
    }
  }'
```

## AI Orchestrator

The AI Orchestrator coordinates all AI services and manages complex workflows.

### Key Responsibilities:

1. **Request Routing**: Routes requests to appropriate AI services
2. **Workflow Management**: Executes multi-step AI workflows
3. **Result Aggregation**: Combines results from multiple services
4. **Service Health Monitoring**: Tracks service status and performance
5. **Load Balancing**: Distributes requests across service instances

### Workflow Example:

```json
{
  "id": "analyze-and-optimize",
  "name": "System Analysis and Optimization",
  "steps": [
    {
      "id": "analyze",
      "type": "optimization",
      "service": "optimization",
      "parameters": {
        "task": "analyze"
      }
    },
    {
      "id": "recommend",
      "type": "optimization",
      "service": "optimization",
      "parameters": {
        "task": "recommend"
      },
      "dependencies": ["analyze"]
    }
  ],
  "timeout": "5m"
}
```

## Configuration

AI services are configured through the `AIServiceConfig` structure:

```yaml
ai:
  ollama:
    host: "localhost"
    port: 11434
    timeout: "30s"
  
  models:
    path: "/opt/aios/models"
    default_model: "llama2"
    max_tokens: 2048
    temperature: 0.7
  
  computer_vision:
    enabled: true
    model_path: "/opt/aios/models/cv"
    confidence_threshold: 0.8
    max_image_size: "10MB"
  
  voice:
    enabled: false
    model_path: "/opt/aios/models/voice"
    wake_word: "aios"
    sample_rate: 16000
  
  performance:
    max_concurrent_requests: 10
    request_timeout: "30s"
    model_cache_size: 5
  
  security:
    enable_sandbox: true
    allowed_operations: ["read", "analyze"]
    data_retention: "24h"
```

## Model Management

### Supported Models:

1. **Language Models (LLM)**:
   - Llama 2 (7B, 13B, 70B)
   - Code Llama
   - Mistral
   - Custom fine-tuned models

2. **Computer Vision Models**:
   - YOLO for object detection
   - Tesseract for OCR
   - Custom UI analysis models

3. **Optimization Models**:
   - Performance prediction models
   - Resource optimization algorithms
   - Anomaly detection models

### Model Operations:

```bash
# List available models
curl http://localhost:8080/api/v1/ai/models

# Load a model
curl -X POST http://localhost:8080/api/v1/models/load \
  -d '{"model_id": "llama2"}'

# Get model status
curl http://localhost:8080/api/v1/models/llama2/status
```

## Performance and Monitoring

### Metrics:

- Request latency and throughput
- Model loading times and memory usage
- Service health and error rates
- Resource utilization

### Observability:

- **Prometheus metrics** at `/metrics`
- **Jaeger tracing** for request flows
- **Structured logging** with correlation IDs
- **Health checks** at `/health`

### Performance Optimization:

1. **Model Caching**: Keep frequently used models in memory
2. **Request Batching**: Batch similar requests for efficiency
3. **Load Balancing**: Distribute load across service instances
4. **Resource Limits**: Set appropriate CPU/memory limits
5. **Async Processing**: Use async patterns for long-running tasks

## Security and Privacy

### Security Features:

1. **Sandboxing**: AI operations run in isolated environments
2. **Input Validation**: All inputs are validated and sanitized
3. **Rate Limiting**: Prevent abuse and resource exhaustion
4. **Access Control**: Role-based access to AI services
5. **Audit Logging**: All AI operations are logged

### Privacy Protection:

1. **Local Processing**: All AI processing happens locally
2. **Data Retention**: Configurable data retention policies
3. **Encryption**: Data encrypted in transit and at rest
4. **Anonymization**: Personal data is anonymized when possible

## Development and Testing

### Running Tests:

```bash
# Run all AI service tests
go test ./internal/ai -v

# Run specific service tests
go test ./internal/ai -run TestLLMService

# Run benchmarks
go test ./internal/ai -bench=.
```

### Adding New Services:

1. Implement the service interface
2. Add service to orchestrator initialization
3. Update routing logic
4. Add API endpoints
5. Write comprehensive tests
6. Update documentation

## Troubleshooting

### Common Issues:

1. **Model Loading Failures**:
   - Check model file permissions
   - Verify sufficient memory
   - Check Ollama service status

2. **High Latency**:
   - Monitor resource usage
   - Check model cache hit rates
   - Optimize request batching

3. **Service Unavailable**:
   - Check service health endpoints
   - Verify network connectivity
   - Review service logs

### Debug Commands:

```bash
# Check AI service status
curl http://localhost:8080/api/v1/ai/status

# View service logs
docker logs aios-daemon

# Monitor resource usage
curl http://localhost:8080/api/v1/system/resources
```

## Future Enhancements

### Planned Features:

1. **Voice Services**: Speech-to-text and text-to-speech
2. **Multi-modal AI**: Combined vision and language models
3. **Federated Learning**: Collaborative model training
4. **Edge AI**: Optimized models for edge devices
5. **Custom Model Training**: Fine-tuning capabilities

### Roadmap:

- **Q1 2024**: Voice services implementation
- **Q2 2024**: Multi-modal AI capabilities
- **Q3 2024**: Advanced optimization algorithms
- **Q4 2024**: Federated learning framework

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines on contributing to the AI Services Layer.

## License

The AI Services Layer is part of AIOS and is licensed under the same terms as the main project.
