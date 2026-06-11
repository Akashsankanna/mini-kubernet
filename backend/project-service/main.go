package main

import (
	"log"
	"net/http"
	"time"

	"project-service/config"
	gitvalidator "project-service/internal/git"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Project struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	GithubURL string    `json:"github_url"`
	Framework string    `json:"framework"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type CreateProjectRequest struct {
	Name      string `json:"name" binding:"required"`
	GithubURL string `json:"github_url" binding:"required"`
}

func main() {

	config.ConnectDB()

	router := gin.Default()
	router.SetTrustedProxies(nil)

	router.Use(cors.Default())

	projectsGroup := router.Group("/api/v1/projects")
	{
		projectsGroup.POST("", createProject)
		projectsGroup.GET("", getProjects)
		projectsGroup.GET("/:id", getProject)
		projectsGroup.DELETE("/:id", deleteProject)
	}

	router.GET("/health", healthCheck)

	port := "8084"

	router.Run(":" + port)
}
func createProject(c *gin.Context) {

	var req CreateProjectRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if err := gitvalidator.ValidateGitHubRepository(req.GithubURL); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	repoPath, frameworkName, err := gitvalidator.CloneRepository(req.GithubURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("Repository cloned: %s", repoPath)
	log.Printf("Detected framework: %s", frameworkName)

	project := &Project{
		Name:      req.Name,
		GithubURL: req.GithubURL,
		Framework: frameworkName,
		Status:    "created",
	}

	err = config.DB.QueryRow(`
		INSERT INTO projects (
			name,
			github_url,
			framework,
			status
		)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`,
		project.Name,
		project.GithubURL,
		project.Framework,
		project.Status,
	).Scan(
		&project.ID,
		&project.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func getProjects(c *gin.Context) {

	rows, err := config.DB.Query(`
		SELECT id, name, github_url, status, created_at
		FROM projects
		ORDER BY created_at DESC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	defer rows.Close()

	var projects []Project

	for rows.Next() {

		var project Project

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.GithubURL,
			&project.Framework,
			&project.Status,
			&project.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		projects = append(projects, project)
	}

	c.JSON(http.StatusOK, projects)
}

func getProject(c *gin.Context) {

	id := c.Param("id")

	var project Project

	err := config.DB.QueryRow(`
		SELECT id, name, github_url, status, created_at
		FROM projects
		WHERE id = $1
	`,
		id,
	).Scan(
		&project.ID,
		&project.Name,
		&project.GithubURL,
		&project.Framework,
		&project.Status,
		&project.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "project not found",
		})
		return
	}

	c.JSON(http.StatusOK, project)
}

func deleteProject(c *gin.Context) {

	id := c.Param("id")

	result, err := config.DB.Exec(`
		DELETE FROM projects
		WHERE id = $1
	`,
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "project not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "project deleted",
	})
}

func healthCheck(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "project-service",
	})
}
