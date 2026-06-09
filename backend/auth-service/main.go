package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	db        *sql.DB
	jwtSecret []byte
)

const requestTimeout = 5 * time.Second

// ==============================
// MODELS
// ==============================

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// ==============================
// INITIALIZATION
// ==============================

func init() {

	var err error

	dbURL := os.Getenv("DATABASE_URL")

	if dbURL == "" {

		// FIXED DATABASE NAME
		dbURL = "postgres://postgres:akash45@localhost:5432/mini kubernet?sslmode=disable"
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
		auth.POST("/login", login)
		auth.POST("/login/otp/request", requestOTPLoginHandler)
		auth.POST("/login/otp/verify", verifyOTPLoginHandler)
		auth.GET("/validate", validateToken)
		auth.POST("/logout", logout)
		auth.POST("/refresh", refreshToken)
	}

	router.GET("/health", healthCheck)

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

	log.Printf("🚀 Auth service running on port %s", port)

	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("Server failed:", err)
	}
}

// ==============================
// LOGIN
// ==============================

func login(c *gin.Context) {

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	var user User
	var hashedPassword string

	err := db.QueryRowContext(
		ctx,
		`
		SELECT
			id,
			username,
			email,
			role,
			password_hash
		FROM users
		WHERE username = $1 OR email = $1
		LIMIT 1
		`,
		req.Username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Role,
		&hashedPassword,
	)

	if err != nil {

		if err == sql.ErrNoRows {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid credentials",
			})

			return
		}

		log.Println("Login DB error:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "database error",
		})

		return
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(req.Password),
	)

	if err != nil {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})

		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
			Issuer:    "auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)

	if err != nil {

		log.Println("JWT generation error:", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})

		return
	}

	c.JSON(http.StatusOK, TokenResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime,
	})
}

// ==============================
// VALIDATE TOKEN
// ==============================

func validateToken(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")

	if authHeader == "" {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "missing authorization header",
		})

		return
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid authorization format",
		})

		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {

			// SECURITY FIX
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}

			return jwtSecret, nil
		},
	)

	if err != nil || !token.Valid {

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid token",
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":    true,
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

// ==============================
// LOGOUT
// ==============================

func logout(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// ==============================
// REFRESH TOKEN
// ==============================

func refreshToken(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"message": "refresh token endpoint",
	})
}

// ==============================
// HEALTH CHECK
// ==============================

func healthCheck(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	err := db.PingContext(ctx)

	if err != nil {

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "auth-service",
	})
}
