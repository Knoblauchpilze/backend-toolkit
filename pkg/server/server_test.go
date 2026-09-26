package server

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
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

func TestUnit_Server(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("expect failure when adding unsupported route", func(t *testing.T) {
		s := newTestServer(4000)

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
				assert.Equal(t, ErrUnsupportedMethod, err, "Actual err: %v", err)
			})
		}
	})

	t.Run("route responds with envelope wrapping", func(t *testing.T) {
		s := newTestServerWithOkHandler(t, 4001)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4001")

		err := s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
		assertIsOkResponse(t, response)
	})

	t.Run("raw route responds with no envelope wrapping", func(t *testing.T) {
		s := newTestServer(4006)
		helloHandler := func(c *gin.Context) {
			c.String(http.StatusOK, "Hello")
		}
		route := rest.NewRawRoute(http.MethodGet, "/", helloHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4006")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
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
		s := newTestServer(4008)
		stringHandler := func(c *gin.Context) {
			c.String(http.StatusOK, "test-string-output")
		}
		route := rest.NewRoute(http.MethodGet, "/", stringHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4008")

		err = s.Stop()
		require.NoError(t, err, "Actual err: %v", err)

		<-done

		assert.Equal(t, http.StatusOK, response.StatusCode)
		assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
		assert.Regexp(t, uuidRegex, response.Header.Get(requestIdHeader))
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "SUCCESS", actual.Status)
		assert.Equal(t, http.StatusOK, actual.StatusCode)
		assert.Equal(t, `"test-string-output"`, string(actual.Details))
	})

	t.Run("base path is prefixed to route path", func(t *testing.T) {
		s := newTestServerWithPath(4002, "prefix")
		route := rest.NewRoute(http.MethodGet, "/route", testHttpHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4002/prefix/route")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
		assertIsOkResponse(t, response)
	})

	t.Run("returns error envelope when handler panics", func(t *testing.T) {
		s := newTestServer(4003)
		errorHandler := func(c *gin.Context) {
			panic(errors.New("this handler panics"))
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4003")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
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
		s := newTestServer(4007)
		errorHandler := func(c *gin.Context) {
			panic(errors.New("this handler panics"))
		}
		route := rest.NewRawRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4007")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
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
		s := newTestServer(4009)
		errorHandler := func(c *gin.Context) {
			c.JSON(http.StatusBadGateway, db.ErrAlreadyCommitted)
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4009")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, http.StatusBadGateway, response.StatusCode)
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "ERROR", actual.Status)
		assert.Equal(t, http.StatusBadGateway, actual.StatusCode)
		assert.Equal(t, `{"code":102,"message":"an unexpected error occurred"}`, string(actual.Details))
	})

	t.Run("returns error envelope when handler returns error", func(t *testing.T) {
		s := newTestServer(4004)
		errorHandler := func(c *gin.Context) {
			c.Error(db.ErrAlreadyCommitted) // nolint: errcheck
		}
		route := rest.NewRoute(http.MethodGet, "/", errorHandler)
		err := s.AddRoute(route)
		require.NoError(t, err, "Actual err: %v", err)

		done := asyncRunServerAndAssertStopWithoutError(t, s)

		response := doRequest(t, http.MethodGet, "http://localhost:4004")

		err = s.Stop()
		<-done

		require.NoError(t, err, "Actual err: %v", err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		actual := unmarshalResponseAndAssertRequestId(t, response)
		assert.Equal(t, "ERROR", actual.Status)
		assert.Equal(t, http.StatusInternalServerError, actual.StatusCode)
		assert.Equal(t, `{"message":"an unexpected error occurred. Code: 102"}`, string(actual.Details))
	})
}

func newTestServer(port uint16) *Server {
	return newTestServerWithPath(port, "/")
}

func newTestServerWithPath(port uint16, path string) *Server {
	config := Config{
		BasePath:        path,
		Port:            port,
		ShutdownTimeout: 2 * time.Second,
	}

	return NewWithLogger(config, slog.Default())
}

func newTestServerWithOkHandler(t *testing.T, port uint16) *Server {
	t.Helper()

	s := newTestServer(port)

	route := rest.NewRoute(http.MethodGet, "/", testHttpHandler)
	err := s.AddRoute(route)
	require.NoError(t, err, "Actual err: %v", err)

	return s
}

func asyncRunServerAndAssertStopWithoutError(
	t *testing.T, s *Server,
) <-chan struct{} {
	t.Helper()

	done := make(chan struct{}, 1)

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		err := process.SafeRunSync(s.Start)
		require.NoError(t, err, "Actual err: %v", err)
	}()

	const reasonableTimeForServerToBeUp = 50 * time.Millisecond
	time.Sleep(reasonableTimeForServerToBeUp)

	return done
}
