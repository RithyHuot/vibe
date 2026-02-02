# Use Alpine as base image
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

# Copy pre-built binary from GoReleaser
COPY vibe /usr/local/bin/vibe

# Create config directory
RUN mkdir -p /home/vibe/.config/vibe && \
    chown -R vibe:vibe /home/vibe

# Switch to non-root user
USER vibe

# Set entrypoint
ENTRYPOINT ["vibe"]

# Default command
CMD ["--help"]
