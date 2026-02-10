package middleware

/*
Recovery middleware for panic-safe HTTP handling.

Summary
-------
- Protects the HTTP server from panics occurring in handlers or downstream middleware.
- Converts panics into HTTP 500 Internal Server Error responses.
- Logs the panic value together with a stack trace.
- Prevents a single faulty request from crashing the entire process.
- Intended to be used early in the middleware chain.
*/

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/bennof/gobfwebservice/server"
)

// Recovery is an HTTP middleware that intercepts panics during request handling.
// If a panic occurs, it logs the panic and stack trace and responds with
// a 500 Internal Server Error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure panics do not propagate and crash the server
		defer func() {
			if rec := recover(); rec != nil {
				// Log panic details and stack trace for diagnostics
				log.Printf("panic: %v\n%s", rec, debug.Stack())

				// Return a generic error response to the client
				server.InternalServerError(w, r)
			}
		}()

		// Delegate request handling to the next handler
		next.ServeHTTP(w, r)
	})
}
