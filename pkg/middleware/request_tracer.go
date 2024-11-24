package middleware

import (
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
)

func RequestTracer(log echo.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestId, exists := tryGetRequestIdHeader(c.Response())
			if exists {
				if requestLog, err := logger.Duplicate(log); err == nil {
					requestLog.SetPrefix(requestId)
					c.SetLogger(requestLog)
				}
			}

			return next(c)
		}
	}
}

func tryGetRequestIdHeader(resp *echo.Response) (string, bool) {
	requestIds, ok := resp.Header()[requestIdHeader]
	if !ok || len(requestIds) > 1 {
		return "", false
	}

	return requestIds[0], true
}
