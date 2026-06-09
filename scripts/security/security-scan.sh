#!/bin/bash

# Security scanning script
# Scans containers and code for vulnerabilities

set -e

ENVIRONMENT=${1:-prod}

print_info() {
    echo "[INFO] $1"
}

print_error() {
    echo "[ERROR] $1"
}

VULNERABILITIES_FOUND=0

# Scan running images
print_info "Scanning container images for vulnerabilities..."
for pod in $(kubectl get pods -n kubernet-${ENVIRONMENT} -o jsonpath='{.items[*].metadata.name}'); do
    IMAGE=$(kubectl get pod $pod -n kubernet-${ENVIRONMENT} -o jsonpath='{.spec.containers[0].image}')
    print_info "Scanning $IMAGE..."
    
    if ! trivy image --exit-code 0 --severity CRITICAL "$IMAGE"; then
        VULNERABILITIES_FOUND=$((VULNERABILITIES_FOUND + 1))
    fi
done

# Scan filesystem
print_info "Scanning Kubernetes manifests..."
trivy config k8s/

# Scan Terraform
print_info "Scanning Terraform files..."
trivy config terraform/

if [ $VULNERABILITIES_FOUND -gt 0 ]; then
    print_error "Found $VULNERABILITIES_FOUND images with CRITICAL vulnerabilities!"
    exit 1
else
    print_info "✓ Security scan passed!"
    exit 0
fi
