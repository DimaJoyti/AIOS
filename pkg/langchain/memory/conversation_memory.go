package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultConversationMemory implements ConversationMemory interface
type DefaultConversationMemory struct {
	messages    []llm.Message
	maxTokens   int
	maxMessages int
	memoryKeys  []string
	store       MemoryStore
	logger      *logrus.Logger
	tracer      trace.Tracer
	mu          sync.RWMutex
}

// ConversationMemoryConfig represents configuration for conversation memory
type ConversationMemoryConfig struct {
	MaxTokens   int         `json:"max_tokens,omitempty"`
	MaxMessages int         `json:"max_messages,omitempty"`
	MemoryKeys  []string    `json:"memory_keys,omitempty"`
	Store       MemoryStore `json:"-"`
	Persistent  bool        `json:"persistent,omitempty"`
}

// NewConversationMemory creates a new conversation memory
func NewConversationMemory(config *ConversationMemoryConfig, logger *logrus.Logger) (ConversationMemory, error) {
	if config.MaxTokens <= 0 {
		config.MaxTokens = 2000 // Default max tokens
	}

	if config.MaxMessages <= 0 {
		config.MaxMessages = 100 // Default max messages
	}

	memoryKeys := config.MemoryKeys
	if len(memoryKeys) == 0 {
		memoryKeys = []string{"history", "chat_history"}
	}

	memory := &DefaultConversationMemory{
		messages:    make([]llm.Message, 0),
		maxTokens:   config.MaxTokens,
		maxMessages: config.MaxMessages,
		memoryKeys:  memoryKeys,
		store:       config.Store,
		logger:      logger,
		tracer:      otel.Tracer("langchain.memory.conversation"),
	}

	// Load existing messages if store is provided
	if config.Persistent && config.Store != nil {
		if err := memory.loadFromStore(context.Background()); err != nil {
			logger.WithError(err).Warn("Failed to load conversation history from store")
		}
	}

	return memory, nil
}

// LoadMemoryVariables loads memory variables for the given input
func (m *DefaultConversationMemory) LoadMemoryVariables(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "conversation_memory.load_variables")
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	variables := make(map[string]interface{})

	// Format messages as string for history
	history := m.formatMessagesAsString()

	for _, key := range m.memoryKeys {
		switch key {
		case "history", "chat_history":
			variables[key] = history
		case "messages":
			variables[key] = m.messages
		case "recent_messages":
			recentCount := 10 // Default recent count
			if len(m.messages) < recentCount {
				recentCount = len(m.messages)
			}
			variables[key] = m.messages[len(m.messages)-recentCount:]
		}
	}

	span.SetAttributes(
		attribute.Int("memory.message_count", len(m.messages)),
		attribute.Int("memory.variables_count", len(variables)),
	)

	return variables, nil
}

// SaveContext saves the context from input and output
func (m *DefaultConversationMemory) SaveContext(ctx context.Context, input map[string]interface{}, output map[string]interface{}) error {
	ctx, span := m.tracer.Start(ctx, "conversation_memory.save_context")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Extract messages from input and output
	if userInput, exists := input["input"]; exists {
		userMessage := llm.Message{
			Role:      "user",
			Content:   fmt.Sprintf("%v", userInput),
			Timestamp: time.Now(),
		}
		m.messages = append(m.messages, userMessage)
	}

	if assistantOutput, exists := output["text"]; exists {
		assistantMessage := llm.Message{
			Role:      "assistant",
			Content:   fmt.Sprintf("%v", assistantOutput),
			Timestamp: time.Now(),
		}
		m.messages = append(m.messages, assistantMessage)
	}

	// Trim messages if necessary
	m.trimMessages()

	// Save to store if available
	if m.store != nil {
		if err := m.saveToStore(ctx); err != nil {
			m.logger.WithError(err).Error("Failed to save conversation to store")
			return err
		}
	}

	span.SetAttributes(
		attribute.Int("memory.total_messages", len(m.messages)),
	)

	return nil
}

// Clear clears all memory
func (m *DefaultConversationMemory) Clear(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "conversation_memory.clear")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = make([]llm.Message, 0)

	// Clear from store if available
	if m.store != nil {
		if err := m.store.Delete(ctx, "conversation_messages"); err != nil {
			return fmt.Errorf("failed to clear conversation from store: %w", err)
		}
	}

	return nil
}

// GetMemoryKeys returns the keys that this memory system provides
func (m *DefaultConversationMemory) GetMemoryKeys() []string {
	return m.memoryKeys
}

// GetMemoryType returns the type of this memory system
func (m *DefaultConversationMemory) GetMemoryType() string {
	return "conversation"
}

// AddMessage adds a message to the conversation
func (m *DefaultConversationMemory) AddMessage(ctx context.Context, message llm.Message) error {
	ctx, span := m.tracer.Start(ctx, "conversation_memory.add_message")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.messages = append(m.messages, message)
	m.trimMessages()

	// Save to store if available
	if m.store != nil {
		if err := m.saveToStore(ctx); err != nil {
			return fmt.Errorf("failed to save message to store: %w", err)
		}
	}

	span.SetAttributes(
		attribute.String("message.role", message.Role),
		attribute.Int("message.content_length", len(message.Content)),
	)

	return nil
}

// GetMessages returns all messages in the conversation
func (m *DefaultConversationMemory) GetMessages(ctx context.Context) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modification
	messages := make([]llm.Message, len(m.messages))
	copy(messages, m.messages)

	return messages, nil
}

// GetRecentMessages returns the most recent N messages
func (m *DefaultConversationMemory) GetRecentMessages(ctx context.Context, count int) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if count <= 0 {
		return []llm.Message{}, nil
	}

	start := len(m.messages) - count
	if start < 0 {
		start = 0
	}

	messages := make([]llm.Message, len(m.messages[start:]))
	copy(messages, m.messages[start:])

	return messages, nil
}

// GetMessagesByTimeRange returns messages within a time range
func (m *DefaultConversationMemory) GetMessagesByTimeRange(ctx context.Context, start, end time.Time) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var filteredMessages []llm.Message
	for _, message := range m.messages {
		if message.Timestamp.After(start) && message.Timestamp.Before(end) {
			filteredMessages = append(filteredMessages, message)
		}
	}

	return filteredMessages, nil
}

// SetMaxTokens sets the maximum number of tokens to keep in memory
func (m *DefaultConversationMemory) SetMaxTokens(maxTokens int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxTokens = maxTokens
	m.trimMessages()
}

// SetMaxMessages sets the maximum number of messages to keep in memory
func (m *DefaultConversationMemory) SetMaxMessages(maxMessages int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.maxMessages = maxMessages
	m.trimMessages()
}

// Helper methods

func (m *DefaultConversationMemory) formatMessagesAsString() string {
	if len(m.messages) == 0 {
		return ""
	}

	var result string
	for _, message := range m.messages {
		switch message.Role {
		case "user":
			result += fmt.Sprintf("Human: %s\n", message.Content)
		case "assistant":
			result += fmt.Sprintf("Assistant: %s\n", message.Content)
		case "system":
			result += fmt.Sprintf("System: %s\n", message.Content)
		default:
			result += fmt.Sprintf("%s: %s\n", message.Role, message.Content)
		}
	}

	return result
}

func (m *DefaultConversationMemory) trimMessages() {
	// Trim by message count
	if len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}

	// Trim by token count (approximate)
	if m.maxTokens > 0 {
		totalTokens := 0
		for i := len(m.messages) - 1; i >= 0; i-- {
			// Rough token estimation: 1 token ≈ 4 characters
			messageTokens := len(m.messages[i].Content) / 4
			if totalTokens+messageTokens > m.maxTokens {
				m.messages = m.messages[i+1:]
				break
			}
			totalTokens += messageTokens
		}
	}
}

func (m *DefaultConversationMemory) saveToStore(ctx context.Context) error {
	if m.store == nil {
		return nil
	}

	return m.store.Store(ctx, "conversation_messages", m.messages)
}

func (m *DefaultConversationMemory) loadFromStore(ctx context.Context) error {
	if m.store == nil {
		return nil
	}

	data, err := m.store.Retrieve(ctx, "conversation_messages")
	if err != nil {
		return err
	}

	if messages, ok := data.([]llm.Message); ok {
		m.messages = messages
		m.trimMessages()
	}

	return nil
}

// BufferWindowMemory keeps only the last N messages
type BufferWindowMemory struct {
	*DefaultConversationMemory
	windowSize int
}

// NewBufferWindowMemory creates a new buffer window memory
func NewBufferWindowMemory(windowSize int, logger *logrus.Logger) (*BufferWindowMemory, error) {
	config := &ConversationMemoryConfig{
		MaxMessages: windowSize,
		MaxTokens:   0, // No token limit for window memory
	}

	baseMemory, err := NewConversationMemory(config, logger)
	if err != nil {
		return nil, err
	}

	return &BufferWindowMemory{
		DefaultConversationMemory: baseMemory.(*DefaultConversationMemory),
		windowSize:                windowSize,
	}, nil
}

// TokenBufferMemory keeps messages within a token limit
type TokenBufferMemory struct {
	*DefaultConversationMemory
	tokenLimit int
}

// NewTokenBufferMemory creates a new token buffer memory
func NewTokenBufferMemory(tokenLimit int, logger *logrus.Logger) (*TokenBufferMemory, error) {
	config := &ConversationMemoryConfig{
		MaxTokens:   tokenLimit,
		MaxMessages: 0, // No message limit for token memory
	}

	baseMemory, err := NewConversationMemory(config, logger)
	if err != nil {
		return nil, err
	}

	return &TokenBufferMemory{
		DefaultConversationMemory: baseMemory.(*DefaultConversationMemory),
		tokenLimit:                tokenLimit,
	}, nil
}

// ConversationSummaryMemory combines conversation memory with summarization
type ConversationSummaryMemory struct {
	*DefaultConversationMemory
	summaryLLM    llm.LLM
	summary       string
	summaryTokens int
}

// NewConversationSummaryMemory creates a new conversation summary memory
func NewConversationSummaryMemory(summaryLLM llm.LLM, config *ConversationMemoryConfig, logger *logrus.Logger) (*ConversationSummaryMemory, error) {
	baseMemory, err := NewConversationMemory(config, logger)
	if err != nil {
		return nil, err
	}

	return &ConversationSummaryMemory{
		DefaultConversationMemory: baseMemory.(*DefaultConversationMemory),
		summaryLLM:                summaryLLM,
		summary:                   "",
		summaryTokens:             0,
	}, nil
}

// LoadMemoryVariables loads memory variables including summary
func (m *ConversationSummaryMemory) LoadMemoryVariables(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	variables, err := m.DefaultConversationMemory.LoadMemoryVariables(ctx, input)
	if err != nil {
		return nil, err
	}

	// Add summary to variables
	if m.summary != "" {
		variables["summary"] = m.summary
	}

	return variables, nil
}

// SaveContext saves context and updates summary if needed
func (m *ConversationSummaryMemory) SaveContext(ctx context.Context, input map[string]interface{}, output map[string]interface{}) error {
	// Save to base memory first
	if err := m.DefaultConversationMemory.SaveContext(ctx, input, output); err != nil {
		return err
	}

	// Update summary if we have too many tokens
	totalTokens := m.estimateTokens()
	if totalTokens > m.maxTokens {
		return m.updateSummary(ctx)
	}

	return nil
}

func (m *ConversationSummaryMemory) estimateTokens() int {
	total := m.summaryTokens
	for _, message := range m.messages {
		total += len(message.Content) / 4 // Rough estimation
	}
	return total
}

func (m *ConversationSummaryMemory) updateSummary(ctx context.Context) error {
	if m.summaryLLM == nil {
		return fmt.Errorf("summary LLM not configured")
	}

	// Create summary prompt
	currentHistory := m.formatMessagesAsString()
	prompt := fmt.Sprintf("Progressively summarize the lines of conversation provided, adding onto the previous summary returning a new summary.\n\nCURRENT SUMMARY:\n%s\n\nNEW LINES OF CONVERSATION:\n%s\n\nNEW SUMMARY:", m.summary, currentHistory)

	req := &llm.CompletionRequest{
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	response, err := m.summaryLLM.Complete(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	m.summary = response.Content
	m.summaryTokens = len(response.Content) / 4 // Rough estimation

	// Clear old messages since they're now summarized
	m.messages = make([]llm.Message, 0)

	return nil
}
