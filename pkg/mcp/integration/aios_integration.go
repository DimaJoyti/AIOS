package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aios/aios/pkg/ai"
	"github.com/aios/aios/pkg/langchain/chains"
	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langchain/memory"
	"github.com/aios/aios/pkg/langgraph"
	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/aios/aios/pkg/mcp/resources"
	"github.com/aios/aios/pkg/mcp/server"
	"github.com/aios/aios/pkg/mcp/tools"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AIOSMCPIntegration integrates MCP with AIOS services
type AIOSMCPIntegration struct {
	mcpServer       *server.MCPServer
	aiOrchestrator  ai.Orchestrator
	llmManager      llm.LLMManager
	chainManager    chains.ChainManager
	memoryManager   memory.MemoryManager
	graphExecutor   langgraph.GraphExecutor
	toolManager     tools.ToolManager
	resourceManager resources.ResourceManager
	logger          *logrus.Logger
	tracer          trace.Tracer
}

// IntegrationConfig represents integration configuration
type IntegrationConfig struct {
	MCPServerConfig      *server.ServerConfig             `json:"mcp_server_config"`
	EnableAIOrchestrator bool                             `json:"enable_ai_orchestrator"`
	EnableLangchain      bool                             `json:"enable_langchain"`
	EnableLanggraph      bool                             `json:"enable_langgraph"`
	EnableTools          bool                             `json:"enable_tools"`
	EnableResources      bool                             `json:"enable_resources"`
	ToolsConfig          *ToolManagerConfig               `json:"tools_config"`
	ResourcesConfig      *resources.ResourceManagerConfig `json:"resources_config"`
	Metadata             map[string]interface{}           `json:"metadata"`
}

// ToolManagerConfig represents tool manager configuration
type ToolManagerConfig struct {
	EnableFileSystem bool                   `json:"enable_filesystem"`
	EnableGit        bool                   `json:"enable_git"`
	EnableBuild      bool                   `json:"enable_build"`
	EnableProcess    bool                   `json:"enable_process"`
	EnableNetwork    bool                   `json:"enable_network"`
	FileSystemPaths  []string               `json:"filesystem_paths"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewAIOSMCPIntegration creates a new AIOS MCP integration
func NewAIOSMCPIntegration(
	config *IntegrationConfig,
	aiOrchestrator ai.Orchestrator,
	llmManager llm.LLMManager,
	chainManager chains.ChainManager,
	memoryManager memory.MemoryManager,
	graphExecutor langgraph.GraphExecutor,
	logger *logrus.Logger,
) (*AIOSMCPIntegration, error) {

	integration := &AIOSMCPIntegration{
		aiOrchestrator: aiOrchestrator,
		llmManager:     llmManager,
		chainManager:   chainManager,
		memoryManager:  memoryManager,
		graphExecutor:  graphExecutor,
		logger:         logger,
		tracer:         otel.Tracer("mcp.integration.aios"),
	}

	// Create MCP server
	mcpServer, err := server.NewMCPServer(config.MCPServerConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}
	integration.mcpServer = mcpServer

	// Initialize components based on configuration
	if config.EnableTools {
		if err := integration.initializeTools(config.ToolsConfig); err != nil {
			return nil, fmt.Errorf("failed to initialize tools: %w", err)
		}
	}

	if config.EnableResources {
		if err := integration.initializeResources(config.ResourcesConfig); err != nil {
			return nil, fmt.Errorf("failed to initialize resources: %w", err)
		}
	}

	// Register MCP handlers
	if err := integration.registerHandlers(); err != nil {
		return nil, fmt.Errorf("failed to register handlers: %w", err)
	}

	return integration, nil
}

// Start starts the integration
func (i *AIOSMCPIntegration) Start(ctx context.Context) error {
	ctx, span := i.tracer.Start(ctx, "aios_mcp_integration.start")
	defer span.End()

	i.logger.Info("Starting AIOS MCP integration")

	// Start MCP server
	if err := i.mcpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	i.logger.Info("AIOS MCP integration started successfully")

	return nil
}

// Stop stops the integration
func (i *AIOSMCPIntegration) Stop(ctx context.Context) error {
	ctx, span := i.tracer.Start(ctx, "aios_mcp_integration.stop")
	defer span.End()

	i.logger.Info("Stopping AIOS MCP integration")

	// Stop MCP server
	if err := i.mcpServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop MCP server: %w", err)
	}

	i.logger.Info("AIOS MCP integration stopped successfully")

	return nil
}

// GetMCPServer returns the MCP server
func (i *AIOSMCPIntegration) GetMCPServer() *server.MCPServer {
	return i.mcpServer
}

// GetToolManager returns the tool manager
func (i *AIOSMCPIntegration) GetToolManager() tools.ToolManager {
	return i.toolManager
}

// GetResourceManager returns the resource manager
func (i *AIOSMCPIntegration) GetResourceManager() resources.ResourceManager {
	return i.resourceManager
}

// Helper methods

func (i *AIOSMCPIntegration) initializeTools(config *ToolManagerConfig) error {
	// Create tool manager
	toolManager := tools.NewToolManager(i.logger)
	i.toolManager = toolManager

	// Register tools based on configuration
	if config.EnableFileSystem {
		fileSystemTool := tools.NewFileSystemTool(".", config.FileSystemPaths, i.logger)
		if err := toolManager.RegisterTool(fileSystemTool); err != nil {
			return fmt.Errorf("failed to register filesystem tool: %w", err)
		}
	}

	if config.EnableGit {
		gitTool := tools.NewGitTool(i.logger)
		if err := toolManager.RegisterTool(gitTool); err != nil {
			return fmt.Errorf("failed to register git tool: %w", err)
		}
	}

	// Register additional tools as needed...

	return nil
}

func (i *AIOSMCPIntegration) initializeResources(config *resources.ResourceManagerConfig) error {
	// Create resource manager
	resourceManager, err := resources.NewResourceManager(config, i.logger)
	if err != nil {
		return fmt.Errorf("failed to create resource manager: %w", err)
	}
	i.resourceManager = resourceManager

	// Register default resources
	if err := i.registerDefaultResources(); err != nil {
		return fmt.Errorf("failed to register default resources: %w", err)
	}

	return nil
}

func (i *AIOSMCPIntegration) registerDefaultResources() error {
	// Register AIOS configuration as a resource
	configResource := &AIOSConfigResource{
		uri:         "aios://config",
		name:        "AIOS Configuration",
		description: "Current AIOS system configuration",
		logger:      i.logger,
	}

	if err := i.resourceManager.RegisterResource(configResource); err != nil {
		return fmt.Errorf("failed to register config resource: %w", err)
	}

	// Register LLM models as resources
	if i.llmManager != nil {
		llmResource := &LLMModelsResource{
			uri:         "aios://llm/models",
			name:        "LLM Models",
			description: "Available LLM models and their status",
			llmManager:  i.llmManager,
			logger:      i.logger,
		}

		if err := i.resourceManager.RegisterResource(llmResource); err != nil {
			return fmt.Errorf("failed to register LLM models resource: %w", err)
		}
	}

	// Register memory systems as resources
	if i.memoryManager != nil {
		memoryResource := &MemorySystemsResource{
			uri:           "aios://memory/systems",
			name:          "Memory Systems",
			description:   "Available memory systems and their status",
			memoryManager: i.memoryManager,
			logger:        i.logger,
		}

		if err := i.resourceManager.RegisterResource(memoryResource); err != nil {
			return fmt.Errorf("failed to register memory systems resource: %w", err)
		}
	}

	return nil
}

func (i *AIOSMCPIntegration) registerHandlers() error {
	// Register tools handler
	if i.toolManager != nil {
		toolsHandler := NewToolsHandler(i.toolManager, i.logger)
		if err := i.mcpServer.RegisterHandler(
			[]string{protocol.MethodListTools, protocol.MethodCallTool},
			toolsHandler,
		); err != nil {
			return fmt.Errorf("failed to register tools handler: %w", err)
		}
	}

	// Register resources handler
	if i.resourceManager != nil {
		resourcesHandler := NewResourcesHandler(i.resourceManager, i.logger)
		if err := i.mcpServer.RegisterHandler(
			[]string{protocol.MethodListResources, protocol.MethodReadResource},
			resourcesHandler,
		); err != nil {
			return fmt.Errorf("failed to register resources handler: %w", err)
		}
	}

	// Register AI orchestrator handler
	if i.aiOrchestrator != nil {
		aiHandler := NewAIHandler(i.aiOrchestrator, i.llmManager, i.chainManager, i.graphExecutor, i.logger)
		if err := i.mcpServer.RegisterHandler(
			[]string{"ai/complete", "ai/chain", "ai/graph"},
			aiHandler,
		); err != nil {
			return fmt.Errorf("failed to register AI handler: %w", err)
		}
	}

	return nil
}

// Resource implementations

// AIOSConfigResource provides access to AIOS configuration
type AIOSConfigResource struct {
	uri         string
	name        string
	description string
	logger      *logrus.Logger
}

func (r *AIOSConfigResource) GetURI() string                         { return r.uri }
func (r *AIOSConfigResource) GetName() string                        { return r.name }
func (r *AIOSConfigResource) GetDescription() string                 { return r.description }
func (r *AIOSConfigResource) GetMimeType() string                    { return "application/json" }
func (r *AIOSConfigResource) GetAnnotations() map[string]interface{} { return nil }
func (r *AIOSConfigResource) GetLastModified() time.Time             { return time.Now() }
func (r *AIOSConfigResource) GetSize() int64                         { return 0 }
func (r *AIOSConfigResource) IsWatchable() bool                      { return false }
func (r *AIOSConfigResource) Watch(ctx context.Context, callback resources.ResourceCallback) error {
	return nil
}
func (r *AIOSConfigResource) StopWatch() error                    { return nil }
func (r *AIOSConfigResource) Validate() error                     { return nil }
func (r *AIOSConfigResource) GetCategory() string                 { return "configuration" }
func (r *AIOSConfigResource) GetTags() []string                   { return []string{"aios", "config"} }
func (r *AIOSConfigResource) GetMetadata() map[string]interface{} { return nil }

func (r *AIOSConfigResource) ReadContent(ctx context.Context) ([]protocol.ResourceContent, error) {
	// Return current AIOS configuration
	config := map[string]interface{}{
		"version": "1.0.0",
		"status":  "running",
		"modules": []string{"langchain", "langgraph", "mcp"},
	}

	content, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}

	return []protocol.ResourceContent{
		{
			URI:      r.uri,
			MimeType: "application/json",
			Text:     string(content),
		},
	}, nil
}

// LLMModelsResource provides access to LLM models
type LLMModelsResource struct {
	uri         string
	name        string
	description string
	llmManager  llm.LLMManager
	logger      *logrus.Logger
}

func (r *LLMModelsResource) GetURI() string                         { return r.uri }
func (r *LLMModelsResource) GetName() string                        { return r.name }
func (r *LLMModelsResource) GetDescription() string                 { return r.description }
func (r *LLMModelsResource) GetMimeType() string                    { return "application/json" }
func (r *LLMModelsResource) GetAnnotations() map[string]interface{} { return nil }
func (r *LLMModelsResource) GetLastModified() time.Time             { return time.Now() }
func (r *LLMModelsResource) GetSize() int64                         { return 0 }
func (r *LLMModelsResource) IsWatchable() bool                      { return true }
func (r *LLMModelsResource) Watch(ctx context.Context, callback resources.ResourceCallback) error {
	return nil
}
func (r *LLMModelsResource) StopWatch() error                    { return nil }
func (r *LLMModelsResource) Validate() error                     { return nil }
func (r *LLMModelsResource) GetCategory() string                 { return "ai" }
func (r *LLMModelsResource) GetTags() []string                   { return []string{"llm", "models", "ai"} }
func (r *LLMModelsResource) GetMetadata() map[string]interface{} { return nil }

func (r *LLMModelsResource) ReadContent(ctx context.Context) ([]protocol.ResourceContent, error) {
	// Get health status of LLM models (as a proxy for available models)
	healthStatus := r.llmManager.GetHealthStatus(ctx)

	content, err := json.Marshal(map[string]interface{}{
		"models": healthStatus,
		"count":  len(healthStatus),
	})
	if err != nil {
		return nil, err
	}

	return []protocol.ResourceContent{
		{
			URI:      r.uri,
			MimeType: "application/json",
			Text:     string(content),
		},
	}, nil
}

// MemorySystemsResource provides access to memory systems
type MemorySystemsResource struct {
	uri           string
	name          string
	description   string
	memoryManager memory.MemoryManager
	logger        *logrus.Logger
}

func (r *MemorySystemsResource) GetURI() string                         { return r.uri }
func (r *MemorySystemsResource) GetName() string                        { return r.name }
func (r *MemorySystemsResource) GetDescription() string                 { return r.description }
func (r *MemorySystemsResource) GetMimeType() string                    { return "application/json" }
func (r *MemorySystemsResource) GetAnnotations() map[string]interface{} { return nil }
func (r *MemorySystemsResource) GetLastModified() time.Time             { return time.Now() }
func (r *MemorySystemsResource) GetSize() int64                         { return 0 }
func (r *MemorySystemsResource) IsWatchable() bool                      { return true }
func (r *MemorySystemsResource) Watch(ctx context.Context, callback resources.ResourceCallback) error {
	return nil
}
func (r *MemorySystemsResource) StopWatch() error                    { return nil }
func (r *MemorySystemsResource) Validate() error                     { return nil }
func (r *MemorySystemsResource) GetCategory() string                 { return "memory" }
func (r *MemorySystemsResource) GetTags() []string                   { return []string{"memory", "systems"} }
func (r *MemorySystemsResource) GetMetadata() map[string]interface{} { return nil }

func (r *MemorySystemsResource) ReadContent(ctx context.Context) ([]protocol.ResourceContent, error) {
	// Get memory system types
	memoryTypes := r.memoryManager.GetMemoryTypes()

	content, err := json.Marshal(map[string]interface{}{
		"memory_types": memoryTypes,
		"count":        len(memoryTypes),
	})
	if err != nil {
		return nil, err
	}

	return []protocol.ResourceContent{
		{
			URI:      r.uri,
			MimeType: "application/json",
			Text:     string(content),
		},
	}, nil
}
