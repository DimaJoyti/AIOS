package cv

import (
	"context"
	"math/rand"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
)

// Config represents configuration for CV services
type Config struct {
	ModelPath            string  `yaml:"model_path"`
	YOLOModelPath        string  `yaml:"yolo_model_path"`
	ModelsPath           string  `yaml:"models_path"`
	TesseractPath        string  `yaml:"tesseract_path"`
	ConfidenceThreshold  float64 `yaml:"confidence_threshold"`
	Language             string  `yaml:"language"`
	PageSegmentationMode int     `yaml:"page_segmentation_mode"`
	MaxProcessingTime    int     `yaml:"max_processing_time_ms"`
	EnableGPU            bool    `yaml:"enable_gpu"`
}

// MockDetector provides mock object detection for development
type MockDetector struct {
	logger *logrus.Logger
}

// NewMockDetector creates a new mock object detector
func NewMockDetector(logger *logrus.Logger) *MockDetector {
	return &MockDetector{
		logger: logger,
	}
}

// LoadModel mock implementation
func (m *MockDetector) LoadModel(modelPath string) error {
	m.logger.Info("Mock detector: model loaded", "path", modelPath)
	return nil
}

// SetConfidenceThreshold mock implementation
func (m *MockDetector) SetConfidenceThreshold(threshold float64) {
	m.logger.Info("Mock detector: confidence threshold set", "threshold", threshold)
}

// DetectObjects mock implementation
func (m *MockDetector) DetectObjects(ctx context.Context, imageData []byte) (*models.ObjectDetection, error) {
	start := time.Now()
	m.logger.Info("Mock detector: detecting objects", "image_size", len(imageData))

	// Simulate processing time
	time.Sleep(time.Millisecond * time.Duration(50+rand.Intn(100)))

	// Generate mock detections
	objects := []models.DetectedObject{
		{
			Class:      "person",
			Confidence: 0.92 + rand.Float64()*0.07,
			Bounds: models.Rectangle{
				X:      50 + rand.Intn(100),
				Y:      100 + rand.Intn(50),
				Width:  80 + rand.Intn(40),
				Height: 180 + rand.Intn(60),
			},
		},
		{
			Class:      "laptop",
			Confidence: 0.85 + rand.Float64()*0.1,
			Bounds: models.Rectangle{
				X:      200 + rand.Intn(100),
				Y:      150 + rand.Intn(50),
				Width:  120 + rand.Intn(30),
				Height: 80 + rand.Intn(20),
			},
		},
		{
			Class:      "chair",
			Confidence: 0.78 + rand.Float64()*0.15,
			Bounds: models.Rectangle{
				X:      400 + rand.Intn(100),
				Y:      200 + rand.Intn(100),
				Width:  60 + rand.Intn(20),
				Height: 100 + rand.Intn(40),
			},
		},
	}

	// Randomly include/exclude some objects
	finalObjects := make([]models.DetectedObject, 0)
	for _, obj := range objects {
		if rand.Float64() > 0.3 { // 70% chance to include
			finalObjects = append(finalObjects, obj)
		}
	}

	result := &models.ObjectDetection{
		Objects:   finalObjects,
		Count:     len(finalObjects),
		Timestamp: time.Now(),
	}

	m.logger.WithFields(logrus.Fields{
		"objects_detected": len(finalObjects),
		"processing_time":  time.Since(start),
	}).Info("Mock detector: detection completed")

	return result, nil
}

// MockOCR provides mock OCR for development
type MockOCR struct {
	logger   *logrus.Logger
	language string
	psm      int
}

// NewMockOCR creates a new mock OCR engine
func NewMockOCR(logger *logrus.Logger) *MockOCR {
	return &MockOCR{
		logger:   logger,
		language: "eng",
		psm:      6,
	}
}

// SetLanguage mock implementation
func (m *MockOCR) SetLanguage(language string) error {
	m.language = language
	m.logger.Info("Mock OCR: language set", "language", language)
	return nil
}

// SetPageSegmentationMode mock implementation
func (m *MockOCR) SetPageSegmentationMode(mode int) error {
	m.psm = mode
	m.logger.Info("Mock OCR: PSM set", "mode", mode)
	return nil
}

// ExtractText mock implementation
func (m *MockOCR) ExtractText(ctx context.Context, imageData []byte) (*models.TextRecognition, error) {
	start := time.Now()
	m.logger.Info("Mock OCR: extracting text", "image_size", len(imageData))

	// Simulate processing time
	time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(200)))

	// Generate mock text based on image size and context
	mockTexts := []string{
		"Welcome to AIOS Desktop Environment",
		"File Edit View Tools Help",
		"New Document",
		"Save As...",
		"Settings",
		"About AIOS",
		"Terminal",
		"Applications",
		"System Monitor",
		"Task Manager",
		"Network Settings",
		"Display Configuration",
		"User Account",
		"Security Settings",
		"Backup & Restore",
		"Software Updates",
	}

	// Select random text elements
	numTexts := 1 + rand.Intn(3)
	selectedTexts := make([]string, 0, numTexts)
	for i := 0; i < numTexts; i++ {
		selectedTexts = append(selectedTexts, mockTexts[rand.Intn(len(mockTexts))])
	}

	fullText := ""
	regions := make([]models.TextRegion, 0)
	currentX := 50
	currentY := 50

	for _, text := range selectedTexts {
		if fullText != "" {
			fullText += " "
		}
		fullText += text

		// Generate word-level details
		textWords := splitIntoWords(text)
		for _, word := range textWords {
			wordWidth := len(word) * 8 // Approximate character width
			regions = append(regions, models.TextRegion{
				Text:       word,
				Confidence: 0.85 + rand.Float64()*0.1,
				Language:   m.language,
				Bounds: models.Rectangle{
					X:      currentX,
					Y:      currentY,
					Width:  wordWidth,
					Height: 16,
				},
			})
			currentX += wordWidth + 5 // Add space between words
		}
		currentY += 25 // Move to next line
		currentX = 50  // Reset X position
	}

	confidence := 0.80 + rand.Float64()*0.15

	result := &models.TextRecognition{
		Text:       fullText,
		Confidence: confidence,
		Language:   m.language,
		Regions:    regions,
		Timestamp:  time.Now(),
	}

	m.logger.WithFields(logrus.Fields{
		"text_length":     len(result.Text),
		"confidence":      result.Confidence,
		"regions_count":   len(regions),
		"processing_time": time.Since(start),
	}).Info("Mock OCR: extraction completed")

	return result, nil
}

// splitIntoWords splits text into individual words
func splitIntoWords(text string) []string {
	words := make([]string, 0)
	currentWord := ""

	for _, char := range text {
		if char == ' ' || char == '\t' || char == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(char)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// ImageAnalyzer provides general image analysis capabilities
type ImageAnalyzer struct {
	logger *logrus.Logger
	config Config
}

// NewImageAnalyzer creates a new image analyzer
func NewImageAnalyzer(logger *logrus.Logger, config Config) *ImageAnalyzer {
	return &ImageAnalyzer{
		logger: logger,
		config: config,
	}
}

// AnalyzeImage performs general image analysis
func (a *ImageAnalyzer) AnalyzeImage(ctx context.Context, imageData []byte) (*models.ImageDescription, error) {
	start := time.Now()
	a.logger.WithField("image_size", len(imageData)).Info("Analyzing image")

	// Simulate processing time
	time.Sleep(time.Millisecond * time.Duration(200+rand.Intn(300)))

	// Generate mock analysis based on image characteristics
	tags := []string{"desktop", "interface", "application", "window", "ui"}

	// Add random additional tags
	additionalTags := []string{"button", "menu", "text", "icon", "toolbar", "statusbar", "dialog"}
	for _, tag := range additionalTags {
		if rand.Float64() > 0.5 {
			tags = append(tags, tag)
		}
	}

	// Generate description based on detected elements
	descriptions := []string{
		"A desktop computer interface with multiple application windows",
		"Software application with various UI elements and controls",
		"Modern graphical user interface with buttons and menus",
		"Desktop environment showing system applications and tools",
		"Computer screen displaying productivity software interface",
	}

	description := descriptions[rand.Intn(len(descriptions))]
	confidence := 0.75 + rand.Float64()*0.2

	result := &models.ImageDescription{
		Description: description,
		Details:     []string{"UI elements detected", "Software interface visible"},
		Objects:     tags[:min(3, len(tags))], // Use first 3 tags as objects
		Scene:       "desktop_environment",
		Colors:      []string{"blue", "gray", "white"},
		Metadata: map[string]interface{}{
			"image_size":      len(imageData),
			"tags_count":      len(tags),
			"analysis_type":   "general",
			"mock":            true,
			"confidence":      confidence,
			"processing_time": time.Since(start),
		},
		Timestamp: time.Now(),
	}

	a.logger.WithFields(logrus.Fields{
		"description_length": len(description),
		"tags_count":         len(tags),
		"confidence":         confidence,
		"processing_time":    time.Since(start),
	}).Info("Image analysis completed")

	return result, nil
}

// ClassifyImage classifies the content of an image
func (a *ImageAnalyzer) ClassifyImage(ctx context.Context, imageData []byte) (*models.ImageClassification, error) {
	start := time.Now()
	a.logger.WithField("image_size", len(imageData)).Info("Classifying image")

	// Simulate processing time
	time.Sleep(time.Millisecond * time.Duration(150+rand.Intn(200)))

	// Generate mock classification results
	classes := []models.ClassificationResult{
		{
			Class:       "desktop_interface",
			Confidence:  0.85 + rand.Float64()*0.1,
			Probability: 0.85 + rand.Float64()*0.1,
		},
		{
			Class:       "application_window",
			Confidence:  0.75 + rand.Float64()*0.15,
			Probability: 0.75 + rand.Float64()*0.15,
		},
		{
			Class:       "productivity_software",
			Confidence:  0.65 + rand.Float64()*0.2,
			Probability: 0.65 + rand.Float64()*0.2,
		},
		{
			Class:       "system_interface",
			Confidence:  0.55 + rand.Float64()*0.25,
			Probability: 0.55 + rand.Float64()*0.25,
		},
	}

	// Sort by confidence
	for i := 0; i < len(classes)-1; i++ {
		for j := i + 1; j < len(classes); j++ {
			if classes[j].Confidence > classes[i].Confidence {
				classes[i], classes[j] = classes[j], classes[i]
			}
		}
	}

	topClass := classes[0]

	result := &models.ImageClassification{
		Classes:    classes,
		TopClass:   topClass.Class,
		Confidence: topClass.Confidence,
		Timestamp:  time.Now(),
	}

	a.logger.WithFields(logrus.Fields{
		"top_class":       topClass.Class,
		"top_confidence":  topClass.Confidence,
		"classes_count":   len(classes),
		"processing_time": time.Since(start),
	}).Info("Image classification completed")

	return result, nil
}

// GenerateMockScreenshot generates mock screenshot data for testing
func GenerateMockScreenshot(width, height int) []byte {
	// Generate mock PNG-like data
	size := width * height * 3 // RGB
	data := make([]byte, size)

	// Fill with gradient pattern
	for i := 0; i < size; i += 3 {
		x := (i / 3) % width
		y := (i / 3) / width

		// Create gradient effect
		r := byte((x * 255) / width)
		g := byte((y * 255) / height)
		b := byte(((x + y) * 255) / (width + height))

		data[i] = r
		data[i+1] = g
		data[i+2] = b
	}

	return data
}

// GenerateMockUIElements generates mock UI elements for testing
func GenerateMockUIElements(imageWidth, imageHeight int) []models.UIElement {
	elements := []models.UIElement{
		{
			ID:         "window-main",
			Type:       "window",
			Text:       "AIOS Desktop",
			Bounds:     models.Rectangle{X: 0, Y: 0, Width: imageWidth, Height: imageHeight},
			Confidence: 0.95,
			Properties: map[string]interface{}{
				"title":     "AIOS Desktop Environment",
				"resizable": true,
				"maximized": true,
			},
		},
		{
			ID:         "button-menu",
			Type:       "button",
			Text:       "Menu",
			Bounds:     models.Rectangle{X: 10, Y: 10, Width: 60, Height: 30},
			Confidence: 0.88,
			Properties: map[string]interface{}{
				"clickable": true,
				"enabled":   true,
			},
		},
		{
			ID:         "textfield-search",
			Type:       "textfield",
			Text:       "Search...",
			Bounds:     models.Rectangle{X: 100, Y: 10, Width: 200, Height: 30},
			Confidence: 0.92,
			Properties: map[string]interface{}{
				"placeholder": "Search applications and files",
				"editable":    true,
			},
		},
		{
			ID:         "button-settings",
			Type:       "button",
			Text:       "Settings",
			Bounds:     models.Rectangle{X: imageWidth - 80, Y: 10, Width: 70, Height: 30},
			Confidence: 0.85,
			Properties: map[string]interface{}{
				"clickable": true,
				"icon":      "gear",
			},
		},
	}

	return elements
}
