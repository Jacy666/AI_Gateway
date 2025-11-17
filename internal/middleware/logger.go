package middleware

import (
	"time"

	"github.com/Jacy666/AI_Gateway/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Logger middleware logs HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		logger.InfoLogger.Printf("[%s] %s %s - %d - %v",
			clientIP, method, path, statusCode, latency)
	}
}
