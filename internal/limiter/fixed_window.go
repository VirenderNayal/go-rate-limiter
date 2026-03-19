// /internal/limiter/fixed_window.go

package limiter

import (
	"context"
	"fmt"
	"time"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

type FixedWindowLimiter struct {
	store  store.FixedWindowStore
	limit  int64
	window int64
}

func (l *FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().Unix()
	bucket := now / l.window

	count, err := l.store.IncrementAndGet(ctx, key, l.window, bucket)

	if err != nil {
		return false, fmt.Errorf("store error: %w", err)
	}

	if count > l.limit {
		return false, apperrors.ErrRateLimited
	}

	return true, nil
}

func NewFixedWindowLimiter(store store.FixedWindowStore, limit, window int64) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		store, limit, window,
	}
}
