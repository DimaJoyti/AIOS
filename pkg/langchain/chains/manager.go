package chains

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ChainManager defines the interface for managing chains
type ChainManager interface {
	// Chain management
	RegisterChain(chain Chain) error
	UnregisterChain(chainID string) error
	GetChain(chainID string) (Chain, error)
	ListChains() []Chain

	// Execution
	ExecuteChain(ctx context.Context, chainID string, input ChainInput) (ChainOutput, error)

	// Configuration
	GetChainTypes() []string
	CreateChain(chainType string, config *ChainConfig) (Chain, error)

	// Monitoring
	GetMetrics() *ChainManagerMetrics
	GetStatus() *ChainManagerStatus
}

// Note: Chain, ChainInput, and ChainOutput are defined in interfaces.go

// Note: ChainConfig is defined in interfaces.go

// ChainManagerMetrics represents chain manager metrics
type ChainManagerMetrics struct {
	TotalExecutions      int64         `json:"total_executions"`
	SuccessfulExecutions int64         `json:"successful_executions"`
	FailedExecutions     int64         `json:"failed_executions"`
	AverageLatency       time.Duration `json:"average_latency"`
	ErrorRate            float64       `json:"error_rate"`
	StartTime            time.Time     `json:"start_time"`
	LastUpdate           time.Time     `json:"last_update"`
}

// ChainManagerStatus represents chain manager status
type ChainManagerStatus struct {
	Running          bool            `json:"running"`
	ChainCount       int             `json:"chain_count"`
	AvailableChains  int             `json:"available_chains"`
	ActiveExecutions int             `json:"active_executions"`
	ChainStatuses    map[string]bool `json:"chain_statuses"`
	LastActivity     time.Time       `json:"last_activity"`
}

// DefaultChainManager implements the ChainManager interface
type DefaultChainManager struct {
	chains  map[string]Chain
	logger  *logrus.Logger
	metrics *ChainManagerMetrics
}

// NewDefaultChainManager creates a new default chain manager
func NewDefaultChainManager(logger *logrus.Logger) ChainManager {
	return &DefaultChainManager{
		chains: make(map[string]Chain),
		logger: logger,
		metrics: &ChainManagerMetrics{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
	}
}

// RegisterChain registers a chain
func (m *DefaultChainManager) RegisterChain(chain Chain) error {
	chainID := chain.GetChainType() // Use GetChainType as identifier
	m.chains[chainID] = chain
	m.logger.WithField("chain_id", chainID).Info("Chain registered")
	return nil
}

// UnregisterChain unregisters a chain
func (m *DefaultChainManager) UnregisterChain(chainID string) error {
	delete(m.chains, chainID)
	m.logger.WithField("chain_id", chainID).Info("Chain unregistered")
	return nil
}

// GetChain gets a chain by ID
func (m *DefaultChainManager) GetChain(chainID string) (Chain, error) {
	chain, exists := m.chains[chainID]
	if !exists {
		return nil, fmt.Errorf("chain not found: %s", chainID)
	}
	return chain, nil
}

// ListChains lists all chains
func (m *DefaultChainManager) ListChains() []Chain {
	chains := make([]Chain, 0, len(m.chains))
	for _, chain := range m.chains {
		chains = append(chains, chain)
	}
	return chains
}

// ExecuteChain executes a chain
func (m *DefaultChainManager) ExecuteChain(ctx context.Context, chainID string, input ChainInput) (ChainOutput, error) {
	startTime := time.Now()

	chain, err := m.GetChain(chainID)
	if err != nil {
		m.updateMetrics(false, time.Since(startTime))
		return nil, err
	}

	output, err := chain.Run(ctx, input)
	if err != nil {
		m.updateMetrics(false, time.Since(startTime))
		return nil, err
	}

	m.updateMetrics(true, time.Since(startTime))
	return output, nil
}

// GetChainTypes returns available chain types
func (m *DefaultChainManager) GetChainTypes() []string {
	return []string{
		"llm_chain",
		"sequential_chain",
		"router_chain",
		"transform_chain",
		"retrieval_chain",
	}
}

// CreateChain creates a new chain
func (m *DefaultChainManager) CreateChain(chainType string, config *ChainConfig) (Chain, error) {
	// TODO: Implement chain creation
	return nil, fmt.Errorf("chain creation not yet implemented for type: %s", chainType)
}

// GetMetrics returns chain manager metrics
func (m *DefaultChainManager) GetMetrics() *ChainManagerMetrics {
	m.metrics.LastUpdate = time.Now()
	return m.metrics
}

// GetStatus returns chain manager status
func (m *DefaultChainManager) GetStatus() *ChainManagerStatus {
	chainStatuses := make(map[string]bool)
	availableCount := 0

	for id, chain := range m.chains {
		// TODO: Implement availability check
		available := true // Assume available for now
		chainStatuses[id] = available
		if available {
			availableCount++
		}
		_ = chain // Use chain to avoid unused variable warning
	}

	return &ChainManagerStatus{
		Running:         true,
		ChainCount:      len(m.chains),
		AvailableChains: availableCount,
		ChainStatuses:   chainStatuses,
		LastActivity:    time.Now(),
	}
}

// updateMetrics updates chain manager metrics
func (m *DefaultChainManager) updateMetrics(success bool, latency time.Duration) {
	m.metrics.TotalExecutions++
	if success {
		m.metrics.SuccessfulExecutions++
	} else {
		m.metrics.FailedExecutions++
	}

	if m.metrics.TotalExecutions > 0 {
		m.metrics.ErrorRate = float64(m.metrics.FailedExecutions) / float64(m.metrics.TotalExecutions)
	}

	// Update average latency
	if m.metrics.AverageLatency == 0 {
		m.metrics.AverageLatency = latency
	} else {
		m.metrics.AverageLatency = (m.metrics.AverageLatency + latency) / 2
	}
}

// TODO: Chain implementations will be provided by the actual chain packages
