package middleware

import (
	"log/slog"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v5"
)

func RequestTracer(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
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

func tryGetRequestIdHeader(resp http.ResponseWriter) (string, bool) {
	requestIds, ok := resp.Header()[requestIdHeader]
	if !ok || len(requestIds) > 1 {
		return "", false
	}

	return requestIds[0], true
}
