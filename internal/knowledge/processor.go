package knowledge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/aios/aios/pkg/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Helper function to create int64 pointer
func int64Ptr(i int64) *int64 {
	return &i
}

// DocumentProcessor handles document processing operations
type DocumentProcessor struct {
	config     *config.Config
	logger     *logrus.Logger
	tracer     trace.Tracer
	repository *Repository
	chunker    *TextChunker
}

// TextChunker handles text chunking operations
type TextChunker struct {
	maxChunkSize      int
	overlapSize       int
	preserveSentences bool
}

// NewDocumentProcessor creates a new document processor instance
func NewDocumentProcessor(config *config.Config, repository *Repository, logger *logrus.Logger) (*DocumentProcessor, error) {
	chunker := &TextChunker{
		maxChunkSize:      1000,
		overlapSize:       200,
		preserveSentences: true,
	}

	return &DocumentProcessor{
		config:     config,
		logger:     logger,
		tracer:     otel.Tracer("knowledge.processor"),
		repository: repository,
		chunker:    chunker,
	}, nil
}

// Start starts the document processor
func (p *DocumentProcessor) Start(ctx context.Context) error {
	p.logger.Info("Starting Document Processor...")
	return nil
}

// Stop stops the document processor
func (p *DocumentProcessor) Stop(ctx context.Context) error {
	p.logger.Info("Stopping Document Processor...")
	return nil
}

// ProcessDocument processes a document upload request
func (p *DocumentProcessor) ProcessDocument(ctx context.Context, req *DocumentUploadRequest, knowledgeBaseID uuid.UUID) (string, error) {
	ctx, span := p.tracer.Start(ctx, "processor.process_document")
	defer span.End()

	// Generate document ID
	docID := uuid.New()

	// Calculate content hash
	hash := sha256.Sum256([]byte(req.Content))
	hashStr := hex.EncodeToString(hash[:])

	// Convert metadata
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create document
	doc := &Document{
		ID:               docID,
		KnowledgeBaseID:  knowledgeBaseID,
		Title:            req.FileName,
		Content:          req.Content,
		ContentType:      "text",
		Language:         "en",
		MimeType:         &req.MimeType,
		Metadata:         metadata,
		ProcessingStatus: "processing",
		Version:          1,
		FileSize:         int64Ptr(int64(len(req.Content))),
		FileHash:         &hashStr,
	}

	// Process content based on MIME type
	processedContent, err := p.processContentByType(req.Content, req.MimeType)
	if err != nil {
		return "", fmt.Errorf("failed to process content: %w", err)
	}

	doc.Content = processedContent

	// Chunk the document
	chunks, err := p.chunker.ChunkText(processedContent)
	if err != nil {
		return "", fmt.Errorf("failed to chunk document: %w", err)
	}

	// Store document in database
	if err := p.repository.CreateDocument(ctx, doc); err != nil {
		return "", fmt.Errorf("failed to store document: %w", err)
	}

	// Create and store document chunks
	for i, chunk := range chunks {
		chunkID := uuid.New()
		chunkMetadata := map[string]interface{}{
			"document_id": docID.String(),
			"chunk_index": i,
		}

		dbChunk := &DocumentChunk{
			ID:            chunkID,
			DocumentID:    docID,
			ChunkIndex:    i,
			Content:       chunk,
			ContentLength: len(chunk),
			ChunkType:     "text",
			Metadata:      chunkMetadata,
		}

		if err := p.repository.CreateDocumentChunk(ctx, dbChunk); err != nil {
			p.logger.WithError(err).WithField("chunk_index", i).Warn("Failed to store chunk")
		}
	}

	// Update document status
	now := time.Now()
	doc.ProcessingStatus = "completed"
	doc.ProcessedAt = &now
	if err := p.repository.UpdateDocument(ctx, doc); err != nil {
		p.logger.WithError(err).Warn("Failed to update document status")
	}

	p.logger.WithFields(logrus.Fields{
		"document_id": docID.String(),
		"file_name":   req.FileName,
		"chunks":      len(chunks),
		"size":        len(req.Content),
	}).Info("Document processed successfully")

	return docID.String(), nil
}

// processContentByType processes content based on MIME type
func (p *DocumentProcessor) processContentByType(content, mimeType string) (string, error) {
	switch mimeType {
	case "text/plain":
		return content, nil
	case "text/html":
		return p.extractTextFromHTML(content), nil
	case "text/markdown":
		return p.processMarkdown(content), nil
	case "application/json":
		return p.formatJSON(content), nil
	default:
		// Treat as plain text
		return content, nil
	}
}

// extractTextFromHTML extracts text content from HTML
func (p *DocumentProcessor) extractTextFromHTML(html string) string {
	// Simple HTML text extraction
	// In a real implementation, you'd use a proper HTML parser
	text := html

	// Remove script and style tags
	text = removeHTMLTags(text, "script")
	text = removeHTMLTags(text, "style")

	// Remove all HTML tags
	text = removeAllHTMLTags(text)

	// Clean up whitespace
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

// processMarkdown processes markdown content
func (p *DocumentProcessor) processMarkdown(markdown string) string {
	// Simple markdown processing
	// Remove markdown syntax while preserving content structure
	text := markdown

	// Remove headers but keep content
	lines := strings.Split(text, "\n")
	var processed []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			// Remove # symbols but keep header text
			line = strings.TrimLeft(line, "# ")
		}
		if line != "" {
			processed = append(processed, line)
		}
	}

	return strings.Join(processed, " ")
}

// formatJSON formats JSON content for better readability
func (p *DocumentProcessor) formatJSON(jsonContent string) string {
	// Simple JSON formatting - in practice you'd use proper JSON parsing
	return jsonContent
}

// removeHTMLTags removes specific HTML tags and their content
func removeHTMLTags(html, tag string) string {
	startTag := fmt.Sprintf("<%s", tag)
	endTag := fmt.Sprintf("</%s>", tag)

	for {
		start := strings.Index(strings.ToLower(html), startTag)
		if start == -1 {
			break
		}

		// Find the end of the opening tag
		tagEnd := strings.Index(html[start:], ">")
		if tagEnd == -1 {
			break
		}
		tagEnd += start + 1

		// Find the closing tag
		end := strings.Index(strings.ToLower(html[tagEnd:]), endTag)
		if end == -1 {
			break
		}
		end += tagEnd + len(endTag)

		// Remove the entire tag and its content
		html = html[:start] + html[end:]
	}

	return html
}

// removeAllHTMLTags removes all HTML tags
func removeAllHTMLTags(html string) string {
	result := ""
	inTag := false

	for _, char := range html {
		if char == '<' {
			inTag = true
		} else if char == '>' {
			inTag = false
		} else if !inTag {
			result += string(char)
		}
	}

	return result
}

// ChunkText splits text into chunks
func (c *TextChunker) ChunkText(text string) ([]string, error) {
	if len(text) <= c.maxChunkSize {
		return []string{text}, nil
	}

	var chunks []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return []string{}, nil
	}

	currentChunk := ""

	for _, word := range words {
		// Check if adding this word would exceed the chunk size
		testChunk := currentChunk
		if testChunk != "" {
			testChunk += " "
		}
		testChunk += word

		if len(testChunk) > c.maxChunkSize && currentChunk != "" {
			// Save current chunk and start a new one
			chunks = append(chunks, strings.TrimSpace(currentChunk))

			// Start new chunk with overlap if configured
			if c.overlapSize > 0 && len(currentChunk) > c.overlapSize {
				overlapWords := strings.Fields(currentChunk)
				overlapStart := len(overlapWords) - (c.overlapSize / 10) // Approximate word count for overlap
				if overlapStart < 0 {
					overlapStart = 0
				}
				currentChunk = strings.Join(overlapWords[overlapStart:], " ") + " " + word
			} else {
				currentChunk = word
			}
		} else {
			currentChunk = testChunk
		}
	}

	// Add the last chunk if it's not empty
	if strings.TrimSpace(currentChunk) != "" {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	return chunks, nil
}

// GetDocument returns a document by ID
func (p *DocumentProcessor) GetDocument(ctx context.Context, docID string) (*Document, error) {
	id, err := uuid.Parse(docID)
	if err != nil {
		return nil, fmt.Errorf("invalid document ID: %w", err)
	}

	return p.repository.GetDocument(ctx, id)
}

// ListDocuments returns a list of all processed documents
func (p *DocumentProcessor) ListDocuments(ctx context.Context, knowledgeBaseID uuid.UUID, limit, offset int) ([]*Document, error) {
	return p.repository.ListDocuments(ctx, knowledgeBaseID, limit, offset)
}

// DeleteDocument removes a document
func (p *DocumentProcessor) DeleteDocument(ctx context.Context, docID string) error {
	id, err := uuid.Parse(docID)
	if err != nil {
		return fmt.Errorf("invalid document ID: %w", err)
	}

	err = p.repository.DeleteDocument(ctx, id)
	if err != nil {
		return err
	}

	p.logger.WithField("document_id", docID).Info("Document deleted")
	return nil
}
