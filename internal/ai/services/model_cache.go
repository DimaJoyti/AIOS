package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ModelCache provides caching for AI model responses
type ModelCache struct {
	textCache  map[string]*CachedTextResponse
	imageCache map[string]*CachedImageResponse
	audioCache map[string]*CachedAudioResponse
	videoCache map[string]*CachedVideoResponse
	mutex      sync.RWMutex
	logger     *logrus.Logger
	maxSize    int
	defaultTTL time.Duration
	stats      *CacheStats
}

// CachedTextResponse represents a cached text response
type CachedTextResponse struct {
	Response  *TextGenerationResponse `json:"response"`
	CreatedAt time.Time               `json:"created_at"`
	ExpiresAt time.Time               `json:"expires_at"`
	HitCount  int                     `json:"hit_count"`
	LastHit   time.Time               `json:"last_hit"`
}

// CachedImageResponse represents a cached image response
type CachedImageResponse struct {
	Response  *ImageGenerationResponse `json:"response"`
	CreatedAt time.Time                `json:"created_at"`
	ExpiresAt time.Time                `json:"expires_at"`
	HitCount  int                      `json:"hit_count"`
	LastHit   time.Time                `json:"last_hit"`
}

// CachedAudioResponse represents a cached audio response
type CachedAudioResponse struct {
	Response  *AudioProcessingResponse `json:"response"`
	CreatedAt time.Time                `json:"created_at"`
	ExpiresAt time.Time                `json:"expires_at"`
	HitCount  int                      `json:"hit_count"`
	LastHit   time.Time                `json:"last_hit"`
}

// CachedVideoResponse represents a cached video response
type CachedVideoResponse struct {
	Response  *VideoProcessingResponse `json:"response"`
	CreatedAt time.Time                `json:"created_at"`
	ExpiresAt time.Time                `json:"expires_at"`
	HitCount  int                      `json:"hit_count"`
	LastHit   time.Time                `json:"last_hit"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalRequests int64     `json:"total_requests"`
	CacheHits     int64     `json:"cache_hits"`
	CacheMisses   int64     `json:"cache_misses"`
	HitRate       float64   `json:"hit_rate"`
	TotalSize     int       `json:"total_size"`
	MaxSize       int       `json:"max_size"`
	Evictions     int64     `json:"evictions"`
	LastCleanup   time.Time `json:"last_cleanup"`
}

// NewModelCache creates a new model cache
func NewModelCache(logger *logrus.Logger) *ModelCache {
	return &ModelCache{
		textCache:  make(map[string]*CachedTextResponse),
		imageCache: make(map[string]*CachedImageResponse),
		audioCache: make(map[string]*CachedAudioResponse),
		videoCache: make(map[string]*CachedVideoResponse),
		logger:     logger,
		maxSize:    10000, // Default max cache entries
		defaultTTL: 1 * time.Hour,
		stats: &CacheStats{
			MaxSize: 10000,
		},
	}
}

// GetTextResponse retrieves a cached text response
func (mc *ModelCache) GetTextResponse(request *TextGenerationRequest) *TextGenerationResponse {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.generateTextCacheKey(request)
	cached, exists := mc.textCache[key]

	mc.stats.TotalRequests++

	if !exists || time.Now().After(cached.ExpiresAt) {
		mc.stats.CacheMisses++
		mc.updateHitRate()
		return nil
	}

	// Update hit statistics
	cached.HitCount++
	cached.LastHit = time.Now()
	mc.stats.CacheHits++
	mc.updateHitRate()

	mc.logger.WithFields(logrus.Fields{
		"cache_key": key,
		"hit_count": cached.HitCount,
	}).Debug("Cache hit for text response")

	return cached.Response
}

// SetTextResponse caches a text response
func (mc *ModelCache) SetTextResponse(request *TextGenerationRequest, response *TextGenerationResponse) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.generateTextCacheKey(request)
	now := time.Now()

	cached := &CachedTextResponse{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(mc.defaultTTL),
		HitCount:  0,
		LastHit:   now,
	}

	mc.textCache[key] = cached
	mc.stats.TotalSize++

	// Check if we need to evict entries
	if mc.stats.TotalSize > mc.maxSize {
		mc.evictOldestEntries()
	}

	mc.logger.WithField("cache_key", key).Debug("Cached text response")
}

// GetImageResponse retrieves a cached image response
func (mc *ModelCache) GetImageResponse(request *ImageGenerationRequest) *ImageGenerationResponse {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.generateImageCacheKey(request)
	cached, exists := mc.imageCache[key]

	mc.stats.TotalRequests++

	if !exists || time.Now().After(cached.ExpiresAt) {
		mc.stats.CacheMisses++
		mc.updateHitRate()
		return nil
	}

	cached.HitCount++
	cached.LastHit = time.Now()
	mc.stats.CacheHits++
	mc.updateHitRate()

	return cached.Response
}

// SetImageResponse caches an image response
func (mc *ModelCache) SetImageResponse(request *ImageGenerationRequest, response *ImageGenerationResponse) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.generateImageCacheKey(request)
	now := time.Now()

	cached := &CachedImageResponse{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(mc.defaultTTL),
		HitCount:  0,
		LastHit:   now,
	}

	mc.imageCache[key] = cached
	mc.stats.TotalSize++

	if mc.stats.TotalSize > mc.maxSize {
		mc.evictOldestEntries()
	}
}

// GetAudioResponse retrieves a cached audio response
func (mc *ModelCache) GetAudioResponse(request *AudioProcessingRequest) *AudioProcessingResponse {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.generateAudioCacheKey(request)
	cached, exists := mc.audioCache[key]

	mc.stats.TotalRequests++

	if !exists || time.Now().After(cached.ExpiresAt) {
		mc.stats.CacheMisses++
		mc.updateHitRate()
		return nil
	}

	cached.HitCount++
	cached.LastHit = time.Now()
	mc.stats.CacheHits++
	mc.updateHitRate()

	return cached.Response
}

// SetAudioResponse caches an audio response
func (mc *ModelCache) SetAudioResponse(request *AudioProcessingRequest, response *AudioProcessingResponse) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.generateAudioCacheKey(request)
	now := time.Now()

	cached := &CachedAudioResponse{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(mc.defaultTTL),
		HitCount:  0,
		LastHit:   now,
	}

	mc.audioCache[key] = cached
	mc.stats.TotalSize++

	if mc.stats.TotalSize > mc.maxSize {
		mc.evictOldestEntries()
	}
}

// GetVideoResponse retrieves a cached video response
func (mc *ModelCache) GetVideoResponse(request *VideoProcessingRequest) *VideoProcessingResponse {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	key := mc.generateVideoCacheKey(request)
	cached, exists := mc.videoCache[key]

	mc.stats.TotalRequests++

	if !exists || time.Now().After(cached.ExpiresAt) {
		mc.stats.CacheMisses++
		mc.updateHitRate()
		return nil
	}

	cached.HitCount++
	cached.LastHit = time.Now()
	mc.stats.CacheHits++
	mc.updateHitRate()

	return cached.Response
}

// SetVideoResponse caches a video response
func (mc *ModelCache) SetVideoResponse(request *VideoProcessingRequest, response *VideoProcessingResponse) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	key := mc.generateVideoCacheKey(request)
	now := time.Now()

	cached := &CachedVideoResponse{
		Response:  response,
		CreatedAt: now,
		ExpiresAt: now.Add(mc.defaultTTL),
		HitCount:  0,
		LastHit:   now,
	}

	mc.videoCache[key] = cached
	mc.stats.TotalSize++

	if mc.stats.TotalSize > mc.maxSize {
		mc.evictOldestEntries()
	}
}

// GetStats returns cache statistics
func (mc *ModelCache) GetStats() *CacheStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	return &CacheStats{
		TotalRequests: mc.stats.TotalRequests,
		CacheHits:     mc.stats.CacheHits,
		CacheMisses:   mc.stats.CacheMisses,
		HitRate:       mc.stats.HitRate,
		TotalSize:     mc.stats.TotalSize,
		MaxSize:       mc.stats.MaxSize,
		Evictions:     mc.stats.Evictions,
		LastCleanup:   mc.stats.LastCleanup,
	}
}

// ClearCache clears all cached responses
func (mc *ModelCache) ClearCache() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	mc.textCache = make(map[string]*CachedTextResponse)
	mc.imageCache = make(map[string]*CachedImageResponse)
	mc.audioCache = make(map[string]*CachedAudioResponse)
	mc.videoCache = make(map[string]*CachedVideoResponse)
	mc.stats.TotalSize = 0

	mc.logger.Info("Cache cleared")
}

// CleanupExpired removes expired cache entries
func (mc *ModelCache) CleanupExpired() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	now := time.Now()
	removed := 0

	// Clean text cache
	for key, cached := range mc.textCache {
		if now.After(cached.ExpiresAt) {
			delete(mc.textCache, key)
			removed++
		}
	}

	// Clean image cache
	for key, cached := range mc.imageCache {
		if now.After(cached.ExpiresAt) {
			delete(mc.imageCache, key)
			removed++
		}
	}

	// Clean audio cache
	for key, cached := range mc.audioCache {
		if now.After(cached.ExpiresAt) {
			delete(mc.audioCache, key)
			removed++
		}
	}

	// Clean video cache
	for key, cached := range mc.videoCache {
		if now.After(cached.ExpiresAt) {
			delete(mc.videoCache, key)
			removed++
		}
	}

	mc.stats.TotalSize -= removed
	mc.stats.LastCleanup = now

	if removed > 0 {
		mc.logger.WithField("removed_entries", removed).Info("Cleaned up expired cache entries")
	}
}

// generateTextCacheKey generates a cache key for text requests
func (mc *ModelCache) generateTextCacheKey(request *TextGenerationRequest) string {
	// Create a deterministic key based on request parameters
	data := map[string]interface{}{
		"model_id":      request.ModelID,
		"prompt":        request.Prompt,
		"system_prompt": request.SystemPrompt,
		"messages":      request.Messages,
		"config":        request.Config,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// generateImageCacheKey generates a cache key for image requests
func (mc *ModelCache) generateImageCacheKey(request *ImageGenerationRequest) string {
	data := map[string]interface{}{
		"model_id":        request.ModelID,
		"prompt":          request.Prompt,
		"negative_prompt": request.NegativePrompt,
		"width":           request.Width,
		"height":          request.Height,
		"steps":           request.Steps,
		"guidance":        request.Guidance,
		"seed":            request.Seed,
		"count":           request.Count,
		"format":          request.Format,
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// generateAudioCacheKey generates a cache key for audio requests
func (mc *ModelCache) generateAudioCacheKey(request *AudioProcessingRequest) string {
	// For audio, we hash the audio data along with other parameters
	hasher := sha256.New()
	hasher.Write([]byte(request.ModelID))
	hasher.Write(request.AudioData)
	hasher.Write([]byte(request.Format))
	hasher.Write([]byte(request.Task))
	hasher.Write([]byte(request.Language))
	hasher.Write([]byte(request.Prompt))

	return hex.EncodeToString(hasher.Sum(nil))
}

// generateVideoCacheKey generates a cache key for video requests
func (mc *ModelCache) generateVideoCacheKey(request *VideoProcessingRequest) string {
	// For video, we hash the video data along with other parameters
	hasher := sha256.New()
	hasher.Write([]byte(request.ModelID))
	hasher.Write(request.VideoData)
	hasher.Write([]byte(request.Format))
	hasher.Write([]byte(request.Task))
	hasher.Write([]byte(request.Prompt))

	return hex.EncodeToString(hasher.Sum(nil))
}

// evictOldestEntries removes the oldest cache entries when cache is full
func (mc *ModelCache) evictOldestEntries() {
	// Simple LRU eviction - remove 10% of entries
	evictCount := mc.maxSize / 10
	evicted := 0

	// Find oldest entries across all caches
	type cacheEntry struct {
		key       string
		createdAt time.Time
		cacheType string
	}

	var entries []cacheEntry

	// Collect all entries
	for key, cached := range mc.textCache {
		entries = append(entries, cacheEntry{key, cached.CreatedAt, "text"})
	}
	for key, cached := range mc.imageCache {
		entries = append(entries, cacheEntry{key, cached.CreatedAt, "image"})
	}
	for key, cached := range mc.audioCache {
		entries = append(entries, cacheEntry{key, cached.CreatedAt, "audio"})
	}
	for key, cached := range mc.videoCache {
		entries = append(entries, cacheEntry{key, cached.CreatedAt, "video"})
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].createdAt.After(entries[j].createdAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove oldest entries
	for _, entry := range entries {
		if evicted >= evictCount {
			break
		}

		switch entry.cacheType {
		case "text":
			delete(mc.textCache, entry.key)
		case "image":
			delete(mc.imageCache, entry.key)
		case "audio":
			delete(mc.audioCache, entry.key)
		case "video":
			delete(mc.videoCache, entry.key)
		}

		evicted++
	}

	mc.stats.TotalSize -= evicted
	mc.stats.Evictions += int64(evicted)

	mc.logger.WithField("evicted_entries", evicted).Debug("Evicted cache entries")
}

// updateHitRate updates the cache hit rate
func (mc *ModelCache) updateHitRate() {
	if mc.stats.TotalRequests > 0 {
		mc.stats.HitRate = float64(mc.stats.CacheHits) / float64(mc.stats.TotalRequests)
	}
}
