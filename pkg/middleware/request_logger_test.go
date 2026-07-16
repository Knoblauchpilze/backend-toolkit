package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestLogger_CallsNextMiddleware(t *testing.T) {
	callable, called, _ := createCallableHandler(RequestLogger)

	callable()

	assert.True(t, *called)
}

func TestUnit_RequestLogger_PrintsRequestTiming(t *testing.T) {
	var out bytes.Buffer
	slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))

	w, r := newTestRouter(
		func(c *gin.Context) {
			SetContextLogger(c, slogLogger)
			c.Next()
		},
		RequestLogger(),
	)
	r.GET("/", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := newGetRequest("http://example.com/")
	r.ServeHTTP(w, req)
	afterCall := time.Now()

	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "INFO", actual.Level)

	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)

	assert.Equal(t, "Request processed", actual.Message)
	assert.Equal(t, "GET", actual.Method)
	assert.Equal(t, "example.com/", actual.Uri)
	assert.Regexp(t, "[0-9]+.[0-9][mµn]s", actual.Duration)
	require.Equal(t, http.StatusOK, actual.Status)
}
