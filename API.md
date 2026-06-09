# API Documentation

## Base URL
```
http://api.example.com/v1
```

## Authentication

All requests require JWT token in header:
```
Authorization: Bearer <token>
```

### Login Endpoint

```
POST /auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}

Response 200:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2024-01-01T00:00:00Z"
}
```

---

## Build Service API

### Create Build

```
POST /build/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "service_name": "api-gateway",
  "git_repo": "https://github.com/org/repo.git",
  "git_branch": "main",
  "dockerfile": "Dockerfile",
  "registry": "gcr.io/my-project"
}

Response 202:
{
  "build_id": "build-api-gateway-1704067200",
  "status": "pending",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get Build Status

```
GET /build/{build_id}
Authorization: Bearer <token>

Response 200:
{
  "build_id": "build-api-gateway-1704067200",
  "status": "completed",
  "progress": 100,
  "log": "Building Docker image...\nImage built successfully",
  "completed_at": "2024-01-01T00:05:00Z"
}
```

### Cancel Build

```
DELETE /build/{build_id}
Authorization: Bearer <token>

Response 200:
{
  "message": "build cancelled",
  "build_id": "build-api-gateway-1704067200"
}
```

---

## Deploy Service API

### Create Deployment

```
POST /deploy/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "service_name": "api-gateway",
  "image": "gcr.io/my-project/api-gateway:v1.0.0",
  "replicas": 3,
  "namespace": "kubernet-prod",
  "environment": {
    "PORT": "8080",
    "LOG_LEVEL": "info"
  }
}

Response 202:
{
  "deployment_id": "deploy-api-gateway-1704067200",
  "service_name": "api-gateway",
  "status": "creating",
  "ready_replicas": 0,
  "desired_replicas": 3,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Get Deployment Status

```
GET /deploy/{deployment_id}
Authorization: Bearer <token>

Response 200:
{
  "deployment_id": "deploy-api-gateway-1704067200",
  "service_name": "api-gateway",
  "status": "ready",
  "ready_replicas": 3,
  "desired_replicas": 3,
  "created_at": "2024-01-01T00:00:00Z"
}
```

### Rollback Deployment

```
PATCH /deploy/{deployment_id}/rollback
Authorization: Bearer <token>

Response 200:
{
  "message": "deployment rolled back",
  "deployment_id": "deploy-api-gateway-1704067200"
}
```

### Delete Deployment

```
DELETE /deploy/{deployment_id}
Authorization: Bearer <token>

Response 200:
{
  "message": "deployment deleted",
  "deployment_id": "deploy-api-gateway-1704067200"
}
```

---

## Error Handling

### Error Response Format

```json
{
  "error": "error message",
  "code": "ERROR_CODE",
  "details": "additional details"
}
```

### Common Error Codes

| Code | Status | Description |
|------|--------|-------------|
| INVALID_TOKEN | 401 | Invalid or expired token |
| UNAUTHORIZED | 403 | User not authorized |
| NOT_FOUND | 404 | Resource not found |
| CONFLICT | 409 | Resource already exists |
| INTERNAL_ERROR | 500 | Internal server error |

---

## Rate Limiting

All endpoints are rate limited to:
- **Authenticated users**: 1000 requests/minute
- **Anonymous**: 100 requests/minute

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1704067260
```

---

## Examples

### Complete Workflow

```bash
# 1. Login
TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}' \
  | jq -r .token)

# 2. Create build
BUILD=$(curl -X POST http://localhost:8080/build/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name":"api-gateway",
    "git_repo":"https://github.com/org/repo.git",
    "git_branch":"main",
    "registry":"gcr.io/my-project"
  }' | jq -r .build_id)

# 3. Monitor build
curl http://localhost:8080/build/$BUILD \
  -H "Authorization: Bearer $TOKEN" | jq .status

# 4. Deploy
DEPLOY=$(curl -X POST http://localhost:8080/deploy/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "service_name":"api-gateway",
    "image":"gcr.io/my-project/api-gateway:latest",
    "replicas":3
  }' | jq -r .deployment_id)

# 5. Check deployment status
curl http://localhost:8080/deploy/$DEPLOY \
  -H "Authorization: Bearer $TOKEN" | jq .status
```

---

## Webhooks (Future)

```
POST /webhooks/github
POST /webhooks/docker-registry
POST /webhooks/k8s-events
```

---

## SDK Usage

### Go

```go
import "github.com/mini-kubernet/sdk-go"

client := kubernet.NewClient("http://localhost:8080", token)

// Build
buildReq := &kubernet.BuildRequest{
    ServiceName: "api-gateway",
    GitRepo:    "https://github.com/org/repo.git",
    Registry:   "gcr.io/my-project",
}
build, err := client.Build.Create(buildReq)

// Deploy
deployReq := &kubernet.DeployRequest{
    ServiceName: "api-gateway",
    Image:       "gcr.io/my-project/api-gateway:latest",
    Replicas:    3,
}
deploy, err := client.Deploy.Create(deployReq)
```

### Python

```python
from kubernet import Client

client = Client("http://localhost:8080", token)

# Build
build = client.build.create({
    "service_name": "api-gateway",
    "git_repo": "https://github.com/org/repo.git",
    "registry": "gcr.io/my-project"
})

# Deploy
deploy = client.deploy.create({
    "service_name": "api-gateway",
    "image": "gcr.io/my-project/api-gateway:latest",
    "replicas": 3
})
```

---

## Versioning

API version is in the URL path: `/v1/`, `/v2/`, etc.

Current version: **v1**

Breaking changes will introduce new versions rather than breaking existing ones.
