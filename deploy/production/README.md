# SLOzy Production Deployment Guide

## Overview

This guide provides instructions for deploying SLOzy to production with all security fixes enabled.

## Prerequisites

- Go 1.21 or later
- Linux/macOS/Windows server
- Docker and Docker Compose (optional)
- TLS certificates (for HTTPS)

## Security Configuration

### 1. Environment Variables

Create a `.env` file with your production configuration:

```bash
# Server Configuration
PORT=8080
ENVIRONMENT=production

# TLS/SSL Configuration
TLS_CERT_PATH=/path/to/cert.pem
TLS_KEY_PATH=/path/to/key.pem

# Rate Limiting
API_RATE_LIMIT=10
CREATE_SLO_RATE_LIMIT=5

# CORS Origins (comma-separated)
ALLOWED_ORIGINS=https://yourdomain.com,https://app.yourdomain.com

# Telegram Alerts (optional)
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id

# Authentication (if implementing)
API_TOKEN_SECRET=your_random_secret_here
```

### 2. Reverse Proxy Configuration

#### Nginx Configuration:

```nginx
server {
    listen 443 ssl http2;
    server_name slozy.yourdomain.com;

    # SSL Configuration
    ssl_certificate /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;

    # Security Headers
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;

    # Rate Limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=5r/s;
    limit_req zone=api burst=10 nodelay;

    # Proxy to Go Server
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # Timeouts
        proxy_connect_timeout 5s;
        proxy_read_timeout 60s;
        proxy_send_timeout 60s;
    }

    # Static Files with Caching
    location /static/ {
        proxy_pass http://localhost:8080;
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}

# HTTP to HTTPS Redirect
server {
    listen 80;
    server_name slozy.yourdomain.com;
    return 301 https://$server_name$request_uri;
}
```

## Deployment Options

### Option 1: Direct Binary Deployment

1. Build for production:

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o slozy-server .

# Make it executable
chmod +x slozy-server
```

2. Create production directory structure:

```bash
sudo mkdir -p /opt/slozy/{bin,data,templates,static,logs,config}
sudo cp slozy-server /opt/slozy/bin/
sudo cp -r data/templates /opt/slozy/
sudo cp -r cmd/slozy-server/static /opt/slozy/
```

3. Create systemd service:

```bash
sudo tee /etc/systemd/system/slozy.service > /dev/null <<EOF
[Unit]
Description=SLOzy SLO Management Server
After=network.target

[Service]
Type=simple
User=slozy
Group=slozy
WorkingDirectory=/opt/slozy
ExecStart=/opt/slozy/bin/slozy-server
Restart=always
RestartSec=5
EnvironmentFile=/opt/slozy/config/.env
StandardOutput=append:/opt/slozy/logs/slozy.log
StandardError=append:/opt/slozy/logs/error.log

[Install]
WantedBy=multi-user.target
EOF

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable slozy
sudo systemctl start slozy
```

### Option 2: Docker Deployment

1. Create Dockerfile:

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build optimized binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o slozy-server ./cmd/slozy-server

# Runtime image
FROM scratch
WORKDIR /app

# Create non-root user
RUN addgroup -g 1000 slozy && adduser -D -s /bin/sh -u 1000 -G slozy slozy

# Copy binary and create directories
COPY --from=builder /app/slozy-server .
RUN mkdir -p data templates static logs

USER slozy
EXPOSE 8080

CMD ["./slozy-server"]
```

2. Create docker-compose.yml:

```yaml
version: '3.8'

services:
  slozy:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENVIRONMENT=production
      - API_RATE_LIMIT=${API_RATE_LIMIT:-10}
      - CREATE_SLO_RATE_LIMIT=${CREATE_SLO_RATE_LIMIT:-5}
      - ALLOWED_ORIGINS=${ALLOWED_ORIGINS:-https://yourdomain.com}
      - TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
      - TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}
    volumes:
      - ./data:/app/data
      - ./templates:/app/templates
      - ./logs:/app/logs
    restart: unless-stopped
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl:ro
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - slozy
    restart: unless-stopped
```

3. Deploy:

```bash
# Build and start
docker-compose up -d

# View logs
docker-compose logs -f slozy
```

## Production Checklist

### Security
- [ ] TLS/SSL configured
- [ ] HSTS enabled
- [ ] Security headers present
- [ ] Rate limiting configured
- [ ] CORS restricted to production domains
- [ ] Input validation enabled
- [ ] File access restrictions active

### Monitoring
- [ ] Application logs configured
- [ ] Error logging implemented
- [ ] Health endpoint accessible
- [ ] Metrics collection (Prometheus) set up
- [ ] Alerting configured

### Performance
- [ ] Gzip compression enabled in reverse proxy
- [ ] Static assets served with caching
- [ ] Database queries optimized
- [ ] Connection pooling configured

### Backup
- [ ] Data directory backed up
- [ ] Configuration files backed up
- [ ] Disaster recovery plan documented

### Scaling
- [ ] Auto-scaling configured (if using cloud)
- [ ] Load balancer configured
- [ ] Session affinity configured
- [ ] Horizontal scaling ready

## Post-Deployment Verification

1. Test security endpoints:

```bash
# Test security headers
curl -I https://slozy.yourdomain.com/api/slo/create

# Test rate limiting
curl -X POST https://slozy.yourdomain.com/api/slo/create \
  -H "Content-Type: application/json" \
  -d '{"serviceName":"test","sloName":"test","target":99.9,"metricQuery":"test"}' \
  --repeat 20

# Test path traversal
curl -X POST https://slozy.yourdomain.com/api/slo/create \
  -H "Content-Type: application/json" \
  -d '{"serviceName":"../../../etc/passwd","sloName":"test","target":99.9,"metricQuery":"test"}'
```

2. Monitor logs:

```bash
# Check application logs
sudo journalctl -u slozy -f

# Or with Docker
docker-compose logs -f slozy
```

## Troubleshooting

### Common Issues

1. **Port Already in Use**
```bash
sudo netstat -tulpn | grep :8080
sudo kill -9 <PID>
```

2. **Permission Denied**
```bash
sudo chown -R slozy:slozy /opt/slozy
sudo chmod +x /opt/slozy/bin/slozy-server
```

3. **SSL Certificate Issues**
```bash
# Verify certificate
openssl x509 -in /path/to/cert.pem -text -noout

# Test SSL
curl -I https://slozy.yourdomain.com
```

## Monitoring and Alerting

### Recommended Tools

1. **Prometheus + Grafana**
   - Monitor request rates, response times
   - Track rate limit hits
   - Alert on high error rates

2. **ELK Stack (Elasticsearch, Logstash, Kibana)**
   - Centralized logging
   - Real-time log analysis
   - Security event monitoring

3. **OpenTelemetry**
   - Distributed tracing
   - Performance metrics
   - Service dependency tracking

### Key Metrics to Monitor

- Request throughput
- Error rate (4xx, 5xx)
- Response time percentiles
- Rate limit violations
- File system usage
- Memory and CPU usage
- TLS handshake failures

## Security Best Practices

1. **Regular Updates**
   - Keep Go runtime updated
   - Monitor for security vulnerabilities
   - Update dependencies regularly

2. **Secret Management**
   - Use environment variables or secret store
   - Never commit secrets to Git
   - Rotate secrets regularly

3. **Network Security**
   - Run behind firewall
   - Use VPN for admin access
   - Limit inbound/outbound connections

4. **Access Control**
   - Implement authentication/authorization
   - Use role-based access control
   - Audit access logs regularly

## Support

For production issues:
1. Check service logs
2. Verify configuration
3. Review security metrics
4. Check resource utilization

For security vulnerabilities, follow your organization's disclosure policy.