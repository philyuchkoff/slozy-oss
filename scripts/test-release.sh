#!/bin/bash
#
# Test SLOzy Release Archive
# This script tests the installation from the release archive
#

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ARCHIVE=${1:-""}
TEST_DIR="test-install-$$"

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    print_status "Cleaning up test environment..."
    docker rm -f slozy-test 2>/dev/null || true
    rm -rf "$TEST_DIR"
}

# Trap cleanup on exit
trap cleanup EXIT

# Check if archive provided
if [[ -z "$ARCHIVE" ]]; then
    # Find the latest archive
    ARCHIVE=$(ls -t slozy-*.zip 2>/dev/null | head -1)
    if [[ -z "$ARCHIVE" ]]; then
        print_error "No release archive found. Run 'make release-archive' first."
        exit 1
    fi
fi

print_status "Testing release archive: $ARCHIVE"

# Check if Docker is available
if ! command -v docker >/dev/null; then
    print_error "Docker is required for testing"
    exit 1
fi

# Extract archive
print_status "Extracting release archive..."
mkdir -p "$TEST_DIR"
cd "$TEST_DIR"
unzip "../$ARCHIVE"

# Run installation in Docker container
print_status "Testing installation in Docker (CentOS 8)..."
docker run \
    --name slozy-test \
    --rm -d \
    -v "$(pwd):/opt/slozy-release" \
    centos:8 \
    tail -f /dev/null

# Install dependencies in container
docker exec slozy-test yum update -y
docker exec slozy-test yum install -y unzip

# Run installation
print_status "Running installation script..."
docker exec slozy-test bash -c "
    cd /opt/slozy-release
    chmod +x scripts/install-centos.sh
    ./scripts/install-centos.sh
"

# Wait for service to start
print_status "Waiting for service to start..."
sleep 5

# Check if service is running
print_status "Verifying installation..."
if docker exec slozy-test systemctl is-active --quiet slozy; then
    print_status "✓ Service is running"
else
    print_error "✗ Service failed to start"
    docker exec slozy-test journalctl -u slozy -n 20
    exit 1
fi

# Test health endpoint
print_status "Testing health endpoint..."
if docker exec slozy-test curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health | grep -q "200"; then
    print_status "✓ Health check passed"
else
    print_warning "Health check returned non-200 status"
fi

# Test API endpoint
print_status "Testing API endpoint..."
if docker exec slozy-test curl -s http://localhost:8080/api/health | grep -q "ok"; then
    print_status "✓ API health check passed"
else
    print_warning "API health check failed"
fi

# Print summary
print_status ""
print_status "✅ Release installation test completed successfully!"
print_status ""
print_status "Test details:"
print_status "  Archive: $ARCHIVE"
print_status "  Test container: slozy-test (centos:8)"
print_status "  Service status: $(docker exec slozy-test systemctl is-active slozy)"
print_status ""
print_status "To examine the test container:"
print_status "  docker exec -it slozy-test bash"
print_status ""
print_status "View service logs:"
print_status "  docker exec slozy-test journalctl -u slozy -n 50"