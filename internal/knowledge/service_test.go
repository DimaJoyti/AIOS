package knowledge

import (
	"testing"

	"github.com/aios/aios/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeService(t *testing.T) {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	cfg := &config.Config{}

	t.Run("CreateService", func(t *testing.T) {
		// For testing, we'll skip the database connection
		// In a real test, you'd use a test database
		t.Skip("Skipping database-dependent test")

		// service, err := NewService(cfg, db, logger)
		// require.NoError(t, err)
		// assert.NotNil(t, service)
	})

	t.Run("DocumentProcessing", func(t *testing.T) {
		// Test document processing logic
		processor, err := NewDocumentProcessor(cfg, nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, processor)
	})

	t.Run("WebCrawling", func(t *testing.T) {
		// Test web crawler
		crawler, err := NewWebCrawler(cfg, nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, crawler)
	})

	t.Run("VectorSearch", func(t *testing.T) {
		// Test vector searcher
		searcher, err := NewVectorSearcher(cfg, nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, searcher)
	})
}

func TestDocumentChunking(t *testing.T) {
	chunker := &TextChunker{
		maxChunkSize:      1000,
		overlapSize:       200,
		preserveSentences: true,
	}

	t.Run("ChunkText", func(t *testing.T) {
		text := "This is a test document. It has multiple sentences. Each sentence should be preserved when chunking. The chunker should create overlapping chunks for better context preservation."

		chunks, err := chunker.ChunkText(text)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 0)

		// Verify chunks are not empty
		for _, chunk := range chunks {
			assert.NotEmpty(t, chunk)
		}
	})

	t.Run("ChunkLongText", func(t *testing.T) {
		// Create a long text that will require multiple chunks
		longText := ""
		for i := 0; i < 100; i++ {
			longText += "This is sentence number " + string(rune(i)) + ". It contains some meaningful content that should be chunked properly. "
		}

		chunks, err := chunker.ChunkText(longText)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1, "Long text should create multiple chunks")

		// Verify chunk sizes
		for _, chunk := range chunks {
			assert.LessOrEqual(t, len(chunk), chunker.maxChunkSize+chunker.overlapSize)
		}
	})
}

func TestSearchRequest(t *testing.T) {
	t.Run("CreateSearchRequest", func(t *testing.T) {
		req := SearchRequest{
			Query:       "test query",
			MaxResults:  10,
			UseRAG:      true,
			ContextSize: 5,
		}

		assert.Equal(t, "test query", req.Query)
		assert.Equal(t, 10, req.MaxResults)
		assert.True(t, req.UseRAG)
		assert.Equal(t, 5, req.ContextSize)
	})
}

func TestCrawlRequest(t *testing.T) {
	t.Run("CreateCrawlRequest", func(t *testing.T) {
		req := CrawlRequest{
			URL:         "https://example.com",
			MaxPages:    100,
			MaxDepth:    3,
			FollowLinks: true,
			Metadata:    map[string]string{"source": "test"},
		}

		assert.Equal(t, "https://example.com", req.URL)
		assert.Equal(t, 100, req.MaxPages)
		assert.Equal(t, 3, req.MaxDepth)
		assert.True(t, req.FollowLinks)
		assert.Equal(t, "test", req.Metadata["source"])
	})
}

func TestDocumentUploadRequest(t *testing.T) {
	t.Run("CreateDocumentUploadRequest", func(t *testing.T) {
		req := DocumentUploadRequest{
			FileName: "test.txt",
			Content:  "This is test content",
			MimeType: "text/plain",
			Metadata: map[string]string{"author": "test"},
		}

		assert.Equal(t, "test.txt", req.FileName)
		assert.Equal(t, "This is test content", req.Content)
		assert.Equal(t, "text/plain", req.MimeType)
		assert.Equal(t, "test", req.Metadata["author"])
	})
}
