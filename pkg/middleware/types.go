package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v5"
)

type (
	NewHandlerFunc = gin.HandlerFunc

	HandlerFunc    = echo.HandlerFunc
	MiddlewareFunc = echo.MiddlewareFunc
)
