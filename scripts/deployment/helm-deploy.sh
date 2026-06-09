#!/bin/bash
set -e

# Helm deployment wrapper script
# This script provides a simpler interface for deploying with Helm

ENVIRONMENT=${1:-prod}
SERVICE=${2:-all}
ACTION=${3:-upgrade}

NAMESPACE="kubernet-${ENVIRONMENT}"

print_info() {
    echo "[INFO] $1"
}

print_error() {
    echo "[ERROR] $1"
    exit 1
}

if [ "$ACTION" != "upgrade" ] && [ "$ACTION" != "install" ] && [ "$ACTION" != "dry-run" ]; then
    print_error "Action must be: upgrade, install, or dry-run"
fi

helm_deploy() {
    local SERVICE=$1
    local VALUES_BASE="helm/charts/${SERVICE}/values.yaml"
    local VALUES_ENV="helm/charts/${SERVICE}/values-${ENVIRONMENT}.yaml"
    
    if [ ! -f "$VALUES_BASE" ]; then
        print_error "Chart not found: $VALUES_BASE"
    fi
    
    print_info "Deploying $SERVICE to $ENVIRONMENT..."
    
    HELM_CMD="helm $ACTION $SERVICE helm/charts/$SERVICE --namespace $NAMESPACE"
    HELM_CMD="$HELM_CMD --values $VALUES_BASE"
    
    if [ -f "$VALUES_ENV" ]; then
        HELM_CMD="$HELM_CMD --values $VALUES_ENV"
    fi
    
    if [ "$ACTION" = "dry-run" ]; then
        HELM_CMD="$HELM_CMD --dry-run --debug"
    fi
    
    eval $HELM_CMD
}

case $SERVICE in
    all)
        for svc in auth-service api-gateway build-service deploy-service frontend postgres redis; do
            if [ -d "helm/charts/$svc" ]; then
                helm_deploy "$svc"
            fi
        done
        ;;
    *)
        helm_deploy "$SERVICE"
        ;;
esac

print_info "Helm deployment completed!"
