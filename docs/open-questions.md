# День 1. Открытые вопросы и реестр рисков

Источник: [ТЗ](../Smart_Home_Hub_TZ.md). Дата: 25 октября 2026 года. Связь: [матрица требований](requirements.md), [backlog](backlog.md), [план](DEVELOPMENT_PLAN_DAILY.md).


**Обновление Д002 (25.10.2026):** пользователь подтвердил отсутствие замеров и подготовку вариантов. Созданы [предварительный BOM](../hardware/bom/BOM.md), [power budget](../hardware/bom/power-budget.md), [схема](../hardware/wiring/wiring.md), [pin map](../hardware/wiring/pin-map.md) и [реестр источников](../hardware/research/SOURCES.md). На конец Д002 все 35 вопросов оставались открытыми; документальный/расчётный прогресс не заменяет принятие решения и HIL. [Отчёт Д002](DAY_002_RESULT.md).

**Реестр содержит 35 вопросов.** На Д003 часть вопросов получила проектное решение v0, часть — уточнение контракта; статусы указаны ниже. Предложения, не закреплённые в ADR, остаются вариантами исследования. Изменения обязательного объёма ТЗ не приняты. Выбор точного BOM/библиотек/OS versions и численные ML quality gates остаются открытыми там, где нет доказательств.

P0 — решить до зависимой реализации, иначе высок риск переделки/невыполнения обязательного требования. P1 — уточнить contract/config до соответствующего этапа. Это приоритет разработки, а не сообщение о существующем production incident.

Рабочий ответственный за техническое исследование — разработчик проекта. Владелец проекта определяет доступное железо/аккаунты, пользовательские ожидания и изменения объёма; при совместной работе это могут быть один человек. Не нужно блокировать независимые работы ожиданием всех ответов сразу.

Для закрытия: записать дату, evidence эксперимента/решения, выбранный вариант, ответственного, затронутые requirement IDs, ADR и влияние на срок. Если меняется исходное требование — отдельно отразить согласование изменения; предложение в этом файле само по себе его не заменяет.


**Обновление Д003 (25.10.2026):** созданы [архитектурные контракты v0 и 7 ADR](architecture/README.md), [логическая схема данных](architecture/data-model-v0.md) и [20 сценариев проверки](architecture/scenarios.md). «Решение v0» означает зафиксированную архитектуру, а не выполненную программную/HIL проверку. Подробный итог: [Д003](DAY_003_RESULT.md).

**Обновление Д004–Д010 (25.10.2026):** [локальная часть выполнена](DAYS_004_010_RESULT.md): Go/Python executable spikes, isolated PipeWire monitor, synthetic screen/inference/thermal и loopback TLS, два replay и20 code tests. Пользователь подтвердил отсутствие доступной/настроенной Pi. Physical gates остаются BLOCKED, [M0_NOT_READY](milestones/M0_TECHNICAL_BASELINE.md); target checks не закрыты host results.

## Сводка

| ID | Вопрос | Приоритет | Решить до | Статус |
|---|---|---|---|---|
| Q-001 | Граница автономной работы | P1 | Д003, проверка Д045/Д125 | Решение v0; проверка впереди |
| Q-002 | BOM, электрика и независимые LED-каналы | P0 | Д002–Д008, до монтажа Д034 | Частично решён |
| Q-003 | Носитель и миграция на SSD | P1 | Д006, процедура до Д139 | Открыт |
| Q-004 | Thermal states, fan curve и неисправности | P0 | Д003–Д007, до Д132 | Открыт |
| Q-005 | Цвет, presets и восстановление LED | P1 | Д003–Д005, до Д026 | Частично решён |
| Q-006 | Scene runtime и границы Python/Go | P0 | Д005, окончательно Д071 | Частично решён |
| Q-007 | Audio graph, capabilities и routing | P0 | Д004/Д008, до Д036–Д037 | Открыт |
| Q-008 | Доступность Spotify Connect | P1 | Д004, до Д042 | Открыт |
| Q-009 | Bluetooth trust и pairing policy | P1 | Д003, до Д038 | Частично решён |
| Q-010 | LAN, HTTPS, mDNS и trust bootstrap | P1 | Д003–Д009 | Открыт |
| Q-011 | Матрица платформ screen-client | P0 | Д005, доступ к стендам до Д061 | Открыт |
| Q-012 | Screen pairing и device-token lifecycle | P1 | Д003, до Д056 | Частично решён |
| Q-013 | Нагрузочный профиль и измерение производительности | P0 | Д005, до benchmarks каждого этапа | Открыт |
| Q-014 | Stream loss и screen inactivity | P1 | Д005, до Д060/Д065 | Частично решён |
| Q-015 | Качество Scene ML и конфликты меток | P0 | Д005, пороги до Д072 | Открыт |
| Q-016 | Камеры и драйверы Ubuntu | P0 | Д002–Д005, до Д081–Д083 | Открыт |
| Q-017 | Общая камера, privacy toggles и индикатор | P0 | Д003, до Д081–Д082 | Решение v0; проверка впереди |
| Q-018 | Жесты: activation, quality и reaction time | P0 | Д005, пороги до Д087 | Открыт |
| Q-019 | Presence fusion и отсутствие источников | P1 | Д003, до Д096–Д098 | Частично решён |
| Q-020 | Модели датчиков, единицы и ambient policy | P0 | Д002–Д005, до Д091 | Открыт |
| Q-021 | Privacy для runtime, fixtures, logs и backups | P0 | Д003–Д005, до collectors/CV | Частично решён |
| Q-022 | Habit schema, eligibility и confidence | P0 | Д003/Д021, до collector Д035 | Частично решён |
| Q-023 | Automation arbitration, overrides и timers | P0 | Д003, до Д032; детали до Д101 | Частично решён |
| Q-024 | Auth, roles, scopes и API semantics | P0 | Д003, до Д017–Д020 | Частично решён |
| Q-025 | MQTT discovery, topics и HA compatibility | P1 | Д003, protocol до Д121–Д123 | Открыт |
| Q-026 | Публичный repo, CI images и доверие runner | P0 | Д011/Д015, окончательно до Д141 | Открыт |
| Q-027 | Telegram linking, commands и alerts | P1 | Д003, до Д124 | Открыт |
| Q-028 | SQLite/config/backup/schema rollback | P0 | Д012–Д016, детали до Д129/Д143 | Частично решён |
| Q-029 | PWA и пользовательские flows | P1 | Д022, детали до соответствующих UI этапов | Открыт |
| Q-030 | Events, WS, logs и security retention | P1 | Д012–Д014, до Д023 | Частично решён |
| Q-031 | Readiness, degradation и watchdog | P0 | Д003/Д013, до Д126/Д142 | Решение v0; проверка впереди |
| Q-032 | Защищённый Safe Mode без SQLite | P0 | Д003, реализация до Д128 | Решение v0; проверка впереди |
| Q-033 | DSP format, bands и quality | P1 | Д005/Д037, до Д047–Д049 | Открыт |
| Q-034 | Metrics/alerts profile и laptop telemetry | P1 | Д023, каталог до Д131 | Открыт |
| Q-035 | Coverage, acceptance evidence и test scope | P1 | Д001 постановка, методика до Д005/Д015 | Открыт |

## Q-001. Граница автономной работы

- **Источник:** §1,9,25,52 ТЗ.
- **Приоритет / срок:** P1; Д003, проверка Д045/Д125.
- **Что выяснить:** Какие пользовательские сценарии обязательны без WAN, как показывать недоступность Spotify/Telegram?
- **Предлагаемый путь:** Считать local audio/LED/sensors/rules/local inference базовыми; внешние сервисы деградируют независимо.
- **Условие закрытия:** Список offline-сценариев и ожидаемых ошибок согласован и внесён в acceptance.
- **Затронутые требования:** ARCH-001, TG-006.
- **Прогресс Д003:** ADR-001 и сценарий S12: local manual LED/local audio/sensors/non-ML rules не зависят от WAN/Telegram/ML. Spotify/Telegram — независимые external capabilities; все offline smoke остаются будущими проверками.
- **Прогресс Д004–Д010:** Local probes используют synthetic/loopback inputs и не зависят от Telegram/cloud inference. Это не acceptance автономности ещё не реализованного Core; соответствующие runtime failure checks впереди.
- **Статус:** Решение v0; проверка впереди; продуктовая проверка не выполнена.

## Q-002. BOM, электрика и независимые LED-каналы

- **Источник:** §3–6,19,45 ТЗ.
- **Приоритет / срок:** P0; Д002–Д008, до монтажа Д034.
- **Что выяснить:** Каковы длины лент/число LED, реальный PSU, pinout, bus addresses, уровни/защита и обязательность элементов §5?
- **Предлагаемый путь:** Рассчитать нагрузку и подтвердить два канала вместе с audio; §5 смешивает обязательное/рекомендуемое, предложить включить перечисленную защиту в BOM.
- **Условие закрытия:** BOM/wiring/pin map, расчёт тока и HIL spike; обязательность hardware элементов явно записана.
- **Затронутые требования:** HW-001, HW-005, HW-006, HW-007, HW-008, HW-009, HW-010, HW-011, HW-012, LIGHT-001, LIGHT-002, POWER-001, POWER-002, POWER-003, POWER-004, AMBI-010, BTN-001, DEV-003, TEST-008.
- **Прогресс Д002:** BOM B01–B44 и варианты S0–S3 готовы; размеры/наличие остаются неизвестны. По S02 выявлен PWM/audio conflict; предложен SPI-chain с independent logical zones, физическая независимость и hard-off пока не подтверждены. Wiring/pin map — draft, fuse values и cable sections TBD.
- **Прогресс Д003:** ADR-007 принимает zone/transport separation и независимость control state; SPI-chain остаётся provisional. Hardware topology, counts, protection и physical hard-off проверяются Д004/Д008, замеры отсутствуют.
- **Прогресс Д004–Д010:** Go spike подтвердил независимые logical slices и idle-aware current limits на synthetic fixtures. Physical SPI chain, power sequencing, PSU/wiring/counts всё ещё не проверены: доступной Pi нет.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

**Уточнение владельца26.10.2026:** Pi4B с2GB RAM есть; SSH/физические замеры ещё не предоставлены. [Ограничения2GB](development/pi4b-2gb.md). Размеры LED/BOM/питание и hardware gates остаются открытыми.

## Q-003. Носитель и миграция на SSD

- **Источник:** §3.1,35 ТЗ.
- **Приоритет / срок:** P1; Д006, процедура до Д139.
- **Что выяснить:** Какие storage/USB power/boot конфигурации доступны и что переносится при migration?
- **Предлагаемый путь:** Начать на доступном носителе, предусмотреть проверенный backup и возврат; не выбирать модель без проверки.
- **Условие закрытия:** Проверяемая процедура boot/data migration с rollback и список оборудования.
- **Затронутые требования:** HW-002.
- **Прогресс Д002:** BOM разделяет bootstrap microSD и целевой USB SSD; exact носитель/enclosure/наличие и USB budget неизвестны. Migration/rollback остаются будущей проверкой Д139.
- **Прогресс Д004–Д010:** Подготовлены install/recovery/seed templates и candidate image manifest. Image/storage boot/microSD→SSD migration не выполнялись; published SHA не равен verification отсутствующего image.
- **Статус:** открыт; подготовлен проектный вариант, финальное решение/стендовое доказательство не получено.

## Q-004. Thermal states, fan curve и неисправности

- **Источник:** §3.2,38 ТЗ.
- **Приоритет / срок:** P0; Д003–Д007, до Д132.
- **Что выяснить:** Как 50/65/75°C соотносятся с четырьмя состояниями, hysteresis и fan speeds? Как реагировать на failed read?
- **Предлагаемый путь:** Разделить thermal states, fan curve и warnings; использовать safe fan state при потере температуры.
- **Условие закрытия:** Таблица переходов/диапазонов, схема управления и тесты порогов.
- **Затронутые требования:** HW-003, HW-004, THERM-001, THERM-002, THERM-003.
- **Прогресс Д002:** Зарезервированы 5V fan и GPIO16 control через отдельный driver; current/model/default failure state не выбраны. Постоянно работающий fan допустим для раннего стенда; thresholds и production control остаются решением Д003/Д007.
- **Прогресс Д004–Д010:** Thermal policy v0 прошла synthetic transitions/hysteresis/invalid samples. 50/65/75°C и hysteresis3°C — project defaults; physical fan/current/default polarity и actual load остаются BLOCKED.
- **Статус:** открыт; подготовлен проектный вариант, финальное решение/стендовое доказательство не получено.

## Q-005. Цвет, presets и восстановление LED

- **Источник:** §4,11.5–11.8,14,36 ТЗ.
- **Приоритет / срок:** P1; Д003–Д005, до Д026.
- **Что выяснить:** Каков порядок modifiers/gamma/fade/limiter, Night Shift defaults/tolerance и смысл last state при live modes?
- **Предлагаемый путь:** Зафиксировать color pipeline и golden fixtures; сохранять logical mode/settings, не историю кадров; live-source recovery описать явно.
- **Условие закрытия:** Color contract, default profile/preset semantics и restart scenarios.
- **Затронутые требования:** LIGHT-004, LIGHT-005, LIGHT-006, LIGHT-007, LIGHT-009, LIGHT-010, AMBI-011, AMBI-012, AMBI-013, AMBI-014, AMBI-015, AMBI-017, BTN-004, WEB-007, REL-008.
- **Прогресс Д003:** Domain/state contracts разделяют desired/applied state, transient shutdown fade не затирает saved state. Safety limiter финальный и непреодолимый. Точный gamma/color/fade порядок, coefficients и golden tolerances всё ещё нужны до Д026.
- **Прогресс Д004–Д010:** Synthetic weighted edge colors/corner exclusion и logical slice mapping проверены. Gamma/fade/full color pipeline, calibration и restart live modes требуют production implementation/HIL.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-006. Scene runtime и границы Python/Go

- **Источник:** §2,7,12,36,40 ТЗ.
- **Приоритет / срок:** P0; Д005, окончательно Д071.
- **Что выяснить:** Каким runtime Go-client выполняет локальный Scene ML на четырёх targets и что означает ML Runtime failure для laptop ML?
- **Предлагаемый путь:** Сравнить embedded inference и локальный worker; независимо проверить отказ Pi Runtime и laptop inference.
- **Условие закрытия:** ADR, запуск runtime на всех targets и failure contract без передачи кадров Hub.
- **Затронутые требования:** ARCH-003, AMBI-001, SCENE-002, REL-003, PRIV-002.
- **Прогресс Д003:** ADR-001 фиксирует local-only границы: Pi Runtime и laptop Scene ML независимы. S12 проверяет их отказы отдельно. Embedded/native worker и delivery модели на четыре targets ещё не выбраны.
- **Прогресс Д004–Д010:** Separate stdlib Python logistic demo-worker выполнен локально: <=1Hz, bounded queue, failure не остановил synthetic30Hz loop. Это не actual Scene model/quality; ONNX и four-target packaging не проверены.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-007. Audio graph, capabilities и routing

- **Источник:** §9–10,43 ТЗ.
- **Приоритет / срок:** P0; Д004/Д008, до Д036–Д037.
- **Что выяснить:** Какие host daemons/API дают A2DP/Spotify/local playback, общий PCM и все outputs? Какие controls поддерживает каждый source?
- **Предлагаемый путь:** Проверить один audio graph на Pi и ограниченный IPC из Core; source-specific limitations не выдавать за выполненные controls.
- **Условие закрытия:** Воспроизводимый spike, source×output×control matrix и production setup.
- **Затронутые требования:** AUDIO-001, AUDIO-003, AUDIO-006, AUDIO-007, AUDIO-008, AUDIO-009, AUDIO-010, AUDIO-012, AUDIO-013, AUDIO-014, AUDIO-015, AUDIO-016, AUDIO-017.
- **Прогресс Д002:** Документация rpi_ws281x подтверждает конфликт PWM с встроенным analog audio. Вариант SPI-chain сохраняет ресурс PWM, но actual audio coexistence, host graph, permissions и PCM tap не тестировались; USB DAC сам по себе не закрывает встроенный 3.5mm requirement.
- **Прогресс Д004–Д010:** Isolated PipeWire1.0.5/WirePlumber0.4.17 null-sink→monitor PCM действительно работает, consumer reconnect не остановил playback. Host recipe не переносится на WirePlumber0.5 без migration. Analog/Bluetooth/USB/Spotify и Pi coexistence BLOCKED.
- **Статус:** открыт; подготовлен проектный вариант, финальное решение/стендовое доказательство не получено.

## Q-008. Доступность Spotify Connect

- **Источник:** §9.1 ТЗ.
- **Приоритет / срок:** P1; Д004, до Д042.
- **Что выяснить:** Есть ли подходящая учётная запись и разрешённый совместимый клиент на выбранной ОС/architecture?
- **Предлагаемый путь:** Проверить условия/совместимость при реализации; не обещать внешнюю функцию по одному локальному stub.
- **Условие закрытия:** Фактический playback/control smoke и documented prerequisites либо блокер этапа.
- **Затронутые требования:** AUDIO-002.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-009. Bluetooth trust и pairing policy

- **Источник:** §9.2,21,39 ТЗ.
- **Приоритет / срок:** P1; Д003, до Д038.
- **Что выяснить:** Кто включает PAIRING_MODE, каковы TTL, trusted reconnect, удаление устройства и outbound pairing?
- **Предлагаемый путь:** Новые устройства только в explicit window; trusted reconnect отдельно; pairing events логировать без secrets.
- **Условие закрытия:** Pairing state machine и negative tests неизвестных устройств.
- **Затронутые требования:** AUDIO-005, DEV-007.
- **Прогресс Д003:** State machine pairing: explicit admin window, default candidate TTL120sec, trusted reconnect отдельно; новые pair вне окна запрещены. Actual BlueZ flow/role concurrency/permissions остаются spike Д038.
- **Прогресс Д004–Д010:** Actual Bluetooth unavailable: нет target adapter/phone session и доступа к system D-Bus. PAIRING_MODE/trusted reconnect/roles на hardware не закрыты виртуальным PipeWire test.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-010. LAN, HTTPS, mDNS и trust bootstrap

- **Источник:** §29–30 ТЗ.
- **Приоритет / срок:** P1; Д003–Д009.
- **Что выяснить:** Кто выпускает CA/certificates, как доверять на телефоне, что при конфликте имени и нужен ли remote VPN?
- **Предлагаемый путь:** Подготовить локальный CA/trust runbook; LAN-only по умолчанию; VPN отдельно и только при необходимости.
- **Условие закрытия:** Проверка HTTPS/mDNS с клиентов и согласованный remote-access scope.
- **Затронутые требования:** API-005, API-006, API-007, API-008.
- **Прогресс Д004–Д010:** Ephemeral loopback HTTPS proxy/CA/SAN, negative trust/hostname и leaf replacement прошли. mDNS/Pi LAN/phone trust/reboot и actual Caddy не проверялись; examples/runbook готовы.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-011. Матрица платформ screen-client

- **Источник:** §11.1 ТЗ.
- **Приоритет / срок:** P0; Д005, доступ к стендам до Д061.
- **Что выяснить:** Какие OS versions/Linux sessions/permissions/multiple displays/protected content поддерживаются, доступны ли обе macOS architectures?
- **Предлагаемый путь:** Зафиксировать проверяемую platform matrix и реальные test machines; cross-compile не считать проверкой capture.
- **Условие закрытия:** По каждой обязательной platform есть окружение, package install и capture evidence.
- **Затронутые требования:** AMBI-002, AMBI-003, AMBI-004, AMBI-005, AMBI-006.
- **Прогресс Д004–Д010:** Four-target matrix записана. Linux live capture attempt дала OSError в доступном окружении; synthetic analyzer проходит. Windows/macOS arm64/macOS Intel отсутствуют и остаются BLOCKED.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-012. Screen pairing и device-token lifecycle

- **Источник:** §11.3,27 ТЗ.
- **Приоритет / срок:** P1; Д003, до Д056.
- **Что выяснить:** Каковы code format/TTL/rate limits, scopes, revocation активного stream и локальное token storage?
- **Предлагаемый путь:** Одноразовый ограниченный code, device-scoped token, revoke прекращает commands/stream; storage выбрать по ОС.
- **Условие закрытия:** Protocol/security contract и replay/expired/revoked tests.
- **Затронутые требования:** AMBI-009.
- **Прогресс Д003:** Commands/events фиксируют one-time code, resource binding, source spoof rejection, revoke active WS/stream и epoch invalidation. Code format/TTL/rate и secure platform token store требуют деталей Д020/Д056.
- **Прогресс Д004–Д010:** Strict v0 packet validator отклоняет wrong binding/epoch/version/sequence/RGB/extra image fields. Реального token pairing/revoke/WS transport нет; это не проверка network authentication.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-013. Нагрузочный профиль и измерение производительности

- **Источник:** §11.7,12.2,15.1,45–46 ТЗ.
- **Приоритет / срок:** P0; Д005, до benchmarks каждого этапа.
- **Что выяснить:** При каких LED count/resolution/network/concurrency/активных модулях действуют targets, как измерять clocks/latency/p95?
- **Предлагаемый путь:** Задать reference joint load, intervals/точки измерения и погрешность; учитывать applied frames.
- **Условие закрытия:** Benchmark plan с численными профилями и методикой; после прогона отчёты воспроизводимы.
- **Затронутые требования:** AMBI-016, AMBI-018, SCENE-004, PRES-005, TEST-009, PERF-001, PERF-002, PERF-003, PERF-004, PERF-005, PERF-007, PERF-008.
- **Прогресс Д004–Д010:** Методики всех8 performance targets/reference profile определены, tiny host timings сохранены. REST/WS/full physical Ambilight/gesture/presence targets пока NOT_RUN, synthetic30Hz не physical FPS.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-014. Stream loss и screen inactivity

- **Источник:** §11.9–11.10 ТЗ.
- **Приоритет / срок:** P1; Д005, до Д060/Д065.
- **Что выяснить:** Каковы timeout/static thresholds, что значит monitor off, когда fallback/off и как восстанавливаться при override?
- **Предлагаемый путь:** Явная state machine active/stale/inactive/disconnected с configurable thresholds и общим arbitration.
- **Условие закрытия:** Timeline fixtures и реальные unplug/static/recovery сценарии.
- **Затронутые требования:** AMBI-019, AMBI-020, AMBI-021.
- **Прогресс Д003:** Screen state разделяет connection/activity/render eligibility. Fallback не перезаписывает desired AMBILIGHT; Master OFF invalidates epoch, reconnect не включает свет. Точные timeouts/static detection требуют Д060/Д065.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-015. Качество Scene ML и конфликты меток

- **Источник:** §12 ТЗ.
- **Приоритет / срок:** P0; Д005, пороги до Д072.
- **Что выяснить:** Какие precision/recall/F1 нужны каждому классу, что означают CALM/DYNAMIC, как разрешать противоположные labels?
- **Предлагаемый путь:** Задать rubric/holdout/thresholds до обучения; baseline и per-label calibration; ML влияет только на разрешённые параметры.
- **Условие закрытия:** Dataset manifest, quality gate и label→parameter policy; численные пороги не выдуманы в День 1.
- **Затронутые требования:** SCENE-001, SCENE-003, SCENE-006, SCENE-007.
- **Прогресс Д004–Д010:** Записаны provisional quality gates/rubric/split/freeze policy до обучения. Actual девятиклассовая модель/dataset evaluation не выполнялись; tiny demo не quality evidence.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-016. Камеры и драйверы Ubuntu

- **Источник:** §13.1,15 ТЗ.
- **Приоритет / срок:** P0; Д002–Д005, до Д081–Д083.
- **Что выяснить:** Какие USB webcam/Pi Camera доступны и поддерживаются выбранной Ubuntu/capture stack?
- **Предлагаемый путь:** Проверить оба типа отдельно, закрепить formats/resolution и permissions после spike.
- **Условие закрытия:** BOM и capture/reconnect HIL по двум типам; неподдерживаемая камера заменена/срок пересчитан.
- **Затронутые требования:** GEST-001.
- **Прогресс Д002:** В BOM предусмотрены отдельные проверки USB UVC и Pi Camera Module 2/3, без фиктивного exact model/наличия. Изучены официальные camera/Ubuntu docs; поддержку выбранной Ubuntu/kernel/packages надо подтвердить на стенде.
- **Прогресс Д004–Д010:** Ubuntu Server26.04.1 arm64+raspi выбран documentary candidate, published checksum fetched. Pi/CSI/UVC/kernel compatibility на target не проверена; камера не подключалась.
- **Статус:** открыт; подготовлен проектный вариант, финальное решение/стендовое доказательство не получено.

## Q-017. Общая камера, privacy toggles и индикатор

- **Источник:** §13.2,15.1–15.2,36 ТЗ.
- **Приоритет / срок:** P0; Д003, до Д081–Д082.
- **Что выяснить:** Может ли presence использовать камеру при выключенном Gesture Mode? Что означает camera disabled until required после restart и что при отказе indicator?
- **Предлагаемый путь:** Один camera owner, независимые explicit consumer permissions, indicator на любой capture; после restart capture выключен до разрешённой потребности.
- **Условие закрытия:** Согласованная lifecycle/consent/failure state machine и HIL privacy tests.
- **Затронутые требования:** GEST-002, PRES-003, PRES-006, REL-011.
- **Прогресс Д003:** ADR-003: один capture owner, два независимых consumer permissions, current-boot arm, default OFF после restart и indicator ordering. Known indicator I/O failure блокирует capture; физический burnout без feedback автоматически не диагностируется. HIL Д081–Д100 впереди.
- **Статус:** Решение v0; проверка впереди; продуктовая проверка не выполнена.

## Q-018. Жесты: activation, quality и reaction time

- **Источник:** §13.3–13.4,46 ТЗ.
- **Приоритет / срок:** P0; Д005, пороги до Д087.
- **Что выяснить:** Какой activation gesture, window duration, cooldown, зеркальность/rotation thresholds? Как совместить 300–500мс confirmation с <500мс reaction?
- **Предлагаемый путь:** Измерять от начала устойчивого жеста с confirmation; выбрать 300мс как кандидат и измерить весь budget, задать false actions/min.
- **Условие закрытия:** Gesture vocabulary/default mappings, timing/quality gates и negative fixtures.
- **Затронутые требования:** GEST-003, GEST-004, GEST-005, GEST-006, GEST-009, GEST-010, GEST-011, GEST-012, GEST-013, GEST-014, GEST-015, GEST-016, GEST-017, GEST-018, GEST-019, PERF-006.
- **Прогресс Д004–Д010:** Gesture quality/timing методика и negative-sequence threshold draft описаны. Confirmation внутри500ms budget, actual gestures/model/activation hardware tests не проводились.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-019. Presence fusion и отсутствие источников

- **Источник:** §15 ТЗ.
- **Приоритет / срок:** P1; Д003, до Д096–Д098.
- **Что выяснить:** Каковы weights/TTL/hysteresis, связь телефона с user и смысл UNKNOWN/AWAY при sleep/MAC randomization?
- **Предлагаемый путь:** Не считать один missing ping доказательством AWAY; выводить freshness/причины и проверять source conflicts.
- **Условие закрытия:** Fusion table/timelines и настройка phone/network identifiers.
- **Затронутые требования:** PRES-001, PRES-002.
- **Прогресс Д003:** Domain/events различают UNKNOWN/AWAY, evidence freshness и nullable signals. Трёхзначные conditions не трактуют UNKNOWN как AWAY. Weights/TTL/hysteresis и real phone/network behavior ещё не выбраны.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-020. Модели датчиков, единицы и ambient policy

- **Источник:** §16,21 ТЗ.
- **Приоритет / срок:** P0; Д002–Д005, до Д091.
- **Что выяснить:** Какие модели/шины/addresses/rates/calibration/air-quality units? Точно ли brighter room должно уменьшать LED brightness?
- **Предлагаемый путь:** Сохранить направление из ТЗ до явного изменения; не называть косвенный air-quality index измеренным CO₂; stale отделить от нуля.
- **Условие закрытия:** BOM, measurement contracts/units и согласованная lux→brightness curve.
- **Затронутые требования:** SENS-001, SENS-002, SENS-003, SENS-004, SENS-005, SENS-006, SENS-007, SENS-008.
- **Прогресс Д002:** Предложены 3.3V VEML7700 0x10, BME280 0x76/0x77, SGP40 0x59 и PIR EKMC1601111. Адреса документально различны; exact breakout/pull-ups/наличие/calibration не проверены. SGP40 оформлен как VOC Index; ambient direction из ТЗ не изменён.
- **Прогресс Д004–Д010:** Read-only preflight не нашёл I2C/GPIO devices; sensor adapters/readings не реализованы/не проверены. Кандидаты/адреса остаются документальными, реальные quantities и платы неизвестны.
- **Статус:** открыт; подготовлен проектный вариант, финальное решение/стендовое доказательство не получено.

## Q-021. Privacy для runtime, fixtures, logs и backups

- **Источник:** §11–13,17,25,31,33,35,40 ТЗ.
- **Приоритет / срок:** P0; Д003–Д005, до collectors/CV.
- **Что выяснить:** Какие synthetic/licensed fixtures разрешены, что попадает в diagnostics; как delete/reset действует на backups и восстановленные модели?
- **Предлагаемый путь:** Runtime images только RAM/laptop; redact logs, не коммитить private fixtures; определить purge/restore policy backup copies до обещания полного удаления.
- **Условие закрытия:** Data-flow/storage inventory, fixture license policy и проверяемый delete/reset/restore contract.
- **Затронутые требования:** AUDIO-011, DSP-011, AMBI-007, AMBI-008, SCENE-008, GEST-007, HAB-011, HAB-013, WEB-008, TG-005, STORE-003, REC-005, PRIV-001, PRIV-003, TEST-005.
- **Прогресс Д003:** ADR-006/data model фиксируют RAM-only realtime data, purpose-specific persistence, privacy_epoch, отмену stale training и purge Hub-managed backups. Restore сверяет protected tombstone вне БД. Реализация/fixture licenses/backup integration и внешние export limits должны быть проверены Д118/Д129.
- **Прогресс Д004–Д010:** Probes не сохраняли raw PCM/screenshots; fixed RGB/temperature fixtures synthetic, TLS keys temporary0600 и удалены. Full camera/habit/backup/privacy implementation не выполнена; no inference cloud used.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-022. Habit schema, eligibility и confidence

- **Источник:** §17 ТЗ.
- **Приоритет / срок:** P0; Д003/Д021, до collector Д035.
- **Что выяснить:** Как получить volume target без поля volume, сколько реальных данных достаточно, отсчёт 30 дней и confidence defaults, retention границы?
- **Предлагаемый путь:** Добавить минимальный контекст, не треки; timestamp gate плюс data quality; 60/85% использовать лишь как предложенные defaults.
- **Условие закрытия:** Feature/schema contract, thresholds/calibration plan, eligibility/retention boundary tests.
- **Затронутые требования:** HAB-001, HAB-003, HAB-004, HAB-006, HAB-007, HAB-008, HAB-009, HAB-012, HAB-014.
- **Прогресс Д003:** Схема v0 добавляет nullable volume/presence/ambient/monitor/origin/quality, UTC timestamps+timezone и learning epoch. >=30 реальных дней недостаточно без data quality; confidence numbers/калибровка остаются открыты. Feedback/delete semantics заданы.
- **Прогресс Д004–Д010:** Temporal/data-sufficiency/calibration методика записана, но collector/user observations/30 реальных дней ещё не начаты; recommendations/AUTO не объявлены готовыми.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-023. Automation arbitration, overrides и timers

- **Источник:** §16.2,18–19,49 ТЗ.
- **Приоритет / срок:** P0; Д003, до Д032; детали до Д101.
- **Что выяснить:** Как решаются равные priority, partial multi-actions, unknown/stale, циклы, source manual, scope/TTL override, delayed restart/cancel, timezone/missed schedules?
- **Предлагаемый путь:** Единый command layer, deterministic tie-break, idempotency и cancellation/revalidation; не обходить music protection.
- **Условие закрытия:** Rule/command semantics, conflict tables и timeline/restart tests.
- **Затронутые требования:** SENS-009, HAB-005, AUTO-001, AUTO-004, AUTO-005, AUTO-006, AUTO-007, AUTO-008, AUTO-009, AUTO-010, AUTO-011, AUTO-012, AUTO-013, AUTO-014, AUTO-015, AUTO-016, AUTO-017, AUTO-018, AUTO-019, AUTO-020, AUTO-021, AUTO-022, AUTO-023, AUTO-024, AUTO-025, BTN-002, WEB-009, REL-010.
- **Прогресс Д003:** ADR-002/commands фиксируют field-scoped manual override900sec, Master OFF wildcard, source mapping, stable tie-break, guarded ownership, delayed revalidation/cancel и no overdue catch-up. Partial hardware outcomes и неидемпотентный uncertain replay заданы. Queue/depth/clock tolerances и implementation tests впереди.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-024. Auth, roles, scopes и API semantics

- **Источник:** §26–28 ТЗ.
- **Приоритет / срок:** P0; Д003, до Д017–Д020.
- **Что выяснить:** Как bootstrap admin, параметры Argon2id/session expiry, SameSite/CSRF/WS origin, scopes/revoke, точные rate limits и CRUD resources?
- **Предлагаемый путь:** Без default passwords; hash secrets; least privilege; параметры и дополнительные защитные меры оформить в auth ADR.
- **Условие закрытия:** Permission/API/session contract и negative tests; примеры rate limits не объявлены обязательными числами.
- **Затронутые требования:** ARCH-005, AUTH-001, AUTH-002, AUTH-003, AUTH-004, AUTH-005, AUTH-007, AUTH-008, AUTH-010, AUTH-011, API-002, API-004.
- **Прогресс Д003:** Командный gateway назначает source/principal, токены ограничены scopes+resource bindings, revoke проверяется перед cached result/execution. Peer-auth emergency path определён отдельно. Argon/session defaults, CSRF и конкретные API/auth schemas завершаются Д017–Д020.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-025. MQTT discovery, topics и HA compatibility

- **Источник:** §21–23,48 ТЗ.
- **Приоритет / срок:** P1; Д003, protocol до Д121–Д123.
- **Что выяснить:** Каков discovery payload, IDs, retained/LWT/QoS, ACL namespace, unknown-client audit и связь с HA?
- **Предлагаемый путь:** Versioned idempotent device registry, broker auth/ACL; базовый REST/MQTT без обязательного HA Discovery.
- **Условие закрытия:** Topic/schema/ACL contract, fake/real discovery и внешний client smoke.
- **Затронутые требования:** MQTT-001, MQTT-003, MQTT-004, MQTT-005, MQTT-006, MQTT-007, MQTT-009, SIM-005.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-026. Публичный repo, CI images и доверие runner

- **Источник:** §2,33,41,43–44 ТЗ.
- **Приоритет / срок:** P0; Д011/Д015, окончательно до Д141.
- **Что выяснить:** Кто владеет repo/registry, какие versions/digests/scanners, где privileged exceptions, как защищены release tags и runner от fork PR?
- **Предлагаемый путь:** Проверки PR на изолированных runners; production runner только доверенный deploy; секреты вне repo/artifacts.
- **Условие закрытия:** CI/release permissions map, image policy, trusted-trigger tests и runner runbook.
- **Затронутые требования:** ARCH-006, STORE-006, STORE-008, CI-001, CI-002, CI-003, CI-004, CI-005, CI-007, CI-008, CI-010, CI-011, DEPLOY-001, DEPLOY-002, DEPLOY-003, DEPLOY-004.
- **Прогресс Д004–Д010:** Host versions/evidence/config examples готовы, Go tests с-race проходят. Docker daemon/production container permissions/non-root Pi driver, registry/runner/deploy gates не проверены. Прямой Go binary заменил неработающий snap launcher без смены системной конфигурации.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-027. Telegram linking, commands и alerts

- **Источник:** §25,38 ТЗ.
- **Приоритет / срок:** P1; Д003, до Д124.
- **Что выяснить:** Кто создаёт bot token, как linked user доказывает local identity, какие inline actions/alerts допустимы?
- **Предлагаемый путь:** Одноразовая привязка из local authenticated flow, scoped Core token, redacted payload и независимая failure policy.
- **Условие закрытия:** Command/linking contract, unlinked denial и offline bot tests.
- **Затронутые требования:** TG-001, TG-003, TG-004, TG-007, TG-008, TG-009, TG-010, TG-011, TG-012, TG-013, OBS-007.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-028. SQLite/config/backup/schema rollback

- **Источник:** §31–32,35,44 ТЗ.
- **Приоритет / срок:** P0; Д012–Д016, детали до Д129/Д143.
- **Что выяснить:** Приоритет config, validation/reload, migration compatibility, backup consistency/time/retention, restore secrets и потеря новых writes при rollback?
- **Предлагаемый путь:** Online snapshot/WAL-aware backup, manifest и pre-deploy recovery point; явно определить schema rollback и maintenance window.
- **Условие закрытия:** Storage/config ADR, реальный restore и failed migration rollback test.
- **Затронутые требования:** STORE-001, STORE-002, STORE-004, STORE-005, REC-001, REC-002, REC-003, REC-004, DEPLOY-006.
- **Прогресс Д003:** ADR-006 и data-model-v0 задают ownership config, logical schema/indexes, transaction/effect boundary и privacy-aware backups/restore. Safe Mode recovery по ADR-005. SQL migrations/backup implementation/schema rollback и measured WAL behavior ещё не проверены.
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-029. PWA и пользовательские flows

- **Источник:** §24 ТЗ.
- **Приоритет / срок:** P1; Д022, детали до соответствующих UI этапов.
- **Что выяснить:** Какие browser/mobile targets, install/offline behavior, cache/private data и UX Settings/Devices нужны?
- **Предлагаемый путь:** PWA shell без обещания hardware control при недоступном Hub; не кешировать sensitive API responses; UI внутри этапов.
- **Условие закрытия:** UI route/flow list, browser matrix и проверяемый PWA contract.
- **Затронутые требования:** WEB-002, WEB-003, WEB-010, WEB-011, WEB-012, WEB-013, WEB-014, WEB-015, WEB-016, WEB-017, WEB-018, WEB-019.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-030. Events, WS, logs и security retention

- **Источник:** §8,24.2,38–39 ТЗ.
- **Приоритет / срок:** P1; Д012–Д014, до Д023.
- **Что выяснить:** Какие event schemas/versioning/queue overflow/order/reconnect, correlation IDs, log retention и redaction?
- **Предлагаемый путь:** Bounded queues/latest-frame для быстрых потоков, надёжная semantics управляющих events; IDs и redaction.
- **Условие закрытия:** Event/log contract и slow consumer/reconnect/security event tests.
- **Затронутые требования:** ARCH-008, DEV-005, WEB-005, WEB-006, OBS-001, OBS-002, OBS-036, OBS-037, OBS-038, OBS-039, OBS-040, OBS-041.
- **Прогресс Д003:** Events v0 задаёт envelope/version/producer identity, bounded delivery classes, sequence/epoch, snapshot/resync и direct durable audit writes. Exact queue sizes, log retention и validation/load checks остаются Д014/Д023/Д131.
- **Прогресс Д014–Д015:** Bounded bus/envelope реализованы, default queue64/max subscriptions128/payload16KiB; payload ownership, overflow, race/filter/order tests и host benchmark выполнены. JSON logs с redaction проверены. WS/reconnect/durable security audit/retention остаются открытыми; [отчёт](DAYS_011_015_RESULT.md).
- **Статус:** Частично решён; продуктовая проверка не выполнена.

## Q-031. Readiness, degradation и watchdog

- **Источник:** §36–38,44 ТЗ.
- **Приоритет / срок:** P0; Д003/Д013, до Д126/Д142.
- **Что выяснить:** Должен ли ML/MQTT outage делать /ready false при работающем ядре? Какие modules обязательны профилю и когда watchdog restart?
- **Предлагаемый путь:** Отдельные critical ready/degraded details плюс deploy smoke обязательного профиля; timeout/backoff без crash loops.
- **Условие закрытия:** Health/dependency matrix и failure/release tests без ложного успешного deploy.
- **Затронутые требования:** REL-001, REL-002, REL-004, REL-005, REL-012, REL-013, REL-014, OBS-004, OBS-005, CI-009, DEPLOY-005.
- **Прогресс Д003:** ADR-004 разделяет baseline critical ready и optional degraded; все dependencies проверяются. Release manifest дополнительно проверяет required local services и smoke. Safe Mode ready503 при live emergency loop. Watchdog intervals/profile smoke и fault/deploy tests впереди.
- **Прогресс Д004–Д010:** Release-profile template готов, M0 evaluated NOT_READY. Python TLS stub /health — bootstrap marker, не реализация Core readiness. Required-service smoke/rollback остаются будущими.
- **Прогресс Д013–Д015:** Настоящие Core health/ready реализованы; critical/stale/optional/disabled/start/stop/Safe Mode truth table проверена injection tests. Native processes проверены с memory_fake ready200 и missing storage ready503; Compose lifecycle/restart проверен. Full release profile, watchdog и hardware M0 остаются открытыми.
- **Статус:** Решение v0; проверка впереди; продуктовая проверка не выполнена.

## Q-032. Защищённый Safe Mode без SQLite

- **Источник:** §26–27,36 ТЗ.
- **Приоритет / срок:** P0; Д003, реализация до Д128.
- **Что выяснить:** Как manual lighting/audio аутентифицируются без БД, что с session/token revoke и как выходить из Safe Mode?
- **Предлагаемый путь:** Ограниченный локальный emergency path; не открывать anonymous LAN controls; credentials/recovery описать отдельно.
- **Условие закрытия:** Safe Mode access matrix и HIL disk/corruption/security tests.
- **Затронутые требования:** REL-006, REL-007.
- **Прогресс Д003:** ADR-005 выбирает local Unix socket CLI с restrictive OS peer permissions и RAM-only allowlisted controls. DB-backed LAN auth fail-closed, automation/persistence disabled. Recovery explicit с сохранением emergency Master OFF до согласованного reconcile. Socket/container UID mapping и HIL Д128 ещё не выполнены.
- **Статус:** Решение v0; проверка впереди; продуктовая проверка не выполнена.

## Q-033. DSP format, bands и quality

- **Источник:** §10,45 ТЗ.
- **Приоритет / срок:** P1; Д005/Д037, до Д047–Д049.
- **Что выяснить:** Каковы PCM sample rate/channels/windows/bands, normalization/BPM confidence и пороги beat/BPM quality?
- **Предлагаемый путь:** Один общий PCM contract и synthetic tones/metronome baseline; silence/unknown не выдавать за достоверный beat/BPM.
- **Условие закрытия:** DSP spec, golden fixtures и заранее принятые error/precision/recall thresholds.
- **Затронутые требования:** DSP-001, DSP-002, DSP-003, DSP-004, DSP-006, DSP-007.
- **Прогресс Д004–Д010:** Реально проверен 48kHz stereo S16LE contract на synthetic tone/silence и isolated monitor. Frequency active window/RMS/sample counts валидны; source audio/DSP FFT/beat/BPM quality и sensitivity ещё впереди.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-034. Metrics/alerts profile и laptop telemetry

- **Источник:** §38,47 ТЗ.
- **Приоритет / срок:** P1; Д023, каталог до Д131.
- **Что выяснить:** Как scene_inference_ms из ноутбука попадает в metrics, units/types/labels, alert thresholds/dedupe/clear и optional Prometheus overhead?
- **Предлагаемый путь:** Core endpoint обязателен; collector/Grafana optional; technical timings без frames/private labels; resource limits.
- **Условие закрытия:** Metrics/alert catalog, scrape tests и profile load report.
- **Затронутые требования:** OBS-003, OBS-006, OBS-008, OBS-009, OBS-010, OBS-011, OBS-012, OBS-013, OBS-014, OBS-015, OBS-016, OBS-017, OBS-018, OBS-019, OBS-020, OBS-021, OBS-022, OBS-023, OBS-024, OBS-025, OBS-026, OBS-027, OBS-028, OBS-029, OBS-030, OBS-031, OBS-032, OBS-033, OBS-034, OBS-035.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Q-035. Coverage, acceptance evidence и test scope

- **Источник:** §45–46,50–51 ТЗ.
- **Приоритет / срок:** P1; Д001 постановка, методика до Д005/Д015.
- **Что выяснить:** Что входит в core business logic/overall across Go/Python/TS, как объединять coverage, кто принимает и где хранить evidence?
- **Предлагаемый путь:** Раздельные reports и прозрачный weighted line metric без vendor/generated; каждый пункт §50 имеет PASS/FAIL/BLOCKED и artifact ID.
- **Условие закрытия:** Согласованный scope/denominator, test plan и acceptance template; кодовые tests ещё не выполнены.
- **Затронутые требования:** TEST-001, TEST-002, TEST-003, TEST-006.
- **Прогресс Д004–Д010:** 17 Python и3 Go meaningful spike tests проходят; randomized limiter1000 cases, два replay consistent. Это не измерение80/70% coverage ещё не реализованного Core.
- **Статус:** открыт; решение и доказательства ещё не получены.

## Историческая постановка Дня 2

Собрать реальные размеры монитора/стола, желаемое число LED, список уже имеющегося железа и доступность test laptops. Подготовить BOM и wiring draft по Q-002/Q-003/Q-004/Q-016/Q-020. Проверить конфликты пинов/шин, мощность и условия подключения до закупки/монтажа. Если исходных данных нет, сделать обозначенные варианты расчёта, а не выдавать предположения за фактическую конфигурацию.

Обращаться к актуальным первичным datasheets/documentation при выборе компонентов и инструментов. В День 1 такие внешние проверки ещё не выполнялись; наличие кандидата в плане не является подтверждением совместимости.

## Ближайшие действия после текущего выполнения

Независимые scaffold/config/storage/auth/simulation задачи Д011 доступны. Для закрытия M0 нужны фактические Pi4/headless/cooling/access, короткий LED/audio/PCM HIL и LAN/mobile trust. Документы и executable probes готовы; отсутствующий hardware нельзя заменить новыми PASS-SIM. См. сводный execution report и M0 gate matrix.
