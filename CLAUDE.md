# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Fastapify is a Go module built on top of Gin that provides automatic request/response binding and OpenAPI 3.0 (Swagger) documentation generation. It uses Go generics to create typed HTTP handlers that automatically bind JSON bodies, URI parameters, and query parameters from struct tags.

## Build & Run

```bash
# Build the example
cd examples/user-api && go build -o user-api .

# Run the example (serves on :8080, docs at /docs)
cd examples/user-api && go run .

# Run tests (none exist yet)
go test ./...
```

## Architecture

The public API surface is in `fastapify.go`, which re-exports types from internal packages and provides the `Wrapper` struct and `Bind` function.

### Internal Packages

- **`internal/router/`** — `Router` wraps `gin.Engine`, stores `RouteMeta` (method, path, body/response types), and provides HTTP method functions (GET, POST, etc.) that return a `RouteBuilder` for chaining `.Body()` and `.Response()` schema declarations. Handles `{param}` → `:param` path conversion for Gin.

- **`internal/openapi/`** — `BuildOpenAPI` generates an OpenAPI 3.0 spec from `[]RouteMeta` using reflection. `SetupSwagger` serves the spec JSON and an embedded Swagger UI HTML page. Schema generation uses struct field tags (`json`, `form`, `binding`) to derive property names, types, and required fields.

- **`internal/response/`** — Defines `ApiResponse[T]` (generic success wrapper), `ApiError` (structured error type with status code/message/code), and `HandleError` which pattern-matches on `ApiError` vs `validator.ValidationErrors` vs generic errors.

### Key Patterns

- Route registration uses a builder pattern: `app.GET("/path", handler).Body(ReqStruct{}).Response(RespStruct{})` — Body/Response calls attach `reflect.Type` metadata for OpenAPI generation.
- `Bind()` auto-detects binding source by HTTP method: GET/DELETE → query params, POST/PUT/PATCH → JSON body. URI params are always bound if the path contains `:` or `*`.
- Routes use `{param}` syntax (OpenAPI-style) which gets converted to Gin's `:param` internally.
- Tags for Swagger grouping are auto-derived from the first path segment.
