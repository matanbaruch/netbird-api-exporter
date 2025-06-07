# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o netbird-api-exporter .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

# Create app directory with proper permissions for nobody user
RUN mkdir -p /app && \
    chown -R 65534:65534 /app

WORKDIR /app

# Copy the binary from builder and set permissions
COPY --from=builder --chown=65534:65534 /app/netbird-api-exporter .
RUN chmod +x netbird-api-exporter

# Switch to non-root user (nobody)
USER 65534

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["./netbird-api-exporter"]
