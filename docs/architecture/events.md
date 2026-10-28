# Событийный контракт v0

Связи: [ADR-001](adr/001-module-boundaries.md), [команды](commands.md), [состояния](state-machines.md), [синтетические примеры](examples/events.json). Источник: §8/24.2/38–40 ТЗ.

## 1. Envelope

```json
{
  "event_id": "example-event-001",
  "schema_version": 1,
  "type": "LIGHT_MODE_CHANGED",
  "producer": "core.lighting",
  "occurred_at": "2026-10-25T09:00:00Z",
  "received_at": "2026-10-25T09:00:00Z",
  "boot_epoch": "example-boot-a",
  "sequence": 1,
  "correlation_id": "example-flow-001",
  "causation_id": "example-command-001",
  "resource": {"kind": "zone", "id": "monitor.backlight"},
  "revision": 1,
  "payload": {"old_mode": "STATIC", "new_mode": "OFF"}
}
```

Timestamp от внешнего producer не считается доверенным временем Core. Core назначает received_at/producer identity после service/device auth, проверяет type/schema/size/rate и timestamp plausibility. `sequence` монотонен для producer+boot_epoch, не глобален. Correlation/event ID не разрешает доступ к чужому resource.

`schema_version=1` — версия envelope. Неизвестный обязательный type/version отклоняется; additive optional fields допустимы при совместимом parser. Breaking change требует нового contract version. Topic/type string не используется как произвольный Go type или SQL selector.

## 2. Наблюдение, факт и команда

- `GESTURE_RECOGNIZED`, `SENSOR_UPDATED`, `SCENE_CHANGED` — наблюдения после validation, сами ничего не исполняют.
- `LIGHT_MODE_CHANGED`, `COMMAND_COMPLETED` — факты об изменённом control/applied state. Payload явно различает desired/applied; UI не делает вывод о физическом свете по одному accepted command.
- `USER_ACTION` — принятый ручной intent без secrets/track metadata; habits collector отдельно решает, какие поля законно сохранить.
- Нельзя создать manual override, отправив event с таким именем. Команда возникает только в trusted mapping/rule/gateway, с проверкой command layer.

## 3. Каталог событий

| Type | Producer → consumers | Минимальный payload / ограничения |
|---|---|---|
| BEAT_DETECTED | DSP → lighting/UI | source_epoch, beat strength, confidence; RAM-only |
| AUDIO_FEATURES_UPDATED | DSP → lighting/UI | overall/bass/mids/treble/BPM nullable, measured_at; RAM-only |
| AUDIO_STATE_CHANGED | Audio → UI/rules/habits | playing/volume/output/source kind; track metadata только current authorized UI, не habits/audit |
| GESTURE_RECOGNIZED | Runtime→Core mapper/UI | gesture/confidence, consumer epoch/sequence; frames отсутствуют |
| SCENE_CHANGED | Screen-client→Ambilight/UI | labels/confidence/inference timing, stream epoch; не записывать историю |
| TV_SIGNAL_LOST | Ambilight → rules/UI | client/session/zone, reason; имя из §8, источник здесь screen-client, не обещание отдельного TV tuner |
| SCREEN_ACTIVITY_CHANGED | Client → Core | monitor active/static/unavailable; никакого screenshot |
| LIGHT_MODE_CHANGED | Lighting → rules/UI | zone, old/new desired mode, revision, cause; effect state RAM |
| LIGHTING_STATE_APPLIED | Lighting → UI | zone/applied_revision, fallback/availability, technical frame counters |
| USER_ACTION | Dispatcher → local collector/audit | actor ref, verb, affected paths, normalized minimal state; no secret/track |
| DEVICE_CONNECTED / DEVICE_DISCONNECTED | Device Manager → UI/rules | device id, status, capabilities refs |
| SENSOR_UPDATED | Sensor Manager → presence/rules/UI | measurement kind/value/unit/quality/freshness |
| USER_PRESENT / USER_AWAY | Presence → rules/UI | state, reason codes, evidence freshness; only on corresponding confirmed transitions |
| PRESENCE_STATE_CHANGED | Presence → rules/UI/habits | все четыре состояния, включая UNKNOWN/PROBABLY_PRESENT |
| AUTOMATION_TRIGGERED | Automation → UI/audit | rule/trigger ID; отдельный execution result, не утверждение успеха |
| AUTOMATION_COMPLETED | Automation → UI/audit | APPLIED/PARTIAL/FAILED/SKIPPED, reason/action outcomes |
| COMMAND_COMPLETED | Dispatcher → operation subscribers | command ID/status/revision, safe error codes |
| CAMERA_STATE_CHANGED | Camera owner → privacy/UI/health | state/consumers/indicator/epoch; frames запрещены |
| SYSTEM_WARNING | System/Thermal/health → alerts/UI | code/severity/dependency/technical details с redaction |
| SYSTEM_STATE_CHANGED | System → UI | lifecycle/ready/degraded/reason codes |
| SECURITY_EVENT | Auth/integrations → restricted audit | action/result/principal ref; не passwords/tokens |

ТЗ приводит примеры событий, а не закрытый список. Дополнительные типы выше — проектные контракты v0. Envelope не должен превращаться в механизм хранения запрещённых payload.

## 4. Delivery и backpressure

Event Bus — **внутрипроцессный**, без Kafka/Redis и без event sourcing. Объём очередей ограничен конфигурацией; численные defaults подобрать benchmark на Д014, не оставлять unbounded.

| Класс | Семантика |
|---|---|
| Color/DSP/preview/metrics | Latest-value/coalescing на producer+resource; промежуточные значения могут пропускаться |
| Control state changes | Per-resource revision ordering; gap/overflow → invalidate subscription и snapshot resync, а не молчаливое состояние UI |
| Gesture/onset/automation triggers | Не coalesce как gauge; dedupe ID/epoch; при overflow явный error/drop counter, пропущенный gesture не воспроизводить позже |
| Security/automation audit | Владелец пишет обязательный audit напрямую через storage repository, а Event Bus уведомляет; loss уведомления не должен терять обязательную запись |

Producer не блокирует render loop на медленном UI. Контрольная команда не зависит от delivery в браузер: её operation outcome/query и persisted desired state — отдельный источник истины. Audit transaction commit происходит до сообщения об успешной durable записи.

После restart in-memory event stream не replay. Durable jobs/commands имеют отдельный lifecycle reconciliation; Event Bus не выдаёт exactly-once гарантий. External observation повторяется — consume по source epoch+sequence/id и ограниченному dedupe cache.

## 5. WebSocket snapshot и resync

`/ws/events` защищён session/token auth, resource/owner filtering и origin policy. Типы/частоты фильтруются сервером, не только frontend. Отзыв session/token закрывает активное соединение. Sensitive fields не публикуются в общий broadcast до фильтрации.

Race-free initial sync: сервер сначала регистрирует subscription с bounded buffer и subscription epoch, затем берёт согласованный Core state snapshot/revision watermark, отправляет `SNAPSHOT`, затем только изменения новее watermarks. Если buffer переполнен — `RESYNC_REQUIRED` и новый snapshot; old epoch discard. Snapshot/current state не содержит camera/screen frames. Reconnect не обещает replay по last event ID из прошлого boot.

Для browser latency учитывается публикация→получение, а не только socket write. WS throttling не ограничивает render FPS. Нельзя поставить spectrum limit равным rate limit управляющего HTTP API и считать target 30FPS выполненным.

## 6. Privacy и persistence

По умолчанию события — RAM. Camera frames не покидают camera pipeline, screen frames — laptop. PCM, DSP features history, Ambilight frames и scene labels history на диск не пишутся. Debug/exception/log middleware не должны сериализовать весь payload.

Persistence — только purpose-specific records из [data model](data-model-v0.md): minimum habit observations, automation/security audit с redaction, desired settings и operation metadata. Realtime data не попадает в generic event table/outbox. User/ServicePrincipal запрещено подменять actor_id в received observation.

System technical metadata для `/metrics` может включать inference latency, но не лица/кадры/полную scene history. Telegram receives только allowlisted status/commands/alerts без ML/CV/habit payload.

## Реализованный срез Д014–Д015

`internal/events` реализует envelope v1 с объектом resource `{kind,id}` и bounded
RAM-only delivery. На этом этапе каталог содержит SENSOR_UPDATED,
AUDIO_FEATURES_UPDATED, PRESENCE_STATE_CHANGED, DEVICE_CONNECTED,
SYSTEM_STATE_CHANGED; два дополнительных типа SCREEN_COLORS_OBSERVED и
MQTT_DEVICE_OBSERVED передают синтетические зональные цвета и discovery/state/disconnect
соответственно. Остальные типы таблицы подключаются вместе со своими модулями.
Это внутренний trusted Go API, не внешний ingress и не способ обойти будущую auth.

Exact subscription filter — `kind:id`; пустой filter означает все ресурсы.
Latest допускается только с одним exact resource и ёмкостью 1. Максимум subscriptions
ограничен config; identity producer в симуляции один, смена producer/epoch требует
новой subscription у будущего bridge. Control overflow закрывает канал/Done с
ErrOverflow: buffered events могут оставаться, consumer обязан проверить Err и
запросить snapshot будущего API. Trigger overflow увеличивает drop counter.
Один mutex задаёт порядок публикаций; последовательность/epoch/revision назначает
producer, bus не выполняет dedupe, durable запись или внешнюю аутентификацию.
Обработка дубликатов внешних наблюдений будет в gateway, а WS resync — на Д023.
Подробности и измерения: [baseline](../research/event-bus-baseline.md).
