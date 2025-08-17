package ai

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// CVService implements the ComputerVisionService interface
type CVService struct {
	config AIServiceConfig
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewCVService creates a new computer vision service
func NewCVService(config AIServiceConfig, logger *logrus.Logger) *CVService {
	tracer := otel.Tracer("cv-service")

	return &CVService{
		config: config,
		logger: logger,
		tracer: tracer,
	}
}

// AnalyzeScreen analyzes a screenshot for UI elements and content
func (s *CVService) AnalyzeScreen(ctx context.Context, screenshot []byte) (*models.ScreenAnalysis, error) {
	ctx, span := s.tracer.Start(ctx, "cv.AnalyzeScreen")
	defer span.End()

	s.logger.WithField("image_size", len(screenshot)).Info("Analyzing screen")

	// Decode image to get dimensions
	img, format, err := s.decodeImage(screenshot)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	s.logger.WithFields(logrus.Fields{
		"width":  bounds.Dx(),
		"height": bounds.Dy(),
		"format": format,
	}).Debug("Image decoded successfully")

	// Detect UI elements
	elements, err := s.detectUIElements(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to detect UI elements: %w", err)
	}

	// Analyze layout
	layout := s.analyzeLayout(img, elements)

	// Extract text regions
	textRegions, err := s.extractTextRegions(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text regions: %w", err)
	}

	// Identify possible actions
	actions := s.identifyPossibleActions(elements)

	// Analyze accessibility
	accessibility := s.analyzeAccessibility(elements, textRegions)

	analysis := &models.ScreenAnalysis{
		Elements:      elements,
		Layout:        layout,
		Text:          textRegions,
		Actions:       actions,
		Accessibility: accessibility,
		Metadata: map[string]interface{}{
			"image_width":        bounds.Dx(),
			"image_height":       bounds.Dy(),
			"image_format":       format,
			"elements_count":     len(elements),
			"text_regions_count": len(textRegions),
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"elements_found": len(elements),
		"text_regions":   len(textRegions),
		"actions":        len(actions),
	}).Info("Screen analysis completed")

	return analysis, nil
}

// DetectUI detects UI elements in an image
func (s *CVService) DetectUI(ctx context.Context, imageData []byte) (*models.UIElements, error) {
	ctx, span := s.tracer.Start(ctx, "cv.DetectUI")
	defer span.End()

	img, _, err := s.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	elements, err := s.detectUIElements(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to detect UI elements: %w", err)
	}

	// Categorize elements by type
	uiElements := &models.UIElements{
		Buttons:    []models.UIElement{},
		TextFields: []models.UIElement{},
		Images:     []models.UIElement{},
		Links:      []models.UIElement{},
		Menus:      []models.UIElement{},
		Windows:    []models.UIElement{},
		Other:      []models.UIElement{},
	}

	for _, element := range elements {
		switch element.Type {
		case "button":
			uiElements.Buttons = append(uiElements.Buttons, element)
		case "textfield", "input":
			uiElements.TextFields = append(uiElements.TextFields, element)
		case "image":
			uiElements.Images = append(uiElements.Images, element)
		case "link":
			uiElements.Links = append(uiElements.Links, element)
		case "menu":
			uiElements.Menus = append(uiElements.Menus, element)
		case "window":
			uiElements.Windows = append(uiElements.Windows, element)
		default:
			uiElements.Other = append(uiElements.Other, element)
		}
	}

	return uiElements, nil
}

// RecognizeText performs OCR on an image
func (s *CVService) RecognizeText(ctx context.Context, imageData []byte) (*models.TextRecognition, error) {
	ctx, span := s.tracer.Start(ctx, "cv.RecognizeText")
	defer span.End()

	s.logger.WithField("image_size", len(imageData)).Info("Performing OCR")

	img, _, err := s.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	textRegions, err := s.extractTextRegions(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to extract text: %w", err)
	}

	// Combine all text regions into a single text
	var fullText bytes.Buffer
	for i, region := range textRegions {
		if i > 0 {
			fullText.WriteString(" ")
		}
		fullText.WriteString(region.Text)
	}

	// Calculate overall confidence
	var totalConfidence float64
	for _, region := range textRegions {
		totalConfidence += region.Confidence
	}
	averageConfidence := totalConfidence / float64(len(textRegions))
	if len(textRegions) == 0 {
		averageConfidence = 0.0
	}

	recognition := &models.TextRecognition{
		Text:       fullText.String(),
		Regions:    textRegions,
		Language:   "en", // TODO: Implement language detection
		Confidence: averageConfidence,
		Timestamp:  time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"text_length": len(recognition.Text),
		"regions":     len(textRegions),
		"confidence":  averageConfidence,
	}).Info("OCR completed")

	return recognition, nil
}

// ClassifyImage classifies the content of an image
func (s *CVService) ClassifyImage(ctx context.Context, imageData []byte) (*models.ImageClassification, error) {
	ctx, span := s.tracer.Start(ctx, "cv.ClassifyImage")
	defer span.End()

	// TODO: Implement actual image classification
	// For now, return mock classification results
	classes := []models.ClassificationResult{
		{Class: "desktop", Confidence: 0.85, Probability: 0.85},
		{Class: "application", Confidence: 0.75, Probability: 0.75},
		{Class: "user_interface", Confidence: 0.90, Probability: 0.90},
	}

	classification := &models.ImageClassification{
		Classes:    classes,
		TopClass:   "user_interface",
		Confidence: 0.90,
		Timestamp:  time.Now(),
	}

	return classification, nil
}

// DetectObjects detects objects in an image
func (s *CVService) DetectObjects(ctx context.Context, imageData []byte) (*models.ObjectDetection, error) {
	ctx, span := s.tracer.Start(ctx, "cv.DetectObjects")
	defer span.End()

	_, _, err := s.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// TODO: Implement actual object detection
	// For now, return mock detection results
	objects := []models.DetectedObject{
		{
			Class:      "window",
			Confidence: 0.92,
			Bounds:     models.Rectangle{X: 100, Y: 50, Width: 800, Height: 600},
			Properties: map[string]interface{}{"title": "Application Window"},
		},
		{
			Class:      "button",
			Confidence: 0.88,
			Bounds:     models.Rectangle{X: 200, Y: 100, Width: 100, Height: 30},
			Properties: map[string]interface{}{"text": "OK"},
		},
	}

	detection := &models.ObjectDetection{
		Objects:   objects,
		Count:     len(objects),
		Timestamp: time.Now(),
	}

	return detection, nil
}

// AnalyzeLayout analyzes the layout structure of a UI
func (s *CVService) AnalyzeLayout(ctx context.Context, imageData []byte) (*models.LayoutAnalysis, error) {
	ctx, span := s.tracer.Start(ctx, "cv.AnalyzeLayout")
	defer span.End()

	img, _, err := s.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	elements, err := s.detectUIElements(ctx, img)
	if err != nil {
		return nil, fmt.Errorf("failed to detect UI elements: %w", err)
	}

	layout := s.analyzeLayout(img, elements)

	// Detect layout patterns
	patterns := s.detectLayoutPatterns(elements)

	// Generate layout suggestions
	suggestions := s.generateLayoutSuggestions(elements, patterns)

	// Build hierarchy
	hierarchy := s.buildLayoutHierarchy(elements)

	// Convert LayoutInfo to LayoutStructure
	structure := models.LayoutStructure{
		Type:      layout.Type,
		Columns:   0, // TODO: Calculate from layout
		Rows:      0, // TODO: Calculate from layout
		Regions:   layout.Regions,
		Alignment: "left", // TODO: Determine from layout
		Spacing:   10,     // TODO: Calculate from layout
	}

	analysis := &models.LayoutAnalysis{
		Structure:   structure,
		Hierarchy:   hierarchy,
		Patterns:    patterns,
		Suggestions: suggestions,
		Metadata: map[string]interface{}{
			"elements_analyzed": len(elements),
			"patterns_found":    len(patterns),
		},
		Timestamp: time.Now(),
	}

	return analysis, nil
}

// GenerateDescription generates a natural language description of an image
func (s *CVService) GenerateDescription(ctx context.Context, imageData []byte) (*models.ImageDescription, error) {
	ctx, span := s.tracer.Start(ctx, "cv.GenerateDescription")
	defer span.End()

	// TODO: Implement actual image description generation
	// This would typically involve a vision-language model
	description := &models.ImageDescription{
		Description: "A computer desktop interface with various application windows and UI elements",
		Details: []string{
			"Multiple windows are visible",
			"There are buttons and text fields",
			"The interface appears to be a modern desktop environment",
		},
		Objects: []string{"window", "button", "text", "menu"},
		Scene:   "desktop_interface",
		Mood:    "professional",
		Colors:  []string{"blue", "white", "gray"},
		Metadata: map[string]interface{}{
			"analysis_method": "computer_vision",
			"confidence":      0.75,
		},
		Timestamp: time.Now(),
	}

	return description, nil
}

// CompareImages compares two images for similarity
func (s *CVService) CompareImages(ctx context.Context, image1, image2 []byte) (*models.ImageComparison, error) {
	ctx, span := s.tracer.Start(ctx, "cv.CompareImages")
	defer span.End()

	// TODO: Implement actual image comparison
	// This would involve feature extraction and similarity calculation
	comparison := &models.ImageComparison{
		Similarity:   0.85,
		Differences:  []models.ImageDifference{},
		MatchRegions: []models.Rectangle{},
		Analysis:     "Images are highly similar with minor differences in UI state",
		Metadata: map[string]interface{}{
			"comparison_method": "feature_matching",
			"algorithm":         "sift",
		},
		Timestamp: time.Now(),
	}

	return comparison, nil
}

// Helper methods

func (s *CVService) decodeImage(imageData []byte) (image.Image, string, error) {
	reader := bytes.NewReader(imageData)

	// Try PNG first
	reader.Seek(0, 0)
	if img, err := png.Decode(reader); err == nil {
		return img, "png", nil
	}

	// Try JPEG
	reader.Seek(0, 0)
	if img, err := jpeg.Decode(reader); err == nil {
		return img, "jpeg", nil
	}

	// Try generic decode
	reader.Seek(0, 0)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("unsupported image format: %w", err)
	}

	return img, format, nil
}

func (s *CVService) detectUIElements(ctx context.Context, img image.Image) ([]models.UIElement, error) {
	// TODO: Implement actual UI element detection
	// This would involve computer vision algorithms to detect buttons, text fields, etc.
	bounds := img.Bounds()

	elements := []models.UIElement{
		{
			ID:         "element-1",
			Type:       "window",
			Text:       "Main Window",
			Bounds:     models.Rectangle{X: 0, Y: 0, Width: bounds.Dx(), Height: bounds.Dy()},
			Confidence: 0.95,
			Properties: map[string]interface{}{"title": "Main Window"},
		},
		{
			ID:         "element-2",
			Type:       "button",
			Text:       "OK",
			Bounds:     models.Rectangle{X: 100, Y: 100, Width: 80, Height: 30},
			Confidence: 0.88,
			Properties: map[string]interface{}{"clickable": true},
		},
	}

	return elements, nil
}

func (s *CVService) extractTextRegions(ctx context.Context, img image.Image) ([]models.TextRegion, error) {
	// TODO: Implement actual OCR
	// This would involve OCR libraries like Tesseract
	regions := []models.TextRegion{
		{
			Text:       "Sample Text",
			Bounds:     models.Rectangle{X: 50, Y: 50, Width: 100, Height: 20},
			Confidence: 0.92,
			Language:   "en",
		},
	}

	return regions, nil
}

func (s *CVService) analyzeLayout(img image.Image, elements []models.UIElement) models.LayoutInfo {
	bounds := img.Bounds()

	return models.LayoutInfo{
		Type:        "desktop",
		Dimensions:  models.Rectangle{X: 0, Y: 0, Width: bounds.Dx(), Height: bounds.Dy()},
		Regions:     []models.Rectangle{},
		Orientation: "landscape",
		Density:     float64(len(elements)) / float64(bounds.Dx()*bounds.Dy()) * 1000000, // elements per million pixels
	}
}

func (s *CVService) identifyPossibleActions(elements []models.UIElement) []models.PossibleAction {
	actions := []models.PossibleAction{}

	for _, element := range elements {
		switch element.Type {
		case "button":
			actions = append(actions, models.PossibleAction{
				Type:        "click",
				Target:      element.ID,
				Description: fmt.Sprintf("Click %s button", element.Text),
				Parameters:  map[string]interface{}{"x": element.Bounds.X, "y": element.Bounds.Y},
				Confidence:  0.9,
			})
		case "textfield", "input":
			actions = append(actions, models.PossibleAction{
				Type:        "type",
				Target:      element.ID,
				Description: fmt.Sprintf("Type text in %s field", element.Text),
				Parameters:  map[string]interface{}{"x": element.Bounds.X, "y": element.Bounds.Y},
				Confidence:  0.85,
			})
		}
	}

	return actions
}

func (s *CVService) analyzeAccessibility(elements []models.UIElement, textRegions []models.TextRegion) models.AccessibilityInfo {
	score := 75.0 // Base score
	issues := []string{}
	suggestions := []string{}

	// Check for text contrast (simplified)
	if len(textRegions) == 0 {
		issues = append(issues, "No text detected")
		score -= 20
	}

	// Check for interactive elements
	interactiveCount := 0
	for _, element := range elements {
		if element.Type == "button" || element.Type == "link" {
			interactiveCount++
		}
	}

	if interactiveCount == 0 {
		issues = append(issues, "No interactive elements detected")
		score -= 15
	}

	suggestions = append(suggestions, "Ensure sufficient color contrast for text")
	suggestions = append(suggestions, "Provide alternative text for images")
	suggestions = append(suggestions, "Use semantic HTML elements")

	return models.AccessibilityInfo{
		Score:       score,
		Issues:      issues,
		Suggestions: suggestions,
		Compliance:  "AA", // WCAG 2.1 AA level
	}
}

func (s *CVService) detectLayoutPatterns(elements []models.UIElement) []models.LayoutPattern {
	// TODO: Implement pattern detection algorithms
	return []models.LayoutPattern{
		{
			Type:        "grid",
			Confidence:  0.8,
			Description: "Grid layout pattern detected",
			Examples:    []models.Rectangle{},
		},
	}
}

func (s *CVService) generateLayoutSuggestions(elements []models.UIElement, patterns []models.LayoutPattern) []models.LayoutSuggestion {
	return []models.LayoutSuggestion{
		{
			Type:        "alignment",
			Description: "Consider aligning elements to improve visual hierarchy",
			Impact:      "Improved user experience and readability",
			Priority:    "medium",
		},
	}
}

func (s *CVService) buildLayoutHierarchy(elements []models.UIElement) []models.LayoutNode {
	nodes := []models.LayoutNode{}

	for _, element := range elements {
		node := models.LayoutNode{
			ID:       element.ID,
			Type:     element.Type,
			Bounds:   element.Bounds,
			Children: []models.LayoutNode{},
		}
		nodes = append(nodes, node)
	}

	return nodes
}

// AdvancedObjectDetection performs advanced object detection with segmentation
func (s *CVService) AdvancedObjectDetection(ctx context.Context, imageData []byte) (*models.ObjectDetectionResponse, error) {
	ctx, span := s.tracer.Start(ctx, "cv.AdvancedObjectDetection")
	defer span.End()

	start := time.Now()
	s.logger.WithField("image_size", len(imageData)).Info("Performing advanced object detection")

	if !s.config.CVEnabled {
		return nil, fmt.Errorf("computer vision service is disabled")
	}

	// TODO: Implement actual advanced object detection using YOLO v8, DETR, or similar
	// This would involve:
	// 1. Loading advanced object detection model
	// 2. Image preprocessing and normalization
	// 3. Running inference with segmentation
	// 4. Post-processing with NMS and confidence filtering

	// Mock implementation with realistic object detection results
	objects := []models.AdvancedDetectedObject{
		{
			ID:         "obj_1",
			Class:      "person",
			Confidence: 0.95,
			BoundingBox: models.BoundingBox{
				X:      100,
				Y:      150,
				Width:  80,
				Height: 200,
			},
			Segmentation: []models.Point{
				{X: 100, Y: 150}, {X: 180, Y: 150}, {X: 180, Y: 350}, {X: 100, Y: 350},
			},
			Attributes: map[string]interface{}{
				"age_group": "adult",
				"gender":    "unknown",
				"pose":      "standing",
			},
		},
		{
			ID:         "obj_2",
			Class:      "car",
			Confidence: 0.88,
			BoundingBox: models.BoundingBox{
				X:      300,
				Y:      200,
				Width:  150,
				Height: 100,
			},
			Segmentation: []models.Point{
				{X: 300, Y: 200}, {X: 450, Y: 200}, {X: 450, Y: 300}, {X: 300, Y: 300},
			},
			Attributes: map[string]interface{}{
				"color": "blue",
				"type":  "sedan",
			},
		},
		{
			ID:         "obj_3",
			Class:      "bicycle",
			Confidence: 0.82,
			BoundingBox: models.BoundingBox{
				X:      50,
				Y:      250,
				Width:  60,
				Height: 80,
			},
			Segmentation: []models.Point{
				{X: 50, Y: 250}, {X: 110, Y: 250}, {X: 110, Y: 330}, {X: 50, Y: 330},
			},
			Attributes: map[string]interface{}{
				"color": "red",
				"type":  "mountain",
			},
		},
	}

	response := &models.ObjectDetectionResponse{
		Objects:        objects,
		TotalCount:     len(objects),
		Model:          "yolo-v8-advanced",
		Confidence:     0.88,
		ProcessingTime: time.Since(start),
		Metadata: map[string]interface{}{
			"image_size":       len(imageData),
			"detection_mode":   "advanced",
			"segmentation":     true,
			"classes_detected": []string{"person", "car", "bicycle"},
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"objects_detected": len(objects),
		"processing_time":  time.Since(start),
		"confidence":       response.Confidence,
	}).Info("Advanced object detection completed")

	return response, nil
}

// GenerateImage generates images from text descriptions
func (s *CVService) GenerateImage(ctx context.Context, prompt string, parameters map[string]interface{}) (*models.ImageGenerationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "cv.GenerateImage")
	defer span.End()

	start := time.Now()
	s.logger.WithField("prompt", prompt).Info("Generating image from text")

	if !s.config.CVEnabled {
		return nil, fmt.Errorf("computer vision service is disabled")
	}

	// TODO: Implement actual image generation using Stable Diffusion, DALL-E, or similar
	// This would involve:
	// 1. Loading image generation model
	// 2. Text encoding and conditioning
	// 3. Diffusion process with sampling
	// 4. Image post-processing and upscaling

	// Extract parameters with defaults
	width := 512
	height := 512
	steps := 20
	guidance := 7.5

	if w, exists := parameters["width"]; exists {
		if wInt, ok := w.(int); ok {
			width = wInt
		}
	}
	if h, exists := parameters["height"]; exists {
		if hInt, ok := h.(int); ok {
			height = hInt
		}
	}
	if s, exists := parameters["steps"]; exists {
		if sInt, ok := s.(int); ok {
			steps = sInt
		}
	}
	if g, exists := parameters["guidance"]; exists {
		if gFloat, ok := g.(float64); ok {
			guidance = gFloat
		}
	}

	// Mock image generation - create dummy image data
	numImages := 1
	if count, exists := parameters["num_images"]; exists {
		if countInt, ok := count.(int); ok {
			numImages = countInt
		}
	}

	images := make([][]byte, numImages)
	for i := 0; i < numImages; i++ {
		// Create mock image data
		imageSize := width * height * 3 // RGB
		imageData := make([]byte, imageSize)
		for j := range imageData {
			imageData[j] = byte((i + j) % 256)
		}
		images[i] = imageData
	}

	response := &models.ImageGenerationResponse{
		Images:   images,
		Prompt:   prompt,
		Model:    "stable-diffusion-xl",
		Width:    width,
		Height:   height,
		Steps:    steps,
		Guidance: guidance,
		Seed:     time.Now().UnixNano(),
		Metadata: map[string]interface{}{
			"processing_time": time.Since(start).Milliseconds(),
			"prompt_length":   len(prompt),
			"num_images":      numImages,
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"prompt":          prompt,
		"images_count":    len(images),
		"dimensions":      fmt.Sprintf("%dx%d", width, height),
		"processing_time": time.Since(start),
	}).Info("Image generation completed")

	return response, nil
}

// AnalyzeVideo analyzes video content for objects, activities, and scenes
func (s *CVService) AnalyzeVideo(ctx context.Context, videoData []byte) (*models.VideoAnalysisResponse, error) {
	ctx, span := s.tracer.Start(ctx, "cv.AnalyzeVideo")
	defer span.End()

	start := time.Now()
	s.logger.WithField("video_size", len(videoData)).Info("Analyzing video content")

	if !s.config.CVEnabled {
		return nil, fmt.Errorf("computer vision service is disabled")
	}

	// TODO: Implement actual video analysis
	// This would involve:
	// 1. Video frame extraction and preprocessing
	// 2. Temporal object tracking
	// 3. Activity recognition
	// 4. Scene segmentation and classification
	// 5. Audio-visual synchronization analysis

	// Mock implementation
	estimatedDuration := time.Duration(len(videoData)/1000000) * time.Second
	frameRate := 30.0

	scenes := []models.VideoScene{
		{
			StartTime:   0,
			EndTime:     estimatedDuration / 3,
			Description: "Opening scene with outdoor environment",
			Objects:     []string{"person", "tree", "building", "sky"},
			Activities:  []string{"walking", "talking"},
			Confidence:  0.92,
		},
		{
			StartTime:   estimatedDuration / 3,
			EndTime:     2 * estimatedDuration / 3,
			Description: "Action sequence with vehicles",
			Objects:     []string{"car", "person", "road", "traffic_light"},
			Activities:  []string{"driving", "running", "crossing"},
			Confidence:  0.88,
		},
		{
			StartTime:   2 * estimatedDuration / 3,
			EndTime:     estimatedDuration,
			Description: "Indoor conversation scene",
			Objects:     []string{"person", "table", "chair", "window"},
			Activities:  []string{"sitting", "talking", "gesturing"},
			Confidence:  0.85,
		},
	}

	allObjects := []string{"person", "tree", "building", "sky", "car", "road", "traffic_light", "table", "chair", "window"}
	allActivities := []string{"walking", "talking", "driving", "running", "crossing", "sitting", "gesturing"}

	response := &models.VideoAnalysisResponse{
		Summary:    fmt.Sprintf("Video analysis of %.1f second clip with %d scenes", estimatedDuration.Seconds(), len(scenes)),
		Scenes:     scenes,
		Objects:    allObjects,
		Activities: allActivities,
		Duration:   estimatedDuration,
		FrameRate:  frameRate,
		Resolution: "1920x1080",
		Metadata: map[string]interface{}{
			"processing_time":  time.Since(start).Milliseconds(),
			"video_size":       len(videoData),
			"scenes_count":     len(scenes),
			"objects_count":    len(allObjects),
			"activities_count": len(allActivities),
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"duration":        estimatedDuration,
		"scenes_count":    len(scenes),
		"objects_count":   len(allObjects),
		"processing_time": time.Since(start),
	}).Info("Video analysis completed")

	return response, nil
}
