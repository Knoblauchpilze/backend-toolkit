package middleware

import (
	"net/http"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
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
	code := http.StatusInternalServerError
	if errorWithCode, ok := err.(errors.ErrorWithCode); ok {
		code = errorCodeToHttpErrorCode(errorWithCode.Code())
	}

	return &echo.HTTPError{
		Code:     code,
		Message:  err.Error(),
		Internal: err,
	}
}

func errorCodeToHttpErrorCode(code errors.ErrorCode) int {
	switch code {
	default:
		return http.StatusInternalServerError
	}
}
