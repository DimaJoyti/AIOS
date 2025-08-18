package fixtures

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aios/aios/pkg/models"
	"gopkg.in/yaml.v3"
)

// FixtureManager manages test fixtures and data
type FixtureManager struct {
	basePath string
	cache    map[string]interface{}
	mu       sync.RWMutex
}

// NewFixtureManager creates a new fixture manager
func NewFixtureManager(basePath string) *FixtureManager {
	return &FixtureManager{
		basePath: basePath,
		cache:    make(map[string]interface{}),
	}
}

// LoadFixture loads a fixture from file
func (f *FixtureManager) LoadFixture(name string, target interface{}) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check cache first
	if cached, exists := f.cache[name]; exists {
		return f.copyValue(cached, target)
	}

	// Try different file extensions
	extensions := []string{".json", ".yaml", ".yml"}
	var data []byte
	var err error

	for _, ext := range extensions {
		filePath := filepath.Join(f.basePath, name+ext)
		data, err = os.ReadFile(filePath)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("fixture %s not found: %w", name, err)
	}

	// Determine format and unmarshal
	if strings.HasSuffix(name, ".json") || json.Valid(data) {
		err = json.Unmarshal(data, target)
	} else {
		err = yaml.Unmarshal(data, target)
	}

	if err != nil {
		return fmt.Errorf("failed to unmarshal fixture %s: %w", name, err)
	}

	// Cache the result
	f.cache[name] = f.cloneValue(target)

	return nil
}

// SaveFixture saves a fixture to file
func (f *FixtureManager) SaveFixture(name string, data interface{}, format string) error {
	var content []byte
	var err error
	var ext string

	switch strings.ToLower(format) {
	case "json":
		content, err = json.MarshalIndent(data, "", "  ")
		ext = ".json"
	case "yaml", "yml":
		content, err = yaml.Marshal(data)
		ext = ".yaml"
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal fixture: %w", err)
	}

	filePath := filepath.Join(f.basePath, name+ext)

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(filePath, content, 0644)
}

// ListFixtures lists all available fixtures
func (f *FixtureManager) ListFixtures() ([]string, error) {
	var fixtures []string

	err := filepath.WalkDir(f.basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".json" || ext == ".yaml" || ext == ".yml" {
				relPath, _ := filepath.Rel(f.basePath, path)
				name := strings.TrimSuffix(relPath, ext)
				fixtures = append(fixtures, name)
			}
		}

		return nil
	})

	return fixtures, err
}

// ClearCache clears the fixture cache
func (f *FixtureManager) ClearCache() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cache = make(map[string]interface{})
}

// copyValue copies a value using reflection
func (f *FixtureManager) copyValue(src, dst interface{}) error {
	srcVal := reflect.ValueOf(src)
	dstVal := reflect.ValueOf(dst)

	if dstVal.Kind() != reflect.Ptr {
		return fmt.Errorf("destination must be a pointer")
	}

	dstVal.Elem().Set(srcVal)
	return nil
}

// cloneValue creates a deep copy of a value
func (f *FixtureManager) cloneValue(src interface{}) interface{} {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}

	clone := reflect.New(srcVal.Type())
	clone.Elem().Set(srcVal)

	return clone.Interface()
}

// TestDataBuilder provides a fluent interface for building test data
type TestDataBuilder struct {
	data map[string]interface{}
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder() *TestDataBuilder {
	return &TestDataBuilder{
		data: make(map[string]interface{}),
	}
}

// WithField sets a field value
func (b *TestDataBuilder) WithField(key string, value interface{}) *TestDataBuilder {
	b.data[key] = value
	return b
}

// WithRandomString sets a field to a random string
func (b *TestDataBuilder) WithRandomString(key string, length int) *TestDataBuilder {
	b.data[key] = randomString(length)
	return b
}

// WithRandomInt sets a field to a random integer
func (b *TestDataBuilder) WithRandomInt(key string, min, max int) *TestDataBuilder {
	b.data[key] = randomInt(min, max)
	return b
}

// WithTimestamp sets a field to current timestamp
func (b *TestDataBuilder) WithTimestamp(key string) *TestDataBuilder {
	b.data[key] = time.Now()
	return b
}

// Build builds the test data into the target struct
func (b *TestDataBuilder) Build(target interface{}) error {
	data, err := json.Marshal(b.data)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, target)
}

// BuildMap returns the test data as a map
func (b *TestDataBuilder) BuildMap() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range b.data {
		result[k] = v
	}
	return result
}

// UserFixtures provides user-related test fixtures
type UserFixtures struct {
	manager *FixtureManager
}

// NewUserFixtures creates new user fixtures
func NewUserFixtures(manager *FixtureManager) *UserFixtures {
	return &UserFixtures{manager: manager}
}

// CreateTestUser creates a test user
func (u *UserFixtures) CreateTestUser() *models.User {
	return &models.User{
		ID:        randomString(10),
		Username:  "testuser_" + randomString(5),
		Email:     fmt.Sprintf("test_%s@example.com", randomString(5)),
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateAdminUser creates a test admin user
func (u *UserFixtures) CreateAdminUser() *models.User {
	user := u.CreateTestUser()
	user.Username = "admin_" + randomString(5)
	user.Email = fmt.Sprintf("admin_%s@example.com", randomString(5))
	user.IsAdmin = true
	return user
}

// CreateUserWithPreferences creates a user with preferences
func (u *UserFixtures) CreateUserWithPreferences() (*models.User, *models.UserPreferences) {
	user := u.CreateTestUser()
	prefs := &models.UserPreferences{
		UserID:                user.ID,
		Theme:                 "dark",
		Language:              "en",
		Notifications:         true,
		AutoSave:              true,
		PreferredAIModel:      "llama2",
		VoiceSettings:         map[string]interface{}{"enabled": false},
		DesktopLayout:         map[string]interface{}{"layout": "tiling"},
		AccessibilitySettings: map[string]interface{}{"high_contrast": false},
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}
	return user, prefs
}

// AIFixtures provides AI-related test fixtures
type AIFixtures struct {
	manager *FixtureManager
}

// NewAIFixtures creates new AI fixtures
func NewAIFixtures(manager *FixtureManager) *AIFixtures {
	return &AIFixtures{manager: manager}
}

// CreateTestAIModel creates a test AI model
func (a *AIFixtures) CreateTestAIModel() *models.AIModel {
	return &models.AIModel{
		ID:        randomString(10),
		Name:      "test-model-" + randomString(5),
		Version:   "1.0.0",
		Type:      "llm",
		Provider:  "ollama",
		IsActive:  true,
		IsDefault: false,
		SizeBytes: int64(randomInt(1000000, 10000000)),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestLLMResponse creates a test LLM response
func (a *AIFixtures) CreateTestLLMResponse() *models.LLMResponse {
	return &models.LLMResponse{
		Text:           "This is a test response from the AI model.",
		Confidence:     0.95,
		TokensUsed:     50,
		Model:          "test-model",
		ProcessingTime: time.Millisecond * 100,
		Timestamp:      time.Now(),
	}
}

// CreateTestChatMessage creates a test chat message
func (a *AIFixtures) CreateTestChatMessage(role, content string) *models.ChatMessage {
	return &models.ChatMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// CreateTestObjectDetection creates a test object detection result
func (a *AIFixtures) CreateTestObjectDetection() *models.ObjectDetection {
	return &models.ObjectDetection{
		Objects: []models.DetectedObject{
			{
				Class:      "person",
				Confidence: 0.95,
				Bounds: models.Rectangle{
					X:      100,
					Y:      100,
					Width:  200,
					Height: 300,
				},
			},
			{
				Class:      "laptop",
				Confidence: 0.88,
				Bounds: models.Rectangle{
					X:      300,
					Y:      150,
					Width:  150,
					Height: 100,
				},
			},
		},
		Count:     2,
		Timestamp: time.Now(),
	}
}

// CreateTestSpeechRecognition creates a test speech recognition result
func (a *AIFixtures) CreateTestSpeechRecognition() *models.SpeechRecognition {
	return &models.SpeechRecognition{
		Text:       "Hello, this is a test speech recognition result.",
		Confidence: 0.92,
		Language:   "en",
		Duration:   time.Second * 3,
		Words: []models.WordRecognition{
			{Word: "Hello", Confidence: 0.95, StartTime: 0, EndTime: 500 * time.Millisecond},
			{Word: "this", Confidence: 0.90, StartTime: 500 * time.Millisecond, EndTime: 800 * time.Millisecond},
			{Word: "is", Confidence: 0.88, StartTime: 800 * time.Millisecond, EndTime: 1000 * time.Millisecond},
		},
		Timestamp: time.Now(),
	}
}

// DatabaseFixtures provides database-related test fixtures
type DatabaseFixtures struct {
	manager *FixtureManager
}

// NewDatabaseFixtures creates new database fixtures
func NewDatabaseFixtures(manager *FixtureManager) *DatabaseFixtures {
	return &DatabaseFixtures{manager: manager}
}

// CreateTestDatabase creates test database data
func (d *DatabaseFixtures) CreateTestDatabase() map[string]interface{} {
	return map[string]interface{}{
		"users": []map[string]interface{}{
			{
				"id":         "user1",
				"username":   "testuser1",
				"email":      "test1@example.com",
				"is_active":  true,
				"created_at": time.Now(),
			},
			{
				"id":         "user2",
				"username":   "testuser2",
				"email":      "test2@example.com",
				"is_active":  true,
				"created_at": time.Now(),
			},
		},
		"ai_models": []map[string]interface{}{
			{
				"id":        "model1",
				"name":      "test-llm",
				"type":      "llm",
				"provider":  "ollama",
				"is_active": true,
			},
		},
	}
}

// FixtureLoader provides utilities for loading fixtures in tests
type FixtureLoader struct {
	t       *testing.T
	manager *FixtureManager
}

// NewFixtureLoader creates a new fixture loader for tests
func NewFixtureLoader(t *testing.T, basePath string) *FixtureLoader {
	return &FixtureLoader{
		t:       t,
		manager: NewFixtureManager(basePath),
	}
}

// Load loads a fixture and fails the test if it can't be loaded
func (l *FixtureLoader) Load(name string, target interface{}) {
	err := l.manager.LoadFixture(name, target)
	if err != nil {
		l.t.Fatalf("Failed to load fixture %s: %v", name, err)
	}
}

// MustLoad loads a fixture and panics if it can't be loaded
func (l *FixtureLoader) MustLoad(name string, target interface{}) {
	err := l.manager.LoadFixture(name, target)
	if err != nil {
		panic(fmt.Sprintf("Failed to load fixture %s: %v", name, err))
	}
}

// Helper functions

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

func randomInt(min, max int) int {
	return min + int(time.Now().UnixNano())%(max-min+1)
}
