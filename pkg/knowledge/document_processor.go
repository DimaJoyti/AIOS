package knowledge

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultDocumentProcessor implements the DocumentProcessor interface
type DefaultDocumentProcessor struct {
	logger     *logrus.Logger
	tracer     trace.Tracer
	config     *DocumentProcessorConfig
	chunkers   map[ChunkingStrategy]TextChunker
	extractors map[string]MetadataExtractor
}

// DocumentProcessorConfig represents configuration for document processing
type DocumentProcessorConfig struct {
	DefaultChunkSize        int              `json:"default_chunk_size"`
	DefaultChunkOverlap     int              `json:"default_chunk_overlap"`
	DefaultStrategy         ChunkingStrategy `json:"default_strategy"`
	EnableCleaning          bool             `json:"enable_cleaning"`
	EnableLanguageDetection bool             `json:"enable_language_detection"`
	EnableEntityExtraction  bool             `json:"enable_entity_extraction"`
	EnableKeywordExtraction bool             `json:"enable_keyword_extraction"`
	EnableSummarization     bool             `json:"enable_summarization"`
}

// TextChunker handles text chunking
type TextChunker interface {
	Chunk(text string, chunkSize int, overlap int) ([]*DocumentChunk, error)
}

// MetadataExtractor extracts metadata from documents
type MetadataExtractor interface {
	Extract(doc *Document) (map[string]interface{}, error)
}

// NewDefaultDocumentProcessor creates a new default document processor
func NewDefaultDocumentProcessor(logger *logrus.Logger) (DocumentProcessor, error) {
	config := &DocumentProcessorConfig{
		DefaultChunkSize:        1000,
		DefaultChunkOverlap:     200,
		DefaultStrategy:         ChunkingStrategyRecursive,
		EnableCleaning:          true,
		EnableLanguageDetection: true,
		EnableEntityExtraction:  true,
		EnableKeywordExtraction: true,
		EnableSummarization:     false,
	}

	processor := &DefaultDocumentProcessor{
		logger:     logger,
		tracer:     otel.Tracer("knowledge.document_processor"),
		config:     config,
		chunkers:   make(map[ChunkingStrategy]TextChunker),
		extractors: make(map[string]MetadataExtractor),
	}

	// Initialize chunkers
	processor.chunkers[ChunkingStrategyFixed] = &FixedSizeChunker{}
	processor.chunkers[ChunkingStrategySentence] = &SentenceChunker{}
	processor.chunkers[ChunkingStrategyParagraph] = &ParagraphChunker{}
	processor.chunkers[ChunkingStrategyRecursive] = &RecursiveChunker{}
	processor.chunkers[ChunkingStrategySemantic] = &SemanticChunker{}

	// Initialize metadata extractors
	processor.extractors["text"] = &TextMetadataExtractor{}
	processor.extractors["pdf"] = &PDFMetadataExtractor{}

	return processor, nil
}

// ProcessDocument processes a document and returns processed result
func (dp *DefaultDocumentProcessor) ProcessDocument(ctx context.Context, doc *Document) (*ProcessedDocument, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_document")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
		attribute.Int("document.content_length", len(doc.Content)),
	)

	// Clean text if enabled
	cleanedContent := doc.Content
	if dp.config.EnableCleaning {
		cleanedContent = dp.cleanText(cleanedContent)
	}

	// Detect language if enabled
	language := doc.Language
	if dp.config.EnableLanguageDetection && language == "" {
		detectedLang, err := dp.DetectLanguage(ctx, cleanedContent)
		if err != nil {
			dp.logger.WithError(err).Warn("Failed to detect language")
		} else {
			language = detectedLang
		}
	}

	// Chunk document
	chunks, err := dp.ChunkDocument(ctx, doc, dp.config.DefaultStrategy)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}

	// Extract metadata
	metadata, err := dp.ExtractMetadata(ctx, doc)
	if err != nil {
		dp.logger.WithError(err).Warn("Failed to extract metadata")
		metadata = make(map[string]interface{})
	}

	// Extract entities if enabled
	var entities []*Entity
	if dp.config.EnableEntityExtraction {
		entities, err = dp.extractEntities(ctx, cleanedContent)
		if err != nil {
			dp.logger.WithError(err).Warn("Failed to extract entities")
		}
	}

	// Extract keywords if enabled
	var keywords []string
	if dp.config.EnableKeywordExtraction {
		keywords = dp.extractKeywords(cleanedContent)
	}

	// Generate summary if enabled
	var summary string
	if dp.config.EnableSummarization {
		summary = dp.generateSummary(cleanedContent)
	}

	// Analyze sentiment
	sentiment := dp.analyzeSentiment(cleanedContent)

	// Extract topics
	topics := dp.extractTopics(cleanedContent)

	// Update document with extracted metadata
	if doc.Metadata == nil {
		doc.Metadata = make(map[string]interface{})
	}
	for key, value := range metadata {
		doc.Metadata[key] = value
	}

	processedDoc := &ProcessedDocument{
		Document:    doc,
		Chunks:      chunks,
		Entities:    entities,
		Keywords:    keywords,
		Summary:     summary,
		Language:    language,
		Sentiment:   sentiment,
		Topics:      topics,
		ProcessedAt: time.Now(),
	}

	dp.logger.WithFields(logrus.Fields{
		"document_id":     doc.ID,
		"chunks_count":    len(chunks),
		"entities_count":  len(entities),
		"keywords_count":  len(keywords),
		"language":        language,
		"processing_time": time.Since(startTime),
	}).Info("Document processed successfully")

	return processedDoc, nil
}

// ChunkDocument chunks a document using the specified strategy
func (dp *DefaultDocumentProcessor) ChunkDocument(ctx context.Context, doc *Document, strategy ChunkingStrategy) ([]*DocumentChunk, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.chunk_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("chunking.strategy", string(strategy)),
	)

	chunker, exists := dp.chunkers[strategy]
	if !exists {
		return nil, fmt.Errorf("unsupported chunking strategy: %s", strategy)
	}

	chunks, err := chunker.Chunk(doc.Content, dp.config.DefaultChunkSize, dp.config.DefaultChunkOverlap)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to chunk document: %w", err)
	}

	// Set document ID and metadata for chunks
	for i, chunk := range chunks {
		chunk.DocumentID = doc.ID
		chunk.ChunkIndex = i
		chunk.CreatedAt = time.Now()
		if chunk.ID == "" {
			chunk.ID = uuid.New().String()
		}
	}

	return chunks, nil
}

// ExtractMetadata extracts metadata from a document
func (dp *DefaultDocumentProcessor) ExtractMetadata(ctx context.Context, doc *Document) (map[string]interface{}, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.extract_metadata")
	defer span.End()

	extractor, exists := dp.extractors[doc.ContentType]
	if !exists {
		extractor = dp.extractors["text"] // Default to text extractor
	}

	return extractor.Extract(doc)
}

// DetectLanguage detects the language of the text
func (dp *DefaultDocumentProcessor) DetectLanguage(ctx context.Context, text string) (string, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.detect_language")
	defer span.End()

	// Simple language detection based on character patterns
	// In a real implementation, you would use a proper language detection library

	// Count different character types
	latinCount := 0
	cyrillicCount := 0
	arabicCount := 0
	cjkCount := 0

	for _, r := range text {
		switch {
		case unicode.In(r, unicode.Latin):
			latinCount++
		case unicode.In(r, unicode.Cyrillic):
			cyrillicCount++
		case unicode.In(r, unicode.Arabic):
			arabicCount++
		case unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul):
			cjkCount++
		}
	}

	total := latinCount + cyrillicCount + arabicCount + cjkCount
	if total == 0 {
		return "unknown", nil
	}

	// Determine dominant script
	if float64(latinCount)/float64(total) > 0.7 {
		return "en", nil // Default to English for Latin script
	} else if float64(cyrillicCount)/float64(total) > 0.7 {
		return "ru", nil // Default to Russian for Cyrillic script
	} else if float64(arabicCount)/float64(total) > 0.7 {
		return "ar", nil // Arabic
	} else if float64(cjkCount)/float64(total) > 0.7 {
		return "zh", nil // Default to Chinese for CJK
	}

	return "unknown", nil
}

// CleanText cleans and normalizes text
func (dp *DefaultDocumentProcessor) CleanText(ctx context.Context, text string) (string, error) {
	return dp.cleanText(text), nil
}

// cleanText performs text cleaning
func (dp *DefaultDocumentProcessor) cleanText(text string) string {
	// Remove excessive whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Remove control characters
	text = regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`).ReplaceAllString(text, "")

	// Normalize quotes
	text = strings.ReplaceAll(text, "\u201c", "\"") // Left double quotation mark
	text = strings.ReplaceAll(text, "\u201d", "\"") // Right double quotation mark
	text = strings.ReplaceAll(text, "\u2018", "'")  // Left single quotation mark
	text = strings.ReplaceAll(text, "\u2019", "'")  // Right single quotation mark

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// extractEntities extracts entities from text (simplified implementation)
func (dp *DefaultDocumentProcessor) extractEntities(ctx context.Context, text string) ([]*Entity, error) {
	var entities []*Entity

	// Simple entity extraction using regex patterns
	// In a real implementation, you would use NLP libraries like spaCy or Stanford NER

	// Extract email addresses
	emailRegex := regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`)
	emails := emailRegex.FindAllString(text, -1)
	for _, email := range emails {
		entity := &Entity{
			ID:         uuid.New().String(),
			Name:       email,
			Type:       "email",
			Confidence: 0.9,
			Source:     "regex",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		entities = append(entities, entity)
	}

	// Extract URLs
	urlRegex := regexp.MustCompile(`https?://[^\s]+`)
	urls := urlRegex.FindAllString(text, -1)
	for _, url := range urls {
		entity := &Entity{
			ID:         uuid.New().String(),
			Name:       url,
			Type:       "url",
			Confidence: 0.9,
			Source:     "regex",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		entities = append(entities, entity)
	}

	// Extract phone numbers (simple pattern)
	phoneRegex := regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b|\b\(\d{3}\)\s*\d{3}-\d{4}\b`)
	phones := phoneRegex.FindAllString(text, -1)
	for _, phone := range phones {
		entity := &Entity{
			ID:         uuid.New().String(),
			Name:       phone,
			Type:       "phone",
			Confidence: 0.8,
			Source:     "regex",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

// extractKeywords extracts keywords from text (simplified implementation)
func (dp *DefaultDocumentProcessor) extractKeywords(text string) []string {
	// Simple keyword extraction based on word frequency
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// Count word frequencies
	for _, word := range words {
		// Remove punctuation
		word = regexp.MustCompile(`[^\w]`).ReplaceAllString(word, "")
		if len(word) > 3 { // Only consider words longer than 3 characters
			wordCount[word]++
		}
	}

	// Sort by frequency
	type wordFreq struct {
		word  string
		count int
	}

	var frequencies []wordFreq
	for word, count := range wordCount {
		frequencies = append(frequencies, wordFreq{word, count})
	}

	// Sort by count (descending)
	for i := 0; i < len(frequencies)-1; i++ {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[i].count < frequencies[j].count {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	// Return top keywords
	var keywords []string
	maxKeywords := 10
	if len(frequencies) < maxKeywords {
		maxKeywords = len(frequencies)
	}

	for i := 0; i < maxKeywords; i++ {
		keywords = append(keywords, frequencies[i].word)
	}

	return keywords
}

// generateSummary generates a summary of the text (simplified implementation)
func (dp *DefaultDocumentProcessor) generateSummary(text string) string {
	// Simple extractive summarization - take first few sentences
	sentences := strings.Split(text, ".")
	if len(sentences) <= 3 {
		return text
	}

	summary := strings.Join(sentences[:3], ".") + "."
	return summary
}

// analyzeSentiment analyzes sentiment of the text (simplified implementation)
func (dp *DefaultDocumentProcessor) analyzeSentiment(text string) *Sentiment {
	// Simple sentiment analysis based on keyword matching
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "like", "happy", "positive"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "hate", "dislike", "sad", "negative", "poor", "worst"}

	text = strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}

	total := positiveCount + negativeCount
	if total == 0 {
		return &Sentiment{
			Score:      0.0,
			Magnitude:  0.0,
			Label:      "neutral",
			Confidence: 0.5,
		}
	}

	score := float32(positiveCount-negativeCount) / float32(total)
	magnitude := float32(total) / float32(len(strings.Fields(text)))

	label := "neutral"
	if score > 0.1 {
		label = "positive"
	} else if score < -0.1 {
		label = "negative"
	}

	return &Sentiment{
		Score:      score,
		Magnitude:  magnitude,
		Label:      label,
		Confidence: 0.7,
	}
}

// extractTopics extracts topics from text (simplified implementation)
func (dp *DefaultDocumentProcessor) extractTopics(text string) []*Topic {
	// Simple topic extraction based on keyword clustering
	keywords := dp.extractKeywords(text)

	// Group keywords into topics (simplified)
	var topics []*Topic
	if len(keywords) > 0 {
		topic := &Topic{
			ID:         uuid.New().String(),
			Name:       "main_topic",
			Keywords:   keywords[:min(5, len(keywords))],
			Confidence: 0.6,
			CreatedAt:  time.Now(),
		}
		topics = append(topics, topic)
	}

	return topics
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
