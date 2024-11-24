package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const sampleRequestId = "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1"

var sampleJsonData = []byte(`{"value":12}`)

func TestUnit_EnvelopeResponseWriter_AutomaticallySetsSuccessStatusWhenNoStatusIsUsed(t *testing.T) {
	out := httptest.NewRecorder()

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)

	rw.Write(sampleJsonData)

	expectedJson := `
	{
		"requestId": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
		"status": "SUCCESS",
		"details": {
			"value": 12
		}
	}`
	assert.JSONEq(t, expectedJson, out.Body.String())
}

func TestUnit_EnvelopeResponseWriter_ForwardsProvidedWriterHeaders(t *testing.T) {
	out := httptest.NewRecorder()
	out.Header().Add("Key1", "val1")
	out.Header().Add("Key1", "val2")

	out.Header().Add("Key2", "other-value")

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)
	actual := rw.Header()

	expected := http.Header{
		"Key1": []string{"val1", "val2"},
		"Key2": []string{"other-value"},
	}
	assert.Equal(t, expected, actual)
}

func TestUnit_EnvelopeResponseWriter_SetsStatusCodeOnCallToWriteHeader(t *testing.T) {
	out := httptest.NewRecorder()

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)

	rw.WriteHeader(http.StatusUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, out.Code)
}

func TestUnit_EnvelopeResponseWriter_WrapsSuccessResponse(t *testing.T) {
	out := httptest.NewRecorder()

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)

	rw.WriteHeader(http.StatusCreated)
	rw.Write(sampleJsonData)

	assert.Equal(t, http.StatusCreated, out.Code)
	expectedJson := `
	{
		"requestId": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
		"status": "SUCCESS",
		"details": {
			"value": 12
		}
	}`
	assert.JSONEq(t, expectedJson, out.Body.String())
}

func TestUnit_EnvelopeResponseWriter_WrapsErrorResponse(t *testing.T) {
	out := httptest.NewRecorder()

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)

	rw.WriteHeader(http.StatusUnauthorized)
	rw.Write(sampleJsonData)

	assert.Equal(t, http.StatusUnauthorized, out.Code)
	expectedJson := `
	{
		"requestId": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
		"status": "ERROR",
		"details": {
			"value": 12
		}
	}`
	assert.JSONEq(t, expectedJson, out.Body.String())
}

func TestUnit_EnvelopeResponseWriter_WrapsPlainDataAsDetailsString(t *testing.T) {
	out := httptest.NewRecorder()

	rw := NewResponseEnvelopeWriter(out, sampleRequestId)

	rw.Write([]byte("some-data"))

	expectedJson := `
	{
		"requestId": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
		"status": "SUCCESS",
		"details": "some-data"
	}`
	actual := out.Body.String()
	assert.JSONEq(t, expectedJson, actual)
}
