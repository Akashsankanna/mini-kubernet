package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

var (
	db        *sql.DB
	jwtSecret []byte
)

const requestTimeout = 5 * time.Second

// ==============================
// INITIALIZATION
// ==============================

func init() {

	var err error

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {

		// FIXED DATABASE NAME
		dbURL = "postgres://postgres:akash45@localhost:5432/mini_kubernet?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)

	if err != nil {
		log.Fatal("Database connection failed:", err)
	}

	// Connection Pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	err = db.PingContext(ctx)

	if err != nil {
		log.Fatal("Database ping failed:", err)
	}

	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		secret = "change-this-secret-in-production"
	}

	jwtSecret = []byte(secret)

	log.Println("✅ Database connected")
}

// ==============================
// MAIN
// ==============================

func main() {

	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS
	corsConfig := cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
		},
		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	router.Use(cors.New(corsConfig))

	// Routes
	auth := router.Group("/api/v1/auth")
	{
		auth.POST("/register", registerHandler)
		auth.POST("/login", loginHandler)
		auth.POST("/google", googleLoginHandler)

		auth.POST("/login/otp/request", requestOTPLoginHandler)
		auth.POST("/login/otp/verify", verifyOTPLoginHandler)

		auth.GET("/verify", verifyTokenHandler)
		auth.POST("/logout", logoutHandler)

		auth.GET("/profile",
			authMiddleware(),
			getProfileHandler,
		)

		auth.PUT("/profile",
			authMiddleware(),
			updateProfileHandler,
		)

		auth.PUT("/change-password",
			authMiddleware(),
			changePasswordHandler,
		)
	}

	router.GET("/health", healthCheckHandler)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:           ":" + port,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	log.Printf("Auth service running on port %s", port)

	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Server failed:", err)
	}
}
