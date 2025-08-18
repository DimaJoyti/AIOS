package docprocessing

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultProcessingPipeline implements ProcessingPipeline interface
type DefaultProcessingPipeline struct {
	stages         []ProcessingStage
	maxConcurrency int
	timeout        time.Duration
	logger         *logrus.Logger
	tracer         trace.Tracer
	mu             sync.RWMutex
}

// NewProcessingPipeline creates a new processing pipeline
func NewProcessingPipeline(logger *logrus.Logger) *DefaultProcessingPipeline {
	return &DefaultProcessingPipeline{
		stages:         make([]ProcessingStage, 0),
		maxConcurrency: 10,
		timeout:        5 * time.Minute,
		logger:         logger,
		tracer:         otel.Tracer("docprocessing.pipeline"),
	}
}

// ProcessDocument processes a single document through the pipeline
func (p *DefaultProcessingPipeline) ProcessDocument(ctx context.Context, doc *Document) (*ProcessingResult, error) {
	ctx, span := p.tracer.Start(ctx, "pipeline.process_document")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.source", doc.Source),
		attribute.Int("pipeline.stage_count", len(p.stages)),
	)

	startTime := time.Now()
	result := &ProcessingResult{
		Document: doc,
		Success:  false,
		Duration: 0,
		Metadata: make(map[string]interface{}),
	}

	// Create a timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()

	// Process through each stage
	currentDoc := doc
	for i, stage := range p.stages {
		stageSpan := trace.SpanFromContext(timeoutCtx)
		stageSpan.SetAttributes(
			attribute.String("stage.name", stage.GetStageName()),
			attribute.String("stage.type", stage.GetStageType()),
			attribute.Int("stage.index", i),
		)

		processedDoc, err := stage.Process(timeoutCtx, currentDoc)
		if err != nil {
			result.Error = &ProcessingError{
				Stage:     stage.GetStageName(),
				Message:   fmt.Sprintf("Stage %s failed", stage.GetStageName()),
				Cause:     err,
				Document:  currentDoc,
				Timestamp: time.Now(),
			}
			span.RecordError(result.Error)
			result.Duration = time.Since(startTime)
			return result, result.Error
		}

		currentDoc = processedDoc
		p.logger.WithFields(logrus.Fields{
			"document_id": doc.ID,
			"stage":       stage.GetStageName(),
			"stage_index": i,
		}).Debug("Stage completed successfully")
	}

	result.Document = currentDoc
	result.Success = true
	result.Duration = time.Since(startTime)
	result.Metadata["processing_stages"] = len(p.stages)
	result.Metadata["processing_time"] = result.Duration.String()

	span.SetAttributes(
		attribute.Bool("processing.success", true),
		attribute.String("processing.duration", result.Duration.String()),
	)

	return result, nil
}

// ProcessDocuments processes multiple documents
func (p *DefaultProcessingPipeline) ProcessDocuments(ctx context.Context, docs []*Document) ([]*ProcessingResult, error) {
	ctx, span := p.tracer.Start(ctx, "pipeline.process_documents")
	defer span.End()

	span.SetAttributes(
		attribute.Int("documents.count", len(docs)),
		attribute.Int("pipeline.max_concurrency", p.maxConcurrency),
	)

	if len(docs) == 0 {
		return []*ProcessingResult{}, nil
	}

	// Create channels for work distribution
	docChan := make(chan *Document, len(docs))
	resultChan := make(chan *ProcessingResult, len(docs))

	// Send documents to channel
	for _, doc := range docs {
		docChan <- doc
	}
	close(docChan)

	// Start workers
	var wg sync.WaitGroup
	workerCount := min(p.maxConcurrency, len(docs))

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for doc := range docChan {
				result, err := p.ProcessDocument(ctx, doc)
				if err != nil && result == nil {
					result = &ProcessingResult{
						Document: doc,
						Success:  false,
						Error:    err,
						Duration: 0,
					}
				}
				resultChan <- result
			}
		}(i)
	}

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]*ProcessingResult, 0, len(docs))
	for result := range resultChan {
		results = append(results, result)
	}

	// Calculate summary metrics
	successCount := 0
	totalDuration := time.Duration(0)
	for _, result := range results {
		if result.Success {
			successCount++
		}
		totalDuration += result.Duration
	}

	span.SetAttributes(
		attribute.Int("processing.success_count", successCount),
		attribute.Int("processing.failure_count", len(results)-successCount),
		attribute.String("processing.total_duration", totalDuration.String()),
	)

	p.logger.WithFields(logrus.Fields{
		"total_documents":    len(docs),
		"successful":         successCount,
		"failed":            len(results) - successCount,
		"total_duration":    totalDuration,
		"average_duration":  totalDuration / time.Duration(len(docs)),
	}).Info("Batch processing completed")

	return results, nil
}

// ProcessStream processes documents from a stream
func (p *DefaultProcessingPipeline) ProcessStream(ctx context.Context, docStream <-chan *Document) (<-chan *ProcessingResult, error) {
	ctx, span := p.tracer.Start(ctx, "pipeline.process_stream")
	defer span.End()

	resultChan := make(chan *ProcessingResult, p.maxConcurrency)

	go func() {
		defer close(resultChan)

		// Create worker pool
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, p.maxConcurrency)

		for doc := range docStream {
			select {
			case semaphore <- struct{}{}:
				wg.Add(1)
				go func(d *Document) {
					defer func() {
						<-semaphore
						wg.Done()
					}()

					result, err := p.ProcessDocument(ctx, d)
					if err != nil && result == nil {
						result = &ProcessingResult{
							Document: d,
							Success:  false,
							Error:    err,
							Duration: 0,
						}
					}

					select {
					case resultChan <- result:
					case <-ctx.Done():
						return
					}
				}(doc)
			case <-ctx.Done():
				wg.Wait()
				return
			}
		}

		wg.Wait()
	}()

	return resultChan, nil
}

// AddStage adds a processing stage to the pipeline
func (p *DefaultProcessingPipeline) AddStage(stage ProcessingStage) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if stage == nil {
		return fmt.Errorf("stage cannot be nil")
	}

	if err := stage.Validate(); err != nil {
		return fmt.Errorf("stage validation failed: %w", err)
	}

	p.stages = append(p.stages, stage)

	p.logger.WithFields(logrus.Fields{
		"stage_name": stage.GetStageName(),
		"stage_type": stage.GetStageType(),
		"total_stages": len(p.stages),
	}).Info("Added processing stage")

	return nil
}

// RemoveStage removes a processing stage from the pipeline
func (p *DefaultProcessingPipeline) RemoveStage(stageName string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, stage := range p.stages {
		if stage.GetStageName() == stageName {
			p.stages = append(p.stages[:i], p.stages[i+1:]...)
			p.logger.WithField("stage_name", stageName).Info("Removed processing stage")
			return nil
		}
	}

	return fmt.Errorf("stage not found: %s", stageName)
}

// GetStages returns all processing stages
func (p *DefaultProcessingPipeline) GetStages() []ProcessingStage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stages := make([]ProcessingStage, len(p.stages))
	copy(stages, p.stages)
	return stages
}

// Configure configures the pipeline with options
func (p *DefaultProcessingPipeline) Configure(options map[string]interface{}) error {
	if maxConcurrency, ok := options["max_concurrency"].(int); ok {
		p.maxConcurrency = maxConcurrency
	}

	if timeout, ok := options["timeout"].(time.Duration); ok {
		p.timeout = timeout
	}

	return nil
}

// ContentExtractionStage implements ProcessingStage for content extraction
type ContentExtractionStage struct {
	extractors map[string]ContentExtractor
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewContentExtractionStage creates a new content extraction stage
func NewContentExtractionStage(extractors []ContentExtractor, logger *logrus.Logger) *ContentExtractionStage {
	extractorMap := make(map[string]ContentExtractor)
	for _, extractor := range extractors {
		for _, contentType := range extractor.GetSupportedTypes() {
			extractorMap[contentType] = extractor
		}
	}

	return &ContentExtractionStage{
		extractors: extractorMap,
		logger:     logger,
		tracer:     otel.Tracer("docprocessing.stages.extraction"),
	}
}

// Process processes a document in this stage
func (ces *ContentExtractionStage) Process(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := ces.tracer.Start(ctx, "content_extraction_stage.process")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
	)

	// Find appropriate extractor
	extractor := ces.findExtractor(doc.ContentType)
	if extractor == nil {
		// No extractor found, return document as-is
		ces.logger.WithFields(logrus.Fields{
			"document_id":   doc.ID,
			"content_type":  doc.ContentType,
		}).Debug("No extractor found for content type")
		return doc, nil
	}

	// Extract content
	extractedDoc, err := extractor.Extract(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("content extraction failed: %w", err)
	}

	span.SetAttributes(
		attribute.String("extractor.type", fmt.Sprintf("%T", extractor)),
	)

	return extractedDoc, nil
}

// GetStageName returns the stage name
func (ces *ContentExtractionStage) GetStageName() string {
	return "content_extraction"
}

// GetStageType returns the stage type
func (ces *ContentExtractionStage) GetStageType() string {
	return "extraction"
}

// Configure configures the stage with options
func (ces *ContentExtractionStage) Configure(options map[string]interface{}) error {
	// Content extraction stage doesn't need configuration
	return nil
}

// Validate validates the stage configuration
func (ces *ContentExtractionStage) Validate() error {
	if len(ces.extractors) == 0 {
		return fmt.Errorf("no extractors configured")
	}
	return nil
}

func (ces *ContentExtractionStage) findExtractor(contentType string) ContentExtractor {
	// Direct match
	if extractor, exists := ces.extractors[contentType]; exists {
		return extractor
	}

	// Partial match
	for supportedType, extractor := range ces.extractors {
		if extractor.CanExtract(contentType) {
			return extractor
		}
		_ = supportedType
	}

	return nil
}

// TextProcessingStage implements ProcessingStage for text processing
type TextProcessingStage struct {
	processors []TextProcessor
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewTextProcessingStage creates a new text processing stage
func NewTextProcessingStage(processors []TextProcessor, logger *logrus.Logger) *TextProcessingStage {
	return &TextProcessingStage{
		processors: processors,
		logger:     logger,
		tracer:     otel.Tracer("docprocessing.stages.text_processing"),
	}
}

// Process processes a document in this stage
func (tps *TextProcessingStage) Process(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := tps.tracer.Start(ctx, "text_processing_stage.process")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("processors.count", len(tps.processors)),
	)

	processedDoc := *doc
	currentText := doc.Content
	currentMetadata := make(map[string]interface{})

	// Copy existing metadata
	for k, v := range doc.Metadata {
		currentMetadata[k] = v
	}

	// Apply each processor
	for i, processor := range tps.processors {
		processedText, updatedMetadata, err := processor.Process(ctx, currentText, currentMetadata)
		if err != nil {
			return nil, fmt.Errorf("text processor %s failed: %w", processor.GetProcessorType(), err)
		}

		currentText = processedText
		currentMetadata = updatedMetadata

		tps.logger.WithFields(logrus.Fields{
			"document_id":    doc.ID,
			"processor":      processor.GetProcessorType(),
			"processor_index": i,
		}).Debug("Text processor completed")
	}

	processedDoc.Content = currentText
	processedDoc.Metadata = currentMetadata

	return &processedDoc, nil
}

// GetStageName returns the stage name
func (tps *TextProcessingStage) GetStageName() string {
	return "text_processing"
}

// GetStageType returns the stage type
func (tps *TextProcessingStage) GetStageType() string {
	return "processing"
}

// Configure configures the stage with options
func (tps *TextProcessingStage) Configure(options map[string]interface{}) error {
	// Text processing stage doesn't need configuration
	return nil
}

// Validate validates the stage configuration
func (tps *TextProcessingStage) Validate() error {
	if len(tps.processors) == 0 {
		return fmt.Errorf("no processors configured")
	}
	return nil
}

// ChunkingStage implements ProcessingStage for document chunking
type ChunkingStage struct {
	chunker DocumentChunker
	logger  *logrus.Logger
	tracer  trace.Tracer
}

// NewChunkingStage creates a new chunking stage
func NewChunkingStage(chunker DocumentChunker, logger *logrus.Logger) *ChunkingStage {
	return &ChunkingStage{
		chunker: chunker,
		logger:  logger,
		tracer:  otel.Tracer("docprocessing.stages.chunking"),
	}
}

// Process processes a document in this stage
func (cs *ChunkingStage) Process(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := cs.tracer.Start(ctx, "chunking_stage.process")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("chunker.type", cs.chunker.GetChunkerType()),
	)

	// Create chunks
	chunks, err := cs.chunker.ChunkDocument(ctx, doc)
	if err != nil {
		return nil, fmt.Errorf("document chunking failed: %w", err)
	}

	// Add chunks to document metadata
	processedDoc := *doc
	if processedDoc.Metadata == nil {
		processedDoc.Metadata = make(map[string]interface{})
	}

	processedDoc.Metadata["chunks"] = chunks
	processedDoc.Metadata["chunk_count"] = len(chunks)
	processedDoc.Metadata["chunker_type"] = cs.chunker.GetChunkerType()

	span.SetAttributes(
		attribute.Int("document.chunk_count", len(chunks)),
	)

	cs.logger.WithFields(logrus.Fields{
		"document_id":  doc.ID,
		"chunk_count":  len(chunks),
		"chunker_type": cs.chunker.GetChunkerType(),
	}).Debug("Document chunking completed")

	return &processedDoc, nil
}

// GetStageName returns the stage name
func (cs *ChunkingStage) GetStageName() string {
	return "chunking"
}

// GetStageType returns the stage type
func (cs *ChunkingStage) GetStageType() string {
	return "chunking"
}

// Configure configures the stage with options
func (cs *ChunkingStage) Configure(options map[string]interface{}) error {
	return cs.chunker.Configure(options)
}

// Validate validates the stage configuration
func (cs *ChunkingStage) Validate() error {
	if cs.chunker == nil {
		return fmt.Errorf("chunker not configured")
	}
	return nil
}

// Helper functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// PipelineBuilder provides a fluent interface for building pipelines
type PipelineBuilder struct {
	pipeline *DefaultProcessingPipeline
	stages   []StageConfig
}

// NewPipelineBuilder creates a new pipeline builder
func NewPipelineBuilder(logger *logrus.Logger) *PipelineBuilder {
	return &PipelineBuilder{
		pipeline: NewProcessingPipeline(logger),
		stages:   make([]StageConfig, 0),
	}
}

// AddContentExtraction adds a content extraction stage
func (pb *PipelineBuilder) AddContentExtraction(extractors []ContentExtractor) *PipelineBuilder {
	stage := NewContentExtractionStage(extractors, pb.pipeline.logger)
	pb.pipeline.AddStage(stage)
	return pb
}

// AddTextProcessing adds a text processing stage
func (pb *PipelineBuilder) AddTextProcessing(processors []TextProcessor) *PipelineBuilder {
	stage := NewTextProcessingStage(processors, pb.pipeline.logger)
	pb.pipeline.AddStage(stage)
	return pb
}

// AddChunking adds a chunking stage
func (pb *PipelineBuilder) AddChunking(chunker DocumentChunker) *PipelineBuilder {
	stage := NewChunkingStage(chunker, pb.pipeline.logger)
	pb.pipeline.AddStage(stage)
	return pb
}

// WithConcurrency sets the maximum concurrency
func (pb *PipelineBuilder) WithConcurrency(maxConcurrency int) *PipelineBuilder {
	pb.pipeline.maxConcurrency = maxConcurrency
	return pb
}

// WithTimeout sets the processing timeout
func (pb *PipelineBuilder) WithTimeout(timeout time.Duration) *PipelineBuilder {
	pb.pipeline.timeout = timeout
	return pb
}

// Build builds the pipeline
func (pb *PipelineBuilder) Build() ProcessingPipeline {
	// Sort stages by order if specified
	sort.Slice(pb.stages, func(i, j int) bool {
		return pb.stages[i].Order < pb.stages[j].Order
	})

	return pb.pipeline
}
