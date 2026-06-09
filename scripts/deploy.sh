#!/bin/bash

# Advanced Kubernetes Deployment Script
# Deploy all services to Kubernetes

set -e

NAMESPACE="kubernet-prod"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-gcr.io/my-project}"
DOCKER_TAG="${DOCKER_TAG:-latest}"

echo "=========================================="
echo "Deploying Mini Kubernetes Project"
echo "=========================================="

# Create namespace
echo "Creating namespace..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Create secrets
echo "Creating secrets..."
kubectl create secret generic db-secret \
  --from-literal=url=${DATABASE_URL} \
  -n $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic jwt-secret \
  --from-literal=key=${JWT_SECRET_KEY} \
  -n $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Deploy databases
echo "Deploying PostgreSQL..."
kubectl apply -f kubernetes/postgres-statefulset.yaml

echo "Deploying Redis..."
kubectl apply -f kubernetes/redis-deployment.yaml

# Wait for databases
echo "Waiting for databases to be ready..."
kubectl wait --for=condition=ready pod \
  -l app=postgres -n $NAMESPACE --timeout=300s

kubectl wait --for=condition=ready pod \
  -l app=redis -n $NAMESPACE --timeout=300s

# Deploy services
echo "Deploying services..."
kubectl apply -f kubernetes/services/

# Wait for services
echo "Waiting for services to be ready..."
kubectl wait --for=condition=ready pod \
  -l app=api-gateway -n $NAMESPACE --timeout=300s

# Install Istio
echo "Installing Istio service mesh..."
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

helm upgrade --install istio-base istio/base \
  -n istio-system --create-namespace

helm upgrade --install istiod istio/istiod \
  -n istio-system

# Deploy Istio configs
echo "Deploying Istio configurations..."
kubectl apply -f kubernetes/istio/

# Deploy HPA
echo "Deploying Horizontal Pod Autoscalers..."
kubectl apply -f kubernetes/hpa/

# Install monitoring
echo "Installing Prometheus..."
kubectl apply -f kubernetes/monitoring/prometheus.yaml

echo "Installing Grafana..."
kubectl apply -f kubernetes/monitoring/grafana.yaml

# Verify deployment
echo "=========================================="
echo "Deployment Summary"
echo "=========================================="
kubectl get namespaces
kubectl get all -n $NAMESPACE
kubectl get virtualservices -n $NAMESPACE
kubectl get hpa -n $NAMESPACE

echo ""
echo "Deployment completed successfully!"
echo ""
echo "Next steps:"
echo "1. Configure kubectl: kubectl config use-context <cluster-context>"
echo "2. Port forward Grafana: kubectl port-forward svc/grafana 3000:3000 -n $NAMESPACE"
echo "3. Access Grafana: http://localhost:3000 (admin/admin)"
