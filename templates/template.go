package templates

// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Benjamin Benno Falkner

// Package templates provides template management with shared layouts and pre-rendering capabilities.
//
// # Overview
//
// The templates package offers a flexible system for managing HTML templates with the following features:
//
//   - Shared layouts: Define common layouts (header, footer, etc.) once and reuse across templates
//   - Dynamic rendering: Render templates directly to HTTP responses
//   - Pre-rendering: Render templates to strings or bytes for caching, static site generation, or email
//   - Template reloading: Hot-reload templates during development
//   - Layout inheritance: Each view template automatically inherits from shared layouts
//
// # Directory Structure
//
// Templates should be organized as follows:
//
//	templates/
//	├── layout/
//	│   ├── base.html      # Main layout
//	│   └── admin.html     # Admin layout
//	├── home.html          # View templates
//	├── about.html
//	└── contact.html
//
// # Layout Templates
//
// Layout templates define the common structure using {{define}} and {{block}}:
//
//	{{define "base"}}
//	<!DOCTYPE html>
//	<html>
//	<head>
//	    <title>{{block "title" .}}Default Title{{end}}</title>
//	</head>
//	<body>
//	    {{block "content" .}}{{end}}
//	</body>
//	</html>
//	{{end}}
//
// # View Templates
//
// View templates extend layouts by defining their blocks:
//
//	{{template "base" .}}
//	{{define "title"}}My Page{{end}}
//	{{define "content"}}
//	    <h1>Hello World</h1>
//	{{end}}
//
// # Usage Examples
//
// ## Basic HTTP Rendering
//
//	tplSet, _ := templates.LoadTemplates("templates")
//	tplSet.Render(w, "home.html", data)
//
// ## Pre-rendering for Static Sites
//
//	html, _ := tplSet.RenderToString("about.html", data)
//	os.WriteFile("static/about.html", []byte(html), 0644)
//
// # Performance Benefits
//
//   - Pre-rendering: Generate static HTML at build time for faster serving
//   - Caching: Render once, serve many times using RenderToBytes
//   - Layout sharing: Parse layout files once, clone for each view
//
// # Development vs Production
//
//	if devMode {
//	    tplSet.Reload() // Reload templates on each request
//	}

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// TemplateSet manages a collection of templates with shared layouts.
// All templates in the set share common layout files and can be rendered
// to HTTP responses, strings, or byte buffers.
type TemplateSet struct {
	Views   map[string]*template.Template // Map of template name to parsed template
	baseDir string                        // Base directory for template reloading
}

// LoadTemplates loads all templates from a directory with shared layouts.
// It first loads all layout files from the "layout" subdirectory, then
// loads each view template and clones the layouts into it.
//
// Directory structure:
//
//	dir/
//	├── layout/*.html  (shared layouts)
//	└── *.html         (view templates)
//
// Returns an error if layouts cannot be loaded or if any view template fails to parse.
func LoadTemplates(dir string) (*TemplateSet, error) {
	// Load layouts
	layoutPattern := filepath.Join(dir, "layout", "*.html")
	layouts, err := template.ParseGlob(layoutPattern)
	if err != nil {
		log.Printf("failed to load layouts (skip): %w", err)
	}

	set := &TemplateSet{
		Views:   make(map[string]*template.Template),
		baseDir: dir,
	}

	// Load view templates
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	for _, entry := range entries {
		// Skip directories and non-html files
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if filepath.Ext(name) != ".html" {
			continue
		}

		var tpl *template.Template

		// Clone layouts and add view template
		if layouts != nil {
			tpl, err := layouts.Clone()
			if err != nil {
				return nil, fmt.Errorf("failed to clone layout for %s: %w", name, err)
			}
			_, err = tpl.ParseFiles(filepath.Join(dir, name))
			if err != nil {
				return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
			}
		} else {
			tpl, err = template.ParseFiles(filepath.Join(dir, name))
			if err != nil {
				return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
			}
		}

		set.Views[name] = tpl
	}

	return set, nil
}

// Get returns the parsed template by name.
// Returns an error if the template doesn't exist.
//
// Example:
//
//	tpl, err := tplSet.Get("email.html")
//	if err != nil {
//	    return err
//	}
//	var buf bytes.Buffer
//	tpl.Execute(&buf, data)
func (ts *TemplateSet) Get(name string) (*template.Template, error) {
	tpl, ok := ts.Views[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}
	return tpl, nil
}

// Render renders a template by name directly to an HTTP response.
// Sets the Content-Type header to "text/html; charset=utf-8" and executes the template.
//
// Use this for dynamic page rendering in HTTP handlers.
//
// Example:
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    data := map[string]interface{}{"Title": "Home"}
//	    tplSet.Render(w, "home.html", data)
//	}
func (ts *TemplateSet) Render(w http.ResponseWriter, name string, data interface{}) error {
	tpl, ok := ts.Views[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tpl.Execute(w, data)
}

// RenderWithLayout renders a template using a specific named layout.
// This allows selecting which layout to use at render time.
//
// Example:
//
//	tplSet.RenderWithLayout(w, "dashboard.html", "admin", data)
func (ts *TemplateSet) RenderWithLayout(w http.ResponseWriter, templateName, layoutName string, data interface{}) error {
	tpl, ok := ts.Views[templateName]
	if !ok {
		return fmt.Errorf("template %s not found", templateName)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tpl.ExecuteTemplate(w, layoutName, data)
}

// RenderToString renders a template to a string.
// Useful for email generation, static site generation, or debugging.
//
// Example (Email):
//
//	emailBody, _ := tplSet.RenderToString("welcome-email.html", userData)
//	sendEmail(user.Email, "Welcome!", emailBody)
//
// Example (Static Site):
//
//	html, _ := tplSet.RenderToString("about.html", nil)
//	os.WriteFile("static/about.html", []byte(html), 0644)
func (ts *TemplateSet) RenderToString(name string, data interface{}) (string, error) {
	buf, err := ts.RenderToBytes(name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderToBytes renders a template to a byte buffer.
// This is the most efficient method for pre-rendering and caching,
// as it avoids string conversion overhead.
//
// Use this when you need to:
//   - Cache rendered HTML in memory
//   - Write pre-rendered HTML to files
//   - Store rendered content in a database
//   - Send rendered HTML over network protocols
//
// Example (Caching):
//
//	buf, _ := tplSet.RenderToBytes("product.html", product)
//	cache.Set("product-"+id, buf.Bytes(), 10*time.Minute)
//
// Example (File Generation):
//
//	buf, _ := tplSet.RenderToBytes("sitemap.html", pages)
//	os.WriteFile("public/sitemap.html", buf.Bytes(), 0644)
func (ts *TemplateSet) RenderToBytes(name string, data interface{}) (*bytes.Buffer, error) {
	tpl, ok := ts.Views[name]
	if !ok {
		return nil, fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &buf, nil
}

// RenderToStringWithLayout renders a template with a specific layout to a string.
//
// Example:
//
//	html, _ := tplSet.RenderToStringWithLayout("report.html", "pdf-layout", data)
//	generatePDF(html)
func (ts *TemplateSet) RenderToStringWithLayout(templateName, layoutName string, data interface{}) (string, error) {
	buf, err := ts.RenderToBytesWithLayout(templateName, layoutName, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RenderToBytesWithLayout renders a template with a specific layout to a byte buffer.
//
// Example:
//
//	buf, _ := tplSet.RenderToBytesWithLayout("invoice.html", "print-layout", invoice)
//	cache.Set("invoice-"+id, buf.Bytes(), time.Hour)
func (ts *TemplateSet) RenderToBytesWithLayout(templateName, layoutName string, data interface{}) (*bytes.Buffer, error) {
	tpl, ok := ts.Views[templateName]
	if !ok {
		return nil, fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, layoutName, data); err != nil {
		return nil, err
	}

	return &buf, nil
}

// Has checks if a template exists in the set.
//
// Example:
//
//	if tplSet.Has("custom-page.html") {
//	    tplSet.Render(w, "custom-page.html", data)
//	} else {
//	    tplSet.Render(w, "default.html", data)
//	}
func (ts *TemplateSet) Has(name string) bool {
	_, ok := ts.Views[name]
	return ok
}

// Reload reloads all templates from disk.
// Useful in development mode to pick up template changes without restarting the server.
//
// Example (Development Mode):
//
//	func handler(w http.ResponseWriter, r *http.Request) {
//	    if devMode {
//	        tplSet.Reload()
//	    }
//	    tplSet.Render(w, "page.html", data)
//	}
//
// Note: In production, you typically load templates once at startup.
func (ts *TemplateSet) Reload() error {
	newSet, err := LoadTemplates(ts.baseDir)
	if err != nil {
		return err
	}

	ts.Views = newSet.Views
	return nil
}

type TemplateSetConfig struct {
	Folder string `json:"Folder"`
}

// DefaultTemplateSetConfig returns a default template configuration
// using the given folder path.
func DefaultTemplateSetConfig(folder string) TemplateSetConfig {
	return TemplateSetConfig{
		Folder: folder,
	}
}
