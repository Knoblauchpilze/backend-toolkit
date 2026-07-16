package middleware

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
)

// Recover catches panics that occur inside handlers and converts them into
// HTTP 500 errors stored in the Gin context for ErrorConverter to handle.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				recoveredErr, ok := r.(error)
				if !ok {
					recoveredErr = fmt.Errorf("%v", r)
				}

				stack := make([]byte, 4<<10) // 4 KB
				length := runtime.Stack(stack, false)

				log := GetContextLogger(c)
				log.Error(createErrorLog(c.Request, recoveredErr, stack[:length]))

				_ = c.Error(recoveredErr)
				c.Abort()
			}
		}()
		c.Next()
	}
}

func createErrorLog(req *http.Request, err error, stack []byte) string {
	var out string

	out += fmt.Sprintf("%v", req.Method)
	out += fmt.Sprintf(" %v", pathFromRequest(req))
	out += fmt.Sprintf(" generated panic: %v. Stack: %v", err, string(stack))

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
