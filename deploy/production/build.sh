#!/bin/bash

# SLOzy Production Build Script
# This script builds optimized production binaries for multiple platforms

set -e

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "v1.0.0")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS="-w -s -X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE"

echo "Building SLOzy Production Binaries"
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "=================================="

# Create output directory
mkdir -p dist

# Build for multiple architectures
platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r GOOS GOARCH <<< "$platform"

    echo -n "Building for $GOOS/$GOARCH... "

    if [ "$GOOS" = "windows" ]; then
        OUT="slozy-server.exe"
    else
        OUT="slozy-server"
    fi

    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
        go build -ldflags="$LDFLAGS" -o "dist/$OUT.$GOOS.$GOARCH" .

    echo "✓"

    # Create tarball
    cd dist
    TAR_OUT="slozy-${VERSION}-${GOOS}-${GOARCH}.tar.gz"
    tar -czf "$TAR_OUT" "$OUT.$GOOS.$GOARCH"
    rm "$OUT.$GOOS.$GOARCH"
    cd ..
    echo "  → $TAR_OUT created"
done

echo ""
echo "Build complete! Files in dist/:"
ls -lh dist/*.tar.gz

# Generate checksums
echo ""
echo "Generating checksums..."
cd dist
sha256sum *.tar.gz > checksums.txt
echo "✓ checksums.txt created"
cd ..

# Create a production configuration template
cat > dist/.env.production <<EOF
# SLOzy Production Configuration Template
# Copy and modify these values for your deployment

# Server Configuration
PORT=8080
ENVIRONMENT=production
VERSION=$VERSION

# TLS/SSL Configuration (uncomment for HTTPS)
# TLS_CERT_PATH=/path/to/cert.pem
# TLS_KEY_PATH=/path/to/key.pem

# Rate Limiting (requests per second)
API_RATE_LIMIT=10
CREATE_SLO_RATE_LIMIT=5

# CORS Origins (comma-separated, NO WILDCARDS in production!)
ALLOWED_ORIGINS=https://yourdomain.com

# Telegram Alerts (optional but recommended)
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
EOF

echo ""
echo "✅ Production build complete!"
echo "📦 Deploy: extract tarball for your platform"
echo "⚙️  Configure: copy .env.production to .env and customize"
echo "🔐 Secure: run behind HTTPS reverse proxy"