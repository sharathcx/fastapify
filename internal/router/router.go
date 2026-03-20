package router

import (
	"context"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/sharathcx/fastapify/internal/bind"
)

// RouteMeta holds metadata about a registered route for OpenAPI generation.
type RouteMeta struct {
	Method       string
	Path         string
	Tag          string
	BodyType     reflect.Type
	ResponseType reflect.Type
}

// Router wraps a gin.Engine and tracks route metadata.
type Router struct {
	Engine *gin.Engine
	Routes []RouteMeta
}

// RouteBuilder provides a chainable API for setting route metadata.
type RouteBuilder struct {
	router *Router
	index  int
}

// New creates a new Router wrapping the given gin.Engine.
func New(r *gin.Engine) *Router {
	return &Router{Engine: r}
}

func (r *Router) handle(method, path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	ginPath := ToGinPath(path)

	routeIdx := len(r.Routes)
	r.Routes = append(r.Routes, RouteMeta{
		Method: method,
		Path:   path,
		Tag:    deriveTag(path),
	})
	meta := &r.Routes[routeIdx]

	// Automatic Validation Middleware
	autoValidator := func(c *gin.Context) {
		if meta.BodyType != nil {
			req := reflect.New(meta.BodyType).Interface()
			if !bind.Bind(c, req) {
				c.Abort()
				return
			}
			c.Set("fastapify_body", req)
		}
		c.Next()
	}

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+2)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, autoValidator)
	handlers = append(handlers, handler)
	r.Engine.Handle(method, ginPath, handlers...)

	return &RouteBuilder{router: r, index: routeIdx}
}

func (r *Router) GET(path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return r.handle(http.MethodGet, path, handler, middleware...)
}

func (r *Router) POST(path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return r.handle(http.MethodPost, path, handler, middleware...)
}

func (r *Router) PUT(path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return r.handle(http.MethodPut, path, handler, middleware...)
}

func (r *Router) PATCH(path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return r.handle(http.MethodPatch, path, handler, middleware...)
}

func (r *Router) DELETE(path string, handler gin.HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return r.handle(http.MethodDelete, path, handler, middleware...)
}

// Body sets the request body schema type for Swagger documentation.
func (rb *RouteBuilder) Body(schema any) *RouteBuilder {
	rb.router.Routes[rb.index].BodyType = reflect.TypeOf(schema)
	return rb
}

// Response sets the response schema type for Swagger documentation.
func (rb *RouteBuilder) Response(schema any) *RouteBuilder {
	rb.router.Routes[rb.index].ResponseType = reflect.TypeOf(schema)
	return rb
}

// --- Path Helpers ---

// ToGinPath converts OpenAPI-style {param} paths to Gin-style :param paths.
func ToGinPath(path string) string {
	ginPath := strings.ReplaceAll(path, "{", ":")
	return strings.ReplaceAll(ginPath, "}", "")
}

var paramPattern = regexp.MustCompile(`\{(\w+)\}`)

// ExtractParamNames returns parameter names from an OpenAPI-style path.
func ExtractParamNames(path string) []string {
	matches := paramPattern.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m[1])
	}
	return names
}

func deriveTag(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return strings.Title(trimmed[:idx])
	}
	return strings.Title(trimmed)
}

// --- Middleware ---

// TimeoutMiddleware returns middleware that sets a request context timeout.
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
