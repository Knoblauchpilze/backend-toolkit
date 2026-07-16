package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs timing and status information for each request after it
// has been processed. It uses the logger stored in the Gin context, which may
// have been enriched with a request ID by RequestTracer.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		log := GetContextLogger(c)
		elapsed := time.Since(start)

		log.Info(
			"Request processed",
			slog.String("method", c.Request.Method),
			slog.String("uri", fmt.Sprintf("%s%s", c.Request.Host, c.Request.URL.Path)),
			slog.String("duration", fmt.Sprintf("%v", elapsed)),
			slog.Int("status", c.Writer.Status()),
		)
	}
}
