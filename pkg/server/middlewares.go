package server

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v4"
)

func buildMiddlewaresForRoute(route rest.Route, log echo.Logger) []echo.MiddlewareFunc {
	var out []echo.MiddlewareFunc

	if route.UseResponseEnvelope() {
		out = append(out, middleware.ResponseEnvelope())
	}

	out = append(
		out,
		middleware.RequestTracer(log),
		middleware.ErrorConverter(),
		middleware.Recover(),
	)

	return out
}
