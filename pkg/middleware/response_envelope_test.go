package middleware

import (
	"io"
	"net/http"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ResponseEnvelope_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableHandler(ResponseEnvelope)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_ResponseEnvelope_WrapsPlainOutputInResponseEnvelope(t *testing.T) {
	next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

	middleware := ResponseEnvelope()
	callable := middleware(next)

	ctx, rw := generateTestEchoContext()

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, rw.Code)
	body, err := io.ReadAll(rw.Body)
	require.Nil(t, err)
	actual := string(body)
	// https://stackoverflow.com/questions/136505/searching-for-uuids-in-text-with-regex
	expected := `{"requestId":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"SUCCESS","details":"my-output"}`
	assert.Regexp(t, expected, actual)
}

func TestUnit_ResponseEnvelope_WrapsJsonOutputInResponseEnvelope(t *testing.T) {
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

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, rw.Code)
	body, err := io.ReadAll(rw.Body)
	require.Nil(t, err)
	actual := string(body)
	expected := `{"requestId":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"SUCCESS","details":{"Key":"value"}}`
	assert.Regexp(t, expected, actual)
}

func TestUnit_ResponseEnvelope_CorrectlyUpdatesContentLengthToAccountForEnvelope(t *testing.T) {
	next := createHandlerFuncWithPlainOutput(http.StatusOK, "my-output")

	middleware := ResponseEnvelope()
	callable := middleware(next)

	ctx, rw := generateTestEchoContext()

	err := callable(ctx)
	require.Nil(t, err)

	length := rw.Header().Get("Content-Length")
	// The length accounts for:
	//  - 50 characters for the request identifier and quotes
	//  - 18 characters for the status and quotes
	//  - 10 characters for the details header and quotes
	//  - 11 character for the plain output
	assert.Equal(t, "93", length)
}

func TestUnit_ResponseEnvelope_WhenStatusIsNot200Ok_ExpectStatusReflectsIt(t *testing.T) {
	next := createHandlerFuncWithPlainOutput(http.StatusBadGateway, "my-output")

	middleware := ResponseEnvelope()
	callable := middleware(next)

	ctx, rw := generateTestEchoContext()

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, http.StatusBadGateway, rw.Code)
	body, err := io.ReadAll(rw.Body)
	require.Nil(t, err)
	actual := string(body)
	expected := `{"requestId":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"ERROR","details":"my-output"}`
	assert.Regexp(t, expected, actual)
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
