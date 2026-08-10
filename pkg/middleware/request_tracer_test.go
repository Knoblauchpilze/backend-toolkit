package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestTracer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls next middleware", func(t *testing.T) {
		handler, called := createHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, RequestTracer(slog.Default()))

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, *called)
	})

	t.Run("when request defines logger uses it for logging", func(t *testing.T) {
		called := false
		handler := func(c *gin.Context) {
			called = true
			log := rest.GetContextLogger(c.Request.Context())
			log.Info("test-message")
			c.Status(http.StatusNoContent)
		}

		var out1 bytes.Buffer
		log1 := slog.New(slog.NewJSONHandler(&out1, &slog.HandlerOptions{Level: slog.LevelDebug}))

		r := createTestGinRouterWithHandler(t, handler, RequestTracer(log1))

		req, out2 := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		require.True(t, called)
		assert.Empty(t, out1.Bytes())
		assert.Regexp(
			t,
			`\{"time":"[^"]+","level":"INFO","msg":"test-message"\}`,
			out2.String(),
		)
	})

	t.Run("when request does not define logger uses default if request id is set", func(t *testing.T) {
		called := false
		handler := func(c *gin.Context) {
			called = true
			log := rest.GetContextLogger(c.Request.Context())
			log.Info("test-message")
			c.Status(http.StatusNoContent)
		}

		var out bytes.Buffer
		log := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))

		r := createTestGinRouterWithHandler(t, handler, RequestTracer(log))

		ctx := rest.WithContextRequestId(t.Context(), "my-request-id")
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/", nil)
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		require.True(t, called)
		assert.Regexp(
			t,
			`\{"time":"[^"]+","level":"INFO","msg":"test-message","requestId":"my-request-id"\}`,
			out.String(),
		)
	})

	t.Run("when request id not set leaves logger unchanged", func(t *testing.T) {
		called := false
		handler := func(c *gin.Context) {
			called = true
			log := rest.GetContextLogger(c.Request.Context())
			log.Info("test-message")
			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, RequestTracer(slog.Default()))

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		require.True(t, called)
		assert.Regexp(
			t,
			`\{"time":"[^"]+","level":"INFO","msg":"test-message"\}`,
			out.String(),
		)
	})

	t.Run("when request id set adds request id to logger", func(t *testing.T) {
		called := false
		handler := func(c *gin.Context) {
			called = true
			log := rest.GetContextLogger(c.Request.Context())
			log.Info("test-message")
			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, RequestTracer(slog.Default()))

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		ctx := rest.WithContextRequestId(req.Context(), "my-request-id")
		req = req.WithContext(ctx)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		require.True(t, called)
		assert.Regexp(
			t,
			`\{"time":"[^"]+","level":"INFO","msg":"test-message","requestId":"my-request-id"\}`,
			out.String(),
		)
	})
}
