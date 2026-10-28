# Health и metrics: Д013–Д015

HTTP contracts: [OpenAPI](../api/openapi.yaml). GET-only `/health`, `/ready`, `/metrics`, `/api/v1/version`; отдельно stage POST `/api/v1/auth/login` без session issuance. Unknown routes404 JSON; unsupported methods405 JSON+Allow:GET. Включая internal errors, request ID проходит bounded ASCII validation или заменяется random128-bit hex; не auth и не metric label. Responses no-store/nosniff. Panic recovery буферизует response (limit64KiB) и возвращает sanitized500, не partial success. Нет CORS, auth, command/API write и WS transport.

| Состояние | `/health` | `/ready` |
|---|---|---|
| STARTING, adapter unavailable | 200 live | 503 |
| RUNNING, memory_fake + synthetic adapters READY | 200 | 200, simulated=true/persistent=false |
| SQLite/storage NOT_IMPLEMENTED или critical STALE/UNAVAILABLE | 200 | 503 |
| Optional ML/MQTT unavailable/not implemented | 200 | 200+degraded при готовых critical |
| Disabled ML/MQTT | 200 | Не degraded |
| SAFE_MODE health evaluator | 200 live | 503; emergency control ещё не реализован |
| STOPPING | 503 | 503 до закрытия listener |

Это cached/in-memory dependency state, без outbound calls на каждом HTTP request. Реально запускаются fake suite/observer/HTTP; optional actual services отсутствуют. STARTING → RUNNING выставляется после успешного fake Start, fatal worker завершает приложение. Stale/Safe Mode truth table проверена injection tests, auto TTL/checker/watchdog/recovery реализуются с реальными интеграциями позже. Readiness не проверяет весь release manifest и не закрывает M0/M1.

Prometheus `/metrics`: `hub_http_requests_total`, `hub_http_errors_total`, `hub_http_request_duration_ms_{bucket,sum,count}`; cumulative buckets1/5/10/50/100/1000/+Inf **ms**. Шесть bounded route labels (включая login), неизвестные URLs всегда `unmatched`, нет user IDs/payload/query. Snapshot содержит только завершённые запросы (сам текущий scrape ещё не учтён). Event counters: published/dropped/overflow/rejected + active subscriptions. Они описывают реальный bus, не CPU/temperature реальной Pi. История событий/аудио/кадров не пишется в metrics/logs.

Producer IDs/sequences/timestamps в internal bus не аутентификация внешнего устройства. Gateway auth/validation/rate limits и per-resource browser permissions появятся на следующих этапах. Подключение Prometheus/Grafana optional и здесь не запущено.

## Дополнение Д016–Д018

SQLite READY требует open/migrations и реального writable probe, persistent=true
только после этого. Simulated остаётся true с fake adapters. Required auth добавляет
пятую critical dependency credentials_only/NOT_IMPLEMENTED до sessions Д019. Explicit
core-only SQLite dev profile ready200 не является authenticated production profile.
POST `/api/v1/auth/login` — bounded credential verifier, valid503/no cookie,
invalid401/limit429. Metrics имеют шестой bounded route label для login; request
body/password/username не логируются. [Stage contract](../security/admin-bootstrap.md).
