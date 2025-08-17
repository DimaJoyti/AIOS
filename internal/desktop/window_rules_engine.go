package desktop

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// WindowRulesEngine manages automated window behavior rules
type WindowRulesEngine struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  RulesEngineConfig
	mu      sync.RWMutex
	
	// Rules management
	rules           []WindowRule
	ruleGroups      map[string][]WindowRule
	activeRules     map[string]*ActiveRule
	ruleHistory     []RuleExecution
	
	// Rule evaluation
	evaluationQueue chan *RuleEvaluationRequest
	running         bool
	stopCh          chan struct{}
	
	// Callbacks
	onRuleMatched   func(*WindowRule, *models.Window)
	onRuleExecuted  func(*RuleExecution)
}

// RulesEngineConfig defines rules engine configuration
type RulesEngineConfig struct {
	MaxConcurrentEvaluations int           `json:"max_concurrent_evaluations"`
	EvaluationTimeout        time.Duration `json:"evaluation_timeout"`
	RuleHistorySize          int           `json:"rule_history_size"`
	EnableProfiling          bool          `json:"enable_profiling"`
	DefaultPriority          int           `json:"default_priority"`
	CacheResults             bool          `json:"cache_results"`
	CacheTTL                 time.Duration `json:"cache_ttl"`
}

// WindowRule defines a window management rule
type WindowRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Priority    int                    `json:"priority"`
	Conditions  []RuleCondition        `json:"conditions"`
	Actions     []RuleAction           `json:"actions"`
	Triggers    []RuleTrigger          `json:"triggers"`
	Schedule    *RuleSchedule          `json:"schedule,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastMatched time.Time              `json:"last_matched"`
	MatchCount  int                    `json:"match_count"`
}

// RuleCondition defines when a rule should be applied
type RuleCondition struct {
	Type     string      `json:"type"`     // "window_title", "application", "size", "position", "workspace", "time", "custom"
	Operator string      `json:"operator"` // "equals", "contains", "matches", "greater_than", "less_than", "in_range"
	Value    interface{} `json:"value"`
	Negate   bool        `json:"negate"`
	Weight   float64     `json:"weight"`
}

// RuleAction defines what should happen when a rule matches
type RuleAction struct {
	Type       string                 `json:"type"` // "move", "resize", "workspace", "minimize", "maximize", "close", "focus", "tile", "custom"
	Parameters map[string]interface{} `json:"parameters"`
	Delay      time.Duration          `json:"delay"`
	Animation  bool                   `json:"animation"`
}

// RuleTrigger defines when rules should be evaluated
type RuleTrigger struct {
	Event     string                 `json:"event"` // "window_created", "window_focused", "window_moved", "workspace_changed", "time_based"
	Condition map[string]interface{} `json:"condition,omitempty"`
}

// RuleSchedule defines time-based rule execution
type RuleSchedule struct {
	StartTime    string   `json:"start_time"`    // "09:00"
	EndTime      string   `json:"end_time"`      // "17:00"
	DaysOfWeek   []string `json:"days_of_week"`  // ["monday", "tuesday", ...]
	Timezone     string   `json:"timezone"`
	Recurring    bool     `json:"recurring"`
	CronExpression string `json:"cron_expression,omitempty"`
}

// ActiveRule represents a rule that is currently being processed
type ActiveRule struct {
	Rule      *WindowRule `json:"rule"`
	WindowID  string      `json:"window_id"`
	StartTime time.Time   `json:"start_time"`
	Status    string      `json:"status"` // "evaluating", "executing", "completed", "failed"
	Progress  float64     `json:"progress"`
}

// RuleExecution represents the execution of a rule
type RuleExecution struct {
	RuleID      string                 `json:"rule_id"`
	WindowID    string                 `json:"window_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
	ActionsRun  []string               `json:"actions_run"`
	Context     map[string]interface{} `json:"context"`
}

// RuleEvaluationRequest represents a request to evaluate rules
type RuleEvaluationRequest struct {
	Window    *models.Window         `json:"window"`
	Event     string                 `json:"event"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
	Response  chan *RuleEvaluationResult
}

// RuleEvaluationResult represents the result of rule evaluation
type RuleEvaluationResult struct {
	MatchedRules []WindowRule `json:"matched_rules"`
	Executions   []RuleExecution `json:"executions"`
	Error        error        `json:"error,omitempty"`
}

// NewWindowRulesEngine creates a new window rules engine
func NewWindowRulesEngine(logger *logrus.Logger, config RulesEngineConfig) *WindowRulesEngine {
	tracer := otel.Tracer("window-rules-engine")
	
	engine := &WindowRulesEngine{
		logger:          logger,
		tracer:          tracer,
		config:          config,
		rules:           make([]WindowRule, 0),
		ruleGroups:      make(map[string][]WindowRule),
		activeRules:     make(map[string]*ActiveRule),
		ruleHistory:     make([]RuleExecution, 0),
		evaluationQueue: make(chan *RuleEvaluationRequest, 100),
		stopCh:          make(chan struct{}),
	}
	
	// Load default rules
	engine.loadDefaultRules()
	
	return engine
}

// Start starts the rules engine
func (wre *WindowRulesEngine) Start(ctx context.Context) error {
	ctx, span := wre.tracer.Start(ctx, "windowRulesEngine.Start")
	defer span.End()
	
	wre.mu.Lock()
	defer wre.mu.Unlock()
	
	if wre.running {
		return fmt.Errorf("rules engine already running")
	}
	
	wre.running = true
	
	// Start evaluation workers
	for i := 0; i < wre.config.MaxConcurrentEvaluations; i++ {
		go wre.evaluationWorker()
	}
	
	wre.logger.Info("Window rules engine started")
	return nil
}

// Stop stops the rules engine
func (wre *WindowRulesEngine) Stop(ctx context.Context) error {
	ctx, span := wre.tracer.Start(ctx, "windowRulesEngine.Stop")
	defer span.End()
	
	wre.mu.Lock()
	defer wre.mu.Unlock()
	
	if !wre.running {
		return nil
	}
	
	wre.running = false
	close(wre.stopCh)
	
	wre.logger.Info("Window rules engine stopped")
	return nil
}

// EvaluateRules evaluates rules for a window and event
func (wre *WindowRulesEngine) EvaluateRules(ctx context.Context, window *models.Window, event string, context map[string]interface{}) (*RuleEvaluationResult, error) {
	ctx, span := wre.tracer.Start(ctx, "windowRulesEngine.EvaluateRules")
	defer span.End()
	
	request := &RuleEvaluationRequest{
		Window:    window,
		Event:     event,
		Context:   context,
		Timestamp: time.Now(),
		Response:  make(chan *RuleEvaluationResult, 1),
	}
	
	// Send to evaluation queue
	select {
	case wre.evaluationQueue <- request:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wre.config.EvaluationTimeout):
		return nil, fmt.Errorf("evaluation timeout")
	}
	
	// Wait for result
	select {
	case result := <-request.Response:
		return result, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wre.config.EvaluationTimeout):
		return nil, fmt.Errorf("evaluation timeout")
	}
}

// evaluationWorker processes rule evaluation requests
func (wre *WindowRulesEngine) evaluationWorker() {
	for {
		select {
		case <-wre.stopCh:
			return
		case request := <-wre.evaluationQueue:
			result := wre.processEvaluationRequest(request)
			select {
			case request.Response <- result:
			default:
				// Response channel might be closed
			}
		}
	}
}

// processEvaluationRequest processes a single evaluation request
func (wre *WindowRulesEngine) processEvaluationRequest(request *RuleEvaluationRequest) *RuleEvaluationResult {
	wre.mu.RLock()
	rules := make([]WindowRule, len(wre.rules))
	copy(rules, wre.rules)
	wre.mu.RUnlock()
	
	result := &RuleEvaluationResult{
		MatchedRules: make([]WindowRule, 0),
		Executions:   make([]RuleExecution, 0),
	}
	
	// Evaluate each rule
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		
		// Check if rule should be triggered for this event
		if !wre.shouldTriggerRule(&rule, request.Event) {
			continue
		}
		
		// Check if rule schedule allows execution
		if !wre.isRuleScheduleActive(&rule) {
			continue
		}
		
		// Evaluate rule conditions
		if wre.evaluateRuleConditions(&rule, request.Window, request.Context) {
			result.MatchedRules = append(result.MatchedRules, rule)
			
			// Execute rule actions
			execution := wre.executeRuleActions(&rule, request.Window, request.Context)
			result.Executions = append(result.Executions, execution)
			
			// Update rule statistics
			wre.updateRuleStats(&rule)
			
			// Notify callback
			if wre.onRuleMatched != nil {
				wre.onRuleMatched(&rule, request.Window)
			}
		}
	}
	
	return result
}

// shouldTriggerRule checks if a rule should be triggered for an event
func (wre *WindowRulesEngine) shouldTriggerRule(rule *WindowRule, event string) bool {
	if len(rule.Triggers) == 0 {
		return true // No specific triggers means trigger on all events
	}
	
	for _, trigger := range rule.Triggers {
		if trigger.Event == event || trigger.Event == "*" {
			return true
		}
	}
	
	return false
}

// isRuleScheduleActive checks if a rule's schedule allows execution
func (wre *WindowRulesEngine) isRuleScheduleActive(rule *WindowRule) bool {
	if rule.Schedule == nil {
		return true
	}
	
	now := time.Now()
	
	// Check day of week
	if len(rule.Schedule.DaysOfWeek) > 0 {
		currentDay := strings.ToLower(now.Weekday().String())
		dayAllowed := false
		for _, day := range rule.Schedule.DaysOfWeek {
			if strings.ToLower(day) == currentDay {
				dayAllowed = true
				break
			}
		}
		if !dayAllowed {
			return false
		}
	}
	
	// Check time range (simplified)
	if rule.Schedule.StartTime != "" && rule.Schedule.EndTime != "" {
		currentTime := now.Format("15:04")
		if currentTime < rule.Schedule.StartTime || currentTime > rule.Schedule.EndTime {
			return false
		}
	}
	
	return true
}

// evaluateRuleConditions evaluates all conditions for a rule
func (wre *WindowRulesEngine) evaluateRuleConditions(rule *WindowRule, window *models.Window, context map[string]interface{}) bool {
	if len(rule.Conditions) == 0 {
		return true
	}
	
	totalWeight := 0.0
	matchedWeight := 0.0
	
	for _, condition := range rule.Conditions {
		weight := condition.Weight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight
		
		if wre.evaluateCondition(&condition, window, context) {
			matchedWeight += weight
		}
	}
	
	// Rule matches if more than 50% of weighted conditions match
	return matchedWeight/totalWeight > 0.5
}

// evaluateCondition evaluates a single condition
func (wre *WindowRulesEngine) evaluateCondition(condition *RuleCondition, window *models.Window, context map[string]interface{}) bool {
	var result bool
	
	switch condition.Type {
	case "window_title":
		result = wre.evaluateStringCondition(window.Title, condition)
	case "application":
		result = wre.evaluateStringCondition(window.Application, condition)
	case "size":
		result = wre.evaluateSizeCondition(window.Size, condition)
	case "position":
		result = wre.evaluatePositionCondition(window.Position, condition)
	case "workspace":
		result = wre.evaluateIntCondition(window.Workspace, condition)
	case "time":
		result = wre.evaluateTimeCondition(condition)
	case "custom":
		result = wre.evaluateCustomCondition(condition, window, context)
	default:
		result = false
	}
	
	// Apply negation if specified
	if condition.Negate {
		result = !result
	}
	
	return result
}

// evaluateStringCondition evaluates string-based conditions
func (wre *WindowRulesEngine) evaluateStringCondition(value string, condition *RuleCondition) bool {
	conditionValue, ok := condition.Value.(string)
	if !ok {
		return false
	}
	
	switch condition.Operator {
	case "equals":
		return value == conditionValue
	case "contains":
		return strings.Contains(strings.ToLower(value), strings.ToLower(conditionValue))
	case "matches":
		matched, _ := regexp.MatchString(conditionValue, value)
		return matched
	case "starts_with":
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(conditionValue))
	case "ends_with":
		return strings.HasSuffix(strings.ToLower(value), strings.ToLower(conditionValue))
	default:
		return false
	}
}

// evaluateSizeCondition evaluates size-based conditions
func (wre *WindowRulesEngine) evaluateSizeCondition(size models.Size, condition *RuleCondition) bool {
	// Implementation would depend on the specific size condition format
	return true // Simplified
}

// evaluatePositionCondition evaluates position-based conditions
func (wre *WindowRulesEngine) evaluatePositionCondition(position models.Position, condition *RuleCondition) bool {
	// Implementation would depend on the specific position condition format
	return true // Simplified
}

// evaluateIntCondition evaluates integer-based conditions
func (wre *WindowRulesEngine) evaluateIntCondition(value int, condition *RuleCondition) bool {
	conditionValue, ok := condition.Value.(float64) // JSON numbers are float64
	if !ok {
		return false
	}
	
	intConditionValue := int(conditionValue)
	
	switch condition.Operator {
	case "equals":
		return value == intConditionValue
	case "greater_than":
		return value > intConditionValue
	case "less_than":
		return value < intConditionValue
	case "greater_equal":
		return value >= intConditionValue
	case "less_equal":
		return value <= intConditionValue
	default:
		return false
	}
}

// evaluateTimeCondition evaluates time-based conditions
func (wre *WindowRulesEngine) evaluateTimeCondition(condition *RuleCondition) bool {
	// Implementation would depend on the specific time condition format
	return true // Simplified
}

// evaluateCustomCondition evaluates custom conditions
func (wre *WindowRulesEngine) evaluateCustomCondition(condition *RuleCondition, window *models.Window, context map[string]interface{}) bool {
	// Implementation would allow for custom condition evaluation
	return true // Simplified
}

// executeRuleActions executes all actions for a rule
func (wre *WindowRulesEngine) executeRuleActions(rule *WindowRule, window *models.Window, context map[string]interface{}) RuleExecution {
	execution := RuleExecution{
		RuleID:     rule.ID,
		WindowID:   window.ID,
		Timestamp:  time.Now(),
		Success:    true,
		ActionsRun: make([]string, 0),
		Context:    context,
	}
	
	start := time.Now()
	
	for _, action := range rule.Actions {
		// Apply delay if specified
		if action.Delay > 0 {
			time.Sleep(action.Delay)
		}
		
		// Execute action
		if err := wre.executeAction(&action, window); err != nil {
			execution.Success = false
			execution.Error = err.Error()
			break
		}
		
		execution.ActionsRun = append(execution.ActionsRun, action.Type)
	}
	
	execution.Duration = time.Since(start)
	
	// Record execution
	wre.recordExecution(execution)
	
	return execution
}

// executeAction executes a single action
func (wre *WindowRulesEngine) executeAction(action *RuleAction, window *models.Window) error {
	switch action.Type {
	case "move":
		return wre.executeMove(action, window)
	case "resize":
		return wre.executeResize(action, window)
	case "workspace":
		return wre.executeWorkspace(action, window)
	case "minimize":
		return wre.executeMinimize(action, window)
	case "maximize":
		return wre.executeMaximize(action, window)
	case "focus":
		return wre.executeFocus(action, window)
	case "close":
		return wre.executeClose(action, window)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// Action execution methods (simplified implementations)

func (wre *WindowRulesEngine) executeMove(action *RuleAction, window *models.Window) error {
	// Implementation would move the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "move",
	}).Debug("Executing move action")
	return nil
}

func (wre *WindowRulesEngine) executeResize(action *RuleAction, window *models.Window) error {
	// Implementation would resize the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "resize",
	}).Debug("Executing resize action")
	return nil
}

func (wre *WindowRulesEngine) executeWorkspace(action *RuleAction, window *models.Window) error {
	// Implementation would move window to workspace
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "workspace",
	}).Debug("Executing workspace action")
	return nil
}

func (wre *WindowRulesEngine) executeMinimize(action *RuleAction, window *models.Window) error {
	// Implementation would minimize the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "minimize",
	}).Debug("Executing minimize action")
	return nil
}

func (wre *WindowRulesEngine) executeMaximize(action *RuleAction, window *models.Window) error {
	// Implementation would maximize the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "maximize",
	}).Debug("Executing maximize action")
	return nil
}

func (wre *WindowRulesEngine) executeFocus(action *RuleAction, window *models.Window) error {
	// Implementation would focus the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "focus",
	}).Debug("Executing focus action")
	return nil
}

func (wre *WindowRulesEngine) executeClose(action *RuleAction, window *models.Window) error {
	// Implementation would close the window
	wre.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"action":    "close",
	}).Debug("Executing close action")
	return nil
}

// Helper methods

func (wre *WindowRulesEngine) updateRuleStats(rule *WindowRule) {
	wre.mu.Lock()
	defer wre.mu.Unlock()
	
	// Find and update the rule
	for i := range wre.rules {
		if wre.rules[i].ID == rule.ID {
			wre.rules[i].LastMatched = time.Now()
			wre.rules[i].MatchCount++
			break
		}
	}
}

func (wre *WindowRulesEngine) recordExecution(execution RuleExecution) {
	wre.mu.Lock()
	defer wre.mu.Unlock()
	
	wre.ruleHistory = append(wre.ruleHistory, execution)
	
	// Maintain history size
	if len(wre.ruleHistory) > wre.config.RuleHistorySize {
		wre.ruleHistory = wre.ruleHistory[len(wre.ruleHistory)-wre.config.RuleHistorySize:]
	}
	
	// Notify callback
	if wre.onRuleExecuted != nil {
		wre.onRuleExecuted(&execution)
	}
}

func (wre *WindowRulesEngine) loadDefaultRules() {
	// Load some default rules
	defaultRules := []WindowRule{
		{
			ID:          "default-browser-maximize",
			Name:        "Maximize Browser Windows",
			Description: "Automatically maximize web browser windows",
			Enabled:     true,
			Priority:    1,
			Conditions: []RuleCondition{
				{
					Type:     "application",
					Operator: "contains",
					Value:    "browser",
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type:       "maximize",
					Parameters: map[string]interface{}{},
					Animation:  true,
				},
			},
			Triggers: []RuleTrigger{
				{Event: "window_created"},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	wre.rules = append(wre.rules, defaultRules...)
}
