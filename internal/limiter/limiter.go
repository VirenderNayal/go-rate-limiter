// /internal/limiter/limiter.go

package limiter

import (
	"context"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type Algorithm string

const (
	FixedWindow   Algorithm = "fixed-window"
	SlidingWindow Algorithm = "sliding-window"
	TokenBucket   Algorithm = "token-bucket"
)

type Config struct {
	Algorithm Algorithm
	Limit     int64
	Window    int64
}

type Option func(*Config)

func WithAlgorithm(algo Algorithm) Option {
	return func(c *Config) {
		c.Algorithm = algo
	}
}

func WithLimit(limit int64) Option {
	return func(c *Config) {
		c.Limit = limit
	}
}

func WithWindow(window int64) Option {
	return func(c *Config) {
		c.Window = window
	}
}

func NewLimiter(cfg *Config) (Limiter, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.Limit == 0 {
		cfg.Limit = 100
	}
	if cfg.Window == 0 {
		cfg.Window = 60
	}
	if cfg.Algorithm == "" {
		cfg.Algorithm = FixedWindow
	}

	switch cfg.Algorithm {
	case FixedWindow:
		return NewFixedWindowLimiter(store.NewMemoryFixedWindowStore(), cfg.Limit, cfg.Window), nil
	case SlidingWindow:
		return NewSlidingWindowLimiter(store.NewMemorySlidingWindowStore(), cfg.Limit, cfg.Window), nil
	case TokenBucket:
		return NewTokenBucketLimiter(store.NewMemoryTokenBucketStore(), cfg.Limit, cfg.Window), nil
	default:
		return nil, apperrors.ErrInvalidAlgorithm
	}
}

func NewLimiterWithOptions(opts ...Option) (Limiter, error) {
	cfg := &Config{
		Algorithm: FixedWindow,
		Limit:     100,
		Window:    60,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return NewLimiter(cfg)
}
