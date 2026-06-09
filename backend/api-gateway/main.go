package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})
}

func main() {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Gateway routes - forward to services
	router.POST("/auth/login", forwardToAuth)
	router.GET("/auth/validate", forwardToAuth)

	router.POST("/build/create", forwardToBuild)
	router.GET("/build/:id", forwardToBuild)

	router.POST("/deploy/create", forwardToDeploy)
	router.GET("/deploy/:id", forwardToDeploy)

	router.GET("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s\n", port)
	router.Run(":" + port)
}

func forwardToAuth(c *gin.Context) {
	// Implement service discovery and forwarding
	c.JSON(http.StatusOK, gin.H{"message": "forwarded to auth-service"})
}

func forwardToBuild(c *gin.Context) {
	// Implement service discovery and forwarding
	c.JSON(http.StatusOK, gin.H{"message": "forwarded to build-service"})
}

func forwardToDeploy(c *gin.Context) {
	// Implement service discovery and forwarding
	c.JSON(http.StatusOK, gin.H{"message": "forwarded to deploy-service"})
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"service": "api-gateway",
	})
}
