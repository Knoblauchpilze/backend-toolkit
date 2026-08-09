package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestUnit_ErrorConverter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets status to internal server error when only c.Error is used", func(t *testing.T) {
		handler := func(c *gin.Context) {
			_ = c.Error(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, ErrorConverter())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.JSONEq(t, `{"message":"an unexpected error occurred. Code: 102"}`, rw.Body.String())
	})

	t.Run("does not override already written response", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.String(http.StatusBadRequest, "bad request")
			_ = c.Error(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, ErrorConverter())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusBadRequest, rw.Code)
		assert.Equal(t, "bad request", rw.Body.String())
	})

	t.Run("does nothing when context is aborted", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.Abort()
			_ = c.Error(db.ErrAlreadyCommitted)
		}

		r := createTestGinRouterWithHandler(t, handler, ErrorConverter())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Empty(t, rw.Body.Bytes())
	})

	t.Run("does nothing when no error exists in context", func(t *testing.T) {
		handler := func(c *gin.Context) {
			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, ErrorConverter())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusNoContent, rw.Code)
		assert.Empty(t, rw.Body.Bytes())
	})

	t.Run("uses last error when multiple errors exist", func(t *testing.T) {
		firstErr := errors.New("first error")
		lastErr := errors.New("last error")

		handler := func(c *gin.Context) {
			_ = c.Error(firstErr)
			_ = c.Error(lastErr)
		}

		r := createTestGinRouterWithHandler(t, handler, ErrorConverter())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.JSONEq(t, `{"message":"last error"}`, rw.Body.String())
	})
}
