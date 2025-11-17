package middleware

import (
	"net/http"

	"github.com/Jacy666/AI_Gateway/pkg/logger"
	"github.com/Jacy666/AI_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
)

// ErrorHandler middleware handles panics and errors globally
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.ErrorLogger.Printf("Panic recovered: %v", err)
				response.Error(c, http.StatusInternalServerError, "Internal server error")
				c.Abort()
			}
		}()

		c.Next()

		// Check if there were any errors during request processing
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()
			logger.ErrorLogger.Printf("Request error: %v", err.Err)

			// If no response has been written yet, write error response
			if !c.Writer.Written() {
				response.Error(c, http.StatusInternalServerError, err.Err.Error())
			}
		}
	}
}
