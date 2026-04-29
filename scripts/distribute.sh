#!/bin/bash
#
# SLOzy Distribution Builder
# Creates distributable packages with proper structure and documentation
#

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "\${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}" && pwd)"
readonly DIST_DIR="${PROJECT_ROOT}/dist"

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() { echo -e "${BLUE}ℹ️  $*${NC}"; }
log_success() { echo -e "${GREEN}✅ $*${NC}"; }
log_warning() { echo -e "${YELLOW}⚠️  $*${NC}"; }
log_error() { echo -e "${RED}❌ $*${NC}" >&2; }

# Configuration
readonly EXCLUDE_DIRS=(
    '.git'
    '.vscode'
    'cmd/slozy-server/data'
    'cmd/slozy-server/internal/storage'
    'bin'
    'data'
    'dist'
)

readonly EXCLUDE_FILES=(
    'distribute.sh'
    '*.log'
    '.DS_Store'
    '*.swp'
    '*.tmp'
)

readonly ESSENTIAL_FILES=(
    'LICENSE'
    'go.mod'
    'go.sum'
    'Makefile'
    '.env.example'
    'docker-compose.yml'
    'Dockerfile'
)

# Functions
cleanup() {
    local exit_code=$?
    if [[ $exit_code -ne 0 ]]; then
        log_error "Script failed with exit code $exit_code"
        # Don't delete archive on failure, only partial files
        if [[ -d "$DIST_DIR/slozy" ]]; then
            log_info "Cleaning up partial distribution..."
            rm -rf "$DIST_DIR/slozy"
        fi
    fi
    exit $exit_code
}

get_version_info() {
    local version
    local commit

    version=$(git describe --tags --abbrev=0 2>/dev/null || echo "dev")
    commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    echo "$version-$commit-$(date +%Y%m%d)"
}

create_rsync_excludes() {
    local excludes=()

    for dir in "${EXCLUDE_DIRS[@]}"; do
        excludes+=("--exclude=$dir")
    done

    for file in "${EXCLUDE_FILES[@]}"; do
        excludes+=("--exclude=$file")
    done

    printf '%s\n' "${excludes[@]}"
}

fix_directory_nesting() {
    local target_dir="$1"
    local nested_dir="$target_dir/cmd/slozy-server/cmd/slozy-server"

    if [[ -d "$nested_dir" ]]; then
        log_warning "Fixing double directory nesting..."
        mkdir -p "$target_dir/cmd/slozy-server"
        mv "$nested_dir"/* "$target_dir/cmd/slozy-server/"
        rmdir "$nested_dir"
        rmdir "$target_dir/cmd/slozy-server/cmd" 2>/dev/null || true
    fi
}

create_distribution_readme() {
    local name="$1"
    local build_date="$2"
    local git_commit="$3"

    cat > "$DIST_DIR/slozy/README.md" << EOF
# SLOzy - Service Level Objective Management Tool

## Quick Start

### Prerequisites
- Go 1.21 or later
- No external dependencies required

### Installation

1. **Extract Archive**
   \`\`\`bash
   unzip ${name}
   cd slozy
   \`\`\`

2. **Build**
   \`\`\`bash
   make build
   \`\`\`

3. **Run**
   \`\`\`bash
   make run
   # or
   ./bin/slozy
   \`\`\`

4. **Access Web UI**
   Open http://localhost:8080

## Features
- 🎯 SLO Creation with targets
- 📊 Multiple indicators (latency, availability, throughput)
- 🔔 Integrated alerting (Prometheus, AlertManager, Grafana)
- 📈 Auto-generated dashboards
- 🎨 Web-based interface
- 📋 Pre-built templates

## Directory Structure
\`\`\`
slozy/
├── cmd/slozy-server/          # Source code
│   ├── internal/              # Packages
│   │   ├── config/        # Configuration
│   │   ├── generators/    # File generators
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/    # HTTP middleware
│   │   └── models/       # Data models
│   └── main.go               # Entry point
├── static/                   # Web UI
├── templates/                # SLO templates
├── data/                     # Generated SLOs (auto-created)
├── bin/                      # Build output (auto-created)
├── Makefile                  # Build commands
└── README.md                 # This file
\`\`\`

## Environment Variables
\`\`\`bash
export PORT=8080                # Server port
export DATA_DIR=./data            # Data directory
export TEMPLATES_DIR=./templates    # Templates directory
export STATIC_DIR=./static           # Static files

# Alerting (optional)
export TELEGRAM_BOT_TOKEN=...   # Bot token
export TELEGRAM_CHAT_ID=...       # Chat ID
export SMTP_HOST=localhost:25       # SMTP server
export SMTP_FROM=slo-alerts@example.com # From email
\`\`\`

## Commands
\`\`\`bash
make build      # Build application
make run        # Run application
make clean      # Clean artifacts
make install    # Install system-wide
\`\`\`

## Support
- 📖 Documentation: https://github.com/philyuchkoff/slozy
- 🐛 Issues: https://github.com/philyuchkoff/slozy/issues
- 📧 Email: support@slozy.dev
- 📜 License: MIT

---
Built on ${build_date} from Git: ${git_commit}
EOF
}

display_archive_contents() {
    local archive="$1"

    log_info "Archive contents:"
    unzip -l "$archive" 2>/dev/null | head -20 || true
}

# Main execution
main() {
    trap cleanup EXIT

    # Ensure dist directory exists
    mkdir -p "$DIST_DIR"

    log_info "Building SLOzy distribution..."

    # Change to project root
    cd "$PROJECT_ROOT"

    # Get version information
    local name="slozy-$(get_version_info)"
    local build_date
    local git_commit

    build_date=$(date +%Y-%m-%d)
    git_commit=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    log_info "Creating ${name}..."

    # Clean and create distribution directory (keep archive if exists)
    rm -rf "$DIST_DIR/slozy"
    mkdir -p "$DIST_DIR"

    # Copy source code with proper excludes
    log_info "Copying source code..."
    rsync -av \
        --exclude=.git \
        --exclude=.vscode \
        --exclude='cmd/slozy-server/data' \
        --exclude='cmd/slozy-server/internal/storage' \
        --exclude=bin \
        --exclude=data \
        --exclude=dist \
        --exclude=distribute.sh \
        --exclude='*.log' \
        --exclude=.DS_Store \
        --exclude='*.swp' \
        --exclude='*.tmp' \
        cmd/ "$DIST_DIR/slozy/"

    # Fix directory structure if needed
    fix_directory_nesting "$DIST_DIR/slozy"

    # Copy essential files
    log_info "Copying essential files..."
    cp "${ESSENTIAL_FILES[@]}" "$DIST_DIR/slozy/" 2>/dev/null || true

    # Copy systemd service if exists
    if [[ -f "systemd/slozy.service" ]]; then
        cp "systemd/slozy.service" "$DIST_DIR/slozy/"
    fi

    # Copy assets if they exist
    if [[ -d "static" ]]; then
        log_info "Copying static assets..."
        cp -r static "$DIST_DIR/slozy/"
    fi

    if [[ -d "templates" ]]; then
        log_info "Copying templates..."
        cp -r templates "$DIST_DIR/slozy/"
    fi

    # Copy systemd service
    if [[ -f "systemd/slozy.service" ]]; then
        log_info "Copying systemd service..."
        cp "systemd/slozy.service" "$DIST_DIR/slozy/"
    fi

    # Create runtime directories
    mkdir -p "$DIST_DIR/slozy/data" "$DIST_DIR/slozy/bin"

    # Create README for distribution
    log_info "Generating distribution README..."
    create_distribution_readme "$name" "$build_date" "$git_commit"

    # Create archive
    log_info "Creating distribution archive..."
    cd "$DIST_DIR"
    zip -r "${name}.zip" slozy/

    # Display results
    local archive_name="${name}"
    local archive_path="$DIST_DIR/$archive_name.zip"

    # Wait a moment for archive to be fully written
    sleep 1

    if [[ ! -f "$archive_path" ]]; then
        log_error "Archive not found at $archive_path"
        exit 1
    fi

    local size
    size=$(du -h "$archive_path" | cut -f1)

    echo
    log_success "Distribution created: ${archive_path}"
    log_info "Archive size: $size"
    echo
    display_archive_contents "$archive_path"
    echo
    log_success "Ready to distribute!"
    echo
    log_info "To test the distribution:"
    echo "  cd /tmp"
    echo "  unzip ${archive_name}"
    echo "  cd slozy"
    echo "  make build"
    echo "  make run"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -h, --help     Show this help message"
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            exit 1
            ;;
    esac
    shift
done

# Run main function
main "$@"