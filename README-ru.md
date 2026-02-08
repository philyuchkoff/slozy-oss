# 🦥 SLOzy - [SLO made lazy and easy](https://slozy.ru/)

Инструмент для простого создания SLO (Service Level Objectives) в формате OpenSLO с полной поддержкой алертинга.

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.19%2B-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="https://img.shields.io/badge/OpenSLO-compatible-4A154B?style=for-the-badge" alt="OpenSLO">
  <img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="License">
</p>

## Что это?

SLOzy - это веб-инструмент, который позволяет командам разработки и продукта создавать SLO без необходимости разбираться в сложностях формата [OpenSLO](https://github.com/OpenSLO/OpenSLO). 
Входные данные валидируются, если что-то не так - получите подсказку.
Просто заполните форму - получите готовый YAML файл и положите его туда, куда скажут SRE, дальше они разберутся сами.
Для самых тяжелых случаев можно нажать "Fill example" и все поля заполнятся подсказками.

## ✨ Особенности

### Основные возможности
- 🎯 **Простой веб-интерфейс** - заполните форму, получите SLO
- 📚 **Шаблоны** - готовые шаблоны для разных типов сервисов (Web API, базы данных, очереди и т.д., можно добавлять и править под любой продукт, примеры лежат в `/templates`)
- 🔧 **Автоматическая генерация** - из простого описания в валидный OpenSLO YAML
- 🏷️ **Автодополнение** - подсказки метрик, команд, владельцев
- 📊 **Предпросмотр** - визуализация SLO до сохранения

### 🔔 Полная поддержка алертинга:
- **OpenSLO Alert Policies** - интегрированы в основной SLO YAML
- **Prometheus Alert Rules** - отдельный YAML файл с правилами
- **AlertManager Config** - пример конфигурации для маршрутизации алертов
- **Grafana Dashboards** - два готовых дашборда для визуализации
- **README файл** - инструкции по развертыванию
- **Burn rate thresholds** - настраиваемые пороги с разными severity уровнями
- **Множественные каналы** - Slack, Email, PagerDuty, webhook

### 📦 ZIP
- **Все в одном ZIP** - все конфигурационные файлы упаковываются в один архив
- **Уникальные имена** - архивы содержат timestamp для уникальности

### 🎨 Функции интерфейса:
- Чекбокс "Enable Alerting" с расширенной конфигурацией
- Настройка Burn Rate Thresholds (warning/critical)
- Указание каналов уведомлений
- Шаблоны с предварительно настроенным алертингом
- Просмотр сгенерированных файлов
- Примеры запросов для разных типов метрик

### ✅ Технические улучшения:
- Форматирование чисел без ошибок плавающей точки
- Правильная генерация PromQL выражений
- Alert Inhibition для предотвращения alert storms (механизм в Alertmanager, который предотвращает отправку менее важных алертов, когда уже сработал более важный алерт для того же объекта)
- Подробные аннотации и метки
- Валидация входных данных

## 🚀 Быстрый старт

### Требования
- Go 1.19 или выше
- Веб-браузер с поддержкой JavaScript

### Установка и запуск
```bash
# Клонируйте репозиторий
git clone https://github.com/philyuchkoff/slozy.git
cd slozy

# Запустите сервер на 8080 порту
go run main.go

# Или запустите на другом порту
PORT=9090 go run main.go
```

Приложение будет доступно по адресу: [http://localhost:8080](http://localhost:8080)

## 🎯 Использование

### 1. Заполнение формы
1. Откройте [http://localhost:8080](http://localhost:8080)
2. Введите базовую информацию о сервисе:
   - Название сервиса
   - Название SLO
   - Команда
   - Описание

### 2. Настройка метрик
1. Выберите тип индикатора (Latency, Availability и т.д.)
2. Укажите целевое значение (например, 99.9%)
3. Задайте пороговое значение (например, 200ms для латентности)
4. Укажите период бюджета ошибок (по умолчанию 28 дней)

### 3. Конфигурация метрик
1. Выберите источник метрик (Prometheus, Datadog, CloudWatch и др.)
2. Введите запрос к метрикам
3. Используйте готовые примеры или создайте свой запрос

### 4. Настройка алертинга (опционально)
1. Включите алертинг
2. Настройте каналы уведомлений (Slack, Email, PagerDuty)
3. Добавьте пороги burn rate:
   - Warning: 5-10x быстрее за 1 час
   - Critical: 50-100x быстрее за 5 минут

### 5. Генерация и скачивание
1. Нажмите "Generate SLO with Alerts"
2. Все конфигурационные файлы будут упакованы в ZIP архив
3. Архив автоматически скачается

## 📄 Генерируемые файлы

В ZIP архиве будут созданы следующие файлы:

```
service-name-slo-name-YYYYMMDD-HHMMSS.zip
├── service-name-slo-name.yaml            # Основной SLO в формате OpenSLO
├── service-name-slo-name-prometheus-alerts.yaml     # Правила алертов Prometheus
├── service-name-slo-name-alertmanager-config.yaml   # Конфигурация Alertmanager
├── service-name-slo-name-grafana-dashboard.json     # Дашборд Grafana
├── service-name-slo-name-multi-burnrate-dashboard.json  # Расширенный дашборд
└── service-name-slo-name-README.md                   # Документация
```

## 📋 Поддерживаемые типы SLO

- **Latency** — время ответа сервиса
- **Availability** — uptime сервиса
- **Throughput** — количество запросов
- **Кастомные метрики**

## 📚 Шаблоны SLO

В SLOzy включены готовые шаблоны для:

### 1. Web API Latency
- Цель: 99.9%
- Порог: 200ms
- Алертинг: Warning (10x/1h), Critical (50x/5m)

### 2. Web API Availability  
- Цель: 99.95%
- Порог: 0.0005
- Алертинг: Warning (5x/1h)

### 3. Database Query Latency
- Цель: 99.9%
- Порог: 100ms
- Алертинг: Critical (20x/30m)

### 4. Можно добавить свои шаблоны 

## 🔧 API Endpoints

```bash
# Создание SLO через curl (возвращает ZIP архив)
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

# Получение списка SLO
curl http://localhost:8080/api/slos

# Получение шаблонов
curl http://localhost:8080/api/templates

# Health check
curl http://localhost:8080/health
```

Доступные API endpoints:
- `GET /` — Веб-интерфейс
- `POST /api/slo/create` — Создание SLO (возвращает ZIP архив)
- `GET /api/slos` — Список созданных SLO
- `GET /api/templates` — Шаблоны SLO
- `GET /health` — Health check

## 🐳 Docker

```bash
# Сборка образа
docker build -t slozy .

# Запуск контейнера
docker run -p 8080:8080 -v $(pwd)/data:/app/data slozy
```

## 🛠️ Разработка

### Сборка для продакшена

```bash
# Сборка для Linux
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o slozy-linux-amd64 main.go

# Сборка для macOS
GOOS=darwin GOARCH=arm64 go build -o slozy-macos-arm64 main.go

# Уменьшение размера бинарника
upx --best slozy-linux-amd64
```

### Переменные окружения

```bash
# Порт сервера (по умолчанию 8080)
export PORT=8080

# Директория для данных (по умолчанию ./data)
export DATA_DIR=/var/lib/slozy

# Уровень логирования (debug, info, warn, error)
export LOG_LEVEL=info
```

## 📊 Технологии

- **Backend**: Go (чистый net/http, без фреймворков)
- **Frontend**: HTML5, CSS3, Vanilla JavaScript
- **Форматы**: YAML, JSON
- **Стандарты**: OpenSLO v1alpha
- Prometheus, Grafana, Alertmanager совместимость

## 🔒 Безопасность

### Рекомендации по использованию в продакшене
1. Используйте HTTPS (рекомендуется nginx или traefik как reverse proxy)
2. Ограничьте доступ к API
3. Не храните чувствительные данные в директории `data/`
4. Используйте переменные окружения для конфигурации

### Ограничения
- Приложение предназначено для внутреннего использования
- Не реализована аутентификация/авторизация
- Генерируемые файлы могут содержать шаблонные данные (замените на свои перед использованием)

### Полезные советы
- Адаптируйте метрики - замените примеры запросов на ваши реальные метрики
- Настройте алертинг - укажите правильные каналы уведомлений для вашей команды
- Проверьте пороги - настройте burn rate thresholds под ваш сервис
- Тестируйте - перед применением в продакшене протестируйте алерты в тестовом окружении

## Telegram Notifications

SLOzy поддерживает отправку уведомлений в Telegram. Для настройки:

### Быстрая настройка:
1. Создайте бота через @BotFather
2. Получите токен бота
3. Получите chat ID (личный, группы или канала)
4. Установите переменные окружения:
   ```bash
   export TELEGRAM_BOT_TOKEN="ваш_токен"
   export TELEGRAM_CHAT_ID="ваш_chat_id"
5. В форме создания SLO укажите: `telegram:ваш_chat_id`
6. Подробные инструкции в [examples/telegram-notifications/README.md](./examples/telegram-notifications/README.md)

## 📈 Roadmap

- [ ] Импорт существующих SLO
- [ ] Экспорт в другие форматы (Pulumi, Terraform)
- [ ] Интеграция с Git (авто-коммит конфигураций)
- [x] Отправка нотификаций в Telegram
- [ ] Планировщик пересмотра SLO
- [ ] История изменений SLO
- [ ] Multi-tenancy поддержка
- [ ] API ключи для автоматизации

## 🤝 Вклад в проект

1. Форкните репозиторий
2. Создайте ветку для вашей функции (`git checkout -b feature/useful-feature`)
3. Зафиксируйте изменения (`git commit -m 'Add useful feature'`)
4. Запушьте ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

### Требования к коду
- Следуйте существующему стилю кодирования
- Добавляйте тесты для новой функциональности
- Обновляйте документацию при необходимости

## 📝 Примеры

Примеры готовых SLO файлов находятся в директории `examples/`.

## 🐛 Отчеты об ошибках

Нашли баг? Пожалуйста, создайте issue с детальным описанием:
1. Шаги для воспроизведения
2. Ожидаемое поведение
3. Фактическое поведение
4. Логи (если есть)

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для деталей.

## 🙏 Благодарности

- Команде OpenSLO за отличную спецификацию
- Сообществу Prometheus за инструменты мониторинга
- Всем контрибьюторам и пользователям

## 📚 Полезные ссылки

- [OpenSLO Specification](https://openslo.com/)
- [SRE для неинженеров](https://srebook.ru/)