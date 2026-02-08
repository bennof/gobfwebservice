package server

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

/*
Package server provides a small wrapper around net/http to run an HTTP server
with typed configuration and graceful shutdown support.

Summary
-------
- Defines a ServerConfig struct for JSON-serializable server settings.
- Wraps http.Server together with a ServeMux for route registration.
- Supports blocking start as well as managed run modes.
- Implements graceful shutdown using OS signals and contexts.
- Allows integration into larger applications via context-based lifecycle control.
*/

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/* ---------- configuration ---------- */

// ServerConfig holds server-specific runtime settings.
// It is designed to be loaded from JSON configuration files.
type ServerConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	ReadTimeout  int    `json:"read_timeout"`  // seconds
	WriteTimeout int    `json:"write_timeout"` // seconds
}

/* ---------- server wrapper ---------- */

// Server represents an HTTP server instance with its configuration
// and routing multiplexer.
type Server struct {
	config     *ServerConfig
	httpServer *http.Server
	mux        *http.ServeMux
}

// NewServer creates a new Server instance using the provided configuration
// and optional ServeMux. If mux is nil, a new ServeMux is created.
func NewServer(cfg *ServerConfig, mux *http.ServeMux) (*Server, error) {
	if mux == nil {
		mux = http.NewServeMux()
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	s := &Server{
		config: cfg,
		mux:    mux,
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      mux,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
	}

	return s, nil
}

/* ---------- accessors ---------- */

// Mux returns the underlying ServeMux used for route registration.
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}

// Config returns the server configuration.
// The returned pointer allows runtime inspection or modification.
func (s *Server) Config() *ServerConfig {
	return s.config
}

/* ---------- lifecycle ---------- */

// Start starts the HTTP server and blocks until it stops.
// This method does not handle graceful shutdown.
func (s *Server) Start() error {
	log.Printf("Starting server on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server using the provided context.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down server...")
	return s.httpServer.Shutdown(ctx)
}

// Run starts the server and installs OS signal handlers for graceful shutdown.
// It listens for SIGINT and SIGTERM and shuts the server down with a fixed timeout.
func (s *Server) Run() error {
	// Channel to receive server runtime errors
	serverErrors := make(chan error, 1)

	// Start server asynchronously
	go func() {
		log.Printf("Server listening on %s", s.httpServer.Addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for either a server error or an OS shutdown signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-quit:
		log.Printf("Received signal: %v", sig)

		// Create shutdown context with a fixed timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := s.httpServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}

// RunWithContext starts the server and shuts it down when either the given
// context is cancelled or an OS shutdown signal is received.
// The shutdown timeout is configurable.
func (s *Server) RunWithContext(ctx context.Context, shutdownTimeout time.Duration) error {
	// Channel to receive server runtime errors
	serverErrors := make(chan error, 1)

	// Start server asynchronously
	go func() {
		log.Printf("Server listening on %s", s.httpServer.Addr)
		serverErrors <- s.httpServer.ListenAndServe()
	}()

	// Setup signal handling for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Wait for server error, context cancellation, or OS signal
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")

	case sig := <-quit:
		log.Printf("Received signal: %v", sig)
	}

	// Create shutdown context with the provided timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Server stopped gracefully")
	return nil
}
