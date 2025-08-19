package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelper provides common testing utilities
type TestHelper struct {
	t       *testing.T
	cleanup []func()
	mu      sync.Mutex
}

// NewTestHelper creates a new test helper
func NewTestHelper(t *testing.T) *TestHelper {
	helper := &TestHelper{
		t:       t,
		cleanup: make([]func(), 0),
	}

	t.Cleanup(helper.Cleanup)
	return helper
}

// Cleanup runs all registered cleanup functions
func (h *TestHelper) Cleanup() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for i := len(h.cleanup) - 1; i >= 0; i-- {
		h.cleanup[i]()
	}
}

// AddCleanup registers a cleanup function
func (h *TestHelper) AddCleanup(fn func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.cleanup = append(h.cleanup, fn)
}

// TempDir creates a temporary directory for testing
func (h *TestHelper) TempDir() string {
	dir, err := os.MkdirTemp("", "aios-test-*")
	require.NoError(h.t, err)

	h.AddCleanup(func() {
		os.RemoveAll(dir)
	})

	return dir
}

// TempFile creates a temporary file for testing
func (h *TestHelper) TempFile(content string) string {
	file, err := os.CreateTemp("", "aios-test-*.tmp")
	require.NoError(h.t, err)

	if content != "" {
		_, err = file.WriteString(content)
		require.NoError(h.t, err)
	}

	err = file.Close()
	require.NoError(h.t, err)

	h.AddCleanup(func() {
		os.Remove(file.Name())
	})

	return file.Name()
}

// CreateTestFile creates a test file with specified content
func (h *TestHelper) CreateTestFile(dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(h.t, err)
	return filePath
}

// AssertEventually asserts that a condition becomes true within a timeout
func (h *TestHelper) AssertEventually(condition func() bool, timeout time.Duration, message string) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if condition() {
				return
			}
		case <-timeoutCh:
			h.t.Fatalf("Condition not met within timeout: %s", message)
		}
	}
}

// AssertNever asserts that a condition never becomes true within a duration
func (h *TestHelper) AssertNever(condition func() bool, duration time.Duration, message string) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	timeoutCh := time.After(duration)

	for {
		select {
		case <-ticker.C:
			if condition() {
				h.t.Fatalf("Condition became true when it shouldn't: %s", message)
			}
		case <-timeoutCh:
			return
		}
	}
}

// WithTimeout runs a function with a timeout context
func (h *TestHelper) WithTimeout(timeout time.Duration, fn func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(ctx)
	}()

	select {
	case <-done:
		// Function completed successfully
	case <-ctx.Done():
		h.t.Fatalf("Function timed out after %v", timeout)
	}
}

// ConcurrentTest runs multiple test functions concurrently
func (h *TestHelper) ConcurrentTest(tests ...func()) {
	var wg sync.WaitGroup
	errors := make(chan error, len(tests))

	for _, test := range tests {
		wg.Add(1)
		go func(testFn func()) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					errors <- fmt.Errorf("test panicked: %v", r)
				}
			}()
			testFn()
		}(test)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		h.t.Error(err)
	}
}

// HTTPTestHelper provides HTTP testing utilities
type HTTPTestHelper struct {
	*TestHelper
	server *httptest.Server
	client *http.Client
}

// NewHTTPTestHelper creates a new HTTP test helper
func NewHTTPTestHelper(t *testing.T, handler http.Handler) *HTTPTestHelper {
	helper := &HTTPTestHelper{
		TestHelper: NewTestHelper(t),
		server:     httptest.NewServer(handler),
		client:     &http.Client{Timeout: 30 * time.Second},
	}

	helper.AddCleanup(helper.server.Close)
	return helper
}

// URL returns the test server URL
func (h *HTTPTestHelper) URL() string {
	return h.server.URL
}

// GET performs a GET request
func (h *HTTPTestHelper) GET(path string) *http.Response {
	resp, err := h.client.Get(h.server.URL + path)
	require.NoError(h.t, err)
	return resp
}

// POST performs a POST request
func (h *HTTPTestHelper) POST(path string, body io.Reader) *http.Response {
	resp, err := h.client.Post(h.server.URL+path, "application/json", body)
	require.NoError(h.t, err)
	return resp
}

// PUT performs a PUT request
func (h *HTTPTestHelper) PUT(path string, body io.Reader) *http.Response {
	req, err := http.NewRequest("PUT", h.server.URL+path, body)
	require.NoError(h.t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	require.NoError(h.t, err)
	return resp
}

// DELETE performs a DELETE request
func (h *HTTPTestHelper) DELETE(path string) *http.Response {
	req, err := http.NewRequest("DELETE", h.server.URL+path, nil)
	require.NoError(h.t, err)

	resp, err := h.client.Do(req)
	require.NoError(h.t, err)
	return resp
}

// AssertJSONResponse asserts that the response contains expected JSON
func (h *HTTPTestHelper) AssertJSONResponse(resp *http.Response, expectedStatus int, expected interface{}) {
	defer resp.Body.Close()

	assert.Equal(h.t, expectedStatus, resp.StatusCode)
	assert.Equal(h.t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(h.t, err)

	if expected != nil {
		var actual interface{}
		err = json.Unmarshal(body, &actual)
		require.NoError(h.t, err)

		assert.Equal(h.t, expected, actual)
	}
}

// AssertErrorResponse asserts that the response contains an error
func (h *HTTPTestHelper) AssertErrorResponse(resp *http.Response, expectedStatus int, expectedMessage string) {
	defer resp.Body.Close()

	assert.Equal(h.t, expectedStatus, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(h.t, err)

	var errorResp map[string]interface{}
	err = json.Unmarshal(body, &errorResp)
	require.NoError(h.t, err)

	if expectedMessage != "" {
		assert.Contains(h.t, errorResp["error"], expectedMessage)
	}
}

// TestDataGenerator provides test data generation utilities
type TestDataGenerator struct{}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator() *TestDataGenerator {
	return &TestDataGenerator{}
}

// RandomString generates a random string of specified length
func (g *TestDataGenerator) RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// RandomEmail generates a random email address
func (g *TestDataGenerator) RandomEmail() string {
	return fmt.Sprintf("%s@%s.com", g.RandomString(8), g.RandomString(6))
}

// RandomInt generates a random integer between min and max
func (g *TestDataGenerator) RandomInt(min, max int) int {
	return min + int(time.Now().UnixNano())%(max-min+1)
}

// TestAssertions provides enhanced assertions
type TestAssertions struct {
	t *testing.T
}

// NewTestAssertions creates new test assertions
func NewTestAssertions(t *testing.T) *TestAssertions {
	return &TestAssertions{t: t}
}

// AssertStructEqual asserts that two structs are equal, ignoring specified fields
func (a *TestAssertions) AssertStructEqual(expected, actual interface{}, ignoreFields ...string) {
	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	if expectedVal.Type() != actualVal.Type() {
		a.t.Fatalf("Types don't match: expected %T, got %T", expected, actual)
	}

	a.compareStructs(expectedVal, actualVal, ignoreFields, "")
}

func (a *TestAssertions) compareStructs(expected, actual reflect.Value, ignoreFields []string, path string) {
	if expected.Type() != actual.Type() {
		a.t.Fatalf("Types don't match at %s: expected %v, got %v", path, expected.Type(), actual.Type())
	}

	switch expected.Kind() {
	case reflect.Struct:
		for i := 0; i < expected.NumField(); i++ {
			field := expected.Type().Field(i)
			fieldPath := path + "." + field.Name

			// Skip ignored fields
			if a.shouldIgnoreField(field.Name, ignoreFields) {
				continue
			}

			expectedField := expected.Field(i)
			actualField := actual.Field(i)

			a.compareStructs(expectedField, actualField, ignoreFields, fieldPath)
		}
	case reflect.Slice, reflect.Array:
		if expected.Len() != actual.Len() {
			a.t.Fatalf("Slice lengths don't match at %s: expected %d, got %d", path, expected.Len(), actual.Len())
		}

		for i := 0; i < expected.Len(); i++ {
			a.compareStructs(expected.Index(i), actual.Index(i), ignoreFields, fmt.Sprintf("%s[%d]", path, i))
		}
	case reflect.Map:
		if expected.Len() != actual.Len() {
			a.t.Fatalf("Map lengths don't match at %s: expected %d, got %d", path, expected.Len(), actual.Len())
		}

		for _, key := range expected.MapKeys() {
			expectedVal := expected.MapIndex(key)
			actualVal := actual.MapIndex(key)

			if !actualVal.IsValid() {
				a.t.Fatalf("Key %v not found in actual map at %s", key.Interface(), path)
			}

			a.compareStructs(expectedVal, actualVal, ignoreFields, fmt.Sprintf("%s[%v]", path, key.Interface()))
		}
	default:
		if !reflect.DeepEqual(expected.Interface(), actual.Interface()) {
			a.t.Fatalf("Values don't match at %s: expected %v, got %v", path, expected.Interface(), actual.Interface())
		}
	}
}

func (a *TestAssertions) shouldIgnoreField(fieldName string, ignoreFields []string) bool {
	for _, ignore := range ignoreFields {
		if fieldName == ignore {
			return true
		}
	}
	return false
}

// GetCallerInfo returns information about the test caller
func GetCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown", 0
	}

	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	filename := parts[len(parts)-1]

	return filename, line
}

// SkipIfShort skips the test if running in short mode
func SkipIfShort(t *testing.T, reason string) {
	if testing.Short() {
		t.Skipf("Skipping test in short mode: %s", reason)
	}
}

// RequireEnv requires an environment variable to be set
func RequireEnv(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// SetEnv sets an environment variable for the duration of the test
func SetEnv(t *testing.T, key, value string) {
	oldValue := os.Getenv(key)
	err := os.Setenv(key, value)
	require.NoError(t, err)

	t.Cleanup(func() {
		if oldValue == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, oldValue)
		}
	})
}
