package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleRequestId = "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1"

type details struct {
	Value int `json:"value"`
}

var sampleJsonData = details{Value: 12}

func TestUnit_EnvelopeResponseWriter(t *testing.T) {
	t.Run("automatically sets success status when no status is used", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("forwards provided writer headers", func(t *testing.T) {
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
	})

	t.Run("sets status code on call to write header", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusUnauthorized)

		assert.Equal(t, http.StatusUnauthorized, out.Code)
		assert.Equal(t, http.StatusUnauthorized, rw.response.StatusCode)
		assert.Equal(t, StatusError, rw.response.Status)
	})

	t.Run("uses first status code when write header is called multiple times", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusAccepted)
		rw.WriteHeader(http.StatusUnauthorized)
		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusAccepted, out.Code)
		assert.Equal(t, http.StatusAccepted, rw.response.StatusCode)
		assert.Equal(t, StatusSuccess, rw.response.Status)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 202,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("wraps success response", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusCreated)
		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusCreated, out.Code)
		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 201,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("sets content length to match output", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusCreated)
		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		lengths, ok := rw.Header()["Content-Length"]
		require.True(t, ok, "Missing Content-Length header")
		require.Len(t, lengths, 1)

		// The length accounts for the response envelope and the JSON format
		// 12 is the length of "{"value":12}
		// 101 is the length of the response envelope wrapper"
		expectedLength := fmt.Sprintf("%d", 12+101)
		actualLength := lengths[0]

		assert.Equal(t, expectedLength, actualLength)
	})

	t.Run("wraps error response", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusUnauthorized)
		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusUnauthorized, out.Code)
		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "ERROR",
			"status_code": 401,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("decodes bytes after write header using committed status", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusAccepted)
		_, err := rw.Write([]byte(`{"value":12}`))
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, http.StatusAccepted, out.Code)
		assert.Equal(t, http.StatusAccepted, rw.response.StatusCode)
		assert.Equal(t, StatusSuccess, rw.response.Status)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 202,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("wraps plain string as details string", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.writeTyped("some-data")
		require.NoError(t, err, "Actual err: %v", err)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": "some-data"
		}`
		actual := out.Body.String()
		assert.JSONEq(t, expectedJson, actual)
	})

	t.Run("decodes json when writing bytes", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.Write([]byte(`{"value":12}`))
		require.Nil(t, err)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": {
				"value": 12
			}
		}`
		actual := out.Body.String()
		assert.JSONEq(t, expectedJson, actual)
	})

	t.Run("decodes json when writing bytes", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		value := details{Value: 45}
		data, err := json.Marshal(value)
		require.Nil(t, err, "Actual err: %v", err)

		_, err = rw.Write(data)
		require.Nil(t, err)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": {
				"value": 45
			}
		}`
		actual := out.Body.String()
		assert.JSONEq(t, expectedJson, actual)
	})

	t.Run("decodes string when writing bytes", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.Write([]byte("An error occurred"))
		require.Nil(t, err)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": "An error occurred"
		}`
		actual := out.Body.String()
		assert.JSONEq(t, expectedJson, actual)
	})

	t.Run("ignores late write header after body write", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		rw.WriteHeader(http.StatusUnauthorized)

		assert.Equal(t, http.StatusOK, out.Code)
		assert.Equal(t, http.StatusOK, rw.response.StatusCode)
		assert.Equal(t, StatusSuccess, rw.response.Status)

		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": {
				"value": 12
			}
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("sets content type and request id header on success", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, "application/json", out.Header().Get("Content-Type"))

		assert.Equal(t, sampleRequestId, out.Header().Get("X-Request-Id"))

		body := out.Body.Bytes()
		contentLength := out.Header().Get("Content-Length")
		assert.Equal(t, fmt.Sprintf("%d", len(body)), contentLength)
	})

	t.Run("sets content type and request id header on error", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		rw.WriteHeader(http.StatusInternalServerError)
		_, err := rw.writeTyped(sampleJsonData)
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, "application/json", out.Header().Get("Content-Type"))
		assert.Equal(t, sampleRequestId, out.Header().Get("X-Request-Id"))

		body := out.Body.Bytes()
		contentLength := out.Header().Get("Content-Length")
		assert.Equal(t, fmt.Sprintf("%d", len(body)), contentLength)
		assert.Equal(t, http.StatusInternalServerError, out.Code)
	})

	t.Run("sets JSON content type also when sending plain string", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.Write([]byte("plain string output"))
		require.NoError(t, err, "Actual err: %v", err)

		assert.Equal(t, "application/json", out.Header().Get("Content-Type"))
		assert.Equal(t, sampleRequestId, out.Header().Get("X-Request-Id"))

		body := out.Body.Bytes()
		contentLength := out.Header().Get("Content-Length")
		assert.Equal(t, fmt.Sprintf("%d", len(body)), contentLength)
		expectedJson := `
		{
			"request_id": "b8e9de68-3d49-4d40-a9a6-f8f3d3eab8f1",
			"status": "SUCCESS",
			"status_code": 200,
			"details": "plain string output"
		}`
		assert.JSONEq(t, expectedJson, out.Body.String())
	})

	t.Run("returns error when writing multiple times to the body", func(t *testing.T) {
		out := httptest.NewRecorder()

		rw := NewResponseEnvelopeWriter(out, sampleRequestId)

		_, err := rw.Write([]byte(`{"value":12}`))
		require.NoError(t, err, "Actual err: %v", err)

		bodyAfterFirstWrite := out.Body.String()
		contentLengthAfterFirstWrite := out.Header().Get("Content-Length")

		_, err = rw.Write([]byte(`{"value":99}`))
		require.Error(t, err)

		assert.ErrorIs(t, err, ErrMultipleBodyWrite, "Actual err: %v", err)
		assert.Equal(t, bodyAfterFirstWrite, out.Body.String())
		assert.Equal(t, contentLengthAfterFirstWrite, out.Header().Get("Content-Length"))
	})
}
