package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorConverter() HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}
		if c.IsAborted() {
			return
		}

		wrapToHttpError(c, c.Errors.Last().Err)
	}
}

func wrapToHttpError(c *gin.Context, err error) {
	c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
}
