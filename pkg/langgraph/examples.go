package langgraph

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ExampleTool demonstrates a simple tool implementation
type ExampleTool struct {
	name        string
	description string
}

func NewExampleTool(name, description string) *ExampleTool {
	return &ExampleTool{
		name:        name,
		description: description,
	}
}

func (t *ExampleTool) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Simple tool that adds a timestamp
	output := make(map[string]interface{})
	for k, v := range input {
		output[k] = v
	}
	output["tool_executed"] = t.name
	output["execution_time"] = time.Now()
	return output, nil
}

func (t *ExampleTool) GetName() string        { return t.name }
func (t *ExampleTool) GetDescription() string { return t.description }
func (t *ExampleTool) GetInputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"input": map[string]interface{}{
				"type":        "string",
				"description": "Input data",
			},
		},
	}
}
func (t *ExampleTool) GetOutputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"tool_executed": map[string]interface{}{
				"type":        "string",
				"description": "Name of the executed tool",
			},
			"execution_time": map[string]interface{}{
				"type":        "string",
				"description": "Time when tool was executed",
			},
		},
	}
}
func (t *ExampleTool) Validate() error { return nil }

// CreateSimpleLinearGraph creates a simple linear graph example
func CreateSimpleLinearGraph(logger *logrus.Logger) (Graph, error) {
	builder := NewGraphBuilder(logger)

	// Create nodes
	startNode := NewStartNode("start", logger)
	tool1 := NewExampleTool("tool1", "First processing tool")
	toolNode1 := NewToolNode("tool1_node", tool1, logger)
	tool2 := NewExampleTool("tool2", "Second processing tool")
	toolNode2 := NewToolNode("tool2_node", tool2, logger)
	endNode := NewEndNode("end", logger)

	// Add nodes to graph
	builder.AddNode(startNode).
		AddNode(toolNode1).
		AddNode(toolNode2).
		AddNode(endNode)

	// Create edges
	edgeFactory := NewEdgeFactory(logger)

	edge1, err := edgeFactory.CreateEdge("default", "start", "tool1_node", nil)
	if err != nil {
		return nil, err
	}

	edge2, err := edgeFactory.CreateEdge("default", "tool1_node", "tool2_node", nil)
	if err != nil {
		return nil, err
	}

	edge3, err := edgeFactory.CreateEdge("default", "tool2_node", "end", nil)
	if err != nil {
		return nil, err
	}

	builder.AddEdge(edge1).AddEdge(edge2).AddEdge(edge3)

	return builder.Build()
}

// CreateConditionalGraph creates a graph with conditional routing
func CreateConditionalGraph(logger *logrus.Logger) (Graph, error) {
	builder := NewGraphBuilder(logger)

	// Create nodes
	startNode := NewStartNode("start", logger)

	// Decision node with condition
	condition := NewStateCondition("decision_value", "greater_than", 5, logger)
	decisionNode := NewConditionalNode("decision", condition, logger)

	// Branch nodes
	tool1 := NewExampleTool("high_value_tool", "Processes high values")
	highValueNode := NewToolNode("high_value", tool1, logger)

	tool2 := NewExampleTool("low_value_tool", "Processes low values")
	lowValueNode := NewToolNode("low_value", tool2, logger)

	endNode := NewEndNode("end", logger)

	// Add nodes
	builder.AddNode(startNode).
		AddNode(decisionNode).
		AddNode(highValueNode).
		AddNode(lowValueNode).
		AddNode(endNode)

	// Create edges
	edgeFactory := NewEdgeFactory(logger)

	// Start to decision
	edge1, err := edgeFactory.CreateEdge("default", "start", "decision", nil)
	if err != nil {
		return nil, err
	}

	// Conditional edges from decision
	highCondition := NewStateCondition("decision_result", "equals", true, logger)
	edge2, err := edgeFactory.CreateEdge("conditional", "decision", "high_value", map[string]interface{}{
		"condition": highCondition,
	})
	if err != nil {
		return nil, err
	}

	lowCondition := NewStateCondition("decision_result", "equals", false, logger)
	edge3, err := edgeFactory.CreateEdge("conditional", "decision", "low_value", map[string]interface{}{
		"condition": lowCondition,
	})
	if err != nil {
		return nil, err
	}

	// Both branches to end
	edge4, err := edgeFactory.CreateEdge("default", "high_value", "end", nil)
	if err != nil {
		return nil, err
	}

	edge5, err := edgeFactory.CreateEdge("default", "low_value", "end", nil)
	if err != nil {
		return nil, err
	}

	builder.AddEdge(edge1).AddEdge(edge2).AddEdge(edge3).AddEdge(edge4).AddEdge(edge5)

	return builder.Build()
}

// CreateParallelGraph creates a graph with parallel execution
func CreateParallelGraph(logger *logrus.Logger) (Graph, error) {
	builder := NewGraphBuilder(logger)

	// Create nodes
	startNode := NewStartNode("start", logger)

	// Parallel processing nodes
	tool1 := NewExampleTool("parallel_tool_1", "First parallel processor")
	parallelNode1 := NewToolNode("parallel_1", tool1, logger)

	tool2 := NewExampleTool("parallel_tool_2", "Second parallel processor")
	parallelNode2 := NewToolNode("parallel_2", tool2, logger)

	tool3 := NewExampleTool("parallel_tool_3", "Third parallel processor")
	parallelNode3 := NewToolNode("parallel_3", tool3, logger)

	// Merge node (function node that combines results)
	mergeFunction := func(ctx context.Context, state GraphState) (GraphState, error) {
		newState := make(GraphState)
		for k, v := range state {
			newState[k] = v
		}

		// Combine parallel results
		results := make([]interface{}, 0)
		if result1, exists := state["parallel_1_result"]; exists {
			results = append(results, result1)
		}
		if result2, exists := state["parallel_2_result"]; exists {
			results = append(results, result2)
		}
		if result3, exists := state["parallel_3_result"]; exists {
			results = append(results, result3)
		}

		newState["merged_results"] = results
		return newState, nil
	}
	mergeNode := NewFunctionNode("merge", mergeFunction, logger)

	endNode := NewEndNode("end", logger)

	// Add nodes
	builder.AddNode(startNode).
		AddNode(parallelNode1).
		AddNode(parallelNode2).
		AddNode(parallelNode3).
		AddNode(mergeNode).
		AddNode(endNode)

	// Create edges
	edgeFactory := NewEdgeFactory(logger)

	// Start to parallel nodes
	edge1, err := edgeFactory.CreateEdge("parallel", "start", "parallel_1", map[string]interface{}{
		"max_concurrency": 1,
	})
	if err != nil {
		return nil, err
	}

	edge2, err := edgeFactory.CreateEdge("parallel", "start", "parallel_2", map[string]interface{}{
		"max_concurrency": 1,
	})
	if err != nil {
		return nil, err
	}

	edge3, err := edgeFactory.CreateEdge("parallel", "start", "parallel_3", map[string]interface{}{
		"max_concurrency": 1,
	})
	if err != nil {
		return nil, err
	}

	// Parallel nodes to merge
	edge4, err := edgeFactory.CreateEdge("default", "parallel_1", "merge", nil)
	if err != nil {
		return nil, err
	}

	edge5, err := edgeFactory.CreateEdge("default", "parallel_2", "merge", nil)
	if err != nil {
		return nil, err
	}

	edge6, err := edgeFactory.CreateEdge("default", "parallel_3", "merge", nil)
	if err != nil {
		return nil, err
	}

	// Merge to end
	edge7, err := edgeFactory.CreateEdge("default", "merge", "end", nil)
	if err != nil {
		return nil, err
	}

	builder.AddEdge(edge1).AddEdge(edge2).AddEdge(edge3).
		AddEdge(edge4).AddEdge(edge5).AddEdge(edge6).AddEdge(edge7)

	return builder.Build()
}

// CreateLoopGraph creates a graph with loop functionality
func CreateLoopGraph(logger *logrus.Logger) (Graph, error) {
	builder := NewGraphBuilder(logger)

	// Create nodes
	startNode := NewStartNode("start", logger)

	// Initialize counter
	initFunction := func(ctx context.Context, state GraphState) (GraphState, error) {
		newState := make(GraphState)
		for k, v := range state {
			newState[k] = v
		}
		newState["counter"] = 0
		newState["max_iterations"] = 5
		return newState, nil
	}
	initNode := NewFunctionNode("init", initFunction, logger)

	// Processing node inside loop
	tool := NewExampleTool("loop_processor", "Processes data in loop")
	processNode := NewToolNode("process", tool, logger)

	// Increment counter
	incrementFunction := func(ctx context.Context, state GraphState) (GraphState, error) {
		newState := make(GraphState)
		for k, v := range state {
			newState[k] = v
		}

		counter := 0
		if c, exists := state["counter"]; exists {
			if cInt, ok := c.(int); ok {
				counter = cInt
			}
		}

		newState["counter"] = counter + 1
		return newState, nil
	}
	incrementNode := NewFunctionNode("increment", incrementFunction, logger)

	endNode := NewEndNode("end", logger)

	// Add nodes
	builder.AddNode(startNode).
		AddNode(initNode).
		AddNode(processNode).
		AddNode(incrementNode).
		AddNode(endNode)

	// Create edges
	edgeFactory := NewEdgeFactory(logger)

	// Linear flow
	edge1, err := edgeFactory.CreateEdge("default", "start", "init", nil)
	if err != nil {
		return nil, err
	}

	edge2, err := edgeFactory.CreateEdge("default", "init", "process", nil)
	if err != nil {
		return nil, err
	}

	edge3, err := edgeFactory.CreateEdge("default", "process", "increment", nil)
	if err != nil {
		return nil, err
	}

	// Loop condition: continue if counter < max_iterations
	loopCondition := NewFunctionCondition(
		func(ctx context.Context, state GraphState) (bool, error) {
			counter := 0
			maxIter := 5

			if c, exists := state["counter"]; exists {
				if cInt, ok := c.(int); ok {
					counter = cInt
				}
			}

			if m, exists := state["max_iterations"]; exists {
				if mInt, ok := m.(int); ok {
					maxIter = mInt
				}
			}

			return counter < maxIter, nil
		},
		"counter < max_iterations",
		logger,
	)

	edge4, err := edgeFactory.CreateEdge("loop", "increment", "process", map[string]interface{}{
		"condition":      loopCondition,
		"max_iterations": 10,
	})
	if err != nil {
		return nil, err
	}

	// Exit condition: go to end when loop is done
	exitCondition := NewFunctionCondition(
		func(ctx context.Context, state GraphState) (bool, error) {
			counter := 0
			maxIter := 5

			if c, exists := state["counter"]; exists {
				if cInt, ok := c.(int); ok {
					counter = cInt
				}
			}

			if m, exists := state["max_iterations"]; exists {
				if mInt, ok := m.(int); ok {
					maxIter = mInt
				}
			}

			return counter >= maxIter, nil
		},
		"counter >= max_iterations",
		logger,
	)

	edge5, err := edgeFactory.CreateEdge("conditional", "increment", "end", map[string]interface{}{
		"condition": exitCondition,
	})
	if err != nil {
		return nil, err
	}

	builder.AddEdge(edge1).AddEdge(edge2).AddEdge(edge3).AddEdge(edge4).AddEdge(edge5)

	return builder.Build()
}

// CreateComplexWorkflowGraph creates a complex workflow combining multiple patterns
func CreateComplexWorkflowGraph(logger *logrus.Logger) (Graph, error) {
	builder := NewGraphBuilder(logger)

	// Create all nodes
	startNode := NewStartNode("start", logger)

	// Input validation
	validateTool := NewExampleTool("validator", "Validates input data")
	validateNode := NewToolNode("validate", validateTool, logger)

	// Decision point
	condition := NewStateCondition("validation_result", "equals", "valid", logger)
	decisionNode := NewConditionalNode("decision", condition, logger)

	// Error handling
	errorTool := NewExampleTool("error_handler", "Handles validation errors")
	errorNode := NewToolNode("error", errorTool, logger)

	// Parallel processing for valid data
	processTool1 := NewExampleTool("processor_1", "First data processor")
	processNode1 := NewToolNode("process_1", processTool1, logger)

	processTool2 := NewExampleTool("processor_2", "Second data processor")
	processNode2 := NewToolNode("process_2", processTool2, logger)

	// Aggregation
	aggregateFunction := func(ctx context.Context, state GraphState) (GraphState, error) {
		newState := make(GraphState)
		for k, v := range state {
			newState[k] = v
		}
		newState["aggregated"] = true
		newState["final_result"] = "Processing completed successfully"
		return newState, nil
	}
	aggregateNode := NewFunctionNode("aggregate", aggregateFunction, logger)

	endNode := NewEndNode("end", logger)

	// Add all nodes
	builder.AddNode(startNode).
		AddNode(validateNode).
		AddNode(decisionNode).
		AddNode(errorNode).
		AddNode(processNode1).
		AddNode(processNode2).
		AddNode(aggregateNode).
		AddNode(endNode)

	// Create edges with different types
	edgeFactory := NewEdgeFactory(logger)

	// Main flow
	edge1, err := edgeFactory.CreateEdge("default", "start", "validate", nil)
	if err != nil {
		return nil, err
	}

	edge2, err := edgeFactory.CreateEdge("default", "validate", "decision", nil)
	if err != nil {
		return nil, err
	}

	// Conditional routing
	validCondition := NewStateCondition("decision_result", "equals", true, logger)
	edge3, err := edgeFactory.CreateEdge("conditional", "decision", "process_1", map[string]interface{}{
		"condition": validCondition,
		"weight":    1.0,
	})
	if err != nil {
		return nil, err
	}

	// Parallel processing
	edge4, err := edgeFactory.CreateEdge("parallel", "decision", "process_2", map[string]interface{}{
		"condition":       validCondition,
		"max_concurrency": 1,
	})
	if err != nil {
		return nil, err
	}

	// Error handling
	errorCondition := NewStateCondition("decision_result", "equals", false, logger)
	edge5, err := edgeFactory.CreateEdge("conditional", "decision", "error", map[string]interface{}{
		"condition": errorCondition,
	})
	if err != nil {
		return nil, err
	}

	// Convergence
	edge6, err := edgeFactory.CreateEdge("default", "process_1", "aggregate", nil)
	if err != nil {
		return nil, err
	}

	edge7, err := edgeFactory.CreateEdge("default", "process_2", "aggregate", nil)
	if err != nil {
		return nil, err
	}

	edge8, err := edgeFactory.CreateEdge("default", "error", "end", nil)
	if err != nil {
		return nil, err
	}

	edge9, err := edgeFactory.CreateEdge("default", "aggregate", "end", nil)
	if err != nil {
		return nil, err
	}

	builder.AddEdge(edge1).AddEdge(edge2).AddEdge(edge3).AddEdge(edge4).
		AddEdge(edge5).AddEdge(edge6).AddEdge(edge7).AddEdge(edge8).AddEdge(edge9)

	return builder.Build()
}

// DemoAllGraphTypes demonstrates all graph types
func DemoAllGraphTypes(logger *logrus.Logger) error {
	logger.Info("Creating and demonstrating all graph types...")

	// Simple linear graph
	linearGraph, err := CreateSimpleLinearGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create linear graph: %w", err)
	}
	logger.Info("✓ Linear graph created successfully")

	// Conditional graph
	conditionalGraph, err := CreateConditionalGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create conditional graph: %w", err)
	}
	logger.Info("✓ Conditional graph created successfully")

	// Parallel graph
	parallelGraph, err := CreateParallelGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create parallel graph: %w", err)
	}
	logger.Info("✓ Parallel graph created successfully")

	// Loop graph
	loopGraph, err := CreateLoopGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create loop graph: %w", err)
	}
	logger.Info("✓ Loop graph created successfully")

	// Complex workflow graph
	complexGraph, err := CreateComplexWorkflowGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create complex workflow graph: %w", err)
	}
	logger.Info("✓ Complex workflow graph created successfully")

	// Validate all graphs
	graphs := []Graph{linearGraph, conditionalGraph, parallelGraph, loopGraph, complexGraph}
	for i, graph := range graphs {
		if err := graph.Validate(); err != nil {
			return fmt.Errorf("graph %d validation failed: %w", i+1, err)
		}
	}

	logger.Info("✓ All graphs validated successfully")
	logger.Info("🎉 Node types and routing system demonstration completed!")

	return nil
}
