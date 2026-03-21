package fastapify

import (
	"context"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	finternal "github.com/sharathcx/fastapify/internal"
)

// --- Core Types ---

// HandlerFunc is the fastapify handler signature.
// Return finternal.ApiResponse[T] for success or *finternal.ApiError for errors.
type HandlerFunc func(c *gin.Context) any

type Wrapper struct {
	Engine *gin.Engine
	Routes []RouteMeta
}

type RouteMeta struct {
	Method       string
	Path         string
	Tag          string
	BodyType     reflect.Type
	ParamsType   reflect.Type
	ResponseType reflect.Type
}

type RouteBuilder struct {
	wrapper *Wrapper
	index   int
}

func New(r *gin.Engine) *Wrapper {
	return &Wrapper{Engine: r}
}

// --- Bind Helper ---

// Req retrieves the automatically bound and validated request data from the context.
func Req[T any](c *gin.Context) *T {
	val, exists := c.Get("fastapify_body")
	if !exists {
		return new(T)
	}
	return val.(*T)
}

// Params retrieves the automatically bound and validated URI params from the context.
func Params[T any](c *gin.Context) *T {
	val, exists := c.Get("fastapify_params")
	if !exists {
		return new(T)
	}
	return val.(*T)
}

// Bind validates and binds the request into the given struct.
// Automatically:
//  1. Binds URI parameters (and makes them immutable)
//  2. Detects HTTP method for Body vs Query binding
//  3. Sends 400 error response on failure
//
// Returns true on success, false on failure.
func Bind(c *gin.Context, req any) bool {
	// Step 1: Bind URI params and save their values for protection
	uriValues := make(map[int]reflect.Value)
	reqVal := reflect.ValueOf(req).Elem()
	if reqVal.Kind() == reflect.Struct {
		_ = c.ShouldBindUri(req)

		// Snapshot URI-tagged fields
		reqType := reqVal.Type()
		for i := 0; i < reqType.NumField(); i++ {
			if reqType.Field(i).Tag.Get("uri") != "" {
				uriValues[i] = reflect.ValueOf(reqVal.Field(i).Interface())
			}
		}
	}

	// Step 2: Bind Body or Query
	var err error
	switch c.Request.Method {
	case http.MethodGet, http.MethodDelete:
		err = c.ShouldBindQuery(req)
	default:
		err = c.ShouldBindJSON(req)
	}

	if err != nil && err.Error() != "EOF" {
		statusCode, response := finternal.HandleError(err)
		c.JSON(statusCode, response)
		return false
	}

	// Step 3: Restore URI values (Protection against body override)
	for i, val := range uriValues {
		reqVal.Field(i).Set(val)
	}

	return true
}

// --- Route Registration ---

func (w *Wrapper) handle(method, path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	ginPath := toGinPath(path)

	routeIdx := len(w.Routes)
	w.Routes = append(w.Routes, RouteMeta{
		Method: method,
		Path:   path,
		Tag:    deriveTag(path),
	})
	meta := &w.Routes[routeIdx]

	// Automatic Validation Middleware
	autoValidator := func(c *gin.Context) {
		if meta.ParamsType != nil {
			params := reflect.New(meta.ParamsType).Interface()
			if err := c.ShouldBindUri(params); err != nil {
				statusCode, response := finternal.HandleError(err)
				c.JSON(statusCode, response)
				c.Abort()
				return
			}
			c.Set("fastapify_params", params)
		}
		if meta.BodyType != nil {
			req := reflect.New(meta.BodyType).Interface()
			if !Bind(c, req) {
				c.Abort()
				return
			}
			c.Set("fastapify_body", req)
		}
		c.Next()
	}

	// Wraps HandlerFunc to automatically write the JSON response
	ginHandler := func(c *gin.Context) {
		result := handler(c)
		if c.Writer.Written() {
			return
		}
		if result == nil {
			return
		}
		switch v := result.(type) {
		case *finternal.ApiError:
			statusCode, response := finternal.HandleError(v)
			c.JSON(statusCode, response)
		default:
			c.JSON(http.StatusOK, result)
		}
	}

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+2)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, autoValidator)
	handlers = append(handlers, ginHandler)
	w.Engine.Handle(method, ginPath, handlers...)

	return &RouteBuilder{wrapper: w, index: routeIdx}
}

func (w *Wrapper) GET(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return w.handle(http.MethodGet, path, handler, middleware...)
}

func (w *Wrapper) POST(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return w.handle(http.MethodPost, path, handler, middleware...)
}

func (w *Wrapper) PUT(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return w.handle(http.MethodPut, path, handler, middleware...)
}

func (w *Wrapper) PATCH(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return w.handle(http.MethodPatch, path, handler, middleware...)
}

func (w *Wrapper) DELETE(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return w.handle(http.MethodDelete, path, handler, middleware...)
}

// --- RouteBuilder Chainable Methods ---

// Params sets the URI params schema type for validation and Swagger documentation.
func (rb *RouteBuilder) Params(schema any) *RouteBuilder {
	rb.wrapper.Routes[rb.index].ParamsType = reflect.TypeOf(schema)
	return rb
}

// Body sets the request body schema type for Swagger documentation.
func (rb *RouteBuilder) Body(schema any) *RouteBuilder {
	rb.wrapper.Routes[rb.index].BodyType = reflect.TypeOf(schema)
	return rb
}

// Response sets the response schema type for Swagger documentation.
func (rb *RouteBuilder) Response(schema any) *RouteBuilder {
	rb.wrapper.Routes[rb.index].ResponseType = reflect.TypeOf(schema)
	return rb
}

// --- Group ---

type Group struct {
	wrapper *Wrapper
	prefix  string
}

func (w *Wrapper) Group(prefix string) *Group {
	return &Group{wrapper: w, prefix: prefix}
}

func (g *Group) GET(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return g.wrapper.handle(http.MethodGet, g.prefix+path, handler, middleware...)
}

func (g *Group) POST(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return g.wrapper.handle(http.MethodPost, g.prefix+path, handler, middleware...)
}

func (g *Group) PUT(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return g.wrapper.handle(http.MethodPut, g.prefix+path, handler, middleware...)
}

func (g *Group) PATCH(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return g.wrapper.handle(http.MethodPatch, g.prefix+path, handler, middleware...)
}

func (g *Group) DELETE(path string, handler HandlerFunc, middleware ...gin.HandlerFunc) *RouteBuilder {
	return g.wrapper.handle(http.MethodDelete, g.prefix+path, handler, middleware...)
}

// --- Helpers ---

func toGinPath(path string) string {
	ginPath := strings.ReplaceAll(path, "{", ":")
	return strings.ReplaceAll(ginPath, "}", "")
}

func deriveTag(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return strings.Title(trimmed[:idx])
	}
	return strings.Title(trimmed)
}

var paramPattern = regexp.MustCompile(`\{(\w+)\}`)

func extractParamNames(path string) []string {
	matches := paramPattern.FindAllStringSubmatch(path, -1)
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, m[1])
	}
	return names
}

// --- Middleware ---

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// --- Re-exported Types & Functions ---

// ApiError is a structured error type with status code, message, and error code.
type ApiError = finternal.ApiError

// ApiResponse is a generic success response wrapper.
type ApiResponse[T any] = finternal.ApiResponse[T]

// ValidationErrorDetail holds a single field validation error.
type ValidationErrorDetail = finternal.ValidationErrorDetail

// Error constants
const (
	ErrValidation       = finternal.ErrValidation
	ErrBadRequest       = finternal.ErrBadRequest
	ErrUnauthorized     = finternal.ErrUnauthorized
	ErrForbidden        = finternal.ErrForbidden
	ErrNotFound         = finternal.ErrNotFound
	ErrResourceConflict = finternal.ErrResourceConflict
	ErrUploadError      = finternal.ErrUploadError
	ErrInternalError    = finternal.ErrInternalError
)

// Error constructors
var (
	NewApiError   = finternal.NewApiError
	NotFound      = finternal.NotFound
	BadRequest    = finternal.BadRequest
	Unauthorized  = finternal.Unauthorized
	Forbidden     = finternal.Forbidden
	Conflict      = finternal.Conflict
	InternalError = finternal.InternalError
	HandleError   = finternal.HandleError
)

// NewApiResponse creates a new success response.
func NewApiResponse[T any](statusCode int, data T, message string) ApiResponse[T] {
	return finternal.NewApiResponse(statusCode, data, message)
}
