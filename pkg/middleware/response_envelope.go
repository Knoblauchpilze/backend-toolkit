package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const requestIdHeader = "X-Request-Id"

func ResponseEnvelope() echo.MiddlewareFunc {
	config := middleware.RequestIDConfig{
		Generator: func() string {
			return uuid.New().String()
		},
		RequestIDHandler: func(c *echo.Context, requestId string) {
			echoResp, err := echo.UnwrapResponse(c.Response())
			if err == nil {
				rw := rest.NewResponseEnvelopeWriter(echoResp.ResponseWriter, requestId)
				echoResp.ResponseWriter = rw
			}
		},
		TargetHeader: requestIdHeader,
	}

	return middleware.RequestIDWithConfig(config)
}
