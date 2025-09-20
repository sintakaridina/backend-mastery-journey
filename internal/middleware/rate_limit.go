package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"grpc-firstls/internal/services"

	"github.com/gin-gonic/gin"
)

func RateLimit(apiKeyService *services.APIKeyService, rateLimitService *services.RateLimitService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for health check endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// Try Authorization header as fallback
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "API key required",
				"message": "Please provide an API key in the X-API-Key header or Authorization header",
			})
			c.Abort()
			return
		}

		// Validate API key
		apiKeyRecord, err := apiKeyService.ValidateAPIKey(apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid API key",
				"message": "The provided API key is invalid or inactive",
			})
			c.Abort()
			return
		}

		// Check rate limit
		rateLimitResult, err := rateLimitService.CheckRateLimit(c.Request.Context(), apiKeyRecord)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Rate limit check failed",
				"message": "Unable to check rate limit",
			})
			c.Abort()
			return
		}

	// Add rate limit headers
	c.Header("X-RateLimit-Limit", strconv.FormatInt(rateLimitResult.Limit, 10))
	c.Header("X-RateLimit-Remaining", strconv.FormatInt(rateLimitResult.Remaining, 10))
	c.Header("X-RateLimit-Reset", rateLimitResult.ResetTime.Format(time.RFC3339))

		// Check if rate limit exceeded
		if !rateLimitResult.Allowed {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "You have exceeded your rate limit. Please try again later.",
				"retry_after": int(time.Until(rateLimitResult.ResetTime).Seconds()),
			})
			c.Abort()
			return
		}

		// Store API key info in context for use in handlers
		c.Set("api_key", apiKeyRecord)
		c.Next()
	}
}
