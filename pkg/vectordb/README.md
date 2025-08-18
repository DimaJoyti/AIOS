# Vector Database Integration

## Overview

The Vector Database Integration provides a comprehensive, production-ready system for storing, indexing, and querying high-dimensional vectors. This system is designed to support semantic search, recommendation systems, and AI-powered applications with enterprise-grade performance and scalability.

## Features

### 🚀 **Core Capabilities**
- **Multi-Provider Support**: Qdrant, Weaviate, Pinecone, and extensible architecture
- **Embedding Integration**: OpenAI, Ollama, HuggingFace, and custom providers
- **High-Level Vector Store**: Document-centric API with automatic embedding generation
- **Advanced Search**: Similarity search, MMR (Maximum Marginal Relevance), filtered search
- **Batch Operations**: Efficient bulk insert, update, delete, and search operations
- **Text Processing**: Automatic document chunking with configurable overlap
- **Metrics & Monitoring**: Comprehensive performance tracking and observability

### 🏗️ **Architecture**
- **Interface-Driven Design**: Clean abstractions for easy extension and testing
- **Factory Pattern**: Configuration-driven component creation
- **Builder Pattern**: Fluent interfaces for complex configurations
- **Observability**: Full OpenTelemetry integration with metrics and tracing
- **Error Handling**: Robust error handling with detailed error messages
- **Type Safety**: Strong typing throughout with comprehensive validation

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/aios/aios/pkg/vectordb"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    manager := vectordb.NewVectorDBManager(logger)
    
    // Configure vector database
    vectorDBConfig := vectordb.NewVectorDBBuilder().
        WithProvider("qdrant").
        WithHost("localhost").
        WithPort(6333).
        WithTimeout(30 * time.Second).
        Build()
    
    // Configure embedding provider
    embeddingConfig := vectordb.NewEmbeddingBuilder().
        WithProvider("openai").
        WithModel("text-embedding-ada-002").
        WithAPIKey("your-openai-api-key").
        WithDimensions(1536).
        Build()
    
    // Create vector store
    storeConfig := vectordb.NewVectorStoreBuilder().
        WithVectorDB(vectorDBConfig).
        WithEmbedding(embeddingConfig).
        WithChunkSize(1000).
        WithOverlap(200).
        Build()
    
    vectorStore, err := manager.CreateVectorStore(storeConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Add documents
    documents := []*vectordb.Document{
        {
            ID:      "doc1",
            Content: "Artificial Intelligence is transforming technology.",
            Metadata: map[string]interface{}{
                "category": "technology",
                "author":   "John Doe",
            },
        },
    }
    
    err = vectorStore.AddDocuments(ctx, "my_collection", documents)
    if err != nil {
        log.Fatal(err)
    }
    
    // Search
    results, err := vectorStore.SimilaritySearch(
        ctx, "my_collection", "AI technology", 5, nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    
    for _, doc := range results {
        log.Printf("Found: %s", doc.Content)
    }
}
```

## Supported Providers

### Vector Databases

#### Qdrant
- **Features**: High performance, real-time updates, advanced filtering
- **Configuration**:
```go
config := vectordb.NewVectorDBBuilder().
    WithProvider("qdrant").
    WithHost("localhost").
    WithPort(6333).
    WithAPIKey("optional-api-key").
    WithTLS(false).
    Build()
```

#### Weaviate (Coming Soon)
- **Features**: GraphQL API, automatic vectorization, hybrid search
- **Use Cases**: Knowledge graphs, semantic search, content management

#### Pinecone (Coming Soon)
- **Features**: Managed service, auto-scaling, high availability
- **Use Cases**: Production applications, large-scale deployments

### Embedding Providers

#### OpenAI
- **Models**: text-embedding-ada-002, text-embedding-3-small, text-embedding-3-large
- **Configuration**:
```go
config := vectordb.NewEmbeddingBuilder().
    WithProvider("openai").
    WithModel("text-embedding-ada-002").
    WithAPIKey("your-api-key").
    WithDimensions(1536).
    WithBatchSize(100).
    Build()
```

#### Ollama
- **Models**: nomic-embed-text, all-minilm, custom models
- **Configuration**:
```go
config := vectordb.NewEmbeddingBuilder().
    WithProvider("ollama").
    WithModel("nomic-embed-text").
    WithBaseURL("http://localhost:11434").
    WithDimensions(768).
    Build()
```

## Advanced Features

### Maximum Marginal Relevance (MMR) Search

MMR search provides diverse results by balancing relevance and diversity:

```go
results, err := vectorStore.MaxMarginalRelevanceSearch(
    ctx,
    "collection",
    "query text",
    topK,      // Number of results to return
    fetchK,    // Number of candidates to fetch
    lambda,    // Balance between relevance (1.0) and diversity (0.0)
    filter,    // Optional metadata filter
)
```

### Document Chunking

Automatically split large documents into manageable chunks:

```go
splitter := vectordb.NewTextSplitter(1000, 200) // chunk size, overlap
chunks := splitter.SplitDocuments(documents)

err := vectorStore.AddDocuments(ctx, "collection", chunks)
```

### Filtered Search

Search with metadata filters:

```go
filter := map[string]interface{}{
    "category": "technology",
    "date": map[string]interface{}{
        "gte": "2024-01-01",
    },
}

results, err := vectorStore.SimilaritySearch(
    ctx, "collection", "query", 10, filter,
)
```

### Batch Operations

Efficient bulk operations:

```go
// Batch add texts
texts := []string{"text1", "text2", "text3"}
metadatas := []map[string]interface{}{
    {"category": "A"},
    {"category": "B"},
    {"category": "C"},
}

err := vectorStore.AddTexts(ctx, "collection", texts, metadatas)

// Batch search
searchRequests := []*vectordb.SearchRequest{
    {Collection: "col1", Vector: embedding1, TopK: 5},
    {Collection: "col2", Vector: embedding2, TopK: 5},
}

results, err := vectorDB.BatchSearch(ctx, searchRequests)
```

## Configuration

### Vector Database Configuration

```go
config := &vectordb.VectorDBConfig{
    Provider:   "qdrant",
    Host:       "localhost",
    Port:       6333,
    APIKey:     "optional-key",
    Database:   "default",
    Timeout:    30 * time.Second,
    MaxRetries: 3,
    RetryDelay: 1 * time.Second,
    TLS:        false,
    Metadata: map[string]interface{}{
        "environment": "production",
        "log_level":   "info",
    },
}
```

### Embedding Configuration

```go
config := &vectordb.EmbeddingConfig{
    Provider:   "openai",
    Model:      "text-embedding-ada-002",
    APIKey:     "your-api-key",
    BaseURL:    "https://api.openai.com/v1",
    Dimensions: 1536,
    MaxTokens:  8191,
    BatchSize:  100,
    Timeout:    30 * time.Second,
    Metadata: map[string]interface{}{
        "version": "v1",
    },
}
```

### Vector Store Configuration

```go
config := &vectordb.VectorStoreConfig{
    VectorDB:  vectorDBConfig,
    Embedding: embeddingConfig,
    ChunkSize: 1000,
    Overlap:   200,
}
```

## Metrics and Monitoring

### Built-in Metrics

```go
metrics := vectordb.NewVectorDBMetrics(logger)

// Record operations
metrics.RecordOperation("search", duration, success)
metrics.RecordVectorCount("collection", count)
metrics.RecordSearchLatency("collection", latency)

// Get statistics
operationStats := metrics.GetOperationStats()
collectionStats := metrics.GetCollectionStats()
```

### OpenTelemetry Integration

All operations are automatically instrumented with OpenTelemetry:

- **Traces**: Request tracing across all operations
- **Metrics**: Operation counts, latencies, error rates
- **Attributes**: Collection names, vector counts, search parameters

## Error Handling

The system provides comprehensive error handling:

```go
// Connection errors
if err := vectorDB.Connect(ctx); err != nil {
    log.Printf("Connection failed: %v", err)
}

// Validation errors
if err := factory.ValidateConfig(config); err != nil {
    log.Printf("Invalid config: %v", err)
}

// Operation errors
if err := vectorStore.AddDocuments(ctx, collection, docs); err != nil {
    log.Printf("Failed to add documents: %v", err)
}
```

## Testing

### Unit Tests

```bash
go test ./pkg/vectordb/...
```

### Integration Tests

```bash
# Start Qdrant
docker run -p 6333:6333 qdrant/qdrant

# Run integration tests
go test -tags=integration ./pkg/vectordb/...
```

### Mock Testing

```go
// Use provided mocks for testing
mockVectorDB := vectordb.NewMockVectorDB()
mockEmbedding := vectordb.NewMockEmbeddingProvider(768, "mock-model")

vectorStore := vectordb.NewVectorStore(mockVectorDB, mockEmbedding, config, logger)
```

## Performance Considerations

### Optimization Tips

1. **Batch Operations**: Use batch operations for bulk data processing
2. **Chunk Size**: Optimize chunk size based on your content and use case
3. **Embedding Dimensions**: Balance between accuracy and performance
4. **Connection Pooling**: Reuse connections for better performance
5. **Indexing**: Use appropriate indexing strategies for your vector database

### Scaling

- **Horizontal Scaling**: Use multiple vector database instances
- **Sharding**: Distribute collections across multiple databases
- **Caching**: Implement caching for frequently accessed vectors
- **Load Balancing**: Distribute queries across multiple instances

## Best Practices

1. **Configuration Management**: Use environment variables for sensitive data
2. **Error Handling**: Implement proper error handling and retry logic
3. **Monitoring**: Set up comprehensive monitoring and alerting
4. **Security**: Use TLS and authentication in production
5. **Testing**: Write comprehensive tests including integration tests
6. **Documentation**: Document your vector schemas and search patterns

## Integration with AIOS

This vector database system integrates seamlessly with the AIOS ecosystem:

- **Langchain Integration**: Use with Langchain chains and agents
- **Memory Systems**: Store and retrieve conversation memory
- **Tool Integration**: Expose as tools through the MCP protocol
- **AI Orchestrator**: Use for semantic search in AI workflows
- **Observability**: Full integration with AIOS monitoring stack

## Examples

See `examples.go` for comprehensive usage examples including:

- Basic document storage and search
- Advanced search with MMR and filtering
- Batch operations and performance optimization
- Integration with different providers
- Metrics and monitoring setup

## Contributing

To add a new vector database provider:

1. Implement the `VectorDB` interface
2. Create a factory implementing `VectorDBFactory`
3. Register the factory with the manager
4. Add comprehensive tests
5. Update documentation

To add a new embedding provider:

1. Implement the `EmbeddingProvider` interface
2. Create a factory implementing `EmbeddingProviderFactory`
3. Register with the embedding manager
4. Add tests and documentation

## License

This vector database integration is part of the AIOS project and follows the same licensing terms.
