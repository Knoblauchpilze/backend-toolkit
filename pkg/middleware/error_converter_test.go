package middleware

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_ErrorConverter_CallsNextMiddleware(t *testing.T) {
	callable, called, _ := createCallableHandler(ErrorConverter)

	callable()

	assert.True(t, *called)
}

func TestUnit_ErrorConverter_WrapsUnknownErrorIntoHttpError(t *testing.T) {
	w, r := newTestRouter(ErrorConverter())
	r.GET("/", func(c *gin.Context) {
		_ = c.Error(fmt.Errorf("some error"))
	})

	req := newGetRequest("/")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assertJsonBody(t, w, `{"message":"some error"}`)
}

func TestUnit_ErrorConverter_WrapsErrorWithCodeIntoHttpError(t *testing.T) {
	w, r := newTestRouter(ErrorConverter())
	r.GET("/", func(c *gin.Context) {
		_ = c.Error(ErrUncaughtPanic)
	})

	req := newGetRequest("/")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assertJsonBody(t, w, `{"message":"an unexpected error occurred. Code: 400"}`)
}

func TestUnit_ErrorConverter_NoError_DoesNotWriteBody(t *testing.T) {
	callable, _, w := createCallableHandler(ErrorConverter)
	callable()
	require.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Body.String())
}
