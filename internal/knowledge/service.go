package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aios/aios/pkg/config"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Service provides knowledge management functionality
type Service struct {
	config     *config.Config
	logger     *logrus.Logger
	tracer     trace.Tracer
	db         *sqlx.DB
	repository *Repository
	crawler    *WebCrawler
	processor  *DocumentProcessor
	searcher   *VectorSearcher
	httpServer *http.Server
}

// CrawlRequest represents a web crawling request
type CrawlRequest struct {
	URL         string            `json:"url"`
	MaxPages    int               `json:"max_pages,omitempty"`
	MaxDepth    int               `json:"max_depth,omitempty"`
	FollowLinks bool              `json:"follow_links,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// CrawlResponse represents a web crawling response
type CrawlResponse struct {
	JobID      string `json:"job_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	PagesFound int    `json:"pages_found,omitempty"`
}

// DocumentUploadRequest represents a document upload request
type DocumentUploadRequest struct {
	FileName string            `json:"file_name"`
	Content  string            `json:"content"`
	MimeType string            `json:"mime_type"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// SearchRequest represents a knowledge search request
type SearchRequest struct {
	Query       string            `json:"query"`
	MaxResults  int               `json:"max_results,omitempty"`
	Filters     map[string]string `json:"filters,omitempty"`
	UseRAG      bool              `json:"use_rag,omitempty"`
	ContextSize int               `json:"context_size,omitempty"`
}

// SearchResponse represents a knowledge search response
type SearchResponse struct {
	Results []SearchResult `json:"results"`
	Total   int            `json:"total"`
	Query   string         `json:"query"`
}

// SearchResult represents a single search result
type SearchResult struct {
	ID       string            `json:"id"`
	Title    string            `json:"title"`
	Content  string            `json:"content"`
	URL      string            `json:"url,omitempty"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// NewService creates a new knowledge service instance
func NewService(config *config.Config, db *sqlx.DB, logger *logrus.Logger) (*Service, error) {
	tracer := otel.Tracer("knowledge.service")

	// Create repository
	repository := NewRepository(db, logger)

	// Initialize components
	crawler, err := NewWebCrawler(config, repository, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create web crawler: %w", err)
	}

	processor, err := NewDocumentProcessor(config, repository, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create document processor: %w", err)
	}

	searcher, err := NewVectorSearcher(config, repository, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector searcher: %w", err)
	}

	return &Service{
		config:     config,
		logger:     logger,
		tracer:     tracer,
		db:         db,
		repository: repository,
		crawler:    crawler,
		processor:  processor,
		searcher:   searcher,
	}, nil
}

// Start starts the knowledge service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting Knowledge Service...")

	// Start components
	if err := s.crawler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start web crawler: %w", err)
	}

	if err := s.processor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start document processor: %w", err)
	}

	if err := s.searcher.Start(ctx); err != nil {
		return fmt.Errorf("failed to start vector searcher: %w", err)
	}

	// Setup HTTP server
	router := mux.NewRouter()
	s.setupRoutes(router)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Services.Knowledge.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		s.logger.WithField("port", s.config.Services.Knowledge.Port).Info("Knowledge service HTTP server starting")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server error")
		}
	}()

	s.logger.Info("Knowledge Service started successfully")
	return nil
}

// Stop stops the knowledge service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Knowledge Service...")

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
		}
	}

	// Stop components
	if err := s.searcher.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Error stopping vector searcher")
	}

	if err := s.processor.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Error stopping document processor")
	}

	if err := s.crawler.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Error stopping web crawler")
	}

	s.logger.Info("Knowledge Service stopped")
	return nil
}

// setupRoutes sets up HTTP routes
func (s *Service) setupRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Knowledge management endpoints
	router.HandleFunc("/crawl", s.crawlHandler).Methods("POST")
	router.HandleFunc("/upload", s.uploadHandler).Methods("POST")
	router.HandleFunc("/search", s.searchHandler).Methods("POST")
	router.HandleFunc("/documents", s.documentsHandler).Methods("GET")
	router.HandleFunc("/sources", s.sourcesHandler).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws", s.websocketHandler)
}

// healthHandler handles health check requests
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// crawlHandler handles web crawling requests
func (s *Service) crawlHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "knowledge.crawl")
	defer span.End()

	var req CrawlRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start crawling job
	jobID, err := s.crawler.StartCrawl(ctx, &req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start crawling")
		http.Error(w, "Failed to start crawling", http.StatusInternalServerError)
		return
	}

	response := CrawlResponse{
		JobID:   jobID,
		Status:  "started",
		Message: "Crawling job started successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// uploadHandler handles document upload requests
func (s *Service) uploadHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "knowledge.upload")
	defer span.End()

	var req DocumentUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Get knowledge base ID from request or context
	knowledgeBaseID := uuid.New() // This should come from the request

	// Process document
	docID, err := s.processor.ProcessDocument(ctx, &req, knowledgeBaseID)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process document")
		http.Error(w, "Failed to process document", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"document_id": docID,
		"status":      "processed",
		"message":     "Document uploaded and processed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// searchHandler handles knowledge search requests
func (s *Service) searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "knowledge.search")
	defer span.End()

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Perform search
	results, err := s.searcher.Search(ctx, &req)
	if err != nil {
		s.logger.WithError(err).Error("Failed to perform search")
		http.Error(w, "Failed to perform search", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// documentsHandler handles document listing requests
func (s *Service) documentsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "knowledge.documents")
	defer span.End()

	// TODO: Get knowledge base ID from request or context
	knowledgeBaseID := uuid.New() // This should come from the request
	limit := 50                   // Default limit
	offset := 0                   // Default offset

	documents, err := s.processor.ListDocuments(ctx, knowledgeBaseID, limit, offset)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list documents")
		http.Error(w, "Failed to list documents", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// sourcesHandler handles knowledge sources listing requests
func (s *Service) sourcesHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "knowledge.sources")
	defer span.End()

	sources, err := s.crawler.ListSources(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to list sources")
		http.Error(w, "Failed to list sources", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

// SearchKnowledge performs a knowledge search
func (s *Service) SearchKnowledge(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	ctx, span := s.tracer.Start(ctx, "knowledge.search_knowledge")
	defer span.End()

	// Perform search using the searcher
	results, err := s.searcher.Search(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	return results, nil
}

// websocketHandler handles WebSocket connections for real-time updates
func (s *Service) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// WebSocket implementation for real-time updates
	// This would handle real-time progress updates for crawling and processing
	s.logger.Info("WebSocket connection requested")
	// TODO: Implement WebSocket handler
}
