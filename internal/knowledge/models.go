package knowledge

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// KnowledgeBase represents a knowledge base in the database
type KnowledgeBase struct {
	ID          uuid.UUID           `db:"id" json:"id"`
	Name        string              `db:"name" json:"name"`
	Description *string             `db:"description" json:"description,omitempty"`
	OwnerID     uuid.UUID           `db:"owner_id" json:"owner_id"`
	Config      KnowledgeBaseConfig `db:"config" json:"config"`
	Stats       KnowledgeBaseStats  `db:"stats" json:"stats"`
	Metadata    MetadataMap         `db:"metadata" json:"metadata"`
	Status      string              `db:"status" json:"status"`
	CreatedAt   time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `db:"updated_at" json:"updated_at"`
}

// KnowledgeBaseConfig holds configuration for a knowledge base
type KnowledgeBaseConfig struct {
	EmbeddingModel    string `json:"embedding_model"`
	ChunkSize         int    `json:"chunk_size"`
	ChunkOverlap      int    `json:"chunk_overlap"`
	IndexingEnabled   bool   `json:"indexing_enabled"`
	GraphEnabled      bool   `json:"graph_enabled"`
	VersioningEnabled bool   `json:"versioning_enabled"`
}

// KnowledgeBaseStats holds statistics for a knowledge base
type KnowledgeBaseStats struct {
	DocumentCount int `json:"document_count"`
	ChunkCount    int `json:"chunk_count"`
	EntityCount   int `json:"entity_count"`
	TotalSize     int `json:"total_size"`
}

// Document represents a document in the knowledge base
type Document struct {
	ID               uuid.UUID      `db:"id" json:"id"`
	KnowledgeBaseID  uuid.UUID      `db:"knowledge_base_id" json:"knowledge_base_id"`
	Title            string         `db:"title" json:"title"`
	Content          string         `db:"content" json:"content"`
	ContentType      string         `db:"content_type" json:"content_type"`
	Language         string         `db:"language" json:"language"`
	Source           *string        `db:"source" json:"source,omitempty"`
	URL              *string        `db:"url" json:"url,omitempty"`
	Author           *string        `db:"author" json:"author,omitempty"`
	FilePath         *string        `db:"file_path" json:"file_path,omitempty"`
	FileSize         *int64         `db:"file_size" json:"file_size,omitempty"`
	FileHash         *string        `db:"file_hash" json:"file_hash,omitempty"`
	MimeType         *string        `db:"mime_type" json:"mime_type,omitempty"`
	Tags             pq.StringArray `db:"tags" json:"tags"`
	Categories       pq.StringArray `db:"categories" json:"categories"`
	Metadata         MetadataMap    `db:"metadata" json:"metadata"`
	ProcessingStatus string         `db:"processing_status" json:"processing_status"`
	ProcessingError  *string        `db:"processing_error" json:"processing_error,omitempty"`
	Version          int            `db:"version" json:"version"`
	ParentID         *uuid.UUID     `db:"parent_id" json:"parent_id,omitempty"`
	CreatedAt        time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time      `db:"updated_at" json:"updated_at"`
	ProcessedAt      *time.Time     `db:"processed_at" json:"processed_at,omitempty"`
}

// DocumentChunk represents a chunk of a document for RAG
type DocumentChunk struct {
	ID             uuid.UUID              `db:"id" json:"id"`
	DocumentID     uuid.UUID              `db:"document_id" json:"document_id"`
	ChunkIndex     int                    `db:"chunk_index" json:"chunk_index"`
	Content        string                 `db:"content" json:"content"`
	ContentLength  int                    `db:"content_length" json:"content_length"`
	ChunkType      string                 `db:"chunk_type" json:"chunk_type"`
	StartPosition  *int                   `db:"start_position" json:"start_position,omitempty"`
	EndPosition    *int                   `db:"end_position" json:"end_position,omitempty"`
	Metadata       map[string]interface{} `db:"metadata" json:"metadata"`
	Embedding      []float64              `db:"embedding" json:"embedding,omitempty"`
	EmbeddingModel string                 `db:"embedding_model" json:"embedding_model"`
	CreatedAt      time.Time              `db:"created_at" json:"created_at"`
}

// CrawlJob represents a web crawling job
type CrawlJob struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	KnowledgeBaseID uuid.UUID              `db:"knowledge_base_id" json:"knowledge_base_id"`
	URL             string                 `db:"url" json:"url"`
	Status          string                 `db:"status" json:"status"`
	MaxPages        int                    `db:"max_pages" json:"max_pages"`
	MaxDepth        int                    `db:"max_depth" json:"max_depth"`
	FollowLinks     bool                   `db:"follow_links" json:"follow_links"`
	PagesFound      int                    `db:"pages_found" json:"pages_found"`
	PagesProcessed  int                    `db:"pages_processed" json:"pages_processed"`
	ErrorMessage    *string                `db:"error_message" json:"error_message,omitempty"`
	Metadata        map[string]interface{} `db:"metadata" json:"metadata"`
	StartedAt       *time.Time             `db:"started_at" json:"started_at,omitempty"`
	CompletedAt     *time.Time             `db:"completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

// CrawledPage represents a crawled web page
type CrawledPage struct {
	ID            uuid.UUID              `db:"id" json:"id"`
	CrawlJobID    uuid.UUID              `db:"crawl_job_id" json:"crawl_job_id"`
	URL           string                 `db:"url" json:"url"`
	Title         *string                `db:"title" json:"title,omitempty"`
	Content       *string                `db:"content" json:"content,omitempty"`
	ContentLength *int                   `db:"content_length" json:"content_length,omitempty"`
	StatusCode    *int                   `db:"status_code" json:"status_code,omitempty"`
	ContentType   *string                `db:"content_type" json:"content_type,omitempty"`
	Links         pq.StringArray         `db:"links" json:"links"`
	Depth         int                    `db:"depth" json:"depth"`
	Metadata      map[string]interface{} `db:"metadata" json:"metadata"`
	CrawledAt     time.Time              `db:"crawled_at" json:"crawled_at"`
}

// Entity represents a knowledge entity
type Entity struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	KnowledgeBaseID uuid.UUID              `db:"knowledge_base_id" json:"knowledge_base_id"`
	Name            string                 `db:"name" json:"name"`
	Type            string                 `db:"type" json:"type"`
	Description     *string                `db:"description" json:"description,omitempty"`
	Properties      map[string]interface{} `db:"properties" json:"properties"`
	Aliases         pq.StringArray         `db:"aliases" json:"aliases"`
	Embedding       []float64              `db:"embedding" json:"embedding,omitempty"`
	Confidence      float64                `db:"confidence" json:"confidence"`
	Source          *string                `db:"source" json:"source,omitempty"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

// EntityRelationship represents a relationship between entities
type EntityRelationship struct {
	ID               uuid.UUID              `db:"id" json:"id"`
	KnowledgeBaseID  uuid.UUID              `db:"knowledge_base_id" json:"knowledge_base_id"`
	SourceEntityID   uuid.UUID              `db:"source_entity_id" json:"source_entity_id"`
	TargetEntityID   uuid.UUID              `db:"target_entity_id" json:"target_entity_id"`
	RelationshipType string                 `db:"relationship_type" json:"relationship_type"`
	Properties       map[string]interface{} `db:"properties" json:"properties"`
	Confidence       float64                `db:"confidence" json:"confidence"`
	Source           *string                `db:"source" json:"source,omitempty"`
	CreatedAt        time.Time              `db:"created_at" json:"created_at"`
}

// SearchQuery represents a search query for analytics
type SearchQuery struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	KnowledgeBaseID uuid.UUID              `db:"knowledge_base_id" json:"knowledge_base_id"`
	UserID          *uuid.UUID             `db:"user_id" json:"user_id,omitempty"`
	QueryText       string                 `db:"query_text" json:"query_text"`
	QueryType       string                 `db:"query_type" json:"query_type"`
	QueryEmbedding  []float64              `db:"query_embedding" json:"query_embedding,omitempty"`
	ResultsCount    int                    `db:"results_count" json:"results_count"`
	ResponseTimeMs  int                    `db:"response_time_ms" json:"response_time_ms"`
	Metadata        map[string]interface{} `db:"metadata" json:"metadata"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
}

// SearchCache represents cached search results
type SearchCache struct {
	ID              uuid.UUID              `db:"id" json:"id"`
	QueryHash       string                 `db:"query_hash" json:"query_hash"`
	KnowledgeBaseID uuid.UUID              `db:"knowledge_base_id" json:"knowledge_base_id"`
	QueryText       string                 `db:"query_text" json:"query_text"`
	Results         map[string]interface{} `db:"results" json:"results"`
	HitCount        int                    `db:"hit_count" json:"hit_count"`
	ExpiresAt       time.Time              `db:"expires_at" json:"expires_at"`
	CreatedAt       time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time              `db:"updated_at" json:"updated_at"`
}

// Implement database/sql/driver interfaces for custom types

// Value implements the driver.Valuer interface for KnowledgeBaseConfig
func (c KnowledgeBaseConfig) Value() (driver.Value, error) {
	return json.Marshal(c)
}

// Scan implements the sql.Scanner interface for KnowledgeBaseConfig
func (c *KnowledgeBaseConfig) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into KnowledgeBaseConfig", value)
	}

	return json.Unmarshal(bytes, c)
}

// Value implements the driver.Valuer interface for KnowledgeBaseStats
func (s KnowledgeBaseStats) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for KnowledgeBaseStats
func (s *KnowledgeBaseStats) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into KnowledgeBaseStats", value)
	}

	return json.Unmarshal(bytes, s)
}

// MetadataMap is a custom type for metadata to implement database interfaces
type MetadataMap map[string]interface{}

// Value implements the driver.Valuer interface for MetadataMap
func (m MetadataMap) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for MetadataMap
func (m *MetadataMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into MetadataMap", value)
	}

	return json.Unmarshal(bytes, m)
}
