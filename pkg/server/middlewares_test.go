package server

import (
	"net/http"
	"net/http/httptest"
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
		assert.Len(t, actual, 1)
	})

	t.Run("for raw route", func(t *testing.T) {
		r := rest.NewRawRoute(http.MethodGet, "/path", testHandler)

		actual := buildMiddlewaresForRoute(r)

		assert.Len(t, actual, 0)
	})
}

func TestIT_BuildMiddlewaresForRoute(t *testing.T) {
	t.Run("route middleware stack wraps response", func(t *testing.T) {
		handler := func(c *echo.Context) error {
			return c.String(http.StatusOK, "Hello")
		}
		route := rest.NewRoute(http.MethodGet, "/path", handler)

		rw, err := runRouteMiddlewareChain(route)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusOK, rw.Code)
		// TODO: This should be JSON
		assert.Equal(t, "text/plain; charset=UTF-8", rw.Header().Get("Content-Type"))
		assert.Contains(t, rw.Body.String(), `"status":"SUCCESS"`)
		assert.Contains(t, rw.Body.String(), `"status_code":200`)
		assert.Contains(t, rw.Body.String(), `"details":"Hello"`)
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
	middlewares := buildMiddlewaresForRoute(route)

	switch route.Method() {
	case http.MethodGet:
		e.GET(route.Path(), route.Handler(), middlewares...)
	case http.MethodPost:
		e.POST(route.Path(), route.Handler(), middlewares...)
	case http.MethodDelete:
		e.DELETE(route.Path(), route.Handler(), middlewares...)
	case http.MethodPatch:
		e.PATCH(route.Path(), route.Handler(), middlewares...)
	default:
		return nil, ErrUnsupportedMethod
	}

	e.ServeHTTP(rw, req)

	return rw, nil
}
