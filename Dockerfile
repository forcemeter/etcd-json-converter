# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git for go mod download
RUN apk add --no-cache git upx

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

# Copy source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o etcd-json-converter .
RUN upx etcd-json-converter

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /app/etcd-json-converter /usr/local/bin/

# Default data directory
VOLUME /data

ENTRYPOINT ["etcd-json-converter"]

CMD ["--help"]