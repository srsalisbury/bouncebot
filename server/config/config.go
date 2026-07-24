// Package config provides configuration loading from environment variables.
package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all server configuration.
type Config struct {
	// Server settings
	Port    int
	DataDir string // Base directory for all data files (rooms, daily puzzles, user progress)

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
	DisconnectGracePeriod     time.Duration
	SoloDisconnectGracePeriod time.Duration

	// Solver settings
	SolverTimeout time.Duration

	// Feature flags
	EnableDailyChallenge bool

	// PublicClientURL is the public base URL of the client SPA (e.g.
	// "https://bouncebot.example.com"). Used to build the redirect target
	// for join-room preview links, since the server and client don't share
	// an origin.
	PublicClientURL string

	// PublicServerURL is this server's own public base URL, INCLUDING any
	// path prefix a reverse proxy strips before forwarding (e.g.
	// "https://bouncebot.example.com/beta/api"). Used to build the og:image
	// URL for join-room preview links. Can't be derived from the incoming
	// request's Host header alone: a stripPrefix-style proxy removes the
	// prefix before the server ever sees the request, so from the server's
	// perspective that prefix doesn't exist.
	PublicServerURL string
}

// DefaultConfig returns configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Port:    8080,
		DataDir: "data",
		AllowedOrigins:        []string{"localhost"},
		AllowSameHost:         true,
		AutoSaveInterval:      30 * time.Second,
		CleanupInterval:       1 * time.Hour,
		RoomMaxAge:            24 * time.Hour,
		DisconnectGracePeriod:     5 * time.Minute,
		SoloDisconnectGracePeriod: 30 * time.Minute,
		SolverTimeout:         30 * time.Second,
		PublicClientURL:       "http://localhost:5173",
		PublicServerURL:       "http://localhost:8080",
	}
}

// RoomsFile returns the path to rooms.json within the data directory.
func (c *Config) RoomsFile() string {
	return filepath.Join(c.DataDir, "rooms.json")
}

// LoadFromEnv loads configuration from environment variables.
// It first loads .env.local then .env (if they exist) to populate env vars.
// Real env vars always take priority since godotenv won't overwrite them.
// Supported variables:
//   - PORT: Server port (default: 8080)
//   - DATA_DIR: Base directory for all data files (default: data)
//   - ALLOWED_ORIGINS: Comma-separated allowed origins (default: localhost)
//   - ALLOW_SAME_HOST: Allow same-host requests (default: true)
//   - AUTO_SAVE_INTERVAL: Auto-save interval in seconds (default: 30)
//   - CLEANUP_INTERVAL: Cleanup interval in seconds (default: 3600)
//   - ROOM_MAX_AGE: Room max age in seconds (default: 86400)
//   - DISCONNECT_GRACE_PERIOD: Player disconnect grace period in seconds (default: 300)
//   - SOLO_DISCONNECT_GRACE_PERIOD: Solo mode disconnect grace period in seconds (default: 1800)
//   - SOLVER_TIMEOUT: Solver timeout in seconds (default: 30)
//   - ENABLE_DAILY_CHALLENGE: Enable daily challenge feature (default: false)
//   - PUBLIC_CLIENT_URL: Public base URL of the client SPA (default: http://localhost:5173)
//   - PUBLIC_SERVER_URL: This server's own public base URL, including any reverse-proxy path prefix (default: http://localhost:8080)
func LoadFromEnv() *Config {
	// Load .env.local (personal overrides, gitignored) then .env (checked-in defaults).
	// Each call is separate so a missing .env.local doesn't prevent loading .env.
	// godotenv never overwrites existing env vars, so .env.local values take priority.
	_ = godotenv.Load(".env.local")
	_ = godotenv.Load(".env")

	cfg := DefaultConfig()

	if v := os.Getenv("PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			cfg.Port = port
		}
	}

	if v := os.Getenv("DATA_DIR"); v != "" {
		cfg.DataDir = v
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

	if v := os.Getenv("SOLO_DISCONNECT_GRACE_PERIOD"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.SoloDisconnectGracePeriod = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("SOLVER_TIMEOUT"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			cfg.SolverTimeout = time.Duration(secs) * time.Second
		}
	}

	if v := os.Getenv("ENABLE_DAILY_CHALLENGE"); v != "" {
		cfg.EnableDailyChallenge = v == "true" || v == "1"
	}

	if v := os.Getenv("PUBLIC_CLIENT_URL"); v != "" {
		cfg.PublicClientURL = strings.TrimSuffix(v, "/")
	}

	if v := os.Getenv("PUBLIC_SERVER_URL"); v != "" {
		cfg.PublicServerURL = strings.TrimSuffix(v, "/")
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
