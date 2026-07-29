package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func RequestTracer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestId, exists := rest.RequestIdFromContext(c)
			if exists {
				c.SetLogger(c.Logger().With("requestId", requestId))
			}

			return next(c)
		}
	}
}
