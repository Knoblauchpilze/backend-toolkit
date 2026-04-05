package middleware

import (
	"github.com/labstack/echo/v5"
)

func ErrorConverter() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if err := next(c); err != nil {
				return wrapToHttpError(err)
			}

			return nil
		}
	}
}
