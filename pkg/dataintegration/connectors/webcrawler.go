package connectors

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/aios/aios/pkg/dataintegration"
)

// WebCrawlerConnector implements the DataConnector interface for web crawling
type WebCrawlerConnector struct {
	client    *http.Client
	connected bool
	config    *CrawlerConfig
}

// CrawlerConfig contains web crawler configuration
type CrawlerConfig struct {
	UserAgent       string        `json:"user_agent"`
	MaxDepth        int           `json:"max_depth"`
	MaxPages        int           `json:"max_pages"`
	Delay           time.Duration `json:"delay"`
	Timeout         time.Duration `json:"timeout"`
	FollowRedirects bool          `json:"follow_redirects"`
	RespectRobots   bool          `json:"respect_robots"`
	Selectors       []string      `json:"selectors"`
	ExcludePatterns []string      `json:"exclude_patterns"`
	IncludePatterns []string      `json:"include_patterns"`
}

// NewWebCrawlerConnector creates a new web crawler connector
func NewWebCrawlerConnector() *WebCrawlerConnector {
	return &WebCrawlerConnector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: &CrawlerConfig{
			UserAgent:       "AIOS-WebCrawler/1.0",
			MaxDepth:        3,
			MaxPages:        100,
			Delay:           1 * time.Second,
			Timeout:         30 * time.Second,
			FollowRedirects: true,
			RespectRobots:   true,
			Selectors:       []string{"body"},
		},
	}
}

// Connector Information

// GetType returns the connector type
func (wc *WebCrawlerConnector) GetType() string {
	return "webcrawler"
}

// GetName returns the connector name
func (wc *WebCrawlerConnector) GetName() string {
	return "Web Crawler"
}

// GetDescription returns the connector description
func (wc *WebCrawlerConnector) GetDescription() string {
	return "Web crawler for extracting content from websites with configurable rules and selectors"
}

// GetVersion returns the connector version
func (wc *WebCrawlerConnector) GetVersion() string {
	return "1.0.0"
}

// GetSupportedOperations returns the list of supported operations
func (wc *WebCrawlerConnector) GetSupportedOperations() []string {
	return []string{
		"crawl_website",
		"extract_page",
		"get_links",
		"extract_content",
		"check_robots",
		"get_sitemap",
	}
}

// Configuration

// GetConfigSchema returns the configuration schema
func (wc *WebCrawlerConnector) GetConfigSchema() *dataintegration.ConnectorConfigSchema {
	return &dataintegration.ConnectorConfigSchema{
		Properties: map[string]*dataintegration.ConfigProperty{
			"base_url": {
				Type:        "string",
				Description: "Base URL to start crawling from",
			},
			"user_agent": {
				Type:        "string",
				Description: "User agent string for HTTP requests",
				Default:     "AIOS-WebCrawler/1.0",
			},
			"max_depth": {
				Type:        "integer",
				Description: "Maximum crawling depth",
				Default:     3,
			},
			"max_pages": {
				Type:        "integer",
				Description: "Maximum number of pages to crawl",
				Default:     100,
			},
			"delay": {
				Type:        "string",
				Description: "Delay between requests (e.g., '1s', '500ms')",
				Default:     "1s",
			},
			"selectors": {
				Type:        "array",
				Description: "CSS selectors for content extraction",
				Default:     []string{"body"},
			},
			"exclude_patterns": {
				Type:        "array",
				Description: "URL patterns to exclude from crawling",
			},
			"include_patterns": {
				Type:        "array",
				Description: "URL patterns to include in crawling",
			},
			"respect_robots": {
				Type:        "boolean",
				Description: "Whether to respect robots.txt",
				Default:     true,
			},
		},
		Required: []string{"base_url"},
	}
}

// ValidateConfig validates the configuration
func (wc *WebCrawlerConnector) ValidateConfig(config map[string]interface{}) error {
	if baseURL, exists := config["base_url"]; !exists || baseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	// Validate URL format
	if baseURL, exists := config["base_url"]; exists {
		if _, err := url.Parse(baseURL.(string)); err != nil {
			return fmt.Errorf("invalid base_url format: %w", err)
		}
	}

	// Validate numeric values
	if maxDepth, exists := config["max_depth"]; exists {
		if depth, ok := maxDepth.(float64); ok && depth < 0 {
			return fmt.Errorf("max_depth must be non-negative")
		}
	}

	if maxPages, exists := config["max_pages"]; exists {
		if pages, ok := maxPages.(float64); ok && pages <= 0 {
			return fmt.Errorf("max_pages must be positive")
		}
	}

	return nil
}

// Connection Management

// Connect establishes a connection (validates configuration)
func (wc *WebCrawlerConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	if err := wc.ValidateConfig(config); err != nil {
		return err
	}

	// Update configuration
	if userAgent, exists := config["user_agent"]; exists {
		wc.config.UserAgent = userAgent.(string)
	}

	if maxDepth, exists := config["max_depth"]; exists {
		if depth, ok := maxDepth.(float64); ok {
			wc.config.MaxDepth = int(depth)
		}
	}

	if maxPages, exists := config["max_pages"]; exists {
		if pages, ok := maxPages.(float64); ok {
			wc.config.MaxPages = int(pages)
		}
	}

	if delay, exists := config["delay"]; exists {
		if delayStr, ok := delay.(string); ok {
			if d, err := time.ParseDuration(delayStr); err == nil {
				wc.config.Delay = d
			}
		}
	}

	if selectors, exists := config["selectors"]; exists {
		if selectorList, ok := selectors.([]interface{}); ok {
			wc.config.Selectors = make([]string, len(selectorList))
			for i, selector := range selectorList {
				wc.config.Selectors[i] = selector.(string)
			}
		}
	}

	if excludePatterns, exists := config["exclude_patterns"]; exists {
		if patternList, ok := excludePatterns.([]interface{}); ok {
			wc.config.ExcludePatterns = make([]string, len(patternList))
			for i, pattern := range patternList {
				wc.config.ExcludePatterns[i] = pattern.(string)
			}
		}
	}

	if includePatterns, exists := config["include_patterns"]; exists {
		if patternList, ok := includePatterns.([]interface{}); ok {
			wc.config.IncludePatterns = make([]string, len(patternList))
			for i, pattern := range patternList {
				wc.config.IncludePatterns[i] = pattern.(string)
			}
		}
	}

	if respectRobots, exists := config["respect_robots"]; exists {
		if respect, ok := respectRobots.(bool); ok {
			wc.config.RespectRobots = respect
		}
	}

	// Update client timeout
	wc.client.Timeout = wc.config.Timeout

	// Test the connection
	if err := wc.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to web crawler: %w", err)
	}

	wc.connected = true
	return nil
}

// Disconnect closes the connection
func (wc *WebCrawlerConnector) Disconnect(ctx context.Context) error {
	wc.connected = false
	return nil
}

// IsConnected returns whether the connector is connected
func (wc *WebCrawlerConnector) IsConnected() bool {
	return wc.connected
}

// TestConnection tests the connection by making a simple HTTP request
func (wc *WebCrawlerConnector) TestConnection(ctx context.Context) error {
	// Create a simple test request
	req, err := http.NewRequestWithContext(ctx, "HEAD", "https://httpbin.org/status/200", nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", wc.config.UserAgent)

	resp, err := wc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("test request failed with status %d", resp.StatusCode)
	}

	return nil
}

// Data Operations

// ExtractData performs data extraction based on parameters
func (wc *WebCrawlerConnector) ExtractData(ctx context.Context, params *dataintegration.ExtractionParams) (*dataintegration.DataExtraction, error) {
	start := time.Now()

	if !wc.connected {
		return nil, fmt.Errorf("connector not connected")
	}

	// Get base URL from custom parameters
	baseURL, exists := params.Custom["base_url"]
	if !exists {
		return nil, fmt.Errorf("base_url is required in extraction parameters")
	}

	// Perform web crawling
	records, err := wc.crawlWebsite(ctx, baseURL.(string), params)
	if err != nil {
		return nil, fmt.Errorf("failed to crawl website: %w", err)
	}

	return &dataintegration.DataExtraction{
		Records:     records,
		TotalCount:  int64(len(records)),
		HasMore:     false,
		ExtractedAt: time.Now(),
		Duration:    time.Since(start),
		Metadata: map[string]interface{}{
			"base_url":    baseURL,
			"pages_crawled": len(records),
			"max_depth":   wc.config.MaxDepth,
		},
	}, nil
}

// StreamData streams data (not supported for web crawler)
func (wc *WebCrawlerConnector) StreamData(ctx context.Context, params *dataintegration.StreamParams) (<-chan *dataintegration.DataRecord, error) {
	return nil, fmt.Errorf("streaming not supported for web crawler")
}

// Health and Monitoring

// GetHealth returns the connector health status
func (wc *WebCrawlerConnector) GetHealth() *dataintegration.ConnectorHealth {
	status := dataintegration.HealthStatusHealthy
	message := "Web crawler is healthy"

	if !wc.connected {
		status = dataintegration.HealthStatusUnhealthy
		message = "Web crawler not connected"
	}

	return &dataintegration.ConnectorHealth{
		Status:       status,
		Message:      message,
		Connected:    wc.connected,
		LastActivity: time.Now(),
	}
}

// GetMetrics returns connector metrics
func (wc *WebCrawlerConnector) GetMetrics() *dataintegration.ConnectorMetrics {
	return &dataintegration.ConnectorMetrics{
		OperationCount:   make(map[string]int64),
		AverageLatency:   make(map[string]time.Duration),
		ErrorCount:       make(map[string]int64),
		LastOperation:    time.Now(),
		ConnectionUptime: time.Hour, // Placeholder
	}
}

// Helper methods

// crawlWebsite performs the actual web crawling
func (wc *WebCrawlerConnector) crawlWebsite(ctx context.Context, baseURL string, params *dataintegration.ExtractionParams) ([]*dataintegration.DataRecord, error) {
	var records []*dataintegration.DataRecord
	visited := make(map[string]bool)
	toVisit := []string{baseURL}
	depth := 0

	for len(toVisit) > 0 && depth < wc.config.MaxDepth && len(records) < wc.config.MaxPages {
		select {
		case <-ctx.Done():
			return records, ctx.Err()
		default:
		}

		currentURL := toVisit[0]
		toVisit = toVisit[1:]

		if visited[currentURL] {
			continue
		}

		visited[currentURL] = true

		// Add delay between requests
		if wc.config.Delay > 0 {
			time.Sleep(wc.config.Delay)
		}

		// Extract page content
		record, links, err := wc.extractPage(ctx, currentURL)
		if err != nil {
			continue // Skip failed pages
		}

		records = append(records, record)

		// Add new links to visit queue
		for _, link := range links {
			if !visited[link] && wc.shouldVisitURL(link) {
				toVisit = append(toVisit, link)
			}
		}

		depth++
	}

	return records, nil
}

// extractPage extracts content from a single page
func (wc *WebCrawlerConnector) extractPage(ctx context.Context, pageURL string) (*dataintegration.DataRecord, []string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("User-Agent", wc.config.UserAgent)

	resp, err := wc.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	// Extract content using selectors
	content := make(map[string]interface{})
	for _, selector := range wc.config.Selectors {
		doc.Find(selector).Each(func(i int, s *goquery.Selection) {
			text := strings.TrimSpace(s.Text())
			if text != "" {
				key := fmt.Sprintf("%s_%d", selector, i)
				content[key] = text
			}
		})
	}

	// Extract links
	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			if absoluteURL, err := url.Parse(href); err == nil {
				if baseURL, err := url.Parse(pageURL); err == nil {
					resolvedURL := baseURL.ResolveReference(absoluteURL)
					links = append(links, resolvedURL.String())
				}
			}
		}
	})

	// Create data record
	record := &dataintegration.DataRecord{
		ID:       fmt.Sprintf("page_%s", pageURL),
		SourceID: "webcrawler",
		Data: map[string]interface{}{
			"url":         pageURL,
			"title":       doc.Find("title").Text(),
			"content":     content,
			"links_count": len(links),
			"status_code": resp.StatusCode,
		},
		Metadata: map[string]interface{}{
			"content_type":   resp.Header.Get("Content-Type"),
			"content_length": resp.Header.Get("Content-Length"),
			"last_modified":  resp.Header.Get("Last-Modified"),
		},
		Timestamp: time.Now(),
	}

	return record, links, nil
}

// shouldVisitURL checks if a URL should be visited based on include/exclude patterns
func (wc *WebCrawlerConnector) shouldVisitURL(urlStr string) bool {
	// Check exclude patterns
	for _, pattern := range wc.config.ExcludePatterns {
		if strings.Contains(urlStr, pattern) {
			return false
		}
	}

	// Check include patterns (if any)
	if len(wc.config.IncludePatterns) > 0 {
		for _, pattern := range wc.config.IncludePatterns {
			if strings.Contains(urlStr, pattern) {
				return true
			}
		}
		return false
	}

	return true
}
