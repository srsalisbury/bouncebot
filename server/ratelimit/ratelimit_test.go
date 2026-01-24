package ratelimit

import (
	"net/http"
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	limiter := NewLimiter(3, time.Minute)

	// First 3 requests should be allowed
	for i := 0; i < 3; i++ {
		if !limiter.Allow("192.168.1.1") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4th request should be rate limited
	if limiter.Allow("192.168.1.1") {
		t.Error("4th request should be rate limited")
	}

	// Different IP should still be allowed
	if !limiter.Allow("192.168.1.2") {
		t.Error("Request from different IP should be allowed")
	}
}

func TestLimiter_WindowExpiry(t *testing.T) {
	// Use a short window for testing
	limiter := NewLimiter(2, 50*time.Millisecond)

	// Exhaust the limit
	if !limiter.Allow("192.168.1.1") {
		t.Error("First request should be allowed")
	}
	if !limiter.Allow("192.168.1.1") {
		t.Error("Second request should be allowed")
	}
	if limiter.Allow("192.168.1.1") {
		t.Error("Third request should be rate limited")
	}

	// Wait for window to expire
	time.Sleep(60 * time.Millisecond)

	// Should be allowed again
	if !limiter.Allow("192.168.1.1") {
		t.Error("Request should be allowed after window expires")
	}
}

func TestLimiter_Cleanup(t *testing.T) {
	limiter := NewLimiter(2, 50*time.Millisecond)

	limiter.Allow("192.168.1.1")
	limiter.Allow("192.168.1.2")

	// Verify entries exist
	limiter.mu.Lock()
	if len(limiter.requests) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(limiter.requests))
	}
	limiter.mu.Unlock()

	// Wait for window to expire and cleanup
	time.Sleep(60 * time.Millisecond)
	limiter.Cleanup()

	// Entries should be cleaned up
	limiter.mu.Lock()
	if len(limiter.requests) != 0 {
		t.Errorf("Expected 0 entries after cleanup, got %d", len(limiter.requests))
	}
	limiter.mu.Unlock()
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "X-Forwarded-For single IP",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Forwarded-For multiple IPs",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50, 70.41.3.18, 150.172.238.178"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.50",
		},
		{
			name:       "X-Real-IP",
			headers:    map[string]string{"X-Real-IP": "203.0.113.51"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.51",
		},
		{
			name:       "X-Forwarded-For takes precedence over X-Real-IP",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.50", "X-Real-IP": "203.0.113.51"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.50",
		},
		{
			name:       "Fall back to RemoteAddr",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1:12345",
			expected:   "192.168.1.1",
		},
		{
			name:       "RemoteAddr without port",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.1",
			expected:   "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				RemoteAddr: tt.remoteAddr,
				Header:     make(http.Header),
			}
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			got := GetClientIP(req)
			if got != tt.expected {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}
