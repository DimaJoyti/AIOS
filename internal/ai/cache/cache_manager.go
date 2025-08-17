package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string
	Value     interface{}
	ExpiresAt time.Time
	CreatedAt time.Time
	AccessCount int64
	LastAccess  time.Time
}

// MemoryCacheManager implements in-memory caching
type MemoryCacheManager struct {
	cache     map[string]*CacheEntry
	mu        sync.RWMutex
	maxSize   int64
	currentSize int64
	ttl       time.Duration
	logger    *logrus.Logger
	tracer    trace.Tracer
	stats     *models.CacheStats
}

// NewMemoryCacheManager creates a new in-memory cache manager
func NewMemoryCacheManager(maxSize int64, ttl time.Duration, logger *logrus.Logger) *MemoryCacheManager {
	return &MemoryCacheManager{
		cache:   make(map[string]*CacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
		logger:  logger,
		tracer:  otel.Tracer("ai.cache_manager"),
		stats: &models.CacheStats{
			MaxSize:   maxSize,
			Timestamp: time.Now(),
		},
	}
}

// Get retrieves a cached result
func (c *MemoryCacheManager) Get(ctx context.Context, key string) (interface{}, bool, error) {
	ctx, span := c.tracer.Start(ctx, "cache.Get")
	defer span.End()

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		c.stats.TotalMisses++
		c.updateHitRate()
		return nil, false, nil
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.cache, key)
		c.currentSize--
		c.mu.Unlock()
		c.mu.RLock()
		
		c.stats.TotalMisses++
		c.updateHitRate()
		return nil, false, nil
	}

	// Update access statistics
	entry.AccessCount++
	entry.LastAccess = time.Now()
	
	c.stats.TotalHits++
	c.updateHitRate()

	c.logger.WithFields(logrus.Fields{
		"key":          key,
		"access_count": entry.AccessCount,
		"age":          time.Since(entry.CreatedAt),
	}).Debug("Cache hit")

	return entry.Value, true, nil
}

// Set stores a result in cache
func (c *MemoryCacheManager) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	ctx, span := c.tracer.Start(ctx, "cache.Set")
	defer span.End()

	c.mu.Lock()
	defer c.mu.Unlock()

	// Use provided TTL or default
	if ttl == 0 {
		ttl = c.ttl
	}

	// Check if we need to evict entries
	if c.currentSize >= c.maxSize {
		c.evictLRU()
	}

	// Create new entry
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		ExpiresAt:   time.Now().Add(ttl),
		CreatedAt:   time.Now(),
		AccessCount: 0,
		LastAccess:  time.Now(),
	}

	// Store in cache
	c.cache[key] = entry
	c.currentSize++
	c.stats.Size = c.currentSize

	c.logger.WithFields(logrus.Fields{
		"key":         key,
		"ttl":         ttl,
		"cache_size":  c.currentSize,
		"expires_at":  entry.ExpiresAt,
	}).Debug("Cache set")

	return nil
}

// Delete removes a cached result
func (c *MemoryCacheManager) Delete(ctx context.Context, key string) error {
	ctx, span := c.tracer.Start(ctx, "cache.Delete")
	defer span.End()

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.cache[key]; exists {
		delete(c.cache, key)
		c.currentSize--
		c.stats.Size = c.currentSize

		c.logger.WithField("key", key).Debug("Cache entry deleted")
	}

	return nil
}

// Clear clears all cached results
func (c *MemoryCacheManager) Clear(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "cache.Clear")
	defer span.End()

	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CacheEntry)
	c.currentSize = 0
	c.stats.Size = 0
	c.stats.Evictions = 0

	c.logger.Info("Cache cleared")

	return nil
}

// GetStats returns cache statistics
func (c *MemoryCacheManager) GetStats(ctx context.Context) (*models.CacheStats, error) {
	ctx, span := c.tracer.Start(ctx, "cache.GetStats")
	defer span.End()

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update current stats
	c.stats.Size = c.currentSize
	c.stats.Timestamp = time.Now()

	// Create a copy to return
	statsCopy := *c.stats
	return &statsCopy, nil
}

// evictLRU evicts the least recently used entry
func (c *MemoryCacheManager) evictLRU() {
	var oldestKey string
	var oldestTime time.Time

	// Find the least recently used entry
	for key, entry := range c.cache {
		if oldestKey == "" || entry.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.LastAccess
		}
	}

	// Remove the oldest entry
	if oldestKey != "" {
		delete(c.cache, oldestKey)
		c.currentSize--
		c.stats.Evictions++

		c.logger.WithFields(logrus.Fields{
			"evicted_key":  oldestKey,
			"last_access":  oldestTime,
			"cache_size":   c.currentSize,
		}).Debug("Cache entry evicted")
	}
}

// updateHitRate calculates and updates the hit rate
func (c *MemoryCacheManager) updateHitRate() {
	total := c.stats.TotalHits + c.stats.TotalMisses
	if total > 0 {
		c.stats.HitRate = float64(c.stats.TotalHits) / float64(total)
		c.stats.MissRate = float64(c.stats.TotalMisses) / float64(total)
	}
}

// CleanupExpired removes expired entries
func (c *MemoryCacheManager) CleanupExpired(ctx context.Context) error {
	ctx, span := c.tracer.Start(ctx, "cache.CleanupExpired")
	defer span.End()

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	var expiredKeys []string

	// Find expired entries
	for key, entry := range c.cache {
		if now.After(entry.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// Remove expired entries
	for _, key := range expiredKeys {
		delete(c.cache, key)
		c.currentSize--
	}

	c.stats.Size = c.currentSize

	if len(expiredKeys) > 0 {
		c.logger.WithFields(logrus.Fields{
			"expired_count": len(expiredKeys),
			"cache_size":    c.currentSize,
		}).Debug("Expired cache entries cleaned up")
	}

	return nil
}

// GenerateKey generates a cache key from input parameters
func GenerateKey(prefix string, params ...interface{}) string {
	hasher := sha256.New()
	hasher.Write([]byte(prefix))
	
	for _, param := range params {
		hasher.Write([]byte(fmt.Sprintf("%v", param)))
	}
	
	return hex.EncodeToString(hasher.Sum(nil))[:16] // Use first 16 characters
}

// StartCleanupRoutine starts a background routine to clean up expired entries
func (c *MemoryCacheManager) StartCleanupRoutine(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.CleanupExpired(ctx); err != nil {
					c.logger.WithError(err).Error("Failed to cleanup expired cache entries")
				}
			}
		}
	}()
}
