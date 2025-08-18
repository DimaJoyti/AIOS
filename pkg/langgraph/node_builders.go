package langgraph

import (
	"context"
	"fmt"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langchain/memory"
	"github.com/aios/aios/pkg/langchain/prompts"
	"github.com/sirupsen/logrus"
)

// NodeBuilder provides a fluent interface for building nodes
type NodeBuilder struct {
	nodeType   string
	id         string
	inputKeys  []string
	outputKeys []string
	metadata   map[string]interface{}
	config     map[string]interface{}
	logger     *logrus.Logger
}

// NewNodeBuilder creates a new node builder
func NewNodeBuilder(logger *logrus.Logger) *NodeBuilder {
	return &NodeBuilder{
		inputKeys:  []string{},
		outputKeys: []string{},
		metadata:   make(map[string]interface{}),
		config:     make(map[string]interface{}),
		logger:     logger,
	}
}

// WithID sets the node ID
func (b *NodeBuilder) WithID(id string) *NodeBuilder {
	b.id = id
	return b
}

// WithType sets the node type
func (b *NodeBuilder) WithType(nodeType string) *NodeBuilder {
	b.nodeType = nodeType
	return b
}

// WithInputKeys sets the input keys
func (b *NodeBuilder) WithInputKeys(keys ...string) *NodeBuilder {
	b.inputKeys = keys
	return b
}

// WithOutputKeys sets the output keys
func (b *NodeBuilder) WithOutputKeys(keys ...string) *NodeBuilder {
	b.outputKeys = keys
	return b
}

// WithMetadata adds metadata
func (b *NodeBuilder) WithMetadata(key string, value interface{}) *NodeBuilder {
	b.metadata[key] = value
	return b
}

// WithConfig adds configuration
func (b *NodeBuilder) WithConfig(key string, value interface{}) *NodeBuilder {
	b.config[key] = value
	return b
}

// BuildStart creates a start node
func (b *NodeBuilder) BuildStart() (*StartNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	return NewStartNode(b.id, b.logger), nil
}

// BuildEnd creates an end node
func (b *NodeBuilder) BuildEnd() (*EndNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	return NewEndNode(b.id, b.logger), nil
}

// BuildLLM creates an LLM node
func (b *NodeBuilder) BuildLLM(llmInstance llm.LLM, promptTemplate prompts.PromptTemplate) (*LLMNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if llmInstance == nil {
		return nil, fmt.Errorf("LLM instance is required")
	}
	return NewLLMNode(b.id, llmInstance, promptTemplate, b.logger), nil
}

// BuildTool creates a tool node
func (b *NodeBuilder) BuildTool(tool Tool) (*ToolNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if tool == nil {
		return nil, fmt.Errorf("tool is required")
	}
	return NewToolNode(b.id, tool, b.logger), nil
}

// BuildConditional creates a conditional node
func (b *NodeBuilder) BuildConditional(condition Condition) (*ConditionalNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if condition == nil {
		return nil, fmt.Errorf("condition is required")
	}
	return NewConditionalNode(b.id, condition, b.logger), nil
}

// BuildFunction creates a function node
func (b *NodeBuilder) BuildFunction(fn func(context.Context, GraphState) (GraphState, error)) (*FunctionNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if fn == nil {
		return nil, fmt.Errorf("function is required")
	}

	node := NewFunctionNode(b.id, fn, b.logger)
	if len(b.inputKeys) > 0 {
		node.SetInputKeys(b.inputKeys)
	}
	if len(b.outputKeys) > 0 {
		node.SetOutputKeys(b.outputKeys)
	}

	return node, nil
}

// BuildPassthrough creates a passthrough node
func (b *NodeBuilder) BuildPassthrough() (*PassthroughNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	return NewPassthroughNode(b.id, b.logger), nil
}

// BuildParallel creates a parallel node
func (b *NodeBuilder) BuildParallel(nodes []Node) (*ParallelNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("at least one child node is required")
	}
	return NewParallelNode(b.id, nodes, b.logger), nil
}

// BuildLoop creates a loop node
func (b *NodeBuilder) BuildLoop(node Node, condition Condition, maxIter int) (*LoopNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if node == nil {
		return nil, fmt.Errorf("child node is required")
	}
	if condition == nil {
		return nil, fmt.Errorf("condition is required")
	}
	if maxIter <= 0 {
		maxIter = 100 // Default max iterations
	}
	return NewLoopNode(b.id, node, condition, maxIter, b.logger), nil
}

// BuildSwitch creates a switch node
func (b *NodeBuilder) BuildSwitch(cases []SwitchCase, defaultNode Node) (*SwitchNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if len(cases) == 0 && defaultNode == nil {
		return nil, fmt.Errorf("at least one case or default node is required")
	}
	return NewSwitchNode(b.id, cases, defaultNode, b.logger), nil
}

// BuildMemory creates a memory node
func (b *NodeBuilder) BuildMemory(memorySystem memory.Memory, operation string) (*MemoryNode, error) {
	if b.id == "" {
		return nil, fmt.Errorf("node ID is required")
	}
	if memorySystem == nil {
		return nil, fmt.Errorf("memory system is required")
	}
	if operation == "" {
		operation = "read" // Default operation
	}
	return NewMemoryNode(b.id, memorySystem, operation, b.logger), nil
}

// NodeFactory creates different types of nodes
type NodeFactory struct {
	logger *logrus.Logger
}

// NewNodeFactory creates a new node factory
func NewNodeFactory(logger *logrus.Logger) *NodeFactory {
	return &NodeFactory{
		logger: logger,
	}
}

// CreateNode creates a node of the specified type
func (f *NodeFactory) CreateNode(nodeType string, id string, options map[string]interface{}) (Node, error) {
	builder := NewNodeBuilder(f.logger).WithID(id).WithType(nodeType)

	// Apply common options
	if inputKeys, exists := options["input_keys"]; exists {
		if keys, ok := inputKeys.([]string); ok {
			builder.WithInputKeys(keys...)
		}
	}

	if outputKeys, exists := options["output_keys"]; exists {
		if keys, ok := outputKeys.([]string); ok {
			builder.WithOutputKeys(keys...)
		}
	}

	if metadata, exists := options["metadata"]; exists {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			for k, v := range metadataMap {
				builder.WithMetadata(k, v)
			}
		}
	}

	switch nodeType {
	case "start":
		return builder.BuildStart()
	case "end":
		return builder.BuildEnd()
	case "passthrough":
		return builder.BuildPassthrough()
	case "llm":
		llmInstance, llmExists := options["llm"]
		promptTemplate, promptExists := options["prompt"]

		if !llmExists {
			return nil, fmt.Errorf("LLM instance is required for LLM node")
		}

		llmObj, ok := llmInstance.(llm.LLM)
		if !ok {
			return nil, fmt.Errorf("invalid LLM instance type")
		}

		var promptObj prompts.PromptTemplate
		if promptExists {
			if prompt, ok := promptTemplate.(prompts.PromptTemplate); ok {
				promptObj = prompt
			}
		}

		return builder.BuildLLM(llmObj, promptObj)

	case "tool":
		toolInstance, exists := options["tool"]
		if !exists {
			return nil, fmt.Errorf("tool instance is required for tool node")
		}

		toolObj, ok := toolInstance.(Tool)
		if !ok {
			return nil, fmt.Errorf("invalid tool instance type")
		}

		return builder.BuildTool(toolObj)

	case "conditional":
		conditionInstance, exists := options["condition"]
		if !exists {
			return nil, fmt.Errorf("condition is required for conditional node")
		}

		conditionObj, ok := conditionInstance.(Condition)
		if !ok {
			return nil, fmt.Errorf("invalid condition type")
		}

		return builder.BuildConditional(conditionObj)

	case "function":
		functionInstance, exists := options["function"]
		if !exists {
			return nil, fmt.Errorf("function is required for function node")
		}

		functionObj, ok := functionInstance.(func(context.Context, GraphState) (GraphState, error))
		if !ok {
			return nil, fmt.Errorf("invalid function type")
		}

		return builder.BuildFunction(functionObj)

	case "parallel":
		nodesInstance, exists := options["nodes"]
		if !exists {
			return nil, fmt.Errorf("child nodes are required for parallel node")
		}

		nodesSlice, ok := nodesInstance.([]Node)
		if !ok {
			return nil, fmt.Errorf("invalid nodes type")
		}

		return builder.BuildParallel(nodesSlice)

	case "loop":
		nodeInstance, nodeExists := options["node"]
		conditionInstance, conditionExists := options["condition"]
		maxIterInstance, maxIterExists := options["max_iterations"]

		if !nodeExists {
			return nil, fmt.Errorf("child node is required for loop node")
		}
		if !conditionExists {
			return nil, fmt.Errorf("condition is required for loop node")
		}

		nodeObj, ok := nodeInstance.(Node)
		if !ok {
			return nil, fmt.Errorf("invalid node type")
		}

		conditionObj, ok := conditionInstance.(Condition)
		if !ok {
			return nil, fmt.Errorf("invalid condition type")
		}

		maxIter := 100 // Default
		if maxIterExists {
			if maxIterInt, ok := maxIterInstance.(int); ok {
				maxIter = maxIterInt
			}
		}

		return builder.BuildLoop(nodeObj, conditionObj, maxIter)

	case "switch":
		casesInstance, casesExists := options["cases"]
		defaultNodeInstance, defaultExists := options["default_node"]

		var cases []SwitchCase
		if casesExists {
			if casesSlice, ok := casesInstance.([]SwitchCase); ok {
				cases = casesSlice
			}
		}

		var defaultNode Node
		if defaultExists {
			if defaultNodeObj, ok := defaultNodeInstance.(Node); ok {
				defaultNode = defaultNodeObj
			}
		}

		return builder.BuildSwitch(cases, defaultNode)

	case "memory":
		memoryInstance, memoryExists := options["memory"]
		operationInstance, operationExists := options["operation"]

		if !memoryExists {
			return nil, fmt.Errorf("memory system is required for memory node")
		}

		memoryObj, ok := memoryInstance.(memory.Memory)
		if !ok {
			return nil, fmt.Errorf("invalid memory system type")
		}

		operation := "read" // Default
		if operationExists {
			if operationStr, ok := operationInstance.(string); ok {
				operation = operationStr
			}
		}

		return builder.BuildMemory(memoryObj, operation)

	default:
		return nil, fmt.Errorf("unsupported node type: %s", nodeType)
	}
}

// GraphBuilder provides a fluent interface for building graphs
type GraphBuilder struct {
	nodes  map[string]Node
	edges  []Edge
	router Router
	config *GraphConfig
	logger *logrus.Logger
}

// NewGraphBuilder creates a new graph builder
func NewGraphBuilder(logger *logrus.Logger) *GraphBuilder {
	return &GraphBuilder{
		nodes:  make(map[string]Node),
		edges:  make([]Edge, 0),
		config: &GraphConfig{},
		logger: logger,
	}
}

// AddNode adds a node to the graph
func (b *GraphBuilder) AddNode(node Node) *GraphBuilder {
	b.nodes[node.GetID()] = node
	return b
}

// AddEdge adds an edge to the graph
func (b *GraphBuilder) AddEdge(edge Edge) *GraphBuilder {
	b.edges = append(b.edges, edge)
	return b
}

// WithRouter sets the router
func (b *GraphBuilder) WithRouter(router Router) *GraphBuilder {
	b.router = router
	return b
}

// WithConfig sets the graph configuration
func (b *GraphBuilder) WithConfig(config *GraphConfig) *GraphBuilder {
	b.config = config
	return b
}

// Build creates the graph
func (b *GraphBuilder) Build() (Graph, error) {
	if len(b.nodes) == 0 {
		return nil, fmt.Errorf("graph must have at least one node")
	}

	// Create default router if none provided
	if b.router == nil {
		b.router = NewConditionalRouter(b.logger)
	}

	// Create default config if none provided
	if b.config == nil {
		b.config = &GraphConfig{
			ID:          "",
			Name:        "default-graph",
			Description: "Default graph configuration",
			Nodes:       []NodeConfig{},
			Edges:       []EdgeConfig{},
			EntryPoints: []string{},
			ExitPoints:  []string{},
			Metadata:    make(map[string]interface{}),
		}
	}

	// Create graph with ID
	graphID := b.config.ID
	if graphID == "" {
		graphID = "graph-" + fmt.Sprintf("%d", time.Now().UnixNano())
	}

	graph := NewGraph(graphID)

	// Add nodes
	for _, node := range b.nodes {
		if err := graph.AddNode(node); err != nil {
			return nil, fmt.Errorf("failed to add node %s: %w", node.GetID(), err)
		}
	}

	// Add edges
	for _, edge := range b.edges {
		if err := graph.AddEdge(edge); err != nil {
			return nil, fmt.Errorf("failed to add edge %s->%s: %w", edge.GetFrom(), edge.GetTo(), err)
		}
	}

	return graph, nil
}

// Quick builder methods for common patterns

// BuildLinearGraph creates a linear graph with the given nodes
func (f *NodeFactory) BuildLinearGraph(nodeIDs []string, nodeConfigs []map[string]interface{}) (Graph, error) {
	builder := NewGraphBuilder(f.logger)

	// Create nodes
	for i, nodeID := range nodeIDs {
		var config map[string]interface{}
		if i < len(nodeConfigs) {
			config = nodeConfigs[i]
		} else {
			config = make(map[string]interface{})
		}

		nodeType := "passthrough"
		if nt, exists := config["type"]; exists {
			if ntStr, ok := nt.(string); ok {
				nodeType = ntStr
			}
		}

		node, err := f.CreateNode(nodeType, nodeID, config)
		if err != nil {
			return nil, fmt.Errorf("failed to create node %s: %w", nodeID, err)
		}

		builder.AddNode(node)
	}

	// Create linear edges
	edgeFactory := NewEdgeFactory(f.logger)
	for i := 0; i < len(nodeIDs)-1; i++ {
		edge, err := edgeFactory.CreateEdge("default", nodeIDs[i], nodeIDs[i+1], nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge %s->%s: %w", nodeIDs[i], nodeIDs[i+1], err)
		}
		builder.AddEdge(edge)
	}

	return builder.Build()
}

// BuildParallelGraph creates a graph with parallel execution
func (f *NodeFactory) BuildParallelGraph(entryNode string, parallelNodes []string, exitNode string) (Graph, error) {
	builder := NewGraphBuilder(f.logger)
	edgeFactory := NewEdgeFactory(f.logger)

	// Create entry node
	entry, err := f.CreateNode("start", entryNode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create entry node: %w", err)
	}
	builder.AddNode(entry)

	// Create parallel nodes
	for _, nodeID := range parallelNodes {
		node, err := f.CreateNode("passthrough", nodeID, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create parallel node %s: %w", nodeID, err)
		}
		builder.AddNode(node)

		// Edge from entry to parallel node
		edge, err := edgeFactory.CreateEdge("parallel", entryNode, nodeID, map[string]interface{}{
			"max_concurrency": 1,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create edge to parallel node %s: %w", nodeID, err)
		}
		builder.AddEdge(edge)
	}

	// Create exit node
	exit, err := f.CreateNode("end", exitNode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create exit node: %w", err)
	}
	builder.AddNode(exit)

	// Edges from parallel nodes to exit
	for _, nodeID := range parallelNodes {
		edge, err := edgeFactory.CreateEdge("default", nodeID, exitNode, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create edge from parallel node %s: %w", nodeID, err)
		}
		builder.AddEdge(edge)
	}

	return builder.Build()
}
