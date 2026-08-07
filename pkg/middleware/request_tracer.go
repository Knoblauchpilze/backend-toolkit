package middleware

import (
	"context"
	"log/slog"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

func RequestTracer(log *slog.Logger) NewHandlerFunc {
	return func(c *gin.Context) {
		requestId, exists := rest.RequestIdFromContext(c.Request.Context())
		if exists {
			ctx := enrichRequestLoggerWithRequestId(c.Request.Context(), requestId, log)
			c.Request = c.Request.WithContext(ctx)
		}

		c.Next()
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
