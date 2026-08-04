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
		callable, called, ctx := createCallableTracerHandler(t)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
	})

	t.Run("when request id not set leaves logger unchanged", func(t *testing.T) {
		_, _, ctx := createCallableTracerHandler(t)
		originalLogger := rest.GetContextLogger(ctx.Request().Context())

		callable, _, _ := createCallableTracerHandler(t)
		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		actual := rest.GetContextLogger(ctx.Request().Context())
		assert.Equal(t, originalLogger, actual)
	})

	t.Run("when request id set adds request id to logger", func(t *testing.T) {
		callable, _, _ := createCallableTracerHandler(t)
		modifer, out := generateLoggerModifier(t)
		ctx, _ := generateTestEchoContext(t, modifer)
		originalLogger := ctx.Logger()

		req := ctx.Request()
		reqCtx := rest.WithContextRequestId(req.Context(), "my-request-id")
		ctx.SetRequest(req.WithContext(reqCtx))

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		actual := rest.GetContextLogger(ctx.Request().Context())
		assert.NotEqual(t, originalLogger, actual)

		actual.Info("test-message")
		assert.Contains(t, out.String(), `"requestId":"my-request-id"`)
	})
}

func createCallableTracerHandler(t *testing.T) (HandlerFunc, *bool, *echo.Context) {
	generator := func() MiddlewareFunc {
		return RequestTracer(slog.Default())
	}
	middleware, called, ctx := createCallableHandler(t, generator)

	return middleware, called, ctx
}
