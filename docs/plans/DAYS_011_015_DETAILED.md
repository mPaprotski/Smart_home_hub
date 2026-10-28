# Подробный рабочий план: Д011–Д015

**Hardware baseline от26.10.2026: только Raspberry Pi 4B с2GB RAM.** [Ресурсный профиль](../development/pi4b-2gb.md). Pi совместные memory/performance gates ещё не выполнены.

Подготовлен 25 октября 2026 года. Основание: [ТЗ v1.0](../../Smart_Home_Hub_TZ.md), [общий план](../DEVELOPMENT_PLAN_DAILY.md), [архитектура Д003](../architecture/README.md), [результат Д004–Д010](../DAYS_004_010_RESULT.md), [M0](../milestones/M0_TECHNICAL_BASELINE.md), [backlog](../backlog.md).

**Реализация программного среза Д011–Д015 выполнена:** [фактический отчёт](../DAYS_011_015_RESULT.md). Ниже сохранён исходный рабочий план; отдельные пункты с внешними зависимостями (remote CI/физический Pi) не объявляются PASS. Product Go/frontend/Python scaffold создан; auth/SQLite/ML runtime остаются следующими этапами.

## 1. Цель пяти дней

Получить первый работающий software skeleton: Hub запускается без Raspberry Pi, читает validated configuration, корректно останавливает свои workers, отвечает health/API-контрактом, передаёт события по bounded Event Bus и принимает fake sensor events. Те же проверки можно воспроизвести локально и в CI без production secrets/оборудования.

Это промежуточный срез **до M1 Д025**. SQLite/migrations начинаются Д016, command dispatcher/permissions — Д017, auth — Д018–Д020, полноценный UI/WS — Д022–Д025. Их отсутствие не маскируется fake success в этих пяти днях.

### Граница открытого M0

Pi пока нет либо она не настроена, target SSH не передан. M0_NOT_READY остаётся: physical cooling/GPIO/I²C/SPI/LED/audio/LAN-mobile tests не завершены. Независимая software development разрешена по решению M0 review.

Доступ к host devices/PID1 зависит от контекста запуска/sandbox: escalated inventory и restricted inventory могут различаться. Перед реализацией перепроверить actual runtime context; x86_64 audio/camera/I²C device не становится evidence Raspberry Pi. В simulation не открывать реальные устройства даже при их наличии.

## 2. Расписание и зависимости

Каждый день: **360 минут предметной работы + 120 минут сопровождения = 8 часов**. Сопровождение:20мин вход/план,40мин документация/tool versions,40мин troubleshooting,20мин journal/следующий шаг. Бюджет не гарантирует, что произвольная несовместимость dependency укладывается в40мин.

| День | Основной результат | Вход | Условие выхода |
|---|---|---|---|
| Д011 | Monorepo/tooling/минимальные entrypoints | Docs/ADR/existing experiments | Clean build и documented commands без Pi |
| Д012 | Validated config, lifecycle, JSON logging | Д011 | Invalid config не стартует; cancel/shutdown ограничены |
| Д013 | HTTP/API/health/OpenAPI/metrics skeleton | Д012 | Readiness честно различает stage/critical/degraded |
| Д014 | In-process bounded Event Bus | Config/lifecycle/contracts | Slow consumer и concurrent shutdown не ломают систему |
| Д015 | Fake adapters, native/Compose simulation и CI v0 | Д011–Д014 | Fake sensor→Core event доказан; checks воспроизводимы |

Д014 не требует закончить весь API: он зависит от config/lifecycle/событийных контрактов; номер дня сохраняет рабочий порядок одного разработчика. Frontend/Python build scaffolds появляются на Д011, их предметные функции — в последующих этапах.

### Evidence levels

- `PASS-HOST`: фактический запуск/tests на developer machine, версии известны.
- `PASS-SIM`: synthetic/fake scenario с проверенным expected output.
- `PASS-COMPOSE`: реальный Docker Compose build/start/smoke/cleanup; одна валидация YAML сюда не входит.
- `PASS-CI`: workflow действительно выполнен в GitHub на указанном commit, доступен run/result.
- `DOC-ONLY`: подготовлен contract/template/инструкция без запуска.
- `BLOCKED`: нет toolchain/network/Docker daemon/repo access/hardware; причина и повторяемый следующий шаг.
- `FAIL`: проверка запущена и дала неверный результат; сохранён reproduction.

Код0 shell-команды, которая просто печатает «OK», не evidence. Ни один из этих software статусов не закрывает physical M0 или product acceptance §50.

## 3. Общие правила реализации

1. Рабочий product root — репозиторий `Smart_home_hub/`, куда перенесены ТЗ, `docs/`, `hardware/`, `configs/`, `deploy/` и `experiments/`. Все команды выполнять из этого каталога; не создавать внутри него ещё один одноимённый project root.
2. Существующие experiments/results и отчёты сохранять. Продвигать spike в product code только после review API/validation/test scope; `internal/` не импортирует `experiments/`.
3. `go.mod` product использует modules. `GO111MODULE=off` из experiment runner не переносить в product Make/CI; product `go env GOMOD` должен указывать на новый root module.
4. Exact tool/dependency versions фиксировать после actual compatibility check. Наличие host Go/Python/Node не означает поддержку выбранного Vite/lint packages.
5. SQLite unavailable/NOT_IMPLEMENTED не объявлять READY. В simulation fake storage/adapter имеют явные kind/status, их нельзя включить в production profile без отказа.
6. Пока auth нет, network binding только loopback; public/LAN control routes не открывать. Production mode с неготовыми auth/storage/hardware даёт понятный startup/ready отказ, а не fallback на anonymous simulation.
7. Dependencies/fakes подключать явным wiring, без import-time hardware access. Запуск теста не должен включать Bluetooth/camera/GPIO либо читать microphone.
8. Tests проверяют business/contract/failure behavior, а не копируют реализацию. Docs/simple scaffold не покрывать искусственными тестами ради количества.
9. Git metadata workspace может быть readonly/служебной: проверить фактический worktree. Не повреждать её повторным init. VCS/remote ограничения фиксировать отдельно; code/files можно подготовить в разрешённой области. Публикация GitHub/создание production runner этим планом не выполняется.

---

## Д011. Monorepo, toolchains и первые entrypoints

### Цель и вход

Создать единый воспроизводимый software root без hardware dependency, сохранив ТЗ/архитектуру/experiments. Вход: ADR-001/006/007, structure §41, существующий `.gitignore`, actual Go/Python/Node inventory и доступ к dependency registries при необходимости.

### Блок 11.1 — инвентаризация и границы репозитория, 60 минут

1. Проверить existing files, `AGENTS.md`, writable root, VCS status и отсутствие уже созданного product code. Не перезаписывать пользовательские файлы шаблоном генератора.
2. Зафиксировать source root и список остающихся historical/experiment directories; пустой одноимённый каталог не считать отдельным project root автоматически.
3. Выбрать module path по известному GitHub owner/repo. Если remote ещё не определён — использовать документированный local module name, не выдумывать аккаунт/не добавлять absolute `replace` на свой home.
4. Создать root `go.mod`; решить single-module layout для Hub/screen-client/bot. Для первой версии один module проще, отдельные binaries остаются отдельными entrypoints.
5. Сопоставить architecture modules с будущими папками; не добавлять круговые `api↔lighting↔automation` imports.
6. Решить, какие generated/build/runtime файлы игнорируются: root `dist`, frontend artifacts, caches, local env/keys, SQLite/WAL, private seeds и local evidence с personal data.
7. Проверить существующие ignore rules: не закрывают ли они будущий source package `internal/core` или полезные templates; спорные правила сузить, не удаляя защиты секретов.
8. Сохранить исходное ТЗ неизменным, не выполнять git push/init в read-only `.git` ради формального пункта.

**Артефакты:** `go.mod`, обновлённый `.gitignore`, `docs/development/repository-layout.md`.

### Блок 11.2 — Go skeleton и dependency boundaries, 90 минут

1. Создать `cmd/hub/main.go` и минимальный entrypoint: help/version/build metadata, возвращаемые exit codes, без GPIO/audio/network side effects.
2. Начать `internal/app`, `internal/config`, `internal/api`, `internal/events`, `internal/devices`, `internal/system`. Не создавать десятки пустых public APIs, заранее ограничивающих реализацию.
3. Определить composition root: `main` собирает dependencies, package initialization не запускает goroutines и не открывает `/dev`.
4. Для `cmd/screen-client` и `telegram` сохранить места/минимальные buildable stubs с явным NOT_IMPLEMENTED, если entrypoints нужны для repo contract. Не выдавать stub exit0 за поддержку четырёх платформ/Telegram.
5. Не копировать `zone_frame_probe.go` целиком в Lighting Engine: сначала описать будущую adapter boundary. Hardware driver dependency не должна быть частью обычного Core build на x86.
6. Начать единый `version`/build info contract без реальных secrets/git credentials; dev build имеет честный `development` marker.
7. Проверить `go list`/build/test на обычном host. Контролировать, что module mode включён, `experiments` не стали production dependency.
8. Зафиксировать, как отделяются experiment tests/coverage от `internal` production reports. Root `go test ./...` может включить старые probes, но это не Core coverage.

**Артефакты:** buildable Hub entrypoint, package skeleton, dependency-boundary note. Полный lifecycle реализуется на Д012.

### Блок 11.3 — frontend/Python scaffold, 90 минут

1. Для `frontend/` определить React/TypeScript/Vite manifest и совместимую Node/tool version. Старую host Node18 не считать подходящей для latest toolchain без `engines` check.
2. Создать минимальный UI shell/build entrypoint: нейтральный development placeholder без выдуманных online devices, logged-in admin и готовых control buttons.
3. Добавить TypeScript strict config, ESLint config, scripts `dev`, `typecheck`, `lint`, `build`; lockfile выбранного package manager хранить в repo.
4. Установку Node/dependencies выполнять через project-local/user-authorized tool environment либо контейнер; не обновлять глобальный runtime незаметно.
5. Для `ml/` создать installable Python package/`pyproject.toml`, pinned development tools, `common` types/entrypoint skeleton. Import package не открывает camera и не загружает model.
6. Выбрать минимальный реальный Python contract для ранней проверки, например typed observation validation с запретом raw image payload, конечными числами и версией envelope. Не реализовывать весь Runtime Д046 раньше времени.
7. Отделить dependency set будущих CV/DSP от minimal contract/test package; `numpy`/camera native libraries не обязательны для Core CI scaffold.
8. Подготовить `.env.example`/`config.example.yaml` с комментариями и фиктивными значениями; secrets пока не нужны disabled components, fixed production session secret не добавлять.
9. Проверить frontend/Python commands на реально доступной версии. Network/toolchain failure фиксировать как BLOCKED конкретного job, не заменять build на `echo`.

**Артефакты:** frontend package/config/lock, Python packaging/common contract, example configs. Product Dashboard/auth/PWA и ML inference остаются будущими этапами.

### Блок 11.4 — единые команды разработчика, 60 минут

1. Создать Make/task interface с явными target names: `help`, `fmt`, `fmt-check`, `lint`, `typecheck`, `test`, `build`, `run-sim`, `check`.
2. Каждая команда вызывает реальный инструмент, сохраняет failure exit code и не скрывает missing dependencies.
3. Разделить изменяющий `fmt` и проверяющий `fmt-check`; CI не форматирует source и не выдаёт auto-fixed tree за исходно корректный.
4. Параметризовать Go/Python/package-manager executable. Локальный workaround `/snap/go/current/bin/go` допускается через переменную вызова, не hardcoded home path в общем Makefile.
5. Не использовать `GO111MODULE=off`, `GOTOOLCHAIN=auto` с незадокументированной загрузкой или привязку к личной `.venv` вне repo как обязательный contract.
6. Базовый `run-sim` Д011 может ещё показывать NOT_IMPLEMENTED до wiring Д015; script не имитирует работающие события. В README явно указать ближайший milestone.
7. Разделить продуктовый `check` и optional hardware/media experiments, которым нужны дополнительные sockets/installed tools.

**Артефакты:** Makefile/task scripts, documented command matrix, missing prerequisites list.

### Блок 11.5 — clean build и quickstart, 60 минут

1. Проверить доступный независимый source copy/checkout без private files и build artifacts. Для download-dependent first install отдельно отметить network requirement.
2. Последовательно выполнить dependency install по lockfiles, format check/type checks, build и meaningful existing tests.
3. Убедиться, что default build/run не зависит от Pi, external bot token, broker credentials или root privileges.
4. Проверить `--help`/`--version`, invalid flag и отсутствие side effects от import/entrypoint.
5. Написать root README quickstart: tools, install/check/run, exact expected stage behavior, docs links, M0 hardware blockers.
6. Записать actual versions/commands/result codes, package source/license notes и ограничения VCS/remote.
7. Создать дневной report; не записывать «CI прошёл», если проверено только локально.

### Проверки Д011

| ID | Проверка | Условие PASS |
|---|---|---|
| D11-01 | Единственный source root | Docs/hardware/experiments сохранены, product imports не зависят от второго nested root |
| D11-02 | Go module/build | Реальные module-aware list/build работают, help/version без hardware side effects |
| D11-03 | Frontend scaffold | Совместимая toolchain, typecheck/lint/build действительно выполнены |
| D11-04 | Python scaffold | Package импортируется/проверяется без CV/native devices; contract negative cases работают |
| D11-05 | Commands | Failure exit codes не скрыты, fmt отдельно от fmt-check |
| D11-06 | Clean quickstart | Повторный запуск из чистой source среды по README, без secrets/Pi |

**Итог:** `docs/DAY_011_RESULT.md`, repository layout/tool manifest/README, проверяемый scaffold. Неразрешённые dependency downloads не закрываются старым успешным `experiments` replay.

---

## Д012. Конфигурация, lifecycle и logging

### Цель и вход

Hub превращается из entrypoint в управляемый процесс с deterministic config resolution, контекстом отмены, bounded startup/shutdown и структурированными логами. Вход: Д011, [ADR-006](../architecture/adr/006-config-data-ownership.md), [data ownership](../architecture/data-model-v0.md), Q-028/Q-030.

### Блок 12.1 — schema и ownership конфигурации, 60 минут

1. Определить typed config model: profile, server, simulation, component enablement, event queue limits, ML/MQTT endpoints, lighting bindings/safety defaults.
2. Разделить environment profile `development/test/production` и runtime adapter mode `simulation/hardware`; это разные оси, а не произвольные неявные aliases.
3. Составить ownership table каждого key: env secret/infrastructure, YAML hardware/defaults, будущий SQLite user setting. Не применять blind env>YAML>DB ко всем полям.
4. Выбрать maintained YAML parser и закрепить version после проверки API/license/security. Нужны strict unknown-field handling и согласованная политика duplicate keys.
5. Зафиксировать разрешённые env overrides только whitelist infrastructure keys. Env variable с похожим именем не переписывает user brightness/preset произвольно.
6. Определить поведение `.env`: development helper/explicit path допускается, production не загружает случайный файл из cwd. Логи не печатают полный env.
7. Для disabled components не требовать бессмысленные tokens; для реально enabled sensitive component missing secret — ошибка. Auth production requirements пока not implemented, production startup не должен открывать anonymous server.
8. Runtime user settings до Д016 представлены contract/provider/default, не фальшивой SQLite записью.

**Артефакты:** `internal/config` schema/loader design, `docs/development/configuration.md`, example profiles.

### Блок 12.2 — parsing и semantic validation, 90 минут

1. Реализовать чтение explicit config path, строгий YAML decode, default application и разрешённые env overrides.
2. Различить missing required config, malformed YAML, unknown/duplicate key и неверное значение. Сообщение содержит field path/code, а не raw secret line.
3. Проверить port/rate/duration/brightness/current ranges, finite numbers, positive counts и допустимые enum.
4. Проверить уникальность zone/device IDs, bounds/non-overlap bindings, сумму monitor side counts, matrix dimensions и отсутствие дублирующих physical owners.
5. Проверить semantic conflicts: simulation не требует реальных GPIO/device nodes; production не может молча переключиться на fake adapters; incompatible enabled backend — явная ошибка.
6. Не подставлять geometry120LED или PSU values из illustrative сценария как измеренную room config. Example simulation geometry должна быть помечена synthetic.
7. Создать immutable effective config snapshot с revision/источниками ключей и безопасным diagnostic view. Sensitive values отсутствуют в diagnostic JSON.
8. Invalid update/parse не даёт half-applied config; полноценный hot reload пока не нужен, restart-required fields обозначены.
9. Покрыть table-driven positive/negative cases, включая bool вместо number/NaN/overflow/пустые env/неправильный config path.

**Артефакты:** loader/validator, deterministic fixtures и config tests. Никакой actual sensor polling на этом этапе.

### Блок 12.3 — lifecycle coordinator, 90 минут

1. Ввести `Run(ctx, dependencies)` и module lifecycle interface с обязательным ownership каждого worker/resource.
2. Composition root создаёт root context, module start order и reverse cleanup. Ошибка третьего module останавливает первые два.
3. Signal handler реагирует на SIGINT/SIGTERM: закрывает admission, отменяет workers, ждёт bounded completion, возвращает понятный exit code.
4. Различить readiness state STARTING/RUNNING/STOPPING/FAILED и signal normal shutdown vs fatal startup failure.
5. Не ждать HTTP server, который ещё не реализован: lifecycle contract должен принимать будущую server component через интерфейс.
6. Исключить goroutine-per-event без владельца, endless retry loop, `context.Background()` внутри worker, отрывающий его от app stop.
7. Для startup/shutdown timeout выбрать проектные defaults и возможность override в tests. Timeout не означает, что process бесконечно «ещё завершается».
8. При forced stop отметить unknown in-flight outcomes; не заявлять graceful SQLite/LED fade, пока Storage/Lighting ещё нет. Hooks для будущих модулей предусмотрены явно.
9. Тестировать start failure/cancel/double stop с fake modules, barriers/channels и timeout guard, не с произвольными sleeps ради зелёного теста.

### Блок 12.4 — JSON logs и redaction, 60 минут

1. Настроить structured logger с level/time/component/event/correlation fields и единым error policy.
2. Зафиксировать allowlist startup/runtime fields; Config/env/API payload целиком не сериализовать.
3. Обернуть sensitive values так, чтобы обычные String/JSON/log paths давали redacted значение. Не полагаться только на regexp по слову `password`.
4. Проверить synthetic secret sentinel в parsed config, failed decode, error wrapping и `%v/%+v` diagnostic paths. Реальный секрет для теста не нужен.
5. Ограничить размеры внешних строк, стабилизировать log codes; request/producer IDs не становятся metric labels с бесконечной cardinality.
6. Сохранить DEBUG/INFO/WARN/ERROR distinctions. DEBUG не разрешает frame/PCM/token dump.
7. Определить stdout/stderr и exit behavior для CLI validation. Будущий log rotation — эксплуатационная задача, но формат/отсутствие secrets нужны сейчас.

### Блок 12.5 — process-level smoke и failure report, 60 минут

1. Запустить Hub с test/simulation config, проверить config snapshot/selected profile и закрытие по SIGTERM.
2. Запустить malformed/missing/production-with-unimplemented-sensitive-components cases: process отказывает объяснимо, сеть/hardware не открывает.
3. Создать start-failure fake module и убедиться, что предыдущие workers остановлены; повторный Run/Stop не оставляет handles.
4. Проверить logs machine-readable JSON и отсутствие secret sentinel во всём captured stdout/stderr.
5. Выполнить targeted unit/race tests, format/lint; сохранить command/result/known limitations.
6. Обновить config/lifecycle docs и report. Ни один fake module не называется настоящим ML/broker/SQLite.

### Проверки Д012

| ID | Проверка | Условие PASS |
|---|---|---|
| D12-01 | Ownership/precedence | Whitelist и профили применяются deterministic, user preferences не переписаны случайным env |
| D12-02 | Syntax/semantic errors | Invalid/duplicate/unknown/range/conflict config не стартует |
| D12-03 | Atomic startup/cleanup | Partial startup failure освобождает уже созданные ресурсы |
| D12-04 | Cancellation | SIGTERM/context cancel/double stop ограничены и без leaked workers |
| D12-05 | Redaction | Synthetic secret отсутствует во всех проверенных log/error paths |
| D12-06 | No hidden hardware | Simulation не открывает реальные устройства, production не fallback в fake |

**Итог:** `docs/DAY_012_RESULT.md`, config/lifecycle/logger code и meaningful tests. Успешный lifecycle ещё не означает готовый auth/SQLite/fade engine.

---

## Д013. HTTP/API, health/readiness, OpenAPI и первые метрики

### Цель и вход

Поднять реальный HTTP server поверх lifecycle, определить общий формат API и достоверный health model. Вход: Д012, [ADR-004](../architecture/adr/004-readiness.md), command/error contracts, Q-024/Q-031/Q-034. Auth и SQLite ещё не реализованы — это влияет на exposure и readiness, а не замалчивается.

### Блок 13.1 — HTTP boundary и routes, 60 минут

1. Выбрать минимальную router/server реализацию на совместимой Go version; stdlib `net/http` достаточно как начальный кандидат, внешняя dependency требует обоснования.
2. Подключить server к composition root: listen error возвращается lifecycle, context cancel вызывает bounded Shutdown, socket освобождается.
3. Default bind до auth — `127.0.0.1`, порт конфигурируется/в tests может быть ephemeral. Занятый порт — startup error, а не незаметный выбор случайного production порта.
4. Зарегистрировать только доступные endpoints: `/health`, `/ready`, `/metrics`, минимальный `/api/v1/version` либо согласованный metadata route.
5. Остальные product resources пока не регистрировать либо возвращать явный NOT_IMPLEMENTED/404 согласно выбранному API contract. Нельзя открывать незащищённый light/reset/token endpoint «для разработки» на LAN.
6. Определить read-header/read/write/idle/shutdown timeouts, header/body size bounds и allowed methods. Численные defaults выбрать документированно, tests могут уменьшать их.
7. Development loopback HTTP — инструмент локальной разработки. Production LAN HTTPS через proxy остаётся обязательным и не считается выполненным от plain HTTP server.
8. Зафиксировать, как worker fatal error/cancel сообщает app readiness/liveness и завершает server без утечки listener.

**Артефакты:** `internal/api/server.go`, routes/wiring, server configuration и API stage note.

### Блок 13.2 — единые responses/errors/request IDs, 90 минут

1. Ввести DTO response/error: safe code/message, request_id/correlation_id, retryable и optional details из allowlist.
2. Не использовать произвольный `err.Error()` внутренней зависимости как public message. Panic recovery не отправляет traceback, env, filesystem/private address или SQL details.
3. Определить JSON content type, UTF-8, пустое body/unknown fields и единицы измерений. Finite numbers/ranges валидируются на boundary.
4. Добавить request ID middleware: принимать bounded допустимый client ID либо генерировать новый. Он не является principal/auth доказательством.
5. Защитить logger от newline/oversized header injection, запрос с неизвестным ID не получает доступа к чужой operation.
6. Обеспечить consistent responses для unknown route/method, malformed request, timeout, unsupported content type и panic.
7. Собрать version/build/profile payload без secrets. Не выводить private key paths, токены, реальные network names и полный config.
8. Подготовить API auth/error placeholders: schema может описывать будущие scopes, но write routes не объявляются уже защищёнными без auth implementation.
9. Использовать `httptest`/test listeners для success/negative cases; actual HTTP server smoke отдельно от вызова handler напрямую.

**Артефакты:** API DTO/middleware/tests, `docs/api/error-contract.md`, синтетические examples.

### Блок 13.3 — dependency health evaluator, 90 минут

1. Реализовать dependency-check interface с enabled/critical/kind/status/reason/checked_at, bound deadline и cancellation.
2. Отделить process liveness `/health` от normal operation readiness `/ready`; optional dependency down не означает dead process.
3. В ready report учитывать все configured dependencies: storage/SQLite, hardware adapters, Runtime/MQTT и disabled/not-implemented states.
4. Не запускать медленные внешние calls синхронно без bound на каждый HTTP request; использовать bounded checker/cache с понятным stale policy и startup status.
5. Real SQLite ещё нет: до Д016 настоящий SQLite check имеет `NOT_IMPLEMENTED/NOT_READY`, не пустой файл с PRAGMA в памяти и не constant true.
6. На Д013 positive/degraded/critical matrix проверить injected dependency doubles в tests. Actual runnable process при обязательной отсутствующей dependency возвращает ready503.
7. На Д015 simulation может использовать explicit fake storage provider с `kind=memory_fake`, `persistent=false`, `simulated=true`. SQLite при этом не называется READY: это отдельный simulation dependency set.
8. Production config не допускает fake provider либо продолжение при обязательных NOT_IMPLEMENTED auth/storage/hardware components. Понятный отказ до exposure, не anonymous fallback.
9. Public health body минимален. Пока auth нет, подробные diagnostics доступны только loopback/tests; будущий authenticated System endpoint запланировать отдельно.
10. Проверить RUNNING+degraded, STARTING, STOPPING, stale timeout и SAFE_MODE semantics. Сам Safe Mode CLI/SQLite recovery реализуется позже, на Д013 проверяется health interpretation, не готовность аварийного управления.
11. Readiness evaluator не подменяет release verifier: full-profile required-service versions/smoke из ADR-004 остаются separate checks, не один HTTP200.

**Артефакты:** `internal/system/health`, dependency fakes для tests, readiness truth table и tests.

### Блок 13.4 — OpenAPI и минимальные HTTP metrics, 60 минут

1. Создать `docs/api/openapi.yaml`: versioned paths, methods, schemas, HTTP statuses, error envelope и health/profile semantics.
2. Описать реально implemented endpoints отдельно от planned resources. Swagger specification не должна обещать готовый `/auth`/`/zones` CRUD.
3. Добавить specification validation в developer check либо отдельную команду, используя закреплённый validator; ошибка specification вызывает failure.
4. Начать Prometheus-compatible `/metrics`: `http_requests_total`, `http_request_duration_ms`, process/app state counters по текущему scope.
5. Сохранить units из ТЗ: duration с суффиксом `_ms` измеряется в ms ровно один раз. Не назвать seconds миллисекундами либо умножить дважды.
6. Labels — bounded route template/method/status class. Raw path с user/device ID, request ID, usernames/IPs и payload не становятся labels.
7. Нереализованные метрики audio/ML не заполнять фиктивными healthy values. В каталоге отметить доступный subset и будущие source owners.
8. Отдельно проверить response/logging для `/metrics` и health: rate/detailed information не должны позволять раскрыть secrets.

**Артефакты:** OpenAPI v0, metric registry/middleware/tests, `docs/observability/metrics-v0.md`.

### Блок 13.5 — сквозной smoke и failure tests, 60 минут

1. Запустить compiled Hub с test config на loopback и проверить HTTP content/status/build marker фактическим client request.
2. Выполнить health matrix через injected dependencies: critical ready/critical failure/optional failure/disabled/stale checker.
3. Проверить actual missing SQLite stage readiness503 и liveness200. Positive test double не писать в отчёте как настоящую SQLite.
4. Проверить panic recovery, неверный метод/path, request ID sanitization и отсутствие synthetic secret в responses/logs.
5. Прервать процесс SIGTERM и убедиться, что listener закрыт, readiness до stop меняется, workers завершаются.
6. Запустить API/lifecycle targeted race tests, specification validation и metrics assertions.
7. Сохранить actual status codes, scope и blocker, если sandbox не позволяет bind loopback. In-process handler tests тогда остаются отдельным PASS, process network smoke BLOCKED.

### Readiness truth table Д013–Д015

| Контекст | Storage | Required adapters | ML/MQTT | Expected ready |
|---|---|---|---|---|
| Д013 actual stage | NOT_IMPLEMENTED SQLite required | Ещё не готовы | Optional | 503; liveness может быть200 |
| Д013 unit double | Explicit test READY | Explicit test READY | READY | 200 только в test scenario |
| Д013 unit double | READY | READY | Enabled optional unavailable | 200 + degraded |
| Д013 unit double | Unavailable/stale critical | READY | Любое | 503 |
| Д015 explicit simulation | memory_fake READY, persistent=false | Fake adapters READY | Fake/disabled, явный kind | 200 simulation-ready, не production |
| Production с fake/unimplemented components | Запрещённая комбинация | Не соответствует profile | Любое | Startup rejection/503; не fallback |
| STOPPING | Любое | Останавливаются | Любое | Не ready, новые commands не принимаются |

### Проверки Д013

| ID | Проверка | Условие PASS |
|---|---|---|
| D13-01 | Server lifecycle | Listen failure/cancel/Shutdown корректны, loopback exposure |
| D13-02 | API errors | JSON/status/code consistent, internal details/secret sentinel не выдаются |
| D13-03 | Request metadata | ID bounded/sanitized, не используется как auth/metric identity |
| D13-04 | Health matrix | Critical/degraded/disabled/NOT_IMPLEMENTED честно различаются |
| D13-05 | OpenAPI | Schema валидна и соответствует actual routes/statuses |
| D13-06 | Metrics | Реальные counters/units/bounded labels, не placeholder healthy telemetry |
| D13-07 | Process smoke | Реальный client/server проход и bounded shutdown; отдельно от handler-only tests |

**Итог:** `docs/DAY_013_RESULT.md`, actual HTTP server/health evaluator/contract/metrics и tests. Это не завершённый admin login, WS либо full System dashboard.

---

## Д014. Event Bus: bounded delivery, отмена, race и benchmarks

### Цель и вход

Реализовать внутрипроцессную событийную основу по [events v0](../architecture/events.md), чтобы slow UI/consumer не останавливал Core и будущий render loop. Вход: Д012 lifecycle/config/logger, typed event contract Д003, Q-030/Q-013. HTTP/WS transport не требуется для самого bus.

### Блок 14.1 — public interface/envelope, 60 минут

1. Определить `Event` с ID/version/type/producer/timestamps/boot_epoch/sequence/correlation/causation/resource/revision/payload.
2. Разделить trusted internal builders и external observations: неизвестный producer/actor из входного payload не становится trusted source.
3. Определить `Publish`, `Subscribe`, `Unsubscribe/Cancel`, `Close`, error/result semantics и ownership channel closure.
4. Subscriber получает read-only receive channel/stream и reason завершения, а не возможность писать/закрывать bus channel.
5. Согласовать typed payload builders или immutable serialization. Caller/subscriber не должен менять общий map/slice после Publish и портить другие consumers.
6. Ввести validation version/type/size/required IDs и bounded filter/resource keys. Raw camera/screen frames/DSP history в generic payload не допускаются.
7. Определить ordering: per-resource revisions генерирует владелец state, sequence принадлежит producer epoch. Bus не изобретает domain revision и не обещает глобальный total order внешних clocks.
8. Чётко указать, что bus RAM-only, restart replay/exactly-once/durable audit не обеспечивает.

**Артефакты:** `internal/events` interfaces/types/validators и краткий design/locking note.

### Блок 14.2 — subscribe/publish/cancel implementation, 90 минут

1. Создать bounded subscription registry/queues с config limits и explicit lifecycle owner.
2. Выбрать одну документированную locking/dispatch strategy; запретить callback/I/O/блокирующий consumer code под global bus lock.
3. Публикация должна иметь bounded behavior. Не создавать goroutine на каждое событие и unbounded backlog, скрывающий slow consumer.
4. Реализовать filter по типу/resource и отсутствие доставки post-cancel. Snapshot/filter data не меняются без синхронизации.
5. Защитить races publish vs unsubscribe/Close: один owner закрывает канал, send-after-close не возникает.
6. Cancellation subscriber освобождает registry/queue и завершает reader без ожидания следующего события.
7. Поддержать app context shutdown и idempotent Close. Publish после закрытия возвращает понятную ошибку, не panic.
8. Структуры payload копировать/делать immutable до передачи; несколько consumers не разделяют доступную для записи byte slice без contract.
9. Ограничить число subscriptions/keys; invalid/excessive subscribe возвращает отказ, а не растущую map.
10. Покрыть producer/subscriber behavior deterministic synchronization primitives. Timeout использовать как аварийный guard, не основной способ упорядочить тест.

### Блок 14.3 — overflow policy по классам событий, 90 минут

1. Для gauges/realtime telemetry реализовать coalescing/latest-value в bounded resource scope: старые значения можно заменить свежими, факт потери учитывается.
2. Для control state changes overflow обязан дать explicit gap/resync/termination reason. Нельзя молча выкинуть update и оставить consumer уверенным в актуальном состоянии.
3. Для gesture/beat/trigger events не применять blindly gauge coalescing. Если очередь заполнена — явный drop/error counter; старый gesture не исполняется через минуту после освобождения.
4. Per-subscriber policy не должна блокировать других subscribers и publisher. Consumer lifecycle не определяет успех уже принятой command.
5. Security/automation audit не доверять bus как durable store: будущие modules пишут обязательный record через Storage, bus лишь уведомляет.
6. Начальные queue capacities задать как project candidates в config, затем сравнить по benchmark. Например control64/latest1-per-key — гипотеза, не новое требование ТЗ.
7. При latest-value per-key ограничить и количество ключей, иначе bounded channel с unbounded map всё равно даёт leak.
8. Добавить counters subscriptions/queue drops/overflows/rejected events; labels bounded, не event_id/user_id.
9. Проверить, что marker `RESYNC_REQUIRED` пока внутренний termination/error contract, а не обещание готового WS snapshot protocol Д023.
10. Явно записать, какие semantics отложены до domain consumers: external event dedupe/trigger lineage/automation loops. Не делать произвольный глобальный dedupe, теряющий законные repeated events.

### Блок 14.4 — race/concurrency/negative tests, 60 минут

1. Несколько publishers и subscribers: проверять completeness/ordering в границах обещанного contract, не случайный interleaving всех goroutines.
2. Slow subscriber: намеренно не читать его queue, продолжать publish и проверять progress fast consumer.
3. Publish/Unsubscribe/Close одновременно: много повторов с barriers, race detector, отсутствие panic/deadlock.
4. Caller мутирует исходный payload после Publish; другой consumer изменяет полученную копию — соседние recipients не меняются.
5. Cancel до subscription, во время ожидания, после Close; повторные Cancel/Close должны быть определены.
6. Reject unknown version/oversized/forbidden fields и invalid capacities, nil/zero state cases по выбранному API.
7. Overflow каждого класса проверять отдельно: latest telemetry, control resync и trigger drop — разные outcomes.
8. Убедиться, что test completion освобождает все принадлежащие bus workers; сравнение только общего числа goroutines без ownership не считать достаточным evidence.

### Блок 14.5 — microbenchmarks и baseline report, 60 минут

1. Добавить publish benchmark без subscribers, с1/10/100 subscribers и с одним permanently slow consumer, с корректным cleanup.
2. Измерить ns/op, allocs/op, throughput, overflow counts и выбранный upper bound resource use. Указать Go/OS/CPU/queue sizes.
3. Для constrained scenario проверить bounded latency/progress, не выдавая microbenchmark publish за WS<100ms или LED30FPS.
4. Сравнить candidate queue capacities на одинаковом fixture; выбрать обоснованный default и пределы конфигурации.
5. Проверить bounded memory на коротком event burst, no goroutine-per-message, стабильность registry после subscribe/unsubscribe cycles.
6. Выполнить `go test -race` и bench отдельно: profiling/race overhead не сравнивать с normal performance run.
7. Сохранить actual raw/summary outputs, known limitations и обновить event contract, если implementation потребовала уточнения.

### Проверки Д014

| ID | Проверка | Условие PASS |
|---|---|---|
| D14-01 | Envelope/API | Validation/ownership/ordering заданы и исполняются |
| D14-02 | Basic delivery/filter | Нужные consumers получают событие, лишние — нет |
| D14-03 | Slow consumer | Publisher/fast consumer продолжают работу, очередь bounded |
| D14-04 | Class overflow | Latest/control/trigger policies различаются и наблюдаемы |
| D14-05 | Cancellation/close | Concurrent stop/unsubscribe безопасны, resources освобождены |
| D14-06 | Payload isolation | Caller/consumer mutation не меняет чужую доставку |
| D14-07 | Race/benchmark | Реальные race tests и воспроизводимые baseline numbers |

**Итог:** `docs/DAY_014_RESULT.md`, bus code/tests/benchmarks, `docs/research/event-bus-baseline.md`. Hardware-independent stage может быть полностью проверен на host; product realtime и durable rule behavior от этого ещё не реализованы.

---

## Д015. Fake adapters, simulation, dev Compose и CI v0

### Цель и вход

Собрать предыдущие дни в работающий vertical slice: **fake sensor → typed observation → Core Event Bus → проверенный consumer**, плюс reproducible lint/test/build и container smoke там, где Docker доступен. Вход: Д011–Д014, [hardware interfaces](../../Smart_Home_Hub_TZ.md), simulation requirements, Q-026/Q-030/Q-035.

### Блок 15.1 — adapter ports и contracts, 60 минут

1. Определить узкие interfaces для device lifecycle/status, sensor read/stream, LED transport/frame, audio feature source, presence и screen input.
2. Согласовать context cancellation, timestamps/units/quality, error/offline и ownership resources. Unknown measurement не равен нулю.
3. Ports принадлежат domain/adapter boundary, не package GPIO-driver library. API/automation не получают `/dev` handle через удобный getter.
4. Для LED выбрать frame-level interface либо адаптацию SetPixel/Show из примера ТЗ с atomic frame semantics. Глобальный limiter принадлежит будущему Lighting Engine, не fake source.
5. Определить factory/wiring: `adapter_mode=simulation` выбирает только fake adapters; отсутствие hardware в production — error, не auto-fake.
6. In-memory simulation state обозначить `memory_fake`, `persistent=false`; fake health/storage не называется проверенной SQLite и запрещён production validator.
7. Чётко отделить test-only DI doubles от user-selectable simulation adapters. Тестовый всегда-ready checker не должен случайно попасть в production binary composition.
8. Зафиксировать capability/status contracts и минимальное покрытие каждого fake к Д015; полный симуляционный UI/сценарии дополняются на Д024.

**Артефакты:** interfaces, simulation factory, adapter contract note и versioned seed/scenario schema.

### Блок 15.2 — deterministic fake implementations, 90 минут

1. Реализовать seeded fake sensor с несколькими measurements, units/observed_at/quality/sequence и explicit stale/offline/error transitions.
2. Связать его с Core event builder: emitted measurement валидируется и превращается в `SENSOR_UPDATED`; observer в test проверяет resource/value/unit/epoch, не просто число logs.
3. Создать fake LED transport, который принимает validated frame в RAM, хранит только latest applied frame и технические counters; actual hardware libraries не импортируются.
4. Для virtual MATRIX поддержать geometry/mapping contract на synthetic input. Graphical visualizer не писать вместо минимального data representation — он запланирован позже.
5. Добавить минимальные fake audio features, presence state sequence, screen color input и MQTT-device observations через общий contract. Это не playback/FFT/person inference/screen capture/broker.
6. Использовать общий fake clock или deterministic step API для tests; production fake generator может иметь bounded ticker, принадлежащий lifecycle.
7. Error injection сделать явным scenario input: delayed start, offline, invalid sample, disconnect/reconnect, worker failure. Не менять global clock/host network/sysfs.
8. Проверить Start/Stop/status и отсутствие goroutines после cancel. Каждая adapter queue bounded, timestamps/seed не зависят от random host state.
9. Исключить persistent history кадров/PCM/features. Synthetic latest frame допустим в RAM; public fixture generator не содержит personal data.
10. Не подключать реальный MQTT/BlueZ/camera «потому что устройство обнаружилось». Simulation должна быть безопасна и на компьютере с такими devices.

### Минимальный scope fakes на Д015

| Fake | Что требуется сейчас | Что завершается позже |
|---|---|---|
| LED strip | Frame validation, latest frame/counters | Lighting effects/limiter/visual UI |
| LED matrix | Width/height/orientation/serpentine mapping fixture | UI visualizer, real panel |
| Sensors | Seeded valid/stale/offline events | Все реальные drivers/calibration |
| Audio | Typed synthetic spectrum/beat/BPM inputs | Playback/source controls/FFT correctness |
| Presence | Все4 итоговых states по сценарию | Fusion из реальных sources |
| Screen-client | Bound synthetic colors/epoch/sequence/error | Pairing/real capture/30FPS/e2e |
| MQTT device | Typed discovery/state/disconnect observations | Mosquitto/auth/ACL/network autodiscovery |
| Storage | Explicit nonpersistent simulation provider | SQLite/WAL/migrations/transactions |

Каждый fake должен делать реальную bounded contract operation, а не быть пустым методом `return nil`. Но это ещё не полная simulation acceptance M1: неготовый scope в README/report остаётся видимым.

### Блок 15.3 — native simulation и dev Compose, 90 минут

1. Собрать `run-sim` native path: config Д012→app lifecycle→bus Д014→fake adapters, HTTP stage Д013 и controlled stop.
2. Реально проверить fake sensor→Core→subscriber с deterministic expected event. Internal assertion/test consumer не превращается в public unauthenticated control endpoint.
3. Readiness simulation учитывает готовый fake state provider/adapters и возвращает явные profile/simulated/persistent flags. Production остаётся отвергающей unimplemented/fake combination.
4. Создать multi-stage development Dockerfile Core с non-root runtime, ограниченными writable paths и healthcheck, реально доступным внутри final image.
5. Подготовить `docker-compose.dev.yml` без `/dev`/host networking/privileged и без mounts реальных credentials/SSH/CA/DB. Sources/build caches — только по необходимости dev flow.
6. Различить bind внутри container и host exposure: внутри private container network можно listen0.0.0.0, но host published ports — `127.0.0.1:...`; это не открытие anonymous LAN API.
7. Включать только реально существующие services этого среза. Frontend placeholder может быть optional dev profile; ML contract/неготовый bot не имитировать healthy production containers. Real broker/Telegram/Grafana ещё не нужны для fake flow.
8. `depends_on` и healthchecks настроить по фактической criticality: running container не равен ready service, optional worker outage не делает core permanently unavailable.
9. При доступном Docker выполнить build→up→HTTP ready/sensor scenario→SIGTERM/down и cleanup. Проверить restart/reconnect и отсутствие lingering containers/volumes с private data.
10. Если daemon/network/build registry недоступны, выполнить native equivalent и Compose validation отдельно: `PASS-HOST`/`DOC-ONLY`, а `PASS-COMPOSE` оставить BLOCKED с причиной. YAML parse не заменяет container start.

**Артефакты:** native simulation runner/scenario, Dockerfiles/dev Compose, smoke script, Docker prerequisites/status.

### Блок 15.4 — GitHub Actions CI v0, 60 минут

1. Подготовить PR/push workflow, выполняющий те же meaningful commands, что developer `check`; не поддерживать разные скрытые списки локальных/CI checks.
2. Go job: format check, vet/lint, tests с race, build Hub. Static/security tools подключать с закреплёнными versions по готовому scope; network scanner failure виден, не suppress в success.
3. Frontend job: lockfile install, ESLint, TypeScript check, build. Unit job запускает реальные тесты business/helper logic, когда она есть; пустое `--passWithNoTests` не считается выполнением требования unit coverage.
4. Если frontend пока только scaffold без содержательной testable logic, отметить unit часть `NOT_APPLICABLE_AT_THIS_STAGE` с follow-up Д022–Д025, не создавать бессмысленный тест статического placeholder ради зелёного числа. Полный §43 frontend test requirement остаётся открытым.
5. Python job: Ruff/mypy/pytest для `ml/common` реального contract package. Native CV/DSP/hardware dependencies не нужны для этого job; existing experiments — optional separate job/command.
6. Все tools/runtimes/actions versions закрепить совместимо с project manifests; third-party Actions по проверенному commit SHA. Default permissions — минимальные read-only contents.
7. PR/fork workflow не использует production self-hosted runner, production secrets, broad write token или checkout непроверенного PR через privileged `pull_request_target`.
8. Добавить job timeout/concurrency cancel для устаревших runs, redacted reports и dependency cache с корректными lockfile keys; cache не хранит env/keys/DB/user models.
9. Проверить workflow syntax/schema/lint и controlled failing command. Не использовать `continue-on-error` на обязательных checks для иллюзии готовности.
10. Если repo/remote credentials не заданы — закончить workflow files и local checks, actual GitHub run пометить BLOCKED. `PASS-CI` только после действительно выполненного remote run на известном commit.

**Артефакты:** `.github/workflows/ci.yml`, pinned tool setup, check mapping и CI execution/limitations report. Production tag deploy/self-hosted runner создаются позже, не этим PR pipeline.

### Блок 15.5 — итоговый acceptance и передача на Д016, 60 минут

1. В чистой source среде выполнить tool setup→lockfile install→check→build→native simulation→stop. Отметить первый network install отдельно от повторного cached run.
2. Пройти sensor success/stale/offline, slow subscriber, invalid config, optional fake failure и cancelled shutdown; assertions ожидают поля/status, не просто живой процесс.
3. Проверить, что simulation не открыла реальные hardware handles и не потребовала личных tokens. Production validation не принимает fake/неimplemented components как готовые.
4. Выполнить доступный Compose smoke и/или сохранить exact blocker; remote CI status также фиксировать отдельно.
5. Сверить OpenAPI/README/config examples с фактическими routes/flags/commands; убрать stale обещание, что auth/UI/SQLite уже готовы.
6. Собрать report по пяти дням с actual code/tests/build artifacts/versions и неготовыми частями. Не считать 20 spike tests Д004–Д010 заменой новых Core tests.
7. Обновить backlog/вопросы/ADRs только в подтверждённой части. Hardware M0_NOT_READY и platform capture blockers остаются.
8. Передать Д016: storage interface expectations, schema v0, migration/transaction/retention constraints, готовые health hooks для настоящей SQLite и список meaningful integration tests.

### Проверки Д015

| ID | Проверка | Условие PASS |
|---|---|---|
| D15-01 | Adapter boundaries | Core/fakes работают через interfaces без driver-specific dependency |
| D15-02 | Sensor→Core vertical slice | Validated deterministic event с expected payload дошёл до real bus consumer |
| D15-03 | Fake behavior/lifecycle | Minimal fake operations работают, offline/error/cancel без утечек |
| D15-04 | Mode separation | Fake возможен только явно в simulation, persistent=false/production restrictions честны |
| D15-05 | Native smoke | Реальный start/ready/event/stop без Pi/secrets |
| D15-06 | Compose | Только actual build/start/smoke/cleanup даёт PASS-COMPOSE; иначе BLOCKED |
| D15-07 | Local quality | Go/TS/Python commands проверяют реальный scope и не подавляют failures |
| D15-08 | Workflow protection | No production self-hosted/secrets, pinned actions/minimal permissions |
| D15-09 | Remote CI | Реальный run даёт PASS-CI; файлы/local tests — не remote execution |
| D15-10 | Handoff | Scope/limits/commands/report готовы для SQLite/auth этапа |

**Итог:** `docs/DAY_015_RESULT.md`, Core simulation slice, fake adapters/dev Compose/CI workflow и factual evidence. Это первый software skeleton, не полная система M1/v1.0.0.

---

## 4. Предполагаемая структура после Д015

Пути ниже — ожидаемые новые artifacts, не утверждение, что они созданы подготовкой плана. Существующие `docs`, `hardware`, `experiments` сохраняются.

```text
Smart_home_hub/
├── go.mod / go.sum при наличии external deps
├── Makefile
├── README.md
├── .env.example
├── config.example.yaml
├── cmd/
│   ├── hub/main.go
│   └── screen-client/          # место/явный stub, не готовый capture client
├── internal/
│   ├── app/                    # composition + lifecycle
│   ├── config/                 # loader/validation
│   ├── api/                    # HTTP/DTO/middleware
│   ├── events/                 # bounded bus
│   ├── devices/                # adapter ports/status
│   ├── simulation/             # factories/fakes/scenarios
│   └── system/                 # health/build metadata
├── ml/
│   ├── pyproject.toml
│   └── common/                 # minimal real observation contract
├── frontend/
│   ├── package.json + lockfile
│   ├── tsconfig / eslint / vite config
│   └── src/                    # minimal honest shell
├── telegram/                   # boundary/explicit stub only
├── configs/                    # validated examples/profiles
├── deploy/docker/              # development image files
├── docker-compose.dev.yml
├── .github/workflows/ci.yml
├── tests/integration/          # software vertical slice
├── docs/api/ / development/ / observability/
├── hardware/                   # existing draft + future HIL
└── experiments/                # existing feasibility, not product dependency
```

Не требовать `go.sum` без external modules лишь ради наличия файла; не создавать пустые package API без смысла. Full repository structure §41 разворачивается по мере появления реальных модулей.

## 5. Два уровня готовности и открытые зависимости

| Область | Можно доказать к Д015 | Что остаётся открытым |
|---|---|---|
| Monorepo/build | Module/locks/commands/build | Public remote identity/CI run при отсутствии доступа |
| Config/lifecycle | Typed validation, cancellation, redacted logs | Настройки ещё не реализованных modules и реальное hardware shutdown |
| HTTP/health | Реальный server, truth table, stage/simulation semantics | Admin auth/SQLite readiness/production LAN access |
| Event Bus | Delivery/overflow/cancel/race/bench | Real consumers, WS snapshot/revoke, durable audit storage |
| Simulation | Fake sensor→Core, latest fake frame, seeded scenarios | Full UI visualizer/E2E, real Lighting/Audio/Automation engines |
| CI/Compose | Local checks и actual available container run | Remote GitHub run/daemon registry access при отсутствии доступа |
| Raspberry Pi | Только prepared hardware boundaries | Весь physical M0/HIL, untouched by software PASS |

### Старт Д016

Для SQLite этапа нужны: существующая Core composition, storage port, migrations folder/schema v0, config paths, health criticality, тестовый isolation/cleanup и common errors. Реализовать SQLite WAL/миграции вместо fake provider, проверить schema failure/Safe Mode prerequisites. Simulation provider не должен незаметно стать production persistence layer.

Auth следует Д018–Д020. До него network controls остаются закрытыми/не реализованными; нельзя вынести ранний scaffold в LAN под видом готового local-first Hub.

## 6. Работа при tool/network/Docker ограничениях

| Ограничение | Что всё ещё можно сделать | Что не отмечать PASS |
|---|---|---|
| Go launcher сломан | Проверить доступный normal binary/user-local toolchain и зафиксировать workaround | Build, если компиляция реально не запускалась |
| Нет dependency network/cache | Подготовить manifests/locks только из реально разрешённых sources, contracts/native stdlib checks | Install/security scan, которые не выполнились |
| Node несовместим с Vite | Выбрать compatible local/container toolchain, проверить engines | Frontend build на неподходящем Node |
| Docker CLI есть, daemon недоступен | Native simulation, Docker/Compose syntax/design | PASS-COMPOSE из `compose config` |
| Нет GitHub repo/credentials | Workflow files/local equivalent/syntax lint | PASS-CI или защищённые branch settings |
| Sandbox не даёт socket bind | In-process handler/integration tests, запрос разрешённого isolated loopback при исполнении | Process HTTP smoke без listener |
| Нет Pi | Весь независимый software scope этих дней | Physical M0/audio/LED/thermal acceptance |

Сначала исправлять воспроизводимую причину в пределах project scope. Не менять глобальный Node, system audio, trust store или host Docker permissions без необходимости, вытекающей из задачи. Блокер сдвигает зависимый criterion, не превращает tests в `echo OK`.

## 7. Шаблон дневного отчёта

```text
День / дата / scope:
Source commit или manifest файлов:
Tool/dependency versions:
Выполненные блоки и реальные новые файлы:
Проверки Dxx-yy, command и exit code:
Evidence level: HOST / SIM / COMPOSE / CI
Actual outputs и существенные scalar measurements:
Negative/failure cases:
Данные/секреты/устройства, которые использовались:
Оставшиеся BLOCKED/NOT_RUN и точная причина:
Изменение ADR/requirements/backlog:
Фактическое время и расход резерва:
Что нужно для следующего дня:
```

Галочки ставятся по фактическому выполнению. Target/CI/Compose evidence хранится отдельно, health stub не становится production Core. Private inputs и полные personal logs не коммитятся вместе с report.

## 8. Связь с требованиями и вопросами

| Область | Требования | Вопросы / ADR |
|---|---|---|
| Архитектура/monorepo | ARCH-002/ARCH-004/ARCH-006…ARCH-010 | Q-026, ADR-001/007 |
| Config/storage boundary | STORE-001…STORE-008 | Q-028, ADR-006; actual SQLite с Д016 |
| HTTP/API/health | API-001…API-004, OBS-001…OBS-005 | Q-024/Q-031/Q-034, ADR-004 |
| Event contracts/privacy | ARCH-008, PRIV-001…PRIV-003 | Q-021/Q-030, events v0 |
| Simulation | SIM-001…SIM-007, DEV-001/DEV-002 | Q-025/Q-030, ADR-001/007 |
| CI/tests/build | CI-001…CI-011, TEST-001…TEST-009 | Q-026/Q-035; полный pipeline позже |

Покрытие области в плане не означает выполнение всех связанных требований за эти пять дней. Например full CI security/multi-arch packaging, auth и real hardware acceptance остаются запланированными этапами.

## 9. Первичные источники для реализации

Проверены при подготовке25.10.2026; exact installed/dependency versions уточнить перед исполнением.

- [Go dependency management](https://go.dev/doc/modules/managing-dependencies): product module/manifest/dependency handling; experiment module-off настройки не переносятся автоматически.
- [Go net/http](https://pkg.go.dev/net/http): server/context/Shutdown semantics для выбранной Go version.
- [Go race detector](https://go.dev/doc/articles/race_detector): meaningful race checks, совместимый runtime/compiler; successful plain tests не заменяют race run.
- [Vite guide](https://vite.dev/guide/): актуальная runtime/tool compatibility перед scaffold. Host Node version из старого inventory не принимается на веру как suitable.
- [Python asyncio tasks](https://docs.python.org/3/library/asyncio-task.html): cancellation/task ownership, если minimal Python worker использует asyncio; сверять с выбранной Python version, а не бездумно копировать latest API.
- [Compose startup/shutdown order](https://docs.docker.com/compose/how-tos/startup-order/): запуск контейнера не означает готовность зависимости; health conditions требуют реальной проверки.
- [GitHub secure use](https://docs.github.com/en/actions/reference/security/secure-use): минимальные permissions, untrusted PR boundaries, Actions pinning и secrets.

Эти документы обосновывают технические проверки, но не заменяют actual build/tests/CI evidence. Новые команды, библиотеки и workflows этим планом не установлены и не запущены.
