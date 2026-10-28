# Конфигурация, startup и ошибки

`hub --config config.example.yaml` требует один явный YAML document, максимум1MiB. Порядок: typed defaults → strict YAML → только разрешённые infrastructure env → semantic validation → composition. Неизвестные/duplicate keys, второй document, NaN/Inf/ranges, overlaps/mapping errors отклоняются до listener. YAML parser `go.yaml.in/yaml/v3 v3.0.5` (MIT и Apache, см. LICENSE модуля): maintained Go YAML project, ограничение aliases есть в parser; собственный input-size bound добавлен в loader. Общий dependency policy/security audit — будущий release gate, здесь pin+strict validation+negative tests.

| Владелец | Данные / overrides |
|---|---|
| Env infrastructure whitelist | `HUB_PROFILE`, `HUB_ADAPTER_MODE`, `HUB_STORAGE`, `HUB_SERVER_HOST`, `HUB_SERVER_PORT`, `HUB_LOG_LEVEL`, `HUB_CONTAINER_NETWORK`, `HUB_DB_PATH` |
| Env secrets | `SESSION_SECRET`, `ML_SERVICE_TOKEN`, `MQTT_PASSWORD`; YAML keys для них запрещены |
| YAML | Server timeouts, event bounds, LED defaults/mapping/budgets, simulation seed/tick, enabled components |
| SQLite | Users/roles, timezone setting, desired zones и command metadata реализованы; sessions пока schema-only, presets — будущий domain module |

`.env.example` — справочник, Core **не загружает `.env`**. Никакой blind env override brightness/user state. Unknown env ignored. Secrets disabled components не обязательны; enabled ML/MQTT без соответствующего token/password дают error. Enabled integration с secret всё равно NOT_IMPLEMENTED/degraded: внешние endpoints/transport/auth интеграций ещё не реализованы, сеть к ним не открывается.

Profiles: development/test; production fail closed NOT_IMPLEMENTED. Adapter mode — отдельное поле; доступна только explicit simulation, hardware fail closed. Storage: `memory_fake`, `not_implemented` либо `sqlite`. NOT_IMPLEMENTED демонстрирует stage readiness503. Memory fake — bounded latest-observation store на5 ресурсов; SQLite — отдельные persistent repositories, не история observations.

Server host — numeric loopback (IPv4/IPv6), port0..65535 (0 нужен изолированным smoke tests). Публичный bind запрещён; только explicit container flag разрешает0.0.0.0. Таймауты0<d<=60s, simulation tick>=1ms. Event queue1..4096, max subscriptions1..4096, payload128..1048576 bytes. Default queue64, max subscriptions128, payload16KiB — baseline, не обещание производительности Pi.

Lighting: capacity1..4096; уникальные zone IDs; offset/count в bounds, без overlaps; brightness finite0..1; ток finite>0; kind strip/matrix; matrix width×height=count; side_counts либо отсутствует, либо ровно4 с суммой=count. Это validation/safety defaults, **не реализация render/current limiter**. Fake LED fixture —4 пикселя и2×2 matrix, не попытка управлять реальным layout YAML. Сценарии arbitrary production layout/limiter — Д026+.

`--check-config` валидирует и печатает safe JSON log без secrets; exit0. Invalid config exit2 с безопасными `CONFIG_PARSE/CONFIG_VALUE/CONFIG_SIZE/CONFIG_READ/NOT_IMPLEMENTED`, без raw YAML/path/env. `Secret` форматирование/JSON/slog redacted; whole Config JSON исключает secrets. HTTP errors не содержат traceback. Логи — JSON stderr; allowlist: event label, bounded route, status, ms duration, sanitized request ID, synthetic profile, event count. Payload/request body/query/env не логируются. Persistent log retention/redaction на будущих integrations — Q-030, не закрыт этим каркасом.

Lifecycle: последовательный Start; на fail cleanup успешно начатых components в reverse order. Component.Start обязан освободить собственный partial init при ошибке и учитывать ctx; Stop — учитывать deadline. App отменяет worker context и ограничивает aggregate stop timeout. SIGTERM/INT — штатный exit0; worker fatal/listen failure — cleanup и exit1. Некооперативную произвольную Go goroutine runtime принудительно не убивает: контракт компонентов требует bounded operations, что соблюдено текущими workers. Бесконечные retries не добавлены.

## Обновление Д016–Д018

Storage поддерживает `sqlite` с typed `sqlite.path/busy_timeout/startup_timeout`.
Default path `.runtime/data/hub.db`, busy500ms (range1ms..5s), startup10s (1ms..60s).
Path не включается в Config JSON. Directory0700/DB0600 обязательны; fallback на memory
при ошибке нет. Existing observation cache остаётся RAM, SQL repositories отдельны.
`auth.required=true` допустим только с SQLite. Без `auth.enabled=true` legacy profile
держит ready503. В `configs/auth.dev.yaml` включены issuer, HTTPS public_origin и
TLS pair; альтернативно разрешён явно trusted host proxy.
[Sessions/tokens](../security/sessions-tokens.md). [Storage](storage.md), [admin](../security/admin-bootstrap.md).
Target: толькоPi4B/2GB; [memory profile](pi4b-2gb.md).
