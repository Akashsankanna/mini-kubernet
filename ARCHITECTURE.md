# Architecture

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      Users / Clients                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│              AWS Route 53 (Global Load Balancer)                  │
│                   (Multi-Region DNS)                              │
└────────────────────────────┬────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        │                    │                    │
        ▼                    ▼                    ▼
┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
│   US East 1      │ │   US West 2      │ │  EU West 1       │
│  EKS Cluster     │ │  EKS Cluster     │ │  EKS Cluster     │
└─────────┬────────┘ └─────────┬────────┘ └─────────┬────────┘
          │                    │                    │
          └────────────────────┼────────────────────┘
                               │
                               ▼
                    ┌──────────────────────┐
                    │   API Gateway (LB)   │
                    │   (Istio IngressGW)  │
                    └──────────┬───────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
        ▼                      ▼                      ▼
   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
   │   Frontend  │    │ Auth Svc    │    │Build Svc    │
   │   React     │    │   (3 pods)  │    │  (2 pods)   │
   │  (3 pods)   │    └─────────────┘    └─────────────┘
   └─────────────┘           │
        │                    │
        └────────────────────┴─────────────┐
                                           │
                                           ▼
                                    ┌──────────────┐
                                    │Deploy Service│
                                    │  (3 pods)    │
                                    └──────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                     Service Mesh (Istio)                          │
│                                                                   │
│  • mTLS Encryption (Pod to Pod)                                  │
│  • Traffic Management (VirtualService, DestinationRule)          │
│  • Security Policies (AuthorizationPolicy)                       │
│  • Observability (Metrics, Tracing)                              │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Data Layer (AWS)                               │
├─────────────────────────────────────────────────────────────────┤
│  PostgreSQL       │  Redis Cache      │  RDS (Multi-AZ)         │
│  (Primary)        │  (Session Store)  │  (Replicated)           │
│                   │                   │                         │
│  - User data      │  - Tokens         │  - Connection pooling   │
│  - Deployments    │  - Rate limiting  │  - Encryption at rest   │
│  - History        │  - Message Queue  │  - Daily backups        │
└───────────────────┴───────────────────┴─────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│              Monitoring & Observability                           │
├─────────────────────────────────────────────────────────────────┤
│  Prometheus       │  Grafana          │  Jaeger (Tracing)       │
│  - 15s scrape     │  - Dashboards     │  - Service dependency   │
│  - 30d retention  │  - Alerts         │  - Request tracing      │
│  - Alert rules    │  - User mgmt      │  - Performance analysis │
└───────────────────┴───────────────────┴─────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                  CI/CD Pipeline (GitHub Actions)                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  Git Push → Test → Build → Scan → Push → Deploy → Monitor        │
│     │        │      │       │       │       │        │           │
│     ▼        ▼      ▼       ▼       ▼       ▼        ▼           │
│   Code     Unit   Docker  Security GCR   Canary  Prometheus     │
│   Tests    Tests  Image   Scan    Image Deploy   Metrics        │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## Component Interaction Flow

```
User Request
    │
    ▼
Route 53 (DNS)
    │
    ▼
Load Balancer (ALB/NLB)
    │
    ▼
Istio Ingress Gateway
    │
    ▼
Istio Virtual Service
    │
    ├──→ Request Routing (weighted distribution)
    │       • 90% v1 (stable)
    │       • 10% v2 (canary)
    │
    ▼
Istio Destination Rule
    │
    ├──→ Load Balancing (Round Robin)
    ├──→ Circuit Breaking
    ├──→ Connection Pool Management
    │
    ▼
Service Pod (Envoy Sidecar)
    │
    ├──→ mTLS Encryption
    ├──→ Observability (Metrics)
    ├──→ Rate Limiting
    │
    ▼
Application Container
    │
    ├──→ Auth Service (JWT validation)
    ├──→ Business Logic
    ├──→ Data Access
    │
    ▼
Database / Cache
    │
    ├──→ PostgreSQL (persistent data)
    ├──→ Redis (session/cache)
    │
    ▼
Response
    │
    ▼
User
```

## Deployment Stages

### Stage 1: Development
```
Local Docker Compose
├── All services running
├── Full database
├── Real-time logs
└── Easy debugging
```

### Stage 2: Staging
```
Single EKS Cluster
├── Full infrastructure
├── Testing workloads
├── Staging database
└── Integration tests
```

### Stage 3: Production (Multi-Region)
```
3x EKS Clusters (us-east, us-west, eu-west)
├── Multi-region deployment
├── Active-active setup
├── Read replicas
├── Disaster recovery
└── Global load balancing
```

## Auto-Scaling Strategy

```
┌──────────────────────────────────┐
│   Prometheus Metrics Collection  │
│   (15s scrape interval)          │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│   AI Scaler Pod (Every 5 min)    │
│   • CPU Predictor (LSTM)         │
│   • Memory Predictor (GB)        │
│   • Request Rate Predictor (RPS) │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│   Calculate Desired Replicas     │
│   • Min replicas: 1-3            │
│   • Max replicas: 5-10           │
│   • Scaling factor: utilization  │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│   Update Deployment Replicas     │
│   via Kubernetes API             │
└────────────┬─────────────────────┘
             │
             ▼
┌──────────────────────────────────┐
│   Monitor New Pod Status         │
│   • Readiness probe              │
│   • Liveness probe               │
│   • Metrics convergence          │
└──────────────────────────────────┘
```

## Data Flow

```
Request → API Gateway → Service Discovery
                             │
                    ┌────────┼────────┐
                    │        │        │
              Auth Service Build Svc Deploy Svc
                    │        │        │
                    └────────┼────────┘
                             │
                         Database
                        PostgreSQL
                             │
                    ┌────────┴────────┐
                    │                 │
              Response           Cache (Redis)
                    │                 │
                    └────────┬────────┘
                             │
                           User
```

## Security Layers

```
┌─────────────────────────────────────┐
│   External (Internet)               │
└────────────────┬────────────────────┘
                 │ TLS/HTTPS
                 ▼
┌─────────────────────────────────────┐
│   AWS WAF (DDoS Protection)         │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Load Balancer (ALB/NLB)           │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   AWS Security Groups               │
│   (Ingress/Egress Rules)            │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Kubernetes Network Policies       │
│   (Pod-to-Pod Rules)                │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Istio mTLS                        │
│   (Service-to-Service Encryption)   │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   RBAC Authorization                │
│   (Service Accounts & Roles)        │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   JWT Token Authentication          │
│   (API Level)                       │
└────────────────┬────────────────────┘
                 │
                 ▼
┌─────────────────────────────────────┐
│   Database (Encrypted at Rest)      │
└─────────────────────────────────────┘
```

## Disaster Recovery

```
Primary Cluster (Active)           Secondary Cluster (Standby)
        │                                     │
        │ Continuous Replication              │
        ├────────────────────────────────────→│
        │                                     │
        │ Health Check Failed                 │
        │ Route 53 Failover Triggered         │
        │                                     │
        └─────────────X       Start Service   │
                              ▼               │
                         Secondary Active ←──┘
```
