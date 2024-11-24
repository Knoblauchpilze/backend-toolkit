package middleware

import (
	"github.com/KnoblauchPilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const requestIdHeader = "X-Request-Id"

func ResponseEnvelope() echo.MiddlewareFunc {
	config := middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
		RequestIDHandler: func(c echo.Context, requestId string) {
			rw := rest.NewResponseEnvelopeWriter(c.Response().Writer, requestId)
			c.Response().Writer = rw
		},
		TargetHeader: requestIdHeader,
	}

	return middleware.RequestIDWithConfig(config)
}
