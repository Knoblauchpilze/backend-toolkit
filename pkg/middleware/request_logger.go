package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func RequestLogger() echo.MiddlewareFunc {
	config := middleware.RequestLoggerConfig{
		LogHost:    true,
		LogMethod:  true,
		LogURIPath: true,
		LogStatus:  true,
		LogValuesFunc: func(c *echo.Context, values middleware.RequestLoggerValues) error {
			createRequestLog(values, c.Logger())
			return nil
		},
	}
	// Voluntarily ignoring errors
	logging, _ := config.ToMiddleware()

	return logging
}

func createRequestLog(values middleware.RequestLoggerValues, log *slog.Logger) {
	elapsed := time.Since(values.StartTime)

	log.Info(
		"Request processed",
		slog.String("method", values.Method),
		slog.String("uri", fmt.Sprintf("%s%s", values.Host, values.URIPath)),
		slog.String("duration", fmt.Sprintf("%v", elapsed)),
		slog.Int("status", values.Status),
	)
}
