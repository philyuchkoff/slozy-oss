# 🦥 SLOzy - [SLO made lazy and easy](https://slozy.ru/)

A tool for creating and managing Service Level Objectives (SLOs) with comprehensive alerting support.

## ✨ Features

### Core Features

- 🎯 **SLO Creation** - Define service level objectives with targets
- 📊 **Multiple Indicators** - Support for latency, availability, throughput
- 🔔 **Integrated Alerting** - Generate Prometheus rules, AlertManager configs
- 📈 **Grafana Dashboards** - Auto-generated dashboards for monitoring
- 📝 **Documentation** - Automatically generate README files
- 🎨 **Web UI** - Intuitive web interface for SLO management
- 📋 **Templates** - Pre-built templates for common services

### Advanced Features

- 🏷️ **Autocompletion** – Smart hints for metrics, teams, owners
- 📈 **Visualization** – Preview SLO burn rates before saving
- 🎛 **Multi-file Generation** – All configs packaged in ZIP
- 🌐 **CORS Support** – Direct integration with frontend
- 📑 **Template Editor** – Edit existing or create new templates

## 🚀 Quick Start

### Prerequisites

- Go 1.22 or later
- No external dependencies required for development
- systemd and Nginx for production deployment

### Installation

#### Development Installation

```bash
# Clone the repository
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Build the application
make build

# Run in development mode
make run
```

#### Production Deployment

##### 📦 Using Distribution Scripts (Recommended)

```bash
# Create distributable package
./scripts/distribute.sh

# Deploy to remote server
./scripts/deploy.sh -s user@server.com dist/slozy-v1.4.0-abc123-20260429.zip -b -r

# With custom options
./scripts/deploy.sh \
  --server deploy@your-server.com \
  --path /opt/slozy \
  --name slozy \
  --backup \
  --restart \
  dist/slozy-v1.4.0-abc123-20260429.zip
```

##### Manual Deployment

```bash
# Clone the repository
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Build for production
make build-prod

# Install and configure
sudo ./scripts/setup.sh

# Configure environment
sudo nano /opt/slozy/.env

# Start service
sudo systemctl start slozy
```

#### Docker Deployment

```bash
# Build the image
docker build -t slozy .

# Run with persistent storage
docker run -d \
  --name slozy \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  slozy
```

#### Access Web Interface

Open your browser to: <http://localhost:8080> (or your configured host/port)

## 📖 Usage

### 1. Creating SLOs

#### Using Web UI (Recommended)

1. Navigate to <http://localhost:8080>
2. Click "Create SLO"
3. Fill in service information:
   - Service name and description
   - SLO name and description  
   - Metric source and query
   - Target percentage and threshold
   - Team and labels
4. Configure alerting:
   - Enable alerting
   - Set notification channels (Slack, email, PagerDuty, Telegram)
   - Configure burn rate thresholds
   - Set severity levels
5. Click "Generate & Download"
6. Receive ZIP archive with:
   - OpenSLO YAML file
   - Prometheus alert rules (optional)
   - AlertManager config (optional)
   - Grafana dashboard JSON (optional)
   - README with instructions

#### Using API

```bash
curl -X POST http://localhost:8080/api/slo/create \
  -H "Content-Type: application/json" \
  -d @slo-config.json \
  -o slo-files.zip
```

### 2. Managing Templates

#### Using Built-in Templates

SLOzy includes pre-built templates:

- **Web API** – Latency and availability for HTTP services
- **Database** – Query latency and availability
- **Kubernetes** – Infrastructure SLOs
- **Message Queue** – Throughput and latency
- **Load Balancer** – Request latency
- **Cache System** – Response time
- **Object Storage** – Availability and latency
- **Mobile App** – Crash-free users and performance
- **CI/CD Pipeline** – Build success rate
- **IoT Devices** – Device availability
- **Monitoring Systems** – Service uptime

#### Creating Custom Templates

1. Navigate to <http://localhost:8080>
2. Click "Templates" tab
3. Click "Create Template"
4. Fill in template details
5. Save template - automatically available

### 3. Advanced Configuration

#### Environment Variables

```bash
# Server Configuration
export PORT=8080                         # Server port
export DATA_DIR=./data                       # Data directory
export TEMPLATES_DIR=./templates              # Templates directory
export STATIC_DIR=./static                      # Static files directory

# Alerting Configuration
export TELEGRAM_BOT_TOKEN=your_bot_token     # Telegram bot token
export TELEGRAM_CHAT_ID=your_chat_id        # Telegram chat ID
export SMTP_HOST=localhost:25                  # SMTP server
export SMTP_FROM=slo-alerts@example.com         # Email from address
export GRAFANA_URL=https://grafana.example.com    # Grafana base URL
```

#### File Structure

```
slozy/
├── cmd/slozy-server/          # Application source
│   ├── internal/              # Internal packages
│   │   ├── config/          # Configuration management
│   │   ├── generators/       # File generators
│   │   ├── handlers/        # HTTP handlers
│   │   ├── middleware/       # HTTP middleware
│   │   └── models/          # Data models
│   └── main.go               # Application entry
├── static/                   # Web UI assets
├── templates/                # SLO templates
├── data/                     # Generated SLO files
├── bin/                      # Build output
├── Makefile                  # Build commands
└── README.md                 # This file
```

## 🧰 Development

### Running in Development Mode

```bash
make dev
```

Requires `air` for hot reload (`go install github.com/air-verse/air@latest`).

### Available Commands

```bash
make build      # Build application
make run        # Run application
make dev        # Run in development mode
make test       # Run tests
make clean      # Clean build artifacts
make install    # Install system-wide
make deps       # Download dependencies
make lint       # Run linter
docker-build   # Build Docker image
docker-run     # Run in Docker container
make web        # Open web interface
```

## 🐳 Docker Support

```bash
# Build image
docker build -t slozy:latest .

# Run container
docker run -p 8080:8080 \
  -v $(PWD)/data:/app/data \
  slozy:latest
```

Or use docker-compose:

```bash
docker-compose up
```

## 📊 Monitoring SLOs

Once SLOs are created:

### 1. OpenSLO Files

The generated YAML files can be used with any OpenSLO-compatible tool.

### 2. Prometheus Rules

Import the generated alert rules into your Prometheus server:

```bash
curl -X POST http://localhost:9090/api/v1/rules \
  -H "Content-Type: application/yaml" \
  --data-binary @service-slo-prometheus-alerts.yaml
```

### 3. Grafana Dashboards

Import the JSON dashboard files:

1. Navigate to Grafana
2. Click "+" → "Import"
3. Upload the dashboard JSON file
4. Select data source
5. Save dashboard

### 4. AlertManager Configuration

The generated configuration includes:

- Routes for different severity levels
- Integration with Slack, email, PagerDuty, Telegram
- Inhibition rules to prevent alert spam

## 📋 Integration Examples

### Kubernetes

```yaml
apiVersion: v1
kind: Deployment
metadata:
  name: slozy
spec:
  replicas: 1
  selector:
    matchLabels:
      app: slozy
  template:
    metadata:
      labels:
        app: slozy
    spec:
      containers:
      - name: slozy
        image: slozy:latest
        ports:
        - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: slozy-service
spec:
  selector:
    app: slozy
  ports:
    - port: 80
      targetPort: 8080
```

### Terraform

```hcl
resource "docker_image" "slozy" {
  name = "slozy"
  build_dir = "${path.module}"
}

resource "docker_container" "slozy" {
  image = docker_image.slozy.latest
  name  = "slozy"
  ports {
    internal = 8080
    external = 8080
  }
}
```

## 🔧 Troubleshooting

### Build Issues

```bash
# Clean build
make clean && make build

# Update dependencies
go mod tidy

# Verify Go version
go version  # Should be 1.21+
```

### Runtime Issues

```bash
# Check port usage
lsof -i :8080

# Use different port
PORT=9090 ./bin/slozy

# Check logs
./bin/slozy 2>&1 | tee slozy.log

# Verify data directory permissions
ls -la data/
```

### Permission Issues

```bash
# macOS/Linux
chmod +x ./bin/slozy

# Windows
icacls slozy.exe /grant Everyone:F
```

## 🏢 Architecture

The project follows Go project layout:

- **cmd/** - Application entry points
- **internal/** - Private application code
- **static/** - Static assets
- **templates/** - SLO templates
- **examples/** - Usage examples

### Key Components

1. **Generators** (cmd/slozy-server/internal/generators/)
   - `openslo.go` - OpenSLO YAML generation
   - `prometheus.go` - Prometheus alert rules
   - `alertmanager.go` - Alertmanager configuration
   - `grafana.go` - Grafana dashboard JSON
   - `readme.go` - Documentation generation

2. **Handlers** (cmd/slozy-server/internal/handlers/)
   - `slo.go` - SLO CRUD operations
   - `template.go` - Template management

3. **Models** (cmd/slozy-server/internal/models/)
   - `slo.go` - Data structures
   - `validator.go` - Input validation

4. **Middleware** (cmd/slozy-server/internal/middleware/)
   - `logging.go` - Request/response logging

## 📄 OpenSLO Compatibility

SLOzy generates fully compliant OpenSLO v1alpha files:

- API version: openslo/v1alpha
- SLO object with metadata and spec
- Time windows with rolling/budgeting support
- Indicator configuration based on type
- Alert policies with burn rate conditions

## 📦 Distribution and Deployment

### Automated Scripts

For easy production deployment:

```bash
# Create distribution package
./scripts/distribute.sh

# Deploy to remote server  
./scripts/deploy.sh -s user@server.com dist/slozy-*.zip -b -r

# Full installation on clean server
sudo ./scripts/setup.sh full
```

See [scripts/README.md](scripts/README.md) for detailed instructions.

### Production Features

- ✅ Graceful shutdown with 30s timeout
- ✅ Configurable rate limiting (5-10 req/s)
- ✅ Structured JSON logging
- ✅ Health monitoring endpoint
- ✅ Security headers and CORS
- ✅ Prometheus/Grafana integration
- ✅ Systemd service management

### Configuration

```bash
# Environment variables
PORT=8080                         # Server port
DATA_DIR=/opt/slozy/data           # Data directory
TEMPLATES_DIR=/opt/slozy/templates   # Templates
STATIC_DIR=/opt/slozy/static         # Static files

# Alerting (optional)  
TELEGRAM_BOT_TOKEN=...              # Telegram bot token
TELEGRAM_CHAT_ID=...               # Chat ID
LOG_LEVEL=info                      # Logging level
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes
4. Add tests if applicable
5. Run linter: `make lint`
6. Run tests: `make test`
7. Submit pull request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
## 🙏 Acknowledgments

Built using Go with standard libraries only.
OpenSLO specification by the OpenSLO community.
Icons and visual design credit to original authors.

---

**Made with ❤️**
