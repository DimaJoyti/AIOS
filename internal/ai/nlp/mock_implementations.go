package nlp

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
)

// MockIntentClassifier provides mock intent classification for development
type MockIntentClassifier struct {
	logger *logrus.Logger
}

// NewMockIntentClassifier creates a new mock intent classifier
func NewMockIntentClassifier(logger *logrus.Logger) *MockIntentClassifier {
	return &MockIntentClassifier{
		logger: logger,
	}
}

// LoadModel mock implementation
func (m *MockIntentClassifier) LoadModel(modelPath string) error {
	m.logger.Info("Mock intent classifier: model loaded", "path", modelPath)
	return nil
}

// ClassifyIntent mock implementation
func (m *MockIntentClassifier) ClassifyIntent(ctx context.Context, text string) (*models.IntentAnalysis, error) {
	m.logger.Info("Mock intent classifier: classifying intent", "text", text)

	// Simple keyword-based intent classification
	text = strings.ToLower(text)
	
	var intent string
	var confidence float64
	var entities []models.NamedEntity

	switch {
	case strings.Contains(text, "open") || strings.Contains(text, "launch") || strings.Contains(text, "start"):
		intent = "open_application"
		confidence = 0.85 + rand.Float64()*0.1
		entities = append(entities, models.NamedEntity{
			Text:       extractTarget(text, []string{"open", "launch", "start"}),
			Type:       "APPLICATION",
			Confidence: 0.8,
		})
	case strings.Contains(text, "close") || strings.Contains(text, "quit") || strings.Contains(text, "exit"):
		intent = "close_application"
		confidence = 0.80 + rand.Float64()*0.15
		entities = append(entities, models.NamedEntity{
			Text:       extractTarget(text, []string{"close", "quit", "exit"}),
			Type:       "APPLICATION",
			Confidence: 0.75,
		})
	case strings.Contains(text, "search") || strings.Contains(text, "find") || strings.Contains(text, "look"):
		intent = "search"
		confidence = 0.88 + rand.Float64()*0.1
		entities = append(entities, models.NamedEntity{
			Text:       extractTarget(text, []string{"search", "find", "look"}),
			Type:       "QUERY",
			Confidence: 0.85,
		})
	case strings.Contains(text, "play") || strings.Contains(text, "music") || strings.Contains(text, "song"):
		intent = "play_media"
		confidence = 0.90 + rand.Float64()*0.08
		entities = append(entities, models.NamedEntity{
			Text:       extractTarget(text, []string{"play", "music", "song"}),
			Type:       "MEDIA",
			Confidence: 0.88,
		})
	case strings.Contains(text, "volume") || strings.Contains(text, "sound"):
		intent = "adjust_volume"
		confidence = 0.92 + rand.Float64()*0.06
		entities = append(entities, models.NamedEntity{
			Text:       extractNumber(text),
			Type:       "NUMBER",
			Confidence: 0.90,
		})
	case strings.Contains(text, "help") || strings.Contains(text, "assist") || strings.Contains(text, "support"):
		intent = "request_help"
		confidence = 0.95 + rand.Float64()*0.04
	case strings.Contains(text, "weather") || strings.Contains(text, "temperature"):
		intent = "get_weather"
		confidence = 0.87 + rand.Float64()*0.1
	case strings.Contains(text, "time") || strings.Contains(text, "clock"):
		intent = "get_time"
		confidence = 0.93 + rand.Float64()*0.05
	default:
		intent = "unknown"
		confidence = 0.3 + rand.Float64()*0.4
	}

	return &models.IntentAnalysis{
		Intent:     intent,
		Confidence: confidence,
		Entities:   entities,
		Context: map[string]interface{}{
			"mock":        true,
			"text_length": len(text),
		},
		Timestamp: time.Now(),
	}, nil
}

// MockEntityExtractor provides mock entity extraction for development
type MockEntityExtractor struct {
	logger *logrus.Logger
}

// NewMockEntityExtractor creates a new mock entity extractor
func NewMockEntityExtractor(logger *logrus.Logger) *MockEntityExtractor {
	return &MockEntityExtractor{
		logger: logger,
	}
}

// LoadModel mock implementation
func (m *MockEntityExtractor) LoadModel(modelPath string) error {
	m.logger.Info("Mock entity extractor: model loaded", "path", modelPath)
	return nil
}

// ExtractEntities mock implementation
func (m *MockEntityExtractor) ExtractEntities(ctx context.Context, text string) (*models.EntityExtraction, error) {
	m.logger.Info("Mock entity extractor: extracting entities", "text", text)

	entities := []models.NamedEntity{}
	
	// Simple pattern-based entity extraction
	words := strings.Fields(strings.ToLower(text))
	
	for i, word := range words {
		switch {
		case isApplication(word):
			entities = append(entities, models.NamedEntity{
				Text:       word,
				Type:       "APPLICATION",
				Confidence: 0.85 + rand.Float64()*0.1,
				StartPos:   i,
				EndPos:     i + 1,
			})
		case isNumber(word):
			entities = append(entities, models.NamedEntity{
				Text:       word,
				Type:       "NUMBER",
				Confidence: 0.95 + rand.Float64()*0.04,
				StartPos:   i,
				EndPos:     i + 1,
			})
		case isTime(word):
			entities = append(entities, models.NamedEntity{
				Text:       word,
				Type:       "TIME",
				Confidence: 0.88 + rand.Float64()*0.1,
				StartPos:   i,
				EndPos:     i + 1,
			})
		case isLocation(word):
			entities = append(entities, models.NamedEntity{
				Text:       word,
				Type:       "LOCATION",
				Confidence: 0.80 + rand.Float64()*0.15,
				StartPos:   i,
				EndPos:     i + 1,
			})
		}
	}

	return &models.EntityExtraction{
		Entities:  entities,
		Relations: []models.EntityRelation{}, // TODO: Add relation extraction
		Timestamp: time.Now(),
	}, nil
}

// MockSentimentAnalyzer provides mock sentiment analysis for development
type MockSentimentAnalyzer struct {
	logger *logrus.Logger
}

// NewMockSentimentAnalyzer creates a new mock sentiment analyzer
func NewMockSentimentAnalyzer(logger *logrus.Logger) *MockSentimentAnalyzer {
	return &MockSentimentAnalyzer{
		logger: logger,
	}
}

// LoadModel mock implementation
func (m *MockSentimentAnalyzer) LoadModel(modelPath string) error {
	m.logger.Info("Mock sentiment analyzer: model loaded", "path", modelPath)
	return nil
}

// AnalyzeSentiment mock implementation
func (m *MockSentimentAnalyzer) AnalyzeSentiment(ctx context.Context, text string) (*models.SentimentAnalysis, error) {
	m.logger.Info("Mock sentiment analyzer: analyzing sentiment", "text", text)

	// Simple keyword-based sentiment analysis
	text = strings.ToLower(text)
	
	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "like", "happy", "pleased"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "hate", "dislike", "angry", "frustrated", "disappointed", "sad"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if strings.Contains(text, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(text, word) {
			negativeCount++
		}
	}
	
	var sentiment models.SentimentScore
	if positiveCount > negativeCount {
		sentiment = models.SentimentScore{
			Label:      "positive",
			Score:      0.7 + rand.Float64()*0.25,
			Confidence: 0.8 + rand.Float64()*0.15,
		}
	} else if negativeCount > positiveCount {
		sentiment = models.SentimentScore{
			Label:      "negative",
			Score:      -0.7 - rand.Float64()*0.25,
			Confidence: 0.8 + rand.Float64()*0.15,
		}
	} else {
		sentiment = models.SentimentScore{
			Label:      "neutral",
			Score:      -0.1 + rand.Float64()*0.2,
			Confidence: 0.6 + rand.Float64()*0.3,
		}
	}

	emotions := []models.EmotionScore{
		{Emotion: "joy", Score: rand.Float64(), Confidence: 0.7 + rand.Float64()*0.2},
		{Emotion: "anger", Score: rand.Float64(), Confidence: 0.7 + rand.Float64()*0.2},
		{Emotion: "sadness", Score: rand.Float64(), Confidence: 0.7 + rand.Float64()*0.2},
		{Emotion: "fear", Score: rand.Float64(), Confidence: 0.7 + rand.Float64()*0.2},
	}

	return &models.SentimentAnalysis{
		Sentiment: sentiment,
		Emotions:  emotions,
		Aspects:   []models.AspectSentiment{}, // TODO: Add aspect-based sentiment
		Timestamp: time.Now(),
	}, nil
}

// MockLanguageDetector provides mock language detection for development
type MockLanguageDetector struct {
	logger *logrus.Logger
}

// NewMockLanguageDetector creates a new mock language detector
func NewMockLanguageDetector(logger *logrus.Logger) *MockLanguageDetector {
	return &MockLanguageDetector{
		logger: logger,
	}
}

// LoadModel mock implementation
func (m *MockLanguageDetector) LoadModel(modelPath string) error {
	m.logger.Info("Mock language detector: model loaded", "path", modelPath)
	return nil
}

// DetectLanguage mock implementation
func (m *MockLanguageDetector) DetectLanguage(ctx context.Context, text string) (string, float64, error) {
	m.logger.Info("Mock language detector: detecting language", "text", text)

	// Simple heuristic-based language detection
	text = strings.ToLower(text)
	
	// Check for common words in different languages
	englishWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by"}
	spanishWords := []string{"el", "la", "y", "o", "pero", "en", "de", "con", "por", "para", "que", "es"}
	frenchWords := []string{"le", "la", "et", "ou", "mais", "dans", "de", "avec", "par", "pour", "que", "est"}
	
	englishCount := countWords(text, englishWords)
	spanishCount := countWords(text, spanishWords)
	frenchCount := countWords(text, frenchWords)
	
	if englishCount >= spanishCount && englishCount >= frenchCount {
		return "en", 0.85 + rand.Float64()*0.1, nil
	} else if spanishCount >= frenchCount {
		return "es", 0.80 + rand.Float64()*0.15, nil
	} else {
		return "fr", 0.75 + rand.Float64()*0.2, nil
	}
}

// Helper functions

func extractTarget(text string, triggers []string) string {
	words := strings.Fields(text)
	for i, word := range words {
		for _, trigger := range triggers {
			if strings.Contains(word, trigger) && i+1 < len(words) {
				return words[i+1]
			}
		}
	}
	return "unknown"
}

func extractNumber(text string) string {
	words := strings.Fields(text)
	for _, word := range words {
		if isNumber(word) {
			return word
		}
	}
	return "0"
}

func isApplication(word string) bool {
	apps := []string{"terminal", "browser", "editor", "calculator", "music", "video", "email", "chat"}
	for _, app := range apps {
		if strings.Contains(word, app) {
			return true
		}
	}
	return false
}

func isNumber(word string) bool {
	numbers := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "ten", "twenty", "thirty", "fifty", "hundred"}
	for _, num := range numbers {
		if strings.Contains(word, num) {
			return true
		}
	}
	return false
}

func isTime(word string) bool {
	timeWords := []string{"morning", "afternoon", "evening", "night", "today", "tomorrow", "yesterday", "hour", "minute", "second"}
	for _, timeWord := range timeWords {
		if strings.Contains(word, timeWord) {
			return true
		}
	}
	return false
}

func isLocation(word string) bool {
	locations := []string{"home", "office", "school", "park", "store", "restaurant", "city", "country"}
	for _, location := range locations {
		if strings.Contains(word, location) {
			return true
		}
	}
	return false
}

func countWords(text string, words []string) int {
	count := 0
	for _, word := range words {
		if strings.Contains(text, word) {
			count++
		}
	}
	return count
}
