package docprocessing

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextExtractor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	extractor := NewTextExtractor(logger)

	t.Run("CanExtract", func(t *testing.T) {
		assert.True(t, extractor.CanExtract("text/plain"))
		assert.True(t, extractor.CanExtract("application/json"))
		assert.False(t, extractor.CanExtract("application/pdf"))
	})

	t.Run("Extract", func(t *testing.T) {
		doc := &Document{
			ID:          "test1",
			Content:     "  This is a test document with   extra   spaces  \n\n\n",
			ContentType: "text/plain",
			Metadata:    make(map[string]interface{}),
		}

		result, err := extractor.Extract(context.Background(), doc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "This is a test document with extra spaces", result.Content)
		assert.Equal(t, "text_extractor", result.Metadata["extracted_by"])
	})
}

func TestHTMLExtractor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	extractor := NewHTMLExtractor(logger)

	t.Run("CanExtract", func(t *testing.T) {
		assert.True(t, extractor.CanExtract("text/html"))
		assert.True(t, extractor.CanExtract("application/xhtml+xml"))
		assert.False(t, extractor.CanExtract("text/plain"))
	})

	t.Run("Extract", func(t *testing.T) {
		htmlContent := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Document</title>
			<meta name="description" content="A test document">
		</head>
		<body>
			<h1>Main Title</h1>
			<p>This is a <strong>test</strong> paragraph with <a href="http://example.com">a link</a>.</p>
			<img src="test.jpg" alt="Test image">
			<script>console.log('test');</script>
		</body>
		</html>
		`

		doc := &Document{
			ID:          "test1",
			Content:     htmlContent,
			ContentType: "text/html",
			Metadata:    make(map[string]interface{}),
		}

		result, err := extractor.Extract(context.Background(), doc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Test Document", result.Title)
		assert.Contains(t, result.Content, "Main Title")
		assert.Contains(t, result.Content, "This is a test paragraph")
		assert.NotContains(t, result.Content, "<script>")
		assert.NotContains(t, result.Content, "console.log")

		// Check extracted metadata
		assert.Equal(t, "html_extractor", result.Metadata["extracted_by"])
		assert.Contains(t, result.Metadata, "links")
		assert.Contains(t, result.Metadata, "images")
		assert.Contains(t, result.Metadata, "meta_tags")
	})
}

func TestMarkdownExtractor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	extractor := NewMarkdownExtractor(logger)

	t.Run("CanExtract", func(t *testing.T) {
		assert.True(t, extractor.CanExtract("text/markdown"))
		assert.False(t, extractor.CanExtract("text/html"))
	})

	t.Run("Extract", func(t *testing.T) {
		markdownContent := `---
title: Test Document
author: John Doe
---

# Main Title

This is a **test** paragraph with [a link](http://example.com).

## Subtitle

- Item 1
- Item 2

` + "```go\nfunc test() {}\n```"

		doc := &Document{
			ID:          "test1",
			Content:     markdownContent,
			ContentType: "text/markdown",
			Metadata:    make(map[string]interface{}),
		}

		result, err := extractor.Extract(context.Background(), doc)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Main Title", result.Title)

		// Check extracted metadata
		assert.Equal(t, "markdown_extractor", result.Metadata["extracted_by"])
		assert.Contains(t, result.Metadata, "headers")
		assert.Contains(t, result.Metadata, "links")
		assert.Contains(t, result.Metadata, "front_matter")

		frontMatter := result.Metadata["front_matter"].(map[string]interface{})
		assert.Equal(t, "Test Document", frontMatter["title"])
		assert.Equal(t, "John Doe", frontMatter["author"])
	})
}

func TestCleaningProcessor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	processor := NewCleaningProcessor(logger)

	t.Run("Process", func(t *testing.T) {
		text := "  This   is   a   test   with\n\n\nextra\t\twhitespace  \n  "
		metadata := make(map[string]interface{})

		result, updatedMetadata, err := processor.Process(context.Background(), text, metadata)
		require.NoError(t, err)
		assert.Equal(t, "This is a test with\nextra whitespace", result)
		assert.True(t, updatedMetadata["cleaning_applied"].(bool))
		assert.Contains(t, updatedMetadata, "original_length")
		assert.Contains(t, updatedMetadata, "cleaned_length")
	})
}

func TestLanguageDetectionProcessor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	processor := NewLanguageDetectionProcessor(logger)

	t.Run("DetectEnglish", func(t *testing.T) {
		text := "This is an English text with common words like the, and, or, but."
		metadata := make(map[string]interface{})

		result, updatedMetadata, err := processor.Process(context.Background(), text, metadata)
		require.NoError(t, err)
		assert.Equal(t, text, result) // Text should be unchanged
		assert.Equal(t, "en", updatedMetadata["detected_language"])
		assert.True(t, updatedMetadata["language_detection_applied"].(bool))
	})

	t.Run("DetectSpanish", func(t *testing.T) {
		text := "Este es un texto en español con palabras como el, la, y, o."
		metadata := make(map[string]interface{})

		result, updatedMetadata, err := processor.Process(context.Background(), text, metadata)
		require.NoError(t, err)
		assert.Equal(t, text, result)
		assert.Equal(t, "es", updatedMetadata["detected_language"])
	})
}

func TestFixedSizeChunker(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	chunker := NewFixedSizeChunker(50, 10, logger)

	t.Run("ChunkDocument", func(t *testing.T) {
		doc := &Document{
			ID:      "test1",
			Content: "This is a long document that should be split into multiple chunks because it exceeds the chunk size limit that we have set for testing purposes.",
		}

		chunks, err := chunker.ChunkDocument(context.Background(), doc)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1)

		// Check chunk properties
		for i, chunk := range chunks {
			assert.Equal(t, doc.ID, chunk.DocumentID)
			assert.Equal(t, i, chunk.ChunkIndex)
			assert.Contains(t, chunk.ID, doc.ID)
			assert.LessOrEqual(t, len(chunk.Content), 50)
			assert.Equal(t, "fixed_size", chunk.Metadata["chunk_type"])
		}

		// Check overlap
		if len(chunks) > 1 {
			// There should be some overlap between consecutive chunks
			assert.True(t, chunks[0].EndPos > chunks[1].StartPos)
		}
	})
}

func TestSentenceChunker(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	chunker := NewSentenceChunker(100, 1, logger)

	t.Run("ChunkDocument", func(t *testing.T) {
		doc := &Document{
			ID:      "test1",
			Content: "This is the first sentence. This is the second sentence! This is the third sentence? This is a very long sentence that might exceed the chunk size limit and should be handled properly.",
		}

		chunks, err := chunker.ChunkDocument(context.Background(), doc)
		require.NoError(t, err)
		assert.Greater(t, len(chunks), 1)

		// Check chunk properties
		for i, chunk := range chunks {
			assert.Equal(t, doc.ID, chunk.DocumentID)
			assert.Equal(t, i, chunk.ChunkIndex)
			assert.Equal(t, "sentence", chunk.Metadata["chunk_type"])
			assert.Contains(t, chunk.Metadata, "sentence_count")
		}
	})
}

func TestProcessingPipeline(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("ProcessDocument", func(t *testing.T) {
		// Create pipeline with stages
		pipeline := NewProcessingPipeline(logger)

		// Add content extraction stage
		extractors := []ContentExtractor{
			NewTextExtractor(logger),
			NewHTMLExtractor(logger),
		}
		extractionStage := NewContentExtractionStage(extractors, logger)
		pipeline.AddStage(extractionStage)

		// Add text processing stage
		processors := []TextProcessor{
			NewCleaningProcessor(logger),
			NewLanguageDetectionProcessor(logger),
		}
		processingStage := NewTextProcessingStage(processors, logger)
		pipeline.AddStage(processingStage)

		// Add chunking stage
		chunker := NewFixedSizeChunker(100, 20, logger)
		chunkingStage := NewChunkingStage(chunker, logger)
		pipeline.AddStage(chunkingStage)

		// Test document
		doc := &Document{
			ID:          "test1",
			Content:     "  This is a test document with   extra   spaces.  It should be processed through the pipeline.  ",
			ContentType: "text/plain",
			Metadata:    make(map[string]interface{}),
		}

		result, err := pipeline.ProcessDocument(context.Background(), doc)
		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.NotNil(t, result.Document)
		assert.Greater(t, result.Duration, time.Duration(0))

		// Check that processing was applied
		assert.Contains(t, result.Document.Metadata, "cleaning_applied")
		assert.Contains(t, result.Document.Metadata, "detected_language")
		assert.Contains(t, result.Document.Metadata, "chunks")
	})

	t.Run("ProcessDocuments", func(t *testing.T) {
		pipeline := NewProcessingPipeline(logger)

		// Add simple processing stage
		processors := []TextProcessor{NewCleaningProcessor(logger)}
		processingStage := NewTextProcessingStage(processors, logger)
		pipeline.AddStage(processingStage)

		// Test documents
		docs := []*Document{
			{
				ID:          "test1",
				Content:     "  Document 1  ",
				ContentType: "text/plain",
				Metadata:    make(map[string]interface{}),
			},
			{
				ID:          "test2",
				Content:     "  Document 2  ",
				ContentType: "text/plain",
				Metadata:    make(map[string]interface{}),
			},
		}

		results, err := pipeline.ProcessDocuments(context.Background(), docs)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, result := range results {
			assert.True(t, result.Success)
			assert.Contains(t, result.Document.Metadata, "cleaning_applied")
		}
	})
}

func TestDocumentProcessor(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("ProcessDocument", func(t *testing.T) {
		// Create simple pipeline
		pipeline := NewProcessingPipeline(logger)
		processors := []TextProcessor{NewCleaningProcessor(logger)}
		processingStage := NewTextProcessingStage(processors, logger)
		pipeline.AddStage(processingStage)

		processor := NewDocumentProcessor(pipeline, logger)

		doc := &Document{
			ID:          "test1",
			Content:     "  Test document  ",
			ContentType: "text/plain",
			Metadata:    make(map[string]interface{}),
		}

		result, err := processor.ProcessDocument(context.Background(), doc)
		require.NoError(t, err)
		assert.True(t, result.Success)

		// Check metrics
		metrics := processor.GetMetrics()
		assert.Equal(t, int64(1), metrics.TotalDocuments)
		assert.Equal(t, int64(1), metrics.ProcessedDocuments)
		assert.Equal(t, int64(0), metrics.FailedDocuments)
	})
}

func TestProcessingManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager := NewProcessingManager(logger)

	t.Run("RegisterAndGetExtractor", func(t *testing.T) {
		extractor := NewTextExtractor(logger)
		err := manager.RegisterExtractor(extractor)
		require.NoError(t, err)

		retrieved, err := manager.GetExtractor("text/plain")
		require.NoError(t, err)
		assert.Equal(t, extractor, retrieved)
	})

	t.Run("RegisterAndGetProcessor", func(t *testing.T) {
		processor := NewCleaningProcessor(logger)
		err := manager.RegisterProcessor(processor)
		require.NoError(t, err)

		retrieved, err := manager.GetProcessor("cleaning")
		require.NoError(t, err)
		assert.Equal(t, processor, retrieved)
	})

	t.Run("RegisterAndGetChunker", func(t *testing.T) {
		chunker := NewFixedSizeChunker(100, 20, logger)
		err := manager.RegisterChunker(chunker)
		require.NoError(t, err)

		retrieved, err := manager.GetChunker("fixed_size")
		require.NoError(t, err)
		assert.Equal(t, chunker, retrieved)
	})

	t.Run("CreateProcessor", func(t *testing.T) {
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
		require.NoError(t, err)
		assert.NotNil(t, processor)

		// Test the created processor
		doc := &Document{
			ID:          "test1",
			Content:     "  Test  ",
			ContentType: "text/plain",
			Metadata:    make(map[string]interface{}),
		}

		result, err := processor.ProcessDocument(context.Background(), doc)
		require.NoError(t, err)
		assert.True(t, result.Success)
	})

	t.Run("ListComponents", func(t *testing.T) {
		extractors := manager.ListExtractors()
		assert.Greater(t, len(extractors), 0)

		processors := manager.ListProcessors()
		assert.Greater(t, len(processors), 0)

		chunkers := manager.ListChunkers()
		assert.Greater(t, len(chunkers), 0)
	})
}

func TestFileSource(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// This test would require actual files, so we'll test the basic functionality
	t.Run("GetSourceType", func(t *testing.T) {
		source := NewFileSource("/tmp", logger)
		assert.Equal(t, "file", source.GetSourceType())
	})

	t.Run("Configure", func(t *testing.T) {
		source := NewFileSource("/tmp", logger)
		err := source.Configure(map[string]interface{}{
			"patterns":      []string{"*.txt", "*.md"},
			"recursive":     false,
			"max_file_size": int64(1024 * 1024),
		})
		assert.NoError(t, err)
	})
}

func TestReaderSource(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("AddReaderAndGetDocument", func(t *testing.T) {
		source := NewReaderSource(logger)
		
		content := "This is test content"
		reader := strings.NewReader(content)
		metadata := map[string]interface{}{
			"test": "value",
		}

		id := source.AddReader(reader, "text/plain", metadata)
		assert.NotEmpty(t, id)

		doc, err := source.GetDocument(context.Background(), id)
		require.NoError(t, err)
		assert.Equal(t, content, doc.Content)
		assert.Equal(t, "text/plain", doc.ContentType)
		assert.Equal(t, "value", doc.Metadata["test"])
	})
}
