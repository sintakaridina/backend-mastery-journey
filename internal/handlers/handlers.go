package handlers

import (
	"net/http"

	"grpc-firstls/internal/database"
	"grpc-firstls/internal/services"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	apiKeyService    *services.APIKeyService
	rateLimitService *services.RateLimitService
}

func NewHandler(apiKeyService *services.APIKeyService, rateLimitService *services.RateLimitService) *Handler {
	return &Handler{
		apiKeyService:    apiKeyService,
		rateLimitService: rateLimitService,
	}
}

func (h *Handler) SetupRoutes(router *gin.Engine) {
	// Health check endpoint (no rate limiting)
	router.GET("/health", h.HealthCheck)

	// API key management endpoints (admin functionality)
	admin := router.Group("/admin")
	{
		admin.POST("/api-keys", h.CreateAPIKey)
		admin.DELETE("/api-keys/:key", h.DeactivateAPIKey)
	}

	// Protected endpoints (with rate limiting)
	api := router.Group("/api")
	{
		api.GET("/status", h.GetStatus)
		api.GET("/rate-limit", h.GetRateLimitStatus)
		api.POST("/test", h.TestEndpoint)
	}
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "rate-limiter-api",
	})
}

func (h *Handler) CreateAPIKey(c *gin.Context) {
	var request struct {
		Name                   string `json:"name" binding:"required"`
		RateLimitRequests      int    `json:"rate_limit_requests"`
		RateLimitWindowSeconds int    `json:"rate_limit_window_seconds"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Set defaults if not provided
	if request.RateLimitRequests <= 0 {
		request.RateLimitRequests = 100
	}
	if request.RateLimitWindowSeconds <= 0 {
		request.RateLimitWindowSeconds = 3600 // 1 hour
	}

	apiKey, err := h.apiKeyService.CreateAPIKey(
		request.Name,
		request.RateLimitRequests,
		request.RateLimitWindowSeconds,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create API key",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"api_key": apiKey,
		"name":    request.Name,
		"rate_limit": gin.H{
			"requests":       request.RateLimitRequests,
			"window_seconds": request.RateLimitWindowSeconds,
		},
	})
}

func (h *Handler) DeactivateAPIKey(c *gin.Context) {
	apiKey := c.Param("key")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "API key required",
			"message": "Please provide an API key in the URL path",
		})
		return
	}

	err := h.apiKeyService.DeactivateAPIKey(apiKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "API key not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "API key deactivated successfully",
	})
}

func (h *Handler) GetStatus(c *gin.Context) {
	apiKey, exists := c.Get("api_key")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "API key not found in context",
		})
		return
	}

	apiKeyRecord := apiKey.(*database.APIKey)

	c.JSON(http.StatusOK, gin.H{
		"status": "authenticated",
		"api_key": gin.H{
			"id":   apiKeyRecord.ID,
			"name": apiKeyRecord.Name,
		},
	})
}

func (h *Handler) GetRateLimitStatus(c *gin.Context) {
	apiKey, exists := c.Get("api_key")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "API key not found in context",
		})
		return
	}

	apiKeyRecord := apiKey.(*database.APIKey)

	rateLimitResult, err := h.rateLimitService.GetRateLimitStatus(c.Request.Context(), apiKeyRecord)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get rate limit status",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rate_limit": gin.H{
			"limit":      rateLimitResult.Limit,
			"remaining":  rateLimitResult.Remaining,
			"reset_time": rateLimitResult.ResetTime,
			"allowed":    rateLimitResult.Allowed,
		},
	})
}

func (h *Handler) TestEndpoint(c *gin.Context) {
	apiKey, exists := c.Get("api_key")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "API key not found in context",
		})
		return
	}

	apiKeyRecord := apiKey.(*database.APIKey)

	var request struct {
		Message string `json:"message"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Request processed successfully",
		"echo":    request.Message,
		"api_key": gin.H{
			"id":   apiKeyRecord.ID,
			"name": apiKeyRecord.Name,
		},
	})
}
