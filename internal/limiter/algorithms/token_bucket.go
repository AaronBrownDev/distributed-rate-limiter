package algorithms

import (
	"sync"
	"time"
)

// TokenBucket represents a thread-safe token bucket rate limiter.
type TokenBucket struct {
	capacity     int64         // Max tokens the bucket can hold
	tokens       int64         // Current token count
	refillRate   int64         // Tokens per period
	refillPeriod time.Duration // How often to add tokens
	lastRefill   time.Time     // Last refill timestamp
	mutex        sync.Mutex    // Mutex for thread safety
}

// NewTokenBucket returns a new TokenBucket that corresponds with the provided arguments.
func NewTokenBucket(capacity, refillRate int64, refillPeriod time.Duration) *TokenBucket {
	return &TokenBucket{
		capacity:     capacity,
		tokens:       capacity,
		refillRate:   refillRate,
		refillPeriod: refillPeriod,
		lastRefill:   time.Now(),
		mutex:        sync.Mutex{},
	}
}

// Allow checks if incoming request is valid.
// Returns true and remaining tokens if valid request
// else returns false and current tokens.
func (tb *TokenBucket) Allow(tokensRequest int64) (bool, int64) {
	// Locks the mutex for thread safety
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	// Finds how much time has passed since last refill
	elapsedTime := time.Since(tb.lastRefill)

	// Compute how many tokens to add based on elapsed time
	increments := elapsedTime / tb.refillPeriod
	newTokens := tb.refillRate * int64(increments)

	// Updates last refill timestamp if tokens were added
	if increments > 0 {
		tb.lastRefill = tb.lastRefill.Add(tb.refillPeriod * increments)
	}

	// Adds token
	if tb.tokens+newTokens > tb.capacity {
		tb.tokens = tb.capacity
	} else {
		tb.tokens += newTokens
	}

	if tb.tokens-tokensRequest < 0 {
		return false, tb.tokens
	} else {
		tb.tokens -= tokensRequest
		return true, tb.tokens
	}
}

// getCurrentTokens is a helper function for getting the current amount of tokens in the bucket
func (tb *TokenBucket) getCurrentTokens() int64 {
	return tb.tokens
}
