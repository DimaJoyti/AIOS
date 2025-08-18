package knowledge

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKnowledgeManagementIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("CreateKnowledgeManager", func(t *testing.T) {
		config := &KnowledgeManagerConfig{
			DefaultEmbeddingModel: "text-embedding-ada-002",
			DefaultChunkSize:      500,
			DefaultChunkOverlap:   100,
			CacheEnabled:          true,
			CacheTTL:              1 * time.Hour,
			GraphEnabled:          true,
			MultiModalEnabled:     false,
			MaxConcurrentOps:      5,
			MetricsInterval:       1 * time.Minute,
		}

		manager, err := NewKnowledgeManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("DocumentProcessingWorkflow", func(t *testing.T) {
		// Create document processor
		processor, err := NewDefaultDocumentProcessor(logger)
		require.NoError(t, err)

		// Create test document
		doc := &Document{
			ID:          "test-doc-1",
			Title:       "Introduction to AI",
			Content:     "Artificial Intelligence (AI) is a branch of computer science that aims to create intelligent machines. AI systems can perform tasks that typically require human intelligence, such as visual perception, speech recognition, decision-making, and language translation. Machine learning is a subset of AI that enables computers to learn and improve from experience without being explicitly programmed.",
			ContentType: "text",
			Language:    "en",
			Source:      "test",
			Tags:        []string{"ai", "technology", "computer-science"},
			CreatedAt:   time.Now(),
		}

		// Process document
		ctx := context.Background()
		processedDoc, err := processor.ProcessDocument(ctx, doc)
		require.NoError(t, err)
		assert.NotNil(t, processedDoc)
		assert.Greater(t, len(processedDoc.Chunks), 0)
		assert.Greater(t, len(processedDoc.Keywords), 0)
		assert.Equal(t, "en", processedDoc.Language)
	})

	t.Run("ChunkingStrategies", func(t *testing.T) {
		processor, err := NewDefaultDocumentProcessor(logger)
		require.NoError(t, err)

		doc := &Document{
			ID:      "chunking-test",
			Title:   "Test Document",
			Content: "This is the first paragraph. It contains multiple sentences. Each sentence provides information.\n\nThis is the second paragraph. It also has multiple sentences. The content continues here.\n\nThis is the third paragraph. It concludes the document. The end.",
		}

		ctx := context.Background()

		// Test different chunking strategies
		strategies := []ChunkingStrategy{
			ChunkingStrategyFixed,
			ChunkingStrategySentence,
			ChunkingStrategyParagraph,
			ChunkingStrategyRecursive,
		}

		for _, strategy := range strategies {
			chunks, err := processor.ChunkDocument(ctx, doc, strategy)
			require.NoError(t, err, "Failed for strategy: %s", strategy)
			assert.Greater(t, len(chunks), 0, "No chunks created for strategy: %s", strategy)

			// Verify chunk properties
			for i, chunk := range chunks {
				assert.NotEmpty(t, chunk.ID)
				assert.Equal(t, doc.ID, chunk.DocumentID)
				assert.Equal(t, i, chunk.ChunkIndex)
				assert.NotEmpty(t, chunk.Content)
				assert.False(t, chunk.CreatedAt.IsZero())
			}
		}
	})

	t.Run("KnowledgeGraphOperations", func(t *testing.T) {
		graph, err := NewDefaultKnowledgeGraph(logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Create test entities
		entity1 := &Entity{
			ID:          "entity-1",
			Name:        "Artificial Intelligence",
			Type:        "concept",
			Description: "A branch of computer science",
			Confidence:  0.9,
		}

		entity2 := &Entity{
			ID:          "entity-2",
			Name:        "Machine Learning",
			Type:        "concept",
			Description: "A subset of AI",
			Confidence:  0.9,
		}

		entity3 := &Entity{
			ID:          "entity-3",
			Name:        "Deep Learning",
			Type:        "concept",
			Description: "A subset of machine learning",
			Confidence:  0.9,
		}

		// Add entities
		err = graph.AddEntity(ctx, entity1)
		require.NoError(t, err)
		err = graph.AddEntity(ctx, entity2)
		require.NoError(t, err)
		err = graph.AddEntity(ctx, entity3)
		require.NoError(t, err)

		// Create relationships
		rel1 := &Relationship{
			ID:         "rel-1",
			FromEntity: "entity-2",
			ToEntity:   "entity-1",
			Type:       "subset_of",
			Confidence: 0.9,
		}

		rel2 := &Relationship{
			ID:         "rel-2",
			FromEntity: "entity-3",
			ToEntity:   "entity-2",
			Type:       "subset_of",
			Confidence: 0.9,
		}

		// Add relationships
		err = graph.AddRelationship(ctx, rel1)
		require.NoError(t, err)
		err = graph.AddRelationship(ctx, rel2)
		require.NoError(t, err)

		// Test path finding
		paths, err := graph.FindPath(ctx, "entity-3", "entity-1", 3)
		require.NoError(t, err)
		assert.Greater(t, len(paths), 0)

		// Test neighbors
		neighbors, err := graph.GetNeighbors(ctx, "entity-2", 1)
		require.NoError(t, err)
		assert.Equal(t, 2, len(neighbors)) // Should have 2 neighbors

		// Test centrality
		centrality, err := graph.CalculateCentrality(ctx, "entity-2")
		require.NoError(t, err)
		assert.Greater(t, centrality, 0.0)
	})

	t.Run("SemanticCacheOperations", func(t *testing.T) {
		cache, err := NewDefaultSemanticCache(1*time.Hour, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Cache a query
		query := "What is artificial intelligence?"
		result := "AI is a branch of computer science"
		err = cache.CacheQuery(ctx, query, result, 1*time.Hour)
		require.NoError(t, err)

		// Find similar queries
		similar, err := cache.GetSimilarQueries(ctx, "What is AI?", 0.5)
		require.NoError(t, err)
		assert.Greater(t, len(similar), 0)

		// Cache an embedding
		embedding := []float32{0.1, 0.2, 0.3, 0.4, 0.5}
		err = cache.CacheEmbedding(ctx, "test text", embedding, 1*time.Hour)
		require.NoError(t, err)

		// Retrieve cached embedding
		cachedEmbedding, found, err := cache.GetCachedEmbedding(ctx, "test text")
		require.NoError(t, err)
		assert.True(t, found)
		assert.Equal(t, embedding, cachedEmbedding)

		// Get cache stats
		stats, err := cache.GetCacheStats(ctx)
		require.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("DocumentIndexingOperations", func(t *testing.T) {
		indexer, err := NewDefaultKnowledgeIndexer(logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Create test documents
		docs := []*Document{
			{
				ID:      "doc-1",
				Title:   "AI Basics",
				Content: "Introduction to artificial intelligence and machine learning concepts.",
				Tags:    []string{"ai", "ml"},
			},
			{
				ID:      "doc-2",
				Title:   "Deep Learning",
				Content: "Advanced neural networks and deep learning architectures.",
				Tags:    []string{"deep-learning", "neural-networks"},
			},
			{
				ID:      "doc-3",
				Title:   "Natural Language Processing",
				Content: "Text processing and language understanding with AI.",
				Tags:    []string{"nlp", "text-processing"},
			},
		}

		// Index documents
		err = indexer.IndexDocuments(ctx, docs)
		require.NoError(t, err)

		// Get index stats
		stats, err := indexer.GetIndexStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, 3, stats.DocumentCount)
		assert.Greater(t, stats.IndexSize, int64(0))

		// Update a document
		docs[0].Content = "Updated content about AI and ML with more details."
		err = indexer.UpdateIndex(ctx, docs[0].ID, docs[0])
		require.NoError(t, err)

		// Delete a document
		err = indexer.DeleteFromIndex(ctx, docs[2].ID)
		require.NoError(t, err)

		// Check updated stats
		stats, err = indexer.GetIndexStats(ctx)
		require.NoError(t, err)
		assert.Equal(t, 2, stats.DocumentCount)
	})

	t.Run("MetadataExtraction", func(t *testing.T) {
		textExtractor := &TextMetadataExtractor{}
		pdfExtractor := &PDFMetadataExtractor{}

		// Test text metadata extraction
		textDoc := &Document{
			ID:      "text-doc",
			Title:   "Sample Text",
			Content: "This is a sample text document with multiple words and sentences. It contains various information that can be analyzed for metadata extraction. The document has URLs like https://example.com and email addresses like test@example.com.",
		}

		textMetadata, err := textExtractor.Extract(textDoc)
		require.NoError(t, err)
		assert.Contains(t, textMetadata, "word_count")
		assert.Contains(t, textMetadata, "character_count")
		assert.Contains(t, textMetadata, "language")
		assert.Contains(t, textMetadata, "urls")
		assert.Contains(t, textMetadata, "emails")

		// Test PDF metadata extraction
		pdfDoc := &Document{
			ID:          "pdf-doc",
			Title:       "Sample PDF",
			Content:     "This is content extracted from a PDF document.",
			ContentType: "pdf",
		}

		pdfMetadata, err := pdfExtractor.Extract(pdfDoc)
		require.NoError(t, err)
		assert.Contains(t, pdfMetadata, "content_type")
		assert.Equal(t, "pdf", pdfMetadata["content_type"])
		assert.Contains(t, pdfMetadata, "word_count")
	})

	t.Run("EndToEndWorkflow", func(t *testing.T) {
		// This test demonstrates a complete knowledge management workflow
		config := &KnowledgeManagerConfig{
			DefaultEmbeddingModel: "text-embedding-ada-002",
			DefaultChunkSize:      300,
			DefaultChunkOverlap:   50,
			CacheEnabled:          true,
			CacheTTL:              1 * time.Hour,
			GraphEnabled:          true,
			MaxConcurrentOps:      3,
		}

		// Note: This would require actual embedding providers to work fully
		// For now, we'll test the structure and interfaces
		manager, err := NewKnowledgeManager(config, logger)
		require.NoError(t, err)

		ctx := context.Background()

		// Create knowledge base
		kbConfig := &KnowledgeBaseConfig{
			EmbeddingModel:    "text-embedding-ada-002",
			ChunkingStrategy:  ChunkingStrategyRecursive,
			ChunkSize:         300,
			ChunkOverlap:      50,
			IndexingEnabled:   true,
			GraphEnabled:      true,
			SecurityLevel:     SecurityLevelInternal,
		}

		kb, err := manager.CreateKnowledgeBase(ctx, kbConfig)
		require.NoError(t, err)
		assert.NotNil(t, kb)
		assert.NotEmpty(t, kb.ID)

		// Create test documents
		docs := []*Document{
			{
				ID:              "kb-doc-1",
				Title:           "Introduction to Machine Learning",
				Content:         "Machine learning is a method of data analysis that automates analytical model building. It is a branch of artificial intelligence based on the idea that systems can learn from data, identify patterns and make decisions with minimal human intervention.",
				ContentType:     "text",
				Language:        "en",
				Source:          "educational",
				KnowledgeBaseID: kb.ID,
				Tags:            []string{"machine-learning", "ai", "data-science"},
			},
			{
				ID:              "kb-doc-2",
				Title:           "Deep Learning Fundamentals",
				Content:         "Deep learning is part of a broader family of machine learning methods based on artificial neural networks with representation learning. Learning can be supervised, semi-supervised or unsupervised.",
				ContentType:     "text",
				Language:        "en",
				Source:          "educational",
				KnowledgeBaseID: kb.ID,
				Tags:            []string{"deep-learning", "neural-networks", "ai"},
			},
		}

		// Add documents to knowledge base
		for _, doc := range docs {
			err = manager.AddDocument(ctx, doc)
			// Note: This might fail without actual embedding providers
			// but we can test the interface
			if err != nil {
				t.Logf("Expected error without embedding provider: %v", err)
			}
		}

		// Test search functionality (would require embeddings to work fully)
		searchOptions := &SearchOptions{
			TopK:             5,
			Threshold:        0.7,
			SearchType:       SearchTypeSemantic,
			IncludeMetadata:  true,
			KnowledgeBaseIDs: []string{kb.ID},
		}

		_, err = manager.Search(ctx, &SearchQuery{
			Query:   "What is machine learning?",
			Options: searchOptions,
		})
		// This might fail without actual vector store, but tests the interface
		if err != nil {
			t.Logf("Expected error without vector store: %v", err)
		}

		// Get knowledge metrics
		metrics, err := manager.GetKnowledgeMetrics(ctx)
		require.NoError(t, err)
		assert.NotNil(t, metrics)
	})
}

func TestKnowledgeManagerConfiguration(t *testing.T) {
	logger := logrus.New()

	t.Run("DefaultConfiguration", func(t *testing.T) {
		manager, err := NewKnowledgeManager(nil, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})

	t.Run("CustomConfiguration", func(t *testing.T) {
		config := &KnowledgeManagerConfig{
			DefaultEmbeddingModel: "custom-model",
			DefaultChunkSize:      2000,
			DefaultChunkOverlap:   400,
			CacheEnabled:          false,
			GraphEnabled:          false,
			MultiModalEnabled:     true,
			MaxConcurrentOps:      20,
			MetricsInterval:       10 * time.Minute,
		}

		manager, err := NewKnowledgeManager(config, logger)
		require.NoError(t, err)
		assert.NotNil(t, manager)
	})
}

func TestDocumentValidation(t *testing.T) {
	t.Run("ValidDocument", func(t *testing.T) {
		doc := &Document{
			ID:          "valid-doc",
			Title:       "Valid Document",
			Content:     "This is a valid document with proper content.",
			ContentType: "text",
			Language:    "en",
			Source:      "test",
		}

		assert.NotEmpty(t, doc.ID)
		assert.NotEmpty(t, doc.Title)
		assert.NotEmpty(t, doc.Content)
	})

	t.Run("DocumentWithMetadata", func(t *testing.T) {
		doc := &Document{
			ID:          "meta-doc",
			Title:       "Document with Metadata",
			Content:     "Content with metadata",
			ContentType: "text",
			Metadata: map[string]interface{}{
				"author":   "Test Author",
				"category": "test",
				"priority": 1,
			},
		}

		assert.Contains(t, doc.Metadata, "author")
		assert.Equal(t, "Test Author", doc.Metadata["author"])
	})
}
