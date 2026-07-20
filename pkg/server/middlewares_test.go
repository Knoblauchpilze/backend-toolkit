package server

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnit_BuildMiddlewaresForRoute(t *testing.T) {
	t.Run("for route", func(t *testing.T) {
		r := rest.NewRoute(http.MethodGet, "/path", testHandler)

		actual := buildMiddlewaresForRoute(r)

		// We can't compare functions in Go so we just check the length
		// of the middlewares slice
		assert.Len(t, actual, 5)
	})

	t.Run("for raw route", func(t *testing.T) {
		r := rest.NewRawRoute(http.MethodGet, "/path", testHandler)

		actual := buildMiddlewaresForRoute(r)

		assert.Len(t, actual, 4)
	})
}

func TestIntegration_BuildMiddlewaresForRoute(t *testing.T) {
	t.Run("route middleware stack wraps response and keeps request id consistent", func(t *testing.T) {
		handler := func(c *echo.Context) error {
			return c.String(http.StatusOK, "Hello")
		}
		route := rest.NewRoute(http.MethodGet, "/path", handler)

		rw, err := runRouteMiddlewareChain(route)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusOK, rw.Code)
		// TODO: This should be JSON
		assert.Equal(t, "text/plain; charset=UTF-8", rw.Header().Get("Content-Type"))

		requestId := rw.Header().Get("X-Request-Id")
		assert.Regexp(t, uuidRegex, requestId)

		length, err := strconv.Atoi(rw.Header().Get("Content-Length"))
		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, rw.Body.Len(), length)

		actual := unmarshalResponseAndAssertRequestId(t, rw.Result())

		assert.Equal(t, requestId, actual.RequestId)
		assert.Equal(t, "SUCCESS", actual.Status)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
		assert.Equal(t, `"Hello"`, string(actual.Details))
	})

	t.Run("raw route middleware stack bypasses response envelope", func(t *testing.T) {
		handler := func(c *echo.Context) error {
			return c.String(http.StatusOK, "Hello")
		}
		route := rest.NewRawRoute(http.MethodGet, "/path", handler)

		rw, err := runRouteMiddlewareChain(route)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "text/plain; charset=UTF-8", rw.Header().Get("Content-Type"))
		assert.Equal(t, "Hello", rw.Body.String())

		requestId := rw.Header().Get("X-Request-Id")
		assert.Regexp(t, uuidRegex, requestId)
	})
}

func testHandler(c *echo.Context) error {
	return nil
}

func runRouteMiddlewareChain(
	route *rest.Route,
) (*httptest.ResponseRecorder, error) {
	e := echo.New()
	req := httptest.NewRequest(route.Method(), "http://example.com"+route.Path(), nil)
	rw := httptest.NewRecorder()
	ctx := e.NewContext(req, rw)

	callable := route.Handler()
	middlewares := buildMiddlewaresForRoute(route)
	for i := len(middlewares) - 1; i >= 0; i-- {
		callable = middlewares[i](callable)
	}

	err := callable(ctx)

	return rw, err
}
