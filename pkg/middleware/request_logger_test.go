package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestLogger_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableHandler(RequestLogger)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_RequestLogger_PrintsRequestTiming(t *testing.T) {
	callable, _, ctx := createCallableHandler(RequestLogger)

	var out bytes.Buffer
	slogLogger := slog.New(slog.NewJSONHandler(&out, &slog.HandlerOptions{Level: slog.LevelDebug}))
	ctx.SetLogger(slogLogger)

	err := callable(ctx)
	require.Nil(t, err)
	afterCall := time.Now()

	actual := unmarshalLogOutput(t, out)
	assert.Equal(t, "INFO", actual.Level)

	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)

	assert.Equal(t, "Request processed", actual.Message)
	assert.Equal(t, "GET", actual.Method)
	assert.Equal(t, "example.com/", actual.Uri)
	assert.Regexp(t, "[0-9]+.[0-9][mµn]s", actual.Duration)
	assert.Equal(t, http.StatusOK, actual.Status)
}

func areTimeCloserThan(t1 time.Time, t2 time.Time, distance time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= distance
}
