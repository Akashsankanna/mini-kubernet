package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ==========================================
// MODELS
// ==========================================

type BuildRequest struct {
	ServiceName string `json:"service_name" binding:"required"`
	GitRepo     string `json:"git_repo" binding:"required"`
	GitBranch   string `json:"git_branch"`
	Dockerfile  string `json:"dockerfile"`
	Registry    string `json:"registry" binding:"required"`
}

type BuildResponse struct {
	BuildID   string    `json:"build_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type BuildStatus struct {
	BuildID     string    `json:"build_id"`
	Status      string    `json:"status"`
	Progress    int       `json:"progress"`
	Log         string    `json:"log"`
	Error       string    `json:"error,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
}

// ==========================================
// GLOBALS
// ==========================================

var (
	builds      = make(map[string]*BuildStatus)
	buildsMutex sync.RWMutex
)

// ==========================================
// MAIN
// ==========================================

func main() {

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS
	router.Use(cors.Default())

	// Routes
	build := router.Group("/api/v1/build")
	{
		build.POST("/create", createBuild)
		build.GET("/:id", getBuildStatus)
		build.DELETE("/:id", cancelBuild)
	}

	router.GET("/health", healthCheck)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8082"
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("🚀 Build Service running on port %s", port)

	err := server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}

// ==========================================
// CREATE BUILD
// ==========================================

func createBuild(c *gin.Context) {

	var req BuildRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	// Default branch
	if req.GitBranch == "" {
		req.GitBranch = "main"
	}

	// Default Dockerfile
	if req.Dockerfile == "" {
		req.Dockerfile = "Dockerfile"
	}

	buildID := generateBuildID()

	status := &BuildStatus{
		BuildID:  buildID,
		Status:   "pending",
		Progress: 0,
		Log:      "",
	}

	buildsMutex.Lock()
	builds[buildID] = status
	buildsMutex.Unlock()

	// Run async
	go executeBuild(buildID, req)

	c.JSON(http.StatusAccepted, BuildResponse{
		BuildID:   buildID,
		Status:    "pending",
		CreatedAt: time.Now(),
	})
}

// ==========================================
// EXECUTE BUILD
// ==========================================

func executeBuild(buildID string, req BuildRequest) {

	updateBuild(buildID, "building", 10, "Starting build...\n", "")

	workspace := filepath.Join(os.TempDir(), buildID)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		20*time.Minute,
	)

	defer cancel()

	// Clone repo
	updateBuild(buildID, "building", 20, "Cloning repository...\n", "")

	cloneCmd := exec.CommandContext(
		ctx,
		"git",
		"clone",
		"-b",
		req.GitBranch,
		req.GitRepo,
		workspace,
	)

	cloneOutput, err := cloneCmd.CombinedOutput()

	if err != nil {

		updateBuild(
			buildID,
			"failed",
			20,
			string(cloneOutput),
			err.Error(),
		)

		return
	}

	// Docker image name
	imageName := fmt.Sprintf(
		"%s/%s:latest",
		req.Registry,
		req.ServiceName,
	)

	updateBuild(buildID, "building", 50, "Building docker image...\n", "")

	// Docker build
	buildCmd := exec.CommandContext(
		ctx,
		"docker",
		"build",
		"-f",
		req.Dockerfile,
		"-t",
		imageName,
		workspace,
	)

	buildOutput, err := buildCmd.CombinedOutput()

	if err != nil {

		updateBuild(
			buildID,
			"failed",
			50,
			string(buildOutput),
			err.Error(),
		)

		return
	}

	updateBuild(
		buildID,
		"building",
		75,
		"Image build successful\n"+string(buildOutput),
		"",
	)

	// Docker push
	pushCmd := exec.CommandContext(
		ctx,
		"docker",
		"push",
		imageName,
	)

	pushOutput, err := pushCmd.CombinedOutput()

	if err != nil {

		updateBuild(
			buildID,
			"failed",
			80,
			string(pushOutput),
			err.Error(),
		)

		return
	}

	buildsMutex.Lock()

	builds[buildID].Status = "completed"
	builds[buildID].Progress = 100
	builds[buildID].Log += string(pushOutput)
	builds[buildID].CompletedAt = time.Now()

	buildsMutex.Unlock()

	log.Printf("✅ Build %s completed", buildID)

	// Cleanup
	_ = os.RemoveAll(workspace)
}

// ==========================================
// GET BUILD STATUS
// ==========================================

func getBuildStatus(c *gin.Context) {

	buildID := c.Param("id")

	buildsMutex.RLock()

	status, exists := builds[buildID]

	buildsMutex.RUnlock()

	if !exists {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "build not found",
		})

		return
	}

	c.JSON(http.StatusOK, status)
}

// ==========================================
// CANCEL BUILD
// ==========================================

func cancelBuild(c *gin.Context) {

	buildID := c.Param("id")

	buildsMutex.Lock()
	defer buildsMutex.Unlock()

	status, exists := builds[buildID]

	if !exists {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "build not found",
		})

		return
	}

	if status.Status == "completed" {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "cannot cancel completed build",
		})

		return
	}

	status.Status = "cancelled"

	c.JSON(http.StatusOK, gin.H{
		"message":  "build cancelled",
		"build_id": buildID,
	})
}

// ==========================================
// HEALTH CHECK
// ==========================================

func healthCheck(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "build-service",
	})
}

// ==========================================
// HELPERS
// ==========================================

func generateBuildID() string {

	hostname := os.Getenv("HOSTNAME")

	if hostname == "" {
		hostname = "local"
	}

	return fmt.Sprintf(
		"build-%s-%s-%s",
		hostname,
		strconv.FormatInt(time.Now().Unix(), 10),
		uuid.New().String()[:8],
	)
}

func updateBuild(
	buildID,
	status string,
	progress int,
	logMsg,
	errMsg string,
) {

	buildsMutex.Lock()
	defer buildsMutex.Unlock()

	build, exists := builds[buildID]

	if !exists {
		return
	}

	build.Status = status
	build.Progress = progress
	build.Log += logMsg

	if errMsg != "" {
		build.Error = errMsg
	}
}
