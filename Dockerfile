# Build stage
FROM golang:1.26-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o keyforge ./cmd/keyforge

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    gnupg \
    age \
    sops

# Create non-root user
RUN addgroup -g 1000 keyforge && \
    adduser -D -u 1000 -G keyforge keyforge

# Copy binary from builder
COPY --from=builder /build/keyforge /usr/local/bin/keyforge

# Create directories for keys
RUN mkdir -p /home/keyforge/.config/sops/age && \
    chown -R keyforge:keyforge /home/keyforge/.config

# Switch to non-root user
USER keyforge
WORKDIR /workspace

ENTRYPOINT ["keyforge"]
CMD ["--help"]
