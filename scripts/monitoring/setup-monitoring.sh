#!/bin/bash

# Monitoring setup script
# Configures monitoring dashboards and alert rules

set -e

ENVIRONMENT=${1:-prod}
NAMESPACE="kubernet-${ENVIRONMENT}"

print_info() {
    echo "[INFO] $1"
}

# Configure Prometheus scrape targets
print_info "Configuring Prometheus scrape targets..."
kubectl apply -f k8s/monitoring/prometheus-config.yaml -n monitoring

# Create Grafana dashboards
print_info "Creating Grafana dashboards..."
kubectl exec -n monitoring grafana-0 -- grafana-cli dashboard import kubernetes-cluster

# Configure alert rules
print_info "Configuring alert rules..."
kubectl apply -f k8s/monitoring/alert-rules.yaml -n monitoring

# Setup AlertManager
print_info "Setting up AlertManager..."
kubectl apply -f k8s/monitoring/alertmanager-config.yaml -n monitoring

# Configure service monitors
print_info "Configuring ServiceMonitors..."
kubectl apply -f k8s/monitoring/servicemonitors.yaml -n monitoring

# Verify monitoring
print_info "Verifying monitoring setup..."
kubectl get servicemonitors -A
kubectl get prometheusrule -A
kubectl get alertmanagerrconfig -A

print_info "Monitoring setup completed!"
print_info "Access Grafana: kubectl port-forward -n monitoring svc/grafana 3000:80"
print_info "Access Prometheus: kubectl port-forward -n monitoring svc/prometheus 9090:9090"
