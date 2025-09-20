package main

import (
	"log"
	"os"

	"grpc-firstls/internal/config"
	"grpc-firstls/internal/database"
	"grpc-firstls/internal/handlers"
	"grpc-firstls/internal/middleware"
	"grpc-firstls/internal/redis"
	"grpc-firstls/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	// Initialize services
	apiKeyService := services.NewAPIKeyService(db)
	rateLimitService := services.NewRateLimitService(redisClient, cfg.RateLimitConfig)

	// Initialize handlers
	handler := handlers.NewHandler(apiKeyService, rateLimitService)

	// Setup router
	router := gin.Default()

	// Add middleware
	router.Use(middleware.CORS())
	router.Use(middleware.RateLimit(apiKeyService, rateLimitService))

	// Setup routes
	handler.SetupRoutes(router)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
