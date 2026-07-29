package rest

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

const (
	requestIdContextKey = "request_id"
)

func GetContextLogger(c *echo.Context) *slog.Logger {
	return c.Logger()
}

func SetContextRequestId(c *echo.Context, requestId string) {
	c.Set(requestIdContextKey, requestId)
}

func RequestIdFromContext(c *echo.Context) (string, bool) {
	val := c.Get(requestIdContextKey)
	s, ok := val.(string)
	return s, ok && s != ""
}
