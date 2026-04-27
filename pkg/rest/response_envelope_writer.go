package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type envelopeResponseWriter[T any] struct {
	response ResponseEnvelope[T]
	writer   http.ResponseWriter
}

func NewResponseEnvelopeWriter[T any](w http.ResponseWriter, requestId string) *envelopeResponseWriter[T] {
	return &envelopeResponseWriter[T]{
		response: ResponseEnvelope[T]{
			RequestId: requestId,
			Status:    "SUCCESS",
		},
		writer: w,
	}
}

func (erw *envelopeResponseWriter[T]) Header() http.Header {
	return erw.writer.Header()
}

func (erw *envelopeResponseWriter[T]) Write(data T) (int, error) {
	erw.response.Details = data
	out, err := json.Marshal(erw.response)
	if err != nil {
		return 0, err
	}

	// Update Content-Length to reflect the actual wrapped payload size
	erw.writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))

	return erw.writer.Write(out)
}

func (erw *envelopeResponseWriter[T]) WriteHeader(statusCode int) {
	if statusCode < 200 || statusCode > 299 {
		erw.response.Status = "ERROR"
	} else {
		erw.response.Status = "SUCCESS"
	}
	erw.writer.WriteHeader(statusCode)
}
