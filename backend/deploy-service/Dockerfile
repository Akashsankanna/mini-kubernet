FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o deploy-service .

FROM alpine:3.20

RUN apk add --no-cache \
    ca-certificates \
    wget \
    tzdata

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/deploy-service .

RUN chown -R app:app /app

USER app

EXPOSE 8083

HEALTHCHECK --interval=30s \
    --timeout=10s \
    --start-period=5s \
    --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8083/health || exit 1

CMD ["./deploy-service"]