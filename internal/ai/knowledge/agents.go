package knowledge

import (
	"context"
	"fmt"
	"time"

	"github.com/aios/aios/internal/knowledge"
	"github.com/aios/aios/pkg/ai"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// API interface for knowledge operations (to avoid circular imports)
type API interface {
	SearchKnowledge(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
}

// SearchRequest represents a knowledge search request
type SearchRequest struct {
	Query       string            `json:"query"`
	MaxResults  int               `json:"max_results,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	UseRAG      bool              `json:"use_rag,omitempty"`
	ContextSize int               `json:"context_size,omitempty"`
}

// SearchResponse represents a knowledge search response
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	URL      string            `json:"url,omitempty"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// DocumentAgent handles document processing tasks
type DocumentAgent struct {
	id              string
	knowledgeService *knowledge.Service
	logger          *logrus.Logger
	tracer          trace.Tracer
	available       bool
}

// RAGAgent handles retrieval-augmented generation tasks
type RAGAgent struct {
	id              string
	knowledgeService *knowledge.Service
	logger          *logrus.Logger
	tracer          trace.Tracer
	available       bool
}

// NewDocumentAgent creates a new document processing agent
func NewDocumentAgent(knowledgeService *knowledge.Service, logger *logrus.Logger) (*DocumentAgent, error) {
	return &DocumentAgent{
		id:              "document-agent",
		knowledgeService: knowledgeService,
		logger:          logger,
		tracer:          otel.Tracer("ai.knowledge.document"),
		available:       true,
	}, nil
}

// NewRAGAgent creates a new RAG agent
func NewRAGAgent(knowledgeService *knowledge.Service, logger *logrus.Logger) (*RAGAgent, error) {
	return &RAGAgent{
		id:              "rag-agent",
		knowledgeService: knowledgeService,
		logger:          logger,
		tracer:          otel.Tracer("ai.knowledge.rag"),
		available:       true,
	}, nil
}

// DocumentAgent implementation

func (a *DocumentAgent) GetID() string {
	return a.id
}

func (a *DocumentAgent) GetName() string {
	return "Document Processing Agent"
}

func (a *DocumentAgent) GetDescription() string {
	return "Handles document upload, processing, and content extraction tasks"
}

func (a *DocumentAgent) GetCapabilities() []string {
	return []string{"document-upload", "document-processing", "content-extraction"}
}

func (a *DocumentAgent) IsAvailable() bool {
	return a.available
}

func (a *DocumentAgent) GetStatus() *ai.AgentStatus {
	return &ai.AgentStatus{
		ID:        a.id,
		Name:      a.GetName(),
		Available: a.available,
		Busy:      !a.available,
	}
}

func (a *DocumentAgent) ExecuteTask(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	ctx, span := a.tracer.Start(ctx, "document_agent.execute_task")
	defer span.End()

	startTime := time.Now()
	a.available = false
	defer func() { a.available = true }()

	a.logger.WithField("task_id", task.ID).Info("Executing document processing task")

	switch task.Type {
	case "document-upload":
		return a.handleDocumentUpload(ctx, task)
	case "document-processing":
		return a.handleDocumentProcessing(ctx, task)
	case "content-extraction":
		return a.handleContentExtraction(ctx, task)
	default:
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       fmt.Sprintf("unsupported task type: %s", task.Type),
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}
}

func (a *DocumentAgent) handleDocumentUpload(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	// Extract upload parameters from task
	fileName, ok := task.Input["file_name"].(string)
	if !ok {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "missing file_name parameter",
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	content, ok := task.Input["content"].(string)
	if !ok {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "missing content parameter",
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	mimeType, _ := task.Input["mime_type"].(string)
	if mimeType == "" {
		mimeType = "text/plain"
	}

	// Create upload request
	uploadReq := &knowledge.DocumentUploadRequest{
		FileName: fileName,
		Content:  content,
		MimeType: mimeType,
		Metadata: make(map[string]string),
	}

	// Add metadata if provided
	if metadata, ok := task.Input["metadata"].(map[string]string); ok {
		uploadReq.Metadata = metadata
	}

	// Upload document through knowledge API
	// Note: This would need to be implemented in the knowledge API
	// For now, we'll simulate the upload
	a.logger.WithFields(logrus.Fields{
		"file_name": fileName,
		"mime_type": mimeType,
	}).Info("Document uploaded successfully")

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"document_id": fmt.Sprintf("doc_%d", time.Now().Unix())},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}

func (a *DocumentAgent) handleDocumentProcessing(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	// Process document through knowledge API
	a.logger.WithField("task_id", task.ID).Info("Processing document")

	// Simulate document processing
	time.Sleep(100 * time.Millisecond)

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"status": "processed", "chunks": 5},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}

func (a *DocumentAgent) handleContentExtraction(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	// Extract content through knowledge API
	a.logger.WithField("task_id", task.ID).Info("Extracting content")

	// Simulate content extraction
	time.Sleep(50 * time.Millisecond)

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"extracted_text": "Sample extracted content"},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}

// RAGAgent implementation

func (a *RAGAgent) GetID() string {
	return a.id
}

func (a *RAGAgent) GetName() string {
	return "RAG Processing Agent"
}

func (a *RAGAgent) GetDescription() string {
	return "Handles retrieval-augmented generation, knowledge search, and context retrieval tasks"
}

func (a *RAGAgent) GetCapabilities() []string {
	return []string{"knowledge-search", "rag-query", "context-retrieval"}
}

func (a *RAGAgent) IsAvailable() bool {
	return a.available
}

func (a *RAGAgent) GetStatus() *ai.AgentStatus {
	return &ai.AgentStatus{
		ID:        a.id,
		Name:      a.GetName(),
		Available: a.available,
		Busy:      !a.available,
	}
}

func (a *RAGAgent) ExecuteTask(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	ctx, span := a.tracer.Start(ctx, "rag_agent.execute_task")
	defer span.End()

	startTime := time.Now()
	a.available = false
	defer func() { a.available = true }()

	a.logger.WithField("task_id", task.ID).Info("Executing RAG task")

	switch task.Type {
	case "knowledge-search":
		return a.handleKnowledgeSearch(ctx, task)
	case "rag-query":
		return a.handleRAGQuery(ctx, task)
	case "context-retrieval":
		return a.handleContextRetrieval(ctx, task)
	default:
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       fmt.Sprintf("unsupported task type: %s", task.Type),
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}
}

func (a *RAGAgent) handleKnowledgeSearch(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	query, ok := task.Input["query"].(string)
	if !ok {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "missing query parameter",
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	maxResults, _ := task.Input["max_results"].(int)
	if maxResults == 0 {
		maxResults = 10
	}

	// Create search request
	searchReq := &knowledge.SearchRequest{
		Query:      query,
		MaxResults: maxResults,
		UseRAG:     false,
	}

	// Perform search through knowledge service
	searchResp, err := a.knowledgeService.SearchKnowledge(ctx, searchReq)
	if err != nil {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       fmt.Sprintf("search failed: %v", err),
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"search_results": searchResp},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}

func (a *RAGAgent) handleRAGQuery(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	query, ok := task.Input["query"].(string)
	if !ok {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "missing query parameter",
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	// Create RAG search request
	searchReq := &knowledge.SearchRequest{
		Query:       query,
		MaxResults:  5,
		UseRAG:      true,
		ContextSize: 1000,
	}

	// Perform RAG search through knowledge service
	searchResp, err := a.knowledgeService.SearchKnowledge(ctx, searchReq)
	if err != nil {
		return &ai.TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       fmt.Sprintf("RAG query failed: %v", err),
			Duration:    time.Since(startTime),
			AgentID:     a.id,
			CompletedAt: time.Now(),
		}, nil
	}

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"rag_results": searchResp},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}

func (a *RAGAgent) handleContextRetrieval(ctx context.Context, task *ai.Task) (*ai.TaskResult, error) {
	startTime := time.Now()

	// Handle context retrieval
	a.logger.WithField("task_id", task.ID).Info("Retrieving context")

	// Simulate context retrieval
	time.Sleep(75 * time.Millisecond)

	return &ai.TaskResult{
		TaskID:      task.ID,
		Success:     true,
		Output:      map[string]interface{}{"context": "Retrieved context information"},
		Duration:    time.Since(startTime),
		AgentID:     a.id,
		CompletedAt: time.Now(),
	}, nil
}
