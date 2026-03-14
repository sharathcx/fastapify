package fastapify

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sharathcx/fastapify/utils"
)

type Wrapper struct {
	Engine *gin.Engine
	Routes []RouteMeta
}

type RouteMeta struct {
	Method string
	Path   string
	Tag    string
	Input  reflect.Type
	Output reflect.Type
}

func New(r *gin.Engine) *Wrapper {
	return &Wrapper{Engine: r}
}

func Get[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	register(w, http.MethodGet, path, handler, middleware...)
}

func Post[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	register(w, http.MethodPost, path, handler, middleware...)
}

func Put[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	register(w, http.MethodPut, path, handler, middleware...)
}

func Patch[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	register(w, http.MethodPatch, path, handler, middleware...)
}

func Delete[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	register(w, http.MethodDelete, path, handler, middleware...)
}

func deriveTag(path string) string {
	trimmed := strings.TrimPrefix(path, "/")
	if idx := strings.Index(trimmed, "/"); idx != -1 {
		return strings.Title(trimmed[:idx])
	}
	return strings.Title(trimmed)
}

func register[Req any, Res any](w *Wrapper, method, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	w.Routes = append(w.Routes, RouteMeta{
		Method: method,
		Path:   path,
		Tag:    deriveTag(path),
		Input:  reflect.TypeOf(*new(Req)),
		Output: reflect.TypeOf(*new(Res)),
	})

	ginPath := strings.ReplaceAll(path, "{", ":")
	ginPath = strings.ReplaceAll(ginPath, "}", "")

	hasUriParams := strings.Contains(path, "{")

	handlers := make([]gin.HandlerFunc, 0, len(middleware)+1)
	handlers = append(handlers, middleware...)
	handlers = append(handlers, func(c *gin.Context) {
		req := new(Req)

		if hasUriParams {
			if err := c.ShouldBindUri(req); err != nil {
				statusCode, response := utils.HandleError(err)
				c.JSON(statusCode, response)
				return
			}
		}

		if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
			if err := c.ShouldBindJSON(req); err != nil {
				statusCode, response := utils.HandleError(err)
				c.JSON(statusCode, response)
				return
			}
		}

		// 3. Business logic invocation
		res, err := handler(c, req)
		if err != nil {
			statusCode, response := utils.HandleError(err)
			c.JSON(statusCode, response)
			return
		}

		if res != nil {
			c.JSON(http.StatusOK, utils.NewApiResponse(http.StatusOK, res, "Success"))
		} else {
			c.JSON(http.StatusOK, utils.NewApiResponse[*Res](http.StatusOK, nil, "Success"))
		}
	})

	w.Engine.Handle(method, ginPath, handlers...)
}
