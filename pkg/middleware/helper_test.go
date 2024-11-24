package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KnoblauchPilze/backend-toolkit/pkg/logger"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func createTestEchoHandlerFuncWithCalledBoolean() (echo.HandlerFunc, *bool) {
	called := false
	call := func(c echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	}
	return call, &called
}

type middlewareGenerator func() echo.MiddlewareFunc

func createCallableHandler(generator middlewareGenerator) (echo.HandlerFunc, *bool, echo.Context) {
	next, called := createTestEchoHandlerFuncWithCalledBoolean()
	ctx, _ := generateTestEchoContext()

	middlewareFunc := generator()
	callable := middlewareFunc(next)

	return callable, called, ctx
}

func generateTestEchoContext() (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	return generateTestEchoContextFromRequest(req)
}

func generateTestEchoContextWithLogger() (echo.Context, *bytes.Buffer) {
	ctx, _ := generateTestEchoContext()

	var out bytes.Buffer
	log := logger.New(&out)
	ctx.SetLogger(logger.Wrap(log))

	return ctx, &out
}

func generateTestEchoContextFromRequest(req *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	rw := httptest.NewRecorder()

	ctx := e.NewContext(req, rw)

	return ctx, rw
}

type message struct {
	Level   string
	Time    time.Time
	Message string
}

func unmarshalLogOutput(t *testing.T, out bytes.Buffer) message {
	var actual message

	err := json.Unmarshal(out.Bytes(), &actual)
	require.Nil(t, err)

	return actual
}

func assertIsHttpErrorWithMessageAndCode(t *testing.T, err error, message string, httpCode int) {
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)

	require.Equal(t, httpCode, httpErr.Code)
	require.Equal(t, message, httpErr.Message)
}
