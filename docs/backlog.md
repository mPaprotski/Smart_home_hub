# День 1. Backlog разработки

**Hardware baseline от26.10.2026: только Raspberry Pi 4B с2GB RAM.** [Ресурсный профиль](development/pi4b-2gb.md). Pi совместные memory/performance gates ещё не выполнены.

Основание: [ТЗ](../Smart_Home_Hub_TZ.md), [матрица требований](requirements.md), [ежедневный план](DEVELOPMENT_PLAN_DAILY.md), [открытые вопросы](open-questions.md).

Дата: 25 октября 2026 года. **BL-01 в работе: реализован программный Core/simulation срез Д011–Д015; SQLite/auth и production готовность ещё впереди.** Завершены документы Дня1, предварительный BOM Дня2, архитектура/7 ADR Дня3, [локальная часть Д004–Д010](DAYS_004_010_RESULT.md) и [проверки Core/native/Compose Д011–Д015](DAYS_011_015_RESULT.md). Аппаратные испытания не выполнены; M0 остаётся NOT_READY. Нумерация BL соответствует 13 этапам §51 ТЗ, а не количеству недель.

Владелец исполнения по умолчанию — разработчик проекта. M0–M11 и номера дней относятся к ежедневному плану: 160 основных рабочих дней плюс банк 20 дней резерва. Не нужно складывать окна BL как независимые оценки: UI/CI/collector/observability развиваются поперёк этапов и частично входят в одни и те же дни.

## Порядок и зависимости

```text
Уточнение требований и hardware spikes (Д001–Д010)
  → BL-01 Core/auth/storage/simulation
  → BL-02 Lighting → BL-03 Audio → BL-04 DSP
  → BL-05 Ambilight → BL-06 Scene ML
  → BL-07 Gestures → BL-08 Sensors/Presence
  → BL-09 Automation → BL-10 Habit Prediction
  → BL-11 MQTT/Telegram → BL-12 Reliability/Observability
  → завершение BL-13 CI/CD → formal acceptance → v1.0.0
```

Это порядок основных работ одного разработчика. BL-13 начинается CI-заготовкой на Д015. Manual override начинается до полного Automation Engine. Habit collector начинается на Д035, чтобы календарное обучение шло параллельно. MQTT condition/Telegram manual source проверяются сначала fake fixtures, затем реальными integration на BL-11. Камера для presence зависит от Camera Manager BL-07, Scene ML от этого Camera Manager не зависит.

Требования-владение обозначают ответственный пакет, а не обещают закрытие всех его сквозных строк в первое рабочее окно. Например, каркас UI принадлежит BL-01, функции отдельных страниц завершаются с профильными пакетами; MUSIC/AMBILIGHT mode contracts Lighting закрываются с BL-04/BL-05. В формальной приёмке проверяется весь комплект зависимостей.

## Обязательные пакеты

### BL-01. Infrastructure Core

**Статус:** в работе, M0_NOT_READY. **Окна:** Д001–Д025; аппаратная подготовка Д002–Д010. **Контроль:** M1 / Д025.

**Зависимости:** Исходное ТЗ; доступность Pi и клиента LAN.

- [x] Разложить ТЗ и сохранить матрицу/backlog/questions (Д001).
- [x] Подготовить предварительный BOM/power budget/wiring/pin map с явными неизвестными и вариантами (Д002); [отчёт](DAY_002_RESULT.md).
- [x] Описать domain/commands/events/states, logical data model, 7 ADR и 20 конфликтных сценариев (Д003); [отчёт](DAY_003_RESULT.md).
- [x] Выполнить host/synthetic zone/PCM/PipeWire/screen/inference/thermal/HTTPS probes, два replay и20 code tests (локальная часть Д004–Д010).
- [x] Подготовить Pi install/preflight/HIL/HTTPS guides и честную M0 evidence matrix.
- [ ] Выполнить actual Pi installation/cooling/GPIO/I²C/SPI/LED-audio/LAN-mobile tests; M0 physical gates открыты.
- [ ] Уточнить аппаратную конфигурацию после замеров и подтвердить LED/audio/capture решения на стенде (Д003–Д010).
- [x] Создать monorepo, config, lifecycle, HTTP/OpenAPI, Event Bus, минимальные fakes и CI skeleton (Д011–Д015); [host/native/Compose evidence](DAYS_011_015_RESULT.md). Remote CI отдельно ещё не запускался.
- [x] Реализовать SQLite/migrations, first persistent command/permissions и admin bootstrap/Argon2id credentials (Д016–Д018); [отчёт и границы](DAYS_016_018_RESULT.md).
- [x] Добавить sessions/CSRF, scoped tokens/security audit, timezone gateway и auth-only WS, domain schema/retention (Д019–Д021); [локальный результат и границы](DAYS_019_021_RESULT.md).
- [ ] Поднять responsive UI, auth, WS/Dashboard, HTTPS/mDNS и все fake adapters (Д022–Д025).

**Готово, когда:** Чистый checkout работает без Pi в simulation; admin login/защита API/WS и persistence проверены. M0 отдельно требует Pi spike, M1 не заменяет аппаратные испытания.

**Требования-владение:** ARCH-001, ARCH-002, ARCH-003, ARCH-004, ARCH-005, ARCH-006, ARCH-007, ARCH-008, ARCH-009, ARCH-010, ARCH-011, ARCH-012, HW-001, HW-002, HW-003, HW-005, HW-006, HW-007, HW-008, HW-010, HW-011, HW-012, WEB-001, WEB-002, WEB-003, WEB-004, WEB-005, WEB-006, WEB-007, WEB-008, WEB-009, WEB-010, WEB-011, WEB-012, WEB-013, WEB-014, WEB-015, WEB-016, WEB-017, WEB-018, WEB-019, AUTH-001, AUTH-002, AUTH-003, AUTH-004, AUTH-005, AUTH-006, AUTH-007, AUTH-008, AUTH-009, AUTH-010, AUTH-011, API-001, API-002, API-003, API-004, API-005, API-006, API-007, STORE-001, STORE-002, STORE-003, STORE-004, STORE-005, STORE-006, STORE-007, STORE-008, SIM-001, SIM-002, SIM-003, SIM-004, SIM-005, SIM-006, SIM-007.

**Открытые вопросы:** Q-001, Q-002, Q-003, Q-004, Q-005, Q-006, Q-010, Q-021, Q-023, Q-024, Q-025, Q-026, Q-028, Q-029, Q-030.

### BL-02. Lighting Engine

**Статус:** запланирован. **Окна:** Д026–Д035. **Контроль:** M2 / Д035.

**Зависимости:** BL-01; подтверждённые LED channels и PSU/BOM.

- [ ] Реализовать arbitrary zones, STRIP/MATRIX и mapping.
- [ ] Подключить реальные WS2812B adapters и единый color/render pipeline.
- [ ] Реализовать zone/global current limits, fade, базовые modes и configurable Night Shift.
- [ ] Сделать multi-zone presets, Lighting UI/visualizer и manual override/Master OFF.
- [ ] Подключить кнопку/remapping; смонтировать и откалибровать две зоны; включить ранний collector BL-10 на Д035.

**Готово, когда:** Две реальные зоны независимы; limiter действует на любой frame; кнопка и presets работают; ручной OFF защищён от автоматического включения. MUSIC/AMBILIGHT modes принимаются с BL-04/BL-05.

**Требования-владение:** LIGHT-001, LIGHT-002, LIGHT-003, LIGHT-004, LIGHT-005, LIGHT-006, LIGHT-007, LIGHT-008, LIGHT-009, LIGHT-010, LIGHT-011, LIGHT-012, LIGHT-013, LIGHT-014, LIGHT-015, LIGHT-016, LIGHT-017, LIGHT-018, LIGHT-019, LIGHT-020, LIGHT-021, POWER-001, POWER-002, POWER-003, POWER-004, POWER-005, BTN-001, BTN-002, BTN-003, BTN-004, BTN-005, DEV-001, DEV-002, DEV-004, DEV-005, DEV-006, DEV-007.

**Открытые вопросы:** Q-002, Q-005, Q-009, Q-023, Q-030.

### BL-03. Audio

**Статус:** запланирован. **Окна:** Д036–Д045. **Контроль:** Audio acceptance / Д045.

**Зависимости:** BL-01/BL-02; audio spike Д004/Д008.

- [ ] Оформить host audio graph и ограниченный control API контейнера.
- [ ] Получить общий PCM tap для DSP всех источников без записи звука.
- [ ] Подключить Bluetooth input/PAIRING_MODE/trusted devices и source capabilities/metadata.
- [ ] Добавить local files и Spotify Connect, все три output types и их выбор UI.
- [ ] Пройти source×output×control matrix, disconnect/reconnect, mute и restart STOPPED.

**Готово, когда:** Все обязательные источники/выходы проверены на Pi; metadata условны доступностью, controls и ограничения явно записаны; local/Bluetooth не зависят от WAN.

**Требования-владение:** AUDIO-001, AUDIO-002, AUDIO-003, AUDIO-004, AUDIO-005, AUDIO-006, AUDIO-007, AUDIO-008, AUDIO-009, AUDIO-010, AUDIO-011, AUDIO-012, AUDIO-013, AUDIO-014, AUDIO-015, AUDIO-016, AUDIO-017.

**Открытые вопросы:** Q-007, Q-008, Q-009, Q-021.

### BL-04. Music Reactive Lighting

**Статус:** запланирован. **Окна:** Д046–Д055. **Контроль:** M3 / Д055.

**Зависимости:** BL-02/BL-03; PCM contract.

- [ ] Поднять Python Runtime lifecycle с независимостью от playback/core.
- [ ] Реализовать FFT/bands/normalization/smoothing/beat/BPM и user settings.
- [ ] Добавить MUSIC_PULSE/MUSIC_BEAT/MUSIC_SPECTRUM/MUSIC_WAVE/MUSIC_CENTER.
- [ ] Вывести spectrum/BPM/beat через bounded WS в UI.
- [ ] Проверить golden audio/effects, отказ DSP и joint load на Pi.

**Готово, когда:** Пять эффектов и realtime UI работают от реального audio; limiter соблюдён; нет DSP history; playback не прекращается при Runtime failure.

**Требования-владение:** DSP-001, DSP-002, DSP-003, DSP-004, DSP-005, DSP-006, DSP-007, DSP-008, DSP-009, DSP-010, DSP-011.

**Открытые вопросы:** Q-021, Q-033.

### BL-05. Screen Client / Ambilight

**Статус:** запланирован. **Окна:** Д056–Д070. **Контроль:** M4 / Д070.

**Зависимости:** BL-01/BL-02; реальные четыре OS targets; capture spike.

- [ ] Реализовать one-time pairing/scoped token и versioned colors protocol.
- [ ] Подключить реальный capture Linux/Windows/macOS arm64/macOS x86_64.
- [ ] Реализовать weighted edges/mapping/black bars/color correction/smoothing.
- [ ] Добавить Cinema, inactivity/disconnect fallback и color-only preview/configuration UI.
- [ ] Упаковать четыре клиента и измерить реальный screen→LED FPS/latency; проверить invalid/slow/disconnected client.

**Готово, когда:** Четыре платформы реально передают корректные края монитора; 30 FPS/<100ms измерены; screenshots не сохраняются/не передаются Hub. Отсутствующий test target остаётся blocker.

**Требования-владение:** AMBI-001, AMBI-002, AMBI-003, AMBI-004, AMBI-005, AMBI-006, AMBI-007, AMBI-008, AMBI-009, AMBI-010, AMBI-011, AMBI-012, AMBI-013, AMBI-014, AMBI-015, AMBI-016, AMBI-017, AMBI-019, AMBI-020, AMBI-021.

**Открытые вопросы:** Q-002, Q-005, Q-006, Q-011, Q-012, Q-013, Q-014, Q-021.

### BL-06. Scene ML

**Статус:** запланирован. **Окна:** Д071–Д080. **Контроль:** M5 / Д080.

**Зависимости:** BL-05; Q-006/Q-015; licensed/synthetic dataset.

- [ ] Выбрать локальный inference runtime, проверить четыре targets и delivery модели.
- [ ] Подготовить temporal dataset/rubric/split/quality gates до обучения.
- [ ] Построить features/baseline, обучить и независимо оценить multi-label classifier.
- [ ] Добавить influence 0–100%, label conflict policy и изменения только разрешённых параметров.
- [ ] Проверить 1 inference/sec без блокирования 30FPS, disable/failure/0%, privacy и dependencies всех пакетов.

**Готово, когда:** Model quality gate пройден; ordinary Ambilight работает без ML; все девять меток проверены, mode не переключается, scene history отсутствует.

**Требования-владение:** SCENE-001, SCENE-002, SCENE-003, SCENE-004, SCENE-005, SCENE-006, SCENE-007, SCENE-008.

**Открытые вопросы:** Q-006, Q-013, Q-015, Q-021.

### BL-07. Gesture Control

**Статус:** запланирован. **Окна:** Д081–Д090. **Контроль:** Gesture acceptance / Д090.

**Зависимости:** BL-01/BL-04 Runtime; оба типа камер; Q-017/Q-018.

- [ ] Реализовать camera owner, privacy toggles и реальный CV indicator.
- [ ] Проверить USB webcam и Pi Camera на целевой Ubuntu.
- [ ] Добавить hand detection, пять статических и пять динамических gestures.
- [ ] Добавить activation window/confirmation/cooldown и editable mappings UI.
- [ ] Измерить false positives/quality/reaction <500ms с confirmation, проверить shared load/disconnect/restart/privacy.

**Готово, когда:** Все десять mappings проверены; камера после restart выключена; RAM-only и indicator соблюдены; quality/timing gates не заменены оптимистичным UI событием.

**Требования-владение:** GEST-001, GEST-002, GEST-003, GEST-004, GEST-005, GEST-006, GEST-007, GEST-008, GEST-009, GEST-010, GEST-011, GEST-012, GEST-013, GEST-014, GEST-015, GEST-016, GEST-017, GEST-018, GEST-019.

**Открытые вопросы:** Q-016, Q-017, Q-018, Q-021.

### BL-08. Sensors / Presence

**Статус:** запланирован. **Окна:** Д091–Д100. **Контроль:** M6 / Д100.

**Зависимости:** BL-01/BL-07 shared camera; датчики; Q-019/Q-020.

- [ ] Реализовать Sensor Manager/fake drivers/quality/stale и реальные пять типов измерений.
- [ ] Подключить Devices UI, inventory/status/capabilities/manual addition и sensor events.
- [ ] Добавить ambient modifier с направлением из ТЗ и защитой music/override.
- [ ] Добавить camera person/Bluetooth/network/PIR и fusion четырёх состояний.
- [ ] Проверить CV scheduling, indicator, source conflicts/offline и совокупную нагрузку на Pi.

**Готово, когда:** Пять датчиков доступны UI/rules; четыре presence source проверены; UNKNOWN отличён от AWAY, камера общая и privacy сохранена.

**Требования-владение:** PRES-001, PRES-002, PRES-003, PRES-004, PRES-005, PRES-006, SENS-001, SENS-002, SENS-003, SENS-004, SENS-005, SENS-006, SENS-007, SENS-008, SENS-009.

**Открытые вопросы:** Q-013, Q-017, Q-019, Q-020, Q-023.

### BL-09. Automation Engine

**Статус:** запланирован. **Окна:** Ранний override Д032; Д101–Д110. **Контроль:** M7 / Д110.

**Зависимости:** BL-01 command/events; BL-02/03/05/07/08 data; MQTT condition smoke после BL-11.

- [ ] Зафиксировать rule DSL, nested AND/OR, typed validation и evaluator всех conditions.
- [ ] Реализовать regular schedules/timezone, multi-actions, priority/tie-break и loop/idempotency policy.
- [ ] Добавить persisted/cancellable delayed jobs с revalidation/restart semantics.
- [ ] Сделать visual Rule Builder, audit с 30-day retention и preview.
- [ ] Проверить manual override всех источников, Master OFF, music protection и event storm.

**Готово, когда:** Rules/delays/priority/override/audit доказаны simulation и HIL; реальные MQTT и Telegram cross-check завершены с BL-11. M7 не закрывает будущие интеграционные проверки автоматически.

**Требования-владение:** AUTO-001, AUTO-002, AUTO-003, AUTO-004, AUTO-005, AUTO-006, AUTO-007, AUTO-008, AUTO-009, AUTO-010, AUTO-011, AUTO-012, AUTO-013, AUTO-014, AUTO-015, AUTO-016, AUTO-017, AUTO-018, AUTO-019, AUTO-020, AUTO-021, AUTO-022, AUTO-023, AUTO-024, AUTO-025.

**Открытые вопросы:** Q-023.

### BL-10. Habit Prediction

**Статус:** запланирован. **Окна:** Schema Д021; collector Д035; Д111–Д120. **Контроль:** M8 / Д120 при выполненном календарном gate.

**Зависимости:** BL-01 storage; BL-02 events; BL-09 execution policy; >=30 реальных дней и достаточно данных.

- [ ] Включить минимальный collector на Д035 и сохранять фактический learning start; расширять по мере готовности источников.
- [ ] Проверить schema/context/data quality, duration, 90-day retention и local training snapshots.
- [ ] Построить baseline/temporal split/model для всех семи targets без sequence modelling.
- [ ] Реализовать eligibility/confidence/SUGGEST/AUTO, Predictions UI и accept/reject feedback.
- [ ] Удалять observations/feedback/model/caches и перезапускать learning period; проводить AUTO через override/capabilities.

**Готово, когда:** 30 дней подтверждены реальными timestamps, а не fixtures; confidence откалиброван; delete и negative feedback проверены. Недостаточная история означает no-action и перенос gate, не фиктивную рекомендацию.

**Требования-владение:** HAB-001, HAB-002, HAB-003, HAB-004, HAB-005, HAB-006, HAB-007, HAB-008, HAB-009, HAB-010, HAB-011, HAB-012, HAB-013, HAB-014.

**Открытые вопросы:** Q-021, Q-022, Q-023.

### BL-11. MQTT / Telegram / HA readiness

**Статус:** запланирован. **Окна:** Д121–Д125; проверки в M9. **Контроль:** Integration acceptance / Д125.

**Зависимости:** BL-01 auth/scopes/API; BL-08 registry; BL-09 events.

- [ ] Развернуть Mosquitto с auth/ACL/LAN boundaries.
- [ ] Определить discovery/topics/LWT/reconnect и добавить idempotent Device registry.
- [ ] Подключить MQTT events к rules/fake simulation и проверить независимый REST/MQTT client.
- [ ] Сделать Go bot только через Core API, явную Telegram ID linking и все семь commands/inline buttons.
- [ ] Проверить scoped permissions, forbidden data, bot/WAN outage и HA independence.

**Готово, когда:** MQTT autodiscovery работает с тестовым устройством; все bot commands доступны linked user; Hub работает без bot/HA; реальный HA Discovery не входит в обязательный объём.

**Требования-владение:** MQTT-001, MQTT-002, MQTT-003, MQTT-004, MQTT-005, MQTT-006, MQTT-007, MQTT-008, TG-001, TG-002, TG-003, TG-004, TG-005, TG-006, TG-007, TG-008, TG-009, TG-010, TG-011, TG-012, TG-013.

**Открытые вопросы:** Q-001, Q-021, Q-025, Q-027.

### BL-12. Reliability / Observability

**Статус:** запланирован. **Окна:** Метрики/logs с Д012; Д126–Д135. **Контроль:** M9 / Д130; performance gate / Д135.

**Зависимости:** Все функциональные BL-01…BL-11; согласованные health/privacy/backup contracts.

- [ ] Добавить watchdog/dependency recovery, graceful shutdown и restart policies состояний.
- [ ] Реализовать защищённый SQLite Safe Mode без persistent writes/automation.
- [ ] Добавить daily backup/7 copies/restore/factory reset с privacy для backups.
- [ ] Завершить metrics/alerts/System UI/Thermal Manager и ограничения logs/disk/RAM.
- [ ] Реализовать memory-pressure controller для Pi4B/2GB: documented9→1 degradation, lazy unload/RSS verification, hysteresis/recovery, UNKNOWN при недоступном CV и fault tests; [policy](development/pi4b-2gb.md).
- [ ] Проверить ML/MQTT/SQLite failures и full joint performance profile; optional dashboards отдельно.

**Готово, когда:** Fault matrix соответствует ТЗ; restore/reset реально проверены; все metrics/alerts доступны; §46 выполнен на зафиксированном стенде.

**Требования-владение:** THERM-001, THERM-002, THERM-003, REC-001, REC-002, REC-003, REC-004, REC-005, REL-001, REL-002, REL-003, REL-004, REL-005, REL-006, REL-007, REL-008, REL-009, REL-010, REL-011, REL-012, REL-013, REL-014, OBS-001, OBS-003, OBS-004, OBS-005, OBS-007, OBS-008, OBS-009, OBS-010, OBS-011, OBS-012, OBS-013, OBS-014, OBS-015, OBS-016, OBS-017, OBS-018, OBS-019, OBS-020, OBS-021, OBS-022, OBS-023, OBS-024, OBS-025, OBS-026, OBS-027, OBS-028, OBS-029, OBS-030, OBS-031, OBS-032, OBS-033, OBS-034, OBS-035, OBS-036, OBS-037, OBS-038, OBS-039, OBS-040, OBS-041, PRIV-001, PRIV-002, PRIV-003, PERF-001, PERF-002, PERF-003, PERF-004, PERF-005, PERF-006, PERF-007, PERF-008.

**Открытые вопросы:** Q-004, Q-005, Q-006, Q-013, Q-017, Q-018, Q-021, Q-023, Q-027, Q-028, Q-030, Q-031, Q-032, Q-034.

### BL-13. CI/CD, стабилизация и выпуск

**Статус:** запланирован. **Окна:** CI v0 Д015; Д136–Д160. **Контроль:** M10 / Д145; M11 / Д160.

**Зависимости:** Проверки каждого BL в течение разработки; production release после BL-01…BL-12.

- [ ] Завершить non-root/multi-stage/pinned/multi-arch images и полный Go/TS/Python security/test pipeline.
- [ ] Проверить systemd/reboot/SSD migration/clean installation и client packages.
- [ ] Изолировать production runner от PR/forks; реализовать tag-only deploy/readiness/schema-safe rollback.
- [ ] Закрыть coverage/golden/benchmarks/E2E/privacy, провести >=48h soak и исправить blockers.
- [ ] Пройти все 62 пункта плюс обязательные требования вне §50; закончить user/ops/architecture docs и выпустить проверенный v1.0.0.

**Готово, когда:** Все acceptance строки имеют evidence final candidate; coverage >=80% core/>=70% overall, CI gates зелёные; tag deploy и rollback испытаны на Pi. Публикация не выполняется в рамках Дня 1.

**Требования-владение:** CI-001, CI-002, CI-003, CI-004, CI-005, CI-006, CI-007, CI-008, CI-009, CI-010, CI-011, DEPLOY-001, DEPLOY-002, DEPLOY-003, DEPLOY-004, DEPLOY-005, DEPLOY-006, TEST-001, TEST-002, TEST-003, TEST-004, TEST-005, TEST-006, TEST-007, TEST-008, TEST-009.

**Открытые вопросы:** Q-002, Q-013, Q-021, Q-026, Q-028, Q-031, Q-035.

## Опциональные работы и рекомендации

Не являются обязательным объёмом первой версии и выполняются только после release blockers/достаточности резерва. Каждую из строк ниже проверять по каталогу; рекомендации не превращать автоматически в обязательные закупки.

| ID требования | Работа | Место |
|---|---|---|
| HW-004 | Вентилятор управляется GPIO через транзистор/MOSFET | Рекомендация ТЗ; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| HW-009 | Для level shifting выбран 74AHCT125 либо 74HCT245 | Рекомендация ТЗ; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| AMBI-018 | Gaming Mode снижает smoothing/latency и ускоряет реакцию | Опционально; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| DEV-003 | Реальная LED-панель добавляется после логической поддержки | Опционально; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| MQTT-009 | Home Assistant MQTT Discovery добавляется в будущем | Опционально; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| API-008 | Remote access при необходимости выполняется через Tailscale/WireGuard | Рекомендация ТЗ; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| OBS-002 | Logs используют correlation/event IDs где возможно | Рекомендация ТЗ; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |
| OBS-006 | Grafana/Prometheus разворачиваются optional Compose profile | Опционально; резерв Д176–Д180 для optional функций либо решение в BOM/architecture для рекомендаций |

## Явно вне обязательного backlog

Исключения §52 и потенциальное развитие §53 перечислены в [границах матрицы](requirements.md). Не создавать текущие задачи на AirPlay, face identity, Kubernetes/k3s, Zigbee/Thread/Matter, sockets/relays/climate, полноценный multi-user UX, mobile app и несколько комнат без отдельного изменения scope.

Облачный ML для основной функциональности, запись runtime camera/screen frames и listening history — запрещённые текущими требованиями направления; они не становятся разрешёнными optional-задачами.

## Работа с резервом и blockers

- 20 резервных дней — банк из ежедневного плана, расходуемый перед зависимой проверкой. Это не обещание выполнить все optional-функции дополнительно.
- Приоритет: невыполненные обязательные критерии → regression/повторный soak → installation/docs → optional.
- Отсутствующее железо, нужная OS platform, аккаунт сервиса или календарные habit data дают статус BLOCKED соответствующей проверке, а не PASS.
- При двух подряд незакрытых недельных рубежах пересчитать оставшийся объём по фактической скорости.
- Новая функция получает отдельные требования, проверку, зависимости и оценку. Не помещать её в резерв молча.

## Ближайшая работа после локальных Д004–Д010

**Д011 — независимый repository/Core scaffold**, с открытыми M0 hardware gates. Выполненные и заблокированные cases перечислены в [execution report](DAYS_004_010_RESULT.md). Для Pi нужны install Д006, cooling/access Д007, HIL Д008 и LAN/mobile Д009 по готовым protocols. Физический PASS не ставится до реально выполненного опыта.

## Общий Definition of Done пакета

Связанные обязательные требования реализованы; positive/negative/edge checks пройдены на подходящем уровне; hardware features подтверждены HIL, где требуются; OpenAPI/config/examples/docs обновлены; simulation не сломана; secrets/private data отсутствуют; результаты привязаны к commit/artifact/model version. Финальный PASS может дать только реально выполненная проверка.

## Детализация ближайшего этапа

[Максимально подробный план Д004–Д010](plans/DAYS_004_010_DETAILED.md): host/sim vs Pi evidence, пятёрка рабочих блоков на день, ожидаемые артефакты, negative cases, критерии M0 и ветка без оборудования. Подготовка этого плана не означает выполнения экспериментов или закрытия требований.

## План следующего software-среза

[Д011–Д015 подробно](plans/DAYS_011_015_DETAILED.md): monorepo/tooling, config/lifecycle/logging, HTTP/health/OpenAPI/metrics, bounded Event Bus, fake adapters/native simulation/dev Compose/CI v0. План подготовлен; implementation ещё не выполнена. Hardware M0_NOT_READY остаётся отдельным ограничением.
