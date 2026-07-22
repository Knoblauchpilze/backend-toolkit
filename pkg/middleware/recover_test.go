package middleware

import (
	"errors"
	"testing"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errSample = errors.New("some error")
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

		ctx, _ := generateTestEchoContext()

		err := callable(ctx)

		assert.NotNil(t, err)
		assert.True(t, *called)
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
		assert.Regexp(t, "GET example.com/ generated panic: some error. Stack: [[:graph:]\\s]*", actual.Message)
	})

	t.Run("returns error from panic", func(t *testing.T) {
		next, _ := createPanicHandler()

		middleware := Recover()
		callable := middleware(next)

		ctx, _ := generateTestEchoContext()

		err := callable(ctx)
		require.NotNil(t, err)

		assert.ErrorIs(t, err, errSample)
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
