package main

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ==========================================
// AUTH MIDDLEWARE
// ==========================================

func authMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Missing authorization header",
			})

			c.Abort()
			return
		}

		// Validate Bearer format
		if !strings.HasPrefix(authHeader, "Bearer ") {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Invalid authorization format",
			})

			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		if tokenString == "" {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Missing token",
			})

			c.Abort()
			return
		}

		// Validate JWT
		claims, err := validateToken(tokenString)

		if err != nil {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Invalid or expired token",
			})

			c.Abort()
			return
		}

		// Store values in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// ==========================================
// ROLE MIDDLEWARE
// ==========================================

func roleMiddleware(requiredRoles ...string) gin.HandlerFunc {

	return func(c *gin.Context) {

		roleValue, exists := c.Get("role")

		if !exists {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Role not found",
			})

			c.Abort()
			return
		}

		userRole, ok := roleValue.(string)

		if !ok {

			c.JSON(http.StatusUnauthorized, AuthResponse{
				Success: false,
				Message: "Invalid role format",
			})

			c.Abort()
			return
		}

		allowed := false

		for _, role := range requiredRoles {

			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {

			c.JSON(http.StatusForbidden, AuthResponse{
				Success: false,
				Message: "Insufficient permissions",
			})

			c.Abort()
			return
		}

		c.Next()
	}
}

// ==========================================
// LOGGING MIDDLEWARE
// ==========================================

func loggingMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		start := time.Now()

		c.Next()

		duration := time.Since(start)

		log.Printf(
			"%s %s %d %s IP=%s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
		)
	}
}

// ==========================================
// ERROR HANDLING MIDDLEWARE
// ==========================================

func errorHandlingMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		defer func() {

			if err := recover(); err != nil {

				log.Printf("PANIC RECOVERED: %v", err)

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					AuthResponse{
						Success: false,
						Message: "Internal server error",
					},
				)
			}
		}()

		c.Next()
	}
}

// ==========================================
// SECURITY HEADERS
// ==========================================

func securityHeadersMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// HSTS
		c.Header(
			"Strict-Transport-Security",
			"max-age=31536000; includeSubDomains",
		)

		// CSP
		c.Header(
			"Content-Security-Policy",
			"default-src 'self'",
		)

		// Referrer Policy
		c.Header(
			"Referrer-Policy",
			"strict-origin-when-cross-origin",
		)

		// Permissions Policy
		c.Header(
			"Permissions-Policy",
			"geolocation=(), microphone=(), camera=()",
		)

		// Disable caching sensitive responses
		c.Header(
			"Cache-Control",
			"no-store",
		)

		c.Next()
	}
}

// ==========================================
// REQUEST VALIDATION
// ==========================================

func requestValidationMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// Limit request body size
		c.Request.Body = http.MaxBytesReader(
			c.Writer,
			c.Request.Body,
			1024*1024, // 1MB
		)

		method := c.Request.Method

		// Validate JSON content-type
		if method == http.MethodPost ||
			method == http.MethodPut ||
			method == http.MethodPatch {

			contentType := c.ContentType()

			// Allow charset=utf-8
			if !strings.HasPrefix(contentType, "application/json") {

				c.JSON(http.StatusBadRequest, AuthResponse{
					Success: false,
					Message: "Content-Type must be application/json",
				})

				c.Abort()
				return
			}
		}

		c.Next()
	}
}
