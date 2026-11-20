# План рефакторинга SLOzy OSS

## 🔴 Высокий приоритет

- [x] **1. Имя модуля**: `go.mod` объявляет `github.com/philyuchkoff/slozy` вместо `slozy-oss`. Исправить во всех импортах.
- [x] **2. Раздутый `main.go`**: вынести конфигурацию middleware в `middleware_setup.go`, роуты в `routes.go`, `printServerInfo` в отдельный файл.
- [x] **3. Тесты расходятся с реализацией**: `security_test.go` ожидает `'unsafe-inline' 'unsafe-eval'` в CSP, но `security.go` этого не содержит.
- [x] **4. Битые цели в Makefile**: `release-archive`, `dev` ссылаются на несуществующие файлы (`build-release.sh`, `.air.toml`). Исправить или удалить.
- [x] **5. `.gitignore`**: содержит `deploy/` (но директория закоммичена), артефакты `AUDIT.md`, `PLAN.md`, `ARCHITECTURE_DRAFT.md`.

## 🟡 Средний приоритет

- [x] **6. Дублирование проверок path traversal**: в `handlers/slo.go:safeFilename()` и `handlers/template.go:loadTemplateFile()`. Вынести в общий `internal/safepath/`.
- [x] **7. Двойная санитизация Prometheus-запросов**: `validator.go:validatePrometheusQuery()` и `prometheus.go:sanitizeQuery()`. Объединить.
- [x] **8. Мёртвый код**: `SLOHandler.tempDir` не используется, методы `Wrap()` не вызываются.
- [x] **9. Утечка горутин в rate limiter**: cleanup-горутина без graceful shutdown. Заменена на периодическую очистку в основном потоке.
- [x] **10. Побочный эффект в GET**: `loadTemplates()` вызывает `os.MkdirAll()`. Создание директории только в `main.go`.
- [x] **11. `log` → `slog`**: заменён стандартный логгер на структурированный (`internal/logger`).
- [x] **12. Жёстко закодированные URL**: Grafana URL теперь берётся из `GRAFANA_URL` env var. GitHub/favicon URLs обновлены на `slozy-oss`.

## 🟢 Низкий приоритет

- [x] **13. Два Dockerfile**: production Dockerfile обновлён (добавлены `static/` и `templates/`), корневой оставлен для разработки.
- [ ] **14. Два nginx-конфига**: `/deploy/nginx.conf` и `/deploy/production/nginx.prod.conf` — унифицировать (опционально).
- [x] **15. Два systemd-сервиса**: оба оставлены (compat для старых систем), обновлены Documentation URL.
- [x] **16. Русские комментарии**: заменены на английские.
- [ ] **17. Нестандартная структура `internal/`**: оставлена как есть (вкусовщина).
