package fastapify

import (
	"github.com/gin-gonic/gin"

	"github.com/sharathcx/fastapify/internal/bind"
	"github.com/sharathcx/fastapify/internal/openapi"
	"github.com/sharathcx/fastapify/internal/response"
	"github.com/sharathcx/fastapify/internal/router"
)

// Re-export types for public API
type ApiError = response.ApiError
type RouteMeta = router.RouteMeta

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

// Re-export middleware
var TimeoutMiddleware = router.TimeoutMiddleware

// Wrapper is the main Fastapify application instance.
type Wrapper struct {
	*router.Router
}

// New creates a new Fastapify application instance.
func New(r *gin.Engine) *Wrapper {
	return &Wrapper{
		Router: router.New(r),
	}
}

// Req retrieves the automatically bound and validated request data from the context.
func Req[T any](c *gin.Context) *T {
	val, exists := c.Get("fastapify_body")
	if !exists {
		return new(T)
	}
	return val.(*T)
}

// Bind validates and binds the request into the given struct.
func Bind(c *gin.Context, req any) bool {
	return bind.Bind(c, req)
}

// SetupSwagger initializes the Swagger UI and OpenAPI JSON endpoint.
func (w *Wrapper) SetupSwagger(jsonPath string) {
	openapi.SetupSwagger(w.Engine, w.Routes, jsonPath)
}
