# Полный пример конфигурации SLO

Этот пример показывает полный набор файлов, которые генерирует SLOzy при создании SLO с алертингом.

## Файлы в ZIP архиве

1. **notification-service-api-latency.yaml** - Основной SLO файл в формате OpenSLO
2. **notification-service-api-latency-prometheus-alerts.yaml** - Правила алертов для Prometheus
3. **notification-service-api-latency-alertmanager-config.yaml** - Конфигурация Alertmanager
4. **notification-service-api-latency-grafana-dashboard.json** - Дашборд Grafana
5. **notification-service-api-latency-multi-burnrate-dashboard.json** - Расширенный дашборд с несколькими burn rate
6. **notification-service-api-latency-README.md** - Документация по развертыванию

## Как использовать

1. Скопируйте файлы в соответствующие директории вашего мониторинга
2. Настройте метрики под ваше окружение
3. Обновите каналы уведомлений
4. Импортируйте дашборды в Grafana
5. Примените конфигурации в Prometheus и Alertmanager

## Пример запроса для создания такого SLO

```bash
curl -X POST http://localhost:8080/api/slo/create \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "notification-service",
    "sloName": "api-latency",
    "sloDescription": "99th percentile latency for notification API",
    "indicatorType": "latency",
    "metricSource": "prometheus",
    "metricQuery": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"notification-service\"}[5m]))",
    "target": 99.9,
    "threshold": "500ms",
    "budgetingPeriod": "28d",
    "team": "backend",
    "alerting": {
      "enabled": true,
      "notificationChannels": ["slack#notifications-alerts", "email:oncall@example.com"],
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
  }' --output notification-service-config.zip
```
