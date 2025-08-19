# AIOS Knowledge Management - Phase 2 Implementation

## Overview

Phase 2 of the AIOS Knowledge Management system provides a complete, production-ready implementation with database persistence, advanced document processing, and enhanced search capabilities. This implementation is built entirely in Go, providing high performance and reliability.

## Architecture

### Core Components

1. **Knowledge Service** (`internal/knowledge/service.go`)
   - Main service orchestrating all knowledge management operations
   - HTTP API for document upload, search, and crawling
   - Integration with database and processing components

2. **Database Layer** (`internal/knowledge/repository.go`, `internal/knowledge/models.go`)
   - PostgreSQL-based persistence with vector support (pgvector)
   - Comprehensive schema for documents, chunks, knowledge bases, and crawl jobs
   - Optimized queries for search and retrieval

3. **Document Processor** (`internal/knowledge/processor.go`)
   - Advanced text chunking with overlap for better context preservation
   - Support for multiple document formats (text, HTML, PDF planned)
   - Metadata extraction and content normalization

4. **Web Crawler** (`internal/knowledge/crawler.go`)
   - Concurrent web crawling with configurable limits
   - Respect for robots.txt and rate limiting
   - Link extraction and content processing

5. **Vector Searcher** (`internal/knowledge/searcher.go`)
   - Foundation for semantic search capabilities
   - Vector embedding integration (OpenAI, local models)
   - Similarity search and ranking

6. **Enhanced MCP Integration** (`internal/mcp/enhanced/`)
   - MCP (Model Context Protocol) server with knowledge tools
   - Seamless integration with AI models and agents
   - Tool registry for extensible functionality

## Database Schema

### Core Tables

- **knowledge_bases**: Container for organized knowledge collections
- **documents**: Individual documents with metadata and processing status
- **document_chunks**: Text chunks for RAG (Retrieval-Augmented Generation)
- **crawl_jobs**: Web crawling job management and status
- **crawled_pages**: Individual crawled pages with content
- **entities**: Knowledge graph entities (for future graph capabilities)
- **entity_relationships**: Relationships between entities
- **search_queries**: Query analytics and caching
- **search_cache**: Performance optimization for repeated queries

### Vector Support

The schema includes vector columns using pgvector extension for:
- Document chunk embeddings
- Entity embeddings
- Query embeddings for semantic search

## API Endpoints

### Knowledge Service (Port 8181)

- `POST /api/v1/documents` - Upload and process documents
- `GET /api/v1/documents` - List processed documents
- `POST /api/v1/search` - Search knowledge base
- `POST /api/v1/crawl` - Start web crawling job
- `GET /api/v1/crawl/{jobId}` - Get crawl job status
- `GET /api/v1/sources` - List knowledge sources

### Enhanced MCP Service (Port 8051)

- WebSocket and HTTP endpoints for MCP protocol
- Tool registration and execution
- Knowledge-aware AI agent integration

## Configuration

### Environment Variables

```bash
DATABASE_URL=postgres://user:password@localhost:5432/aios?sslmode=disable
OPENAI_API_KEY=your_openai_key  # For embeddings
```

### Service Configuration

```yaml
services:
  knowledge:
    enabled: true
    host: "0.0.0.0"
    port: 8181
  mcp:
    enabled: true
    host: "0.0.0.0"
    port: 8051
```

## Building and Running

### Prerequisites

- Go 1.23+
- PostgreSQL 14+ with pgvector extension
- Optional: OpenAI API key for embeddings

### Build Commands

```bash
# Build knowledge service
go build -o bin/aios-knowledge ./cmd/aios-knowledge

# Build enhanced MCP service
go build -o bin/aios-mcp-enhanced ./cmd/aios-mcp-enhanced

# Run tests
go test ./internal/knowledge/...
```

### Database Setup

```bash
# Run migrations
psql -d aios -f scripts/migrations/000001_initial_schema.up.sql
psql -d aios -f scripts/migrations/000002_knowledge_management.up.sql
```

## Usage Examples

### Document Upload

```bash
curl -X POST http://localhost:8181/api/v1/documents \
  -H "Content-Type: application/json" \
  -d '{
    "file_name": "example.txt",
    "content": "This is example content for processing",
    "mime_type": "text/plain",
    "metadata": {"source": "manual_upload"}
  }'
```

### Knowledge Search

```bash
curl -X POST http://localhost:8181/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "example content",
    "max_results": 10,
    "use_rag": true
  }'
```

### Web Crawling

```bash
curl -X POST http://localhost:8181/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com",
    "max_pages": 50,
    "max_depth": 3,
    "follow_links": true
  }'
```

## Performance Optimizations

1. **Database Indexing**
   - Full-text search indexes on content
   - Vector similarity indexes for embeddings
   - Composite indexes for common query patterns

2. **Caching**
   - Search result caching with TTL
   - Query analytics for optimization
   - Connection pooling for database access

3. **Concurrent Processing**
   - Parallel document chunking
   - Concurrent web crawling
   - Asynchronous embedding generation

## Security Features

1. **Input Validation**
   - Content sanitization
   - URL validation for crawling
   - File type restrictions

2. **Rate Limiting**
   - API endpoint rate limiting
   - Crawling rate limits
   - Resource usage monitoring

3. **Access Control**
   - Knowledge base ownership
   - User-scoped operations
   - Audit logging

## Monitoring and Observability

1. **OpenTelemetry Integration**
   - Distributed tracing across services
   - Performance metrics collection
   - Error tracking and alerting

2. **Structured Logging**
   - JSON-formatted logs
   - Correlation IDs for request tracking
   - Performance timing logs

3. **Health Checks**
   - Database connectivity monitoring
   - Service health endpoints
   - Dependency status checks

## Future Enhancements

1. **Advanced Vector Search**
   - Multiple embedding models
   - Hybrid search (keyword + semantic)
   - Re-ranking algorithms

2. **Knowledge Graph**
   - Entity extraction and linking
   - Relationship discovery
   - Graph-based reasoning

3. **Multi-modal Support**
   - Image and video processing
   - Audio transcription
   - Cross-modal search

4. **Collaborative Features**
   - Knowledge base sharing
   - Collaborative editing
   - Version control

## Troubleshooting

### Common Issues

1. **Database Connection Errors**
   - Verify PostgreSQL is running
   - Check DATABASE_URL format
   - Ensure pgvector extension is installed

2. **Build Failures**
   - Run `go mod tidy` to update dependencies
   - Check Go version compatibility
   - Verify all imports are available

3. **Performance Issues**
   - Monitor database query performance
   - Check vector index usage
   - Review concurrent operation limits

### Debug Mode

Set log level to debug for detailed operation logs:
```bash
export LOG_LEVEL=debug
./bin/aios-knowledge
```

## Contributing

1. Follow Go coding standards
2. Add tests for new functionality
3. Update documentation for API changes
4. Use structured logging for observability
5. Include OpenTelemetry tracing for new operations
