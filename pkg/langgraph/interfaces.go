package langgraph

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
)

// GraphState represents the state of a graph execution
type GraphState map[string]interface{}

// Node represents a node in the execution graph
type Node interface {
	// Execute executes the node with the given state
	Execute(ctx context.Context, state GraphState) (GraphState, error)

	// GetID returns the unique identifier of the node
	GetID() string

	// GetType returns the type of the node
	GetType() string

	// GetInputKeys returns the input keys this node expects
	GetInputKeys() []string

	// GetOutputKeys returns the output keys this node produces
	GetOutputKeys() []string

	// Validate validates the node configuration
	Validate() error
}

// Edge represents an edge between nodes in the graph
type Edge interface {
	// GetFrom returns the source node ID
	GetFrom() string

	// GetTo returns the target node ID
	GetTo() string

	// GetCondition returns the condition for this edge (nil for unconditional)
	GetCondition() Condition

	// GetWeight returns the weight of this edge (for routing decisions)
	GetWeight() float64

	// SetCondition sets the condition for this edge
	SetCondition(condition Condition)

	// SetWeight sets the weight for this edge
	SetWeight(weight float64)

	// GetMetadata returns edge metadata
	GetMetadata() map[string]interface{}

	// SetMetadata sets edge metadata
	SetMetadata(key string, value interface{})
}

// Condition represents a condition for conditional routing
type Condition interface {
	// Evaluate evaluates the condition with the given state
	Evaluate(ctx context.Context, state GraphState) (bool, error)

	// GetDescription returns a description of the condition
	GetDescription() string
}

// Graph represents an execution graph
type Graph interface {
	// AddNode adds a node to the graph
	AddNode(node Node) error

	// AddEdge adds an edge to the graph
	AddEdge(edge Edge) error

	// RemoveNode removes a node from the graph
	RemoveNode(nodeID string) error

	// RemoveEdge removes an edge from the graph
	RemoveEdge(from, to string) error

	// GetNode returns a node by ID
	GetNode(nodeID string) (Node, error)

	// GetNodes returns all nodes in the graph
	GetNodes() map[string]Node

	// GetEdges returns all edges from a node
	GetEdges(nodeID string) []Edge

	// GetEntryPoints returns the entry point nodes
	GetEntryPoints() []string

	// GetExitPoints returns the exit point nodes
	GetExitPoints() []string

	// Validate validates the graph structure
	Validate() error

	// Clone creates a copy of the graph
	Clone() Graph
}

// GraphExecutor executes graphs
type GraphExecutor interface {
	// Execute executes a graph with the given initial state
	Execute(ctx context.Context, graph Graph, initialState GraphState) (*ExecutionResult, error)

	// ExecuteWithCallback executes a graph with callback support
	ExecuteWithCallback(ctx context.Context, graph Graph, initialState GraphState, callback ExecutionCallback) (*ExecutionResult, error)

	// ExecuteStream executes a graph with streaming updates
	ExecuteStream(ctx context.Context, graph Graph, initialState GraphState) (<-chan ExecutionUpdate, error)

	// GetExecutionHistory returns the execution history
	GetExecutionHistory() []ExecutionRecord

	// GetMetrics returns execution metrics
	GetMetrics() ExecutionMetrics
}

// ExecutionResult represents the result of graph execution
type ExecutionResult struct {
	ID           string                 `json:"id"`
	FinalState   GraphState             `json:"final_state"`
	ExecutedPath []string               `json:"executed_path"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	NodeResults  map[string]NodeResult  `json:"node_results"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NodeResult represents the result of a single node execution
type NodeResult struct {
	NodeID    string        `json:"node_id"`
	NodeType  string        `json:"node_type"`
	Success   bool          `json:"success"`
	Error     string        `json:"error,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
	Input     GraphState    `json:"input"`
	Output    GraphState    `json:"output"`
}

// ExecutionUpdate represents a streaming update during execution
type ExecutionUpdate struct {
	Type        string                 `json:"type"` // "node_start", "node_complete", "node_error", "execution_complete"
	ExecutionID string                 `json:"execution_id"`
	NodeID      string                 `json:"node_id,omitempty"`
	State       GraphState             `json:"state,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionCallback defines callbacks for graph execution events
type ExecutionCallback interface {
	// OnExecutionStart is called when graph execution starts
	OnExecutionStart(ctx context.Context, graph Graph, initialState GraphState) error

	// OnExecutionEnd is called when graph execution ends
	OnExecutionEnd(ctx context.Context, result *ExecutionResult) error

	// OnNodeStart is called when node execution starts
	OnNodeStart(ctx context.Context, node Node, state GraphState) error

	// OnNodeEnd is called when node execution ends
	OnNodeEnd(ctx context.Context, node Node, result *NodeResult) error

	// OnNodeError is called when node execution fails
	OnNodeError(ctx context.Context, node Node, state GraphState, err error) error

	// OnStateUpdate is called when the graph state is updated
	OnStateUpdate(ctx context.Context, oldState, newState GraphState) error
}

// ExecutionRecord represents a record of graph execution
type ExecutionRecord struct {
	ID           string                 `json:"id"`
	GraphID      string                 `json:"graph_id"`
	InitialState GraphState             `json:"initial_state"`
	FinalState   GraphState             `json:"final_state"`
	ExecutedPath []string               `json:"executed_path"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	NodeCount    int                    `json:"node_count"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionMetrics represents metrics for graph execution
type ExecutionMetrics struct {
	TotalExecutions      int64                           `json:"total_executions"`
	SuccessfulExecutions int64                           `json:"successful_executions"`
	FailedExecutions     int64                           `json:"failed_executions"`
	AverageDuration      time.Duration                   `json:"average_duration"`
	TotalDuration        time.Duration                   `json:"total_duration"`
	NodeMetrics          map[string]NodeExecutionMetrics `json:"node_metrics"`
	GraphMetrics         map[string]ExecutionMetrics     `json:"graph_metrics"`
}

// NodeExecutionMetrics represents metrics for node execution
type NodeExecutionMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageDuration      time.Duration `json:"average_duration"`
	MinDuration          time.Duration `json:"min_duration"`
	MaxDuration          time.Duration `json:"max_duration"`
}

// GraphBuilderInterface provides a fluent interface for building graphs
type GraphBuilderInterface interface {
	// WithID sets the graph ID
	WithID(id string) GraphBuilderInterface

	// AddLLMNode adds an LLM node
	AddLLMNode(id string, llm llm.LLM, prompt string) GraphBuilderInterface

	// AddToolNode adds a tool node
	AddToolNode(id string, tool Tool) GraphBuilderInterface

	// AddConditionalNode adds a conditional node
	AddConditionalNode(id string, condition Condition) GraphBuilderInterface

	// AddCustomNode adds a custom node
	AddCustomNode(node Node) GraphBuilderInterface

	// AddEdge adds an edge between nodes
	AddEdge(from, to string) GraphBuilderInterface

	// AddConditionalEdge adds a conditional edge
	AddConditionalEdge(from, to string, condition Condition) GraphBuilderInterface

	// SetEntryPoint sets the entry point node
	SetEntryPoint(nodeID string) GraphBuilderInterface

	// SetExitPoint sets an exit point node
	SetExitPoint(nodeID string) GraphBuilderInterface

	// WithMetadata adds metadata
	WithMetadata(key string, value interface{}) GraphBuilderInterface

	// Build builds the graph
	Build() (Graph, error)
}

// Tool represents a tool that can be used in tool nodes
type Tool interface {
	// Execute executes the tool with the given input
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// GetName returns the tool name
	GetName() string

	// GetDescription returns the tool description
	GetDescription() string

	// GetInputSchema returns the input schema
	GetInputSchema() map[string]interface{}

	// GetOutputSchema returns the output schema
	GetOutputSchema() map[string]interface{}

	// Validate validates the tool configuration
	Validate() error
}

// GraphOptimizer optimizes graph execution
type GraphOptimizer interface {
	// OptimizeGraph optimizes a graph for better performance
	OptimizeGraph(ctx context.Context, graph Graph) (Graph, error)

	// AnalyzeGraph analyzes a graph and provides optimization suggestions
	AnalyzeGraph(ctx context.Context, graph Graph) (*GraphAnalysis, error)

	// FindBottlenecks identifies performance bottlenecks in the graph
	FindBottlenecks(ctx context.Context, graph Graph, metrics ExecutionMetrics) ([]Bottleneck, error)
}

// GraphAnalysis represents the analysis of a graph
type GraphAnalysis struct {
	GraphID        string                 `json:"graph_id"`
	NodeCount      int                    `json:"node_count"`
	EdgeCount      int                    `json:"edge_count"`
	Complexity     float64                `json:"complexity"`
	EstimatedCost  float64                `json:"estimated_cost"`
	CriticalPath   []string               `json:"critical_path"`
	Bottlenecks    []string               `json:"bottlenecks"`
	Suggestions    []string               `json:"suggestions"`
	OptimizedGraph Graph                  `json:"optimized_graph,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Bottleneck represents a performance bottleneck
type Bottleneck struct {
	NodeID      string        `json:"node_id"`
	Type        string        `json:"type"`
	Severity    float64       `json:"severity"`
	Description string        `json:"description"`
	Suggestion  string        `json:"suggestion"`
	Impact      time.Duration `json:"impact"`
}

// GraphRegistry manages registered graphs
type GraphRegistry interface {
	// RegisterGraph registers a graph with a name
	RegisterGraph(name string, graph Graph) error

	// GetGraph retrieves a graph by name
	GetGraph(name string) (Graph, error)

	// ListGraphs returns all registered graph names
	ListGraphs() []string

	// UnregisterGraph removes a graph from the registry
	UnregisterGraph(name string) error

	// Clone creates a copy of the registry
	Clone() GraphRegistry
}

// GraphComposer composes complex graphs from simpler ones
type GraphComposer interface {
	// ComposeSequential creates a sequential graph from nodes
	ComposeSequential(nodes []Node) (Graph, error)

	// ComposeParallel creates a parallel graph from nodes
	ComposeParallel(nodes []Node) (Graph, error)

	// ComposeConditional creates a conditional graph with branches
	ComposeConditional(condition Condition, trueGraph, falseGraph Graph) (Graph, error)

	// ComposeFromConfig creates a graph from configuration
	ComposeFromConfig(config GraphConfig) (Graph, error)

	// MergeGraphs merges multiple graphs into one
	MergeGraphs(graphs []Graph) (Graph, error)
}

// GraphConfig represents configuration for graph composition
type GraphConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []NodeConfig           `json:"nodes"`
	Edges       []EdgeConfig           `json:"edges"`
	EntryPoints []string               `json:"entry_points"`
	ExitPoints  []string               `json:"exit_points"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NodeConfig represents configuration for a node
type NodeConfig struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Config     map[string]interface{} `json:"config"`
	InputKeys  []string               `json:"input_keys,omitempty"`
	OutputKeys []string               `json:"output_keys,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// EdgeConfig represents configuration for an edge
type EdgeConfig struct {
	From      string                 `json:"from"`
	To        string                 `json:"to"`
	Condition string                 `json:"condition,omitempty"`
	Weight    float64                `json:"weight,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StateManager manages graph state persistence and recovery
type StateManager interface {
	// SaveState saves the current graph state
	SaveState(ctx context.Context, executionID string, state GraphState) error

	// LoadState loads a saved graph state
	LoadState(ctx context.Context, executionID string) (GraphState, error)

	// DeleteState deletes a saved graph state
	DeleteState(ctx context.Context, executionID string) error

	// ListStates lists all saved states
	ListStates(ctx context.Context) ([]string, error)

	// CreateCheckpoint creates a checkpoint of the current state
	CreateCheckpoint(ctx context.Context, executionID string, checkpointID string, state GraphState) error

	// RestoreCheckpoint restores state from a checkpoint
	RestoreCheckpoint(ctx context.Context, executionID string, checkpointID string) (GraphState, error)
}
