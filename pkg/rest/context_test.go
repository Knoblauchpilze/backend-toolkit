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
	t.Run("when no logger injected returns nil logger", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(nil)

		actual := GetContextLogger(ctx)

		assert.Nil(t, actual)
	})

	t.Run("logs to configured output when logger is available", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		var out bytes.Buffer
		slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
		SetContextLogger(ctx, slogLogger)

		GetContextLogger(ctx).Info("test-message")

		assert.Contains(t, out.String(), `"msg":"test-message"`)
	})

	t.Run("returns logger when already available in the context", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		var out bytes.Buffer
		expected := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
		SetContextLogger(ctx, expected)

		actual := GetContextLogger(ctx)
		actual.Info("test-message")

		assert.Equal(t, expected, actual)
		assert.Contains(t, out.String(), `"msg":"test-message"`)
	})
}

func TestUnit_SetContextLogger(t *testing.T) {
	t.Run("stores logger that can be read by GetContextLogger", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		var out bytes.Buffer
		log := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))

		SetContextLogger(ctx, log)

		actual := GetContextLogger(ctx)
		actual.Info("test-message")

		assert.Equal(t, log, actual)
		assert.Contains(t, out.String(), `"msg":"test-message"`)
	})

	t.Run("overrides existing logger when defining a new one", func(t *testing.T) {
		ctx, _ := generateTestEchoContextFromRequest(httptest.NewRequest(http.MethodGet, "/", nil))

		var firstOut bytes.Buffer
		firstLog := slog.New(slog.NewJSONHandler(&firstOut, &slog.HandlerOptions{Level: slog.LevelDebug}))

		var secondOut bytes.Buffer
		secondLog := slog.New(slog.NewJSONHandler(&secondOut, &slog.HandlerOptions{Level: slog.LevelDebug}))

		SetContextLogger(ctx, firstLog)
		SetContextLogger(ctx, secondLog)

		actual := GetContextLogger(ctx)
		actual.Info("test-message")

		assert.Equal(t, secondLog, actual)
		assert.Contains(t, secondOut.String(), `"msg":"test-message"`)
		assert.Empty(t, firstOut.String())
	})
}
