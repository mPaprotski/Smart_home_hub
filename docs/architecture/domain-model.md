# Доменные сущности и инварианты v0

Связь: [архитектурный индекс](README.md), [данные](data-model-v0.md), [команды](commands.md), [события](events.md), [состояния](state-machines.md). Источник: §4/9/11–28/31/36/49 ТЗ.

## 1. Общие правила

- Entity IDs — непрозрачные стабильные строки; конкретный генератор выбирается при реализации. Zone ID может быть читаемым `monitor.backlight`; его переименование — явная операция, не неявный побочный эффект.
- UTC timestamps в API — RFC3339 с timezone; business schedules используют отдельную IANA timezone, первоначально `Europe/Minsk`. Время и день habits вычисляются относительно настройки пользователя, а не timezone host.
- Для duration/confirmation/рендера внутри процесса использовать монотонное время. Между перезапусками хранить UTC deadlines и отдельно обрабатывать недостоверные часы.
- Числа конечны: NaN/Inf запрещены. Brightness/volume/modifiers в internal model — 0…1; UI может отображать проценты. `ml_influence` API/UI — 0…100, к normalized weight преобразуется явно.
- У mutable control state есть `revision`; optimistic concurrency не заменяет общий арбитр. Runtime telemetry имеет freshness/sequence, но не обязана создавать запись SQLite на каждый sample.
- `null` означает неизвестно/неприменимо, а не ноль/false. Нулевое измерение допустимо, если его quality — VALID.
- В MVP один admin, но ownership уже выражено явно. Shared room/device помечаются системными; multi-user UX не добавляется автоматически.

## 2. Пользователи и доступ

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| User | id, username, password_hash, enabled, timezone, created_at | Auth; password только Argon2id hash; первый user admin |
| Role | id, name, permissions | Auth; role admin не равна произвольному machine token |
| UserRole | user_id, role_id | Auth; привязка M:N без обязательного multi-user UI |
| Session | id_hash, user_id, expires_at, revoked_at, created_at | Auth; cookie secret не хранится открытым текстом; logout/revoke проверяются до commands/WS |
| ApiToken | id, secret_hash, principal_id/user_id, scopes, resource_bindings, revoked_at, created_at | Auth; показать secret один раз; scope не отменяет привязку к конкретной зоне/device |
| ServicePrincipal | id, kind, owner_user_id, allowed_operations | Core identity; внутренние adapters/Runtime не получают admin по факту подключения |
| ClientPairing | id, code_hash, expires_at, attempts, consumed_at, requested_binding | Auth/device pairing; code одноразовый, не session/token |
| TelegramLink | telegram_user_id, user_id, enabled, linked_at | Auth/integration; explicit local linking, unknown Telegram user не управляет Hub |

`source` и principal internal command формируются доверенным gateway. Поле `source=button` в пользовательском JSON не даёт аппаратных прав. Gesture runtime может сообщить gesture/confidence, но не выбрать произвольного actor или action вне сохранённой mapping.

## 3. Устройства и логические зоны

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| Device | id, type, adapter_id, status, capabilities, metadata, last_seen_at | Device Manager; ONLINE/OFFLINE/ERROR/DISABLED; metadata валидируется/redacts |
| DeviceCapability | device_id, name, properties | Device Manager; `rgb`, `per_pixel`, `humidity`, `motion`, `audio_output` и т.п.; unsupported command отклоняется |
| PhysicalTransport | id, adapter, physical_device, channel, frame_capacity, availability | Adapter/config layer; один owner hardware handle; transport не является zone |
| ZoneBinding | zone_id, transport_id, offset, pixel_count, mapping | Lighting/config; непересекающиеся slices, общая длина в пределах transport capacity |
| LightingZone | id, owner_user_id nullable, device_id, layout, desired_state, runtime_state, revision | Lighting; произвольное число зон, никакого hardcoded массива из двух элементов |
| ZoneDesiredState | mode, brightness, static_color, effect_params, selected_preset_id, transition_ms | Lighting; последняя принятая пользовательская/разрешённая automation цель, persistence без frame history |
| ZoneRuntimeState | effective_mode, applied_revision, availability, fallback_reason, frame_metrics, modifiers | Lighting RAM; желаемая цель может отличаться от временного fallback/неисправного hardware |
| Preset | id, owner_user_id, name, zone_entries, revision | Lighting/settings; несколько зон; все entries validate до выполнения |

STRIP: `pixel_count`. MATRIX: `width`, `height`, `orientation`, `serpentine_layout`, согласованный `width×height=pixel_count`. Monitor mapping дополнительно хранит counts четырёх сторон, start_corner/direction; их сумма равна длине зоны. Bounds и пересечения проверяются до любого Show().

Обязательные modes: OFF, STATIC, BREATHING, RAINBOW, MUSIC_PULSE, MUSIC_BEAT, MUSIC_SPECTRUM, MUSIC_WAVE, MUSIC_CENTER, AMBILIGHT, NIGHT_SHIFT. Неподдерживаемый mode не «исправляется» молча в STATIC.

Brightness pipeline: desired brightness → разрешённые master/ambient/scene modifiers → effect/color/fade → **финальный zone/global current limiter** → transport frame. Точный gamma/preprocessing порядок и coefficients остаются Q-005 до golden/color spike. Limiter никогда не отменяется ни manual override, ни priority. Frame и визуальная telemetry — RAM-only.

## 4. Аудио

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| AudioSource | id, kind, available, capabilities, connection_state | Audio; BLUETOOTH_A2DP/SPOTIFY_CONNECT/LOCAL_FILES |
| AudioOutput | id, kind, available, capabilities | Audio; ANALOG_3_5/USB/BLUETOOTH; предпочтение сохраняется, live handle нет |
| AudioState | source_id, output_id, playback_state, volume, mute, metadata, revision | Audio RAM; STOPPED/PLAYING/PAUSED/BUFFERING/ERROR; availability отдельно |
| AudioPreferences | preferred_output_id, source_settings, startup_volume_limit | User settings; конкретный startup volume после audio spike, playback всегда STOPPED при restart |
| AudioFeatures | measured_at, source_epoch, overall, bass, mids, treble, beat, bpm nullable, confidence | DSP RAM; не listening history, не SQLite table |

Metadata track/artist/album/cover/progress существуют только пока нужны текущему UI; track metadata, playlist progression и BPM не попадают в command journal/habits/logs как история прослушивания. Локальная библиотека файлов может сохраняться, факт воспроизведения каждого файла — нет.

## 5. Измерения, presence и камера

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| SensorDefinition | device_id, measurement_kind, unit, poll_interval, config | Sensor Manager; exact driver/units из BOM, не универсальный `air_quality_ppm` |
| SensorMeasurement | device_id, kind, value nullable, unit, observed_at, received_at, quality, sequence | Sensor Manager RAM; VALID/STALE/UNAVAILABLE/ERROR; stale не удовлетворяет числовому condition |
| PresenceEvidence | source_id, kind, detected nullable, confidence nullable, observed_at, expires_at, quality | Presence; camera/PIR/Bluetooth/network, без face identity |
| PresenceState | state, evaluated_at, sources, reason, revision | Presence RAM; UNKNOWN отличается от AWAY; fusion weights/TTL Q-019 |
| CameraPermission | gesture_allowed, cv_presence_allowed, revision | Core privacy settings; независимо от факта наличия камеры |
| CameraLease | consumer, process_epoch, expires_at, requested_profile | Capture owner RAM; gesture/presence lease, не API token |
| CameraRuntime | capture_state, consumers, indicator_state, fault_reason, boot_epoch | Camera Manager; после boot leases отсутствуют, capture OFF |
| GestureObservation | gesture, confidence, observed_at, sequence, runtime_epoch | Runtime RAM; только наблюдение, не разрешение на command |
| GestureMapping | owner_user_id, gesture, action_template, confirmation_ms, cooldown_ms | Core settings; activation gate и allowed operations проверяет Core |

PIR — источник motion, а не доказательство identity. Sensor raw stream не сохраняется как time series в MVP; habit collector берёт минимальный требуемый контекст. CV frames остаются внутри локального Runtime memory, не в event bus/WS.

## 6. Правила и защищённое ручное состояние

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| AutomationRule | id, owner_user_id, enabled, priority, trigger, conditions, actions, allow_music_scene_change, revision | Automation; versioned typed DSL; unknown reference — validation error |
| ConditionNode | operator, children либо typed comparison | Automation; AND/OR nested, type/depth limits при implementation |
| ActionTemplate | command_type, target, payload, delay_ms optional | Automation; не произвольный shell/SQL/Python |
| DelayedJob | id, rule_id/revision, trigger_id, not_before, actions, guard, status, attempt | Automation/storage; rule disabled/changed, guard false или override блокируют stale action |
| ManualOverride | id, actor, source, resource_paths, wildcard_scope optional, created_at, expires_at, command_id | Dispatcher; default TTL 900 секунд; persisted при доступной БД |
| AutomationAudit | rule_id, trigger_id, time, actions summary, result, error, correlation_id | Automation/storage; 30 дней, без raw ML/frame/track payload |

Field paths: `zone:<id>:mode`, `zone:<id>:brightness`, `zone:<id>:static_color`, `audio:volume`, `audio:playback`, `audio:output`. Это логические keys, не адреса JSON произвольной структуры. `lighting:*` Master OFF перекрывает любые lighting control paths, включая добавленную в интервале зону.

## 7. Привычки и рекомендации

| Сущность | Основные поля | Владелец и инвариант |
|---|---|---|
| HabitLearningState | user_id, started_at, privacy_epoch, data_quality, model_version nullable | Habits; eligibility не раньше 30 реальных дней, при плохих данных no-action |
| HabitObservation | user_id, timestamp, timezone, day_of_week, zone_id, mode, brightness, preset_id, music_active, tv_active, ambilight_active, gesture_used, manual_override, state_duration_s | Collector/storage; 90-day rolling retention |
| Дополнительный контекст наблюдения | volume nullable, presence_state, ambient_lux nullable, monitor_active nullable, origin, quality flags | Проектное уточнение §17.1/17.2/17.5; без этих полей targets/context неполны; не записывать треки/кадры |
| HabitModel | user_id, model_version, privacy_epoch, feature_schema_version, local_path, trained_at, quality_summary | Local Runtime; artifacts только локально, атомарная замена с проверкой epoch |
| Recommendation | id, user_id, action_template, confidence, explanation_codes, model_version, expires_at, status | Predictions; PENDING/ACCEPTED/REJECTED/EXPIRED/EXECUTED/FAILED |
| PredictionFeedback | recommendation_id, user_id, decision, created_at, privacy_epoch | Habits; ACCEPT/REJECT; reject negative, manual undo не feedback |

Recommendation не bypass command dispatcher. AUTO допускается только при явном включении, достаточной истории/quality, confidence и разрешённом контексте. Конкретные thresholds и калибровка остаются Q-022; 60/85% — предложенный пример, не доказанная accuracy.

## 8. Жизненный цикл и служебные сущности

`CommandRecord` хранит минимальную idempotency/outcome metadata, не raw secrets/frames. `SystemConfig` имеет profile/revision. `DependencyHealth` разделяет configured/enabled/live/ready/degraded и причины. `SecurityEvent` redacted; срок хранения определяется Q-030. `BackupManifest` содержит schema/software versions, checksums, privacy epoch и перечень компонентов, но не публикуется в GitHub artifacts.

Значения status enum описываются однозначно в [state machines](state-machines.md). Результат command и состояние hardware не смешиваются: ACCEPTED не равно APPLIED, APPLIED не гарантирует физический свет без telemetry/осмотра, если adapter не умеет это подтвердить.
