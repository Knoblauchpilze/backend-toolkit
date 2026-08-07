package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

func RequestLogger() NewHandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			log := rest.GetContextLogger(c.Request.Context())
			createRequestLog(log, c.Request, c.Writer.Status(), time.Since(start))
		}()
		c.Next()
	}
}

func createRequestLog(log *slog.Logger, req *http.Request, status int, elapsed time.Duration) {
	log.Info(
		"Request processed",
		slog.String("method", req.Method),
		slog.String("uri", fmt.Sprintf("%s%s", req.Host, req.URL.Path)),
		slog.String("duration", fmt.Sprintf("%v", elapsed)),
		slog.Int("status", status),
	)
}
