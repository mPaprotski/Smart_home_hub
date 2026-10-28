# Smart Home Hub

Локальная система управления светом, аудио и устройствами умного дома на **Raspberry Pi 4B, только2GB RAM**. [Техническое задание](Smart_Home_Hub_TZ.md).

Реализован программный срез дней11–21: Go Core, SQLite schema7, admin bootstrap,
HTTPS sessions/CSRF, scoped tokens, atomic security audit, защищённый timezone
command gateway, auth-only WebSocket и typed domain repositories с retention.
[Результат дней19–21](docs/DAYS_019_021_RESULT.md).

**Это development simulation.** Для сессий нужен явный HTTPS auth profile;
legacy credentials-only профиль сохраняет503 без cookie. Dashboard/business WS —
следующие этапы. Physical M0 открыт: настроенного SSH и реальных измерений нет.
[Ограничения и бюджет2GB](docs/development/pi4b-2gb.md).
[Sessions/tokens и запуск HTTPS](docs/security/sessions-tokens.md).

## Запуск

Команды выполняются из корня этого репозитория `Smart_home_hub` (в workspace это вложенный каталог).

```bash
# Требуются Go1.27.3, Python3.12.3, Node22.23.4; без Pi/секретов.
python3 -m venv .venv
.venv/bin/python -m pip install -r ml/requirements-dev.lock
npm --prefix frontend ci
go mod download
make check
make smoke
make run-sim
```

На текущей машине используйте `make GO=/snap/go/current/bin/go ...`: snap-wrapper в sandbox недоступен. Уже установленный локальный Node автоматически выбирается `scripts/npm.sh`. Для установки Node локально и всех вариантов запуска см. [инструкцию разработки](docs/development/README.md).

Core слушает `http://127.0.0.1:8080`: `/health`, `/ready`, `/api/v1/version`, `/metrics`. Только GET; Ctrl+C завершает процесс. UI shell: `scripts/npm.sh run dev`, отдельно от Core.

```bash
# Фактический Docker smoke: build/start/ready/events/restart/stop/cleanup.
docker compose -p smarthome-d15 -f deploy/docker/compose.dev.yaml build
.venv/bin/python scripts/compose_smoke.py
```

## SQLite и bootstrap

```bash
make GO=/snap/go/current/bin/go build-go
./bin/hub admin bootstrap --config configs/sqlite.dev.yaml --username admin
./bin/hub admin verify --config configs/sqlite.dev.yaml --username admin
./bin/hub --config configs/sqlite.dev.yaml
```

Password вводится с TTY без echo, default password отсутствует. DB `.runtime/data`
приватна. [Storage](docs/development/storage.md), [permissions](docs/security/permission-matrix.md),
[admin/password policy](docs/security/admin-bootstrap.md). Persistent smoke:
`make GO=/snap/go/current/bin/go smoke-persistence`.

## Документация

- [Ежедневный план](docs/DEVELOPMENT_PLAN_DAILY.md), [подробные дни11–15](docs/plans/DAYS_011_015_DETAILED.md), [подробные дни16–18](docs/plans/DAYS_016_018_DETAILED.md), [подробные дни19–21](docs/plans/DAYS_019_021_DETAILED.md)
- [Результаты дней11–15](docs/DAYS_011_015_RESULT.md), [конфигурация](docs/development/configuration.md), [структура](docs/development/repository-layout.md)
- [OpenAPI](docs/api/openapi.yaml), [health/metrics](docs/observability/README.md), [Event Bus baseline](docs/research/event-bus-baseline.md)
- [Требования](docs/requirements.md), [backlog](docs/backlog.md), [открытые вопросы](docs/open-questions.md), [архитектура/ADR](docs/architecture/README.md)
- [BOM](hardware/bom/BOM.md), [питание](hardware/bom/power-budget.md), [wiring](hardware/wiring/wiring.md)
- [Результаты дней4–10](docs/DAYS_004_010_RESULT.md), [отдельные эксперименты](experiments/README.md), [перенос репозитория](docs/REPOSITORY_RELOCATION.md)

Secrets, real seeds, keys, SQLite, local caches и build artifacts исключены через `.gitignore`. Исходное ТЗ и старые результаты сохранены. Исходники подготовлены в целевом Git-репозитории; commit/push и remote CI этим этапом не выполнялись.
