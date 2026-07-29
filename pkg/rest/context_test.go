package rest

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_RequestIdFromContext(t *testing.T) {
	t.Run("returns request id when it exists", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))
		expectedId := uuid.NewString()
		SetContextRequestId(ctx, expectedId)

		requestId, ok := RequestIdFromContext(ctx)

		assert.True(t, ok)
		assert.Equal(t, expectedId, requestId)
	})

	t.Run("returns empy value when unset", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		requestId, ok := RequestIdFromContext(ctx)

		assert.False(t, ok)
		assert.Empty(t, requestId)
	})
}

func TestUnit_GetContextLogger(t *testing.T) {
	t.Run("when no logger injected returns non nil logger", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(nil)

		actual := GetContextLogger(ctx)

		assert.NotNil(t, actual)
	})

	t.Run("when logger is injected logs to expected output", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		var out bytes.Buffer
		slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
		ctx.SetLogger(slogLogger)

		GetContextLogger(ctx).Info("test-message")

		assert.Contains(t, out.String(), `"msg":"test-message"`)
	})
}
