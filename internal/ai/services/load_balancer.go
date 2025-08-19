package services

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// LoadBalancer manages load balancing across AI models
type LoadBalancer struct {
	strategies map[string]LoadBalancingStrategy
	metrics    map[string]*ModelMetrics
	mutex      sync.RWMutex
	logger     *logrus.Logger
}

// LoadBalancingStrategy defines the interface for load balancing strategies
type LoadBalancingStrategy interface {
	SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error)
	GetName() string
}

// ModelMetrics represents performance metrics for a model
type ModelMetrics struct {
	ModelID         string        `json:"model_id"`
	RequestCount    int64         `json:"request_count"`
	SuccessCount    int64         `json:"success_count"`
	ErrorCount      int64         `json:"error_count"`
	AverageLatency  time.Duration `json:"average_latency"`
	TotalLatency    time.Duration `json:"total_latency"`
	LastRequestTime time.Time     `json:"last_request_time"`
	CurrentLoad     int           `json:"current_load"`
	MaxLoad         int           `json:"max_load"`
	SuccessRate     float64       `json:"success_rate"`
	CostPerRequest  float64       `json:"cost_per_request"`
	TotalCost       float64       `json:"total_cost"`
	HealthScore     float64       `json:"health_score"`
	LastUpdate      time.Time     `json:"last_update"`
}

// RoundRobinStrategy implements round-robin load balancing
type RoundRobinStrategy struct {
	currentIndex int
	mutex        sync.Mutex
}

// WeightedRoundRobinStrategy implements weighted round-robin load balancing
type WeightedRoundRobinStrategy struct {
	currentWeights map[string]int
	mutex          sync.Mutex
}

// LeastConnectionsStrategy implements least connections load balancing
type LeastConnectionsStrategy struct{}

// PerformanceBasedStrategy implements performance-based load balancing
type PerformanceBasedStrategy struct{}

// CostOptimizedStrategy implements cost-optimized load balancing
type CostOptimizedStrategy struct{}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(logger *logrus.Logger) *LoadBalancer {
	lb := &LoadBalancer{
		strategies: make(map[string]LoadBalancingStrategy),
		metrics:    make(map[string]*ModelMetrics),
		logger:     logger,
	}

	// Register default strategies
	lb.RegisterStrategy(&RoundRobinStrategy{})
	lb.RegisterStrategy(&WeightedRoundRobinStrategy{currentWeights: make(map[string]int)})
	lb.RegisterStrategy(&LeastConnectionsStrategy{})
	lb.RegisterStrategy(&PerformanceBasedStrategy{})
	lb.RegisterStrategy(&CostOptimizedStrategy{})

	return lb
}

// RegisterStrategy registers a load balancing strategy
func (lb *LoadBalancer) RegisterStrategy(strategy LoadBalancingStrategy) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	lb.strategies[strategy.GetName()] = strategy
	lb.logger.WithField("strategy", strategy.GetName()).Info("Load balancing strategy registered")
}

// SelectModel selects the best model based on requirements
func (lb *LoadBalancer) SelectModel(models []*AIModel, requirements map[string]interface{}) (*AIModel, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Get strategy from requirements, default to performance-based
	strategyName := "performance-based"
	if strategy, ok := requirements["load_balancing_strategy"].(string); ok {
		strategyName = strategy
	}

	strategy, exists := lb.strategies[strategyName]
	if !exists {
		strategy = lb.strategies["performance-based"]
	}

	lb.mutex.RLock()
	metrics := make(map[string]*ModelMetrics)
	for k, v := range lb.metrics {
		metrics[k] = v
	}
	lb.mutex.RUnlock()

	return strategy.SelectModel(models, metrics, requirements)
}

// UpdateMetrics updates metrics for a model
func (lb *LoadBalancer) UpdateMetrics(modelID string, latency time.Duration, cost float64, success bool) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	metrics, exists := lb.metrics[modelID]
	if !exists {
		metrics = &ModelMetrics{
			ModelID:    modelID,
			MaxLoad:    100, // Default max load
			LastUpdate: time.Now(),
		}
		lb.metrics[modelID] = metrics
	}

	// Update metrics
	metrics.RequestCount++
	metrics.TotalLatency += latency
	metrics.AverageLatency = time.Duration(int64(metrics.TotalLatency) / metrics.RequestCount)
	metrics.TotalCost += cost
	metrics.CostPerRequest = metrics.TotalCost / float64(metrics.RequestCount)
	metrics.LastRequestTime = time.Now()
	metrics.LastUpdate = time.Now()

	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	if metrics.RequestCount > 0 {
		metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.RequestCount)
	}

	// Calculate health score (0-1, higher is better)
	metrics.HealthScore = lb.calculateHealthScore(metrics)
}

// GetMetrics returns metrics for a model
func (lb *LoadBalancer) GetMetrics(modelID string) *ModelMetrics {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	if metrics, exists := lb.metrics[modelID]; exists {
		return metrics
	}
	return nil
}

// GetAllMetrics returns metrics for all models
func (lb *LoadBalancer) GetAllMetrics() map[string]*ModelMetrics {
	lb.mutex.RLock()
	defer lb.mutex.RUnlock()

	result := make(map[string]*ModelMetrics)
	for k, v := range lb.metrics {
		result[k] = v
	}
	return result
}

// calculateHealthScore calculates a health score for a model
func (lb *LoadBalancer) calculateHealthScore(metrics *ModelMetrics) float64 {
	// Factors: success rate (40%), latency (30%), load (20%), recency (10%)

	successScore := metrics.SuccessRate

	// Latency score (lower latency = higher score)
	latencyScore := 1.0
	if metrics.AverageLatency > 0 {
		// Normalize latency (assume 1 second is baseline)
		latencyScore = math.Max(0, 1.0-float64(metrics.AverageLatency.Milliseconds())/1000.0)
	}

	// Load score (lower load = higher score)
	loadScore := 1.0
	if metrics.MaxLoad > 0 {
		loadScore = math.Max(0, 1.0-float64(metrics.CurrentLoad)/float64(metrics.MaxLoad))
	}

	// Recency score (more recent = higher score)
	recencyScore := 1.0
	if !metrics.LastRequestTime.IsZero() {
		timeSince := time.Since(metrics.LastRequestTime)
		// Decay over 1 hour
		recencyScore = math.Max(0, 1.0-float64(timeSince.Minutes())/60.0)
	}

	return 0.4*successScore + 0.3*latencyScore + 0.2*loadScore + 0.1*recencyScore
}

// RoundRobin Strategy Implementation

func (rr *RoundRobinStrategy) GetName() string {
	return "round-robin"
}

func (rr *RoundRobinStrategy) SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error) {
	rr.mutex.Lock()
	defer rr.mutex.Unlock()

	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Filter active models
	activeModels := make([]*AIModel, 0)
	for _, model := range models {
		if model.Status == "active" {
			activeModels = append(activeModels, model)
		}
	}

	if len(activeModels) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Select next model in round-robin fashion
	selectedModel := activeModels[rr.currentIndex%len(activeModels)]
	rr.currentIndex++

	return selectedModel, nil
}

// Weighted Round Robin Strategy Implementation

func (wrr *WeightedRoundRobinStrategy) GetName() string {
	return "weighted-round-robin"
}

func (wrr *WeightedRoundRobinStrategy) SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error) {
	wrr.mutex.Lock()
	defer wrr.mutex.Unlock()

	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Filter active models and calculate weights
	type weightedModel struct {
		model  *AIModel
		weight int
	}

	var weightedModels []weightedModel
	for _, model := range models {
		if model.Status != "active" {
			continue
		}

		// Calculate weight based on health score
		weight := 1
		if modelMetrics, exists := metrics[model.ID]; exists {
			weight = int(modelMetrics.HealthScore*10) + 1 // 1-11 range
		}

		weightedModels = append(weightedModels, weightedModel{model, weight})
	}

	if len(weightedModels) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Select model based on current weights
	for _, wm := range weightedModels {
		currentWeight := wrr.currentWeights[wm.model.ID]
		if currentWeight <= 0 {
			wrr.currentWeights[wm.model.ID] = wm.weight - 1
			return wm.model, nil
		}
		wrr.currentWeights[wm.model.ID] = currentWeight - 1
	}

	// Reset weights and select first model
	for _, wm := range weightedModels {
		wrr.currentWeights[wm.model.ID] = wm.weight
	}

	return weightedModels[0].model, nil
}

// Least Connections Strategy Implementation

func (lc *LeastConnectionsStrategy) GetName() string {
	return "least-connections"
}

func (lc *LeastConnectionsStrategy) SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	var bestModel *AIModel
	minLoad := math.MaxInt32

	for _, model := range models {
		if model.Status != "active" {
			continue
		}

		load := 0
		if modelMetrics, exists := metrics[model.ID]; exists {
			load = modelMetrics.CurrentLoad
		}

		if load < minLoad {
			minLoad = load
			bestModel = model
		}
	}

	if bestModel == nil {
		return nil, fmt.Errorf("no active models available")
	}

	return bestModel, nil
}

// Performance Based Strategy Implementation

func (pb *PerformanceBasedStrategy) GetName() string {
	return "performance-based"
}

func (pb *PerformanceBasedStrategy) SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Score models based on performance metrics
	type scoredModel struct {
		model *AIModel
		score float64
	}

	var scoredModels []scoredModel

	for _, model := range models {
		if model.Status != "active" {
			continue
		}

		score := 0.5 // Base score for active models
		if modelMetrics, exists := metrics[model.ID]; exists {
			score = modelMetrics.HealthScore
		}

		scoredModels = append(scoredModels, scoredModel{model, score})
	}

	if len(scoredModels) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Sort by score (highest first)
	sort.Slice(scoredModels, func(i, j int) bool {
		return scoredModels[i].score > scoredModels[j].score
	})

	return scoredModels[0].model, nil
}

// Cost Optimized Strategy Implementation

func (co *CostOptimizedStrategy) GetName() string {
	return "cost-optimized"
}

func (co *CostOptimizedStrategy) SelectModel(models []*AIModel, metrics map[string]*ModelMetrics, requirements map[string]interface{}) (*AIModel, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	// Score models based on cost efficiency (performance per cost)
	type costScoredModel struct {
		model *AIModel
		score float64
	}

	var costScoredModels []costScoredModel

	for _, model := range models {
		if model.Status != "active" {
			continue
		}

		score := 1.0 // Base score
		if modelMetrics, exists := metrics[model.ID]; exists && modelMetrics.CostPerRequest > 0 {
			// Higher success rate and lower cost = higher score
			score = modelMetrics.SuccessRate / modelMetrics.CostPerRequest
		}

		costScoredModels = append(costScoredModels, costScoredModel{model, score})
	}

	if len(costScoredModels) == 0 {
		return nil, fmt.Errorf("no active models available")
	}

	// Sort by cost efficiency (highest first)
	sort.Slice(costScoredModels, func(i, j int) bool {
		return costScoredModels[i].score > costScoredModels[j].score
	})

	return costScoredModels[0].model, nil
}
