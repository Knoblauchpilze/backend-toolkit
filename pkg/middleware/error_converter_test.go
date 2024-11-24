package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/errors"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ErrorConverter_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableHandler(ErrorConverter)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_ErrorConverter_WrapsUnknownErrorIntoHttpError(t *testing.T) {
	next := createErrorHandler(fmt.Errorf("some error"))
	middleware := ErrorConverter()
	callable := middleware(next)
	ctx, rw := generateTestEchoContext()

	err := callable(ctx)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	expected := `
	{
		"message": "some error"
	}`
	assert.JSONEq(t, expected, rw.Body.String())
}

func TestUnit_ErrorConverter_WrapsErrorWithCodeIntoHttpError(t *testing.T) {
	next := createErrorHandler(errors.NewCode(UncaughtPanic))
	middleware := ErrorConverter()
	callable := middleware(next)
	ctx, rw := generateTestEchoContext()

	err := callable(ctx)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	expected := `
	{
		"message": "An unexpected error occurred. Code: 400"
	}`
	assert.JSONEq(t, expected, rw.Body.String())
}

func createErrorHandler(err error) echo.HandlerFunc {
	handler := func(c echo.Context) error {
		return err
	}

	return handler
}
