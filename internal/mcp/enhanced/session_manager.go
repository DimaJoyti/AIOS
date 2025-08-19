package enhanced

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SessionManager manages MCP sessions with enhanced capabilities
type SessionManager struct {
	sessions map[string]*MCPSession
	mutex    sync.RWMutex
	logger   *logrus.Logger
	tracer   trace.Tracer
}

// MCPSession represents an enhanced MCP session with context and memory
type MCPSession struct {
	ID              string                 `json:"id"`
	ClientID        string                 `json:"client_id"`
	CreatedAt       time.Time              `json:"created_at"`
	LastActivity    time.Time              `json:"last_activity"`
	Context         *SessionContext        `json:"context"`
	Memory          *SessionMemory         `json:"memory"`
	ActiveTools     map[string]*ToolState  `json:"active_tools"`
	Capabilities    []string               `json:"capabilities"`
	Metadata        map[string]interface{} `json:"metadata"`
	IsActive        bool                   `json:"is_active"`
	WorkflowState   *WorkflowState         `json:"workflow_state,omitempty"`
}

// SessionContext holds contextual information for the session
type SessionContext struct {
	CurrentKnowledgeBase string                 `json:"current_knowledge_base,omitempty"`
	UserPreferences      map[string]interface{} `json:"user_preferences"`
	ActiveDocuments      []string               `json:"active_documents"`
	SearchHistory        []SearchHistoryItem    `json:"search_history"`
	LastQuery            string                 `json:"last_query,omitempty"`
	ContextWindow        []ContextItem          `json:"context_window"`
}

// SessionMemory holds persistent memory for the session
type SessionMemory struct {
	ShortTerm  []MemoryItem `json:"short_term"`
	LongTerm   []MemoryItem `json:"long_term"`
	Facts      []Fact       `json:"facts"`
	MaxItems   int          `json:"max_items"`
	TTL        time.Duration `json:"ttl"`
}

// ToolState represents the state of an active tool
type ToolState struct {
	ToolName     string                 `json:"tool_name"`
	State        string                 `json:"state"` // idle, running, completed, error
	LastUsed     time.Time              `json:"last_used"`
	Parameters   map[string]interface{} `json:"parameters"`
	Results      interface{}            `json:"results,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// WorkflowState manages multi-step workflows
type WorkflowState struct {
	WorkflowID    string                 `json:"workflow_id"`
	CurrentStep   int                    `json:"current_step"`
	TotalSteps    int                    `json:"total_steps"`
	StepResults   []interface{}          `json:"step_results"`
	WorkflowType  string                 `json:"workflow_type"`
	Parameters    map[string]interface{} `json:"parameters"`
	Status        string                 `json:"status"` // running, paused, completed, error
}

// SearchHistoryItem represents a search query in history
type SearchHistoryItem struct {
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
	Results   int       `json:"results"`
}

// ContextItem represents an item in the context window
type ContextItem struct {
	Type      string      `json:"type"` // query, result, tool_call, response
	Content   interface{} `json:"content"`
	Timestamp time.Time   `json:"timestamp"`
	Relevance float64     `json:"relevance"`
}

// MemoryItem represents a memory item
type MemoryItem struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"` // fact, preference, context, result
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Relevance float64                `json:"relevance"`
	TTL       *time.Time             `json:"ttl,omitempty"`
}

// Fact represents a learned fact
type Fact struct {
	ID         string                 `json:"id"`
	Statement  string                 `json:"statement"`
	Confidence float64                `json:"confidence"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata"`
	Timestamp  time.Time              `json:"timestamp"`
}

// NewSessionManager creates a new enhanced session manager
func NewSessionManager(logger *logrus.Logger) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*MCPSession),
		logger:   logger,
		tracer:   otel.Tracer("mcp.enhanced.session"),
	}
}

// CreateSession creates a new MCP session
func (sm *SessionManager) CreateSession(ctx context.Context, clientID string) (*MCPSession, error) {
	ctx, span := sm.tracer.Start(ctx, "session_manager.create_session")
	defer span.End()

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sessionID := uuid.New().String()
	now := time.Now()

	session := &MCPSession{
		ID:           sessionID,
		ClientID:     clientID,
		CreatedAt:    now,
		LastActivity: now,
		Context: &SessionContext{
			UserPreferences: make(map[string]interface{}),
			ActiveDocuments: make([]string, 0),
			SearchHistory:   make([]SearchHistoryItem, 0),
			ContextWindow:   make([]ContextItem, 0),
		},
		Memory: &SessionMemory{
			ShortTerm: make([]MemoryItem, 0),
			LongTerm:  make([]MemoryItem, 0),
			Facts:     make([]Fact, 0),
			MaxItems:  100,
			TTL:       24 * time.Hour,
		},
		ActiveTools:  make(map[string]*ToolState),
		Capabilities: []string{"knowledge_search", "document_upload", "web_crawl", "rag_query"},
		Metadata:     make(map[string]interface{}),
		IsActive:     true,
	}

	sm.sessions[sessionID] = session

	sm.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"client_id":  clientID,
	}).Info("Created new MCP session")

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *SessionManager) GetSession(sessionID string) (*MCPSession, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// Update last activity
	session.LastActivity = time.Now()
	return session, nil
}

// UpdateSessionContext updates the session context
func (sm *SessionManager) UpdateSessionContext(sessionID string, updates map[string]interface{}) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Update context based on provided updates
	if knowledgeBase, ok := updates["knowledge_base"].(string); ok {
		session.Context.CurrentKnowledgeBase = knowledgeBase
	}

	if query, ok := updates["last_query"].(string); ok {
		session.Context.LastQuery = query
		// Add to search history
		session.Context.SearchHistory = append(session.Context.SearchHistory, SearchHistoryItem{
			Query:     query,
			Timestamp: time.Now(),
		})
	}

	session.LastActivity = time.Now()
	return nil
}

// AddToMemory adds an item to session memory
func (sm *SessionManager) AddToMemory(sessionID string, item MemoryItem) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	// Add to short-term memory
	session.Memory.ShortTerm = append(session.Memory.ShortTerm, item)

	// Manage memory size
	if len(session.Memory.ShortTerm) > session.Memory.MaxItems {
		// Move oldest items to long-term memory if relevant
		for i := 0; i < len(session.Memory.ShortTerm)-session.Memory.MaxItems; i++ {
			oldItem := session.Memory.ShortTerm[i]
			if oldItem.Relevance > 0.7 { // High relevance threshold
				session.Memory.LongTerm = append(session.Memory.LongTerm, oldItem)
			}
		}
		// Keep only recent items in short-term
		session.Memory.ShortTerm = session.Memory.ShortTerm[len(session.Memory.ShortTerm)-session.Memory.MaxItems:]
	}

	session.LastActivity = time.Now()
	return nil
}

// GetRelevantMemory retrieves relevant memory items for a query
func (sm *SessionManager) GetRelevantMemory(sessionID, query string, limit int) ([]MemoryItem, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	var relevant []MemoryItem

	// Search short-term memory
	for _, item := range session.Memory.ShortTerm {
		if sm.isRelevant(item, query) {
			relevant = append(relevant, item)
		}
	}

	// Search long-term memory if needed
	if len(relevant) < limit {
		for _, item := range session.Memory.LongTerm {
			if sm.isRelevant(item, query) {
				relevant = append(relevant, item)
			}
		}
	}

	// Sort by relevance and timestamp
	// TODO: Implement proper relevance scoring

	if len(relevant) > limit {
		relevant = relevant[:limit]
	}

	return relevant, nil
}

// isRelevant checks if a memory item is relevant to a query
func (sm *SessionManager) isRelevant(item MemoryItem, query string) bool {
	// Simple keyword matching for now
	// TODO: Implement semantic similarity
	return item.Content != "" && query != ""
}

// CleanupExpiredSessions removes expired sessions
func (sm *SessionManager) CleanupExpiredSessions(maxAge time.Duration) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	for sessionID, session := range sm.sessions {
		if now.Sub(session.LastActivity) > maxAge {
			delete(sm.sessions, sessionID)
			sm.logger.WithField("session_id", sessionID).Info("Cleaned up expired session")
		}
	}
}

// GetSessionStats returns statistics about active sessions
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_sessions":  len(sm.sessions),
		"active_sessions": 0,
		"total_tools":     0,
		"memory_items":    0,
	}

	for _, session := range sm.sessions {
		if session.IsActive {
			stats["active_sessions"] = stats["active_sessions"].(int) + 1
		}
		stats["total_tools"] = stats["total_tools"].(int) + len(session.ActiveTools)
		stats["memory_items"] = stats["memory_items"].(int) + len(session.Memory.ShortTerm) + len(session.Memory.LongTerm)
	}

	return stats
}

// SerializeSession serializes a session to JSON
func (sm *SessionManager) SerializeSession(sessionID string) ([]byte, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(session)
}

// DeserializeSession deserializes a session from JSON
func (sm *SessionManager) DeserializeSession(data []byte) (*MCPSession, error) {
	var session MCPSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.sessions[session.ID] = &session
	return &session, nil
}
