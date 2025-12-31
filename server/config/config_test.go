package config

import "testing"

func TestIsOriginAllowed(t *testing.T) {
	cfg := &Config{
		AllowedOrigins: []string{"localhost", "example.com"},
	}

	tests := []struct {
		origin string
		want   bool
	}{
		// localhost variations
		{"http://localhost", true},
		{"https://localhost", true},
		{"http://localhost:8080", true},
		{"http://localhost:3000", true},

		// example.com variations
		{"http://example.com", true},
		{"https://example.com", true},
		{"http://example.com:8080", true},

		// Not allowed
		{"http://other.com", false},
		{"http://notlocalhost", false},
		{"http://localhost.evil.com", false},
		{"http://example.com.evil.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			got := cfg.IsOriginAllowed(tt.origin)
			if got != tt.want {
				t.Errorf("IsOriginAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

func TestIsOriginAllowedForRequest_ConfiguredOrigins(t *testing.T) {
	cfg := &Config{
		AllowedOrigins: []string{"localhost"},
		AllowSameHost:  false,
	}

	tests := []struct {
		origin      string
		requestHost string
		want        bool
	}{
		// Configured origin should work regardless of request host
		{"http://localhost:3000", "other.com:8080", true},
		{"http://localhost", "other.com", true},

		// Non-configured origin should not work when AllowSameHost is false
		{"http://other.com:3000", "other.com:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin+"_"+tt.requestHost, func(t *testing.T) {
			got := cfg.IsOriginAllowedForRequest(tt.origin, tt.requestHost)
			if got != tt.want {
				t.Errorf("IsOriginAllowedForRequest(%q, %q) = %v, want %v",
					tt.origin, tt.requestHost, got, tt.want)
			}
		})
	}
}

func TestIsOriginAllowedForRequest_SameHost(t *testing.T) {
	cfg := &Config{
		AllowedOrigins: []string{},
		AllowSameHost:  true,
	}

	tests := []struct {
		name        string
		origin      string
		requestHost string
		want        bool
	}{
		{
			name:        "same host different ports",
			origin:      "http://myserver.com:3000",
			requestHost: "myserver.com:8080",
			want:        true,
		},
		{
			name:        "same host no port in origin",
			origin:      "http://myserver.com",
			requestHost: "myserver.com:8080",
			want:        true,
		},
		{
			name:        "same host no port in request",
			origin:      "http://myserver.com:3000",
			requestHost: "myserver.com",
			want:        true,
		},
		{
			name:        "same host https",
			origin:      "https://myserver.com:3000",
			requestHost: "myserver.com:8080",
			want:        true,
		},
		{
			name:        "different hosts",
			origin:      "http://frontend.com:3000",
			requestHost: "backend.com:8080",
			want:        false,
		},
		{
			name:        "subdomain mismatch",
			origin:      "http://app.myserver.com:3000",
			requestHost: "myserver.com:8080",
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.IsOriginAllowedForRequest(tt.origin, tt.requestHost)
			if got != tt.want {
				t.Errorf("IsOriginAllowedForRequest(%q, %q) = %v, want %v",
					tt.origin, tt.requestHost, got, tt.want)
			}
		})
	}
}

func TestIsOriginAllowedForRequest_SameHostDisabled(t *testing.T) {
	cfg := &Config{
		AllowedOrigins: []string{},
		AllowSameHost:  false,
	}

	// Same host should not be allowed when AllowSameHost is false
	got := cfg.IsOriginAllowedForRequest("http://myserver.com:3000", "myserver.com:8080")
	if got != false {
		t.Errorf("expected same-host to be rejected when AllowSameHost=false")
	}
}

func TestIsOriginAllowedForRequest_InvalidOrigin(t *testing.T) {
	cfg := &Config{
		AllowedOrigins: []string{},
		AllowSameHost:  true,
	}

	// Invalid origin should be rejected
	got := cfg.IsOriginAllowedForRequest("not-a-valid-url", "myserver.com:8080")
	if got != false {
		t.Errorf("expected invalid origin to be rejected")
	}
}
