#!/bin/bash

# SLOzy Production Deployment Script
# Use: ./deploy.sh [options]

set -e

# Default values
ENVIRONMENT="production"
DOMAIN="slozy.yourdomain.com"
DEPLOY_PATH="/opt/slozy"
SERVICE_NAME="slozy"
NGINX_PATH="/etc/nginx/sites-available"
SSL_PATH="/etc/ssl/letsencrypt"

# Parse options
while [[ $# -gt 0 ]]; do
    case $1 in
        --domain)
            DOMAIN="$2"
            shift 2
            ;;
        --deploy-path)
            DEPLOY_PATH="$2"
            shift 2
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --domain DOMAIN          Domain name (default: $DOMAIN)"
            echo "  --deploy-path PATH      Deployment path (default: $DEPLOY_PATH)"
            echo "  --dry-run            Show commands without executing"
            echo "  --help               Show this help"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "🚀 SLOzy Production Deployment"
echo "================================"
echo "Domain: $DOMAIN"
echo "Deploy Path: $DEPLOY_PATH"
if [ "$DRY_RUN" = true ]; then
    echo "Mode: DRY RUN (no changes will be made)"
fi
echo ""

# Check prerequisites
check_prereqs() {
    echo "📋 Checking prerequisites..."

    # Check if running as root
    if [ "$EUID" -ne 0 ] && [ "$DRY_RUN" != true ]; then
        echo "❌ Must run as root"
        exit 1
    fi

    # Check if Go installed
    if ! command -v go &> /dev/null; then
        echo "❌ Go is not installed"
        exit 1
    fi

    # Check if Docker installed (optional)
    if command -v docker &> /dev/null; then
        echo "✅ Docker found"
        USE_DOCKER=true
    else
        echo "⚠️  Docker not found, using binary deployment"
        USE_DOCKER=false
    fi

    echo "✅ Prerequisites checked"
}

# Build application
build_app() {
    echo ""
    echo "🔨 Building SLOzy..."

    if [ "$DRY_RUN" != true ]; then
        # Build binary
        CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
            go build -ldflags="-w -s" -o slozy-server .

        echo "✅ Binary built: slozy-server"
    else
        echo "Would run: CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build..."
    fi
}

# Create system user
create_user() {
    echo ""
    echo "👤 Creating system user..."

    if [ "$DRY_RUN" != true ]; then
        if ! id -u slozy &>/dev/null; then
            useradd -r -s /bin/false -d $DEPLOY_PATH slozy
            echo "✅ User 'slozy' created"
        else
            echo "⚠️  User 'slozy' already exists"
        fi
    else
        echo "Would create user: slozy"
    fi
}

# Create directory structure
create_dirs() {
    echo ""
    echo "📁 Creating directory structure..."

    dirs=("$DEPLOY_PATH" "$DEPLOY_PATH/data" "$DEPLOY_PATH/templates" "$DEPLOY_PATH/static" "$DEPLOY_PATH/logs")

    for dir in "${dirs[@]}"; do
        if [ "$DRY_RUN" != true ]; then
            mkdir -p $dir
            echo "  Created: $dir"
        else
            echo "Would create: $dir"
        fi
    done

    if [ "$DRY_RUN" != true ]; then
        chown -R slozy:slozy $DEPLOY_PATH
        chmod 755 $DEPLOY_PATH
        echo "✅ Permissions set"
    fi
}

# Deploy files
deploy_files() {
    echo ""
    echo "📤 Deploying files..."

    if [ "$DRY_RUN" != true ]; then
        cp slozy-server $DEPLOY_PATH/
        cp -r static/* $DEPLOY_PATH/static/
        echo "✅ Files deployed"
    else
        echo "Would copy: slozy-server to $DEPLOY_PATH/"
        echo "Would copy: static files to $DEPLOY_PATH/static/"
    fi
}

# Configure systemd service
configure_systemd() {
    echo ""
    echo "⚙️  Configuring service..."

    cat > slozy.service <<EOF
[Unit]
Description=SLOzy SLO Management Server
After=network.target

[Service]
Type=simple
User=slozy
Group=slozy
WorkingDirectory=$DEPLOY_PATH
ExecStart=$DEPLOY_PATH/slozy-server
Restart=always
RestartSec=5
Environment=ENVIRONMENT=$ENVIRONMENT

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$DEPLOY_PATH/data $DEPLOY_PATH/templates $DEPLOY_PATH/logs

# Logging
StandardOutput=append:$DEPLOY_PATH/logs/slozy.log
StandardError=append:$DEPLOY_PATH/logs/error.log

[Install]
WantedBy=multi-user.target
EOF

    if [ "$DRY_RUN" != true ]; then
        mv slozy.service /etc/systemd/system/
        systemctl daemon-reload
        systemctl enable slozy
        echo "✅ Service configured"
    else
        echo "Would create: /etc/systemd/system/slozy.service"
    fi
}

# Create sample configuration
create_config() {
    echo ""
    echo "🔧 Creating configuration..."

    cat > $DEPLOY_PATH/.env <<EOF
# SLOzy Production Configuration
ENVIRONMENT=$ENVIRONMENT
PORT=8080

# Security
API_RATE_LIMIT=10
CREATE_SLO_RATE_LIMIT=5
ALLOWED_ORIGINS=https://$DOMAIN

# Optional: Telegram Alerts
# TELEGRAM_BOT_TOKEN=your_bot_token
# TELEGRAM_CHAT_ID=your_chat_id
EOF

    if [ "$DRY_RUN" != true ]; then
        chown slozy:slozy $DEPLOY_PATH/.env
        chmod 600 $DEPLOY_PATH/.env
        echo "✅ Configuration created at $DEPLOY_PATH/.env"
    fi
}

# Deploy with Docker
deploy_docker() {
    echo ""
    echo "🐳 Deploying with Docker..."

    if [ "$DRY_RUN" != true ]; then
        # Copy docker-compose file
        cp deploy/production/docker-compose.prod.yml $DEPLOY_PATH/docker-compose.yml

        # Create .env for Docker
        cat > $DEPLOY_PATH/.env <<EOF
DOMAIN=$DOMAIN
API_RATE_LIMIT=10
CREATE_SLO_RATE_LIMIT=5
ALLOWED_ORIGINS=https://$DOMAIN
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}
TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}
GRAFANA_PASSWORD=${GRAFANA_PASSWORD}
EOF

        # Start services
        cd $DEPLOY_PATH
        docker-compose -f docker-compose.yml up -d

        echo "✅ Docker services started"
    else
        echo "Would start Docker Compose with domain: $DOMAIN"
    fi
}

# Setup reverse proxy
setup_nginx() {
    echo ""
    echo "🌐 Configuring Nginx..."

    if [ "$DRY_RUN" != true ]; then
        # Copy nginx config
        cp deploy/production/nginx.prod.conf $NGINX_PATH/slozy

        # Create symlink
        ln -sf $NGINX_PATH/slozy $NGINX_PATH/../sites-enabled/slozy

        # Test and reload nginx
        nginx -t && systemctl reload nginx
        echo "✅ Nginx configured"
    else
        echo "Would configure Nginx for domain: $DOMAIN"
    fi
}

# Verify deployment
verify() {
    echo ""
    echo "🔍 Verifying deployment..."

    if [ "$DRY_RUN" != true ]; then
        # Wait for service to start
        sleep 5

        # Check if service is running
        if systemctl is-active --quiet slozy; then
            echo "✅ Service is running"
        else
            echo "❌ Service failed to start"
            journalctl -u slozy --no-pager -n 20
            exit 1
        fi

        # Check health endpoint
        if curl -f http://localhost:8080/health >/dev/null 2>&1; then
            echo "✅ Health check passed"
        else
            echo "❌ Health check failed"
            exit 1
        fi

        # Check security headers
        echo "Checking security headers..."
        headers=$(curl -I -s http://localhost:8080/health)
        if echo "$headers" | grep -q "X-Frame-Options"; then
            echo "✅ Security headers present"
        else
            echo "⚠️  Security headers missing"
        fi
    else
        echo "Would verify deployment"
    fi
}

# Main execution
main() {
    check_prereqs

    if [ "$USE_DOCKER" = true ]; then
        deploy_docker
    else
        build_app
        create_user
        create_dirs
        deploy_files
        configure_systemd
        create_config
        systemctl start slozy
    fi

    setup_nginx
    verify

    echo ""
    echo "🎉 Deployment complete!"
    echo ""
    echo "Next steps:"
    echo "1. Configure your domain DNS to point to this server"
    echo "2. Set up SSL certificates (Let's Encrypt recommended)"
    echo "3. Configure monitoring and alerting"
    echo "4. Set up backups"
    echo ""
    echo "Service management:"
    echo "  Start:   sudo systemctl start slozy"
    echo "  Stop:    sudo systemctl stop slozy"
    echo "  Restart:  sudo systemctl restart slozy"
    echo "  Status:   sudo systemctl status slozy"
    echo "  Logs:     sudo journalctl -u slozy -f"
}

# Run main function
main

echo ""
echo "✅ Deployment script completed"