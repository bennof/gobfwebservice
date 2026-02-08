package middleware

/*
Logging middleware for HTTP request tracing.

Summary
-------
- Logs exactly one entry per HTTP request.
- Captures method, path, status code, duration, and request ID.
- Uses Go's global standard logger (log.Printf), so output format and
  destination are controlled by the central logging configuration.
- Designed to be lightweight and free of business logic.
*/

import (
	"log"
	"net/http"
	"time"
)

// statusRecorder wraps an http.ResponseWriter to capture the HTTP status code
// written by the handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader intercepts the status code before delegating to the underlying
// ResponseWriter.
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging is an HTTP middleware that logs basic request information.
// It measures request duration and logs method, path, status code,
// elapsed time, and request ID.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Record start time for duration measurement
		start := time.Now()

		// Wrap the ResponseWriter to capture the status code
		rec := &statusRecorder{
			ResponseWriter: w,
			status:         http.StatusOK, // default if WriteHeader is not called
		}

		// Execute the next handler in the chain
		next.ServeHTTP(rec, r)

		// Log request details after the handler has completed
		dur := time.Since(start)
		log.Printf(
			"%s %s %d %s rid=%s",
			r.Method,
			r.URL.Path,
			rec.status,
			dur,
			GetRequestID(r.Context()),
		)
	})
}
