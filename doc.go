// Package go-rate-limiter provides a pluggable rate limiting library for Go.
//
// This module offers three rate limiting algorithms (fixed window, sliding window,
// token bucket) with in-memory and Redis storage backends, plus drop-in HTTP middleware.
//
// Import the public API:
//
//	import "github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"
//
// For Redis support, also import:
//
//	import "github.com/virendernayal/go-rate-limiter/pkg/ratelimiter/redisstore"
//
// See https://github.com/virendernayal/go-rate-limiter for full documentation.
package goratelimiter
