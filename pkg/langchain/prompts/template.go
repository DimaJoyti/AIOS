package prompts

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
)

// DefaultPromptValue implements PromptValue interface
type DefaultPromptValue struct {
	content   string
	messages  []llm.Message
	variables map[string]interface{}
}

// ToString returns the prompt as a string
func (p *DefaultPromptValue) ToString() string {
	return p.content
}

// ToMessages returns the prompt as a list of messages
func (p *DefaultPromptValue) ToMessages() []llm.Message {
	if len(p.messages) > 0 {
		return p.messages
	}
	
	// Convert string content to a single user message
	return []llm.Message{
		{
			Role:      "user",
			Content:   p.content,
			Timestamp: time.Now(),
		},
	}
}

// GetVariables returns the variables used in this prompt
func (p *DefaultPromptValue) GetVariables() map[string]interface{} {
	return p.variables
}

// DefaultPromptTemplate implements PromptTemplate interface
type DefaultPromptTemplate struct {
	template         string
	inputVariables   []string
	partialVariables map[string]interface{}
	templateFormat   string
	metadata         map[string]interface{}
}

// NewPromptTemplate creates a new prompt template
func NewPromptTemplate(config *PromptTemplateConfig) (PromptTemplate, error) {
	if config.Template == "" {
		return nil, fmt.Errorf("template cannot be empty")
	}

	if config.TemplateFormat == "" {
		config.TemplateFormat = "f-string"
	}

	template := &DefaultPromptTemplate{
		template:         config.Template,
		inputVariables:   config.InputVariables,
		partialVariables: config.PartialVariables,
		templateFormat:   config.TemplateFormat,
		metadata:         config.Metadata,
	}

	// Auto-detect input variables if not provided
	if len(template.inputVariables) == 0 {
		template.inputVariables = template.extractVariables()
	}

	if config.ValidateTemplate {
		if err := template.Validate(); err != nil {
			return nil, fmt.Errorf("template validation failed: %w", err)
		}
	}

	return template, nil
}

// Format formats the template with the given variables
func (t *DefaultPromptTemplate) Format(variables map[string]interface{}) (PromptValue, error) {
	return t.FormatPrompt(variables)
}

// FormatPrompt formats the template and returns a PromptValue
func (t *DefaultPromptTemplate) FormatPrompt(variables map[string]interface{}) (PromptValue, error) {
	// Merge partial variables with provided variables
	allVariables := make(map[string]interface{})
	
	// Add partial variables first
	for k, v := range t.partialVariables {
		allVariables[k] = v
	}
	
	// Add provided variables (can override partial variables)
	for k, v := range variables {
		allVariables[k] = v
	}

	// Check that all required variables are provided
	for _, varName := range t.inputVariables {
		if _, exists := allVariables[varName]; !exists {
			return nil, fmt.Errorf("missing required variable: %s", varName)
		}
	}

	// Format the template based on the template format
	content, err := t.formatTemplate(t.template, allVariables)
	if err != nil {
		return nil, fmt.Errorf("failed to format template: %w", err)
	}

	return &DefaultPromptValue{
		content:   content,
		variables: allVariables,
	}, nil
}

// GetInputVariables returns the list of input variables required by this template
func (t *DefaultPromptTemplate) GetInputVariables() []string {
	return t.inputVariables
}

// Validate validates the template syntax and variables
func (t *DefaultPromptTemplate) Validate() error {
	// Check for basic syntax issues
	if strings.TrimSpace(t.template) == "" {
		return fmt.Errorf("template cannot be empty")
	}

	// Validate variables in template
	extractedVars := t.extractVariables()
	for _, varName := range extractedVars {
		if varName == "" {
			return fmt.Errorf("empty variable name found in template")
		}
	}

	// Try formatting with dummy variables to check syntax
	dummyVars := make(map[string]interface{})
	for _, varName := range t.inputVariables {
		dummyVars[varName] = "test"
	}

	_, err := t.formatTemplate(t.template, dummyVars)
	if err != nil {
		return fmt.Errorf("template syntax error: %w", err)
	}

	return nil
}

// Clone creates a copy of the template
func (t *DefaultPromptTemplate) Clone() PromptTemplate {
	inputVars := make([]string, len(t.inputVariables))
	copy(inputVars, t.inputVariables)

	partialVars := make(map[string]interface{})
	for k, v := range t.partialVariables {
		partialVars[k] = v
	}

	metadata := make(map[string]interface{})
	for k, v := range t.metadata {
		metadata[k] = v
	}

	return &DefaultPromptTemplate{
		template:         t.template,
		inputVariables:   inputVars,
		partialVariables: partialVars,
		templateFormat:   t.templateFormat,
		metadata:         metadata,
	}
}

// Helper methods

func (t *DefaultPromptTemplate) extractVariables() []string {
	var variables []string
	
	switch t.templateFormat {
	case "f-string":
		// Extract variables in {variable} format
		re := regexp.MustCompile(`\{([^}]+)\}`)
		matches := re.FindAllStringSubmatch(t.template, -1)
		
		varSet := make(map[string]bool)
		for _, match := range matches {
			if len(match) > 1 {
				varName := strings.TrimSpace(match[1])
				if !varSet[varName] {
					variables = append(variables, varName)
					varSet[varName] = true
				}
			}
		}
	case "mustache":
		// Extract variables in {{variable}} format
		re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
		matches := re.FindAllStringSubmatch(t.template, -1)
		
		varSet := make(map[string]bool)
		for _, match := range matches {
			if len(match) > 1 {
				varName := strings.TrimSpace(match[1])
				if !varSet[varName] {
					variables = append(variables, varName)
					varSet[varName] = true
				}
			}
		}
	default:
		// Default to f-string format
		return t.extractVariables()
	}
	
	return variables
}

func (t *DefaultPromptTemplate) formatTemplate(template string, variables map[string]interface{}) (string, error) {
	result := template
	
	switch t.templateFormat {
	case "f-string":
		// Replace {variable} with values
		for varName, value := range variables {
			placeholder := fmt.Sprintf("{%s}", varName)
			valueStr := fmt.Sprintf("%v", value)
			result = strings.ReplaceAll(result, placeholder, valueStr)
		}
	case "mustache":
		// Replace {{variable}} with values
		for varName, value := range variables {
			placeholder := fmt.Sprintf("{{%s}}", varName)
			valueStr := fmt.Sprintf("%v", value)
			result = strings.ReplaceAll(result, placeholder, valueStr)
		}
	default:
		return "", fmt.Errorf("unsupported template format: %s", t.templateFormat)
	}
	
	return result, nil
}

// DefaultChatPromptTemplate implements ChatPromptTemplate interface
type DefaultChatPromptTemplate struct {
	*DefaultPromptTemplate
	messages []MessageTemplate
}

// NewChatPromptTemplate creates a new chat prompt template
func NewChatPromptTemplate(config *ChatPromptTemplateConfig) (ChatPromptTemplate, error) {
	if len(config.Messages) == 0 {
		return nil, fmt.Errorf("chat template must have at least one message")
	}

	// Create base template config
	baseConfig := &PromptTemplateConfig{
		Template:         "", // Will be built from messages
		InputVariables:   config.InputVariables,
		PartialVariables: config.PartialVariables,
		TemplateFormat:   config.TemplateFormat,
		ValidateTemplate: config.ValidateTemplate,
		Metadata:         config.Metadata,
	}

	baseTemplate, err := NewPromptTemplate(baseConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create base template: %w", err)
	}

	chatTemplate := &DefaultChatPromptTemplate{
		DefaultPromptTemplate: baseTemplate.(*DefaultPromptTemplate),
		messages:              config.Messages,
	}

	// Auto-detect input variables from all messages if not provided
	if len(chatTemplate.inputVariables) == 0 {
		chatTemplate.inputVariables = chatTemplate.extractVariablesFromMessages()
	}

	if config.ValidateTemplate {
		if err := chatTemplate.Validate(); err != nil {
			return nil, fmt.Errorf("chat template validation failed: %w", err)
		}
	}

	return chatTemplate, nil
}

// FormatPrompt formats the chat template and returns a PromptValue
func (c *DefaultChatPromptTemplate) FormatPrompt(variables map[string]interface{}) (PromptValue, error) {
	// Merge partial variables with provided variables
	allVariables := make(map[string]interface{})
	
	for k, v := range c.partialVariables {
		allVariables[k] = v
	}
	
	for k, v := range variables {
		allVariables[k] = v
	}

	// Check that all required variables are provided
	for _, varName := range c.inputVariables {
		if _, exists := allVariables[varName]; !exists {
			return nil, fmt.Errorf("missing required variable: %s", varName)
		}
	}

	// Format each message
	messages := make([]llm.Message, len(c.messages))
	for i, msgTemplate := range c.messages {
		content, err := c.formatTemplate(msgTemplate.Template, allVariables)
		if err != nil {
			return nil, fmt.Errorf("failed to format message %d: %w", i, err)
		}

		messages[i] = llm.Message{
			Role:      msgTemplate.Role,
			Content:   content,
			Metadata:  msgTemplate.Metadata,
			Timestamp: time.Now(),
		}
	}

	return &DefaultPromptValue{
		messages:  messages,
		variables: allVariables,
	}, nil
}

// AddMessage adds a message template to the chat prompt
func (c *DefaultChatPromptTemplate) AddMessage(role string, template string) error {
	c.messages = append(c.messages, MessageTemplate{
		Role:     role,
		Template: template,
	})
	
	// Update input variables
	c.inputVariables = c.extractVariablesFromMessages()
	
	return nil
}

// AddSystemMessage adds a system message template
func (c *DefaultChatPromptTemplate) AddSystemMessage(template string) error {
	return c.AddMessage("system", template)
}

// AddUserMessage adds a user message template
func (c *DefaultChatPromptTemplate) AddUserMessage(template string) error {
	return c.AddMessage("user", template)
}

// AddAssistantMessage adds an assistant message template
func (c *DefaultChatPromptTemplate) AddAssistantMessage(template string) error {
	return c.AddMessage("assistant", template)
}

// GetMessages returns the message templates
func (c *DefaultChatPromptTemplate) GetMessages() []MessageTemplate {
	return c.messages
}

func (c *DefaultChatPromptTemplate) extractVariablesFromMessages() []string {
	varSet := make(map[string]bool)
	var variables []string

	for _, msg := range c.messages {
		// Temporarily set template to extract variables
		oldTemplate := c.template
		c.template = msg.Template
		
		msgVars := c.extractVariables()
		for _, varName := range msgVars {
			if !varSet[varName] {
				variables = append(variables, varName)
				varSet[varName] = true
			}
		}
		
		c.template = oldTemplate
	}

	return variables
}
