#!/bin/bash
#
# SLOzy Upgrade Script
# This script upgrades an existing SLOzy installation
#

set -euo pipefail

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Default configuration
INSTALL_DIR="/opt/slozy"
SERVICE_NAME="slozy"
BACKUP_DIR="/opt/slozy-backup-$(date +%Y%m%d-%H%M%S)"
ARCH="amd64"

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --backup-dir)
            BACKUP_DIR="$2"
            shift 2
            ;;
        --no-backup)
            NO_BACKUP=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "  --install-dir DIR   Installation directory (default: /opt/slozy)"
            echo "  --backup-dir DIR    Custom backup directory"
            echo "  --no-backup         Skip backup creation"
            echo "  --help              Show this help message"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

print_status "Starting SLOzy upgrade..."
print_status "Installation directory: $INSTALL_DIR"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    print_error "This script must be run as root (use sudo)"
    exit 1
fi

# Check if SLOzy is installed
if [[ ! -f "$INSTALL_DIR/bin/slozy-server" ]]; then
    print_error "SLOzy is not installed at $INSTALL_DIR"
    exit 1
fi

# Check if service is running
SERVICE_STATUS=$(systemctl is-active "$SERVICE_NAME" 2>/dev/null || echo "inactive")

# Create backup unless skipped
if [[ "${NO_BACKUP:-false}" != "true" ]]; then
    print_status "Creating backup at $BACKUP_DIR..."
    mkdir -p "$BACKUP_DIR"
    cp -r "$INSTALL_DIR"/* "$BACKUP_DIR/"
    print_status "Backup created successfully"
fi

# Stop service
if [[ "$SERVICE_STATUS" == "active" ]]; then
    print_status "Stopping SLOzy service..."
    systemctl stop "$SERVICE_NAME"
fi

# Build new version
print_status "Building new version..."
if [[ -d "slozy-source" ]]; then
    rm -rf slozy-source
fi

if [[ ! -f "Makefile" ]]; then
    git clone https://github.com/philyuchkoff/slozy.git slozy-source
    cd slozy-source
fi

make build-prod BINARY_NAME=slozy-server

# Detect architecture
ARCH=$(uname -m | sed 's/x86_64/amd64/' | sed 's/aarch64/arm64/')

# Install new binary
print_status "Installing new binary..."
cp "bin/slozy-server-linux-$ARCH" "$INSTALL_DIR/bin/slozy-server"
chmod +x "$INSTALL_DIR/bin/slozy-server"

# Update static files and templates (preserve customizations)
print_status "Updating assets..."
if [[ -d "$INSTALL_DIR/static.bak" ]]; then
    rm -rf "$INSTALL_DIR/static.bak"
fi
mv "$INSTALL_DIR/static" "$INSTALL_DIR/static.bak"
cp -r static/* "$INSTALL_DIR/static/"
chown -R slozy:slozy "$INSTALL_DIR/static"

# Note: We preserve templates directory as users may have custom templates
print_warning "Templates directory preserved - manually update if needed"

# Start service
print_status "Starting SLOzy service..."
systemctl start "$SERVICE_NAME"
sleep 2

# Check service status
if systemctl is-active --quiet "$SERVICE_NAME"; then
    print_status "✅ SLOzy upgraded successfully!"
    print_status "Service is running"
else
    print_error "❌ Failed to start SLOzy service"
    print_status "Checking logs:"
    journalctl -u "$SERVICE_NAME" --no-pager -n 20

    # Restore from backup if upgrade failed
    if [[ "${NO_BACKUP:-false}" != "true" ]]; then
        print_warning "Attempting to restore from backup..."
        systemctl stop "$SERVICE_NAME"
        cp -r "$BACKUP_DIR"/* "$INSTALL_DIR/"
        systemctl start "$SERVICE_NAME"
        print_status "Backup restored"
    fi
    exit 1
fi

print_status ""
print_status "Upgrade completed!"
print_status "To view new features: journalctl -u $SERVICE_NAME -f"