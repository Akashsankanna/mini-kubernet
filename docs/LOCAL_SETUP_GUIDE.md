# Quick Start: Running Mini Kubernet Locally

## Prerequisites

- Docker Desktop (with Kubernetes enabled)
- kubectl >= 1.28
- helm >= 3.12
- Go >= 1.21 (for backend development)
- Node.js >= 18 (for frontend development)

## Local Development Setup

### 1. Start Docker Desktop

Enable Kubernetes in Docker Desktop preferences.

### 2. Verify Kubernetes

```bash
kubectl cluster-info
kubectl get nodes
```

### 3. Clone Repository

```bash
git clone https://github.com/mini-kubernet/mini-kubernet.git
cd mini-kubernet
```

### 4. Create Namespace

```bash
kubectl create namespace kubernet-local
kubectl config set-context --current --namespace=kubernet-local
```

### 5. Start Infrastructure with Docker Compose

```bash
docker-compose up -d
```

This starts:
- PostgreSQL database
- Redis cache
- NATS message queue

### 6. Deploy with Helm (Dev Environment)

```bash
./scripts/deployment/helm-deploy.sh local
```

Or manually:

```bash
helm install auth-service helm/charts/auth-service \
  --values helm/charts/auth-service/values.yaml \
  --values helm/charts/auth-service/values-dev.yaml \
  -n kubernet-local
```

### 7. Forward Ports

```bash
# Frontend
kubectl port-forward -n kubernet-local svc/frontend 3000:80

# API Gateway
kubectl port-forward -n kubernet-local svc/api-gateway 8080:8080

# Auth Service
kubectl port-forward -n kubernet-local svc/auth-service 8081:8081

# Prometheus
kubectl port-forward -n kubernet-local svc/prometheus 9090:9090

# Grafana
kubectl port-forward -n kubernet-local svc/grafana 3000:80
```

### 8. Access Services

- **Frontend**: http://localhost:3000
- **API Gateway**: http://localhost:8080
- **Auth Service**: http://localhost:8081
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000

## Development Workflow

### Backend Development

```bash
# 1. Build auth-service locally
cd backend/auth-service
go build -o auth-service main.go

# 2. Run locally
./auth-service

# 3. Run tests
go test -v ./...
```

### Frontend Development

```bash
# 1. Install dependencies
cd frontend
npm install

# 2. Start dev server
npm run dev

# 3. Run tests
npm test
```

### Docker-based Development

```bash
# Build images
docker-compose build

# Run services
docker-compose up

# View logs
docker-compose logs -f auth-service

# Stop services
docker-compose down
```

## Testing Locally

### Run Unit Tests

```bash
./tests/unit/run-unit-tests.sh
```

### Run E2E Tests

```bash
./tests/e2e/run-e2e-tests.sh local
```

### Test Authentication Flow

```bash
# 1. Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "full_name": "Test User"
  }'

# 2. Login with password
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!"
  }'

# 3. Request OTP
curl -X POST http://localhost:8080/api/v1/auth/login/otp/request \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'

# 4. Verify OTP
curl -X POST http://localhost:8080/api/v1/auth/login/otp/verify \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "otp": "123456"
  }'
```

## Troubleshooting

### Pod not starting

```bash
# Check pod status
kubectl describe pod <pod-name>

# View logs
kubectl logs <pod-name>

# Check resource limits
kubectl top nodes
kubectl top pods
```

### Database connection errors

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Test connection
kubectl exec -it <pod> -- \
  psql -h host.docker.internal -U postgres -c "SELECT 1;"
```

### Frontend not loading

```bash
# Check if frontend service is running
kubectl get svc frontend

# Check ingress
kubectl get ingress

# View frontend logs
kubectl logs -l app=frontend
```

## Next Steps

1. **Read Documentation**: Check `docs/` directory
2. **Explore APIs**: Read `API.md`
3. **Setup CI/CD**: Configure GitHub Actions
4. **Deploy to Cloud**: Follow `docs/DEPLOYMENT_GUIDE.md`

## Support

- Issues: GitHub Issues
- Discussion: GitHub Discussions
- Docs: [README.md](README.md)
- Architecture: [docs/ARCHITECTURE_DETAILED.md](docs/ARCHITECTURE_DETAILED.md)
