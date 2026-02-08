package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/bennof/go-bfwebservice/example"
	"github.com/bennof/go-bfwebservice/logging"
	"github.com/bennof/go-bfwebservice/middleware"
	"github.com/bennof/go-bfwebservice/server"
	"github.com/bennof/go-bfwebservice/templates"
)

func main() {
	// ------------------------------------------------------------------
	// CLI flag
	// ------------------------------------------------------------------
	out := flag.String("out", "config.json", "output config file (path + filename)")
	flag.Parse()

	// ------------------------------------------------------------------
	// Build default configuration
	// ------------------------------------------------------------------
	cfg := example.ExampleConfig{
		Server: server.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  10,
			WriteTimeout: 10,
		},
		TemplateFolder: templates.DefaultTemplateSetConfig("example/templates"),
		ErrorTemplate:  "error.html",
		Log:            logging.DefaultConfig(),
		Cors:           middleware.DefaultCORSConfig(),
		Rates:          middleware.DefaultRateLimitConfig(),
	}

	// ------------------------------------------------------------------
	// Encode JSON
	// ------------------------------------------------------------------
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}

	// ------------------------------------------------------------------
	// Ensure target directory exists
	// ------------------------------------------------------------------
	dir := filepath.Dir(*out)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("failed to create directory: %v", err)
		}
	}

	// ------------------------------------------------------------------
	// Write file
	// ------------------------------------------------------------------
	if err := os.WriteFile(*out, data, 0644); err != nil {
		log.Fatalf("failed to write config: %v", err)
	}
}
