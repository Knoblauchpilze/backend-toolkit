package rest

import (
	"github.com/labstack/echo/v4"
)

type Route interface {
	Method() string
	Handler() echo.HandlerFunc
	Path() string
	UseResponseEnvelope() bool
}

type Routes []Route

type routeImpl struct {
	method              string
	path                string
	handler             echo.HandlerFunc
	useResponseEnvelope bool
}

func NewRoute(method string, path string, handler echo.HandlerFunc) Route {
	return &routeImpl{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: true,
	}
}

func NewRawRoute(method string, path string, handler echo.HandlerFunc) Route {
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

func (r *routeImpl) Handler() echo.HandlerFunc {
	return r.handler
}

func (r *routeImpl) Path() string {
	return r.path
}

func (r *routeImpl) UseResponseEnvelope() bool {
	return r.useResponseEnvelope
}
