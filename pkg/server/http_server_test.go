package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	uuidRegex       = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	requestIdHeader = "X-Request-Id"
)

func TestUnit_HttpServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("expect failure when adding unsupported route", func(t *testing.T) {
		s := newTestHttpServer()

		unsupportedMethods := []string{
			http.MethodHead,
			http.MethodPut,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		}

		for _, method := range unsupportedMethods {
			t.Run(method, func(t *testing.T) {
				sampleRoute := rest.NewRoute(method, "/", testHttpHandler)
				err := s.AddRoute(sampleRoute)
				assert.ErrorIs(t, err, ErrUnsupportedMethod, "Actual err: %v", err)
			})
		}
	})

	t.Run("route responds with envelope wrapping", func(t *testing.T) {
		s := newTestHttpServer()

		route := rest.NewRoute(http.MethodGet, "/", testHttpHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assertIsOkResponse(t, response)
	})

	t.Run("raw route responds with no envelope wrapping", func(t *testing.T) {
		s := newTestHttpServer()
		helloHandler := func(c *gin.Context) {
			c.String(http.StatusOK, "Hello")
		}
		route := rest.NewRawRoute(http.MethodGet, "/", helloHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Regexp(t, uuidRegex, response.Header.Get(requestIdHeader))
		body, err := io.ReadAll(response.Body)
		require.NoError(t, err, "Actual err: %v", err)
		assertResponseContentLengthMatchesBody(t, response, body)
		assert.Equal(t, "Hello", string(body))
		err = response.Body.Close()
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("route with string output returns enveloped json response", func(t *testing.T) {
		s := newTestHttpServer()
		stringHandler := func(c *gin.Context) {
			c.String(http.StatusOK, "test-string-output")
		}
		route := rest.NewRoute(http.MethodGet, "/", stringHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
		assert.Regexp(t, uuidRegex, response.Header.Get(requestIdHeader))
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "SUCCESS", actual.Status)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
		assert.Equal(t, `"test-string-output"`, string(actual.Details))
	})

	t.Run("base path is prefixed to route path", func(t *testing.T) {
		s := newTestHttpServerWithPath("prefix")
		route := rest.NewRoute(http.MethodGet, "/route", testHttpHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url+"/prefix/route")

		assertIsOkResponse(t, response)
	})

	t.Run("returns error envelope when handler panics", func(t *testing.T) {
		s := newTestHttpServer()
		errorHandler := func(c *gin.Context) {
			panic(errors.New("this handler panics"))
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "ERROR", actual.Status)
		assert.Equal(t, http.StatusInternalServerError, actual.StatusCode)
		assert.Equal(
			t,
			`{"message":"an unexpected error occurred. Code: 400 (cause: this handler panics)"}`,
			string(actual.Details),
		)
	})

	t.Run("returns non-envelope error when raw route handler panics", func(t *testing.T) {
		s := newTestHttpServer()
		errorHandler := func(c *gin.Context) {
			panic(errors.New("this handler panics"))
		}
		route := rest.NewRawRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Regexp(t, uuidRegex, response.Header.Get(requestIdHeader))
		body, err := io.ReadAll(response.Body)
		require.NoError(t, err, "Actual err: %v", err)
		assertResponseContentLengthMatchesBody(t, response, body)
		assert.JSONEq(
			t,
			`{"message":"an unexpected error occurred. Code: 400 (cause: this handler panics)"}`,
			string(body),
		)
		err = response.Body.Close()
		require.NoError(t, err, "Actual err: %v", err)
	})

	t.Run("returns error envelope when handler returns error as JSON", func(t *testing.T) {
		s := newTestHttpServer()
		errorHandler := func(c *gin.Context) {
			c.JSON(http.StatusBadGateway, db.ErrAlreadyCommitted)
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusBadGateway, response.StatusCode)
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "ERROR", actual.Status)
		assert.Equal(t, http.StatusBadGateway, actual.StatusCode)
		assert.Equal(t, `{"code":102,"message":"an unexpected error occurred"}`, string(actual.Details))
	})

	t.Run("returns error envelope when handler returns error", func(t *testing.T) {
		s := newTestHttpServer()
		errorHandler := func(c *gin.Context) {
			c.Error(db.ErrAlreadyCommitted) // nolint: errcheck
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		url := asyncRunHttpServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, url)

		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "ERROR", actual.Status)
		assert.Equal(t, http.StatusInternalServerError, actual.StatusCode)
		assert.Equal(t, `{"message":"an unexpected error occurred. Code: 102"}`, string(actual.Details))
	})
}

func newTestHttpServer() *HttpServer {
	return newTestHttpServerWithPath("/")
}

func newTestHttpServerWithPath(path string) *HttpServer {
	config := Config{
		BasePath:        path,
		ShutdownTimeout: 2 * time.Second,
	}

	return NewWithLogger(config, slog.Default())
}

// Returns the base url the server listens on. Serving is stopped and asserted
// to have terminated without error when the test ends.
func asyncRunHttpServerAndAssertStopWithoutError(
	t *testing.T, s *HttpServer,
) string {
	t.Helper()

	listener, err := s.Bind(0)
	require.NoError(t, err, "Actual err: %v", err)

	ctx, cancel := context.WithCancel(context.Background())
	serveErr := make(chan error, 1)

	go func() {
		serveErr <- process.SafeRunSync(func() error {
			return s.Serve(ctx, listener)
		})
	}()

	t.Cleanup(func() {
		cancel()
		err := <-serveErr
		require.NoError(t, err, "Actual err: %v", err)
	})

	addr, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return fmt.Sprintf("http://localhost:%d", addr.Port)
}

func testHttpHandler(c *gin.Context) {
	c.JSON(http.StatusOK, "OK")
}

func doRequest(
	t *testing.T, method string, url string,
) *http.Response {
	t.Helper()

	req, err := http.NewRequest(method, url, nil)
	require.NoError(t, err, "Actual err: %v", err)

	client := &http.Client{}
	rw, err := client.Do(req)
	require.NoError(t, err, "Actual err: %v", err)

	return rw
}

type responseEnvelope struct {
	RequestId  string          `json:"request_id"`
	Status     string          `json:"status"`
	StatusCode int             `json:"status_code"`
	Details    json.RawMessage `json:"details"`
}

func unmarshalResponseAndAssertRequestId(t *testing.T, resp *http.Response) responseEnvelope {
	t.Helper()

	defer func() {
		err := resp.Body.Close()
		require.NoError(t, err, "Actual err: %v", err)
	}()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "Actual err: %v", err)
	assertResponseContentLengthMatchesBody(t, resp, data)

	var out responseEnvelope
	err = json.Unmarshal(data, &out)
	require.NoError(t, err, "Actual err: %v", err)

	headerRequestId := resp.Header.Get(requestIdHeader)
	assert.Regexp(t, uuidRegex, headerRequestId)
	assert.Regexp(t, uuidRegex, out.RequestId)
	assert.Equal(t, headerRequestId, out.RequestId)

	return out
}

func assertResponseContentLengthMatchesBody(
	t *testing.T,
	resp *http.Response,
	body []byte,
) {
	t.Helper()

	contentLength, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	require.NoError(t, err, "Actual err: %v", err)
	assert.Equal(t, len(body), contentLength)
}

func assertIsOkResponse(t *testing.T, response *http.Response) {
	t.Helper()

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
	actual := unmarshalResponseAndAssertRequestId(t, response)
	assert.Equal(t, "SUCCESS", actual.Status)
	assert.Equal(t, http.StatusOK, actual.StatusCode)
	assert.Equal(t, `"OK"`, string(actual.Details))
}
