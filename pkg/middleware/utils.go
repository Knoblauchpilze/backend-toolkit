package middleware

import (
	stderrors "errors"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v5"
)

func formatHttpStatusCode(status int) string {
	switch {
	case status >= 500:
		return logger.FormatWithColor(status, logger.Red)
	case status >= 400:
		return logger.FormatWithColor(status, logger.Yellow)
	case status >= 300:
		return logger.FormatWithColor(status, logger.Cyan)
	default:
		return logger.FormatWithColor(status, logger.Green)
	}
}

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
