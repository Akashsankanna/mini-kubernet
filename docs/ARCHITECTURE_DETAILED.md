# Mini Kubernet Architecture & Design

## System Overview

Mini Kubernet is a cloud-native microservices platform built on Kubernetes with enterprise-grade security, observability, and deployment automation.

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        External Users                        │
└────────────────────┬────────────────────────────────────────┘
                     │ HTTPS
┌────────────────────▼────────────────────────────────────────┐
│              AWS Application Load Balancer                   │
└────────────────────┬────────────────────────────────────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│         Kubernetes Ingress Controller (NGINX)               │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │         Service Mesh (Istio)                         │   │
│  │  ┌──────────────┬──────────────┬──────────────┐     │   │
│  │  │ Auth Service │ API Gateway  │   Frontend   │     │   │
│  │  │              │              │              │     │   │
│  │  │ ┌──────────┐ │ ┌──────────┐│ ┌──────────┐ │     │   │
│  │  │ │Pod       │ │ │Pod       ││ │Pod       │ │     │   │
│  │  │ │Pod       │ │ │Pod       ││ │Pod       │ │     │   │
│  │  │ └──────────┘ │ └──────────┘│ └──────────┘ │     │   │
│  │  │              │              │              │     │   │
│  │  │ Services (4  │ Services (3  │ Services (3  │     │   │
│  │  │ replicas)    │ replicas)    │ replicas)    │     │   │
│  │  │              │              │              │     │   │
│  │  └──────────────┴──────────────┴──────────────┘     │   │
│  │                                                      │   │
│  │  ┌──────────────┬──────────────┬──────────────┐     │   │
│  │  │Build Service │Deploy Service│  gRPC APIs   │     │   │
│  │  └──────────────┴──────────────┴──────────────┘     │   │
│  │                                                      │   │
│  │  Monitoring: Prometheus, Grafana, Loki             │   │
│  │  Security: Pod Security Policies, RBAC, Network    │   │
│  │  Policies, Vault Secret Management                │   │
│  └──────────────────────────────────────────────────────┘   │
│                                                              │
│         Kubernetes Cluster (1.28+)                         │
│         - 3+ Worker Nodes                                  │
│         - Auto-scaling enabled                            │
│         - Multi-AZ deployment                             │
└──────────────┬──────────────┬──────────────┬───────────────┘
               │              │              │
        ┌──────▼──┐    ┌──────▼──┐    ┌─────▼──────┐
        │   AWS   │    │   AWS   │    │   AWS      │
        │   RDS   │    │Elastic  │    │   Secrets  │
        │Postgres │    │ Cache   │    │  Manager   │
        │         │    │ Redis   │    │            │
        └─────────┘    └─────────┘    └────────────┘
```

## Service Architecture

### Authentication Service (auth-service)
- **Responsibility**: User authentication and authorization
- **Features**:
  - Password-based login with bcrypt hashing
  - OTP (One-Time Password) authentication
  - Google OAuth 2.0 integration
  - JWT token generation and refresh
  - Session management
  - Audit logging
- **API Endpoints**:
  - `POST /api/v1/auth/register` - User registration
  - `POST /api/v1/auth/login` - Password login
  - `POST /api/v1/auth/login/otp/request` - Request OTP
  - `POST /api/v1/auth/login/otp/verify` - Verify OTP
  - `POST /api/v1/auth/oauth/google` - Google OAuth login
  - `POST /api/v1/auth/refresh` - Refresh JWT token

### API Gateway (api-gateway)
- **Responsibility**: Request routing, rate limiting, and request transformation
- **Features**:
  - Request routing to microservices
  - Rate limiting (5 req/IP for OTP, 100 req/IP for general API)
  - Request/response logging
  - JWT validation
  - CORS handling
  - Request transformation
- **Routes**:
  - `/api/v1/auth/*` → auth-service
  - `/api/v1/users/*` → user-service
  - `/api/v1/projects/*` → project-service

### Build Service (build-service)
- **Responsibility**: Container image building and registry management
- **Features**:
  - Docker image building from source
  - Image tagging and versioning
  - ECR integration
  - Build caching
  - Build logs and metrics

### Deploy Service (deploy-service)
- **Responsibility**: Kubernetes deployment management
- **Features**:
  - Deployment automation
  - Rollout management
  - Health check monitoring
  - Automatic rollback on failure
  - Environment-specific configurations

### Frontend (frontend)
- **Responsibility**: User interface
- **Stack**: React 18 + Tailwind CSS + Vite
- **Features**:
  - Responsive design
  - State management with Zustand
  - Protected routes
  - Multi-step authentication (password, OTP, Google OAuth)
  - Real-time updates via WebSocket

## Data Flow

### Authentication Flow

```
User Input
    │
    ▼
Frontend (React)
    │
    ├─► Password Login Flow
    │       │
    │       ▼
    │   API Gateway
    │       │
    │       ▼
    │   Auth Service
    │       │
    │       ├─► PostgreSQL (user lookup)
    │       ├─► Redis (session store)
    │       └─► JWT generation
    │
    ├─► OTP Login Flow
    │       │
    │       ▼
    │   API Gateway
    │       │
    │       ▼
    │   Auth Service
    │       │
    │       ├─► Generate 6-digit OTP (crypto/rand)
    │       ├─► PostgreSQL (store otp_records)
    │       ├─► Email Service (send OTP)
    │       └─► Frontend (auto-fill OTP in dev mode)
    │
    └─► Google OAuth Flow
            │
            ▼
        Google Endpoints
            │
            ▼
        Auth Service
            │
            ├─► Verify ID token
            ├─► PostgreSQL (user lookup/creation)
            └─► JWT generation
    
    ▼
JWT + Refresh Token
    │
    ▼
Frontend (Store in localStorage/sessionStorage)
    │
    ▼
Authenticated API Requests
```

## Data Models

### User Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(255),
    avatar_url VARCHAR(255),
    oauth_provider VARCHAR(50),
    oauth_id VARCHAR(255),
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### OTP Records Table
```sql
CREATE TABLE otp_records (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    otp_code VARCHAR(6) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    attempts INT DEFAULT 0,
    is_used BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Sessions Table
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Audit Log Table
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    action VARCHAR(50),
    resource VARCHAR(255),
    details JSONB,
    ip_address VARCHAR(45),
    user_agent VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Deployment Strategy

### Multi-Environment Setup

```
Development Environment
├── 1 Replica per service
├── Lower resource limits
├── HPA disabled
├── Daily backups
└── 7-day retention

Staging Environment
├── 2 Replicas per service
├── Medium resource limits
├── HPA enabled (2-5 replicas)
├── Daily backups
└── 30-day retention

Production Environment
├── 5 Replicas per service
├── High resource limits
├── HPA enabled (5-20 replicas)
├── Continuous backups
└── 90-day retention
```

### CI/CD Pipeline

```
Git Push
    │
    ▼
GitHub Actions Trigger
    │
    ├─► Security Scan (Trivy)
    │       │
    │       ▼
    │   Code Quality (SonarQube)
    │
    ├─► Build
    │       │
    │       ├─► Go Services
    │       ├─► Frontend (Node.js)
    │       └─► Docker Images
    │
    ├─► Validate
    │       │
    │       ├─► Kubernetes Manifests
    │       ├─► Helm Charts
    │       └─► Terraform
    │
    ├─► Push to ECR
    │
    └─► Deploy
            │
            ├─► Dev (develop branch)
            ├─► Staging (staging branch)
            └─► Prod (main branch + manual approval)
```

## Scalability

### Horizontal Scaling
- **Load Balancing**: AWS ALB + Kubernetes Service
- **Auto-scaling**: HPA based on CPU (70%) and memory (80%)
- **Pod Disruption Budgets**: Ensures minimum availability during node drains
- **Pod Anti-affinity**: Distributes pods across nodes

### Vertical Scaling
- **Resource Requests**: Ensures minimum resources
- **Resource Limits**: Prevents resource hogging
- **QoS Classes**: Guaranteed quality of service

### Database Scaling
- **Read Replicas**: RDS multi-AZ with read replicas
- **Connection Pooling**: PgBouncer for connection management
- **Caching**: Redis for frequently accessed data

## High Availability

### Redundancy
- Multiple replicas of each service (minimum 3)
- Multi-AZ deployment for RDS and Redis
- Automated failover for database and cache

### Health Checks
- **Liveness Probe**: Restarts unhealthy pods
- **Readiness Probe**: Removes unhealthy pods from load balancer
- **Startup Probe**: Ensures pod is ready before traffic

### Disaster Recovery
- **RTO**: 1 hour (Recovery Time Objective)
- **RPO**: 5 minutes (Recovery Point Objective)
- **Backup Strategy**: Automated daily backups with 90-day retention
- **Recovery Testing**: Monthly recovery drills

## Security Architecture

### Network Security
- **Network Policies**: Ingress/egress rules by namespace
- **Service Mesh**: Istio for mTLS between services
- **Ingress TLS**: Let's Encrypt certificates
- **DDoS Protection**: AWS Shield

### Secret Management
- **Vault Integration**: Centralized secret storage
- **External Secrets Operator**: Automatic secret injection
- **Rotation**: Automatic secret rotation
- **Audit**: All secret access logged

### Pod Security
- **Security Context**: Non-root users, read-only filesystems
- **RBAC**: Least privilege access
- **Network Policies**: Restrict pod-to-pod communication
- **Pod Security Standards**: Enforce baseline standards

### Image Security
- **Scanning**: Trivy scans all images before deployment
- **Registry**: Private ECR with encryption
- **Signing**: Images signed with Docker Content Trust
- **Policies**: Admission controllers enforce image policies

## Observability

### Metrics
- **Collection**: Prometheus scrapes metrics every 15s
- **Storage**: 30-day retention
- **Visualization**: Grafana dashboards
- **Alerting**: Alert Manager for critical metrics

### Logging
- **Aggregation**: Grafana Loki
- **Collection**: Promtail DaemonSet on all nodes
- **Retention**: 30-day retention
- **Querying**: LogQL for advanced queries

### Tracing
- **Instrumentation**: OpenTelemetry
- **Export**: Jaeger for distributed tracing
- **Sampling**: 10% sampling for production
- **Storage**: 7-day retention

## Cost Optimization

### Resource Management
- **Requests/Limits**: Prevents overprovisioning
- **HPA**: Scales down during off-peak hours
- **Spot Instances**: 70% cost savings on compute
- **Reserved Instances**: 40% discount for predictable workloads

### Storage
- **Lifecycle Policies**: Archive old logs to S3
- **Compression**: Compress logs to reduce storage
- **Cleanup**: Delete old backups after retention

### Network
- **VPC Endpoints**: Reduced data transfer costs
- **CloudFront**: CDN for static assets
- **Reserved Bandwidth**: Bulk discounts

## Future Roadmap

- [ ] Multi-region deployment
- [ ] Machine learning for anomaly detection
- [ ] Advanced API versioning
- [ ] GraphQL API
- [ ] WebSocket support
- [ ] Event streaming (Kafka)
- [ ] Advanced analytics
