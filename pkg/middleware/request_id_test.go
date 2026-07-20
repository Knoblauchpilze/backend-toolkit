package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestId(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableHandler(RequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
	})

	t.Run("sets response header and context with same value", func(t *testing.T) {
		callable, _, ctx := createCallableHandler(RequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		headerValue := ctx.Response().Header().Get(requestIdHeader)
		assert.NotEmpty(t, headerValue)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, headerValue)

		contextValue, exists := ctx.Get(requestIdContextKey).(string)
		assert.True(t, exists)
		assert.Equal(t, headerValue, contextValue)
	})

	t.Run("does not change request id when already present", func(t *testing.T) {
		existingRequestId := uuid.NewString()

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		req.Header.Set(requestIdHeader, existingRequestId)

		ctx, _ := generateTestEchoContextFromRequest(req)
		next, called := createTestEchoHandlerFuncWithCalledBoolean()

		callable := RequestId()(next)
		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
		assert.Equal(t, existingRequestId, ctx.Response().Header().Get(requestIdHeader))
	})
}

func TestUnit_RequestIdFromContext(t *testing.T) {
	t.Run("returns request id when it exists", func(t *testing.T) {
		ctx, _ := generateTestEchoContext()
		expectedId := uuid.NewString()
		ctx.Set(requestIdContextKey, expectedId)

		requestId, ok := RequestIdFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, expectedId, requestId)
	})

	t.Run("returns that no header exists when unset", func(t *testing.T) {
		ctx, _ := generateTestEchoContext()

		requestId, ok := RequestIdFromContext(ctx)

		assert.False(t, ok)
		assert.Empty(t, requestId)
	})
}
