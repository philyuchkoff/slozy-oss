# Multi-stage Dockerfile for SLOzy Production
FROM golang:1.21-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build arguments
ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

# Build with optimization flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.VCSRef=${VCS_REF}" \
    -a -installsuffix cgo \
    -o slozy-server \
    ./cmd/slozy-server

# Production stage
FROM scratch

# Import CA certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Create non-root user for security
RUN addgroup -g 1001 -S slozy && \
    adduser -u 1001 -S slozy -G slozy

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/slozy-server .

# Create required directories with correct permissions
RUN mkdir -p data templates static logs && \
    chown -R slozy:slozy /app

# Switch to non-root user
USER slozy

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/slozy-server", "-healthcheck"]

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/app/slozy-server"]