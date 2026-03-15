package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	Register(w, http.MethodGet, path, handler, middleware...)
}

func Post[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	Register(w, http.MethodPost, path, handler, middleware...)
}

func Put[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	Register(w, http.MethodPut, path, handler, middleware...)
}

func Patch[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	Register(w, http.MethodPatch, path, handler, middleware...)
}

func Delete[Req any, Res any](w *Wrapper, path string, handler func(*gin.Context, *Req) (*Res, error), middleware ...gin.HandlerFunc) {
	Register(w, http.MethodDelete, path, handler, middleware...)
}
