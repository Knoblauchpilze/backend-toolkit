package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var sampleUuid = uuid.MustParse("08ce96a3-3430-48a8-a3b2-b1c987a207ca")

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

func TestUnit_MarshalNilToEmptySlice_WhenNil_ExpectMarshalToEmptySlice(t *testing.T) {
	assert := assert.New(t)

	var in []int

	actual, err := MarshalNilToEmptySlice(in)

	assert.Nil(err)
	assert.Equal("[]", string(actual))
}

func TestUnit_MarshalNilToEmptySlice_WhenNotNil_ExpectMarshalCorrectData(t *testing.T) {
	assert := assert.New(t)

	in := []int{1, 2}

	actual, err := MarshalNilToEmptySlice(in)

	assert.Nil(err)
	assert.Equal("[1,2]", string(actual))
}

var defaultKey = "my-key"

func TestUnit_FetchIdFromQueryParam_whenNoId_expectNotExistAndNoError(t *testing.T) {
	assert := assert.New(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx, _ := generateTestEchoContextFromRequest(req)

	exists, _, err := FetchIdFromQueryParam(defaultKey, ctx)
	assert.False(exists)
	assert.Nil(err)
}

func TestUnit_FetchIdFromQueryParam_whenIdSetForOtherKey_expectNotExistAndNoError(t *testing.T) {
	assert := assert.New(t)

	req := generateRequestWithQueryParams("not-the-default-key", sampleUuid.String())
	ctx, _ := generateTestEchoContextFromRequest(req)

	exists, _, err := FetchIdFromQueryParam(defaultKey, ctx)
	assert.False(exists)
	assert.Nil(err)
}

func TestUnit_FetchIdFromQueryParam_whenIdSyntaxIsWrong_expectExistAndError(t *testing.T) {
	assert := assert.New(t)

	req := generateRequestWithQueryParams(defaultKey, "not-a-uuid")
	ctx, _ := generateTestEchoContextFromRequest(req)

	exists, _, err := FetchIdFromQueryParam(defaultKey, ctx)
	assert.True(exists)
	assert.Equal("invalid UUID length: 10", err.Error())
}

func TestUnit_FetchIdFromQueryParam_whenIdIsSet_expectExistCorrectIdAndNoError(t *testing.T) {
	assert := assert.New(t)

	req := generateRequestWithQueryParams(defaultKey, sampleUuid.String())
	ctx, _ := generateTestEchoContextFromRequest(req)

	exists, actual, err := FetchIdFromQueryParam(defaultKey, ctx)
	assert.True(exists)
	assert.Equal(sampleUuid, actual)
	assert.Nil(err)
}

func generateRequestWithQueryParams(key string, value string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	q := req.URL.Query()
	q.Add(key, value)

	req.URL.RawQuery = q.Encode()

	return req
}

func generateTestEchoContextFromRequest(
	req *http.Request,
) (echo.Context, *httptest.ResponseRecorder) {
	e := echo.New()
	rw := httptest.NewRecorder()

	ctx := e.NewContext(req, rw)
	return ctx, rw
}
