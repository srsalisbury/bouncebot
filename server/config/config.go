// Package config provides configuration loading from environment variables.
package config

import (
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
	// e.g., "localhost,guido.local,myserver.com"
	// Each hostname allows both http://hostname and http://hostname:port
	AllowedOrigins []string

	// Session timing
	AutoSaveInterval   time.Duration
	CleanupInterval    time.Duration
	SessionMaxAge      time.Duration
	DisconnectGracePeriod time.Duration
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:                  8080,
		DataFile:              "sessions.json",
		AllowedOrigins:        []string{"localhost"},
		AutoSaveInterval:      30 * time.Second,
		CleanupInterval:       1 * time.Hour,
		SessionMaxAge:         24 * time.Hour,
		DisconnectGracePeriod: 30 * time.Second,
	}
}

// LoadFromEnv loads configuration from environment variables.
// Environment variables override defaults. Supported variables:
//   - PORT: Server port (default: 8080)
//   - DATA_FILE: Path to session data file (default: sessions.json)
//   - ALLOWED_ORIGINS: Comma-separated allowed origins (default: localhost)
//   - AUTO_SAVE_INTERVAL: Auto-save interval in seconds (default: 30)
//   - CLEANUP_INTERVAL: Cleanup interval in seconds (default: 3600)
//   - SESSION_MAX_AGE: Session max age in seconds (default: 86400)
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

	if v := os.Getenv("SESSION_MAX_AGE"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.SessionMaxAge = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("DISCONNECT_GRACE_PERIOD"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.DisconnectGracePeriod = time.Duration(secs) * time.Second
		}
	}

	return cfg
}

// IsOriginAllowed checks if the given origin is allowed.
func (c *Config) IsOriginAllowed(origin string) bool {
	for _, allowed := range c.AllowedOrigins {
		// Check exact match or with port
		prefix := "http://" + allowed
		if origin == prefix || strings.HasPrefix(origin, prefix+":") {
			return true
		}
	}
	return false
}
