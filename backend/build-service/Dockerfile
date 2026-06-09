FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates docker-cli
WORKDIR /root/

COPY --from=builder /app/build-service .

EXPOSE 8082
CMD ["./build-service"]
