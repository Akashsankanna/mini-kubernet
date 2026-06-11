package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
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

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:5173",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
	}))

	router.Any("/api/v1/auth/*path", forwardToAuth)
	router.Any("/api/v1/build/*path", forwardToBuild)
	router.Any("/api/v1/deploy/*path", forwardToDeploy)
	router.Any("/api/v1/projects/*path", forwardToProject)
	router.GET("/health", healthCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("API Gateway starting on port %s\n", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func forwardToAuth(c *gin.Context) {

	targetURL := "http://localhost:8081" + c.Request.URL.RequestURI()

	req, err := http.NewRequest(
		c.Request.Method,
		targetURL,
		c.Request.Body,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		body,
	)
}

func forwardToBuild(c *gin.Context) {

	targetURL := "http://localhost:8082" + c.Request.URL.RequestURI()

	req, err := http.NewRequest(
		c.Request.Method,
		targetURL,
		c.Request.Body,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		body,
	)
}

func forwardToDeploy(c *gin.Context) {

	targetURL := "http://localhost:8083" + c.Request.URL.RequestURI()

	req, err := http.NewRequest(
		c.Request.Method,
		targetURL,
		c.Request.Body,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		body,
	)
}

func forwardToProject(c *gin.Context) {

	targetURL := "http://localhost:8084" +
		c.Request.URL.RequestURI()

	req, err := http.NewRequest(
		c.Request.Method,
		targetURL,
		c.Request.Body,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	req.Header = c.Request.Header

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	c.Data(
		resp.StatusCode,
		resp.Header.Get("Content-Type"),
		body,
	)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "api-gateway",
	})
}
