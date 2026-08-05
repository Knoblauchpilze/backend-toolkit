package rest

import (
	"github.com/gin-gonic/gin"
)

type Routes []Route

type Route struct {
	method              string
	path                string
	handler             gin.HandlerFunc
	useResponseEnvelope bool
}

func NewRoute(method string, path string, handler gin.HandlerFunc) *Route {
	return &Route{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: true,
	}
}

func NewRawRoute(method string, path string, handler gin.HandlerFunc) *Route {
	return &Route{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: false,
	}
}

func (r *Route) Method() string {
	return r.method
}

func (r *Route) Handler() gin.HandlerFunc {
	return r.handler
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) UseResponseEnvelope() bool {
	return r.useResponseEnvelope
}
