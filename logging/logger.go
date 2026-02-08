package logging

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package logging provides a configurable wrapper around Go's standard log package.

Summary
-------
- Configures the global default logger used by log.Print*, log.Fatal*, etc.
- Supports logging to stdout and optionally to a file at the same time.
- Can be fully configured via a JSON-serializable Config struct.
- Designed to integrate cleanly with a central application config.
- Keeps dependencies minimal and relies only on the standard library.
*/

import (
	"io"
	"log"
	"os"
)

// Config defines the runtime configuration for the global logger.
// It is intended to be loaded from JSON configuration files.
type Config struct {
	Enabled   bool   `json:"enabled"`    // Enable or disable logging completely
	Level     string `json:"level"`      // debug, info, warn, error (reserved for future use)
	File      string `json:"file"`       // Log file path; empty means stdout only
	Flags     int    `json:"flags"`      // Explicit log flags (overrides computed flags if set)
	UTC       bool   `json:"utc"`        // Use UTC timestamps
	ShortFile bool   `json:"short_file"` // Include short file name and line number
}

// DefaultConfig returns a sane default logger configuration.
// These defaults are suitable for most production services.
func DefaultConfig() Config {
	return Config{
		Enabled:   true,
		Level:     "info",
		File:      "",
		Flags:     0,
		UTC:       true,
		ShortFile: false,
	}
}

// Init configures the global default logger according to the provided config.
// After calling Init, all existing log.Print*, log.Fatal*, and log.Panic*
// calls will use the configured output and flags.
func Init(c ...Config) error {
	// Start with default configuration
	cfg := DefaultConfig()
	if len(c) > 0 {
		cfg = c[0]
	}

	// Disable logging entirely if requested
	if !cfg.Enabled {
		log.SetOutput(io.Discard)
		return nil
	}

	// Default output is stdout
	var out io.Writer = os.Stdout

	// Optionally append logs to a file
	if cfg.File != "" {
		f, err := os.OpenFile(
			cfg.File,
			os.O_CREATE|os.O_WRONLY|os.O_APPEND,
			0644,
		)
		if err != nil {
			return err
		}

		// Write logs to both stdout and file
		out = io.MultiWriter(os.Stdout, f)
	}

	log.SetOutput(out)
	log.SetFlags(resolveFlags(cfg))

	return nil
}

// resolveFlags computes log flags from the configuration.
// If cfg.Flags is non-zero, it overrides all computed flags.
func resolveFlags(cfg Config) int {
	flags := log.Ldate | log.Ltime

	if cfg.UTC {
		flags |= log.LUTC
	}
	if cfg.ShortFile {
		flags |= log.Lshortfile
	}

	// Explicit flags override computed ones
	if cfg.Flags != 0 {
		flags = cfg.Flags
	}

	return flags
}
