package server

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package server provides reusable HTTP error handling helpers for web services.

Summary
-------
- Centralizes rendering of common HTTP error responses (4xx / 5xx).
- Supports an optional, shared HTML template for error pages.
- Falls back to plain status codes if no template is configured.
- Suppresses HTML error pages for static asset requests
  (e.g. JS, CSS, images, fonts) to avoid polluting asset responses.
- Designed to be framework-agnostic and usable with net/http directly.
*/

import (
	"html/template"
	"net/http"
	"path/filepath"
	"strings"
)

var (
	// errorTemplate holds the parsed template used to render error pages.
	errorTemplate *template.Template

	// errorTemplateName is the name of the template block to execute.
	// If empty, errors are returned without rendering HTML.
	errorTemplateName string = ""
)

// SetErrorTemplate configures a shared HTML template for error pages.
// If name is empty, HTML rendering is disabled and only status codes are sent.
func SetErrorTemplate(tpl *template.Template, name string) {
	errorTemplate = tpl
	errorTemplateName = name
}

/* ---------- HTTP error handlers ---------- */

// BadRequest renders a 400 Bad Request error.
func BadRequest(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusBadRequest, "Bad Request", "The request could not be processed.")
}

// Unauthorized renders a 401 Unauthorized error.
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusUnauthorized, "Unauthorized", "You must authenticate to access this resource.")
}

// Forbidden renders a 403 Forbidden error.
func Forbidden(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusForbidden, "Forbidden", "You do not have permission to access this resource.")
}

// NotFound renders a 404 Not Found error.
func NotFound(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusNotFound, "Not Found", "The requested page does not exist.")
}

// MethodNotAllowed renders a 405 Method Not Allowed error.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusMethodNotAllowed, "Method Not Allowed", "The HTTP method used is not allowed for this resource.")
}

// InternalServerError renders a 500 Internal Server Error.
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusInternalServerError, "Internal Server Error", "An error occurred on the server.")
}

// ServiceUnavailable renders a 503 Service Unavailable error.
func ServiceUnavailable(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusServiceUnavailable, "Service Unavailable", "The server is currently unavailable. Please try again later.")
}

// TooManyRequests renders a 429 Too Many Requests error.
func TooManyRequests(w http.ResponseWriter, r *http.Request) {
	RenderError(
		w,
		r,
		http.StatusTooManyRequests,
		"Too Many Requests",
		"You have sent too many requests in a given amount of time. Please try again later.",
	)
}

/* ---------- core rendering ---------- */

// RenderError renders an HTTP error response with the given status code,
// title, and message. Depending on configuration, this either renders an
// HTML template or sends a plain status code.
func RenderError(w http.ResponseWriter, r *http.Request, code int, title, message string) {
	// Suppress HTML error pages for static asset requests
	if isSilentError(w, r, code) {
		return
	}

	// If no template is configured, return only the status code
	if errorTemplateName == "" {
		w.Header().Del("Content-Type")
		w.WriteHeader(code)
		return
	}

	// Render the configured HTML error template
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)

	data := map[string]interface{}{
		"Code":    code,
		"Title":   title,
		"Message": message,
		"Path":    r.URL.Path,
	}

	if err := errorTemplate.ExecuteTemplate(w, errorTemplateName, data); err != nil {
		// Fallback to a plain HTTP error if template rendering fails
		http.Error(w, message, code)
	}
}

/* ---------- helpers ---------- */

// isSilentError returns true for requests targeting static assets.
// In these cases, no HTML error page is rendered to avoid corrupting
// asset responses (e.g. JS, CSS, images, fonts).
func isSilentError(w http.ResponseWriter, r *http.Request, code int) bool {
	ext := strings.ToLower(filepath.Ext(r.URL.Path))

	switch ext {
	case ".js", ".css", ".map", ".ico", ".png", ".svg", ".jpg", ".jpeg", ".webp",
		".woff", ".woff2", ".ttf", ".eot", ".gif", ".pdf", ".json", ".xml":
		w.Header().Del("Content-Type")
		w.WriteHeader(code)
		return true
	}

	return false
}
