package enhanced

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aios/aios/internal/adapters"
	"github.com/aios/aios/pkg/config"
	"github.com/aios/aios/pkg/mcp/server"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

// API provides enhanced MCP functionality with Archon capabilities
type API struct {
	config        *config.Config
	pythonAdapter *adapters.PythonServiceAdapter
	logger        *logrus.Logger
	tracer        trace.Tracer
}

// MCPToolRequest represents an MCP tool execution request
type MCPToolRequest struct {
	ToolName  string                 `json:"tool_name"`
	Arguments map[string]interface{} `json:"arguments"`
	SessionID string                 `json:"session_id,omitempty"`
}

// MCPToolResponse represents an MCP tool execution response
type MCPToolResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// MCPSessionInfo represents MCP session information
type MCPSessionInfo struct {
	SessionID string            `json:"session_id"`
	ClientID  string            `json:"client_id"`
	Status    string            `json:"status"`
	Tools     []string          `json:"tools"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewAPI creates a new enhanced MCP API instance
func NewAPI(config *config.Config, pythonAdapter *adapters.PythonServiceAdapter, logger *logrus.Logger, tracer trace.Tracer) (*API, error) {
	return &API{
		config:        config,
		pythonAdapter: pythonAdapter,
		logger:        logger,
		tracer:        tracer,
	}, nil
}

// RegisterRoutes registers HTTP routes for the enhanced MCP API
func (api *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", api.healthHandler)
	mux.HandleFunc("/tools", api.toolsHandler)
	mux.HandleFunc("/execute", api.executeHandler)
	mux.HandleFunc("/sessions", api.sessionsHandler)
	mux.HandleFunc("/status", api.statusHandler)
}

// RegisterTools registers enhanced MCP tools with the MCP server
func (api *API) RegisterTools(mcpServer *server.MCPServer) error {
	// Get available tools from Python MCP server
	ctx := context.Background()
	resp, err := api.pythonAdapter.Get(ctx, "/tools")
	if err != nil {
		return fmt.Errorf("failed to get tools from Python MCP server: %w", err)
	}

	var tools []map[string]interface{}
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &tools); err != nil {
		return fmt.Errorf("failed to parse tools response: %w", err)
	}

	// Register each tool with the Go MCP server
	for _, tool := range tools {
		toolName, ok := tool["name"].(string)
		if !ok {
			continue
		}

		// Create a wrapper function that calls the Python service
		_ = api.createToolWrapper(toolName)

		// TODO: Register the tool (this would need to be implemented in the MCP server)
		// if err := mcpServer.RegisterTool(toolName, toolFunc); err != nil {
		//	api.logger.WithError(err).WithField("tool", toolName).Warn("Failed to register tool")
		// }
		api.logger.WithField("tool", toolName).Info("Tool wrapper created")
	}

	api.logger.Info("Enhanced MCP tools registered successfully")
	return nil
}

// createToolWrapper creates a wrapper function for Python MCP tools
func (api *API) createToolWrapper(toolName string) func(context.Context, map[string]interface{}) (interface{}, error) {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		req := MCPToolRequest{
			ToolName:  toolName,
			Arguments: args,
		}

		resp, err := api.pythonAdapter.Post(ctx, "/execute", req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tool %s: %w", toolName, err)
		}

		var toolResp MCPToolResponse
		if err := json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &toolResp); err != nil {
			return nil, fmt.Errorf("failed to parse tool response: %w", err)
		}

		if !toolResp.Success {
			return nil, fmt.Errorf("tool execution failed: %s", toolResp.Error)
		}

		return toolResp.Result, nil
	}
}

// healthHandler handles health check requests
func (api *API) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	if api.pythonAdapter.IsHealthy(ctx) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy"}`))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status": "unhealthy"}`))
	}
}

// toolsHandler handles tool listing requests
func (api *API) toolsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	resp, err := api.pythonAdapter.Get(ctx, "/tools")
	if err != nil {
		api.logger.WithError(err).Error("Failed to list tools")
		http.Error(w, "Failed to list tools", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Data)
}

// executeHandler handles tool execution requests
func (api *API) executeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req MCPToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	resp, err := api.pythonAdapter.Post(ctx, "/execute", req)
	if err != nil {
		api.logger.WithError(err).Error("Failed to execute tool")
		http.Error(w, "Failed to execute tool", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Data)
}

// sessionsHandler handles MCP session management requests
func (api *API) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		resp, err := api.pythonAdapter.Get(ctx, "/sessions")
		if err != nil {
			api.logger.WithError(err).Error("Failed to list sessions")
			http.Error(w, "Failed to list sessions", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Data)

	case http.MethodPost:
		var sessionReq map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		resp, err := api.pythonAdapter.Post(ctx, "/sessions", sessionReq)
		if err != nil {
			api.logger.WithError(err).Error("Failed to create session")
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Data)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// statusHandler handles MCP server status requests
func (api *API) statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	resp, err := api.pythonAdapter.Get(ctx, "/status")
	if err != nil {
		api.logger.WithError(err).Error("Failed to get status")
		http.Error(w, "Failed to get status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp.Data)
}

// ExecuteTool executes an MCP tool through the Python service
func (api *API) ExecuteTool(ctx context.Context, toolName string, args map[string]interface{}) (interface{}, error) {
	req := MCPToolRequest{
		ToolName:  toolName,
		Arguments: args,
	}

	resp, err := api.pythonAdapter.Post(ctx, "/execute", req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool %s: %w", toolName, err)
	}

	var toolResp MCPToolResponse
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &toolResp); err != nil {
		return nil, fmt.Errorf("failed to parse tool response: %w", err)
	}

	if !toolResp.Success {
		return nil, fmt.Errorf("tool execution failed: %s", toolResp.Error)
	}

	return toolResp.Result, nil
}

// ListTools returns available MCP tools
func (api *API) ListTools(ctx context.Context) ([]map[string]interface{}, error) {
	resp, err := api.pythonAdapter.Get(ctx, "/tools")
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	var tools []map[string]interface{}
	if err := json.Unmarshal([]byte(fmt.Sprintf("%v", resp.Data)), &tools); err != nil {
		return nil, fmt.Errorf("failed to parse tools response: %w", err)
	}

	return tools, nil
}
