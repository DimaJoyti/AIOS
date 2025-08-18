# Document Processing Pipeline

## Overview

The Document Processing Pipeline provides a comprehensive, production-ready system for ingesting, processing, and transforming documents from various sources. This system is designed to handle multiple document formats, extract content, clean and normalize text, and prepare documents for downstream AI applications including vector storage and semantic search.

## Features

### 🚀 **Core Capabilities**
- **Multi-Source Ingestion**: File system, URLs, streams, and custom readers
- **Content Extraction**: PDF, DOCX, HTML, Markdown, JSON, and plain text
- **Text Processing**: Cleaning, normalization, language detection, and custom processing
- **Document Chunking**: Fixed-size, sentence-based, and custom chunking strategies
- **Pipeline Architecture**: Configurable, extensible processing workflows
- **Batch & Stream Processing**: Efficient processing of single documents or large batches
- **Metrics & Monitoring**: Comprehensive performance tracking and observability

### 🏗️ **Architecture**
- **Interface-Driven Design**: Clean abstractions for easy extension and testing
- **Pipeline Pattern**: Configurable stages with error handling and recovery
- **Factory Pattern**: Dynamic component creation and registration
- **Builder Pattern**: Fluent interfaces for complex configurations
- **Observability**: Full OpenTelemetry integration with metrics and tracing
- **Type Safety**: Strong typing throughout with comprehensive validation

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/aios/aios/pkg/docprocessing"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    manager := docprocessing.NewProcessingManager(logger)
    
    // Configure processing pipeline
    config := &docprocessing.ProcessingConfig{
        PipelineStages: []docprocessing.StageConfig{
            {
                Name:    "extraction",
                Type:    "extraction",
                Enabled: true,
                Order:   1,
            },
            {
                Name:    "processing",
                Type:    "processing",
                Enabled: true,
                Order:   2,
                Options: map[string]interface{}{
                    "processor_types": []string{"cleaning", "language_detection"},
                },
            },
            {
                Name:    "chunking",
                Type:    "chunking",
                Enabled: true,
                Order:   3,
                Options: map[string]interface{}{
                    "chunker_type": "fixed_size",
                    "chunk_size":   1000,
                    "overlap":      200,
                },
            },
        },
        MaxConcurrency:    5,
        ProcessingTimeout: 30 * time.Second,
    }
    
    // Create processor
    processor, err := manager.CreateProcessor(config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Process a document
    doc := &docprocessing.Document{
        ID:          "doc1",
        Content:     "This is a sample document to be processed.",
        ContentType: "text/plain",
        Metadata:    make(map[string]interface{}),
    }
    
    result, err := processor.ProcessDocument(context.Background(), doc)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Processed document: %s", result.Document.ID)
    log.Printf("Success: %v", result.Success)
    log.Printf("Duration: %v", result.Duration)
    log.Printf("Chunks: %v", result.Document.Metadata["chunk_count"])
}
```

## Document Sources

### File Source
Process documents from the file system.

```go
fileSource := docprocessing.NewFileSource("/path/to/documents", logger)

// Configure file patterns and options
fileSource.Configure(map[string]interface{}{
    "patterns":      []string{"*.txt", "*.md", "*.html"},
    "recursive":     true,
    "max_file_size": int64(10 * 1024 * 1024), // 10MB
})

// Get documents
docChan, err := fileSource.GetDocuments(context.Background())
```

### URL Source
Process documents from web URLs.

```go
urls := []string{
    "https://example.com/doc1.html",
    "https://example.com/doc2.html",
}

urlSource := docprocessing.NewURLSource(urls, logger)

// Configure HTTP options
urlSource.Configure(map[string]interface{}{
    "user_agent": "AIOS-DocumentProcessor/1.0",
    "timeout":    30 * time.Second,
    "headers": map[string]string{
        "Authorization": "Bearer token",
    },
})
```

### Reader Source
Process documents from io.Reader streams.

```go
readerSource := docprocessing.NewReaderSource(logger)

// Add readers
id1 := readerSource.AddReader(strings.NewReader("content1"), "text/plain", nil)
id2 := readerSource.AddReader(fileReader, "application/pdf", metadata)

// Get document
doc, err := readerSource.GetDocument(context.Background(), id1)
```

## Content Extractors

### Text Extractor
Handles plain text documents with basic cleaning.

```go
extractor := docprocessing.NewTextExtractor(logger)

// Supported types: text/plain, text/csv, application/json, application/xml
```

### HTML Extractor
Extracts content from HTML documents.

```go
extractor := docprocessing.NewHTMLExtractor(logger)

// Features:
// - Removes scripts and styles
// - Extracts text content
// - Preserves links and images metadata
// - Extracts meta tags
```

### Markdown Extractor
Processes Markdown documents.

```go
extractor := docprocessing.NewMarkdownExtractor(logger)

// Features:
// - Extracts front matter
// - Preserves or converts formatting
// - Extracts headers and links
// - Handles code blocks
```

### JSON Extractor
Extracts text content from JSON documents.

```go
extractor := docprocessing.NewJSONExtractor(logger)

// Converts JSON structure to readable text
```

## Text Processors

### Cleaning Processor
Cleans and normalizes text content.

```go
processor := docprocessing.NewCleaningProcessor(logger)

processor.Configure(map[string]interface{}{
    "remove_extra_whitespace": true,
    "remove_control_chars":    true,
    "remove_empty_lines":      true,
    "trim_lines":              true,
})
```

### Language Detection Processor
Detects the language of text content.

```go
processor := docprocessing.NewLanguageDetectionProcessor(logger)

processor.Configure(map[string]interface{}{
    "supported_languages": []string{"en", "es", "fr", "de"},
    "default_language":    "en",
})
```

### Normalization Processor
Normalizes text for consistent processing.

```go
processor := docprocessing.NewNormalizationProcessor(logger)

processor.Configure(map[string]interface{}{
    "to_lower_case":       true,
    "remove_accents":      true,
    "remove_punctuation":  false,
})
```

## Document Chunkers

### Fixed Size Chunker
Splits documents into fixed-size chunks.

```go
chunker := docprocessing.NewFixedSizeChunker(1000, 200, logger)

// Parameters: chunk size, overlap, logger
// Tries to break at word boundaries
```

### Sentence Chunker
Splits documents based on sentence boundaries.

```go
chunker := docprocessing.NewSentenceChunker(1000, 2, logger)

// Parameters: max chunk size, sentence overlap, logger
// Preserves sentence integrity
```

## Pipeline Configuration

### Complete Configuration Example

```go
config := &docprocessing.ProcessingConfig{
    // Pipeline stages
    PipelineStages: []docprocessing.StageConfig{
        {
            Name:    "extraction",
            Type:    "extraction",
            Enabled: true,
            Order:   1,
            Options: map[string]interface{}{
                "content_types": []string{"text/html", "text/markdown"},
            },
        },
        {
            Name:    "text_processing",
            Type:    "processing",
            Enabled: true,
            Order:   2,
            Options: map[string]interface{}{
                "processor_types": []string{
                    "cleaning",
                    "language_detection",
                    "normalization",
                },
            },
        },
        {
            Name:    "chunking",
            Type:    "chunking",
            Enabled: true,
            Order:   3,
            Options: map[string]interface{}{
                "chunker_type": "sentence",
                "max_chunk_size": 1000,
                "overlap": 2,
            },
        },
    },
    
    // Processing options
    MaxConcurrency:    10,
    ProcessingTimeout: 60 * time.Second,
    RetryAttempts:     3,
    RetryDelay:        1 * time.Second,
    
    // Content extraction options
    ExtractImages:      true,
    ExtractTables:      true,
    ExtractMetadata:    true,
    PreserveFormatting: false,
    
    // Text processing options
    CleanText:          true,
    DetectLanguage:     true,
    NormalizeText:      true,
    SupportedLanguages: []string{"en", "es", "fr"},
    
    // Output options
    IncludeMetadata: true,
    IncludeChunks:   true,
    
    // Monitoring options
    EnableMetrics: true,
    EnableTracing: true,
    LogLevel:      "info",
}
```

## Processing Modes

### Single Document Processing

```go
result, err := processor.ProcessDocument(ctx, document)
```

### Batch Processing

```go
results, err := processor.GetPipeline().ProcessDocuments(ctx, documents)
```

### Stream Processing

```go
resultChan, err := processor.GetPipeline().ProcessStream(ctx, docChan)
for result := range resultChan {
    // Process result
}
```

### File Processing

```go
result, err := processor.ProcessFile(ctx, "/path/to/file.txt")
```

### URL Processing

```go
result, err := processor.ProcessURL(ctx, "https://example.com/document.html")
```

### Reader Processing

```go
result, err := processor.ProcessReader(ctx, reader, "text/plain", metadata)
```

## Custom Components

### Custom Text Processor

```go
type CustomProcessor struct {
    logger *logrus.Logger
}

func (cp *CustomProcessor) Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error) {
    // Custom processing logic
    processedText := strings.ToUpper(text)
    
    if metadata == nil {
        metadata = make(map[string]interface{})
    }
    metadata["custom_processing"] = true
    
    return processedText, metadata, nil
}

func (cp *CustomProcessor) GetProcessorType() string {
    return "custom"
}

func (cp *CustomProcessor) Configure(options map[string]interface{}) error {
    return nil
}

// Register with manager
manager.RegisterProcessor(customProcessor)
```

### Custom Document Chunker

```go
type CustomChunker struct {
    logger *logrus.Logger
}

func (cc *CustomChunker) ChunkDocument(ctx context.Context, doc *Document) ([]*DocumentChunk, error) {
    // Custom chunking logic
    chunks := make([]*DocumentChunk, 0)
    
    // Split by custom logic
    parts := strings.Split(doc.Content, "---")
    
    for i, part := range parts {
        chunk := &DocumentChunk{
            ID:         fmt.Sprintf("%s_custom_%d", doc.ID, i),
            DocumentID: doc.ID,
            Content:    strings.TrimSpace(part),
            ChunkIndex: i,
            Metadata: map[string]interface{}{
                "chunk_type": "custom",
            },
        }
        chunks = append(chunks, chunk)
    }
    
    return chunks, nil
}

func (cc *CustomChunker) GetChunkerType() string {
    return "custom"
}

func (cc *CustomChunker) Configure(options map[string]interface{}) error {
    return nil
}

// Register with manager
manager.RegisterChunker(customChunker)
```

## Metrics and Monitoring

### Processing Metrics

```go
metrics := processor.GetMetrics()

fmt.Printf("Total Documents: %d\n", metrics.TotalDocuments)
fmt.Printf("Processed: %d\n", metrics.ProcessedDocuments)
fmt.Printf("Failed: %d\n", metrics.FailedDocuments)
fmt.Printf("Total Chunks: %d\n", metrics.TotalChunks)
fmt.Printf("Average Process Time: %v\n", metrics.AverageProcessTime)
fmt.Printf("Error Rate: %.2f%%\n", metrics.ErrorRate*100)
fmt.Printf("Throughput: %.2f docs/sec\n", metrics.ThroughputPerSec)
```

### OpenTelemetry Integration

All operations are automatically instrumented with OpenTelemetry:

- **Traces**: Request tracing across all processing stages
- **Metrics**: Operation counts, latencies, error rates
- **Attributes**: Document IDs, content types, processing stages

## Error Handling

### Processing Errors

```go
result, err := processor.ProcessDocument(ctx, doc)
if err != nil {
    if processingErr, ok := err.(*docprocessing.ProcessingError); ok {
        fmt.Printf("Stage: %s\n", processingErr.Stage)
        fmt.Printf("Message: %s\n", processingErr.Message)
        fmt.Printf("Cause: %v\n", processingErr.Cause)
    }
}
```

### Retry Configuration

```go
config := &docprocessing.ProcessingConfig{
    RetryAttempts: 3,
    RetryDelay:    1 * time.Second,
    // ... other config
}
```

## Integration Examples

### Vector Database Integration

```go
// Process documents and store in vector database
results, err := processor.GetPipeline().ProcessDocuments(ctx, documents)
if err != nil {
    return err
}

for _, result := range results {
    if !result.Success {
        continue
    }
    
    // Extract chunks for vector storage
    if chunks, exists := result.Document.Metadata["chunks"]; exists {
        if chunkList, ok := chunks.([]*docprocessing.DocumentChunk); ok {
            for _, chunk := range chunkList {
                // Convert to vector database document
                vectorDoc := &vectordb.Document{
                    ID:       chunk.ID,
                    Content:  chunk.Content,
                    Metadata: chunk.Metadata,
                }
                
                // Add to vector store
                err := vectorStore.AddDocuments(ctx, "collection", []*vectordb.Document{vectorDoc})
                if err != nil {
                    log.Printf("Failed to add chunk to vector store: %v", err)
                }
            }
        }
    }
}
```

### Memory System Integration

```go
// Store processed documents in memory system
for _, result := range results {
    if result.Success {
        memoryEntry := &memory.Entry{
            ID:       result.Document.ID,
            Content:  result.Document.Content,
            Metadata: result.Document.Metadata,
        }
        
        err := memorySystem.Store(ctx, memoryEntry)
        if err != nil {
            log.Printf("Failed to store in memory: %v", err)
        }
    }
}
```

## Performance Optimization

### Concurrency Tuning

```go
config := &docprocessing.ProcessingConfig{
    MaxConcurrency: runtime.NumCPU() * 2, // Adjust based on workload
    // ... other config
}
```

### Memory Management

```go
// For large documents, consider streaming processing
resultChan, err := processor.GetPipeline().ProcessStream(ctx, docChan)

// Process results as they come to avoid memory buildup
for result := range resultChan {
    // Process immediately
    handleResult(result)
}
```

### Batch Size Optimization

```go
// Process documents in optimal batch sizes
batchSize := 100
for i := 0; i < len(documents); i += batchSize {
    end := i + batchSize
    if end > len(documents) {
        end = len(documents)
    }
    
    batch := documents[i:end]
    results, err := processor.GetPipeline().ProcessDocuments(ctx, batch)
    // Handle results
}
```

## Testing

### Unit Tests

```bash
go test ./pkg/docprocessing/...
```

### Integration Tests

```bash
# Test with real files
go test -tags=integration ./pkg/docprocessing/...
```

### Mock Testing

The package provides comprehensive mocks for testing:

```go
// Use mock components for testing
mockExtractor := &MockContentExtractor{}
mockProcessor := &MockTextProcessor{}
mockChunker := &MockDocumentChunker{}
```

## Best Practices

1. **Configuration Management**: Use environment variables for sensitive configuration
2. **Error Handling**: Implement proper error handling and logging
3. **Resource Management**: Close sources and clean up resources properly
4. **Performance Monitoring**: Monitor processing metrics and optimize accordingly
5. **Security**: Validate input documents and sanitize content
6. **Testing**: Write comprehensive tests including edge cases
7. **Documentation**: Document custom processors and chunkers

## Integration with AIOS

This document processing system integrates seamlessly with the AIOS ecosystem:

- **Vector Database**: Direct integration for semantic search
- **Memory Systems**: Store processed documents in memory
- **AI Orchestrator**: Use in AI workflows and chains
- **Tool Framework**: Expose as tools through MCP protocol
- **Observability**: Full integration with AIOS monitoring stack

## Examples

See `examples.go` for comprehensive usage examples including:

- Basic document processing
- Batch processing workflows
- Stream processing patterns
- Custom component development
- Integration with other AIOS services

## Contributing

To add new extractors, processors, or chunkers:

1. Implement the appropriate interface
2. Add comprehensive tests
3. Update documentation
4. Register with the default manager if appropriate

## License

This document processing pipeline is part of the AIOS project and follows the same licensing terms.
