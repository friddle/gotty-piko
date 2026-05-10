# Build stage
FROM docker.linkos.org/library/golang:1.23-alpine AS builder

WORKDIR /app

# Configure Go proxy for China network environment
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GO111MODULE=on

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM docker.linkos.org/library/alpine:3.19

WORKDIR /app

# Use China mirror for Alpine packages
RUN sed -i 's|dl-cdn.alpinelinux.org|mirrors.tuna.tsinghua.edu.cn|g' /etc/apk/repositories \
    && apk --no-cache add ca-certificates curl

# Copy the binary from builder
COPY --from=builder /app/server .

# Expose ports
EXPOSE 80 8022 8023

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:80/health || exit 1

# Run the server
CMD ["./server"]
