package docprocessing

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultDocumentProcessor implements DocumentProcessor interface
type DefaultDocumentProcessor struct {
	pipeline ProcessingPipeline
	metrics  *ProcessingMetrics
	logger   *logrus.Logger
	tracer   trace.Tracer
	mu       sync.RWMutex
}

// NewDocumentProcessor creates a new document processor
func NewDocumentProcessor(pipeline ProcessingPipeline, logger *logrus.Logger) *DefaultDocumentProcessor {
	return &DefaultDocumentProcessor{
		pipeline: pipeline,
		metrics: &ProcessingMetrics{
			LastProcessedAt: time.Now(),
		},
		logger: logger,
		tracer: otel.Tracer("docprocessing.processor"),
	}
}

// ProcessFromSource processes documents from a source
func (dp *DefaultDocumentProcessor) ProcessFromSource(ctx context.Context, source DocumentSource) (<-chan *ProcessingResult, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_from_source")
	defer span.End()

	span.SetAttributes(
		attribute.String("source.type", source.GetSourceType()),
	)

	// Get documents from source
	docChan, err := source.GetDocuments(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents from source: %w", err)
	}

	// Process documents through pipeline
	resultChan, err := dp.pipeline.ProcessStream(ctx, docChan)
	if err != nil {
		return nil, fmt.Errorf("failed to process document stream: %w", err)
	}

	// Wrap result channel to update metrics
	wrappedChan := make(chan *ProcessingResult, 100)
	go func() {
		defer close(wrappedChan)
		for result := range resultChan {
			dp.updateMetrics(result)
			wrappedChan <- result
		}
	}()

	return wrappedChan, nil
}

// ProcessDocument processes a single document
func (dp *DefaultDocumentProcessor) ProcessDocument(ctx context.Context, doc *Document) (*ProcessingResult, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.source", doc.Source),
		attribute.String("document.content_type", doc.ContentType),
	)

	result, err := dp.pipeline.ProcessDocument(ctx, doc)
	if err != nil {
		dp.updateMetrics(&ProcessingResult{
			Document: doc,
			Success:  false,
			Error:    err,
			Duration: 0,
		})
		return result, err
	}

	dp.updateMetrics(result)
	return result, nil
}

// ProcessFile processes a file
func (dp *DefaultDocumentProcessor) ProcessFile(ctx context.Context, filePath string) (*ProcessingResult, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_file")
	defer span.End()

	span.SetAttributes(
		attribute.String("file.path", filePath),
	)

	// Create file source
	fileSource := NewFileSource(filepath.Dir(filePath), dp.logger)
	
	// Get the specific file
	doc, err := fileSource.GetDocument(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load file: %w", err)
	}

	return dp.ProcessDocument(ctx, doc)
}

// ProcessURL processes a document from URL
func (dp *DefaultDocumentProcessor) ProcessURL(ctx context.Context, url string) (*ProcessingResult, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_url")
	defer span.End()

	span.SetAttributes(
		attribute.String("url", url),
	)

	// Create URL source
	urlSource := NewURLSource([]string{url}, dp.logger)
	
	// Get the document
	doc, err := urlSource.GetDocument(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	return dp.ProcessDocument(ctx, doc)
}

// ProcessReader processes content from a reader
func (dp *DefaultDocumentProcessor) ProcessReader(ctx context.Context, reader io.Reader, contentType string, metadata map[string]interface{}) (*ProcessingResult, error) {
	ctx, span := dp.tracer.Start(ctx, "document_processor.process_reader")
	defer span.End()

	span.SetAttributes(
		attribute.String("content_type", contentType),
	)

	// Create reader source
	readerSource := NewReaderSource(dp.logger)
	id := readerSource.AddReader(reader, contentType, metadata)
	
	// Get the document
	doc, err := readerSource.GetDocument(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create document from reader: %w", err)
	}

	return dp.ProcessDocument(ctx, doc)
}

// GetPipeline returns the processing pipeline
func (dp *DefaultDocumentProcessor) GetPipeline() ProcessingPipeline {
	return dp.pipeline
}

// SetPipeline sets the processing pipeline
func (dp *DefaultDocumentProcessor) SetPipeline(pipeline ProcessingPipeline) error {
	if pipeline == nil {
		return fmt.Errorf("pipeline cannot be nil")
	}
	
	dp.mu.Lock()
	defer dp.mu.Unlock()
	
	dp.pipeline = pipeline
	return nil
}

// GetMetrics returns processing metrics
func (dp *DefaultDocumentProcessor) GetMetrics() *ProcessingMetrics {
	dp.mu.RLock()
	defer dp.mu.RUnlock()
	
	// Return a copy to avoid race conditions
	metricsCopy := *dp.metrics
	return &metricsCopy
}

func (dp *DefaultDocumentProcessor) updateMetrics(result *ProcessingResult) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	dp.metrics.TotalDocuments++
	dp.metrics.LastProcessedAt = time.Now()

	if result.Success {
		dp.metrics.ProcessedDocuments++
	} else {
		dp.metrics.FailedDocuments++
	}

	// Update timing metrics
	dp.metrics.TotalProcessTime += result.Duration
	if dp.metrics.ProcessedDocuments > 0 {
		dp.metrics.AverageProcessTime = dp.metrics.TotalProcessTime / time.Duration(dp.metrics.ProcessedDocuments)
	}

	// Update error rate
	if dp.metrics.TotalDocuments > 0 {
		dp.metrics.ErrorRate = float64(dp.metrics.FailedDocuments) / float64(dp.metrics.TotalDocuments)
	}

	// Update throughput (documents per second)
	if dp.metrics.TotalProcessTime > 0 {
		dp.metrics.ThroughputPerSec = float64(dp.metrics.ProcessedDocuments) / dp.metrics.TotalProcessTime.Seconds()
	}

	// Update chunk count if available
	if result.Chunks != nil {
		dp.metrics.TotalChunks += int64(len(result.Chunks))
	} else if result.Document != nil && result.Document.Metadata != nil {
		if chunkCount, exists := result.Document.Metadata["chunk_count"]; exists {
			if count, ok := chunkCount.(int); ok {
				dp.metrics.TotalChunks += int64(count)
			}
		}
	}
}

// DefaultProcessingManager implements ProcessingManager interface
type DefaultProcessingManager struct {
	extractors map[string]ContentExtractor
	processors map[string]TextProcessor
	chunkers   map[string]DocumentChunker
	sources    map[string]DocumentSource
	logger     *logrus.Logger
	tracer     trace.Tracer
	mu         sync.RWMutex
}

// NewProcessingManager creates a new processing manager
func NewProcessingManager(logger *logrus.Logger) *DefaultProcessingManager {
	manager := &DefaultProcessingManager{
		extractors: make(map[string]ContentExtractor),
		processors: make(map[string]TextProcessor),
		chunkers:   make(map[string]DocumentChunker),
		sources:    make(map[string]DocumentSource),
		logger:     logger,
		tracer:     otel.Tracer("docprocessing.manager"),
	}

	// Register default components
	manager.registerDefaults()

	return manager
}

// CreateProcessor creates a new document processor
func (pm *DefaultProcessingManager) CreateProcessor(config *ProcessingConfig) (DocumentProcessor, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Build pipeline from configuration
	builder := NewPipelineBuilder(pm.logger)

	// Configure pipeline settings
	if config.MaxConcurrency > 0 {
		builder.WithConcurrency(config.MaxConcurrency)
	}
	if config.ProcessingTimeout > 0 {
		builder.WithTimeout(config.ProcessingTimeout)
	}

	// Add stages based on configuration
	for _, stageConfig := range config.PipelineStages {
		if !stageConfig.Enabled {
			continue
		}

		switch stageConfig.Type {
		case "extraction":
			extractors := pm.getExtractorsForStage(stageConfig)
			if len(extractors) > 0 {
				builder.AddContentExtraction(extractors)
			}

		case "processing":
			processors := pm.getProcessorsForStage(stageConfig)
			if len(processors) > 0 {
				builder.AddTextProcessing(processors)
			}

		case "chunking":
			chunker := pm.getChunkerForStage(stageConfig)
			if chunker != nil {
				builder.AddChunking(chunker)
			}
		}
	}

	pipeline := builder.Build()
	processor := NewDocumentProcessor(pipeline, pm.logger)

	pm.logger.WithFields(logrus.Fields{
		"stages":          len(config.PipelineStages),
		"max_concurrency": config.MaxConcurrency,
		"timeout":         config.ProcessingTimeout,
	}).Info("Created document processor")

	return processor, nil
}

// RegisterExtractor registers a content extractor
func (pm *DefaultProcessingManager) RegisterExtractor(extractor ContentExtractor) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if extractor == nil {
		return fmt.Errorf("extractor cannot be nil")
	}

	// Register for all supported types
	for _, contentType := range extractor.GetSupportedTypes() {
		pm.extractors[contentType] = extractor
	}

	pm.logger.WithFields(logrus.Fields{
		"extractor_type":    fmt.Sprintf("%T", extractor),
		"supported_types":   extractor.GetSupportedTypes(),
	}).Info("Registered content extractor")

	return nil
}

// RegisterProcessor registers a text processor
func (pm *DefaultProcessingManager) RegisterProcessor(processor TextProcessor) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if processor == nil {
		return fmt.Errorf("processor cannot be nil")
	}

	pm.processors[processor.GetProcessorType()] = processor

	pm.logger.WithField("processor_type", processor.GetProcessorType()).Info("Registered text processor")

	return nil
}

// RegisterChunker registers a document chunker
func (pm *DefaultProcessingManager) RegisterChunker(chunker DocumentChunker) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if chunker == nil {
		return fmt.Errorf("chunker cannot be nil")
	}

	pm.chunkers[chunker.GetChunkerType()] = chunker

	pm.logger.WithField("chunker_type", chunker.GetChunkerType()).Info("Registered document chunker")

	return nil
}

// RegisterSource registers a document source
func (pm *DefaultProcessingManager) RegisterSource(source DocumentSource) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	pm.sources[source.GetSourceType()] = source

	pm.logger.WithField("source_type", source.GetSourceType()).Info("Registered document source")

	return nil
}

// GetExtractor returns a content extractor for the given content type
func (pm *DefaultProcessingManager) GetExtractor(contentType string) (ContentExtractor, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if extractor, exists := pm.extractors[contentType]; exists {
		return extractor, nil
	}

	// Try partial match
	for supportedType, extractor := range pm.extractors {
		if extractor.CanExtract(contentType) {
			return extractor, nil
		}
		_ = supportedType
	}

	return nil, fmt.Errorf("no extractor found for content type: %s", contentType)
}

// GetProcessor returns a text processor by type
func (pm *DefaultProcessingManager) GetProcessor(processorType string) (TextProcessor, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if processor, exists := pm.processors[processorType]; exists {
		return processor, nil
	}

	return nil, fmt.Errorf("no processor found for type: %s", processorType)
}

// GetChunker returns a document chunker by type
func (pm *DefaultProcessingManager) GetChunker(chunkerType string) (DocumentChunker, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if chunker, exists := pm.chunkers[chunkerType]; exists {
		return chunker, nil
	}

	return nil, fmt.Errorf("no chunker found for type: %s", chunkerType)
}

// GetSource returns a document source by type
func (pm *DefaultProcessingManager) GetSource(sourceType string) (DocumentSource, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if source, exists := pm.sources[sourceType]; exists {
		return source, nil
	}

	return nil, fmt.Errorf("no source found for type: %s", sourceType)
}

// ListExtractors returns all registered extractors
func (pm *DefaultProcessingManager) ListExtractors() []ContentExtractor {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	extractors := make([]ContentExtractor, 0, len(pm.extractors))
	seen := make(map[ContentExtractor]bool)

	for _, extractor := range pm.extractors {
		if !seen[extractor] {
			extractors = append(extractors, extractor)
			seen[extractor] = true
		}
	}

	return extractors
}

// ListProcessors returns all registered processors
func (pm *DefaultProcessingManager) ListProcessors() []TextProcessor {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	processors := make([]TextProcessor, 0, len(pm.processors))
	for _, processor := range pm.processors {
		processors = append(processors, processor)
	}

	return processors
}

// ListChunkers returns all registered chunkers
func (pm *DefaultProcessingManager) ListChunkers() []DocumentChunker {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	chunkers := make([]DocumentChunker, 0, len(pm.chunkers))
	for _, chunker := range pm.chunkers {
		chunkers = append(chunkers, chunker)
	}

	return chunkers
}

// ListSources returns all registered sources
func (pm *DefaultProcessingManager) ListSources() []DocumentSource {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	sources := make([]DocumentSource, 0, len(pm.sources))
	for _, source := range pm.sources {
		sources = append(sources, source)
	}

	return sources
}

func (pm *DefaultProcessingManager) registerDefaults() {
	// Register default extractors
	pm.RegisterExtractor(NewTextExtractor(pm.logger))
	pm.RegisterExtractor(NewHTMLExtractor(pm.logger))
	pm.RegisterExtractor(NewMarkdownExtractor(pm.logger))
	pm.RegisterExtractor(NewJSONExtractor(pm.logger))

	// Register default processors
	pm.RegisterProcessor(NewCleaningProcessor(pm.logger))
	pm.RegisterProcessor(NewLanguageDetectionProcessor(pm.logger))
	pm.RegisterProcessor(NewNormalizationProcessor(pm.logger))

	// Register default chunkers
	pm.RegisterChunker(NewFixedSizeChunker(1000, 200, pm.logger))
	pm.RegisterChunker(NewSentenceChunker(1000, 2, pm.logger))
}

func (pm *DefaultProcessingManager) getExtractorsForStage(stageConfig StageConfig) []ContentExtractor {
	extractors := make([]ContentExtractor, 0)

	if types, exists := stageConfig.Options["content_types"].([]string); exists {
		for _, contentType := range types {
			if extractor, err := pm.GetExtractor(contentType); err == nil {
				extractors = append(extractors, extractor)
			}
		}
	} else {
		// Return all extractors
		extractors = pm.ListExtractors()
	}

	return extractors
}

func (pm *DefaultProcessingManager) getProcessorsForStage(stageConfig StageConfig) []TextProcessor {
	processors := make([]TextProcessor, 0)

	if types, exists := stageConfig.Options["processor_types"].([]string); exists {
		for _, processorType := range types {
			if processor, err := pm.GetProcessor(processorType); err == nil {
				processors = append(processors, processor)
			}
		}
	} else {
		// Return default processors
		if processor, err := pm.GetProcessor("cleaning"); err == nil {
			processors = append(processors, processor)
		}
		if processor, err := pm.GetProcessor("language_detection"); err == nil {
			processors = append(processors, processor)
		}
	}

	return processors
}

func (pm *DefaultProcessingManager) getChunkerForStage(stageConfig StageConfig) DocumentChunker {
	chunkerType := "fixed_size"
	if ct, exists := stageConfig.Options["chunker_type"].(string); exists {
		chunkerType = ct
	}

	chunker, err := pm.GetChunker(chunkerType)
	if err != nil {
		pm.logger.WithError(err).WithField("chunker_type", chunkerType).Warn("Failed to get chunker")
		return nil
	}

	// Configure chunker with stage options
	if err := chunker.Configure(stageConfig.Options); err != nil {
		pm.logger.WithError(err).WithField("chunker_type", chunkerType).Warn("Failed to configure chunker")
	}

	return chunker
}
