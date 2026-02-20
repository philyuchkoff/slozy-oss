#!/bin/bash
#
# SLOzy Production Installation Script
# This script installs SLOzy as a production service
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default configuration
INSTALL_DIR="/opt/slozy"
SERVICE_USER="slozy"
SERVICE_GROUP="slozy"
ARCH="amd64"  # Change to arm64 for ARM systems
PORT="8080"

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to detect architecture
detect_arch() {
    case $(uname -m) in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --port)
            PORT="$2"
            shift 2
            ;;
        --arch)
            ARCH="$2"
            shift 2
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "  --install-dir DIR   Installation directory (default: /opt/slozy)"
            echo "  --port PORT         Service port (default: 8080)"
            echo "  --arch ARCH         Architecture (amd64|arm64, default: amd64)"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Auto-detect architecture if not specified
if [[ "$ARCH" == "amd64" && $(detect_arch) == "arm64" ]]; then
    ARCH="arm64"
    print_warning "Detected ARM64 architecture, using $ARCH"
fi

print_status "Starting SLOzy installation..."
print_status "Installation directory: $INSTALL_DIR"
print_status "Service port: $PORT"
print_status "Architecture: $ARCH"

# Check prerequisites
print_status "Checking prerequisites..."

if ! command_exists systemctl; then
    print_error "systemd is required for this installation script"
    exit 1
fi

if [[ $EUID -ne 0 ]]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Create directories
print_status "Creating directories..."
mkdir -p "$INSTALL_DIR"/{bin,data,templates,static,logs}

# Create service user
print_status "Creating service user..."
if ! getent group "$SERVICE_GROUP" >/dev/null; then
    groupadd -r "$SERVICE_GROUP"
fi

if ! getent passwd "$SERVICE_USER" >/dev/null; then
    useradd -r -s /bin/false -g "$SERVICE_GROUP" -d "$INSTALL_DIR" "$SERVICE_USER"
fi

# Download and build SLOzy
print_status "Building SLOzy..."
if [[ -d "slozy-source" ]]; then
    rm -rf slozy-source
fi

# Clone repository if not running from it
if [[ ! -f "Makefile" ]]; then
    git clone https://github.com/philyuchkoff/slozy.git slozy-source
    cd slozy-source
fi

# Build for production
make build-prod BINARY_NAME=slozy-server

# Install binary
print_status "Installing binary..."
cp "bin/slozy-server-linux-$ARCH" "$INSTALL_DIR/bin/slozy-server"
chmod +x "$INSTALL_DIR/bin/slozy-server"

# Copy static files and templates
print_status "Copying assets..."
cp -r static/* "$INSTALL_DIR/static/"
cp -r templates/* "$INSTALL_DIR/templates/"

# Set permissions
print_status "Setting permissions..."
chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
chmod -R 755 "$INSTALL_DIR"
chmod -R 644 "$INSTALL_DIR/static"/* "$INSTALL_DIR/templates"/* 2>/dev/null || true
chmod 755 "$INSTALL_DIR/bin"/*

# Create environment file
print_status "Creating environment configuration..."
cat > "$INSTALL_DIR/.env" << EOF
# SLOzy Production Configuration
PORT=$PORT
HOST=0.0.0.0
DATA_DIR=$INSTALL_DIR/data
TEMPLATES_DIR=$INSTALL_DIR/templates
STATIC_DIR=$INSTALL_DIR/static
LOG_LEVEL=info

# Add your notification configurations below
# TELEGRAM_BOT_TOKEN=your-token
# TELEGRAM_CHAT_ID=your-chat-id
EOF

chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/.env"
chmod 640 "$INSTALL_DIR/.env"

# Install systemd service
print_status "Installing systemd service..."
if [[ -f "systemd/slozy.service" ]]; then
    # Update paths in service file
    sed "s|/opt/slozy|$INSTALL_DIR|g" systemd/slozy.service > /etc/systemd/system/slozy.service
else
    # Create service file
    cat > /etc/systemd/system/slozy.service << EOF
[Unit]
Description=SLOzy - SLO Management Service
Documentation=https://github.com/philyuchkoff/slozy
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
ExecStart=$INSTALL_DIR/bin/slozy-server
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=slozy

[Install]
WantedBy=multi-user.target
EOF
fi

# Reload systemd and enable service
systemctl daemon-reload
systemctl enable slozy

# Create logrotate config
print_status "Setting up log rotation..."
cat > /etc/logrotate.d/slozy << EOF
$INSTALL_DIR/logs/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 $SERVICE_USER $SERVICE_GROUP
    postrotate
        systemctl reload slozy
    endscript
}
EOF

# Success!
print_status "Installation completed successfully!"
print_status ""
print_status "Next steps:"
print_status "1. Edit configuration: $INSTALL_DIR/.env"
print_status "2. Start the service: systemctl start slozy"
print_status "3. Check status: systemctl status slozy"
print_status "4. View logs: journalctl -u slozy -f"
print_status ""
print_status "SLOzy will be available at: http://localhost:$PORT"