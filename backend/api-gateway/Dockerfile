# ==========================
# Build Stage
# ==========================
FROM golang:1.21-alpine AS builder

# Install git for private/public modules
RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o api-gateway .

# ==========================
# Runtime Stage
# ==========================
FROM alpine:3.20

# Install runtime packages
RUN apk add --no-cache ca-certificates wget tzdata

# Create non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary
COPY --from=builder /app/api-gateway .

# Change ownership
RUN chown -R app:app /app

# Run as non-root
USER app

# Expose service port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s \
    --timeout=10s \
    --start-period=5s \
    --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Start application
CMD ["./api-gateway"]