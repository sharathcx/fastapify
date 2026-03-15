package fastapify

import (
	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify/internal/openapi"
	"github.com/sharathcx/fastapify/internal/response"
	"github.com/sharathcx/fastapify/internal/router"
)

// Re-export error types and functions for convenience
type ApiError = response.ApiError

var NewApiError = response.NewApiError

const (
	ErrValidation       = response.ErrValidation
	ErrBadRequest       = response.ErrBadRequest
	ErrUnauthorized     = response.ErrUnauthorized
	ErrForbidden        = response.ErrForbidden
	ErrNotFound         = response.ErrNotFound
	ErrResourceConflict = response.ErrResourceConflict
	ErrUploadError      = response.ErrUploadError
	ErrInternalError    = response.ErrInternalError
)

// Wrapper is the main Fastapify application instance.
type Wrapper struct {
	*router.Wrapper
}

// New creates a new Fastapify application instance.
func New(r *gin.Engine) *Wrapper {
	return &Wrapper{
		Wrapper: router.New(r),
	}
}

// Get registers a new GET route.
func Get[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	router.Get(w.Wrapper, path, handler, middleware...)
}

// Post registers a new POST route.
func Post[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	router.Post(w.Wrapper, path, handler, middleware...)
}

// Put registers a new PUT route.
func Put[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	router.Put(w.Wrapper, path, handler, middleware...)
}

// Patch registers a new PATCH route.
func Patch[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	router.Patch(w.Wrapper, path, handler, middleware...)
}

// Delete registers a new DELETE route.
func Delete[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	router.Delete(w.Wrapper, path, handler, middleware...)
}

// SetupSwagger initializes the Swagger UI and OpenAPI JSON endpoint.
func (w *Wrapper) SetupSwagger(jsonPath string) {
	openapi.SetupSwagger(w.Wrapper, jsonPath)
}