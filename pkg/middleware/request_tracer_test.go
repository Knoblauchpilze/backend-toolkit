package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestTracer_CallsNextMiddleware(t *testing.T) {
	log, _ := newLogger()
	callable, called, ctx := createCallableTracerHandler(log)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_RequestTracer_WhenRequestIdNotSet_LeavesPrefixEmpty(t *testing.T) {
	log, _ := newLogger()
	callable, _, ctx := createCallableTracerHandler(log)

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, "", ctx.Logger().Prefix())
}

func TestUnit_RequestTracer_WhenRequestIdIsNotValid_LeavesPrefixEmpty(t *testing.T) {
	log, _ := newLogger()
	callable, _, ctx := createCallableTracerHandler(log)

	ctx.Response().Header().Set(requestIdHeader, "my-request-id-1")
	ctx.Response().Header().Add(requestIdHeader, "my-request-id-2")

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, "", ctx.Logger().Prefix())
}

func TestUnit_RequestTracer_WhenRequestIdSet_SetsLoggerPrefix(t *testing.T) {
	log, _ := newLogger()
	callable, _, ctx := createCallableTracerHandler(log)

	ctx.Response().Header().Set(requestIdHeader, "my-request-id")

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, "my-request-id", ctx.Logger().Prefix())
}

func TestUnit_RequestTracer_MultipleRequestsWriteToSameOutput(t *testing.T) {
	log, out := newLogger()

	next := func(c echo.Context) error {
		c.Logger().Printf(c.Request().Host)
		return nil
	}

	generator := RequestTracer(log)

	req1 := httptest.NewRequest(http.MethodGet, "http://test1", nil)
	ctx1, _ := generateTestEchoContextFromRequest(req1)
	ctx1.Response().Header().Set(requestIdHeader, "req1")

	req2 := httptest.NewRequest(http.MethodGet, "http://test2", nil)
	ctx2, _ := generateTestEchoContextFromRequest(req2)
	ctx2.Response().Header().Set(requestIdHeader, "req2")

	callable := generator(next)
	err := callable(ctx1)
	require.Nil(t, err)
	err = callable(ctx2)
	require.Nil(t, err)

	expected := "{\"level\":\"debug\",\"message\":\"test1\"}\n{\"level\":\"debug\",\"message\":\"test2\"}\n"
	assert.Equal(t, expected, out.String())
}

func newLogger() (echo.Logger, *bytes.Buffer) {
	var out bytes.Buffer
	return logger.Wrap(logger.New(&out)), &out
}

func createCallableTracerHandler(log echo.Logger) (echo.HandlerFunc, *bool, echo.Context) {
	generator := func() echo.MiddlewareFunc {
		return RequestTracer(log)
	}
	middleware, called, ctx := createCallableHandler(generator)
	ctx.SetLogger(log)

	return middleware, called, ctx

}
