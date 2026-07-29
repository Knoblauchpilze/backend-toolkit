package middleware

import (
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
		originalLogger := ctx.Logger()

		callable, _, _ := createCallableTracerHandler()
		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, originalLogger, ctx.Logger())
	})

	t.Run("when request id set adds request id to logger", func(t *testing.T) {
		callable, _, _ := createCallableTracerHandler()
		ctx, out := generateTestEchoContextWithLogger()
		originalLogger := ctx.Logger()

		rest.SetContextRequestId(ctx, "my-request-id")

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.NotEqual(t, originalLogger, ctx.Logger())

		ctx.Logger().Info("test-message")
		assert.Contains(t, out.String(), `"requestId":"my-request-id"`)
	})
}

func createCallableTracerHandler() (echo.HandlerFunc, *bool, *echo.Context) {
	generator := func() echo.MiddlewareFunc {
		return RequestTracer()
	}
	middleware, called, ctx := createCallableHandler(generator)

	return middleware, called, ctx
}
