package server

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/middleware"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

func buildMiddlewaresForRoute(route rest.Route) []gin.HandlerFunc {
	var out []gin.HandlerFunc

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
