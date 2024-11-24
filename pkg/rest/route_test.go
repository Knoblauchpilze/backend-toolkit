package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var testHandler = func(c echo.Context) error { return nil }

func TestUnit_Route_Method(t *testing.T) {
	r := NewRoute(http.MethodGet, "", testHandler)
	assert.Equal(t, http.MethodGet, r.Method())
}

func TestUnit_Route_Handler(t *testing.T) {
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return nil
	}

	r := NewRoute(http.MethodGet, "", handler)
	actual := r.Handler()
	actual(dummyEchoContext())

	assert.True(t, handlerCalled)
}

func TestUnit_Route_Path(t *testing.T) {
	type testCase struct {
		path     string
		expected string
	}

	tests := []testCase{
		{path: "path", expected: "/path"},
		{path: "/path", expected: "/path"},
		{path: "/path/", expected: "/path"},
		{path: "path/", expected: "/path"},
		{path: ":id", expected: "/:id"},
		{path: "/path/:id/", expected: "/path/:id"},
		{path: "/path/:id/", expected: "/path/:id"},
		{path: "path/:id/", expected: "/path/:id"},
	}

	for _, tc := range tests {
		t.Run("", func(t *testing.T) {
			r := NewRoute(http.MethodGet, tc.path, testHandler)

			assert.Equal(t, tc.expected, r.Path())
		})
	}
}

func dummyEchoContext() echo.Context {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rw := httptest.NewRecorder()

	return e.NewContext(req, rw)
}
