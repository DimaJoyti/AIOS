package knowledge

import (
	"context"
	"time"
)

// KnowledgeManager defines the main interface for knowledge management
type KnowledgeManager interface {
	// Document Management
	AddDocument(ctx context.Context, doc *Document) error
	AddDocuments(ctx context.Context, docs []*Document) error
	GetDocument(ctx context.Context, id string) (*Document, error)
	UpdateDocument(ctx context.Context, doc *Document) error
	DeleteDocument(ctx context.Context, id string) error
	
	// Knowledge Base Operations
	CreateKnowledgeBase(ctx context.Context, config *KnowledgeBaseConfig) (*KnowledgeBase, error)
	GetKnowledgeBase(ctx context.Context, id string) (*KnowledgeBase, error)
	ListKnowledgeBases(ctx context.Context) ([]*KnowledgeBase, error)
	DeleteKnowledgeBase(ctx context.Context, id string) error
	
	// Search and Retrieval
	Search(ctx context.Context, query *SearchQuery) (*SearchResult, error)
	SimilaritySearch(ctx context.Context, query string, options *SearchOptions) ([]*Document, error)
	HybridSearch(ctx context.Context, query string, options *SearchOptions) ([]*Document, error)
	
	// RAG Operations
	RetrieveAndGenerate(ctx context.Context, query string, options *RAGOptions) (*RAGResponse, error)
	RetrieveContext(ctx context.Context, query string, options *RetrievalOptions) (*RetrievalResult, error)
	
	// Knowledge Graph
	ExtractEntities(ctx context.Context, text string) ([]*Entity, error)
	ExtractRelationships(ctx context.Context, text string) ([]*Relationship, error)
	BuildKnowledgeGraph(ctx context.Context, documents []*Document) (*KnowledgeGraph, error)
	QueryKnowledgeGraph(ctx context.Context, query *GraphQuery) (*GraphResult, error)
	
	// Analytics and Insights
	GetKnowledgeMetrics(ctx context.Context) (*KnowledgeMetrics, error)
	GetSearchAnalytics(ctx context.Context, timeRange *TimeRange) (*SearchAnalytics, error)
	
	// Maintenance
	ReindexKnowledgeBase(ctx context.Context, kbID string) error
	OptimizeStorage(ctx context.Context) error
	BackupKnowledgeBase(ctx context.Context, kbID string, destination string) error
}

// DocumentProcessor handles document processing and chunking
type DocumentProcessor interface {
	ProcessDocument(ctx context.Context, doc *Document) (*ProcessedDocument, error)
	ChunkDocument(ctx context.Context, doc *Document, strategy ChunkingStrategy) ([]*DocumentChunk, error)
	ExtractMetadata(ctx context.Context, doc *Document) (map[string]interface{}, error)
	DetectLanguage(ctx context.Context, text string) (string, error)
	CleanText(ctx context.Context, text string) (string, error)
}

// EmbeddingManager handles embedding generation and management
type EmbeddingManager interface {
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)
	GetEmbeddingDimensions() int
	GetEmbeddingModel() string
	CompareEmbeddings(embedding1, embedding2 []float32) float32
}

// KnowledgeGraph represents a knowledge graph
type KnowledgeGraph interface {
	AddEntity(ctx context.Context, entity *Entity) error
	AddRelationship(ctx context.Context, rel *Relationship) error
	GetEntity(ctx context.Context, id string) (*Entity, error)
	GetRelationships(ctx context.Context, entityID string) ([]*Relationship, error)
	FindPath(ctx context.Context, fromID, toID string, maxDepth int) ([]*Path, error)
	QueryGraph(ctx context.Context, query *GraphQuery) (*GraphResult, error)
	GetNeighbors(ctx context.Context, entityID string, depth int) ([]*Entity, error)
	CalculateCentrality(ctx context.Context, entityID string) (float64, error)
}

// RAGPipeline defines the RAG pipeline interface
type RAGPipeline interface {
	Retrieve(ctx context.Context, query string, options *RetrievalOptions) (*RetrievalResult, error)
	Rerank(ctx context.Context, query string, documents []*Document) ([]*Document, error)
	Generate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error)
	Pipeline(ctx context.Context, query string, options *RAGOptions) (*RAGResponse, error)
}

// QueryProcessor handles query understanding and expansion
type QueryProcessor interface {
	ProcessQuery(ctx context.Context, query string) (*ProcessedQuery, error)
	ExpandQuery(ctx context.Context, query string) (*ExpandedQuery, error)
	ExtractIntent(ctx context.Context, query string) (*QueryIntent, error)
	ExtractEntities(ctx context.Context, query string) ([]*QueryEntity, error)
	RewriteQuery(ctx context.Context, query string, context *QueryContext) (string, error)
}

// SemanticCache provides caching for semantic operations
type SemanticCache interface {
	GetSimilarQueries(ctx context.Context, query string, threshold float32) ([]*CachedQuery, error)
	CacheQuery(ctx context.Context, query string, result interface{}, ttl time.Duration) error
	CacheEmbedding(ctx context.Context, text string, embedding []float32, ttl time.Duration) error
	GetCachedEmbedding(ctx context.Context, text string) ([]float32, bool, error)
	InvalidateCache(ctx context.Context, pattern string) error
	GetCacheStats(ctx context.Context) (*CacheStats, error)
}

// KnowledgeIndexer handles indexing operations
type KnowledgeIndexer interface {
	IndexDocument(ctx context.Context, doc *Document) error
	IndexDocuments(ctx context.Context, docs []*Document) error
	UpdateIndex(ctx context.Context, docID string, doc *Document) error
	DeleteFromIndex(ctx context.Context, docID string) error
	RebuildIndex(ctx context.Context, kbID string) error
	GetIndexStats(ctx context.Context) (*IndexStats, error)
}

// MultiModalProcessor handles different content types
type MultiModalProcessor interface {
	ProcessText(ctx context.Context, text string) (*ProcessedContent, error)
	ProcessImage(ctx context.Context, imageData []byte) (*ProcessedContent, error)
	ProcessAudio(ctx context.Context, audioData []byte) (*ProcessedContent, error)
	ProcessVideo(ctx context.Context, videoData []byte) (*ProcessedContent, error)
	ProcessPDF(ctx context.Context, pdfData []byte) (*ProcessedContent, error)
	ExtractTextFromImage(ctx context.Context, imageData []byte) (string, error)
	TranscribeAudio(ctx context.Context, audioData []byte) (string, error)
}

// KnowledgeValidator validates knowledge consistency
type KnowledgeValidator interface {
	ValidateDocument(ctx context.Context, doc *Document) (*ValidationResult, error)
	ValidateKnowledgeBase(ctx context.Context, kbID string) (*ValidationResult, error)
	CheckConsistency(ctx context.Context, kbID string) (*ConsistencyReport, error)
	DetectDuplicates(ctx context.Context, kbID string) ([]*DuplicateGroup, error)
	ValidateRelationships(ctx context.Context, relationships []*Relationship) (*ValidationResult, error)
}

// KnowledgeExporter handles knowledge export operations
type KnowledgeExporter interface {
	ExportKnowledgeBase(ctx context.Context, kbID string, format ExportFormat) ([]byte, error)
	ExportDocuments(ctx context.Context, docIDs []string, format ExportFormat) ([]byte, error)
	ExportKnowledgeGraph(ctx context.Context, kbID string, format GraphExportFormat) ([]byte, error)
	ImportKnowledgeBase(ctx context.Context, data []byte, format ExportFormat) (*KnowledgeBase, error)
}

// KnowledgeCollaborator handles collaborative features
type KnowledgeCollaborator interface {
	ShareKnowledgeBase(ctx context.Context, kbID string, userID string, permissions *Permissions) error
	GetSharedKnowledgeBases(ctx context.Context, userID string) ([]*SharedKnowledgeBase, error)
	CreateAnnotation(ctx context.Context, annotation *Annotation) error
	GetAnnotations(ctx context.Context, docID string) ([]*Annotation, error)
	CreateComment(ctx context.Context, comment *Comment) error
	GetComments(ctx context.Context, docID string) ([]*Comment, error)
}

// KnowledgeRecommender provides recommendation capabilities
type KnowledgeRecommender interface {
	RecommendDocuments(ctx context.Context, userID string, options *RecommendationOptions) ([]*Document, error)
	RecommendQueries(ctx context.Context, query string, options *QueryRecommendationOptions) ([]string, error)
	RecommendRelatedTopics(ctx context.Context, topic string, options *TopicRecommendationOptions) ([]*Topic, error)
	GetTrendingTopics(ctx context.Context, timeRange *TimeRange) ([]*TrendingTopic, error)
	PersonalizeResults(ctx context.Context, userID string, results []*Document) ([]*Document, error)
}

// KnowledgeMonitor provides monitoring and observability
type KnowledgeMonitor interface {
	RecordQuery(ctx context.Context, query *QueryEvent) error
	RecordRetrieval(ctx context.Context, retrieval *RetrievalEvent) error
	RecordGeneration(ctx context.Context, generation *GenerationEvent) error
	GetQueryMetrics(ctx context.Context, timeRange *TimeRange) (*QueryMetrics, error)
	GetPerformanceMetrics(ctx context.Context, timeRange *TimeRange) (*PerformanceMetrics, error)
	GetUsageMetrics(ctx context.Context, timeRange *TimeRange) (*UsageMetrics, error)
	CreateAlert(ctx context.Context, alert *Alert) error
	GetAlerts(ctx context.Context, status AlertStatus) ([]*Alert, error)
}

// KnowledgeVersioning handles versioning and history
type KnowledgeVersioning interface {
	CreateVersion(ctx context.Context, kbID string, description string) (*Version, error)
	GetVersions(ctx context.Context, kbID string) ([]*Version, error)
	RestoreVersion(ctx context.Context, kbID string, versionID string) error
	CompareVersions(ctx context.Context, kbID string, version1, version2 string) (*VersionComparison, error)
	GetDocumentHistory(ctx context.Context, docID string) ([]*DocumentVersion, error)
	RevertDocument(ctx context.Context, docID string, versionID string) error
}

// KnowledgeSecurity handles security and access control
type KnowledgeSecurity interface {
	AuthorizeAccess(ctx context.Context, userID string, resource string, action string) error
	EncryptDocument(ctx context.Context, doc *Document) (*EncryptedDocument, error)
	DecryptDocument(ctx context.Context, encDoc *EncryptedDocument) (*Document, error)
	AuditAccess(ctx context.Context, userID string, resource string, action string) error
	GetAuditLog(ctx context.Context, filters *AuditFilters) ([]*AuditEntry, error)
	SetPermissions(ctx context.Context, resource string, permissions *Permissions) error
	GetPermissions(ctx context.Context, resource string) (*Permissions, error)
}

// KnowledgeWorkflow handles workflow automation
type KnowledgeWorkflow interface {
	CreateWorkflow(ctx context.Context, workflow *Workflow) error
	ExecuteWorkflow(ctx context.Context, workflowID string, input *WorkflowInput) (*WorkflowResult, error)
	GetWorkflows(ctx context.Context) ([]*Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *Workflow) error
	DeleteWorkflow(ctx context.Context, workflowID string) error
	ScheduleWorkflow(ctx context.Context, workflowID string, schedule *Schedule) error
}
