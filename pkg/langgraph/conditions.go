package langgraph

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StateCondition evaluates conditions based on graph state
type StateCondition struct {
	key      string
	operator string
	value    interface{}
	logger   *logrus.Logger
	tracer   trace.Tracer
}

// NewStateCondition creates a new state condition
func NewStateCondition(key, operator string, value interface{}, logger *logrus.Logger) *StateCondition {
	return &StateCondition{
		key:      key,
		operator: operator,
		value:    value,
		logger:   logger,
		tracer:   otel.Tracer("langgraph.conditions.state"),
	}
}

// Evaluate evaluates the state condition
func (c *StateCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "state_condition.evaluate")
	defer span.End()

	stateValue, exists := state[c.key]
	if !exists {
		return c.operator == "not_exists", nil
	}

	switch c.operator {
	case "exists":
		return true, nil
	case "not_exists":
		return false, nil
	case "equals", "eq":
		return reflect.DeepEqual(stateValue, c.value), nil
	case "not_equals", "ne":
		return !reflect.DeepEqual(stateValue, c.value), nil
	case "greater_than", "gt":
		return c.compareNumbers(stateValue, c.value, ">")
	case "greater_equal", "ge":
		return c.compareNumbers(stateValue, c.value, ">=")
	case "less_than", "lt":
		return c.compareNumbers(stateValue, c.value, "<")
	case "less_equal", "le":
		return c.compareNumbers(stateValue, c.value, "<=")
	case "contains":
		return c.contains(stateValue, c.value)
	case "starts_with":
		return c.startsWith(stateValue, c.value)
	case "ends_with":
		return c.endsWith(stateValue, c.value)
	case "matches":
		return c.matches(stateValue, c.value)
	default:
		return false, fmt.Errorf("unsupported operator: %s", c.operator)
	}
}

// GetDescription returns a description of the condition
func (c *StateCondition) GetDescription() string {
	return fmt.Sprintf("state[%s] %s %v", c.key, c.operator, c.value)
}

func (c *StateCondition) compareNumbers(a, b interface{}, op string) (bool, error) {
	aFloat, err := c.toFloat64(a)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to number", a)
	}

	bFloat, err := c.toFloat64(b)
	if err != nil {
		return false, fmt.Errorf("cannot convert %v to number", b)
	}

	switch op {
	case ">":
		return aFloat > bFloat, nil
	case ">=":
		return aFloat >= bFloat, nil
	case "<":
		return aFloat < bFloat, nil
	case "<=":
		return aFloat <= bFloat, nil
	default:
		return false, fmt.Errorf("unsupported comparison operator: %s", op)
	}
}

func (c *StateCondition) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("cannot convert %T to float64", v)
	}
}

func (c *StateCondition) contains(haystack, needle interface{}) (bool, error) {
	haystackStr := fmt.Sprintf("%v", haystack)
	needleStr := fmt.Sprintf("%v", needle)
	return strings.Contains(haystackStr, needleStr), nil
}

func (c *StateCondition) startsWith(str, prefix interface{}) (bool, error) {
	strVal := fmt.Sprintf("%v", str)
	prefixVal := fmt.Sprintf("%v", prefix)
	return strings.HasPrefix(strVal, prefixVal), nil
}

func (c *StateCondition) endsWith(str, suffix interface{}) (bool, error) {
	strVal := fmt.Sprintf("%v", str)
	suffixVal := fmt.Sprintf("%v", suffix)
	return strings.HasSuffix(strVal, suffixVal), nil
}

func (c *StateCondition) matches(str, pattern interface{}) (bool, error) {
	// Simple pattern matching - could be extended with regex
	strVal := fmt.Sprintf("%v", str)
	patternVal := fmt.Sprintf("%v", pattern)
	return strVal == patternVal, nil
}

// FunctionCondition evaluates using a custom function
type FunctionCondition struct {
	function    func(context.Context, GraphState) (bool, error)
	description string
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewFunctionCondition creates a new function condition
func NewFunctionCondition(fn func(context.Context, GraphState) (bool, error), description string, logger *logrus.Logger) *FunctionCondition {
	return &FunctionCondition{
		function:    fn,
		description: description,
		logger:      logger,
		tracer:      otel.Tracer("langgraph.conditions.function"),
	}
}

// Evaluate evaluates the function condition
func (c *FunctionCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "function_condition.evaluate")
	defer span.End()

	return c.function(ctx, state)
}

// GetDescription returns a description of the condition
func (c *FunctionCondition) GetDescription() string {
	return c.description
}

// CompositeCondition combines multiple conditions
type CompositeCondition struct {
	conditions []Condition
	operator   string // "and", "or", "not"
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewCompositeCondition creates a new composite condition
func NewCompositeCondition(conditions []Condition, operator string, logger *logrus.Logger) *CompositeCondition {
	return &CompositeCondition{
		conditions: conditions,
		operator:   operator,
		logger:     logger,
		tracer:     otel.Tracer("langgraph.conditions.composite"),
	}
}

// Evaluate evaluates the composite condition
func (c *CompositeCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "composite_condition.evaluate")
	defer span.End()

	switch c.operator {
	case "and":
		for _, condition := range c.conditions {
			result, err := condition.Evaluate(ctx, state)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil

	case "or":
		for _, condition := range c.conditions {
			result, err := condition.Evaluate(ctx, state)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil

	case "not":
		if len(c.conditions) != 1 {
			return false, fmt.Errorf("NOT operator requires exactly one condition")
		}
		result, err := c.conditions[0].Evaluate(ctx, state)
		if err != nil {
			return false, err
		}
		return !result, nil

	default:
		return false, fmt.Errorf("unsupported composite operator: %s", c.operator)
	}
}

// GetDescription returns a description of the condition
func (c *CompositeCondition) GetDescription() string {
	descriptions := make([]string, len(c.conditions))
	for i, condition := range c.conditions {
		descriptions[i] = condition.GetDescription()
	}

	switch c.operator {
	case "and":
		return fmt.Sprintf("(%s)", strings.Join(descriptions, " AND "))
	case "or":
		return fmt.Sprintf("(%s)", strings.Join(descriptions, " OR "))
	case "not":
		return fmt.Sprintf("NOT (%s)", descriptions[0])
	default:
		return fmt.Sprintf("UNKNOWN(%s)", strings.Join(descriptions, ", "))
	}
}

// AlwaysTrueCondition always returns true
type AlwaysTrueCondition struct{}

func (c *AlwaysTrueCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	return true, nil
}

func (c *AlwaysTrueCondition) GetDescription() string {
	return "always true"
}

// AlwaysFalseCondition always returns false
type AlwaysFalseCondition struct{}

func (c *AlwaysFalseCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	return false, nil
}

func (c *AlwaysFalseCondition) GetDescription() string {
	return "always false"
}

// CountCondition checks iteration counts
type CountCondition struct {
	key       string
	operator  string
	threshold int
	logger    *logrus.Logger
	tracer    trace.Tracer
}

// NewCountCondition creates a new count condition
func NewCountCondition(key, operator string, threshold int, logger *logrus.Logger) *CountCondition {
	return &CountCondition{
		key:       key,
		operator:  operator,
		threshold: threshold,
		logger:    logger,
		tracer:    otel.Tracer("langgraph.conditions.count"),
	}
}

// Evaluate evaluates the count condition
func (c *CountCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "count_condition.evaluate")
	defer span.End()

	value, exists := state[c.key]
	if !exists {
		return false, nil
	}

	count := 0
	switch v := value.(type) {
	case int:
		count = v
	case float64:
		count = int(v)
	case []interface{}:
		count = len(v)
	case map[string]interface{}:
		count = len(v)
	case string:
		count = len(v)
	default:
		return false, fmt.Errorf("cannot get count from type %T", value)
	}

	switch c.operator {
	case "eq", "equals":
		return count == c.threshold, nil
	case "ne", "not_equals":
		return count != c.threshold, nil
	case "gt", "greater_than":
		return count > c.threshold, nil
	case "ge", "greater_equal":
		return count >= c.threshold, nil
	case "lt", "less_than":
		return count < c.threshold, nil
	case "le", "less_equal":
		return count <= c.threshold, nil
	default:
		return false, fmt.Errorf("unsupported count operator: %s", c.operator)
	}
}

// GetDescription returns a description of the condition
func (c *CountCondition) GetDescription() string {
	return fmt.Sprintf("count(%s) %s %d", c.key, c.operator, c.threshold)
}

// ErrorCondition checks for errors in state
type ErrorCondition struct {
	checkKey string
	logger   *logrus.Logger
	tracer   trace.Tracer
}

// NewErrorCondition creates a new error condition
func NewErrorCondition(checkKey string, logger *logrus.Logger) *ErrorCondition {
	return &ErrorCondition{
		checkKey: checkKey,
		logger:   logger,
		tracer:   otel.Tracer("langgraph.conditions.error"),
	}
}

// Evaluate evaluates the error condition
func (c *ErrorCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "error_condition.evaluate")
	defer span.End()

	if c.checkKey != "" {
		if errorValue, exists := state[c.checkKey]; exists {
			return errorValue != nil, nil
		}
		return false, nil
	}

	// Check for common error keys
	errorKeys := []string{"error", "err", "exception", "failure"}
	for _, key := range errorKeys {
		if errorValue, exists := state[key]; exists && errorValue != nil {
			return true, nil
		}
	}

	return false, nil
}

// GetDescription returns a description of the condition
func (c *ErrorCondition) GetDescription() string {
	if c.checkKey != "" {
		return fmt.Sprintf("has error in %s", c.checkKey)
	}
	return "has error"
}

// TimeoutCondition checks for timeouts
type TimeoutCondition struct {
	startKey string
	timeout  int64 // seconds
	logger   *logrus.Logger
	tracer   trace.Tracer
}

// NewTimeoutCondition creates a new timeout condition
func NewTimeoutCondition(startKey string, timeoutSeconds int64, logger *logrus.Logger) *TimeoutCondition {
	return &TimeoutCondition{
		startKey: startKey,
		timeout:  timeoutSeconds,
		logger:   logger,
		tracer:   otel.Tracer("langgraph.conditions.timeout"),
	}
}

// Evaluate evaluates the timeout condition
func (c *TimeoutCondition) Evaluate(ctx context.Context, state GraphState) (bool, error) {
	ctx, span := c.tracer.Start(ctx, "timeout_condition.evaluate")
	defer span.End()

	startValue, exists := state[c.startKey]
	if !exists {
		return false, nil
	}

	startTime, ok := startValue.(int64)
	if !ok {
		return false, fmt.Errorf("start time must be int64 (unix timestamp)")
	}

	currentTime := ctx.Value("current_time")
	if currentTime == nil {
		// Use current time if not provided in context
		currentTime = int64(time.Now().Unix())
	}

	currentTimeInt, ok := currentTime.(int64)
	if !ok {
		return false, fmt.Errorf("current time must be int64 (unix timestamp)")
	}

	elapsed := currentTimeInt - startTime
	return elapsed >= c.timeout, nil
}

// GetDescription returns a description of the condition
func (c *TimeoutCondition) GetDescription() string {
	return fmt.Sprintf("timeout after %d seconds from %s", c.timeout, c.startKey)
}
