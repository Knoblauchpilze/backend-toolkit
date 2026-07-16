package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

// httpError represents an HTTP error with a status code and message.
type httpError struct {
	Code    int
	Message string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Message)
}

func wrapToHttpError(err error) *httpError {
	var existing *httpError
	if stderrors.As(err, &existing) {
		return existing
	}

	code := http.StatusInternalServerError
	if errorWithCode, ok := err.(*errors.ErrorWithCode); ok {
		code = errorCodeToHttpErrorCode(errorWithCode.Code)
	}

	return &httpError{
		Code:    code,
		Message: err.Error(),
	}
}

func errorCodeToHttpErrorCode(code errors.ErrorCode) int {
	switch code {
	default:
		return http.StatusInternalServerError
	}
}
