# Stage 1: Build
FROM golang:1.25.0-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION:-dev} -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o vibe \
    ./cmd/vibe/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    openssh-client \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 vibe && \
    adduser -D -u 1000 -G vibe vibe

# Set working directory
WORKDIR /home/vibe

# Copy binary from builder
COPY --from=builder /build/vibe /usr/local/bin/vibe

# Create config directory
RUN mkdir -p /home/vibe/.config/vibe && \
    chown -R vibe:vibe /home/vibe

# Switch to non-root user
USER vibe

# Set entrypoint
ENTRYPOINT ["vibe"]

# Default command
CMD ["--help"]
