package enhanced

import (
	"context"
	"fmt"
	"time"

	"github.com/aios/aios/internal/ai/knowledge"
	knowledgeService "github.com/aios/aios/internal/knowledge"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// EnhancedToolRegistry manages enhanced MCP tools with agent integration
type EnhancedToolRegistry struct {
	tools            map[string]*EnhancedTool
	sessionManager   *SessionManager
	knowledgeAgent   *knowledge.DocumentAgent
	ragAgent         *knowledge.RAGAgent
	knowledgeService *knowledgeService.Service
	logger           *logrus.Logger
	tracer           trace.Tracer
}

// EnhancedTool represents an enhanced MCP tool with advanced capabilities
type EnhancedTool struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Parameters   map[string]interface{} `json:"parameters"`
	Handler      ToolHandler            `json:"-"`
	Capabilities []string               `json:"capabilities"`
	RequiresAuth bool                   `json:"requires_auth"`
	Streaming    bool                   `json:"streaming"`
	Chainable    bool                   `json:"chainable"`
	Category     string                 `json:"category"`
}

// ToolHandler defines the interface for tool handlers
type ToolHandler func(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error)

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success     bool                   `json:"success"`
	Data        interface{}            `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	StreamData  chan interface{}       `json:"-"`
	IsStreaming bool                   `json:"is_streaming"`
	NextTools   []string               `json:"next_tools,omitempty"`
}

// StreamingResponse represents a streaming response
type StreamingResponse struct {
	Type      string                 `json:"type"` // data, progress, complete, error
	Data      interface{}            `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewEnhancedToolRegistry creates a new enhanced tool registry
func NewEnhancedToolRegistry(
	sessionManager *SessionManager,
	knowledgeAgent *knowledge.DocumentAgent,
	ragAgent *knowledge.RAGAgent,
	knowledgeService *knowledgeService.Service,
	logger *logrus.Logger,
) *EnhancedToolRegistry {
	registry := &EnhancedToolRegistry{
		tools:            make(map[string]*EnhancedTool),
		sessionManager:   sessionManager,
		knowledgeAgent:   knowledgeAgent,
		ragAgent:         ragAgent,
		knowledgeService: knowledgeService,
		logger:           logger,
		tracer:           otel.Tracer("mcp.enhanced.tools"),
	}

	// Register enhanced tools
	registry.registerEnhancedTools()
	return registry
}

// registerEnhancedTools registers all enhanced tools
func (etr *EnhancedToolRegistry) registerEnhancedTools() {
	// Enhanced Knowledge Search with streaming and context
	etr.RegisterTool(&EnhancedTool{
		Name:        "enhanced_knowledge_search",
		Description: "Advanced knowledge search with context awareness and streaming results",
		Parameters: map[string]interface{}{
			"query":          map[string]interface{}{"type": "string", "required": true},
			"max_results":    map[string]interface{}{"type": "integer", "default": 10},
			"use_context":    map[string]interface{}{"type": "boolean", "default": true},
			"stream":         map[string]interface{}{"type": "boolean", "default": false},
			"knowledge_base": map[string]interface{}{"type": "string", "required": false},
		},
		Handler:      etr.handleEnhancedKnowledgeSearch,
		Capabilities: []string{"search", "context", "streaming"},
		Streaming:    true,
		Chainable:    true,
		Category:     "knowledge",
	})

	// Enhanced RAG Query with memory integration
	etr.RegisterTool(&EnhancedTool{
		Name:        "enhanced_rag_query",
		Description: "Advanced RAG query with session memory and context integration",
		Parameters: map[string]interface{}{
			"query":        map[string]interface{}{"type": "string", "required": true},
			"context_size": map[string]interface{}{"type": "integer", "default": 5},
			"use_memory":   map[string]interface{}{"type": "boolean", "default": true},
			"temperature":  map[string]interface{}{"type": "number", "default": 0.7},
		},
		Handler:      etr.handleEnhancedRAGQuery,
		Capabilities: []string{"rag", "memory", "context"},
		Chainable:    true,
		Category:     "ai",
	})

	// Document Analysis with agent integration
	etr.RegisterTool(&EnhancedTool{
		Name:        "enhanced_document_analysis",
		Description: "Advanced document analysis using knowledge agents",
		Parameters: map[string]interface{}{
			"document_id":   map[string]interface{}{"type": "string", "required": true},
			"analysis_type": map[string]interface{}{"type": "string", "enum": []string{"summary", "entities", "topics", "sentiment"}},
			"depth":         map[string]interface{}{"type": "string", "enum": []string{"shallow", "medium", "deep"}, "default": "medium"},
		},
		Handler:      etr.handleEnhancedDocumentAnalysis,
		Capabilities: []string{"analysis", "agents", "nlp"},
		Chainable:    true,
		Category:     "analysis",
	})

	// Workflow Orchestration
	etr.RegisterTool(&EnhancedTool{
		Name:        "workflow_orchestrator",
		Description: "Orchestrate multi-step workflows with tool chaining",
		Parameters: map[string]interface{}{
			"workflow_type": map[string]interface{}{"type": "string", "required": true},
			"steps":         map[string]interface{}{"type": "array", "required": true},
			"parameters":    map[string]interface{}{"type": "object", "default": map[string]interface{}{}},
		},
		Handler:      etr.handleWorkflowOrchestration,
		Capabilities: []string{"workflow", "orchestration", "chaining"},
		Chainable:    false,
		Category:     "workflow",
	})

	// Context Manager
	etr.RegisterTool(&EnhancedTool{
		Name:        "context_manager",
		Description: "Manage session context and memory",
		Parameters: map[string]interface{}{
			"action": map[string]interface{}{"type": "string", "enum": []string{"get", "set", "clear", "search"}},
			"key":    map[string]interface{}{"type": "string", "required": false},
			"value":  map[string]interface{}{"type": "any", "required": false},
			"query":  map[string]interface{}{"type": "string", "required": false},
		},
		Handler:      etr.handleContextManager,
		Capabilities: []string{"context", "memory", "session"},
		Chainable:    true,
		Category:     "session",
	})
}

// RegisterTool registers a new enhanced tool
func (etr *EnhancedToolRegistry) RegisterTool(tool *EnhancedTool) {
	etr.tools[tool.Name] = tool
	etr.logger.WithField("tool_name", tool.Name).Info("Registered enhanced tool")
}

// ExecuteTool executes a tool with enhanced capabilities
func (etr *EnhancedToolRegistry) ExecuteTool(ctx context.Context, sessionID, toolName string, params map[string]interface{}) (*ToolResult, error) {
	ctx, span := etr.tracer.Start(ctx, "enhanced_tools.execute_tool")
	defer span.End()

	tool, exists := etr.tools[toolName]
	if !exists {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("tool not found: %s", toolName),
		}, nil
	}

	// Update session tool state
	session, err := etr.sessionManager.GetSession(sessionID)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("session not found: %s", sessionID),
		}, nil
	}

	// Track tool usage
	toolState := &ToolState{
		ToolName:   toolName,
		State:      "running",
		LastUsed:   time.Now(),
		Parameters: params,
	}
	session.ActiveTools[toolName] = toolState

	// Execute tool
	result, err := tool.Handler(ctx, sessionID, params)
	if err != nil {
		toolState.State = "error"
		toolState.ErrorMessage = err.Error()
		return &ToolResult{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Update tool state
	toolState.State = "completed"
	toolState.Results = result.Data

	// Add to session memory if successful
	if result.Success {
		memoryItem := MemoryItem{
			ID:      uuid.New().String(),
			Type:    "tool_result",
			Content: fmt.Sprintf("Tool %s executed successfully", toolName),
			Metadata: map[string]interface{}{
				"tool_name":  toolName,
				"parameters": params,
				"result":     result.Data,
			},
			Timestamp: time.Now(),
			Relevance: 0.8,
		}
		etr.sessionManager.AddToMemory(sessionID, memoryItem)
	}

	return result, nil
}

// handleEnhancedKnowledgeSearch handles enhanced knowledge search
func (etr *EnhancedToolRegistry) handleEnhancedKnowledgeSearch(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error) {
	query, _ := params["query"].(string)
	maxResults, _ := params["max_results"].(int)
	useContext, _ := params["use_context"].(bool)
	stream, _ := params["stream"].(bool)

	if maxResults == 0 {
		maxResults = 10
	}

	// Get session context if requested
	var contextQuery string
	if useContext {
		session, err := etr.sessionManager.GetSession(sessionID)
		if err == nil && session.Context.LastQuery != "" {
			contextQuery = fmt.Sprintf("%s %s", session.Context.LastQuery, query)
		} else {
			contextQuery = query
		}
	} else {
		contextQuery = query
	}

	// Execute search using knowledge service
	searchRequest := &knowledgeService.SearchRequest{
		Query:      contextQuery,
		MaxResults: maxResults,
		UseRAG:     true,
	}
	results, err := etr.knowledgeService.SearchKnowledge(ctx, searchRequest)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("search failed: %v", err),
		}, nil
	}

	// Update session context
	etr.sessionManager.UpdateSessionContext(sessionID, map[string]interface{}{
		"last_query": query,
	})

	result := &ToolResult{
		Success: true,
		Data:    results,
		Metadata: map[string]interface{}{
			"query":         query,
			"context_query": contextQuery,
			"result_count":  len(results.Results),
			"used_context":  useContext,
		},
		IsStreaming: stream,
	}

	// Handle streaming if requested
	if stream {
		result.StreamData = make(chan interface{}, 10)
		// Convert SearchResult slice to interface{} slice for streaming
		interfaceResults := make([]interface{}, len(results.Results))
		for i, r := range results.Results {
			interfaceResults[i] = r
		}
		go etr.streamSearchResults(result.StreamData, interfaceResults)
	}

	return result, nil
}

// handleEnhancedRAGQuery handles enhanced RAG queries
func (etr *EnhancedToolRegistry) handleEnhancedRAGQuery(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error) {
	query, _ := params["query"].(string)
	contextSize, _ := params["context_size"].(int)
	useMemory, _ := params["use_memory"].(bool)
	temperature, _ := params["temperature"].(float64)

	if contextSize == 0 {
		contextSize = 5
	}
	if temperature == 0 {
		temperature = 0.7
	}

	// Get relevant memory if requested
	var memoryContext []MemoryItem
	if useMemory {
		var err error
		memoryContext, err = etr.sessionManager.GetRelevantMemory(sessionID, query, 3)
		if err != nil {
			etr.logger.WithError(err).Warn("Failed to get relevant memory")
		}
	}

	// Create RAG search request
	searchRequest := &knowledgeService.SearchRequest{
		Query:       query,
		MaxResults:  5,
		UseRAG:      true,
		ContextSize: contextSize,
	}

	// Execute RAG query using knowledge service
	response, err := etr.knowledgeService.SearchKnowledge(ctx, searchRequest)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("RAG query failed: %v", err),
		}, nil
	}

	return &ToolResult{
		Success: true,
		Data:    response,
		Metadata: map[string]interface{}{
			"query":        query,
			"context_size": contextSize,
			"used_memory":  useMemory,
			"memory_items": len(memoryContext),
		},
	}, nil
}

// handleEnhancedDocumentAnalysis handles document analysis
func (etr *EnhancedToolRegistry) handleEnhancedDocumentAnalysis(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error) {
	documentID, _ := params["document_id"].(string)
	analysisType, _ := params["analysis_type"].(string)
	depth, _ := params["depth"].(string)

	if analysisType == "" {
		analysisType = "summary"
	}
	if depth == "" {
		depth = "medium"
	}

	// For now, simulate document retrieval (in real implementation, this would query the knowledge service)
	// TODO: Implement actual document retrieval from knowledge service
	_ = documentID // Mark as used
	err := error(nil)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("failed to get document: %v", err),
		}, nil
	}

	// Perform analysis based on type
	var analysisResult interface{}
	switch analysisType {
	case "summary":
		// Simulate document summarization
		analysisResult = map[string]interface{}{
			"summary": "This is a simulated summary of the document",
			"depth":   depth,
		}
	case "entities":
		// Simulate entity extraction
		analysisResult = map[string]interface{}{
			"entities": []string{"Entity1", "Entity2", "Entity3"},
		}
	case "topics":
		// Simulate topic extraction
		analysisResult = map[string]interface{}{
			"topics": []string{"Topic1", "Topic2", "Topic3"},
		}
	case "sentiment":
		// Simulate sentiment analysis
		analysisResult = map[string]interface{}{
			"sentiment": "positive",
			"score":     0.8,
		}
	default:
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported analysis type: %s", analysisType),
		}, nil
	}

	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("analysis failed: %v", err),
		}, nil
	}

	return &ToolResult{
		Success: true,
		Data:    analysisResult,
		Metadata: map[string]interface{}{
			"document_id":   documentID,
			"analysis_type": analysisType,
			"depth":         depth,
		},
	}, nil
}

// handleWorkflowOrchestration handles workflow orchestration
func (etr *EnhancedToolRegistry) handleWorkflowOrchestration(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error) {
	workflowType, _ := params["workflow_type"].(string)
	steps, _ := params["steps"].([]interface{})
	workflowParams, _ := params["parameters"].(map[string]interface{})

	if workflowParams == nil {
		workflowParams = make(map[string]interface{})
	}

	// Create workflow state
	workflowID := uuid.New().String()
	workflowState := &WorkflowState{
		WorkflowID:   workflowID,
		CurrentStep:  0,
		TotalSteps:   len(steps),
		StepResults:  make([]interface{}, 0),
		WorkflowType: workflowType,
		Parameters:   workflowParams,
		Status:       "running",
	}

	// Update session with workflow state
	session, err := etr.sessionManager.GetSession(sessionID)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("session not found: %s", sessionID),
		}, nil
	}
	session.WorkflowState = workflowState

	// Execute workflow steps
	results := make([]interface{}, 0)
	for i, step := range steps {
		stepMap, ok := step.(map[string]interface{})
		if !ok {
			workflowState.Status = "error"
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("invalid step format at index %d", i),
			}, nil
		}

		toolName, _ := stepMap["tool"].(string)
		stepParams, _ := stepMap["parameters"].(map[string]interface{})

		// Execute step
		stepResult, err := etr.ExecuteTool(ctx, sessionID, toolName, stepParams)
		if err != nil || !stepResult.Success {
			workflowState.Status = "error"
			return &ToolResult{
				Success: false,
				Error:   fmt.Sprintf("workflow step %d failed: %v", i, err),
			}, nil
		}

		results = append(results, stepResult.Data)
		workflowState.StepResults = append(workflowState.StepResults, stepResult.Data)
		workflowState.CurrentStep = i + 1
	}

	workflowState.Status = "completed"

	return &ToolResult{
		Success: true,
		Data:    results,
		Metadata: map[string]interface{}{
			"workflow_id":     workflowID,
			"workflow_type":   workflowType,
			"steps_completed": len(results),
		},
	}, nil
}

// handleContextManager handles context management
func (etr *EnhancedToolRegistry) handleContextManager(ctx context.Context, sessionID string, params map[string]interface{}) (*ToolResult, error) {
	action, _ := params["action"].(string)

	session, err := etr.sessionManager.GetSession(sessionID)
	if err != nil {
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("session not found: %s", sessionID),
		}, nil
	}

	switch action {
	case "get":
		return &ToolResult{
			Success: true,
			Data:    session.Context,
		}, nil

	case "set":
		key, _ := params["key"].(string)
		value := params["value"]
		if key != "" {
			updates := map[string]interface{}{key: value}
			err := etr.sessionManager.UpdateSessionContext(sessionID, updates)
			if err != nil {
				return &ToolResult{
					Success: false,
					Error:   err.Error(),
				}, nil
			}
		}
		return &ToolResult{
			Success: true,
			Data:    "context updated",
		}, nil

	case "clear":
		session.Context = &SessionContext{
			UserPreferences: make(map[string]interface{}),
			ActiveDocuments: make([]string, 0),
			SearchHistory:   make([]SearchHistoryItem, 0),
			ContextWindow:   make([]ContextItem, 0),
		}
		return &ToolResult{
			Success: true,
			Data:    "context cleared",
		}, nil

	case "search":
		query, _ := params["query"].(string)
		memory, err := etr.sessionManager.GetRelevantMemory(sessionID, query, 10)
		if err != nil {
			return &ToolResult{
				Success: false,
				Error:   err.Error(),
			}, nil
		}
		return &ToolResult{
			Success: true,
			Data:    memory,
		}, nil

	default:
		return &ToolResult{
			Success: false,
			Error:   fmt.Sprintf("unsupported action: %s", action),
		}, nil
	}
}

// streamSearchResults streams search results
func (etr *EnhancedToolRegistry) streamSearchResults(stream chan interface{}, results []interface{}) {
	defer close(stream)

	for i, result := range results {
		response := StreamingResponse{
			Type:      "data",
			Data:      result,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"index": i,
				"total": len(results),
			},
		}
		stream <- response
		time.Sleep(100 * time.Millisecond) // Simulate streaming delay
	}

	// Send completion signal
	stream <- StreamingResponse{
		Type:      "complete",
		Data:      "search completed",
		Timestamp: time.Now(),
	}
}

// GetToolList returns a list of available tools
func (etr *EnhancedToolRegistry) GetToolList() map[string]*EnhancedTool {
	return etr.tools
}

// GetToolsByCategory returns tools filtered by category
func (etr *EnhancedToolRegistry) GetToolsByCategory(category string) []*EnhancedTool {
	var tools []*EnhancedTool
	for _, tool := range etr.tools {
		if tool.Category == category {
			tools = append(tools, tool)
		}
	}
	return tools
}
