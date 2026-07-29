package middleware

import (
	"log/slog"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func RequestTracer(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			requestId, exists := rest.RequestIdFromContext(c)
			if exists {
				enrichRequestLoggerWithRequestId(c, requestId, log)
			}

			return next(c)
		}
	}
}

func enrichRequestLoggerWithRequestId(
	c *echo.Context,
	requestId string,
	defaultLogger *slog.Logger,
) {
	reqLog := rest.GetContextLogger(c)
	if reqLog == nil {
		reqLog = defaultLogger
	}

	reqLog = reqLog.With("requestId", requestId)

	rest.SetContextLogger(c, reqLog)
}
