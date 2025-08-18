package cv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
)

// TesseractOCR implements OCR using Tesseract
type TesseractOCR struct {
	logger       *logrus.Logger
	config       Config
	tesseractCmd string
	language     string
	psm          int // Page Segmentation Mode
}

// NewTesseractOCR creates a new Tesseract OCR engine
func NewTesseractOCR(config Config) (*TesseractOCR, error) {
	ocr := &TesseractOCR{
		config:       config,
		tesseractCmd: "tesseract",
		language:     "eng",
		psm:          6, // Uniform block of text
	}

	// Check if custom Tesseract path is specified
	if config.TesseractPath != "" {
		ocr.tesseractCmd = config.TesseractPath
	}

	// Verify Tesseract installation
	if err := ocr.verifyInstallation(); err != nil {
		return nil, fmt.Errorf("Tesseract verification failed: %w", err)
	}

	return ocr, nil
}

// verifyInstallation checks if Tesseract is properly installed
func (ocr *TesseractOCR) verifyInstallation() error {
	cmd := exec.Command(ocr.tesseractCmd, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tesseract not found or not working: %w", err)
	}

	ocr.logger.Info("Tesseract OCR initialized", "version", strings.Split(string(output), "\n")[0])
	return nil
}

// SetLanguage sets the OCR language
func (ocr *TesseractOCR) SetLanguage(language string) error {
	// Validate language by checking if it's available
	cmd := exec.Command(ocr.tesseractCmd, "--list-langs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get available languages: %w", err)
	}

	availableLanguages := strings.Split(string(output), "\n")
	languageFound := false
	for _, lang := range availableLanguages {
		if strings.TrimSpace(lang) == language {
			languageFound = true
			break
		}
	}

	if !languageFound {
		return fmt.Errorf("language '%s' not available", language)
	}

	ocr.language = language
	ocr.logger.Info("OCR language set", "language", language)
	return nil
}

// SetPageSegmentationMode sets the page segmentation mode
func (ocr *TesseractOCR) SetPageSegmentationMode(mode int) error {
	if mode < 0 || mode > 13 {
		return fmt.Errorf("invalid PSM mode: %d (must be 0-13)", mode)
	}

	ocr.psm = mode
	ocr.logger.Info("OCR PSM set", "mode", mode)
	return nil
}

// ExtractText performs OCR on the given image data
func (ocr *TesseractOCR) ExtractText(ctx context.Context, imageData []byte) (*models.TextRecognition, error) {
	start := time.Now()
	ocr.logger.Info("Starting OCR text extraction", "image_size", len(imageData))

	// Create temporary files
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("ocr_input_%d.png", time.Now().UnixNano()))
	outputBase := filepath.Join(tempDir, fmt.Sprintf("ocr_output_%d", time.Now().UnixNano()))

	// Clean up temporary files
	defer func() {
		os.Remove(inputFile)
		os.Remove(outputBase + ".txt")
		os.Remove(outputBase + ".tsv")
		os.Remove(outputBase + ".hocr")
	}()

	// Write image data to temporary file
	if err := os.WriteFile(inputFile, imageData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input image: %w", err)
	}

	// Extract text with confidence scores
	textResult, err := ocr.extractTextWithConfidence(ctx, inputFile, outputBase)
	if err != nil {
		return nil, fmt.Errorf("OCR extraction failed: %w", err)
	}

	// Extract detailed word-level information
	wordDetails, err := ocr.extractWordDetails(ctx, inputFile, outputBase)
	if err != nil {
		ocr.logger.Warn("Failed to extract word details", "error", err)
		// Create a simple region with the entire text as fallback
		wordDetails = []models.TextRegion{
			{
				Text:       textResult.Text,
				Confidence: textResult.Confidence,
				Language:   ocr.language,
				Bounds:     models.Rectangle{X: 0, Y: 0, Width: 100, Height: 20}, // Placeholder bounds
			},
		}
	}
	regions := wordDetails

	result := &models.TextRecognition{
		Text:       textResult.Text,
		Confidence: textResult.Confidence,
		Language:   ocr.language,
		Regions:    regions,
		Timestamp:  time.Now(),
	}

	ocr.logger.WithFields(logrus.Fields{
		"text_length":     len(result.Text),
		"confidence":      result.Confidence,
		"regions_count":   len(regions),
		"processing_time": time.Since(start),
	}).Info("OCR extraction completed")

	return result, nil
}

// extractTextWithConfidence extracts text with overall confidence
func (ocr *TesseractOCR) extractTextWithConfidence(ctx context.Context, inputFile, outputBase string) (*models.TextRecognition, error) {
	// Run Tesseract to extract text
	args := []string{
		inputFile,
		outputBase,
		"-l", ocr.language,
		"--psm", strconv.Itoa(ocr.psm),
		"--oem", "3", // Use LSTM OCR Engine Mode
	}

	cmd := exec.CommandContext(ctx, ocr.tesseractCmd, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("tesseract command failed: %w, output: %s", err, string(output))
	}

	// Read extracted text
	textFile := outputBase + ".txt"
	textData, err := os.ReadFile(textFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read OCR output: %w", err)
	}

	text := strings.TrimSpace(string(textData))

	// Calculate confidence using TSV output
	confidence, err := ocr.calculateConfidenceFromTSV(ctx, inputFile, outputBase)
	if err != nil {
		ocr.logger.Warn("Failed to calculate confidence", "error", err)
		confidence = 0.5 // Default confidence
	}

	return &models.TextRecognition{
		Text:       text,
		Confidence: confidence,
		Language:   ocr.language,
		Regions:    []models.TextRegion{},
		Timestamp:  time.Now(),
	}, nil
}

// calculateConfidenceFromTSV calculates confidence from TSV output
func (ocr *TesseractOCR) calculateConfidenceFromTSV(ctx context.Context, inputFile, outputBase string) (float64, error) {
	// Run Tesseract to generate TSV output
	args := []string{
		inputFile,
		outputBase,
		"-l", ocr.language,
		"--psm", strconv.Itoa(ocr.psm),
		"--oem", "3",
		"tsv",
	}

	cmd := exec.CommandContext(ctx, ocr.tesseractCmd, args...)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("tesseract TSV command failed: %w", err)
	}

	// Read TSV file
	tsvFile := outputBase + ".tsv"
	tsvData, err := os.ReadFile(tsvFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read TSV output: %w", err)
	}

	// Parse TSV and calculate average confidence
	lines := strings.Split(string(tsvData), "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("invalid TSV format")
	}

	var totalConfidence float64
	var wordCount int

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue
		}

		// Field 10 is confidence, field 11 is text
		confidenceStr := fields[10]
		text := fields[11]

		if strings.TrimSpace(text) == "" || confidenceStr == "-1" {
			continue
		}

		confidence, err := strconv.ParseFloat(confidenceStr, 64)
		if err != nil {
			continue
		}

		totalConfidence += confidence
		wordCount++
	}

	if wordCount == 0 {
		return 0, nil
	}

	return totalConfidence / float64(wordCount) / 100.0, nil // Convert to 0-1 range
}

// extractWordDetails extracts detailed word-level information
func (ocr *TesseractOCR) extractWordDetails(ctx context.Context, inputFile, outputBase string) ([]models.TextRegion, error) {
	// Run Tesseract to generate TSV output with word details
	args := []string{
		inputFile,
		outputBase + "_words",
		"-l", ocr.language,
		"--psm", strconv.Itoa(ocr.psm),
		"--oem", "3",
		"tsv",
	}

	cmd := exec.CommandContext(ctx, ocr.tesseractCmd, args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tesseract word details command failed: %w", err)
	}

	// Read TSV file
	tsvFile := outputBase + "_words.tsv"
	tsvData, err := os.ReadFile(tsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read word details TSV: %w", err)
	}

	// Parse TSV and extract word details
	lines := strings.Split(string(tsvData), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid word details TSV format")
	}

	var words []models.TextRegion

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue
		}

		// Parse fields: level, page_num, block_num, par_num, line_num, word_num, left, top, width, height, conf, text
		text := strings.TrimSpace(fields[11])
		if text == "" {
			continue
		}

		left, _ := strconv.Atoi(fields[6])
		top, _ := strconv.Atoi(fields[7])
		width, _ := strconv.Atoi(fields[8])
		height, _ := strconv.Atoi(fields[9])
		confidence, _ := strconv.ParseFloat(fields[10], 64)

		word := models.TextRegion{
			Text:       text,
			Confidence: confidence / 100.0, // Convert to 0-1 range
			Language:   ocr.language,
			Bounds: models.Rectangle{
				X:      left,
				Y:      top,
				Width:  width,
				Height: height,
			},
		}

		words = append(words, word)
	}

	return words, nil
}

// ExtractTextRegions extracts text from specific regions of an image
func (ocr *TesseractOCR) ExtractTextRegions(ctx context.Context, imageData []byte, regions []models.BoundingBox) ([]models.TextRecognition, error) {
	ocr.logger.Info("Extracting text from regions", "regions_count", len(regions))

	results := make([]models.TextRecognition, 0, len(regions))

	for i, region := range regions {
		// Create cropped image for the region
		croppedImage, err := ocr.cropImage(imageData, region)
		if err != nil {
			ocr.logger.Warn("Failed to crop image for region", "region", i, "error", err)
			continue
		}

		// Extract text from cropped region
		result, err := ocr.ExtractText(ctx, croppedImage)
		if err != nil {
			ocr.logger.Warn("Failed to extract text from region", "region", i, "error", err)
			continue
		}

		// Note: TextRecognition doesn't have metadata field, so we skip adding region info
		results = append(results, *result)
	}

	ocr.logger.Info("Text extraction from regions completed", "successful_regions", len(results))
	return results, nil
}

// cropImage crops an image to the specified bounding box
func (ocr *TesseractOCR) cropImage(imageData []byte, region models.BoundingBox) ([]byte, error) {
	// This is a simplified implementation
	// In a real implementation, you would use an image processing library
	// to actually crop the image data

	// For now, return the original image data
	// TODO: Implement actual image cropping using image processing library
	return imageData, nil
}

// DetectTextRegions detects text regions in an image
func (ocr *TesseractOCR) DetectTextRegions(ctx context.Context, imageData []byte) ([]models.BoundingBox, error) {
	start := time.Now()
	ocr.logger.Info("Detecting text regions", "image_size", len(imageData))

	// Create temporary files
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("text_detect_%d.png", time.Now().UnixNano()))
	outputBase := filepath.Join(tempDir, fmt.Sprintf("text_regions_%d", time.Now().UnixNano()))

	defer func() {
		os.Remove(inputFile)
		os.Remove(outputBase + ".tsv")
	}()

	// Write image data to temporary file
	if err := os.WriteFile(inputFile, imageData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input image: %w", err)
	}

	// Run Tesseract with PSM 6 (uniform block of text) to detect text regions
	args := []string{
		inputFile,
		outputBase,
		"-l", ocr.language,
		"--psm", "6",
		"--oem", "3",
		"tsv",
	}

	cmd := exec.CommandContext(ctx, ocr.tesseractCmd, args...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("tesseract text detection failed: %w", err)
	}

	// Parse TSV to extract text regions
	regions, err := ocr.parseTextRegionsFromTSV(outputBase + ".tsv")
	if err != nil {
		return nil, fmt.Errorf("failed to parse text regions: %w", err)
	}

	ocr.logger.Info("Text region detection completed",
		"regions_found", len(regions),
		"processing_time", time.Since(start),
	)

	return regions, nil
}

// parseTextRegionsFromTSV parses text regions from TSV output
func (ocr *TesseractOCR) parseTextRegionsFromTSV(tsvFile string) ([]models.BoundingBox, error) {
	data, err := os.ReadFile(tsvFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read TSV file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid TSV format")
	}

	var regions []models.BoundingBox
	regionMap := make(map[string]models.BoundingBox)

	for _, line := range lines[1:] { // Skip header
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 12 {
			continue
		}

		level, _ := strconv.Atoi(fields[0])
		text := strings.TrimSpace(fields[11])

		// We're interested in paragraph level (level 4) or line level (level 5)
		if level != 4 && level != 5 {
			continue
		}

		if text == "" {
			continue
		}

		left, _ := strconv.Atoi(fields[6])
		top, _ := strconv.Atoi(fields[7])
		width, _ := strconv.Atoi(fields[8])
		height, _ := strconv.Atoi(fields[9])

		// Create unique key for region
		key := fmt.Sprintf("%d_%d_%d_%d", left, top, width, height)

		region := models.BoundingBox{
			X:      left,
			Y:      top,
			Width:  width,
			Height: height,
		}

		regionMap[key] = region
	}

	// Convert map to slice
	for _, region := range regionMap {
		regions = append(regions, region)
	}

	return regions, nil
}

// GetAvailableLanguages returns list of available OCR languages
func (ocr *TesseractOCR) GetAvailableLanguages() ([]string, error) {
	cmd := exec.Command(ocr.tesseractCmd, "--list-langs")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get available languages: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	languages := make([]string, 0)

	for _, line := range lines[1:] { // Skip first line (header)
		lang := strings.TrimSpace(line)
		if lang != "" {
			languages = append(languages, lang)
		}
	}

	return languages, nil
}
