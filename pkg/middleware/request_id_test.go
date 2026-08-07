package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUnit_RequestId(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("calls next middleware", func(t *testing.T) {
		handler, called := createHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, RequestId())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, *called)
	})

	t.Run("sets response header and context with same value", func(t *testing.T) {
		var contextValue string
		requestIdExists := false
		handler := func(c *gin.Context) {
			contextValue, requestIdExists = rest.RequestIdFromContext(c.Request.Context())

			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, RequestId())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		headerValue := rw.Header().Get(requestIdHeader)
		assert.NotEmpty(t, headerValue)
		assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`, headerValue)

		assert.True(t, requestIdExists)
		assert.Equal(t, headerValue, contextValue)
	})

	t.Run("does not change request id when already present", func(t *testing.T) {
		var contextValue string
		requestIdExists := false
		handler := func(c *gin.Context) {
			contextValue, requestIdExists = rest.RequestIdFromContext(c.Request.Context())

			c.Status(http.StatusNoContent)
		}

		r := createTestGinRouterWithHandler(t, handler, RequestId())

		existingRequestId := uuid.NewString()
		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		req.Header.Set(requestIdHeader, existingRequestId)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, requestIdExists)
		assert.Equal(t, existingRequestId, contextValue)
		assert.Equal(t, existingRequestId, rw.Header().Get(requestIdHeader))
	})
}
