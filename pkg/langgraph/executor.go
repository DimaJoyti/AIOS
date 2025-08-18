package langgraph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultGraphExecutor implements the GraphExecutor interface
type DefaultGraphExecutor struct {
	logger           *logrus.Logger
	tracer           trace.Tracer
	executionHistory []ExecutionRecord
	metrics          ExecutionMetrics
	stateManager     StateManager
	mu               sync.RWMutex
}

// ExecutorConfig represents configuration for the graph executor
type ExecutorConfig struct {
	StateManager StateManager `json:"-"`
	MaxHistory   int          `json:"max_history,omitempty"`
}

// NewGraphExecutor creates a new graph executor
func NewGraphExecutor(config *ExecutorConfig, logger *logrus.Logger) GraphExecutor {
	if config.MaxHistory <= 0 {
		config.MaxHistory = 1000 // Default max history
	}

	return &DefaultGraphExecutor{
		logger:           logger,
		tracer:           otel.Tracer("langgraph.executor"),
		executionHistory: make([]ExecutionRecord, 0),
		metrics: ExecutionMetrics{
			NodeMetrics:  make(map[string]NodeExecutionMetrics),
			GraphMetrics: make(map[string]ExecutionMetrics),
		},
		stateManager: config.StateManager,
	}
}

// Execute executes a graph with the given initial state
func (e *DefaultGraphExecutor) Execute(ctx context.Context, graph Graph, initialState GraphState) (*ExecutionResult, error) {
	return e.ExecuteWithCallback(ctx, graph, initialState, nil)
}

// ExecuteWithCallback executes a graph with callback support
func (e *DefaultGraphExecutor) ExecuteWithCallback(ctx context.Context, graph Graph, initialState GraphState, callback ExecutionCallback) (*ExecutionResult, error) {
	ctx, span := e.tracer.Start(ctx, "graph_executor.execute")
	defer span.End()

	executionID := uuid.New().String()
	startTime := time.Now()

	span.SetAttributes(
		attribute.String("execution.id", executionID),
		attribute.Int("graph.node_count", len(graph.GetNodes())),
	)

	e.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"node_count":   len(graph.GetNodes()),
	}).Info("Starting graph execution")

	// Validate graph
	if err := graph.Validate(); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("graph validation failed: %w", err)
	}

	// Initialize execution result
	result := &ExecutionResult{
		ID:           executionID,
		FinalState:   make(GraphState),
		ExecutedPath: make([]string, 0),
		Success:      false,
		StartTime:    startTime,
		NodeResults:  make(map[string]NodeResult),
		Metadata:     make(map[string]interface{}),
	}

	// Copy initial state
	currentState := make(GraphState)
	for k, v := range initialState {
		currentState[k] = v
	}

	// Call execution start callback
	if callback != nil {
		if err := callback.OnExecutionStart(ctx, graph, initialState); err != nil {
			e.logger.WithError(err).Error("Execution start callback failed")
		}
	}

	// Save initial state if state manager is available
	if e.stateManager != nil {
		if err := e.stateManager.SaveState(ctx, executionID, currentState); err != nil {
			e.logger.WithError(err).Error("Failed to save initial state")
		}
	}

	// Execute the graph
	var executionError error
	defer func() {
		// Finalize result
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		result.Success = executionError == nil

		if executionError != nil {
			result.Error = executionError.Error()
			span.RecordError(executionError)
		}

		// Copy final state
		for k, v := range currentState {
			result.FinalState[k] = v
		}

		// Call execution end callback
		if callback != nil {
			if err := callback.OnExecutionEnd(ctx, result); err != nil {
				e.logger.WithError(err).Error("Execution end callback failed")
			}
		}

		// Record execution
		e.recordExecution(result)

		span.SetAttributes(
			attribute.Bool("execution.success", result.Success),
			attribute.Int64("execution.duration_ms", result.Duration.Milliseconds()),
			attribute.Int("execution.nodes_executed", len(result.ExecutedPath)),
		)

		e.logger.WithFields(logrus.Fields{
			"execution_id":   executionID,
			"success":        result.Success,
			"duration":       result.Duration,
			"nodes_executed": len(result.ExecutedPath),
		}).Info("Graph execution completed")
	}()

	// Get entry points
	entryPoints := graph.GetEntryPoints()
	if len(entryPoints) == 0 {
		executionError = fmt.Errorf("graph has no entry points")
		return result, executionError
	}

	// Execute starting from entry points
	visited := make(map[string]bool)

	// For simplicity, start with the first entry point
	// In a more sophisticated implementation, you might execute all entry points in parallel
	currentNode := entryPoints[0]

	for currentNode != "" {
		// Check if we've already visited this node (cycle detection)
		if visited[currentNode] {
			e.logger.WithField("node_id", currentNode).Warn("Cycle detected, stopping execution")
			break
		}

		// Get the node
		node, err := graph.GetNode(currentNode)
		if err != nil {
			executionError = fmt.Errorf("failed to get node %s: %w", currentNode, err)
			return result, executionError
		}

		// Execute the node
		nodeResult, newState, err := e.executeNode(ctx, node, currentState, callback)
		if err != nil {
			executionError = fmt.Errorf("node %s execution failed: %w", currentNode, err)
			return result, executionError
		}

		// Record node execution
		result.NodeResults[currentNode] = *nodeResult
		result.ExecutedPath = append(result.ExecutedPath, currentNode)
		visited[currentNode] = true

		// Update state
		oldState := make(GraphState)
		for k, v := range currentState {
			oldState[k] = v
		}

		for k, v := range newState {
			currentState[k] = v
		}

		// Call state update callback
		if callback != nil {
			if err := callback.OnStateUpdate(ctx, oldState, currentState); err != nil {
				e.logger.WithError(err).Error("State update callback failed")
			}
		}

		// Save state if state manager is available
		if e.stateManager != nil {
			if err := e.stateManager.SaveState(ctx, executionID, currentState); err != nil {
				e.logger.WithError(err).Error("Failed to save state")
			}
		}

		// Determine next node
		nextNode, err := e.getNextNode(ctx, graph, currentNode, currentState)
		if err != nil {
			executionError = fmt.Errorf("failed to determine next node: %w", err)
			return result, executionError
		}

		currentNode = nextNode
	}

	return result, nil
}

// ExecuteStream executes a graph with streaming updates
func (e *DefaultGraphExecutor) ExecuteStream(ctx context.Context, graph Graph, initialState GraphState) (<-chan ExecutionUpdate, error) {
	updateCh := make(chan ExecutionUpdate, 10)

	go func() {
		defer close(updateCh)

		callback := &streamingCallback{
			updateCh: updateCh,
		}

		result, err := e.ExecuteWithCallback(ctx, graph, initialState, callback)

		// Send final update
		finalUpdate := ExecutionUpdate{
			Type:        "execution_complete",
			ExecutionID: result.ID,
			State:       result.FinalState,
			Timestamp:   time.Now(),
		}

		if err != nil {
			finalUpdate.Error = err.Error()
		}

		updateCh <- finalUpdate
	}()

	return updateCh, nil
}

// GetExecutionHistory returns the execution history
func (e *DefaultGraphExecutor) GetExecutionHistory() []ExecutionRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()

	history := make([]ExecutionRecord, len(e.executionHistory))
	copy(history, e.executionHistory)

	return history
}

// GetMetrics returns execution metrics
func (e *DefaultGraphExecutor) GetMetrics() ExecutionMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy of metrics
	metrics := ExecutionMetrics{
		TotalExecutions:      e.metrics.TotalExecutions,
		SuccessfulExecutions: e.metrics.SuccessfulExecutions,
		FailedExecutions:     e.metrics.FailedExecutions,
		AverageDuration:      e.metrics.AverageDuration,
		TotalDuration:        e.metrics.TotalDuration,
		NodeMetrics:          make(map[string]NodeExecutionMetrics),
		GraphMetrics:         make(map[string]ExecutionMetrics),
	}

	// Copy node metrics
	for nodeID, nodeMetrics := range e.metrics.NodeMetrics {
		metrics.NodeMetrics[nodeID] = nodeMetrics
	}

	// Copy graph metrics
	for graphID, graphMetrics := range e.metrics.GraphMetrics {
		metrics.GraphMetrics[graphID] = graphMetrics
	}

	return metrics
}

// Helper methods

func (e *DefaultGraphExecutor) executeNode(ctx context.Context, node Node, state GraphState, callback ExecutionCallback) (*NodeResult, GraphState, error) {
	nodeID := node.GetID()
	startTime := time.Now()

	e.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"node_type": node.GetType(),
	}).Debug("Executing node")

	// Call node start callback
	if callback != nil {
		if err := callback.OnNodeStart(ctx, node, state); err != nil {
			e.logger.WithError(err).Error("Node start callback failed")
		}
	}

	// Execute the node
	newState, err := node.Execute(ctx, state)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// Create node result
	nodeResult := &NodeResult{
		NodeID:    nodeID,
		NodeType:  node.GetType(),
		Success:   err == nil,
		StartTime: startTime,
		EndTime:   endTime,
		Duration:  duration,
		Input:     state,
		Output:    newState,
	}

	if err != nil {
		nodeResult.Error = err.Error()

		// Call node error callback
		if callback != nil {
			if cbErr := callback.OnNodeError(ctx, node, state, err); cbErr != nil {
				e.logger.WithError(cbErr).Error("Node error callback failed")
			}
		}

		return nodeResult, state, err
	}

	// Call node end callback
	if callback != nil {
		if err := callback.OnNodeEnd(ctx, node, nodeResult); err != nil {
			e.logger.WithError(err).Error("Node end callback failed")
		}
	}

	// Update node metrics
	e.updateNodeMetrics(nodeID, node.GetType(), duration, true)

	e.logger.WithFields(logrus.Fields{
		"node_id":   nodeID,
		"node_type": node.GetType(),
		"duration":  duration,
	}).Debug("Node execution completed")

	return nodeResult, newState, nil
}

func (e *DefaultGraphExecutor) getNextNode(ctx context.Context, graph Graph, currentNode string, state GraphState) (string, error) {
	edges := graph.GetEdges(currentNode)

	if len(edges) == 0 {
		// No outgoing edges, execution ends
		return "", nil
	}

	// Evaluate conditions and find the best edge
	var bestEdge Edge
	bestWeight := -1.0

	for _, edge := range edges {
		condition := edge.GetCondition()

		// If no condition, edge is always valid
		if condition == nil {
			if edge.GetWeight() > bestWeight {
				bestEdge = edge
				bestWeight = edge.GetWeight()
			}
			continue
		}

		// Evaluate condition
		matches, err := condition.Evaluate(ctx, state)
		if err != nil {
			e.logger.WithError(err).WithField("edge", fmt.Sprintf("%s->%s", edge.GetFrom(), edge.GetTo())).Error("Failed to evaluate edge condition")
			continue
		}

		if matches && edge.GetWeight() > bestWeight {
			bestEdge = edge
			bestWeight = edge.GetWeight()
		}
	}

	if bestEdge == nil {
		// No valid edge found
		return "", nil
	}

	return bestEdge.GetTo(), nil
}

func (e *DefaultGraphExecutor) recordExecution(result *ExecutionResult) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Create execution record
	record := ExecutionRecord{
		ID:           result.ID,
		InitialState: make(GraphState), // We don't store initial state in this simple implementation
		FinalState:   result.FinalState,
		ExecutedPath: result.ExecutedPath,
		Success:      result.Success,
		Error:        result.Error,
		StartTime:    result.StartTime,
		EndTime:      result.EndTime,
		Duration:     result.Duration,
		NodeCount:    len(result.ExecutedPath),
	}

	// Add to history
	e.executionHistory = append(e.executionHistory, record)

	// Trim history if necessary
	if len(e.executionHistory) > 1000 { // Max history size
		e.executionHistory = e.executionHistory[1:]
	}

	// Update metrics
	e.metrics.TotalExecutions++
	if result.Success {
		e.metrics.SuccessfulExecutions++
	} else {
		e.metrics.FailedExecutions++
	}

	e.metrics.TotalDuration += result.Duration
	e.metrics.AverageDuration = e.metrics.TotalDuration / time.Duration(e.metrics.TotalExecutions)
}

func (e *DefaultGraphExecutor) updateNodeMetrics(nodeID, nodeType string, duration time.Duration, success bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	metrics, exists := e.metrics.NodeMetrics[nodeID]
	if !exists {
		metrics = NodeExecutionMetrics{
			MinDuration: duration,
			MaxDuration: duration,
		}
	}

	metrics.TotalExecutions++
	if success {
		metrics.SuccessfulExecutions++
	} else {
		metrics.FailedExecutions++
	}

	if duration < metrics.MinDuration {
		metrics.MinDuration = duration
	}
	if duration > metrics.MaxDuration {
		metrics.MaxDuration = duration
	}

	totalDuration := metrics.AverageDuration*time.Duration(metrics.TotalExecutions-1) + duration
	metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalExecutions)

	e.metrics.NodeMetrics[nodeID] = metrics
}

// streamingCallback implements ExecutionCallback for streaming updates
type streamingCallback struct {
	updateCh chan<- ExecutionUpdate
}

func (c *streamingCallback) OnExecutionStart(ctx context.Context, graph Graph, initialState GraphState) error {
	c.updateCh <- ExecutionUpdate{
		Type:      "execution_start",
		State:     initialState,
		Timestamp: time.Now(),
	}
	return nil
}

func (c *streamingCallback) OnExecutionEnd(ctx context.Context, result *ExecutionResult) error {
	// Final update will be sent by the executor
	return nil
}

func (c *streamingCallback) OnNodeStart(ctx context.Context, node Node, state GraphState) error {
	c.updateCh <- ExecutionUpdate{
		Type:      "node_start",
		NodeID:    node.GetID(),
		State:     state,
		Timestamp: time.Now(),
	}
	return nil
}

func (c *streamingCallback) OnNodeEnd(ctx context.Context, node Node, result *NodeResult) error {
	c.updateCh <- ExecutionUpdate{
		Type:      "node_complete",
		NodeID:    node.GetID(),
		State:     result.Output,
		Timestamp: time.Now(),
	}
	return nil
}

func (c *streamingCallback) OnNodeError(ctx context.Context, node Node, state GraphState, err error) error {
	c.updateCh <- ExecutionUpdate{
		Type:      "node_error",
		NodeID:    node.GetID(),
		State:     state,
		Error:     err.Error(),
		Timestamp: time.Now(),
	}
	return nil
}

func (c *streamingCallback) OnStateUpdate(ctx context.Context, oldState, newState GraphState) error {
	// State updates are handled by node events
	return nil
}
