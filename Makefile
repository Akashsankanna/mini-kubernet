.PHONY: help build deploy test clean docker-build docker-push

help:
	@echo "Advanced Kubernetes Project - Make Commands"
	@echo "=========================================="
	@echo "make build                - Build all services"
	@echo "make test                 - Run all tests"
	@echo "make docker-build         - Build Docker images"
	@echo "make docker-push          - Push Docker images to registry"
	@echo "make deploy-infra         - Deploy infrastructure (Terraform)"
	@echo "make deploy-k8s           - Deploy to Kubernetes"
	@echo "make deploy-canary        - Deploy canary version"
	@echo "make deploy-multi-cluster - Deploy to multiple clusters"
	@echo "make install-istio        - Install Istio service mesh"
	@echo "make install-prometheus   - Install Prometheus monitoring"
	@echo "make clean                - Clean build artifacts"

DOCKER_REGISTRY ?= gcr.io/my-project
DOCKER_TAG ?= latest
CLUSTERS ?= us-east-1 us-west-2 eu-west-1

# Build targets
build:
	@echo "Building backend services..."
	cd backend/auth-service && go build -o ../../bin/auth-service ./cmd/main.go
	cd backend/build-service && go build -o ../../bin/build-service ./cmd/main.go
	cd backend/deploy-service && go build -o ../../bin/deploy-service ./cmd/main.go
	cd backend/api-gateway && go build -o ../../bin/api-gateway ./cmd/main.go
	@echo "Building frontend..."
	cd frontend && npm install && npm run build
	@echo "Build complete!"

test:
	@echo "Running tests..."
	cd backend/auth-service && go test ./...
	cd backend/build-service && go test ./...
	cd backend/deploy-service && go test ./...
	cd backend/api-gateway && go test ./...

docker-build:
	@echo "Building Docker images..."
	docker build -t $(DOCKER_REGISTRY)/auth-service:$(DOCKER_TAG) -f backend/auth-service/Dockerfile backend/auth-service
	docker build -t $(DOCKER_REGISTRY)/build-service:$(DOCKER_TAG) -f backend/build-service/Dockerfile backend/build-service
	docker build -t $(DOCKER_REGISTRY)/deploy-service:$(DOCKER_TAG) -f backend/deploy-service/Dockerfile backend/deploy-service
	docker build -t $(DOCKER_REGISTRY)/api-gateway:$(DOCKER_TAG) -f backend/api-gateway/Dockerfile backend/api-gateway
	docker build -t $(DOCKER_REGISTRY)/frontend:$(DOCKER_TAG) -f frontend/Dockerfile frontend

docker-push:
	@echo "Pushing Docker images..."
	docker push $(DOCKER_REGISTRY)/auth-service:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/build-service:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/deploy-service:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/api-gateway:$(DOCKER_TAG)
	docker push $(DOCKER_REGISTRY)/frontend:$(DOCKER_TAG)

deploy-infra:
	@echo "Deploying infrastructure..."
	cd terraform && terraform init && terraform plan && terraform apply -auto-approve

deploy-k8s:
	@echo "Deploying to Kubernetes..."
	kubectl apply -f kubernetes/namespace.yaml
	kubectl apply -f kubernetes/postgres-statefulset.yaml
	kubectl apply -f kubernetes/redis-deployment.yaml
	kubectl apply -f kubernetes/services/

deploy-canary:
	@echo "Deploying canary version (10% traffic)..."
	kubectl apply -f kubernetes/canary/
	kubectl patch virtualservice api-gateway -p '{"spec":{"hosts":[{"name":"api-gateway","weight":90},{"name":"api-gateway-canary","weight":10}]}}'

deploy-multi-cluster:
	@echo "Deploying to multiple clusters: $(CLUSTERS)"
	@for cluster in $(CLUSTERS); do \
		echo "Deploying to $$cluster"; \
		kubectl config use-context $$cluster; \
		make deploy-k8s; \
	done

install-istio:
	@echo "Installing Istio..."
	helm repo add istio https://istio-release.storage.googleapis.com/charts
	helm repo update
	helm install istio-base istio/base -n istio-system --create-namespace
	helm install istiod istio/istiod -n istio-system

install-prometheus:
	@echo "Installing Prometheus & Grafana..."
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update
	helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring --create-namespace
	helm install grafana grafana/grafana -n monitoring

clean:
	@echo "Cleaning up..."
	rm -rf bin/
	rm -rf frontend/build
	rm -rf frontend/node_modules
	cd terraform && terraform destroy -auto-approve
	@echo "Clean complete!"
