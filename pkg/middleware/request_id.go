package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const requestIdHeader = "X-Request-Id"

func RequestId() HandlerFunc {
	return func(c *gin.Context) {
		requestId := c.GetHeader(requestIdHeader)
		if requestId == "" {
			requestId = uuid.NewString()
		}

		c.Header(requestIdHeader, requestId)
		ctx := rest.WithContextRequestId(c.Request.Context(), requestId)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
