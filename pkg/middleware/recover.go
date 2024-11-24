package middleware

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type recoveredErrorData struct {
	err   error
	ctx   echo.Context
	req   *http.Request
	stack []byte
}

func Recover() echo.MiddlewareFunc {
	config := middleware.RecoverConfig{
		DisableStackAll: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			data := recoveredErrorData{
				err:   err,
				ctx:   c,
				req:   c.Request(),
				stack: stack,
			}

			c.Logger().Errorf(createErrorLog(data))

			return wrapToHttpError(err)
		},
	}

	return middleware.RecoverWithConfig(config)
}

func createErrorLog(data recoveredErrorData) string {
	var out string

	out += fmt.Sprintf("%v", data.req.Method)
	out += fmt.Sprintf(" %v", pathFromRequest(data.req))
	out += fmt.Sprintf(" generated panic: %v. Stack: %v", data.err, string(data.stack))

	return out
}

func pathFromRequest(req *http.Request) string {
	host := req.Host
	// https://github.com/labstack/echo/blob/5a0b4dd8063575995cbcb746a0fb31266a0de3db/middleware/request_logger.go#L312
	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s%s", host, path)
}
