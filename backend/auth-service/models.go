package main

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// =========================================
// USER MODEL
// =========================================

type User struct {
	ID               int        `json:"id" db:"id"`
	Username         string     `json:"username" db:"username"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	FirstName        string     `json:"first_name" db:"first_name"`
	LastName         string     `json:"last_name" db:"last_name"`
	Role             string     `json:"role" db:"role"`
	Status           string     `json:"status" db:"status"`
	GoogleID         *string    `json:"google_id,omitempty" db:"google_id"`
	Avatar           *string    `json:"avatar,omitempty" db:"avatar"`
	PhoneNumber      *string    `json:"phone_number,omitempty" db:"phone_number"`
	TwoFactorEnabled bool       `json:"two_factor_enabled" db:"two_factor_enabled"`
	LastLogin        *time.Time `json:"last_login,omitempty" db:"last_login"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	EmailVerified    bool       `json:"email_verified" db:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty" db:"email_verified_at"`
}

// =========================================
// OTP MODEL
// =========================================

type OTPRecord struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	OTPCode   string    `json:"-" db:"otp_code"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Attempts  int       `json:"attempts" db:"attempts"`
	IsUsed    bool      `json:"is_used" db:"is_used"`
}

// =========================================
// AUDIT LOG
// =========================================

type AuditLog struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Resource  string    `json:"resource" db:"resource"`
	Status    string    `json:"status" db:"status"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// =========================================
// SESSION MODEL
// =========================================

type Session struct {
	ID        int       `json:"id" db:"id"`
	UserID    int       `json:"user_id" db:"user_id"`
	Token     string    `json:"-" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IPAddress string    `json:"ip_address" db:"ip_address"`
	UserAgent string    `json:"user_agent" db:"user_agent"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// =========================================
// REGISTER REQUEST
// =========================================

type RegisterRequest struct {
	Username  string `json:"username" binding:"required,min=3,max=50"`
	Email     string `json:"email" binding:"required,email,max=255"`
	Password  string `json:"password" binding:"required,min=8,max=72"`
	FirstName string `json:"first_name" binding:"required,max=100"`
	LastName  string `json:"last_name" binding:"required,max=100"`
	Phone     string `json:"phone,omitempty" binding:"omitempty,max=20"`
}

// =========================================
// LOGIN REQUEST
// =========================================

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// =========================================
// OTP LOGIN REQUEST
// =========================================

type OTPLoginRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// =========================================
// OTP VERIFY REQUEST
// =========================================

type OTPVerifyRequest struct {
	Email   string `json:"email" binding:"required,email"`
	OTPCode string `json:"otp_code" binding:"required,len=6,numeric"`
}

// =========================================
// GOOGLE LOGIN REQUEST
// =========================================

type GoogleLoginRequest struct {
	Token string `json:"token" binding:"required"`
}

// =========================================
// REFRESH TOKEN REQUEST
// =========================================

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// =========================================
// CHANGE PASSWORD REQUEST
// =========================================

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=8,max=72"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

// =========================================
// TOKEN RESPONSE
// =========================================

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         *User  `json:"user,omitempty"`
}

// =========================================
// GENERIC AUTH RESPONSE
// =========================================

type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// =========================================
// ACCESS TOKEN CLAIMS
// =========================================

type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`

	jwt.RegisteredClaims
}

// =========================================
// REFRESH TOKEN CLAIMS
// =========================================

type RefreshClaims struct {
	UserID int `json:"user_id"`

	jwt.RegisteredClaims
}

// =========================================
// PAGINATION
// =========================================

type PaginationParams struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=10" binding:"min=1,max=100"`
}
