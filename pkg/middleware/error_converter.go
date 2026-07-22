package middleware

import (
	stderrors "errors"
	"net/http"

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

func wrapToHttpError(err error) error {
	var httpErr *echo.HTTPError
	if stderrors.As(err, &httpErr) {
		return err
	}

	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
