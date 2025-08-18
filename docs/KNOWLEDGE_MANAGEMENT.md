# AIOS Knowledge Management and RAG System

## Overview

The AIOS Knowledge Management and RAG (Retrieval-Augmented Generation) System provides a comprehensive platform for managing, processing, and querying knowledge bases with advanced AI capabilities. The system integrates document processing, semantic search, knowledge graphs, and RAG pipelines to deliver intelligent information retrieval and generation.

## 🏗️ Architecture

### Core Components

```
Knowledge Management System
├── Knowledge Manager (Orchestrator)
├── Document Processing Pipeline
├── RAG Pipeline
├── Knowledge Graph
├── Embedding Management
├── Semantic Search & Indexing
├── Caching Layer
└── Monitoring & Analytics
```

### Key Features

- **🔍 Advanced Search**: Semantic, keyword, and hybrid search capabilities
- **🧠 RAG Pipeline**: Complete retrieval-augmented generation workflow
- **📊 Knowledge Graph**: Entity-relationship modeling and graph queries
- **📝 Document Processing**: Multi-format document processing and chunking
- **⚡ Performance**: Semantic caching and optimized indexing
- **🔒 Enterprise**: Security, versioning, and collaboration features
- **📈 Observability**: Comprehensive monitoring and analytics

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/aios/aios/pkg/knowledge"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create knowledge manager
    config := &knowledge.KnowledgeManagerConfig{
        DefaultEmbeddingModel: "text-embedding-ada-002",
        DefaultChunkSize:      1000,
        DefaultChunkOverlap:   200,
        CacheEnabled:          true,
        GraphEnabled:          true,
    }
    
    manager, err := knowledge.NewKnowledgeManager(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Create knowledge base
    kbConfig := &knowledge.KnowledgeBaseConfig{
        EmbeddingModel:   "text-embedding-ada-002",
        ChunkingStrategy: knowledge.ChunkingStrategyRecursive,
        ChunkSize:        1000,
        IndexingEnabled:  true,
        GraphEnabled:     true,
    }
    
    kb, err := manager.CreateKnowledgeBase(context.Background(), kbConfig)
    if err != nil {
        panic(err)
    }
    
    // Add documents
    doc := &knowledge.Document{
        ID:              "doc-1",
        Title:           "Introduction to AI",
        Content:         "Artificial Intelligence is...",
        ContentType:     "text",
        KnowledgeBaseID: kb.ID,
    }
    
    err = manager.AddDocument(context.Background(), doc)
    if err != nil {
        panic(err)
    }
    
    // Search documents
    searchResult, err := manager.Search(context.Background(), &knowledge.SearchQuery{
        Query: "What is artificial intelligence?",
        Options: &knowledge.SearchOptions{
            TopK:      5,
            Threshold: 0.7,
        },
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d documents\n", len(searchResult.Documents))
    
    // RAG query
    ragResponse, err := manager.RetrieveAndGenerate(context.Background(), 
        "Explain artificial intelligence", &knowledge.RAGOptions{
            ContextLength:  2000,
            IncludeSources: true,
        })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Response: %s\n", ragResponse.Response)
}
```

## 📚 Core Concepts

### Documents

Documents are the fundamental units of knowledge in the system:

```go
type Document struct {
    ID              string                 `json:"id"`
    Title           string                 `json:"title"`
    Content         string                 `json:"content"`
    ContentType     string                 `json:"content_type"`
    Language        string                 `json:"language"`
    Source          string                 `json:"source"`
    Tags            []string               `json:"tags"`
    Categories      []string               `json:"categories"`
    Metadata        map[string]interface{} `json:"metadata"`
    Embedding       []float32              `json:"embedding"`
    KnowledgeBaseID string                 `json:"knowledge_base_id"`
    CreatedAt       time.Time              `json:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at"`
}
```

### Knowledge Bases

Knowledge bases are containers for related documents:

```go
type KnowledgeBase struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Config      *KnowledgeBaseConfig   `json:"config"`
    Stats       *KnowledgeBaseStats    `json:"stats"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}
```

### Chunking Strategies

The system supports multiple document chunking strategies:

- **Fixed**: Fixed-size chunks with configurable overlap
- **Sentence**: Sentence-boundary aware chunking
- **Paragraph**: Paragraph-based chunking
- **Recursive**: Hierarchical chunking with multiple separators
- **Semantic**: Semantic boundary detection (advanced)

### Knowledge Graph

Entities and relationships form a knowledge graph:

```go
type Entity struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Type        string                 `json:"type"`
    Description string                 `json:"description"`
    Properties  map[string]interface{} `json:"properties"`
    Confidence  float32                `json:"confidence"`
}

type Relationship struct {
    ID         string  `json:"id"`
    FromEntity string  `json:"from_entity"`
    ToEntity   string  `json:"to_entity"`
    Type       string  `json:"type"`
    Confidence float32 `json:"confidence"`
}
```

## 🔍 Search Capabilities

### Semantic Search

```go
searchOptions := &knowledge.SearchOptions{
    TopK:       10,
    Threshold:  0.7,
    SearchType: knowledge.SearchTypeSemantic,
}

result, err := manager.Search(ctx, &knowledge.SearchQuery{
    Query:   "machine learning algorithms",
    Options: searchOptions,
})
```

### Hybrid Search

```go
searchOptions := &knowledge.SearchOptions{
    TopK:       10,
    SearchType: knowledge.SearchTypeHybrid,
}

result, err := manager.Search(ctx, &knowledge.SearchQuery{
    Query:   "neural networks deep learning",
    Options: searchOptions,
})
```

### Graph Search

```go
graphQuery := &knowledge.GraphQuery{
    Type:     "find_path",
    FromEntity: "machine-learning",
    ToEntity:   "neural-networks",
    MaxDepth:   3,
}

result, err := graph.QueryGraph(ctx, graphQuery)
```

## 🧠 RAG Pipeline

### Basic RAG

```go
ragOptions := &knowledge.RAGOptions{
    RetrievalOptions: &knowledge.RetrievalOptions{
        TopK:      5,
        Threshold: 0.7,
    },
    GenerationOptions: &knowledge.GenerationOptions{
        Model:       "gpt-4",
        Temperature: 0.7,
        MaxTokens:   500,
    },
    ContextLength:  2000,
    IncludeSources: true,
}

response, err := manager.RetrieveAndGenerate(ctx, 
    "Explain the differences between supervised and unsupervised learning", 
    ragOptions)
```

### Advanced RAG with Citations

```go
ragOptions := &knowledge.RAGOptions{
    RetrievalOptions: &knowledge.RetrievalOptions{
        TopK:             5,
        RerankingEnabled: true,
        HybridSearch:     true,
    },
    GenerationOptions: &knowledge.GenerationOptions{
        Model:        "gpt-4",
        Temperature:  0.7,
        SystemPrompt: "You are an expert AI researcher. Provide detailed, accurate answers with citations.",
    },
    ContextLength:  4000,
    IncludeSources: true,
}

response, err := manager.RetrieveAndGenerate(ctx, query, ragOptions)

// Access citations
for _, citation := range response.Citations {
    fmt.Printf("Source: %s\n", citation.Text)
}
```

## 📊 Knowledge Graph Operations

### Adding Entities and Relationships

```go
// Add entities
entity := &knowledge.Entity{
    ID:          "ai",
    Name:        "Artificial Intelligence",
    Type:        "concept",
    Description: "Branch of computer science",
    Confidence:  0.95,
}

err := graph.AddEntity(ctx, entity)

// Add relationships
relationship := &knowledge.Relationship{
    ID:         "rel-1",
    FromEntity: "machine-learning",
    ToEntity:   "ai",
    Type:       "subset_of",
    Confidence: 0.9,
}

err = graph.AddRelationship(ctx, relationship)
```

### Graph Queries

```go
// Find paths between entities
paths, err := graph.FindPath(ctx, "deep-learning", "ai", 3)

// Get neighbors
neighbors, err := graph.GetNeighbors(ctx, "machine-learning", 2)

// Calculate centrality
centrality, err := graph.CalculateCentrality(ctx, "ai")
```

## ⚡ Performance Optimization

### Semantic Caching

```go
cache, err := knowledge.NewDefaultSemanticCache(24*time.Hour, logger)

// Cache query results
err = cache.CacheQuery(ctx, "What is AI?", result, 1*time.Hour)

// Find similar cached queries
similar, err := cache.GetSimilarQueries(ctx, "What is artificial intelligence?", 0.8)
```

### Batch Processing

```go
// Add multiple documents efficiently
documents := []*knowledge.Document{doc1, doc2, doc3}
err := manager.AddDocuments(ctx, documents)

// Batch embedding generation
texts := []string{"text1", "text2", "text3"}
embeddings, err := embeddingManager.GenerateEmbeddings(ctx, texts)
```

## 🔒 Security and Access Control

### Security Levels

```go
kbConfig := &knowledge.KnowledgeBaseConfig{
    SecurityLevel: knowledge.SecurityLevelConfidential,
    // ... other config
}
```

### Access Control

```go
// Set permissions
permissions := &knowledge.Permissions{
    Read:   []string{"user1", "group1"},
    Write:  []string{"admin1"},
    Delete: []string{"admin1"},
}

err := security.SetPermissions(ctx, "kb-id", permissions)
```

## 📈 Monitoring and Analytics

### Knowledge Metrics

```go
metrics, err := manager.GetKnowledgeMetrics(ctx)
fmt.Printf("Total Documents: %d\n", metrics.TotalDocuments)
fmt.Printf("Index Health: %.2f\n", metrics.IndexHealth)
```

### Search Analytics

```go
timeRange := &knowledge.TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

analytics, err := manager.GetSearchAnalytics(ctx, timeRange)
fmt.Printf("Total Queries: %d\n", analytics.TotalQueries)
fmt.Printf("Average Response Time: %v\n", analytics.AverageResponseTime)
```

## 🛠️ Configuration

### Knowledge Manager Configuration

```go
config := &knowledge.KnowledgeManagerConfig{
    DefaultEmbeddingModel: "text-embedding-ada-002",
    DefaultChunkSize:      1000,
    DefaultChunkOverlap:   200,
    CacheEnabled:          true,
    CacheTTL:              24 * time.Hour,
    GraphEnabled:          true,
    MultiModalEnabled:     false,
    MaxConcurrentOps:      10,
    MetricsInterval:       5 * time.Minute,
}
```

### Knowledge Base Configuration

```go
kbConfig := &knowledge.KnowledgeBaseConfig{
    EmbeddingModel:     "text-embedding-ada-002",
    ChunkingStrategy:   knowledge.ChunkingStrategyRecursive,
    ChunkSize:          1000,
    ChunkOverlap:       200,
    IndexingEnabled:    true,
    GraphEnabled:       true,
    MultiModalEnabled:  false,
    VersioningEnabled:  true,
    SecurityLevel:      knowledge.SecurityLevelInternal,
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all knowledge management tests
go test ./pkg/knowledge/...

# Run integration tests
go test -tags=integration ./pkg/knowledge/...

# Run with coverage
go test -cover ./pkg/knowledge/...
```

## 📖 Examples

See the complete example in `examples/knowledge_management_example.go` for a comprehensive demonstration of all features.

## 🤝 Contributing

1. Follow the established patterns and interfaces
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability

## 📄 License

This knowledge management system is part of the AIOS project and follows the same licensing terms.
