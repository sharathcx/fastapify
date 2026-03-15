package router

import (
	"reflect"

	"github.com/gin-gonic/gin"
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
