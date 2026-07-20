package middleware

import (
	"github.com/gin-gonic/gin"
)

// ErrorConverter converts errors stored in the Gin context (via c.Error) into
// proper HTTP error responses with a JSON body of the form {"message": "..."}.
func ErrorConverter() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		httpErr := wrapToHttpError(err)

		c.JSON(httpErr.Code, gin.H{"message": httpErr.Message})
	}
}
