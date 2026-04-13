package ratelimiter_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/virendernayal/go-rate-limiter/pkg/ratelimiter"
)

func ExampleNewTokenBucketLimiter() {
	store := ratelimiter.NewMemoryTokenBucketStore()
	limiter := ratelimiter.NewTokenBucketLimiter(store, 5, 60) // 5 req per 60s

	allowed, err := limiter.Allow(context.Background(), "user:1")
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("allowed:", allowed)
	// Output: allowed: true
}

func ExampleMiddlewareWithIP() {
	store := ratelimiter.NewMemoryFixedWindowStore()
	limiter := ratelimiter.NewFixedWindowLimiter(store, 100, 60)

	handler := ratelimiter.MiddlewareWithIP(limiter)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("OK"))
		}),
	)

	// Use handler with http.ListenAndServe or in tests.
	_ = handler
}
