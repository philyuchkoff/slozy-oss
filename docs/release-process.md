# Release Process

This document describes the process for creating and deploying SLOzy releases.

## Overview

The release process includes:
1. Building cross-platform binaries
2. Creating a deployment archive with all necessary files
3. Automating installation on CentOS/RHEL systems
4. Verifying the installation

## Creating a Release

### Prerequisites

- Go 1.22 or later
- Git
- Make
- For testing: Docker (optional)

### Step 1: Tag the Release

```bash
# Create and push a new tag
git tag -a v1.2.2 -m "Release v1.2.2"
git push origin v1.2.2
```

### Step 2: Build the Release Archive

```bash
# Build the release archive
make release-archive

# This will create: slozy-v1.2.2.zip
```

### Step 3: Test the Release (Optional)

```bash
# Test the release in a Docker container
make test-release
```

### Step 4: Manual Upload (if not using GitHub Actions)

If you're not using GitHub Actions, upload the archive manually:

```bash
# Upload to GitHub Releases (requires gh CLI)
gh release create v1.2.2 slozy-v1.2.2.zip \
  --title "SLOzy v1.2.2" \
  --notes "Release notes..."
```

## Automated Releases

### GitHub Actions

The project includes a GitHub Actions workflow that automatically:
1. Triggers on tag pushes (v*)
2. Runs tests
3. Builds the release archive
4. Creates a GitHub release with the archive

No manual steps required - just push a tag!

### Local Development

For local development and testing:

```bash
# Build with specific version
./build-release.sh v1.2.2-beta

# Clean release files
make clean-release
```

## Deployment Guide for CentOS/RHEL

### Quick Installation

```bash
# Download the latest release
wget https://github.com/philyuchkoff/slozy/releases/latest/download/slozy-vX.Y.Z.zip

# Extract and install
unzip slozy-vX.Y.Z.zip
cd release-vX.Y.Z
sudo ./scripts/install-centos.sh
```

### What Gets Installed

The installation script:
1. Creates a service user: `slozy`
2. Installs binaries to: `/opt/slozy/bin/`
3. Stores assets in: `/opt/slozy/`
4. Configures systemd service: `slozy.service`
5. Sets up log rotation

### Post-Installation

1. **Configure the service**:
   ```bash
   sudo nano /opt/slozy/.env
   ```

2. **Start the service**:
   ```bash
   sudo systemctl start slozy
   ```

3. **Check status**:
   ```bash
   sudo systemctl status slozy
   ```

4. **View logs**:
   ```bash
   sudo journalctl -u slozy -f
   ```

## Release Archive Contents

The release archive (`slozy-VERSION.zip`) contains:

```
release-VERSION/
├── slozy-server-linux-amd64    # Binary for Intel/AMD
├── slozy-server-linux-arm64    # Binary for ARM
├── static/                     # Web UI assets
├── templates/                  # SLO templates
├── config/
│   ├── .env.example           # Environment template
│   └── slozy.service          # systemd service file
├── scripts/
│   └── install-centos.sh      # Installation script
├── VERSION.txt                # Version info
└── README.md                  # Installation guide
```

## Rolling Back

If a deployment fails:

1. **Check the backup location** (if it exists):
   ```bash
   ls -la /opt/slozy-backup-*
   ```

2. **Stop the service**:
   ```bash
   sudo systemctl stop slozy
   ```

3. **Restore from backup** (if available):
   ```bash
   sudo systemctl stop slozy
   sudo rm -rf /opt/slozy
   sudo mv /opt/slozy-backup-YYYYMMDD-HHMMSS /opt/slozy
   sudo systemctl start slozy
   ```

4. **Or install previous version**:
   ```bash
   # Download previous release
   wget https://github.com/philyuchkoff/slozy/releases/download/VERSION/slozy-VERSION.zip
   # Extract and run install script
   ```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u slozy -n 50

# Check configuration
sudo -u slozy /opt/slozy/bin/slozy-server --check-config
```

### Port Conflicts

Edit `/opt/slozy/.env`:
```bash
PORT=8081  # Change port
```

Then restart:
```bash
sudo systemctl restart slozy
```

### Permission Issues

```bash
# Fix ownership
sudo chown -R slozy:slozy /opt/slozy

# Fix permissions
sudo chmod -R 755 /opt/slozy
sudo chmod +x /opt/slozy/bin/slozy-server
```

## Security Considerations

- The service runs as non-root user `slozy`
- Write permissions are limited to data and logs directories
- systemd service includes security restrictions:
  - `NoNewPrivileges=true`
  - `ProtectSystem=strict`
  - `PrivateTmp=true`

## Next Steps

After deployment:

1. Configure notifications in `/opt/slozy/.env`
2. Set up reverse proxy (nginx or Apache)
3. Configure monitoring
4. Enable alerting

For more details, see the [Deployment Guide](deployment.md).