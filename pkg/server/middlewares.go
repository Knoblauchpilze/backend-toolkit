package server

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func buildMiddlewaresForRoute(route *rest.Route) []echo.MiddlewareFunc {
	out := []echo.MiddlewareFunc{}

	if route.UseResponseEnvelope() {
		out = append(out, middleware.ResponseEnvelope())
	}

	return out
}
