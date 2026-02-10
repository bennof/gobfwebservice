package example

/*
ExampleConfig defines the full configuration for the example service.

Summary
-------
- Aggregates all subsystem configurations into a single JSON-serializable struct.
- Designed to be loaded via the central config management.
- Keeps concerns separated while allowing unified configuration.
*/

import (
	"github.com/bennof/gobfwebservice/logging"
	"github.com/bennof/gobfwebservice/middleware"
	"github.com/bennof/gobfwebservice/server"
	"github.com/bennof/gobfwebservice/templates"
)

// ExampleConfig bundles all configuration sections required by the example service.
type ExampleConfig struct {
	Server         server.ServerConfig         `json:"server"`
	TemplateFolder templates.TemplateSetConfig `json:"templates"`
	ErrorTemplate  string                      `json:"error_template"`
	Log            logging.Config              `json:"logging"`
	Cors           middleware.CORSConfig       `json:"cors"`
	Rates          middleware.RateLimitConfig  `json:"rate_limit"`
}
