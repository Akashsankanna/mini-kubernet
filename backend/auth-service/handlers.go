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

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"google.golang.org/api/idtoken"
)

const requestTimeout = 5 * time.Second

// ========================================
// REGISTER HANDLER
// ========================================

func registerHandler(c *gin.Context) {

	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		logAudit(c, "registration_attempt", "user", "invalid_request")

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request format",
		})

		return
	}

	if !isPasswordStrong(req.Password) {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Password does not meet security requirements",
		})

		return
	}

	if !isValidEmail(req.Email) {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid email format",
		})

		return
	}

	if !checkRateLimit(c.ClientIP(), "registration") {

		logAudit(c, "registration_attempt", req.Email, "rate_limited")

		c.JSON(http.StatusTooManyRequests, AuthResponse{
			Success: false,
			Message: "Too many registration attempts",
		})

		return
	}

	passwordHash, err := hashPassword(req.Password)

	if err != nil {

		log.Println("Password hashing failed:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to process password",
		})

		return
	}

	user := &User{
		Username:  strings.TrimSpace(req.Username),
		Email:     strings.TrimSpace(strings.ToLower(req.Email)),
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Role:      "user",
		Status:    "active",
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	err = db.QueryRowContext(ctx, `
		INSERT INTO users (
			username,
			email,
			password_hash,
			first_name,
			last_name,
			role,
			status,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`,
		user.Username,
		user.Email,
		passwordHash,
		user.FirstName,
		user.LastName,
		user.Role,
		user.Status,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {

		var pqErr *pq.Error

		if errors.As(err, &pqErr) {

			if pqErr.Code == "23505" {

				c.JSON(http.StatusConflict, AuthResponse{
					Success: false,
					Message: "Username or email already exists",
				})

				return
			}
		}

		log.Println("Registration DB error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to create user",
		})

		return
	}

	logAudit(c, "user_registered", user.Username, "success")

	c.JSON(http.StatusCreated, AuthResponse{
		Success: true,
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user_id": user.ID,
			"email":   user.Email,
		},
	})
}

// ========================================
// LOGIN HANDLER
// ========================================

func loginHandler(c *gin.Context) {

	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})

		return
	}

	if !checkRateLimit(c.ClientIP(), "login") {

		c.JSON(http.StatusTooManyRequests, AuthResponse{
			Success: false,
			Message: "Too many login attempts",
		})

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	user := &User{}
	var passwordHash string

	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			email,
			password_hash,
			first_name,
			last_name,
			role,
			status,
			avatar,
			created_at,
			updated_at
		FROM users
		WHERE username = $1 OR email = $1
		LIMIT 1
	`,
		req.Username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.Avatar,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})

		return
	}

	if err != nil {

		log.Println("Login DB error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Internal server error",
		})

		return
	}

	if !verifyPassword(passwordHash, req.Password) {

		logAudit(c, "login_attempt", user.Username, "invalid_password")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})

		return
	}

	if user.Status != "active" {

		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Account is " + user.Status,
		})

		return
	}

	accessToken, err := generateAccessToken(user)

	if err != nil {

		log.Println("Access token error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate access token",
		})

		return
	}

	refreshToken, err := generateRefreshToken(user.ID)

	if err != nil {

		log.Println("Refresh token error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate refresh token",
		})

		return
	}

	err = storeSession(
		user.ID,
		refreshToken,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		log.Println("Session store error:", err)
	}

	_, err = db.ExecContext(
		ctx,
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1",
		user.ID,
	)

	if err != nil {
		log.Println("Last login update error:", err)
	}

	logAudit(c, "login_success", user.Username, "success")

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenExpiry.Seconds()),
		User:         user,
	})
}

// ========================================
// GOOGLE LOGIN
// ========================================

func googleLoginHandler(c *gin.Context) {

	var req GoogleLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request",
		})

		return
	}

	clientID := os.Getenv("GOOGLE_CLIENT_ID")

	if clientID == "" {

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Google OAuth not configured",
		})

		return
	}

	payload, err := idtoken.Validate(
		context.Background(),
		req.Token,
		clientID,
	)

	if err != nil {

		log.Println("Google token validation failed:", err)

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid Google token",
		})

		return
	}

	email, ok := payload.Claims["email"].(string)

	if !ok {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid Google account",
		})

		return
	}

	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	googleID := payload.Subject

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	user := &User{}

	err = db.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			email,
			first_name,
			last_name,
			role,
			avatar,
			status,
			created_at,
			updated_at
		FROM users
		WHERE google_id = $1 OR email = $2
	`,
		googleID,
		email,
	).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Avatar,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {

		username := strings.Split(email, "@")[0]

		user = &User{
			Username:  username,
			Email:     email,
			GoogleID:  googleID,
			Avatar:    picture,
			FirstName: name,
			Status:    "active",
			Role:      "user",
		}

		err = db.QueryRowContext(ctx, `
			INSERT INTO users (
				username,
				email,
				google_id,
				avatar,
				first_name,
				role,
				status,
				email_verified,
				created_at,
				updated_at
			)
			VALUES (
				$1,$2,$3,$4,$5,$6,$7,true,
				CURRENT_TIMESTAMP,
				CURRENT_TIMESTAMP
			)
			RETURNING id
		`,
			user.Username,
			user.Email,
			user.GoogleID,
			user.Avatar,
			user.FirstName,
			user.Role,
			user.Status,
		).Scan(&user.ID)

		if err != nil {

			log.Println("Google user creation failed:", err)

			c.JSON(http.StatusInternalServerError, AuthResponse{
				Success: false,
				Message: "Failed to create user",
			})

			return
		}

	} else if err != nil {

		log.Println("Google login DB error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Internal server error",
		})

		return
	}

	if user.Status != "active" {

		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Account is " + user.Status,
		})

		return
	}

	accessToken, err := generateAccessToken(user)

	if err != nil {

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})

		return
	}

	refreshToken, err := generateRefreshToken(user.ID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate refresh token",
		})

		return
	}

	_ = storeSession(
		user.ID,
		refreshToken,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	now := time.Now()
	user.LastLogin = &now

	_, _ = db.ExecContext(
		ctx,
		"UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1",
		user.ID,
	)

	logAudit(c, "google_login", user.Username, "success")

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenExpiry.Seconds()),
		User:         user,
	})
}

// ========================================
// OTP LOGIN REQUEST
// ========================================

func requestOTPLoginHandler(c *gin.Context) {

	var req OTPLoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request format",
		})

		return
	}

	// Rate limiting
	if !checkRateLimit(c.ClientIP(), "otp") {

		logAudit(c, "otp_request", req.Email, "rate_limited")

		c.JSON(http.StatusTooManyRequests, AuthResponse{
			Success: false,
			Message: "Too many OTP requests. Please try again later",
		})

		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Check if user exists
	var userID int
	var userName, firstName string

	err := db.QueryRowContext(ctx, `
		SELECT id, username, first_name FROM users WHERE email = $1
	`, email).Scan(&userID, &userName, &firstName)

	if err == sql.ErrNoRows {

		// For security, don't reveal if email exists
		logAudit(c, "otp_request", email, "user_not_found")

		c.JSON(http.StatusOK, AuthResponse{
			Success: true,
			Message: "If email exists, OTP has been sent",
		})

		return
	}

	if err != nil {

		log.Println("OTP request DB error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to process request",
		})

		return
	}

	// Generate OTP
	otp := generateOTP()

	// Store OTP in database with 5-minute expiry
	expiresAt := time.Now().Add(OTPExpiry)

	_, err = db.ExecContext(ctx, `
		INSERT INTO otp_records (
			user_id,
			otp_code,
			expires_at,
			attempts,
			is_used,
			created_at
		)
		VALUES ($1, $2, $3, 0, false, CURRENT_TIMESTAMP)
		ON CONFLICT (user_id)
		DO UPDATE SET
			otp_code = $2,
			expires_at = $3,
			attempts = 0,
			is_used = false,
			created_at = CURRENT_TIMESTAMP
	`, userID, otp, expiresAt)

	if err != nil {

		log.Println("OTP storage failed:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate OTP",
		})

		return
	}

	// Send OTP email (for now, log it)
	err = sendOTPEmail(email, firstName, otp)

	if err != nil {

		log.Println("OTP email send failed:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to send OTP",
		})

		return
	}

	logAudit(c, "otp_requested", email, "success")

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "OTP sent to your email",
		Data: map[string]interface{}{
			"masked_email": maskEmail(email),
			"otp":          otp, // For development/testing only - REMOVE IN PRODUCTION
			"expires_in":   int(OTPExpiry.Seconds()),
		},
	})
}

// ========================================
// OTP LOGIN VERIFY
// ========================================

func verifyOTPLoginHandler(c *gin.Context) {

	var req OTPVerifyRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request format",
		})

		return
	}

	email := strings.TrimSpace(strings.ToLower(req.Email))

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	// Find user by email
	user := &User{}
	var otpCode string
	var otpExpiresAt time.Time
	var otpAttempts int
	var otpIsUsed bool

	err := db.QueryRowContext(ctx, `
		SELECT
			u.id,
			u.username,
			u.email,
			u.first_name,
			u.last_name,
			u.role,
			u.avatar,
			u.status,
			u.created_at,
			u.updated_at,
			COALESCE(o.otp_code, ''),
			COALESCE(o.expires_at, CURRENT_TIMESTAMP),
			COALESCE(o.attempts, 0),
			COALESCE(o.is_used, false)
		FROM users u
		LEFT JOIN otp_records o ON u.id = o.user_id
		WHERE u.email = $1
	`, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Avatar,
		&user.Status,
		&user.CreatedAt,
		&user.UpdatedAt,
		&otpCode,
		&otpExpiresAt,
		&otpAttempts,
		&otpIsUsed,
	)

	if err == sql.ErrNoRows {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "User not found",
		})

		return
	}

	if err != nil {

		log.Println("OTP verify DB error:", err)

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Internal server error",
		})

		return
	}

	// Check if OTP is expired
	if time.Now().After(otpExpiresAt) {

		logAudit(c, "otp_verify", email, "expired")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "OTP expired. Please request a new one",
		})

		return
	}

	// Check if OTP is already used
	if otpIsUsed {

		logAudit(c, "otp_verify", email, "already_used")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "OTP already used. Please request a new one",
		})

		return
	}

	// Check maximum attempts
	if otpAttempts >= 3 {

		logAudit(c, "otp_verify", email, "max_attempts")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Maximum attempts exceeded. Please request a new OTP",
		})

		return
	}

	// Verify OTP
	if req.OTPCode != otpCode {

		// Increment attempts
		_, _ = db.ExecContext(ctx, `
			UPDATE otp_records
			SET attempts = attempts + 1
			WHERE user_id = $1
		`, user.ID)

		logAudit(c, "otp_verify", email, "invalid_otp")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid OTP",
		})

		return
	}

	// Mark OTP as used
	_, _ = db.ExecContext(ctx, `
		UPDATE otp_records
		SET is_used = true, attempts = 0
		WHERE user_id = $1
	`, user.ID)

	// Check user status
	if user.Status != "active" {

		c.JSON(http.StatusForbidden, AuthResponse{
			Success: false,
			Message: "Account is " + user.Status,
		})

		return
	}

	// Generate tokens
	accessToken, err := generateAccessToken(user)

	if err != nil {

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate token",
		})

		return
	}

	refreshToken, err := generateRefreshToken(user.ID)

	if err != nil {

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to generate refresh token",
		})

		return
	}

	// Store session
	_ = storeSession(user.ID, refreshToken, c.ClientIP(), c.Request.UserAgent())

	// Update last login
	_, _ = db.ExecContext(ctx, `
		UPDATE users
		SET last_login = CURRENT_TIMESTAMP
		WHERE id = $1
	`, user.ID)

	logAudit(c, "otp_login_success", user.Username, "success")

	c.JSON(http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(AccessTokenExpiry.Seconds()),
		User:         user,
	})
}

// ========================================
// VERIFY TOKEN
// ========================================

func verifyTokenHandler(c *gin.Context) {

	token := c.GetHeader("Authorization")

	if token == "" {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Missing authorization header",
		})

		return
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims, err := validateToken(token)

	if err != nil {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid token",
		})

		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Token valid",
		Data: map[string]interface{}{
			"user_id":  claims.UserID,
			"username": claims.Username,
			"email":    claims.Email,
			"role":     claims.Role,
		},
	})
}

// ========================================
// LOGOUT
// ========================================

func logoutHandler(c *gin.Context) {

	token := c.GetHeader("Authorization")

	if token == "" {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Missing authorization token",
		})

		return
	}

	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims, err := validateToken(token)

	if err != nil {

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid token",
		})

		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err = db.ExecContext(
		ctx,
		"UPDATE sessions SET is_active = false WHERE token = $1",
		token,
	)

	if err != nil {
		log.Println("Logout session update failed:", err)
	}

	logAudit(c, "logout", claims.Username, "success")

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// ========================================
// HEALTH CHECK
// ========================================

func healthCheckHandler(c *gin.Context) {

	err := HealthCheck(db)

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
		"version": "1.0.0",
	})
}

// ========================================
// HELPERS
// ========================================

func storeSession(userID int, token, ip, userAgent string) error {

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		INSERT INTO sessions (
			user_id,
			token,
			expires_at,
			ip_address,
			user_agent,
			created_at,
			is_active
		)
		VALUES (
			$1,$2,$3,$4,$5,
			CURRENT_TIMESTAMP,
			true
		)
	`,
		userID,
		token,
		time.Now().Add(RefreshTokenExpiry),
		ip,
		userAgent,
	)

	return err
}

func logAudit(c *gin.Context, action, resource, status string) {

	userID := 0

	authHeader := c.GetHeader("Authorization")

	if strings.HasPrefix(authHeader, "Bearer ") {

		token := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := validateToken(token)

		if err == nil && claims != nil {
			userID = claims.UserID
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		INSERT INTO audit_logs (
			user_id,
			action,
			resource,
			status,
			ip_address,
			user_agent,
			created_at
		)
		VALUES (
			$1,$2,$3,$4,$5,$6,
			CURRENT_TIMESTAMP
		)
	`,
		userID,
		action,
		resource,
		status,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	if err != nil {
		log.Println("Audit log failed:", err)
	}
}

func checkRateLimit(identifier, endpoint string) bool {

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	config := getRateLimitConfig()

	var attempts int

	err := db.QueryRowContext(ctx, `
		SELECT attempts
		FROM rate_limits
		WHERE identifier = $1
		AND endpoint = $2
		AND reset_at > CURRENT_TIMESTAMP
	`,
		identifier,
		endpoint,
	).Scan(&attempts)

	if err != nil && err != sql.ErrNoRows {
		return true
	}

	maxAttempts := config.LoginAttempts

	switch endpoint {
	case "otp":
		maxAttempts = config.OTPAttempts
	case "registration":
		maxAttempts = config.RegistrationAttempts
	}

	if attempts >= maxAttempts {
		return false
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO rate_limits (
			identifier,
			endpoint,
			attempts,
			reset_at,
			created_at
		)
		VALUES (
			$1,$2,1,$3,
			CURRENT_TIMESTAMP
		)
		ON CONFLICT (identifier, endpoint)
		DO UPDATE SET
			attempts = rate_limits.attempts + 1,
			reset_at = $3
	`,
		identifier,
		endpoint,
		time.Now().Add(
			time.Duration(config.WindowSeconds)*time.Second,
		),
	)

	if err != nil {
		log.Println("Rate limit update failed:", err)
	}

	return true
}

func maskEmail(email string) string {

	if len(email) < 5 {
		return "***"
	}

	return email[:2] + "****" + email[len(email)-2:]
}

func sendOTPEmail(email, name, otp string) error {

	log.Printf(
		"OTP %s sent to %s (%s)",
		otp,
		name,
		email,
	)

	return nil
}
