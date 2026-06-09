package main

import (
	"database/sql"
	"fmt"
	"log"
)

const (
	UsersTableSQL = `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(50) NOT NULL UNIQUE,
		email VARCHAR(255) NOT NULL UNIQUE,
		password_hash VARCHAR(255),
		first_name VARCHAR(100),
		last_name VARCHAR(100),
		role VARCHAR(20) DEFAULT 'user',
		status VARCHAR(20) DEFAULT 'active',
		google_id VARCHAR(255),
		avatar VARCHAR(500),
		phone_number VARCHAR(20),
		two_factor_enabled BOOLEAN DEFAULT false,
		last_login TIMESTAMP,
		email_verified BOOLEAN DEFAULT false,
		email_verified_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
	CREATE INDEX IF NOT EXISTS idx_users_google_id ON users(google_id);
	`

	OTPTableSQL = `
	CREATE TABLE IF NOT EXISTS otp_records (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		otp_code VARCHAR(10) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		attempts INTEGER DEFAULT 0,
		is_used BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_otp_user_id ON otp_records(user_id);
	CREATE INDEX IF NOT EXISTS idx_otp_expires_at ON otp_records(expires_at);
	`

	SessionsTableSQL = `
	CREATE TABLE IF NOT EXISTS sessions (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL,
		token VARCHAR(512) NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		ip_address VARCHAR(45),
		user_agent TEXT,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);
	CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token);
	`

	AuditLogsTableSQL = `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id SERIAL PRIMARY KEY,
		user_id INTEGER,
		action VARCHAR(50) NOT NULL,
		resource VARCHAR(100),
		status VARCHAR(20),
		ip_address VARCHAR(45),
		user_agent TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
	);

	CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_logs(user_id);
	CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
	CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs(created_at);
	`

	RateLimitTableSQL = `
	CREATE TABLE IF NOT EXISTS rate_limits (
		id SERIAL PRIMARY KEY,
		identifier VARCHAR(255) NOT NULL,
		endpoint VARCHAR(255),
		attempts INTEGER DEFAULT 0,
		reset_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(identifier, endpoint)
	);

	CREATE INDEX IF NOT EXISTS idx_rate_limit_identifier ON rate_limits(identifier);
	`
)

// InitDB initializes all database tables and migrations
func InitDB(db *sql.DB) error {
	log.Println("🔄 Running database migrations...")

	tables := []string{
		UsersTableSQL,
		OTPTableSQL,
		SessionsTableSQL,
		AuditLogsTableSQL,
		RateLimitTableSQL,
	}

	for i, table := range tables {
		if _, err := db.Exec(table); err != nil {
			return fmt.Errorf("migration %d failed: %w", i, err)
		}
	}

	log.Println("✅ Database migrations completed successfully")
	return nil
}

// SeedDefaultData creates default admin user if not exists
func SeedDefaultData(db *sql.DB) error {
	log.Println("🌱 Checking database seed data...")

	// Check if admin exists
	var exists int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'admin'").Scan(&exists)
	if err != nil {
		return err
	}

	if exists > 0 {
		log.Println("✅ Admin user already exists")
		return nil
	}

	// Create default admin
	hashedPassword, _ := hashPassword("Admin@123456")
	_, err = db.Exec(`
		INSERT INTO users (username, email, password_hash, first_name, last_name, role, status, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, "admin", "admin@kubernet.io", hashedPassword, "Admin", "User", "admin", "active", true)

	if err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	log.Println("✅ Default admin user created")
	return nil
}

// HealthCheck verifies database connectivity
func HealthCheck(db *sql.DB) error {
	return db.Ping()
}
