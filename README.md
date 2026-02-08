# 🦥 SLOzy - [SLO made lazy and easy](https://slozy.ru/)

A tool for creating SLOs (Service Level Objectives) in OpenSLO format with full alerting support.

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.19%2B-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/OpenSLO-compatible-4A154B?style=for-the-badge" alt="OpenSLO">
  <img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="License">
</p>

## What is this?

SLOzy is a web tool that allows development and product teams to create SLOs without needing to understand the complexities of the [OpenSLO](https://github.com/OpenSLO/OpenSLO) format.

Input data is validated; if something is wrong, you'll get a hint.

Just fill out the form, get a ready-made YAML file, and put it where the SREs tell you to—they'll handle the rest.

For the toughest cases, you can click "Fill example" and all fields will be filled with hints.

## ✨ Features

### Core Features
- 🎯 **Simple Web Interface** – fill out the form, get an SLO
- 📚 **Templates** – ready-made templates for different service types (Web API, databases, queues, etc.; you can add and edit them for any product; examples are in `/templates`)
- 🔧 **Automatic Generation** – from a simple description to valid OpenSLO YAML
- 🏷️ **Autocompletion** – hints for metrics, teams, owners
- 📊 **Preview** – visualize the SLO before saving

### 🔔 Full Alerting Support:
- **OpenSLO Alert Policies** – integrated into the main SLO YAML
- **Prometheus Alert Rules** – a separate YAML file with rules
- **AlertManager Config** – example configuration for alert routing
- **Grafana Dashboards** – two ready-made dashboards for visualization
- **README File** – deployment instructions
- **Burn Rate Thresholds** – configurable thresholds with different severity levels
- **Multiple Channels** – Slack, Email, PagerDuty, webhook

### 📦 ZIP
- **All in One ZIP** – all configuration files are packaged into a single archive
- **Unique Names** – archives contain a timestamp for uniqueness

### 🎨 Interface Features:
- "Enable Alerting" checkbox with advanced configuration
- Burn Rate Thresholds configuration (warning/critical)
- Notification channels specification
- Templates with pre-configured alerting
- View generated files
- Query examples for different metric types

### ✅ Technical Improvements:
- Number formatting without floating-point errors
- Correct PromQL expression generation
- Alert Inhibition to prevent alert storms (an Alertmanager mechanism that prevents sending less important alerts when a more important alert is already firing for the same object)
- Detailed annotations and labels
- Input data validation

## 🚀 Quick Start

### Requirements
- Go 1.19 or higher
- Web browser with JavaScript support

### Installation and Running
```bash
# Clone the repository
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Start the server on port 8080
go run main.go

# Or run on a different port
PORT=9090 go run main.go
```

The application will be available at: [http://localhost:8080](http://localhost:8080)

## 🎯 Usage

### 1. Filling Out the Form
1. Open [http://localhost:8080](http://localhost:8080)
2. Enter basic service information:
   - Service name
   - SLO name
   - Team
   - Description

### 2. Configuring Metrics
1. Choose the indicator type (Latency, Availability, etc.)
2. Specify the target value (e.g., 99.9%)
3. Set the threshold value (e.g., 200ms for latency)
4. Specify the error budget period (default 28 days)

### 3. Metric Configuration
1. Choose the metric source (Prometheus, Datadog, CloudWatch, etc.)
2. Enter the metric query
3. Use ready-made examples or create your own query

### 4. Configuring Alerting (Optional)
1. Enable alerting
2. Configure notification channels (Slack, Email, PagerDuty)
3. Add burn rate thresholds:
   - Warning: 5-10x faster over 1 hour
   - Critical: 50-100x faster over 5 minutes

### 5. Generation and Download
1. Click "Generate SLO with Alerts"
2. All configuration files will be packaged into a ZIP archive
3. The archive will be downloaded automatically

## 📄 Generated Files

The following files will be created in the ZIP archive:

```
service-name-slo-name-YYYYMMDD-HHMMSS.zip
├── service-name-slo-name.yaml            # Main SLO in OpenSLO format
├── service-name-slo-name-prometheus-alerts.yaml     # Prometheus alert rules
├── service-name-slo-name-alertmanager-config.yaml   # Alertmanager configuration
├── service-name-slo-name-grafana-dashboard.json     # Grafana dashboard
├── service-name-slo-name-multi-burnrate-dashboard.json  # Extended dashboard
└── service-name-slo-name-README.md                   # Documentation
```

## 📋 Supported SLO Types

- **Latency** – service response time
- **Availability** – service uptime
- **Throughput** – request count
- **Custom metrics**

## 📚 SLO Templates

SLOzy includes ready-made templates for:

### 1. Web API Latency
- Target: 99.9%
- Threshold: 200ms
- Alerting: Warning (10x/1h), Critical (50x/5m)

### 2. Web API Availability  
- Target: 99.95%
- Threshold: 0.0005
- Alerting: Warning (5x/1h)

### 3. Database Query Latency
- Target: 99.9%
- Threshold: 100ms
- Alerting: Critical (20x/30m)

### 4. and more
See `/templates` directory

### 5. You can add your own templates

## 🔧 API Endpoints

```bash
# Create SLO via curl (returns a ZIP archive)
curl -X POST http://localhost:8080/api/slo/create \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "payment-service",
    "sloName": "checkout-availability",
    "indicatorType": "availability",
    "target": 99.95,
    "threshold": "0.0005",
    "metricSource": "prometheus",
    "metricQuery": "sum(rate(http_requests_total{job=\"payment\", status=~\"2..\"}[5m])) / sum(rate(http_requests_total{job=\"payment\"}[5m]))",
    "team": "backend",
    "alerting": {
      "enabled": true,
      "notificationChannels": ["slack#slo-alerts"],
      "burnRateThresholds": [
        {
          "rate": 5.0,
          "period": "1h",
          "severity": "warning",
          "alertName": "HighBurnRate"
        }
      ]
    }
  }' --output slo-configuration.zip

# Get list of SLOs
curl http://localhost:8080/api/slos

# Get templates
curl http://localhost:8080/api/templates

# Health check
curl http://localhost:8080/health
```

Available API endpoints:
- `GET /` – Web interface
- `POST /api/slo/create` – Create SLO (returns ZIP archive)
- `GET /api/slos` – List of created SLOs
- `GET /api/templates` – SLO templates
- `GET /health` – Health check

## 🐳 Docker

```bash
# Build image
docker build -t slozy .

# Run container
docker run -p 8080:8080 -v $(pwd)/data:/app/data slozy
```

## 🛠️ Development

### Production Build

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o slozy-linux-amd64 main.go

# Build for macOS
GOOS=darwin GOARCH=arm64 go build -o slozy-macos-arm64 main.go

# Reduce binary size
upx --best slozy-linux-amd64
```

### Environment Variables

```bash
# Server port (default 8080)
export PORT=8080

# Data directory (default ./data)
export DATA_DIR=/var/lib/slozy

# Log level (debug, info, warn, error)
export LOG_LEVEL=info
```

## 📊 Technologies

- **Backend**: Go (pure net/http, no frameworks)
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Formats**: YAML, JSON
- **Standards**: OpenSLO v1alpha
- Prometheus, Grafana, Alertmanager compatibility

## 🔒 Security

### Recommendations for Production Use
1. Use HTTPS (recommend nginx or traefik as a reverse proxy)
2. Restrict API access
3. Do not store sensitive data in the `data/` directory
4. Use environment variables for configuration

### Limitations
- The application is intended for internal use
- Authentication/authorization is not implemented
- Generated files may contain template data (replace with your own before use)

### Useful Tips
- Adapt metrics – replace example queries with your real metrics
- Configure alerting – specify the correct notification channels for your team
- Check thresholds – adjust burn rate thresholds for your service
- Test – test alerts in a test environment before applying to production

## Telegram Notifications

SLOzy supports sending notifications to Telegram. To set up:

### Quick Setup:
1. Create a bot via @BotFather
2. Get the bot token
3. Get the chat ID (personal, group, or channel)
4. Set environment variables:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_token"
   export TELEGRAM_CHAT_ID="your_chat_id"
5. In the SLO creation form, specify: `telegram:your_chat_id`
6. Detailed instructions in [examples/telegram-notifications/README.md](./examples/telegram-notifications/README.md)

## 📈 Roadmap

- [ ] Import existing SLOs (e.g., YAML from Terraform or JSON from API)
- [ ] Simplify terminology in forms
- [ ] Git integration (auto-commit configurations)
- [ ] Clearer connection between templates and parameters ("highlighting" fields prefilled from the template)
- [x] Telegram notification sending
- [ ] SLO review scheduler
- [ ] SLO change history
- [ ] Multi-tenancy support
- [ ] API keys for automation
- [ ] Preview generated YAML files instead of immediate download
- [ ] Tooltips for complex fields, "micro-hints", "examples"
- [ ] Ability to share pre-filled SLO configuration
- [ ] Visual separation: cards, indentation, icons for each block
- [ ] Save drafts to localStorage
- [ ] Drag-and-drop for prioritizing Existing SLOs list

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/useful-feature`)
3. Commit your changes (`git commit -m 'Add useful feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 Examples

Examples of ready-made SLO files are in the `examples/` directory.

## 🐛 Bug Reports

Found a bug? Please create an issue with detailed description:
1. Steps to reproduce
2. Expected behavior
3. Actual behavior
4. Logs (if any)

## 📄 License

This project is distributed under the MIT license. See the [LICENSE](LICENSE) file for details.

## 📚 Useful Links

- [OpenSLO Specification](https://openslo.com/)
- [SRE for Non-Engineers](https://srebook.ru/)
