package config

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package config provides a small, generic configuration manager for Go services.

Overview
--------
This package offers a lightweight, type-safe way to manage application
configuration that is backed by a JSON file.

Core ideas:
- A configuration always consists of two parts:
  1. A strongly typed config struct (generic parameter T)
  2. The filename it was loaded from or saved to
- The filename is treated as part of the configuration state
- Callers work directly on the config struct via a pointer
- Persistence is explicit (Load / Save / SaveAs)

Design goals:
- Minimal API surface
- No reflection beyond encoding/json
- Explicit, predictable behavior (no magic reloading)
- Suitable for CLIs, services, and small tools
- Easy to extend with defaults, validation, or env overrides

Typical usage:

	cfg := config.New("config.json", MyConfig{})
	_ = cfg.Load("config.json")

	cfg.Get().Port = 8080
	_ = cfg.Save()

Thread-safety:
- This type is NOT concurrency-safe by design
- Intended to be configured at startup or in single-threaded CLI tools
*/

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
)

// Config wraps a typed configuration together with its associated file path.
//
// Both fields are intentionally unexported to enforce controlled access
// via methods (encapsulation).
type Config[T any] struct {
	filename string
	cfg      T
}

// New creates a new Config instance with an initial filename and config value.
//
// The filename may be empty and set later via SetFilename or Load.
func New[T any](filename string, cfg T) *Config[T] {
	return &Config[T]{
		filename: filename,
		cfg:      cfg,
	}
}

/* --------------------------------------------------------------------------
   Filename handling
   -------------------------------------------------------------------------- */

// Filename returns the currently associated configuration file path.
func (c *Config[T]) Filename() string {
	return c.filename
}

// SetFilename sets or updates the configuration file path.
//
// This does not read or write any files.
func (c *Config[T]) SetFilename(path string) {
	c.filename = path
}

/* --------------------------------------------------------------------------
   Config access
   -------------------------------------------------------------------------- */

// Get returns a pointer to the underlying configuration struct.
//
// Mutating the returned value directly modifies the stored configuration.
// This is intentional to keep usage ergonomic.
func (c *Config[T]) Get() *T {
	return &c.cfg
}

/* --------------------------------------------------------------------------
   Persistence
   -------------------------------------------------------------------------- */

// Load reads a JSON configuration file and unmarshals it into the config.
//
// On success, the internal filename is updated to the loaded path.
func (c *Config[T]) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &c.cfg); err != nil {
		return err
	}

	c.filename = path
	return nil
}

// Save writes the current configuration to the previously configured filename.
//
// Returns ErrNoFilename if no filename has been set.
func (c *Config[T]) Save() error {
	if c.filename == "" {
		return ErrNoFilename
	}
	return c.SaveAs(c.filename)
}

// SaveAs writes the current configuration to the given file path.
//
// Parent directories are created automatically.
// The internal filename is updated on success.
func (c *Config[T]) SaveAs(filename string) error {
	b, err := json.MarshalIndent(c.cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
	}

	if err := os.WriteFile(filename, b, 0644); err != nil {
		return err
	}

	c.filename = filename
	return nil
}

// ErrNoFilename is returned when Save is called without a configured filename.
var ErrNoFilename = errors.New("no filename set")
