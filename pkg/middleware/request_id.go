package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const requestIdHeader = "X-Request-Id"

func RequestId() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			requestId := req.Header.Get(requestIdHeader)
			if requestId == "" {
				requestId = uuid.NewString()
			}

			c.Response().Header().Set(requestIdHeader, requestId)
			ctx := rest.WithContextRequestId(req.Context(), requestId)
			c.SetRequest(req.WithContext(ctx))

			return next(c)
		}
	}
}
