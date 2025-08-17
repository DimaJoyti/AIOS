package security

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity     int           // Maximum number of tokens
	tokens       int           // Current number of tokens
	refillRate   int           // Tokens added per second
	lastRefill   time.Time     // Last time tokens were added
	mu           sync.Mutex
}

// RateLimitStats represents rate limiting statistics
type RateLimitStats struct {
	TotalRequests   int64                    `json:"total_requests"`
	AllowedRequests int64                    `json:"allowed_requests"`
	BlockedRequests int64                    `json:"blocked_requests"`
	ActiveBuckets   int                      `json:"active_buckets"`
	BucketStats     map[string]BucketStats   `json:"bucket_stats"`
}

// BucketStats represents statistics for a specific bucket
type BucketStats struct {
	ID             string    `json:"id"`
	Capacity       int       `json:"capacity"`
	CurrentTokens  int       `json:"current_tokens"`
	RefillRate     int       `json:"refill_rate"`
	LastRefill     time.Time `json:"last_refill"`
	TotalRequests  int64     `json:"total_requests"`
	AllowedRequests int64    `json:"allowed_requests"`
	BlockedRequests int64    `json:"blocked_requests"`
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*TokenBucket),
	}
}

// Allow checks if a request is allowed for the given identifier
func (rl *RateLimiter) Allow(identifier string, rateLimit int) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rateLimit,
			tokens:     rateLimit,
			refillRate: rateLimit / 60, // Refill rate per second (assuming rate limit is per minute)
			lastRefill: time.Now(),
		}
		if bucket.refillRate == 0 {
			bucket.refillRate = 1 // Minimum refill rate
		}
		rl.buckets[identifier] = bucket
	}
	rl.mu.Unlock()

	return bucket.consume()
}

// AllowN checks if N requests are allowed for the given identifier
func (rl *RateLimiter) AllowN(identifier string, rateLimit int, n int) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rateLimit,
			tokens:     rateLimit,
			refillRate: rateLimit / 60,
			lastRefill: time.Now(),
		}
		if bucket.refillRate == 0 {
			bucket.refillRate = 1
		}
		rl.buckets[identifier] = bucket
	}
	rl.mu.Unlock()

	return bucket.consumeN(n)
}

// GetRemainingTokens returns the number of remaining tokens for an identifier
func (rl *RateLimiter) GetRemainingTokens(identifier string) int {
	rl.mu.RLock()
	bucket, exists := rl.buckets[identifier]
	rl.mu.RUnlock()

	if !exists {
		return 0
	}

	bucket.refill()
	bucket.mu.Lock()
	defer bucket.mu.Unlock()
	return bucket.tokens
}

// SetRateLimit updates the rate limit for an identifier
func (rl *RateLimiter) SetRateLimit(identifier string, rateLimit int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[identifier]
	if !exists {
		bucket = &TokenBucket{
			capacity:   rateLimit,
			tokens:     rateLimit,
			refillRate: rateLimit / 60,
			lastRefill: time.Now(),
		}
		if bucket.refillRate == 0 {
			bucket.refillRate = 1
		}
		rl.buckets[identifier] = bucket
		return
	}

	bucket.mu.Lock()
	bucket.capacity = rateLimit
	bucket.refillRate = rateLimit / 60
	if bucket.refillRate == 0 {
		bucket.refillRate = 1
	}
	// Adjust current tokens if they exceed new capacity
	if bucket.tokens > bucket.capacity {
		bucket.tokens = bucket.capacity
	}
	bucket.mu.Unlock()
}

// RemoveIdentifier removes rate limiting for an identifier
func (rl *RateLimiter) RemoveIdentifier(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.buckets, identifier)
}

// GetStats returns rate limiting statistics
func (rl *RateLimiter) GetStats() RateLimitStats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := RateLimitStats{
		ActiveBuckets: len(rl.buckets),
		BucketStats:   make(map[string]BucketStats),
	}

	for id, bucket := range rl.buckets {
		bucket.mu.Lock()
		bucketStats := BucketStats{
			ID:            id,
			Capacity:      bucket.capacity,
			CurrentTokens: bucket.tokens,
			RefillRate:    bucket.refillRate,
			LastRefill:    bucket.lastRefill,
		}
		bucket.mu.Unlock()

		stats.BucketStats[id] = bucketStats
	}

	return stats
}

// CleanupInactive removes inactive buckets (buckets that haven't been used recently)
func (rl *RateLimiter) CleanupInactive(maxAge time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	for id, bucket := range rl.buckets {
		bucket.mu.Lock()
		lastUsed := bucket.lastRefill
		bucket.mu.Unlock()

		if lastUsed.Before(cutoff) {
			delete(rl.buckets, id)
		}
	}
}

// TokenBucket methods

// consume attempts to consume one token from the bucket
func (tb *TokenBucket) consume() bool {
	return tb.consumeN(1)
}

// consumeN attempts to consume N tokens from the bucket
func (tb *TokenBucket) consumeN(n int) bool {
	tb.refill()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	if tb.tokens >= n {
		tb.tokens -= n
		return true
	}
	return false
}

// refill adds tokens to the bucket based on elapsed time
func (tb *TokenBucket) refill() {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// Calculate tokens to add based on elapsed time
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate
	
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}
}

// AdvancedRateLimiter provides more sophisticated rate limiting
type AdvancedRateLimiter struct {
	buckets    map[string]*AdvancedTokenBucket
	mu         sync.RWMutex
	globalStats *GlobalRateLimitStats
}

// AdvancedTokenBucket provides burst handling and sliding window
type AdvancedTokenBucket struct {
	capacity       int           // Maximum tokens
	tokens         int           // Current tokens
	refillRate     int           // Tokens per second
	burstCapacity  int           // Maximum burst size
	lastRefill     time.Time     // Last refill time
	requestHistory []time.Time   // Sliding window of requests
	windowSize     time.Duration // Size of sliding window
	mu             sync.Mutex
}

// GlobalRateLimitStats tracks global rate limiting statistics
type GlobalRateLimitStats struct {
	TotalRequests     int64     `json:"total_requests"`
	AllowedRequests   int64     `json:"allowed_requests"`
	BlockedRequests   int64     `json:"blocked_requests"`
	AverageLatency    float64   `json:"average_latency_ms"`
	PeakRequestsPerSec int64    `json:"peak_requests_per_sec"`
	LastReset         time.Time `json:"last_reset"`
	mu                sync.RWMutex
}

// NewAdvancedRateLimiter creates a new advanced rate limiter
func NewAdvancedRateLimiter() *AdvancedRateLimiter {
	return &AdvancedRateLimiter{
		buckets: make(map[string]*AdvancedTokenBucket),
		globalStats: &GlobalRateLimitStats{
			LastReset: time.Now(),
		},
	}
}

// AllowWithBurst checks if a request is allowed with burst handling
func (arl *AdvancedRateLimiter) AllowWithBurst(identifier string, rateLimit, burstLimit int) bool {
	arl.mu.Lock()
	bucket, exists := arl.buckets[identifier]
	if !exists {
		bucket = &AdvancedTokenBucket{
			capacity:       rateLimit,
			tokens:         rateLimit,
			refillRate:     rateLimit / 60,
			burstCapacity:  burstLimit,
			lastRefill:     time.Now(),
			requestHistory: make([]time.Time, 0),
			windowSize:     time.Minute,
		}
		if bucket.refillRate == 0 {
			bucket.refillRate = 1
		}
		arl.buckets[identifier] = bucket
	}
	arl.mu.Unlock()

	allowed := bucket.consumeWithBurst()
	
	// Update global stats
	arl.globalStats.mu.Lock()
	arl.globalStats.TotalRequests++
	if allowed {
		arl.globalStats.AllowedRequests++
	} else {
		arl.globalStats.BlockedRequests++
	}
	arl.globalStats.mu.Unlock()

	return allowed
}

// consumeWithBurst attempts to consume a token with burst handling
func (atb *AdvancedTokenBucket) consumeWithBurst() bool {
	atb.refill()
	atb.cleanupHistory()

	atb.mu.Lock()
	defer atb.mu.Unlock()

	now := time.Now()
	
	// Check sliding window rate limit
	recentRequests := 0
	for _, reqTime := range atb.requestHistory {
		if now.Sub(reqTime) <= atb.windowSize {
			recentRequests++
		}
	}

	// Allow if within burst capacity or regular capacity
	if recentRequests < atb.burstCapacity && atb.tokens > 0 {
		atb.tokens--
		atb.requestHistory = append(atb.requestHistory, now)
		return true
	}

	return false
}

// refill adds tokens to the advanced bucket
func (atb *AdvancedTokenBucket) refill() {
	atb.mu.Lock()
	defer atb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(atb.lastRefill)
	
	tokensToAdd := int(elapsed.Seconds()) * atb.refillRate
	
	if tokensToAdd > 0 {
		atb.tokens += tokensToAdd
		if atb.tokens > atb.capacity {
			atb.tokens = atb.capacity
		}
		atb.lastRefill = now
	}
}

// cleanupHistory removes old entries from request history
func (atb *AdvancedTokenBucket) cleanupHistory() {
	atb.mu.Lock()
	defer atb.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-atb.windowSize)
	
	// Remove old entries
	validRequests := make([]time.Time, 0, len(atb.requestHistory))
	for _, reqTime := range atb.requestHistory {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	atb.requestHistory = validRequests
}

// GetGlobalStats returns global rate limiting statistics
func (arl *AdvancedRateLimiter) GetGlobalStats() GlobalRateLimitStats {
	arl.globalStats.mu.RLock()
	defer arl.globalStats.mu.RUnlock()
	
	stats := *arl.globalStats
	return stats
}

// ResetGlobalStats resets global statistics
func (arl *AdvancedRateLimiter) ResetGlobalStats() {
	arl.globalStats.mu.Lock()
	defer arl.globalStats.mu.Unlock()
	
	arl.globalStats.TotalRequests = 0
	arl.globalStats.AllowedRequests = 0
	arl.globalStats.BlockedRequests = 0
	arl.globalStats.AverageLatency = 0
	arl.globalStats.PeakRequestsPerSec = 0
	arl.globalStats.LastReset = time.Now()
}
