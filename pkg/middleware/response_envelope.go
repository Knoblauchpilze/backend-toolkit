package middleware

import (
	"bufio"
	"net"
	"net/http"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ginEnvelopeWriter wraps a Gin ResponseWriter and an EnvelopeResponseWriter.
// It satisfies the gin.ResponseWriter interface so that c.Writer can be
// replaced with it: all writes go through the envelope logic while
// Gin-specific methods are delegated to the original writer.
type ginEnvelopeWriter struct {
	envelopeWriter *rest.EnvelopeResponseWriter[any]
	original       gin.ResponseWriter
}

// http.ResponseWriter methods - routed through the envelope writer
func (w *ginEnvelopeWriter) Header() http.Header              { return w.envelopeWriter.Header() }
func (w *ginEnvelopeWriter) Write(data []byte) (int, error)   { return w.envelopeWriter.Write(data) }
func (w *ginEnvelopeWriter) WriteHeader(code int)             { w.envelopeWriter.WriteHeader(code) }

// gin.ResponseWriter-specific methods - delegated to the original writer
func (w *ginEnvelopeWriter) Flush()                            { w.original.Flush() }
func (w *ginEnvelopeWriter) Status() int                       { return w.original.Status() }
func (w *ginEnvelopeWriter) Size() int                         { return w.original.Size() }
func (w *ginEnvelopeWriter) Written() bool                     { return w.original.Written() }
func (w *ginEnvelopeWriter) WriteHeaderNow()                   { w.original.WriteHeaderNow() }
func (w *ginEnvelopeWriter) Pusher() http.Pusher               { return w.original.Pusher() }
func (w *ginEnvelopeWriter) WriteString(s string) (int, error) { return w.envelopeWriter.Write([]byte(s)) }

// http.Hijacker - delegated to the original writer
func (w *ginEnvelopeWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.original.Hijack()
}

// http.CloseNotifier - delegated to the original writer
//
//nolint:staticcheck // CloseNotify is part of the gin.ResponseWriter interface.
func (w *ginEnvelopeWriter) CloseNotify() <-chan bool {
	return w.original.CloseNotify()
}

// ResponseEnvelope generates a unique request ID (or reuses an existing
// X-Request-Id header from the incoming request), sets it as the X-Request-Id
// response header, and wraps the response writer so that every response body
// is automatically enclosed in a ResponseEnvelope JSON structure.
func ResponseEnvelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader(requestIdHeader)
		if requestId == "" {
			requestId = uuid.New().String()
		}
		c.Header(requestIdHeader, requestId)

		originalWriter := c.Writer
		envelopeWriter := rest.NewResponseEnvelopeWriter(originalWriter, requestId, rest.DecodeJSONOrString)
		c.Writer = &ginEnvelopeWriter{
			envelopeWriter: envelopeWriter,
			original:       originalWriter,
		}

		c.Next()
	}
}
