package chains

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultSequentialChain implements the SequentialChain interface
type DefaultSequentialChain struct {
	chains     []Chain
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
	metadata   map[string]interface{}
}

// SequentialChainConfig represents configuration for a sequential chain
type SequentialChainConfig struct {
	Chains     []Chain                `json:"-"`
	InputKeys  []string               `json:"input_keys,omitempty"`
	OutputKeys []string               `json:"output_keys,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewSequentialChain creates a new sequential chain
func NewSequentialChain(config *SequentialChainConfig, logger *logrus.Logger) (SequentialChain, error) {
	if len(config.Chains) == 0 {
		return nil, fmt.Errorf("at least one chain is required")
	}

	// Determine input keys from the first chain if not provided
	inputKeys := config.InputKeys
	if len(inputKeys) == 0 {
		inputKeys = config.Chains[0].GetInputKeys()
	}

	// Determine output keys from the last chain if not provided
	outputKeys := config.OutputKeys
	if len(outputKeys) == 0 {
		outputKeys = config.Chains[len(config.Chains)-1].GetOutputKeys()
	}

	chain := &DefaultSequentialChain{
		chains:     config.Chains,
		inputKeys:  inputKeys,
		outputKeys: outputKeys,
		logger:     logger,
		tracer:     otel.Tracer("langchain.chains.sequential"),
		metadata:   config.Metadata,
	}

	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("chain validation failed: %w", err)
	}

	return chain, nil
}

// Run executes the chain with the given input
func (c *DefaultSequentialChain) Run(ctx context.Context, input ChainInput) (ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "sequential_chain.run")
	defer span.End()

	executionID := uuid.New().String()
	span.SetAttributes(
		attribute.String("chain.type", c.GetChainType()),
		attribute.String("chain.execution_id", executionID),
		attribute.Int("chain.count", len(c.chains)),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"chain_type":   c.GetChainType(),
		"chain_count":  len(c.chains),
	}).Info("Starting sequential chain execution")

	start := time.Now()
	currentInput := input

	// Execute chains sequentially
	for i, chain := range c.chains {
		chainSpan := trace.SpanFromContext(ctx)
		chainSpan.SetAttributes(
			attribute.Int("chain.step", i),
			attribute.String("chain.step_type", chain.GetChainType()),
		)

		c.logger.WithFields(logrus.Fields{
			"execution_id": executionID,
			"step":         i,
			"step_type":    chain.GetChainType(),
		}).Debug("Executing chain step")

		stepStart := time.Now()
		output, err := chain.Run(ctx, currentInput)
		stepDuration := time.Since(stepStart)

		if err != nil {
			span.RecordError(err)
			c.logger.WithError(err).WithFields(logrus.Fields{
				"execution_id": executionID,
				"step":         i,
				"step_type":    chain.GetChainType(),
				"duration":     stepDuration,
			}).Error("Chain step failed")
			return nil, fmt.Errorf("chain step %d (%s) failed: %w", i, chain.GetChainType(), err)
		}

		c.logger.WithFields(logrus.Fields{
			"execution_id": executionID,
			"step":         i,
			"step_type":    chain.GetChainType(),
			"duration":     stepDuration,
		}).Debug("Chain step completed")

		// Merge output with current input for next chain
		// This allows chains to pass data through the sequence
		for key, value := range output {
			currentInput[key] = value
		}
	}

	// Extract final output based on output keys
	finalOutput := make(ChainOutput)
	for _, key := range c.outputKeys {
		if value, exists := currentInput[key]; exists {
			finalOutput[key] = value
		}
	}

	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int64("chain.duration_ms", duration.Milliseconds()),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"duration":     duration,
		"output_keys":  len(finalOutput),
	}).Info("Sequential chain execution completed")

	return finalOutput, nil
}

// GetInputKeys returns the expected input keys
func (c *DefaultSequentialChain) GetInputKeys() []string {
	return c.inputKeys
}

// GetOutputKeys returns the output keys this chain produces
func (c *DefaultSequentialChain) GetOutputKeys() []string {
	return c.outputKeys
}

// GetChainType returns the type of this chain
func (c *DefaultSequentialChain) GetChainType() string {
	return "sequential"
}

// Validate validates the chain configuration
func (c *DefaultSequentialChain) Validate() error {
	if len(c.chains) == 0 {
		return fmt.Errorf("at least one chain is required")
	}

	if len(c.inputKeys) == 0 {
		return fmt.Errorf("input keys cannot be empty")
	}

	if len(c.outputKeys) == 0 {
		return fmt.Errorf("output keys cannot be empty")
	}

	// Validate each chain
	for i, chain := range c.chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("chain %d validation failed: %w", i, err)
		}
	}

	// Validate chain compatibility (output of one chain should be compatible with input of next)
	for i := 0; i < len(c.chains)-1; i++ {
		currentChain := c.chains[i]
		nextChain := c.chains[i+1]

		currentOutputKeys := currentChain.GetOutputKeys()
		nextInputKeys := nextChain.GetInputKeys()

		// Check if next chain's input keys are satisfied by current chain's output or initial input
		for _, inputKey := range nextInputKeys {
			found := false

			// Check in current chain's output
			for _, outputKey := range currentOutputKeys {
				if outputKey == inputKey {
					found = true
					break
				}
			}

			// Check in initial input keys if not found in current output
			if !found {
				for _, initialKey := range c.inputKeys {
					if initialKey == inputKey {
						found = true
						break
					}
				}
			}

			if !found {
				return fmt.Errorf("chain %d requires input key '%s' which is not provided by chain %d or initial input", i+1, inputKey, i)
			}
		}
	}

	return nil
}

// AddChain adds a chain to the sequence
func (c *DefaultSequentialChain) AddChain(chain Chain) error {
	if chain == nil {
		return fmt.Errorf("chain cannot be nil")
	}

	c.chains = append(c.chains, chain)

	// Update output keys to match the last chain
	c.outputKeys = chain.GetOutputKeys()

	return c.Validate()
}

// GetChains returns all chains in the sequence
func (c *DefaultSequentialChain) GetChains() []Chain {
	return c.chains
}

// RemoveChain removes a chain from the sequence
func (c *DefaultSequentialChain) RemoveChain(index int) error {
	if index < 0 || index >= len(c.chains) {
		return fmt.Errorf("invalid chain index: %d", index)
	}

	// Remove chain at index
	c.chains = append(c.chains[:index], c.chains[index+1:]...)

	if len(c.chains) == 0 {
		return fmt.Errorf("cannot remove last chain")
	}

	// Update output keys to match the new last chain
	c.outputKeys = c.chains[len(c.chains)-1].GetOutputKeys()

	return c.Validate()
}

// SimpleSequentialChain is a simplified version that passes output directly to next chain
type SimpleSequentialChain struct {
	*DefaultSequentialChain
}

// NewSimpleSequentialChain creates a new simple sequential chain
func NewSimpleSequentialChain(chains []Chain, logger *logrus.Logger) (*SimpleSequentialChain, error) {
	if len(chains) == 0 {
		return nil, fmt.Errorf("at least one chain is required")
	}

	// For simple sequential chains, each chain should have single input/output
	for i, chain := range chains {
		inputKeys := chain.GetInputKeys()
		outputKeys := chain.GetOutputKeys()

		if len(inputKeys) != 1 {
			return nil, fmt.Errorf("chain %d must have exactly one input key for simple sequential chain", i)
		}

		if len(outputKeys) != 1 {
			return nil, fmt.Errorf("chain %d must have exactly one output key for simple sequential chain", i)
		}
	}

	config := &SequentialChainConfig{
		Chains:     chains,
		InputKeys:  chains[0].GetInputKeys(),
		OutputKeys: chains[len(chains)-1].GetOutputKeys(),
	}

	baseChain, err := NewSequentialChain(config, logger)
	if err != nil {
		return nil, err
	}

	return &SimpleSequentialChain{
		DefaultSequentialChain: baseChain.(*DefaultSequentialChain),
	}, nil
}

// Run executes the simple sequential chain
func (c *SimpleSequentialChain) Run(ctx context.Context, input ChainInput) (ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "simple_sequential_chain.run")
	defer span.End()

	if len(c.chains) == 0 {
		return nil, fmt.Errorf("no chains to execute")
	}

	// For simple sequential chain, pass output of one chain as input to next
	currentOutput := input

	for i, chain := range c.chains {
		// For simple chains, use the single output key as input for next chain
		if i > 0 {
			// Get the output from previous chain
			prevOutputKeys := c.chains[i-1].GetOutputKeys()
			if len(prevOutputKeys) != 1 {
				return nil, fmt.Errorf("previous chain must have exactly one output key")
			}

			// Get the input key for current chain
			currentInputKeys := chain.GetInputKeys()
			if len(currentInputKeys) != 1 {
				return nil, fmt.Errorf("current chain must have exactly one input key")
			}

			// Create new input with the output from previous chain
			prevOutputKey := prevOutputKeys[0]
			currentInputKey := currentInputKeys[0]

			if value, exists := currentOutput[prevOutputKey]; exists {
				currentOutput = ChainInput{currentInputKey: value}
			} else {
				return nil, fmt.Errorf("previous chain did not produce expected output key: %s", prevOutputKey)
			}
		}

		output, err := chain.Run(ctx, currentOutput)
		if err != nil {
			return nil, fmt.Errorf("chain %d failed: %w", i, err)
		}

		currentOutput = ChainInput(output)
	}

	return ChainOutput(currentOutput), nil
}
