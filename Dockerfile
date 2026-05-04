# Runtime stage only - GoReleaser provides the pre-built binary
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    gnupg \
    age \
    sops

# Create non-root user
RUN addgroup -g 1000 keyforge && \
    adduser -D -u 1000 -G keyforge keyforge

# Copy pre-built binary from GoReleaser
COPY keyforge /usr/local/bin/keyforge

# Create directories for keys
RUN mkdir -p /home/keyforge/.config/sops/age && \
    chown -R keyforge:keyforge /home/keyforge/.config

# Switch to non-root user
USER keyforge
WORKDIR /workspace

ENTRYPOINT ["keyforge"]
CMD ["--help"]
