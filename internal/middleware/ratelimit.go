package middleware

import (
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/virendernayal/go-rate-limiter/internal/apperrors"
	"github.com/virendernayal/go-rate-limiter/internal/limiter"
)

func IPKeyFunc(r *http.Request) string {
	clientIP := r.Header.Get("X-Forwarded-For")

	if clientIP != "" {
		clientIP = strings.SplitN(clientIP, ",", 2)[0]
		clientIP = strings.TrimSpace(clientIP)
	}

	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	host, _, err := net.SplitHostPort(clientIP)
	if err == nil {
		clientIP = host
	}

	return clientIP
}

func RateLimitMiddleware(limiter limiter.Limiter, keyFunc func(*http.Request) string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)

			allowed, err := limiter.Allow(r.Context(), key)

			if err != nil {
				if errors.Is(err, apperrors.ErrRateLimited) {
					http.Error(w, "rate limited", http.StatusTooManyRequests)
					return
				}

				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			if !allowed {
				http.Error(w, "rate limited", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
