# Telegram Notifications для SLO алертов

Как настроить Telegram нотификации для SLO алертов.

## 1: Создание Telegram бота

1. Откройте Telegram и найдите @BotFather
2. Отправьте команду `/newbot`
3. Придумайте имя для бота (например, `SLO Alert Bot`)
4. Получите токен бота (например: `1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ`)
5. Сохраните токен

## 2: Получение Chat ID

### Для личных уведомлений:
1. Найдите бота @userinfobot в Telegram
2. Отправьте ему `/start`
3. Он покажет ваш chat ID

### Для групповых чатов:
1. Добавьте бота в группу
2. Добавьте @RawDataBot в группу
3. Он покажет chat ID группы

### Для каналов:
1. Создайте канал
2. Добавьте бота в канал как администратора
3. Отправьте сообщение в канал
4. Получите chat ID из ссылки: `https://t.me/c/1234567890/1`

## 3: Настройка окружения

```bash
# Установите переменные окружения
export TELEGRAM_BOT_TOKEN="1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ"
export TELEGRAM_CHAT_ID="-1001234567890"

# Или добавьте в .env файл
TELEGRAM_BOT_TOKEN=ваш_токен
TELEGRAM_CHAT_ID=ваш_chat_id
```

## 4. Создание SLO с Telegram нотификациями
### Через веб-интерфейс:
В форме создания SLO в разделе "Notification Channels" укажите: `telegram:-1001234567890` - укажите свой, это пример!

Можно комбинировать: `slack#slo-alerts, telegram:-1001234567890, email:sre@example.com`

### Через API:
```bash
curl -X POST http://localhost:8080/api/slo/create \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "user-api",
    "sloName": "api-latency",
    "indicatorType": "latency",
    "target": 99.9,
    "threshold": "200ms",
    "metricQuery": "histogram_quantile(0.99, rate(http_request_duration_seconds_bucket{job=\"user-api\"}[5m]))",
    "team": "backend",
    "alerting": {
      "enabled": true,
      "notificationChannels": [
        "telegram:-1001234567890",
        "slack#slo-alerts"
      ],
      "burnRateThresholds": [
        {
          "rate": 10.0,
          "period": "1h",
          "severity": "warning",
          "alertName": "HighBurnRate"
        }
      ]
    }
  }' --output slo-config.zip
  ```

  ## 5: Если все сделано правильно

  SLOzy сгенерирует конфигурацию Alertmanager с поддержкой Telegram:

  ```yaml
  # alertmanager-config.yaml
receivers:
  - name: 'telegram-alerts'
    telegram_configs:
      - api_url: 'https://api.telegram.org/bot1234567890:ABCDEFGHIJKLMNOPQRSTUVWXYZ/sendMessage'
        chat_id: '-1001234567890'
        message: |
          🚨 *CRITICAL SLO Alert*
          
          *Service:* user-api
          *SLO:* api-latency
          *Severity:* critical
          
          user-api is burning error budget 10.0 times faster than allowed for 1h
          
          [View Dashboard](https://grafana.example.com/d/user-api-api-latency)
        parse_mode: 'Markdown'
        disable_notifications: false
```

## Полезные советы

### Разные чаты для разных severity:
```yaml
# warning → группа команды
# critical → канал on-call
notificationChannels:
  - "telegram:-1001111111111"  # warning группа
  - "telegram:-1002222222222"  # critical канал
```

### Комбинированные уведомления:
```yaml
notificationChannels:
  - "telegram:-1001234567890"      # Telegram группа
  - "slack#slo-alerts"             # Slack канал
  - "email:sre@example.com"        # Email рассылка
  - "pagerduty"                    # PagerDuty для critical
  ```

### Форматирование сообщений:
Telegram поддерживает:

- Markdown: *жирный*, _курсив_, [ссылка](url)
- HTML: <b>жирный</b>, <i>курсив</i>
- Emoji: 🚨 ⚠️ ✅ ❌

## Безопасность:
- не коммитьте токены в Git :)
- используйте переменные окружения
- ограничьте права бота
- используйте разные токены для разных окружений

## Устранение проблем
### Бот не отправляет сообщения:
- Проверьте токен бота
- Убедитесь, что бот добавлен в чат
- Проверьте права бота в чате
- Убедитесь, что бот не заблокирован

### Неверный chat ID:
- Для личных чатов: только цифры
- Для групп/каналов: начинается с `-100`

### Rate limiting:
Telegram API имеет лимиты:

- 30 сообщений в секунду для бота
- 20 сообщений в минуту в группу

Используйте `group_wait` в Alertmanager для агрегации