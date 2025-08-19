package langgraph

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Router defines the interface for routing decisions
type Router interface {
	// Route determines the next node(s) to execute
	Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error)

	// GetType returns the router type
	GetType() string

	// Configure configures the router with options
	Configure(options map[string]interface{}) error
}

// ConditionalRouter routes based on edge conditions
type ConditionalRouter struct {
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewConditionalRouter creates a new conditional router
func NewConditionalRouter(logger *logrus.Logger) *ConditionalRouter {
	return &ConditionalRouter{
		logger: logger,
		tracer: otel.Tracer("langgraph.routing.conditional"),
	}
}

// Route determines the next nodes based on conditions
func (r *ConditionalRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "conditional_router.route")
	defer span.End()

	var nextNodes []string

	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		condition := edge.GetCondition()
		if condition == nil {
			// No condition means always route
			nextNodes = append(nextNodes, edge.GetTo())
			continue
		}

		matches, err := condition.Evaluate(ctx, state)
		if err != nil {
			r.logger.WithError(err).WithFields(logrus.Fields{
				"from": edge.GetFrom(),
				"to":   edge.GetTo(),
			}).Error("Failed to evaluate edge condition")
			continue
		}

		if matches {
			nextNodes = append(nextNodes, edge.GetTo())
		}
	}

	return nextNodes, nil
}

// GetType returns the router type
func (r *ConditionalRouter) GetType() string {
	return "conditional"
}

// Configure configures the router
func (r *ConditionalRouter) Configure(options map[string]interface{}) error {
	// Conditional router doesn't need configuration
	return nil
}

// WeightedRouter routes based on edge weights
type WeightedRouter struct {
	random *rand.Rand
	logger *logrus.Logger
	tracer trace.Tracer
	mu     sync.Mutex
}

// NewWeightedRouter creates a new weighted router
func NewWeightedRouter(logger *logrus.Logger) *WeightedRouter {
	return &WeightedRouter{
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
		logger: logger,
		tracer: otel.Tracer("langgraph.routing.weighted"),
	}
}

// Route determines the next node based on weights
func (r *WeightedRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "weighted_router.route")
	defer span.End()

	var validEdges []Edge
	var totalWeight float64

	// Collect valid edges and calculate total weight
	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		// Check condition if present
		condition := edge.GetCondition()
		if condition != nil {
			matches, err := condition.Evaluate(ctx, state)
			if err != nil {
				r.logger.WithError(err).Error("Failed to evaluate edge condition")
				continue
			}
			if !matches {
				continue
			}
		}

		validEdges = append(validEdges, edge)
		totalWeight += edge.GetWeight()
	}

	if len(validEdges) == 0 {
		return []string{}, nil
	}

	if totalWeight <= 0 {
		// If no weights, select randomly
		r.mu.Lock()
		selectedIndex := r.random.Intn(len(validEdges))
		r.mu.Unlock()
		return []string{validEdges[selectedIndex].GetTo()}, nil
	}

	// Weighted random selection
	r.mu.Lock()
	randomValue := r.random.Float64() * totalWeight
	r.mu.Unlock()

	currentWeight := 0.0
	for _, edge := range validEdges {
		currentWeight += edge.GetWeight()
		if randomValue <= currentWeight {
			return []string{edge.GetTo()}, nil
		}
	}

	// Fallback to last edge
	return []string{validEdges[len(validEdges)-1].GetTo()}, nil
}

// GetType returns the router type
func (r *WeightedRouter) GetType() string {
	return "weighted"
}

// Configure configures the router
func (r *WeightedRouter) Configure(options map[string]interface{}) error {
	if seed, exists := options["seed"]; exists {
		if seedInt, ok := seed.(int64); ok {
			r.mu.Lock()
			r.random = rand.New(rand.NewSource(seedInt))
			r.mu.Unlock()
		}
	}
	return nil
}

// PriorityRouter routes based on edge priorities
type PriorityRouter struct {
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewPriorityRouter creates a new priority router
func NewPriorityRouter(logger *logrus.Logger) *PriorityRouter {
	return &PriorityRouter{
		logger: logger,
		tracer: otel.Tracer("langgraph.routing.priority"),
	}
}

// Route determines the next node based on priority
func (r *PriorityRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "priority_router.route")
	defer span.End()

	var validEdges []Edge

	// Collect valid edges
	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		// Check condition if present
		condition := edge.GetCondition()
		if condition != nil {
			matches, err := condition.Evaluate(ctx, state)
			if err != nil {
				r.logger.WithError(err).Error("Failed to evaluate edge condition")
				continue
			}
			if !matches {
				continue
			}
		}

		validEdges = append(validEdges, edge)
	}

	if len(validEdges) == 0 {
		return []string{}, nil
	}

	// Sort by weight (using weight as priority, higher = higher priority)
	sort.Slice(validEdges, func(i, j int) bool {
		return validEdges[i].GetWeight() > validEdges[j].GetWeight()
	})

	// Return highest priority edge
	return []string{validEdges[0].GetTo()}, nil
}

// GetType returns the router type
func (r *PriorityRouter) GetType() string {
	return "priority"
}

// Configure configures the router
func (r *PriorityRouter) Configure(options map[string]interface{}) error {
	// Priority router doesn't need configuration
	return nil
}

// RoundRobinRouter distributes load evenly
type RoundRobinRouter struct {
	counters map[string]int
	logger   *logrus.Logger
	tracer   trace.Tracer
	mu       sync.Mutex
}

// NewRoundRobinRouter creates a new round-robin router
func NewRoundRobinRouter(logger *logrus.Logger) *RoundRobinRouter {
	return &RoundRobinRouter{
		counters: make(map[string]int),
		logger:   logger,
		tracer:   otel.Tracer("langgraph.routing.round_robin"),
	}
}

// Route determines the next node using round-robin
func (r *RoundRobinRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "round_robin_router.route")
	defer span.End()

	var validEdges []Edge

	// Collect valid edges
	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		// Check condition if present
		condition := edge.GetCondition()
		if condition != nil {
			matches, err := condition.Evaluate(ctx, state)
			if err != nil {
				r.logger.WithError(err).Error("Failed to evaluate edge condition")
				continue
			}
			if !matches {
				continue
			}
		}

		validEdges = append(validEdges, edge)
	}

	if len(validEdges) == 0 {
		return []string{}, nil
	}

	// Round-robin selection
	r.mu.Lock()
	counter := r.counters[currentNode]
	selectedIndex := counter % len(validEdges)
	r.counters[currentNode] = counter + 1
	r.mu.Unlock()

	return []string{validEdges[selectedIndex].GetTo()}, nil
}

// GetType returns the router type
func (r *RoundRobinRouter) GetType() string {
	return "round_robin"
}

// Configure configures the router
func (r *RoundRobinRouter) Configure(options map[string]interface{}) error {
	if reset, exists := options["reset"]; exists {
		if resetBool, ok := reset.(bool); ok && resetBool {
			r.mu.Lock()
			r.counters = make(map[string]int)
			r.mu.Unlock()
		}
	}
	return nil
}

// ParallelRouter routes to multiple nodes simultaneously
type ParallelRouter struct {
	maxParallel int
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewParallelRouter creates a new parallel router
func NewParallelRouter(maxParallel int, logger *logrus.Logger) *ParallelRouter {
	if maxParallel <= 0 {
		maxParallel = 10 // Default max parallel
	}

	return &ParallelRouter{
		maxParallel: maxParallel,
		logger:      logger,
		tracer:      otel.Tracer("langgraph.routing.parallel"),
	}
}

// Route determines multiple next nodes for parallel execution
func (r *ParallelRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "parallel_router.route")
	defer span.End()

	var nextNodes []string

	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		// Check condition if present
		condition := edge.GetCondition()
		if condition != nil {
			matches, err := condition.Evaluate(ctx, state)
			if err != nil {
				r.logger.WithError(err).Error("Failed to evaluate edge condition")
				continue
			}
			if !matches {
				continue
			}
		}

		nextNodes = append(nextNodes, edge.GetTo())

		// Limit parallel execution
		if len(nextNodes) >= r.maxParallel {
			break
		}
	}

	return nextNodes, nil
}

// GetType returns the router type
func (r *ParallelRouter) GetType() string {
	return "parallel"
}

// Configure configures the router
func (r *ParallelRouter) Configure(options map[string]interface{}) error {
	if maxParallel, exists := options["max_parallel"]; exists {
		if maxParallelInt, ok := maxParallel.(int); ok && maxParallelInt > 0 {
			r.maxParallel = maxParallelInt
		}
	}
	return nil
}

// HealthBasedRouter routes based on node health
type HealthBasedRouter struct {
	healthChecker  func(nodeID string) bool
	fallbackRouter Router
	logger         *logrus.Logger
	tracer         trace.Tracer
}

// NewHealthBasedRouter creates a new health-based router
func NewHealthBasedRouter(healthChecker func(string) bool, fallbackRouter Router, logger *logrus.Logger) *HealthBasedRouter {
	return &HealthBasedRouter{
		healthChecker:  healthChecker,
		fallbackRouter: fallbackRouter,
		logger:         logger,
		tracer:         otel.Tracer("langgraph.routing.health"),
	}
}

// Route determines the next node based on health
func (r *HealthBasedRouter) Route(ctx context.Context, currentNode string, state GraphState, edges []Edge) ([]string, error) {
	ctx, span := r.tracer.Start(ctx, "health_router.route")
	defer span.End()

	var healthyEdges []Edge

	// Filter edges to healthy nodes
	for _, edge := range edges {
		if edge.GetFrom() != currentNode {
			continue
		}

		// Check condition if present
		condition := edge.GetCondition()
		if condition != nil {
			matches, err := condition.Evaluate(ctx, state)
			if err != nil {
				r.logger.WithError(err).Error("Failed to evaluate edge condition")
				continue
			}
			if !matches {
				continue
			}
		}

		// Check node health
		if r.healthChecker != nil && !r.healthChecker(edge.GetTo()) {
			r.logger.WithField("node_id", edge.GetTo()).Warn("Node is unhealthy, skipping")
			continue
		}

		healthyEdges = append(healthyEdges, edge)
	}

	if len(healthyEdges) == 0 {
		// No healthy nodes, use fallback router with all edges
		if r.fallbackRouter != nil {
			return r.fallbackRouter.Route(ctx, currentNode, state, edges)
		}
		return []string{}, fmt.Errorf("no healthy nodes available and no fallback router")
	}

	// Use fallback router with healthy edges
	if r.fallbackRouter != nil {
		return r.fallbackRouter.Route(ctx, currentNode, state, healthyEdges)
	}

	// Default to first healthy edge
	return []string{healthyEdges[0].GetTo()}, nil
}

// GetType returns the router type
func (r *HealthBasedRouter) GetType() string {
	return "health_based"
}

// Configure configures the router
func (r *HealthBasedRouter) Configure(options map[string]interface{}) error {
	// Delegate to fallback router if available
	if r.fallbackRouter != nil {
		return r.fallbackRouter.Configure(options)
	}
	return nil
}

// RouterFactory creates routers
type RouterFactory struct {
	logger *logrus.Logger
}

// NewRouterFactory creates a new router factory
func NewRouterFactory(logger *logrus.Logger) *RouterFactory {
	return &RouterFactory{
		logger: logger,
	}
}

// CreateRouter creates a router of the specified type
func (f *RouterFactory) CreateRouter(routerType string, options map[string]interface{}) (Router, error) {
	switch routerType {
	case "conditional":
		return NewConditionalRouter(f.logger), nil
	case "weighted":
		return NewWeightedRouter(f.logger), nil
	case "priority":
		return NewPriorityRouter(f.logger), nil
	case "round_robin":
		return NewRoundRobinRouter(f.logger), nil
	case "parallel":
		maxParallel := 10
		if mp, exists := options["max_parallel"]; exists {
			if mpInt, ok := mp.(int); ok {
				maxParallel = mpInt
			}
		}
		return NewParallelRouter(maxParallel, f.logger), nil
	case "health_based":
		var healthChecker func(string) bool
		var fallbackRouter Router

		if hc, exists := options["health_checker"]; exists {
			if hcFunc, ok := hc.(func(string) bool); ok {
				healthChecker = hcFunc
			}
		}

		if fr, exists := options["fallback_router"]; exists {
			if frRouter, ok := fr.(Router); ok {
				fallbackRouter = frRouter
			}
		}

		return NewHealthBasedRouter(healthChecker, fallbackRouter, f.logger), nil
	default:
		return nil, fmt.Errorf("unsupported router type: %s", routerType)
	}
}
