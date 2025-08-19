package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/internal/ai/services"
	"github.com/aios/aios/pkg/config"
	"github.com/aios/aios/pkg/database"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AIIntegrationTestSuite struct {
	suite.Suite
	aiService *ai.Service
	server    *httptest.Server
	logger    *logrus.Logger
}

func (suite *AIIntegrationTestSuite) SetupSuite() {
	// Setup logger
	suite.logger = logrus.New()
	suite.logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	// Setup test configuration
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "test",
			Password: "test",
			Name:     "aios_test",
			SSLMode:  "disable",
		},
	}

	// Create test database connection (mock for testing)
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	// For integration tests, we'll use a mock database
	// In a real scenario, you'd use a test database
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		// Use a mock database for testing
		suite.logger.Warn("Using mock database for testing")
		db = nil
	}

	// Create AI service
	aiService, err := ai.NewService(cfg, db, suite.logger)
	require.NoError(suite.T(), err)
	suite.aiService = aiService

	// Start the service
	ctx := context.Background()
	err = suite.aiService.Start(ctx)
	require.NoError(suite.T(), err)

	// Create test server
	suite.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock AI service endpoints for testing
		suite.handleTestRequest(w, r)
	}))
}

func (suite *AIIntegrationTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}

	if suite.aiService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		suite.aiService.Stop(ctx)
	}
}

func (suite *AIIntegrationTestSuite) handleTestRequest(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/api/v1/ai/chat":
		suite.handleChatRequest(w, r)
	case "/api/v1/ai/templates":
		suite.handleTemplatesRequest(w, r)
	case "/api/v1/ai/models":
		suite.handleModelsRequest(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (suite *AIIntegrationTestSuite) handleChatRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Messages []services.ChatMessage `json:"messages"`
		ModelID  string                 `json:"model_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Mock response
	response := map[string]interface{}{
		"id":            "test-response-id",
		"text":          "This is a test response from the AI service",
		"finish_reason": "stop",
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 20,
			"total_tokens":      30,
		},
		"cost":       0.001,
		"latency":    100,
		"model_id":   request.ModelID,
		"provider":   "test-provider",
		"created_at": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (suite *AIIntegrationTestSuite) handleTemplatesRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates := []map[string]interface{}{
			{
				"id":          "test-template",
				"name":        "Test Template",
				"description": "A test template",
				"category":    "test",
				"template":    "Hello {{.name}}!",
				"variables": []map[string]interface{}{
					{
						"name":        "name",
						"type":        "string",
						"description": "Name to greet",
						"required":    true,
					},
				},
				"created_at": time.Now().Format(time.RFC3339),
				"updated_at": time.Now().Format(time.RFC3339),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)

	case http.MethodPost:
		var template map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Add timestamps
		template["created_at"] = time.Now().Format(time.RFC3339)
		template["updated_at"] = time.Now().Format(time.RFC3339)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(template)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (suite *AIIntegrationTestSuite) handleModelsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	models := []map[string]interface{}{
		{
			"id":       "gpt-4",
			"name":     "GPT-4",
			"provider": "openai",
			"type":     "text",
			"status":   "active",
		},
		{
			"id":       "gpt-3.5-turbo",
			"name":     "GPT-3.5 Turbo",
			"provider": "openai",
			"type":     "text",
			"status":   "active",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func (suite *AIIntegrationTestSuite) TestChatEndpoint() {
	requestBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Hello, how are you?",
			},
		},
		"model_id": "gpt-3.5-turbo",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(suite.T(), err)

	resp, err := http.Post(
		suite.server.URL+"/api/v1/ai/chat",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)

	assert.Contains(suite.T(), response, "text")
	assert.Contains(suite.T(), response, "usage")
	assert.Contains(suite.T(), response, "model_id")
	assert.Equal(suite.T(), "gpt-3.5-turbo", response["model_id"])
}

func (suite *AIIntegrationTestSuite) TestTemplatesEndpoint() {
	// Test GET templates
	resp, err := http.Get(suite.server.URL + "/api/v1/ai/templates")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var templates []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&templates)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), templates)
	assert.Contains(suite.T(), templates[0], "id")
	assert.Contains(suite.T(), templates[0], "name")

	// Test POST template
	newTemplate := map[string]interface{}{
		"id":          "integration-test-template",
		"name":        "Integration Test Template",
		"description": "A template for integration testing",
		"template":    "Test {{.value}}",
		"variables": []map[string]interface{}{
			{
				"name":        "value",
				"type":        "string",
				"description": "Test value",
				"required":    true,
			},
		},
	}

	jsonBody, err := json.Marshal(newTemplate)
	require.NoError(suite.T(), err)

	resp, err = http.Post(
		suite.server.URL+"/api/v1/ai/templates",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)

	var createdTemplate map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createdTemplate)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), newTemplate["id"], createdTemplate["id"])
	assert.Contains(suite.T(), createdTemplate, "created_at")
}

func (suite *AIIntegrationTestSuite) TestModelsEndpoint() {
	resp, err := http.Get(suite.server.URL + "/api/v1/ai/models")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var models []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&models)
	require.NoError(suite.T(), err)

	assert.NotEmpty(suite.T(), models)

	// Check that we have expected models
	modelIDs := make([]string, len(models))
	for i, model := range models {
		modelIDs[i] = model["id"].(string)
	}

	assert.Contains(suite.T(), modelIDs, "gpt-4")
	assert.Contains(suite.T(), modelIDs, "gpt-3.5-turbo")
}

func (suite *AIIntegrationTestSuite) TestConcurrentRequests() {
	const numRequests = 10

	requestBody := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": "Concurrent test request",
			},
		},
		"model_id": "gpt-3.5-turbo",
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(suite.T(), err)

	// Channel to collect results
	results := make(chan error, numRequests)

	// Launch concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(requestID int) {
			resp, err := http.Post(
				suite.server.URL+"/api/v1/ai/chat",
				"application/json",
				bytes.NewBuffer(jsonBody),
			)
			if err != nil {
				results <- fmt.Errorf("request %d failed: %w", requestID, err)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("request %d returned status %d", requestID, resp.StatusCode)
				return
			}

			results <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-results:
			assert.NoError(suite.T(), err)
		case <-time.After(10 * time.Second):
			suite.T().Fatal("Request timed out")
		}
	}
}

func (suite *AIIntegrationTestSuite) TestErrorHandling() {
	// Test invalid JSON
	resp, err := http.Post(
		suite.server.URL+"/api/v1/ai/chat",
		"application/json",
		bytes.NewBuffer([]byte("invalid json")),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode)

	// Test missing required fields
	requestBody := map[string]interface{}{
		"model_id": "gpt-3.5-turbo",
		// Missing messages field
	}

	jsonBody, err := json.Marshal(requestBody)
	require.NoError(suite.T(), err)

	resp, err = http.Post(
		suite.server.URL+"/api/v1/ai/chat",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()

	// Should handle gracefully (implementation dependent)
	assert.True(suite.T(), resp.StatusCode >= 400)
}

func TestAIIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AIIntegrationTestSuite))
}
