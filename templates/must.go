package templates

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package templates provides small helper utilities related to template handling.

Summary
-------
- Contains generic helper functions to simplify error handling.
- Intended for use during application startup or template initialization.
- Panics the application in case of unrecoverable errors.
*/

import "log"

// Must returns the provided value if err is nil.
// If err is non-nil, the function logs the error and terminates the program.
//
// This helper is intended for initialization-time operations where
// a failure should abort execution (e.g. parsing templates).
func Must[T any](v T, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	return v
}
