package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnit_SanitizePath(t *testing.T) {
	type testCase struct {
		in       string
		expected string
	}

	testCases := []testCase{
		{in: "", expected: "/"},
		{in: "/", expected: "/"},
		{in: "//", expected: "/"},
		{in: "path", expected: "/path"},
		{in: "path/", expected: "/path"},
		{in: "path//", expected: "/path"},
		{in: "/path", expected: "/path"},
		{in: "//path", expected: "/path"},
		{in: "/path/", expected: "/path"},
		{in: "//path/", expected: "/path"},
		{in: "/path//", expected: "/path"},
		{in: "//path//", expected: "/path"},
		{in: "path/id", expected: "/path/id"},
		{in: "path//id", expected: "/path/id"},
		{in: "path/id/", expected: "/path/id"},
		{in: "/path/id", expected: "/path/id"},
		{in: "/path/id/", expected: "/path/id"},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			actual := sanitizePath(testCase.in)

			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestUnit_ConcatenateEndpoints(t *testing.T) {
	type testCase struct {
		basePath string
		path     string
		expected string
	}

	testCases := []testCase{
		{basePath: "", path: "", expected: "/"},
		{basePath: "", path: "/some/path", expected: "/some/path"},
		{basePath: "/some/path", path: "", expected: "/some/path"},
		{basePath: "/some/endpoint", path: "/some/path", expected: "/some/endpoint/some/path"},
		{basePath: "/some/endpoint", path: "some/path", expected: "/some/endpoint/some/path"},
		{basePath: "some/endpoint", path: "some/path", expected: "/some/endpoint/some/path"},
		{basePath: "some/endpoint", path: "/path/", expected: "/some/endpoint/path"},
		{basePath: "/some/endpoint", path: "/path/", expected: "/some/endpoint/path"},
		{basePath: "/some/endpoint/", path: "/path/", expected: "/some/endpoint/path"},
		{basePath: "some/endpoint", path: "path/", expected: "/some/endpoint/path"},
	}

	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			actual := ConcatenateEndpoints(testCase.basePath, testCase.path)

			assert.Equal(t, testCase.expected, actual)
		})
	}
}
