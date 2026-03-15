package router

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify/internal/response"
)

func deriveTag(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return strings.Title(trimmed[:idx])
	}
	return strings.Title(trimmed)
}

func Register[Req any, Res any](w *Wrapper, method, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	// 1. Normalize for Swagger (OpenAPI uses {param})
	swaggerPath := path
	if strings.Contains(path, ":") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, ":") {
				parts[i] = "{" + part[1:] + "}"
			}
		}
		swaggerPath = strings.Join(parts, "/")
	}

	w.Routes = append(w.Routes, RouteMeta{
		Method: method,
		Path:   swaggerPath,
		Tag:    deriveTag(path),
		Input:  reflect.TypeOf(*new(Req)),
		Output: reflect.TypeOf(*new(Res)),
	})

	// 2. Normalize for Gin (uses :param)
	ginPath := strings.ReplaceAll(path, "{", ":")
	ginPath = strings.ReplaceAll(ginPath, "}", "")

	hasUriParams := strings.Contains(path, "{") || strings.Contains(path, ":")

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, func(c *gin.Context) {
		req := new(Req)

		if hasUriParams {
			if err := c.ShouldBindUri(req); err != nil {
				statusCode, res := response.HandleError(err)
				c.JSON(statusCode, res)
				return
			}
		}

		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			if err := c.ShouldBindJSON(req); err != nil {
				statusCode, res := response.HandleError(err)
				c.JSON(statusCode, res)
				return
			}
		}

		// 3. Business logic invocation
		res, err := handler(c, req)
		if err != nil {
			statusCode, response := response.HandleError(err)
			c.JSON(statusCode, response)
			return
		}

		if res != nil {
			c.JSON(http.StatusOK, response.NewApiResponse(http.StatusOK, res, "Success"))
		} else {
			c.JSON(http.StatusOK, response.NewApiResponse[*Res](http.StatusOK, nil, "Success"))
		}
	})

	w.Engine.Handle(method, ginPath, handlers...)
}
