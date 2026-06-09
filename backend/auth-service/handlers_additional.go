package main

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const dbTimeout = 5 * time.Second

// ================================
// CHANGE PASSWORD
// ================================

func changePasswordHandler(c *gin.Context) {

	var req ChangePasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	usernameRaw, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid session",
		})
		return
	}

	username, ok := usernameRaw.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid session",
		})
		return
	}

	if req.OldPassword == req.NewPassword {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "New password must be different",
		})
		return
	}

	if !isPasswordStrong(req.NewPassword) {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Password does not meet security requirements",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var currentHash string

	err := db.QueryRowContext(
		ctx,
		"SELECT password_hash FROM users WHERE id = $1",
		userID,
	).Scan(&currentHash)

	if err != nil {

		logAudit(c, "password_change", username, "failed")

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to verify password",
		})
		return
	}

	if !verifyPassword(currentHash, req.OldPassword) {

		logAudit(c, "password_change", username, "invalid_password")

		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Current password incorrect",
		})
		return
	}

	newHash, err := hashPassword(req.NewPassword)
	if err != nil {

		logAudit(c, "password_change", username, "hash_failed")

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to hash password",
		})
		return
	}

	_, err = db.ExecContext(
		ctx,
		`UPDATE users
		 SET password_hash = $1,
		     updated_at = CURRENT_TIMESTAMP
		 WHERE id = $2`,
		newHash,
		userID,
	)

	if err != nil {

		logAudit(c, "password_change", username, "db_failed")

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update password",
		})
		return
	}

	logAudit(c, "password_changed", username, "success")

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Password changed successfully",
	})
}

// ================================
// GET PROFILE
// ================================

func getProfileHandler(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	user := &User{}

	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			email,
			first_name,
			last_name,
			role,
			status,
			avatar,
			phone_number,
			two_factor_enabled,
			last_login,
			email_verified,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.Avatar,
		&user.PhoneNumber,
		&user.TwoFactorEnabled,
		&user.LastLogin,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to fetch profile",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Profile fetched successfully",
		Data:    user,
	})
}

// ================================
// UPDATE PROFILE
// ================================

func updateProfileHandler(c *gin.Context) {

	var update struct {
		FirstName   string `json:"first_name" binding:"omitempty,max=100"`
		LastName    string `json:"last_name" binding:"omitempty,max=100"`
		PhoneNumber string `json:"phone_number" binding:"omitempty,max=20"`
		Avatar      string `json:"avatar" binding:"omitempty,url"`
	}

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	usernameRaw, _ := c.Get("username")
	username, _ := usernameRaw.(string)

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			phone_number = COALESCE($3, phone_number),
			avatar = COALESCE($4, avatar),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`,
		nullIfEmpty(update.FirstName),
		nullIfEmpty(update.LastName),
		nullIfEmpty(update.PhoneNumber),
		nullIfEmpty(update.Avatar),
		userID,
	)

	if err != nil {

		logAudit(c, "profile_update", username, "failed")

		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update profile",
		})
		return
	}

	logAudit(c, "profile_update", username, "success")

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// ================================
// LIST USERS
// ================================

func listUsersHandler(c *gin.Context) {

	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var total int

	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to count users",
		})
		return
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			username,
			email,
			first_name,
			last_name,
			role,
			status,
			created_at,
			last_login
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to fetch users",
		})
		return
	}

	defer rows.Close()

	var users []User

	for rows.Next() {

		var user User

		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
			&user.Status,
			&user.CreatedAt,
			&user.LastLogin,
		)

		if err != nil {
			continue
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed while reading users",
		})
		return
	}

	c.Header("X-Total-Count", strconv.Itoa(total))
	c.Header("X-Page-Number", strconv.Itoa(page))

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "Users fetched successfully",
		Data: map[string]interface{}{
			"users": users,
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

// ================================
// GET USER
// ================================

func getUserHandler(c *gin.Context) {

	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	user := &User{}

	err := db.QueryRowContext(ctx, `
		SELECT
			id,
			username,
			email,
			first_name,
			last_name,
			role,
			status,
			avatar,
			phone_number,
			two_factor_enabled,
			last_login,
			email_verified,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
		&user.Status,
		&user.Avatar,
		&user.PhoneNumber,
		&user.TwoFactorEnabled,
		&user.LastLogin,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, AuthResponse{
			Success: false,
			Message: "User not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to fetch user",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Data:    user,
	})
}

// ================================
// UPDATE USER
// ================================

func updateUserHandler(c *gin.Context) {

	id := c.Param("id")

	var update struct {
		Role   string `json:"role" binding:"omitempty,oneof=user admin moderator"`
		Status string `json:"status" binding:"omitempty,oneof=active inactive suspended"`
	}

	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Invalid request",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET
			role = COALESCE($1, role),
			status = COALESCE($2, status),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`,
		nullIfEmpty(update.Role),
		nullIfEmpty(update.Status),
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to update user",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User updated successfully",
	})
}

// ================================
// DELETE USER
// ================================

func deleteUserHandler(c *gin.Context) {

	id := c.Param("id")

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := db.ExecContext(ctx, `
		UPDATE users
		SET
			status = 'inactive',
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to delete user",
		})
		return
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}

// ================================
// GET AUDIT LOGS
// ================================

func getAuditLogsHandler(c *gin.Context) {

	page := 1
	limit := 50

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	offset := (page - 1) * limit

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			user_id,
			action,
			resource,
			status,
			ip_address,
			created_at
		FROM audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to fetch audit logs",
		})
		return
	}

	defer rows.Close()

	var logs []AuditLog

	for rows.Next() {

		var logEntry AuditLog

		err := rows.Scan(
			&logEntry.ID,
			&logEntry.UserID,
			&logEntry.Action,
			&logEntry.Resource,
			&logEntry.Status,
			&logEntry.IPAddress,
			&logEntry.CreatedAt,
		)

		if err != nil {
			continue
		}

		logs = append(logs, logEntry)
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Data: map[string]interface{}{
			"logs": logs,
			"page": page,
		},
	})
}

// ================================
// GET SESSIONS
// ================================

func getSessionsHandler(c *gin.Context) {

	userID, exists := c.Get("user_id")

	if !exists {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	rows, err := db.QueryContext(ctx, `
		SELECT
			id,
			ip_address,
			user_agent,
			created_at,
			is_active
		FROM sessions
		WHERE user_id = $1
		AND is_active = true
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Message: "Failed to fetch sessions",
		})
		return
	}

	defer rows.Close()

	var sessions []Session

	for rows.Next() {

		var session Session

		err := rows.Scan(
			&session.ID,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.IsActive,
		)

		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}

	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Data: map[string]interface{}{
			"sessions": sessions,
		},
	})
}

// ================================
// HELPER
// ================================

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
