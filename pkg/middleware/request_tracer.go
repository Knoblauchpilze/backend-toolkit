package middleware

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

func RequestTracer() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestId, exists := tryGetRequestIdHeader(c.Response())
			if exists {
				c.SetLogger(c.Logger().With("requestId", requestId))
			}

			return next(c)
		}
	}
}

func tryGetRequestIdHeader(resp http.ResponseWriter) (string, bool) {
	requestIds, ok := resp.Header()[requestIdHeader]
	if !ok || len(requestIds) > 1 {
		return "", false
	}

	return requestIds[0], true
}
