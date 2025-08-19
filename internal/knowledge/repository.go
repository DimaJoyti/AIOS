package knowledge

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Repository provides database operations for knowledge management
type Repository struct {
	db     *sqlx.DB
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewRepository creates a new knowledge repository
func NewRepository(db *sqlx.DB, logger *logrus.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
		tracer: otel.Tracer("knowledge.repository"),
	}
}

// Knowledge Base operations

// CreateKnowledgeBase creates a new knowledge base
func (r *Repository) CreateKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateKnowledgeBase")
	defer span.End()

	query := `
		INSERT INTO knowledge.knowledge_bases (id, name, description, owner_id, config, stats, metadata, status)
		VALUES (:id, :name, :description, :owner_id, :config, :stats, :metadata, :status)
	`

	_, err := r.db.NamedExecContext(ctx, query, kb)
	if err != nil {
		r.logger.WithError(err).Error("Failed to create knowledge base")
		return fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return nil
}

// GetKnowledgeBase retrieves a knowledge base by ID
func (r *Repository) GetKnowledgeBase(ctx context.Context, id uuid.UUID) (*KnowledgeBase, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetKnowledgeBase")
	defer span.End()

	var kb KnowledgeBase
	query := `SELECT * FROM knowledge.knowledge_bases WHERE id = $1`

	err := r.db.GetContext(ctx, &kb, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("knowledge base not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get knowledge base")
		return nil, fmt.Errorf("failed to get knowledge base: %w", err)
	}

	return &kb, nil
}

// ListKnowledgeBases retrieves all knowledge bases for a user
func (r *Repository) ListKnowledgeBases(ctx context.Context, ownerID uuid.UUID) ([]*KnowledgeBase, error) {
	ctx, span := r.tracer.Start(ctx, "repository.ListKnowledgeBases")
	defer span.End()

	var kbs []*KnowledgeBase
	query := `SELECT * FROM knowledge.knowledge_bases WHERE owner_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &kbs, query, ownerID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list knowledge bases")
		return nil, fmt.Errorf("failed to list knowledge bases: %w", err)
	}

	return kbs, nil
}

// UpdateKnowledgeBase updates a knowledge base
func (r *Repository) UpdateKnowledgeBase(ctx context.Context, kb *KnowledgeBase) error {
	ctx, span := r.tracer.Start(ctx, "repository.UpdateKnowledgeBase")
	defer span.End()

	query := `
		UPDATE knowledge.knowledge_bases 
		SET name = :name, description = :description, config = :config, stats = :stats, 
		    metadata = :metadata, status = :status, updated_at = NOW()
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, kb)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update knowledge base")
		return fmt.Errorf("failed to update knowledge base: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge base not found: %s", kb.ID)
	}

	return nil
}

// DeleteKnowledgeBase deletes a knowledge base
func (r *Repository) DeleteKnowledgeBase(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "repository.DeleteKnowledgeBase")
	defer span.End()

	query := `DELETE FROM knowledge.knowledge_bases WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete knowledge base")
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("knowledge base not found: %s", id)
	}

	return nil
}

// Document operations

// CreateDocument creates a new document
func (r *Repository) CreateDocument(ctx context.Context, doc *Document) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateDocument")
	defer span.End()

	query := `
		INSERT INTO knowledge.documents (
			id, knowledge_base_id, title, content, content_type, language, source, url, author,
			file_path, file_size, file_hash, mime_type, tags, categories, metadata, 
			processing_status, version, parent_id
		) VALUES (
			:id, :knowledge_base_id, :title, :content, :content_type, :language, :source, :url, :author,
			:file_path, :file_size, :file_hash, :mime_type, :tags, :categories, :metadata,
			:processing_status, :version, :parent_id
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, doc)
	if err != nil {
		r.logger.WithError(err).Error("Failed to create document")
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// GetDocument retrieves a document by ID
func (r *Repository) GetDocument(ctx context.Context, id uuid.UUID) (*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetDocument")
	defer span.End()

	var doc Document
	query := `SELECT * FROM knowledge.documents WHERE id = $1`

	err := r.db.GetContext(ctx, &doc, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("document not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get document")
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return &doc, nil
}

// ListDocuments retrieves documents for a knowledge base
func (r *Repository) ListDocuments(ctx context.Context, knowledgeBaseID uuid.UUID, limit, offset int) ([]*Document, error) {
	ctx, span := r.tracer.Start(ctx, "repository.ListDocuments")
	defer span.End()

	var docs []*Document
	query := `
		SELECT * FROM knowledge.documents 
		WHERE knowledge_base_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &docs, query, knowledgeBaseID, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list documents")
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	return docs, nil
}

// UpdateDocument updates a document
func (r *Repository) UpdateDocument(ctx context.Context, doc *Document) error {
	ctx, span := r.tracer.Start(ctx, "repository.UpdateDocument")
	defer span.End()

	query := `
		UPDATE knowledge.documents 
		SET title = :title, content = :content, content_type = :content_type, language = :language,
		    source = :source, url = :url, author = :author, file_path = :file_path, file_size = :file_size,
		    file_hash = :file_hash, mime_type = :mime_type, tags = :tags, categories = :categories,
		    metadata = :metadata, processing_status = :processing_status, processing_error = :processing_error,
		    version = :version, processed_at = :processed_at, updated_at = NOW()
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, doc)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update document")
		return fmt.Errorf("failed to update document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", doc.ID)
	}

	return nil
}

// DeleteDocument deletes a document
func (r *Repository) DeleteDocument(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "repository.DeleteDocument")
	defer span.End()

	query := `DELETE FROM knowledge.documents WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.WithError(err).Error("Failed to delete document")
		return fmt.Errorf("failed to delete document: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("document not found: %s", id)
	}

	return nil
}

// Document Chunk operations

// CreateDocumentChunk creates a new document chunk
func (r *Repository) CreateDocumentChunk(ctx context.Context, chunk *DocumentChunk) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateDocumentChunk")
	defer span.End()

	query := `
		INSERT INTO knowledge.document_chunks (
			id, document_id, chunk_index, content, content_length, chunk_type,
			start_position, end_position, metadata, embedding, embedding_model
		) VALUES (
			:id, :document_id, :chunk_index, :content, :content_length, :chunk_type,
			:start_position, :end_position, :metadata, :embedding, :embedding_model
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, chunk)
	if err != nil {
		r.logger.WithError(err).Error("Failed to create document chunk")
		return fmt.Errorf("failed to create document chunk: %w", err)
	}

	return nil
}

// GetDocumentChunks retrieves chunks for a document
func (r *Repository) GetDocumentChunks(ctx context.Context, documentID uuid.UUID) ([]*DocumentChunk, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetDocumentChunks")
	defer span.End()

	var chunks []*DocumentChunk
	query := `
		SELECT * FROM knowledge.document_chunks 
		WHERE document_id = $1 
		ORDER BY chunk_index
	`

	err := r.db.SelectContext(ctx, &chunks, query, documentID)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get document chunks")
		return nil, fmt.Errorf("failed to get document chunks: %w", err)
	}

	return chunks, nil
}

// SearchDocumentChunks performs vector similarity search on document chunks
func (r *Repository) SearchDocumentChunks(ctx context.Context, knowledgeBaseID uuid.UUID, embedding []float64, limit int) ([]*DocumentChunk, error) {
	ctx, span := r.tracer.Start(ctx, "repository.SearchDocumentChunks")
	defer span.End()

	var chunks []*DocumentChunk
	query := `
		SELECT dc.* FROM knowledge.document_chunks dc
		JOIN knowledge.documents d ON dc.document_id = d.id
		WHERE d.knowledge_base_id = $1 AND dc.embedding IS NOT NULL
		ORDER BY dc.embedding <-> $2::vector
		LIMIT $3
	`

	// Convert embedding to PostgreSQL vector format
	embeddingStr := fmt.Sprintf("[%v]", embedding)

	err := r.db.SelectContext(ctx, &chunks, query, knowledgeBaseID, embeddingStr, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to search document chunks")
		return nil, fmt.Errorf("failed to search document chunks: %w", err)
	}

	return chunks, nil
}

// Crawl Job operations

// CreateCrawlJob creates a new crawl job
func (r *Repository) CreateCrawlJob(ctx context.Context, job *CrawlJob) error {
	ctx, span := r.tracer.Start(ctx, "repository.CreateCrawlJob")
	defer span.End()

	query := `
		INSERT INTO knowledge.crawl_jobs (
			id, knowledge_base_id, url, status, max_pages, max_depth, follow_links, metadata
		) VALUES (
			:id, :knowledge_base_id, :url, :status, :max_pages, :max_depth, :follow_links, :metadata
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, job)
	if err != nil {
		r.logger.WithError(err).Error("Failed to create crawl job")
		return fmt.Errorf("failed to create crawl job: %w", err)
	}

	return nil
}

// UpdateCrawlJob updates a crawl job
func (r *Repository) UpdateCrawlJob(ctx context.Context, job *CrawlJob) error {
	ctx, span := r.tracer.Start(ctx, "repository.UpdateCrawlJob")
	defer span.End()

	query := `
		UPDATE knowledge.crawl_jobs 
		SET status = :status, pages_found = :pages_found, pages_processed = :pages_processed,
		    error_message = :error_message, metadata = :metadata, started_at = :started_at,
		    completed_at = :completed_at, updated_at = NOW()
		WHERE id = :id
	`

	_, err := r.db.NamedExecContext(ctx, query, job)
	if err != nil {
		r.logger.WithError(err).Error("Failed to update crawl job")
		return fmt.Errorf("failed to update crawl job: %w", err)
	}

	return nil
}

// GetCrawlJob retrieves a crawl job by ID
func (r *Repository) GetCrawlJob(ctx context.Context, id uuid.UUID) (*CrawlJob, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetCrawlJob")
	defer span.End()

	var job CrawlJob
	query := `SELECT * FROM knowledge.crawl_jobs WHERE id = $1`

	err := r.db.GetContext(ctx, &job, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("crawl job not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get crawl job")
		return nil, fmt.Errorf("failed to get crawl job: %w", err)
	}

	return &job, nil
}

// Search Cache operations

// GetSearchCache retrieves cached search results
func (r *Repository) GetSearchCache(ctx context.Context, queryHash string, knowledgeBaseID uuid.UUID) (*SearchCache, error) {
	ctx, span := r.tracer.Start(ctx, "repository.GetSearchCache")
	defer span.End()

	var cache SearchCache
	query := `
		SELECT * FROM knowledge.search_cache 
		WHERE query_hash = $1 AND knowledge_base_id = $2 AND expires_at > NOW()
	`

	err := r.db.GetContext(ctx, &cache, query, queryHash, knowledgeBaseID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Cache miss
		}
		r.logger.WithError(err).Error("Failed to get search cache")
		return nil, fmt.Errorf("failed to get search cache: %w", err)
	}

	// Update hit count
	go func() {
		updateQuery := `UPDATE knowledge.search_cache SET hit_count = hit_count + 1, updated_at = NOW() WHERE id = $1`
		r.db.ExecContext(context.Background(), updateQuery, cache.ID)
	}()

	return &cache, nil
}

// SetSearchCache stores search results in cache
func (r *Repository) SetSearchCache(ctx context.Context, cache *SearchCache) error {
	ctx, span := r.tracer.Start(ctx, "repository.SetSearchCache")
	defer span.End()

	query := `
		INSERT INTO knowledge.search_cache (
			id, query_hash, knowledge_base_id, query_text, results, expires_at
		) VALUES (
			:id, :query_hash, :knowledge_base_id, :query_text, :results, :expires_at
		)
		ON CONFLICT (query_hash, knowledge_base_id) 
		DO UPDATE SET results = EXCLUDED.results, expires_at = EXCLUDED.expires_at, updated_at = NOW()
	`

	_, err := r.db.NamedExecContext(ctx, query, cache)
	if err != nil {
		r.logger.WithError(err).Error("Failed to set search cache")
		return fmt.Errorf("failed to set search cache: %w", err)
	}

	return nil
}

// CleanupExpiredCache removes expired cache entries
func (r *Repository) CleanupExpiredCache(ctx context.Context) error {
	ctx, span := r.tracer.Start(ctx, "repository.CleanupExpiredCache")
	defer span.End()

	query := `DELETE FROM knowledge.search_cache WHERE expires_at < NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		r.logger.WithError(err).Error("Failed to cleanup expired cache")
		return fmt.Errorf("failed to cleanup expired cache: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		r.logger.WithField("rows_deleted", rowsAffected).Info("Cleaned up expired cache entries")
	}

	return nil
}
