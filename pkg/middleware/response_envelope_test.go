package middleware

import (
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ResponseEnvelope(t *testing.T) {
	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableHandler(t, ResponseEnvelope)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
	})

	t.Run("wraps plain output in envelope with request id", func(t *testing.T) {
		next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext(t, addRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

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
		next := createHandlerFuncWithJsonOutput(http.StatusOK, sample)

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext(t, addRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

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

		next := createHandlerFuncWithJsonOutput(http.StatusOK, testStruct{Key: "value"})

		callable := ResponseEnvelope()(next)
		ctx, rw := generateTestEchoContext(t)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

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
		next := createHandlerFuncWithPlainOutput(http.StatusBadGateway, "my-output")

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext(t, addRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

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

func createHandlerFuncWithPlainOutput(httpCode int, out string) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return c.String(httpCode, out)
	}
}

func createHandlerFuncWithJsonOutput[T any](httpCode int, out T) echo.HandlerFunc {
	return func(c *echo.Context) error {
		return c.JSON(httpCode, out)
	}
}
