package rest

import (
	"github.com/labstack/echo/v5"
)

type Routes []Route

type Route struct {
	method              string
	path                string
	handler             echo.HandlerFunc
	useResponseEnvelope bool
}

func NewRoute(method string, path string, handler echo.HandlerFunc) *Route {
	return &Route{
		method:              method,
		path:                sanitizePath(path),
		handler:             handler,
		useResponseEnvelope: true,
	}
}

func NewRawRoute(method string, path string, handler echo.HandlerFunc) *Route {
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

func (r *Route) Handler() echo.HandlerFunc {
	return r.handler
}

func (r *Route) Path() string {
	return r.path
}

func (r *Route) UseResponseEnvelope() bool {
	return r.useResponseEnvelope
}
