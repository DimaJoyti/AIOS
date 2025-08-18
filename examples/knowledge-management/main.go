package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/knowledge"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	fmt.Println("🧠 AIOS Knowledge Management System Demo")
	fmt.Println("========================================")

	// Run the comprehensive demo
	if err := runKnowledgeManagementDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Knowledge Management Demo completed successfully!")
}

func runKnowledgeManagementDemo(logger *logrus.Logger) error {
	ctx := context.Background()

	// Step 1: Create Knowledge Manager
	fmt.Println("\n1. Creating Knowledge Manager...")

	// Note: This demo uses a mock configuration for demonstration purposes
	// In a real environment, you would configure with actual embedding providers
	fmt.Println("   Note: Using mock configuration for demonstration")
	fmt.Println("   In production, configure with actual embedding providers (OpenAI, Ollama, etc.)")

	config := &knowledge.KnowledgeManagerConfig{
		DefaultEmbeddingModel: "mock-embedding-model",
		DefaultChunkSize:      1000,
		DefaultChunkOverlap:   200,
		CacheEnabled:          true,
		CacheTTL:              24 * time.Hour,
		GraphEnabled:          true,
		MultiModalEnabled:     false,
		MaxConcurrentOps:      10,
		MetricsInterval:       5 * time.Minute,
	}

	manager, err := knowledge.NewKnowledgeManager(config, logger)
	if err != nil {
		// For demo purposes, show what would happen if properly configured
		fmt.Printf("   ⚠️  Knowledge Manager creation failed (expected in demo): %v\n", err)
		fmt.Println("   📝 This is expected as the demo doesn't have embedding provider credentials")
		fmt.Println("   🔧 To use in production:")
		fmt.Println("      - Set OPENAI_API_KEY environment variable for OpenAI embeddings")
		fmt.Println("      - Or configure Ollama for local embeddings")
		fmt.Println("      - Or use other supported embedding providers")

		// Show what the demo would do if it worked
		return runMockKnowledgeDemo()
	}
	fmt.Println("✓ Knowledge Manager created successfully")

	// Step 2: Create Knowledge Base
	fmt.Println("\n2. Creating Knowledge Base...")
	kbConfig := &knowledge.KnowledgeBaseConfig{
		EmbeddingModel:    "text-embedding-ada-002",
		ChunkingStrategy:  knowledge.ChunkingStrategyRecursive,
		ChunkSize:         1000,
		ChunkOverlap:      200,
		IndexingEnabled:   true,
		GraphEnabled:      true,
		MultiModalEnabled: false,
		VersioningEnabled: true,
		SecurityLevel:     knowledge.SecurityLevelInternal,
	}

	kb, err := manager.CreateKnowledgeBase(ctx, kbConfig)
	if err != nil {
		return fmt.Errorf("failed to create knowledge base: %w", err)
	}
	fmt.Printf("✓ Knowledge Base created: %s\n", kb.ID)

	// Step 3: Add Documents
	fmt.Println("\n3. Adding Documents to Knowledge Base...")
	documents := createSampleDocuments(kb.ID)

	for i, doc := range documents {
		fmt.Printf("   Adding document %d: %s\n", i+1, doc.Title)
		if err := manager.AddDocument(ctx, doc); err != nil {
			// Note: This might fail without actual embedding providers
			fmt.Printf("   ⚠️  Document addition failed (expected without embedding provider): %v\n", err)
		} else {
			fmt.Printf("   ✓ Document added successfully\n")
		}
	}

	// Step 4: Demonstrate Document Processing
	fmt.Println("\n4. Demonstrating Document Processing...")
	processor, err := knowledge.NewDefaultDocumentProcessor(logger)
	if err != nil {
		return fmt.Errorf("failed to create document processor: %w", err)
	}

	sampleDoc := documents[0]
	processedDoc, err := processor.ProcessDocument(ctx, sampleDoc)
	if err != nil {
		return fmt.Errorf("failed to process document: %w", err)
	}

	fmt.Printf("   ✓ Document processed successfully:\n")
	fmt.Printf("     - Chunks: %d\n", len(processedDoc.Chunks))
	fmt.Printf("     - Keywords: %d\n", len(processedDoc.Keywords))
	fmt.Printf("     - Entities: %d\n", len(processedDoc.Entities))
	fmt.Printf("     - Language: %s\n", processedDoc.Language)
	if processedDoc.Sentiment != nil {
		fmt.Printf("     - Sentiment: %s (%.2f)\n", processedDoc.Sentiment.Label, processedDoc.Sentiment.Score)
	}

	// Step 5: Demonstrate Chunking Strategies
	fmt.Println("\n5. Demonstrating Different Chunking Strategies...")
	strategies := []knowledge.ChunkingStrategy{
		knowledge.ChunkingStrategyFixed,
		knowledge.ChunkingStrategySentence,
		knowledge.ChunkingStrategyParagraph,
		knowledge.ChunkingStrategyRecursive,
	}

	for _, strategy := range strategies {
		chunks, err := processor.ChunkDocument(ctx, sampleDoc, strategy)
		if err != nil {
			fmt.Printf("   ⚠️  %s chunking failed: %v\n", strategy, err)
			continue
		}
		fmt.Printf("   ✓ %s chunking: %d chunks\n", strategy, len(chunks))
	}

	// Step 6: Demonstrate Knowledge Graph
	fmt.Println("\n6. Building Knowledge Graph...")
	graph, err := knowledge.NewDefaultKnowledgeGraph(logger)
	if err != nil {
		return fmt.Errorf("failed to create knowledge graph: %w", err)
	}

	// Create entities and relationships
	entities, relationships := createSampleKnowledgeGraph()

	for _, entity := range entities {
		if err := graph.AddEntity(ctx, entity); err != nil {
			fmt.Printf("   ⚠️  Failed to add entity %s: %v\n", entity.Name, err)
		} else {
			fmt.Printf("   ✓ Added entity: %s\n", entity.Name)
		}
	}

	for _, rel := range relationships {
		if err := graph.AddRelationship(ctx, rel); err != nil {
			fmt.Printf("   ⚠️  Failed to add relationship: %v\n", err)
		} else {
			fmt.Printf("   ✓ Added relationship: %s -> %s\n", rel.FromEntity, rel.ToEntity)
		}
	}

	// Demonstrate path finding
	fmt.Println("\n   Finding paths in knowledge graph...")
	paths, err := graph.FindPath(ctx, "ai", "neural-networks", 3)
	if err != nil {
		fmt.Printf("   ⚠️  Path finding failed: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d paths between AI and Neural Networks\n", len(paths))
	}

	// Step 7: Demonstrate Search (would require embeddings)
	fmt.Println("\n7. Demonstrating Search Capabilities...")
	searchOptions := &knowledge.SearchOptions{
		TopK:             5,
		Threshold:        0.7,
		SearchType:       knowledge.SearchTypeSemantic,
		IncludeMetadata:  true,
		KnowledgeBaseIDs: []string{kb.ID},
	}

	searchQuery := &knowledge.SearchQuery{
		Query:   "What is artificial intelligence?",
		Options: searchOptions,
	}

	searchResult, err := manager.Search(ctx, searchQuery)
	if err != nil {
		fmt.Printf("   ⚠️  Search failed (expected without vector store): %v\n", err)
	} else {
		fmt.Printf("   ✓ Search completed: %d results\n", len(searchResult.Documents))
	}

	// Step 8: Demonstrate RAG Pipeline (would require LLM)
	fmt.Println("\n8. Demonstrating RAG Pipeline...")
	ragOptions := &knowledge.RAGOptions{
		RetrievalOptions: &knowledge.RetrievalOptions{
			TopK:      5,
			Threshold: 0.7,
		},
		GenerationOptions: &knowledge.GenerationOptions{
			Model:       "gpt-4",
			Temperature: 0.7,
			MaxTokens:   500,
		},
		ContextLength:  2000,
		IncludeSources: true,
	}

	ragResponse, err := manager.RetrieveAndGenerate(ctx, "Explain machine learning", ragOptions)
	if err != nil {
		fmt.Printf("   ⚠️  RAG pipeline failed (expected without LLM): %v\n", err)
	} else {
		fmt.Printf("   ✓ RAG response generated: %d characters\n", len(ragResponse.Response))
	}

	// Step 9: Demonstrate Caching
	fmt.Println("\n9. Demonstrating Semantic Caching...")
	cache, err := knowledge.NewDefaultSemanticCache(1*time.Hour, logger)
	if err != nil {
		return fmt.Errorf("failed to create semantic cache: %w", err)
	}

	// Cache some queries
	queries := []string{
		"What is machine learning?",
		"How does AI work?",
		"Explain neural networks",
	}

	for _, query := range queries {
		result := fmt.Sprintf("Answer to: %s", query)
		if err := cache.CacheQuery(ctx, query, result, 1*time.Hour); err != nil {
			fmt.Printf("   ⚠️  Failed to cache query: %v\n", err)
		} else {
			fmt.Printf("   ✓ Cached query: %s\n", query)
		}
	}

	// Find similar queries
	similar, err := cache.GetSimilarQueries(ctx, "What is AI?", 0.5)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to find similar queries: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d similar queries\n", len(similar))
	}

	// Step 10: Get Knowledge Metrics
	fmt.Println("\n10. Getting Knowledge Metrics...")
	metrics, err := manager.GetKnowledgeMetrics(ctx)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to get metrics: %v\n", err)
	} else {
		fmt.Printf("   ✓ Knowledge Metrics:\n")
		fmt.Printf("     - Total Documents: %d\n", metrics.TotalDocuments)
		fmt.Printf("     - Total Knowledge Bases: %d\n", metrics.TotalKnowledgeBases)
		fmt.Printf("     - Storage Used: %d bytes\n", metrics.StorageUsed)
		fmt.Printf("     - Index Health: %.2f\n", metrics.IndexHealth)
	}

	return nil
}

func createSampleDocuments(kbID string) []*knowledge.Document {
	return []*knowledge.Document{
		{
			ID:              "doc-ai-intro",
			Title:           "Introduction to Artificial Intelligence",
			Content:         "Artificial Intelligence (AI) is a branch of computer science that aims to create intelligent machines capable of performing tasks that typically require human intelligence. These tasks include learning, reasoning, problem-solving, perception, and language understanding. AI systems can be categorized into narrow AI, which is designed for specific tasks, and general AI, which would have human-like cognitive abilities across all domains.",
			ContentType:     "text",
			Language:        "en",
			Source:          "educational",
			Author:          "AI Research Team",
			Tags:            []string{"ai", "artificial-intelligence", "computer-science", "technology"},
			Categories:      []string{"education", "technology"},
			KnowledgeBaseID: kbID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              "doc-ml-basics",
			Title:           "Machine Learning Fundamentals",
			Content:         "Machine Learning (ML) is a subset of artificial intelligence that enables computers to learn and improve from experience without being explicitly programmed. ML algorithms build mathematical models based on training data to make predictions or decisions. The main types of machine learning include supervised learning, unsupervised learning, and reinforcement learning. Applications of ML include image recognition, natural language processing, recommendation systems, and autonomous vehicles.",
			ContentType:     "text",
			Language:        "en",
			Source:          "educational",
			Author:          "ML Research Team",
			Tags:            []string{"machine-learning", "ml", "algorithms", "data-science"},
			Categories:      []string{"education", "data-science"},
			KnowledgeBaseID: kbID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              "doc-deep-learning",
			Title:           "Deep Learning and Neural Networks",
			Content:         "Deep Learning is a subset of machine learning based on artificial neural networks with multiple layers. These deep neural networks can automatically learn hierarchical representations of data, making them particularly effective for tasks like image recognition, speech processing, and natural language understanding. Popular deep learning architectures include Convolutional Neural Networks (CNNs) for image processing, Recurrent Neural Networks (RNNs) for sequential data, and Transformers for natural language processing.",
			ContentType:     "text",
			Language:        "en",
			Source:          "educational",
			Author:          "Deep Learning Team",
			Tags:            []string{"deep-learning", "neural-networks", "cnn", "rnn", "transformers"},
			Categories:      []string{"education", "advanced-ai"},
			KnowledgeBaseID: kbID,
			CreatedAt:       time.Now(),
		},
		{
			ID:              "doc-nlp",
			Title:           "Natural Language Processing",
			Content:         "Natural Language Processing (NLP) is a field of artificial intelligence that focuses on the interaction between computers and human language. NLP combines computational linguistics with machine learning and deep learning to enable computers to understand, interpret, and generate human language. Key NLP tasks include text classification, sentiment analysis, named entity recognition, machine translation, and question answering. Modern NLP systems often use transformer-based models like BERT and GPT.",
			ContentType:     "text",
			Language:        "en",
			Source:          "educational",
			Author:          "NLP Research Team",
			Tags:            []string{"nlp", "natural-language-processing", "bert", "gpt", "transformers"},
			Categories:      []string{"education", "language-ai"},
			KnowledgeBaseID: kbID,
			CreatedAt:       time.Now(),
		},
	}
}

func createSampleKnowledgeGraph() ([]*knowledge.Entity, []*knowledge.Relationship) {
	entities := []*knowledge.Entity{
		{
			ID:          "ai",
			Name:        "Artificial Intelligence",
			Type:        "concept",
			Description: "Branch of computer science creating intelligent machines",
			Confidence:  0.95,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "machine-learning",
			Name:        "Machine Learning",
			Type:        "concept",
			Description: "Subset of AI that learns from data",
			Confidence:  0.95,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "deep-learning",
			Name:        "Deep Learning",
			Type:        "concept",
			Description: "ML using deep neural networks",
			Confidence:  0.95,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "neural-networks",
			Name:        "Neural Networks",
			Type:        "concept",
			Description: "Computing systems inspired by biological neural networks",
			Confidence:  0.90,
			CreatedAt:   time.Now(),
		},
		{
			ID:          "nlp",
			Name:        "Natural Language Processing",
			Type:        "concept",
			Description: "AI field focused on language understanding",
			Confidence:  0.90,
			CreatedAt:   time.Now(),
		},
	}

	relationships := []*knowledge.Relationship{
		{
			ID:         "rel-1",
			FromEntity: "machine-learning",
			ToEntity:   "ai",
			Type:       "subset_of",
			Confidence: 0.95,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "rel-2",
			FromEntity: "deep-learning",
			ToEntity:   "machine-learning",
			Type:       "subset_of",
			Confidence: 0.95,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "rel-3",
			FromEntity: "deep-learning",
			ToEntity:   "neural-networks",
			Type:       "uses",
			Confidence: 0.90,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "rel-4",
			FromEntity: "nlp",
			ToEntity:   "ai",
			Type:       "subset_of",
			Confidence: 0.90,
			CreatedAt:  time.Now(),
		},
		{
			ID:         "rel-5",
			FromEntity: "nlp",
			ToEntity:   "machine-learning",
			Type:       "uses",
			Confidence: 0.85,
			CreatedAt:  time.Now(),
		},
	}

	return entities, relationships
}

// runMockKnowledgeDemo shows what the knowledge management demo would do
func runMockKnowledgeDemo() error {
	fmt.Println("\n📋 Mock Knowledge Management Demo")
	fmt.Println("   (Showing what would happen with proper configuration)")

	fmt.Println("\n2. Creating Knowledge Bases...")
	fmt.Println("   ✓ Would create: Technical Documentation KB")
	fmt.Println("   ✓ Would create: Project Knowledge KB")
	fmt.Println("   ✓ Would create: Team Expertise KB")

	fmt.Println("\n3. Adding Documents...")
	fmt.Println("   ✓ Would add: AI and Machine Learning Overview")
	fmt.Println("   ✓ Would add: Natural Language Processing Guide")
	fmt.Println("   ✓ Would add: Deep Learning Fundamentals")

	fmt.Println("\n4. Processing and Indexing...")
	fmt.Println("   ✓ Would chunk documents into 1000-char segments")
	fmt.Println("   ✓ Would generate embeddings for each chunk")
	fmt.Println("   ✓ Would index in vector database")

	fmt.Println("\n5. Building Knowledge Graph...")
	fmt.Println("   ✓ Would extract entities: AI, Machine Learning, NLP, etc.")
	fmt.Println("   ✓ Would create relationships between concepts")
	fmt.Println("   ✓ Would build semantic connections")

	fmt.Println("\n6. Performing Semantic Search...")
	fmt.Println("   ✓ Would search: 'What is machine learning?'")
	fmt.Println("   ✓ Would return relevant document chunks")
	fmt.Println("   ✓ Would rank by semantic similarity")

	fmt.Println("\n7. RAG Pipeline Demo...")
	fmt.Println("   ✓ Would retrieve relevant context")
	fmt.Println("   ✓ Would generate augmented responses")
	fmt.Println("   ✓ Would provide source citations")

	fmt.Println("\n8. Analytics and Insights...")
	fmt.Println("   ✓ Would show knowledge base statistics")
	fmt.Println("   ✓ Would display search patterns")
	fmt.Println("   ✓ Would provide usage metrics")

	return nil
}
