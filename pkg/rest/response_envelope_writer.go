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

	// wroteHeader controls whether either WriteHeader or writing the response body
	// was already called. In both cases, the response's status is committed to the
	// value it had at that point, making following updates meaningless. To avoid
	// changing the Status/StatusCode to values that do not reflect the underlying
	// HTTP response's values, this field, when true, prevents any changes to any
	// of those two values.
	wroteHeader bool
}

func NewResponseEnvelopeWriter[T any](w http.ResponseWriter, requestId string, decoder ResponseEnvelopeDecoder[T]) *envelopeResponseWriter[T] {
	return &envelopeResponseWriter[T]{
		response: ResponseEnvelope[T]{
			RequestId:  requestId,
			Status:     StatusSuccess,
			StatusCode: http.StatusOK,
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

	erw.wroteHeader = true

	// Update Content-Length to reflect the actual wrapped payload size
	erw.writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))

	return erw.writer.Write(out)
}

func (erw *envelopeResponseWriter[T]) WriteHeader(statusCode int) {
	if erw.wroteHeader {
		return
	}

	erw.wroteHeader = true
	erw.response.StatusCode = statusCode
	erw.response.Status = statusFromHTTPCode(statusCode)
	erw.writer.WriteHeader(statusCode)
}

func statusFromHTTPCode(statusCode int) Status {
	if statusCode >= 200 && statusCode <= 299 {
		return StatusSuccess
	}

	return StatusError
}
