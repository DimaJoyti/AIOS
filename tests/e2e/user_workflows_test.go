package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	baseURL       string
	httpClient    *http.Client
	testDocuments []string
	testTemplates []string
	sessionID     string
}

func (suite *E2ETestSuite) SetupSuite() {
	// Configure test environment
	suite.baseURL = os.Getenv("AIOS_TEST_URL")
	if suite.baseURL == "" {
		suite.baseURL = "http://localhost:3000"
	}

	suite.httpClient = &http.Client{
		Timeout: 30 * time.Second,
	}

	// Generate unique session ID for this test run
	suite.sessionID = fmt.Sprintf("e2e-test-%d", time.Now().Unix())

	// Create test documents
	suite.createTestDocuments()
}

func (suite *E2ETestSuite) TearDownSuite() {
	// Cleanup test documents
	suite.cleanupTestDocuments()
}

func (suite *E2ETestSuite) createTestDocuments() {
	testDir := "test_documents"
	os.MkdirAll(testDir, 0755)

	// Create test text file
	textFile := filepath.Join(testDir, "test_document.txt")
	textContent := `This is a test document for AIOS end-to-end testing.
It contains multiple paragraphs to test document processing capabilities.

The document includes various types of content:
- Lists and bullet points
- Technical terms and concepts
- Multiple sentences and paragraphs

This content will be used to verify that the document upload,
processing, and analysis features work correctly in the AIOS platform.`

	err := os.WriteFile(textFile, []byte(textContent), 0644)
	require.NoError(suite.T(), err)
	suite.testDocuments = append(suite.testDocuments, textFile)

	// Create test markdown file
	mdFile := filepath.Join(testDir, "test_readme.md")
	mdContent := `# Test README

This is a test markdown document for AIOS testing.

## Features

- Document processing
- AI chat functionality
- Template management
- Real-time analytics

## Usage

1. Upload documents
2. Process with AI
3. Generate insights
4. Export results

### Code Example

` + "```go\nfunc main() {\n    fmt.Println(\"Hello AIOS\")\n}\n```" + `

## Conclusion

This document tests markdown processing capabilities.`

	err = os.WriteFile(mdFile, []byte(mdContent), 0644)
	require.NoError(suite.T(), err)
	suite.testDocuments = append(suite.testDocuments, mdFile)
}

func (suite *E2ETestSuite) cleanupTestDocuments() {
	for _, doc := range suite.testDocuments {
		os.Remove(doc)
	}
	os.Remove("test_documents")
}

func (suite *E2ETestSuite) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, suite.baseURL+endpoint, reqBody)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Session-ID", suite.sessionID)

	return suite.httpClient.Do(req)
}

func (suite *E2ETestSuite) TestCompleteUserWorkflow() {
	// Test the complete user journey from landing page to AI interaction

	// 1. Test landing page accessibility
	suite.T().Run("LandingPage", func(t *testing.T) {
		resp, err := suite.httpClient.Get(suite.baseURL)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		// Check for key elements
		bodyStr := string(body)
		assert.Contains(t, bodyStr, "AIOS")
		assert.Contains(t, bodyStr, "dashboard")
	})

	// 2. Test dashboard accessibility
	suite.T().Run("Dashboard", func(t *testing.T) {
		resp, err := suite.httpClient.Get(suite.baseURL + "/dashboard")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// 3. Test AI models endpoint
	suite.T().Run("AIModels", func(t *testing.T) {
		resp, err := suite.makeRequest("GET", "/api/ai/models", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var models []map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&models)
			require.NoError(t, err)
			assert.NotEmpty(t, models)
		}
	})

	// 4. Test chat functionality
	suite.T().Run("ChatFunctionality", func(t *testing.T) {
		chatRequest := map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": "Hello, this is an end-to-end test message.",
				},
			},
			"model_id":   "gpt-3.5-turbo",
			"session_id": suite.sessionID,
		}

		resp, err := suite.makeRequest("POST", "/api/ai/chat", chatRequest)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "text")
			assert.Contains(t, response, "usage")
			assert.NotEmpty(t, response["text"])
		} else {
			// Log the error for debugging
			body, _ := io.ReadAll(resp.Body)
			suite.T().Logf("Chat request failed with status %d: %s", resp.StatusCode, string(body))
		}
	})
}

func (suite *E2ETestSuite) TestDocumentWorkflow() {
	// Test complete document upload and processing workflow

	suite.T().Run("DocumentUpload", func(t *testing.T) {
		if len(suite.testDocuments) == 0 {
			t.Skip("No test documents available")
		}

		for _, docPath := range suite.testDocuments {
			t.Run(filepath.Base(docPath), func(t *testing.T) {
				// Open test document
				file, err := os.Open(docPath)
				require.NoError(t, err)
				defer file.Close()

				// Create multipart form
				var buf bytes.Buffer
				writer := multipart.NewWriter(&buf)

				part, err := writer.CreateFormFile("file", filepath.Base(docPath))
				require.NoError(t, err)

				_, err = io.Copy(part, file)
				require.NoError(t, err)

				err = writer.Close()
				require.NoError(t, err)

				// Create request
				req, err := http.NewRequest("POST", suite.baseURL+"/api/documents/upload", &buf)
				require.NoError(t, err)

				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("X-Session-ID", suite.sessionID)

				// Send request
				resp, err := suite.httpClient.Do(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				if resp.StatusCode == http.StatusOK {
					var response map[string]interface{}
					err = json.NewDecoder(resp.Body).Decode(&response)
					require.NoError(t, err)

					assert.Contains(t, response, "id")
					assert.Contains(t, response, "name")
					assert.Contains(t, response, "status")

					// Store document ID for later tests
					if docID, ok := response["id"].(string); ok {
						suite.testTemplates = append(suite.testTemplates, docID)
					}
				} else {
					body, _ := io.ReadAll(resp.Body)
					t.Logf("Document upload failed with status %d: %s", resp.StatusCode, string(body))
				}
			})
		}
	})

	suite.T().Run("DocumentList", func(t *testing.T) {
		resp, err := suite.makeRequest("GET", "/api/documents/upload", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			// Should have documents or be empty
			assert.NotNil(t, response)
		}
	})
}

func (suite *E2ETestSuite) TestTemplateWorkflow() {
	// Test template creation, execution, and management

	suite.T().Run("TemplateCreation", func(t *testing.T) {
		template := map[string]interface{}{
			"id":          fmt.Sprintf("e2e-test-template-%d", time.Now().Unix()),
			"name":        "E2E Test Template",
			"description": "Template created during end-to-end testing",
			"category":    "test",
			"template":    "Analyze the following text and provide {{.analysis_type}} analysis:\n\n{{.text}}",
			"variables": []map[string]interface{}{
				{
					"name":        "text",
					"type":        "string",
					"description": "Text to analyze",
					"required":    true,
				},
				{
					"name":        "analysis_type",
					"type":        "string",
					"description": "Type of analysis to perform",
					"required":    false,
					"default":     "detailed",
				},
			},
			"config": map[string]interface{}{
				"model_id":    "gpt-3.5-turbo",
				"temperature": 0.7,
				"max_tokens":  500,
			},
			"tags": []string{"test", "e2e", "analysis"},
		}

		resp, err := suite.makeRequest("POST", "/api/ai/templates", template)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, template["id"], response["id"])
			assert.Contains(t, response, "created_at")

			// Store template ID for execution test
			if templateID, ok := response["id"].(string); ok {
				suite.testTemplates = append(suite.testTemplates, templateID)
			}
		} else {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Template creation failed with status %d: %s", resp.StatusCode, string(body))
		}
	})

	suite.T().Run("TemplateExecution", func(t *testing.T) {
		if len(suite.testTemplates) == 0 {
			t.Skip("No test templates available")
		}

		templateID := suite.testTemplates[0]

		execution := map[string]interface{}{
			"variables": map[string]interface{}{
				"text":          "This is a sample text for analysis during end-to-end testing.",
				"analysis_type": "summary",
			},
			"user_id":    "e2e-test-user",
			"session_id": suite.sessionID,
		}

		endpoint := fmt.Sprintf("/api/ai/templates/%s/execute", templateID)
		resp, err := suite.makeRequest("POST", endpoint, execution)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var response map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&response)
			require.NoError(t, err)

			assert.Contains(t, response, "text")
			assert.Contains(t, response, "template_id")
			assert.Equal(t, templateID, response["template_id"])
		} else {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Template execution failed with status %d: %s", resp.StatusCode, string(body))
		}
	})

	suite.T().Run("TemplateList", func(t *testing.T) {
		resp, err := suite.makeRequest("GET", "/api/ai/templates", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var templates []map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&templates)
			require.NoError(t, err)

			// Should have at least our test template
			assert.NotEmpty(t, templates)

			// Check if our test template is in the list
			found := false
			for _, template := range templates {
				if strings.Contains(template["id"].(string), "e2e-test-template") {
					found = true
					break
				}
			}
			assert.True(t, found, "Test template should be in the list")
		}
	})
}

func (suite *E2ETestSuite) TestErrorHandling() {
	// Test various error scenarios

	suite.T().Run("InvalidChatRequest", func(t *testing.T) {
		invalidRequest := map[string]interface{}{
			"model_id": "non-existent-model",
			// Missing required messages field
		}

		resp, err := suite.makeRequest("POST", "/api/ai/chat", invalidRequest)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400, "Should return error status")
	})

	suite.T().Run("InvalidTemplateExecution", func(t *testing.T) {
		execution := map[string]interface{}{
			"variables": map[string]interface{}{
				// Missing required variables
			},
		}

		resp, err := suite.makeRequest("POST", "/api/ai/templates/non-existent/execute", execution)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400, "Should return error status")
	})

	suite.T().Run("InvalidDocumentUpload", func(t *testing.T) {
		// Try to upload invalid file type
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		part, err := writer.CreateFormFile("file", "test.exe")
		require.NoError(t, err)

		part.Write([]byte("invalid file content"))
		writer.Close()

		req, err := http.NewRequest("POST", suite.baseURL+"/api/documents/upload", &buf)
		require.NoError(t, err)

		req.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err := suite.httpClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= 400, "Should reject invalid file type")
	})
}

func (suite *E2ETestSuite) TestPerformanceUnderLoad() {
	// Test system performance under concurrent load

	suite.T().Run("ConcurrentChatRequests", func(t *testing.T) {
		const numRequests = 10
		const timeout = 30 * time.Second

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		results := make(chan error, numRequests)

		for i := 0; i < numRequests; i++ {
			go func(requestID int) {
				chatRequest := map[string]interface{}{
					"messages": []map[string]interface{}{
						{
							"role":    "user",
							"content": fmt.Sprintf("Concurrent test request #%d", requestID),
						},
					},
					"model_id":   "gpt-3.5-turbo",
					"session_id": fmt.Sprintf("%s-concurrent-%d", suite.sessionID, requestID),
				}

				resp, err := suite.makeRequest("POST", "/api/ai/chat", chatRequest)
				if err != nil {
					results <- err
					return
				}
				defer resp.Body.Close()

				if resp.StatusCode != http.StatusOK {
					results <- fmt.Errorf("request %d failed with status %d", requestID, resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Collect results
		successCount := 0
		for i := 0; i < numRequests; i++ {
			select {
			case err := <-results:
				if err == nil {
					successCount++
				} else {
					t.Logf("Request failed: %v", err)
				}
			case <-ctx.Done():
				t.Fatal("Test timed out")
			}
		}

		// At least 70% of requests should succeed
		successRate := float64(successCount) / float64(numRequests)
		assert.True(t, successRate >= 0.7,
			"Success rate should be at least 70%%, got %.2f%%", successRate*100)

		t.Logf("Concurrent requests: %d/%d succeeded (%.2f%%)",
			successCount, numRequests, successRate*100)
	})
}

func TestE2EWorkflows(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
