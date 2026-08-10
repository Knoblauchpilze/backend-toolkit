package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_RequestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls next middleware", func(t *testing.T) {
		handler, called := createHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, RequestLogger())

		req, _ := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, *called)
	})

	t.Run("prints request timing", func(t *testing.T) {
		handler, _ := createHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, RequestLogger())

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		afterCall := time.Now()

		actual := unmarshalLogOutput(t, *out)
		assert.Equal(t, "INFO", actual.Level)

		safetyMargin := 5 * time.Second
		assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)

		assert.Equal(t, "Request processed", actual.Message)
		assert.Equal(t, "GET", actual.Method)
		assert.Equal(t, "example.com/", actual.Uri)
		assert.Regexp(t, "[0-9]+.[0-9][mµn]s", actual.Duration)
		assert.Equal(t, http.StatusNoContent, actual.Status)
	})

	t.Run("prints request timing when handler panics", func(t *testing.T) {
		handler, called := createPanicHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, RequestLogger())

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()

		recovered := false
		func() {
			// Recover as there's nothing else preventing the panic to
			// bubble up.
			defer func() {
				recover() //nolint:errcheck
				recovered = true
			}()

			r.ServeHTTP(rw, req)
		}()

		require.True(t, *called)
		require.True(t, recovered)
		actual := unmarshalLogOutput(t, *out)
		assert.Equal(t, "Request processed", actual.Message)
	})
}

func areTimeCloserThan(t1 time.Time, t2 time.Time, distance time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= distance
}
