package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const requestIdHeader = "X-Request-Id"
const requestIdContextKey = "request_id"

func RequestId() echo.MiddlewareFunc {
	cfg := middleware.RequestIDConfig{
		Generator:    func() string { return uuid.New().String() },
		TargetHeader: requestIdHeader,
		RequestIDHandler: func(c *echo.Context, requestId string) {
			c.Set(requestIdContextKey, requestId)
		},
	}

	return middleware.RequestIDWithConfig(cfg)
}

func RequestIdFromContext(c *echo.Context) (string, bool) {
	val := c.Get(requestIdContextKey)
	s, ok := val.(string)
	return s, ok && s != ""
}
