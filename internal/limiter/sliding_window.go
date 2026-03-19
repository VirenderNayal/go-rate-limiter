// /internal/limiter/sliding_window.go

package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

type SlidingWindowLimiter struct {
	store  store.SlidingWindowStore
	limit  int64
	window int64
}

func (l *SlidingWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {

	now := time.Now().Unix()
	windowStart := now - l.window

	count, err := l.store.AddAndCountTimestamps(ctx, key, now, windowStart)

	if err != nil {
		return false, fmt.Errorf("store error: %w", err)
	}

	if count > l.limit {
		return false, apperrors.ErrRateLimited
	}

	return true, nil
}

func NewSlidingWindowLimiter(store store.SlidingWindowStore, limit, window int64) *SlidingWindowLimiter {
	return &SlidingWindowLimiter{
		store, limit, window,
	}
}
