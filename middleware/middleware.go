package middleware

import "net/http"

// Middleware defines a standard HTTP middleware.
type Middleware func(http.Handler) http.Handler
