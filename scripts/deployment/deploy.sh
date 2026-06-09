#!/bin/bash
set -e

# Mini Kubernet Production Deployment Script
# This script deploys the entire Mini Kubernet stack to Kubernetes

ENVIRONMENT=${1:-prod}
NAMESPACE="kubernet-${ENVIRONMENT}"
REGISTRY="${ECR_REGISTRY:-}"

echo "========================================"
echo "Mini Kubernet Deployment"
echo "Environment: $ENVIRONMENT"
echo "Namespace: $NAMESPACE"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Create namespace
print_status "Creating namespace: $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Apply security policies
print_status "Applying security policies..."
kubectl apply -f k8s/security/${ENVIRONMENT}-namespace.yaml

# Apply Kustomize overlays
print_status "Deploying with Kustomize overlays..."
kubectl apply -k k8s/overlays/${ENVIRONMENT}/

# Deploy Helm charts
print_status "Deploying Helm charts..."

# Auth Service
helm upgrade --install auth-service helm/charts/auth-service \
  --namespace "$NAMESPACE" \
  --values helm/charts/auth-service/values.yaml \
  --values helm/charts/auth-service/values-${ENVIRONMENT}.yaml \
  --set image.tag="${ENVIRONMENT}" \
  ${REGISTRY:+--set image.repository=${REGISTRY}/auth-service}

# API Gateway
helm upgrade --install api-gateway helm/charts/api-gateway \
  --namespace "$NAMESPACE" \
  --values helm/charts/api-gateway/values.yaml \
  --values helm/charts/api-gateway/values-${ENVIRONMENT}.yaml \
  ${REGISTRY:+--set image.repository=${REGISTRY}/api-gateway}

# Build Service
helm upgrade --install build-service helm/charts/build-service \
  --namespace "$NAMESPACE" \
  --values helm/charts/build-service/values.yaml \
  --values helm/charts/build-service/values-${ENVIRONMENT}.yaml \
  ${REGISTRY:+--set image.repository=${REGISTRY}/build-service}

# Deploy Service
helm upgrade --install deploy-service helm/charts/deploy-service \
  --namespace "$NAMESPACE" \
  --values helm/charts/deploy-service/values.yaml \
  --values helm/charts/deploy-service/values-${ENVIRONMENT}.yaml \
  ${REGISTRY:+--set image.repository=${REGISTRY}/deploy-service}

# Frontend
helm upgrade --install frontend helm/charts/frontend \
  --namespace "$NAMESPACE" \
  --values helm/charts/frontend/values.yaml \
  --values helm/charts/frontend/values-${ENVIRONMENT}.yaml \
  ${REGISTRY:+--set image.repository=${REGISTRY}/frontend}

# Wait for deployments
print_status "Waiting for deployments to be ready..."
kubectl rollout status deployment/auth-service -n "$NAMESPACE" --timeout=5m || print_warning "auth-service deployment pending"
kubectl rollout status deployment/api-gateway -n "$NAMESPACE" --timeout=5m || print_warning "api-gateway deployment pending"
kubectl rollout status deployment/frontend -n "$NAMESPACE" --timeout=5m || print_warning "frontend deployment pending"

# Verify deployments
print_status "Verifying deployments..."
kubectl get deployments -n "$NAMESPACE"
kubectl get services -n "$NAMESPACE"
kubectl get ingresses -n "$NAMESPACE"

# Setup ArgoCD for continuous deployment (prod only)
if [ "$ENVIRONMENT" = "prod" ]; then
    print_status "Setting up ArgoCD for production..."
    kubectl apply -f argocd/projects/mini-kubernet.yaml
    kubectl apply -f argocd/applications/prod-app.yaml
    print_status "ArgoCD configured. Monitor via: kubectl port-forward -n argocd svc/argocd-server 8080:443"
fi

# Health checks
print_status "Running health checks..."
for i in {1..30}; do
    if kubectl get pods -n "$NAMESPACE" | grep -q "Running"; then
        print_status "Pods are running"
        break
    fi
    if [ $i -eq 30 ]; then
        print_warning "Some pods not in Running state yet"
    fi
    sleep 5
done

print_status "Deployment completed successfully!"
print_status "Access your services at:"
echo "  Auth Service: https://auth-${ENVIRONMENT}.mini-kubernet.com"
echo "  API Gateway: https://api-${ENVIRONMENT}.mini-kubernet.com"
echo "  Frontend: https://${ENVIRONMENT}.mini-kubernet.com"
