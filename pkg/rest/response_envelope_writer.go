package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ResponseEnvelopeDecoder[T any] func(data []byte) (T, error)

type envelopeResponseWriter[T any] struct {
	response ResponseEnvelope[T]
	writer   http.ResponseWriter
	decoder  ResponseEnvelopeDecoder[T]
}

func NewResponseEnvelopeWriter[T any](w http.ResponseWriter, requestId string, decoder ResponseEnvelopeDecoder[T]) *envelopeResponseWriter[T] {
	return &envelopeResponseWriter[T]{
		response: ResponseEnvelope[T]{
			RequestId: requestId,
			Status:    StatusSuccess,
		},
		writer:  w,
		decoder: decoder,
	}
}

func (erw *envelopeResponseWriter[T]) Header() http.Header {
	return erw.writer.Header()
}

func (erw *envelopeResponseWriter[T]) Write(data []byte) (int, error) {
	details, err := erw.decoder(data)
	if err != nil {
		return 0, err
	}

	return erw.WriteTyped(details)
}

func (erw *envelopeResponseWriter[T]) WriteTyped(data T) (int, error) {
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
		erw.response.Status = StatusError
	} else {
		erw.response.Status = StatusSuccess
	}
	erw.writer.WriteHeader(statusCode)
}
