# AIOS Enhanced Services

This directory contains the enhanced AIOS services that integrate Archon's knowledge management capabilities using Go-based microservices.

## Services Overview

### Knowledge Service
- **Command**: `cmd/aios-knowledge/`
- **Implementation**: `internal/knowledge/`
- **Technology**: Go + Gorilla Mux + OpenTelemetry
- **Port**: 8181
- **Features**:
  - Web crawling with goquery
  - Document processing and chunking
  - Vector-based search and RAG
  - Real-time WebSocket updates
  - RESTful API endpoints

### Enhanced MCP Service
- **Command**: `cmd/aios-mcp-enhanced/`
- **Implementation**: `internal/mcp/enhanced/`
- **Technology**: Go + WebSocket + MCP Protocol
- **Ports**: 8051 (MCP), 8052 (Management)
- **Features**:
  - 10 comprehensive MCP tools
  - Knowledge search integration
  - Document upload and processing
  - Web crawling capabilities
  - Project management tools
  - Session management
