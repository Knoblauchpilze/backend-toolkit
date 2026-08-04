package middleware

import (
	stderrors "errors"
	"fmt"
	"net/http"
	"runtime"

	"github.com/Knoblauchpilze/backend-toolkit/pkg/errors"
	"github.com/Knoblauchpilze/backend-toolkit/pkg/rest"
	"github.com/labstack/echo/v5"
)

const (
	errUncaughtPanic errors.ErrorCode = 400
)

type recoveredErrorData struct {
	err   error
	ctx   *echo.Context
	req   *http.Request
	stack []byte
}

func Recover() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c *echo.Context) (err error) {
			defer func() {
				if r := recover(); r != nil {
					recoveredErr, ok := r.(error)
					if !ok {
						recoveredErr = fmt.Errorf("%v", r)
					}

					if _, ok := recoveredErr.(*errors.ErrorWithCode); !ok {
						recoveredErr = errors.WrapCode(recoveredErr, errUncaughtPanic)
					}

					stack := make([]byte, 4<<10) // 4 KB
					length := runtime.Stack(stack, false)

					data := recoveredErrorData{
						err:   recoveredErr,
						ctx:   c,
						req:   c.Request(),
						stack: stack[:length],
					}

					log := rest.GetContextLogger(c.Request().Context())
					log.Error(createErrorLog(data))

					err = wrapToHttpError(recoveredErr)
				}
			}()

			err = next(c)
			return wrapToHttpError(err)
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

func wrapToHttpError(err error) error {
	if err == nil {
		return nil
	}

	var httpErr *echo.HTTPError
	if stderrors.As(err, &httpErr) {
		return err
	}

	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
}
