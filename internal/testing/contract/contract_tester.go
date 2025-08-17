package contract

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

// ContractTester provides API contract testing capabilities
type ContractTester struct {
	baseURL string
	client  *http.Client
	schemas map[string]*JSONSchema
}

// JSONSchema represents a JSON schema for validation
type JSONSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]*JSONSchema `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Items      *JSONSchema            `json:"items,omitempty"`
	Format     string                 `json:"format,omitempty"`
	Pattern    string                 `json:"pattern,omitempty"`
	Minimum    *float64               `json:"minimum,omitempty"`
	Maximum    *float64               `json:"maximum,omitempty"`
	MinLength  *int                   `json:"minLength,omitempty"`
	MaxLength  *int                   `json:"maxLength,omitempty"`
}

// Contract represents an API contract
type Contract struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Endpoint    string            `json:"endpoint"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Request     *ContractRequest  `json:"request,omitempty"`
	Response    *ContractResponse `json:"response"`
}

// ContractRequest represents the expected request format
type ContractRequest struct {
	Schema   *JSONSchema            `json:"schema"`
	Examples []interface{}          `json:"examples"`
	Headers  map[string]string      `json:"headers"`
}

// ContractResponse represents the expected response format
type ContractResponse struct {
	StatusCode int                    `json:"status_code"`
	Schema     *JSONSchema            `json:"schema"`
	Headers    map[string]string      `json:"headers"`
	Examples   []interface{}          `json:"examples"`
}

// ContractTestResult represents the result of a contract test
type ContractTestResult struct {
	Contract    Contract              `json:"contract"`
	Passed      bool                  `json:"passed"`
	Errors      []ContractError       `json:"errors"`
	Duration    time.Duration         `json:"duration"`
	Timestamp   time.Time             `json:"timestamp"`
	RequestData interface{}           `json:"request_data,omitempty"`
	ResponseData interface{}          `json:"response_data,omitempty"`
}

// ContractError represents a contract validation error
type ContractError struct {
	Type        string `json:"type"`
	Field       string `json:"field"`
	Expected    string `json:"expected"`
	Actual      string `json:"actual"`
	Description string `json:"description"`
}

// NewContractTester creates a new contract tester
func NewContractTester(baseURL string) *ContractTester {
	return &ContractTester{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		schemas: make(map[string]*JSONSchema),
	}
}

// RegisterSchema registers a JSON schema for validation
func (ct *ContractTester) RegisterSchema(name string, schema *JSONSchema) {
	ct.schemas[name] = schema
}

// TestContract tests a single API contract
func (ct *ContractTester) TestContract(ctx context.Context, contract Contract) (*ContractTestResult, error) {
	start := time.Now()
	
	result := &ContractTestResult{
		Contract:  contract,
		Passed:    true,
		Errors:    make([]ContractError, 0),
		Timestamp: start,
	}
	
	// Test with each example if available
	if contract.Request != nil && len(contract.Request.Examples) > 0 {
		for _, example := range contract.Request.Examples {
			if err := ct.testWithExample(ctx, contract, example, result); err != nil {
				return nil, err
			}
		}
	} else {
		// Test without request body
		if err := ct.testWithExample(ctx, contract, nil, result); err != nil {
			return nil, err
		}
	}
	
	result.Duration = time.Since(start)
	return result, nil
}

// testWithExample tests the contract with a specific example
func (ct *ContractTester) testWithExample(ctx context.Context, contract Contract, example interface{}, result *ContractTestResult) error {
	// Prepare request
	var requestBody io.Reader
	if example != nil {
		jsonData, err := json.Marshal(example)
		if err != nil {
			return fmt.Errorf("failed to marshal request example: %w", err)
		}
		requestBody = bytes.NewReader(jsonData)
		result.RequestData = example
	}
	
	// Create HTTP request
	url := ct.baseURL + contract.Endpoint
	req, err := http.NewRequestWithContext(ctx, contract.Method, url, requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// Set headers
	if contract.Headers != nil {
		for key, value := range contract.Headers {
			req.Header.Set(key, value)
		}
	}
	if contract.Request != nil && contract.Request.Headers != nil {
		for key, value := range contract.Request.Headers {
			req.Header.Set(key, value)
		}
	}
	
	// Set content type for JSON requests
	if example != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Execute request
	resp, err := ct.client.Do(req)
	if err != nil {
		result.Passed = false
		result.Errors = append(result.Errors, ContractError{
			Type:        "request_error",
			Description: fmt.Sprintf("Request failed: %v", err),
		})
		return nil
	}
	defer resp.Body.Close()
	
	// Validate response
	ct.validateResponse(resp, contract.Response, result)
	
	return nil
}

// validateResponse validates the HTTP response against the contract
func (ct *ContractTester) validateResponse(resp *http.Response, expected *ContractResponse, result *ContractTestResult) {
	// Validate status code
	if resp.StatusCode != expected.StatusCode {
		result.Passed = false
		result.Errors = append(result.Errors, ContractError{
			Type:        "status_code",
			Expected:    fmt.Sprintf("%d", expected.StatusCode),
			Actual:      fmt.Sprintf("%d", resp.StatusCode),
			Description: "Status code mismatch",
		})
	}
	
	// Validate headers
	if expected.Headers != nil {
		for key, expectedValue := range expected.Headers {
			actualValue := resp.Header.Get(key)
			if actualValue != expectedValue {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "header",
					Field:       key,
					Expected:    expectedValue,
					Actual:      actualValue,
					Description: fmt.Sprintf("Header %s mismatch", key),
				})
			}
		}
	}
	
	// Validate response body schema
	if expected.Schema != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			result.Passed = false
			result.Errors = append(result.Errors, ContractError{
				Type:        "response_read_error",
				Description: fmt.Sprintf("Failed to read response body: %v", err),
			})
			return
		}
		
		var responseData interface{}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &responseData); err != nil {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "json_parse_error",
					Description: fmt.Sprintf("Failed to parse JSON response: %v", err),
				})
				return
			}
		}
		
		result.ResponseData = responseData
		
		// Validate against schema
		ct.validateJSONSchema(responseData, expected.Schema, "", result)
	}
}

// validateJSONSchema validates data against a JSON schema
func (ct *ContractTester) validateJSONSchema(data interface{}, schema *JSONSchema, path string, result *ContractTestResult) {
	if schema == nil {
		return
	}
	
	// Type validation
	if !ct.validateType(data, schema.Type) {
		result.Passed = false
		result.Errors = append(result.Errors, ContractError{
			Type:        "type_mismatch",
			Field:       path,
			Expected:    schema.Type,
			Actual:      ct.getActualType(data),
			Description: fmt.Sprintf("Type mismatch at %s", path),
		})
		return
	}
	
	// Object validation
	if schema.Type == "object" && data != nil {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			result.Passed = false
			result.Errors = append(result.Errors, ContractError{
				Type:        "object_type_error",
				Field:       path,
				Description: "Expected object but got different type",
			})
			return
		}
		
		// Check required fields
		for _, required := range schema.Required {
			if _, exists := dataMap[required]; !exists {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "missing_required_field",
					Field:       ct.joinPath(path, required),
					Expected:    "present",
					Actual:      "missing",
					Description: fmt.Sprintf("Required field %s is missing", required),
				})
			}
		}
		
		// Validate properties
		if schema.Properties != nil {
			for key, value := range dataMap {
				if propSchema, exists := schema.Properties[key]; exists {
					ct.validateJSONSchema(value, propSchema, ct.joinPath(path, key), result)
				}
			}
		}
	}
	
	// Array validation
	if schema.Type == "array" && data != nil {
		dataArray, ok := data.([]interface{})
		if !ok {
			result.Passed = false
			result.Errors = append(result.Errors, ContractError{
				Type:        "array_type_error",
				Field:       path,
				Description: "Expected array but got different type",
			})
			return
		}
		
		// Validate items
		if schema.Items != nil {
			for i, item := range dataArray {
				itemPath := fmt.Sprintf("%s[%d]", path, i)
				ct.validateJSONSchema(item, schema.Items, itemPath, result)
			}
		}
	}
	
	// String validation
	if schema.Type == "string" && data != nil {
		str, ok := data.(string)
		if ok {
			// Length validation
			if schema.MinLength != nil && len(str) < *schema.MinLength {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "string_too_short",
					Field:       path,
					Expected:    fmt.Sprintf("min length %d", *schema.MinLength),
					Actual:      fmt.Sprintf("length %d", len(str)),
					Description: "String is too short",
				})
			}
			
			if schema.MaxLength != nil && len(str) > *schema.MaxLength {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "string_too_long",
					Field:       path,
					Expected:    fmt.Sprintf("max length %d", *schema.MaxLength),
					Actual:      fmt.Sprintf("length %d", len(str)),
					Description: "String is too long",
				})
			}
		}
	}
	
	// Number validation
	if (schema.Type == "number" || schema.Type == "integer") && data != nil {
		var num float64
		var ok bool
		
		switch v := data.(type) {
		case float64:
			num, ok = v, true
		case int:
			num, ok = float64(v), true
		case int64:
			num, ok = float64(v), true
		}
		
		if ok {
			if schema.Minimum != nil && num < *schema.Minimum {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "number_too_small",
					Field:       path,
					Expected:    fmt.Sprintf("min %f", *schema.Minimum),
					Actual:      fmt.Sprintf("%f", num),
					Description: "Number is too small",
				})
			}
			
			if schema.Maximum != nil && num > *schema.Maximum {
				result.Passed = false
				result.Errors = append(result.Errors, ContractError{
					Type:        "number_too_large",
					Field:       path,
					Expected:    fmt.Sprintf("max %f", *schema.Maximum),
					Actual:      fmt.Sprintf("%f", num),
					Description: "Number is too large",
				})
			}
		}
	}
}

// validateType validates the data type
func (ct *ContractTester) validateType(data interface{}, expectedType string) bool {
	if data == nil {
		return expectedType == "null"
	}
	
	actualType := ct.getActualType(data)
	return actualType == expectedType
}

// getActualType returns the actual type of the data
func (ct *ContractTester) getActualType(data interface{}) string {
	if data == nil {
		return "null"
	}
	
	switch data.(type) {
	case bool:
		return "boolean"
	case string:
		return "string"
	case float64, int, int64:
		return "number"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		return reflect.TypeOf(data).String()
	}
}

// joinPath joins path components
func (ct *ContractTester) joinPath(base, component string) string {
	if base == "" {
		return component
	}
	return base + "." + component
}

// TestContracts tests multiple contracts
func (ct *ContractTester) TestContracts(ctx context.Context, contracts []Contract) ([]*ContractTestResult, error) {
	results := make([]*ContractTestResult, 0, len(contracts))
	
	for _, contract := range contracts {
		result, err := ct.TestContract(ctx, contract)
		if err != nil {
			return nil, fmt.Errorf("failed to test contract %s: %w", contract.Name, err)
		}
		results = append(results, result)
	}
	
	return results, nil
}

// GenerateContractReport generates a report for contract test results
func (ct *ContractTester) GenerateContractReport(results []*ContractTestResult) string {
	var report strings.Builder
	
	report.WriteString("API Contract Test Report\n")
	report.WriteString("========================\n\n")
	
	passed := 0
	total := len(results)
	
	for _, result := range results {
		if result.Passed {
			passed++
			report.WriteString(fmt.Sprintf("✓ %s - PASSED (%v)\n", result.Contract.Name, result.Duration))
		} else {
			report.WriteString(fmt.Sprintf("✗ %s - FAILED (%v)\n", result.Contract.Name, result.Duration))
			for _, err := range result.Errors {
				report.WriteString(fmt.Sprintf("  - %s: %s\n", err.Type, err.Description))
			}
		}
	}
	
	report.WriteString(fmt.Sprintf("\nSummary: %d/%d contracts passed (%.1f%%)\n", 
		passed, total, float64(passed)/float64(total)*100))
	
	return report.String()
}

// LoadContractsFromFile loads contracts from a JSON file
func (ct *ContractTester) LoadContractsFromFile(filename string) ([]Contract, error) {
	// This would load contracts from a file
	// Implementation depends on file format (JSON, YAML, etc.)
	return nil, fmt.Errorf("not implemented")
}

// SaveContractsToFile saves contracts to a JSON file
func (ct *ContractTester) SaveContractsToFile(contracts []Contract, filename string) error {
	// This would save contracts to a file
	// Implementation depends on desired file format
	return fmt.Errorf("not implemented")
}
