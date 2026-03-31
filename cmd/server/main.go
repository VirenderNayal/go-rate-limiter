package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/virendernayal/go-rate-limiter/internal/limiter"
	"github.com/virendernayal/go-rate-limiter/internal/middleware"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

func main() {
	ctx := context.Background()

	// REDIS connection
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Redis connection failed: %v", err)
	}

	fmt.Println("Connected to Redis!")

	// LIMITER
	rateLimiter, _ := limiter.NewLimiter(&limiter.Config{Limit: 3, Window: 1, Algorithm: limiter.SlidingWindow})

	// SERVERSETUP

	mux := http.NewServeMux()

	// Public route (no middleware)
	mux.HandleFunc("GET /health", healthHandler)

	// Protected route (with middleware)
	mux.Handle("GET /ping", middleware.RateLimitMiddleware(rateLimiter, middleware.IPKeyFunc)(http.HandlerFunc(pingHandler)))

	server := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	// start server in goroutine
	go func() {
		log.Println("server running on port : 8080")

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// wait for signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped cleanly")
}
