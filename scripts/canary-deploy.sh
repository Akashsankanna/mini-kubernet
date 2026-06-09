#!/bin/bash

# Canary Deployment Script
# Gradually roll out new version

set -e

NAMESPACE="kubernet-prod"
SERVICE="api-gateway"

echo "=========================================="
echo "Canary Deployment Script"
echo "=========================================="

# Start with 10% traffic to new version
echo "Starting canary deployment with 10% traffic..."
kubectl apply -f kubernetes/canary/api-gateway-canary.yaml

echo "Applying 10% traffic to canary version..."
kubectl patch virtualservice $SERVICE \
  -n $NAMESPACE --type merge \
  -p '{"spec":{"http":[{"route":[{"destination":{"host":"'$SERVICE'","subset":"v1"},"weight":90},{"destination":{"host":"'$SERVICE'","subset":"v2"},"weight":10}]}]}}'

echo "Waiting for 5 minutes..."
sleep 300

# Check metrics
echo "Checking canary metrics..."
ERROR_RATE=$(kubectl logs -n $NAMESPACE -l app=$SERVICE,version=v2 --tail=100 | grep -c "error" || true)

if [ $ERROR_RATE -gt 10 ]; then
  echo "High error rate detected! Rolling back canary deployment..."
  kubectl delete deployment $SERVICE-canary -n $NAMESPACE
  exit 1
fi

# Increase to 25%
echo "Increasing traffic to 25%..."
kubectl patch virtualservice $SERVICE \
  -n $NAMESPACE --type merge \
  -p '{"spec":{"http":[{"route":[{"destination":{"host":"'$SERVICE'","subset":"v1"},"weight":75},{"destination":{"host":"'$SERVICE'","subset":"v2"},"weight":25}]}]}}'

sleep 300

# Increase to 50%
echo "Increasing traffic to 50%..."
kubectl patch virtualservice $SERVICE \
  -n $NAMESPACE --type merge \
  -p '{"spec":{"http":[{"route":[{"destination":{"host":"'$SERVICE'","subset":"v1"},"weight":50},{"destination":{"host":"'$SERVICE'","subset":"v2"},"weight":50}]}]}}'

sleep 300

# Full traffic to new version
echo "Rolling out fully to new version (100%)..."
kubectl patch virtualservice $SERVICE \
  -n $NAMESPACE --type merge \
  -p '{"spec":{"http":[{"route":[{"destination":{"host":"'$SERVICE'","subset":"v2"},"weight":100}]}]}}'

# Update stable version
kubectl set image deployment/$SERVICE \
  $SERVICE=gcr.io/my-project/$SERVICE:canary \
  -n $NAMESPACE

# Remove canary deployment
kubectl delete deployment $SERVICE-canary -n $NAMESPACE

echo "=========================================="
echo "Canary deployment completed successfully!"
echo "=========================================="
