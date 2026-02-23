package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/db"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/logger"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/process"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestUnit_Server_WhenAddingUnSupportedRoutes_ExpectFailure(t *testing.T) {
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
			assert.True(
				t,
				errors.IsErrorWithCode(err, UnsupportedMethod),
				"Actual err: %v",
				err,
			)
		})
	}
}

func TestUnit_Server_AnswersToRequestsWithResponseEnvelope(t *testing.T) {
	s := newTestServerWithOkHandler(t, 4001)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4001")

	err := s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assertIsOkResponse(t, response)
}

func TestUnit_Server_WhenRegisteringRawRoute_AnswersToRequestsWithoutResponseEnvelope(t *testing.T) {
	s := newTestServer(4006)
	helloHandler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "Hello")
	}
	route := rest.NewRawRoute(http.MethodGet, "/", helloHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4006")

	err = s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, "Hello", string(body))
}

func TestUnit_Server_WhenConfigDefinesABasePath_ExpectPrefixedToRoutes(t *testing.T) {
	s := newTestServerWithPath(4002, "prefix")
	route := rest.NewRoute(http.MethodGet, "/route", testHttpHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4002/prefix/route")

	err = s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assertIsOkResponse(t, response)
}

func TestUnit_Server_WhenHandlerPanics_ExpectErrorResponseEnvelope(t *testing.T) {
	s := newTestServer(4003)
	errorHandler := func(c *echo.Context) error {
		panic(fmt.Errorf("this handler panics"))
	}
	route := rest.NewRoute(http.MethodGet, "/", errorHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4003")

	err = s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	actual := unmarshalResponseAndAssertRequestId(t, response)
	assert.Equal(t, "ERROR", actual.Status)
	assert.Equal(t, `{"message":"this handler panics"}`, string(actual.Details))
}

func TestUnit_Server_WhenHandlerReturnsError_ExpectErrorResponseEnvelope(t *testing.T) {
	s := newTestServer(4004)
	errorHandler := func(c *echo.Context) error {
		return errors.NewCode(db.AlreadyCommitted)
	}
	route := rest.NewRoute(http.MethodGet, "/", errorHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4004")

	err = s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	actual := unmarshalResponseAndAssertRequestId(t, response)
	assert.Equal(t, "ERROR", actual.Status)
	assert.Equal(t, `{"message":"An unexpected error occurred. Code: 102"}`, string(actual.Details))
}

func TestUnit_Server_ExpectRequestIsProvidedALoggerWithARequestIdAsPrefix(t *testing.T) {
	s := newTestServer(4005)
	errorHandler := func(c *echo.Context) error {
		prefix := c.Logger().Prefix()
		err := uuid.Validate(prefix)
		assert.Nil(t, err, "Actual err: %v (prefix: %s)", err, prefix)
		return testHttpHandler(c)
	}
	route := rest.NewRoute(http.MethodGet, "/", errorHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	done := asyncRunServerAndAssertStopWithoutError(t, s)

	response := doRequest(t, http.MethodGet, "http://localhost:4005")

	err = s.Stop()
	<-done

	assert.Nil(t, err, "Actual err: %v", err)
	assertIsOkResponse(t, response)
}

type responseEnvelope struct {
	RequestId string          `json:"requestId"`
	Status    string          `json:"status"`
	Details   json.RawMessage `json:"details"`
}

func newTestServer(port uint16) Server {
	return newTestServerWithPath(port, "/")
}

func newTestServerWithPath(port uint16, path string) Server {
	config := Config{
		BasePath:        path,
		Port:            port,
		ShutdownTimeout: 2 * time.Second,
	}
	log := logger.New(os.Stdout)

	return NewWithLogger(config, log)
}

func newTestServerWithOkHandler(t *testing.T, port uint16) Server {
	s := newTestServer(port)

	route := rest.NewRoute(http.MethodGet, "/", testHttpHandler)
	err := s.AddRoute(route)
	assert.Nil(t, err, "Actual err: %v", err)

	return s
}

func testHttpHandler(c *echo.Context) error {
	return c.JSON(http.StatusOK, "OK")
}

func asyncRunServerAndAssertStopWithoutError(
	t *testing.T, s Server,
) <-chan struct{} {
	done := make(chan struct{}, 1)

	go func() {
		defer func() {
			done <- struct{}{}
		}()

		err := process.SafeRunSync(s.Start)
		assert.Nil(t, err, "Actual err: %v", err)
	}()

	const reasonableTimeForServerToBeUp = 50 * time.Millisecond
	time.Sleep(reasonableTimeForServerToBeUp)

	return done
}

func doRequest(
	t *testing.T, method string, url string,
) *http.Response {
	req, err := http.NewRequest(method, url, nil)
	assert.Nil(t, err, "Actual err: %v", err)

	client := &http.Client{}
	rw, err := client.Do(req)
	assert.Nil(t, err, "Actual err: %v", err)

	return rw
}

func unmarshalResponseAndAssertRequestId(t *testing.T, resp *http.Response) responseEnvelope {
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	assert.Nil(t, err, "Actual err: %v", err)

	var out responseEnvelope
	err = json.Unmarshal(data, &out)
	assert.Nil(t, err, "Actual err: %v", err)

	const idRegex = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	assert.Regexp(t, idRegex, out.RequestId)

	return out
}

func assertIsOkResponse(t *testing.T, response *http.Response) {
	assert.Equal(t, http.StatusOK, response.StatusCode)
	actual := unmarshalResponseAndAssertRequestId(t, response)
	assert.Equal(t, "SUCCESS", actual.Status)
	assert.Equal(t, `"OK"`, string(actual.Details))
}
