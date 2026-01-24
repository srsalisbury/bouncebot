// Package ratelimit provides IP-based rate limiting for HTTP endpoints.
package ratelimit

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// clientIPKey is the context key for the client IP address.
type clientIPKey struct{}

// ContextWithClientIP returns a new context with the client IP address.
func ContextWithClientIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, clientIPKey{}, ip)
}

// ClientIPFromContext returns the client IP address from the context.
// Returns an empty string if not set.
func ClientIPFromContext(ctx context.Context) string {
	ip, _ := ctx.Value(clientIPKey{}).(string)
	return ip
}

// InjectClientIPMiddleware wraps an http.Handler to inject the client IP into the context.
func InjectClientIPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := GetClientIP(r)
		ctx := ContextWithClientIP(r.Context(), ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Limiter implements a sliding window rate limiter per IP address.
type Limiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewLimiter creates a new rate limiter.
// limit specifies the maximum number of requests allowed per window.
// window specifies the time window for rate limiting.
func NewLimiter(limit int, window time.Duration) *Limiter {
	return &Limiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given IP should be allowed.
// Returns true if within rate limit, false if rate limited.
func (l *Limiter) Allow(ip string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	// Get existing requests for this IP
	times := l.requests[ip]

	// Filter out requests outside the window
	var validTimes []time.Time
	for _, t := range times {
		if t.After(windowStart) {
			validTimes = append(validTimes, t)
		}
	}

	// Check if under limit
	if len(validTimes) >= l.limit {
		l.requests[ip] = validTimes
		return false
	}

	// Add new request and allow
	l.requests[ip] = append(validTimes, now)
	return true
}

// Cleanup removes stale entries from the limiter.
// Should be called periodically to prevent memory growth.
func (l *Limiter) Cleanup() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-l.window)

	for ip, times := range l.requests {
		var validTimes []time.Time
		for _, t := range times {
			if t.After(windowStart) {
				validTimes = append(validTimes, t)
			}
		}
		if len(validTimes) == 0 {
			delete(l.requests, ip)
		} else {
			l.requests[ip] = validTimes
		}
	}
}

// StartCleanup starts a background goroutine that periodically cleans up stale entries.
// Returns a channel that should be closed to stop the cleanup goroutine.
func (l *Limiter) StartCleanup(interval time.Duration) chan struct{} {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				l.Cleanup()
			case <-stop:
				return
			}
		}
	}()
	return stop
}

// GetClientIP extracts the client IP address from an HTTP request.
// It checks X-Forwarded-For and X-Real-IP headers for proxied requests.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For first (may contain multiple IPs, take the first)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can be comma-separated list of IPs
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr (strip port if present)
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
