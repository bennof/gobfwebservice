# go-bfwebservice

A lightweight, modular Go web service foundation with a strong focus on
clarity, composability, and production-ready defaults.

This project provides a **small but complete core** for building HTTP services
in Go without introducing framework lock-in or unnecessary abstractions.

---

## Philosophy

- **Explicit over magic**
- **Functional composition** for middleware
- **Standard library first**
- **Config-driven**, JSON-serializable
- **Small, reusable building blocks**
- No hidden globals (except where Go already has them, e.g. `log`)

The goal is not to replace existing frameworks, but to offer a **clean base**
for services where you want full control and understanding of the runtime.

---

## Features

### Core Infrastructure
- HTTP server with graceful shutdown
- Centralized error handling with optional HTML templates
- Global logging initialization (stdout / file)
- Unified JSON configuration

### Middleware (Composable, Functional)
- Request ID
- Logging
- Panic recovery
- Timeout
- CORS
- Rate limiting (memory-bounded, resource-safe)

All middleware follows this type:

```go
type Middleware func(http.Handler) http.Handler