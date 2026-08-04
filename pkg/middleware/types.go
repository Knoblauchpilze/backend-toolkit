package middleware

import (
	"github.com/labstack/echo/v5"
)

type (
	HandlerFunc    = echo.HandlerFunc
	MiddlewareFunc = echo.MiddlewareFunc
)
