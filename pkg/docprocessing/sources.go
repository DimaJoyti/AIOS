package docprocessing

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// FileSource implements DocumentSource for file system sources
type FileSource struct {
	basePath    string
	patterns    []string
	recursive   bool
	maxFileSize int64
	logger      *logrus.Logger
	tracer      trace.Tracer
}

// NewFileSource creates a new file source
func NewFileSource(basePath string, logger *logrus.Logger) *FileSource {
	return &FileSource{
		basePath:    basePath,
		patterns:    []string{"*"},
		recursive:   true,
		maxFileSize: 100 * 1024 * 1024, // 100MB default
		logger:      logger,
		tracer:      otel.Tracer("docprocessing.sources.file"),
	}
}

// GetDocuments retrieves documents from the file system
func (fs *FileSource) GetDocuments(ctx context.Context) (<-chan *Document, error) {
	ctx, span := fs.tracer.Start(ctx, "file_source.get_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("source.base_path", fs.basePath),
		attribute.Bool("source.recursive", fs.recursive),
	)

	docChan := make(chan *Document, 100)

	go func() {
		defer close(docChan)

		err := fs.walkFiles(ctx, fs.basePath, func(filePath string, info os.FileInfo) error {
			if info.Size() > fs.maxFileSize {
				fs.logger.WithFields(logrus.Fields{
					"file": filePath,
					"size": info.Size(),
					"max":  fs.maxFileSize,
				}).Warn("File too large, skipping")
				return nil
			}

			doc, err := fs.createDocumentFromFile(filePath, info)
			if err != nil {
				fs.logger.WithError(err).WithField("file", filePath).Error("Failed to create document")
				return nil
			}

			select {
			case docChan <- doc:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})

		if err != nil {
			fs.logger.WithError(err).Error("Error walking files")
		}
	}()

	return docChan, nil
}

// GetDocument retrieves a specific document by file path
func (fs *FileSource) GetDocument(ctx context.Context, id string) (*Document, error) {
	ctx, span := fs.tracer.Start(ctx, "file_source.get_document")
	defer span.End()

	span.SetAttributes(attribute.String("document.id", id))

	// Treat ID as file path
	filePath := id
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(fs.basePath, filePath)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}

	return fs.createDocumentFromFile(filePath, info)
}

// GetSourceType returns the source type
func (fs *FileSource) GetSourceType() string {
	return "file"
}

// Configure configures the file source
func (fs *FileSource) Configure(options map[string]interface{}) error {
	if patterns, ok := options["patterns"].([]string); ok {
		fs.patterns = patterns
	}

	if recursive, ok := options["recursive"].(bool); ok {
		fs.recursive = recursive
	}

	if maxSize, ok := options["max_file_size"].(int64); ok {
		fs.maxFileSize = maxSize
	}

	return nil
}

// Close closes the file source
func (fs *FileSource) Close() error {
	return nil
}

func (fs *FileSource) walkFiles(ctx context.Context, root string, fn func(string, os.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if !fs.recursive && path != root {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file matches patterns
		if !fs.matchesPatterns(filepath.Base(path)) {
			return nil
		}

		return fn(path, info)
	})
}

func (fs *FileSource) matchesPatterns(filename string) bool {
	for _, pattern := range fs.patterns {
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
	}
	return false
}

func (fs *FileSource) createDocumentFromFile(filePath string, info os.FileInfo) (*Document, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", md5.Sum(content))

	// Detect content type
	contentType := fs.detectContentType(filePath, content)

	doc := &Document{
		ID:          filePath,
		Source:      fs.GetSourceType(),
		Title:       filepath.Base(filePath),
		Content:     string(content),
		ContentType: contentType,
		Metadata: map[string]interface{}{
			"file_path":   filePath,
			"file_size":   info.Size(),
			"file_mode":   info.Mode().String(),
			"modified_at": info.ModTime(),
			"extension":   filepath.Ext(filePath),
		},
		CreatedAt: info.ModTime(),
		UpdatedAt: info.ModTime(),
		Size:      info.Size(),
		Checksum:  checksum,
	}

	return doc, nil
}

func (fs *FileSource) detectContentType(filePath string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".doc":
		return "application/msword"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".xml":
		return "application/xml"
	case ".json":
		return "application/json"
	case ".csv":
		return "text/csv"
	case ".md":
		return "text/markdown"
	default:
		// Use http.DetectContentType for unknown types
		return http.DetectContentType(content)
	}
}

// URLSource implements DocumentSource for URL sources
type URLSource struct {
	urls       []string
	httpClient *http.Client
	userAgent  string
	headers    map[string]string
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// NewURLSource creates a new URL source
func NewURLSource(urls []string, logger *logrus.Logger) *URLSource {
	return &URLSource{
		urls: urls,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "AIOS-DocumentProcessor/1.0",
		headers:   make(map[string]string),
		logger:    logger,
		tracer:    otel.Tracer("docprocessing.sources.url"),
	}
}

// GetDocuments retrieves documents from URLs
func (us *URLSource) GetDocuments(ctx context.Context) (<-chan *Document, error) {
	ctx, span := us.tracer.Start(ctx, "url_source.get_documents")
	defer span.End()

	span.SetAttributes(attribute.Int("source.url_count", len(us.urls)))

	docChan := make(chan *Document, 10)

	go func() {
		defer close(docChan)

		for _, url := range us.urls {
			doc, err := us.fetchDocument(ctx, url)
			if err != nil {
				us.logger.WithError(err).WithField("url", url).Error("Failed to fetch document")
				continue
			}

			select {
			case docChan <- doc:
			case <-ctx.Done():
				return
			}
		}
	}()

	return docChan, nil
}

// GetDocument retrieves a specific document by URL
func (us *URLSource) GetDocument(ctx context.Context, id string) (*Document, error) {
	return us.fetchDocument(ctx, id)
}

// GetSourceType returns the source type
func (us *URLSource) GetSourceType() string {
	return "url"
}

// Configure configures the URL source
func (us *URLSource) Configure(options map[string]interface{}) error {
	if userAgent, ok := options["user_agent"].(string); ok {
		us.userAgent = userAgent
	}

	if timeout, ok := options["timeout"].(time.Duration); ok {
		us.httpClient.Timeout = timeout
	}

	if headers, ok := options["headers"].(map[string]string); ok {
		us.headers = headers
	}

	return nil
}

// Close closes the URL source
func (us *URLSource) Close() error {
	return nil
}

func (us *URLSource) fetchDocument(ctx context.Context, url string) (*Document, error) {
	ctx, span := us.tracer.Start(ctx, "url_source.fetch_document")
	defer span.End()

	span.SetAttributes(attribute.String("document.url", url))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", us.userAgent)
	for key, value := range us.headers {
		req.Header.Set(key, value)
	}

	resp, err := us.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", md5.Sum(content))

	// Get content type from response
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	doc := &Document{
		ID:          url,
		Source:      us.GetSourceType(),
		Title:       us.extractTitleFromURL(url),
		Content:     string(content),
		ContentType: contentType,
		Metadata: map[string]interface{}{
			"url":            url,
			"status_code":    resp.StatusCode,
			"content_length": len(content),
			"headers":        resp.Header,
			"fetched_at":     time.Now(),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Size:      int64(len(content)),
		Checksum:  checksum,
	}

	return doc, nil
}

func (us *URLSource) extractTitleFromURL(url string) string {
	// Simple title extraction from URL
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return url
}

// ReaderSource implements DocumentSource for io.Reader sources
type ReaderSource struct {
	readers []ReaderInfo
	logger  *logrus.Logger
	tracer  trace.Tracer
}

// ReaderInfo contains information about a reader
type ReaderInfo struct {
	Reader      io.Reader
	ID          string
	ContentType string
	Metadata    map[string]interface{}
}

// NewReaderSource creates a new reader source
func NewReaderSource(logger *logrus.Logger) *ReaderSource {
	return &ReaderSource{
		readers: make([]ReaderInfo, 0),
		logger:  logger,
		tracer:  otel.Tracer("docprocessing.sources.reader"),
	}
}

// AddReader adds a reader to the source
func (rs *ReaderSource) AddReader(reader io.Reader, contentType string, metadata map[string]interface{}) string {
	id := uuid.New().String()

	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	rs.readers = append(rs.readers, ReaderInfo{
		Reader:      reader,
		ID:          id,
		ContentType: contentType,
		Metadata:    metadata,
	})

	return id
}

// GetDocuments retrieves documents from readers
func (rs *ReaderSource) GetDocuments(ctx context.Context) (<-chan *Document, error) {
	ctx, span := rs.tracer.Start(ctx, "reader_source.get_documents")
	defer span.End()

	span.SetAttributes(attribute.Int("source.reader_count", len(rs.readers)))

	docChan := make(chan *Document, 10)

	go func() {
		defer close(docChan)

		for _, readerInfo := range rs.readers {
			doc, err := rs.createDocumentFromReader(readerInfo)
			if err != nil {
				rs.logger.WithError(err).WithField("reader_id", readerInfo.ID).Error("Failed to create document from reader")
				continue
			}

			select {
			case docChan <- doc:
			case <-ctx.Done():
				return
			}
		}
	}()

	return docChan, nil
}

// GetDocument retrieves a specific document by reader ID
func (rs *ReaderSource) GetDocument(ctx context.Context, id string) (*Document, error) {
	for _, readerInfo := range rs.readers {
		if readerInfo.ID == id {
			return rs.createDocumentFromReader(readerInfo)
		}
	}
	return nil, fmt.Errorf("reader not found: %s", id)
}

// GetSourceType returns the source type
func (rs *ReaderSource) GetSourceType() string {
	return "reader"
}

// Configure configures the reader source
func (rs *ReaderSource) Configure(options map[string]interface{}) error {
	// Reader source doesn't need configuration
	return nil
}

// Close closes the reader source
func (rs *ReaderSource) Close() error {
	// Close any closeable readers
	for _, readerInfo := range rs.readers {
		if closer, ok := readerInfo.Reader.(io.Closer); ok {
			closer.Close()
		}
	}
	return nil
}

func (rs *ReaderSource) createDocumentFromReader(readerInfo ReaderInfo) (*Document, error) {
	content, err := io.ReadAll(readerInfo.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	// Calculate checksum
	checksum := fmt.Sprintf("%x", md5.Sum(content))

	// Merge metadata
	metadata := make(map[string]interface{})
	for k, v := range readerInfo.Metadata {
		metadata[k] = v
	}
	metadata["reader_id"] = readerInfo.ID
	metadata["content_length"] = len(content)

	doc := &Document{
		ID:          readerInfo.ID,
		Source:      rs.GetSourceType(),
		Title:       fmt.Sprintf("Reader-%s", readerInfo.ID[:8]),
		Content:     string(content),
		ContentType: readerInfo.ContentType,
		Metadata:    metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Size:        int64(len(content)),
		Checksum:    checksum,
	}

	return doc, nil
}
