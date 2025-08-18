package langgraph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/langchain/memory"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StartNode represents the entry point of a graph
type StartNode struct {
	id         string
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewStartNode creates a new start node
func NewStartNode(id string, logger *logrus.Logger) *StartNode {
	return &StartNode{
		id:         id,
		inputKeys:  []string{},
		outputKeys: []string{"started"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.start"),
	}
}

func (n *StartNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "start_node.execute")
	defer span.End()

	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	newState["started"] = true
	newState["start_time"] = time.Now()

	return newState, nil
}

func (n *StartNode) GetID() string           { return n.id }
func (n *StartNode) GetType() string         { return "start" }
func (n *StartNode) GetInputKeys() []string  { return n.inputKeys }
func (n *StartNode) GetOutputKeys() []string { return n.outputKeys }
func (n *StartNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	return nil
}

// EndNode represents the exit point of a graph
type EndNode struct {
	id         string
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewEndNode creates a new end node
func NewEndNode(id string, logger *logrus.Logger) *EndNode {
	return &EndNode{
		id:         id,
		inputKeys:  []string{},
		outputKeys: []string{"completed"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.end"),
	}
}

func (n *EndNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "end_node.execute")
	defer span.End()

	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	newState["completed"] = true
	newState["end_time"] = time.Now()

	if startTime, exists := state["start_time"]; exists {
		if st, ok := startTime.(time.Time); ok {
			newState["duration"] = time.Since(st)
		}
	}

	return newState, nil
}

func (n *EndNode) GetID() string           { return n.id }
func (n *EndNode) GetType() string         { return "end" }
func (n *EndNode) GetInputKeys() []string  { return n.inputKeys }
func (n *EndNode) GetOutputKeys() []string { return n.outputKeys }
func (n *EndNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	return nil
}

// ParallelNode executes multiple nodes in parallel
type ParallelNode struct {
	id         string
	nodes      []Node
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewParallelNode creates a new parallel node
func NewParallelNode(id string, nodes []Node, logger *logrus.Logger) *ParallelNode {
	return &ParallelNode{
		id:         id,
		nodes:      nodes,
		inputKeys:  []string{},
		outputKeys: []string{"parallel_results"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.parallel"),
	}
}

func (n *ParallelNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "parallel_node.execute")
	defer span.End()

	type result struct {
		nodeID string
		state  GraphState
		err    error
	}

	resultCh := make(chan result, len(n.nodes))
	var wg sync.WaitGroup

	// Execute all nodes in parallel
	for _, node := range n.nodes {
		wg.Add(1)
		go func(n Node) {
			defer wg.Done()

			nodeState, err := n.Execute(ctx, state)
			resultCh <- result{
				nodeID: n.GetID(),
				state:  nodeState,
				err:    err,
			}
		}(node)
	}

	// Wait for all nodes to complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make(map[string]interface{})
	var lastError error

	for res := range resultCh {
		if res.err != nil {
			lastError = res.err
			n.logger.WithError(res.err).WithField("node_id", res.nodeID).Error("Parallel node execution failed")
		} else {
			results[res.nodeID] = res.state
		}
	}

	if lastError != nil {
		return state, fmt.Errorf("parallel execution failed: %w", lastError)
	}

	// Merge results into new state
	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}
	newState["parallel_results"] = results

	return newState, nil
}

func (n *ParallelNode) GetID() string           { return n.id }
func (n *ParallelNode) GetType() string         { return "parallel" }
func (n *ParallelNode) GetInputKeys() []string  { return n.inputKeys }
func (n *ParallelNode) GetOutputKeys() []string { return n.outputKeys }
func (n *ParallelNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if len(n.nodes) == 0 {
		return fmt.Errorf("parallel node must have at least one child node")
	}
	return nil
}

// LoopNode executes a node repeatedly based on a condition
type LoopNode struct {
	id         string
	node       Node
	condition  Condition
	maxIter    int
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewLoopNode creates a new loop node
func NewLoopNode(id string, node Node, condition Condition, maxIter int, logger *logrus.Logger) *LoopNode {
	return &LoopNode{
		id:         id,
		node:       node,
		condition:  condition,
		maxIter:    maxIter,
		inputKeys:  []string{},
		outputKeys: []string{"loop_results", "iterations"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.loop"),
	}
}

func (n *LoopNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "loop_node.execute")
	defer span.End()

	currentState := make(GraphState)
	for k, v := range state {
		currentState[k] = v
	}

	var results []GraphState
	iterations := 0

	for iterations < n.maxIter {
		// Check condition
		shouldContinue, err := n.condition.Evaluate(ctx, currentState)
		if err != nil {
			return state, fmt.Errorf("loop condition evaluation failed: %w", err)
		}

		if !shouldContinue {
			break
		}

		// Execute node
		newState, err := n.node.Execute(ctx, currentState)
		if err != nil {
			return state, fmt.Errorf("loop node execution failed at iteration %d: %w", iterations, err)
		}

		results = append(results, newState)
		currentState = newState
		iterations++
	}

	// Update state with results
	finalState := make(GraphState)
	for k, v := range currentState {
		finalState[k] = v
	}
	finalState["loop_results"] = results
	finalState["iterations"] = iterations

	return finalState, nil
}

func (n *LoopNode) GetID() string           { return n.id }
func (n *LoopNode) GetType() string         { return "loop" }
func (n *LoopNode) GetInputKeys() []string  { return n.inputKeys }
func (n *LoopNode) GetOutputKeys() []string { return n.outputKeys }
func (n *LoopNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.node == nil {
		return fmt.Errorf("loop node cannot be nil")
	}
	if n.condition == nil {
		return fmt.Errorf("loop condition cannot be nil")
	}
	if n.maxIter <= 0 {
		return fmt.Errorf("max iterations must be positive")
	}
	return nil
}

// SwitchNode routes to different nodes based on conditions
type SwitchNode struct {
	id          string
	cases       []SwitchCase
	defaultNode Node
	inputKeys   []string
	outputKeys  []string
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// SwitchCase represents a case in a switch node
type SwitchCase struct {
	Condition Condition
	Node      Node
}

// NewSwitchNode creates a new switch node
func NewSwitchNode(id string, cases []SwitchCase, defaultNode Node, logger *logrus.Logger) *SwitchNode {
	return &SwitchNode{
		id:          id,
		cases:       cases,
		defaultNode: defaultNode,
		inputKeys:   []string{},
		outputKeys:  []string{"switch_result", "selected_case"},
		logger:      logger,
		tracer:      otel.Tracer("langgraph.nodes.switch"),
	}
}

func (n *SwitchNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "switch_node.execute")
	defer span.End()

	// Evaluate cases in order
	for i, switchCase := range n.cases {
		matches, err := switchCase.Condition.Evaluate(ctx, state)
		if err != nil {
			n.logger.WithError(err).WithField("case_index", i).Error("Switch case condition evaluation failed")
			continue
		}

		if matches {
			// Execute the matching node
			result, err := switchCase.Node.Execute(ctx, state)
			if err != nil {
				return state, fmt.Errorf("switch case %d execution failed: %w", i, err)
			}

			// Add switch metadata
			newState := make(GraphState)
			for k, v := range result {
				newState[k] = v
			}
			newState["selected_case"] = i
			newState["switch_result"] = result

			return newState, nil
		}
	}

	// No case matched, use default
	if n.defaultNode != nil {
		result, err := n.defaultNode.Execute(ctx, state)
		if err != nil {
			return state, fmt.Errorf("switch default node execution failed: %w", err)
		}

		// Add switch metadata
		newState := make(GraphState)
		for k, v := range result {
			newState[k] = v
		}
		newState["selected_case"] = -1 // Indicates default
		newState["switch_result"] = result

		return newState, nil
	}

	// No matching case and no default
	return state, fmt.Errorf("no matching case found and no default node provided")
}

func (n *SwitchNode) GetID() string           { return n.id }
func (n *SwitchNode) GetType() string         { return "switch" }
func (n *SwitchNode) GetInputKeys() []string  { return n.inputKeys }
func (n *SwitchNode) GetOutputKeys() []string { return n.outputKeys }
func (n *SwitchNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if len(n.cases) == 0 && n.defaultNode == nil {
		return fmt.Errorf("switch node must have at least one case or a default node")
	}
	return nil
}

// MemoryNode provides memory operations
type MemoryNode struct {
	id           string
	memorySystem memory.Memory
	operation    string // "read", "write", "clear"
	inputKeys    []string
	outputKeys   []string
	logger       *logrus.Logger
	tracer       trace.Tracer
}

// NewMemoryNode creates a new memory node
func NewMemoryNode(id string, memorySystem memory.Memory, operation string, logger *logrus.Logger) *MemoryNode {
	return &MemoryNode{
		id:           id,
		memorySystem: memorySystem,
		operation:    operation,
		inputKeys:    []string{"memory_input"},
		outputKeys:   []string{"memory_output"},
		logger:       logger,
		tracer:       otel.Tracer("langgraph.nodes.memory"),
	}
}

func (n *MemoryNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "memory_node.execute")
	defer span.End()

	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	switch n.operation {
	case "read":
		variables, err := n.memorySystem.LoadMemoryVariables(ctx, state)
		if err != nil {
			return state, fmt.Errorf("memory read failed: %w", err)
		}
		newState["memory_output"] = variables

	case "write":
		input := make(map[string]interface{})
		output := make(map[string]interface{})

		if memInput, exists := state["memory_input"]; exists {
			if inputMap, ok := memInput.(map[string]interface{}); ok {
				input = inputMap
			}
		}

		if memOutput, exists := state["memory_output"]; exists {
			if outputMap, ok := memOutput.(map[string]interface{}); ok {
				output = outputMap
			}
		}

		if err := n.memorySystem.SaveContext(ctx, input, output); err != nil {
			return state, fmt.Errorf("memory write failed: %w", err)
		}

	case "clear":
		if err := n.memorySystem.Clear(ctx); err != nil {
			return state, fmt.Errorf("memory clear failed: %w", err)
		}

	default:
		return state, fmt.Errorf("unsupported memory operation: %s", n.operation)
	}

	return newState, nil
}

func (n *MemoryNode) GetID() string           { return n.id }
func (n *MemoryNode) GetType() string         { return "memory" }
func (n *MemoryNode) GetInputKeys() []string  { return n.inputKeys }
func (n *MemoryNode) GetOutputKeys() []string { return n.outputKeys }
func (n *MemoryNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.memorySystem == nil {
		return fmt.Errorf("memory system cannot be nil")
	}
	validOps := map[string]bool{"read": true, "write": true, "clear": true}
	if !validOps[n.operation] {
		return fmt.Errorf("invalid memory operation: %s", n.operation)
	}
	return nil
}
