#!/bin/bash

# Initialize all Helm charts for other services
set -e

SERVICES=("api-gateway" "build-service" "deploy-service" "frontend" "postgres" "redis")

for service in "${SERVICES[@]}"; do
    CHART_DIR="helm/charts/$service"
    
    if [ ! -d "$CHART_DIR" ]; then
        echo "Creating Helm chart structure for $service..."
        mkdir -p "$CHART_DIR/templates"
        
        # Create Chart.yaml
        cat > "$CHART_DIR/Chart.yaml" <<EOF
apiVersion: v2
name: $service
description: A Helm chart for Mini Kubernet $service
type: application
version: 1.0.0
appVersion: "1.0.0"
EOF
        
        # Create values.yaml with basic structure
        cat > "$CHART_DIR/values.yaml" <<EOF
replicaCount: 3

image:
  repository: "$service"
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 8080

resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 250m
    memory: 256Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
EOF
        
        # Create values-dev.yaml
        cat > "$CHART_DIR/values-dev.yaml" <<EOF
replicaCount: 1

resources:
  limits:
    cpu: 250m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 128Mi

autoscaling:
  enabled: false
EOF
        
        # Create values-staging.yaml
        cat > "$CHART_DIR/values-staging.yaml" <<EOF
replicaCount: 2

resources:
  limits:
    cpu: 400m
    memory: 384Mi
  requests:
    cpu: 200m
    memory: 192Mi

autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
EOF
        
        # Create values-prod.yaml
        cat > "$CHART_DIR/values-prod.yaml" <<EOF
replicaCount: 5

resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi

autoscaling:
  enabled: true
  minReplicas: 5
  maxReplicas: 20
EOF
        
        # Create basic templates
        cat > "$CHART_DIR/templates/deployment.yaml" <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Chart.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ .Chart.Name }}
  template:
    metadata:
      labels:
        app: {{ .Chart.Name }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - containerPort: 8080
        resources:
          {{- toYaml .Values.resources | nindent 12 }}
EOF
        
        echo "✓ Created Helm chart for $service"
    else
        echo "✓ Helm chart already exists for $service"
    fi
done

echo "All Helm charts initialized!"
