// Package ratelimiter provides a pluggable rate limiting library with multiple
// algorithms (fixed window, sliding window, token bucket) and storage backends.
//
// The core package includes in-memory stores and has no external dependencies.
// For Redis-backed distributed rate limiting, import the [redisstore] sub-package.
//
// Basic usage:
//
//	store := ratelimiter.NewMemoryTokenBucketStore()
//	limiter := ratelimiter.NewTokenBucketLimiter(store, 100, 60) // 100 req per 60s
//	handler := ratelimiter.MiddlewareWithIP(limiter)(myHandler)
package ratelimiter

import (
	"net/http"

	"github.com/virendernayal/go-rate-limiter/internal/limiter"
	"github.com/virendernayal/go-rate-limiter/internal/middleware"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

// Limiter is the core interface. Call Allow with a key to check if a request
// should be permitted.
type Limiter = limiter.Limiter

// Algorithm selects the rate limiting strategy.
type Algorithm = limiter.Algorithm

const (
	// FixedWindow divides time into fixed intervals; counter resets at each boundary.
	FixedWindow = limiter.FixedWindow
	// SlidingWindow tracks request timestamps in a rolling window for smoother limiting.
	SlidingWindow = limiter.SlidingWindow
	// TokenBucket refills tokens at a steady rate, allowing controlled bursts.
	TokenBucket = limiter.TokenBucket
)

// NewMemoryFixedWindowStore creates an in-memory store for the fixed window algorithm.
var NewMemoryFixedWindowStore = store.NewMemoryFixedWindowStore

// NewMemorySlidingWindowStore creates an in-memory store for the sliding window algorithm.
var NewMemorySlidingWindowStore = store.NewMemorySlidingWindowStore

// NewMemoryTokenBucketStore creates an in-memory store for the token bucket algorithm.
var NewMemoryTokenBucketStore = store.NewMemoryTokenBucketStore

// FixedWindowStore is the storage interface for the fixed window algorithm.
// Implement this to provide a custom backend.
type FixedWindowStore = store.FixedWindowStore

// SlidingWindowStore is the storage interface for the sliding window algorithm.
// Implement this to provide a custom backend.
type SlidingWindowStore = store.SlidingWindowStore

// TokenBucketStore is the storage interface for the token bucket algorithm.
// Implement this to provide a custom backend.
type TokenBucketStore = store.TokenBucketStore

// NewFixedWindowLimiter creates a limiter using the fixed window algorithm.
// limit is the max requests allowed; window is the time window in seconds.
var NewFixedWindowLimiter = limiter.NewFixedWindowLimiter

// NewSlidingWindowLimiter creates a limiter using the sliding window algorithm.
// limit is the max requests allowed; window is the time window in seconds.
var NewSlidingWindowLimiter = limiter.NewSlidingWindowLimiter

// NewTokenBucketLimiter creates a limiter using the token bucket algorithm.
// limit is the bucket capacity; window is the refill period in seconds.
var NewTokenBucketLimiter = limiter.NewTokenBucketLimiter

// Middleware returns HTTP middleware that rate-limits requests using the given
// key extraction function. Requests that exceed the limit receive a 429 response.
func Middleware(l Limiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(l, keyFunc)
}

// MiddlewareWithIP returns HTTP middleware that rate-limits by client IP address.
// It reads X-Forwarded-For first, falling back to RemoteAddr.
func MiddlewareWithIP(l Limiter) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(l, middleware.IPKeyFunc)
}
