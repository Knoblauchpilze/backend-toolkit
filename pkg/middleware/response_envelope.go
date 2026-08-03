package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

const (
	defaultRequestId = "unknown"
)

func ResponseEnvelope() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestId, ok := rest.RequestIdFromContext(c.Request().Context())
			if !ok {
				requestId = defaultRequestId
			}

			echoResp, err := echo.UnwrapResponse(c.Response())
			if err == nil {
				rw := rest.NewResponseEnvelopeWriter(
					echoResp.ResponseWriter,
					requestId,
				)
				echoResp.ResponseWriter = rw
			}

			return next(c)
		}
	}
}
