package rest

import (
	"encoding/json"
	"net/http"
)

type envelopeResponseWriter struct {
	response responseEnvelope
	writer   http.ResponseWriter
}

func NewResponseEnvelopeWriter(w http.ResponseWriter, requestId string) *envelopeResponseWriter {
	return &envelopeResponseWriter{
		response: responseEnvelope{
			RequestId: requestId,
			Status:    "SUCCESS",
		},
		writer: w,
	}
}

func (erw *envelopeResponseWriter) Header() http.Header {
	return erw.writer.Header()
}

func (erw *envelopeResponseWriter) Write(data []byte) (int, error) {
	erw.response.Details = data
	out, err := json.Marshal(erw.response)
	if err != nil {
		// Attempt to marshal as string
		asString := string(data)
		encodedData, err := json.Marshal(&asString)
		if err != nil {
			// Fallback to writing no response envelope
			return erw.writer.Write(data)
		}

		return erw.Write(encodedData)
	}

	return erw.writer.Write(out)
}

func (erw *envelopeResponseWriter) WriteHeader(statusCode int) {
	if statusCode < 200 || statusCode > 299 {
		erw.response.Status = "ERROR"
	} else {
		erw.response.Status = "SUCCESS"
	}
	erw.writer.WriteHeader(statusCode)
}
