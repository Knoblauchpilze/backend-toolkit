package middleware

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_Recover_CallsNextMiddleware(t *testing.T) {
	callable, called, ctx := createCallableHandler(Recover)

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_Recover_PreventsPanic(t *testing.T) {
	next, called := createPanicHandler()

	middleware := Recover()
	callable := middleware(next)

	ctx, _ := generateTestEchoContext()

	err := callable(ctx)

	assert.Nil(t, err)
	assert.True(t, *called)
}

func TestUnit_Recover_LogsError(t *testing.T) {
	next, _ := createPanicHandler()

	middleware := Recover()
	callable := middleware(next)

	ctx, out := generateTestEchoContextWithLogger()

	err := callable(ctx)
	require.Nil(t, err)
	afterCall := time.Now()

	actual := unmarshalLogOutput(t, *out)
	assert.Equal(t, "error", actual.Level)
	safetyMargin := 5 * time.Second
	assert.True(t, areTimeCloserThan(actual.Time, afterCall, safetyMargin), "%v and %v are not within %v", afterCall, actual.Time, safetyMargin)
	// https://golangforall.com/en/post/golang-regexp-matching-newline.html
	assert.Regexp(t, "GET example.com/ generated panic: some error. Stack: [[:graph:]\\s]*", actual.Message)
}

func TestUnit_Recover_SetsStatusCodeToError(t *testing.T) {
	next, _ := createPanicHandler()

	middleware := Recover()
	callable := middleware(next)

	ctx, rw := generateTestEchoContext()

	err := callable(ctx)
	require.Nil(t, err)

	assert.Equal(t, http.StatusInternalServerError, rw.Code)
	body, err := io.ReadAll(rw.Body)
	require.Nil(t, err)
	expected := `
	{
		"message":"some error"
	}`
	assert.JSONEq(t, expected, string(body))
}

func createPanicHandler() (echo.HandlerFunc, *bool) {
	var called bool
	handler := func(c echo.Context) error {
		called = true
		panic(fmt.Errorf("some error"))
	}

	return handler, &called
}
