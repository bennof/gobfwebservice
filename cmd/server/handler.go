package main

/*
Example handlers for quick testing.

Summary
-------
- HelloHTML: returns a simple HTML page.
- HelloJSON: returns a simple JSON response for API testing.
*/

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/bennof/go-bfwebservice/middleware"
)

// HelloHTML writes a minimal HTML response.
func HelloHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>Hello</title></head>
<body>
  <h1>Hello (HTML)</h1>
  <p>This is a test handler.</p>
</body>
</html>`))
}

// HelloJSON writes a minimal JSON response.
func HelloJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	resp := map[string]any{
		"ok":         true,
		"message":    "Hello (JSON)",
		"path":       r.URL.RequestURI(),
		"method":     r.Method,
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"request_id": middleware.GetRequestID(r.Context()),
	}

	_ = json.NewEncoder(w).Encode(resp)
}
