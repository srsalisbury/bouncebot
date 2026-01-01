// Package config provides configuration loading from environment variables.
package config

import (
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all server configuration.
type Config struct {
	// Server settings
	Port     int
	DataFile string

	// CORS/WebSocket allowed origins (comma-separated hostnames)
	// e.g., "localhost,myserver.com"
	// Each hostname allows both http://hostname and http://hostname:port
	AllowedOrigins []string

	// AllowSameHost allows requests where the Origin header's hostname
	// matches the server's Host header. This makes CORS work automatically
	// when frontend and backend are served from the same host.
	AllowSameHost bool

	// Room timing
	AutoSaveInterval      time.Duration
	CleanupInterval       time.Duration
	RoomMaxAge            time.Duration
	DisconnectGracePeriod time.Duration
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:                  8080,
		DataFile:              "rooms.json",
		AllowedOrigins:        []string{"localhost"},
		AllowSameHost:         true,
		AutoSaveInterval:      30 * time.Second,
		CleanupInterval:       1 * time.Hour,
		RoomMaxAge:            24 * time.Hour,
		DisconnectGracePeriod: 30 * time.Second,
	}
}

// LoadFromEnv loads configuration from environment variables.
// Environment variables override defaults. Supported variables:
//   - PORT: Server port (default: 8080)
//   - DATA_FILE: Path to room data file (default: rooms.json)
//   - ALLOWED_ORIGINS: Comma-separated allowed origins (default: localhost)
//   - ALLOW_SAME_HOST: Allow same-host requests (default: true)
//   - AUTO_SAVE_INTERVAL: Auto-save interval in seconds (default: 30)
//   - CLEANUP_INTERVAL: Cleanup interval in seconds (default: 3600)
//   - ROOM_MAX_AGE: Room max age in seconds (default: 86400)
//   - DISCONNECT_GRACE_PERIOD: Player disconnect grace period in seconds (default: 30)
func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}

	if v := os.Getenv("DATA_FILE"); v != "" {
		cfg.DataFile = v
	}

	if v := os.Getenv("ALLOWED_ORIGINS"); v != "" {
		origins := strings.Split(v, ",")
		cfg.AllowedOrigins = make([]string, 0, len(origins))
		for _, o := range origins {
			o = strings.TrimSpace(o)
			if o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	if v := os.Getenv("ALLOW_SAME_HOST"); v != "" {
		cfg.AllowSameHost = v == "true" || v == "1"
	}

	if v := os.Getenv("AUTO_SAVE_INTERVAL"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.AutoSaveInterval = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("CLEANUP_INTERVAL"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.CleanupInterval = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("ROOM_MAX_AGE"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.RoomMaxAge = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("DISCONNECT_GRACE_PERIOD"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.DisconnectGracePeriod = time.Duration(secs) * time.Second
		}
	}

	return cfg
}

// IsOriginAllowed checks if the given origin is allowed based on configured origins only.
func (c *Config) IsOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		// Check both http and https, with or without port
		for _, scheme := range []string{"http://", "https://"} {
			prefix := scheme + allowed
			if origin == prefix || strings.HasPrefix(origin, prefix+":") {
				return true
			}
		}
	}
	return false
}

// IsOriginAllowedForRequest checks if the given origin is allowed,
// considering both configured origins and same-host policy.
// requestHost is the Host header from the incoming request.
func (c *Config) IsOriginAllowedForRequest(origin, requestHost string) bool {
	// Check configured origins first
	if c.IsOriginAllowed(origin) {
		return true
	}

	// Check same-host policy
	if c.AllowSameHost {
		parsedOrigin, err := url.Parse(origin)
		if err != nil {
			return false
		}
		originHost := parsedOrigin.Hostname()

		// Strip port from request host for comparison
		parsedReq, err := url.Parse("http://" + requestHost)
		if err != nil {
			return false
		}
		reqHost := parsedReq.Hostname()

		if originHost == reqHost {
			return true
		}
	}

	return false
}
