package prompts

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
)

// PromptValue represents a formatted prompt ready for LLM consumption
type PromptValue interface {
	// ToString returns the prompt as a string
	ToString() string

	// ToMessages returns the prompt as a list of messages
	ToMessages() []llm.Message

	// GetVariables returns the variables used in this prompt
	GetVariables() map[string]interface{}
}

// PromptTemplate defines the interface for prompt templates
type PromptTemplate interface {
	// Format formats the template with the given variables
	Format(variables map[string]interface{}) (PromptValue, error)

	// FormatPrompt formats the template and returns a PromptValue
	FormatPrompt(variables map[string]interface{}) (PromptValue, error)

	// GetInputVariables returns the list of input variables required by this template
	GetInputVariables() []string

	// Validate validates the template syntax and variables
	Validate() error

	// Clone creates a copy of the template
	Clone() PromptTemplate
}

// ChatPromptTemplate defines the interface for chat-based prompt templates
type ChatPromptTemplate interface {
	PromptTemplate

	// AddMessage adds a message template to the chat prompt
	AddMessage(role string, template string) error

	// AddSystemMessage adds a system message template
	AddSystemMessage(template string) error

	// AddUserMessage adds a user message template
	AddUserMessage(template string) error

	// AddAssistantMessage adds an assistant message template
	AddAssistantMessage(template string) error

	// GetMessages returns the message templates
	GetMessages() []MessageTemplate
}

// MessageTemplate represents a single message template
type MessageTemplate struct {
	Role     string                 `json:"role"`
	Template string                 `json:"template"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptTemplateConfig represents configuration for prompt templates
type PromptTemplateConfig struct {
	Template         string                 `json:"template"`
	InputVariables   []string               `json:"input_variables"`
	PartialVariables map[string]interface{} `json:"partial_variables,omitempty"`
	TemplateFormat   string                 `json:"template_format,omitempty"` // "f-string", "jinja2", "mustache"
	ValidateTemplate bool                   `json:"validate_template,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ChatPromptTemplateConfig represents configuration for chat prompt templates
type ChatPromptTemplateConfig struct {
	Messages         []MessageTemplate      `json:"messages"`
	InputVariables   []string               `json:"input_variables"`
	PartialVariables map[string]interface{} `json:"partial_variables,omitempty"`
	TemplateFormat   string                 `json:"template_format,omitempty"`
	ValidateTemplate bool                   `json:"validate_template,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PromptTemplateManager manages prompt templates
type PromptTemplateManager interface {
	// RegisterTemplate registers a prompt template with a name
	RegisterTemplate(name string, template PromptTemplate) error

	// GetTemplate retrieves a prompt template by name
	GetTemplate(name string) (PromptTemplate, error)

	// ListTemplates returns all registered template names
	ListTemplates() []string

	// DeleteTemplate removes a template by name
	DeleteTemplate(name string) error

	// LoadFromFile loads templates from a file
	LoadFromFile(filepath string) error

	// SaveToFile saves templates to a file
	SaveToFile(filepath string) error

	// Clone creates a copy of the manager
	Clone() PromptTemplateManager
}

// PromptOptimizer optimizes prompts for better performance
type PromptOptimizer interface {
	// OptimizePrompt optimizes a prompt for the given LLM
	OptimizePrompt(ctx context.Context, prompt PromptValue, llm llm.LLM) (PromptValue, error)

	// AnalyzePrompt analyzes a prompt and provides optimization suggestions
	AnalyzePrompt(ctx context.Context, prompt PromptValue) (*PromptAnalysis, error)

	// ComparePrompts compares multiple prompts and ranks them
	ComparePrompts(ctx context.Context, prompts []PromptValue, llm llm.LLM) (*PromptComparison, error)
}

// PromptAnalysis represents the analysis of a prompt
type PromptAnalysis struct {
	TokenCount       int                    `json:"token_count"`
	Complexity       float64                `json:"complexity"`
	Clarity          float64                `json:"clarity"`
	Suggestions      []string               `json:"suggestions"`
	EstimatedCost    float64                `json:"estimated_cost"`
	OptimizedVersion string                 `json:"optimized_version,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PromptComparison represents the comparison of multiple prompts
type PromptComparison struct {
	Prompts []PromptValue `json:"prompts"`
	Scores  []float64     `json:"scores"`
	Best    int           `json:"best"`
	Worst   int           `json:"worst"`
	Metrics map[string]interface{} `json:"metrics"`
}

// PromptCache caches formatted prompts for performance
type PromptCache interface {
	// Get retrieves a cached prompt
	Get(key string) (PromptValue, bool)

	// Set stores a prompt in the cache
	Set(key string, prompt PromptValue, ttl time.Duration) error

	// Delete removes a prompt from the cache
	Delete(key string) error

	// Clear clears all cached prompts
	Clear() error

	// Stats returns cache statistics
	Stats() map[string]interface{}
}

// PromptValidator validates prompt templates and values
type PromptValidator interface {
	// ValidateTemplate validates a prompt template
	ValidateTemplate(template PromptTemplate) error

	// ValidatePrompt validates a formatted prompt
	ValidatePrompt(prompt PromptValue) error

	// ValidateVariables validates that all required variables are provided
	ValidateVariables(template PromptTemplate, variables map[string]interface{}) error

	// GetValidationRules returns the current validation rules
	GetValidationRules() map[string]interface{}

	// SetValidationRules sets custom validation rules
	SetValidationRules(rules map[string]interface{}) error
}

// PromptMetrics collects metrics about prompt usage
type PromptMetrics interface {
	// RecordPromptUsage records usage of a prompt template
	RecordPromptUsage(templateName string, variables map[string]interface{}, duration time.Duration)

	// RecordPromptPerformance records performance metrics for a prompt
	RecordPromptPerformance(templateName string, tokenCount int, cost float64, success bool)

	// GetUsageStats returns usage statistics for prompt templates
	GetUsageStats(templateName string) map[string]interface{}

	// GetPerformanceStats returns performance statistics
	GetPerformanceStats(templateName string) map[string]interface{}

	// GetTopTemplates returns the most used templates
	GetTopTemplates(limit int) []string
}

// PromptVersioning manages versions of prompt templates
type PromptVersioning interface {
	// SaveVersion saves a new version of a prompt template
	SaveVersion(name string, template PromptTemplate, version string) error

	// GetVersion retrieves a specific version of a prompt template
	GetVersion(name string, version string) (PromptTemplate, error)

	// GetLatestVersion retrieves the latest version of a prompt template
	GetLatestVersion(name string) (PromptTemplate, error)

	// ListVersions returns all versions of a prompt template
	ListVersions(name string) ([]string, error)

	// CompareVersions compares two versions of a prompt template
	CompareVersions(name string, version1, version2 string) (*VersionComparison, error)

	// RollbackVersion rolls back to a previous version
	RollbackVersion(name string, version string) error
}

// VersionComparison represents the comparison between two template versions
type VersionComparison struct {
	Name        string                 `json:"name"`
	Version1    string                 `json:"version1"`
	Version2    string                 `json:"version2"`
	Differences []string               `json:"differences"`
	Metrics     map[string]interface{} `json:"metrics"`
}

// PromptBuilder provides a fluent interface for building prompts
type PromptBuilder interface {
	// WithTemplate sets the template string
	WithTemplate(template string) PromptBuilder

	// WithVariable adds a variable
	WithVariable(name string, value interface{}) PromptBuilder

	// WithVariables adds multiple variables
	WithVariables(variables map[string]interface{}) PromptBuilder

	// WithSystemMessage adds a system message (for chat templates)
	WithSystemMessage(message string) PromptBuilder

	// WithUserMessage adds a user message (for chat templates)
	WithUserMessage(message string) PromptBuilder

	// WithAssistantMessage adds an assistant message (for chat templates)
	WithAssistantMessage(message string) PromptBuilder

	// WithMetadata adds metadata
	WithMetadata(key string, value interface{}) PromptBuilder

	// Build builds the prompt template
	Build() (PromptTemplate, error)

	// BuildAndFormat builds and formats the prompt template
	BuildAndFormat() (PromptValue, error)
}
