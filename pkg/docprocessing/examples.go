package docprocessing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// ExampleBasicProcessing demonstrates basic document processing
func ExampleBasicProcessing() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create processing manager
	manager := NewProcessingManager(logger)

	// Create a simple processing configuration
	config := &ProcessingConfig{
		PipelineStages: []StageConfig{
			{
				Name:    "extraction",
				Type:    "extraction",
				Enabled: true,
				Order:   1,
			},
			{
				Name:    "text_processing",
				Type:    "processing",
				Enabled: true,
				Order:   2,
				Options: map[string]interface{}{
					"processor_types": []string{"cleaning", "language_detection"},
				},
			},
			{
				Name:    "chunking",
				Type:    "chunking",
				Enabled: true,
				Order:   3,
				Options: map[string]interface{}{
					"chunker_type": "fixed_size",
					"chunk_size":   500,
					"overlap":      100,
				},
			},
		},
		MaxConcurrency:    5,
		ProcessingTimeout: 30 * time.Second,
	}

	// Create document processor
	processor, err := manager.CreateProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	// Create test documents
	documents := []*Document{
		{
			ID:          "doc1",
			Source:      "example",
			Title:       "Sample Text Document",
			Content:     "  This is a sample text document with some extra whitespace.  It contains multiple sentences. Each sentence provides some information about the document processing capabilities.  ",
			ContentType: "text/plain",
			Metadata: map[string]interface{}{
				"category": "example",
				"author":   "system",
			},
		},
		{
			ID:          "doc2",
			Source:      "example",
			Title:       "HTML Document",
			Content:     `<html><head><title>HTML Example</title></head><body><h1>Main Title</h1><p>This is an <strong>HTML</strong> document with <a href="http://example.com">links</a> and formatting.</p></body></html>`,
			ContentType: "text/html",
			Metadata: map[string]interface{}{
				"category": "web",
				"format":   "html",
			},
		},
		{
			ID:          "doc3",
			Source:      "example",
			Title:       "Markdown Document",
			Content:     "# Markdown Example\n\nThis is a **markdown** document with [links](http://example.com) and formatting.\n\n## Section 2\n\n- Item 1\n- Item 2\n- Item 3",
			ContentType: "text/markdown",
			Metadata: map[string]interface{}{
				"category": "documentation",
				"format":   "markdown",
			},
		},
	}

	ctx := context.Background()

	logger.Info("🚀 Starting basic document processing example")

	// Process documents individually
	for _, doc := range documents {
		logger.WithFields(logrus.Fields{
			"document_id":   doc.ID,
			"content_type":  doc.ContentType,
			"original_size": len(doc.Content),
		}).Info("Processing document")

		result, err := processor.ProcessDocument(ctx, doc)
		if err != nil {
			logger.WithError(err).WithField("document_id", doc.ID).Error("Failed to process document")
			continue
		}

		logger.WithFields(logrus.Fields{
			"document_id":       doc.ID,
			"success":           result.Success,
			"duration":          result.Duration,
			"processed_size":    len(result.Document.Content),
			"detected_language": result.Document.Metadata["detected_language"],
			"chunk_count":       result.Document.Metadata["chunk_count"],
		}).Info("Document processed successfully")

		// Display chunks if available
		if chunks, exists := result.Document.Metadata["chunks"]; exists {
			if chunkList, ok := chunks.([]*DocumentChunk); ok {
				logger.WithField("document_id", doc.ID).Info("Document chunks:")
				for i, chunk := range chunkList {
					logger.WithFields(logrus.Fields{
						"chunk_index": i,
						"chunk_id":    chunk.ID,
						"chunk_size":  len(chunk.Content),
						"start_pos":   chunk.StartPos,
						"end_pos":     chunk.EndPos,
					}).Info("Chunk details")
				}
			}
		}
	}

	// Get processing metrics
	metrics := processor.GetMetrics()
	logger.WithFields(logrus.Fields{
		"total_documents":      metrics.TotalDocuments,
		"processed_documents":  metrics.ProcessedDocuments,
		"failed_documents":     metrics.FailedDocuments,
		"total_chunks":         metrics.TotalChunks,
		"average_process_time": metrics.AverageProcessTime,
		"error_rate":           metrics.ErrorRate,
		"throughput_per_sec":   metrics.ThroughputPerSec,
	}).Info("Processing metrics")

	logger.Info("✅ Basic document processing example completed")

	return nil
}

// ExampleBatchProcessing demonstrates batch document processing
func ExampleBatchProcessing() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	manager := NewProcessingManager(logger)

	// Create configuration for batch processing
	config := &ProcessingConfig{
		PipelineStages: []StageConfig{
			{
				Name:    "extraction",
				Type:    "extraction",
				Enabled: true,
				Order:   1,
			},
			{
				Name:    "processing",
				Type:    "processing",
				Enabled: true,
				Order:   2,
				Options: map[string]interface{}{
					"processor_types": []string{"cleaning", "language_detection", "normalization"},
				},
			},
		},
		MaxConcurrency:    10, // Higher concurrency for batch processing
		ProcessingTimeout: 60 * time.Second,
	}

	processor, err := manager.CreateProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	// Generate multiple test documents
	documents := make([]*Document, 0, 20)
	for i := 0; i < 20; i++ {
		doc := &Document{
			ID:          fmt.Sprintf("batch_doc_%d", i),
			Source:      "batch_example",
			Title:       fmt.Sprintf("Batch Document %d", i),
			Content:     fmt.Sprintf("  This is batch document number %d. It contains some sample content for testing batch processing capabilities. The document has multiple sentences and various formatting.  ", i),
			ContentType: "text/plain",
			Metadata: map[string]interface{}{
				"batch_id": "batch_001",
				"index":    i,
				"category": "batch_test",
			},
		}
		documents = append(documents, doc)
	}

	ctx := context.Background()

	logger.WithField("document_count", len(documents)).Info("🚀 Starting batch processing example")

	startTime := time.Now()

	// Process all documents in batch
	results, err := processor.GetPipeline().ProcessDocuments(ctx, documents)
	if err != nil {
		return fmt.Errorf("batch processing failed: %w", err)
	}

	totalDuration := time.Since(startTime)

	// Analyze results
	successCount := 0
	failureCount := 0
	totalProcessingTime := time.Duration(0)

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			failureCount++
		}
		totalProcessingTime += result.Duration
	}

	logger.WithFields(logrus.Fields{
		"total_documents":       len(documents),
		"successful":            successCount,
		"failed":                failureCount,
		"total_wall_time":       totalDuration,
		"total_processing_time": totalProcessingTime,
		"average_per_doc":       totalProcessingTime / time.Duration(len(documents)),
		"throughput_per_sec":    float64(len(documents)) / totalDuration.Seconds(),
		"concurrency_benefit":   float64(totalProcessingTime.Nanoseconds()) / float64(totalDuration.Nanoseconds()),
	}).Info("Batch processing completed")

	// Display sample results
	logger.Info("Sample processing results:")
	for i, result := range results[:min(5, len(results))] {
		logger.WithFields(logrus.Fields{
			"index":             i,
			"document_id":       result.Document.ID,
			"success":           result.Success,
			"duration":          result.Duration,
			"detected_language": result.Document.Metadata["detected_language"],
			"original_length":   result.Document.Metadata["original_length"],
			"cleaned_length":    result.Document.Metadata["cleaned_length"],
		}).Info("Sample result")
	}

	logger.Info("✅ Batch processing example completed")

	return nil
}

// ExampleStreamProcessing demonstrates stream-based document processing
func ExampleStreamProcessing() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	manager := NewProcessingManager(logger)

	config := &ProcessingConfig{
		PipelineStages: []StageConfig{
			{
				Name:    "extraction",
				Type:    "extraction",
				Enabled: true,
				Order:   1,
			},
			{
				Name:    "processing",
				Type:    "processing",
				Enabled: true,
				Order:   2,
				Options: map[string]interface{}{
					"processor_types": []string{"cleaning"},
				},
			},
		},
		MaxConcurrency: 5,
	}

	processor, err := manager.CreateProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	ctx := context.Background()

	logger.Info("🚀 Starting stream processing example")

	// Create a document stream
	docChan := make(chan *Document, 10)

	// Start processing stream
	resultChan, err := processor.GetPipeline().ProcessStream(ctx, docChan)
	if err != nil {
		return fmt.Errorf("failed to start stream processing: %w", err)
	}

	// Send documents to stream
	go func() {
		defer close(docChan)

		for i := 0; i < 15; i++ {
			doc := &Document{
				ID:          fmt.Sprintf("stream_doc_%d", i),
				Source:      "stream_example",
				Title:       fmt.Sprintf("Stream Document %d", i),
				Content:     fmt.Sprintf("  Stream document %d with content to be processed.  ", i),
				ContentType: "text/plain",
				Metadata: map[string]interface{}{
					"stream_id": "stream_001",
					"index":     i,
				},
			}

			select {
			case docChan <- doc:
				logger.WithField("document_id", doc.ID).Debug("Sent document to stream")
			case <-ctx.Done():
				return
			}

			// Simulate streaming delay
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// Process results as they come
	processedCount := 0
	for result := range resultChan {
		processedCount++

		logger.WithFields(logrus.Fields{
			"processed_count": processedCount,
			"document_id":     result.Document.ID,
			"success":         result.Success,
			"duration":        result.Duration,
		}).Info("Processed stream document")

		if result.Error != nil {
			logger.WithError(result.Error).WithField("document_id", result.Document.ID).Error("Stream processing error")
		}
	}

	logger.WithField("total_processed", processedCount).Info("✅ Stream processing example completed")

	return nil
}

// ExampleReaderSource demonstrates processing from io.Reader sources
func ExampleReaderSource() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	manager := NewProcessingManager(logger)

	config := &ProcessingConfig{
		PipelineStages: []StageConfig{
			{
				Name:    "extraction",
				Type:    "extraction",
				Enabled: true,
				Order:   1,
			},
			{
				Name:    "processing",
				Type:    "processing",
				Enabled: true,
				Order:   2,
				Options: map[string]interface{}{
					"processor_types": []string{"cleaning", "language_detection"},
				},
			},
		},
	}

	processor, err := manager.CreateProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	ctx := context.Background()

	logger.Info("🚀 Starting reader source example")

	// Test different content types
	testContents := []struct {
		content     string
		contentType string
		metadata    map[string]interface{}
	}{
		{
			content:     "  This is plain text content with extra spaces.  ",
			contentType: "text/plain",
			metadata:    map[string]interface{}{"source": "string"},
		},
		{
			content:     `<html><body><h1>HTML Content</h1><p>This is HTML content.</p></body></html>`,
			contentType: "text/html",
			metadata:    map[string]interface{}{"source": "html_string"},
		},
		{
			content:     `{"title": "JSON Document", "content": "This is JSON content", "tags": ["test", "example"]}`,
			contentType: "application/json",
			metadata:    map[string]interface{}{"source": "json_string"},
		},
	}

	for i, testContent := range testContents {
		reader := strings.NewReader(testContent.content)

		result, err := processor.ProcessReader(ctx, reader, testContent.contentType, testContent.metadata)
		if err != nil {
			logger.WithError(err).WithField("index", i).Error("Failed to process reader")
			continue
		}

		logger.WithFields(logrus.Fields{
			"index":             i,
			"content_type":      testContent.contentType,
			"success":           result.Success,
			"duration":          result.Duration,
			"original_length":   len(testContent.content),
			"processed_length":  len(result.Document.Content),
			"detected_language": result.Document.Metadata["detected_language"],
		}).Info("Processed reader content")
	}

	logger.Info("✅ Reader source example completed")

	return nil
}

// ExampleCustomProcessor demonstrates creating custom processors
func ExampleCustomProcessor() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create a custom processor that counts words
	wordCountProcessor := &WordCountProcessor{
		logger: logger,
	}

	// Create a custom chunker that splits on paragraphs
	paragraphChunker := &ParagraphChunker{
		logger: logger,
	}

	manager := NewProcessingManager(logger)

	// Register custom components
	err := manager.RegisterProcessor(wordCountProcessor)
	if err != nil {
		return fmt.Errorf("failed to register word count processor: %w", err)
	}

	err = manager.RegisterChunker(paragraphChunker)
	if err != nil {
		return fmt.Errorf("failed to register paragraph chunker: %w", err)
	}

	config := &ProcessingConfig{
		PipelineStages: []StageConfig{
			{
				Name:    "extraction",
				Type:    "extraction",
				Enabled: true,
				Order:   1,
			},
			{
				Name:    "custom_processing",
				Type:    "processing",
				Enabled: true,
				Order:   2,
				Options: map[string]interface{}{
					"processor_types": []string{"cleaning", "word_count"},
				},
			},
			{
				Name:    "custom_chunking",
				Type:    "chunking",
				Enabled: true,
				Order:   3,
				Options: map[string]interface{}{
					"chunker_type": "paragraph",
				},
			},
		},
	}

	processor, err := manager.CreateProcessor(config)
	if err != nil {
		return fmt.Errorf("failed to create processor: %w", err)
	}

	ctx := context.Background()

	logger.Info("🚀 Starting custom processor example")

	doc := &Document{
		ID:          "custom_test",
		Source:      "example",
		Title:       "Custom Processing Test",
		Content:     "This is the first paragraph with several words.\n\nThis is the second paragraph with different content.\n\nThis is the third paragraph with even more content to demonstrate custom processing.",
		ContentType: "text/plain",
		Metadata:    make(map[string]interface{}),
	}

	result, err := processor.ProcessDocument(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to process document: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"document_id":  doc.ID,
		"success":      result.Success,
		"duration":     result.Duration,
		"word_count":   result.Document.Metadata["word_count"],
		"chunk_count":  result.Document.Metadata["chunk_count"],
		"chunker_type": result.Document.Metadata["chunker_type"],
	}).Info("Custom processing completed")

	// Display chunks
	if chunks, exists := result.Document.Metadata["chunks"]; exists {
		if chunkList, ok := chunks.([]*DocumentChunk); ok {
			logger.Info("Paragraph chunks:")
			for i, chunk := range chunkList {
				logger.WithFields(logrus.Fields{
					"chunk_index": i,
					"chunk_id":    chunk.ID,
					"content":     chunk.Content,
				}).Info("Paragraph chunk")
			}
		}
	}

	logger.Info("✅ Custom processor example completed")

	return nil
}

// WordCountProcessor is a custom processor that counts words
type WordCountProcessor struct {
	logger *logrus.Logger
}

func (wcp *WordCountProcessor) Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error) {
	words := strings.Fields(text)
	wordCount := len(words)

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	metadata["word_count"] = wordCount
	metadata["word_count_applied"] = true

	return text, metadata, nil
}

func (wcp *WordCountProcessor) GetProcessorType() string {
	return "word_count"
}

func (wcp *WordCountProcessor) Configure(options map[string]interface{}) error {
	return nil
}

// ParagraphChunker is a custom chunker that splits on paragraphs
type ParagraphChunker struct {
	logger *logrus.Logger
}

func (pc *ParagraphChunker) ChunkDocument(ctx context.Context, doc *Document) ([]*DocumentChunk, error) {
	paragraphs := strings.Split(doc.Content, "\n\n")
	chunks := make([]*DocumentChunk, 0, len(paragraphs))

	currentPos := 0
	for i, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}

		chunk := &DocumentChunk{
			ID:         fmt.Sprintf("%s_para_%d", doc.ID, i),
			DocumentID: doc.ID,
			Content:    paragraph,
			ChunkIndex: i,
			StartPos:   currentPos,
			EndPos:     currentPos + len(paragraph),
			Metadata: map[string]interface{}{
				"chunk_type":      "paragraph",
				"paragraph_index": i,
				"original_doc_id": doc.ID,
			},
		}

		chunks = append(chunks, chunk)
		currentPos += len(paragraph) + 2 // +2 for \n\n
	}

	return chunks, nil
}

func (pc *ParagraphChunker) GetChunkerType() string {
	return "paragraph"
}

func (pc *ParagraphChunker) Configure(options map[string]interface{}) error {
	return nil
}

// DemoDocumentProcessing demonstrates all document processing capabilities
func DemoDocumentProcessing() error {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("🚀 Starting Document Processing Demo")

	// Run all examples
	examples := []struct {
		name string
		fn   func() error
	}{
		{"Basic Processing", ExampleBasicProcessing},
		{"Batch Processing", ExampleBatchProcessing},
		{"Stream Processing", ExampleStreamProcessing},
		{"Reader Source", ExampleReaderSource},
		{"Custom Processor", ExampleCustomProcessor},
	}

	for _, example := range examples {
		logger.WithField("example", example.name).Info("Running example")

		if err := example.fn(); err != nil {
			logger.WithError(err).WithField("example", example.name).Error("Example failed")
			return err
		}

		logger.WithField("example", example.name).Info("✅ Example completed")
		time.Sleep(1 * time.Second) // Brief pause between examples
	}

	logger.Info("🎉 Document Processing Demo completed successfully!")

	return nil
}

// min function is defined in pipeline.go
