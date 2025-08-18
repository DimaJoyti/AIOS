package docprocessing

import (
	"context"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TextExtractor implements ContentExtractor for plain text documents
type TextExtractor struct {
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewTextExtractor creates a new text extractor
func NewTextExtractor(logger *logrus.Logger) *TextExtractor {
	return &TextExtractor{
		logger: logger,
		tracer: otel.Tracer("docprocessing.extractors.text"),
	}
}

// Extract extracts content from plain text documents
func (te *TextExtractor) Extract(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := te.tracer.Start(ctx, "text_extractor.extract")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
	)

	// For plain text, content is already extracted
	// Just clean up and normalize if needed
	extractedDoc := *doc

	// Basic text cleaning
	content := strings.TrimSpace(doc.Content)

	// Remove excessive whitespace
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")

	// Remove control characters except newlines and tabs
	content = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`).ReplaceAllString(content, "")

	extractedDoc.Content = content

	// Add extraction metadata
	if extractedDoc.Metadata == nil {
		extractedDoc.Metadata = make(map[string]interface{})
	}
	extractedDoc.Metadata["extracted_by"] = "text_extractor"
	extractedDoc.Metadata["original_length"] = len(doc.Content)
	extractedDoc.Metadata["extracted_length"] = len(content)

	return &extractedDoc, nil
}

// CanExtract checks if the extractor can handle the document type
func (te *TextExtractor) CanExtract(contentType string) bool {
	textTypes := []string{
		"text/plain",
		"text/csv",
		"application/json",
		"application/xml",
		"text/xml",
	}

	for _, t := range textTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}

// GetSupportedTypes returns supported content types
func (te *TextExtractor) GetSupportedTypes() []string {
	return []string{
		"text/plain",
		"text/csv",
		"application/json",
		"application/xml",
		"text/xml",
	}
}

// HTMLExtractor implements ContentExtractor for HTML documents
type HTMLExtractor struct {
	preserveFormatting    bool
	enableLinkExtraction  bool
	enableImageExtraction bool
	logger                *logrus.Logger
	tracer                trace.Tracer
}

// NewHTMLExtractor creates a new HTML extractor
func NewHTMLExtractor(logger *logrus.Logger) *HTMLExtractor {
	return &HTMLExtractor{
		preserveFormatting:    false,
		enableLinkExtraction:  true,
		enableImageExtraction: true,
		logger:                logger,
		tracer:                otel.Tracer("docprocessing.extractors.html"),
	}
}

// Extract extracts content from HTML documents
func (he *HTMLExtractor) Extract(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := he.tracer.Start(ctx, "html_extractor.extract")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
	)

	extractedDoc := *doc

	// Extract title from HTML
	title := he.extractTitle(doc.Content)
	if title != "" {
		extractedDoc.Title = title
	}

	// Extract text content
	textContent := he.extractTextContent(doc.Content)
	extractedDoc.Content = textContent

	// Extract metadata
	metadata := make(map[string]interface{})
	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	metadata["extracted_by"] = "html_extractor"
	metadata["original_length"] = len(doc.Content)
	metadata["extracted_length"] = len(textContent)

	if he.enableLinkExtraction {
		links := he.extractLinks(doc.Content)
		metadata["links"] = links
		metadata["link_count"] = len(links)
	}

	if he.enableImageExtraction {
		images := he.extractImages(doc.Content)
		metadata["images"] = images
		metadata["image_count"] = len(images)
	}

	// Extract meta tags
	metaTags := he.extractMetaTags(doc.Content)
	metadata["meta_tags"] = metaTags

	extractedDoc.Metadata = metadata

	return &extractedDoc, nil
}

// CanExtract checks if the extractor can handle the document type
func (he *HTMLExtractor) CanExtract(contentType string) bool {
	htmlTypes := []string{
		"text/html",
		"application/xhtml+xml",
	}

	for _, t := range htmlTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}

	return false
}

// GetSupportedTypes returns supported content types
func (he *HTMLExtractor) GetSupportedTypes() []string {
	return []string{
		"text/html",
		"application/xhtml+xml",
	}
}

func (he *HTMLExtractor) extractTitle(html string) string {
	titleRegex := regexp.MustCompile(`<title[^>]*>([^<]+)</title>`)
	matches := titleRegex.FindStringSubmatch(html)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	return ""
}

func (he *HTMLExtractor) extractTextContent(html string) string {
	// Remove script and style tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	html = scriptRegex.ReplaceAllString(html, "")

	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	html = styleRegex.ReplaceAllString(html, "")

	// Remove HTML comments
	commentRegex := regexp.MustCompile(`<!--.*?-->`)
	html = commentRegex.ReplaceAllString(html, "")

	// Remove HTML tags
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	text := tagRegex.ReplaceAllString(html, " ")

	// Decode HTML entities (basic ones)
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

func (he *HTMLExtractor) extractLinks(html string) []string {
	linkRegex := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["'][^>]*>`)
	matches := linkRegex.FindAllStringSubmatch(html, -1)

	links := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])
		}
	}

	return links
}

func (he *HTMLExtractor) extractImages(html string) []string {
	imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["'][^>]*>`)
	matches := imgRegex.FindAllStringSubmatch(html, -1)

	images := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			images = append(images, match[1])
		}
	}

	return images
}

func (he *HTMLExtractor) extractMetaTags(html string) map[string]string {
	metaTags := make(map[string]string)

	// Extract meta tags with name attribute
	nameRegex := regexp.MustCompile(`<meta[^>]+name=["']([^"']+)["'][^>]+content=["']([^"']+)["'][^>]*>`)
	matches := nameRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 2 {
			metaTags[match[1]] = match[2]
		}
	}

	// Extract meta tags with property attribute (Open Graph, etc.)
	propertyRegex := regexp.MustCompile(`<meta[^>]+property=["']([^"']+)["'][^>]+content=["']([^"']+)["'][^>]*>`)
	matches = propertyRegex.FindAllStringSubmatch(html, -1)
	for _, match := range matches {
		if len(match) > 2 {
			metaTags[match[1]] = match[2]
		}
	}

	return metaTags
}

// MarkdownExtractor implements ContentExtractor for Markdown documents
type MarkdownExtractor struct {
	preserveFormatting     bool
	enableHeaderExtraction bool
	enableLinkExtraction   bool
	logger                 *logrus.Logger
	tracer                 trace.Tracer
}

// NewMarkdownExtractor creates a new Markdown extractor
func NewMarkdownExtractor(logger *logrus.Logger) *MarkdownExtractor {
	return &MarkdownExtractor{
		preserveFormatting:     true,
		enableHeaderExtraction: true,
		enableLinkExtraction:   true,
		logger:                 logger,
		tracer:                 otel.Tracer("docprocessing.extractors.markdown"),
	}
}

// Extract extracts content from Markdown documents
func (me *MarkdownExtractor) Extract(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := me.tracer.Start(ctx, "markdown_extractor.extract")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
	)

	extractedDoc := *doc

	content := doc.Content

	// Extract title from first header if not already set
	if doc.Title == "" || doc.Title == doc.ID {
		title := me.extractTitle(content)
		if title != "" {
			extractedDoc.Title = title
		}
	}

	// Extract metadata
	metadata := make(map[string]interface{})
	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	metadata["extracted_by"] = "markdown_extractor"
	metadata["original_length"] = len(doc.Content)

	if me.enableHeaderExtraction {
		headers := me.extractHeaders(content)
		metadata["headers"] = headers
		metadata["header_count"] = len(headers)
	}

	if me.enableLinkExtraction {
		links := me.extractMarkdownLinks(content)
		metadata["links"] = links
		metadata["link_count"] = len(links)
	}

	// Extract front matter if present
	frontMatter := me.extractFrontMatter(content)
	if len(frontMatter) > 0 {
		metadata["front_matter"] = frontMatter
		// Remove front matter from content
		content = me.removeFrontMatter(content)
	}

	if !me.preserveFormatting {
		// Convert to plain text
		content = me.convertToPlainText(content)
	}

	metadata["extracted_length"] = len(content)
	extractedDoc.Content = content
	extractedDoc.Metadata = metadata

	return &extractedDoc, nil
}

// CanExtract checks if the extractor can handle the document type
func (me *MarkdownExtractor) CanExtract(contentType string) bool {
	return strings.Contains(contentType, "text/markdown") ||
		strings.Contains(contentType, "text/x-markdown")
}

// GetSupportedTypes returns supported content types
func (me *MarkdownExtractor) GetSupportedTypes() []string {
	return []string{
		"text/markdown",
		"text/x-markdown",
	}
}

func (me *MarkdownExtractor) extractTitle(content string) string {
	// Look for first H1 header
	h1Regex := regexp.MustCompile(`^#\s+(.+)$`)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if matches := h1Regex.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

func (me *MarkdownExtractor) extractHeaders(content string) []map[string]interface{} {
	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	lines := strings.Split(content, "\n")

	headers := make([]map[string]interface{}, 0)

	for i, line := range lines {
		if matches := headerRegex.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 2 {
			level := len(matches[1])
			text := strings.TrimSpace(matches[2])

			headers = append(headers, map[string]interface{}{
				"level": level,
				"text":  text,
				"line":  i + 1,
			})
		}
	}

	return headers
}

func (me *MarkdownExtractor) extractMarkdownLinks(content string) []map[string]string {
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	links := make([]map[string]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 2 {
			links = append(links, map[string]string{
				"text": match[1],
				"url":  match[2],
			})
		}
	}

	return links
}

func (me *MarkdownExtractor) extractFrontMatter(content string) map[string]interface{} {
	// Simple YAML front matter extraction
	frontMatterRegex := regexp.MustCompile(`^---\n(.*?)\n---\n`)
	matches := frontMatterRegex.FindStringSubmatch(content)

	frontMatter := make(map[string]interface{})

	if len(matches) > 1 {
		// Basic YAML parsing (key: value pairs)
		lines := strings.Split(matches[1], "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					frontMatter[key] = value
				}
			}
		}
	}

	return frontMatter
}

func (me *MarkdownExtractor) removeFrontMatter(content string) string {
	frontMatterRegex := regexp.MustCompile(`^---\n.*?\n---\n`)
	return frontMatterRegex.ReplaceAllString(content, "")
}

func (me *MarkdownExtractor) convertToPlainText(content string) string {
	// Remove markdown formatting
	text := content

	// Remove headers
	text = regexp.MustCompile(`^#{1,6}\s+`).ReplaceAllString(text, "")

	// Remove bold and italic
	text = regexp.MustCompile(`\*\*([^*]+)\*\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`\*([^*]+)\*`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`__([^_]+)__`).ReplaceAllString(text, "$1")
	text = regexp.MustCompile(`_([^_]+)_`).ReplaceAllString(text, "$1")

	// Remove links but keep text
	text = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`).ReplaceAllString(text, "$1")

	// Remove code blocks
	text = regexp.MustCompile("```[^`]*```").ReplaceAllString(text, "")
	text = regexp.MustCompile("`([^`]+)`").ReplaceAllString(text, "$1")

	// Remove list markers
	text = regexp.MustCompile(`^\s*[-*+]\s+`).ReplaceAllString(text, "")
	text = regexp.MustCompile(`^\s*\d+\.\s+`).ReplaceAllString(text, "")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}

// JSONExtractor implements ContentExtractor for JSON documents
type JSONExtractor struct {
	extractKeys   []string
	flattenObject bool
	logger        *logrus.Logger
	tracer        trace.Tracer
}

// NewJSONExtractor creates a new JSON extractor
func NewJSONExtractor(logger *logrus.Logger) *JSONExtractor {
	return &JSONExtractor{
		extractKeys:   []string{},
		flattenObject: false,
		logger:        logger,
		tracer:        otel.Tracer("docprocessing.extractors.json"),
	}
}

// Extract extracts content from JSON documents
func (je *JSONExtractor) Extract(ctx context.Context, doc *Document) (*Document, error) {
	ctx, span := je.tracer.Start(ctx, "json_extractor.extract")
	defer span.End()

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.String("document.content_type", doc.ContentType),
	)

	extractedDoc := *doc

	// For JSON, we'll extract text values and create a readable format
	content := je.extractTextFromJSON(doc.Content)
	extractedDoc.Content = content

	// Add extraction metadata
	metadata := make(map[string]interface{})
	for k, v := range doc.Metadata {
		metadata[k] = v
	}

	metadata["extracted_by"] = "json_extractor"
	metadata["original_length"] = len(doc.Content)
	metadata["extracted_length"] = len(content)

	extractedDoc.Metadata = metadata

	return &extractedDoc, nil
}

// CanExtract checks if the extractor can handle the document type
func (je *JSONExtractor) CanExtract(contentType string) bool {
	return strings.Contains(contentType, "application/json")
}

// GetSupportedTypes returns supported content types
func (je *JSONExtractor) GetSupportedTypes() []string {
	return []string{
		"application/json",
	}
}

func (je *JSONExtractor) extractTextFromJSON(jsonContent string) string {
	// Simple JSON text extraction - remove structural characters and extract values
	text := jsonContent

	// Remove JSON structural characters
	text = strings.ReplaceAll(text, "{", " ")
	text = strings.ReplaceAll(text, "}", " ")
	text = strings.ReplaceAll(text, "[", " ")
	text = strings.ReplaceAll(text, "]", " ")
	text = strings.ReplaceAll(text, "\"", " ")
	text = strings.ReplaceAll(text, ",", " ")
	text = strings.ReplaceAll(text, ":", " ")

	// Clean up whitespace
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	return text
}
