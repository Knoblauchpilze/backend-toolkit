package middleware

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Recover_CallsNextMiddleware(t *testing.T) {
	callable, called, _ := createCallableHandler(Recover)

	callable()

	assert.True(t, *called)
}

func TestUnit_Recover_PreventsPanic(t *testing.T) {
	panicHandler, called := createPanicHandler()

	w, r := newTestRouter(Recover())
	r.GET("/", panicHandler)

	r.ServeHTTP(w, newGetRequest("/"))

	assert.True(t, *called)
	// No crash: test passes if we reach here
}

func TestUnit_Recover_LogsError(t *testing.T) {
	ctx, out := generateTestGinContextWithLogger()

	w, r := newTestRouter(
		func(c *gin.Context) {
			SetContextLogger(c, GetContextLogger(ctx))
			c.Next()
		},
		Recover(),
	)
	r.GET("/", func(c *gin.Context) {
		panic(fmt.Errorf("some error"))
	})

	r.ServeHTTP(w, newGetRequest("http://example.com/"))
	afterCall := time.Now()

	actual := unmarshalLogOutput(t, *out)
	assert.Equal(t, "ERROR", actual.Level)
	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
	// https://golangforall.com/en/post/golang-regexp-matching-newline.html
	assert.Regexp(t, "GET example.com/ generated panic: some error. Stack: [[:graph:]\\s]*", actual.Message)
}

func TestUnit_Recover_SetsStatusCodeToError(t *testing.T) {
	w, r := newTestRouter(ErrorConverter(), Recover())
	r.GET("/", func(c *gin.Context) {
		panic(fmt.Errorf("some error"))
	})

	r.ServeHTTP(w, newGetRequest("/"))

	require.Equal(t, http.StatusInternalServerError, w.Code)
	assertJsonBody(t, w, `{"message":"some error"}`)
}

func areTimeCloserThan(t1 time.Time, t2 time.Time, distance time.Duration) bool {
	diff := t1.Sub(t2).Abs()
	return diff <= distance
}

func createPanicHandler() (gin.HandlerFunc, *bool) {
	called := false
	handler := func(c *gin.Context) {
		called = true
		panic(fmt.Errorf("some error"))
	}
	return handler, &called
}
