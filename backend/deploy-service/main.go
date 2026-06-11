package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"

	"k8s.io/client-go/tools/clientcmd"
)

// ==========================================
// MODELS
// ==========================================

type DeploymentRequest struct {
	ServiceName string            `json:"service_name" binding:"required"`
	Image       string            `json:"image" binding:"required"`
	Replicas    int32             `json:"replicas"`
	Namespace   string            `json:"namespace"`
	Environment map[string]string `json:"environment,omitempty"`
}

type DeploymentStatus struct {
	DeploymentID    string    `json:"deployment_id"`
	ServiceName     string    `json:"service_name"`
	Status          string    `json:"status"`
	ReadyReplicas   int32     `json:"ready_replicas"`
	DesiredReplicas int32     `json:"desired_replicas"`
	CreatedAt       time.Time `json:"created_at"`
}

// ==========================================
// GLOBALS
// ==========================================

var (
	k8sClient   kubernetes.Interface
	deployments = make(map[string]*DeploymentStatus)
	mutex       sync.RWMutex
)

// ==========================================
// INIT
// ==========================================

func init() {

	var (
		config *rest.Config
		err    error
	)

	// Try in-cluster config
	config, err = rest.InClusterConfig()

	if err != nil {

		log.Println("Using local kubeconfig")

		home, err := os.UserHomeDir()
		if err != nil {
			log.Println("Failed to get user home directory:", err)
			return
		}

		kubeconfig := filepath.Join(
			home,
			".kube",
			"config",
		)

		config, err = clientcmd.BuildConfigFromFlags(
			"",
			kubeconfig,
		)

		if err != nil {

			log.Println("Failed to load kubeconfig:", err)
			return
		}
	}

	k8sClient, err = kubernetes.NewForConfig(config)

	if err != nil {
		log.Println("Failed to create Kubernetes client:", err)
		return
	}

	log.Println("✅ Kubernetes client initialized")
}

// ==========================================
// MAIN
// ==========================================

func main() {

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(cors.Default())

	// Routes
	deploy := router.Group("/api/v1/deploy")
	{
		deploy.POST("/create", createDeployment)
		deploy.GET("/:id", getDeploymentStatus)
		deploy.PATCH("/:id/rollback", rollbackDeployment)
		deploy.DELETE("/:id", deleteDeployment)
	}

	router.GET("/health", healthCheck)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    15 * time.Second,
		WriteTimeout:   15 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("🚀 Deploy Service running on port %s", port)

	err := server.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		log.Fatal("Server failed:", err)
	}
}

// ==========================================
// CREATE DEPLOYMENT
// ==========================================

func createDeployment(c *gin.Context) {

	var req DeploymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	if req.Namespace == "" {
		req.Namespace = "default"
	}

	if req.Replicas <= 0 {
		req.Replicas = 1
	}

	deploymentID := generateDeploymentID(req.ServiceName)

	status := &DeploymentStatus{
		DeploymentID:    deploymentID,
		ServiceName:     req.ServiceName,
		Status:          "creating",
		ReadyReplicas:   0,
		DesiredReplicas: req.Replicas,
		CreatedAt:       time.Now(),
	}

	mutex.Lock()
	deployments[deploymentID] = status
	mutex.Unlock()

	go deployToKubernetes(
		context.Background(),
		req,
		status,
	)

	c.JSON(http.StatusAccepted, status)
}

// ==========================================
// GET DEPLOYMENT STATUS
// ==========================================

func getDeploymentStatus(c *gin.Context) {

	deploymentID := c.Param("id")

	mutex.RLock()

	status, exists := deployments[deploymentID]

	mutex.RUnlock()

	if !exists {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "deployment not found",
		})

		return
	}

	c.JSON(http.StatusOK, status)
}

// ==========================================
// ROLLBACK DEPLOYMENT
// ==========================================

func rollbackDeployment(c *gin.Context) {

	deploymentID := c.Param("id")

	mutex.RLock()

	status, exists := deployments[deploymentID]

	mutex.RUnlock()

	if !exists {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "deployment not found",
		})

		return
	}

	status.Status = "rolled_back"

	c.JSON(http.StatusOK, gin.H{
		"message":       "deployment rolled back",
		"deployment_id": deploymentID,
	})
}

// ==========================================
// DELETE DEPLOYMENT
// ==========================================

func deleteDeployment(c *gin.Context) {

	deploymentID := c.Param("id")

	mutex.Lock()

	status, exists := deployments[deploymentID]

	if exists {
		status.Status = "deleted"
		delete(deployments, deploymentID)
	}

	mutex.Unlock()

	if !exists {

		c.JSON(http.StatusNotFound, gin.H{
			"error": "deployment not found",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "deployment deleted",
		"deployment_id": deploymentID,
	})
}

// ==========================================
// HEALTH CHECK
// ==========================================

func healthCheck(c *gin.Context) {

	if k8sClient == nil {

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "degraded",
			"service": "deploy-service",
			"k8s":     "unavailable",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "deploy-service",
	})
}

// ==========================================
// DEPLOY TO KUBERNETES
// ==========================================

func deployToKubernetes(
	ctx context.Context,
	req DeploymentRequest,
	status *DeploymentStatus,
) {

	if k8sClient == nil {

		log.Println("Kubernetes client unavailable")

		status.Status = "failed"

		return
	}

	status.Status = "deploying"

	labels := map[string]string{
		"app": req.ServiceName,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.ServiceName,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &req.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  req.ServiceName,
							Image: req.Image,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	_, err := k8sClient.
		AppsV1().
		Deployments(req.Namespace).
		Create(
			ctx,
			deployment,
			metav1.CreateOptions{},
		)

	if err != nil {

		log.Println("Deployment failed:", err)

		status.Status = "failed"

		return
	}

	log.Printf("✅ Deployment %s created", req.ServiceName)

	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.ServiceName,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       80,
					TargetPort: intstr.FromInt(80),
				},
			},
		},
	}

	_, err = k8sClient.
		CoreV1().
		Services(req.Namespace).
		Create(
			ctx,
			service,
			metav1.CreateOptions{},
		)

	if err != nil {

		log.Println("Service creation failed:", err)

		status.Status = "failed"

		return
	}

	log.Printf("✅ Service %s created", req.ServiceName)

	time.Sleep(2 * time.Second)

	status.Status = "ready"
	status.ReadyReplicas = req.Replicas

	log.Printf(
		"✅ Deployment %s completed",
		status.DeploymentID,
	)
}

// ==========================================
// HELPERS
// ==========================================

func generateDeploymentID(service string) string {

	return fmt.Sprintf(
		"deploy-%s-%s-%s",
		service,
		strconv.FormatInt(time.Now().Unix(), 10),
		uuid.New().String()[:8],
	)
}
