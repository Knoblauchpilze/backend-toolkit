package middleware

import (
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

const requestIdHeader = "X-Request-Id"

func RequestId() echo.MiddlewareFunc {
	cfg := middleware.RequestIDConfig{
		Generator:    func() string { return uuid.New().String() },
		TargetHeader: requestIdHeader,
		RequestIDHandler: func(c *echo.Context, requestId string) {
			req := c.Request()
			ctx := rest.WithContextRequestId(req.Context(), requestId)
			c.SetRequest(req.WithContext(ctx))
		},
	}

	return middleware.RequestIDWithConfig(cfg)
}
