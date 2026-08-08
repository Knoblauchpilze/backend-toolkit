package middleware

import (
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
)

const (
	defaultRequestId = "unknown"
)

type ginEnvelopeWriter struct {
	gin.ResponseWriter
	envelopeWriter *rest.EnvelopeResponseWriter
}

func ResponseEnvelope() HandlerFunc {
	return func(c *gin.Context) {
		requestId, ok := rest.RequestIdFromContext(c.Request.Context())
		if !ok {
			requestId = defaultRequestId
		}

		c.Writer = &ginEnvelopeWriter{
			ResponseWriter: c.Writer,
			envelopeWriter: rest.NewResponseEnvelopeWriter(c.Writer, requestId),
		}

		c.Next()
	}
}

func (w *ginEnvelopeWriter) Header() http.Header {
	return w.envelopeWriter.Header()
}

func (w *ginEnvelopeWriter) Write(data []byte) (int, error) {
	return w.envelopeWriter.Write(data)
}

func (w *ginEnvelopeWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}

func (w *ginEnvelopeWriter) WriteHeader(statusCode int) {
	w.envelopeWriter.WriteHeader(statusCode)
}
