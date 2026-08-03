package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

func RequestLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			start := time.Now()
			defer func() {
				req := c.Request()
				log := rest.GetContextLogger(req.Context())
				status := http.StatusOK
				if echoResp, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil {
					status = echoResp.Status
				}
				createRequestLog(log, req, status, time.Since(start))
			}()
			return next(c)
		}
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
