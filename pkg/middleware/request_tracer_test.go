package middleware

import (
	"log/slog"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestTracer(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableTracerHandler()

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
	})

	t.Run("when request id not set leaves logger unchanged", func(t *testing.T) {
		_, _, ctx := createCallableTracerHandler()
		originalLogger := rest.GetContextLogger(ctx)

		callable, _, _ := createCallableTracerHandler()
		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		actual := rest.GetContextLogger(ctx)
		assert.Equal(t, originalLogger, actual)
	})

	t.Run("when request id set adds request id to logger", func(t *testing.T) {
		callable, _, _ := createCallableTracerHandler()
		ctx, out := generateTestEchoContextWithLogger()
		originalLogger := ctx.Logger()

		rest.SetContextRequestId(ctx, "my-request-id")

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		actual := rest.GetContextLogger(ctx)
		assert.NotEqual(t, originalLogger, actual)

		actual.Info("test-message")
		assert.Contains(t, out.String(), `"requestId":"my-request-id"`)
	})
}

func createCallableTracerHandler() (echo.HandlerFunc, *bool, *echo.Context) {
	generator := func() echo.MiddlewareFunc {
		return RequestTracer(slog.Default())
	}
	middleware, called, ctx := createCallableHandler(generator)

	return middleware, called, ctx
}
