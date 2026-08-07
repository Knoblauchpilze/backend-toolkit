package rest

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
)

const (
	requestIdHeader = "X-Request-Id"

	errMultipleBodyWrite errors.ErrorCode = 350
)

var (
	ErrMultipleBodyWrite = errors.FromCodeAndDetails(errMultipleBodyWrite, "response body already written")
)

type EnvelopeResponseWriter struct {
	response ResponseEnvelope[any]
	writer   http.ResponseWriter

	// wroteHeader controls whether either WriteHeader or writing the response body
	// was already called. In both cases, the response's status is committed to the
	// value it had at that point, making following updates meaningless. To avoid
	// changing the Status/StatusCode to values that do not reflect the underlying
	// HTTP response's values, this field, when true, prevents any changes to any
	// of those two values.
	wroteHeader bool
	wroteBody   bool
}

func NewResponseEnvelopeWriter(w http.ResponseWriter, requestId string) *EnvelopeResponseWriter {
	return &EnvelopeResponseWriter{
		response: ResponseEnvelope[any]{
			RequestId:  requestId,
			Status:     StatusSuccess,
			StatusCode: http.StatusOK,
		},
		writer: w,
	}
}

func (erw *EnvelopeResponseWriter) Header() http.Header {
	return erw.writer.Header()
}

func (erw *EnvelopeResponseWriter) Write(data []byte) (int, error) {
	if erw.wroteBody {
		return 0, ErrMultipleBodyWrite
	}

	details, err := decodeJSONOrString(data)
	if err != nil {
		return 0, err
	}

	return erw.writeTyped(details)
}

func (erw *EnvelopeResponseWriter) writeTyped(data any) (int, error) {
	if erw.wroteBody {
		return 0, ErrMultipleBodyWrite
	}

	erw.response.Details = data
	out, err := json.Marshal(erw.response)
	if err != nil {
		return 0, err
	}

	// Set transport-layer headers before committing response
	erw.writer.Header().Set("Content-Type", "application/json")
	erw.writer.Header().Set(requestIdHeader, erw.response.RequestId)
	erw.writer.Header().Set("Content-Length", fmt.Sprintf("%d", len(out)))

	erw.wroteHeader = true
	erw.wroteBody = true

	return erw.writer.Write(out)
}

func (erw *EnvelopeResponseWriter) WriteHeader(statusCode int) {
	if erw.wroteHeader {
		return
	}

	erw.wroteHeader = true
	erw.response.StatusCode = statusCode
	erw.response.Status = statusFromHTTPCode(statusCode)

	// Set transport-layer headers at write header time to ensure they're committed
	erw.writer.Header().Set("Content-Type", "application/json")
	erw.writer.Header().Set(requestIdHeader, erw.response.RequestId)

	erw.writer.WriteHeader(statusCode)
}

func statusFromHTTPCode(statusCode int) Status {
	if statusCode >= 200 && statusCode <= 299 {
		return StatusSuccess
	}

	return StatusError
}
