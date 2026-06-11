# ==========================
# Build Stage
# ==========================
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o build-service .

# ==========================
# Runtime Stage
# ==========================
FROM alpine:3.20

# Runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    docker-cli \
    wget \
    tzdata

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary
COPY --from=builder /app/build-service .

# Set permissions
RUN chown -R app:app /app

# Run as non-root
USER app

# Service port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s \
    --timeout=10s \
    --start-period=5s \
    --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8082/health || exit 1

# Start service
CMD ["./build-service"]