# Multi-stage Dockerfile for getblobz
# Produces a minimal Alpine-based image with the getblobz binary

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w -extldflags '-static'" -o getblobz main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 getblobz && \
    adduser -D -u 1000 -G getblobz getblobz

# Create directories for data and config
RUN mkdir -p /data /config && \
    chown -R getblobz:getblobz /data /config

WORKDIR /home/getblobz

# Copy binary from builder
COPY --from=builder /build/getblobz /usr/local/bin/getblobz

# Switch to non-root user
USER getblobz

# Set volumes
VOLUME ["/data", "/config"]

# Environment variables
ENV GETBLOBZ_OUTPUT_PATH=/data \
    GETBLOBZ_STATE_DATABASE=/data/.sync-state.db

# Default command shows help
ENTRYPOINT ["getblobz"]
CMD ["--help"]

# Metadata
LABEL org.opencontainers.image.title="getblobz" \
      org.opencontainers.image.description="Azure Blob Storage sync tool" \
      org.opencontainers.image.vendor="haepapa" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/haepapa/getblobz"
