#!/bin/bash

# Multi-Cluster Deployment Script

set -e

CLUSTERS=("us-east-1" "us-west-2" "eu-west-1")
DOCKER_REGISTRY="${DOCKER_REGISTRY:-gcr.io/my-project}"

echo "=========================================="
echo "Multi-Cluster Deployment"
echo "=========================================="

for cluster in "${CLUSTERS[@]}"; do
  echo ""
  echo "Deploying to cluster: $cluster"
  echo "=========================================="
  
  # Switch context
  kubectl config use-context $cluster
  
  # Run deployment
  bash scripts/deploy.sh
  
  # Verify
  echo "Verifying deployment in $cluster..."
  kubectl get pods -n kubernet-prod
  
  echo "Deployment to $cluster completed!"
done

echo ""
echo "=========================================="
echo "Multi-cluster deployment completed!"
echo "=========================================="
