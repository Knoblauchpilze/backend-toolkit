package middleware

import (
	"context"
	"log/slog"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func RequestTracer(log *slog.Logger) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			ctx := req.Context()
			requestId, exists := rest.RequestIdFromContext(ctx)
			if exists {
				ctx = enrichRequestLoggerWithRequestId(ctx, requestId, log)
				c.SetRequest(req.WithContext(ctx))
			}

			return next(c)
		}
	}
}

func enrichRequestLoggerWithRequestId(
	ctx context.Context,
	requestId string,
	defaultLogger *slog.Logger,
) context.Context {
	reqLog := rest.GetContextLogger(ctx)
	if reqLog == nil {
		reqLog = defaultLogger
	}

	reqLog = reqLog.With("requestId", requestId)

	return rest.WithContextLogger(ctx, reqLog)
}
