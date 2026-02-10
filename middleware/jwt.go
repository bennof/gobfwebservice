// middleware/bearer_context.go
//
// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner
//
// -----------------------------------------------------------------------------
// Overview
// -----------------------------------------------------------------------------
//
// This file provides lightweight Bearer-token middleware.
//
// Purpose:
//   - Extract a Bearer token from the Authorization header
//   - Optionally parse token claims
//   - Store token and/or claims in the request context
//
// Design goals:
//   - No token validation logic (can be handled by nginx auth_request)
//   - No hard dependency on JWT
//   - Zero allocations if no Bearer header is present
//   - Type-safe or map-based claim access
//
// Typical usage:
//
//   mux.Handle(
//       "/api",
//       middleware.BearerContextTyped(authSvc.ParseJWTClaims)(
//           handler,
//       ),
//   )
//
// Or (map-based):
//
//   middleware.BearerContextMap(parseFunc)
//
// -----------------------------------------------------------------------------

package middleware

import (
	"context"
	"net/http"
	"strings"
)

// -----------------------------------------------------------------------------
// Context keys (unexported)
// -----------------------------------------------------------------------------

// ctxKeyBearerToken stores the raw Bearer token string.
type ctxKeyBearerToken struct{}

// ctxKeyBearerClaims stores typed claims (generic).
type ctxKeyBearerClaims[T any] struct{}

// ctxKeyBearerClaimsMap stores untyped (map-based) claims.
type ctxKeyBearerClaimsMap struct{}

// -----------------------------------------------------------------------------
// Parser types
// -----------------------------------------------------------------------------

// BearerParser parses a token into a typed claims object.
//
// Parsing errors are ignored by the middleware;
// absence of claims simply means "unauthenticated".
type BearerParser[T any] func(token string) (*T, error)

// BearerMapParser parses a token into a map-based claims structure.
type BearerMapParser func(token string) (map[string]any, error)

// -----------------------------------------------------------------------------
// Middleware: token only
// -----------------------------------------------------------------------------

// BearerContext extracts the Bearer token from the Authorization header
// and stores it in the request context.
//
// No validation or parsing is performed.
func BearerContext() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(h), "bearer ") {
				token := strings.TrimSpace(h[len("Bearer "):])
				ctx := context.WithValue(r.Context(), ctxKeyBearerToken{}, token)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// -----------------------------------------------------------------------------
// Middleware: typed claims
// -----------------------------------------------------------------------------

// BearerContextTyped extracts a Bearer token and parses it into typed claims.
//
// If parsing fails, the request continues without claims.
// This is intentional to allow:
//   - nginx-based verification
//   - optional authentication
func BearerContextTyped[T any](parser BearerParser[T]) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(h), "bearer ") {
				token := strings.TrimSpace(h[len("Bearer "):])

				if claims, err := parser(token); err == nil {
					ctx := context.WithValue(r.Context(), ctxKeyBearerClaims[T]{}, claims)
					ctx = context.WithValue(ctx, ctxKeyBearerToken{}, token)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// -----------------------------------------------------------------------------
// Middleware: map-based claims
// -----------------------------------------------------------------------------

// BearerContextMap extracts a Bearer token and parses it into a map.
//
// Useful for:
//   - reverse proxies
//   - dynamic claim inspection
//   - systems without a fixed claim schema
func BearerContextMap(parser BearerMapParser) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if strings.HasPrefix(strings.ToLower(h), "bearer ") {
				token := strings.TrimSpace(h[len("Bearer "):])

				if claims, err := parser(token); err == nil {
					ctx := context.WithValue(r.Context(), ctxKeyBearerClaimsMap{}, claims)
					ctx = context.WithValue(ctx, ctxKeyBearerToken{}, token)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// -----------------------------------------------------------------------------
// Getters (read-only)
// -----------------------------------------------------------------------------

// GetBearerToken returns the raw Bearer token from context.
func GetBearerToken(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyBearerToken{})
	t, ok := v.(string)
	return t, ok
}

// GetBearerClaimsTyped returns typed claims from context.
func GetBearerClaimsTyped[T any](ctx context.Context) (*T, bool) {
	v := ctx.Value(ctxKeyBearerClaims[T]{})
	c, ok := v.(*T)
	return c, ok
}

// GetBearerClaimsMap returns map-based claims from context.
func GetBearerClaimsMap(ctx context.Context) (map[string]any, bool) {
	v := ctx.Value(ctxKeyBearerClaimsMap{})
	m, ok := v.(map[string]any)
	return m, ok
}
