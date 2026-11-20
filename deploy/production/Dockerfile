# Multi-stage production Dockerfile for SLOzy OSS
# For development, use the root Dockerfile (single-stage alpine)

FROM golang:1.27.3-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

ARG VERSION=dev
ARG BUILD_DATE
ARG VCS_REF

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s -X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.VCSRef=${VCS_REF}" \
    -a -installsuffix cgo \
    -o slozy-server \
    ./cmd/slozy-server

# Production stage
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN addgroup -g 1001 -S slozy && \
    adduser -u 1001 -S slozy -G slozy

WORKDIR /app

COPY --from=builder /app/slozy-server .

# Copy static files and templates
COPY static/ ./static/
COPY templates/ ./templates/

RUN mkdir -p data logs && \
    chown -R slozy:slozy /app

USER slozy

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/app/slozy-server", "-healthcheck"]

EXPOSE 8080

ENTRYPOINT ["/app/slozy-server"]
