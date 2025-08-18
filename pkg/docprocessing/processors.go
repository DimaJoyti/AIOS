package docprocessing

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// CleaningProcessor implements TextProcessor for text cleaning
type CleaningProcessor struct {
	removeExtraWhitespace  bool
	removeControlChars     bool
	normalizeUnicode       bool
	enableEmptyLineRemoval bool
	enableLineTrimming     bool
	logger                 *logrus.Logger
	tracer                 trace.Tracer
}

// NewCleaningProcessor creates a new cleaning processor
func NewCleaningProcessor(logger *logrus.Logger) *CleaningProcessor {
	return &CleaningProcessor{
		removeExtraWhitespace:  true,
		removeControlChars:     true,
		normalizeUnicode:       true,
		enableEmptyLineRemoval: true,
		enableLineTrimming:     true,
		logger:                 logger,
		tracer:                 otel.Tracer("docprocessing.processors.cleaning"),
	}
}

// Process cleans the text content
func (cp *CleaningProcessor) Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error) {
	ctx, span := cp.tracer.Start(ctx, "cleaning_processor.process")
	defer span.End()

	span.SetAttributes(
		attribute.Int("text.original_length", len(text)),
	)

	originalLength := len(text)
	cleanedText := text

	// Remove control characters except newlines and tabs
	if cp.removeControlChars {
		cleanedText = cp.removeControlCharacters(cleanedText)
	}

	// Normalize whitespace
	if cp.removeExtraWhitespace {
		cleanedText = cp.normalizeWhitespace(cleanedText)
	}

	// Remove empty lines
	if cp.enableEmptyLineRemoval {
		cleanedText = cp.removeEmptyLines(cleanedText)
	}

	// Trim lines
	if cp.enableLineTrimming {
		cleanedText = cp.trimLines(cleanedText)
	}

	// Update metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["cleaning_applied"] = true
	metadata["original_length"] = originalLength
	metadata["cleaned_length"] = len(cleanedText)
	metadata["reduction_ratio"] = float64(originalLength-len(cleanedText)) / float64(originalLength)

	span.SetAttributes(
		attribute.Int("text.cleaned_length", len(cleanedText)),
		attribute.Float64("text.reduction_ratio", metadata["reduction_ratio"].(float64)),
	)

	return cleanedText, metadata, nil
}

// GetProcessorType returns the processor type
func (cp *CleaningProcessor) GetProcessorType() string {
	return "cleaning"
}

// Configure configures the cleaning processor
func (cp *CleaningProcessor) Configure(options map[string]interface{}) error {
	if removeWhitespace, ok := options["remove_extra_whitespace"].(bool); ok {
		cp.removeExtraWhitespace = removeWhitespace
	}

	if removeControl, ok := options["remove_control_chars"].(bool); ok {
		cp.removeControlChars = removeControl
	}

	if normalize, ok := options["normalize_unicode"].(bool); ok {
		cp.normalizeUnicode = normalize
	}

	if removeEmpty, ok := options["remove_empty_lines"].(bool); ok {
		cp.enableEmptyLineRemoval = removeEmpty
	}

	if trim, ok := options["trim_lines"].(bool); ok {
		cp.enableLineTrimming = trim
	}

	return nil
}

func (cp *CleaningProcessor) removeControlCharacters(text string) string {
	return regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(text, "")
}

func (cp *CleaningProcessor) normalizeWhitespace(text string) string {
	// Replace multiple spaces with single space
	text = regexp.MustCompile(`[ \t]+`).ReplaceAllString(text, " ")
	// Replace multiple newlines with single newline
	text = regexp.MustCompile(`\n+`).ReplaceAllString(text, "\n")
	return text
}

func (cp *CleaningProcessor) removeEmptyLines(text string) string {
	lines := strings.Split(text, "\n")
	nonEmptyLines := make([]string, 0, len(lines))

	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines = append(nonEmptyLines, line)
		}
	}

	return strings.Join(nonEmptyLines, "\n")
}

func (cp *CleaningProcessor) trimLines(text string) string {
	lines := strings.Split(text, "\n")
	trimmedLines := make([]string, len(lines))

	for i, line := range lines {
		trimmedLines[i] = strings.TrimSpace(line)
	}

	return strings.Join(trimmedLines, "\n")
}

// LanguageDetectionProcessor implements TextProcessor for language detection
type LanguageDetectionProcessor struct {
	supportedLanguages []string
	defaultLanguage    string
	logger             *logrus.Logger
	tracer             trace.Tracer
}

// NewLanguageDetectionProcessor creates a new language detection processor
func NewLanguageDetectionProcessor(logger *logrus.Logger) *LanguageDetectionProcessor {
	return &LanguageDetectionProcessor{
		supportedLanguages: []string{"en", "es", "fr", "de", "it", "pt", "ru", "zh", "ja", "ko"},
		defaultLanguage:    "en",
		logger:             logger,
		tracer:             otel.Tracer("docprocessing.processors.language"),
	}
}

// Process detects the language of the text
func (ldp *LanguageDetectionProcessor) Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error) {
	ctx, span := ldp.tracer.Start(ctx, "language_detection_processor.process")
	defer span.End()

	// Simple language detection based on character patterns
	detectedLanguage := ldp.detectLanguage(text)

	// Update metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["detected_language"] = detectedLanguage
	metadata["language_detection_applied"] = true

	span.SetAttributes(
		attribute.String("text.detected_language", detectedLanguage),
	)

	return text, metadata, nil
}

// GetProcessorType returns the processor type
func (ldp *LanguageDetectionProcessor) GetProcessorType() string {
	return "language_detection"
}

// Configure configures the language detection processor
func (ldp *LanguageDetectionProcessor) Configure(options map[string]interface{}) error {
	if languages, ok := options["supported_languages"].([]string); ok {
		ldp.supportedLanguages = languages
	}

	if defaultLang, ok := options["default_language"].(string); ok {
		ldp.defaultLanguage = defaultLang
	}

	return nil
}

func (ldp *LanguageDetectionProcessor) detectLanguage(text string) string {
	// Simple language detection based on character patterns
	// This is a basic implementation - in production, you'd use a proper language detection library

	text = strings.ToLower(text)

	// Check for common patterns
	if regexp.MustCompile(`\b(the|and|or|but|in|on|at|to|for|of|with|by)\b`).MatchString(text) {
		return "en"
	}

	if regexp.MustCompile(`\b(el|la|los|las|y|o|pero|en|con|por|para|de)\b`).MatchString(text) {
		return "es"
	}

	if regexp.MustCompile(`\b(le|la|les|et|ou|mais|dans|sur|avec|par|pour|de)\b`).MatchString(text) {
		return "fr"
	}

	if regexp.MustCompile(`\b(der|die|das|und|oder|aber|in|auf|mit|von|zu|für)\b`).MatchString(text) {
		return "de"
	}

	// Check for non-Latin scripts
	if regexp.MustCompile(`[\u4e00-\u9fff]`).MatchString(text) {
		return "zh"
	}

	if regexp.MustCompile(`[\u3040-\u309f\u30a0-\u30ff]`).MatchString(text) {
		return "ja"
	}

	if regexp.MustCompile(`[\uac00-\ud7af]`).MatchString(text) {
		return "ko"
	}

	if regexp.MustCompile(`[\u0400-\u04ff]`).MatchString(text) {
		return "ru"
	}

	return ldp.defaultLanguage
}

// NormalizationProcessor implements TextProcessor for text normalization
type NormalizationProcessor struct {
	toLowerCase              bool
	enableAccentRemoval      bool
	enablePunctuationRemoval bool
	logger                   *logrus.Logger
	tracer                   trace.Tracer
}

// NewNormalizationProcessor creates a new normalization processor
func NewNormalizationProcessor(logger *logrus.Logger) *NormalizationProcessor {
	return &NormalizationProcessor{
		toLowerCase:              false,
		enableAccentRemoval:      false,
		enablePunctuationRemoval: false,
		logger:                   logger,
		tracer:                   otel.Tracer("docprocessing.processors.normalization"),
	}
}

// Process normalizes the text content
func (np *NormalizationProcessor) Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error) {
	ctx, span := np.tracer.Start(ctx, "normalization_processor.process")
	defer span.End()

	normalizedText := text

	if np.toLowerCase {
		normalizedText = strings.ToLower(normalizedText)
	}

	if np.enableAccentRemoval {
		normalizedText = np.removeAccents(normalizedText)
	}

	if np.enablePunctuationRemoval {
		normalizedText = np.removePunctuation(normalizedText)
	}

	// Update metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["normalization_applied"] = true
	metadata["to_lower_case"] = np.toLowerCase
	metadata["remove_accents"] = np.enableAccentRemoval
	metadata["remove_punctuation"] = np.enablePunctuationRemoval

	return normalizedText, metadata, nil
}

// GetProcessorType returns the processor type
func (np *NormalizationProcessor) GetProcessorType() string {
	return "normalization"
}

// Configure configures the normalization processor
func (np *NormalizationProcessor) Configure(options map[string]interface{}) error {
	if toLowerCase, ok := options["to_lower_case"].(bool); ok {
		np.toLowerCase = toLowerCase
	}

	if removeAccents, ok := options["remove_accents"].(bool); ok {
		np.enableAccentRemoval = removeAccents
	}

	if removePunctuation, ok := options["remove_punctuation"].(bool); ok {
		np.enablePunctuationRemoval = removePunctuation
	}

	return nil
}

func (np *NormalizationProcessor) removeAccents(text string) string {
	// Basic accent removal - in production, use a proper Unicode normalization library
	replacements := map[string]string{
		"á": "a", "à": "a", "ä": "a", "â": "a", "ā": "a", "ã": "a",
		"é": "e", "è": "e", "ë": "e", "ê": "e", "ē": "e",
		"í": "i", "ì": "i", "ï": "i", "î": "i", "ī": "i",
		"ó": "o", "ò": "o", "ö": "o", "ô": "o", "ō": "o", "õ": "o",
		"ú": "u", "ù": "u", "ü": "u", "û": "u", "ū": "u",
		"ñ": "n", "ç": "c",
		"Á": "A", "À": "A", "Ä": "A", "Â": "A", "Ā": "A", "Ã": "A",
		"É": "E", "È": "E", "Ë": "E", "Ê": "E", "Ē": "E",
		"Í": "I", "Ì": "I", "Ï": "I", "Î": "I", "Ī": "I",
		"Ó": "O", "Ò": "O", "Ö": "O", "Ô": "O", "Ō": "O", "Õ": "O",
		"Ú": "U", "Ù": "U", "Ü": "U", "Û": "U", "Ū": "U",
		"Ñ": "N", "Ç": "C",
	}

	for accented, plain := range replacements {
		text = strings.ReplaceAll(text, accented, plain)
	}

	return text
}

func (np *NormalizationProcessor) removePunctuation(text string) string {
	return regexp.MustCompile(`[^\p{L}\p{N}\s]`).ReplaceAllString(text, " ")
}

// SentenceChunker implements DocumentChunker for sentence-based chunking
type SentenceChunker struct {
	maxChunkSize int
	overlap      int
	logger       *logrus.Logger
	tracer       trace.Tracer
}

// NewSentenceChunker creates a new sentence chunker
func NewSentenceChunker(maxChunkSize, overlap int, logger *logrus.Logger) *SentenceChunker {
	return &SentenceChunker{
		maxChunkSize: maxChunkSize,
		overlap:      overlap,
		logger:       logger,
		tracer:       otel.Tracer("docprocessing.chunkers.sentence"),
	}
}

// ChunkDocument splits a document into sentence-based chunks
func (sc *SentenceChunker) ChunkDocument(ctx context.Context, doc *Document) ([]*DocumentChunk, error) {
	ctx, span := sc.tracer.Start(ctx, "sentence_chunker.chunk_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("chunker.max_chunk_size", sc.maxChunkSize),
		attribute.Int("chunker.overlap", sc.overlap),
	)

	sentences := sc.splitIntoSentences(doc.Content)
	chunks := sc.createChunks(doc, sentences)

	span.SetAttributes(
		attribute.Int("document.sentence_count", len(sentences)),
		attribute.Int("document.chunk_count", len(chunks)),
	)

	return chunks, nil
}

// GetChunkerType returns the chunker type
func (sc *SentenceChunker) GetChunkerType() string {
	return "sentence"
}

// Configure configures the sentence chunker
func (sc *SentenceChunker) Configure(options map[string]interface{}) error {
	if maxSize, ok := options["max_chunk_size"].(int); ok {
		sc.maxChunkSize = maxSize
	}

	if overlap, ok := options["overlap"].(int); ok {
		sc.overlap = overlap
	}

	return nil
}

func (sc *SentenceChunker) splitIntoSentences(text string) []string {
	// Simple sentence splitting - in production, use a proper NLP library
	sentenceRegex := regexp.MustCompile(`[.!?]+\s+`)
	sentences := sentenceRegex.Split(text, -1)

	// Clean up sentences
	cleanSentences := make([]string, 0, len(sentences))
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if sentence != "" {
			cleanSentences = append(cleanSentences, sentence)
		}
	}

	return cleanSentences
}

func (sc *SentenceChunker) createChunks(doc *Document, sentences []string) []*DocumentChunk {
	chunks := make([]*DocumentChunk, 0)
	currentChunk := ""
	currentPos := 0
	chunkIndex := 0

	for i, sentence := range sentences {
		// Check if adding this sentence would exceed the chunk size
		if len(currentChunk)+len(sentence)+1 > sc.maxChunkSize && currentChunk != "" {
			// Create chunk
			chunk := &DocumentChunk{
				ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
				DocumentID: doc.ID,
				Content:    strings.TrimSpace(currentChunk),
				ChunkIndex: chunkIndex,
				StartPos:   currentPos - len(currentChunk),
				EndPos:     currentPos,
				Metadata: map[string]interface{}{
					"chunk_type":      "sentence",
					"sentence_count":  strings.Count(currentChunk, ".") + strings.Count(currentChunk, "!") + strings.Count(currentChunk, "?"),
					"original_doc_id": doc.ID,
				},
			}

			chunks = append(chunks, chunk)
			chunkIndex++

			// Handle overlap
			if sc.overlap > 0 && i > 0 {
				overlapStart := max(0, i-sc.overlap)
				currentChunk = strings.Join(sentences[overlapStart:i], " ") + " "
			} else {
				currentChunk = ""
			}
		}

		if currentChunk != "" {
			currentChunk += " "
		}
		currentChunk += sentence
		currentPos += len(sentence) + 1
	}

	// Add the last chunk if it has content
	if currentChunk != "" {
		chunk := &DocumentChunk{
			ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
			DocumentID: doc.ID,
			Content:    strings.TrimSpace(currentChunk),
			ChunkIndex: chunkIndex,
			StartPos:   currentPos - len(currentChunk),
			EndPos:     currentPos,
			Metadata: map[string]interface{}{
				"chunk_type":      "sentence",
				"sentence_count":  strings.Count(currentChunk, ".") + strings.Count(currentChunk, "!") + strings.Count(currentChunk, "?"),
				"original_doc_id": doc.ID,
			},
		}

		chunks = append(chunks, chunk)
	}

	return chunks
}

// FixedSizeChunker implements DocumentChunker for fixed-size chunking
type FixedSizeChunker struct {
	chunkSize int
	overlap   int
	logger    *logrus.Logger
	tracer    trace.Tracer
}

// NewFixedSizeChunker creates a new fixed-size chunker
func NewFixedSizeChunker(chunkSize, overlap int, logger *logrus.Logger) *FixedSizeChunker {
	return &FixedSizeChunker{
		chunkSize: chunkSize,
		overlap:   overlap,
		logger:    logger,
		tracer:    otel.Tracer("docprocessing.chunkers.fixed_size"),
	}
}

// ChunkDocument splits a document into fixed-size chunks
func (fsc *FixedSizeChunker) ChunkDocument(ctx context.Context, doc *Document) ([]*DocumentChunk, error) {
	ctx, span := fsc.tracer.Start(ctx, "fixed_size_chunker.chunk_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("chunker.chunk_size", fsc.chunkSize),
		attribute.Int("chunker.overlap", fsc.overlap),
	)

	text := doc.Content
	chunks := make([]*DocumentChunk, 0)
	chunkIndex := 0
	start := 0

	for start < len(text) {
		end := start + fsc.chunkSize
		if end > len(text) {
			end = len(text)
		}

		// Try to break at word boundary
		if end < len(text) {
			for i := end; i > start && i > end-100; i-- {
				if unicode.IsSpace(rune(text[i])) {
					end = i
					break
				}
			}
		}

		chunkContent := strings.TrimSpace(text[start:end])
		if chunkContent != "" {
			chunk := &DocumentChunk{
				ID:         fmt.Sprintf("%s_chunk_%d", doc.ID, chunkIndex),
				DocumentID: doc.ID,
				Content:    chunkContent,
				ChunkIndex: chunkIndex,
				StartPos:   start,
				EndPos:     end,
				Metadata: map[string]interface{}{
					"chunk_type":      "fixed_size",
					"chunk_size":      len(chunkContent),
					"original_doc_id": doc.ID,
				},
			}

			chunks = append(chunks, chunk)
			chunkIndex++
		}

		start = end - fsc.overlap
		if start < 0 {
			start = 0
		}
		if start >= end {
			start = end
		}
	}

	span.SetAttributes(
		attribute.Int("document.chunk_count", len(chunks)),
	)

	return chunks, nil
}

// GetChunkerType returns the chunker type
func (fsc *FixedSizeChunker) GetChunkerType() string {
	return "fixed_size"
}

// Configure configures the fixed-size chunker
func (fsc *FixedSizeChunker) Configure(options map[string]interface{}) error {
	if chunkSize, ok := options["chunk_size"].(int); ok {
		fsc.chunkSize = chunkSize
	}

	if overlap, ok := options["overlap"].(int); ok {
		fsc.overlap = overlap
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
