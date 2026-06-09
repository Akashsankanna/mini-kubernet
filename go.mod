module github.com/advanced-k8s/mini-kubernet

go 1.21

require (
	github.com/gin-contrib/cors v1.5.0
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.0.0
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.0.5
	github.com/nats-io/nats.go v1.28.0
	github.com/prometheus/client_golang v1.17.0
	k8s.io/client-go v0.28.3
	k8s.io/api v0.28.3
	golang.org/x/crypto v0.14.0
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	google.golang.org/api v0.150.0
	gopkg.in/yaml.v2 v2.4.0
)
