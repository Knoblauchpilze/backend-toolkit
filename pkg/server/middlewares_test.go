package server

import (
	"net/http"
	"testing"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestUnit_BuildMiddlewaresForRoute_ForRoute(t *testing.T) {
	r := rest.NewRoute(http.MethodGet, "/path", testHandler)

	actual := buildMiddlewaresForRoute(r, nil)

	// We can't compare functions in Go so we just check the length
	// of the middlewares slice
	assert.Len(t, actual, 4)
}

func TestUnit_BuildMiddlewaresForRoute_ForRawRoute(t *testing.T) {
	r := rest.NewRawRoute(http.MethodGet, "/path", testHandler)

	actual := buildMiddlewaresForRoute(r, nil)

	assert.Len(t, actual, 3)
}

var testHandler = func(c echo.Context) error { return nil }
