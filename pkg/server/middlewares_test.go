package server

import (
	"net/http"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
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

var testHandler = func(c *echo.Context) error { return nil }
