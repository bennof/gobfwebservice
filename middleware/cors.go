package middleware

/*
CORS middleware implementation with configurable policies.

Summary
-------
- Provides a configurable CORS (Cross-Origin Resource Sharing) middleware.
- Uses a JSON-serializable configuration struct.
- Supports sensible defaults via DefaultCORSConfig().
- Allows optional configuration by using a variadic constructor.
- Handles CORS preflight (OPTIONS) requests automatically.
*/

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig defines the configuration options for the CORS middleware.
// All fields are JSON-serializable and intended to be part of a global app config.
type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins"`   // List of allowed origins (e.g. "*", "https://example.com")
	AllowedMethods   []string `json:"allowed_methods"`   // Allowed HTTP methods
	AllowedHeaders   []string `json:"allowed_headers"`   // Allowed request headers
	AllowCredentials bool     `json:"allow_credentials"` // Whether credentials (cookies, auth headers) are allowed
	MaxAge           int      `json:"max_age"`           // Preflight cache duration in seconds
}

// DefaultCORSConfig returns a permissive default CORS configuration.
// This default is suitable for development and simple public APIs.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           600,
	}
}

// CORS creates a CORS middleware using the provided configuration.
// If no configuration is supplied, DefaultCORSConfig() is used.
//
// Usage:
//
//	middleware.CORS()              // default configuration
//	middleware.CORS(customConfig)  // custom configuration
func CORS(cfg ...CORSConfig) Middleware {
	// Start with default configuration
	c := DefaultCORSConfig()
	if len(cfg) > 0 {
		c = cfg[0]
	}

	// Precompute header values for efficiency
	origins := strings.Join(c.AllowedOrigins, ", ")
	methods := strings.Join(c.AllowedMethods, ", ")
	headers := strings.Join(c.AllowedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS response headers
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if c.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if c.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(c.MaxAge))
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// Delegate to the next handler
			next.ServeHTTP(w, r)
		})
	}
}
