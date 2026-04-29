#!/bin/bash
#
# SLOzy Production Setup Script
# Performs complete production setup of SLOzy on a server
#

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}" && pwd)"

# Color codes
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $*${NC}"; }
log_success() { echo -e "${GREEN}✅ $*${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $*${NC}"; }
log_error() { echo -e "${RED}❌ $*${NC}" >&2; }

# Configuration
DEFAULT_DEPLOY_PATH="/opt/slozy"
SERVICE_USER="slozy"
SERVICE_NAME="slozy"
NGINX_SITE="slozy"

# Check for root privileges
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        exit 1
    fi
}

# Create directories
create_directories() {
    log_info "Creating directories..."

    mkdir -p "$DEPLOY_PATH"/{bin,data,templates,static,logs,config}
    mkdir -p /etc/{slozy,nginx/sites-available,systemd/system}
}

# Create user and set permissions
create_user() {
    log_info "Creating slozy user..."

    # Create user group if not exists
    groupadd -f "$SERVICE_USER" 2>/dev/null || true

    # Create user if not exists
    useradd -r -s /bin/false -g "$SERVICE_USER" -d "$DEPLOY_PATH" "$SERVICE_USER" 2>/dev/null || true

    # Set permissions
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DEPLOY_PATH"
}

# Install binary
install_binary() {
    log_info "Installing SLOzy binary..."

    # Create the binary if needed
    cd "$DEPLOY_PATH"
    go build -o bin/slozy-server ./slozy-server 2>/dev/null || {
        log_warning "Go not found or build failed"
        log_info "Copying from project root instead..."
        cp -r "$PROJECT_ROOT"/cmd "$DEPLOY_PATH/"
        cd "$DEPLOY_PATH"
    }

    # Set executable permission
    chmod +x "$DEPLOY_PATH"/bin/slozy-server
}

# Install systemd service
install_service() {
    log_info "Installing systemd service..."

    # Copy service file (will be created later)
    if [[ -f "$PROJECT_ROOT/systemd/slozy.service" ]]; then
        cp "$PROJECT_ROOT/systemd/slozy.service" /etc/systemd/system/
    else
        # Create service file
        cat > /etc/systemd/system/slozy.service << 'EOF'
[Unit]
Description=SLOzy - SLO Management Service
Documentation=https://github.com/philyuchkoff/slozy
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
ExecStart=$DEPLOY_PATH/bin/slozy-server
WorkingDirectory=$DEPLOY_PATH
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF
    fi

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
}

# Configure nginx reverse proxy
configure_nginx() {
    log_info "Configuring nginx reverse proxy..."

    # Check if nginx is installed
    if ! command -v nginx >/dev/null; then
        log_warning "Nginx not found. Skipping nginx configuration."
        return 1
    fi

    # Create nginx site configuration
    cat > /etc/nginx/sites-available/"$NGINX_SITE" << 'EOF'
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proto;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_set_header X-Forwarded-Host \$host;
        proxy_set_header X-Forwarded-Port \$server_port;

        # Headers for slozy alerts
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_http_version 1.1;
        proxy_set_header X-Forwarded-For \$remote_addr;
    }

    access_log /var/log/nginx/"$NGINX_SITE".access.log;
    error_log /var/log/nginx/"$NGINX_SITE".error.log;

    # Increase proxy timeouts
    proxy_connect_timeout 60s;
    proxy_send_timeout 60s;
    proxy_read_timeout 60s;
}
EOF

    # Enable the site
    if [[ -d /etc/nginx/sites-enabled ]]; then
        ln -sfn /etc/nginx/sites-available/"$NGINX_SITE" /etc/nginx/sites-enabled/
    else
        mkdir -p /etc/nginx/sites-enabled
        ln -sfn /etc/nginx/sites-available/"$NGINX_SITE" /etc/nginx/sites-enabled/
    fi

    # Test and reload nginx
    nginx -t && systemctl reload nginx
}

# Create environment file
create_env_file() {
    log_info "Creating environment configuration..."

    cat > "$DEPLOY_PATH"/.env << 'EOF'
# SLOzy Configuration
# Modify these values as needed

# Server configuration
PORT=8080
DATA_DIR=$DEPLOY_PATH/data
TEMPLATES_DIR=$DEPLOY_PATH/templates
STATIC_DIR=$DEPLOY_PATH/static

# Alerting configuration (optional)
# TELEGRAM_BOT_TOKEN=your_bot_token
# TELEGRAM_CHAT_ID=your_chat_id

# Logging
LOG_LEVEL=info
EOF

    chown "$SERVICE_USER:$SERVICE_USER" "$DEPLOY_PATH"/.env
    chmod 600 "$DEPLOY_PATH"/.env
}

# Setup log rotation
setup_logrotate() {
    log_info "Setting up log rotation..."

    cat > /etc/logrotate.d/slozy << 'EOF'
$DEPLOY_PATH/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644
    olddir $DEPLOY_PATH/logs/old
    postrotate
        systemctl reload slozy 2>/dev/null || true
}
EOF

    # Enable logrotate
    if [[ -f /etc/cron.daily/logrotate ]]; then
        chmod +x /etc/cron.daily/logrotate
    fi
}

# Finalize setup
finalize_setup() {
    log_info "Finalizing setup..."

    # Create directories with correct permissions
    mkdir -p "$DEPLOY_PATH"/{bin,data,templates,static,logs,config}
    chown -R "$SERVICE_USER:$SERVICE_USER" "$DEPLOY_PATH"

    # Start the service
    systemctl start "$SERVICE_NAME"
    systemctl status "$SERVICE_NAME"

    log_success "Setup complete!"
    echo
    log_success "SLOzy is now running at http://$(hostname -f):80"
    echo
    log_success "To manage service:"
    echo "  sudo systemctl {start|stop|restart|status} $SERVICE_NAME"
    echo "  sudo journalctl -u $SERVICE_NAME -f"
}

# Main execution
main() {
    check_root

    case "${1:-full}" in
        nginx)
            create_directories
            create_user
            install_binary
            create_env_file
            setup_logrotate
            configure_nginx
            ;;
        service|full)
            create_directories
            create_user
            install_binary
            install_service
            create_env_file
            setup_logrotate
            finalize_setup
            ;;
        *)
            echo "Usage: $0 [nginx|service|full]"
            echo "  nginx    - Setup with nginx reverse proxy"
            echo "  service - Setup systemd service only"
            echo "  full     - Complete setup (default)"
            exit 1
            ;;
    esac
}

# Show usage if no arguments
if [[ $# -eq 0 ]]; then
    main
fi

# Run main
main "$@"