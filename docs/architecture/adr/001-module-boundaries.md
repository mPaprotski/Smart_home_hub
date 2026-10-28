# ADR-001. Модули и владение ресурсами

Дата: 25.10.2026. Статус: **принято как проектное решение, реализация впереди**.

Источник: §2/7/8/25/31/40/42/47/54 ТЗ. Требования: ARCH-001…ARCH-012, TG-002, STORE-001…STORE-003, PRIV-001…PRIV-003. Вопросы: Q-001/Q-006/Q-024/Q-030.

## Контекст

Несколько интерфейсов управляют одним hardware. Прямые вызовы adapters из bot/ML/rules создадут разные проверки прав, конфликтующие состояния и обход manual override. При этом небольшому проекту не нужна распределённая шина или десятки сервисов.

## Решение

Модульный Go Core и отдельный локальный Python Runtime сохраняются. Core владеет SQLite и hardware/host-audio adapters через интерфейсы. Frontend, bot, screen-client и Runtime не открывают SQLite и не управляют LED непосредственно. Runtime владеет локальным capture/inference/DSP worker memory и отправляет validated observations; authority camera permissions/gesture mappings остаётся в Core.

| Модуль | Владеет | Допустимые зависимости |
|---|---|---|
| API/Auth | Sessions/scopes/gateway metadata | Command/query ports; Auth repositories |
| Dispatcher | Admission, ordering, overrides, outcomes | Policy ports, typed handlers, Storage unit-of-work |
| Lighting | Zones/presets/render/current limiter | Transport interfaces, typed realtime inputs, repositories |
| Audio | Source/output/playback | Host audio adapters; PCM port для DSP |
| Sensors/Devices | Inventory/availability/measurements | GPIO/I²C/SPI/MQTT adapter ports |
| Presence | Fusion/evidence freshness | Validated observations, no camera raw frames in Core |
| Automation/Predictions | Rules/jobs/proposals/feedback | State query ports, command dispatcher, own repositories |
| System | Config/health/thermal/backup/recovery | Lifecycle/adapters/Storage interfaces |
| Events | Typed pub/sub envelope | Не зависит от конкретных modules |

Для команды typed synchronous handler dependency допустима и нужна, чтобы получить result. Event Bus используется для наблюдений/изменений/подписок, не как бесконтрольный замкнутый request bus. Общие domain types не содержат драйверных GPIO library types.

Обычные services могут вызывать только разрешённые API/ports. Internal Docker network само по себе не даёт admin: service identity, allowlist payload и минимальные scopes обязательны в реализации. Конкретный service-auth transport/token format остаётся Д017–Д020.

## Альтернативы

Прямое управление adapters из каждого интерфейса — отклонено из-за дублирования policy. Отдельный microservice на каждую подсистему — лишняя эксплуатационная/ресурсная цена для Pi4. Полный event sourcing — не требуется и повышает риск хранения запрещённых realtime данных.

## Последствия и проверка

Нужен common command dispatcher и отделение facts от commands. Real/fake adapters подключаются без изменения business logic. WAN/bot/ML outage не разрушает Core local controls. Проверки: imports/mounts/ports review, negative scope tests, fake/real adapter substitution, Runtime/bot failure cases Д015–Д025/Д126. Scene inference runtime остаётся Q-006 до Д005/Д071; здесь принят его local-only boundary, не библиотека.

## Уточнение26.10.2026: control plane на Pi4B/2GB

Принята [политика2GB и degradation priorities](../../development/pi4b-2gb.md).
Pi — единственная2GB плата/control plane/hardware hub; laptop владеет screen capture,
Ambilight analysis и Scene ML. Один лёгкий Python Runtime на Pi, lazy model loading,
bounded workers, проверяемый RSS release. SQLite/Compose; PostgreSQL/Kubernetes и
тяжёлые Python workers вне scope. Automatic memory-pressure controller ещё не написан;
это архитектурное решение, не runtime PASS. Safety/auth не деградируют как optional ML.
