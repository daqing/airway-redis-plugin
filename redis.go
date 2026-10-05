// Package redis is the shared Redis integration for Airway applications:
// environment-based configuration, a process-wide client handle and a health
// check. It is the foundation for features like Redis-backed download
// aggregation and rate limiting.
//
// It complements the framework's lib/redis_client: same REDIS URL convention,
// but Setup verifies connectivity with PING and absent configuration is a
// first-class state (Enabled reports false) instead of a missing client.
package redis

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

// Config carries the Redis connection settings.
type Config struct {
	// URL in the form redis://[username:password@]host:port/db.
	URL string
}

// FromEnv reads the configuration from AIRWAY_REDIS or REDIS (the same
// variables the Airway server honors); the process environment wins over
// nothing — callers load .env beforehand if they want file fallbacks.
func FromEnv() Config {
	url := os.Getenv("AIRWAY_REDIS")
	if url == "" {
		url = os.Getenv("REDIS")
	}
	return Config{URL: url}
}

// Enabled reports whether a Redis URL is configured.
func (c Config) Enabled() bool {
	return c.URL != ""
}

var (
	mu     sync.RWMutex
	client *redis.Client
)

// Setup parses cfg.URL, builds the process-wide client and verifies it with
// PING. Calling Setup again replaces the previous client (closing it). A
// Config with an empty URL is an error — check Enabled first when Redis is
// optional.
func Setup(cfg Config) error {
	if cfg.URL == "" {
		return fmt.Errorf("redis: empty URL; check Config.Enabled before Setup")
	}

	opts, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return fmt.Errorf("redis: invalid URL: %w", err)
	}

	c := redis.NewClient(opts)
	if err := c.Ping(context.Background()).Err(); err != nil {
		_ = c.Close()
		return fmt.Errorf("redis: ping failed: %w", err)
	}

	mu.Lock()
	old := client
	client = c
	mu.Unlock()

	if old != nil {
		_ = old.Close()
	}

	return nil
}

// Current returns the process-wide client, or nil when Setup never ran.
func Current() *redis.Client {
	mu.RLock()
	defer mu.RUnlock()
	return client
}

// Must returns the process-wide client and panics when Setup never ran.
func Must() *redis.Client {
	c := Current()
	if c == nil {
		panic("redis: client not initialized, call Setup first")
	}
	return c
}

// Ping health-checks the process-wide client.
func Ping(ctx context.Context) error {
	c := Current()
	if c == nil {
		return fmt.Errorf("redis: client not initialized, call Setup first")
	}
	return c.Ping(ctx).Err()
}
