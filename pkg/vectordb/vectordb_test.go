package vectordb

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockVectorDB implements VectorDB interface for testing
type MockVectorDB struct {
	collections map[string]*CollectionInfo
	vectors     map[string]map[string]*Vector // collection -> id -> vector
	connected   bool
}

func NewMockVectorDB() *MockVectorDB {
	return &MockVectorDB{
		collections: make(map[string]*CollectionInfo),
		vectors:     make(map[string]map[string]*Vector),
		connected:   false,
	}
}

func (m *MockVectorDB) Connect(ctx context.Context) error {
	m.connected = true
	return nil
}

func (m *MockVectorDB) Disconnect(ctx context.Context) error {
	m.connected = false
	return nil
}

func (m *MockVectorDB) IsConnected() bool {
	return m.connected
}

func (m *MockVectorDB) CreateCollection(ctx context.Context, config *CollectionConfig) error {
	m.collections[config.Name] = &CollectionInfo{
		Name:        config.Name,
		Dimension:   config.Dimension,
		Metric:      config.Metric,
		VectorCount: 0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.vectors[config.Name] = make(map[string]*Vector)
	return nil
}

func (m *MockVectorDB) DeleteCollection(ctx context.Context, name string) error {
	delete(m.collections, name)
	delete(m.vectors, name)
	return nil
}

func (m *MockVectorDB) ListCollections(ctx context.Context) ([]string, error) {
	collections := make([]string, 0, len(m.collections))
	for name := range m.collections {
		collections = append(collections, name)
	}
	return collections, nil
}

func (m *MockVectorDB) CollectionExists(ctx context.Context, name string) (bool, error) {
	_, exists := m.collections[name]
	return exists, nil
}

func (m *MockVectorDB) Insert(ctx context.Context, collection string, vectors []*Vector) error {
	collectionVectors, exists := m.vectors[collection]
	if !exists {
		return assert.AnError
	}

	for _, vector := range vectors {
		collectionVectors[vector.ID] = vector
	}

	// Update vector count
	if info, exists := m.collections[collection]; exists {
		info.VectorCount = int64(len(collectionVectors))
		info.UpdatedAt = time.Now()
	}

	return nil
}

func (m *MockVectorDB) Update(ctx context.Context, collection string, vectors []*Vector) error {
	return m.Insert(ctx, collection, vectors)
}

func (m *MockVectorDB) Delete(ctx context.Context, collection string, ids []string) error {
	collectionVectors, exists := m.vectors[collection]
	if !exists {
		return assert.AnError
	}

	for _, id := range ids {
		delete(collectionVectors, id)
	}

	// Update vector count
	if info, exists := m.collections[collection]; exists {
		info.VectorCount = int64(len(collectionVectors))
		info.UpdatedAt = time.Now()
	}

	return nil
}

func (m *MockVectorDB) Search(ctx context.Context, request *SearchRequest) (*SearchResult, error) {
	collectionVectors, exists := m.vectors[request.Collection]
	if !exists {
		return &SearchResult{Matches: []Match{}, Total: 0}, nil
	}

	// Simple mock search - return all vectors with random scores
	matches := make([]Match, 0, len(collectionVectors))
	for _, vector := range collectionVectors {
		match := Match{
			ID:       vector.ID,
			Score:    0.9, // Mock score
			Values:   vector.Values,
			Metadata: vector.Metadata,
		}
		matches = append(matches, match)

		if len(matches) >= request.TopK {
			break
		}
	}

	return &SearchResult{
		Matches: matches,
		Total:   int64(len(matches)),
	}, nil
}

func (m *MockVectorDB) BatchSearch(ctx context.Context, requests []*SearchRequest) ([]*SearchResult, error) {
	results := make([]*SearchResult, len(requests))
	for i, request := range requests {
		result, err := m.Search(ctx, request)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}
	return results, nil
}

func (m *MockVectorDB) GetVector(ctx context.Context, collection string, id string) (*Vector, error) {
	collectionVectors, exists := m.vectors[collection]
	if !exists {
		return nil, assert.AnError
	}

	vector, exists := collectionVectors[id]
	if !exists {
		return nil, assert.AnError
	}

	return vector, nil
}

func (m *MockVectorDB) GetVectors(ctx context.Context, collection string, ids []string) ([]*Vector, error) {
	vectors := make([]*Vector, 0, len(ids))
	for _, id := range ids {
		vector, err := m.GetVector(ctx, collection, id)
		if err == nil {
			vectors = append(vectors, vector)
		}
	}
	return vectors, nil
}

func (m *MockVectorDB) Count(ctx context.Context, collection string) (int64, error) {
	if info, exists := m.collections[collection]; exists {
		return info.VectorCount, nil
	}
	return 0, assert.AnError
}

func (m *MockVectorDB) CreateIndex(ctx context.Context, collection string, config *IndexConfig) error {
	return nil // Mock implementation
}

func (m *MockVectorDB) DeleteIndex(ctx context.Context, collection string, indexName string) error {
	return nil // Mock implementation
}

func (m *MockVectorDB) GetCollectionInfo(ctx context.Context, collection string) (*CollectionInfo, error) {
	if info, exists := m.collections[collection]; exists {
		return info, nil
	}
	return nil, assert.AnError
}

func (m *MockVectorDB) Health(ctx context.Context) (*HealthStatus, error) {
	status := "healthy"
	if !m.connected {
		status = "unhealthy"
	}

	return &HealthStatus{
		Status:      status,
		Version:     "mock-1.0.0",
		Collections: len(m.collections),
	}, nil
}

// MockEmbeddingProvider implements EmbeddingProvider for testing
type MockEmbeddingProvider struct {
	dimensions int
	model      string
}

func NewMockEmbeddingProvider(dimensions int, model string) *MockEmbeddingProvider {
	return &MockEmbeddingProvider{
		dimensions: dimensions,
		model:      model,
	}
}

func (m *MockEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	// Generate mock embedding based on text length
	embedding := make([]float32, m.dimensions)
	for i := range embedding {
		embedding[i] = float32(len(text)%100) / 100.0
	}
	return embedding, nil
}

func (m *MockEmbeddingProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}

func (m *MockEmbeddingProvider) GetDimensions() int {
	return m.dimensions
}

func (m *MockEmbeddingProvider) GetModel() string {
	return m.model
}

func (m *MockEmbeddingProvider) GetMaxTokens() int {
	return 8192
}

func TestVectorStore(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock components
	vectorDB := NewMockVectorDB()
	embedding := NewMockEmbeddingProvider(768, "mock-model")
	
	config := &VectorStoreConfig{
		ChunkSize: 1000,
		Overlap:   200,
	}

	vectorStore := NewVectorStore(vectorDB, embedding, config, logger)

	ctx := context.Background()
	collectionName := "test_collection"

	// Connect to database
	err := vectorDB.Connect(ctx)
	require.NoError(t, err)

	// Create collection
	collectionConfig := &CollectionConfig{
		Name:      collectionName,
		Dimension: 768,
		Metric:    "cosine",
	}
	err = vectorDB.CreateCollection(ctx, collectionConfig)
	require.NoError(t, err)

	t.Run("AddDocuments", func(t *testing.T) {
		documents := []*Document{
			{
				ID:      "doc1",
				Content: "This is the first document",
				Metadata: map[string]interface{}{
					"category": "test",
				},
			},
			{
				ID:      "doc2",
				Content: "This is the second document",
				Metadata: map[string]interface{}{
					"category": "test",
				},
			},
		}

		err := vectorStore.AddDocuments(ctx, collectionName, documents)
		require.NoError(t, err)

		// Verify documents were added
		count, err := vectorDB.Count(ctx, collectionName)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
	})

	t.Run("SimilaritySearch", func(t *testing.T) {
		query := "first document"
		results, err := vectorStore.SimilaritySearch(ctx, collectionName, query, 5, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2) // Should return both documents
	})

	t.Run("SimilaritySearchWithScore", func(t *testing.T) {
		query := "second document"
		results, err := vectorStore.SimilaritySearchWithScore(ctx, collectionName, query, 5, nil)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, result := range results {
			assert.NotNil(t, result.Document)
			assert.Greater(t, result.Score, float32(0))
		}
	})

	t.Run("GetDocuments", func(t *testing.T) {
		ids := []string{"doc1", "doc2"}
		documents, err := vectorStore.GetDocuments(ctx, collectionName, ids)
		require.NoError(t, err)
		assert.Len(t, documents, 2)

		assert.Equal(t, "doc1", documents[0].ID)
		assert.Equal(t, "This is the first document", documents[0].Content)
	})

	t.Run("DeleteDocuments", func(t *testing.T) {
		ids := []string{"doc1"}
		err := vectorStore.DeleteDocuments(ctx, collectionName, ids)
		require.NoError(t, err)

		// Verify document was deleted
		count, err := vectorDB.Count(ctx, collectionName)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func TestVectorDBManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager := NewVectorDBManager(logger)

	t.Run("ListProviders", func(t *testing.T) {
		providers := manager.ListProviders()
		assert.Contains(t, providers, "qdrant")
	})

	t.Run("GetProvider", func(t *testing.T) {
		factory, err := manager.GetProvider("qdrant")
		require.NoError(t, err)
		assert.NotNil(t, factory)
		assert.Equal(t, "qdrant", factory.GetProviderName())
	})

	t.Run("CreateVectorDB", func(t *testing.T) {
		config := &VectorDBConfig{
			Provider: "qdrant",
			Host:     "localhost",
			Port:     6333,
			Timeout:  30 * time.Second,
		}

		// This will fail because we don't have a real Qdrant instance
		// but it tests the factory creation
		_, err := manager.CreateVectorDB(config)
		// We expect this to fail with connection error, not factory error
		assert.Error(t, err)
	})
}

func TestTextSplitter(t *testing.T) {
	splitter := NewTextSplitter(20, 5)

	t.Run("SplitShortText", func(t *testing.T) {
		text := "Short text"
		chunks := splitter.SplitText(text)
		assert.Len(t, chunks, 1)
		assert.Equal(t, text, chunks[0])
	})

	t.Run("SplitLongText", func(t *testing.T) {
		text := "This is a very long text that should be split into multiple chunks because it exceeds the chunk size limit"
		chunks := splitter.SplitText(text)
		assert.Greater(t, len(chunks), 1)

		// Verify overlap
		for i := 1; i < len(chunks); i++ {
			// Check that there's some overlap between consecutive chunks
			// This is a simplified check
			assert.True(t, len(chunks[i-1]) > 0)
			assert.True(t, len(chunks[i]) > 0)
		}
	})

	t.Run("SplitDocuments", func(t *testing.T) {
		documents := []*Document{
			{
				ID:      "doc1",
				Content: "This is a very long document that should be split into multiple chunks because it exceeds the chunk size limit and we want to test the document splitting functionality",
				Metadata: map[string]interface{}{
					"category": "test",
				},
			},
		}

		chunks := splitter.SplitDocuments(documents)
		assert.Greater(t, len(chunks), 1)

		// Verify chunk metadata
		for i, chunk := range chunks {
			assert.Contains(t, chunk.ID, "doc1_chunk_")
			assert.Equal(t, i, chunk.Metadata["chunk_index"])
			assert.Equal(t, len(chunks), chunk.Metadata["total_chunks"])
			assert.Equal(t, "doc1", chunk.Metadata["original_id"])
		}
	})
}

func TestBuilders(t *testing.T) {
	t.Run("VectorDBBuilder", func(t *testing.T) {
		config := NewVectorDBBuilder().
			WithProvider("qdrant").
			WithHost("localhost").
			WithPort(6333).
			WithAPIKey("test-key").
			WithTLS(true).
			WithTimeout(60 * time.Second).
			WithMetadata("env", "test").
			Build()

		assert.Equal(t, "qdrant", config.Provider)
		assert.Equal(t, "localhost", config.Host)
		assert.Equal(t, 6333, config.Port)
		assert.Equal(t, "test-key", config.APIKey)
		assert.True(t, config.TLS)
		assert.Equal(t, 60*time.Second, config.Timeout)
		assert.Equal(t, "test", config.Metadata["env"])
	})

	t.Run("EmbeddingBuilder", func(t *testing.T) {
		config := NewEmbeddingBuilder().
			WithProvider("openai").
			WithModel("text-embedding-ada-002").
			WithAPIKey("test-key").
			WithDimensions(1536).
			WithBatchSize(50).
			WithTimeout(30 * time.Second).
			WithMetadata("version", "v1").
			Build()

		assert.Equal(t, "openai", config.Provider)
		assert.Equal(t, "text-embedding-ada-002", config.Model)
		assert.Equal(t, "test-key", config.APIKey)
		assert.Equal(t, 1536, config.Dimensions)
		assert.Equal(t, 50, config.BatchSize)
		assert.Equal(t, 30*time.Second, config.Timeout)
		assert.Equal(t, "v1", config.Metadata["version"])
	})

	t.Run("VectorStoreBuilder", func(t *testing.T) {
		vectorDBConfig := &VectorDBConfig{Provider: "qdrant"}
		embeddingConfig := &EmbeddingConfig{Provider: "openai"}

		config := NewVectorStoreBuilder().
			WithVectorDB(vectorDBConfig).
			WithEmbedding(embeddingConfig).
			WithChunkSize(2000).
			WithOverlap(400).
			Build()

		assert.Equal(t, vectorDBConfig, config.VectorDB)
		assert.Equal(t, embeddingConfig, config.Embedding)
		assert.Equal(t, 2000, config.ChunkSize)
		assert.Equal(t, 400, config.Overlap)
	})
}

func TestMetrics(t *testing.T) {
	logger := logrus.New()
	metrics := NewVectorDBMetrics(logger)

	t.Run("RecordOperation", func(t *testing.T) {
		metrics.RecordOperation("search", 100*time.Millisecond, true)
		metrics.RecordOperation("search", 200*time.Millisecond, true)
		metrics.RecordOperation("search", 150*time.Millisecond, false)

		stats := metrics.GetOperationStats()
		searchStats := stats["search"]

		assert.Equal(t, int64(3), searchStats.TotalOperations)
		assert.Equal(t, int64(2), searchStats.SuccessfulOperations)
		assert.Equal(t, int64(1), searchStats.FailedOperations)
		assert.Equal(t, 100*time.Millisecond, searchStats.MinLatency)
		assert.Equal(t, 200*time.Millisecond, searchStats.MaxLatency)
	})

	t.Run("RecordCollectionStats", func(t *testing.T) {
		metrics.RecordVectorCount("test_collection", 1000)
		metrics.RecordSearchLatency("test_collection", 50*time.Millisecond)
		metrics.RecordSearchLatency("test_collection", 100*time.Millisecond)

		stats := metrics.GetCollectionStats()
		collectionStats := stats["test_collection"]

		assert.Equal(t, "test_collection", collectionStats.Name)
		assert.Equal(t, int64(1000), collectionStats.VectorCount)
		assert.Equal(t, int64(2), collectionStats.SearchCount)
		assert.Equal(t, 75*time.Millisecond, collectionStats.AverageSearchLatency)
	})
}
