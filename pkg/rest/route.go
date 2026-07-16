package rest

import (
	"github.com/gin-gonic/gin"
)

// HandlerFunc is the function signature for route handlers in this toolkit.
// It accepts a Gin context and returns an error, allowing handlers to signal
// failures without writing the response directly.
type HandlerFunc func(*gin.Context) error

type Route interface {
	Method() string
	Handler() HandlerFunc
	Path() string
	UseResponseEnvelope() bool
}

type Routes []Route

type routeImpl struct {
	method              string
	path                string
	handler             HandlerFunc
	useResponseEnvelope bool
}

func NewRoute(method string, path string, handler HandlerFunc) Route {
	return &routeImpl{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: true,
	}
}

func NewRawRoute(method string, path string, handler HandlerFunc) Route {
	return &routeImpl{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: false,
	}
}

func (r *routeImpl) Method() string {
	return r.method
}

func (r *routeImpl) Handler() HandlerFunc {
	return r.handler
}

func (r *routeImpl) Path() string {
	return r.path
}

func (r *routeImpl) UseResponseEnvelope() bool {
	return r.useResponseEnvelope
}
