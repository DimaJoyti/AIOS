package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
)

// DefaultToolManager implements ToolManager
type DefaultToolManager struct {
	tools  map[string]MCPTool
	mu     sync.RWMutex
	logger *logrus.Logger
}

// NewToolManager creates a new tool manager
func NewToolManager(logger *logrus.Logger) ToolManager {
	return &DefaultToolManager{
		tools:  make(map[string]MCPTool),
		logger: logger,
	}
}

// RegisterTool registers a tool
func (tm *DefaultToolManager) RegisterTool(tool MCPTool) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if err := tool.Validate(); err != nil {
		return fmt.Errorf("tool validation failed: %w", err)
	}

	tm.tools[tool.GetName()] = tool
	tm.logger.WithField("tool_name", tool.GetName()).Info("Tool registered")
	return nil
}

// UnregisterTool unregisters a tool
func (tm *DefaultToolManager) UnregisterTool(name string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.tools, name)
	tm.logger.WithField("tool_name", name).Info("Tool unregistered")
	return nil
}

// GetTool retrieves a tool by name
func (tm *DefaultToolManager) GetTool(name string) (MCPTool, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tool, exists := tm.tools[name]
	if !exists {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	return tool, nil
}

// ListTools returns all registered tools
func (tm *DefaultToolManager) ListTools() []protocol.Tool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tools := make([]protocol.Tool, 0, len(tm.tools))
	for _, tool := range tm.tools {
		tools = append(tools, protocol.Tool{
			Name:        tool.GetName(),
			Description: tool.GetDescription(),
			InputSchema: tool.GetInputSchema(),
		})
	}

	return tools
}

// CallTool calls a tool with the given arguments
func (tm *DefaultToolManager) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	tool, err := tm.GetTool(name)
	if err != nil {
		return nil, err
	}

	return tool.Execute(ctx, arguments)
}

// ValidateTool validates a tool configuration
func (tm *DefaultToolManager) ValidateTool(tool MCPTool) error {
	return tool.Validate()
}

// RegisterToolProvider registers all tools from a tool provider
func (tm *DefaultToolManager) RegisterToolProvider(provider ToolProvider) error {
	tools := provider.GetTools()
	for _, tool := range tools {
		if err := tm.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register tool %s: %w", tool.GetName(), err)
		}
	}
	return nil
}

// FileSystemToolProvider provides filesystem tools
type FileSystemToolProvider struct {
	tools []MCPTool
}

// NewFileSystemTools creates filesystem tools
func NewFileSystemTools(logger *logrus.Logger) (ToolProvider, error) {
	tools := []MCPTool{
		NewListDirectoryTool(logger),
		NewReadFileTool(logger),
		NewWriteFileTool(logger),
		NewDeleteFileTool(logger),
	}

	return &FileSystemToolProvider{
		tools: tools,
	}, nil
}

// GetTools returns all filesystem tools
func (fsp *FileSystemToolProvider) GetTools() []MCPTool {
	return fsp.tools
}

// Simple tool implementations for testing

// ListDirectoryTool lists directory contents
type ListDirectoryTool struct {
	logger *logrus.Logger
}

func NewListDirectoryTool(logger *logrus.Logger) MCPTool {
	return &ListDirectoryTool{logger: logger}
}

func (t *ListDirectoryTool) GetName() string        { return "list_directory" }
func (t *ListDirectoryTool) GetDescription() string { return "List directory contents" }
func (t *ListDirectoryTool) GetCategory() string    { return "filesystem" }
func (t *ListDirectoryTool) GetTags() []string      { return []string{"filesystem", "directory"} }
func (t *ListDirectoryTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "Directory path to list",
			},
		},
		"required": []string{"path"},
	}
}
func (t *ListDirectoryTool) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"files": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}
}
func (t *ListDirectoryTool) GetTimeout() time.Duration { return 30 * time.Second }
func (t *ListDirectoryTool) IsAsync() bool             { return false }

func (t *ListDirectoryTool) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return &protocol.CallToolResult{
			IsError: true,
			Content: []protocol.ToolContent{
				{Type: "text", Text: "path argument is required and must be a string"},
			},
		}, nil
	}

	// Simple mock implementation
	return &protocol.CallToolResult{
		IsError: false,
		Content: []protocol.ToolContent{
			{Type: "text", Text: fmt.Sprintf("Listed directory: %s", path)},
		},
	}, nil
}

func (t *ListDirectoryTool) Validate() error {
	return nil
}

// ReadFileTool reads file contents
type ReadFileTool struct {
	logger *logrus.Logger
}

func NewReadFileTool(logger *logrus.Logger) MCPTool {
	return &ReadFileTool{logger: logger}
}

func (t *ReadFileTool) GetName() string           { return "read_file" }
func (t *ReadFileTool) GetDescription() string    { return "Read file contents" }
func (t *ReadFileTool) GetCategory() string       { return "filesystem" }
func (t *ReadFileTool) GetTags() []string         { return []string{"filesystem", "file"} }
func (t *ReadFileTool) GetTimeout() time.Duration { return 30 * time.Second }
func (t *ReadFileTool) IsAsync() bool             { return false }
func (t *ReadFileTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"path": map[string]interface{}{
				"type":        "string",
				"description": "File path to read",
			},
		},
		"required": []string{"path"},
	}
}
func (t *ReadFileTool) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"content": map[string]interface{}{
				"type": "string",
			},
		},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	path, ok := arguments["path"].(string)
	if !ok {
		return &protocol.CallToolResult{
			IsError: true,
			Content: []protocol.ToolContent{
				{Type: "text", Text: "path argument is required and must be a string"},
			},
		}, nil
	}

	return &protocol.CallToolResult{
		IsError: false,
		Content: []protocol.ToolContent{
			{Type: "text", Text: fmt.Sprintf("Read file: %s", path)},
		},
	}, nil
}

func (t *ReadFileTool) Validate() error {
	return nil
}

// Placeholder implementations for other tools
type WriteFileTool struct{ logger *logrus.Logger }

func NewWriteFileTool(logger *logrus.Logger) MCPTool             { return &WriteFileTool{logger} }
func (t *WriteFileTool) GetName() string                         { return "write_file" }
func (t *WriteFileTool) GetDescription() string                  { return "Write file contents" }
func (t *WriteFileTool) GetCategory() string                     { return "filesystem" }
func (t *WriteFileTool) GetTags() []string                       { return []string{"filesystem", "file"} }
func (t *WriteFileTool) GetTimeout() time.Duration               { return 30 * time.Second }
func (t *WriteFileTool) IsAsync() bool                           { return false }
func (t *WriteFileTool) GetInputSchema() map[string]interface{}  { return map[string]interface{}{} }
func (t *WriteFileTool) GetOutputSchema() map[string]interface{} { return map[string]interface{}{} }
func (t *WriteFileTool) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	return &protocol.CallToolResult{IsError: false, Content: []protocol.ToolContent{{Type: "text", Text: "File written"}}}, nil
}
func (t *WriteFileTool) Validate() error { return nil }

type DeleteFileTool struct{ logger *logrus.Logger }

func NewDeleteFileTool(logger *logrus.Logger) MCPTool             { return &DeleteFileTool{logger} }
func (t *DeleteFileTool) GetName() string                         { return "delete_file" }
func (t *DeleteFileTool) GetDescription() string                  { return "Delete file" }
func (t *DeleteFileTool) GetCategory() string                     { return "filesystem" }
func (t *DeleteFileTool) GetTags() []string                       { return []string{"filesystem", "file"} }
func (t *DeleteFileTool) GetTimeout() time.Duration               { return 30 * time.Second }
func (t *DeleteFileTool) IsAsync() bool                           { return false }
func (t *DeleteFileTool) GetInputSchema() map[string]interface{}  { return map[string]interface{}{} }
func (t *DeleteFileTool) GetOutputSchema() map[string]interface{} { return map[string]interface{}{} }
func (t *DeleteFileTool) Execute(ctx context.Context, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	return &protocol.CallToolResult{IsError: false, Content: []protocol.ToolContent{{Type: "text", Text: "File deleted"}}}, nil
}
func (t *DeleteFileTool) Validate() error { return nil }
