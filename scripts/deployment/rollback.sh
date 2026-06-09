#!/bin/bash

# Rollback script for Mini Kubernet
# Rolls back to the previous version of a deployment

set -e

ENVIRONMENT=${1:-prod}
SERVICE=${2:-all}
NAMESPACE="kubernet-${ENVIRONMENT}"

print_info() {
    echo "[INFO] $1"
}

print_error() {
    echo "[ERROR] $1"
    exit 1
}

if [ "$SERVICE" = "all" ]; then
    SERVICES=("auth-service" "api-gateway" "build-service" "deploy-service" "frontend")
else
    SERVICES=("$SERVICE")
fi

for svc in "${SERVICES[@]}"; do
    print_info "Rolling back $svc..."
    
    # Get current revision
    CURRENT_REV=$(kubectl rollout history deployment/$svc -n "$NAMESPACE" | tail -1 | awk '{print $1}')
    PREVIOUS_REV=$((CURRENT_REV - 1))
    
    if [ $PREVIOUS_REV -lt 1 ]; then
        print_error "No previous revision available for $svc"
    fi
    
    # Perform rollback
    kubectl rollout undo deployment/$svc -n "$NAMESPACE" --to-revision=$PREVIOUS_REV
    
    # Wait for rollback
    kubectl rollout status deployment/$svc -n "$NAMESPACE" --timeout=5m
    
    print_info "✓ $svc rolled back to revision $PREVIOUS_REV"
done

print_info "Rollback completed successfully!"
