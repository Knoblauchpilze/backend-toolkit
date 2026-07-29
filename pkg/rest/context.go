package rest

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

func GetContextLogger(c *echo.Context) *slog.Logger {
	return c.Logger()
}