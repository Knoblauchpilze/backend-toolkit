package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ResponseEnvelopeDecoder[T any] func(data []byte) (T, error)

const requestIdHeader = "X-Request-Id"

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
	wroteBody   bool
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
	if erw.wroteBody {
		return 0, ErrMultipleBodyWrite
	}

	details, err := erw.decoder(data)
	if err != nil {
		return 0, err
	}

	return erw.writeTyped(details)
}

func (erw *envelopeResponseWriter[T]) writeTyped(data T) (int, error) {
	if erw.wroteBody {
		return 0, ErrMultipleBodyWrite
	}

	erw.response.Details = data
	out, err := json.Marshal(erw.response)
	if err != nil {
		return 0, err
	}

	// Set transport-layer headers before committing response
	erw.writer.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	erw.writer.Header().Set(requestIdHeader, erw.response.RequestId)
	erw.writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))

	erw.wroteHeader = true
	erw.wroteBody = true

	return erw.writer.Write(out)
}

func (erw *envelopeResponseWriter[T]) WriteHeader(statusCode int) {
	if erw.wroteHeader {
		return
	}

	erw.wroteHeader = true
	erw.response.StatusCode = statusCode
	erw.response.Status = statusFromHTTPCode(statusCode)

	// Set transport-layer headers at write header time to ensure they're committed
	erw.writer.Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	erw.writer.Header().Set(requestIdHeader, erw.response.RequestId)

	erw.writer.WriteHeader(statusCode)
}

func statusFromHTTPCode(statusCode int) Status {
	if statusCode >= 200 && statusCode <= 299 {
		return StatusSuccess
	}

	return StatusError
}
