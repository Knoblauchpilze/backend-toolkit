package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/labstack/echo/v5"
)

type recoveredErrorData struct {
	err   error
	ctx   *echo.Context
	req   *http.Request
	stack []byte
}

func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					recoveredErr, ok := r.(error)
					if !ok {
						recoveredErr = fmt.Errorf("%v", r)
					}

					stack := make([]byte, 4<<10) // 4 KB
					length := runtime.Stack(stack, false)

					data := recoveredErrorData{
						err:   recoveredErr,
						ctx:   c,
						req:   c.Request(),
						stack: stack[:length],
					}

					c.Logger().Error(createErrorLog(data))

					err = wrapToHttpError(recoveredErr)
				}
			}()
			return next(c)
		}
	}
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
	path := req.URL.Path
	if path == "" {
		path = "/"
	}

	return fmt.Sprintf("%s%s", host, path)
}
