package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v5"
)

type (
	// TODO: Remove the New prefix
	NewHandlerFunc = gin.HandlerFunc

	HandlerFunc    = echo.HandlerFunc
	MiddlewareFunc = echo.MiddlewareFunc
)
