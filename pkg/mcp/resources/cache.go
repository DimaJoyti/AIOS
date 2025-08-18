package resources

import (
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
)

// CacheStats represents cache statistics for the memory cache
type MemoryCacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Entries     int       `json:"entries"`
	HitRate     float64   `json:"hit_rate"`
	TotalSize   int64     `json:"total_size"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	Entries     int       `json:"entries"`
	HitRate     float64   `json:"hit_rate"`
	TotalSize   int64     `json:"total_size"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// CacheEntry represents a cached resource entry
type CacheEntry struct {
	Content     []*protocol.ResourceContent
	ExpiresAt   time.Time
	Size        int64
	AccessCount int64
	LastAccess  time.Time
}

// MemoryResourceCache implements ResourceCache using in-memory storage
type MemoryResourceCache struct {
	entries       map[string]*CacheEntry
	mu            sync.RWMutex
	maxSize       int64
	currentSize   int64
	defaultTTL    time.Duration
	stats         CacheStats
	logger        *logrus.Logger
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

// NewMemoryResourceCache creates a new memory-based resource cache
func NewMemoryResourceCache(logger *logrus.Logger) (ResourceCache, error) {
	cache := &MemoryResourceCache{
		entries:     make(map[string]*CacheEntry),
		maxSize:     100 * 1024 * 1024, // 100MB default
		defaultTTL:  30 * time.Minute,
		logger:      logger,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine
	cache.cleanupTicker = time.NewTicker(5 * time.Minute)
	go cache.cleanupExpired()

	return cache, nil
}

// Get retrieves content from cache
func (c *MemoryResourceCache) Get(uri string) ([]protocol.ResourceContent, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[uri]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		c.mu.RUnlock()
		c.mu.Lock()
		delete(c.entries, uri)
		c.currentSize -= entry.Size
		c.mu.Unlock()
		c.mu.RLock()
		c.stats.Misses++
		return nil, false
	}

	// Update access stats
	entry.AccessCount++
	entry.LastAccess = time.Now()
	c.stats.Hits++

	// Convert from []*protocol.ResourceContent to []protocol.ResourceContent
	result := make([]protocol.ResourceContent, len(entry.Content))
	for i, content := range entry.Content {
		result[i] = *content
	}
	return result, true
}

// Set stores content in cache
func (c *MemoryResourceCache) Set(uri string, content []protocol.ResourceContent, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Calculate content size
	var size int64
	for _, item := range content {
		size += int64(len(item.Text) + len(item.Blob))
	}

	// Check if we need to make space
	if c.currentSize+size > c.maxSize {
		c.evictLRU(size)
	}

	// Remove existing entry if present
	if existing, exists := c.entries[uri]; exists {
		c.currentSize -= existing.Size
	}

	// Use default TTL if not specified
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	// Convert content to pointer slice
	contentPtrs := make([]*protocol.ResourceContent, len(content))
	for i := range content {
		contentPtrs[i] = &content[i]
	}

	// Create new entry
	entry := &CacheEntry{
		Content:     contentPtrs,
		ExpiresAt:   time.Now().Add(ttl),
		Size:        size,
		AccessCount: 1,
		LastAccess:  time.Now(),
	}

	c.entries[uri] = entry
	c.currentSize += size
	c.stats.Entries = len(c.entries)
	c.stats.TotalSize = c.currentSize

	c.logger.WithFields(logrus.Fields{
		"uri":  uri,
		"size": size,
		"ttl":  ttl,
	}).Debug("Resource cached")

	return nil
}

// Delete removes content from cache
func (c *MemoryResourceCache) Delete(uri string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, exists := c.entries[uri]; exists {
		delete(c.entries, uri)
		c.currentSize -= entry.Size
		c.stats.Entries = len(c.entries)
		c.stats.TotalSize = c.currentSize
	}
	return nil
}

// Clear removes all content from cache
func (c *MemoryResourceCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*CacheEntry)
	c.currentSize = 0
	c.stats.Entries = 0
	c.stats.TotalSize = 0
	return nil
}

// Size returns the number of cached entries
func (c *MemoryResourceCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// Stats returns cache statistics
func (c *MemoryResourceCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := c.stats
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}
	return stats
}

// evictLRU evicts least recently used entries to make space
func (c *MemoryResourceCache) evictLRU(neededSpace int64) {
	// Find entries to evict based on LRU
	type entryInfo struct {
		uri        string
		lastAccess time.Time
		size       int64
	}

	var entries []entryInfo
	for uri, entry := range c.entries {
		entries = append(entries, entryInfo{
			uri:        uri,
			lastAccess: entry.LastAccess,
			size:       entry.Size,
		})
	}

	// Sort by last access time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].lastAccess.After(entries[j].lastAccess) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Evict entries until we have enough space
	var freedSpace int64
	for _, entry := range entries {
		if c.currentSize-freedSpace+neededSpace <= c.maxSize {
			break
		}

		delete(c.entries, entry.uri)
		freedSpace += entry.size
		c.logger.WithField("uri", entry.uri).Debug("Evicted cache entry")
	}

	c.currentSize -= freedSpace
}

// cleanupExpired removes expired entries
func (c *MemoryResourceCache) cleanupExpired() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.mu.Lock()
			now := time.Now()
			var expiredURIs []string

			for uri, entry := range c.entries {
				if now.After(entry.ExpiresAt) {
					expiredURIs = append(expiredURIs, uri)
				}
			}

			for _, uri := range expiredURIs {
				if entry, exists := c.entries[uri]; exists {
					delete(c.entries, uri)
					c.currentSize -= entry.Size
				}
			}

			c.stats.Entries = len(c.entries)
			c.stats.TotalSize = c.currentSize
			c.stats.LastCleanup = now
			c.mu.Unlock()

			if len(expiredURIs) > 0 {
				c.logger.WithField("expired_count", len(expiredURIs)).Debug("Cleaned up expired cache entries")
			}

		case <-c.stopCleanup:
			c.cleanupTicker.Stop()
			return
		}
	}
}

// GetStats returns cache statistics
func (c *MemoryResourceCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"hits":         c.stats.Hits,
		"misses":       c.stats.Misses,
		"entries":      c.stats.Entries,
		"hit_rate":     c.stats.HitRate,
		"total_size":   c.stats.TotalSize,
		"last_cleanup": c.stats.LastCleanup,
	}
}

// SetTTL sets the default TTL for cached content
func (c *MemoryResourceCache) SetTTL(ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTTL = ttl
}

// Close stops the cache cleanup goroutine
func (c *MemoryResourceCache) Close() error {
	close(c.stopCleanup)
	return nil
}
