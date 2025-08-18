package chains

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/aios/aios/pkg/langchain/prompts"
)

// ChainInput represents input to a chain
type ChainInput map[string]interface{}

// ChainOutput represents output from a chain
type ChainOutput map[string]interface{}

// Chain defines the interface for all chains
type Chain interface {
	// Run executes the chain with the given input
	Run(ctx context.Context, input ChainInput) (ChainOutput, error)

	// GetInputKeys returns the expected input keys
	GetInputKeys() []string

	// GetOutputKeys returns the output keys this chain produces
	GetOutputKeys() []string

	// GetChainType returns the type of this chain
	GetChainType() string

	// Validate validates the chain configuration
	Validate() error
}

// LLMChain represents a chain that uses an LLM
type LLMChain interface {
	Chain

	// GetLLM returns the LLM used by this chain
	GetLLM() llm.LLM

	// GetPrompt returns the prompt template used by this chain
	GetPrompt() prompts.PromptTemplate

	// SetLLM sets the LLM for this chain
	SetLLM(llm llm.LLM) error

	// SetPrompt sets the prompt template for this chain
	SetPrompt(prompt prompts.PromptTemplate) error
}

// SequentialChain represents a chain that runs multiple chains in sequence
type SequentialChain interface {
	Chain

	// AddChain adds a chain to the sequence
	AddChain(chain Chain) error

	// GetChains returns all chains in the sequence
	GetChains() []Chain

	// RemoveChain removes a chain from the sequence
	RemoveChain(index int) error
}

// ParallelChain represents a chain that runs multiple chains in parallel
type ParallelChain interface {
	Chain

	// AddChain adds a chain to run in parallel
	AddChain(name string, chain Chain) error

	// GetChains returns all chains
	GetChains() map[string]Chain

	// RemoveChain removes a chain
	RemoveChain(name string) error

	// SetMaxConcurrency sets the maximum number of concurrent executions
	SetMaxConcurrency(max int)
}

// ConditionalChain represents a chain with conditional execution
type ConditionalChain interface {
	Chain

	// AddCondition adds a condition and associated chain
	AddCondition(condition ChainCondition, chain Chain) error

	// SetDefaultChain sets the default chain to run if no conditions match
	SetDefaultChain(chain Chain) error

	// GetConditions returns all conditions and their chains
	GetConditions() []ConditionChainPair
}

// ChainCondition represents a condition for conditional chains
type ChainCondition interface {
	// Evaluate evaluates the condition with the given input
	Evaluate(ctx context.Context, input ChainInput) (bool, error)

	// GetDescription returns a description of the condition
	GetDescription() string
}

// ConditionChainPair represents a condition and its associated chain
type ConditionChainPair struct {
	Condition ChainCondition
	Chain     Chain
}

// ChainExecutor manages chain execution with advanced features
type ChainExecutor interface {
	// Execute executes a chain with retry and timeout support
	Execute(ctx context.Context, chain Chain, input ChainInput, options *ExecutionOptions) (ChainOutput, error)

	// ExecuteWithCallback executes a chain with callback support
	ExecuteWithCallback(ctx context.Context, chain Chain, input ChainInput, callback ChainCallback) (ChainOutput, error)

	// GetExecutionHistory returns the execution history
	GetExecutionHistory() []ExecutionRecord

	// GetMetrics returns execution metrics
	GetMetrics() ExecutionMetrics
}

// ExecutionOptions represents options for chain execution
type ExecutionOptions struct {
	Timeout        time.Duration `json:"timeout,omitempty"`
	RetryCount     int           `json:"retry_count,omitempty"`
	RetryDelay     time.Duration `json:"retry_delay,omitempty"`
	MaxRetryDelay  time.Duration `json:"max_retry_delay,omitempty"`
	BackoffFactor  float64       `json:"backoff_factor,omitempty"`
	FailFast       bool          `json:"fail_fast,omitempty"`
	ContinueOnError bool         `json:"continue_on_error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ChainCallback defines callbacks for chain execution events
type ChainCallback interface {
	// OnStart is called when chain execution starts
	OnStart(ctx context.Context, chain Chain, input ChainInput) error

	// OnEnd is called when chain execution ends successfully
	OnEnd(ctx context.Context, chain Chain, input ChainInput, output ChainOutput) error

	// OnError is called when chain execution fails
	OnError(ctx context.Context, chain Chain, input ChainInput, err error) error

	// OnRetry is called when chain execution is retried
	OnRetry(ctx context.Context, chain Chain, input ChainInput, attempt int, err error) error
}

// ExecutionRecord represents a record of chain execution
type ExecutionRecord struct {
	ID          string                 `json:"id"`
	ChainType   string                 `json:"chain_type"`
	Input       ChainInput             `json:"input"`
	Output      ChainOutput            `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	RetryCount  int                    `json:"retry_count"`
	Success     bool                   `json:"success"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionMetrics represents metrics for chain execution
type ExecutionMetrics struct {
	TotalExecutions   int64         `json:"total_executions"`
	SuccessfulExecutions int64      `json:"successful_executions"`
	FailedExecutions  int64         `json:"failed_executions"`
	AverageDuration   time.Duration `json:"average_duration"`
	TotalDuration     time.Duration `json:"total_duration"`
	RetryRate         float64       `json:"retry_rate"`
	ErrorRate         float64       `json:"error_rate"`
	ChainTypeMetrics  map[string]ExecutionMetrics `json:"chain_type_metrics,omitempty"`
}

// ChainBuilder provides a fluent interface for building chains
type ChainBuilder interface {
	// WithLLM sets the LLM for the chain
	WithLLM(llm llm.LLM) ChainBuilder

	// WithPrompt sets the prompt template for the chain
	WithPrompt(prompt prompts.PromptTemplate) ChainBuilder

	// WithInputKeys sets the expected input keys
	WithInputKeys(keys []string) ChainBuilder

	// WithOutputKeys sets the output keys
	WithOutputKeys(keys []string) ChainBuilder

	// WithRetry sets retry configuration
	WithRetry(count int, delay time.Duration) ChainBuilder

	// WithTimeout sets execution timeout
	WithTimeout(timeout time.Duration) ChainBuilder

	// WithCallback adds a callback
	WithCallback(callback ChainCallback) ChainBuilder

	// WithMetadata adds metadata
	WithMetadata(key string, value interface{}) ChainBuilder

	// Build builds the chain
	Build() (Chain, error)

	// BuildLLMChain builds an LLM chain
	BuildLLMChain() (LLMChain, error)

	// BuildSequentialChain builds a sequential chain
	BuildSequentialChain() (SequentialChain, error)

	// BuildParallelChain builds a parallel chain
	BuildParallelChain() (ParallelChain, error)

	// BuildConditionalChain builds a conditional chain
	BuildConditionalChain() (ConditionalChain, error)
}

// ChainRegistry manages registered chains
type ChainRegistry interface {
	// RegisterChain registers a chain with a name
	RegisterChain(name string, chain Chain) error

	// GetChain retrieves a chain by name
	GetChain(name string) (Chain, error)

	// ListChains returns all registered chain names
	ListChains() []string

	// UnregisterChain removes a chain from the registry
	UnregisterChain(name string) error

	// Clone creates a copy of the registry
	Clone() ChainRegistry
}

// ChainComposer composes complex chains from simpler ones
type ChainComposer interface {
	// ComposeSequential creates a sequential chain from multiple chains
	ComposeSequential(chains []Chain) (SequentialChain, error)

	// ComposeParallel creates a parallel chain from multiple chains
	ComposeParallel(chains map[string]Chain) (ParallelChain, error)

	// ComposeConditional creates a conditional chain with conditions
	ComposeConditional(conditions []ConditionChainPair, defaultChain Chain) (ConditionalChain, error)

	// ComposeFromConfig creates a chain from configuration
	ComposeFromConfig(config ChainConfig) (Chain, error)
}

// ChainConfig represents configuration for chain composition
type ChainConfig struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	LLM         string                 `json:"llm,omitempty"`
	Prompt      string                 `json:"prompt,omitempty"`
	InputKeys   []string               `json:"input_keys,omitempty"`
	OutputKeys  []string               `json:"output_keys,omitempty"`
	Chains      []ChainConfig          `json:"chains,omitempty"`
	Conditions  []ConditionConfig      `json:"conditions,omitempty"`
	Options     *ExecutionOptions      `json:"options,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConditionConfig represents configuration for a condition
type ConditionConfig struct {
	Type        string                 `json:"type"`
	Expression  string                 `json:"expression"`
	Chain       ChainConfig            `json:"chain"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChainOptimizer optimizes chain execution
type ChainOptimizer interface {
	// OptimizeChain optimizes a chain for better performance
	OptimizeChain(ctx context.Context, chain Chain) (Chain, error)

	// AnalyzeChain analyzes a chain and provides optimization suggestions
	AnalyzeChain(ctx context.Context, chain Chain) (*ChainAnalysis, error)

	// BenchmarkChain benchmarks a chain's performance
	BenchmarkChain(ctx context.Context, chain Chain, inputs []ChainInput) (*ChainBenchmark, error)
}

// ChainAnalysis represents the analysis of a chain
type ChainAnalysis struct {
	ChainType       string                 `json:"chain_type"`
	Complexity      float64                `json:"complexity"`
	EstimatedCost   float64                `json:"estimated_cost"`
	Bottlenecks     []string               `json:"bottlenecks"`
	Suggestions     []string               `json:"suggestions"`
	OptimizedChain  Chain                  `json:"optimized_chain,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ChainBenchmark represents benchmark results for a chain
type ChainBenchmark struct {
	ChainType       string                 `json:"chain_type"`
	TotalRuns       int                    `json:"total_runs"`
	SuccessfulRuns  int                    `json:"successful_runs"`
	FailedRuns      int                    `json:"failed_runs"`
	AverageDuration time.Duration          `json:"average_duration"`
	MinDuration     time.Duration          `json:"min_duration"`
	MaxDuration     time.Duration          `json:"max_duration"`
	Percentiles     map[string]time.Duration `json:"percentiles"`
	ThroughputRPS   float64                `json:"throughput_rps"`
	ErrorRate       float64                `json:"error_rate"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}
