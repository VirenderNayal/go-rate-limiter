// /internal/limiter/limiter_redis_test.go
package limiter

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/store"
)

func setupRedis(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1,
	})

	ctx := context.Background()

	// ensure redis is reachable
	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("redis not running: %v", err)
	}

	// clean db before test
	if err := client.FlushDB(ctx).Err(); err != nil {
		t.Fatalf("failed to flush redis: %v", err)
	}

	return client
}

func TestRedisLimiter_BasicAllowThenDeny(t *testing.T) {
	client := setupRedis(t)

	limit := int64(10)
	ctx := context.Background()

	limiters := []Limiter{
		NewFixedWindowLimiter(store.NewRedisFixedWindowStore(client), limit, 60),
		NewSlidingWindowLimiter(store.NewRedisSlidingWindowStore(client), limit, 60),
		NewTokenBucketLimiter(store.NewRedisTokenBucketStore(client), limit, 60),
	}

	keys := []string{"fw-user1", "sw-user1", "tb-user1"}

	for idx, limiter := range limiters {
		for i := 0; i < int(limit); i++ {
			allowed, err := limiter.Allow(ctx, keys[idx])
			if !allowed || err != nil {
				t.Fatalf("expected request %d to pass", i+1)
			}
		}

		allowed, err := limiter.Allow(ctx, keys[idx])

		if allowed {
			t.Fatal("expected request to be denied")
		}

		if !errors.Is(err, apperrors.ErrRateLimited) {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	}
}

func TestRedisLimiter_MultiUserIsolation(t *testing.T) {
	client := setupRedis(t)

	limiter := NewFixedWindowLimiter(
		store.NewRedisFixedWindowStore(client),
		5,
		60,
	)

	ctx := context.Background()

	for i := 0; i < 5; i++ {
		allowed, _ := limiter.Allow(ctx, "user1")
		if !allowed {
			t.Fatal("user1 should be allowed")
		}
	}

	// user2 unaffected
	allowed, _ := limiter.Allow(ctx, "user2")
	if !allowed {
		t.Fatal("user2 should not be affected")
	}
}

func TestRedisFixedWindow_Reset(t *testing.T) {
	client := setupRedis(t)

	limiter := NewFixedWindowLimiter(
		store.NewRedisFixedWindowStore(client),
		2,
		1, // 1 second window
	)

	ctx := context.Background()
	key := "user1"

	limiter.Allow(ctx, key)
	limiter.Allow(ctx, key)

	allowed, _ := limiter.Allow(ctx, key)
	if allowed {
		t.Fatal("should be limited")
	}

	time.Sleep(2 * time.Second)

	allowed, _ = limiter.Allow(ctx, key)
	if !allowed {
		t.Fatal("should allow after reset")
	}
}

func TestRedisTokenBucket_Refill(t *testing.T) {
	client := setupRedis(t)

	limiter := NewTokenBucketLimiter(
		store.NewRedisTokenBucketStore(client),
		5,
		5, // 1 token/sec
	)

	ctx := context.Background()
	key := "user1"

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

func TestRedisLimiter_Concurrency(t *testing.T) {
	client := setupRedis(t)

	limiter := NewFixedWindowLimiter(
		store.NewRedisFixedWindowStore(client),
		100,
		60,
	)

	ctx := context.Background()
	key := "user1"

	var wg sync.WaitGroup
	var mu sync.Mutex

	success := 0

	for i := 0; i < 200; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

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
		}()
	}

	wg.Wait()

	if success > 100 {
		t.Fatalf("allowed more than limit: %d", success)
	}
}
