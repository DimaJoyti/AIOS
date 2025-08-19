package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PromptManager manages AI prompts and templates
type PromptManager struct {
	templates map[string]*PromptTemplate
	chains    map[string]*PromptChain
	mutex     sync.RWMutex
	logger    *logrus.Logger
	tracer    trace.Tracer
}

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Template    string                 `json:"template"`
	Variables   []PromptVariable       `json:"variables"`
	Examples    []PromptExample        `json:"examples"`
	Config      PromptConfig           `json:"config"`
	Metadata    map[string]interface{} `json:"metadata"`
	Version     string                 `json:"version"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	Tags        []string               `json:"tags"`
	IsActive    bool                   `json:"is_active"`
}

// PromptVariable represents a variable in a prompt template
type PromptVariable struct {
	Name         string              `json:"name"`
	Type         string              `json:"type"` // string, number, boolean, array, object
	Description  string              `json:"description"`
	Required     bool                `json:"required"`
	DefaultValue interface{}         `json:"default_value,omitempty"`
	Validation   *VariableValidation `json:"validation,omitempty"`
}

// VariableValidation defines validation rules for prompt variables
type VariableValidation struct {
	MinLength *int     `json:"min_length,omitempty"`
	MaxLength *int     `json:"max_length,omitempty"`
	Pattern   *string  `json:"pattern,omitempty"`
	Enum      []string `json:"enum,omitempty"`
	Min       *float64 `json:"min,omitempty"`
	Max       *float64 `json:"max,omitempty"`
}

// PromptExample represents an example usage of a prompt template
type PromptExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Variables   map[string]interface{} `json:"variables"`
	Expected    string                 `json:"expected"`
}

// PromptConfig holds configuration for prompt execution
type PromptConfig struct {
	ModelID          string                 `json:"model_id"`
	Temperature      float64                `json:"temperature"`
	MaxTokens        int                    `json:"max_tokens"`
	TopP             float64                `json:"top_p"`
	FrequencyPenalty float64                `json:"frequency_penalty"`
	PresencePenalty  float64                `json:"presence_penalty"`
	StopSequences    []string               `json:"stop_sequences"`
	SystemPrompt     string                 `json:"system_prompt"`
	Parameters       map[string]interface{} `json:"parameters"`
}

// PromptChain represents a sequence of prompts for complex workflows
type PromptChain struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []PromptChainStep      `json:"steps"`
	Config      PromptChainConfig      `json:"config"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// PromptChainStep represents a step in a prompt chain
type PromptChainStep struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	TemplateID    string                 `json:"template_id"`
	Variables     map[string]interface{} `json:"variables"`
	Condition     *StepCondition         `json:"condition,omitempty"`
	OutputMapping map[string]string      `json:"output_mapping"`
	RetryConfig   *RetryConfig           `json:"retry_config,omitempty"`
}

// StepCondition defines when a step should be executed
type StepCondition struct {
	Type       string      `json:"type"` // always, if, unless
	Variable   string      `json:"variable,omitempty"`
	Operator   string      `json:"operator,omitempty"` // eq, ne, gt, lt, contains
	Value      interface{} `json:"value,omitempty"`
	Expression string      `json:"expression,omitempty"`
}

// RetryConfig defines retry behavior for a step
type RetryConfig struct {
	MaxRetries int           `json:"max_retries"`
	Delay      time.Duration `json:"delay"`
	Backoff    string        `json:"backoff"` // linear, exponential
}

// PromptChainConfig holds configuration for prompt chain execution
type PromptChainConfig struct {
	Parallel       bool          `json:"parallel"`
	StopOnError    bool          `json:"stop_on_error"`
	Timeout        time.Duration `json:"timeout"`
	MaxConcurrency int           `json:"max_concurrency"`
	DefaultRetries int           `json:"default_retries"`
}

// PromptExecution represents the execution of a prompt
type PromptExecution struct {
	ID         string                 `json:"id"`
	TemplateID string                 `json:"template_id"`
	ChainID    string                 `json:"chain_id,omitempty"`
	Variables  map[string]interface{} `json:"variables"`
	Result     *PromptResult          `json:"result"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
	Duration   time.Duration          `json:"duration"`
	Status     string                 `json:"status"` // pending, running, completed, failed
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// PromptResult represents the result of prompt execution
type PromptResult struct {
	Text         string                 `json:"text"`
	FinishReason string                 `json:"finish_reason"`
	Usage        TokenUsage             `json:"usage"`
	Cost         float64                `json:"cost"`
	Latency      time.Duration          `json:"latency"`
	ModelID      string                 `json:"model_id"`
	Provider     string                 `json:"provider"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// NewPromptManager creates a new prompt manager
func NewPromptManager(logger *logrus.Logger) *PromptManager {
	return &PromptManager{
		templates: make(map[string]*PromptTemplate),
		chains:    make(map[string]*PromptChain),
		logger:    logger,
		tracer:    otel.Tracer("ai.services.prompt_manager"),
	}
}

// CreateTemplate creates a new prompt template
func (pm *PromptManager) CreateTemplate(template *PromptTemplate) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.IsActive = true

	// Validate template
	if err := pm.validateTemplate(template); err != nil {
		return fmt.Errorf("template validation failed: %w", err)
	}

	pm.templates[template.ID] = template

	pm.logger.WithFields(logrus.Fields{
		"template_id": template.ID,
		"name":        template.Name,
		"category":    template.Category,
	}).Info("Prompt template created")

	return nil
}

// GetTemplate retrieves a prompt template by ID
func (pm *PromptManager) GetTemplate(templateID string) (*PromptTemplate, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	template, exists := pm.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// ListTemplates returns all prompt templates
func (pm *PromptManager) ListTemplates() []*PromptTemplate {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	templates := make([]*PromptTemplate, 0, len(pm.templates))
	for _, template := range pm.templates {
		if template.IsActive {
			templates = append(templates, template)
		}
	}

	return templates
}

// GetTemplatesByCategory returns templates filtered by category
func (pm *PromptManager) GetTemplatesByCategory(category string) []*PromptTemplate {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	var templates []*PromptTemplate
	for _, template := range pm.templates {
		if template.IsActive && template.Category == category {
			templates = append(templates, template)
		}
	}

	return templates
}

// RenderTemplate renders a prompt template with variables
func (pm *PromptManager) RenderTemplate(ctx context.Context, templateID string, variables map[string]interface{}) (string, error) {
	ctx, span := pm.tracer.Start(ctx, "prompt_manager.render_template")
	defer span.End()

	promptTemplate, err := pm.GetTemplate(templateID)
	if err != nil {
		return "", err
	}

	// Validate variables
	if err := pm.validateVariables(promptTemplate, variables); err != nil {
		return "", fmt.Errorf("variable validation failed: %w", err)
	}

	// Apply default values for missing variables
	finalVariables := pm.applyDefaults(promptTemplate, variables)

	// Render template
	tmpl, err := template.New("prompt").Parse(promptTemplate.Template)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, finalVariables); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// ExecuteTemplate executes a prompt template and returns the AI response
func (pm *PromptManager) ExecuteTemplate(ctx context.Context, modelManager *ModelManager, templateID string, variables map[string]interface{}) (*PromptExecution, error) {
	ctx, span := pm.tracer.Start(ctx, "prompt_manager.execute_template")
	defer span.End()

	execution := &PromptExecution{
		ID:         generateExecutionID(),
		TemplateID: templateID,
		Variables:  variables,
		StartTime:  time.Now(),
		Status:     "running",
		Metadata:   make(map[string]interface{}),
	}

	// Render template
	prompt, err := pm.RenderTemplate(ctx, templateID, variables)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		execution.EndTime = time.Now()
		execution.Duration = time.Since(execution.StartTime)
		return execution, err
	}

	// Get template configuration
	template, _ := pm.GetTemplate(templateID)

	// Create text generation request
	request := &TextGenerationRequest{
		ModelID: template.Config.ModelID,
		Prompt:  prompt,
		Config: ModelConfig{
			MaxTokens:        template.Config.MaxTokens,
			Temperature:      template.Config.Temperature,
			TopP:             template.Config.TopP,
			FrequencyPenalty: template.Config.FrequencyPenalty,
			PresencePenalty:  template.Config.PresencePenalty,
			StopSequences:    template.Config.StopSequences,
			SystemPrompt:     template.Config.SystemPrompt,
		},
		Metadata: execution.Metadata,
	}

	// Execute with model manager
	response, err := modelManager.GenerateText(ctx, request)
	if err != nil {
		execution.Status = "failed"
		execution.Error = err.Error()
		execution.EndTime = time.Now()
		execution.Duration = time.Since(execution.StartTime)
		return execution, err
	}

	// Update execution with result
	execution.Result = &PromptResult{
		Text:         response.Text,
		FinishReason: response.FinishReason,
		Usage:        response.Usage,
		Cost:         response.Cost,
		Latency:      response.Latency,
		ModelID:      response.ModelID,
		Provider:     response.Provider,
		Metadata:     response.Metadata,
	}

	execution.Status = "completed"
	execution.EndTime = time.Now()
	execution.Duration = time.Since(execution.StartTime)

	pm.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"template_id":  templateID,
		"duration":     execution.Duration,
		"cost":         execution.Result.Cost,
	}).Info("Prompt template executed")

	return execution, nil
}

// CreateChain creates a new prompt chain
func (pm *PromptManager) CreateChain(chain *PromptChain) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	chain.CreatedAt = time.Now()
	chain.UpdatedAt = time.Now()
	chain.IsActive = true

	// Validate chain
	if err := pm.validateChain(chain); err != nil {
		return fmt.Errorf("chain validation failed: %w", err)
	}

	pm.chains[chain.ID] = chain

	pm.logger.WithFields(logrus.Fields{
		"chain_id": chain.ID,
		"name":     chain.Name,
		"steps":    len(chain.Steps),
	}).Info("Prompt chain created")

	return nil
}

// ExecuteChain executes a prompt chain
func (pm *PromptManager) ExecuteChain(ctx context.Context, modelManager *ModelManager, chainID string, variables map[string]interface{}) ([]*PromptExecution, error) {
	ctx, span := pm.tracer.Start(ctx, "prompt_manager.execute_chain")
	defer span.End()

	chain, err := pm.getChain(chainID)
	if err != nil {
		return nil, err
	}

	var executions []*PromptExecution
	chainVariables := make(map[string]interface{})

	// Copy initial variables
	for k, v := range variables {
		chainVariables[k] = v
	}

	for _, step := range chain.Steps {
		// Check step condition
		if !pm.evaluateStepCondition(step.Condition, chainVariables) {
			continue
		}

		// Prepare step variables
		stepVariables := make(map[string]interface{})
		for k, v := range chainVariables {
			stepVariables[k] = v
		}
		for k, v := range step.Variables {
			stepVariables[k] = v
		}

		// Execute step
		execution, err := pm.ExecuteTemplate(ctx, modelManager, step.TemplateID, stepVariables)
		if err != nil {
			if chain.Config.StopOnError {
				return executions, err
			}
			// Continue with next step
			continue
		}

		executions = append(executions, execution)

		// Apply output mapping
		for outputVar, sourceVar := range step.OutputMapping {
			if execution.Result != nil {
				switch sourceVar {
				case "text":
					chainVariables[outputVar] = execution.Result.Text
				case "cost":
					chainVariables[outputVar] = execution.Result.Cost
				case "usage.total_tokens":
					chainVariables[outputVar] = execution.Result.Usage.TotalTokens
				}
			}
		}
	}

	return executions, nil
}

// validateTemplate validates a prompt template
func (pm *PromptManager) validateTemplate(promptTemplate *PromptTemplate) error {
	if promptTemplate.ID == "" {
		return fmt.Errorf("template ID is required")
	}
	if promptTemplate.Name == "" {
		return fmt.Errorf("template name is required")
	}
	if promptTemplate.Template == "" {
		return fmt.Errorf("template content is required")
	}

	// Validate template syntax
	_, err := template.New("validation").Parse(promptTemplate.Template)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	return nil
}

// validateChain validates a prompt chain
func (pm *PromptManager) validateChain(chain *PromptChain) error {
	if chain.ID == "" {
		return fmt.Errorf("chain ID is required")
	}
	if chain.Name == "" {
		return fmt.Errorf("chain name is required")
	}
	if len(chain.Steps) == 0 {
		return fmt.Errorf("chain must have at least one step")
	}

	// Validate that all referenced templates exist
	for _, step := range chain.Steps {
		if _, err := pm.GetTemplate(step.TemplateID); err != nil {
			return fmt.Errorf("step %s references non-existent template %s", step.ID, step.TemplateID)
		}
	}

	return nil
}

// validateVariables validates variables against template requirements
func (pm *PromptManager) validateVariables(template *PromptTemplate, variables map[string]interface{}) error {
	for _, variable := range template.Variables {
		value, exists := variables[variable.Name]

		if variable.Required && !exists {
			return fmt.Errorf("required variable '%s' is missing", variable.Name)
		}

		if exists && variable.Validation != nil {
			if err := pm.validateVariableValue(variable, value); err != nil {
				return fmt.Errorf("validation failed for variable '%s': %w", variable.Name, err)
			}
		}
	}

	return nil
}

// validateVariableValue validates a single variable value
func (pm *PromptManager) validateVariableValue(variable PromptVariable, value interface{}) error {
	validation := variable.Validation

	switch variable.Type {
	case "string":
		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string, got %T", value)
		}

		if validation.MinLength != nil && len(str) < *validation.MinLength {
			return fmt.Errorf("string too short (min: %d)", *validation.MinLength)
		}

		if validation.MaxLength != nil && len(str) > *validation.MaxLength {
			return fmt.Errorf("string too long (max: %d)", *validation.MaxLength)
		}

		if validation.Enum != nil {
			found := false
			for _, enum := range validation.Enum {
				if str == enum {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value not in allowed enum")
			}
		}

	case "number":
		var num float64
		switch v := value.(type) {
		case float64:
			num = v
		case int:
			num = float64(v)
		default:
			return fmt.Errorf("expected number, got %T", value)
		}

		if validation.Min != nil && num < *validation.Min {
			return fmt.Errorf("number too small (min: %f)", *validation.Min)
		}

		if validation.Max != nil && num > *validation.Max {
			return fmt.Errorf("number too large (max: %f)", *validation.Max)
		}
	}

	return nil
}

// applyDefaults applies default values for missing variables
func (pm *PromptManager) applyDefaults(template *PromptTemplate, variables map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy provided variables
	for k, v := range variables {
		result[k] = v
	}

	// Apply defaults for missing variables
	for _, variable := range template.Variables {
		if _, exists := result[variable.Name]; !exists && variable.DefaultValue != nil {
			result[variable.Name] = variable.DefaultValue
		}
	}

	return result
}

// evaluateStepCondition evaluates whether a step should be executed
func (pm *PromptManager) evaluateStepCondition(condition *StepCondition, variables map[string]interface{}) bool {
	if condition == nil || condition.Type == "always" {
		return true
	}

	// Simple condition evaluation (can be extended)
	if condition.Variable != "" {
		value, exists := variables[condition.Variable]
		if !exists {
			return condition.Type == "unless"
		}

		switch condition.Operator {
		case "eq":
			return (value == condition.Value) == (condition.Type == "if")
		case "ne":
			return (value != condition.Value) == (condition.Type == "if")
		}
	}

	return true
}

// getChain retrieves a prompt chain by ID
func (pm *PromptManager) getChain(chainID string) (*PromptChain, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	chain, exists := pm.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}

	return chain, nil
}

// generateExecutionID generates a unique execution ID
func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}
