package middleware

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/utils"
)

// Recovery protects the process from handler panics and converts them to a
// JSON 500 envelope.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[panic] %v", r)
				utils.Server(c, "internal server error")
			}
		}()
		c.Next()
	}
}
