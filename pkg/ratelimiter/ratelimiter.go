package ratelimiter

import (
	"net/http"

	"github.com/virendernayal/go-rate-limiter/internal/limiter"
	"github.com/virendernayal/go-rate-limiter/internal/middleware"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

type Limiter = limiter.Limiter
type Algorithm = limiter.Algorithm

const (
	FixedWindow   = limiter.FixedWindow
	SlidingWindow = limiter.SlidingWindow
	TokenBucket   = limiter.TokenBucket
)

// stores — memory implementations, no Redis dependency
var NewMemoryFixedWindowStore = store.NewMemoryFixedWindowStore
var NewMemorySlidingWindowStore = store.NewMemorySlidingWindowStore
var NewMemoryTokenBucketStore = store.NewMemoryTokenBucketStore

// store interfaces — so callers can implement their own
type FixedWindowStore = store.FixedWindowStore
type SlidingWindowStore = store.SlidingWindowStore
type TokenBucketStore = store.TokenBucketStore

// constructors
var NewFixedWindowLimiter = limiter.NewFixedWindowLimiter
var NewSlidingWindowLimiter = limiter.NewSlidingWindowLimiter
var NewTokenBucketLimiter = limiter.NewTokenBucketLimiter

// middleware
// Middleware applies rate limiting using a custom key extractor
func Middleware(l Limiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(l, keyFunc)
}

// MiddlewareWithIP applies rate limiting keyed by client IP (recommended default)
func MiddlewareWithIP(l Limiter) func(http.Handler) http.Handler {
	return middleware.RateLimitMiddleware(l, middleware.IPKeyFunc)
}
