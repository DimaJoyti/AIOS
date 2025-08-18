package chains

import (
	"context"
	"fmt"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langchain/prompts"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultLLMChain implements the LLMChain interface
type DefaultLLMChain struct {
	llm        llm.LLM
	prompt     prompts.PromptTemplate
	inputKeys  []string
	outputKeys []string
	logger     *logrus.Logger
	tracer     trace.Tracer
	metadata   map[string]interface{}
}

// LLMChainConfig represents configuration for an LLM chain
type LLMChainConfig struct {
	LLM        llm.LLM                `json:"-"`
	Prompt     prompts.PromptTemplate `json:"-"`
	InputKeys  []string               `json:"input_keys,omitempty"`
	OutputKeys []string               `json:"output_keys,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// NewLLMChain creates a new LLM chain
func NewLLMChain(config *LLMChainConfig, logger *logrus.Logger) (LLMChain, error) {
	if config.LLM == nil {
		return nil, fmt.Errorf("LLM is required")
	}

	if config.Prompt == nil {
		return nil, fmt.Errorf("prompt template is required")
	}

	// Set default input keys from prompt if not provided
	inputKeys := config.InputKeys
	if len(inputKeys) == 0 {
		inputKeys = config.Prompt.GetInputVariables()
	}

	// Set default output keys if not provided
	outputKeys := config.OutputKeys
	if len(outputKeys) == 0 {
		outputKeys = []string{"text"}
	}

	chain := &DefaultLLMChain{
		llm:        config.LLM,
		prompt:     config.Prompt,
		inputKeys:  inputKeys,
		outputKeys: outputKeys,
		logger:     logger,
		tracer:     otel.Tracer("langchain.chains.llm"),
		metadata:   config.Metadata,
	}

	if err := chain.Validate(); err != nil {
		return nil, fmt.Errorf("chain validation failed: %w", err)
	}

	return chain, nil
}

// Run executes the chain with the given input
func (c *DefaultLLMChain) Run(ctx context.Context, input ChainInput) (ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "llm_chain.run")
	defer span.End()

	executionID := uuid.New().String()
	span.SetAttributes(
		attribute.String("chain.type", c.GetChainType()),
		attribute.String("chain.execution_id", executionID),
		attribute.String("llm.provider", string(c.llm.GetProvider())),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id": executionID,
		"chain_type":   c.GetChainType(),
		"input_keys":   c.inputKeys,
	}).Info("Starting LLM chain execution")

	start := time.Now()

	// Validate input
	if err := c.validateInput(input); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Convert input to variables for prompt formatting
	variables := make(map[string]interface{})
	for _, key := range c.inputKeys {
		if value, exists := input[key]; exists {
			variables[key] = value
		}
	}

	// Format the prompt
	promptValue, err := c.prompt.FormatPrompt(variables)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to format prompt: %w", err)
	}

	// Create completion request
	messages := promptValue.ToMessages()
	req := &llm.CompletionRequest{
		Messages: messages,
	}

	// Execute LLM request
	response, err := c.llm.Complete(ctx, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("LLM completion failed: %w", err)
	}

	// Create output
	output := ChainOutput{
		"text": response.Content,
	}

	// Add additional output keys if specified
	for _, key := range c.outputKeys {
		if key != "text" {
			switch key {
			case "model":
				output[key] = response.Model
			case "usage":
				output[key] = response.Usage
			case "metadata":
				output[key] = response.Metadata
			}
		}
	}

	duration := time.Since(start)
	span.SetAttributes(
		attribute.Int("llm.usage.prompt_tokens", response.Usage.PromptTokens),
		attribute.Int("llm.usage.completion_tokens", response.Usage.CompletionTokens),
		attribute.Int("llm.usage.total_tokens", response.Usage.TotalTokens),
		attribute.Int64("chain.duration_ms", duration.Milliseconds()),
	)

	c.logger.WithFields(logrus.Fields{
		"execution_id":      executionID,
		"duration":          duration,
		"prompt_tokens":     response.Usage.PromptTokens,
		"completion_tokens": response.Usage.CompletionTokens,
		"total_tokens":      response.Usage.TotalTokens,
	}).Info("LLM chain execution completed")

	return output, nil
}

// GetInputKeys returns the expected input keys
func (c *DefaultLLMChain) GetInputKeys() []string {
	return c.inputKeys
}

// GetOutputKeys returns the output keys this chain produces
func (c *DefaultLLMChain) GetOutputKeys() []string {
	return c.outputKeys
}

// GetChainType returns the type of this chain
func (c *DefaultLLMChain) GetChainType() string {
	return "llm"
}

// Validate validates the chain configuration
func (c *DefaultLLMChain) Validate() error {
	if c.llm == nil {
		return fmt.Errorf("LLM is required")
	}

	if c.prompt == nil {
		return fmt.Errorf("prompt template is required")
	}

	if len(c.inputKeys) == 0 {
		return fmt.Errorf("input keys cannot be empty")
	}

	if len(c.outputKeys) == 0 {
		return fmt.Errorf("output keys cannot be empty")
	}

	// Validate that prompt input variables match chain input keys
	promptVars := c.prompt.GetInputVariables()
	for _, promptVar := range promptVars {
		found := false
		for _, inputKey := range c.inputKeys {
			if inputKey == promptVar {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("prompt variable %s not found in input keys", promptVar)
		}
	}

	return nil
}

// GetLLM returns the LLM used by this chain
func (c *DefaultLLMChain) GetLLM() llm.LLM {
	return c.llm
}

// GetPrompt returns the prompt template used by this chain
func (c *DefaultLLMChain) GetPrompt() prompts.PromptTemplate {
	return c.prompt
}

// SetLLM sets the LLM for this chain
func (c *DefaultLLMChain) SetLLM(newLLM llm.LLM) error {
	if newLLM == nil {
		return fmt.Errorf("LLM cannot be nil")
	}

	c.llm = newLLM
	return c.Validate()
}

// SetPrompt sets the prompt template for this chain
func (c *DefaultLLMChain) SetPrompt(newPrompt prompts.PromptTemplate) error {
	if newPrompt == nil {
		return fmt.Errorf("prompt template cannot be nil")
	}

	c.prompt = newPrompt
	
	// Update input keys to match prompt variables
	c.inputKeys = newPrompt.GetInputVariables()
	
	return c.Validate()
}

// Helper methods

func (c *DefaultLLMChain) validateInput(input ChainInput) error {
	for _, key := range c.inputKeys {
		if _, exists := input[key]; !exists {
			return fmt.Errorf("missing required input key: %s", key)
		}
	}
	return nil
}

// StreamingLLMChain extends DefaultLLMChain with streaming support
type StreamingLLMChain struct {
	*DefaultLLMChain
}

// NewStreamingLLMChain creates a new streaming LLM chain
func NewStreamingLLMChain(config *LLMChainConfig, logger *logrus.Logger) (*StreamingLLMChain, error) {
	baseChain, err := NewLLMChain(config, logger)
	if err != nil {
		return nil, err
	}

	return &StreamingLLMChain{
		DefaultLLMChain: baseChain.(*DefaultLLMChain),
	}, nil
}

// Stream executes the chain with streaming output
func (c *StreamingLLMChain) Stream(ctx context.Context, input ChainInput) (<-chan ChainOutput, error) {
	ctx, span := c.tracer.Start(ctx, "streaming_llm_chain.stream")
	defer span.End()

	outputCh := make(chan ChainOutput, 10)

	go func() {
		defer close(outputCh)

		// Validate input
		if err := c.validateInput(input); err != nil {
			outputCh <- ChainOutput{"error": err.Error()}
			return
		}

		// Convert input to variables for prompt formatting
		variables := make(map[string]interface{})
		for _, key := range c.inputKeys {
			if value, exists := input[key]; exists {
				variables[key] = value
			}
		}

		// Format the prompt
		promptValue, err := c.prompt.FormatPrompt(variables)
		if err != nil {
			outputCh <- ChainOutput{"error": fmt.Sprintf("failed to format prompt: %v", err)}
			return
		}

		// Create completion request
		messages := promptValue.ToMessages()
		req := &llm.CompletionRequest{
			Messages: messages,
			Stream:   true,
		}

		// Execute streaming LLM request
		streamCh, err := c.llm.Stream(ctx, req)
		if err != nil {
			outputCh <- ChainOutput{"error": fmt.Sprintf("LLM streaming failed: %v", err)}
			return
		}

		// Process streaming responses
		var fullContent string
		for streamResp := range streamCh {
			if streamResp.Error != nil {
				outputCh <- ChainOutput{"error": streamResp.Error.Error()}
				return
			}

			fullContent += streamResp.Content

			output := ChainOutput{
				"text":    streamResp.Content,
				"partial": !streamResp.Done,
				"done":    streamResp.Done,
			}

			if streamResp.Done {
				output["full_text"] = fullContent
			}

			outputCh <- output
		}
	}()

	return outputCh, nil
}
