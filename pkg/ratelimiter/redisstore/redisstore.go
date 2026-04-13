// Package redisstore provides Redis-backed store implementations for go-rate-limiter.
//
// Import this package only if you need Redis support. The core ratelimiter
// package has no Redis dependency.
//
//	import "github.com/virendernayal/go-rate-limiter/pkg/ratelimiter/redisstore"
package redisstore

import (
	"github.com/redis/go-redis/v9"
	"github.com/virendernayal/go-rate-limiter/internal/store"
	"github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"
)

// NewFixedWindowStore creates a Redis-backed fixed window store.
func NewFixedWindowStore(client *redis.Client) ratelimiter.FixedWindowStore {
	return store.NewRedisFixedWindowStore(client)
}

// NewSlidingWindowStore creates a Redis-backed sliding window store.
func NewSlidingWindowStore(client *redis.Client) ratelimiter.SlidingWindowStore {
	return store.NewRedisSlidingWindowStore(client)
}

// NewTokenBucketStore creates a Redis-backed token bucket store.
func NewTokenBucketStore(client *redis.Client) ratelimiter.TokenBucketStore {
	return store.NewRedisTokenBucketStore(client)
}
