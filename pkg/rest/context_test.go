package rest

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
