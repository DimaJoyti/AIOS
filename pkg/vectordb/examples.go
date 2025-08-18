package vectordb

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ExampleBasicUsage demonstrates basic vector database usage
func ExampleBasicUsage() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager
	manager := NewVectorDBManager(logger)

	// Configure vector database
	vectorDBConfig := NewVectorDBBuilder().
		WithProvider("qdrant").
		WithHost("localhost").
		WithPort(6333).
		WithTimeout(30 * time.Second).
		Build()

	// Configure embedding provider
	embeddingConfig := NewEmbeddingBuilder().
		WithProvider("openai").
		WithModel("text-embedding-ada-002").
		WithAPIKey("your-openai-api-key").
		WithDimensions(1536).
		Build()

	// Configure vector store
	storeConfig := NewVectorStoreBuilder().
		WithVectorDB(vectorDBConfig).
		WithEmbedding(embeddingConfig).
		WithChunkSize(1000).
		WithOverlap(200).
		Build()

	// Create vector store
	vectorStore, err := manager.CreateVectorStore(storeConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector store: %w", err)
	}

	ctx := context.Background()
	collectionName := "documents"

	// Create collection (assuming we have access to the underlying VectorDB)
	vectorDB, err := manager.CreateVectorDB(vectorDBConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	if err := vectorDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to vector database: %w", err)
	}
	defer vectorDB.Disconnect(ctx)

	collectionConfig := &CollectionConfig{
		Name:        collectionName,
		Dimension:   1536,
		Metric:      "cosine",
		Description: "Document collection for semantic search",
	}

	if err := vectorDB.CreateCollection(ctx, collectionConfig); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Add documents
	documents := []*Document{
		{
			ID:      "doc1",
			Content: "Artificial Intelligence is transforming the way we work and live.",
			Metadata: map[string]interface{}{
				"category": "technology",
				"author":   "John Doe",
				"date":     "2024-01-15",
			},
		},
		{
			ID:      "doc2",
			Content: "Machine Learning algorithms can identify patterns in large datasets.",
			Metadata: map[string]interface{}{
				"category": "technology",
				"author":   "Jane Smith",
				"date":     "2024-01-16",
			},
		},
		{
			ID:      "doc3",
			Content: "Natural Language Processing enables computers to understand human language.",
			Metadata: map[string]interface{}{
				"category": "technology",
				"author":   "Bob Johnson",
				"date":     "2024-01-17",
			},
		},
	}

	if err := vectorStore.AddDocuments(ctx, collectionName, documents); err != nil {
		return fmt.Errorf("failed to add documents: %w", err)
	}

	logger.Info("Added documents to vector store")

	// Perform similarity search
	query := "How does AI impact our daily lives?"
	results, err := vectorStore.SimilaritySearchWithScore(ctx, collectionName, query, 3, nil)
	if err != nil {
		return fmt.Errorf("failed to perform similarity search: %w", err)
	}

	logger.WithField("query", query).Info("Search results:")
	for i, result := range results {
		logger.WithFields(logrus.Fields{
			"rank":     i + 1,
			"id":       result.Document.ID,
			"score":    result.Score,
			"content":  result.Document.Content,
			"category": result.Document.Metadata["category"],
			"author":   result.Document.Metadata["author"],
		}).Info("Search result")
	}

	return nil
}

// ExampleAdvancedUsage demonstrates advanced vector database features
func ExampleAdvancedUsage() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create manager with metrics
	manager := NewVectorDBManager(logger)
	metrics := NewVectorDBMetrics(logger)

	// Configure with Ollama for local embeddings
	vectorDBConfig := NewVectorDBBuilder().
		WithProvider("qdrant").
		WithHost("localhost").
		WithPort(6333).
		WithTimeout(30*time.Second).
		WithMetadata("environment", "development").
		Build()

	embeddingConfig := NewEmbeddingBuilder().
		WithProvider("ollama").
		WithModel("nomic-embed-text").
		WithBaseURL("http://localhost:11434").
		WithDimensions(768).
		WithTimeout(60 * time.Second).
		Build()

	storeConfig := NewVectorStoreBuilder().
		WithVectorDB(vectorDBConfig).
		WithEmbedding(embeddingConfig).
		WithChunkSize(500).
		WithOverlap(100).
		Build()

	vectorStore, err := manager.CreateVectorStore(storeConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector store: %w", err)
	}

	ctx := context.Background()
	collectionName := "knowledge_base"

	// Create vector database and collection
	vectorDB, err := manager.CreateVectorDB(vectorDBConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector database: %w", err)
	}

	if err := vectorDB.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer vectorDB.Disconnect(ctx)

	collectionConfig := &CollectionConfig{
		Name:      collectionName,
		Dimension: 768,
		Metric:    "cosine",
	}

	if err := vectorDB.CreateCollection(ctx, collectionConfig); err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	// Add large document with automatic chunking
	largeDocument := &Document{
		ID: "large_doc",
		Content: `
		Vector databases are specialized databases designed to store and query high-dimensional vectors efficiently. 
		They are essential for applications involving machine learning, artificial intelligence, and similarity search.
		
		Key features of vector databases include:
		1. High-dimensional vector storage
		2. Similarity search capabilities
		3. Scalable indexing algorithms
		4. Real-time query processing
		5. Integration with machine learning workflows
		
		Popular vector databases include Qdrant, Weaviate, Pinecone, and Chroma. Each has its own strengths and 
		use cases. Qdrant is known for its performance and ease of use, while Weaviate offers strong semantic 
		search capabilities.
		
		Vector databases are commonly used in:
		- Recommendation systems
		- Semantic search
		- Image and video search
		- Natural language processing
		- Anomaly detection
		- Content moderation
		`,
		Metadata: map[string]interface{}{
			"type":     "article",
			"topic":    "vector databases",
			"language": "english",
		},
	}

	// Split document into chunks
	splitter := NewTextSplitter(storeConfig.ChunkSize, storeConfig.Overlap)
	chunks := splitter.SplitDocuments([]*Document{largeDocument})

	logger.WithField("chunks", len(chunks)).Info("Split document into chunks")

	// Add chunks to vector store
	if err := vectorStore.AddDocuments(ctx, collectionName, chunks); err != nil {
		return fmt.Errorf("failed to add document chunks: %w", err)
	}

	// Perform different types of searches
	queries := []string{
		"What are vector databases used for?",
		"Tell me about Qdrant features",
		"How do recommendation systems work?",
	}

	for _, query := range queries {
		start := time.Now()

		// Regular similarity search
		results, err := vectorStore.SimilaritySearch(ctx, collectionName, query, 3, nil)
		if err != nil {
			return fmt.Errorf("failed to perform search: %w", err)
		}

		duration := time.Since(start)
		metrics.RecordSearchLatency(collectionName, duration)

		logger.WithFields(logrus.Fields{
			"query":    query,
			"results":  len(results),
			"duration": duration,
		}).Info("Similarity search completed")

		// MMR search for diverse results
		mmrResults, err := vectorStore.MaxMarginalRelevanceSearch(
			ctx, collectionName, query, 3, 6, 0.7, nil,
		)
		if err != nil {
			return fmt.Errorf("failed to perform MMR search: %w", err)
		}

		logger.WithFields(logrus.Fields{
			"query":       query,
			"mmr_results": len(mmrResults),
		}).Info("MMR search completed")
	}

	// Filtered search
	filter := map[string]interface{}{
		"type": "article",
	}

	filteredResults, err := vectorStore.SimilaritySearch(
		ctx, collectionName, "vector database applications", 5, filter,
	)
	if err != nil {
		return fmt.Errorf("failed to perform filtered search: %w", err)
	}

	logger.WithField("filtered_results", len(filteredResults)).Info("Filtered search completed")

	// Get collection statistics
	count, err := vectorDB.Count(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get count: %w", err)
	}

	info, err := vectorDB.GetCollectionInfo(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to get collection info: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"collection":   info.Name,
		"vector_count": count,
		"dimension":    info.Dimension,
		"metric":       info.Metric,
		"created_at":   info.CreatedAt,
	}).Info("Collection statistics")

	// Print metrics
	operationStats := metrics.GetOperationStats()
	collectionStats := metrics.GetCollectionStats()

	logger.Info("Operation statistics:")
	for operation, stats := range operationStats {
		logger.WithFields(logrus.Fields{
			"operation":   operation,
			"total":       stats.TotalOperations,
			"successful":  stats.SuccessfulOperations,
			"failed":      stats.FailedOperations,
			"avg_latency": stats.AverageLatency,
		}).Info("Operation stats")
	}

	logger.Info("Collection statistics:")
	for collection, stats := range collectionStats {
		logger.WithFields(logrus.Fields{
			"collection":         collection,
			"vector_count":       stats.VectorCount,
			"search_count":       stats.SearchCount,
			"avg_search_latency": stats.AverageSearchLatency,
			"last_accessed":      stats.LastAccessed,
		}).Info("Collection stats")
	}

	return nil
}

// ExampleBatchOperations demonstrates batch operations
func ExampleBatchOperations() error {
	logger := logrus.New()

	// Setup (similar to previous examples)
	manager := NewVectorDBManager(logger)

	vectorDBConfig := NewVectorDBBuilder().
		WithProvider("qdrant").
		WithHost("localhost").
		WithPort(6333).
		Build()

	embeddingConfig := NewEmbeddingBuilder().
		WithProvider("openai").
		WithModel("text-embedding-ada-002").
		WithAPIKey("your-api-key").
		Build()

	storeConfig := NewVectorStoreBuilder().
		WithVectorDB(vectorDBConfig).
		WithEmbedding(embeddingConfig).
		Build()

	vectorStore, err := manager.CreateVectorStore(storeConfig)
	if err != nil {
		return err
	}

	ctx := context.Background()
	collectionName := "batch_test"

	// Create collection
	vectorDB, err := manager.CreateVectorDB(vectorDBConfig)
	if err != nil {
		return err
	}

	if err := vectorDB.Connect(ctx); err != nil {
		return err
	}
	defer vectorDB.Disconnect(ctx)

	collectionConfig := &CollectionConfig{
		Name:      collectionName,
		Dimension: 1536,
		Metric:    "cosine",
	}

	if err := vectorDB.CreateCollection(ctx, collectionConfig); err != nil {
		return err
	}

	// Batch add texts
	texts := []string{
		"The quick brown fox jumps over the lazy dog",
		"Machine learning is a subset of artificial intelligence",
		"Python is a popular programming language for data science",
		"Vector databases enable semantic search capabilities",
		"Natural language processing helps computers understand text",
	}

	metadatas := make([]map[string]interface{}, len(texts))
	for i := range metadatas {
		metadatas[i] = map[string]interface{}{
			"index":    i,
			"category": "example",
		}
	}

	if err := vectorStore.AddTexts(ctx, collectionName, texts, metadatas); err != nil {
		return fmt.Errorf("failed to add texts: %w", err)
	}

	logger.WithField("count", len(texts)).Info("Added texts in batch")

	// Batch search
	queries := []string{
		"animal behavior",
		"artificial intelligence",
		"programming languages",
	}

	searchRequests := make([]*SearchRequest, len(queries))

	// Generate embeddings for queries
	embeddingProvider, err := NewOpenAIEmbeddingFactory().Create(embeddingConfig)
	if err != nil {
		return err
	}

	queryEmbeddings, err := embeddingProvider.GenerateEmbeddings(ctx, queries)
	if err != nil {
		return err
	}

	for i, embedding := range queryEmbeddings {
		searchRequests[i] = &SearchRequest{
			Collection: collectionName,
			Vector:     embedding,
			TopK:       3,
			Include:    []string{"metadata"},
		}
	}

	// Perform batch search
	results, err := vectorDB.BatchSearch(ctx, searchRequests)
	if err != nil {
		return fmt.Errorf("failed to perform batch search: %w", err)
	}

	logger.Info("Batch search results:")
	for i, result := range results {
		logger.WithFields(logrus.Fields{
			"query":   queries[i],
			"matches": len(result.Matches),
		}).Info("Query result")

		for j, match := range result.Matches {
			logger.WithFields(logrus.Fields{
				"rank":  j + 1,
				"id":    match.ID,
				"score": match.Score,
			}).Debug("Match")
		}
	}

	return nil
}

// DemoVectorDatabase demonstrates all vector database capabilities
func DemoVectorDatabase() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("🚀 Starting Vector Database Demo")

	// Note: These examples require actual database connections
	// In a real demo, you would uncomment and run these:

	/*
		// Run basic usage example
		logger.Info("📝 Running basic usage example...")
		if err := ExampleBasicUsage(); err != nil {
			logger.WithError(err).Error("Basic usage example failed")
			return err
		}
		logger.Info("✅ Basic usage example completed")

		// Run advanced usage example
		logger.Info("🔧 Running advanced usage example...")
		if err := ExampleAdvancedUsage(); err != nil {
			logger.WithError(err).Error("Advanced usage example failed")
			return err
		}
		logger.Info("✅ Advanced usage example completed")

		// Run batch operations example
		logger.Info("📦 Running batch operations example...")
		if err := ExampleBatchOperations(); err != nil {
			logger.WithError(err).Error("Batch operations example failed")
			return err
		}
		logger.Info("✅ Batch operations example completed")
	*/

	logger.Info("📚 Vector Database system is ready for use!")
	logger.Info("🔧 Configure your vector database and embedding providers")
	logger.Info("📊 Use the examples above as templates for your implementation")

	logger.Info("🎉 Vector Database Demo completed successfully!")

	return nil
}
