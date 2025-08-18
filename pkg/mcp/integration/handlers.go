package integration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aios/aios/pkg/ai"
	"github.com/aios/aios/pkg/langchain/chains"
	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langgraph"
	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/aios/aios/pkg/mcp/resources"
	"github.com/aios/aios/pkg/mcp/tools"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ToolsHandler handles MCP tools requests
type ToolsHandler struct {
	toolManager tools.ToolManager
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewToolsHandler creates a new tools handler
func NewToolsHandler(toolManager tools.ToolManager, logger *logrus.Logger) *ToolsHandler {
	return &ToolsHandler{
		toolManager: toolManager,
		logger:      logger,
		tracer:      otel.Tracer("mcp.handlers.tools"),
	}
}

// HandleRequest handles tools requests
func (h *ToolsHandler) HandleRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := h.tracer.Start(ctx, "tools_handler.handle_request")
	defer span.End()

	switch request.GetMethod() {
	case protocol.MethodListTools:
		return h.handleListTools(ctx, request)
	case protocol.MethodCallTool:
		return h.handleCallTool(ctx, request)
	default:
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeMethodNotFound,
			fmt.Sprintf("method not supported: %s", request.GetMethod()),
			nil,
		)
	}
}

// HandleNotification handles tools notifications
func (h *ToolsHandler) HandleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	// Tools don't typically handle notifications
	return nil
}

// GetSupportedMethods returns supported methods
func (h *ToolsHandler) GetSupportedMethods() []string {
	return []string{protocol.MethodListTools, protocol.MethodCallTool}
}

func (h *ToolsHandler) handleListTools(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params protocol.ListToolsParams
	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid list tools params",
			nil,
		)
	}

	tools := h.toolManager.ListTools()

	result := protocol.ListToolsResult{
		Tools: tools,
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

func (h *ToolsHandler) handleCallTool(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params protocol.CallToolParams
	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid call tool params",
			nil,
		)
	}

	result, err := h.toolManager.CallTool(ctx, params.Name, params.Arguments)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			err.Error(),
			nil,
		)
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

// ResourcesHandler handles MCP resources requests
type ResourcesHandler struct {
	resourceManager resources.ResourceManager
	logger          *logrus.Logger
	tracer          trace.Tracer
}

// NewResourcesHandler creates a new resources handler
func NewResourcesHandler(resourceManager resources.ResourceManager, logger *logrus.Logger) *ResourcesHandler {
	return &ResourcesHandler{
		resourceManager: resourceManager,
		logger:          logger,
		tracer:          otel.Tracer("mcp.handlers.resources"),
	}
}

// HandleRequest handles resources requests
func (h *ResourcesHandler) HandleRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := h.tracer.Start(ctx, "resources_handler.handle_request")
	defer span.End()

	switch request.GetMethod() {
	case protocol.MethodListResources:
		return h.handleListResources(ctx, request)
	case protocol.MethodReadResource:
		return h.handleReadResource(ctx, request)
	default:
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeMethodNotFound,
			fmt.Sprintf("method not supported: %s", request.GetMethod()),
			nil,
		)
	}
}

// HandleNotification handles resources notifications
func (h *ResourcesHandler) HandleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	// Resources don't typically handle notifications
	return nil
}

// GetSupportedMethods returns supported methods
func (h *ResourcesHandler) GetSupportedMethods() []string {
	return []string{protocol.MethodListResources, protocol.MethodReadResource}
}

func (h *ResourcesHandler) handleListResources(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params protocol.ListResourcesParams
	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid list resources params",
			nil,
		)
	}

	result, err := h.resourceManager.ListResources(params.Cursor, 100) // Default limit
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			err.Error(),
			nil,
		)
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

func (h *ResourcesHandler) handleReadResource(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params protocol.ReadResourceParams
	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid read resource params",
			nil,
		)
	}

	result, err := h.resourceManager.ReadResource(ctx, params.URI)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			err.Error(),
			nil,
		)
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

// AIHandler handles AI-related requests
type AIHandler struct {
	aiOrchestrator ai.Orchestrator
	llmManager     llm.LLMManager
	chainManager   chains.ChainManager
	graphExecutor  langgraph.GraphExecutor
	logger         *logrus.Logger
	tracer         trace.Tracer
}

// NewAIHandler creates a new AI handler
func NewAIHandler(
	aiOrchestrator ai.Orchestrator,
	llmManager llm.LLMManager,
	chainManager chains.ChainManager,
	graphExecutor langgraph.GraphExecutor,
	logger *logrus.Logger,
) *AIHandler {
	return &AIHandler{
		aiOrchestrator: aiOrchestrator,
		llmManager:     llmManager,
		chainManager:   chainManager,
		graphExecutor:  graphExecutor,
		logger:         logger,
		tracer:         otel.Tracer("mcp.handlers.ai"),
	}
}

// HandleRequest handles AI requests
func (h *AIHandler) HandleRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := h.tracer.Start(ctx, "ai_handler.handle_request")
	defer span.End()

	switch request.GetMethod() {
	case "ai/complete":
		return h.handleAIComplete(ctx, request)
	case "ai/chain":
		return h.handleAIChain(ctx, request)
	case "ai/graph":
		return h.handleAIGraph(ctx, request)
	default:
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeMethodNotFound,
			fmt.Sprintf("method not supported: %s", request.GetMethod()),
			nil,
		)
	}
}

// HandleNotification handles AI notifications
func (h *AIHandler) HandleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	// AI handler doesn't typically handle notifications
	return nil
}

// GetSupportedMethods returns supported methods
func (h *AIHandler) GetSupportedMethods() []string {
	return []string{"ai/complete", "ai/chain", "ai/graph"}
}

func (h *AIHandler) handleAIComplete(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params struct {
		Model    string                 `json:"model"`
		Messages []llm.Message          `json:"messages"`
		Options  map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid AI complete params",
			nil,
		)
	}

	// Get LLM instance
	llmInstance, err := h.llmManager.GetLLM(params.Model)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			fmt.Sprintf("failed to get LLM: %v", err),
			nil,
		)
	}

	// Create completion request
	completionReq := &llm.CompletionRequest{
		Messages: params.Messages,
	}

	// Execute completion
	response, err := llmInstance.Complete(ctx, completionReq)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			fmt.Sprintf("completion failed: %v", err),
			nil,
		)
	}

	result := map[string]interface{}{
		"content": response.Content,
		"usage":   response.Usage,
		"model":   response.Model,
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

func (h *AIHandler) handleAIChain(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params struct {
		ChainName string                 `json:"chain_name"`
		Input     map[string]interface{} `json:"input"`
		Options   map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid AI chain params",
			nil,
		)
	}

	// Get chain
	chain, err := h.chainManager.GetChain(params.ChainName)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			fmt.Sprintf("failed to get chain: %v", err),
			nil,
		)
	}

	// Execute chain
	output, err := chain.Run(ctx, params.Input)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			fmt.Sprintf("chain execution failed: %v", err),
			nil,
		)
	}

	result := map[string]interface{}{
		"output": output,
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}

func (h *AIHandler) handleAIGraph(ctx context.Context, request protocol.Request) (protocol.Response, error) {
	var params struct {
		GraphID      string                 `json:"graph_id"`
		InitialState map[string]interface{} `json:"initial_state"`
		Options      map[string]interface{} `json:"options,omitempty"`
	}

	if err := json.Unmarshal(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			"invalid AI graph params",
			nil,
		)
	}

	// For now, return a placeholder response
	// In a real implementation, you would:
	// 1. Get the graph from a graph registry
	// 2. Execute the graph with the initial state
	// 3. Return the execution result

	result := map[string]interface{}{
		"message":       "Graph execution not yet implemented",
		"graph_id":      params.GraphID,
		"initial_state": params.InitialState,
	}

	return protocol.NewResponse(request.GetRequestID(), result)
}
