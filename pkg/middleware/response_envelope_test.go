package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ResponseEnvelope(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		handler, called := createHandlerWithCalledBoolean()

		r := createTestGinRouterWithHandler(t, handler, ResponseEnvelope())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.True(t, *called)
	})

	t.Run("wraps plain output in envelope with request id", func(t *testing.T) {
		handler := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

		r := createTestGinRouterWithHandler(t, handler, ResponseEnvelope())

		ctx := rest.WithContextRequestId(t.Context(), sampleRequestId.String())
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, echo.MIMEApplicationJSON, rw.Header().Get(echo.HeaderContentType))
		assert.Equal(t, sampleRequestId.String(), rw.Header().Get(requestIdHeader))
		length := rw.Header().Get("Content-Length")
		// The length includes the value (`my-output`) and the envelope
		assert.Equal(t, "112", length)
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		actual := string(body)
		expected := `{"request_id":"f1df556f-f7c1-4022-832c-e1a7b8a6f3d5","status":"SUCCESS","status_code":200,"details":"my-output"}`
		assert.Regexp(t, expected, actual)
	})

	t.Run("wraps json output in envelope with request id", func(t *testing.T) {
		type testStruct struct {
			Key string
		}
		sample := testStruct{
			Key: "value",
		}
		handler := createHandlerFuncWithJsonOutput(http.StatusOK, sample)

		r := createTestGinRouterWithHandler(t, handler, ResponseEnvelope())

		ctx := rest.WithContextRequestId(t.Context(), sampleRequestId.String())
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, echo.MIMEApplicationJSON, rw.Header().Get(echo.HeaderContentType))
		assert.Equal(t, sampleRequestId.String(), rw.Header().Get(requestIdHeader))
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)

		length, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, len(body), length)

		actual := string(body)
		expected := `{"request_id":"f1df556f-f7c1-4022-832c-e1a7b8a6f3d5","status":"SUCCESS","status_code":200,"details":{"Key":"value"}}`
		assert.Regexp(t, expected, actual)
	})

	t.Run("wraps output in envelope with unknown request id", func(t *testing.T) {
		type testStruct struct {
			Key string
		}
		handler := createHandlerFuncWithJsonOutput(http.StatusOK, testStruct{Key: "value"})

		r := createTestGinRouterWithHandler(t, handler, ResponseEnvelope())

		req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, echo.MIMEApplicationJSON, rw.Header().Get(echo.HeaderContentType))
		assert.Equal(t, defaultRequestId, rw.Header().Get(requestIdHeader))
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)

		length, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, len(body), length)

		expected := `{"request_id":"unknown","status":"SUCCESS","status_code":200,"details":{"Key":"value"}}`
		assert.Regexp(t, expected, string(body))
	})

	t.Run("when status is not 200 ok expect status reflects it", func(t *testing.T) {
		handler := createHandlerFuncWithPlainOutput(http.StatusBadGateway, "my-output")

		r := createTestGinRouterWithHandler(t, handler, ResponseEnvelope())

		ctx := rest.WithContextRequestId(t.Context(), sampleRequestId.String())
		req := httptest.NewRequestWithContext(ctx, http.MethodGet, "http://example.com/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)

		assert.Equal(t, http.StatusBadGateway, rw.Code)
		assert.Equal(t, echo.MIMEApplicationJSON, rw.Header().Get(echo.HeaderContentType))
		assert.Equal(t, sampleRequestId.String(), rw.Header().Get(requestIdHeader))
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)

		length, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, len(body), length)

		actual := string(body)
		expected := `{"request_id":"f1df556f-f7c1-4022-832c-e1a7b8a6f3d5","status":"ERROR","status_code":502,"details":"my-output"}`
		assert.Regexp(t, expected, actual)
	})
}

func createHandlerFuncWithPlainOutput(httpCode int, out string) HandlerFunc {
	return func(c *gin.Context) {
		c.String(httpCode, out)
	}
}

func createHandlerFuncWithJsonOutput[T any](httpCode int, out T) HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(httpCode, out)
	}
}
