package middleware

import (
	stderrors "errors"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errSample = stderrors.New("some error")
)

func TestUnit_Recover(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableHandler(Recover)

		err := callable(ctx)

		assert.Nil(t, err)
		assert.True(t, *called)
	})

	t.Run("prevents panic", func(t *testing.T) {
		next, called := createPanicHandler()

		middleware := Recover()
		callable := middleware(next)

		ctx, out := generateTestEchoContextWithLogger()

		err := callable(ctx)

		assert.NotNil(t, err)
		assert.True(t, *called)

		expected := `GET example.com/ generated panic: an unexpected error occurred. Code: 400 (cause: some error)`
		assert.Contains(t, out.String(), expected)
	})

	t.Run("logs error", func(t *testing.T) {
		next, _ := createPanicHandler()

		middleware := Recover()
		callable := middleware(next)

		ctx, out := generateTestEchoContextWithLogger()

		err := callable(ctx)
		require.NotNil(t, err)
		afterCall := time.Now()

		actual := unmarshalLogOutput(t, *out)
		assert.Equal(t, "ERROR", actual.Level)
		safetyMargin := 5 * time.Second
		assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
		// https://golangforall.com/en/post/golang-regexp-matching-newline.html
		assert.Regexp(
			t,
			"GET example.com/ generated panic: an unexpected error occurred\\. Code: 400 \\(cause: some error\\)\\. Stack: [[:graph:]\\s]*",
			actual.Message,
		)
	})

	t.Run("returns error from panic", func(t *testing.T) {
		next, _ := createPanicHandler()

		middleware := Recover()
		callable := middleware(next)

		ctx, _ := generateTestEchoContextWithLogger()

		err := callable(ctx)
		require.NotNil(t, err)

		errWithCode, ok := errors.AsErrorWithCode(err)
		require.True(t, ok)
		assert.Equal(t, errUncaughtPanic, errWithCode.Code)
		assert.ErrorIs(t, errWithCode.Cause, errSample)
	})

	t.Run("does not rewrap error with code", func(t *testing.T) {
		next, _ := createPanicHandlerWithErrorCode()

		middleware := Recover()
		callable := middleware(next)

		ctx, _ := generateTestEchoContextWithLogger()

		err := callable(ctx)
		require.NotNil(t, err)

		errWithCode, ok := errors.AsErrorWithCode(err)
		require.True(t, ok)
		assert.Equal(t, errors.ErrorCode(123), errWithCode.Code)
		assert.ErrorIs(t, errWithCode.Cause, errSample)
	})
}

func createPanicHandler() (echo.HandlerFunc, *bool) {
	var called bool
	handler := func(c *echo.Context) error {
		called = true
		panic(errSample)
	}

	return handler, &called
}

func createPanicHandlerWithErrorCode() (echo.HandlerFunc, *bool) {
	var called bool
	handler := func(c *echo.Context) error {
		called = true
		errWithCode := errors.WrapCode(errSample, errors.ErrorCode(123))
		panic(errWithCode)
	}

	return handler, &called
}
