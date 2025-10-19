# Multi-stage build for kubeflake keygen server
# Stage 1: Build the Go binary
FROM golang:1.25-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy the entire source code (no external deps, so skip go mod download)
COPY . .

# Build the keygen server binary
# CGO_ENABLED=0 for a static binary
# -ldflags="-w -s" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build \
  -ldflags="-w -s" \
  -o /out/keygen-server \
  ./examples/server

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Add ca-certificates for HTTPS requests (metadata services)
RUN apk --no-cache add ca-certificates wget

# Create non-root user
RUN addgroup -g 1000 keygen && \
  adduser -D -u 1000 -G keygen keygen

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /out/keygen-server /usr/local/bin/keygen-server

# Change ownership to non-root user
RUN chown -R keygen:keygen /app

# Switch to non-root user
USER keygen

# Expose the default port
EXPOSE 8083

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8083/health || exit 1

# Run the server
ENTRYPOINT ["/usr/local/bin/keygen-server"]
CMD ["--address", ":8083"]
