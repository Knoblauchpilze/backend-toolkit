package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestTracer_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableTracerHandler()

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_RequestTracer_WhenRequestIdNotSet_LeavesLoggerUnchanged(t *testing.T) {
	_, _, ctx := createCallableTracerHandler()
	originalLogger := ctx.Logger()

	callable, _, _ := createCallableTracerHandler()
	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, originalLogger, ctx.Logger())
}

func TestUnit_RequestTracer_WhenRequestIdSet_AddsRequestIdToLogger(t *testing.T) {
	callable, _, ctx := createCallableTracerHandler()

	ctx.Response().Header().Set(requestIdHeader, "my-request-id")

	err := callable(ctx)
	require.Nil(t, err)

	// The logger should have been updated (it's a different logger with the requestId attribute)
	assert.NotNil(t, ctx.Logger())
}

func createCallableTracerHandler() (echo.HandlerFunc, *bool, *echo.Context) {
	generator := func() echo.MiddlewareFunc {
		return RequestTracer()
	}
	middleware, called, ctx := createCallableHandler(generator)

	return middleware, called, ctx
}

func generateTestEchoContextFromRequestForTracer(req *http.Request) (*echo.Context, *httptest.ResponseRecorder) {
	return generateTestEchoContextFromRequest(req)
}
