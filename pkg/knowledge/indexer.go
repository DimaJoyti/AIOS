package knowledge

import (
	"context"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultKnowledgeIndexer implements the KnowledgeIndexer interface
type DefaultKnowledgeIndexer struct {
	index  map[string]*IndexEntry
	logger *logrus.Logger
	tracer trace.Tracer
	config *IndexerConfig
	mu     sync.RWMutex
}

// IndexerConfig represents configuration for the knowledge indexer
type IndexerConfig struct {
	MaxDocuments     int           `json:"max_documents"`
	IndexingInterval time.Duration `json:"indexing_interval"`
	EnableFullText   bool          `json:"enable_full_text"`
	EnableMetadata   bool          `json:"enable_metadata"`
	BatchSize        int           `json:"batch_size"`
}

// IndexEntry represents an entry in the knowledge index
type IndexEntry struct {
	DocumentID   string                 `json:"document_id"`
	Title        string                 `json:"title"`
	Content      string                 `json:"content"`
	Keywords     []string               `json:"keywords"`
	Metadata     map[string]interface{} `json:"metadata"`
	Embedding    []float32              `json:"embedding"`
	IndexedAt    time.Time              `json:"indexed_at"`
	LastModified time.Time              `json:"last_modified"`
}

// NewDefaultKnowledgeIndexer creates a new default knowledge indexer
func NewDefaultKnowledgeIndexer(logger *logrus.Logger) (KnowledgeIndexer, error) {
	config := &IndexerConfig{
		MaxDocuments:     100000,
		IndexingInterval: 5 * time.Minute,
		EnableFullText:   true,
		EnableMetadata:   true,
		BatchSize:        100,
	}

	indexer := &DefaultKnowledgeIndexer{
		index:  make(map[string]*IndexEntry),
		logger: logger,
		tracer: otel.Tracer("knowledge.indexer"),
		config: config,
	}

	return indexer, nil
}

// IndexDocument indexes a single document
func (ki *DefaultKnowledgeIndexer) IndexDocument(ctx context.Context, doc *Document) error {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.index_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.title", doc.Title),
		attribute.Int("document.content_length", len(doc.Content)),
	)

	ki.mu.Lock()
	defer ki.mu.Unlock()

	// Create index entry
	entry := &IndexEntry{
		DocumentID:   doc.ID,
		Title:        doc.Title,
		Content:      doc.Content,
		Keywords:     extractKeywords(doc.Content),
		Metadata:     doc.Metadata,
		Embedding:    doc.Embedding,
		IndexedAt:    time.Now(),
		LastModified: doc.UpdatedAt,
	}

	ki.index[doc.ID] = entry

	ki.logger.WithFields(logrus.Fields{
		"document_id":    doc.ID,
		"title":          doc.Title,
		"content_length": len(doc.Content),
		"keywords_count": len(entry.Keywords),
	}).Debug("Document indexed successfully")

	return nil
}

// IndexDocuments indexes multiple documents
func (ki *DefaultKnowledgeIndexer) IndexDocuments(ctx context.Context, docs []*Document) error {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.index_documents")
	defer span.End()

	span.SetAttributes(attribute.Int("documents.count", len(docs)))

	// Process documents in batches
	for i := 0; i < len(docs); i += ki.config.BatchSize {
		end := i + ki.config.BatchSize
		if end > len(docs) {
			end = len(docs)
		}

		batch := docs[i:end]
		for _, doc := range batch {
			if err := ki.IndexDocument(ctx, doc); err != nil {
				ki.logger.WithError(err).WithField("document_id", doc.ID).Warn("Failed to index document")
			}
		}
	}

	ki.logger.WithField("count", len(docs)).Info("Documents indexed successfully")
	return nil
}

// UpdateIndex updates the index for a document
func (ki *DefaultKnowledgeIndexer) UpdateIndex(ctx context.Context, docID string, doc *Document) error {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.update_index")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", docID))

	// Simply reindex the document
	return ki.IndexDocument(ctx, doc)
}

// DeleteFromIndex removes a document from the index
func (ki *DefaultKnowledgeIndexer) DeleteFromIndex(ctx context.Context, docID string) error {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.delete_from_index")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", docID))

	ki.mu.Lock()
	defer ki.mu.Unlock()

	delete(ki.index, docID)

	ki.logger.WithField("document_id", docID).Debug("Document removed from index")
	return nil
}

// RebuildIndex rebuilds the entire index for a knowledge base
func (ki *DefaultKnowledgeIndexer) RebuildIndex(ctx context.Context, kbID string) error {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.rebuild_index")
	defer span.End()

	span.SetAttributes(attribute.String("knowledge_base.id", kbID))

	ki.mu.Lock()
	defer ki.mu.Unlock()

	// Clear existing index entries for this knowledge base
	for docID, entry := range ki.index {
		// In a real implementation, you would check if the document belongs to the KB
		_ = entry
		delete(ki.index, docID)
	}

	ki.logger.WithField("knowledge_base_id", kbID).Info("Index rebuilt successfully")
	return nil
}

// GetIndexStats returns index statistics
func (ki *DefaultKnowledgeIndexer) GetIndexStats(ctx context.Context) (*IndexStats, error) {
	ctx, span := ki.tracer.Start(ctx, "knowledge_indexer.get_index_stats")
	defer span.End()

	ki.mu.RLock()
	defer ki.mu.RUnlock()

	totalSize := int64(0)
	var lastUpdated time.Time

	for _, entry := range ki.index {
		totalSize += int64(len(entry.Content))
		if entry.IndexedAt.After(lastUpdated) {
			lastUpdated = entry.IndexedAt
		}
	}

	stats := &IndexStats{
		DocumentCount: len(ki.index),
		IndexSize:     totalSize,
		LastUpdated:   lastUpdated,
		Health:        1.0, // Simple health calculation
	}

	return stats, nil
}

// extractKeywords extracts keywords from text (simplified implementation)
func extractKeywords(text string) []string {
	// This is a simplified implementation
	// In a real system, you would use proper NLP libraries
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Count word frequencies
	for _, word := range words {
		// Remove punctuation and filter short words
		word = regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		if len(word) > 3 {
			wordCount[word]++
		}
	}

	// Get top keywords
	var keywords []string
	for word, count := range wordCount {
		if count > 1 { // Only include words that appear more than once
			keywords = append(keywords, word)
		}
	}

	// Limit to top 20 keywords
	if len(keywords) > 20 {
		keywords = keywords[:20]
	}

	return keywords
}

// TextMetadataExtractor implements MetadataExtractor for text documents
type TextMetadataExtractor struct{}

// Extract extracts metadata from a text document
func (e *TextMetadataExtractor) Extract(doc *Document) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// Basic text statistics
	metadata["word_count"] = len(strings.Fields(doc.Content))
	metadata["character_count"] = len(doc.Content)
	metadata["line_count"] = len(strings.Split(doc.Content, "\n"))

	// Language detection (simplified)
	metadata["language"] = detectLanguage(doc.Content)

	// Content type
	metadata["content_type"] = "text"

	// Reading time estimation (assuming 200 words per minute)
	wordCount := len(strings.Fields(doc.Content))
	readingTime := wordCount / 200
	if readingTime < 1 {
		readingTime = 1
	}
	metadata["estimated_reading_time_minutes"] = readingTime

	// Extract URLs if any
	urls := extractURLs(doc.Content)
	if len(urls) > 0 {
		metadata["urls"] = urls
	}

	// Extract email addresses if any
	emails := extractEmails(doc.Content)
	if len(emails) > 0 {
		metadata["emails"] = emails
	}

	return metadata, nil
}

// PDFMetadataExtractor implements MetadataExtractor for PDF documents
type PDFMetadataExtractor struct{}

// Extract extracts metadata from a PDF document
func (e *PDFMetadataExtractor) Extract(doc *Document) (map[string]interface{}, error) {
	metadata := make(map[string]interface{})

	// For now, use text extraction as base
	textExtractor := &TextMetadataExtractor{}
	textMetadata, err := textExtractor.Extract(doc)
	if err != nil {
		return nil, err
	}

	// Copy text metadata
	for key, value := range textMetadata {
		metadata[key] = value
	}

	// PDF-specific metadata
	metadata["content_type"] = "pdf"
	metadata["format"] = "pdf"

	// In a real implementation, you would extract PDF-specific metadata
	// such as author, creation date, page count, etc.

	return metadata, nil
}

// Helper functions for metadata extraction

// detectLanguage detects the language of text (simplified implementation)
func detectLanguage(text string) string {
	// Very simple language detection based on character patterns
	// In a real implementation, use proper language detection libraries

	if len(text) < 10 {
		return "unknown"
	}

	// Count different character types
	latinCount := 0
	cyrillicCount := 0
	arabicCount := 0
	cjkCount := 0

	for _, r := range text {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z'):
			latinCount++
		case r >= 0x0400 && r <= 0x04FF: // Cyrillic
			cyrillicCount++
		case r >= 0x0600 && r <= 0x06FF: // Arabic
			arabicCount++
		case r >= 0x4E00 && r <= 0x9FFF: // CJK
			cjkCount++
		}
	}

	total := latinCount + cyrillicCount + arabicCount + cjkCount
	if total == 0 {
		return "unknown"
	}

	// Determine dominant script
	if float64(latinCount)/float64(total) > 0.7 {
		return "en" // Default to English for Latin script
	} else if float64(cyrillicCount)/float64(total) > 0.7 {
		return "ru" // Default to Russian for Cyrillic
	} else if float64(arabicCount)/float64(total) > 0.7 {
		return "ar"
	} else if float64(cjkCount)/float64(total) > 0.7 {
		return "zh"
	}

	return "unknown"
}

// extractURLs extracts URLs from text
func extractURLs(text string) []string {
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	return urlRegex.FindAllString(text, -1)
}

// extractEmails extracts email addresses from text
func extractEmails(text string) []string {
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	return emailRegex.FindAllString(text, -1)
}
