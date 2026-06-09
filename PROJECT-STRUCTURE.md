# Project Structure Documentation

## Directory Overview

```
mini-kubernet/
│
├── backend/                          # Go microservices
│   ├── auth-service/                 # Authentication & JWT
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   └── go.mod
│   │
│   ├── build-service/                # Docker image building
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   └── go.mod
│   │
│   ├── deploy-service/               # Kubernetes deployment
│   │   ├── Dockerfile
│   │   ├── main.go
│   │   └── go.mod
│   │
│   └── api-gateway/                  # Main API gateway
│       ├── Dockerfile
│       ├── main.go
│       └── go.mod
│
├── frontend/                         # React.js application
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   ├── App.js
│   │   ├── index.js
│   │   └── components/
│   ├── Dockerfile
│   ├── package.json
│   └── .gitignore
│
├── kubernetes/                       # Kubernetes manifests
│   ├── namespace.yaml               # Namespaces
│   ├── postgres-statefulset.yaml    # PostgreSQL database
│   ├── redis-deployment.yaml        # Redis cache
│   │
│   ├── services/                    # Microservices deployments
│   │   ├── auth-service.yaml
│   │   ├── build-service.yaml
│   │   ├── deploy-service.yaml
│   │   └── api-gateway.yaml
│   │
│   ├── istio/                       # Service mesh configs
│   │   └── virtual-service.yaml     # Traffic routing & mTLS
│   │
│   ├── hpa/                         # Auto-scaling
│   │   └── autoscaler.yaml          # HPA policies
│   │
│   ├── canary/                      # Canary deployment
│   │   ├── api-gateway-canary.yaml
│   │   └── virtual-service-canary.yaml
│   │
│   ├── ai-scaling/                  # AI-based auto-scaler
│   │   └── ai-scaler.yaml
│   │
│   ├── monitoring/                  # Prometheus & Grafana
│   │   ├── prometheus.yaml
│   │   └── grafana.yaml
│   │
│   └── multi-cluster/               # Multi-region configs
│       └── cluster-config.yaml
│
├── terraform/                       # Infrastructure as Code
│   ├── main.tf                     # Provider configuration
│   ├── variables.tf                # Variable definitions
│   ├── networking.tf               # VPC, subnets, IGW
│   ├── eks.tf                      # EKS cluster setup
│   ├── databases.tf                # RDS, ElastiCache
│   ├── outputs.tf                  # Outputs
│   └── terraform.tfvars            # Variable values (gitignored)
│
├── monitoring/                      # Monitoring configurations
│   ├── prometheus-config.yaml      # Scrape configs & rules
│   └── alert-rules.yaml            # Alert definitions
│
├── scripts/                         # Utility scripts
│   ├── deploy.sh                   # Deploy to K8s
│   ├── canary-deploy.sh            # Canary deployment
│   ├── multi-cluster-deploy.sh     # Multi-region deploy
│   ├── install-istio.sh            # Istio installation
│   └── ai-scaler.py                # AI scaling logic
│
├── .github/                         # GitHub configuration
│   └── workflows/                  # CI/CD pipelines
│       ├── test.yml                # Run tests
│       ├── build.yml               # Build & push images
│       ├── deploy.yml              # Deploy to K8s
│       └── multi-cluster.yml       # Multi-cluster deploy
│
├── .gitignore                       # Git ignore rules
├── .env.example                     # Environment template
├── docker-compose.yml               # Local development
├── docker-compose.prod.yml          # Production compose
├── go.mod                          # Go modules
├── Makefile                         # Build automation
│
├── README.md                        # Quick start guide
├── SETUP.md                         # Installation instructions
├── ARCHITECTURE.md                  # System design
├── ADVANCED-FEATURES.md             # Feature documentation
├── API.md                           # API reference
├── CONTRIBUTING.md                  # Development guide
└── LICENSE                          # MIT License
```

## Key Files Description

### Backend Services
- **auth-service**: JWT authentication, token validation, user management
- **build-service**: Docker image building, registry pushing
- **deploy-service**: Kubernetes deployment orchestration
- **api-gateway**: Request routing, rate limiting, API versioning

### Kubernetes
- **Deployments**: Define how services run
- **Services**: Expose services internally/externally
- **StatefulSets**: Stateful components (PostgreSQL)
- **VirtualServices**: Istio traffic routing
- **DestinationRules**: Connection pooling & LB policies
- **HPA**: Horizontal Pod Autoscalers
- **Namespace**: Logical isolation

### Infrastructure
- **EKS Cluster**: AWS Elastic Kubernetes Service
- **VPC**: Network isolation
- **RDS**: PostgreSQL database (Multi-AZ)
- **ElastiCache**: Redis for caching
- **Security Groups**: Network access control

### CI/CD
- **Test Workflow**: Unit tests, coverage
- **Build Workflow**: Docker image creation
- **Deploy Workflow**: Canary & stable deployments
- **Multi-Cluster**: Regional deployment coordination

## Service Dependencies

```
Frontend (React)
    ↓
API Gateway (Port 8080)
    ├── → Auth Service (Port 8081) → PostgreSQL
    ├── → Build Service (Port 8082) → Docker Daemon
    └── → Deploy Service (Port 8083) → Kubernetes API

Cache Layer: Redis
Message Queue: NATS/Redis
Database: PostgreSQL
```

## Configuration Hierarchy

```
Environment Defaults (.env.example)
        ↓
User Override (.env)
        ↓
Kubernetes ConfigMaps
        ↓
Environment Variables
        ↓
Application
```

## Data Storage

- **PostgreSQL**: User data, deployments, history
- **Redis**: Sessions, rate limits, message queue
- **EBS Volumes**: Persistent storage (PVC)
- **S3**: Build artifacts, logs, backups

## Network Architecture

```
Internet → Route 53 (DNS)
    ↓
Load Balancer (ALB/NLB)
    ↓
Istio Ingress Gateway (Port 80/443)
    ↓
Kubernetes Service (ClusterIP)
    ↓
Pod (with Envoy sidecar)
    ↓
Application Container
```

## Logging & Monitoring

- **Logs**: kubectl logs, CloudWatch
- **Metrics**: Prometheus (15s scrape)
- **Dashboards**: Grafana
- **Tracing**: Jaeger (future)
- **Alerts**: AlertManager

## Scaling Configuration

```yaml
HPA Rules:
- API Gateway: 3-10 replicas, 70% CPU
- Auth Service: 2-8 replicas, 75% CPU
- Build Service: 1-5 replicas, 80% CPU
- Deploy Service: 2-6 replicas, 70% CPU

AI Scaler:
- Predicts 30min ahead
- Uses LSTM, GB, ARIMA models
- Proactive scaling 5min before load
```

## Security Model

```
Layer 1: Network
- VPC with private subnets
- Security groups
- Network policies

Layer 2: API
- JWT authentication
- RBAC authorization
- Rate limiting

Layer 3: Transport
- TLS/HTTPS
- mTLS (Istio)
- Certificate management

Layer 4: Data
- Encryption at rest
- Encryption in transit
- Secrets management
```

## Performance Targets

| Metric | Target | Threshold |
|--------|--------|-----------|
| API Response Time (p95) | < 100ms | 200ms |
| Database Query | < 50ms | 100ms |
| Cache Hit Rate | > 90% | < 80% |
| Error Rate | < 0.1% | > 0.5% |
| Availability | 99.95% | 99.9% |
| Pod Startup | < 30s | 60s |

## Common Tasks

### Deploy
```bash
make deploy-k8s
```

### Scale
```bash
kubectl scale deployment api-gateway --replicas=5
```

### Update Image
```bash
kubectl set image deployment/api-gateway api-gateway=new:image
```

### View Logs
```bash
kubectl logs -f deployment/api-gateway
```

### Port Forward
```bash
kubectl port-forward svc/api-gateway 8080:8080
```

### Execute in Pod
```bash
kubectl exec -it <pod> -- sh
```

### Debug
```bash
kubectl describe pod <pod>
kubectl events <pod>
```
