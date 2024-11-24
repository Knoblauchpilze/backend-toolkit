package middleware

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RequestLogger() echo.MiddlewareFunc {
	config := middleware.RequestLoggerConfig{
		LogHost:    true,
		LogMethod:  true,
		LogURIPath: true,
		LogStatus:  true,
		LogError:   true,
		LogValuesFunc: func(c echo.Context, values middleware.RequestLoggerValues) error {
			c.Logger().Infof(createRequestLog(values))
			return nil
		},
	}
	// Voluntarily ignoring errors
	logging, _ := config.ToMiddleware()

	return logging
}

func createRequestLog(values middleware.RequestLoggerValues) string {
	var out string

	elapsed := time.Since(values.StartTime)

	out += fmt.Sprintf("%v", values.Method)
	out += fmt.Sprintf(" %v%v", values.Host, values.URIPath)
	out += fmt.Sprintf(" processed in %v", elapsed)
	out += fmt.Sprintf(" -> %s", formatHttpStatusCode(values.Status))

	return out
}
