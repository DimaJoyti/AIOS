package langgraph

import (
	"context"
	"fmt"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langchain/prompts"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// LLMNode represents a node that uses an LLM
type LLMNode struct {
	id         string
	llm        llm.LLM
	prompt     prompts.PromptTemplate
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewLLMNode creates a new LLM node
func NewLLMNode(id string, llmInstance llm.LLM, promptTemplate prompts.PromptTemplate, logger *logrus.Logger) *LLMNode {
	inputKeys := []string{"input"}
	if promptTemplate != nil {
		inputKeys = promptTemplate.GetInputVariables()
	}

	return &LLMNode{
		id:         id,
		llm:        llmInstance,
		prompt:     promptTemplate,
		inputKeys:  inputKeys,
		outputKeys: []string{"output", "text"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.llm"),
	}
}

// Execute executes the LLM node
func (n *LLMNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "llm_node.execute")
	defer span.End()

	// Extract variables from state
	variables := make(map[string]interface{})
	for _, key := range n.inputKeys {
		if value, exists := state[key]; exists {
			variables[key] = value
		}
	}

	// Format prompt if available
	var messages []llm.Message
	if n.prompt != nil {
		promptValue, err := n.prompt.FormatPrompt(variables)
		if err != nil {
			return state, fmt.Errorf("failed to format prompt: %w", err)
		}
		messages = promptValue.ToMessages()
	} else {
		// Use input directly as user message
		if input, exists := state["input"]; exists {
			messages = []llm.Message{
				{
					Role:    "user",
					Content: fmt.Sprintf("%v", input),
				},
			}
		} else {
			return state, fmt.Errorf("no input provided and no prompt template")
		}
	}

	// Execute LLM
	req := &llm.CompletionRequest{
		Messages: messages,
	}

	response, err := n.llm.Complete(ctx, req)
	if err != nil {
		return state, fmt.Errorf("LLM execution failed: %w", err)
	}

	// Update state with output
	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	newState["output"] = response.Content
	newState["text"] = response.Content
	newState["llm_response"] = response

	return newState, nil
}

// GetID returns the node ID
func (n *LLMNode) GetID() string {
	return n.id
}

// GetType returns the node type
func (n *LLMNode) GetType() string {
	return "llm"
}

// GetInputKeys returns the input keys
func (n *LLMNode) GetInputKeys() []string {
	return n.inputKeys
}

// GetOutputKeys returns the output keys
func (n *LLMNode) GetOutputKeys() []string {
	return n.outputKeys
}

// Validate validates the node
func (n *LLMNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.llm == nil {
		return fmt.Errorf("LLM cannot be nil")
	}
	return nil
}

// ToolNode represents a node that executes a tool
type ToolNode struct {
	id         string
	tool       Tool
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewToolNode creates a new tool node
func NewToolNode(id string, tool Tool, logger *logrus.Logger) *ToolNode {
	return &ToolNode{
		id:         id,
		tool:       tool,
		inputKeys:  []string{"input"},
		outputKeys: []string{"output"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.tool"),
	}
}

// Execute executes the tool node
func (n *ToolNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "tool_node.execute")
	defer span.End()

	// Prepare input for tool
	toolInput := make(map[string]interface{})
	for _, key := range n.inputKeys {
		if value, exists := state[key]; exists {
			toolInput[key] = value
		}
	}

	// Execute tool
	toolOutput, err := n.tool.Execute(ctx, toolInput)
	if err != nil {
		return state, fmt.Errorf("tool execution failed: %w", err)
	}

	// Update state with output
	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	// Merge tool output into state
	for k, v := range toolOutput {
		newState[k] = v
	}

	return newState, nil
}

// GetID returns the node ID
func (n *ToolNode) GetID() string {
	return n.id
}

// GetType returns the node type
func (n *ToolNode) GetType() string {
	return "tool"
}

// GetInputKeys returns the input keys
func (n *ToolNode) GetInputKeys() []string {
	return n.inputKeys
}

// GetOutputKeys returns the output keys
func (n *ToolNode) GetOutputKeys() []string {
	return n.outputKeys
}

// Validate validates the node
func (n *ToolNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.tool == nil {
		return fmt.Errorf("tool cannot be nil")
	}
	return n.tool.Validate()
}

// ConditionalNode represents a node that routes based on conditions
type ConditionalNode struct {
	id         string
	condition  Condition
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewConditionalNode creates a new conditional node
func NewConditionalNode(id string, condition Condition, logger *logrus.Logger) *ConditionalNode {
	return &ConditionalNode{
		id:         id,
		condition:  condition,
		inputKeys:  []string{"input"},
		outputKeys: []string{"condition_result"},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.conditional"),
	}
}

// Execute executes the conditional node
func (n *ConditionalNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "conditional_node.execute")
	defer span.End()

	// Evaluate condition
	result, err := n.condition.Evaluate(ctx, state)
	if err != nil {
		return state, fmt.Errorf("condition evaluation failed: %w", err)
	}

	// Update state with condition result
	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	newState["condition_result"] = result

	return newState, nil
}

// GetID returns the node ID
func (n *ConditionalNode) GetID() string {
	return n.id
}

// GetType returns the node type
func (n *ConditionalNode) GetType() string {
	return "conditional"
}

// GetInputKeys returns the input keys
func (n *ConditionalNode) GetInputKeys() []string {
	return n.inputKeys
}

// GetOutputKeys returns the output keys
func (n *ConditionalNode) GetOutputKeys() []string {
	return n.outputKeys
}

// Validate validates the node
func (n *ConditionalNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.condition == nil {
		return fmt.Errorf("condition cannot be nil")
	}
	return nil
}

// FunctionNode represents a node that executes a custom function
type FunctionNode struct {
	id         string
	function   func(context.Context, GraphState) (GraphState, error)
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewFunctionNode creates a new function node
func NewFunctionNode(id string, fn func(context.Context, GraphState) (GraphState, error), logger *logrus.Logger) *FunctionNode {
	return &FunctionNode{
		id:         id,
		function:   fn,
		inputKeys:  []string{}, // Will be determined dynamically
		outputKeys: []string{}, // Will be determined dynamically
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.function"),
	}
}

// Execute executes the function node
func (n *FunctionNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "function_node.execute")
	defer span.End()

	return n.function(ctx, state)
}

// GetID returns the node ID
func (n *FunctionNode) GetID() string {
	return n.id
}

// GetType returns the node type
func (n *FunctionNode) GetType() string {
	return "function"
}

// GetInputKeys returns the input keys
func (n *FunctionNode) GetInputKeys() []string {
	return n.inputKeys
}

// GetOutputKeys returns the output keys
func (n *FunctionNode) GetOutputKeys() []string {
	return n.outputKeys
}

// Validate validates the node
func (n *FunctionNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if n.function == nil {
		return fmt.Errorf("function cannot be nil")
	}
	return nil
}

// SetInputKeys sets the input keys for the function node
func (n *FunctionNode) SetInputKeys(keys []string) {
	n.inputKeys = keys
}

// SetOutputKeys sets the output keys for the function node
func (n *FunctionNode) SetOutputKeys(keys []string) {
	n.outputKeys = keys
}

// PassthroughNode represents a node that passes state through unchanged
type PassthroughNode struct {
	id         string
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewPassthroughNode creates a new passthrough node
func NewPassthroughNode(id string, logger *logrus.Logger) *PassthroughNode {
	return &PassthroughNode{
		id:         id,
		inputKeys:  []string{},
		outputKeys: []string{},
		logger:     logger,
		tracer:     otel.Tracer("langgraph.nodes.passthrough"),
	}
}

// Execute executes the passthrough node (returns state unchanged)
func (n *PassthroughNode) Execute(ctx context.Context, state GraphState) (GraphState, error) {
	ctx, span := n.tracer.Start(ctx, "passthrough_node.execute")
	defer span.End()

	// Return a copy of the state
	newState := make(GraphState)
	for k, v := range state {
		newState[k] = v
	}

	return newState, nil
}

// GetID returns the node ID
func (n *PassthroughNode) GetID() string {
	return n.id
}

// GetType returns the node type
func (n *PassthroughNode) GetType() string {
	return "passthrough"
}

// GetInputKeys returns the input keys
func (n *PassthroughNode) GetInputKeys() []string {
	return n.inputKeys
}

// GetOutputKeys returns the output keys
func (n *PassthroughNode) GetOutputKeys() []string {
	return n.outputKeys
}

// Validate validates the node
func (n *PassthroughNode) Validate() error {
	if n.id == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	return nil
}
