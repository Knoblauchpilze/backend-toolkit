package middleware

import (
	stderrors "errors"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/labstack/echo/v5"
)

func wrapToHttpError(err error) error {
	var httpErr *echo.HTTPError
	if stderrors.As(err, &httpErr) {
		return err
	}

	code := http.StatusInternalServerError
	if errorWithCode, ok := err.(errors.ErrorWithCode); ok {
		code = errorCodeToHttpErrorCode(errorWithCode.Code())
	}

	return echo.NewHTTPError(code, err.Error())
}

func errorCodeToHttpErrorCode(code errors.ErrorCode) int {
	switch code {
	default:
		return http.StatusInternalServerError
	}
}
