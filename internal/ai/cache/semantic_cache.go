package cache

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SemanticCacheEntry represents a semantically cached item
type SemanticCacheEntry struct {
	ID          string
	Query       string
	QueryVector []float64
	Response    interface{}
	Metadata    map[string]interface{}
	ExpiresAt   time.Time
	CreatedAt   time.Time
	AccessCount int64
	LastAccess  time.Time
	Confidence  float64
}

// SemanticCacheManager implements semantic caching for AI responses
type SemanticCacheManager struct {
	entries             map[string]*SemanticCacheEntry
	queryIndex          map[string][]string // Simple text-based index
	similarityThreshold float64
	maxEntries          int
	ttl                 time.Duration
	logger              *logrus.Logger
	tracer              trace.Tracer
	stats               *models.CacheStats
}

// NewSemanticCacheManager creates a new semantic cache manager
func NewSemanticCacheManager(maxEntries int, ttl time.Duration, similarityThreshold float64, logger *logrus.Logger) *SemanticCacheManager {
	return &SemanticCacheManager{
		entries:             make(map[string]*SemanticCacheEntry),
		queryIndex:          make(map[string][]string),
		similarityThreshold: similarityThreshold,
		maxEntries:          maxEntries,
		ttl:                 ttl,
		logger:              logger,
		tracer:              otel.Tracer("ai.semantic_cache"),
		stats: &models.CacheStats{
			MaxSize:   int64(maxEntries),
			Timestamp: time.Now(),
		},
	}
}

// GetSimilar retrieves cached results for semantically similar queries
func (s *SemanticCacheManager) GetSimilar(ctx context.Context, query string, queryVector []float64) (*SemanticCacheEntry, float64, error) {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.GetSimilar")
	defer span.End()

	start := time.Now()

	// Find similar queries
	candidates := s.findCandidates(query)

	var bestMatch *SemanticCacheEntry
	var bestSimilarity float64

	for _, candidateID := range candidates {
		entry, exists := s.entries[candidateID]
		if !exists {
			continue
		}

		// Check if expired
		if time.Now().After(entry.ExpiresAt) {
			delete(s.entries, candidateID)
			s.removeFromIndex(candidateID)
			continue
		}

		// Calculate similarity
		var similarity float64
		if len(queryVector) > 0 && len(entry.QueryVector) > 0 {
			similarity = s.cosineSimilarity(queryVector, entry.QueryVector)
		} else {
			similarity = s.textSimilarity(query, entry.Query)
		}

		// Check if similarity meets threshold
		if similarity >= s.similarityThreshold && similarity > bestSimilarity {
			bestMatch = entry
			bestSimilarity = similarity
		}
	}

	if bestMatch != nil {
		// Update access statistics
		bestMatch.AccessCount++
		bestMatch.LastAccess = time.Now()

		s.stats.TotalHits++
		s.updateHitRate()

		s.logger.WithFields(logrus.Fields{
			"query":         query,
			"matched_query": bestMatch.Query,
			"similarity":    bestSimilarity,
			"access_count":  bestMatch.AccessCount,
			"search_time":   time.Since(start),
		}).Debug("Semantic cache hit")

		return bestMatch, bestSimilarity, nil
	}

	s.stats.TotalMisses++
	s.updateHitRate()

	s.logger.WithFields(logrus.Fields{
		"query":       query,
		"candidates":  len(candidates),
		"search_time": time.Since(start),
	}).Debug("Semantic cache miss")

	return nil, 0, nil
}

// Set stores a result in semantic cache
func (s *SemanticCacheManager) Set(ctx context.Context, query string, queryVector []float64, response interface{}, metadata map[string]interface{}, ttl time.Duration) error {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.Set")
	defer span.End()

	// Use provided TTL or default
	if ttl == 0 {
		ttl = s.ttl
	}

	// Check if we need to evict entries
	if len(s.entries) >= s.maxEntries {
		s.evictLRU()
	}

	// Generate unique ID
	entryID := GenerateKey("semantic", query, time.Now().UnixNano())

	// Create new entry
	entry := &SemanticCacheEntry{
		ID:          entryID,
		Query:       query,
		QueryVector: queryVector,
		Response:    response,
		Metadata:    metadata,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
		Confidence:  1.0,
	}

	// Store in cache
	s.entries[entryID] = entry
	s.addToIndex(entryID, query)
	s.stats.Size = int64(len(s.entries))

	s.logger.WithFields(logrus.Fields{
		"query":      query,
		"entry_id":   entryID,
		"ttl":        ttl,
		"cache_size": len(s.entries),
		"expires_at": entry.ExpiresAt,
	}).Debug("Semantic cache set")

	return nil
}

// Delete removes a cached result
func (s *SemanticCacheManager) Delete(ctx context.Context, entryID string) error {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.Delete")
	defer span.End()

	if entry, exists := s.entries[entryID]; exists {
		delete(s.entries, entryID)
		s.removeFromIndex(entryID)
		s.stats.Size = int64(len(s.entries))

		s.logger.WithFields(logrus.Fields{
			"entry_id": entryID,
			"query":    entry.Query,
		}).Debug("Semantic cache entry deleted")
	}

	return nil
}

// Clear clears all cached results
func (s *SemanticCacheManager) Clear(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.Clear")
	defer span.End()

	s.entries = make(map[string]*SemanticCacheEntry)
	s.queryIndex = make(map[string][]string)
	s.stats.Size = 0
	s.stats.Evictions = 0

	s.logger.Info("Semantic cache cleared")

	return nil
}

// GetStats returns cache statistics
func (s *SemanticCacheManager) GetStats(ctx context.Context) (*models.CacheStats, error) {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.GetStats")
	defer span.End()

	// Update current stats
	s.stats.Size = int64(len(s.entries))
	s.stats.Timestamp = time.Now()

	// Create a copy to return
	statsCopy := *s.stats
	return &statsCopy, nil
}

// findCandidates finds potential candidate entries for similarity comparison
func (s *SemanticCacheManager) findCandidates(query string) []string {
	var candidates []string
	queryWords := strings.Fields(strings.ToLower(query))

	// Simple keyword-based candidate selection
	for _, word := range queryWords {
		if len(word) > 3 { // Skip short words
			if entryIDs, exists := s.queryIndex[word]; exists {
				candidates = append(candidates, entryIDs...)
			}
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueCandidates []string
	for _, id := range candidates {
		if !seen[id] {
			seen[id] = true
			uniqueCandidates = append(uniqueCandidates, id)
		}
	}

	return uniqueCandidates
}

// addToIndex adds an entry to the text-based index
func (s *SemanticCacheManager) addToIndex(entryID, query string) {
	words := strings.Fields(strings.ToLower(query))
	for _, word := range words {
		if len(word) > 3 { // Skip short words
			s.queryIndex[word] = append(s.queryIndex[word], entryID)
		}
	}
}

// removeFromIndex removes an entry from the text-based index
func (s *SemanticCacheManager) removeFromIndex(entryID string) {
	for word, entryIDs := range s.queryIndex {
		for i, id := range entryIDs {
			if id == entryID {
				s.queryIndex[word] = append(entryIDs[:i], entryIDs[i+1:]...)
				break
			}
		}
		// Clean up empty entries
		if len(s.queryIndex[word]) == 0 {
			delete(s.queryIndex, word)
		}
	}
}

// cosineSimilarity calculates cosine similarity between two vectors
func (s *SemanticCacheManager) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// textSimilarity calculates simple text similarity using Jaccard index
func (s *SemanticCacheManager) textSimilarity(a, b string) float64 {
	wordsA := strings.Fields(strings.ToLower(a))
	wordsB := strings.Fields(strings.ToLower(b))

	setA := make(map[string]bool)
	setB := make(map[string]bool)

	for _, word := range wordsA {
		setA[word] = true
	}
	for _, word := range wordsB {
		setB[word] = true
	}

	// Calculate intersection
	intersection := 0
	for word := range setA {
		if setB[word] {
			intersection++
		}
	}

	// Calculate union
	union := len(setA) + len(setB) - intersection

	if union == 0 {
		return 0
	}

	return float64(intersection) / float64(union)
}

// evictLRU evicts the least recently used entry
func (s *SemanticCacheManager) evictLRU() {
	var oldestID string
	var oldestTime time.Time

	// Find the least recently used entry
	for id, entry := range s.entries {
		if oldestID == "" || entry.LastAccess.Before(oldestTime) {
			oldestID = id
			oldestTime = entry.LastAccess
		}
	}

	// Remove the oldest entry
	if oldestID != "" {
		delete(s.entries, oldestID)
		s.removeFromIndex(oldestID)
		s.stats.Evictions++

		s.logger.WithFields(logrus.Fields{
			"evicted_id":  oldestID,
			"last_access": oldestTime,
			"cache_size":  len(s.entries),
		}).Debug("Semantic cache entry evicted")
	}
}

// updateHitRate calculates and updates the hit rate
func (s *SemanticCacheManager) updateHitRate() {
	total := s.stats.TotalHits + s.stats.TotalMisses
	if total > 0 {
		s.stats.HitRate = float64(s.stats.TotalHits) / float64(total)
		s.stats.MissRate = float64(s.stats.TotalMisses) / float64(total)
	}
}

// CleanupExpired removes expired entries
func (s *SemanticCacheManager) CleanupExpired(ctx context.Context) error {
	ctx, span := s.tracer.Start(ctx, "semantic_cache.CleanupExpired")
	defer span.End()

	now := time.Now()
	var expiredIDs []string

	// Find expired entries
	for id, entry := range s.entries {
		if now.After(entry.ExpiresAt) {
			expiredIDs = append(expiredIDs, id)
		}
	}

	// Remove expired entries
	for _, id := range expiredIDs {
		delete(s.entries, id)
		s.removeFromIndex(id)
	}

	s.stats.Size = int64(len(s.entries))

	if len(expiredIDs) > 0 {
		s.logger.WithFields(logrus.Fields{
			"expired_count": len(expiredIDs),
			"cache_size":    len(s.entries),
		}).Debug("Expired semantic cache entries cleaned up")
	}

	return nil
}
