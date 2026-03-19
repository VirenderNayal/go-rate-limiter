// /internal/limiter/token_bucket.go

package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

type TokenBucketLimiter struct {
	store  store.TokenBucketStore
	limit  int64
	window int64
}

func (l *TokenBucketLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()

	allowed, err := l.store.AllowAndUpdate(ctx, key, l.limit, l.window, now)

	if err != nil {
		return false, fmt.Errorf("store error: %w", err)
	}

	if !allowed {
		return false, apperrors.ErrRateLimited
	}

	return true, nil
}

func NewTokenBucketLimiter(store store.TokenBucketStore, limit, window int64) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		store, limit, window,
	}
}
