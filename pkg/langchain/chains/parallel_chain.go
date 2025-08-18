package chains

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
	"golang.org/x/time/rate"
)

// DefaultParallelChain implements the ParallelChain interface
type DefaultParallelChain struct {
	chains         map[string]Chain
	inputKeys      []string
	outputKeys     []string
	maxConcurrency int
	limiter        *rate.Limiter
	logger         *logrus.Logger
	tracer         trace.Tracer
	metadata       map[string]interface{}
	mu             sync.RWMutex
}

// ParallelChainConfig represents configuration for a parallel chain
type ParallelChainConfig struct {
	Chains         map[string]Chain       `json:"-"`
	InputKeys      []string               `json:"input_keys,omitempty"`
	OutputKeys     []string               `json:"output_keys,omitempty"`
	MaxConcurrency int                    `json:"max_concurrency,omitempty"`
	RateLimit      float64                `json:"rate_limit,omitempty"` // requests per second
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ChainResult represents the result of a single chain execution
type ChainResult struct {
	Name     string
	Output   ChainOutput
	Error    error
	Duration time.Duration
}

// NewParallelChain creates a new parallel chain
func NewParallelChain(config *ParallelChainConfig, logger *logrus.Logger) (ParallelChain, error) {
	if len(config.Chains) == 0 {
		return nil, fmt.Errorf("at least one chain is required")
	}

	// Set default max concurrency
	maxConcurrency := config.MaxConcurrency
	if maxConcurrency <= 0 {
		maxConcurrency = len(config.Chains) // Default to number of chains
	}

	// Determine input keys from all chains if not provided
	inputKeys := config.InputKeys
	if len(inputKeys) == 0 {
		inputKeySet := make(map[string]bool)
		for _, chain := range config.Chains {
			for _, key := range chain.GetInputKeys() {
				inputKeySet[key] = true
			}
		}
		
		for key := range inputKeySet {
			inputKeys = append(inputKeys, key)
		}
	}

	// Determine output keys from all chains if not provided
	outputKeys := config.OutputKeys
	if len(outputKeys) == 0 {
		for name, chain := range config.Chains {
			for _, key := range chain.GetOutputKeys() {
				outputKeys = append(outputKeys, fmt.Sprintf("%s_%s", name, key))
			}
		}
	}

	// Create rate limiter if specified
	var limiter *rate.Limiter
	if config.RateLimit > 0 {
		limiter = rate.NewLimiter(rate.Limit(config.RateLimit), maxConcurrency)
	}

	chain := &DefaultParallelChain{
		chains:         config.Chains,
		inputKeys:      inputKeys,
		outputKeys:     outputKeys,
		maxConcurrency: maxConcurrency,
		limiter:        limiter,
		logger:         logger,
		tracer:         otel.Tracer("langchain.chains.parallel"),
		metadata:       config.Metadata,
	}

	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("chain validation failed: %w", err)
	}

	return chain, nil
}

// Run executes the chain with the given input
func (c *DefaultParallelChain) Run(ctx context.Context, input ChainInput) (ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "parallel_chain.run")
	defer span.End()

	executionID := uuid.New().String()
	span.SetAttributes(
		attribute.String("chain.type", c.GetChainType()),
		attribute.String("chain.execution_id", executionID),
		attribute.Int("chain.count", len(c.chains)),
		attribute.Int("chain.max_concurrency", c.maxConcurrency),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id":    executionID,
		"chain_type":      c.GetChainType(),
		"chain_count":     len(c.chains),
		"max_concurrency": c.maxConcurrency,
	}).Info("Starting parallel chain execution")

	start := time.Now()

	// Create channels for results
	resultCh := make(chan ChainResult, len(c.chains))
	
	// Create semaphore for concurrency control
	semaphore := make(chan struct{}, c.maxConcurrency)

	// Start goroutines for each chain
	var wg sync.WaitGroup
	c.mu.RLock()
	chains := make(map[string]Chain)
	for name, chain := range c.chains {
		chains[name] = chain
	}
	c.mu.RUnlock()

	for name, chain := range chains {
		wg.Add(1)
		go func(chainName string, ch Chain) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Apply rate limiting if configured
			if c.limiter != nil {
				if err := c.limiter.Wait(ctx); err != nil {
					resultCh <- ChainResult{
						Name:  chainName,
						Error: fmt.Errorf("rate limit wait failed: %w", err),
					}
					return
				}
			}

			c.logger.WithFields(logrus.Fields{
				"execution_id": executionID,
				"chain_name":   chainName,
				"chain_type":   ch.GetChainType(),
			}).Debug("Executing parallel chain")

			chainStart := time.Now()
			output, err := ch.Run(ctx, input)
			duration := time.Since(chainStart)

			result := ChainResult{
				Name:     chainName,
				Output:   output,
				Error:    err,
				Duration: duration,
			}

			if err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"execution_id": executionID,
					"chain_name":   chainName,
					"duration":     duration,
				}).Error("Parallel chain failed")
			} else {
				c.logger.WithFields(logrus.Fields{
					"execution_id": executionID,
					"chain_name":   chainName,
					"duration":     duration,
				}).Debug("Parallel chain completed")
			}

			resultCh <- result
		}(name, chain)
	}

	// Close result channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	results := make(map[string]ChainResult)
	var errors []error

	for result := range resultCh {
		results[result.Name] = result
		if result.Error != nil {
			errors = append(errors, fmt.Errorf("chain %s failed: %w", result.Name, result.Error))
		}
	}

	// Check if any chains failed
	if len(errors) > 0 {
		span.RecordError(fmt.Errorf("parallel chain execution failed: %d chains failed", len(errors)))
		return nil, fmt.Errorf("parallel chain execution failed: %d chains failed", len(errors))
	}

	// Combine outputs
	finalOutput := make(ChainOutput)
	for name, result := range results {
		for key, value := range result.Output {
			// Prefix output keys with chain name to avoid conflicts
			outputKey := fmt.Sprintf("%s_%s", name, key)
			finalOutput[outputKey] = value
		}
		
		// Also add the raw output under the chain name
		finalOutput[name] = result.Output
	}

	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("chain.duration_ms", duration.Milliseconds()),
		attribute.Int("chain.successful_count", len(results)),
		attribute.Int("chain.failed_count", len(errors)),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id":     executionID,
		"duration":         duration,
		"successful_count": len(results),
		"failed_count":     len(errors),
	}).Info("Parallel chain execution completed")

	return finalOutput, nil
}

// GetInputKeys returns the expected input keys
func (c *DefaultParallelChain) GetInputKeys() []string {
	return c.inputKeys
}

// GetOutputKeys returns the output keys this chain produces
func (c *DefaultParallelChain) GetOutputKeys() []string {
	return c.outputKeys
}

// GetChainType returns the type of this chain
func (c *DefaultParallelChain) GetChainType() string {
	return "parallel"
}

// Validate validates the chain configuration
func (c *DefaultParallelChain) Validate() error {
	if len(c.chains) == 0 {
		return fmt.Errorf("at least one chain is required")
	}

	if len(c.inputKeys) == 0 {
		return fmt.Errorf("input keys cannot be empty")
	}

	if len(c.outputKeys) == 0 {
		return fmt.Errorf("output keys cannot be empty")
	}

	if c.maxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	// Validate each chain
	for name, chain := range c.chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain %s validation failed: %w", name, err)
		}
	}

	return nil
}

// AddChain adds a chain to run in parallel
func (c *DefaultParallelChain) AddChain(name string, chain Chain) error {
	if name == "" {
		return fmt.Errorf("chain name cannot be empty")
	}

	if chain == nil {
		return fmt.Errorf("chain cannot be nil")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.chains[name]; exists {
		return fmt.Errorf("chain with name %s already exists", name)
	}

	c.chains[name] = chain

	// Update input keys to include new chain's input keys
	for _, key := range chain.GetInputKeys() {
		found := false
		for _, existingKey := range c.inputKeys {
			if existingKey == key {
				found = true
				break
			}
		}
		if !found {
			c.inputKeys = append(c.inputKeys, key)
		}
	}

	// Update output keys to include new chain's output keys
	for _, key := range chain.GetOutputKeys() {
		outputKey := fmt.Sprintf("%s_%s", name, key)
		c.outputKeys = append(c.outputKeys, outputKey)
	}

	return c.Validate()
}

// GetChains returns all chains
func (c *DefaultParallelChain) GetChains() map[string]Chain {
	c.mu.RLock()
	defer c.mu.RUnlock()

	chains := make(map[string]Chain)
	for name, chain := range c.chains {
		chains[name] = chain
	}
	return chains
}

// RemoveChain removes a chain
func (c *DefaultParallelChain) RemoveChain(name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.chains[name]; !exists {
		return fmt.Errorf("chain with name %s does not exist", name)
	}

	delete(c.chains, name)

	if len(c.chains) == 0 {
		return fmt.Errorf("cannot remove last chain")
	}

	// Rebuild input and output keys
	c.rebuildKeys()

	return c.Validate()
}

// SetMaxConcurrency sets the maximum number of concurrent executions
func (c *DefaultParallelChain) SetMaxConcurrency(max int) {
	if max <= 0 {
		max = len(c.chains)
	}
	c.maxConcurrency = max
}

// Helper methods

func (c *DefaultParallelChain) rebuildKeys() {
	// Rebuild input keys
	inputKeySet := make(map[string]bool)
	for _, chain := range c.chains {
		for _, key := range chain.GetInputKeys() {
			inputKeySet[key] = true
		}
	}

	c.inputKeys = make([]string, 0, len(inputKeySet))
	for key := range inputKeySet {
		c.inputKeys = append(c.inputKeys, key)
	}

	// Rebuild output keys
	c.outputKeys = make([]string, 0)
	for name, chain := range c.chains {
		for _, key := range chain.GetOutputKeys() {
			outputKey := fmt.Sprintf("%s_%s", name, key)
			c.outputKeys = append(c.outputKeys, outputKey)
		}
	}
}
