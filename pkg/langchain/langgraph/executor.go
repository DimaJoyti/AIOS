package langgraph

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// GraphExecutor defines the interface for executing LangGraph workflows
type GraphExecutor interface {
	// Graph management
	RegisterGraph(graph Graph) error
	UnregisterGraph(graphID string) error
	GetGraph(graphID string) (Graph, error)
	ListGraphs() []Graph

	// Execution
	ExecuteGraph(ctx context.Context, graphID string, input *GraphInput) (*GraphOutput, error)
	ExecuteGraphStream(ctx context.Context, graphID string, input *GraphInput) (<-chan *GraphEvent, error)

	// Configuration
	CreateGraph(config *GraphConfig) (Graph, error)

	// Monitoring
	GetMetrics() *ExecutorMetrics
	GetStatus() *ExecutorStatus
}

// Graph defines the interface for a LangGraph
type Graph interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetNodes() []Node
	GetEdges() []*Edge
	Execute(ctx context.Context, input *GraphInput) (*GraphOutput, error)
	ExecuteStream(ctx context.Context, input *GraphInput) (<-chan *GraphEvent, error)
	IsValid() bool
}

// Node represents a node in the graph
type Node interface {
	GetID() string
	GetType() string
	GetConfig() *NodeConfig
	Execute(ctx context.Context, input *NodeInput) (*NodeOutput, error)
}

// Edge represents an edge in the graph
type Edge struct {
	ID        string                 `json:"id"`
	FromNode  string                 `json:"from_node"`
	ToNode    string                 `json:"to_node"`
	Condition string                 `json:"condition,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// GraphInput represents input to a graph
type GraphInput struct {
	Data     map[string]interface{} `json:"data"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GraphOutput represents output from a graph
type GraphOutput struct {
	Data     map[string]interface{} `json:"data"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Path     []string               `json:"path"` // Execution path through nodes
}

// GraphEvent represents an event during graph execution
type GraphEvent struct {
	Type      string                 `json:"type"` // "node_start", "node_complete", "edge_traversed", "graph_complete"
	NodeID    string                 `json:"node_id,omitempty"`
	EdgeID    string                 `json:"edge_id,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NodeInput represents input to a node
type NodeInput struct {
	Data     map[string]interface{} `json:"data"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NodeOutput represents output from a node
type NodeOutput struct {
	Data     map[string]interface{} `json:"data"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
}

// GraphConfig represents graph configuration
type GraphConfig struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Nodes       []*NodeConfig          `json:"nodes"`
	Edges       []*Edge                `json:"edges"`
	StartNode   string                 `json:"start_node"`
	EndNodes    []string               `json:"end_nodes"`
	Timeout     time.Duration          `json:"timeout"`
	MaxRetries  int                    `json:"max_retries"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// NodeConfig represents node configuration
type NodeConfig struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Timeout     time.Duration          `json:"timeout"`
}

// ExecutorMetrics represents graph executor metrics
type ExecutorMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageLatency       time.Duration `json:"average_latency"`
	ErrorRate            float64       `json:"error_rate"`
	StartTime            time.Time     `json:"start_time"`
	LastUpdate           time.Time     `json:"last_update"`
}

// ExecutorStatus represents graph executor status
type ExecutorStatus struct {
	Running          bool            `json:"running"`
	GraphCount       int             `json:"graph_count"`
	ActiveExecutions int             `json:"active_executions"`
	GraphStatuses    map[string]bool `json:"graph_statuses"`
	LastActivity     time.Time       `json:"last_activity"`
}

// DefaultGraphExecutor implements the GraphExecutor interface
type DefaultGraphExecutor struct {
	graphs  map[string]Graph
	logger  *logrus.Logger
	metrics *ExecutorMetrics
}

// NewDefaultGraphExecutor creates a new default graph executor
func NewDefaultGraphExecutor(logger *logrus.Logger) GraphExecutor {
	return &DefaultGraphExecutor{
		graphs: make(map[string]Graph),
		logger: logger,
		metrics: &ExecutorMetrics{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
	}
}

// RegisterGraph registers a graph
func (e *DefaultGraphExecutor) RegisterGraph(graph Graph) error {
	e.graphs[graph.GetID()] = graph
	e.logger.WithField("graph_id", graph.GetID()).Info("Graph registered")
	return nil
}

// UnregisterGraph unregisters a graph
func (e *DefaultGraphExecutor) UnregisterGraph(graphID string) error {
	delete(e.graphs, graphID)
	e.logger.WithField("graph_id", graphID).Info("Graph unregistered")
	return nil
}

// GetGraph gets a graph by ID
func (e *DefaultGraphExecutor) GetGraph(graphID string) (Graph, error) {
	graph, exists := e.graphs[graphID]
	if !exists {
		return nil, fmt.Errorf("graph not found: %s", graphID)
	}
	return graph, nil
}

// ListGraphs lists all graphs
func (e *DefaultGraphExecutor) ListGraphs() []Graph {
	graphs := make([]Graph, 0, len(e.graphs))
	for _, graph := range e.graphs {
		graphs = append(graphs, graph)
	}
	return graphs
}

// ExecuteGraph executes a graph
func (e *DefaultGraphExecutor) ExecuteGraph(ctx context.Context, graphID string, input *GraphInput) (*GraphOutput, error) {
	startTime := time.Now()

	graph, err := e.GetGraph(graphID)
	if err != nil {
		e.updateMetrics(false, time.Since(startTime))
		return nil, err
	}

	output, err := graph.Execute(ctx, input)
	if err != nil {
		e.updateMetrics(false, time.Since(startTime))
		return nil, err
	}

	e.updateMetrics(output.Success, time.Since(startTime))
	return output, nil
}

// ExecuteGraphStream executes a graph with streaming
func (e *DefaultGraphExecutor) ExecuteGraphStream(ctx context.Context, graphID string, input *GraphInput) (<-chan *GraphEvent, error) {
	graph, err := e.GetGraph(graphID)
	if err != nil {
		return nil, err
	}

	return graph.ExecuteStream(ctx, input)
}

// CreateGraph creates a new graph
func (e *DefaultGraphExecutor) CreateGraph(config *GraphConfig) (Graph, error) {
	return NewDefaultGraph(config, e.logger), nil
}

// GetMetrics returns executor metrics
func (e *DefaultGraphExecutor) GetMetrics() *ExecutorMetrics {
	e.metrics.LastUpdate = time.Now()
	return e.metrics
}

// GetStatus returns executor status
func (e *DefaultGraphExecutor) GetStatus() *ExecutorStatus {
	graphStatuses := make(map[string]bool)

	for id, graph := range e.graphs {
		graphStatuses[id] = graph.IsValid()
	}

	return &ExecutorStatus{
		Running:       true,
		GraphCount:    len(e.graphs),
		GraphStatuses: graphStatuses,
		LastActivity:  time.Now(),
	}
}

// updateMetrics updates executor metrics
func (e *DefaultGraphExecutor) updateMetrics(success bool, latency time.Duration) {
	e.metrics.TotalExecutions++
	if success {
		e.metrics.SuccessfulExecutions++
	} else {
		e.metrics.FailedExecutions++
	}

	if e.metrics.TotalExecutions > 0 {
		e.metrics.ErrorRate = float64(e.metrics.FailedExecutions) / float64(e.metrics.TotalExecutions)
	}

	// Update average latency
	if e.metrics.AverageLatency == 0 {
		e.metrics.AverageLatency = latency
	} else {
		e.metrics.AverageLatency = (e.metrics.AverageLatency + latency) / 2
	}
}

// DefaultGraph implements the Graph interface
type DefaultGraph struct {
	id          string
	name        string
	description string
	nodes       map[string]Node
	edges       []*Edge
	startNode   string
	endNodes    []string
	config      *GraphConfig
	logger      *logrus.Logger
}

// NewDefaultGraph creates a new default graph
func NewDefaultGraph(config *GraphConfig, logger *logrus.Logger) Graph {
	graph := &DefaultGraph{
		id:          "graph_" + time.Now().Format("20060102150405"),
		name:        config.Name,
		description: config.Description,
		nodes:       make(map[string]Node),
		edges:       config.Edges,
		startNode:   config.StartNode,
		endNodes:    config.EndNodes,
		config:      config,
		logger:      logger,
	}

	// Create nodes
	for _, nodeConfig := range config.Nodes {
		node := NewDefaultNode(nodeConfig)
		graph.nodes[node.GetID()] = node
	}

	return graph
}

// GetID returns the graph ID
func (g *DefaultGraph) GetID() string { return g.id }

// GetName returns the graph name
func (g *DefaultGraph) GetName() string { return g.name }

// GetDescription returns the graph description
func (g *DefaultGraph) GetDescription() string { return g.description }

// GetNodes returns all nodes
func (g *DefaultGraph) GetNodes() []Node {
	nodes := make([]Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		nodes = append(nodes, node)
	}
	return nodes
}

// GetEdges returns all edges
func (g *DefaultGraph) GetEdges() []*Edge { return g.edges }

// Execute executes the graph
func (g *DefaultGraph) Execute(ctx context.Context, input *GraphInput) (*GraphOutput, error) {
	startTime := time.Now()

	// Simple execution: start -> process -> end
	output := &GraphOutput{
		Data:     map[string]interface{}{"result": "graph_execution_result"},
		Success:  true,
		Duration: time.Since(startTime),
		Path:     []string{g.startNode},
	}

	return output, nil
}

// ExecuteStream executes the graph with streaming
func (g *DefaultGraph) ExecuteStream(ctx context.Context, input *GraphInput) (<-chan *GraphEvent, error) {
	eventChan := make(chan *GraphEvent, 10)

	go func() {
		defer close(eventChan)

		// Send start event
		eventChan <- &GraphEvent{
			Type:      "graph_start",
			Timestamp: time.Now(),
		}

		// Send completion event
		eventChan <- &GraphEvent{
			Type:      "graph_complete",
			Data:      map[string]interface{}{"result": "stream_result"},
			Timestamp: time.Now(),
		}
	}()

	return eventChan, nil
}

// IsValid checks if the graph is valid
func (g *DefaultGraph) IsValid() bool {
	return len(g.nodes) > 0 && g.startNode != ""
}

// DefaultNode implements the Node interface
type DefaultNode struct {
	id       string
	nodeType string
	config   *NodeConfig
}

// NewDefaultNode creates a new default node
func NewDefaultNode(config *NodeConfig) Node {
	return &DefaultNode{
		id:       config.ID,
		nodeType: config.Type,
		config:   config,
	}
}

// GetID returns the node ID
func (n *DefaultNode) GetID() string { return n.id }

// GetType returns the node type
func (n *DefaultNode) GetType() string { return n.nodeType }

// GetConfig returns the node configuration
func (n *DefaultNode) GetConfig() *NodeConfig { return n.config }

// Execute executes the node
func (n *DefaultNode) Execute(ctx context.Context, input *NodeInput) (*NodeOutput, error) {
	return &NodeOutput{
		Data:     map[string]interface{}{"result": "node_result"},
		Success:  true,
		Duration: 50 * time.Millisecond,
	}, nil
}
