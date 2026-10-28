# Разработка: рабочий срез Д011–Д018

Все команды — из корня `Smart_home_hub`. Нужны Go1.27.3 (race tests также требуют C compiler), Python3.12.3 с venv, Node22.23.4 и npm. Версии закреплены в `.tool-versions`, `go.mod`, lockfiles. Host Node18 не подходит Vite8. Глобальные Node/Go и system services не изменяются.

## Первый запуск

```bash
python3 -m venv .venv
.venv/bin/python -m pip install -r ml/requirements-dev.lock
npm --prefix frontend ci
go mod download
make check
make smoke
make run-sim
```

`make run-sim` работает в foreground; Ctrl+C или SIGTERM завершает workers и освобождает listener. По умолчанию Core слушает `127.0.0.1:8080`. В другом терминале:

```bash
curl --fail http://127.0.0.1:8080/health
curl --fail http://127.0.0.1:8080/ready
curl --fail http://127.0.0.1:8080/api/v1/version
curl --fail http://127.0.0.1:8080/metrics
```

Frontend — отдельный статический development shell, пока без подключения к API:

```bash
npm --prefix frontend run dev
```

Проверенных online устройств, входа и управляющих кнопок пока нет. `screen-client`/`bot` умеют `--version`; обычный запуск возвращает exit2/NOT_IMPLEMENTED. Python package — validator observation contract, не inference service.

## Если используется эта рабочая машина

Здесь `/snap/bin/go` недоступен в restricted sandbox, а direct toolchain работает:

```bash
make GO=/snap/go/current/bin/go check
make GO=/snap/go/current/bin/go smoke
make GO=/snap/go/current/bin/go run-sim
```

В этом workspace установлен project-local Node22.23.4. `scripts/npm.sh` автоматически использует его; системный Node18 сохранён. На другой машине с Node18 допускается явная локальная установка:

```bash
npm --prefix .tools install --no-save node@22.23.4
scripts/npm.sh ci
```

`make` вызывает wrapper для frontend. В чистом CI используется setup-node с точной версией; `.tools` не хранится в Git. Restricted sandbox может запретить loopback sockets; actual server/race smoke нужно выполнять в контексте с разрешёнными локальными sockets. Пропускать assertions ради PASS нельзя.

## Команды и область проверок

| Команда | Реальная работа |
|---|---|
| `make fmt` | Изменяет форматирование Go/Python/frontend |
| `make fmt-check` | Проверяет без изменения исходников |
| `make lint` | Go vet, Ruff, ESLint |
| `make typecheck` | strict mypy и TypeScript |
| `make test` | Core Go race tests и Python observation negative cases |
| `make openapi-check` | Валидирует OpenAPI3.0.3 |
| `make build` | Три Go binary, frontend bundle, Python sdist/wheel в `bin/` |
| `make check` | Все перечисленные quality gates и build |
| `make smoke` | Настоящий процесс Hub: HTTP, события→observer, SIGTERM, закрытие порта, secret sentinel |
| `make bench` | Event bus bounded slow-subscriber baseline |
| `make compose-check` | Только syntax; не проверка запуска Docker |

У статического UI пока нет предметной логики, поэтому frontend unit test stage — N/A. Не добавлены пустые тесты и `passWithNoTests`; lint/type/build обязательны. Старые эксперименты запускаются отдельно по `experiments/README.md`, не импортируются продуктом, не включаются в Core test count.

## Development Compose

```bash
docker compose -p smarthome-d15 -f deploy/docker/compose.dev.yaml build
.venv/bin/python scripts/compose_smoke.py
```

Smoke проверяет healthy/nonroot/read-only container, loopback publication, события, restart, graceful stop и cleanup. Он откажется трогать уже существующие контейнеры проекта `smarthome-d15`. Host port выбирается автоматически; итог — JSON в stdout.

Для обычного foreground dev launch используйте другое имя проекта:

```bash
docker compose -p smarthome-dev -f deploy/docker/compose.dev.yaml up --build
# Остановка в другом терминале:
docker compose -p smarthome-dev -f deploy/docker/compose.dev.yaml down
```

По умолчанию host `127.0.0.1:8080`; `HUB_DEV_PORT=18080` меняет только host publish. Внутри контейнера bind `0.0.0.0` разрешён явным `HUB_CONTAINER_NETWORK=true`. Файл — development-only: rootfs read-only, UID10001, no privileges/devices/mounts/secrets/volumes, CPU/RAM/PID limits. Это один Core container; UI запускается нативно отдельно, реальных ML/MQTT пока нет. Production profile в Core отклоняется до старта listener.

## CI и следующий этап

`.github/workflows/ci.yml`: hosted ubuntu runner, read-only contents, pinned action SHA, no persisted checkout credentials, без production secrets/self-hosted/PR write. Выполняет locked install, quality, native smoke, actionlint, Compose syntax/build/smoke. Подготовка файла не означает remote PASS-CI: на этом этапе ничего не закоммичено и не отправлено.

Д016–Д021 реализованы: SQLite/settings, first command, HTTPS sessions/CSRF, machine tokens/audit и typed domain schema. Далее UI и business WS. `ready=true` относится только к явно выбранному dev dependency set; SQLite persistence и synthetic hardware — отдельные признаки. M0 и M1 не закрыты.

## Дополнения Д016–Д018

Target — только [Pi4B/2GB](pi4b-2gb.md). Core dev container128MiB RAM/no swap,
Go soft limit96MiB; это не тест всей2GB системы. [Новый отчёт](../DAYS_016_018_RESULT.md).
SQLite/credential verification уже реализованы; старые упоминания будущего SQLite
выше относятся к исходному scaffold Д015.

```bash
make GO=/snap/go/current/bin/go smoke-persistence
./bin/hub admin bootstrap --config configs/sqlite.dev.yaml --username admin
./bin/hub admin verify --config configs/sqlite.dev.yaml --username admin
./bin/hub --config configs/sqlite.dev.yaml
```

[Storage](storage.md), [admin](../security/admin-bootstrap.md), [permissions](../security/permission-matrix.md).
Persistent Compose применяется с двумя файлами:

```bash
docker compose -p smarthome-d18-smoke -f deploy/docker/compose.dev.yaml -f deploy/docker/compose.sqlite.yaml build
.venv/bin/python scripts/persistence_compose_smoke.py
```

Для рабочей dev DB выбрать другой project name; smoke удаляет только собственный
новый test volume. `auth.required=true` сохраняет ready503 в legacy profile без issuer.
В явном auth.dev HTTPS profile session issuer уже работает.

## Дополнения Д019–Д021

[Результат](../DAYS_019_021_RESULT.md), [security policy](../security/sessions-tokens.md),
[SQLite schema7](../storage/data-dictionary.md).

```bash
make smoke-auth compose-auth-check
docker compose -p smarthome-d21-smoke -f deploy/docker/compose.dev.yaml -f deploy/docker/compose.sqlite.yaml build
.venv/bin/python scripts/auth_compose_smoke.py
```

Smoke создаёт только собственный новый project/volume и disposable CA/keys, проверяет
TLS chain, session/token restart/revoke, then удаляет свои ресурсы. Не запускать его
project name для рабочих данных. Production/hardware и физический M0 остаются закрыты.
