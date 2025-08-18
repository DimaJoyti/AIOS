# Langgraph: Advanced Node Types and Routing System

## Overview

Langgraph provides a comprehensive graph execution framework with advanced node types and sophisticated routing capabilities. This system enables the creation of complex workflows with conditional logic, parallel processing, loops, and intelligent routing.

## Node Types

### Basic Nodes

#### StartNode
- **Purpose**: Entry point for graph execution
- **Features**: Initializes execution state, records start time
- **Usage**: `NewStartNode(id, logger)`

#### EndNode
- **Purpose**: Exit point for graph execution
- **Features**: Finalizes execution state, calculates duration
- **Usage**: `NewEndNode(id, logger)`

#### PassthroughNode
- **Purpose**: Passes state through without modification
- **Features**: Useful for routing and state propagation
- **Usage**: `NewPassthroughNode(id, logger)`

### Processing Nodes

#### LLMNode
- **Purpose**: Executes Large Language Model operations
- **Features**: Prompt templating, response processing
- **Usage**: `NewLLMNode(id, llm, promptTemplate, logger)`

#### ToolNode
- **Purpose**: Executes external tools and functions
- **Features**: Input/output schema validation, error handling
- **Usage**: `NewToolNode(id, tool, logger)`

#### FunctionNode
- **Purpose**: Executes custom Go functions
- **Features**: Direct function execution, flexible input/output
- **Usage**: `NewFunctionNode(id, function, logger)`

### Control Flow Nodes

#### ConditionalNode
- **Purpose**: Makes routing decisions based on conditions
- **Features**: State-based evaluation, boolean logic
- **Usage**: `NewConditionalNode(id, condition, logger)`

#### ParallelNode
- **Purpose**: Executes multiple child nodes concurrently
- **Features**: Concurrent execution, result aggregation
- **Usage**: `NewParallelNode(id, childNodes, logger)`

#### LoopNode
- **Purpose**: Repeats execution based on conditions
- **Features**: Iteration control, max iteration limits
- **Usage**: `NewLoopNode(id, childNode, condition, maxIter, logger)`

#### SwitchNode
- **Purpose**: Routes to different nodes based on multiple conditions
- **Features**: Multiple case evaluation, default fallback
- **Usage**: `NewSwitchNode(id, cases, defaultNode, logger)`

### Memory Nodes

#### MemoryNode
- **Purpose**: Manages memory operations (read/write/clear)
- **Features**: Integration with memory systems, state persistence
- **Usage**: `NewMemoryNode(id, memorySystem, operation, logger)`

## Condition Types

### StateCondition
Evaluates conditions based on graph state values.

```go
// Check if value is greater than 5
condition := NewStateCondition("value", "greater_than", 5, logger)

// Supported operators:
// - equals, eq
// - not_equals, ne
// - greater_than, gt
// - greater_equal, ge
// - less_than, lt
// - less_equal, le
// - contains
// - starts_with
// - ends_with
// - exists
// - not_exists
```

### FunctionCondition
Uses custom functions for complex evaluation logic.

```go
condition := NewFunctionCondition(
    func(ctx context.Context, state GraphState) (bool, error) {
        // Custom logic here
        return someComplexLogic(state), nil
    },
    "custom condition description",
    logger,
)
```

### CompositeCondition
Combines multiple conditions with logical operators.

```go
// AND condition
andCondition := NewCompositeCondition(
    []Condition{condition1, condition2}, 
    "and", 
    logger,
)

// OR condition
orCondition := NewCompositeCondition(
    []Condition{condition1, condition2}, 
    "or", 
    logger,
)

// NOT condition
notCondition := NewCompositeCondition(
    []Condition{condition1}, 
    "not", 
    logger,
)
```

### CountCondition
Evaluates based on collection sizes or numeric values.

```go
// Check if array has more than 3 items
condition := NewCountCondition("items", "greater_than", 3, logger)
```

## Routing Strategies

### ConditionalRouter
Routes based on edge conditions. Default router for most use cases.

```go
router := NewConditionalRouter(logger)
```

### WeightedRouter
Routes based on edge weights using weighted random selection.

```go
router := NewWeightedRouter(logger)
// Configure with seed for deterministic behavior
router.Configure(map[string]interface{}{"seed": int64(12345)})
```

### PriorityRouter
Routes to the highest priority (weight) edge that matches conditions.

```go
router := NewPriorityRouter(logger)
```

### RoundRobinRouter
Distributes load evenly across available edges.

```go
router := NewRoundRobinRouter(logger)
```

### ParallelRouter
Routes to multiple nodes for concurrent execution.

```go
router := NewParallelRouter(maxParallel, logger)
```

### HealthBasedRouter
Routes based on node health with fallback strategies.

```go
healthChecker := func(nodeID string) bool {
    // Check node health
    return isNodeHealthy(nodeID)
}
fallbackRouter := NewConditionalRouter(logger)
router := NewHealthBasedRouter(healthChecker, fallbackRouter, logger)
```

## Edge Types

### ConditionalEdge
Standard edge with optional conditions.

```go
condition := NewStateCondition("route", "equals", "path1", logger)
edge := NewConditionalEdge("from", "to", condition, 1.0, logger)
```

### WeightedEdge
Edge with weight and priority for routing decisions.

```go
edge := NewWeightedEdge("from", "to", 0.7, 1, logger)
```

### ParallelEdge
Edge for parallel execution with concurrency control.

```go
edge := NewParallelEdge("from", "to", maxConcurrency, logger)
```

### LoopEdge
Edge for loop constructs with iteration limits.

```go
condition := NewCountCondition("iterations", "less_than", 5, logger)
edge := NewLoopEdge("from", "to", condition, maxIterations, logger)
```

## Builder Pattern

### NodeBuilder
Fluent interface for building nodes.

```go
builder := NewNodeBuilder(logger)
node, err := builder.
    WithID("my_node").
    WithType("function").
    WithInputKeys("input1", "input2").
    WithOutputKeys("output1").
    WithMetadata("category", "processing").
    BuildFunction(myFunction)
```

### EdgeBuilder
Fluent interface for building edges.

```go
builder := NewEdgeBuilder(logger)
edge, err := builder.
    From("node1").
    To("node2").
    WithCondition(condition).
    WithWeight(2.5).
    WithMetadata("type", "conditional").
    BuildConditional()
```

### GraphBuilder
Fluent interface for building complete graphs.

```go
builder := NewGraphBuilder(logger)
graph, err := builder.
    AddNode(startNode).
    AddNode(processNode).
    AddNode(endNode).
    AddEdge(edge1).
    AddEdge(edge2).
    WithRouter(router).
    Build()
```

## Factory Pattern

### NodeFactory
Creates nodes from configuration.

```go
factory := NewNodeFactory(logger)
node, err := factory.CreateNode("tool", "my_tool", map[string]interface{}{
    "tool": myTool,
    "input_keys": []string{"input"},
    "output_keys": []string{"output"},
})
```

### EdgeFactory
Creates edges from configuration.

```go
factory := NewEdgeFactory(logger)
edge, err := factory.CreateEdge("weighted", "from", "to", map[string]interface{}{
    "weight": 1.5,
    "priority": 2,
})
```

### RouterFactory
Creates routers from configuration.

```go
factory := NewRouterFactory(logger)
router, err := factory.CreateRouter("weighted", map[string]interface{}{
    "seed": int64(12345),
})
```

## Example Usage

### Simple Linear Workflow

```go
// Create nodes
startNode := NewStartNode("start", logger)
toolNode := NewToolNode("process", myTool, logger)
endNode := NewEndNode("end", logger)

// Build graph
builder := NewGraphBuilder(logger)
graph, err := builder.
    AddNode(startNode).
    AddNode(toolNode).
    AddNode(endNode).
    AddEdge(NewEdge("start", "process", nil, 1.0)).
    AddEdge(NewEdge("process", "end", nil, 1.0)).
    Build()
```

### Conditional Workflow

```go
// Create decision condition
condition := NewStateCondition("value", "greater_than", 10, logger)

// Create conditional edges
highValueEdge := NewConditionalEdge("decision", "high_process", 
    NewStateCondition("decision_result", "equals", true, logger), 1.0, logger)
lowValueEdge := NewConditionalEdge("decision", "low_process", 
    NewStateCondition("decision_result", "equals", false, logger), 1.0, logger)

// Build graph with conditional routing
graph, err := builder.
    AddNode(NewConditionalNode("decision", condition, logger)).
    AddNode(NewToolNode("high_process", highValueTool, logger)).
    AddNode(NewToolNode("low_process", lowValueTool, logger)).
    AddEdge(highValueEdge).
    AddEdge(lowValueEdge).
    Build()
```

### Parallel Processing

```go
// Create parallel nodes
parallelNodes := []Node{
    NewToolNode("process1", tool1, logger),
    NewToolNode("process2", tool2, logger),
    NewToolNode("process3", tool3, logger),
}

parallelNode := NewParallelNode("parallel", parallelNodes, logger)

// Use parallel router
router := NewParallelRouter(3, logger)
graph, err := builder.
    AddNode(parallelNode).
    WithRouter(router).
    Build()
```

## Best Practices

1. **Use appropriate node types**: Choose the right node type for your use case
2. **Design clear conditions**: Make conditions readable and maintainable
3. **Handle errors gracefully**: Implement proper error handling in custom functions
4. **Optimize routing**: Choose the right routing strategy for your workflow
5. **Monitor performance**: Use metrics and logging for production deployments
6. **Test thoroughly**: Write comprehensive tests for your graphs
7. **Document workflows**: Provide clear documentation for complex graphs

## Performance Considerations

- **Parallel execution**: Use ParallelNode and ParallelRouter for CPU-intensive tasks
- **Memory management**: Be mindful of state size in long-running workflows
- **Condition optimization**: Keep condition evaluation fast and simple
- **Router selection**: Choose appropriate routers based on your routing needs
- **Resource cleanup**: Ensure proper cleanup in custom functions and tools

## Integration with AIOS

This Langgraph system integrates seamlessly with the AIOS ecosystem:

- **LLM Integration**: Direct integration with AIOS LLM management
- **Memory Systems**: Integration with AIOS memory management
- **Tool Ecosystem**: Compatible with AIOS tool framework
- **Observability**: Full OpenTelemetry integration
- **MCP Protocol**: Exposed through MCP for external access

The advanced node types and routing system provide the foundation for building sophisticated AI workflows that can handle complex decision-making, parallel processing, and adaptive execution patterns.
