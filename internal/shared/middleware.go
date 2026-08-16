
package shared

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandleMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		defer func() {
			if r := recover(); r != nil {

				log.Printf("[PANIC RECOVERED] %v", r)

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()

		c.Next()

		if len(c.Errors) > 0 {

			err := c.Errors.Last()

			log.Printf("[ERROR] %s %s → %v", c.Request.Method, c.Request.URL.Path, err)

			if c.Writer.Written() {
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
	}
}
