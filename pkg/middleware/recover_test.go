package middleware

import (
	stderrors "errors"
	"net/http"
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
		callable, called, ctx := createCallableHandler(t, Recover)

		err := callable(ctx)

		assert.Nil(t, err)
		assert.True(t, *called)
	})

	t.Run("returns nil when next returns nil error", func(t *testing.T) {
		called := false
		next := func(c *echo.Context) error {
			called = true
			return nil
		}

		middleware := Recover()
		callable := middleware(next)
		ctx, _ := generateTestEchoContext(t)

		err := callable(ctx)

		assert.True(t, called)
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("prevents panic", func(t *testing.T) {
		next, called := createPanicHandler()

		middleware := Recover()
		callable := middleware(next)

		modifer, out := generateLoggerModifier(t)
		ctx, _ := generateTestEchoContext(t, modifer)

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

		modifer, out := generateLoggerModifier(t)
		ctx, _ := generateTestEchoContext(t, modifer)

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

		ctx, _ := generateTestEchoContext(t, addLogger)

		err := callable(ctx)

		assertIsHttpErrorWithMessageAndCode(
			t,
			err,
			"an unexpected error occurred. Code: 400 (cause: some error)",
			http.StatusInternalServerError,
		)
	})

	t.Run("does not rewrap error with code", func(t *testing.T) {
		called := false
		next := func(c *echo.Context) error {
			called = true
			errWithCode := errors.WrapCode(errSample, errors.ErrorCode(123))
			panic(errWithCode)
		}

		middleware := Recover()
		callable := middleware(next)

		ctx, _ := generateTestEchoContext(t, addLogger)

		err := callable(ctx)

		require.True(t, called)
		assertIsHttpErrorWithMessageAndCode(
			t,
			err,
			"an unexpected error occurred. Code: 123 (cause: some error)",
			http.StatusInternalServerError,
		)
	})

	t.Run("converts non panicked errors", func(t *testing.T) {
		called := false
		next := func(c *echo.Context) error {
			called = true
			return errSample
		}

		middleware := Recover()
		callable := middleware(next)
		ctx, _ := generateTestEchoContext(t, addLogger)

		err := callable(ctx)

		require.True(t, called)
		assertIsHttpErrorWithMessageAndCode(
			t,
			err,
			"some error",
			http.StatusInternalServerError)
	})
}

func createPanicHandler() (HandlerFunc, *bool) {
	var called bool
	handler := func(c *echo.Context) error {
		called = true
		panic(errSample)
	}

	return handler, &called
}
