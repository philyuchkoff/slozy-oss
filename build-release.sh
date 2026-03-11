#!/bin/bash
#
# SLOzy Release Build Script
# This script creates a release archive with all necessary files for deployment
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get version from git tag or argument
VERSION=${1:-$(git describe --tags --abbrev=0 2>/dev/null || echo "latest")}
RELEASE_DIR="release-${VERSION}"
ARCHIVE_NAME="slozy-${VERSION}.zip"

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Clean previous release
print_status "Cleaning previous release..."
rm -rf "$RELEASE_DIR" "$ARCHIVE_NAME"

# Create release directory
print_status "Creating release directory..."
mkdir -p "$RELEASE_DIR"

# Build binaries for different architectures
print_status "Building binaries..."
make build-prod ARCH=amd64 BINARY_NAME=slozy-server
cp "bin/slozy-server-linux-amd64" "$RELEASE_DIR/"

make build-prod ARCH=arm64 BINARY_NAME=slozy-server
cp "bin/slozy-server-linux-arm64" "$RELEASE_DIR/"

# Copy static assets and templates
print_status "Copying assets..."
mkdir -p "$RELEASE_DIR"/{static,templates,config,scripts}
cp -r static/* "$RELEASE_DIR/static/"
cp -r templates/* "$RELEASE_DIR/templates/"

# Copy configuration files
print_status "Copying configuration files..."
cat > "$RELEASE_DIR/config/.env.example" << 'EOF'
# SLOzy Production Configuration
PORT=8080
HOST=0.0.0.0
DATA_DIR=/opt/slozy/data
TEMPLATES_DIR=/opt/slozy/templates
STATIC_DIR=/opt/slozy/static
LOG_LEVEL=info

# Notification configurations (optional)
# TELEGRAM_BOT_TOKEN=your-bot-token
# TELEGRAM_CHAT_ID=your-chat-id
# SMTP_HOST=localhost:25
# SMTP_FROM=slo-alerts@example.com
# GRAFANA_URL=https://grafana.example.com
EOF

# Copy systemd service files
print_status "Copying systemd service files..."
cp systemd/slozy.service "$RELEASE_DIR/config/"
cp systemd/slozy-compat.service "$RELEASE_DIR/config/"

# Create the installation script for CentOS
print_status "Creating CentOS installation script..."
cat > "$RELEASE_DIR/scripts/install-centos.sh" << 'EOF'
#!/bin/bash
#
# SLOzy CentOS Installation Script
# This script installs or updates SLOzy on CentOS/RHEL systems
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/opt/slozy"
SERVICE_USER="slozy"
SERVICE_GROUP="slozy"
SERVICE_NAME="slozy"
BACKUP_DIR="/opt/slozy-backup-$(date +%Y%m%d-%H%M%S)"

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if running on CentOS/RHEL
check_os() {
    if [[ ! -f /etc/redhat-release ]]; then
        print_error "This script is designed for CentOS/RHEL systems"
        exit 1
    fi
    print_status "Detected $(cat /etc/redhat-release)"
}

# Function to check architecture
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

# Function to install dependencies
install_dependencies() {
    print_status "Installing dependencies..."
    yum update -y
    yum install -y curl wget unzip
}

# Function to create service user
create_service_user() {
    print_status "Creating service user..."
    if ! getent group "$SERVICE_GROUP" >/dev/null; then
        groupadd -r "$SERVICE_GROUP"
    fi

    if ! getent passwd "$SERVICE_USER" >/dev/null; then
        useradd -r -s /bin/false -g "$SERVICE_GROUP" -d "$INSTALL_DIR" "$SERVICE_USER"
    fi
}

# Function to backup existing installation
backup_existing() {
    if [[ -d "$INSTALL_DIR" ]]; then
        print_warning "Existing installation found. Backing up to $BACKUP_DIR"
        mkdir -p "$BACKUP_DIR"
        cp -r "$INSTALL_DIR"/* "$BACKUP_DIR/" 2>/dev/null || true
    fi
}

# Function to install files
install_files() {
    local arch=$(detect_arch)
    local binary_name="slozy-server-linux-${arch}"

    print_status "Installing SLOzy to $INSTALL_DIR..."

    # Create directories
    mkdir -p "$INSTALL_DIR"/{bin,data,templates,static,logs,config}

    # Install binary
    cp "../$binary_name" "$INSTALL_DIR/bin/slozy-server"
    chmod +x "$INSTALL_DIR/bin/slozy-server"

    # Install assets
    cp -r ../static/* "$INSTALL_DIR/static/"
    cp -r ../templates/* "$INSTALL_DIR/templates/"

    # Install configs
    cp ../config/.env.example "$INSTALL_DIR/.env"
    cp ../config/slozy.service /etc/systemd/system/

    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chmod -R 755 "$INSTALL_DIR"
    chmod 640 "$INSTALL_DIR/.env"
}

# Function to configure service
configure_service() {
    print_status "Configuring systemd service..."

    # Update paths in service file
    sed -i "s|/opt/slozy|$INSTALL_DIR|g" /etc/systemd/system/slozy.service

    # Reload systemd
    systemctl daemon-reload
    systemctl enable slozy
}

# Function to check if port is available
check_port() {
    local port=${1:-8080}
    if ss -tuln | grep -q ":$port "; then
        print_warning "Port $port is already in use"
        read -p "Do you want to use a different port? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            read -p "Enter new port: " new_port
            sed -i "s/PORT=.*/PORT=$new_port/" "$INSTALL_DIR/.env"
            check_port "$new_port"
        fi
    fi
}

# Function to restart service
restart_service() {
    print_status "Restarting SLOzy service..."
    systemctl restart slozy

    # Wait for service to start
    sleep 3

    if systemctl is-active --quiet slozy; then
        print_status "Service started successfully"
    else
        print_error "Failed to start service"
        systemctl status slozy
        exit 1
    fi
}

# Function to verify installation
verify_installation() {
    print_status "Verifying installation..."

    # Check service status
    if systemctl is-active --quiet slozy; then
        print_status "✓ Service is running"
    else
        print_error "✗ Service is not running"
        return 1
    fi

    # Check port
    local port=$(grep "^PORT=" "$INSTALL_DIR/.env" | cut -d'=' -f2)
    if ss -tuln | grep -q ":$port "; then
        print_status "✓ Service is listening on port $port"
    else
        print_error "✗ Service is not listening on expected port"
        return 1
    fi

    # Check binary
    if [[ -x "$INSTALL_DIR/bin/slozy-server" ]]; then
        print_status "✓ Binary is installed and executable"
    else
        print_error "✗ Binary not found or not executable"
        return 1
    fi

    # Test HTTP endpoint
    sleep 2
    if curl -s -o /dev/null -w "%{http_code}" "http://localhost:$port/health" | grep -q "200"; then
        print_status "✓ Health check passed"
    else
        print_warning "Health check endpoint not responding (this may be normal on first start)"
    fi
}

# Function to show post-install info
show_post_install() {
    local port=$(grep "^PORT=" "$INSTALL_DIR/.env" | cut -d'=' -f2)

    print_status ""
    print_status "Installation completed successfully!"
    print_status ""
    print_status "Service Management:"
    print_status "  Start:   systemctl start slozy"
    print_status "  Stop:    systemctl stop slozy"
    print_status "  Restart: systemctl restart slozy"
    print_status "  Status:  systemctl status slozy"
    print_status ""
    print_status "Logs:"
    print_status "  journalctl -u slozy -f"
    print_status ""
    print_status "Configuration:"
    print_status "  Edit: $INSTALL_DIR/.env"
    print_status ""
    print_status "Application URL:"
    print_status "  http://localhost:$port"
    print_status ""
    if [[ -n "$BACKUP_DIR" && -d "$BACKUP_DIR" ]]; then
        print_status "Backup location:"
        print_status "  $BACKUP_DIR"
    fi
}

# Main execution
main() {
    print_status "Starting SLOzy installation on CentOS/RHEL..."

    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi

    # Change to script directory
    cd "$(dirname "$0")"

    # Execute installation steps
    check_os
    install_dependencies
    create_service_user
    backup_existing
    install_files
    configure_service
    check_port
    restart_service
    verify_installation
    show_post_install

    print_status "All done!"
}

# Run main function
main "$@"
EOF

chmod +x "$RELEASE_DIR/scripts/install-centos.sh"

# Create version info file
print_status "Creating version info..."
cat > "$RELEASE_DIR/VERSION.txt" << EOF
SLOzy Release: $VERSION
Build Date: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
Git Commit: $(git rev-parse HEAD 2>/dev/null || echo "unknown")
Architecture: amd64, arm64
EOF

# Create README for release
print_status "Creating release README..."
cat > "$RELEASE_DIR/README.md" << EOF
# SLOzy Release $VERSION

This archive contains everything needed to deploy SLOzy on CentOS/RHEL systems.

## Quick Installation

\`\`\`bash
# Extract the archive
unzip slozy-$VERSION.zip
cd release-$VERSION

# Run the installation script (requires sudo)
sudo ./scripts/install-centos.sh
\`\`\`

## Manual Installation

If you prefer manual installation:

1. **Copy files to target system**:
   \`\`\`bash
   # Copy the appropriate binary for your architecture
   cp slozy-server-linux-amd64 /opt/slozy/bin/slozy-server  # for Intel/AMD
   # or
   cp slozy-server-linux-arm64 /opt/slozy/bin/slozy-server  # for ARM
   \`\`\`

2. **Create service user**:
   \`\`\`bash
   sudo groupadd -r slozy
   sudo useradd -r -s /bin/false -g slozy -d /opt/slozy slozy
   \`\`\`

3. **Install files**:
   \`\`\`bash
   sudo mkdir -p /opt/slozy/{bin,data,templates,static,logs,config}
   sudo cp slozy-server-linux-* /opt/slozy/bin/slozy-server
   sudo cp -r static/* /opt/slozy/static/
   sudo cp -r templates/* /opt/slozy/templates/
   sudo cp config/.env.example /opt/slozy/.env
   sudo cp config/slozy.service /etc/systemd/system/
   \`\`\`

4. **Set permissions**:
   \`\`\`bash
   sudo chown -R slozy:slozy /opt/slozy
   sudo chmod +x /opt/slozy/bin/slozy-server
   \`\`\`

5. **Configure service**:
   \`\`\`bash
   sudo systemctl daemon-reload
   sudo systemctl enable slozy
   sudo systemctl start slozy
   \`\`\`

## Configuration

Edit the configuration file at \`/opt/slozy/.env\` to customize:

- Server port (default: 8080)
- Data directories
- Notification settings (Telegram, SMTP, etc.)

## Service Management

\`\`\`bash
# Start service
sudo systemctl start slozy

# Stop service
sudo systemctl stop slozy

# Restart service
sudo systemctl restart slozy

# Check status
sudo systemctl status slozy

# View logs
sudo journalctl -u slozy -f
\`\`\`

## Troubleshooting

1. **Service won't start**:
   \`\`\`bash
   sudo journalctl -u slozy -n 50
   \`\`\`

2. **Port already in use**:
   Edit \`/opt/slozy/.env\` and change the PORT value, then restart the service.

3. **Permission issues**:
   \`\`\`bash
   sudo chown -R slozy:slozy /opt/slozy
   sudo chmod -R 755 /opt/slozy
   \`\`\`

## Support

For issues and questions:
- GitHub: https://github.com/philyuchkoff/slozy
- Documentation: https://slozy.ru
EOF

# Create zip archive
print_status "Creating zip archive..."
cd "$RELEASE_DIR"
zip -r "../$ARCHIVE_NAME" .
cd ..

# Clean up release directory
print_status "Cleaning up..."
rm -rf "$RELEASE_DIR"

# Success!
print_status ""
print_status "✅ Release archive created successfully!"
print_status ""
print_status "Archive name: $ARCHIVE_NAME"
print_status "Size: $(du -h "$ARCHIVE_NAME" | cut -f1)"
print_status ""
print_status "To deploy on CentOS:"
print_status "1. Copy $ARCHIVE_NAME to target machine"
print_status "2. Extract: unzip $ARCHIVE_NAME"
print_status "3. Run: sudo cd release-$VERSION && ./scripts/install-centos.sh"