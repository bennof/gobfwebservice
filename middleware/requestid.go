package middleware

/*
Request ID middleware for HTTP request correlation.

Summary
-------
- Ensures every incoming HTTP request has a unique request identifier.
- Accepts an existing request ID from the X-Request-ID header if present.
- Generates a new request ID otherwise.
- Injects the request ID into the request context.
- Returns the request ID to the client via the X-Request-ID response header.
- Enables log correlation across middleware, handlers, and services.
*/

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// ctxKeyRequestID is an unexported context key type used to avoid
// collisions with other context values.
type ctxKeyRequestID struct{}

// RequestID is an HTTP middleware that injects a request ID into the
// request context and response headers.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to reuse an incoming request ID if provided by the client
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			// Generate a new request ID if none is present
			id = uuid.NewString()
		}

		// Store the request ID in the context
		ctx := context.WithValue(r.Context(), ctxKeyRequestID{}, id)

		// Expose the request ID to the client
		w.Header().Set("X-Request-ID", id)

		// Continue request handling with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from the given context.
// It returns an empty string if no request ID is present.
func GetRequestID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyRequestID{}).(string); ok {
		return v
	}
	return ""
}
