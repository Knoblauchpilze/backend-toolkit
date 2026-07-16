package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	requestIdHeader = "X-Request-Id"
	loggerKey       = "toolkit_logger"
)

// GetContextLogger retrieves the logger stored in the Gin context by this
// toolkit. Falls back to slog.Default() if no logger has been set.
func GetContextLogger(c *gin.Context) *slog.Logger {
	if val, exists := c.Get(loggerKey); exists {
		if log, ok := val.(*slog.Logger); ok {
			return log
		}
	}
	return slog.Default()
}

// SetContextLogger stores a logger in the Gin context for use by other
// middlewares and handlers within this toolkit.
func SetContextLogger(c *gin.Context, log *slog.Logger) {
	c.Set(loggerKey, log)
}

// RequestTracer reads the X-Request-Id response header (set by the
// ResponseEnvelope middleware) and attaches it as an attribute on the context
// logger so all subsequent log lines for that request carry the request ID.
func RequestTracer() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId, exists := tryGetRequestIdHeader(c.Writer)
		if exists {
			log := GetContextLogger(c)
			SetContextLogger(c, log.With("requestId", requestId))
		}

		c.Next()
	}
}

func tryGetRequestIdHeader(resp http.ResponseWriter) (string, bool) {
	requestIds, ok := resp.Header()[requestIdHeader]
	if !ok || len(requestIds) > 1 {
		return "", false
	}

	return requestIds[0], true
}
