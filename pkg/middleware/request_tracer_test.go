package middleware

import (
	"log/slog"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestTracer_CallsNextMiddleware(t *testing.T) {
	callable, called, _ := createCallableTracerHandler()

	callable()

	assert.True(t, *called)
}

func TestUnit_RequestTracer_WhenRequestIdNotSet_LeavesLoggerUnchanged(t *testing.T) {
	originalLogger := slog.Default()
	var loggerAfterTracer *slog.Logger

	w, r := newTestRouter(RequestTracer())
	r.GET("/", func(c *gin.Context) {
		loggerAfterTracer = GetContextLogger(c)
		c.Status(200)
	})

	r.ServeHTTP(w, newGetRequest("/"))
	require.NotNil(t, loggerAfterTracer)
	// No request ID was set so the logger should still be the default
	assert.Equal(t, originalLogger, loggerAfterTracer)
}

func TestUnit_RequestTracer_WhenRequestIdSet_AddsRequestIdToLogger(t *testing.T) {
	originalLogger := slog.Default()
	var loggerAfterTracer *slog.Logger

	w, r := newTestRouter(
		func(c *gin.Context) {
			// Simulate what ResponseEnvelope does: set the header before RequestTracer runs
			c.Header(requestIdHeader, "my-request-id")
			c.Next()
		},
		RequestTracer(),
	)
	r.GET("/", func(c *gin.Context) {
		loggerAfterTracer = GetContextLogger(c)
		c.Status(200)
	})

	r.ServeHTTP(w, newGetRequest("/"))
	require.NotNil(t, loggerAfterTracer)
	// Logger should have been enriched with the request ID
	assert.NotEqual(t, originalLogger, loggerAfterTracer)
}

func createCallableTracerHandler() (func(), *bool, *gin.Context) {
	callable, called, _ := createCallableHandler(func() gin.HandlerFunc {
		return RequestTracer()
	})
	ctx, _ := generateTestGinContext()
	return callable, called, ctx
}
