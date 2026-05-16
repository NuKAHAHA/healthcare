package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// LoginLimiter is a Redis-backed brute-force guard for the login endpoint.
// It tracks by (IP + email) so switching IP isn't sufficient to bypass it.
type LoginLimiter struct {
	limiter     *redis_rate.Limiter
	maxAttempts int
	windowMin   int
	blockMin    int
}

func NewLoginLimiter(rdb *redis.Client, maxAttempts, windowMin, blockMin int) *LoginLimiter {
	return &LoginLimiter{
		limiter:     redis_rate.NewLimiter(rdb),
		maxAttempts: maxAttempts,
		windowMin:   windowMin,
		blockMin:    blockMin,
	}
}

// LoginAttemptResult carries the outcome of a rate-limit check.
type LoginAttemptResult struct {
	Allowed    bool
	Remaining  int64         // attempts left in the current window (after this one)
	RetryAfter time.Duration // how long until the next attempt is allowed
}

// Allow checks the quota and decrements the counter.
func (l *LoginLimiter) Allow(ctx context.Context, ip, email string) LoginAttemptResult {
	key := "login:" + ip + "|" + strings.ToLower(strings.TrimSpace(email))
	limit := redis_rate.Limit{
		Rate:   l.maxAttempts,
		Burst:  l.maxAttempts,
		Period: time.Duration(l.windowMin) * time.Minute,
	}
	res, err := l.limiter.Allow(ctx, key, limit)
	if err != nil {
		// Redis unavailable — fail open
		return LoginAttemptResult{Allowed: true, Remaining: int64(l.maxAttempts)}
	}
	return LoginAttemptResult{
		Allowed:    res.Allowed > 0,
		Remaining:  int64(res.Remaining),
		RetryAfter: res.RetryAfter,
	}
}

// Reset clears the counter on successful login.
func (l *LoginLimiter) Reset(ctx context.Context, ip, email string) {
	key := "login:" + ip + "|" + strings.ToLower(strings.TrimSpace(email))
	_ = l.limiter.Reset(ctx, key)
}

// GlobalRateLimit is a middleware that applies a per-IP request rate limit
// to every route. It uses a sliding window algorithm via Redis.
func GlobalRateLimit(rdb *redis.Client, requestsPerMinute int) func(http.Handler) http.Handler {
	limiter := redis_rate.NewLimiter(rdb)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := ClientIP(r)
			key := fmt.Sprintf("global:%s", ip)
			limit := redis_rate.PerMinute(requestsPerMinute)

			res, err := limiter.Allow(r.Context(), key, limit)
			if err != nil {
				// Redis unavailable — fail open
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit.Rate))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", res.Remaining))

			if res.Allowed == 0 {
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", res.RetryAfter.Seconds()))
				sendError(w, http.StatusTooManyRequests, "rate_limited", "Too many requests. Please slow down.")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP extracts the real client IP, respecting X-Forwarded-For from trusted proxies.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the last IP added by the trusted proxy (rightmost)
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	// Fall back to remote addr
	host := r.RemoteAddr
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}
