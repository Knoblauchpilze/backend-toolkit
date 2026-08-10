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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	sampleRequestId = uuid.MustParse("f1df556f-f7c1-4022-832c-e1a7b8a6f3d5")
)

func createPanicHandlerWithCalledBoolean() (HandlerFunc, *bool) {
	var called bool
	handler := func(c *gin.Context) {
		called = true
		panic(errSample)
	}

	return handler, &called
}

func createHandlerWithCalledBoolean() (HandlerFunc, *bool) {
	called := false
	call := func(c *gin.Context) {
		called = true
		c.Status(http.StatusNoContent)
	}
	return call, &called
}

func createTestGinRouterWithHandler(
	t *testing.T,
	handler HandlerFunc,
	middlewares ...HandlerFunc,
) *gin.Engine {
	t.Helper()

	r := gin.New()

	for _, middleware := range middlewares {
		r.Use(middleware)
	}

	r.GET("/", handler)

	return r
}

func createTestRequestWithLogger(
	t *testing.T,
	method string,
) (*http.Request, *bytes.Buffer) {
	t.Helper()

	var out bytes.Buffer

	slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx := rest.WithContextLogger(t.Context(), slogLogger)

	req := httptest.NewRequestWithContext(ctx, method, "http://example.com/", nil)

	return req, &out
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
