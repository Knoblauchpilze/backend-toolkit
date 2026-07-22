package middleware

import (
	"net/http"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ErrorConverter(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableHandler(ErrorConverter)

		err := callable(ctx)

		assert.Nil(t, err)
		assert.True(t, *called)
	})

	t.Run("wraps unknown error into http error", func(t *testing.T) {
		next := createErrorHandler(errSample)
		middleware := ErrorConverter()
		callable := middleware(next)
		ctx, _ := generateTestEchoContext()

		err := callable(ctx)

		assertIsHttpErrorWithMessageAndCode(t, err, "some error", http.StatusInternalServerError)
	})

	t.Run("wraps error with code into http error", func(t *testing.T) {
		next := createErrorHandler(ErrUncaughtPanic)
		middleware := ErrorConverter()
		callable := middleware(next)
		ctx, _ := generateTestEchoContext()

		err := callable(ctx)

		assertIsHttpErrorWithMessageAndCode(t, err, "an unexpected error occurred. Code: 400", http.StatusInternalServerError)
	})

	t.Run("wraps error with code and cause into http error", func(t *testing.T) {
		errWithCause := errors.WrapCode(errSample, errUncaughtPanic)
		next := createErrorHandler(errWithCause)
		middleware := ErrorConverter()
		callable := middleware(next)
		ctx, _ := generateTestEchoContext()

		err := callable(ctx)

		assertIsHttpErrorWithMessageAndCode(t, err, "an unexpected error occurred. Code: 400 (cause: some error)", http.StatusInternalServerError)
	})
}

func createErrorHandler(err error) echo.HandlerFunc {
	handler := func(c *echo.Context) error {
		return err
	}

	return handler
}
