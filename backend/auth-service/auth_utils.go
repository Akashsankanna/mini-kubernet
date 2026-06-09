package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	AccessTokenExpiry  = 15 * time.Minute   // Short-lived access token
	RefreshTokenExpiry = 7 * 24 * time.Hour // Refresh token for 7 days
	OTPExpiry          = 5 * time.Minute    // OTP valid for 5 minutes
)

// HashPassword securely hashes a password using bcrypt
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

// VerifyPassword checks if password matches hash
func verifyPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateOTP creates a 6-digit OTP
func generateOTP() string {
	otp := make([]byte, 3)
	_, err := rand.Read(otp)
	if err != nil {
		return "000000"
	}

	otpInt := int(otp[0])<<16 | int(otp[1])<<8 | int(otp[2])
	otpString := fmt.Sprintf("%06d", otpInt%1000000)
	return otpString
}

// GenerateAccessToken creates a JWT access token
func generateAccessToken(user *User) (string, error) {
	claims := &Claims{
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		UserID:   user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kubernet-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// GenerateRefreshToken creates a refresh token
func generateRefreshToken(userID int) (string, error) {
	claims := &RefreshClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kubernet-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(getJWTSecret())
}

// ValidateToken verifies and parses JWT token
func validateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// ValidateRefreshToken verifies refresh token
func validateRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return getJWTSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	return claims, nil
}

// GetJWTSecret retrieves JWT secret from environment
func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	if secret == "" {
		secret = "your-super-secret-key-change-in-production"
	}
	return []byte(secret)
}

// IsPasswordStrong validates password strength
func isPasswordStrong(password string) bool {
	if len(password) < 8 {
		return false
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	for _, ch := range password {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		case ch == '!' || ch == '@' || ch == '#' || ch == '$' || ch == '%' || ch == '^' || ch == '&' || ch == '*':
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

// IsValidEmail checks if email format is valid
func isValidEmail(email string) bool {
	// Basic email validation
	return len(email) > 5 && len(email) < 255
}

// GetMaxOTPAttempts returns max OTP verification attempts
func getMaxOTPAttempts() int {
	maxAttempts, _ := strconv.Atoi(os.Getenv("MAX_OTP_ATTEMPTS"))
	if maxAttempts == 0 {
		maxAttempts = 3
	}
	return maxAttempts
}

// GetRateLimitConfig returns rate limiting configuration
type RateLimitConfig struct {
	LoginAttempts        int
	OTPAttempts          int
	RegistrationAttempts int
	WindowSeconds        int
}

func getRateLimitConfig() RateLimitConfig {
	loginAttempts, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_LOGIN_ATTEMPTS"))
	if loginAttempts == 0 {
		loginAttempts = 5
	}

	otpAttempts, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_OTP_ATTEMPTS"))
	if otpAttempts == 0 {
		otpAttempts = 3
	}

	regAttempts, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_REG_ATTEMPTS"))
	if regAttempts == 0 {
		regAttempts = 10
	}

	window, _ := strconv.Atoi(os.Getenv("RATE_LIMIT_WINDOW_SECONDS"))
	if window == 0 {
		window = 900 // 15 minutes
	}

	return RateLimitConfig{
		LoginAttempts:        loginAttempts,
		OTPAttempts:          otpAttempts,
		RegistrationAttempts: regAttempts,
		WindowSeconds:        window,
	}
}
