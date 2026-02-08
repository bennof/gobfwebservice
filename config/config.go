package config

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package config provides a small, generic configuration manager for Go services.

Summary
-------
- Manages a typed configuration struct together with its source file path.
- Supports loading from and saving to JSON files.
- Keeps all internal fields private to enforce encapsulation.
- Exposes controlled access via getters and setters.
- Returns references to the underlying config so modifications are persisted.
- Designed to be extended with defaults, validation, or environment overrides.

Typical usage:
- Create a Config[T] with New(...)
- Load() or modify the config via Get()
- Save() or SaveAs() to persist changes
*/

import (
	"encoding/json"
	"errors"
	"os"
)

// Config wraps a typed configuration together with the file it is loaded from.
// The fields are intentionally private to prevent uncontrolled modification.
type Config[T any] struct {
	filename string
	cfg      T
}

// New creates a new Config instance with an initial filename and config value.
func New[T any](filename string, cfg T) *Config[T] {
	return &Config[T]{
		filename: filename,
		cfg:      cfg,
	}
}

/* ---------- filename handling ---------- */

// Filename returns the currently associated configuration file path.
func (c *Config[T]) Filename() string {
	return c.filename
}

// SetFilename sets or updates the configuration file path.
func (c *Config[T]) SetFilename(path string) {
	c.filename = path
}

/* ---------- config access ---------- */

// Get returns a pointer to the underlying configuration.
// Modifying the returned value directly affects the stored config.
func (c *Config[T]) Get() *T {
	return &c.cfg
}

/* ---------- persistence ---------- */

// Load reads a JSON configuration file and unmarshals it into the config.
// The internal filename is updated to the loaded path.
func (c *Config[T]) Load(filepath string) error {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &c.cfg); err != nil {
		return err
	}
	c.filename = filepath
	return nil
}

// Save writes the current configuration to the previously set filename.
// Returns an error if no filename is configured.
func (c *Config[T]) Save() error {
	if c.filename == "" {
		return ErrNoFilename
	}
	return c.SaveAs(c.filename)
}

// SaveAs writes the current configuration to the given file path
// and updates the internal filename accordingly.
func (c *Config[T]) SaveAs(filepath string) error {
	b, err := json.MarshalIndent(c.cfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath, b, 0644); err != nil {
		return err
	}
	c.filename = filepath
	return nil
}

// ErrNoFilename is returned when Save is called without a configured filename.
var ErrNoFilename = errors.New("no filename set")
