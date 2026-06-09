#!/bin/bash

# Install Istio Service Mesh

set -e

ISTIO_VERSION="1.17.2"

echo "=========================================="
echo "Installing Istio Service Mesh"
echo "=========================================="

# Add Istio repo
helm repo add istio https://istio-release.storage.googleapis.com/charts
helm repo update

# Create namespace
kubectl create namespace istio-system --dry-run=client -o yaml | kubectl apply -f -

# Install base
echo "Installing Istio base..."
helm install istio-base istio/base -n istio-system

# Install istiod
echo "Installing Istiod..."
helm install istiod istio/istiod -n istio-system

# Create ingressgateway namespace and inject sidecar
kubectl create namespace kubernet-prod --dry-run=client -o yaml | kubectl apply -f -
kubectl label namespace kubernet-prod istio-injection=enabled --overwrite

# Install ingress gateway
echo "Installing Istio Ingress Gateway..."
kubectl apply -f kubernetes/istio/

echo "=========================================="
echo "Istio installation completed!"
echo "=========================================="
echo ""
echo "Verify installation:"
echo "kubectl get pods -n istio-system"
echo "kubectl get pods -n kubernet-prod"
