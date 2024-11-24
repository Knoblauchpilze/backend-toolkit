package rest

import (
	"github.com/labstack/echo/v4"
)

type Route interface {
	Method() string
	Handler() echo.HandlerFunc
	Path() string
}

type Routes []Route

type routeImpl struct {
	method  string
	path    string
	handler echo.HandlerFunc
}

func NewRoute(method string, path string, handler echo.HandlerFunc) Route {
	return &routeImpl{
		method:  method,
		path:    sanitizePath(path),
		handler: handler,
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
