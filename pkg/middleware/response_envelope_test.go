package middleware

import (
	"io"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ResponseEnvelope(t *testing.T) {
	sampleRequestId := uuid.MustParse("a57f4ca1-1a22-4990-b3de-5f836a3ea4e9").String()

	t.Run("calls next middleware", func(t *testing.T) {
		callable, called, ctx := createCallableHandler(ResponseEnvelope)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.True(t, *called)
	})

	t.Run("wraps plain output in response envelope", func(t *testing.T) {
		next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext()
		ctx.Set(requestIdContextKey, sampleRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusOK, rw.Code)
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		actual := string(body)
		expected := `{"request_id":"a57f4ca1-1a22-4990-b3de-5f836a3ea4e9","status":"SUCCESS","status_code":200,"details":"my-output"}`
		assert.Regexp(t, expected, actual)
	})

	t.Run("wraps json output in response envelope", func(t *testing.T) {
		type testStruct struct {
			Key string
		}
		sample := testStruct{
			Key: "value",
		}
		next := createHandlerFuncWithJsonOutput(http.StatusOK, sample)

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext()
		ctx.Set(requestIdContextKey, sampleRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusOK, rw.Code)
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		actual := string(body)
		expected := `{"request_id":"a57f4ca1-1a22-4990-b3de-5f836a3ea4e9","status":"SUCCESS","status_code":200,"details":{"Key":"value"}}`
		assert.Regexp(t, expected, actual)
	})

	t.Run("correctly updates content length to account for envelope", func(t *testing.T) {
		next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext()
		ctx.Set(requestIdContextKey, sampleRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		length := rw.Header().Get("Content-Length")
		// The length accounts for:
		//  - 51 characters for the request identifier and quotes
		//  - 18 characters for the status and quotes
		//  - 17 characters for the HTTP status and quotes
		//  - 10 characters for the details header and quotes
		//  - 11 characters for the plain output
		//  - 5 characters for commas separating fields
		assert.Equal(t, "112", length)

		out, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)

		expectedBody := `{"request_id":"a57f4ca1-1a22-4990-b3de-5f836a3ea4e9","status":"SUCCESS","status_code":200,"details":"my-output"}`
		assert.Regexp(t, expectedBody, string(out))
	})

	t.Run("when status is not 200 ok expect status reflects it", func(t *testing.T) {
		next := createHandlerFuncWithPlainOutput(http.StatusBadGateway, "my-output")

		middleware := ResponseEnvelope()
		callable := middleware(next)

		ctx, rw := generateTestEchoContext()
		ctx.Set(requestIdContextKey, sampleRequestId)

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusBadGateway, rw.Code)
		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)
		actual := string(body)
		expected := `{"request_id":"a57f4ca1-1a22-4990-b3de-5f836a3ea4e9","status":"ERROR","status_code":502,"details":"my-output"}`
		assert.Regexp(t, expected, actual)
	})

	t.Run("uses unknown request id when context does not provide one", func(t *testing.T) {
		next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

		callable := ResponseEnvelope()(next)
		ctx, rw := generateTestEchoContext()

		err := callable(ctx)
		require.NoError(t, err, "Actual err: %v", err)

		body, err := io.ReadAll(rw.Body)
		require.NoError(t, err, "Actual err: %v", err)

		expected := `{"request_id":"unknown","status":"SUCCESS","status_code":200,"details":"my-output"}`
		assert.Regexp(t, expected, string(body))
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
