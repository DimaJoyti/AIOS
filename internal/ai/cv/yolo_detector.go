package cv

import (
	"context"
	"encoding/json"
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

// YOLODetector implements object detection using YOLO models
type YOLODetector struct {
	logger      *logrus.Logger
	config      Config
	modelPath   string
	initialized bool
}

// NewYOLODetector creates a new YOLO object detector
func NewYOLODetector(config Config) (*YOLODetector, error) {
	detector := &YOLODetector{
		config:    config,
		modelPath: config.YOLOModelPath,
	}

	// Check if YOLO model exists
	if detector.modelPath == "" {
		detector.modelPath = filepath.Join(config.ModelsPath, "yolo", "yolov8n.pt")
	}

	// Initialize the detector
	if err := detector.LoadModel(detector.modelPath); err != nil {
		return nil, fmt.Errorf("failed to load YOLO model: %w", err)
	}

	return detector, nil
}

// LoadModel loads a YOLO model from the specified path
func (d *YOLODetector) LoadModel(modelPath string) error {
	d.logger.Info("Loading YOLO model", "path", modelPath)

	// Check if model file exists
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		// Try to download the model if it doesn't exist
		if err := d.downloadModel(modelPath); err != nil {
			return fmt.Errorf("model not found and download failed: %w", err)
		}
	}

	d.modelPath = modelPath
	d.initialized = true
	d.logger.Info("YOLO model loaded successfully", "path", modelPath)
	return nil
}

// SetConfidenceThreshold sets the confidence threshold for detections
func (d *YOLODetector) SetConfidenceThreshold(threshold float64) {
	d.config.ConfidenceThreshold = threshold
	d.logger.Info("Updated confidence threshold", "threshold", threshold)
}

// DetectObjects performs object detection on the given image data
func (d *YOLODetector) DetectObjects(ctx context.Context, imageData []byte) (*models.ObjectDetection, error) {
	if !d.initialized {
		return nil, fmt.Errorf("YOLO detector not initialized")
	}

	start := time.Now()
	d.logger.Info("Starting YOLO object detection", "image_size", len(imageData))

	// Create temporary file for input image
	tempDir := os.TempDir()
	inputFile := filepath.Join(tempDir, fmt.Sprintf("yolo_input_%d.jpg", time.Now().UnixNano()))
	outputFile := filepath.Join(tempDir, fmt.Sprintf("yolo_output_%d.txt", time.Now().UnixNano()))

	// Clean up temporary files
	defer func() {
		os.Remove(inputFile)
		os.Remove(outputFile)
	}()

	// Write image data to temporary file
	if err := os.WriteFile(inputFile, imageData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write input image: %w", err)
	}

	// Run YOLO detection
	detections, err := d.runYOLOInference(ctx, inputFile, outputFile)
	if err != nil {
		return nil, fmt.Errorf("YOLO inference failed: %w", err)
	}

	// Filter detections by confidence threshold
	filteredDetections := make([]models.DetectedObject, 0)
	for _, detection := range detections {
		if detection.Confidence >= d.config.ConfidenceThreshold {
			filteredDetections = append(filteredDetections, detection)
		}
	}

	result := &models.ObjectDetection{
		Objects:   filteredDetections,
		Count:     len(filteredDetections),
		Timestamp: time.Now(),
	}

	d.logger.WithFields(logrus.Fields{
		"objects_detected": len(filteredDetections),
		"processing_time":  time.Since(start),
	}).Info("YOLO detection completed")

	return result, nil
}

// runYOLOInference executes YOLO inference using Python script or CLI
func (d *YOLODetector) runYOLOInference(ctx context.Context, inputFile, outputFile string) ([]models.DetectedObject, error) {
	// Try to use ultralytics CLI if available
	if d.hasUltralyticsInstalled() {
		return d.runUltralyticsInference(ctx, inputFile, outputFile)
	}

	// Fall back to Python script
	return d.runPythonInference(ctx, inputFile, outputFile)
}

// hasUltralyticsInstalled checks if ultralytics is installed
func (d *YOLODetector) hasUltralyticsInstalled() bool {
	cmd := exec.Command("yolo", "version")
	return cmd.Run() == nil
}

// runUltralyticsInference runs inference using ultralytics CLI
func (d *YOLODetector) runUltralyticsInference(ctx context.Context, inputFile, outputFile string) ([]models.DetectedObject, error) {
	// Build YOLO command
	args := []string{
		"predict",
		fmt.Sprintf("model=%s", d.modelPath),
		fmt.Sprintf("source=%s", inputFile),
		fmt.Sprintf("conf=%f", d.config.ConfidenceThreshold),
		"save_txt=true",
		"save_conf=true",
		fmt.Sprintf("project=%s", filepath.Dir(outputFile)),
		fmt.Sprintf("name=%s", strings.TrimSuffix(filepath.Base(outputFile), ".txt")),
	}

	cmd := exec.CommandContext(ctx, "yolo", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.logger.Error("YOLO command failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("YOLO command failed: %w", err)
	}

	// Parse results from output directory
	return d.parseYOLOResults(outputFile)
}

// runPythonInference runs inference using Python script
func (d *YOLODetector) runPythonInference(ctx context.Context, inputFile, outputFile string) ([]models.DetectedObject, error) {
	// Create Python script for YOLO inference
	scriptContent := d.generatePythonScript(inputFile, outputFile)
	scriptFile := filepath.Join(os.TempDir(), fmt.Sprintf("yolo_script_%d.py", time.Now().UnixNano()))

	defer os.Remove(scriptFile)

	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to create Python script: %w", err)
	}

	// Run Python script
	cmd := exec.CommandContext(ctx, "python3", scriptFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.logger.Error("Python YOLO script failed", "error", err, "output", string(output))
		return nil, fmt.Errorf("Python YOLO script failed: %w", err)
	}

	// Parse results
	return d.parseYOLOResults(outputFile)
}

// generatePythonScript creates a Python script for YOLO inference
func (d *YOLODetector) generatePythonScript(inputFile, outputFile string) string {
	return fmt.Sprintf(`
import sys
try:
    from ultralytics import YOLO
    import cv2
    import json
except ImportError as e:
    print(f"Required package not installed: {e}")
    sys.exit(1)

# Load model
model = YOLO('%s')

# Run inference
results = model('%s', conf=%f)

# Parse results
detections = []
for result in results:
    boxes = result.boxes
    if boxes is not None:
        for box in boxes:
            x1, y1, x2, y2 = box.xyxy[0].tolist()
            conf = box.conf[0].item()
            cls = int(box.cls[0].item())
            class_name = model.names[cls]
            
            detection = {
                'class': class_name,
                'confidence': conf,
                'x': int(x1),
                'y': int(y1),
                'width': int(x2 - x1),
                'height': int(y2 - y1)
            }
            detections.append(detection)

# Save results
with open('%s', 'w') as f:
    json.dump(detections, f)

print(f"Detected {len(detections)} objects")
`, d.modelPath, inputFile, d.config.ConfidenceThreshold, outputFile)
}

// parseYOLOResults parses YOLO detection results from output file
func (d *YOLODetector) parseYOLOResults(outputFile string) ([]models.DetectedObject, error) {
	// Check if output file exists
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		return []models.DetectedObject{}, nil
	}

	// Read output file
	data, err := os.ReadFile(outputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read output file: %w", err)
	}

	// Parse JSON results
	var rawDetections []map[string]interface{}
	if err := json.Unmarshal(data, &rawDetections); err != nil {
		// Try parsing as text format (YOLO txt format)
		return d.parseTextResults(string(data))
	}

	// Convert to DetectedObject format
	detections := make([]models.DetectedObject, 0, len(rawDetections))
	for _, raw := range rawDetections {
		detection := models.DetectedObject{
			Class:      raw["class"].(string),
			Confidence: raw["confidence"].(float64),
			Bounds: models.Rectangle{
				X:      int(raw["x"].(float64)),
				Y:      int(raw["y"].(float64)),
				Width:  int(raw["width"].(float64)),
				Height: int(raw["height"].(float64)),
			},
		}
		detections = append(detections, detection)
	}

	return detections, nil
}

// parseTextResults parses YOLO text format results
func (d *YOLODetector) parseTextResults(content string) ([]models.DetectedObject, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	detections := make([]models.DetectedObject, 0, len(lines))

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 6 {
			continue
		}

		// Parse YOLO format: class_id x_center y_center width height confidence
		classID, _ := strconv.Atoi(parts[0])
		xCenter, _ := strconv.ParseFloat(parts[1], 64)
		yCenter, _ := strconv.ParseFloat(parts[2], 64)
		width, _ := strconv.ParseFloat(parts[3], 64)
		height, _ := strconv.ParseFloat(parts[4], 64)
		confidence, _ := strconv.ParseFloat(parts[5], 64)

		// Convert to absolute coordinates (assuming image dimensions)
		// Note: This is simplified - real implementation would need actual image dimensions
		imgWidth, imgHeight := 640, 640 // Default YOLO input size

		x := int((xCenter - width/2) * float64(imgWidth))
		y := int((yCenter - height/2) * float64(imgHeight))
		w := int(width * float64(imgWidth))
		h := int(height * float64(imgHeight))

		detection := models.DetectedObject{
			Class:      d.getClassName(classID),
			Confidence: confidence,
			Bounds: models.Rectangle{
				X:      x,
				Y:      y,
				Width:  w,
				Height: h,
			},
		}
		detections = append(detections, detection)
	}

	return detections, nil
}

// getClassName returns the class name for a given class ID
func (d *YOLODetector) getClassName(classID int) string {
	// COCO class names (simplified)
	cocoClasses := []string{
		"person", "bicycle", "car", "motorcycle", "airplane", "bus", "train", "truck",
		"boat", "traffic light", "fire hydrant", "stop sign", "parking meter", "bench",
		"bird", "cat", "dog", "horse", "sheep", "cow", "elephant", "bear", "zebra",
		"giraffe", "backpack", "umbrella", "handbag", "tie", "suitcase", "frisbee",
		"skis", "snowboard", "sports ball", "kite", "baseball bat", "baseball glove",
		"skateboard", "surfboard", "tennis racket", "bottle", "wine glass", "cup",
		"fork", "knife", "spoon", "bowl", "banana", "apple", "sandwich", "orange",
		"broccoli", "carrot", "hot dog", "pizza", "donut", "cake", "chair", "couch",
		"potted plant", "bed", "dining table", "toilet", "tv", "laptop", "mouse",
		"remote", "keyboard", "cell phone", "microwave", "oven", "toaster", "sink",
		"refrigerator", "book", "clock", "vase", "scissors", "teddy bear", "hair drier",
		"toothbrush",
	}

	if classID >= 0 && classID < len(cocoClasses) {
		return cocoClasses[classID]
	}
	return fmt.Sprintf("class_%d", classID)
}

// calculateAverageConfidence calculates the average confidence of detections
func (d *YOLODetector) calculateAverageConfidence(detections []models.DetectedObject) float64 {
	if len(detections) == 0 {
		return 0.0
	}

	total := 0.0
	for _, detection := range detections {
		total += detection.Confidence
	}
	return total / float64(len(detections))
}

// downloadModel downloads a YOLO model if it doesn't exist
func (d *YOLODetector) downloadModel(modelPath string) error {
	d.logger.Info("Downloading YOLO model", "path", modelPath)

	// Create model directory
	if err := os.MkdirAll(filepath.Dir(modelPath), 0755); err != nil {
		return fmt.Errorf("failed to create model directory: %w", err)
	}

	// Download using Python script
	scriptContent := fmt.Sprintf(`
from ultralytics import YOLO
import os

# Download and save model
model = YOLO('%s')
model.save('%s')
print(f"Model saved to {os.path.abspath('%s')}")
`, filepath.Base(modelPath), modelPath, modelPath)

	scriptFile := filepath.Join(os.TempDir(), "download_yolo.py")
	defer os.Remove(scriptFile)

	if err := os.WriteFile(scriptFile, []byte(scriptContent), 0644); err != nil {
		return fmt.Errorf("failed to create download script: %w", err)
	}

	cmd := exec.Command("python3", scriptFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		d.logger.Error("Model download failed", "error", err, "output", string(output))
		return fmt.Errorf("model download failed: %w", err)
	}

	d.logger.Info("YOLO model downloaded successfully", "path", modelPath)
	return nil
}
