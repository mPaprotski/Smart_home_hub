# Raspberry Pi 4B: только2GB RAM

26.10.2026 владелец подтвердил: доступна только Raspberry Pi 4 Model B с2GB RAM.
Это обязательный hardware baseline, а не вариант покупки. Доступ по SSH и фактический
MemTotal пока не получены. Прежние measured evidence/hash относятся к предыдущим
срезам; ТЗ изменено сейчас по явному указанию владельца.

## Насколько это ограничивает проект

Для control plane/SQLite/LED/sensors/MQTT маленький footprint остаётся разумной целью.
Это инженерная оценка, не подтверждённый hardware PASS. Сильнее ограничены одновременные
Python ML/CV/DSP процессы, загрузка нескольких моделей, большие frames/queues, обучение,
build на самой Pi и полноценный monitoring stack. Требования не вычёркнуты; пик общей
RAM и CPU должен измеряться на2GB плате, а model selection зависит от этих измерений.
Нельзя обещать, что весь функционал и все FPS/latency targets заработают одновременно,
только на основании нынешних x86 тестов.

## Начальный бюджет

[Машиночитаемый профиль](../../configs/pi4b-2gb-budget.example.json) — план,
не результат замера и не production manifest.

| Группа | Плановый объём, MiB | Статус |
|---|---:|---|
| ОС/kernel/Docker | 512 | Измерить на Ubuntu headless |
| Core | 128 | Hard cap dev container, soft Go limit96MiB |
| Python ML/CV/DSP | 384 | Кандидат; реальный runtime ещё не реализован |
| Host audio | 192 | Кандидат, нужен совместный граф |
| MQTT | 48 | Кандидат |
| Bot | 32 | Кандидат |
| Static frontend/proxy | 32 | Кандидат; без Node/Vite server |
| Остальное | 128 | Кандидат |
| Итого | 1456 | Не гарантия потребления |

Плановый working envelope <=1536MiB оставляет номинальный резерв512MiB. Реальная
доступная память меньше2GiB: приёмка использует фактический MemTotal/MemAvailable,
а не сумму container caps. Начальная цель MemAvailable>=256MiB под нагрузкой требует
проверки/уточнения на плате. Не учитывать host memory и container RSS дважды.

## Как укладываться

1. Один Go Core и один Python runtime с ограниченной конкуренцией; latest-frame,
   bounded buffers, no media history, reuse allocations.
2. Компактные модели, минимально достаточный входной размер, контролируемые
   inference rates. Не загружать несколько больших frameworks без измерений.
3. Обучение в отдельных окнах/малых batch, без конкуренции с тяжёлыми CV задачами.
   Невозможность уложиться — blocker/оптимизация, не автоматическое удаление функции.
4. Hosted CI/development machine собирают images/frontend; Pi разворачивает готовое.
   Full Grafana/Prometheus optional/off; метрики Core сохраняются.
5. Argon2id baseline не ослабляется:19456KiB/t2/p1, один hashing job одновременно.
   Текущий x86 benchmark — только host baseline; Pi latency предстоит измерить.
6. Core dev container:128MiB hard RAM,128MiB memory+swap total,96MiB Go soft limit.
   Это не эмуляция целой Pi и не memory cap всех будущих сервисов.
7. Не менять swap/host services сейчас. Swap не заменяет RAM и не служит обоснованием
   выполнения realtime targets. OOM protection не выключать.

Docker semantics RAM/swap:[официальная документация](https://docs.docker.com/engine/containers/resource_constraints/).

## Физическая проверка

Зафиксировать model, MemTotal/MemAvailable, ОС/kernel/cgroup/swap baseline. Затем
проверить Core/auth flood, аудио, LED, CV/DSP и finally совместный режим; минимум30мин
steady-state плюс startup/restart/model load и отдельное окно обучения. Снять memory
peak, OOM events, swap activity, CPU/temp и целевые FPS/latency. Начальный duration —
проектный test fixture, не доказательство длительной устойчивости.

При pressure сначала сокращать optional workload; control plane/manual actions не
зависят от ML. Документировать реально включённые компоненты, чтобы отсутствие
тяжёлого модуля не выглядело как full-profile PASS. M0_NOT_READY остаётся до тестов.

## Приоритеты и размещение: уточнение26.10.2026

Pi — control plane/hardware hub. Laptop: screen capture, Ambilight analysis и Scene ML.
Pi: Core/SQLite, lighting/audio/sensors/MQTT/bot/static frontend и один лёгкий Python
Runtime. В нём lazy loading моделей, bounded workers/buffers; все модели постоянно
в памяти не держим. PostgreSQL/Kubernetes/тяжёлые Python workers исключены из scope.

Порядок сохранения функций:

1. Smart Home Core.
2. Lighting.
3. Audio.
4. Automation.
5. Ambilight.
6. Sensors / MQTT.
7. Presence.
8. Gesture ML.
9. Остальные ML-функции.

Деградация идёт9→1. Security/auth, emergency OFF, thermal и критические safety sensors
всегда относятся к Core, даже если их код находится в другом модуле. Disabled CV
означает UNKNOWN/stale presence, не AWAY. Scene labels не являются зависимостью
laptop color stream. Manual control не теряется из-за optional Runtime outage.

Normal target70–75% трактуется как `1-MemAvailable/MemTotal <=0.70..0.75`:
ориентир available headroom25–30% фактического MemTotal. Это оценка доступной памяти,
а не сумма RSS или raw used: [определение Linux MemAvailable](https://docs.kernel.org/filesystems/proc.html#meminfo).
Ранее указанные256MiB — отдельный предварительный floor для load/fault gate, не
нормальная цель и не разрешение держать систему постоянно у границы OOM.

Lazy unload проверяется по process/cgroup memory: Python/framework allocator может
сохранять память. При необходимости — bounded Runtime restart без Core restart.
Hysteresis и valid current permissions/leases обязательны при recovery; никаких
новых camera permissions. CPU/thermal/FPS проверяются вместе с RAM.

Host SSD swap optional, сейчас не настраивается; Core dev container остаётся no-swap.
Ни swap, ни отсутствие реально включённых ML consumers не доказывают full-profile
PASS. Автоматический memory-pressure controller **ещё не реализован**: закреплены
policy и будущие acceptance cases.
