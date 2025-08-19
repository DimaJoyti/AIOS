package system

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SemanticSearchEngine provides AI-powered semantic file search
type SemanticSearchEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config SemanticSearchConfig
	mu     sync.RWMutex

	// AI integration
	aiOrchestrator *ai.Orchestrator

	// Search components
	indexer        *SemanticIndexer
	vectorStore    *VectorStore
	queryProcessor *QueryProcessor
	ranker         *SearchRanker

	// Search state
	searchHistory []SearchEvent
	userProfiles  map[string]*SearchProfile

	// Performance metrics
	totalSearches    int
	averageLatency   time.Duration
	satisfactionRate float64
}

// SemanticSearchConfig defines semantic search configuration
type SemanticSearchConfig struct {
	IndexingEnabled     bool          `json:"indexing_enabled"`
	RealTimeIndexing    bool          `json:"real_time_indexing"`
	VectorDimensions    int           `json:"vector_dimensions"`
	MaxResults          int           `json:"max_results"`
	MinSimilarity       float64       `json:"min_similarity"`
	QueryExpansion      bool          `json:"query_expansion"`
	PersonalizedRanking bool          `json:"personalized_ranking"`
	CacheResults        bool          `json:"cache_results"`
	CacheTTL            time.Duration `json:"cache_ttl"`
}

// SemanticIndexer handles file content indexing
type SemanticIndexer struct {
	logger         *logrus.Logger
	aiOrchestrator *ai.Orchestrator
	mu             sync.RWMutex

	// Indexing state
	indexedFiles   map[string]*IndexedFile
	indexingQueue  chan IndexingTask
	lastIndexing   time.Time
	indexingActive bool
}

// VectorStore manages semantic vectors
type VectorStore struct {
	vectors    map[string][]float64
	metadata   map[string]*VectorMetadata
	dimensions int
	mu         sync.RWMutex
}

// QueryProcessor handles query understanding and expansion
type QueryProcessor struct {
	logger         *logrus.Logger
	aiOrchestrator *ai.Orchestrator

	// Query processing
	synonyms       map[string][]string
	queryHistory   []ProcessedQuery
	expansionRules []ExpansionRule
}

// SearchRanker handles result ranking and personalization
type SearchRanker struct {
	logger         *logrus.Logger
	aiOrchestrator *ai.Orchestrator

	// Ranking models
	models       map[string]RankingModel
	userProfiles map[string]*SearchProfile
}

// SearchEvent represents a search event
type SearchEvent struct {
	QueryID        string                 `json:"query_id"`
	UserID         string                 `json:"user_id"`
	Query          string                 `json:"query"`
	ProcessedQuery *ProcessedQuery        `json:"processed_query"`
	Results        []SearchResult         `json:"results"`
	Timestamp      time.Time              `json:"timestamp"`
	Latency        time.Duration          `json:"latency"`
	ClickedResults []string               `json:"clicked_results"`
	Satisfaction   float64                `json:"satisfaction"`
	Context        map[string]interface{} `json:"context"`
}

// SearchProfile represents a user's search profile
type SearchProfile struct {
	UserID         string             `json:"user_id"`
	SearchHistory  []SearchEvent      `json:"search_history"`
	PreferredTypes []string           `json:"preferred_types"`
	QueryPatterns  map[string]float64 `json:"query_patterns"`
	ClickBehavior  map[string]float64 `json:"click_behavior"`
	PersonalVector []float64          `json:"personal_vector"`
	LastUpdated    time.Time          `json:"last_updated"`
}

// IndexedFile represents an indexed file
type IndexedFile struct {
	FilePath      string        `json:"file_path"`
	Content       string        `json:"content"`
	ContentVector []float64     `json:"content_vector"`
	Metadata      *FileMetadata `json:"metadata"`
	IndexedAt     time.Time     `json:"indexed_at"`
	LastModified  time.Time     `json:"last_modified"`
	IndexVersion  int           `json:"index_version"`
}

// IndexingTask represents a file indexing task
type IndexingTask struct {
	FilePath  string                 `json:"file_path"`
	Content   string                 `json:"content"`
	Priority  int                    `json:"priority"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

// VectorMetadata represents metadata for a vector
type VectorMetadata struct {
	FilePath   string                 `json:"file_path"`
	FileType   string                 `json:"file_type"`
	Size       int64                  `json:"size"`
	ModTime    time.Time              `json:"mod_time"`
	Tags       []string               `json:"tags"`
	Categories []string               `json:"categories"`
	Importance float64                `json:"importance"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ProcessedQuery represents a processed search query
type ProcessedQuery struct {
	OriginalQuery string                 `json:"original_query"`
	ExpandedQuery string                 `json:"expanded_query"`
	QueryVector   []float64              `json:"query_vector"`
	Intent        string                 `json:"intent"`
	Entities      []string               `json:"entities"`
	Keywords      []string               `json:"keywords"`
	Filters       map[string]interface{} `json:"filters"`
	ProcessedAt   time.Time              `json:"processed_at"`
}

// SearchResult represents a search result
type SearchResult struct {
	FilePath   string                 `json:"file_path"`
	Title      string                 `json:"title"`
	Snippet    string                 `json:"snippet"`
	Score      float64                `json:"score"`
	Similarity float64                `json:"similarity"`
	Relevance  float64                `json:"relevance"`
	FileType   string                 `json:"file_type"`
	Size       int64                  `json:"size"`
	ModTime    time.Time              `json:"mod_time"`
	Tags       []string               `json:"tags"`
	Categories []string               `json:"categories"`
	Highlights []string               `json:"highlights"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ExpansionRule defines query expansion rules
type ExpansionRule struct {
	Pattern    string   `json:"pattern"`
	Expansions []string `json:"expansions"`
	Weight     float64  `json:"weight"`
	Context    string   `json:"context"`
}

// RankingModel interface for different ranking approaches
type RankingModel interface {
	Rank(results []SearchResult, query *ProcessedQuery, profile *SearchProfile) []SearchResult
	GetName() string
	Train(searchHistory []SearchEvent) error
}

// NewSemanticSearchEngine creates a new semantic search engine
func NewSemanticSearchEngine(logger *logrus.Logger, config SemanticSearchConfig, aiOrchestrator *ai.Orchestrator) *SemanticSearchEngine {
	tracer := otel.Tracer("semantic-search-engine")

	engine := &SemanticSearchEngine{
		logger:         logger,
		tracer:         tracer,
		config:         config,
		aiOrchestrator: aiOrchestrator,
		searchHistory:  make([]SearchEvent, 0),
		userProfiles:   make(map[string]*SearchProfile),
	}

	// Initialize components
	engine.indexer = NewSemanticIndexer(logger, aiOrchestrator)
	engine.vectorStore = NewVectorStore(config.VectorDimensions)
	engine.queryProcessor = NewQueryProcessor(logger, aiOrchestrator)
	engine.ranker = NewSearchRanker(logger, aiOrchestrator)

	return engine
}

// NewSemanticIndexer creates a new semantic indexer
func NewSemanticIndexer(logger *logrus.Logger, aiOrchestrator *ai.Orchestrator) *SemanticIndexer {
	return &SemanticIndexer{
		logger:         logger,
		aiOrchestrator: aiOrchestrator,
		indexedFiles:   make(map[string]*IndexedFile),
		indexingQueue:  make(chan IndexingTask, 1000),
		lastIndexing:   time.Now(),
	}
}

// NewVectorStore creates a new vector store
func NewVectorStore(dimensions int) *VectorStore {
	return &VectorStore{
		vectors:    make(map[string][]float64),
		metadata:   make(map[string]*VectorMetadata),
		dimensions: dimensions,
	}
}

// NewQueryProcessor creates a new query processor
func NewQueryProcessor(logger *logrus.Logger, aiOrchestrator *ai.Orchestrator) *QueryProcessor {
	return &QueryProcessor{
		logger:         logger,
		aiOrchestrator: aiOrchestrator,
		synonyms:       make(map[string][]string),
		queryHistory:   make([]ProcessedQuery, 0),
		expansionRules: make([]ExpansionRule, 0),
	}
}

// NewSearchRanker creates a new search ranker
func NewSearchRanker(logger *logrus.Logger, aiOrchestrator *ai.Orchestrator) *SearchRanker {
	return &SearchRanker{
		logger:         logger,
		aiOrchestrator: aiOrchestrator,
		models:         make(map[string]RankingModel),
		userProfiles:   make(map[string]*SearchProfile),
	}
}

// Search performs semantic search
func (sse *SemanticSearchEngine) Search(ctx context.Context, query string, userID string, options map[string]interface{}) ([]SearchResult, error) {
	ctx, span := sse.tracer.Start(ctx, "semanticSearchEngine.Search")
	defer span.End()

	start := time.Now()
	queryID := fmt.Sprintf("search_%d", time.Now().UnixNano())

	// Process query
	processedQuery, err := sse.queryProcessor.ProcessQuery(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to process query: %w", err)
	}

	// Get user profile
	profile := sse.getUserProfile(userID)

	// Perform vector search
	candidates, err := sse.vectorSearch(ctx, processedQuery)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Rank results
	rankedResults := sse.ranker.RankResults(candidates, processedQuery, profile)

	// Apply filters and limits
	filteredResults := sse.applyFilters(rankedResults, options)

	// Limit results
	if len(filteredResults) > sse.config.MaxResults {
		filteredResults = filteredResults[:sse.config.MaxResults]
	}

	// Record search event
	searchEvent := SearchEvent{
		QueryID:        queryID,
		UserID:         userID,
		Query:          query,
		ProcessedQuery: processedQuery,
		Results:        filteredResults,
		Timestamp:      start,
		Latency:        time.Since(start),
		Context:        options,
	}

	sse.recordSearchEvent(searchEvent)

	sse.logger.WithFields(logrus.Fields{
		"query_id":     queryID,
		"user_id":      userID,
		"query":        query,
		"result_count": len(filteredResults),
		"latency":      searchEvent.Latency,
	}).Debug("Semantic search completed")

	return filteredResults, nil
}

// ProcessQuery processes a search query
func (qp *QueryProcessor) ProcessQuery(ctx context.Context, query string, userID string) (*ProcessedQuery, error) {
	processed := &ProcessedQuery{
		OriginalQuery: query,
		ProcessedAt:   time.Now(),
	}

	// Expand query if enabled
	expandedQuery := qp.expandQuery(query)
	processed.ExpandedQuery = expandedQuery

	// Extract entities and keywords
	processed.Entities = qp.extractEntities(expandedQuery)
	processed.Keywords = qp.extractKeywords(expandedQuery)

	// Determine intent
	processed.Intent = qp.determineIntent(expandedQuery)

	// Generate query vector using AI
	if qp.aiOrchestrator != nil {
		vector, err := qp.generateQueryVector(ctx, expandedQuery)
		if err == nil {
			processed.QueryVector = vector
		}
	}

	return processed, nil
}

// expandQuery expands the query with synonyms and related terms
func (qp *QueryProcessor) expandQuery(query string) string {
	words := strings.Fields(strings.ToLower(query))
	expandedWords := make([]string, 0)

	for _, word := range words {
		expandedWords = append(expandedWords, word)

		// Add synonyms
		if synonyms, exists := qp.synonyms[word]; exists {
			expandedWords = append(expandedWords, synonyms...)
		}
	}

	return strings.Join(expandedWords, " ")
}

// extractEntities extracts named entities from the query
func (qp *QueryProcessor) extractEntities(query string) []string {
	// Simplified entity extraction
	// In a real implementation, this would use NLP models
	entities := make([]string, 0)

	// Look for file extensions
	words := strings.Fields(query)
	for _, word := range words {
		if strings.HasPrefix(word, ".") && len(word) > 1 {
			entities = append(entities, word)
		}
	}

	return entities
}

// extractKeywords extracts keywords from the query
func (qp *QueryProcessor) extractKeywords(query string) []string {
	// Simplified keyword extraction
	words := strings.Fields(strings.ToLower(query))

	// Filter out common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
	}

	keywords := make([]string, 0)
	for _, word := range words {
		if !stopWords[word] && len(word) > 2 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// determineIntent determines the search intent
func (qp *QueryProcessor) determineIntent(query string) string {
	query = strings.ToLower(query)

	if strings.Contains(query, "recent") || strings.Contains(query, "latest") {
		return "recent"
	} else if strings.Contains(query, "large") || strings.Contains(query, "big") {
		return "size"
	} else if strings.Contains(query, "code") || strings.Contains(query, "function") {
		return "code"
	} else if strings.Contains(query, "document") || strings.Contains(query, "text") {
		return "document"
	} else if strings.Contains(query, "image") || strings.Contains(query, "photo") {
		return "image"
	}

	return "general"
}

// generateQueryVector generates a semantic vector for the query
func (qp *QueryProcessor) generateQueryVector(ctx context.Context, query string) ([]float64, error) {
	// Create AI request for query vectorization
	aiRequest := &models.AIRequest{
		ID:    fmt.Sprintf("query-vector-%d", time.Now().Unix()),
		Type:  "embedding",
		Input: query,
		Parameters: map[string]interface{}{
			"task": "text_embedding",
		},
		Timeout:   3 * time.Second,
		Timestamp: time.Now(),
	}

	response, err := qp.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		return nil, err
	}

	// Parse vector from response
	if vector, ok := response.Result.([]float64); ok {
		return vector, nil
	}

	return nil, fmt.Errorf("invalid vector response")
}

// vectorSearch performs vector similarity search
func (sse *SemanticSearchEngine) vectorSearch(ctx context.Context, query *ProcessedQuery) ([]SearchResult, error) {
	if len(query.QueryVector) == 0 {
		return sse.keywordSearch(query)
	}

	sse.vectorStore.mu.RLock()
	defer sse.vectorStore.mu.RUnlock()

	results := make([]SearchResult, 0)

	for filePath, vector := range sse.vectorStore.vectors {
		similarity := sse.calculateCosineSimilarity(query.QueryVector, vector)

		if similarity >= sse.config.MinSimilarity {
			metadata := sse.vectorStore.metadata[filePath]

			result := SearchResult{
				FilePath:   filePath,
				Title:      extractTitle(filePath),
				Snippet:    extractSnippet(filePath, query.Keywords),
				Similarity: similarity,
				Score:      similarity,
				FileType:   metadata.FileType,
				Size:       metadata.Size,
				ModTime:    metadata.ModTime,
				Tags:       metadata.Tags,
				Categories: metadata.Categories,
				Metadata:   metadata.Metadata,
			}

			results = append(results, result)
		}
	}

	// Sort by similarity
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	return results, nil
}

// keywordSearch performs keyword-based search as fallback
func (sse *SemanticSearchEngine) keywordSearch(query *ProcessedQuery) ([]SearchResult, error) {
	// Simplified keyword search implementation
	results := make([]SearchResult, 0)

	sse.indexer.mu.RLock()
	defer sse.indexer.mu.RUnlock()

	for filePath, indexedFile := range sse.indexer.indexedFiles {
		score := sse.calculateKeywordScore(query.Keywords, indexedFile.Content)

		if score > 0 {
			result := SearchResult{
				FilePath: filePath,
				Title:    extractTitle(filePath),
				Snippet:  extractSnippet(filePath, query.Keywords),
				Score:    score,
				FileType: indexedFile.Metadata.FileType,
				Size:     indexedFile.Metadata.Size,
				ModTime:  indexedFile.Metadata.ModTime,
			}

			results = append(results, result)
		}
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateCosineSimilarity calculates cosine similarity between vectors
func (sse *SemanticSearchEngine) calculateCosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// calculateKeywordScore calculates keyword-based relevance score
func (sse *SemanticSearchEngine) calculateKeywordScore(keywords []string, content string) float64 {
	content = strings.ToLower(content)
	score := 0.0

	for _, keyword := range keywords {
		count := strings.Count(content, strings.ToLower(keyword))
		score += float64(count) * (1.0 / float64(len(keyword))) // Longer keywords get higher weight
	}

	return score
}

// RankResults ranks search results
func (sr *SearchRanker) RankResults(results []SearchResult, query *ProcessedQuery, profile *SearchProfile) []SearchResult {
	// Apply personalization if enabled and profile exists
	if profile != nil {
		for i := range results {
			personalScore := sr.calculatePersonalScore(&results[i], profile)
			results[i].Relevance = (results[i].Score + personalScore) / 2.0
		}
	} else {
		for i := range results {
			results[i].Relevance = results[i].Score
		}
	}

	// Sort by relevance
	sort.Slice(results, func(i, j int) bool {
		return results[i].Relevance > results[j].Relevance
	})

	return results
}

// calculatePersonalScore calculates personalized relevance score
func (sr *SearchRanker) calculatePersonalScore(result *SearchResult, profile *SearchProfile) float64 {
	score := 0.0

	// File type preference
	if weight, exists := profile.ClickBehavior[result.FileType]; exists {
		score += weight * 0.3
	}

	// Category preference
	for _, category := range result.Categories {
		if weight, exists := profile.ClickBehavior[category]; exists {
			score += weight * 0.2
		}
	}

	// Recency preference (if user prefers recent files)
	recencyScore := 1.0 - (time.Since(result.ModTime).Hours() / (24 * 30)) // Month-based decay
	if recencyScore < 0 {
		recencyScore = 0
	}
	score += recencyScore * 0.1

	return score
}

// Helper methods

func (sse *SemanticSearchEngine) getUserProfile(userID string) *SearchProfile {
	sse.mu.RLock()
	defer sse.mu.RUnlock()

	if profile, exists := sse.userProfiles[userID]; exists {
		return profile
	}

	// Create new profile
	profile := &SearchProfile{
		UserID:         userID,
		SearchHistory:  make([]SearchEvent, 0),
		PreferredTypes: make([]string, 0),
		QueryPatterns:  make(map[string]float64),
		ClickBehavior:  make(map[string]float64),
		PersonalVector: make([]float64, sse.config.VectorDimensions),
		LastUpdated:    time.Now(),
	}

	sse.userProfiles[userID] = profile
	return profile
}

func (sse *SemanticSearchEngine) applyFilters(results []SearchResult, options map[string]interface{}) []SearchResult {
	filtered := make([]SearchResult, 0)

	for _, result := range results {
		include := true

		// File type filter
		if fileTypes, ok := options["file_types"].([]string); ok {
			found := false
			for _, ft := range fileTypes {
				if result.FileType == ft {
					found = true
					break
				}
			}
			if !found {
				include = false
			}
		}

		// Size filter
		if maxSize, ok := options["max_size"].(int64); ok {
			if result.Size > maxSize {
				include = false
			}
		}

		if include {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

func (sse *SemanticSearchEngine) recordSearchEvent(event SearchEvent) {
	sse.mu.Lock()
	defer sse.mu.Unlock()

	sse.searchHistory = append(sse.searchHistory, event)
	sse.totalSearches++

	// Update average latency
	if sse.totalSearches == 1 {
		sse.averageLatency = event.Latency
	} else {
		sse.averageLatency = (sse.averageLatency*time.Duration(sse.totalSearches-1) + event.Latency) / time.Duration(sse.totalSearches)
	}

	// Maintain history size
	if len(sse.searchHistory) > 10000 {
		sse.searchHistory = sse.searchHistory[1000:]
	}
}

func extractTitle(filePath string) string {
	// Extract filename as title
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return filePath
}

func extractSnippet(filePath string, keywords []string) string {
	// This would extract relevant snippets from file content
	// For now, return a placeholder
	return fmt.Sprintf("Snippet from %s containing keywords: %s", extractTitle(filePath), strings.Join(keywords, ", "))
}
