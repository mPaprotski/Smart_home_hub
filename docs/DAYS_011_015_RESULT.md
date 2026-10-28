# Дни11–15: результат реализации

Дата выполнения:25.10.2026 UTC. Основание: [подробный план](plans/DAYS_011_015_DETAILED.md), исходное [ТЗ](../Smart_Home_Hub_TZ.md). Рабочий root — Git-репозиторий `Smart_home_hub`, исходники пока не закоммичены/не отправлены. **Программный срез реализован; physical M0 и полный M1 не закрыты. Remote GitHub CI не запускался.**

## Что работает

| День | Реализованный результат | Подробности |
|---|---|---|
| 11 | Root Go module/3 entrypoints, React19+TS6+Vite8 shell, installable Python contract, locked tools, Make | [Д011](DAY_011_RESULT.md) |
| 12 | Strict YAML/env config, secrets redaction, validation/mapping, reverse cleanup и bounded shutdown, JSON logs | [Д012](DAY_012_RESULT.md) |
| 13 | Actual net/http server, liveness/readiness matrix, JSON errors/request IDs/recovery, OpenAPI3.0.3, bounded metrics | [Д013](DAY_013_RESULT.md) |
| 14 | RAM-only bounded Event Bus, envelope v1, catalog/filter/policies, payload ownership, concurrent close/race/benchmark | [Д014](DAY_014_RESULT.md) |
| 15 | Seeded sources, sensor→bus→Core observer, LED/matrix fakes, bounded memory store, native/Compose smoke и CI workflow | [Д015](DAY_015_RESULT.md) |

Код разделён по [структуре](development/repository-layout.md); [quickstart/check commands](development/README.md), [config](development/configuration.md), [HTTP/health/metrics](observability/README.md), [OpenAPI](api/openapi.yaml), [event baseline](research/event-bus-baseline.md).

## Проверки и evidence

- Go:23 предметных tests, `go test -race` на product packages, vet/build; старые experiments не включены в count.
- Python:14 pytest cases, Ruff check/format, strict mypy; sdist/wheel действительно собраны, runtime dependencies empty.
- Frontend: Prettier/ESLint/strict TypeScript/Vite production bundle. Unit tests N/A: пока статический shell без API/предметной логики.
- OpenAPI: validator OK; actual routes/errors/methods/readiness проверены unit и process smoke.
- Native: реальные процессы Hub с `memory_fake` ready200 и `not_implemented` ready503, оба live200. Events дошли до Core; SIGTERM exit0, listener закрыт. Production/hardware/public-bind/invalid config — exit2 до listener; secret sentinel не напечатан.
- Docker: image реально собран, Compose actual healthy/run/events/restart/SIGTERM/cleanup; UID10001, read-only rootfs и loopback host publication проверены через inspect.
- Workflow: actionlint и Compose config прошли; pinned action SHA проверены по официальным Git refs. Remote CI **BLOCKED до отдельного commit/push**, не PASS-CI.
- Чистая source copy: новая venv и npm ci, full check/build/native smoke; см. evidence. Это не Git checkout опубликованного commit.

[Машиночитаемый итог](results/days011-015/summary.json), [quality log](results/days011-015/quality.log), [native JSON](results/days011-015/native-smoke.json), [config JSON](results/days011-015/config-smoke.json), [Compose JSON](results/days011-015/compose-smoke.json), [clean source log](results/days011-015/clean-source.log), [benchmark](results/days011-015/event-bus-benchmark.txt). Логи содержат только этот synthetic development scope, пути нормализованы; secrets/real devices не использовались.

Среда: Go1.27.3 linux/amd64; Python3.12.3; project-local Node22.23.4; npm9.2.0; Docker29.1.3/Compose2.37.1. Образы pinned digest, действия pinned commit SHA, package versions/lockfiles сохранены. Системный Node18/host services не обновлялись. Socket tests запускались в разрешённом host context: restricted sandbox сам по себе запрещает listen.

## Границы результата

Memory fake не называется SQLite: persistent=false и данные исчезают после stop. Storage `not_implemented` оставляет ready503. Production/hardware режимы отклоняются; auth/Argon2id/sessions/scoped tokens/commands/persistence/WS/ML inference/real MQTT/Telegram/capture/LED drivers не готовы. HTTP до auth read-only/loopback; container opt-in отдельный и dev Compose публикует только127.0.0.1.

Simulation fakes — минимальные operations:4-pixel strip,2×2 matrix и mapping fixtures, derived sensor/audio/presence/screen/MQTT observations. Они не управляют physical layout из YAML, не проверяют DSP correctness, pairing/fusion/реальный broker или Lighting current limiter. Health stale/Safe Mode проверены injection; watchdog/auto TTL/emergency auth будущие. Event bus не durable log, не external ingress auth/dedupe; WS resync ещё впереди. Полный simulation UI/acceptance — Д024–Д025.

Первый Compose restart smoke обнаружил ошибку тестового клиента: автоматически опубликованный host port сменился после restart. Harness исправлен: адрес читается заново, повторный end-to-end проход успешен. Неуспешный первый запуск не объявлен PASS.

Исходное ТЗ не изменено: SHA256 `84bf6c3c61b2feff16a23a12226b9646d33b4ffde3754b9c0082db0e4c57ec09`. Старые experiments/results сохранены. Global requirement acceptance не переведён массово в PASS: доказан только описанный срез.

## Передача на Д016

1. Добавить SQLite connection/repository/migrations/backup policy и заменить отдельный memory provider, сохранив simulation identity.
2. Протестировать persistence/restart/error/transaction boundaries; required SQLite readiness должна отражать реальный результат migration/connection.
3. Далее command layer и auth до появления управляющих routes/LAN exposure; Core observer не исполняет события как команды.
4. После commit/push выполнить remote workflow и записать run/commit: сейчас только local equivalent. Физический Pi/M0 вести отдельной evidence matrix.
