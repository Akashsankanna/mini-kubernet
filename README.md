<<<<<<< HEAD
# Advanced Kubernetes Deployment Platform

A production-grade, highly scalable Kubernetes platform featuring multi-cluster deployment, canary deployments, Istio service mesh, and AI-based auto-scaling.

## Architecture Overview

```
Frontend (React) 
    ↓
API Gateway (Go/Fiber)
    ↓
────────────────────────────
│         │         │
Auth    Build    Deploy
Service Service Service
│         │         │
JWT    Docker  Kubernetes
API    API     API
│
PostgreSQL Layer
│
Redis Cache + NATS Message Queue
```

## Features

### 1. Multi-Cluster Deployment
- Deploy across multiple AWS regions
- Automated failover and load balancing
- Consistent state management across clusters
- Global traffic distribution

### 2. Canary Deployment
- Gradual traffic shifting (10% → 100%)
- Automated rollback on metrics threshold
- A/B testing capabilities
- Zero-downtime deployments

### 3. Istio Service Mesh
- Advanced traffic routing
- Mutual TLS (mTLS) security
- Service-to-service authentication
- Distributed tracing integration
- Traffic observability

### 4. AI-Based Auto Scaling
- ML-driven metrics prediction
- Proactive resource allocation
- Cost optimization
- Custom scaling policies

## Tech Stack

| Layer | Technology |
|-------|------------|
| **Frontend** | React 18, Redux, Tailwind CSS |
| **Backend** | Go 1.21, Gin/Fiber |
| **Database** | PostgreSQL 15, Redis 7.0 |
| **Container** | Docker 24 |
| **Orchestration** | Kubernetes 1.28 |
| **Service Mesh** | Istio 1.17 |
| **CI/CD** | GitHub Actions |
| **Infrastructure** | Terraform, AWS |
| **Monitoring** | Prometheus, Grafana |
| **Message Queue** | Redis, NATS |
| **IaC** | Terraform 1.5+ |

## Project Structure

```
mini-kubernet/
├── backend/
│   ├── auth-service/          # JWT authentication
│   ├── build-service/         # Docker image building
│   ├── deploy-service/        # Kubernetes deployment
│   └── api-gateway/           # Main API gateway
├── frontend/                  # React application
├── kubernetes/                # K8s manifests
│   ├── namespaces/
│   ├── services/
│   ├── deployments/
│   ├── canary/
│   ├── multi-cluster/
│   ├── istio/
│   └── monitoring/
├── terraform/                 # Infrastructure as Code
│   ├── aws/
│   ├── networking/
│   └── databases/
├── monitoring/                # Prometheus & Grafana configs
├── scripts/                   # Utility scripts
├── .github/workflows/         # CI/CD pipelines
├── Dockerfile                 # Multi-service docker setup
├── docker-compose.yml         # Local development
├── Makefile                   # Build automation
└── README.md

```

## Prerequisites

```
Go 1.21+
Node.js 18+
Docker 24+
Kubernetes 1.28+
Terraform 1.5+
Helm 3.12+
Istio 1.17+
AWS CLI v2
kubectl 1.28+
```

## Quick Start

### 1. Local Development Setup

```bash
# Clone repository
git clone https://github.com/advanced-k8s/mini-kubernet.git
cd mini-kubernet

# Install dependencies
go mod download
npm install --prefix frontend

# Start local environment
docker-compose up -d

# Run tests
make test
```

### 2. Build & Push Images

```bash
# Build Docker images
make docker-build

# Push to registry (set DOCKER_REGISTRY)
DOCKER_REGISTRY=gcr.io/your-project make docker-push
```

### 3. Deploy to Kubernetes

```bash
# Deploy infrastructure
make deploy-infra

# Install Istio
make install-istio

# Deploy applications
make deploy-k8s

# Deploy canary version
make deploy-canary
```

### 4. Multi-Cluster Deployment

```bash
# Deploy to multiple clusters
CLUSTERS="us-east-1 us-west-2 eu-west-1" make deploy-multi-cluster
```

## Services

### Auth Service
- JWT token generation and validation
- User authentication
- Role-based access control (RBAC)

**Endpoints:**
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Refresh token
- `GET /auth/validate` - Validate token

### Build Service
- Docker image building
- Container registry integration
- Build pipeline orchestration

**Endpoints:**
- `POST /build/create` - Start build
- `GET /build/{id}` - Get build status
- `DELETE /build/{id}` - Cancel build

### Deploy Service
- Kubernetes deployment management
- Version management
- Health checks

**Endpoints:**
- `POST /deploy/create` - Create deployment
- `GET /deploy/{id}` - Get deployment status
- `PATCH /deploy/{id}/rollback` - Rollback deployment
- `DELETE /deploy/{id}` - Delete deployment

### API Gateway
- Request routing
- Rate limiting
- API versioning

**Base URL:** `https://api.example.com/v1`

## Kubernetes Manifests

### Basic Deployments
```bash
kubectl apply -f kubernetes/deployments/
```

### Service Mesh (Istio)
```bash
kubectl apply -f kubernetes/istio/
```

### Canary Deployment
```bash
kubectl apply -f kubernetes/canary/
# Roll out: 10% → 25% → 50% → 100%
```

### Multi-Cluster Setup
```bash
kubectl apply -f kubernetes/multi-cluster/
```

## Monitoring

### Prometheus
- Metrics collection
- Custom dashboards
- Alert rules

Access: `http://prometheus.example.com`

### Grafana
- Visualization dashboards
- Alert management
- User management

Access: `http://grafana.example.com`
Default: admin/admin (change on first login)

## CI/CD Pipeline

GitHub Actions workflows included:

1. **Test Pipeline** (`.github/workflows/test.yml`)
   - Unit tests
   - Integration tests
   - Code coverage

2. **Build Pipeline** (`.github/workflows/build.yml`)
   - Docker image build
   - Security scanning
   - Registry push

3. **Deploy Pipeline** (`.github/workflows/deploy.yml`)
   - Canary deployment
   - Health checks
   - Automated rollback

4. **Multi-Cluster Pipeline** (`.github/workflows/multi-cluster.yml`)
   - Parallel deployment
   - Cross-cluster sync

## Database

### PostgreSQL
- User data
- Deployment history
- Configuration storage

**Connection:** `postgresql://user:pass@postgres:5432/kubernet`

### Redis
- Session cache
- Rate limiting
- Message queue

**Connection:** `redis://redis:6379/0`

## API Examples

### Authentication
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

### Deploy Service
```bash
curl -X POST http://localhost:8082/deploy/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {token}" \
  -d '{
    "service": "api-gateway",
    "image": "gcr.io/project/api-gateway:v1.0.0",
    "replicas": 3
  }'
```

## Scaling

### Auto-Scaling Configuration
```yaml
kind: HorizontalPodAutoscaler
apiVersion: autoscaling/v2
metadata:
  name: api-gateway-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## Security

- **Service-to-Service:** mTLS via Istio
- **Authentication:** JWT tokens
- **Authorization:** RBAC policies
- **Encryption:** TLS for all traffic
- **Network Policies:** Istio VirtualServices

## Troubleshooting

### Check Service Status
```bash
kubectl get pods -n default
kubectl logs -f deployment/api-gateway
kubectl describe pod {pod-name}
```

### View Metrics
```bash
kubectl top nodes
kubectl top pods
```

### Istio Debugging
```bash
kubectl logs -n istio-system -l app=istiod
istioctl analyze
istioctl proxy-config listeners {pod-name}
```

## Contributing

1. Create feature branch: `git checkout -b feature/xyz`
2. Commit changes: `git commit -m "Add feature xyz"`
3. Push branch: `git push origin feature/xyz`
4. Open Pull Request

## License

MIT License - See LICENSE file

## Support

For issues and questions:
- GitHub Issues: https://github.com/advanced-k8s/mini-kubernet/issues
- Documentation: See `/docs` folder
- Email: support@example.com
=======
# mini-kubernet
Cloud-native microservices platform built with Go, React, Kubernetes, Docker, Terraform, ArgoCD, and GitHub Actions. Features JWT authentication, PostgreSQL, Redis, NATS messaging, CI/CD automation, GitOps deployment, monitoring with Prometheus/Grafana, Istio service mesh, and scalable cloud-native architecture.
>>>>>>> c9e41752e2f3ddae54ed269b1196e95747e8884e
