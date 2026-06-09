# Kubernet - Complete Deployment Guide

## Production-Ready Setup with OTP, Google OAuth, and Advanced Features

This guide covers deploying the complete Kubernet platform with authentication, Kubernetes management, and monitoring.

---

## Quick Start (Local Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- Node.js 20+
- PostgreSQL 15
- Git

### 1. Clone and Setup

```bash
cd /path/to/mini-kubernet
cp backend/auth-service/.env.example backend/auth-service/.env
cp frontend/.env.example frontend/.env
```

### 2. Configure Environment

**Backend (.env)**
```env
DATABASE_URL=postgres://postgres:akash45@localhost:5432/mini_kubernet?sslmode=disable
JWT_SECRET_KEY=your-super-secret-key-change-in-production
GOOGLE_CLIENT_ID=your-google-oauth-client-id
CORS_ORIGINS=http://localhost:3000
PORT=8081
GIN_MODE=debug
```

**Frontend (.env)**
```env
VITE_API_URL=http://localhost:8081/api/v1
VITE_GOOGLE_CLIENT_ID=your-google-oauth-client-id
```

### 3. Start Services

```bash
# Start database
docker-compose up -d postgres redis

# Wait for database to be ready
sleep 5

# Start backend (from backend/auth-service)
cd backend/auth-service
go run main.go

# Start frontend (from frontend)
cd frontend
npm install
npm run dev
```

Access:
- Frontend: http://localhost:3000
- API: http://localhost:8081/api/v1
- Health: http://localhost:8081/health

---

## Docker Compose Production Setup

### Build Images

```bash
# Build all services
docker-compose build

# Or build specific services
docker-compose build auth-service
docker-compose build frontend
```

### Run Stack

```bash
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f auth-service
```

---

## Google OAuth Setup

### 1. Create Google OAuth Application

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a new project
3. Enable **Google+ API**
4. Create OAuth 2.0 credentials:
   - Type: Web Application
   - Authorized redirect URIs:
     - `http://localhost:3000`
     - `http://localhost:3000/login`
     - `https://your-domain.com`
     - `https://your-domain.com/login`

### 2. Configure Credentials

Copy the Client ID and Client Secret to:
- Backend: `.env` → `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`
- Frontend: `.env` → `VITE_GOOGLE_CLIENT_ID`

---

## Database Migrations

The application auto-runs migrations on startup. Tables created:

- `users` - User accounts and profiles
- `otp_records` - One-Time Password storage
- `sessions` - Session management
- `audit_logs` - Audit trail
- `rate_limits` - Rate limiting tracking

### Manual Migration (if needed)

```bash
# Connect to database
psql -U postgres -d mini_kubernet -h localhost

# View table structure
\dt

# Check user count
SELECT COUNT(*) FROM users;
```

---

## API Endpoints

### Authentication

#### Register
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890"
}
```

#### Login with Password
```bash
POST /api/v1/auth/login
{
  "username": "john_doe",
  "password": "SecurePass123!"
}
```

#### OTP Login - Request
```bash
POST /api/v1/auth/login/otp/request
{
  "email": "john@example.com"
}
```

#### OTP Login - Verify
```bash
POST /api/v1/auth/login/otp/verify
{
  "email": "john@example.com",
  "otp_code": "123456"
}
```

#### Google Login
```bash
POST /api/v1/auth/login/google
{
  "token": "google-access-token"
}
```

#### Refresh Token
```bash
POST /api/v1/auth/token/refresh
{
  "refresh_token": "refresh-token-from-login"
}
```

#### Verify Token
```bash
GET /api/v1/auth/verify
Authorization: Bearer <access_token>
```

#### Logout
```bash
POST /api/v1/auth/logout
Authorization: Bearer <access_token>
```

#### Change Password
```bash
POST /api/v1/auth/password/change
Authorization: Bearer <access_token>
{
  "old_password": "CurrentPass123!",
  "new_password": "NewPass123!"
}
```

### User Profile

#### Get Profile
```bash
GET /api/v1/user/profile
Authorization: Bearer <access_token>
```

#### Update Profile
```bash
PUT /api/v1/user/profile
Authorization: Bearer <access_token>
{
  "first_name": "John",
  "last_name": "Doe",
  "phone_number": "+1234567890",
  "avatar": "https://example.com/avatar.jpg"
}
```

### Admin Endpoints

#### List Users
```bash
GET /api/v1/admin/users?page=1&limit=20
Authorization: Bearer <admin-token>
```

#### Get User Details
```bash
GET /api/v1/admin/users/{user_id}
Authorization: Bearer <admin-token>
```

#### Update User
```bash
PUT /api/v1/admin/users/{user_id}
Authorization: Bearer <admin-token>
{
  "role": "admin",
  "status": "active"
}
```

#### Delete User
```bash
DELETE /api/v1/admin/users/{user_id}
Authorization: Bearer <admin-token>
```

#### Get Audit Logs
```bash
GET /api/v1/admin/audit-logs?page=1&limit=50
Authorization: Bearer <admin-token>
```

---

## Security Features

### Implemented
✅ JWT Authentication with Refresh Tokens
✅ OTP Login (Time-based One-Time Passwords)
✅ Google OAuth 2.0 Integration
✅ Password Hashing (bcrypt)
✅ Rate Limiting (configurable)
✅ CORS Protection
✅ Audit Logging
✅ Session Management
✅ SQL Injection Prevention
✅ XSS Protection Headers
✅ CSRF Protection

### Configuration

```env
# Password Policy
MIN_PASSWORD_LENGTH=8
REQUIRE_SPECIAL_CHARS=true

# Rate Limiting
RATE_LIMIT_LOGIN_ATTEMPTS=5
RATE_LIMIT_OTP_ATTEMPTS=3
RATE_LIMIT_WINDOW_SECONDS=900

# Token Expiry
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=7d
OTP_EXPIRY_MINUTES=5

# Session Management
MAX_ACTIVE_SESSIONS=5
SESSION_TIMEOUT=30m
```

---

## Kubernetes Deployment

### Prerequisites
- Kubernetes 1.28+
- kubectl configured
- Docker images pushed to registry

### Deploy to Kubernetes

```bash
# Apply database secret
kubectl apply -f kubernetes/secrets.yaml

# Deploy database
kubectl apply -f kubernetes/postgres-statefulset.yaml

# Deploy services
kubectl apply -f kubernetes/services/auth-service.yaml
kubectl apply -f kubernetes/services/api-gateway.yaml

# Check deployment
kubectl get pods
kubectl get services
```

### Environment Variables in Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
data:
  JWT_SECRET_KEY: "your-super-secret-key"
  CORS_ORIGINS: "https://your-domain.com"
  GIN_MODE: "release"
---
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
type: Opaque
stringData:
  DATABASE_URL: "postgresql://user:password@postgres:5432/mini_kubernet"
  GOOGLE_CLIENT_SECRET: "your-google-secret"
```

---

## Monitoring & Logging

### Health Checks

```bash
# Check auth service
curl http://localhost:8081/health

# Check frontend
curl http://localhost:3000/health
```

### View Logs

```bash
# Docker Compose
docker-compose logs auth-service -f

# Kubernetes
kubectl logs deployment/auth-service -f

# Local Go
# Check stdout when running `go run main.go`
```

### Prometheus Metrics

Metrics are exported at: `http://localhost:8081/metrics`

---

## Troubleshooting

### Database Connection Error
```bash
# Check database is running
docker-compose ps postgres

# Test connection
psql -U postgres -h localhost -d mini_kubernet -c "SELECT 1"

# Check DATABASE_URL in .env
```

### Port Already in Use
```bash
# Kill process on port
lsof -ti:8081 | xargs kill -9
lsof -ti:3000 | xargs kill -9
```

### JWT Secret Not Set
```bash
# Generate secure secret
openssl rand -base64 32

# Set in .env
JWT_SECRET_KEY=generated-secret-here
```

### Google OAuth Not Working
1. Verify Client ID matches between backend and frontend
2. Check redirect URIs in Google Console
3. Ensure `@react-oauth/google` is installed: `npm install @react-oauth/google`

---

## Performance Optimization

### Backend
- Connection pooling (25 max open connections)
- Gzip compression
- Static asset caching
- Database query optimization

### Frontend
- Code splitting with Vite
- Lazy loading routes
- CSS-in-JS optimization
- Image lazy loading
- Browser caching headers

### Database
- Proper indexing on user lookups
- Automatic index creation
- Connection pooling

---

## Scaling Recommendations

### Horizontal Scaling
- Run multiple auth-service instances behind load balancer
- Use Kubernetes HPA for auto-scaling
- Redis for distributed session storage

### Database Scaling
- Read replicas for queries
- Write master for mutations
- Connection pooling via PgBouncer

### Caching
- Redis for OTP codes
- Session caching
- API response caching

---

## Production Checklist

Before deploying to production:

- [ ] Change all default passwords
- [ ] Generate strong JWT secret
- [ ] Configure proper CORS origins
- [ ] Set up HTTPS/TLS certificates
- [ ] Enable audit logging
- [ ] Configure rate limiting
- [ ] Set up monitoring & alerting
- [ ] Backup database regularly
- [ ] Enable database encryption
- [ ] Review and update dependencies
- [ ] Run security audit
- [ ] Configure firewall rules
- [ ] Set up log aggregation
- [ ] Test failover & recovery

---

## Additional Resources

- [Go Documentation](https://golang.org/doc)
- [React Documentation](https://react.dev)
- [PostgreSQL Documentation](https://www.postgresql.org/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs)
- [Docker Documentation](https://docs.docker.com)

---

## Support & Issues

For issues or questions:
1. Check the logs: `docker-compose logs`
2. Review environment variables
3. Verify database connectivity
4. Check API endpoints with curl/Postman

