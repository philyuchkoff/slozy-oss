.PHONY: build run test clean deps lint docker-build docker-run

BINARY_NAME=slozy
VERSION=0.1.0

build:
	@echo "Building SLOzy..."
	go build -o bin/$(BINARY_NAME) ./cmd/slozy-server

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
