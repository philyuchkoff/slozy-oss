# SLOzy Production Quick Start

This guide gets you running SLOzy in production in under 10 minutes.

## 🚀 One-Command Installation

```bash
curl -sfL https://raw.githubusercontent.com/philyuchkoff/slozy/main/deploy/install.sh | sudo bash -s -- --install-dir /opt/slozy --port 8080
```

## 📋 Manual Quick Start

### Prerequisites Check

```bash
# Check OS and systemd
cat /etc/os-release
systemctl --version

# Check if port 8080 is available
ss -tulpn | grep 8080 || echo "Port 8080 is available"
```

### Step 1: Clone and Build

```bash
# Clone repository
git clone https://github.com/philyuchkoff/slozy.git
cd /tmp/slozy

# Install Go if not present (Ubuntu/Debian)
sudo apt update && sudo apt install -y golang-go
# Or download from: https://golang.org/dl/

# Build for production
make build-prod
```

### Step 2: Install Application

```bash
# Create directories
sudo mkdir -p /opt/slozy/{bin,data,templates,static,logs}

# Create service user
sudo useradd -r -s /bin/false -g slozy -d /opt/slozy slozy 2>/dev/null || sudo groupadd slozy 2>/dev/null

# Install binary
sudo cp bin/slozy-server-linux-amd64 /opt/slozy/bin/slozy-server
sudo chmod +x /opt/slozy/bin/slozy-server

# Copy assets
sudo cp -r static/* /opt/slozy/static/
sudo cp -r templates/* /opt/slozy/templates/
```

### Step 3: Configure

```bash
# Create environment file
sudo tee /opt/slozy/.env > /dev/null <<EOF
PORT=8080
HOST=0.0.0.0
DATA_DIR=/opt/slozy/data
TEMPLATES_DIR=/opt/slozy/templates
STATIC_DIR=/opt/slozy/static
LOG_LEVEL=info
EOF

# Set permissions
sudo chown -R slozy:slozy /opt/slozy
sudo chmod 640 /opt/slozy/.env
```

### Step 4: Install Service

```bash
# Create systemd service
sudo tee /etc/systemd/system/slozy.service > /dev/null <<EOF
[Unit]
Description=SLOzy - SLO Management Service
After=network.target

[Service]
Type=simple
User=slozy
ExecStart=/opt/slozy/bin/slozy-server
WorkingDirectory=/opt/slozy
EnvironmentFile=/opt/slozy/.env
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable slozy
sudo systemctl start slozy
```

### Step 5: Verify Installation

```bash
# Check service status
sudo systemctl status slozy

# Check logs
sudo journalctl -u slozy -n 20

# Test health endpoint
curl http://localhost:8080/health
```

## 🔧 Optional: Nginx Reverse Proxy

```bash
# Install Nginx
sudo apt install nginx -y

# Create Nginx config
sudo tee /etc/nginx/sites-available/slozy > /dev/null <<EOF
server {
    listen 80;
    server_name slozy.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

# Enable site
sudo ln -s /etc/nginx/sites-available/slozy /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default

# Test and restart
sudo nginx -t && sudo systemctl restart nginx
```

## 🔑 Add SSL Certificate (Let's Encrypt)

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx -y

# Get certificate
sudo certbot --nginx -d slozy.example.com --non-interactive --email admin@example.com --agree-tos
```

## 📊 Access SLOzy

SLOzy is now available at:
- Direct: http://your-server:8080
- Via Nginx: http://slozy.example.com
- With SSL: https://slozy.example.com

### Next Steps

1. Configure notification channels in `/opt/slozy/.env`
2. Set up monitoring for the SLOzy service
3. Configure backup schedule
4. Review [Full Deployment Guide](deployment.md) for advanced configuration

## 🔸 Common Commands

```bash
# View logs
sudo journalctl -u slozy -f

# Restart service
sudo systemctl restart slozy

# Update SLOzy
cd /tmp/slozy && git pull && make build-prod && sudo cp bin/slozy-server-linux-amd64 /opt/slozy/bin/slozy-server && sudo systemctl restart slozy

# Backup data
sudo tar -czf slozy-backup-$(date +%Y%m%d).tar.gz /opt/slozy/data /opt/slozy/templates /opt/slozy/.env
```