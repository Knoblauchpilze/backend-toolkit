package middleware

import (
	stderrors "errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errSample = stderrors.New("some error")
)

func TestUnit_Recover(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls next middleware and traps panic", func(t *testing.T) {
		handler, called := createPanicHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, *called)
	})

	t.Run("does not modify response from next middleware when it does not panic", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusNoContent, rw.Code)
		out, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, []byte{}, out)
	})

	t.Run("sets status to internal server error when next panics", func(t *testing.T) {
		handler, _ := createPanicHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("sets status to internal server error when next panics with code", func(t *testing.T) {
		handler := func(c *gin.Context) {
			panic(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
	})

	t.Run("returns body with error description when next panics", func(t *testing.T) {
		handler, _ := createPanicHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		out, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		expected := []byte("\"an unexpected error occurred. Code: 400 (cause: some error)\"")
		assert.Equal(t, expected, out)
	})

	t.Run("returns body with error description when next panics with code", func(t *testing.T) {
		handler := func(c *gin.Context) {
			panic(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		out, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		expected := []byte("\"an unexpected error occurred. Code: 102\"")
		assert.Equal(t, expected, out)
	})

	t.Run("logs error when logger is available", func(t *testing.T) {
		handler, _ := createPanicHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		afterCall := time.Now()

		actual := unmarshalLogOutput(t, *out)
		assert.Equal(t, "ERROR", actual.Level)
		safetyMargin := 5 * time.Second
		assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
		// https://golangforall.com/en/post/golang-regexp-matching-newline.html
		assert.Regexp(
			t,
			"GET example.com/ generated panic: an unexpected error occurred\\. Code: 400 \\(cause: some error\\)\\. Stack: [[:graph:]\\s]*",
			actual.Message,
		)
	})

	t.Run("logs error with code when logger is available", func(t *testing.T) {
		handler := func(c *gin.Context) {
			panic(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, Recover())

		req, out := createTestRequestWithLogger(t, http.MethodGet)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		afterCall := time.Now()

		actual := unmarshalLogOutput(t, *out)
		assert.Equal(t, "ERROR", actual.Level)
		safetyMargin := 5 * time.Second
		assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
		// https://golangforall.com/en/post/golang-regexp-matching-newline.html
		assert.Regexp(
			t,
			"GET example.com/ generated panic: an unexpected error occurred\\. Code: 102\\. Stack: [[:graph:]\\s]*",
			actual.Message,
		)
	})
}
