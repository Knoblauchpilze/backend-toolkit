package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func createTestGinHandlerWithCalledBoolean() (gin.HandlerFunc, *bool) {
	called := false
	handler := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}
	return handler, &called
}

type middlewareGenerator func() gin.HandlerFunc

// createCallableHandler builds a small Gin router with the middleware under
// test followed by a no-op handler. It returns a function that executes one
// GET / request against that router, plus a "called" flag and the recorder.
func createCallableHandler(generator middlewareGenerator) (func(), *bool, *httptest.ResponseRecorder) {
	next, called := createTestGinHandlerWithCalledBoolean()

	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(generator())
	r.GET("/", next)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	callable := func() {
		r.ServeHTTP(w, req)
	}

	return callable, called, w
}

// newTestRouter creates a test Gin engine with the provided middlewares.
func newTestRouter(middlewares ...gin.HandlerFunc) (*httptest.ResponseRecorder, *gin.Engine) {
	w := httptest.NewRecorder()
	r := gin.New()
	r.Use(middlewares...)
	return w, r
}

func newGetRequest(path string) *http.Request {
	return httptest.NewRequest(http.MethodGet, path, nil)
}

func assertJsonBody(t *testing.T, w *httptest.ResponseRecorder, expected string) {
	t.Helper()
	assert.JSONEq(t, expected, w.Body.String())
}

func generateTestGinContext() (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
	return c, w
}

func generateTestGinContextWithLogger() (*gin.Context, *bytes.Buffer) {
	c, _ := generateTestGinContext()

	var out bytes.Buffer
	slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
	SetContextLogger(c, slogLogger)

	return c, &out
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
	var actual message

	err := json.Unmarshal(out.Bytes(), &actual)
	require.Nil(t, err)

	return actual
}

func assertIsHttpError(t *testing.T, err error, message string, httpCode int) {
	t.Helper()

	httpErr, ok := err.(*httpError)
	require.True(t, ok, "Expected *httpError, got %T: %v", err, err)
	require.Equal(t, httpCode, httpErr.Code)
	require.Equal(t, message, httpErr.Message)
}
