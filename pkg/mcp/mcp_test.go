package mcp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aios/aios/pkg/mcp/client"
	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/aios/aios/pkg/mcp/resources"
	"github.com/aios/aios/pkg/mcp/server"
	"github.com/aios/aios/pkg/mcp/tools"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPProtocolMessages(t *testing.T) {
	t.Run("BaseMessage", func(t *testing.T) {
		msg := &protocol.BaseMessage{
			ID:        "test-123",
			Type:      protocol.MessageTypeRequest,
			Method:    "test_method",
			Timestamp: time.Now(),
		}

		assert.Equal(t, "test-123", msg.GetID())
		assert.Equal(t, protocol.MessageTypeRequest, msg.GetType())
		assert.Equal(t, "test_method", msg.GetMethod())
	})

	t.Run("MCPError", func(t *testing.T) {
		err := &protocol.MCPError{
			Code:    protocol.ErrorCodeInvalidRequest,
			Message: "Invalid request",
		}

		assert.Equal(t, protocol.ErrorCodeInvalidRequest, err.Code)
		assert.Equal(t, "Invalid request", err.Message)
		assert.Equal(t, "Invalid request", err.Error())
	})

	t.Run("ResourceContent", func(t *testing.T) {
		content := &protocol.ResourceContent{
			URI:      "file:///test.txt",
			MimeType: "text/plain",
			Text:     "Hello, World!",
		}

		assert.Equal(t, "file:///test.txt", content.URI)
		assert.Equal(t, "text/plain", content.MimeType)
		assert.Equal(t, "Hello, World!", content.Text)
	})
}

func TestResourceManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := &resources.ResourceManagerConfig{
		EnableCache:    true,
		EnableWatcher:  false,
		EnableMetrics:  true,
		MaxResources:   1000,
		AllowedSchemes: []string{"file", "http", "https", "test"},
	}

	t.Run("CreateResourceManager", func(t *testing.T) {
		manager, err := resources.NewResourceManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("RegisterAndGetResource", func(t *testing.T) {
		manager, err := resources.NewResourceManager(config, logger)
		require.NoError(t, err)

		resource := &TestResource{
			uri:         "test://example",
			name:        "Test Resource",
			description: "A test resource",
			mimeType:    "text/plain",
			content:     "Test content",
		}

		err = manager.RegisterResource(resource)
		require.NoError(t, err)

		retrieved, err := manager.GetResource("test://example")
		require.NoError(t, err)
		assert.Equal(t, resource.GetURI(), retrieved.GetURI())
		assert.Equal(t, resource.GetName(), retrieved.GetName())
	})

	t.Run("ListResources", func(t *testing.T) {
		manager, err := resources.NewResourceManager(config, logger)
		require.NoError(t, err)

		// Register multiple resources
		for i := 0; i < 3; i++ {
			resource := &TestResource{
				uri:  fmt.Sprintf("test://example%d", i),
				name: fmt.Sprintf("Test Resource %d", i),
			}
			err = manager.RegisterResource(resource)
			require.NoError(t, err)
		}

		resourceList, err := manager.ListResources("", 10)
		require.NoError(t, err)
		assert.Len(t, resourceList.Resources, 3)
	})

	t.Run("ReadResourceContent", func(t *testing.T) {
		manager, err := resources.NewResourceManager(config, logger)
		require.NoError(t, err)

		// Register a resource with content
		resource := &TestResource{
			uri:      "test://content",
			name:     "Content Resource",
			mimeType: "text/plain",
			content:  "Test content",
		}
		err = manager.RegisterResource(resource)
		require.NoError(t, err)

		// Read content
		readContent, err := manager.ReadResource(context.Background(), "test://content")
		require.NoError(t, err)
		assert.Len(t, readContent.Contents, 1)
		assert.Equal(t, "Test content", readContent.Contents[0].Text)
	})
}

func TestResourceCache(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateCache", func(t *testing.T) {
		cache, err := resources.NewMemoryResourceCache(logger)
		require.NoError(t, err)
		assert.NotNil(t, cache)
	})

	t.Run("SetAndGetContent", func(t *testing.T) {
		cache, err := resources.NewMemoryResourceCache(logger)
		require.NoError(t, err)

		content := []protocol.ResourceContent{
			{
				URI:      "test://cache",
				MimeType: "text/plain",
				Text:     "Cached content",
			},
		}

		cache.Set("test://cache", content, 5*time.Minute)

		retrieved, found := cache.Get("test://cache")
		assert.True(t, found)
		assert.Len(t, retrieved, 1)
		assert.Equal(t, "Cached content", retrieved[0].Text)
	})

	t.Run("CacheExpiration", func(t *testing.T) {
		cache, err := resources.NewMemoryResourceCache(logger)
		require.NoError(t, err)

		content := []protocol.ResourceContent{
			{
				URI:  "test://expire",
				Text: "Expiring content",
			},
		}

		// Set with very short TTL
		cache.Set("test://expire", content, 1*time.Millisecond)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		_, found := cache.Get("test://expire")
		assert.False(t, found)
	})

	t.Run("CacheStats", func(t *testing.T) {
		cache, err := resources.NewMemoryResourceCache(logger)
		require.NoError(t, err)

		content := []protocol.ResourceContent{
			{URI: "test://stats", Text: "Stats content"},
		}

		cache.Set("test://stats", content, 5*time.Minute)
		cache.Get("test://stats")   // Hit
		cache.Get("test://missing") // Miss

		stats := cache.GetStats()
		assert.NotNil(t, stats)
		// Note: Specific stats depend on implementation
	})
}

func TestToolManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateToolManager", func(t *testing.T) {
		manager := tools.NewToolManager(logger)
		assert.NotNil(t, manager)
	})

	t.Run("RegisterAndListTools", func(t *testing.T) {
		manager := tools.NewToolManager(logger)

		// Register filesystem tools
		fsTools, err := tools.NewFileSystemTools(logger)
		require.NoError(t, err)

		err = manager.RegisterToolProvider(fsTools)
		require.NoError(t, err)

		toolList := manager.ListTools()
		assert.Greater(t, len(toolList), 0)
	})

	t.Run("CallTool", func(t *testing.T) {
		manager := tools.NewToolManager(logger)

		// Register filesystem tools
		fsTools, err := tools.NewFileSystemTools(logger)
		require.NoError(t, err)

		err = manager.RegisterToolProvider(fsTools)
		require.NoError(t, err)

		// Call list_directory tool
		result, err := manager.CallTool(context.Background(), "list_directory", map[string]interface{}{
			"path": "/tmp",
		})

		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestMCPServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateServer", func(t *testing.T) {
		config := &server.ServerConfig{
			Address: "localhost",
			Port:    8080,
			Metadata: map[string]interface{}{
				"name":        "test-server",
				"version":     "1.0.0",
				"description": "Test MCP Server",
			},
		}

		srv, err := server.NewMCPServer(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, srv)
	})

	t.Run("ServerCapabilities", func(t *testing.T) {
		config := &server.ServerConfig{
			Address: "localhost",
			Port:    8080,
			Metadata: map[string]interface{}{
				"name":        "test-server",
				"version":     "1.0.0",
				"description": "Test MCP Server",
			},
		}

		srv, err := server.NewMCPServer(config, logger)
		require.NoError(t, err)

		serverConfig := srv.GetConfig()
		assert.NotNil(t, serverConfig)
	})
}

func TestMCPClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateClient", func(t *testing.T) {
		config := &client.ClientConfig{
			ServerAddress: "localhost",
			ServerPort:    8080,
			ClientInfo: protocol.ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			ConnectTimeout: 30 * time.Second,
			RequestTimeout: 10 * time.Second,
			Metadata: map[string]interface{}{
				"description": "Test MCP Client",
			},
		}

		cli, err := client.NewMCPClient(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, cli)
	})

	t.Run("ClientCapabilities", func(t *testing.T) {
		config := &client.ClientConfig{
			ServerAddress: "localhost",
			ServerPort:    8080,
			ClientInfo: protocol.ClientInfo{
				Name:    "test-client",
				Version: "1.0.0",
			},
			ConnectTimeout: 30 * time.Second,
			RequestTimeout: 10 * time.Second,
			Metadata: map[string]interface{}{
				"description": "Test MCP Client",
			},
		}

		cli, err := client.NewMCPClient(config, logger)
		require.NoError(t, err)

		// Test basic client functionality
		connected := cli.IsConnected()
		assert.False(t, connected) // Should be false since we haven't connected yet
	})
}

func TestResourceValidator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateValidator", func(t *testing.T) {
		validator, err := resources.NewResourceValidator([]string{"file", "http", "https"}, logger)
		require.NoError(t, err)
		assert.NotNil(t, validator)
	})

	t.Run("ValidateFileURI", func(t *testing.T) {
		validator, err := resources.NewResourceValidator([]string{"file"}, logger)
		require.NoError(t, err)

		err = validator.ValidateURI("file:///tmp/test.txt")
		assert.NoError(t, err)
	})

	t.Run("ValidateHTTPURI", func(t *testing.T) {
		validator, err := resources.NewResourceValidator([]string{"https"}, logger)
		require.NoError(t, err)

		err = validator.ValidateURI("https://example.com/resource")
		assert.NoError(t, err)
	})

	t.Run("RejectInvalidScheme", func(t *testing.T) {
		validator, err := resources.NewResourceValidator([]string{"file"}, logger)
		require.NoError(t, err)

		err = validator.ValidateURI("ftp://example.com/file")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not allowed")
	})

	t.Run("RejectPathTraversal", func(t *testing.T) {
		validator, err := resources.NewResourceValidator([]string{"file"}, logger)
		require.NoError(t, err)

		err = validator.ValidateURI("file:///tmp/../etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
	})
}

func TestResourceWatcher(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateWatcher", func(t *testing.T) {
		watcher, err := resources.NewResourceWatcher(logger)
		require.NoError(t, err)
		assert.NotNil(t, watcher)
	})

	t.Run("WatchResource", func(t *testing.T) {
		watcher, err := resources.NewResourceWatcher(logger)
		require.NoError(t, err)

		err = watcher.Watch("file:///tmp/test.txt")
		assert.NoError(t, err)

		assert.True(t, watcher.IsWatching("file:///tmp/test.txt"))
	})

	t.Run("UnwatchResource", func(t *testing.T) {
		watcher, err := resources.NewResourceWatcher(logger)
		require.NoError(t, err)

		err = watcher.Watch("file:///tmp/test.txt")
		require.NoError(t, err)

		err = watcher.Unwatch("file:///tmp/test.txt")
		assert.NoError(t, err)

		assert.False(t, watcher.IsWatching("file:///tmp/test.txt"))
	})
}

func TestResourceMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateMetrics", func(t *testing.T) {
		metrics, err := resources.NewResourceMetrics(logger)
		require.NoError(t, err)
		assert.NotNil(t, metrics)
	})

	t.Run("RecordAccess", func(t *testing.T) {
		metrics, err := resources.NewResourceMetrics(logger)
		require.NoError(t, err)

		metrics.RecordResourceAccess("test://resource", "read", 100*time.Millisecond, true)

		concreteMetrics := metrics.(*resources.DefaultResourceMetrics)
		snapshot := concreteMetrics.GetMetrics()
		assert.Equal(t, int64(1), snapshot.TotalAccesses)
		assert.Equal(t, int64(1), snapshot.SuccessfulAccesses)
		assert.Equal(t, int64(0), snapshot.FailedAccesses)
	})

	t.Run("RecordCacheHit", func(t *testing.T) {
		metrics, err := resources.NewResourceMetrics(logger)
		require.NoError(t, err)

		metrics.RecordCacheHit("test://resource", true)
		metrics.RecordCacheHit("test://resource", false)

		concreteMetrics := metrics.(*resources.DefaultResourceMetrics)
		snapshot := concreteMetrics.GetMetrics()
		assert.Equal(t, int64(1), snapshot.CacheHits)
		assert.Equal(t, int64(1), snapshot.CacheMisses)
		assert.Equal(t, 0.5, snapshot.CacheHitRate)
	})

	t.Run("RecordError", func(t *testing.T) {
		metrics, err := resources.NewResourceMetrics(logger)
		require.NoError(t, err)

		testErr := fmt.Errorf("test error")
		concreteMetrics := metrics.(*resources.DefaultResourceMetrics)
		concreteMetrics.RecordError("test://resource", "read", testErr)

		snapshot := concreteMetrics.GetMetrics()
		assert.Contains(t, snapshot.ErrorCounts, "test error")
		assert.Equal(t, int64(1), snapshot.ErrorsByURI["test://resource"])
	})
}

// TestResource implements MCPResource for testing
type TestResource struct {
	uri         string
	name        string
	description string
	mimeType    string
	content     string
	lastMod     time.Time
	size        int64
	annotations map[string]interface{}
	metadata    map[string]interface{}
	tags        []string
	category    string
}

func (r *TestResource) GetURI() string         { return r.uri }
func (r *TestResource) GetName() string        { return r.name }
func (r *TestResource) GetDescription() string { return r.description }
func (r *TestResource) GetMimeType() string    { return r.mimeType }
func (r *TestResource) GetAnnotations() map[string]interface{} {
	if r.annotations == nil {
		return make(map[string]interface{})
	}
	return r.annotations
}
func (r *TestResource) GetLastModified() time.Time { return r.lastMod }
func (r *TestResource) GetSize() int64             { return r.size }
func (r *TestResource) IsWatchable() bool          { return false }
func (r *TestResource) GetCategory() string        { return r.category }
func (r *TestResource) GetTags() []string          { return r.tags }
func (r *TestResource) GetMetadata() map[string]interface{} {
	if r.metadata == nil {
		return make(map[string]interface{})
	}
	return r.metadata
}

func (r *TestResource) ReadContent(ctx context.Context) ([]protocol.ResourceContent, error) {
	return []protocol.ResourceContent{
		{
			URI:      r.uri,
			MimeType: r.mimeType,
			Text:     r.content,
		},
	}, nil
}

func (r *TestResource) Watch(ctx context.Context, callback resources.ResourceCallback) error {
	return nil
}

func (r *TestResource) StopWatch() error {
	return nil
}

func (r *TestResource) Validate() error {
	if r.uri == "" {
		return fmt.Errorf("URI cannot be empty")
	}
	return nil
}
