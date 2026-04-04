package server

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func buildMiddlewaresForRoute(route rest.Route) []echo.MiddlewareFunc {
	var out []echo.MiddlewareFunc

	if route.UseResponseEnvelope() {
		out = append(out, middleware.ResponseEnvelope())
	}

	out = append(
		out,
		middleware.RequestTracer(),
		middleware.ErrorConverter(),
		middleware.Recover(),
	)

	return out
}
