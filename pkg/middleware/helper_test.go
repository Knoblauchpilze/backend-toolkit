package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/require"
)

var (
	sampleRequestId = uuid.MustParse("f1df556f-f7c1-4022-832c-e1a7b8a6f3d5")
)

func createTestEchoHandlerFuncWithCalledBoolean() (echo.HandlerFunc, *bool) {
	called := false
	call := func(c *echo.Context) error {
		called = true
		return c.NoContent(http.StatusOK)
	}
	return call, &called
}

type middlewareGenerator func() echo.MiddlewareFunc

func createCallableHandler(
	t *testing.T,
	generator middlewareGenerator,
) (echo.HandlerFunc, *bool, *echo.Context) {
	next, called := createTestEchoHandlerFuncWithCalledBoolean()
	ctx, _ := generateTestEchoContext(t)

	middlewareFunc := generator()
	callable := middlewareFunc(next)

	return callable, called, ctx
}

type contextModifier func(*testing.T, *echo.Context)

func generateTestEchoContext(
	t *testing.T,
	modifiers ...contextModifier,
) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	ctx, rw := generateTestEchoContextFromRequest(t, req)

	for _, modifier := range modifiers {
		modifier(t, ctx)
	}

	return ctx, rw
}

func addLogger(t *testing.T, ctx *echo.Context) {
	t.Helper()

	var out bytes.Buffer
	slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
	req := ctx.Request()
	reqCtx := rest.WithContextLogger(req.Context(), slogLogger)
	ctx.SetRequest(req.WithContext(reqCtx))
}

func addRequestId(t *testing.T, ctx *echo.Context) {
	t.Helper()

	req := ctx.Request()
	reqCtx := rest.WithContextRequestId(req.Context(), sampleRequestId.String())
	ctx.SetRequest(req.WithContext(reqCtx))
}

func generateLoggerModifier(
	t *testing.T,
) (contextModifier, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer

	modifier := func(t *testing.T, ctx *echo.Context) {
		slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
		req := ctx.Request()
		reqCtx := rest.WithContextLogger(req.Context(), slogLogger)
		ctx.SetRequest(req.WithContext(reqCtx))
	}

	return modifier, &out
}

func generateTestEchoContextFromRequest(
	t *testing.T,
	req *http.Request,
) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	rw := httptest.NewRecorder()

	ctx := e.NewContext(req, rw)

	return ctx, rw
}

type message struct {
	Time     time.Time `json:"time"`
	Level    string    `json:"level"`
	Message  string    `json:"msg"`
	Method   string    `json:"method"`
	Uri      string    `json:"uri"`
	Duration string    `json:"duration"`
	Status   int       `json:"status"`
}

func unmarshalLogOutput(t *testing.T, out bytes.Buffer) message {
	t.Helper()

	var actual message

	err := json.Unmarshal(out.Bytes(), &actual)
	require.Nil(t, err)

	return actual
}

func assertIsHttpErrorWithMessageAndCode(t *testing.T, err error, message string, httpCode int) {
	t.Helper()

	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)

	require.Equal(t, httpCode, httpErr.Code)
	require.Equal(t, message, httpErr.Message)
}
