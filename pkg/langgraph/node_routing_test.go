package langgraph

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvancedNodes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("StartNode", func(t *testing.T) {
		node := NewStartNode("start", logger)
		
		assert.Equal(t, "start", node.GetID())
		assert.Equal(t, "start", node.GetType())
		assert.NoError(t, node.Validate())
		
		state := GraphState{"input": "test"}
		result, err := node.Execute(context.Background(), state)
		
		require.NoError(t, err)
		assert.True(t, result["started"].(bool))
		assert.NotNil(t, result["start_time"])
	})

	t.Run("EndNode", func(t *testing.T) {
		node := NewEndNode("end", logger)
		
		assert.Equal(t, "end", node.GetID())
		assert.Equal(t, "end", node.GetType())
		assert.NoError(t, node.Validate())
		
		state := GraphState{
			"input":      "test",
			"start_time": time.Now(),
		}
		result, err := node.Execute(context.Background(), state)
		
		require.NoError(t, err)
		assert.True(t, result["completed"].(bool))
		assert.NotNil(t, result["end_time"])
		assert.NotNil(t, result["duration"])
	})

	t.Run("ParallelNode", func(t *testing.T) {
		// Create child nodes
		tool1 := NewExampleTool("tool1", "Test tool 1")
		node1 := NewToolNode("node1", tool1, logger)
		
		tool2 := NewExampleTool("tool2", "Test tool 2")
		node2 := NewToolNode("node2", tool2, logger)
		
		parallelNode := NewParallelNode("parallel", []Node{node1, node2}, logger)
		
		assert.Equal(t, "parallel", parallelNode.GetID())
		assert.Equal(t, "parallel", parallelNode.GetType())
		assert.NoError(t, parallelNode.Validate())
		
		state := GraphState{"input": "test"}
		result, err := parallelNode.Execute(context.Background(), state)
		
		require.NoError(t, err)
		assert.NotNil(t, result["parallel_results"])
		
		parallelResults := result["parallel_results"].(map[string]interface{})
		assert.Contains(t, parallelResults, "node1")
		assert.Contains(t, parallelResults, "node2")
	})

	t.Run("LoopNode", func(t *testing.T) {
		// Create a simple tool node
		tool := NewExampleTool("loop_tool", "Test loop tool")
		childNode := NewToolNode("child", tool, logger)
		
		// Create a condition that runs 3 times
		condition := NewFunctionCondition(
			func(ctx context.Context, state GraphState) (bool, error) {
				iterations := 0
				if iter, exists := state["iterations"]; exists {
					if iterInt, ok := iter.(int); ok {
						iterations = iterInt
					}
				}
				return iterations < 3, nil
			},
			"iterations < 3",
			logger,
		)
		
		loopNode := NewLoopNode("loop", childNode, condition, 5, logger)
		
		assert.Equal(t, "loop", loopNode.GetID())
		assert.Equal(t, "loop", loopNode.GetType())
		assert.NoError(t, loopNode.Validate())
		
		state := GraphState{"input": "test"}
		result, err := loopNode.Execute(context.Background(), state)
		
		require.NoError(t, err)
		assert.Equal(t, 3, result["iterations"])
		assert.NotNil(t, result["loop_results"])
	})

	t.Run("SwitchNode", func(t *testing.T) {
		// Create case nodes
		tool1 := NewExampleTool("case1_tool", "Case 1 tool")
		case1Node := NewToolNode("case1", tool1, logger)
		
		tool2 := NewExampleTool("case2_tool", "Case 2 tool")
		case2Node := NewToolNode("case2", tool2, logger)
		
		defaultTool := NewExampleTool("default_tool", "Default tool")
		defaultNode := NewToolNode("default", defaultTool, logger)
		
		// Create conditions
		condition1 := NewStateCondition("switch_value", "equals", "case1", logger)
		condition2 := NewStateCondition("switch_value", "equals", "case2", logger)
		
		cases := []SwitchCase{
			{Condition: condition1, Node: case1Node},
			{Condition: condition2, Node: case2Node},
		}
		
		switchNode := NewSwitchNode("switch", cases, defaultNode, logger)
		
		assert.Equal(t, "switch", switchNode.GetID())
		assert.Equal(t, "switch", switchNode.GetType())
		assert.NoError(t, switchNode.Validate())
		
		// Test case 1
		state1 := GraphState{"switch_value": "case1"}
		result1, err := switchNode.Execute(context.Background(), state1)
		require.NoError(t, err)
		assert.Equal(t, 0, result1["selected_case"])
		
		// Test case 2
		state2 := GraphState{"switch_value": "case2"}
		result2, err := switchNode.Execute(context.Background(), state2)
		require.NoError(t, err)
		assert.Equal(t, 1, result2["selected_case"])
		
		// Test default case
		state3 := GraphState{"switch_value": "unknown"}
		result3, err := switchNode.Execute(context.Background(), state3)
		require.NoError(t, err)
		assert.Equal(t, -1, result3["selected_case"])
	})
}

func TestConditions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("StateCondition", func(t *testing.T) {
		condition := NewStateCondition("value", "greater_than", 5, logger)
		
		// Test true case
		state1 := GraphState{"value": 10}
		result1, err := condition.Evaluate(context.Background(), state1)
		require.NoError(t, err)
		assert.True(t, result1)
		
		// Test false case
		state2 := GraphState{"value": 3}
		result2, err := condition.Evaluate(context.Background(), state2)
		require.NoError(t, err)
		assert.False(t, result2)
		
		// Test missing key
		state3 := GraphState{"other": 10}
		result3, err := condition.Evaluate(context.Background(), state3)
		require.NoError(t, err)
		assert.False(t, result3)
	})

	t.Run("CompositeCondition", func(t *testing.T) {
		condition1 := NewStateCondition("value1", "greater_than", 5, logger)
		condition2 := NewStateCondition("value2", "less_than", 10, logger)
		
		// Test AND condition
		andCondition := NewCompositeCondition([]Condition{condition1, condition2}, "and", logger)
		
		state1 := GraphState{"value1": 7, "value2": 8}
		result1, err := andCondition.Evaluate(context.Background(), state1)
		require.NoError(t, err)
		assert.True(t, result1)
		
		state2 := GraphState{"value1": 3, "value2": 8}
		result2, err := andCondition.Evaluate(context.Background(), state2)
		require.NoError(t, err)
		assert.False(t, result2)
		
		// Test OR condition
		orCondition := NewCompositeCondition([]Condition{condition1, condition2}, "or", logger)
		
		state3 := GraphState{"value1": 3, "value2": 8}
		result3, err := orCondition.Evaluate(context.Background(), state3)
		require.NoError(t, err)
		assert.True(t, result3)
		
		state4 := GraphState{"value1": 3, "value2": 15}
		result4, err := orCondition.Evaluate(context.Background(), state4)
		require.NoError(t, err)
		assert.False(t, result4)
	})

	t.Run("CountCondition", func(t *testing.T) {
		condition := NewCountCondition("items", "greater_than", 2, logger)
		
		// Test with array
		state1 := GraphState{"items": []interface{}{1, 2, 3, 4}}
		result1, err := condition.Evaluate(context.Background(), state1)
		require.NoError(t, err)
		assert.True(t, result1)
		
		// Test with string
		state2 := GraphState{"items": "ab"}
		result2, err := condition.Evaluate(context.Background(), state2)
		require.NoError(t, err)
		assert.False(t, result2)
		
		// Test with number
		state3 := GraphState{"items": 5}
		result3, err := condition.Evaluate(context.Background(), state3)
		require.NoError(t, err)
		assert.True(t, result3)
	})
}

func TestRouting(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("ConditionalRouter", func(t *testing.T) {
		router := NewConditionalRouter(logger)
		
		// Create edges with conditions
		condition1 := NewStateCondition("route", "equals", "path1", logger)
		edge1 := NewConditionalEdge("start", "node1", condition1, 1.0, logger)
		
		condition2 := NewStateCondition("route", "equals", "path2", logger)
		edge2 := NewConditionalEdge("start", "node2", condition2, 1.0, logger)
		
		edges := []Edge{edge1, edge2}
		
		// Test routing to path1
		state1 := GraphState{"route": "path1"}
		nextNodes1, err := router.Route(context.Background(), "start", state1, edges)
		require.NoError(t, err)
		assert.Equal(t, []string{"node1"}, nextNodes1)
		
		// Test routing to path2
		state2 := GraphState{"route": "path2"}
		nextNodes2, err := router.Route(context.Background(), "start", state2, edges)
		require.NoError(t, err)
		assert.Equal(t, []string{"node2"}, nextNodes2)
		
		// Test no matching route
		state3 := GraphState{"route": "unknown"}
		nextNodes3, err := router.Route(context.Background(), "start", state3, edges)
		require.NoError(t, err)
		assert.Empty(t, nextNodes3)
	})

	t.Run("WeightedRouter", func(t *testing.T) {
		router := NewWeightedRouter(logger)
		
		// Configure with fixed seed for deterministic testing
		err := router.Configure(map[string]interface{}{"seed": int64(12345)})
		require.NoError(t, err)
		
		// Create weighted edges
		edge1 := NewWeightedEdge("start", "node1", 0.7, 1, logger)
		edge2 := NewWeightedEdge("start", "node2", 0.3, 2, logger)
		
		edges := []Edge{edge1, edge2}
		
		// Test multiple routes (should be weighted)
		state := GraphState{}
		results := make(map[string]int)
		
		for i := 0; i < 100; i++ {
			nextNodes, err := router.Route(context.Background(), "start", state, edges)
			require.NoError(t, err)
			require.Len(t, nextNodes, 1)
			results[nextNodes[0]]++
		}
		
		// node1 should be selected more often due to higher weight
		assert.Greater(t, results["node1"], results["node2"])
	})

	t.Run("PriorityRouter", func(t *testing.T) {
		router := NewPriorityRouter(logger)
		
		// Create edges with different priorities (using weight as priority)
		edge1 := NewWeightedEdge("start", "node1", 1.0, 1, logger) // Lower priority
		edge2 := NewWeightedEdge("start", "node2", 2.0, 2, logger) // Higher priority
		edge3 := NewWeightedEdge("start", "node3", 3.0, 3, logger) // Highest priority
		
		edges := []Edge{edge1, edge2, edge3}
		
		state := GraphState{}
		nextNodes, err := router.Route(context.Background(), "start", state, edges)
		require.NoError(t, err)
		
		// Should select highest priority (highest weight)
		assert.Equal(t, []string{"node3"}, nextNodes)
	})

	t.Run("ParallelRouter", func(t *testing.T) {
		router := NewParallelRouter(2, logger) // Max 2 parallel
		
		// Create multiple edges
		edge1 := NewEdge("start", "node1", nil, 1.0)
		edge2 := NewEdge("start", "node2", nil, 1.0)
		edge3 := NewEdge("start", "node3", nil, 1.0)
		
		edges := []Edge{edge1, edge2, edge3}
		
		state := GraphState{}
		nextNodes, err := router.Route(context.Background(), "start", state, edges)
		require.NoError(t, err)
		
		// Should select max 2 nodes for parallel execution
		assert.Len(t, nextNodes, 2)
		assert.Contains(t, []string{"node1", "node2", "node3"}, nextNodes[0])
		assert.Contains(t, []string{"node1", "node2", "node3"}, nextNodes[1])
	})
}

func TestEdgeBuilders(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("EdgeBuilder", func(t *testing.T) {
		builder := NewEdgeBuilder(logger)
		condition := NewStateCondition("test", "equals", "value", logger)
		
		edge, err := builder.
			From("node1").
			To("node2").
			WithCondition(condition).
			WithWeight(2.5).
			WithMetadata("type", "test").
			Build()
		
		require.NoError(t, err)
		assert.Equal(t, "node1", edge.GetFrom())
		assert.Equal(t, "node2", edge.GetTo())
		assert.Equal(t, condition, edge.GetCondition())
		assert.Equal(t, 2.5, edge.GetWeight())
		assert.Equal(t, "test", edge.GetMetadata()["type"])
	})

	t.Run("EdgeFactory", func(t *testing.T) {
		factory := NewEdgeFactory(logger)
		
		// Test different edge types
		options := map[string]interface{}{
			"weight": 1.5,
			"metadata": map[string]interface{}{
				"category": "test",
			},
		}
		
		edge, err := factory.CreateEdge("weighted", "start", "end", options)
		require.NoError(t, err)
		assert.Equal(t, "start", edge.GetFrom())
		assert.Equal(t, "end", edge.GetTo())
		assert.Equal(t, 1.5, edge.GetWeight())
	})
}

func TestNodeBuilders(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("NodeFactory", func(t *testing.T) {
		factory := NewNodeFactory(logger)
		
		// Test start node creation
		startNode, err := factory.CreateNode("start", "start_node", nil)
		require.NoError(t, err)
		assert.Equal(t, "start_node", startNode.GetID())
		assert.Equal(t, "start", startNode.GetType())
		
		// Test tool node creation
		tool := NewExampleTool("test_tool", "Test tool")
		toolOptions := map[string]interface{}{
			"tool": tool,
		}
		toolNode, err := factory.CreateNode("tool", "tool_node", toolOptions)
		require.NoError(t, err)
		assert.Equal(t, "tool_node", toolNode.GetID())
		assert.Equal(t, "tool", toolNode.GetType())
	})

	t.Run("GraphBuilder", func(t *testing.T) {
		builder := NewGraphBuilder(logger)
		
		// Create nodes
		startNode := NewStartNode("start", logger)
		endNode := NewEndNode("end", logger)
		
		// Create edge
		edge := NewEdge("start", "end", nil, 1.0)
		
		graph, err := builder.
			AddNode(startNode).
			AddNode(endNode).
			AddEdge(edge).
			Build()
		
		require.NoError(t, err)
		assert.NotNil(t, graph)
		
		// Validate graph structure
		assert.NoError(t, graph.Validate())
		
		nodes := graph.GetNodes()
		assert.Len(t, nodes, 2)
		assert.Contains(t, nodes, "start")
		assert.Contains(t, nodes, "end")
		
		edges := graph.GetEdges("start")
		assert.Len(t, edges, 1)
		assert.Equal(t, "end", edges[0].GetTo())
	})
}

func TestExamples(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("DemoAllGraphTypes", func(t *testing.T) {
		err := DemoAllGraphTypes(logger)
		assert.NoError(t, err)
	})

	t.Run("CreateSimpleLinearGraph", func(t *testing.T) {
		graph, err := CreateSimpleLinearGraph(logger)
		require.NoError(t, err)
		assert.NoError(t, graph.Validate())
		
		nodes := graph.GetNodes()
		assert.Len(t, nodes, 4) // start, tool1, tool2, end
	})

	t.Run("CreateConditionalGraph", func(t *testing.T) {
		graph, err := CreateConditionalGraph(logger)
		require.NoError(t, err)
		assert.NoError(t, graph.Validate())
		
		nodes := graph.GetNodes()
		assert.Len(t, nodes, 5) // start, decision, high_value, low_value, end
	})

	t.Run("CreateParallelGraph", func(t *testing.T) {
		graph, err := CreateParallelGraph(logger)
		require.NoError(t, err)
		assert.NoError(t, graph.Validate())
		
		nodes := graph.GetNodes()
		assert.Len(t, nodes, 6) // start, 3 parallel nodes, merge, end
	})
}
