package middleware

import (
	"io"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ResponseEnvelope_CallsNextMiddleware(t *testing.T) {
	callable, called, _ := createCallableHandler(ResponseEnvelope)

	callable()

	assert.True(t, *called)
}

func TestUnit_ResponseEnvelope_WrapsPlainOutputInResponseEnvelope(t *testing.T) {
	w, r := newTestRouter(ResponseEnvelope())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "my-output")
	})

	r.ServeHTTP(w, newGetRequest("/"))

	assert.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Body)
	require.Nil(t, err)
	actual := string(body)
	expected := `{"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"SUCCESS","status_code":200,"details":"my-output"}`
	assert.Regexp(t, expected, actual)
}

func TestUnit_ResponseEnvelope_WrapsJsonOutputInResponseEnvelope(t *testing.T) {
	type testStruct struct {
		Key string
	}
	sample := testStruct{Key: "value"}

	w, r := newTestRouter(ResponseEnvelope())
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, sample)
	})

	r.ServeHTTP(w, newGetRequest("/"))

	assert.Equal(t, http.StatusOK, w.Code)
	body, err := io.ReadAll(w.Body)
	require.Nil(t, err)
	actual := string(body)
	expected := `{"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"SUCCESS","status_code":200,"details":{"Key":"value"}}`
	assert.Regexp(t, expected, actual)
}

func TestUnit_ResponseEnvelope_CorrectlyUpdatesContentLengthToAccountForEnvelope(t *testing.T) {
	w, r := newTestRouter(ResponseEnvelope())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "my-output")
	})

	r.ServeHTTP(w, newGetRequest("/"))

	length := w.Header().Get("Content-Length")
	// The length accounts for:
	//  - 51 characters for the request identifier and quotes
	//  - 18 characters for the status and quotes
	//  - 17 characters for the HTTP status and quotes
	//  - 10 characters for the details header and quotes
	//  - 11 characters for the plain output
	//  - 5 characters for commas separating fields
	assert.Equal(t, "112", length)

	out, err := io.ReadAll(w.Body)
	require.NoError(t, err, "Actual err: %v", err)

	expectedBody := `{"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"SUCCESS","status_code":200,"details":"my-output"}`
	assert.Regexp(t, expectedBody, string(out))
}

func TestUnit_ResponseEnvelope_WhenStatusIsNot200Ok_ExpectStatusReflectsIt(t *testing.T) {
	w, r := newTestRouter(ResponseEnvelope())
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusBadGateway, "my-output")
	})

	r.ServeHTTP(w, newGetRequest("/"))

	assert.Equal(t, http.StatusBadGateway, w.Code)
	body, err := io.ReadAll(w.Body)
	require.Nil(t, err)
	actual := string(body)
	expected := `{"request_id":"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}","status":"ERROR","status_code":502,"details":"my-output"}`
	assert.Regexp(t, expected, actual)
}
