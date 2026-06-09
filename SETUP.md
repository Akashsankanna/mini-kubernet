## Setup Instructions

### Prerequisites
- Docker 24+
- Kubernetes 1.28+
- kubectl 1.28+
- Helm 3.12+
- Terraform 1.5+
- Go 1.21+
- Node.js 18+

### Local Development

1. **Clone and setup environment:**
```bash
cp .env.example .env
# Edit .env with your values
```

2. **Start local environment:**
```bash
docker-compose up -d
```

3. **Verify services are running:**
```bash
docker-compose ps
```

### Infrastructure Deployment (AWS)

1. **Configure AWS credentials:**
```bash
aws configure
# Enter your AWS credentials
```

2. **Initialize Terraform:**
```bash
cd terraform
terraform init
```

3. **Deploy infrastructure:**
```bash
terraform plan
terraform apply
```

4. **Configure kubectl:**
```bash
aws eks update-kubeconfig --region us-east-1 --name mini-kubernet-prod
```

### Kubernetes Deployment

1. **Build and push Docker images:**
```bash
make docker-build
DOCKER_REGISTRY=gcr.io/your-project make docker-push
```

2. **Deploy to Kubernetes:**
```bash
bash scripts/deploy.sh
```

3. **Install Istio:**
```bash
bash scripts/install-istio.sh
```

### Canary Deployment

```bash
bash scripts/canary-deploy.sh
```

### Multi-Cluster Deployment

```bash
bash scripts/multi-cluster-deploy.sh
```

### Monitoring

1. **Access Prometheus:**
```bash
kubectl port-forward svc/prometheus 9090:9090 -n kubernet-prod
# Visit http://localhost:9090
```

2. **Access Grafana:**
```bash
kubectl port-forward svc/grafana 3000:3000 -n kubernet-prod
# Visit http://localhost:3000 (admin/admin)
```

### Cleanup

```bash
# Delete Kubernetes deployments
kubectl delete namespace kubernet-prod

# Destroy infrastructure
cd terraform && terraform destroy
```

## Troubleshooting

### Check pod status:
```bash
kubectl get pods -n kubernet-prod
kubectl describe pod <pod-name> -n kubernet-prod
kubectl logs <pod-name> -n kubernet-prod
```

### Check service connectivity:
```bash
kubectl exec -it <pod-name> -n kubernet-prod -- sh
curl http://service-name:port/health
```

### View Istio configuration:
```bash
kubectl get virtualservices -n kubernet-prod
kubectl get destinationrules -n kubernet-prod
istioctl analyze
```

### Monitor resource usage:
```bash
kubectl top nodes
kubectl top pods -n kubernet-prod
```
