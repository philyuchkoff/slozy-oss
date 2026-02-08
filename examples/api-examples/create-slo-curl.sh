#!/bin/bash

# Пример создания SLO через curl с алертингом

BASE_URL="http://localhost:8080"
OUTPUT_FILE="slo-configuration.zip"

echo "Создание SLO для веб-API (latency)..."
curl -X POST "${BASE_URL}/api/slo/create" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "user-api",
    "sloName": "api-latency",
    "sloDescription": "99th percentile latency for API endpoints",
    "indicatorType": "latency",
    "metricSource": "prometheus",
    "metricQuery": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"user-api\"}[5m]))",
    "target": 99.9,
    "threshold": "200ms",
    "budgetingPeriod": "28d",
    "team": "backend",
    "alerting": {
      "enabled": true,
      "notificationChannels": ["slack#slo-alerts", "email:sre@example.com"],
      "burnRateThresholds": [
        {
          "rate": 10.0,
          "period": "1h",
          "severity": "warning",
          "alertName": "HighBurnRate"
        },
        {
          "rate": 50.0,
          "period": "5m",
          "severity": "critical",
          "alertName": "CriticalBurnRate"
        }
      ]
    }
  }' \
  --output "user-api-latency.zip"

echo ""
echo "Создание SLO для платежного сервиса (доступность)..."

curl -X POST "${BASE_URL}/api/slo/create" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "payment-service",
    "sloName": "checkout-availability",
    "sloDescription": "Checkout API availability for successful transactions",
    "indicatorType": "availability",
    "metricSource": "prometheus",
    "metricQuery": "sum(rate(http_requests_total{job=\"payment-service\", endpoint=\"/checkout\", status=~\"2..\"}[5m])) / sum(rate(http_requests_total{job=\"payment-service\", endpoint=\"/checkout\"}[5m]))",
    "target": 99.95,
    "threshold": "0.0005",
    "budgetingPeriod": "28d",
    "team": "payment",
    "alerting": {
      "enabled": true,
      "notificationChannels": ["slack#payment-alerts"],
      "burnRateThresholds": [
        {
          "rate": 5.0,
          "period": "1h",
          "severity": "warning"
        }
      ]
    }
  }' \
  --output "payment-service-availability.zip"

echo ""
echo "Получение списка существующих SLO..."

curl -s "${BASE_URL}/api/slos" | jq .

echo ""
echo "Получение шаблонов..."

curl -s "${BASE_URL}/api/templates" | jq .

echo ""
echo "Примеры успешно выполнены!"
echo "Созданные файлы:"
ls -la *.zip
