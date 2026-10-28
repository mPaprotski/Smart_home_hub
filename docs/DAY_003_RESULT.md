# День 3 — результат архитектурной проработки

Дата: 25 октября 2026 года. Статус: **документальная задача Д003 выполнена**. Приняты рабочие проектные решения v0 без изменения обязательного объёма ТЗ. Реализация, программные тесты и аппаратные доказательства пока впереди.

Вход: [ТЗ](../Smart_Home_Hub_TZ.md), [требования](requirements.md), [День 2](DAY_002_RESULT.md), [открытые вопросы](open-questions.md).

## Материалы

- [Архитектурный индекс](architecture/README.md) — границы системы, связи документов, статусы решений.
- [Модель сущностей](architecture/domain-model.md) — пользователи/devices/zones/presets/audio/sensors/presence/camera/rules/overrides/habits и общие инварианты.
- [Логическая схема данных v0](architecture/data-model-v0.md) — таблицы/FK/indexes, ownership, RAM-only данные, transaction boundaries, retention и privacy delete/restore.
- [Командный контракт](architecture/commands.md) — gateway/auth/arbitration/idempotency/outcomes/delayed work.
- [Событийный контракт](architecture/events.md) — envelope/catalog/delivery/backpressure/WS resync/privacy.
- [Словарь и схемы состояний](architecture/state-machines.md) — lifecycle/ready/Safe Mode/devices/audio/camera/screen/presence/jobs/learning/shutdown.
- [20 сценариев](architecture/scenarios.md) — ожидаемое поведение при конфликтах/отказах и будущие проверки.
- [Команды JSON](architecture/examples/commands.json), [события JSON](architecture/examples/events.json), [health JSON](architecture/examples/readiness.json) — синтетические примеры, не настоящий runtime traffic.

## Семь ADR

| ADR | Что определено | Что ещё требует доказательства |
|---|---|---|
| [001](architecture/adr/001-module-boundaries.md) | Core owns hardware/storage; Runtime/bot через ограниченные ports | Imports/mounts/auth, real/fake substitution, failure isolation |
| [002](architecture/adr/002-command-arbitration.md) | Все mutations через общий arbiter; manual scope/TTL, priority, jobs | Concurrency/clock/race/load tests и реальные effects |
| [003](architecture/adr/003-camera-lifecycle.md) | Один owner, отдельные permissions, current-boot arm, indicator order | Capture stack/indicator/две camera types, crash cleanup |
| [004](architecture/adr/004-readiness.md) | Critical ready отдельно от degraded и required-service release smoke | Health endpoints и broken deployment/rollback |
| [005](architecture/adr/005-sqlite-safe-mode.md) | Emergency local CLI/Unix peer auth, no anonymous LAN fallback | Socket/container OS identity, failure/recovery HIL |
| [006](architecture/adr/006-config-data-ownership.md) | Config owners, minimal habits, retention, privacy epoch/managed backup purge | Migrations, local workers, actual delete/restore/backup integration |
| [007](architecture/adr/007-led-topology.md) | Logical zones отделены от transport; SPI chain candidate | Counts/сигнал/audio coexistence/non-root/physical hard-off на Д004/Д008 |

## Результаты разбора конфликтов

- Ручной Master OFF создаёт `lighting:*` hold и не отменяется вернувшимся screen stream.
- User-selected live mode продолжает render под override; lease защищает control fields, а не замораживает все RGB frames.
- При shutdown fade OFF не записывается как last desired OFF; после boot LED settings восстанавливаются, music STOPPED, camera OFF.
- Presence использует камеру только с отдельным разрешением; после Core boot нужен current-boot arm. Не обещано физическое обнаружение сгоревшего indicator без feedback circuit.
- SQLite failure не открывает сеть: manual control доступен локальному OS-authorized operator; recovery explicit и не отменяет emergency OFF незаметно.
- Delayed absence revalidates guard; UNKNOWN не AWAY. Missed jobs/schedules по default не catch-up.
- Exactly-once для hardware не заявлен: uncertain NextTrack не replay, absolute desired states reconcile.
- Habit delete invalidates old training/model epoch; Hub-managed backups и restore учитывают privacy generation, внешние copies вне доступа не объявляются удалёнными.

## Что проверено в документах

- [x] Созданы все обещанные contracts, state machines, data model и семь ADR.
- [x] Разобраны 20 сценариев с expected outcomes и этапом будущей проверки; product-test status NOT_RUN.
- [x] JSON-примеры синтаксически корректны, envelopes согласованы с обязательными полями/каталогами, scopes/epochs/outcomes не подменены secrets.
- [x] Относительные links, requirement/Q/ADR IDs и Markdown code fences проверены; основные diagrams доступны как Mermaid source.
- [x] Backlog/план/реестр вопросов обновлены; hardware provisional status и отсутствие замеров сохранены.
- [x] Исходное ТЗ, артефакты Д002 и исторические отчёты Д001/Д002 не переписаны.

Внешний Mermaid renderer и работа реальных services не проверялись: наличие корректного source не выдано за UI screenshot или исполняемый архитектурный тест. Schema v0 — логическая модель, не готовый migration SQL/OpenAPI.

## Открытые границы

Архитектурные решения по Q-001/Q-017/Q-031/Q-032 зафиксированы, реализационные критерии закрытия ещё впереди. По Q-002/Q-005/Q-006/Q-009/Q-012/Q-014/Q-019/Q-021/Q-022/Q-023/Q-024/Q-028/Q-030 уточнена часть контракта. Полный реестр остаётся из 35 вопросов, без ложного закрытия HIL/ML/platform gates.

Exact OS/drivers/audio graph/Scene runtime, measured quality/performance, BOM counts/power wiring, thresholds и численные queue/log/auth defaults выбираются в назначенные дни. Принятые проектные defaults 900sec override, 120sec pairing, 24h completed-command metadata не являются новыми требованиями ТЗ; они могут быть изменены с обновлением ADR/contracts.

## Следующий шаг

**Д004 — LED/audio spike:** проверить independent logical zones, подходящий transport и обязательные outputs/общий PCM path, права host/container. Основа — ADR-007 и hardware draft Д002. При недоступной Pi подготовить протокол эксперимента, оставив реальный запуск явно открытым. В рамках Д003 эксперимент не выполнялся.
