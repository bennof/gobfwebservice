package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bennof/gobfwebservice/config"
	"github.com/bennof/gobfwebservice/example"
	"github.com/bennof/gobfwebservice/logging"
	"github.com/bennof/gobfwebservice/middleware"
	"github.com/bennof/gobfwebservice/server"
	"github.com/bennof/gobfwebservice/templates"
)

var CFG config.Config[example.ExampleConfig]

func main() {
	// A command is required as the first argument.
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	// Dispatch subcommands explicitly.
	switch cmd {
	case "init-config":
		runInitConfig(args)

	case "serve":
		runServer(args)

	default:
		fmt.Printf("unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

// usage prints a short help text describing available commands.
func usage() {
	fmt.Println(`auth-cli commands:

	serve

  init-config   -out config.json
  
`)
}

// fatal terminates the program on unrecoverable errors.
//
// In a CLI context this is acceptable and keeps the control flow simple.
func fatal(err error) {
	if err != nil {
		panic(err)
	}
}

func runInitConfig(args []string) {
	// ------------------------------------------------------------------
	// CLI flag
	// ------------------------------------------------------------------
	fs := flag.NewFlagSet("init-config", flag.ExitOnError)
	cfgPath := fs.String("out", "config.json", "output config file")
	fs.Parse(args)

	fmt.Println("Initializing default configuration...")

	// Obtain a mutable reference to the internal config
	cfg := CFG.Get()

	// ------------------------------------------------------------------
	// Build default configuration
	// ------------------------------------------------------------------

	cfg.Server = server.ServerConfig{
		Host:         "localhost",
		Port:         8080,
		ReadTimeout:  10,
		WriteTimeout: 10,
	}
	cfg.TemplateFolder = templates.DefaultTemplateSetConfig("example/templates")
	cfg.ErrorTemplate = "error.html"
	cfg.Log = logging.DefaultConfig()
	cfg.Cors = middleware.DefaultCORSConfig()
	cfg.Rates = middleware.DefaultRateLimitConfig()

	// ------------------------------------------------------------------
	// Write file
	// ------------------------------------------------------------------
	// Persist configuration to disk
	if err := CFG.SaveAs(*cfgPath); err != nil {
		fatal(err)
	}

	fmt.Printf("Configuration written to %s\n", *cfgPath)
}

func runServer(args []string) {
	fs := flag.NewFlagSet("init-config", flag.ExitOnError)
	cfgFile := fs.String("config", "config.json", "path to config file")
	fs.Parse(args)

	// ------------------------------------------------------------
	// Load config
	// ------------------------------------------------------------
	if err := CFG.Load(*cfgFile); err != nil {
		fatal(err)
	}

	cfg := CFG.Get()

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
