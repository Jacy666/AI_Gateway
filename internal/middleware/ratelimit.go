package middleware

import (
	"net/http"

	"github.com/Jacy666/AI_Gateway/internal/config"
	"github.com/Jacy666/AI_Gateway/pkg/response"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter middleware implements token bucket rate limiting
func RateLimiter(cfg *config.Config) gin.HandlerFunc {
	// Create a rate limiter using token bucket algorithm
	// golang.org/x/time/rate implements the token bucket algorithm
	limiter := rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond), cfg.RateLimit.BurstSize)

	return func(c *gin.Context) {
		// Check if request is allowed
		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RateLimiterPerIP creates a rate limiter per IP address
func RateLimiterPerIP(cfg *config.Config) gin.HandlerFunc {
	// Map to store rate limiters per IP
	limiters := make(map[string]*rate.Limiter)

	getLimiter := func(ip string) *rate.Limiter {
		limiter, exists := limiters[ip]
		if !exists {
			limiter = rate.NewLimiter(rate.Limit(cfg.RateLimit.RequestsPerSecond), cfg.RateLimit.BurstSize)
			limiters[ip] = limiter
		}
		return limiter
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := getLimiter(ip)

		if !limiter.Allow() {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded. Please try again later.")
			c.Abort()
			return
		}

		c.Next()
	}
}
