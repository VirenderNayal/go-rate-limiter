// /internal/store/store.go

package store

import "context"

type FixedWindowStore interface {
	IncrementAndGet(ctx context.Context, key string, window int64, bucket int64) (int64, error)
}

type SlidingWindowStore interface {
	AddAndCountTimestamps(ctx context.Context, key string, now int64, windowStart int64) (int64, error)
}

type TokenBucketStore interface {
	AllowAndUpdate(ctx context.Context, key string, limit int64, window int64, now int64) (bool, error)
}
