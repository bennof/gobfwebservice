package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/bennof/gobfwebservice/example"
	"github.com/bennof/gobfwebservice/logging"
	"github.com/bennof/gobfwebservice/middleware"
	"github.com/bennof/gobfwebservice/server"
	"github.com/bennof/gobfwebservice/templates"
)

func main() {
	// ------------------------------------------------------------
	// CLI
	// ------------------------------------------------------------
	cfgFile := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	// ------------------------------------------------------------
	// Load config
	// ------------------------------------------------------------
	data, err := os.ReadFile(*cfgFile)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var cfg example.ExampleConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	// ------------------------------------------------------------
	// Init logging (global)
	// ------------------------------------------------------------
	if err := logging.Init(cfg.Log); err != nil {
		log.Fatalf("failed to init logging: %v", err)
	}

	// ------------------------------------------------------------
	// Templates + error handling
	// ------------------------------------------------------------
	tmpl, err := templates.LoadTemplates(cfg.TemplateFolder.Folder)
	if err != nil {
		log.Fatalf("failed to load templates: %v", err)
	}

	server.SetErrorTemplate(
		templates.Must(tmpl.Get(cfg.ErrorTemplate)),
		cfg.ErrorTemplate,
	)

	// ------------------------------------------------------------
	// Routing
	// ------------------------------------------------------------
	mux := http.NewServeMux()

	// plain HTML
	mux.HandleFunc("/", HelloHTML)

	// API with middleware stack
	mux.Handle("/api/",
		middleware.CORS(cfg.Cors)(
			middleware.RateLimit(cfg.Rates)(
				middleware.Recovery(
					middleware.RequestID(
						middleware.Logging(
							http.HandlerFunc(HelloJSON),
						),
					),
				),
			),
		),
	)

	// ------------------------------------------------------------
	// Server
	// ------------------------------------------------------------
	srv, err := server.NewServer(&cfg.Server, mux)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
