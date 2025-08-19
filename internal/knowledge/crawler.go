package knowledge

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aios/aios/pkg/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// WebCrawler handles web crawling operations
type WebCrawler struct {
	config     *config.Config
	logger     *logrus.Logger
	tracer     trace.Tracer
	repository *Repository
	httpClient *http.Client
	jobs       map[string]*CrawlJobRuntime
	jobsMutex  sync.RWMutex
}

// CrawlJobRuntime represents runtime state for an active crawling job
type CrawlJobRuntime struct {
	*CrawlJob
	Results []CrawlResult
	ctx     context.Context
	cancel  context.CancelFunc
}

// CrawlResult represents a single crawled page
type CrawlResult struct {
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Content     string            `json:"content"`
	Links       []string          `json:"links"`
	Metadata    map[string]string `json:"metadata"`
	CrawledAt   time.Time         `json:"crawled_at"`
	StatusCode  int               `json:"status_code"`
	ContentType string            `json:"content_type"`
	Size        int               `json:"size"`
}

// Source represents a crawled source
type Source struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	LastCrawled time.Time         `json:"last_crawled"`
	PageCount   int               `json:"page_count"`
	Metadata    map[string]string `json:"metadata"`
}

// NewWebCrawler creates a new web crawler instance
func NewWebCrawler(config *config.Config, repository *Repository, logger *logrus.Logger) (*WebCrawler, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	return &WebCrawler{
		config:     config,
		logger:     logger,
		tracer:     otel.Tracer("knowledge.crawler"),
		repository: repository,
		httpClient: httpClient,
		jobs:       make(map[string]*CrawlJobRuntime),
	}, nil
}

// Start starts the web crawler
func (c *WebCrawler) Start(ctx context.Context) error {
	c.logger.Info("Starting Web Crawler...")
	return nil
}

// Stop stops the web crawler
func (c *WebCrawler) Stop(ctx context.Context) error {
	c.logger.Info("Stopping Web Crawler...")

	// Cancel all active jobs
	c.jobsMutex.Lock()
	defer c.jobsMutex.Unlock()

	for _, job := range c.jobs {
		if job.cancel != nil {
			job.cancel()
		}
	}

	return nil
}

// StartCrawl starts a new crawling job
func (c *WebCrawler) StartCrawl(ctx context.Context, req *CrawlRequest) (string, error) {
	ctx, span := c.tracer.Start(ctx, "crawler.start_crawl")
	defer span.End()

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Create job
	jobID := uuid.New()
	jobCtx, cancel := context.WithCancel(ctx)

	// Convert metadata
	metadata := make(map[string]interface{})
	for k, v := range req.Metadata {
		metadata[k] = v
	}

	// Create database job record
	dbJob := &CrawlJob{
		ID:              jobID,
		KnowledgeBaseID: uuid.New(), // TODO: Get from request or context
		URL:             parsedURL.String(),
		Status:          "running",
		MaxPages:        req.MaxPages,
		MaxDepth:        req.MaxDepth,
		FollowLinks:     req.FollowLinks,
		Metadata:        metadata,
		PagesFound:      0,
		PagesProcessed:  0,
	}

	// Create runtime job
	job := &CrawlJobRuntime{
		CrawlJob: dbJob,
		Results:  make([]CrawlResult, 0),
		ctx:      jobCtx,
		cancel:   cancel,
	}

	// Set defaults
	if job.MaxPages == 0 {
		job.MaxPages = 100
	}
	if job.MaxDepth == 0 {
		job.MaxDepth = 3
	}

	// Store job in database
	if err := c.repository.CreateCrawlJob(ctx, dbJob); err != nil {
		return "", fmt.Errorf("failed to create crawl job: %w", err)
	}

	// Store job in memory
	c.jobsMutex.Lock()
	c.jobs[jobID.String()] = job
	c.jobsMutex.Unlock()

	// Start crawling in goroutine
	go c.runCrawlJob(job)

	c.logger.WithFields(logrus.Fields{
		"job_id": jobID.String(),
		"url":    req.URL,
	}).Info("Crawl job started")

	return jobID.String(), nil
}

// runCrawlJob executes a crawling job
func (c *WebCrawler) runCrawlJob(job *CrawlJobRuntime) {
	defer func() {
		endTime := time.Now()
		job.CompletedAt = &endTime
		if job.Status == "running" {
			job.Status = "completed"
		}

		// Update job in database
		c.repository.UpdateCrawlJob(context.Background(), job.CrawlJob)
	}()

	c.logger.WithField("job_id", job.ID).Info("Starting crawl job execution")

	// Update job status
	startTime := time.Now()
	job.StartedAt = &startTime
	job.Status = "running"

	// Crawl the initial URL
	visited := make(map[string]bool)
	queue := []string{job.URL}
	depth := 0

	for len(queue) > 0 && job.PagesFound < job.MaxPages && depth <= job.MaxDepth {
		select {
		case <-job.ctx.Done():
			job.Status = "cancelled"
			job.ErrorMessage = stringPtr("job cancelled")
			return
		default:
		}

		currentURL := queue[0]
		queue = queue[1:]

		if visited[currentURL] {
			continue
		}

		visited[currentURL] = true

		// Crawl the page
		result, err := c.crawlPage(job.ctx, currentURL)
		if err != nil {
			c.logger.WithError(err).WithField("url", currentURL).Warn("Failed to crawl page")
			continue
		}

		job.Results = append(job.Results, *result)
		job.PagesFound++

		// Add links to queue if following links is enabled
		if job.FollowLinks && depth < job.MaxDepth {
			for _, link := range result.Links {
				if !visited[link] && c.shouldFollowLink(job.URL, link) {
					queue = append(queue, link)
				}
			}
		}

		depth++
	}

	c.logger.WithFields(logrus.Fields{
		"job_id":      job.ID,
		"pages_found": job.PagesFound,
	}).Info("Crawl job completed")
}

// crawlPage crawls a single page
func (c *WebCrawler) crawlPage(ctx context.Context, pageURL string) (*CrawlResult, error) {
	ctx, span := c.tracer.Start(ctx, "crawler.crawl_page")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "AIOS-Crawler/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract content
	title := doc.Find("title").First().Text()
	content := c.extractTextContent(doc)
	links := c.extractLinks(doc, pageURL)

	result := &CrawlResult{
		URL:         pageURL,
		Title:       strings.TrimSpace(title),
		Content:     content,
		Links:       links,
		Metadata:    make(map[string]string),
		CrawledAt:   time.Now(),
		StatusCode:  resp.StatusCode,
		ContentType: resp.Header.Get("Content-Type"),
		Size:        len(content),
	}

	// Extract metadata
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, exists := s.Attr("name"); exists {
			if content, exists := s.Attr("content"); exists {
				result.Metadata[name] = content
			}
		}
		if property, exists := s.Attr("property"); exists {
			if content, exists := s.Attr("content"); exists {
				result.Metadata[property] = content
			}
		}
	})

	return result, nil
}

// extractTextContent extracts clean text content from HTML
func (c *WebCrawler) extractTextContent(doc *goquery.Document) string {
	// Remove script and style elements
	doc.Find("script, style, nav, footer, header").Remove()

	// Extract text from main content areas
	var content strings.Builder

	// Try to find main content
	mainSelectors := []string{"main", "article", ".content", "#content", ".main", "#main"}
	for _, selector := range mainSelectors {
		if mainContent := doc.Find(selector).First(); mainContent.Length() > 0 {
			content.WriteString(mainContent.Text())
			break
		}
	}

	// If no main content found, use body
	if content.Len() == 0 {
		content.WriteString(doc.Find("body").Text())
	}

	// Clean up whitespace
	text := strings.TrimSpace(content.String())
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return text
}

// extractLinks extracts all links from the page
func (c *WebCrawler) extractLinks(doc *goquery.Document, baseURL string) []string {
	var links []string
	seen := make(map[string]bool)

	base, err := url.Parse(baseURL)
	if err != nil {
		return links
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Resolve relative URLs
		linkURL, err := base.Parse(href)
		if err != nil {
			return
		}

		absoluteURL := linkURL.String()
		if !seen[absoluteURL] {
			seen[absoluteURL] = true
			links = append(links, absoluteURL)
		}
	})

	return links
}

// shouldFollowLink determines if a link should be followed
func (c *WebCrawler) shouldFollowLink(baseURL, linkURL string) bool {
	base, err := url.Parse(baseURL)
	if err != nil {
		return false
	}

	link, err := url.Parse(linkURL)
	if err != nil {
		return false
	}

	// Only follow links from the same domain
	if base.Host != link.Host {
		return false
	}

	// Skip certain file types
	skipExtensions := []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".zip", ".tar", ".gz"}
	for _, ext := range skipExtensions {
		if strings.HasSuffix(strings.ToLower(link.Path), ext) {
			return false
		}
	}

	return true
}

// GetJob returns a crawl job by ID
func (c *WebCrawler) GetJob(jobID string) (*CrawlJob, error) {
	c.jobsMutex.RLock()
	defer c.jobsMutex.RUnlock()

	job, exists := c.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job.CrawlJob, nil
}

// ListSources returns a list of crawled sources
func (c *WebCrawler) ListSources(ctx context.Context) ([]Source, error) {
	// This would typically query a database
	// For now, return sources from active jobs
	c.jobsMutex.RLock()
	defer c.jobsMutex.RUnlock()

	var sources []Source
	for _, job := range c.jobs {
		// Convert metadata
		metadata := make(map[string]string)
		for k, v := range job.Metadata {
			if str, ok := v.(string); ok {
				metadata[k] = str
			}
		}

		// Get last crawled time
		lastCrawled := job.CreatedAt
		if job.StartedAt != nil {
			lastCrawled = *job.StartedAt
		}

		source := Source{
			ID:          job.ID.String(),
			URL:         job.URL,
			Title:       fmt.Sprintf("Crawl Job %s", job.ID.String()[:8]),
			Description: fmt.Sprintf("Crawled %d pages", job.PagesFound),
			Type:        "web_crawl",
			Status:      job.Status,
			LastCrawled: lastCrawled,
			PageCount:   job.PagesFound,
			Metadata:    metadata,
		}
		sources = append(sources, source)
	}

	return sources, nil
}
