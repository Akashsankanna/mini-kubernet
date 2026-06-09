# Mini Kubernet Enterprise Deployment Guide

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [Infrastructure Setup](#infrastructure-setup)
3. [Kubernetes Cluster Setup](#kubernetes-cluster-setup)
4. [Application Deployment](#application-deployment)
5. [Post-Deployment Validation](#post-deployment-validation)
6. [Operations & Maintenance](#operations--maintenance)

## Prerequisites

### Required Tools
- `terraform` >= 1.5.0
- `kubectl` >= 1.28
- `helm` >= 3.12
- `aws-cli` >= 2.10
- `docker` >= 24.0 (for local builds)
- `git` >= 2.35

### AWS Credentials
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
```

### Environment Variables
```bash
export ENVIRONMENT=prod          # dev, staging, or prod
export ECR_REGISTRY="123456789.dkr.ecr.us-east-1.amazonaws.com"
export DOMAIN="mini-kubernet.com"
```

## Infrastructure Setup

### 1. Initialize Terraform

```bash
cd terraform

# Choose environment (dev, staging, prod)
terraform init \
  -backend-config="key=kubernetes/${ENVIRONMENT}/terraform.tfstate"

terraform plan -var-file="environments/${ENVIRONMENT}/terraform.tfvars"
terraform apply -var-file="environments/${ENVIRONMENT}/terraform.tfvars"
```

### 2. Configure kubectl Access

```bash
aws eks update-kubeconfig \
  --region us-east-1 \
  --name mini-kubernet-${ENVIRONMENT}
```

### 3. Verify Cluster

```bash
kubectl cluster-info
kubectl get nodes
kubectl get namespaces
```

## Kubernetes Cluster Setup

### 1. Install Required Add-ons

```bash
# Install Istio service mesh (optional)
./scripts/install-istio.sh

# Install cert-manager for TLS
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Install ArgoCD for GitOps
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### 2. Create Namespaces and Security Policies

```bash
kubectl apply -f k8s/security/prod-namespace.yaml
kubectl apply -f k8s/security/staging-namespace.yaml
kubectl apply -f k8s/security/dev-namespace.yaml
```

### 3. Deploy Monitoring Stack

```bash
# Create monitoring namespace
kubectl create namespace monitoring

# Deploy Prometheus
kubectl apply -f k8s/monitoring/prometheus.yaml

# Deploy Grafana
helm repo add grafana https://grafana.github.io/helm-charts
helm install grafana grafana/grafana -n monitoring

# Deploy Loki for logging
kubectl apply -f k8s/logging/loki.yaml
kubectl apply -f k8s/logging/promtail.yaml
```

### 4. Configure Secrets Management

```bash
# Deploy Vault
kubectl create namespace vault
kubectl apply -f k8s/vault/vault.yaml

# Initialize Vault
kubectl exec -n vault vault-0 -- vault operator init -key-shares=5 -key-threshold=3

# Unseal Vault (requires 3 of 5 keys)
kubectl exec -n vault vault-0 -- vault operator unseal <key1>
kubectl exec -n vault vault-0 -- vault operator unseal <key2>
kubectl exec -n vault vault-0 -- vault operator unseal <key3>

# Configure Kubernetes authentication
kubectl exec -n vault vault-0 -- vault auth enable kubernetes
```

## Application Deployment

### 1. Build and Push Images

```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REGISTRY

# Build images
docker-compose -f docker-compose.yml build

# Tag and push
for service in auth-service api-gateway build-service deploy-service frontend; do
  docker tag mini-kubernet_${service}:latest ${ECR_REGISTRY}/${service}:${ENVIRONMENT}
  docker push ${ECR_REGISTRY}/${service}:${ENVIRONMENT}
done
```

### 2. Deploy with Helm

```bash
# Option A: Using deployment script
./scripts/deployment/deploy.sh ${ENVIRONMENT}

# Option B: Manual Helm deployment
helm upgrade --install auth-service helm/charts/auth-service \
  --namespace kubernet-${ENVIRONMENT} \
  --values helm/charts/auth-service/values.yaml \
  --values helm/charts/auth-service/values-${ENVIRONMENT}.yaml
```

### 3. Deploy with Kustomize

```bash
kubectl apply -k k8s/overlays/${ENVIRONMENT}/
```

### 4. Configure ArgoCD (Production)

```bash
# Apply ArgoCD project and applications
kubectl apply -f argocd/projects/mini-kubernet.yaml
kubectl apply -f argocd/applications/${ENVIRONMENT}-app.yaml

# Port-forward to ArgoCD UI
kubectl port-forward -n argocd svc/argocd-server 8080:443

# Get default password
kubectl get secret -n argocd argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## Post-Deployment Validation

### 1. Check Service Health

```bash
# Check all pods are running
kubectl get pods -n kubernet-${ENVIRONMENT}

# Check service endpoints
kubectl get svc -n kubernet-${ENVIRONMENT}

# Check ingress
kubectl get ingress -n kubernet-${ENVIRONMENT}
```

### 2. Verify Connectivity

```bash
# Test auth service
curl -k https://auth-${ENVIRONMENT}.mini-kubernet.com/health

# Test API gateway
curl -k https://api-${ENVIRONMENT}.mini-kubernet.com/health

# Test frontend
curl -k https://${ENVIRONMENT}.mini-kubernet.com
```

### 3. Run E2E Tests

```bash
./tests/e2e/run-e2e-tests.sh ${ENVIRONMENT}
```

### 4. Check Metrics and Logs

```bash
# Access Grafana
kubectl port-forward -n monitoring svc/grafana 3000:80

# View logs via Loki
kubectl port-forward -n logging svc/loki 3100:3100

# Access Prometheus
kubectl port-forward -n monitoring svc/prometheus 9090:9090
```

## Operations & Maintenance

### Scaling Services

```bash
# Manual scaling
kubectl scale deployment auth-service -n kubernet-${ENVIRONMENT} --replicas=10

# Enable autoscaling
kubectl autoscale deployment auth-service -n kubernet-${ENVIRONMENT} --min=3 --max=20
```

### Rolling Updates

```bash
# Update image
kubectl set image deployment/auth-service \
  auth-service=${ECR_REGISTRY}/auth-service:new-tag \
  -n kubernet-${ENVIRONMENT}

# Watch rollout
kubectl rollout status deployment/auth-service -n kubernet-${ENVIRONMENT}
```

### Database Migrations

```bash
# Connect to database
kubectl run -it --rm psql-client --image=postgres --restart=Never \
  -n kubernet-${ENVIRONMENT} -- \
  psql -h postgres.kubernet-${ENVIRONMENT}.svc.cluster.local -U postgres

# Run migrations
./backend/scripts/migrate.sh ${ENVIRONMENT}
```

### Backup and Recovery

```bash
# Backup database
aws rds create-db-snapshot \
  --db-instance-identifier mini-kubernet-${ENVIRONMENT} \
  --db-snapshot-identifier mini-kubernet-${ENVIRONMENT}-$(date +%Y%m%d-%H%M%S)

# Backup persistent volumes
kubectl get pvc -n kubernet-${ENVIRONMENT} -o json | \
  kubectl exec -i -n kubernet-${ENVIRONMENT} - -- \
  tar czf /backup/pvc-backup-$(date +%Y%m%d-%H%M%S).tar.gz
```

### Disaster Recovery

```bash
# Restore from database snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier mini-kubernet-${ENVIRONMENT}-restored \
  --db-snapshot-identifier mini-kubernet-${ENVIRONMENT}-backup-id

# Restore persistent volumes
kubectl create configmap restore-backup \
  --from-file=/backup/pvc-backup.tar.gz
```

### Troubleshooting

```bash
# View pod logs
kubectl logs -n kubernet-${ENVIRONMENT} -l app=auth-service --tail=100

# Debug pod issues
kubectl describe pod <pod-name> -n kubernet-${ENVIRONMENT}

# Check events
kubectl get events -n kubernet-${ENVIRONMENT} --sort-by='.lastTimestamp'

# Port-forward for debugging
kubectl port-forward -n kubernet-${ENVIRONMENT} pod/auth-service-xxx 8081:8081
```

## Monitoring & Alerting

### Key Metrics to Monitor

- Pod CPU and memory usage
- Request latency (p50, p95, p99)
- Error rates by service
- Database connection count
- Redis memory usage
- Ingress controller latency
- Certificate expiration dates

### Alert Thresholds

Configure in `k8s/monitoring/alert-rules.yaml`:
- High error rate: > 5%
- High latency: p95 > 500ms
- Pod restarts: > 5 in 1 hour
- Disk usage: > 85%
- Database connections: > 80% of max

## Support & Documentation

- Architecture documentation: `docs/ARCHITECTURE.md`
- Runbooks: `docs/runbooks/`
- API documentation: `API.md`
- Troubleshooting guide: `docs/TROUBLESHOOTING.md`
