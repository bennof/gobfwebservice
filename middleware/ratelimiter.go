package middleware

/*
Rate limiting middleware with bounded memory usage.

Summary
-------
- Implements a simple, resource-protecting rate limiter.
- Limits requests per client IP within a fixed time window.
- Adds a hard cap on the number of tracked clients to prevent
  unbounded memory growth.
- Uses a global reset timer to clear all counters periodically.
- Designed for low-resource systems and small services where
  predictable memory usage is more important than perfect fairness.
*/

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/bennof/go-bfwebservice/server"
)

/* ---------- configuration ---------- */

// RateLimitConfig defines the configuration for the rate limiting middleware.
// It is JSON-serializable and intended to be part of a global application config.
type RateLimitConfig struct {
	MaxRequests int           `json:"max_requests"` // Maximum requests per client IP within the window
	MaxClients  int           `json:"max_clients"`  // Maximum number of distinct clients tracked per window
	Window      time.Duration `json:"window"`       // Time window for rate limiting
}

// DefaultRateLimitConfig returns a conservative default configuration.
// These defaults are suitable for small services running on limited hardware.
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		MaxRequests: 100,
		MaxClients:  1000,
		Window:      time.Minute,
	}
}

/* ---------- middleware ---------- */

// RateLimit creates a rate limiting middleware based on the given configuration.
// The middleware enforces both per-client request limits and a global cap on
// the number of tracked clients to ensure bounded memory usage.
func RateLimit(cfg ...RateLimitConfig) Middleware {
	// Start with default configuration
	c := DefaultRateLimitConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	var (
		mu    sync.Mutex
		hits  = map[string]int{} // request counters per client IP
		reset = time.Now().Add(c.Window)
	)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now()

			mu.Lock()
			// Reset all counters when the time window expires
			if now.After(reset) {
				hits = map[string]int{}
				reset = now.Add(c.Window)
			}

			// Extract client IP (RemoteAddr is usually "IP:PORT")
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				mu.Unlock()
				server.BadRequest(w, r)
				return
			}

			// Reject new clients if the map size limit is reached
			if _, exists := hits[host]; !exists && len(hits) >= c.MaxClients {
				mu.Unlock()
				server.TooManyRequests(w, r)
				return
			}

			// Increment request counter for this client
			hits[host]++
			count := hits[host]
			mu.Unlock()

			// Enforce per-client request limit
			if count > c.MaxRequests {
				server.TooManyRequests(w, r)
				return
			}

			// Delegate to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
