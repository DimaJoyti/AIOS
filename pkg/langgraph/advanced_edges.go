package langgraph

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ConditionalEdge represents an edge with conditions
type ConditionalEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewConditionalEdge creates a new conditional edge
func NewConditionalEdge(from, to string, condition Condition, weight float64, logger *logrus.Logger) *ConditionalEdge {
	return &ConditionalEdge{
		from:      from,
		to:        to,
		condition: condition,
		weight:    weight,
		metadata:  make(map[string]interface{}),
		logger:    logger,
	}
}

func (e *ConditionalEdge) GetFrom() string { return e.from }
func (e *ConditionalEdge) GetTo() string { return e.to }
func (e *ConditionalEdge) GetCondition() Condition { return e.condition }
func (e *ConditionalEdge) GetWeight() float64 { return e.weight }
func (e *ConditionalEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *ConditionalEdge) SetWeight(weight float64) { e.weight = weight }
func (e *ConditionalEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *ConditionalEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }

// WeightedEdge represents an edge with weight for routing decisions
type WeightedEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	priority  int
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewWeightedEdge creates a new weighted edge
func NewWeightedEdge(from, to string, weight float64, priority int, logger *logrus.Logger) *WeightedEdge {
	return &WeightedEdge{
		from:      from,
		to:        to,
		weight:    weight,
		priority:  priority,
		metadata:  make(map[string]interface{}),
		logger:    logger,
	}
}

func (e *WeightedEdge) GetFrom() string { return e.from }
func (e *WeightedEdge) GetTo() string { return e.to }
func (e *WeightedEdge) GetCondition() Condition { return e.condition }
func (e *WeightedEdge) GetWeight() float64 { return e.weight }
func (e *WeightedEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *WeightedEdge) SetWeight(weight float64) { e.weight = weight }
func (e *WeightedEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *WeightedEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }
func (e *WeightedEdge) GetPriority() int { return e.priority }
func (e *WeightedEdge) SetPriority(priority int) { e.priority = priority }

// ParallelEdge represents an edge for parallel execution
type ParallelEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	maxConcurrency int
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewParallelEdge creates a new parallel edge
func NewParallelEdge(from, to string, maxConcurrency int, logger *logrus.Logger) *ParallelEdge {
	return &ParallelEdge{
		from:           from,
		to:             to,
		weight:         1.0,
		maxConcurrency: maxConcurrency,
		metadata:       make(map[string]interface{}),
		logger:         logger,
	}
}

func (e *ParallelEdge) GetFrom() string { return e.from }
func (e *ParallelEdge) GetTo() string { return e.to }
func (e *ParallelEdge) GetCondition() Condition { return e.condition }
func (e *ParallelEdge) GetWeight() float64 { return e.weight }
func (e *ParallelEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *ParallelEdge) SetWeight(weight float64) { e.weight = weight }
func (e *ParallelEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *ParallelEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }
func (e *ParallelEdge) GetMaxConcurrency() int { return e.maxConcurrency }
func (e *ParallelEdge) SetMaxConcurrency(maxConcurrency int) { e.maxConcurrency = maxConcurrency }

// LoopEdge represents an edge for loop constructs
type LoopEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	maxIterations int
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewLoopEdge creates a new loop edge
func NewLoopEdge(from, to string, condition Condition, maxIterations int, logger *logrus.Logger) *LoopEdge {
	return &LoopEdge{
		from:          from,
		to:            to,
		condition:     condition,
		weight:        1.0,
		maxIterations: maxIterations,
		metadata:      make(map[string]interface{}),
		logger:        logger,
	}
}

func (e *LoopEdge) GetFrom() string { return e.from }
func (e *LoopEdge) GetTo() string { return e.to }
func (e *LoopEdge) GetCondition() Condition { return e.condition }
func (e *LoopEdge) GetWeight() float64 { return e.weight }
func (e *LoopEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *LoopEdge) SetWeight(weight float64) { e.weight = weight }
func (e *LoopEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *LoopEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }
func (e *LoopEdge) GetMaxIterations() int { return e.maxIterations }
func (e *LoopEdge) SetMaxIterations(maxIterations int) { e.maxIterations = maxIterations }

// TimeoutEdge represents an edge with timeout conditions
type TimeoutEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	timeout   time.Duration
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewTimeoutEdge creates a new timeout edge
func NewTimeoutEdge(from, to string, timeout time.Duration, logger *logrus.Logger) *TimeoutEdge {
	return &TimeoutEdge{
		from:      from,
		to:        to,
		weight:    1.0,
		timeout:   timeout,
		metadata:  make(map[string]interface{}),
		logger:    logger,
	}
}

func (e *TimeoutEdge) GetFrom() string { return e.from }
func (e *TimeoutEdge) GetTo() string { return e.to }
func (e *TimeoutEdge) GetCondition() Condition { return e.condition }
func (e *TimeoutEdge) GetWeight() float64 { return e.weight }
func (e *TimeoutEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *TimeoutEdge) SetWeight(weight float64) { e.weight = weight }
func (e *TimeoutEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *TimeoutEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }
func (e *TimeoutEdge) GetTimeout() time.Duration { return e.timeout }
func (e *TimeoutEdge) SetTimeout(timeout time.Duration) { e.timeout = timeout }

// ErrorEdge represents an edge for error handling
type ErrorEdge struct {
	from      string
	to        string
	condition Condition
	weight    float64
	errorTypes []string
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewErrorEdge creates a new error edge
func NewErrorEdge(from, to string, errorTypes []string, logger *logrus.Logger) *ErrorEdge {
	return &ErrorEdge{
		from:       from,
		to:         to,
		weight:     1.0,
		errorTypes: errorTypes,
		metadata:   make(map[string]interface{}),
		logger:     logger,
	}
}

func (e *ErrorEdge) GetFrom() string { return e.from }
func (e *ErrorEdge) GetTo() string { return e.to }
func (e *ErrorEdge) GetCondition() Condition { return e.condition }
func (e *ErrorEdge) GetWeight() float64 { return e.weight }
func (e *ErrorEdge) SetCondition(condition Condition) { e.condition = condition }
func (e *ErrorEdge) SetWeight(weight float64) { e.weight = weight }
func (e *ErrorEdge) GetMetadata() map[string]interface{} { return e.metadata }
func (e *ErrorEdge) SetMetadata(key string, value interface{}) { e.metadata[key] = value }
func (e *ErrorEdge) GetErrorTypes() []string { return e.errorTypes }
func (e *ErrorEdge) SetErrorTypes(errorTypes []string) { e.errorTypes = errorTypes }

// EdgeBuilder provides a fluent interface for building edges
type EdgeBuilder struct {
	from      string
	to        string
	condition Condition
	weight    float64
	metadata  map[string]interface{}
	logger    *logrus.Logger
}

// NewEdgeBuilder creates a new edge builder
func NewEdgeBuilder(logger *logrus.Logger) *EdgeBuilder {
	return &EdgeBuilder{
		weight:   1.0,
		metadata: make(map[string]interface{}),
		logger:   logger,
	}
}

// From sets the source node
func (b *EdgeBuilder) From(nodeID string) *EdgeBuilder {
	b.from = nodeID
	return b
}

// To sets the target node
func (b *EdgeBuilder) To(nodeID string) *EdgeBuilder {
	b.to = nodeID
	return b
}

// WithCondition sets the condition
func (b *EdgeBuilder) WithCondition(condition Condition) *EdgeBuilder {
	b.condition = condition
	return b
}

// WithWeight sets the weight
func (b *EdgeBuilder) WithWeight(weight float64) *EdgeBuilder {
	b.weight = weight
	return b
}

// WithMetadata adds metadata
func (b *EdgeBuilder) WithMetadata(key string, value interface{}) *EdgeBuilder {
	b.metadata[key] = value
	return b
}

// Build creates the edge
func (b *EdgeBuilder) Build() (Edge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewEdge(b.from, b.to, b.condition, b.weight)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildConditional creates a conditional edge
func (b *EdgeBuilder) BuildConditional() (*ConditionalEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewConditionalEdge(b.from, b.to, b.condition, b.weight, b.logger)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildWeighted creates a weighted edge
func (b *EdgeBuilder) BuildWeighted(priority int) (*WeightedEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewWeightedEdge(b.from, b.to, b.weight, priority, b.logger)
	edge.SetCondition(b.condition)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildParallel creates a parallel edge
func (b *EdgeBuilder) BuildParallel(maxConcurrency int) (*ParallelEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewParallelEdge(b.from, b.to, maxConcurrency, b.logger)
	edge.SetCondition(b.condition)
	edge.SetWeight(b.weight)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildLoop creates a loop edge
func (b *EdgeBuilder) BuildLoop(maxIterations int) (*LoopEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}
	if b.condition == nil {
		return nil, fmt.Errorf("loop edge requires a condition")
	}

	edge := NewLoopEdge(b.from, b.to, b.condition, maxIterations, b.logger)
	edge.SetWeight(b.weight)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildTimeout creates a timeout edge
func (b *EdgeBuilder) BuildTimeout(timeout time.Duration) (*TimeoutEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewTimeoutEdge(b.from, b.to, timeout, b.logger)
	edge.SetCondition(b.condition)
	edge.SetWeight(b.weight)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// BuildError creates an error edge
func (b *EdgeBuilder) BuildError(errorTypes []string) (*ErrorEdge, error) {
	if b.from == "" {
		return nil, fmt.Errorf("source node cannot be empty")
	}
	if b.to == "" {
		return nil, fmt.Errorf("target node cannot be empty")
	}

	edge := NewErrorEdge(b.from, b.to, errorTypes, b.logger)
	edge.SetCondition(b.condition)
	edge.SetWeight(b.weight)
	for k, v := range b.metadata {
		edge.SetMetadata(k, v)
	}

	return edge, nil
}

// EdgeFactory creates different types of edges
type EdgeFactory struct {
	logger *logrus.Logger
}

// NewEdgeFactory creates a new edge factory
func NewEdgeFactory(logger *logrus.Logger) *EdgeFactory {
	return &EdgeFactory{
		logger: logger,
	}
}

// CreateEdge creates an edge of the specified type
func (f *EdgeFactory) CreateEdge(edgeType string, from, to string, options map[string]interface{}) (Edge, error) {
	builder := NewEdgeBuilder(f.logger).From(from).To(to)

	// Apply common options
	if weight, exists := options["weight"]; exists {
		if weightFloat, ok := weight.(float64); ok {
			builder.WithWeight(weightFloat)
		}
	}

	if condition, exists := options["condition"]; exists {
		if conditionObj, ok := condition.(Condition); ok {
			builder.WithCondition(conditionObj)
		}
	}

	if metadata, exists := options["metadata"]; exists {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			for k, v := range metadataMap {
				builder.WithMetadata(k, v)
			}
		}
	}

	switch edgeType {
	case "default":
		return builder.Build()
	case "conditional":
		return builder.BuildConditional()
	case "weighted":
		priority := 0
		if p, exists := options["priority"]; exists {
			if pInt, ok := p.(int); ok {
				priority = pInt
			}
		}
		return builder.BuildWeighted(priority)
	case "parallel":
		maxConcurrency := 1
		if mc, exists := options["max_concurrency"]; exists {
			if mcInt, ok := mc.(int); ok {
				maxConcurrency = mcInt
			}
		}
		return builder.BuildParallel(maxConcurrency)
	case "loop":
		maxIterations := 100
		if mi, exists := options["max_iterations"]; exists {
			if miInt, ok := mi.(int); ok {
				maxIterations = miInt
			}
		}
		return builder.BuildLoop(maxIterations)
	case "timeout":
		timeout := 30 * time.Second
		if t, exists := options["timeout"]; exists {
			if tDuration, ok := t.(time.Duration); ok {
				timeout = tDuration
			}
		}
		return builder.BuildTimeout(timeout)
	case "error":
		errorTypes := []string{}
		if et, exists := options["error_types"]; exists {
			if etSlice, ok := et.([]string); ok {
				errorTypes = etSlice
			}
		}
		return builder.BuildError(errorTypes)
	default:
		return nil, fmt.Errorf("unsupported edge type: %s", edgeType)
	}
}
