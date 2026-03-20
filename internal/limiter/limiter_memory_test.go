package limiter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

func TestLimiter_BasicAllowThenDeny(t *testing.T) {
	limit := int64(10)

	limiters := []Limiter{
		NewFixedWindowLimiter(store.NewMemoryFixedWindowStore(), limit, 60),
		NewSlidingWindowLimiter(store.NewMemorySlidingWindowStore(), limit, 60),
		NewTokenBucketLimiter(store.NewMemoryTokenBucketStore(), limit, 60),
	}

	key := "user1"
	ctx := context.Background()

	for idx, limiter := range limiters {
		var limiterName string
		switch idx {
		case 0:
			limiterName = "Fixed Window Limiter"
		case 1:
			limiterName = "Sliding Window Limiter"
		case 2:
			limiterName = "Token Bucker Limiter"
		}

		for i := 0; i < int(limit); i++ {
			allowed, err := limiter.Allow(ctx, key)
			if !allowed || err != nil {
				t.Fatalf("%s : expected request %d to be allowed", limiterName, i+1)
			}
		}

		allowed, err := limiter.Allow(ctx, key)

		if allowed {
			t.Fatalf("expected request to be denied")
		}

		if !errors.Is(err, apperrors.ErrRateLimited) {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	}
}

func TestLimiter_DefaultConfig(t *testing.T) {
	limiter, err := NewLimiter(nil)
	if err != nil {
		t.Fatal(err)
	}

	allowed, _ := limiter.Allow(context.Background(), "user1")
	if !allowed {
		t.Fatal("expected first request to pass with default config")
	}
}

func TestLimiter_MultiUserIsolation(t *testing.T) {
	limiter := NewFixedWindowLimiter(store.NewMemoryFixedWindowStore(), 5, 60)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		allowed, _ := limiter.Allow(ctx, "user1")
		if !allowed {
			t.Fatal("user1 should be allowed")
		}
	}

	// user2 should still work
	allowed, _ := limiter.Allow(ctx, "user2")
	if !allowed {
		t.Fatal("user2 should not be affected")
	}
}

func TestFixedWindow_Reset(t *testing.T) {
	limiter := NewFixedWindowLimiter(store.NewMemoryFixedWindowStore(), 2, 1) // 1 sec window
	ctx := context.Background()

	key := "user1"

	// exhaust
	limiter.Allow(ctx, key)
	limiter.Allow(ctx, key)

	allowed, _ := limiter.Allow(ctx, key)
	if allowed {
		t.Fatal("should be limited")
	}

	time.Sleep(2 * time.Second)

	// should reset
	allowed, _ = limiter.Allow(ctx, key)
	if !allowed {
		t.Fatal("should allow after window reset")
	}
}

func TestTokenBucket_Refill(t *testing.T) {
	limiter := NewTokenBucketLimiter(store.NewMemoryTokenBucketStore(), 5, 5) // 1 token/sec
	ctx := context.Background()

	key := "user1"

	// exhaust
	for i := 0; i < 5; i++ {
		limiter.Allow(ctx, key)
	}

	allowed, err := limiter.Allow(ctx, key)
	if allowed {
		t.Fatal("should be rate limited")
	}

	if !errors.Is(err, apperrors.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}

	time.Sleep(2 * time.Second)

	allowed, _ = limiter.Allow(ctx, key)
	if !allowed {
		t.Fatal("should allow after refill")
	}
}

func TestLimiter_Concurrency(t *testing.T) {
	limiter := NewFixedWindowLimiter(store.NewMemoryFixedWindowStore(), 100, 60)

	ctx := context.Background()
	key := "user1"

	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0

	for range 200 {
		wg.Go(func() {

			allowed, err := limiter.Allow(ctx, key)

			if allowed {
				mu.Lock()
				success++
				mu.Unlock()
			} else {
				if !errors.Is(err, apperrors.ErrRateLimited) {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}

	wg.Wait()

	if success > 100 {
		t.Fatalf("allowed more than limit: %d", success)
	}
}

func TestTokenBucket_ConcurrencyStress(t *testing.T) {
	limit := int64(1000)
	limiter := NewTokenBucketLimiter(store.NewMemoryTokenBucketStore(), limit, 60)

	ctx := context.Background()
	key := "user1"

	var wg sync.WaitGroup
	var mu sync.Mutex

	success := 0

	for i := 0; i < 2000; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			allowed, _ := limiter.Allow(ctx, key)

			if allowed {
				mu.Lock()
				success++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if success == int(limit)+1 {
		t.Fatalf("allowed too many requests: %d", success)
	}
}
