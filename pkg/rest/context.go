package rest

import (
	"log/slog"

	"github.com/labstack/echo/v5"
)

const (
	requestIdContextKey = "request_id"
	loggerContextKey    = "toolkit_logger"
)

func GetContextLogger(c *echo.Context) *slog.Logger {
	val := c.Get(loggerContextKey)
	if log, ok := val.(*slog.Logger); ok && log != nil {
		return log
	}

	return nil
}

func SetContextLogger(c *echo.Context, log *slog.Logger) {
	c.Set(loggerContextKey, log)
}

func SetContextRequestId(c *echo.Context, requestId string) {
	c.Set(requestIdContextKey, requestId)
}

func RequestIdFromContext(c *echo.Context) (string, bool) {
	val := c.Get(requestIdContextKey)
	s, ok := val.(string)
	return s, ok && s != ""
}
