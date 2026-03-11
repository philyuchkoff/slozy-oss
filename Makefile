.PHONY: build run dev test clean deps lint docker-build docker-run install web generate-sample build-prod deploy-prod setup-prod release-archive clean-release

BINARY_NAME=slozy-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-X main.Version=$(VERSION) -s -w"

# Build targets
build:
	@echo "Building SLOzy..."
	@mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/slozy-server

build-prod:
	@echo "Building SLOzy for production..."
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/slozy-server
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 ./cmd/slozy-server
	@echo "✅ Production builds created:"
	@ls -lh bin/$(BINARY_NAME)-linux-*

run:
	@echo "Starting SLOzy..."
	go run ./cmd/slozy-server

dev:
	@echo "Starting SLOzy in development mode..."
	air -c .air.toml

test:
	@echo "Running tests..."
	go test ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/ data/

deps:
	@echo "Downloading dependencies..."
	go mod download

lint:
	@echo "Linting..."
	golangci-lint run

docker-build:
	@echo "Building Docker image..."
	docker build -t slozy:$(VERSION) .

docker-run:
	@echo "Running SLOzy in Docker..."
	docker run -p 8080:8080 -v $(PWD)/data:/app/data slozy:$(VERSION)

install:
	@echo "Installing SLOzy..."
	go install ./cmd/slozy-server

web:
	@echo "Starting web interface..."
	open http://localhost:8080

generate-sample:
	@echo "Generating sample SLO..."
	curl -X POST http://localhost:8080/api/slo/create \
		-H "Content-Type: application/json" \
		-d '{
			"serviceName": "user-api",
			"sloName": "api-latency",
			"indicatorType": "latency",
			"target": 99.9,
			"threshold": "200ms",
			"metricSource": "prometheus",
			"metricQuery": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"user-api\"}[5m]))"
		}'

# Production deployment targets
setup-prod:
	@echo "Setting up production environment..."
	@echo "Creating directories..."
	sudo mkdir -p /opt/slozy/{bin,data,templates,static,logs}
	sudo groupadd -f slozy 2>/dev/null || true
	sudo useradd -r -s /bin/false -g slozy -d /opt/slozy slozy 2>/dev/null || true
	sudo chown -R slozy:slozy /opt/slozy
	@echo "✅ Production directories created"

deploy-prod: setup-prod build-prod
	@echo "Deploying SLOzy to production..."
	sudo cp bin/$(BINARY_NAME)-linux-amd64 /opt/slozy/bin/slozy-server
	sudo chmod +x /opt/slozy/bin/slozy-server
	sudo -u slozy cp -r static/* /opt/slozy/static/
	sudo -u slozy cp -r templates/* /opt/slozy/templates/
	sudo cp systemd/slozy.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable slozy
	@echo "✅ SLOzy deployed successfully!"
	@echo "Run 'systemctl start slozy' to start the service"

service-install:
	@echo "Installing systemd service..."
	sudo cp systemd/slozy.service /etc/systemd/system/
	sudo systemctl daemon-reload
	sudo systemctl enable slozy
	@echo "✅ Service installed"

service-start:
	sudo systemctl start slozy
	sudo systemctl status slozy

service-stop:
	sudo systemctl stop slozy

service-logs:
	sudo journalctl -u slozy -f

service-restart:
	sudo systemctl restart slozy

# Release management
release-archive:
	@echo "Creating release archive..."
	@./build-release.sh $(VERSION)
	@echo "✅ Release archive created: slozy-$(VERSION).zip"

test-release:
	@echo "Testing release archive..."
	@./scripts/test-release.sh slozy-$(VERSION).zip

clean-release:
	@echo "Cleaning release files..."
	@rm -f slozy-*.zip
	@rm -rf release-*
	@rm -rf test-install-*
