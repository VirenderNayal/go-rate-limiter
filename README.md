# go-rate-limiter

[![Go Reference](https://pkg.go.dev/badge/github.com/virendernayal/go-rate-limiter.svg)](https://pkg.go.dev/github.com/virendernayal/go-rate-limiter/pkg/ratelimiter)
[![Go Report Card](https://goreportcard.com/badge/github.com/virendernayal/go-rate-limiter)](https://goreportcard.com/report/github.com/virendernayal/go-rate-limiter)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A flexible, production-ready rate limiting library for Go with pluggable storage backends and HTTP middleware support.

## Features

- **Three algorithms** — Fixed Window, Sliding Window, Token Bucket
- **Pluggable storage** — In-memory (single process) or Redis (distributed)
- **Zero-dependency core** — Redis is an opt-in sub-package; the core has no external dependencies
- **HTTP middleware** — Drop-in middleware for `net/http` with IP-based or custom key extraction
- **Thread-safe** — Mutex-protected in-memory stores, Lua-scripted atomic Redis operations
- **Custom store support** — Implement the store interface to bring your own backend

## Requirements

- Go 1.25+
- Redis 6+ (only if using the `redisstore` sub-package)

## Installation

```bash
go get github.com/virendernayal/go-rate-limiter
```

To also use the Redis-backed stores:

```bash
go get github.com/virendernayal/go-rate-limiter/pkg/ratelimiter/redisstore
```

## Quick Start

```go
package main

import (
	"net/http"

	"github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"
)

func main() {
	// Create a token bucket limiter: 10 requests per 60 seconds (in-memory)
	store := ratelimiter.NewMemoryTokenBucketStore()
	limiter := ratelimiter.NewTokenBucketLimiter(store, 10, 60)

	mux := http.NewServeMux()
	mux.Handle("GET /api/data", ratelimiter.MiddlewareWithIP(limiter)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		}),
	))

	http.ListenAndServe(":8080", mux)
}
```

Rate-limited requests receive a `429 Too Many Requests` response.

## Algorithms

| Algorithm | Constant | Best For | Trade-off |
|---|---|---|---|
| **Fixed Window** | `ratelimiter.FixedWindow` | Simple, low-overhead limiting | Can allow 2x burst at window boundaries |
| **Sliding Window** | `ratelimiter.SlidingWindow` | Accurate per-user limiting | Higher memory — stores individual timestamps |
| **Token Bucket** | `ratelimiter.TokenBucket` | Smooth rate with burst tolerance | Slightly more complex state management |

## Usage

### Import Paths

| Package | Import | Purpose |
|---|---|---|
| Core | `github.com/virendernayal/go-rate-limiter/pkg/ratelimiter` | Algorithms, in-memory stores, middleware |
| Redis | `github.com/virendernayal/go-rate-limiter/pkg/ratelimiter/redisstore` | Redis-backed stores (opt-in) |

### With In-Memory Store

Each algorithm has a corresponding memory store constructor:

```go
import "github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"

// Fixed Window — 100 requests per 60-second window
store := ratelimiter.NewMemoryFixedWindowStore()
limiter := ratelimiter.NewFixedWindowLimiter(store, 100, 60)

// Sliding Window — 100 requests in a rolling 60-second window
store := ratelimiter.NewMemorySlidingWindowStore()
limiter := ratelimiter.NewSlidingWindowLimiter(store, 100, 60)

// Token Bucket — capacity of 100, refills over 60 seconds
store := ratelimiter.NewMemoryTokenBucketStore()
limiter := ratelimiter.NewTokenBucketLimiter(store, 100, 60)
```

### With Redis Store (Distributed)

For multi-instance deployments, import the `redisstore` sub-package. The core `ratelimiter` package has **no Redis dependency** — you only pull it in when you need it.

```go
import (
	"github.com/redis/go-redis/v9"
	"github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"
	"github.com/virendernayal/go-rate-limiter/pkg/ratelimiter/redisstore"
)

client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

// Pick the Redis store matching your algorithm:
store := redisstore.NewTokenBucketStore(client)
limiter := ratelimiter.NewTokenBucketLimiter(store, 100, 60)
```

Available Redis stores:

| Constructor | Algorithm |
|---|---|
| `redisstore.NewFixedWindowStore(client)` | Fixed Window |
| `redisstore.NewSlidingWindowStore(client)` | Sliding Window |
| `redisstore.NewTokenBucketStore(client)` | Token Bucket |

### HTTP Middleware

**Rate limit by client IP** (recommended):

```go
handler := ratelimiter.MiddlewareWithIP(limiter)(yourHandler)
```

IP extraction order: `X-Forwarded-For` header first, then `RemoteAddr`.

**Rate limit by custom key** (e.g., API key, user ID, tenant):

```go
keyFunc := func(r *http.Request) string {
	return r.Header.Get("X-API-Key")
}
handler := ratelimiter.Middleware(limiter, keyFunc)(yourHandler)
```

### Direct Usage (Without Middleware)

The `Limiter` interface exposes a single method:

```go
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
```

Use it directly in any context — gRPC interceptors, queue consumers, CLI tools, etc:

```go
allowed, err := limiter.Allow(ctx, "user:123")
if err != nil {
	// handle error
}
if !allowed {
	// rate limited — back off or reject
}
```

### Custom Store Implementation

Implement the appropriate interface to plug in any backend (Memcached, DynamoDB, SQLite, etc.):

```go
// For Fixed Window algorithm
type FixedWindowStore interface {
	IncrementAndGet(ctx context.Context, key string, window, bucket int64) (int64, error)
}

// For Sliding Window algorithm
type SlidingWindowStore interface {
	AddAndCountTimestamps(ctx context.Context, key string, now, windowStart int64) (int64, error)
}

// For Token Bucket algorithm
type TokenBucketStore interface {
	AllowAndUpdate(ctx context.Context, key string, limit, window, now int64) (bool, error)
}
```

Then pass your store to the corresponding limiter constructor:

```go
myStore := NewMyCustomTokenBucketStore(/* ... */)
limiter := ratelimiter.NewTokenBucketLimiter(myStore, 100, 60)
```

## API Reference

Full documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/virendernayal/go-rate-limiter/pkg/ratelimiter).

### Constructors

| Function | Description |
|---|---|
| `NewFixedWindowLimiter(store, limit, window)` | Create a fixed window limiter |
| `NewSlidingWindowLimiter(store, limit, window)` | Create a sliding window limiter |
| `NewTokenBucketLimiter(store, limit, window)` | Create a token bucket limiter |
| `NewMemoryFixedWindowStore()` | In-memory store for fixed window |
| `NewMemorySlidingWindowStore()` | In-memory store for sliding window |
| `NewMemoryTokenBucketStore()` | In-memory store for token bucket |

### Middleware

| Function | Description |
|---|---|
| `MiddlewareWithIP(limiter)` | Rate limit by client IP address |
| `Middleware(limiter, keyFunc)` | Rate limit by custom key extraction function |

### Parameters

| Parameter | Type | Description |
|---|---|---|
| `limit` | `int64` | Maximum number of requests allowed in the window |
| `window` | `int64` | Time window duration in seconds |

## Project Structure

```
go-rate-limiter/
├── cmd/server/              # Example HTTP server
├── internal/
│   ├── apperrors/           # Error definitions (ErrRateLimited, ErrInvalidAlgorithm)
│   ├── limiter/             # Core algorithms & Limiter interface
│   ├── middleware/          # HTTP rate limit middleware & IP extraction
│   └── store/              # Memory & Redis store implementations
├── pkg/ratelimiter/         # Public API — import this
│   └── redisstore/          # Redis store constructors (opt-in dependency)
├── Makefile                 # Build, test, lint, coverage targets
├── LICENSE                  # MIT License
└── go.mod
```

## Development

### Running Tests

```bash
# All tests (skips Redis tests if Redis is unavailable)
make test

# Only Redis integration tests (requires Redis on localhost:6379)
make test-redis
```

### Other Targets

```bash
make build       # Compile all packages
make vet         # Run go vet
make lint        # Run golangci-lint (must be installed)
make cover       # Generate HTML coverage report
make clean       # Remove build artifacts
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/my-feature`)
3. Commit your changes (`git commit -m 'Add my feature'`)
4. Push to the branch (`git push origin feature/my-feature`)
5. Open a Pull Request

Please ensure `make vet` and `make test` pass before submitting.

## License

This project is licensed under the MIT License — see the [LICENSE](LICENSE) file for details.
