# SLOzy Production Deployment Guide

This guide provides comprehensive instructions for deploying SLOzy in a production environment. It covers all the steps and configurations needed for a successful deployment.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Architecture Overview](#architecture-overview)
3. [Installation Methods](#installation-methods)
4. [Configuration](#configuration)
5. [Security Considerations](#security-considerations)
6. [Monitoring and Logging](#monitoring-and-logging)
7. [Troubleshooting](#troubleshooting)
8. [Maintenance](#maintenance)
9. [Migration Guide](#migration-guide)

## Prerequisites

### System Requirements

- **Operating System**: Ubuntu 20.04+, CentOS 8+, RHEL 8+, Debian 11+
- **Architecture**: x86_64 (amd64) or ARM64 (arm64)
- **Memory**: Minimum 512MB RAM, 1GB recommended
- **Storage**: Minimum 1GB disk space, 5GB recommended
- **CPU**: 1 core minimum, 2+ cores recommended

### Software Dependencies

- **Go**: 1.22 or later (for building from source)
- **systemd**: Required for service management
- **Nginx** (optional but recommended): For reverse proxy and SSL termination
- **Git**: For cloning the repository

### Network Requirements

- **Port 8080**: Default SLOzy application port
- **Port 80/443**: If using Nginx reverse proxy
- **Outbound**: Internet access for optional notifications (Telegram, Slack, etc.)

## Architecture Overview

```
                    Internet
                        |
                    [Nginx]
                     :443
                        |
                SSL Termination
                        |
                    http://
                        |
                 [SLOzy App]
                 :8080
                        |
         ┌──────────────┼──────────────┐
         │              │              │
    /opt/slozy/   Local Storage   Notification
       data/          &            Services
                    Templates
```

### Components

1. **SLOzy Application**: The main Go application serving the web interface and API
2. **Nginx** (optional): Reverse proxy providing SSL termination and rate limiting
3. **systemd**: Service manager for automatic startup and restarts
4. **Storage**: Local filesystem for data, templates, and static files

## Installation Methods

### Method 1: Automated Installation (Recommended)

The automated script handles all installation steps:

```bash
# Clone the repository
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Run the installation script
sudo ./deploy/install.sh OPTIONS

# Available options:
#   --install-dir DIR   Installation directory (default: /opt/slozy)
#   --port PORT         Service port (default: 8080)
#   --arch ARCH         Architecture (amd64|arm64, default: amd64)
```

Example with custom options:
```bash
sudo ./deploy/install.sh --install-dir /opt/slozy --port 9090
```

### Method 2: Manual Installation

#### Step 1: Create Directories and User

```bash
# Create installation directory
sudo mkdir -p /opt/slozy/{bin,data,templates,static,logs}

# Create service user and group
sudo groupadd slozy
sudo useradd -r -s /bin/false -g slozy -d /opt/slozy slozy

# Set ownership
sudo chown -R slozy:slozy /opt/slozy
```

#### Step 2: Build and Install

```bash
# Clone and build
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Build for production
make build-prod

# Install binary
sudo cp bin/slozy-server-linux-amd64 /opt/slozy/bin/slozy-server
sudo chmod +x /opt/slozy/bin/slozy-server

# Copy assets
sudo -u slozy cp -r static/* /opt/slozy/static/
sudo -u slozy cp -r templates/* /opt/slozy/templates/
```

#### Step 3: Configure Environment

```bash
# Create environment file
sudo -u slozy tee /opt/slozy/.env > /dev/null <<EOF
PORT=8080
HOST=0.0.0.0
DATA_DIR=/opt/slozy/data
TEMPLATES_DIR=/opt/slozy/templates
STATIC_DIR=/opt/slozy/static
LOG_LEVEL=info
EOF
```

#### Step 4: Install Service

```bash
# Install systemd service
sudo cp systemd/slozy.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable slozy
```

### Method 3: Docker Deployment

```bash
# Build Docker image
docker build -t slozy:latest .

# Run with persistent storage
docker run -d \
  --name slozy \
  -p 8080:8080 \
  -v /opt/slozy/data:/app/data \
  -v /opt/slozy/templates:/app/templates \
  -e PORT=8080 \
  slozy:latest
```

## Configuration

### Environment Variables

Create/edit `/opt/slozy/.env`:

```bash
# Server Configuration
PORT=8080                          # Application port
HOST=0.0.0.0                       # Bind address
DATA_DIR=/opt/slozy/data           # Data storage
TEMPLATES_DIR=/opt/slozy/templates # Templates directory
STATIC_DIR=/opt/slozy/static        # Static files
LOG_LEVEL=info                     # Log level (debug, info, warn, error)

# Notification Settings
TELEGRAM_BOT_TOKEN=your_token      # Telegram bot token
TELEGRAM_CHAT_ID=your_chat_id      # Telegram chat ID

# Security
CORS_ORIGINS=http://localhost:3000 # Allowed CORS origins
RATE_LIMIT_REQUESTS_PER_SECOND=10  # API rate limit
```

### Nginx Configuration (Optional)

1. Copy the Nginx configuration:
```bash
sudo cp deploy/nginx.conf /etc/nginx/sites-available/slozy
sudo ln -s /etc/nginx/sites-available/slozy /etc/nginx/sites-enabled/
```

2. Configure SSL (recommended for production):
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain SSL certificate
sudo certbot --nginx -d slozy.example.com
```

3. Test and reload Nginx:
```bash
sudo nginx -t
sudo systemctl reload nginx
```

### SSL/TLS Configuration

#### Self-Signed Certificate (for testing)
```bash
sudo mkdir -p /etc/ssl/slozy
sudo openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout /etc/ssl/slozy/slozy.key \
  -out /etc/ssl/slozy/slozy.crt \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=slozy.local"
```

#### Let's Encrypt Certificate (production)
```bash
# Install certbot
sudo apt update
sudo apt install certbot

# Generate certificate
sudo certbot certonly --standalone -d slozy.example.com
```

## Security Considerations

### 1. System Hardening

```bash
# Set proper permissions
sudo chmod 750 /opt/slozy
sudo chmod 640 /opt/slozy/.env
sudo chown slozy:slozy /opt/slozy/.env

# Secure binary
sudo chmod 750 /opt/slozy/bin/slozy-server
```

### 2. Firewall Configuration

```bash
# Ubuntu/Debian (UFW)
sudo ufw allow 8080/tcp
sudo ufw allow from 10.0.0.0/8 to any port 8080  # Restrict to internal

# CentOS/RHEL (firewalld)
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 3. AppArmor/Selinux

For additional security, consider creating an AppArmor profile:

```bash
# Create profile
sudo aa-genprof /opt/slozy/bin/slozy-server

# Apply profile
sudo apparmor_parser -r /etc/apparmor.d/opt.slozy.bin.slozy-server
```

### 4. Rate Limiting

Configure in Nginx (see nginx.conf) or using application settings:

```bash
# In .env
RATE_LIMIT_REQUESTS_PER_SECOND=10
RATE_LIMIT_BURST=20
```

## Monitoring and Logging

### 1. Service Status

```bash
# Check service status
sudo systemctl status slozy

# View logs
sudo journalctl -u slozy -f

# Recent errors
sudo journalctl -u slozy --since "1 hour ago" -p err
```

### 2. Log Rotation

Already configured by the installation script in `/etc/logrotate.d/slozy`

### 3. Monitoring with Prometheus

Add to Prometheus configuration:

```yaml
scrape_configs:
  - job_name: 'slozy'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: /metrics
    scrape_interval: 15s
```

### 4. Health Checks

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response
{"status":"healthy","time":"2024-01-01T12:00:00Z"}
```

## Troubleshooting

### Common Issues

#### 1. Service Won't Start

```bash
# Check logs
sudo journalctl -u slozy -n 50

# Common fixes:
# - Check port conflicts: sudo ss -tulpn | grep 8080
# - Verify permissions: ls -la /opt/slozy
# - Check .env file syntax
```

#### 2. Namespace Restrictions Error

If you see errors related to namespaces, use the compatible service:

```bash
sudo cp systemd/slozy-compat.service /etc/systemd/system/slozy.service
sudo systemctl daemon-reload
sudo systemctl restart slozy
```

#### 3. Permission Denied

```bash
# Fix ownership
sudo chown -R slozy:slozy /opt/slozy

# Fix permissions
sudo chmod 755 /opt/slozy
sudo chmod 750 /opt/slozy/bin
```

#### 4. Can't Access from Remote

```bash
# Check if binding to all interfaces
grep "HOST=" /opt/slozy/.env  # Should be 0.0.0.0

# Check firewall
sudo ufw status  # or firewall-cmd --list-all

# Check if port is listening
sudo netstat -tulpn | grep 8080
```

### Debug Mode

Enable debug logging:

```bash
# Edit .env
echo "LOG_LEVEL=debug" | sudo tee -a /opt/slozy/.env

# Restart service
sudo systemctl restart slozy

# View detailed logs
sudo journalctl -u slozy -f
```

## Maintenance

### Applying Updates

#### Using Upgrade Script

```bash
# Pull latest changes
git pull origin main

# Run upgrade
sudo ./deploy/upgrade.sh

# With backup disabled
sudo ./deploy/upgrade.sh --no-backup
```

#### Manual Update

```bash
# Stop service
sudo systemctl stop slozy

# Backup current version
sudo cp -r /opt/slozy /opt/slozy-backup-$(date +%Y%m%d)

# Build and install new version
make build-prod
sudo cp bin/slozy-server-linux-amd64 /opt/slozy/bin/slozy-server

# Update assets if needed
sudo -u slozy cp -r static/* /opt/slozy/static/
sudo -u slozy cp -r templates/* /opt/slozy/templates/

# Start service
sudo systemctl start slozy

# Verify
sudo systemctl status slozy
```

### Backup and Restore

#### Backup

```bash
#!/bin/bash
# backup-slozy.sh

BACKUP_DIR="/opt/backups/slozy-$(date +%Y%m%d-%H%M%S)"
sudo mkdir -p "$BACKUP_DIR"

# Backup application data
sudo cp -r /opt/slozy/data "$BACKUP_DIR/"
sudo cp -r /opt/slozy/templates "$BACKUP_DIR/"
sudo cp /opt/slozy/.env "$BACKUP_DIR/"

# Backup configuration
sudo cp /etc/systemd/system/slozy.service "$BACKUP_DIR/"

echo "Backup created at: $BACKUP_DIR"
```

#### Restore

```bash
#!/bin/bash
# restore-slozy.sh
BACKUP_DIR=$1

if [[ -z "$BACKUP_DIR" ]]; then
    echo "Usage: $0 /path/to/backup"
    exit 1
fi

# Stop service
sudo systemctl stop slozy

# Restore data
sudo cp -r "$BACKUP_DIR/data" /opt/slozy/
sudo cp -r "$BACKUP_DIR/templates" /opt/slozy/
sudo cp "$BACKUP_DIR/.env" /opt/slozy/

# Set permissions
sudo chown -R slozy:slozy /opt/slozy

# Start service
sudo systemctl start slozy
```

### Performance Tuning

#### System Limits

Add to `/etc/security/limits.d/slozy.conf`:
```
slozy soft nofile 65536
slozy hard nofile 65536
slozy soft nproc 4096
slozy hard nproc 4096
```

#### Nginx Tuning

In nginx.conf, tune worker settings:
```nginx
worker_processes auto;
worker_connections 1024;

# Enable caching for static files
location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
    expires 1y;
    add_header Cache-Control "public, immutable";
}
```

## Migration Guide

### Migrating from Development to Production

1. **Backup Development Data**
```bash
tar -czf slozy-dev-backup.tar.gz data/ templates/
```

2. **Export SLOs** (if using database storage)
```bash
curl -H "Accept: application/json" http://localhost:8080/api/slos > slos-export.json
```

3. **Install Production Version**
Follow the installation steps above.

4. **Import Data**
```bash
# Copy files
sudo cp -r data/* /opt/slozy/data/
sudo cp -r templates/* /opt/slozy/templates/

# Set permissions
sudo chown -R slozy:slozy /opt/slozy
```

### Migrating from Docker to Native

1. **Export Data from Docker**
```bash
# Copy data from container
docker cp slozy-container:/app/data ./data-backup
docker cp slozy-container:/app/templates ./templates-backup
```

2. **Install Native Version**
Follow the manual installation steps.

3. **Import Data**
```bash
# Restore data
sudo cp -r data-backup/* /opt/slozy/data/
sudo cp -r templates-backup/* /opt/slozy/templates/
```

### Path Migration

If you installed to a different path than `/opt/slozy`:

1. **Stop the service**
```bash
sudo systemctl stop slozy
```

2. **Move directories**
```bash
sudo mv /old/path/slozy /opt/slozy
```

3. **Update paths in configuration**
```bash
# Edit service file
sudo sed -i 's|/old/path|/opt/slozy|g' /etc/systemd/system/slozy.service
sudo sed -i 's|/old/path|/opt/slozy|g' /opt/slozy/.env

# Reload daemon
sudo systemctl daemon-reload
```

4. **Start service**
```bash
sudo systemctl start slozy
```

## Production Checklist

Before going live with SLOzy:

- [ ] Install with production paths (`/opt/slozy`)
- [ ] Configure SSL/TLS certificates
- [ ] Set up Nginx reverse proxy (recommended)
- [ ] Configure notification channels
- [ ] Set up proper logging and log rotation
- [ ] Configure firewall rules
- [ ] Set up monitoring and alerting
- [ ] Create backup schedule
- [ ] Test health checks
- [ ] Verify rate limiting
- [ ] Document contact information
- [ ] Create incident response plan

## Support

For additional help:

1. Check the [troubleshooting section](#troubleshooting)
2. Review [GitHub Issues](https://github.com/philyuchkoff/slozy/issues)
3. Check application logs: `sudo journalctl -u slozy -f`
4. Visit the [documentation](https://github.com/philyuchkoff/slozy/tree/main/docs)

---

Last updated: $(date)