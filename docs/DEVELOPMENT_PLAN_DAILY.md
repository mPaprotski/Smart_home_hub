# Подробный ежедневный план разработки Smart Home Hub

**Hardware baseline от26.10.2026: только Raspberry Pi 4B с2GB RAM.** [Ресурсный профиль](development/pi4b-2gb.md). Pi совместные memory/performance gates ещё не выполнены.

Основание: [ТЗ Smart Home Hub, версия 1.0](../Smart_Home_Hub_TZ.md). Подготовлено 25 октября 2026 года. В рабочей области обнаружено только ТЗ; наличие готового кода и настроенного оборудования не предполагается.

## 1. Срок, загрузка и правила использования

**План: 160 основных рабочих дней + 20 дней резерва = 180 рабочих дней, 36 пятидневных недель, примерно 8–9 календарных месяцев без отпусков и праздников.** Расчёт — на одного разработчика, выполняющего Go backend, React UI, Python/ML, интеграцию оборудования, тестирование и эксплуатационные работы. Это оценка, подлежащая пересмотру после аппаратных и ML-экспериментов.

Рабочий день — до 8 часов: около 6 часов сосредоточенной работы и 2 часов на настройку, изучение документации, разбор ошибок, ревью и фиксацию результатов. База составляет около 960 часов сосредоточенной работы, резерв — ещё 120. При доступных 3 часах в день базовый объём займёт примерно 320 рабочих дней, резерв — ещё 40; переключение контекста может увеличить срок.

Д001–Д180 — рабочие дни, а не календарные даты. Дата старта и график занятости не заданы, поэтому даты не выдумываются. Обучение привычкам и длительные испытания измеряются фактическим календарным временем.

План охватывает всю первую версию из §50, включая сложные функции, которые в небольшом MVP обычно выносят в отдельные релизы. Виртуальная матрица обязательна. Физическая LED-панель, Gaming Mode, Grafana и демонстрация с Home Assistant — необязательные задачи резерва. Исключения §52 и развитие §53 не входят в обязательный объём.

### Ежедневный цикл

1. 15–20 минут: проверить зависимости, стенд и блокеры; определить демонстрируемый результат.
2. Около 1 часа: уточнить контракт, подготовить пример входа и критерий проверки.
3. Около 3–4 часов: реализовать задачи дня небольшими изменениями.
4. Около 1–1,5 часа: выполнить указанные проверки и исправить ошибки.
5. Оставшееся время: обновить документацию, сохранить изменения, записать фактические затраты и следующий шаг.

Для аппаратных экспериментов и приёмки распределение часов меняется. Внутри дня задачи выполняются в указанном порядке; проверки входят в оценку дня, а не добавляются сверх неё.

### Условие завершения каждого дня

- Результат воспроизводится из чистого checkout с тестовыми конфигурациями.
- Изменившиеся контракты и инструкции обновлены; секреты и пользовательские данные отсутствуют в репозитории.
- Новая бизнес-логика проверена содержательными тестами; аппаратная работа подтверждена протоколом. Для простых настроек и документации достаточно проверки результата.
- Для аппаратных функций сохранён симуляционный путь, где применимо.
- В журнале указаны результат, ограничения, затраченное время и блокеры. Незавершённый критерий переносится явно.

## 2. Входные условия и допущения

- Разработчик знаком с Go, TypeScript/React, Python, Linux и основами электроники. Обучение стеку с нуля требует отдельного резерва.
- Pi 4B/2GB доступна к Д006, LED-стенд — к Д027, аудиооборудование — к Д036, компьютеры четырёх платформ — к Д061–Д068, камера — к Д081, датчики — к Д091.
- Для screen-client нужны реальные Linux x86_64, Windows x86_64, macOS arm64 и macOS x86_64. Успешной cross-compilation недостаточно для подтверждения capture и ML.
- Поставка компонентов идёт параллельно разработке. Ожидание доставки не включено в рабочие дни; при задержках выполняются независимые симуляционные задачи, аппаратный рубеж переносится.
- Проверяются USB webcam и Pi Camera отдельно. Конкретные модели датчиков выбираются в начале; поддержка всех существующих моделей не предполагается.
- Для Spotify Connect требуется доступный разрешённый сервисом сценарий и подходящая учётная запись. Совместимость версий ОС, библиотек и условия сервиса проверяются во время реализации; документ не утверждает их актуальное состояние.
- Интернет нужен для зависимостей, CI/CD, Telegram и Spotify. Локальный свет, локальное аудио, сенсоры, автоматизация и локальный ML должны работать без WAN.
- Создание GitHub-репозитория, публикация релизов, покупки и установка стенда — будущие действия плана, а не уже выполненные операции.

## 3. Вопросы ТЗ и предлагаемые решения

| Вопрос | Рабочее решение плана | День решения |
|---|---|---|
| Нет точного BOM, длин лент и pinout двух зон | Рассчитать реальные нагрузки, подтвердить независимые WS2812B-каналы и исключить конфликт LED/audio/GPIO | Д002–Д008 |
| Не выбран audio graph, DSP нужен общий PCM | Проверить единый host audio graph и ограниченные интерфейсы управления из контейнера; не рассчитывать на полный privileged контейнер | Д004, Д036–Д037 |
| Scene ML внутри Go-клиента, основной ML на Python | Выбрать локальный runtime: embedded inference либо локальный процесс; кадры остаются на ноутбуке | Д005, Д071 |
| Запрет записи кадров и необходимость CV/ML-фикстур | Использовать синтетические либо лицензированные тестовые материалы; runtime-кадры камеры/экрана не записывать | Д005, Д072, Д077 |
| Gesture Mode и presence используют одну камеру | Один capture owner, независимые запросы gesture/presence, общий CV-индикатор; после restart камера выключена | Д003, Д081–Д082 |
| Confirmation 300–500 мс и реакция <500 мс | Измерять от начала устойчивого жеста до действия; целиться в нижний confirmation threshold, учесть всю обработку | Д005, Д087–Д090 |
| `/ready` проверяет ML/MQTT, но их отказ допускается | Разделить critical readiness и degraded status; deploy отдельно проверяет сервисы, обязательные для профиля | Д003, Д013, Д141–Д142 |
| Защищённое управление при SQLite Safe Mode | Определить аварийный локальный доступ и поведение сессий без БД; не открывать незащищённое LAN API | Д003, Д128 |
| Manual override из разных интерфейсов | Единый command dispatcher с actor/source/scope; Master OFF блокирует все зоны на 15 минут по умолчанию | Д017, Д032, Д108 |
| Равные priority, restart delayed jobs не определены | Детерминированный tie-break, idempotency, cancellation/revalidation и явная persistence policy | Д102–Д106 |
| Ambient light работает в необычном направлении | Выполнить зависимость ТЗ: больше lux → меньше LED brightness; проверить пользовательское ожидание | Д095 |
| Нет численных ML quality gates и полной thermal policy | До обучения определить holdout/метрики/пороги; thermal states и fan speeds разделить, добавить гистерезис | Д005–Д007, Д072, Д087 |
| Habit targets включают volume, наблюдения — нет | Дополнить минимальную локальную схему volume и контекстом presence/ambient, не записывать треки | Д003, Д021 |
| Нужно минимум 30 дней обучения | Запустить реальный collector на Д035; готовность проверять фактическим временем и качеством данных | Д035, Д115–Д120 |

Это предложения для разработки, а не утверждённые изменения ТЗ. Решения фиксируются в ADR/журнале; если эксперимент не подтверждает путь, зависимый этап переоценивается.

## 4. Зависимости и контрольные рубежи

Критическая цепочка: аппаратные эксперименты → защищённое ядро → безопасный свет → общий PCM/DSP → screen capture/Ambilight → Scene ML → CV/сенсоры/presence → автоматизация → привычки → надёжность → deploy/rollback → приёмка.

CI, UI, метрики и тесты начинаются с ядра и развиваются внутри каждого этапа. Полное CI/CD завершается позже, но проверка качества не откладывается до конца.

| Рубеж | День | Проверяемый результат |
|---|---:|---|
| M0 | Д010 | Подтверждены аппаратные подходы, BOM и критические технические решения |
| M1 | Д025 | Защищённый Hub/UI и simulation без Pi, первый CI |
| M2 | Д035 | Две LED-зоны, limiter, presets, кнопка, override; старт habit collection |
| M3 | Д055 | Аудиоисточники/выходы, DSP и пять музыкальных эффектов |
| M4 | Д070 | Четыре платформы клиента, Ambilight с целевыми FPS/latency |
| M5 | Д080 | Локальный Scene ML с проверенным влиянием и деградацией |
| M6 | Д100 | Жесты, обе камеры, датчики и presence fusion |
| M7 | Д110 | Rule Builder, schedules, delays, priority и override |
| M8 | Д120 | Habit Predictor, feedback, retention/delete; реальные 30 дней при наличии данных |
| M9 | Д130 | MQTT/Telegram, отказоустойчивость, backup/restore/reset |
| M10 | Д145 | Tag-based deployment, readiness и rollback |
| M11 | Д160 | 62 пункта приёмки, документация, v1.0.0 |

Рубеж не закрывается по календарю автоматически. Непройденный критерий блокирует зависимые задачи и требует использования резерва либо пересмотра оценки.

**Фактическое выполнение Д004–Д010:** [локальные результаты и аппаратные blockers](DAYS_004_010_RESULT.md). Запуски host/simulation сохранены, но M0_NOT_READY из-за отсутствия доступной Pi. Плановые ожидания ниже не являются отметками PASS.

## 5. Основной ежедневный план

У каждой недели указаны входные зависимости и разделы ТЗ. Внутри недели дни последовательны. Задачи могут переставляться только при сохранении зависимостей и аппаратных проверок.

### Неделя 01. Требования и технические риски — Д001–Д005

Связь с ТЗ: §1–17, 26–30, 46, 50. Вход: Исходное ТЗ.

#### Д001. Декомпозиция ТЗ

- Присвоить требованиям идентификаторы.
- Связать 62 пункта приёмки с будущими проверками.
- Разделить обязательное, optional и future.

**Проверка дня:** Каждый пункт §50 имеет владельца этапа и способ проверки; исключения §52 не включены в обязательный объём.

**Результат:** Матрица требований, backlog, журнал вопросов.

#### Д002. BOM и электрическая схема

- Уточнить длины лент, число LED и размеры монитора.
- Рассчитать PSU, проводку, предохранители и injection.
- Выбрать кандидатов камер/датчиков и исключить конфликты адресов/пинов.

**Проверка дня:** BOM покрывает обе зоны, охлаждение, CV-индикатор, кнопку и все датчики; указан срок доступности.

**Результат:** BOM, схема питания, карта GPIO/I²C/SPI.

**Фактический прогресс Д002:** [проектный вариант подготовлен](DAY_002_RESULT.md); пользователь выбрал варианты без замеров. Финальные количества/номиналы и HIL остаются открытыми. Установлено документальное ограничение PWM LED / analog audio; SPI-chain — кандидат для Д003–Д008, не принятое окончательное решение.

#### Д003. Доменные контракты и решения

- Описать устройства, зоны, команды, события и пользователей.
- Согласовать readiness/Safe Mode/override/camera lifecycle.
- Определить ownership данных и недостающие поля привычек.

**Проверка дня:** Все источники управления проходят общий command layer; privacy и неоднозначности отражены в решениях.

**Результат:** ADR, словарь состояний, схема данных v0.

**Фактический прогресс Д003:** [контракты v0 и 7 ADR подготовлены](DAY_003_RESULT.md); описаны 20 конфликтных/аварийных сценариев. Физическая LED topology, версии stack и HIL не подтверждены; продуктовые проверки впереди.

**Подробный план Д004–Д010:** [шаги, бюджеты времени, эксперименты, проверки и gates](plans/DAYS_004_010_DETAILED.md). Это план будущих работ; статусы выполнения дней не изменены.

#### Д004. Spike LED и аудио

- Проверить кандидатов независимых LED-каналов и конфликты с audio.
- Исследовать A2DP sink, output routing и общий PCM tap.
- Выделить host/container boundaries и минимальные права.

**Проверка дня:** Есть воспроизводимый путь для двух зон и аудиографа; неподтверждённые места отмечены как блокеры Pi spike.

**Результат:** Отчёт эксперимента, список выбранных адаптеров.

**Фактическое исполнение:** [отчёт Д004](DAY_004_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

#### Д005. Spike screen/ML и измерения

- Проверить capture/inference подходы для четырёх платформ.
- Разделить 30 FPS colors и медленный ML.
- Определить quality gates, dataset policy и точки измерения latency.

**Проверка дня:** У каждой цели §46 есть методика; распознавание не требует облака или передачи кадров Hub.

**Результат:** Протокол v0, план измерений, реестр рисков.

**Фактическое исполнение:** [отчёт Д005](DAY_005_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

### Неделя 02. Подготовка Pi и подтверждение решений — Д006–Д010

Связь с ТЗ: §3–6, 29, 32–34, 42. Вход: Результаты Д001–Д005; Pi доступна.

#### Д006. Headless установка Pi

- Установить выбранную Ubuntu Server и SSH.
- Проверить питание, архитектуру, диск и LAN.
- Снять базовые CPU/RAM/temperature показатели.

**Проверка дня:** Pi доступна headless; thermal sysfs читается; повторная установка описана без секретов.

**Результат:** Install guide, инвентаризация стенда.

**Фактическое исполнение:** [отчёт Д006](DAY_006_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

#### Д007. Охлаждение и hardware access

- Проверить GPIO/I²C/SPI и управление вентилятором через драйвер.
- Выполнить ограниченный нагрузочный тест.
- Задать гистерезис и безопасное поведение при потере температуры.

**Проверка дня:** Fan управляется; нет дребезга порогов; необходимые permissions известны.

**Результат:** Thermal spike report, карта доступа к устройствам.

**Фактическое исполнение:** [отчёт Д007](DAY_007_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

#### Д008. LED/audio spike на целевой Pi

- Проверить короткие тестовые ленты двух каналов с ограничением питания.
- Повторить A2DP/PCM/output эксперименты на Pi.
- Проверить LED и audio одновременно.

**Проверка дня:** Две зоны независимы; аудио не ломает LED; критические конфликты устранены или требуют переоценки.

**Результат:** Подтверждённые драйверы и host audio setup.

**Фактическое исполнение:** [отчёт Д008](DAY_008_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

#### Д009. HTTPS и mDNS

- Выбрать reverse proxy и локальный CA.
- Проверить smarthome.local с телефона/ноутбука после установки доверия.
- Описать выпуск/отзыв сертификатов и LAN boundaries.

**Проверка дня:** Локальное имя разрешается; HTTPS доверен тестовым клиентом; ключи исключены из git.

**Результат:** TLS/mDNS bootstrap guide, config templates.

**Фактическое исполнение:** [отчёт Д009](DAY_009_RESULT.md); выполнена доступная локальная/документальная часть; target/HIL ограничения указаны отдельно.

#### Д010. Ревизия технической базы M0

- Повторить критические эксперименты из чистого состояния.
- Зафиксировать проверенные версии, BOM и ограничения.
- Пересчитать план при неподтверждённых capture/audio/LED решениях.

**Проверка дня:** M0 закрыт только с подтверждённым аппаратным путём; каждый риск имеет ответственного этапа и запасной вариант.

**Результат:** Отчёт M0, актуальный backlog и ADR.

**Фактическое исполнение:** [отчёт Д010](DAY_010_RESULT.md); M0 review выполнен, hardware gate NOT_READY.

### Неделя 03. Monorepo и Go Core — Д011–Д015

Связь с ТЗ: §7–8, 28, 32–34, 38, 41–45. Вход: M0.

**Подробный план Д011–Д015:** [рабочие блоки, контракты, проверки и software gates](plans/DAYS_011_015_DETAILED.md). Программный срез реализован и проверен: [результат Д011–Д015](DAYS_011_015_RESULT.md). Hardware M0 открыт; remote CI ещё не запускался.

#### Д011. Каркас monorepo

- Создать Go/Python/frontend/deploy/docs структуру.
- Добавить example config/env, gitignore и версии toolchains.
- Сделать команды запуска и форматирования.

**Проверка дня:** Чистый checkout не содержит secrets и проходит базовые команды без Pi.

**Результат:** Repository skeleton, quickstart.

#### Д012. Конфигурация и lifecycle

- Загрузить env/YAML/runtime settings с явным приоритетом.
- Валидировать диапазоны и несовместимые параметры.
- Добавить context cancellation и JSON logs с redaction.

**Проверка дня:** Неверный config объяснимо отклоняется; shutdown не оставляет фоновых задач; секреты не логируются.

**Результат:** Config package, lifecycle tests.

#### Д013. HTTP и readiness

- Поднять /api/v1 и единый error envelope.
- Реализовать /health, /ready и critical/degraded dependencies.
- Начать OpenAPI, request IDs и HTTP metrics.

**Проверка дня:** Health различает процесс, готовое ядро и необязательный отказ; API-контракт проверяем.

**Результат:** HTTP server, OpenAPI v0, endpoint tests.

#### Д014. Event Bus

- Определить envelope с version/time/event/correlation IDs.
- Реализовать bounded subscriptions и cancellation.
- Проверить overflow, медленного consumer и concurrent stop.

**Проверка дня:** Медленный subscriber не блокирует realtime pipeline; race tests и benchmark проходят.

**Результат:** Event bus, unit tests, benchmark baseline.

#### Д015. CI v0 и fake adapters

- Добавить lint/test/build имеющихся стеков.
- Ввести hardware interfaces и fake implementation.
- Поднять dev Compose и smoke без устройств Pi.

**Проверка дня:** CI работает без production secrets/Pi; fake sensor событие достигает Core.

**Результат:** CI workflow v0, Compose dev, simulation skeleton.

### Неделя 04. SQLite, auth и команды — Д016–Д020

**Подробный план Д016–Д018:** [рабочие блоки, проверки и условия завершения](plans/DAYS_016_018_DETAILED.md). Реализованный software slice и точные границы: [результат Д016–Д018](DAYS_016_018_RESULT.md).

Связь с ТЗ: §18, 26–28, 31, 33, 39, 49. Вход: Каркас Д011–Д015.

#### Д016. SQLite WAL и migrations

- Создать users/roles/sessions/zones/devices/settings schema.
- Настроить WAL, busy timeout и короткие transactions.
- Проверить пустую/старую схему, busy и неуспешную migration.

**Проверка дня:** Данные переживают restart; failed migration не оставляет полусхему.

**Результат:** Storage layer, migrations, integration tests.

#### Д017. Command dispatcher и permissions

- Ввести actor/source/scopes/affected resources.
- Реализовать User/Role ownership без multi-user UI.
- Заложить единый arbitration hook для manual/automation/ML.

**Проверка дня:** UI/бот/ML не могут обходить права и business validation; неизвестная capability отклоняется.

**Результат:** Command model, permission matrix, handler tests.

#### Д018. Admin bootstrap и Argon2id

- Сделать одноразовый локальный bootstrap без default password.
- Добавить password hashing/verification.
- Ограничить login attempts и унифицировать ошибки.

**Проверка дня:** В БД/логах нет пароля; повторный bootstrap запрещён; неверный вход rate-limited.

**Результат:** Bootstrap CLI, login API, tests.

**Подробный план Д019–Д021:** [рабочие блоки, auth boundaries, schema/privacy и проверки](plans/DAYS_019_021_DETAILED.md). Software slice реализован и локально проверен: [результат Д019–Д021](DAYS_019_021_RESULT.md). Аппаратные и следующие UI/ML gates остаются открытыми.

#### Д019. Sessions и CSRF

- Добавить expiration/logout/revoke-all.
- Настроить HttpOnly/Secure/SameSite, origin/CSRF policy.
- Защитить WebSocket handshake и протухшие сессии.

**Проверка дня:** Отозванная session не действует REST/WS; cookies корректны через reverse proxy HTTPS.

**Результат:** Auth middleware, session integration tests.

#### Д020. Machine tokens и security audit

- Реализовать scopes и create/list/revoke с показом секрета один раз.
- Хранить hash и metadata.
- Добавить rate limits и audit login/token/settings.

**Проверка дня:** Revoked token и недостаточный scope дают отказ; audit не содержит secret.

**Результат:** Token API, security events, OpenAPI auth.

### Неделя 05. UI и полноценная simulation — Д021–Д025

Связь с ТЗ: §17, 24, 28–29, 31, 34, 38. Вход: Auth Д016–Д020.

#### Д021. Схема остальных данных

- Спроектировать presets/rules/logs/gesture mappings/observations/feedback.
- Добавить volume/presence/ambient context, timestamps и индексы.
- Задать data minimization и retention boundaries.

**Проверка дня:** Схема покрывает habit targets без track history, кадров и DSP history.

**Результат:** Domain migrations, data dictionary.

#### Д022. React shell и авторизация

- Поднять React/TypeScript/Vite с responsive navigation.
- Реализовать login/logout и типизированный API client.
- Добавить loading/error/401 и keyboard access.

**Проверка дня:** На телефоне доступен защищённый Dashboard по HTTPS; logout закрывает доступ.

**Результат:** Frontend shell, auth flow, UI checks.

#### Д023. WebSocket и Dashboard

- Подключить /ws/events с auth/filtering/reconnect.
- Показать CPU/RAM/temperature и статусы модулей.
- Ограничить очереди и частоту realtime сообщений.

**Проверка дня:** Reconnect не дублирует subscribers; медленный браузер не тормозит Core.

**Результат:** WS client/server, Dashboard, reconnect tests.

#### Д024. Fake-источники и virtual LEDs

- Добавить fake audio/sensors/presence/screen/MQTT.
- Ввести seed, fake clock и fault сценарии.
- Поднять virtual strip/matrix preview.

**Проверка дня:** Один сценарий повторяется детерминированно; simulation не импортирует реальные GPIO drivers.

**Результат:** Simulation suite, fixtures, visualizer v0.

#### Д025. Контрольный срез M1

- Пройти clean start/auth/WS/storage/shutdown без Pi.
- Прогнать lint/typecheck/tests/CI.
- Обновить quickstart и HTTPS instructions.

**Проверка дня:** M1 воспроизводим на обычном компьютере; неавторизованный доступ закрыт.

**Результат:** Протокол M1, dev build, OpenAPI update.

### Неделя 06. Lighting Engine и Power Limiter — Д026–Д030

Связь с ТЗ: §4–6, 14, 20–21, 42, 45. Вход: M1; driver решения Д008.

#### Д026. Zone/device models и mapping

- Реализовать STRIP/MATRIX и capabilities/status.
- Поддержать произвольные зоны и geometry.
- Проверить serpentine/orientation/corner/direction.

**Проверка дня:** Бизнес-логика не зашита на две зоны; mapping покрыт table/golden tests.

**Результат:** Zone/device models, mapping library.

#### Д027. Реальные LED adapters

- Подключить выбранный WS2812B driver и lifecycle.
- Разделить logical zones/physical channels.
- Ограничить контейнер только нужными device permissions.

**Проверка дня:** Обе зоны адресуются независимо на коротком стенде и через fake driver.

**Результат:** LED adapter, HIL log, permissions guide.

#### Д028. Zone/global Power Limiter

- Оценивать ток с калибруемыми коэффициентами и idle current.
- Применять zone и общий PSU limits к кадру всех зон.
- Проверить zero/invalid limits и extremes.

**Проверка дня:** Ни эффект, ни fade не обходят limiter; итоговая оценка укладывается в оба ограничения.

**Результат:** Limiter, numeric/property tests, calibration model.

#### Д029. Render loop и цветовой pipeline

- Реализовать monotonic timing, bounded queue и latest-frame policy.
- Добавить RGB/HSV/temperature equivalent/master modifier.
- Выполнять fade перед финальным limiter.

**Проверка дня:** LED >=30 FPS на согласованной нагрузке; нет гонок и всплесков тока при сменах.

**Результат:** Render pipeline, benchmark, golden fixtures.

#### Д030. Базовые эффекты и Night Shift

- Реализовать OFF/STATIC/BREATHING/RAINBOW/NIGHT_SHIFT.
- Сохранить transition duration и parameters.
- Проверить warm equivalent, 10–20% brightness и отсутствие анимации Night Shift.

**Проверка дня:** Все эффекты корректны на STRIP/MATRIX разной длины; Night Shift плавно включается.

**Результат:** Effect implementations, tests, color pipeline doc.

### Неделя 07. Presets, кнопка и две зоны — Д031–Д035

Связь с ТЗ: §4–6, 17, 19–20, 24, 49. Вход: Lighting Д026–Д030.

#### Д031. Lighting UI и multi-zone presets

- Добавить per-zone mode/color/brightness и master brightness.
- Сделать preset CRUD/editor/preview.
- Показать current estimate, applied limit и ошибки.

**Проверка дня:** Preset воспроизводимо меняет обе зоны; UI объясняет ограниченную brightness.

**Результат:** Lighting page, preset editor, API tests.

#### Д032. Manual override и Master OFF

- Ввести affected-state scope, TTL и cancel.
- Создавать override в общем dispatcher для всех manual sources.
- Master OFF выключает и блокирует все зоны.

**Проверка дня:** Automation не включает свет 15 минут после Master OFF; длительность настраивается.

**Результат:** Override service, conflict tests, UI indicator.

#### Д033. Физическая кнопка

- Реализовать debounce и single/double/long state machine.
- Хранить mappings в SQLite и редактировать UI.
- Проверить конкурирующие сигналы и ложный single при double.

**Проверка дня:** Все три press types работают на кнопке и в simulation без двойных действий.

**Результат:** Button driver, remapping UI, HIL protocol.

#### Д034. Монтаж и калибровка зон

- Собрать monitor.backlight/desk.wall по электрической схеме.
- Проверить polarity/ground/injection/нагрев и фактический ток.
- Сверить corners/direction и brightness limits.

**Проверка дня:** Реальная установка стабильна; physical mapping и расчётная модель тока проверены измерением.

**Результат:** As-built wiring, calibration config, HIL report.

#### Д035. M2 и начало habit collection

- Пройти presets/button/override/restore LED state.
- Включить локальный USER_ACTION collector и state snapshots.
- Записать learning_started_at, доступные поля и gaps.

**Проверка дня:** M2 закрыт; события реально сохраняются без треков/кадров; стартовал календарный learning period.

**Результат:** Протокол M2, collector v0, журнал начала обучения.

### Неделя 08. Audio graph и Bluetooth input — Д036–Д040

Связь с ТЗ: §9, 32, 39, 42, 47. Вход: M2; audio spike Д004/Д008.

#### Д036. Production audio graph

- Зафиксировать host daemon/BlueZ и ограниченный IPC control.
- Разделить playback/capture/metadata/transport.
- Автоматизировать host setup и startup.

**Проверка дня:** Аудиограф поднимается после reboot; полный privileged Core не требуется.

**Результат:** Audio architecture, setup script, adapter contracts.

#### Д037. Общий PCM tap

- Снять monitor stream выбранного playback bus.
- Нормализовать sample rate/channels и bounded buffers.
- Проверить silence/overrun/source switching без записи звука.

**Проверка дня:** Все источники могут использовать один формат DSP input; tap не нарушает playback.

**Результат:** PCM bridge, fixture tests, latency baseline.

#### Д038. Bluetooth PAIRING_MODE

- Показать имя Smart Home Hub.
- Ограничить окно новых pairing явной активацией и TTL.
- Добавить trusted reconnect и security events.

**Проверка дня:** Новый телефон вне окна не сопрягается; доверенный reconnect соответствует принятой политике.

**Результат:** Bluetooth adapter, pairing API/UI, HIL log.

#### Д039. Bluetooth transport/metadata

- Подключить доступные Play/Pause/Next/Previous и track/progress.
- Отразить unsupported capabilities.
- Обработать disconnect/reconnect и missing metadata.

**Проверка дня:** UI показывает доступные команды; отсутствие данных не ломает playback.

**Результат:** A2DP control adapter, metadata contract, tests.

#### Д040. Совместный Bluetooth/LED test

- Проиграть телефон на двух работающих LED-зонах.
- Проверить pause/resume/volume/reconnect и PCM tap.
- Измерить CPU/RAM и аппаратные конфликты.

**Проверка дня:** Реальный сигнал поступает DSP; LED независимы; reconnect не требует shell вмешательства.

**Результат:** Audio HIL report, device compatibility notes.

### Неделя 09. Источники, выходы и Audio UI — Д041–Д045

Связь с ТЗ: §9, 24, 39–40. Вход: Audio graph Д036–Д040.

#### Д041. Локальные аудиофайлы

- Реализовать библиотеку/playlist и playback.
- Валидировать paths, formats и missing files.
- Подключить controls/metadata/PCM без истории треков.

**Проверка дня:** Play/Pause/Next/Previous работают; path traversal закрыт; файл не пишет listening history.

**Результат:** Local audio adapter, fixtures, integration tests.

#### Д042. Spotify Connect

- Проверить совместимый клиент и условия тестовой учётной записи.
- Подключить source lifecycle/audio graph/metadata.
- Изолировать credentials и обработать service outage.

**Проверка дня:** Spotify работает при доступном сервисе; его отказ не влияет на local/Bluetooth.

**Результат:** Spotify adapter, configs, prerequisite guide.

#### Д043. Выбор audio output

- Добавить discovery/selection 3.5 mm, USB DAC и Bluetooth output.
- Проверить Bluetooth input/output concurrency и реальную маршрутизацию.
- Обработать исчезновение output и безопасный volume fallback.

**Проверка дня:** Каждый обязательный выход проверен на Pi; смена не создаёт неконтролируемую громкость.

**Результат:** Output manager, HIL matrix, preference storage.

#### Д044. Завершение Audio UI

- Добавить source/output selector и pairing controls.
- Показать track/artist/album/cover/progress по capabilities.
- Согласовать mute/volume/reconnect и unsupported states.

**Проверка дня:** С телефона доступны все обязательные команды; состояния соответствуют API.

**Результат:** Audio page, typed client, interaction checks.

#### Д045. Audio acceptance

- Пройти source × output matrix доступных комбинаций.
- Проверить restart STOPPED, disconnect, mute и WAN outage.
- Обновить setup по найденным ошибкам.

**Проверка дня:** Обязательные источники/выходы подтверждены; local/Bluetooth работают без WAN.

**Результат:** Audio acceptance protocol, smoke scenario.

### Неделя 10. Python Runtime и DSP — Д046–Д050

Связь с ТЗ: §7, 10, 36, 38, 45–46. Вход: PCM Д037; audio Д041–Д045.

#### Д046. ML Runtime lifecycle

- Создать Python service API/health и versioned features.
- Настроить timeout/backoff/reconnect.
- Разделить DSP/gesture/presence workers и failure boundaries.

**Проверка дня:** Падение Python не останавливает playback и статический свет.

**Результат:** Runtime skeleton, contracts, failure tests.

#### Д047. FFT и частотные полосы

- Добавить PCM windows/window function/FFT.
- Вычислить overall/bass/mids/treble с документированными частотами.
- Проверить синусоиды, silence и sample rates.

**Проверка дня:** Golden/numeric tests подтверждают распределение энергии; нет NaN и overflow.

**Результат:** FFT pipeline, deterministic audio fixtures.

#### Д048. Normalization и smoothing

- Добавить auto normalization и manual sensitivity.
- Настроить noise floor, clamp и adaptation time.
- Разделить smoothing features и LED rendering.

**Проверка дня:** Тишина не вызывает мерцание; разные уровни записи дают управляемый сигнал.

**Результат:** DSP settings, numerical tests, comparison report.

#### Д049. Beat detection и BPM

- Реализовать beat/onset и BPM estimation.
- Проверить metronome, смену темпа и неритмичное аудио.
- Добавить confidence/unknown вместо фиктивного BPM.

**Проверка дня:** Измерены beat precision/recall и BPM error по критериям Д005.

**Результат:** Beat/BPM module, quality report.

#### Д050. DSP на реальной Pi

- Измерить FFT/feature latency и resource usage.
- Проверить PCM loss, worker restart и backpressure.
- Проследить отсутствие audio/FFT записи на диск.

**Проверка дня:** Pipeline стабилен и восстанавливается самостоятельно; playback остаётся доступным.

**Результат:** DSP benchmark, recovery tests, runtime tuning.

### Неделя 11. Музыкальные эффекты и spectrum UI — Д051–Д055

Связь с ТЗ: §10, 24, 34, 45–46. Вход: Lighting M2; DSP Д046–Д050.

#### Д051. MUSIC_PULSE и MUSIC_BEAT

- Связать overall volume с brightness, beat с decay pulse.
- Использовать общий fade/brightness/limiter.
- Проверить silence, пики и missing DSP.

**Проверка дня:** Оба режима работают на desk.wall и сохраняют power limits.

**Результат:** Effects, golden tests, physical demo.

#### Д052. MUSIC_SPECTRUM

- Сопоставить bass/mids/treble участкам и цветам.
- Обработать короткую STRIP и MATRIX.
- Подключить sensitivity/smoothing.

**Проверка дня:** Тестовые частоты активируют ожидаемые сегменты без выхода за границы.

**Результат:** Spectrum effect, mapping tests.

#### Д053. MUSIC_WAVE и MUSIC_CENTER

- Двигать импульсы по времени, а не кадрам.
- Добавить симметрию чётных/нечётных длин.
- Ограничить число активных импульсов.

**Проверка дня:** Скорость стабильна при смене FPS; эффекты помещаются в render budget.

**Результат:** Wave/center effects, stress/golden tests.

#### Д054. Realtime spectrum UI

- Передавать spectrum/BPM/beat с bounded WS rate.
- Сделать sensitivity/smoothing/effect controls.
- Добавить no signal/reconnect и проверить browser memory.

**Проверка дня:** Мобильный UI не замедляет render; DSP history не накапливается.

**Результат:** Spectrum visualization, WS contract, load check.

#### Д055. Контрольный срез M3

- Проверить пять эффектов с разными источниками/output.
- Измерить FPS/CPU/RAM/thermal под нагрузкой.
- Уронить DSP и проверить playback/fallback.

**Проверка дня:** M3 закрыт; аудио независимо от DSP, эффекты и UI согласованы.

**Результат:** Протокол M3, resource baseline, demo.

### Неделя 12. Screen-client и первый Ambilight — Д056–Д060

Связь с ТЗ: §11, 27–28, 34, 40, 45–46. Вход: M3; capture spike Д005.

#### Д056. Pairing screen-client

- Сделать one-time code с TTL/rate limits.
- Обменять на scoped device token и безопасно хранить локально.
- Добавить revoke и смену Hub.

**Проверка дня:** Повторный/просроченный code отклоняется; revoked client теряет доступ.

**Результат:** Pairing API, CLI, security tests.

#### Д057. Протокол потока цветов

- Зафиксировать mapping/version/sequence/time payload.
- Ограничить frame size/rate/queues и выбрать latest frame.
- Обработать stale/out-of-order/reconnect.

**Проверка дня:** Клиент не передаёт кадры экрана; плохой input не расходует неограниченную память.

**Результат:** Streaming protocol, parser/fuzz tests.

#### Д058. Linux capture

- Реализовать monitor selection и разрешения.
- Проверить выбранные Linux sessions, scaling и multiple displays.
- Обработать запрет доступа без зависания.

**Проверка дня:** Реальный Linux захватывает нужный экран; screenshot files не создаются.

**Результат:** Linux adapter, supported-session checklist.

#### Д059. Weighted average и mapping

- Разбить четыре края на LED regions с edge weights.
- Поддержать counts/start_corner/direction.
- Прогнать synthetic color-border fixtures.

**Проверка дня:** Golden colors совпадают с ожидаемыми физическими LED.

**Результат:** Screen analyzer, golden test set.

#### Д060. Первый Ambilight E2E

- Передать Linux colors через Hub в virtual/physical monitor zone.
- Добавить lifecycle и basic fallback.
- Снять capture/network/render baseline.

**Проверка дня:** Мониторная рамка повторяет края; Hub payload содержит только colors/metadata.

**Результат:** First E2E, packet review, latency baseline.

### Неделя 13. Cross-platform capture и image pipeline — Д061–Д065

Связь с ТЗ: §11–12, 40, 43, 45. Вход: Первый E2E; платформенные стенды.

#### Д061. Windows x86_64 capture

- Реализовать platform adapter и permissions/errors.
- Проверить DPI/multi-monitor/disconnect/lock.
- Собрать binary с нужными dependencies.

**Проверка дня:** На реальном Windows корректные цвета выбранного экрана; ограничения protected content описаны.

**Результат:** Windows adapter, smoke protocol.

#### Д062. macOS arm64 capture

- Реализовать screen recording permissions и revoke.
- Проверить Retina/multiple displays.
- Собрать native package с выбранным toolchain.

**Проверка дня:** Capture работает на Apple Silicon и корректно прекращается при отзыве доступа.

**Результат:** macOS arm64 build, permission guide.

#### Д063. macOS x86_64 capture

- Собрать и запустить Intel вариант.
- Проверить native dependencies и analyzer consistency.
- Устранить platform-specific resources/build paths.

**Проверка дня:** Intel Mac проходит pairing/capture/stream; компиляция не заменяет запуск.

**Результат:** macOS x86_64 package, compatibility matrix.

#### Д064. Black bars и коррекция цвета

- Добавить black-bar detection/near-black suppression.
- Реализовать gamma/saturation/boost/min-max brightness с фиксированным порядком.
- Проверить subtitles, fade-to-black и яркие края.

**Проверка дня:** Golden tests покрывают bars/black и не дают нестабильных скачков crop.

**Результат:** Image pipeline, fixtures, correction documentation.

#### Д065. Cinema и inactivity

- Настроить Cinema smoothing/transitions/update rate.
- Детектировать monitor off/static timeout/stream loss.
- Определить fallback/off и возврат к активному source.

**Проверка дня:** Cinema плавен; потеря/неактивность переводят зону в заданное состояние.

**Результат:** Cinema profile, inactivity state machine, tests.

### Неделя 14. Ambilight UI, packaging и performance — Д066–Д070

Связь с ТЗ: §11, 24, 27, 43, 45–46. Вход: Cross-platform Д061–Д065.

#### Д066. Ambilight UI и calibration

- Добавить geometry editor, update rate и Cinema settings.
- Показать только цветовую рамку, client health/reconnect.
- Дать тестовые стороны/углы для проверки wiring.

**Проверка дня:** Mapping настраивается без YAML; screenshot не передаётся в Hub.

**Результат:** Ambilight page, calibration flow.

#### Д067. Latency optimization

- Измерить capture/analyze/network/queue/render.
- Проверить полную e2e задержку визуальным стендом или общей опорой времени.
- Сократить allocations/copies и queues по профилю.

**Проверка дня:** В эталонном LAN получены 30 FPS и <100 мс либо открыт блокер M4; clock skew учтён.

**Результат:** Benchmark method, traces, optimization report.

#### Д068. Пакеты четырёх платформ

- Автоматизировать build/checksums/config.
- Проверить install/token storage/update/uninstall на каждой ОС.
- Документировать permissions и ограничения capture.

**Проверка дня:** Каждый пакет запускается в чистой пользовательской среде своей платформы.

**Результат:** Client packages/scripts, install guide.

#### Д069. Network/fallback/security tests

- Эмулировать jitter/loss/disconnect/duplicate frames.
- Проверить revoke/invalid mapping/rate abuse.
- Нагрузить stream при активном UI и других зонах.

**Проверка дня:** Stream не блокирует commands/render; fallback детерминирован; auth соблюдён.

**Результат:** Failure suite, security/performance report.

#### Д070. Контрольный срез M4

- Проверить реальные стороны/corners/Cinema/inactivity.
- Повторить FPS/e2e для платформ в оговорённой сети.
- Закрыть критические дефекты и зафиксировать условия.

**Проверка дня:** M4 закрыт при подтверждённой производительности и всех четырёх платформах.

**Результат:** Протокол M4, compatibility/performance table.

### Неделя 15. Scene ML: данные и модель — Д071–Д075

Связь с ТЗ: §12, 40, 45. Вход: M4; dataset policy Д005.

#### Д071. Inference integration на ноутбуке

- Выбрать локальный runtime с поддержкой четырёх targets.
- Отделить worker от 30 FPS loop.
- Описать model manifest/checksum и offline loading.

**Проверка дня:** Runtime реально исполняется на всех architectures; кадры остаются локально.

**Результат:** Scene runtime ADR, prototype packages.

#### Д072. Scene evaluation dataset

- Определить rubric для девяти multi-label классов.
- Подготовить лицензированные/синтетические sequences и split по исходным роликам.
- До обучения задать per-class precision/recall/F1 thresholds.

**Проверка дня:** Нет утечки кадров одного ролика между train/test; каждый класс представлен.

**Результат:** Dataset manifest, annotation guide, quality gate.

#### Д073. Multi-frame features и baseline

- Рассчитать brightness/contrast/colorfulness/motion.
- Проверить preprocessing детерминированными fixtures.
- Оценить эвристический baseline на frozen holdout.

**Проверка дня:** Признаки воспроизводимы; baseline даёт точку сравнения для ML.

**Результат:** Feature pipeline, baseline report.

#### Д074. Обучение multi-label classifier

- Обучить простую локальную модель на признаках.
- Настроить per-label thresholds только на validation.
- Сохранить model artifact, параметры и воспроизводимый script.

**Проверка дня:** Модель допускает несколько меток; качество и ограничения сравнены с baseline.

**Результат:** Training pipeline, model artifact, evaluation draft.

#### Д075. Локальный inference worker

- Запускать не чаще одного inference в секунду на temporal buffer.
- Отправлять Hub labels/confidence/metadata.
- Проверить timeout/missing model/invalid result.

**Проверка дня:** ML не блокирует 30 FPS; сбой оставляет обычный Ambilight.

**Результат:** Client inference integration, failure tests.

### Неделя 16. Scene adaptation и privacy — Д076–Д080

Связь с ТЗ: §12, 24, 36, 40, 45–46. Вход: Модель Д071–Д075.

#### Д076. ML influence и adaptation policy

- Сопоставить метки пяти разрешённым параметрам.
- Разрешить конфликты меток и clamp ranges.
- Реализовать influence 0–100% без смены режима.

**Проверка дня:** 0% идентичен обычному Ambilight; комбинации меток безопасны и предсказуемы.

**Результат:** Adaptation policy, table tests.

#### Д077. Scene UI и privacy audit

- Добавить influence/current labels/disable.
- Проверить frame dumps/cache/logs/crash diagnostics.
- Разделить test assets и runtime data и исключить scene history.

**Проверка дня:** На диске и в Hub нет runtime images/scene history; UI показывает только разрешённое.

**Результат:** Scene UI, privacy audit results.

#### Д078. Проверка качества Scene ML

- Посчитать финальные метрики на untouched holdout.
- Проверить dark motion/cartoons/subtitles/transitions.
- Дорабатывать по train/validation без подгонки test.

**Проверка дня:** Quality gate Д072 пройден либо ML остаётся блокером; эвристика не объявлена моделью.

**Результат:** Quality report, documented limitations.

#### Д079. Performance ML на четырёх платформах

- Измерить inference/RAM/CPU вместе с 30 FPS stream.
- Проверить offline startup и bundled dependencies.
- Проверить 1 inference/sec и worker cancellation.

**Проверка дня:** Все пакеты запускаются по инструкции без скрытых зависимостей и утечки кадров.

**Результат:** ML-enabled compatibility matrix, benchmark.

#### Д080. Контрольный срез M5

- Пройти все labels/influence extremes и disable.
- Уронить worker и проверить обычный Ambilight/reconnect.
- Повторить latency со включённым ML.

**Проверка дня:** M5 закрыт; ML не ухудшает обязательный realtime pipeline.

**Результат:** Протокол M5, versioned model package.

### Неделя 17. Камера и статические жесты — Д081–Д085

Связь с ТЗ: §13, 15, 36, 40, 42. Вход: Python Runtime; камеры доступны.

#### Д081. Camera owner и lifecycle

- Создать единственный capture owner для gesture/presence subscribers.
- Добавить независимые privacy toggles и camera OFF after restart.
- Обработать busy/disconnect/open failure.

**Проверка дня:** Два consumer не открывают камеру дважды; ownership и остановка детерминированы.

**Результат:** Camera manager, lifecycle tests.

#### Д082. Аппаратный CV indicator

- Включать LED до capture, выключать после release.
- Проверить worker crash и indicator failure.
- Реализовать принятую safe policy при невозможности показать активность.

**Проверка дня:** Ни один CV consumer не работает со скрытой активностью по согласованной политике.

**Результат:** Indicator adapter, HIL privacy protocol.

#### Д083. USB webcam и Pi Camera

- Проверить оба типа отдельно на целевой Ubuntu.
- Настроить formats/resolution/permissions и reconnect.
- Сопоставить реальные и fake camera inputs.

**Проверка дня:** Каждый тип проходит start/stop/reconnect на Pi; инструкция отражает фактические ограничения.

**Результат:** Camera support matrix, setup guide, HIL report.

#### Д084. Hand detection/landmarks

- Подключить lightweight local runtime.
- Нормализовать координаты/confidence.
- Проверить всю область кадра, расстояние и отсутствие рук.

**Проверка дня:** Landmarks воспроизводимы; inference local, frames только RAM.

**Результат:** Hand pipeline, preprocessing tests, timing baseline.

#### Д085. Статические жесты

- Реализовать OPEN_PALM/FIST/THUMB_UP/THUMB_DOWN/TWO_FINGERS.
- Разделить classifier и resolved action.
- Проверить partial hand/low confidence/negative examples.

**Проверка дня:** Классификация проходит fixtures и не вызывает hardware напрямую.

**Результат:** Static classifiers, labelled fixtures, tests.

### Неделя 18. Динамические жесты и качество — Д086–Д090

Связь с ТЗ: §13–15, 24, 45–46. Вход: Camera/hand pipeline Д081–Д085.

#### Д086. Динамические жесты

- Добавить swipe left/right, palm-face-down с downward movement.
- Реализовать clockwise/counter-clockwise rotation.
- Проверить temporal tracking и зеркальность камеры.

**Проверка дня:** Каждый dynamic gesture имеет positive/negative sequence test; direction однозначен.

**Результат:** Temporal recognizers, sequence fixtures.

#### Д087. Activation/cooldown protection

- Выбрать activation gesture без конфликта с actions.
- Реализовать confirmation 300–500 мс, cooldown и active-window timeout.
- Зафиксировать precision/recall и false actions/min quality thresholds.

**Проверка дня:** Движение вне окна не выполняет commands; все таймеры проверяются fake clock.

**Результат:** Gesture gate state machine, quality protocol.

#### Д088. Mappings UI и routing

- Сохранить editable mappings SQLite.
- Показать detected/confidence/resolved action/privacy state.
- Провести commands через capabilities/auth/manual override.

**Проверка дня:** Все десять жестов ТЗ имеют defaults и переназначение; UI соответствует принятой команде.

**Результат:** Gestures page, API, integration tests.

#### Д089. Gesture evaluation на Pi

- Проверить свет/фон/дистанцию/пользователей по матрице.
- Измерить от начала устойчивого жеста до действия и отдельно inference.
- Оптимизировать resolution/FPS по профилю.

**Проверка дня:** <500 мс и quality gate доказаны для оговорённых условий либо оформлен блокер.

**Результат:** Gesture quality report, latency traces.

#### Д090. Gesture acceptance и recovery

- Пройти mappings/activation/cooldown/disconnect.
- Совместить audio/reactive LED/Ambilight/CV.
- Проверить indicator, camera OFF reboot и отсутствие записанных кадров.

**Проверка дня:** Privacy и realtime соблюдены; recovery не требует общего restart.

**Результат:** Gesture acceptance protocol, regression checks.

### Неделя 19. Сенсоры и ambient brightness — Д091–Д095

Связь с ТЗ: §16, 21, 24, 38, 42. Вход: Карта шин; комплект датчиков.

#### Д091. Sensor Manager

- Ввести units/timestamps/quality/stale/capabilities.
- Реализовать polling/retry/backoff/cancellation.
- Отделить offline от реального нулевого значения.

**Проверка дня:** Неисправный датчик не блокирует остальные; fake и hardware имеют общий контракт.

**Результат:** Manager, interfaces, simulation tests.

#### Д092. Освещённость и PIR

- Подключить выбранные ambient/PIR models.
- Проверить calibration/units/debounce/offline.
- Сопоставить physical events с simulation events.

**Проверка дня:** Lux/motion имеют корректные timestamps и quality flags.

**Результат:** Drivers, calibration notes, HIL log.

#### Д093. Температура, влажность, воздух

- Подключить выбранные модели и адреса шин.
- Проверить прогрев/sample intervals/ranges.
- Указать фактические единицы air quality без подмены индекса измеренным CO₂.

**Проверка дня:** Все пять обязательных величин доступны; failed read не становится ложным значением.

**Результат:** Sensor adapters, HIL report, units dictionary.

#### Д094. Devices UI и inventory

- Показать ONLINE/OFFLINE/ERROR/DISABLED, capabilities и freshness.
- Добавить ручное добавление/disable/adapter settings.
- Подключить sensor events/alerts.

**Проверка дня:** Пользователь различает missing/stale/valid data; API пригоден Automation Engine.

**Результат:** Devices page, device API, realtime tests.

#### Д095. Ambient brightness modifier

- Реализовать больше lux → меньше brightness по ТЗ.
- Добавить clamps/smoothing/fallback.
- Учитывать override и protection музыкальной сцены.

**Проверка дня:** Modifier не перезаписывает preset и не обходит limiter; потеря sensor безопасна.

**Результат:** Ambient policy, table tests, calibration chart.

### Неделя 20. Presence fusion и CV scheduler — Д096–Д100

Связь с ТЗ: §15–16, 24, 36, 40, 45–46. Вход: Sensors и Camera Manager.

#### Д096. Camera person detection

- Использовать общий capture owner и local person-only detector.
- Выполнять inference каждые 2–3 секунды.
- Проверить empty scene/disabled CV/worker failure.

**Проверка дня:** Person yes/no работает на Pi без identity recognition и записи кадров.

**Результат:** Presence CV worker, fixtures, timing report.

#### Д097. Bluetooth/network presence

- Связать trusted phone/network device с local user.
- Настроить interval/expiry/minimal identifiers.
- Проверить sleeping phone/MAC randomization/network failure.

**Проверка дня:** Отсутствие ответа не означает мгновенный AWAY; ограничения источников документированы.

**Результат:** Presence adapters, setup guide, tests.

#### Д098. Fusion state machine

- Объединить camera/PIR/Bluetooth/network с TTL/weights/hysteresis.
- Реализовать PRESENT/PROBABLY_PRESENT/AWAY/UNKNOWN.
- Проверить conflicting/stale signals.

**Проверка дня:** UNKNOWN отличим от AWAY; шум одного источника не вызывает частые переключения.

**Результат:** Fusion engine, table-driven timeline tests.

#### Д099. Presence UI и общий CV budget

- Показать source contributions/freshness без кадров.
- Ограничить конкуренцию gesture/person workers.
- Проверить независимые toggles и indicator.

**Проверка дня:** Два CV режима делят одну камеру и укладываются в согласованный CPU budget.

**Результат:** Presence controls, scheduler, load report.

#### Д100. Контрольный срез M6

- Проверить пять sensors, четыре presence source и gestures.
- Смоделировать stale/offline/reconnect.
- Снять CPU/RAM/thermal при audio+LED+Ambilight+CV.

**Проверка дня:** M6 закрыт на Pi 4B/2GB под совместной нагрузкой, а не только по отдельности.

**Результат:** Протокол M6, resource profile, tuning config.

### Неделя 21. Automation Core — Д101–Д105

Связь с ТЗ: §18, 31, 45, 49. Вход: Command layer; события всех источников.

#### Д101. Rule DSL и validation

- Описать versioned nested AND/OR и multi-actions.
- Валидировать types/depth/capabilities/references.
- Добавить CRUD и enable/disable.

**Проверка дня:** Невалидные и чрезмерно глубокие rules отклоняются до исполнения.

**Результат:** Rule schema, OpenAPI, validation tests.

#### Д102. Evaluator всех conditions

- Подключить time/day/presence/monitor/music/Ambilight/brightness/ambient/gesture/ML/MQTT/device/sensor.
- Задать UNKNOWN/stale semantics.
- Разделить trigger и state conditions.

**Проверка дня:** Каждый condition имеет positive/negative/unknown tests; nested evaluation детерминирован.

**Результат:** Condition engine, complete test matrix.

#### Д103. Регулярные schedules

- Добавить weekday/daily schedules и timezone.
- Определить missed runs/clock jumps/restart semantics.
- Проверить DST в поддерживаемых timezone через fake clock.

**Проверка дня:** Нет duplicate run после restart/перевода часов; одноразовые schedules не добавляются.

**Результат:** Scheduler, time tests, documented semantics.

#### Д104. Multi-actions и priority

- Арбитрировать per affected resource с deterministic tie-break.
- Ввести command idempotency и event loop protection.
- Определить partial failure поведение.

**Проверка дня:** Конфликтующие rules дают одинаковый результат с понятным объяснением.

**Результат:** Action executor, priority/loop tests.

#### Д105. Delayed actions

- Реализовать wait/cancel/revalidation.
- Определить persistence/restart policy.
- Проверить user return, rule delete/disable и дубли job.

**Проверка дня:** Устаревший timer не выключает свет после возвращения; restart не дублирует action.

**Результат:** Delayed scheduler, persistence tests.

### Неделя 22. Rule Builder и arbitration — Д106–Д110

Связь с ТЗ: §18, 24, 49–50. Вход: Automation Core Д101–Д105.

#### Д106. Audit и 30-дневная retention

- Сохранять rule/trigger/time/actions/result/error/correlation.
- Добавить фильтры и cleanup.
- Различать failed/partial/skipped без sensitive payload.

**Проверка дня:** Каждый запуск объясним; retention удаляет только устаревшие записи.

**Результат:** Audit API, retention tests.

#### Д107. Визуальный nested Rule Builder

- Сделать AND/OR groups и typed inputs.
- Добавить actions/priority/delays/schedules.
- Проверить сохранение вложенной структуры на телефоне.

**Проверка дня:** Пользователь создаёт и повторно открывает сложное правило без потери данных.

**Результат:** Automations page, editor checks.

#### Д108. Override и музыкальная сцена

- Проверить manual commands UI/button/gesture/бот через единый dispatcher.
- Блокировать scoped automation/ambient/prediction во время override.
- Разрешать смену music scene только explicit rule setting.

**Проверка дня:** Master OFF и ручные изменения не отменяются rules; исключения явно видны пользователю.

**Результат:** Conflict policy, integration tests, UI explanations.

#### Д109. Preview и сквозные сценарии

- Добавить dry-run с why matched/skipped без side effects.
- Проверить Movie Mode, away delay, Night Shift schedule и sensor rules.
- Прогнать concurrent priorities.

**Проверка дня:** Dry-run не вызывает hardware; реальные actions и audit соответствуют условиям.

**Результат:** Scenario suite, rule examples, help.

#### Д110. Контрольный срез M7

- Нагрузить engine rules/events в simulation.
- Проверить storm/duplicates/stop/restart.
- Повторить ключевые cases на железе.

**Проверка дня:** M7 закрыт; LED FPS сохранён, delays/priority/override подтверждены.

**Результат:** Протокол M7, benchmarks, accepted rule set.

### Неделя 23. Habit data и prediction baseline — Д111–Д115

Связь с ТЗ: §17, 31, 40, 45. Вход: Collector с Д035; M7.

#### Д111. Проверка фактических наблюдений

- Сверить volume/music/monitor/Ambilight/presence/ambient/gesture/override fields.
- Добавить duration aggregation и manual/automatic source.
- Найти gaps/duplicates без выдумывания missing values.

**Проверка дня:** Данные пригодны для заявленных targets и не содержат треков/frames.

**Результат:** Data quality report, collector fixes/migrations.

#### Д112. 90-дневная retention и training data

- Реализовать rolling cleanup observations и связанную feedback/model policy.
- Создать local training snapshot.
- Проверить interrupted cleanup/restart/disk limits.

**Проверка дня:** Данные старше 90 дней удаляются; пользовательская история не покидает устройство.

**Результат:** Retention job, local training pipeline, tests.

#### Д113. Baseline привычек

- Сформировать time/day/context features без sequence modelling.
- Разделить данные по времени и исключить future leakage.
- Сравнить frequency baseline с простой моделью.

**Проверка дня:** Есть воспроизводимые метрики и explicit sparse-data handling.

**Результат:** Habit baseline, time-split evaluation.

#### Д114. Prediction targets и model lifecycle

- Добавить preset/brightness/mode/Ambilight/Night Shift/music start/volume.
- Фильтровать по capabilities и контексту.
- Версионировать модель и atomically replace.

**Проверка дня:** Невозможные actions не предлагаются; missing model даёт no action.

**Результат:** Predictor, fixtures, model lifecycle tests.

#### Д115. Learning gate и confidence

- Проверять минимум 30 полных календарных дней с learning_started_at.
- Добавить SUGGEST/AUTO и configurable thresholds.
- Калибровать confidence и no-action при нехватке данных.

**Проверка дня:** Проверены границы 60/85% и 30 дней; fake-clock test не объявлен реальным обучением.

**Результат:** Policy engine, boundary tests, eligibility API.

### Неделя 24. Recommendations UI, feedback и privacy — Д116–Д120

Связь с ТЗ: §17–18, 24, 40, 50. Вход: Habit pipeline Д111–Д115.

#### Д116. Accept/reject feedback

- Создать recommendation ID и idempotent feedback.
- Использовать reject как negative feedback.
- Не учитывать ручную отмену выполненной automation как feedback.

**Проверка дня:** Дубли не меняют модель дважды; manual cancellation не считается rejection.

**Результат:** Prediction API, feedback tests.

#### Д117. Predictions page

- Показать remaining learning time/confidence/reason/mode.
- Добавить thresholds, accept/reject и insufficient data.
- Дать explicit enable AUTO.

**Проверка дня:** SUGGEST не выполняет action; пользователь понимает блокировку и предлагаемый результат.

**Результат:** Predictions UI, interaction checks.

#### Д118. Удаление learning history

- Удалить observations/feedback/model/recommendations.
- Остановить in-flight training и очистить caches.
- Начать новый 30-day period и определить политику backup copies.

**Проверка дня:** Старые jobs/model не возвращают удалённую историю; restore не обходит заявленную privacy policy.

**Результат:** Delete workflow, concurrency tests, backup privacy policy.

#### Д119. Prediction → command/automation

- Провести AUTO через общий priority/override/capability gate.
- Проверить user away/manual OFF/active music.
- Разделить suggestion и executed action в audit.

**Проверка дня:** Высокий confidence не обходит ручной override; no-user context обрабатывается по политике.

**Результат:** Execution integration tests, audit links.

#### Д120. Контрольный срез M8

- Проверить реальные elapsed days и пригодность данных с Д035.
- Показать допустимую recommendation/reject/delete локально.
- Отдельно зафиксировать полезность и eligibility.

**Проверка дня:** M8 закрыт с реальным временным gate; синтетика не заменяет пользовательские 30 дней.

**Результат:** Протокол M8, learning evidence, quality report.

### Неделя 25. MQTT и Telegram — Д121–Д125

Связь с ТЗ: §21–23, 25, 27, 39–40, 48. Вход: Auth/scopes и device/command layer.

#### Д121. Mosquitto auth/ACL

- Отключить anonymous и ограничить network exposure.
- Создать отдельные credentials/ACL по клиентам.
- Проверить LWT/reconnect и чужие publish/subscribe.

**Проверка дня:** Брокер не пускает анонимных клиентов; устройство не меняет чужие topics.

**Результат:** Broker templates, ACL integration tests.

#### Д122. MQTT autodiscovery

- Задать versioned discovery contract и idempotent registry.
- Подключить state/set/LWT/offline/resync.
- Валидировать payload и unknown client audit.

**Проверка дня:** Fake и тестовое реальное устройство появляются без дублей и корректно теряются.

**Результат:** MQTT adapter, discovery spec, tests.

#### Д123. HA-ready REST/MQTT

- Документировать topics/capabilities и scoped external token.
- Проверить независимого клиента без HA dependency.
- Подключить MQTT events к Automation/simulation.

**Проверка дня:** Внешний клиент читает state и выполняет разрешённую command; core автономен.

**Результат:** Integration guide, sample client, event tests.

#### Д124. Telegram service и user linking

- Создать отдельный Go container с доступом только к Core API.
- Реализовать явную one-time привязку Telegram ID к local user.
- Ограничить scopes и retry/backoff.

**Проверка дня:** Непривязанный пользователь не управляет Hub; бот не монтирует БД/hardware.

**Результат:** Telegram skeleton, linking flow, auth tests.

#### Д125. Telegram commands/buttons

- Добавить /status /lights /music /presets /ambilight /night /system.
- Проверить stale callbacks/permissions/manual override.
- Исключить frames/habit history/ML/CV из любых ответов.

**Проверка дня:** Команды работают; WAN/бот outage не влияет на local controls.

**Результат:** Command tests, privacy review, usage guide.

### Неделя 26. Reliability, backup и recovery — Д126–Д130

Связь с ТЗ: §35–38, 50. Вход: Все функции интегрированы.

#### Д126. Watchdog и dependency recovery

- Добавить heartbeat/timeouts/backoff/circuit policy.
- Отделить ML/MQTT degradation от core failure.
- Проверить crash loops и автоматическое восстановление.

**Проверка дня:** ML failure оставляет свет/audio/обычный Ambilight/rules; MQTT failure не мешает GPIO.

**Результат:** Watchdog, fault-injection suite.

#### Д127. Graceful shutdown/restart state

- Остановить новые commands/automation, fade OFF, flush SQLite, stop workers.
- Восстановить LED last state, music STOPPED, camera OFF, automation enabled.
- Настроить docker/systemd stop timeout.

**Проверка дня:** Остановка и forced restart не создают лишнее включение audio/camera и corruption.

**Результат:** Lifecycle implementation, restart HIL tests.

#### Д128. SQLite Safe Mode

- Смоделировать disk full/corruption/unwritable/busy exhaustion.
- Сохранить защищённое manual lighting/audio по согласованному emergency path.
- Запретить persistent changes/automation.

**Проверка дня:** Без БД не открывается незащищённое LAN управление; recovery требует проверенного хранилища.

**Результат:** Safe Mode, fault tests, recovery runbook.

#### Д129. Daily backup и restore

- Использовать SQLite online backup/согласованный snapshot вместо копии active WAL файла.
- Сохранять config/rules/presets и последние 7 копий.
- Проверить checksums/permissions/restore чистой среды.

**Проверка дня:** Backup под записью консистентен; restore даёт рабочую систему; private data не в release artifacts.

**Результат:** Backup/restore CLI, schedule, integration tests.

#### Д130. Factory reset и M9

- Подтвердить reset явным действием, остановить collectors/jobs.
- Удалить user/runtime/learning state и обработать backups/accesses по политике.
- Повторить fault/restore/reset acceptance.

**Проверка дня:** M9 закрыт; ПО остаётся, прежние данные не возвращаются из jobs/caches.

**Результат:** Reset workflow, M9 protocol, recovery docs.

### Неделя 27. Observability и ресурсы — Д131–Д135

Связь с ТЗ: §3, 38–39, 45–46. Вход: Recovery Д126–Д130.

#### Д131. Полный metrics catalog

- Сверить каждую метрику §38 и units.
- Добавить latency buckets/degraded flags.
- Удалить high-cardinality и personal labels.

**Проверка дня:** Обязательные metrics доступны и отражают реальные состояния.

**Результат:** Metrics catalog, endpoint checks.

#### Д132. Production Thermal Manager

- Реализовать NORMAL/WARM/HOT/CRITICAL и configurable hysteresis.
- Разделить fan curve и warning thresholds.
- Определить safe sensor failure и приоритетное снижение CV/DSP нагрузки.

**Проверка дня:** Пороги проходят tests/HIL; тепловая деградация объяснима и видна.

**Результат:** Thermal service, HIL report, operating limits.

#### Д133. Alerts и System UI

- Добавить temperature/power limit/sensor offline/ML down/disk low.
- Ввести dedupe/rate limit/recovery.
- Доставлять UI/log/Telegram без sensitive data.

**Проверка дня:** Alert storm и недоступный Telegram не блокируют realtime.

**Результат:** Alert manager, System page, delivery tests.

#### Д134. Disk/log/memory bounds

- Настроить rotation и retention jobs.
- Проверить отсутствие запрещённых payload в logs/metrics.
- Задать resource budgets и измерить длительные WS/stream сессии.

**Проверка дня:** Нет бесконечного роста RAM/disk; low disk не вызывает лавину writes.

**Результат:** Resource budget, log policy, longevity report.

#### Д135. Совместный performance profiling

- Совместить audio/DSP/Ambilight/gestures/presence/sensors/rules.
- Измерить REST/WS p95, LED FPS, inference и thermal.
- Оптимизировать подтверждённые bottlenecks.

**Проверка дня:** Все цели §46 выполнены на объявленном reference profile либо блокируют приёмку.

**Результат:** Performance report, traces, tuning config.

### Неделя 28. Production packaging и полный CI — Д136–Д140

Связь с ТЗ: §33, 37, 41, 43–44. Вход: Все сервисы и resource baseline.

#### Д136. Production Docker/Compose

- Сделать multi-stage/non-root/read-only где возможно.
- Выделить volumes/tmpfs/device permissions.
- Закрепить base images и проверить arm64/amd64.

**Проверка дня:** Пять сервисов запускаются без лишних host mounts/privileges.

**Результат:** Production Dockerfiles/Compose, access inventory.

#### Д137. Полный CI качества

- Включить gofmt/vet/golangci-lint/govulncheck/gosec/race.
- Добавить ESLint/TS/unit/build и Ruff/mypy/pytest/dependency scan.
- Настроить secrets/container/dependency security reports.

**Проверка дня:** PR checks работают без Pi и production secrets; обязательные проверки не silently skipped.

**Результат:** CI workflows, tool versions, reports.

#### Д138. Multi-arch и client packaging

- Собирать arm64/amd64 images и четыре клиента.
- Паковать native dependencies/models/checksums.
- Связать artifact manifest с commit/tag.

**Проверка дня:** Packages не содержат БД/credentials/private configs; происхождение однозначно.

**Результат:** Build/package workflows, manifests.

#### Д139. Systemd и microSD → SSD

- Описать bootstrap/network ordering/restart/healthchecks.
- Подготовить перенос boot/data с backup и возвратом.
- Проверить USB storage power и reboot на реальной Pi.

**Проверка дня:** Pi загружается с целевого SSD, данные сохранены, возврат документирован.

**Результат:** Systemd unit, install scripts, SSD runbook.

#### Д140. Чистая production установка

- Развернуть package на чистом volume/стенде.
- Пройти admin bootstrap/TLS/device config/pairing.
- Проверить permissions и reboot по инструкции.

**Проверка дня:** Нет неописанных ручных шагов; production install воспроизводим.

**Результат:** Install acceptance log, updated guide.

### Неделя 29. Tag deploy и rollback — Д141–Д145

Связь с ТЗ: §37–38, 43–44, 50. Вход: Production packaging Д136–Д140.

#### Д141. Release workflow и runner

- Разрешить deploy из доверенного tag workflow.
- Изолировать self-hosted runner от public PR/fork кода.
- Ограничить secrets/permissions и concurrent deploy.

**Проверка дня:** PR не выполняется на production runner; branch push не запускает deployment.

**Результат:** Release workflow, runner guide, deployment lock.

#### Д142. Deploy/readiness gate

- Использовать exact digests/config validation/disk checks.
- Делать pre-deploy backup и controlled migrations.
- Проверять critical readiness, required-profile services и smoke.

**Проверка дня:** Broken image/schema не объявляется успешным; optional degradation отображается корректно.

**Результат:** Deploy script, readiness/smoke checks.

#### Д143. Rollback images и schema

- Хранить manifest прошлой рабочей версии.
- Определить forward/backward migration compatibility.
- Проверить restore pre-deploy snapshot без concurrent writes.

**Проверка дня:** Старый image не запускается с несовместимой БД; восстановление данных явное.

**Результат:** Rollback script, migration policy, tests.

#### Д144. Аварийная репетиция release

- Создать pre-release с broken health/download/migration.
- Проверить auto rollback/lock/audit.
- Выполнить ручной recovery по runbook.

**Проверка дня:** Каждый сбой возвращает рабочую версию либо явный безопасный отказ с инструкцией.

**Результат:** Release failure matrix, rehearsal evidence.

#### Д145. Контрольный срез M10

- Провести pre-release tag через checks/package/deploy.
- Проверить reboot/health/key functions.
- Отрепетировать secret rotation и fork protections.

**Проверка дня:** M10 закрыт полным tag flow и доказанным rollback на Pi.

**Результат:** Протокол M10, tested release checklist.

### Неделя 30. E2E, security и 48h soak — Д146–Д150

Связь с ТЗ: §33–46, 50. Вход: M10; production profile.

#### Д146. Покрытие и traceability

- Связать requirements с unit/integration/E2E/HIL/simulation.
- Измерить core >=80% и overall >=70% по явному составу.
- Закрыть meaningful gaps рискованных веток.

**Проверка дня:** Есть отдельные отчёты языков и прозрачный aggregate без generated/vendor/fixtures.

**Результат:** Coverage report, traceability matrix.

#### Д147. Полный simulation E2E

- Запустить fake screen → analyzer → core → virtual strip и golden output.
- Добавить audio/gesture/presence/rules/MQTT fault sequences.
- Проверить детерминизм/race/CI diagnostics.

**Проверка дня:** Suite стабилен без Pi; ошибка указывает нарушенное требование.

**Результат:** Simulation E2E suite, CI reports.

#### Д148. Security/privacy review

- Проверить auth/scopes/CSRF/WS origin/rate limits/pairing/revoke.
- Проверить secrets в repo/images/logs и LAN isolation.
- Подтвердить RAM-only camera/laptop-only screen/local-only habits.

**Проверка дня:** Нет критических обходов/утечек; privacy доказана наблюдением, а не только config.

**Результат:** Security/privacy report, defect list.

#### Д149. Старт 48-часового soak

- Запустить representative audio+LED+screen+CV+sensors+rules.
- Включить bounded diagnostic logs и resource assertions.
- Сделать controlled disconnect/reconnect без остановки стенда.

**Проверка дня:** За первый цикл нет необъяснимого роста RAM/thermal failures; отклонения зафиксированы.

**Результат:** Soak start timestamp, monitoring artifacts.

#### Д150. Продолжение soak и triage

- Проверить второй суточный цикл и daily backup.
- Разбирать defects в отдельной среде, не меняя soak reference.
- Подготовить результаты и обязательные fixes.

**Проверка дня:** Наблюдение продолжается без изменения reference build; итог >=48 часов фиксируется не раньше фактического завершения, обычно на Д151.

**Результат:** промежуточный soak report, prioritized fixes; финальная отметка после фактических 48 часов.

### Неделя 31. Исправления и формальная приёмка — Д151–Д155

Связь с ТЗ: §45–46, 50. Вход: Результаты Д146–Д150.

#### Д151. Release blockers и regression

- После фактических 48 часов завершить и сохранить soak report; закрыть P0/P1 из soak/security/E2E.
- Добавить regression checks по причине дефекта.
- Повторить affected HIL/performance и при необходимости новый soak.

**Проверка дня:** Критических открытых дефектов нет; крупные изменения не обходят повторную стабилизацию.

**Результат:** Fix set, regression report, candidate build.

#### Д152. Приёмка загрузки/audio/Ambilight

- Пройти §50 пункты 1–20 на Pi/телефоне/мониторе.
- Проверить HTTPS/auth/outputs/DSP/Cinema/ML/fallback.
- Измерить FPS/e2e по установленной методике.

**Проверка дня:** Каждый пункт 1–20 имеет pass/fail и evidence; отсутствующее доказательство не pass.

**Результат:** Acceptance protocol A.

#### Д153. Приёмка жестов/кнопки/sensors/rules

- Пройти §50 пункты 21–39.
- Проверить indicator/RAM-only/press types/five sensors/nested rules/delay/priority/override.
- Повторить restart между группами.

**Проверка дня:** Пункты 21–39 подтверждены реальным стендом, а не только mocks.

**Результат:** Acceptance protocol B.

#### Д154. Приёмка habits/integrations

- Пройти §50 пункты 40–49 и targets §17.
- Проверить real 30 days/confidence/reject/delete/MQTT/Telegram.
- Повторить без WAN и проверить HA-ready contracts.

**Проверка дня:** Пункты 40–49 закрыты; useful prediction и eligibility описаны раздельно.

**Результат:** Acceptance protocol C, learning evidence.

#### Д155. Приёмка отказов/deployment

- Пройти §50 пункты 50–62 с fault injection.
- Проверить Safe Mode/restart/backup/restore/reset/release/rollback.
- Сверить полный список 62 критериев.

**Проверка дня:** Незакрытый обязательный пункт блокирует v1.0.0; все отклонения записаны.

**Результат:** Acceptance protocol D, complete 62-point matrix.

### Неделя 32. Документация и v1.0.0 — Д156–Д160

Связь с ТЗ: §35, 41, 44, 50, 54. Вход: Формальная приёмка Д152–Д155.

#### Д156. Пользовательское руководство

- Написать onboarding с телефона и инструкции всех страниц.
- Объяснить privacy/learning/delete/source limitations.
- Добавить troubleshooting по реальным дефектам.

**Проверка дня:** Новый пользователь выполняет основные сценарии без чтения исходников.

**Результат:** User guide, FAQ, onboarding checklist.

#### Д157. Эксплуатация и hardware docs

- Свести BOM/wiring/power/thermal/calibration.
- Описать TLS/backup/restore/reset/SSD/update/rollback.
- Проверить команды на чистом стенде.

**Проверка дня:** Документы соответствуют реальному production profile и HIL результатам.

**Результат:** Operations runbook, hardware docs.

#### Д158. Архитектурный отчёт и demo

- Описать модули/event-command flow/schema/API/models.
- Собрать performance/quality/privacy/acceptance evidence.
- Подготовить demo без персональных данных.

**Проверка дня:** Материалы позволяют воспроизвести измерения и использовать проект в учебной работе.

**Результат:** Architecture report, demo script, evidence index.

#### Д159. Финальная репетиция release

- Проверить clean tree/licenses/models/secrets/manifest.
- Повторить clean install/restore и smoke release candidate.
- Подготовить recovery point и release notes.

**Проверка дня:** Release candidate совпадает с проверенным artifact; release blockers отсутствуют.

**Результат:** Release notes, final checklist, tested candidate.

#### Д160. v1.0.0 и M11

- Создать tag проверенного commit и провести штатный deploy.
- Проверить health/smoke/metrics и ручные controls.
- Сохранить финальную приёмку и backlog развития.

**Проверка дня:** M11 закрыт только с успешным deploy и всеми обязательными критериями.

**Результат:** v1.0.0, финальный отчёт, backlog vNext.

## 6. Резерв: Д161–Д180, до 20 рабочих дней

Это **банк времени**, а не обязательные четыре недели после выпуска. Дни резерва вставляются перед зависимым этапом или релизом при возникновении проблемы; исходная нумерация служит идентификатором задачи. Например, проблема macOS на Д062 требует использования соответствующего резерва до M4, а не после v1.0.0. Фактический график при этом сдвигается.

Приоритет использования: незакрытые обязательные критерии → повторные испытания после исправлений → документация и удобство установки → optional-функции. Резерв нельзя одновременно засчитывать как время исправлений и время расширения.

### Неделя резерва R1: оборудование, аудио и захват

#### Д161. Несовместимость периферии

- Повторить сбой на минимальном стенде; проверить драйвер, kernel/device permissions и wiring.
- Заменить неподдерживаемый компонент кандидатом из BOM; обновить pinout и конфигурацию.
- Проверить аппаратный контракт и simulation, исключить регрессии соседних устройств.

**Результат и проверка:** заменённый adapter проходит прежний HIL-протокол; новая совместимость документирована.

#### Д162. Конфликты Bluetooth и audio routing

- Отдельно проверить pairing, input, output, PCM monitor и source switching.
- Исправить аудиограф/permissions/reconnect; проверить ограничение аппаратной одновременности.
- Повторить затронутые source × output комбинации вместе с LED.

**Результат и проверка:** обязательные источники/выходы работают; ограничения не скрывают невыполненное требование.

#### Д163. Linux capture и display sessions

- Воспроизвести проблему в заявленных session types, monitor layouts и scaling.
- Исправить permissions/portal/capture adapter или уточнить проверяемую границу поддержки.
- Повторить pairing, capture, stream и fallback на Linux.

**Результат и проверка:** поддерживаемая Linux-конфигурация проходит M4; изменение требований фиксируется отдельно.

#### Д164. Windows/macOS capture

- Устранить сбои platform API, permissions, Retina/DPI и native dependencies.
- Проверить отказ и отзыв разрешений, locked session и disconnected monitor.
- Повторно установить пакет в чистой среде проблемной платформы.

**Результат и проверка:** исправленный пакет проходит реальную платформенную матрицу.

#### Д165. Недостигнутая Ambilight latency

- По traces определить capture, analyzer, сеть, очередь или LED как источник задержки.
- Исправить одну подтверждённую причину; сравнить до/после на одинаковом стенде.
- Повторить 30 FPS/<100 мс вместе с ML, audio и UI.

**Результат и проверка:** целевая задержка достигнута без снижения обязательного качества; иначе требуется дополнительная переоценка.

### Неделя резерва R2: ML, приватность и автоматизация

#### Д166. Недостаточное качество Scene ML

- Разобрать ошибки по классам и типам сцен, проверить разметку и split.
- Улучшить данные/признаки/модель с подбором по validation; не подгонять существующий test.
- При исчерпании test как независимого набора выделить новый holdout и повторить оценку.

**Результат и проверка:** заранее принятый quality gate пройден; ограничения модели опубликованы.

#### Д167. Ложные жесты или превышение latency

- Разделить задержку захвата, detection, confirmation, routing и hardware action.
- Настроить gating/cooldown/features с сохранением пользовательского сценария.
- Повторить negative cases и quality/latency matrix на Pi.

**Результат и проверка:** false-positive и response targets достигнуты; смена точки отсчёта не используется как исправление.

#### Д168. CV/presence и ресурсные конфликты

- Проверить camera ownership, scheduler, memory buffers и indicator lifecycle.
- Исправить starvation и oversubscription; проверить все четыре presence source.
- Повторить combined load с отключением/переподключением камеры.

**Результат и проверка:** CV не ухудшает audio/LED/Ambilight; camera privacy соблюдается.

#### Д169. Конфликты правил и отложенных действий

- Воспроизвести последовательность событий с fake clock и сохранёнными IDs.
- Исправить priority/tie-break/override/idempotency/cancellation.
- Проверить restart между trigger и action и возврат пользователя до таймера.

**Результат и проверка:** regression test подтверждает единственное ожидаемое действие и audit.

#### Д170. Learning data и confidence

- Проверить реальные gaps, eligibility, imbalance и калибровку confidence.
- Исправить collector и обучение, сохранив local-only и temporal split.
- Повторить feedback/delete/relearning; при нехватке календарных данных оставить gate открытым.

**Результат и проверка:** нет фиктивных данных и преждевременного AUTO; ожидание 30 дней не заменяется одним резервным днём.

### Неделя резерва R3: надёжность и повторная приёмка

#### Д171. SQLite recovery и backup consistency

- Воспроизвести повреждение, disk full или restore mismatch на копии стенда.
- Исправить snapshot/permissions/manifest/Safe Mode без открытия LAN доступа.
- Восстановить систему с активными ранее rules/presets и проверить целостность.

**Результат и проверка:** свежий backup проверенно восстанавливается; аварийный доступ соответствует security policy.

#### Д172. Deployment и несовместимые миграции

- Повторить broken deploy с exact artifact digests и предыдущей схемой.
- Исправить readiness/locking/backup/rollback порядок.
- Проверить возврат рабочей версии и повторное штатное обновление.

**Результат и проверка:** обе операции воспроизводимы без ручного редактирования БД.

#### Д173. Исправления security/privacy

- Закрыть найденные обходы auth/scopes/origin, утечки или лишние permissions.
- Отозвать/заменить затронутые секреты, если проблема возникла на реальном стенде.
- Повторить конкретный exploit/negative test и затронутые пользовательские сценарии.

**Результат и проверка:** критические замечания закрыты с evidence; секреты отсутствуют в публичных артефактах.

#### Д174. Повтор длительного soak после существенных исправлений

- Развернуть исправленный candidate на эталонном production profile.
- Запустить автоматическое наблюдение не менее 48 реальных часов; независимую документацию делать параллельно.
- Сравнить ресурсы, reconnect, backups и errors с предыдущим прогоном.

**Результат и проверка:** итог фиксируется только после истечения 48 часов; этот пункт использует один день активного труда, но несколько календарных дней наблюдения.

#### Д175. Повтор затронутой формальной приёмки

- Выбрать пункты §50, на которые повлияли исправления, и соседние критические сценарии.
- Повторить их на candidate с окончательным manifest.
- Обновить evidence и решение о готовности; при незакрытом blocker перенести релиз.

**Результат и проверка:** матрица приёмки относится к фактически выпускаемой сборке.

### Неделя резерва R4: оставшийся резерв и optional-функции

Следующие задачи выполняются только при закрытых обязательных критериях и неиспользованном резерве. Иначе эти дни продолжают исправление наиболее важного blocker.

#### Д176. Gaming Mode

- Добавить профиль минимального smoothing и быстрого перехода.
- Сравнить e2e latency и визуальную стабильность с Cinema на одном наборе сцен.
- Проверить переключение профилей без обхода power limits.

**Результат и проверка:** дополнительный профиль документирован; Cinema не регрессировал.

#### Д177. Optional Grafana profile

- Добавить Prometheus/Grafana Compose profile с ограниченным доступом.
- Собрать dashboard из реально реализованных метрик, настроить retention.
- Проверить RAM/disk budget и запуск системы без optional profile.

**Результат и проверка:** наблюдаемость удобнее, базовый production profile остаётся автономным.

#### Д178. Реальная LED matrix при наличии панели

- Настроить width/height/orientation/serpentine и питание.
- Проверить mapping и применимые эффекты на реальной панели.
- Сверить virtual/physical output и общий PSU limiter.

**Результат и проверка:** аппаратная матрица соответствует уже обязательной логической поддержке; отсутствие купленной панели не блокирует v1.

#### Д179. Демонстрационная интеграция Home Assistant

- Подключить отдельный HA-стенд через опубликованные REST/MQTT contracts.
- Проверить read/control, scopes и отказ внешнего клиента.
- Записать пример настройки без включения HA в обязательные зависимости Core.

**Результат и проверка:** интеграция воспроизводима; полноценный MQTT Discovery остаётся отдельным будущим требованием.

#### Д180. Удобство установки и итог резерва

- Провести установку по документации с участием нового пользователя.
- Устранить наиболее частые вопросы onboarding и recovery.
- Сохранить фактический расход резерва, итоговую оценку и backlog следующего релиза.

**Результат и проверка:** нет недокументированных действий; optional-изменения прошли затронутые проверки.

## 7. Привязка к 62 пунктам приёмки ТЗ

Таблица указывает основную реализацию и формальную проверку. Предварительные unit/integration/HIL проверки выполняются в дни разработки; последняя колонка не заменяет их. При заполнении итогового протокола к каждой строке добавляются фактические status, commit/artifact ID, окружение и ссылка на evidence.

| № §50 | Что должно быть принято | Реализация / ранняя проверка | Формальная проверка |
|---:|---|---|---|
| 1 | Загрузка Pi с Ubuntu | Д006, Д139–Д140 | Д152 |
| 2 | systemd → Compose autostart | Д136, Д139–Д140 | Д152 |
| 3 | HTTPS smarthome.local | Д009, Д022, Д140 | Д152 |
| 4 | Admin login | Д018–Д020, Д022 | Д152 |
| 5 | Dashboard Pi/devices | Д023, Д094, Д133 | Д152 |
| 6 | Bluetooth phone connection | Д038–Д040 | Д152 |
| 7 | Запуск музыки | Д039–Д045 | Д152 |
| 8 | Выбранный audio output | Д043–Д045 | Д152 |
| 9 | desk.wall bass/beat/FFT | Д047–Д055 | Д152 |
| 10 | Realtime spectrum/BPM | Д049, Д054–Д055 | Д152 |
| 11 | Запуск laptop client | Д058, Д061–Д063, Д068 | Д152 |
| 12 | Screen pairing | Д056, Д069 | Д152 |
| 13 | AMBILIGHT zone mode | Д060, Д066 | Д152 |
| 14 | Цвета четырёх сторон | Д059–Д060, Д064, Д070 | Д152 |
| 15 | Ambilight около 30 FPS | Д067, Д070, Д079 | Д152 |
| 16 | Ambilight <100 мс e2e | Д067, Д070, Д080 | Д152 |
| 17 | Cinema плавность | Д065, Д070 | Д152 |
| 18 | Scene ML adaptation | Д071–Д080 | Д152 |
| 19 | Ambilight без ML | Д075, Д080, Д126 | Д152 |
| 20 | Screen disconnect fallback | Д060, Д065, Д069 | Д152 |
| 21 | Gesture Mode | Д081–Д083, Д088 | Д153 |
| 22 | OPEN_PALM Play/Pause | Д085, Д088–Д090 | Д153 |
| 23 | Swipe tracks | Д086, Д088–Д090 | Д153 |
| 24 | PALM_FACE_DOWN Night Shift | Д086, Д088–Д090 | Д153 |
| 25 | Rotation volume | Д086, Д088–Д090 | Д153 |
| 26 | Camera frames не сохраняются | Д081–Д090, Д148 | Д153 |
| 27 | CV activity indicator | Д082, Д090, Д099 | Д153 |
| 28 | Single press ON/OFF | Д033–Д035 | Д153 |
| 29 | Double press Night Shift | Д033–Д035 | Д153 |
| 30 | Long press next preset | Д033–Д035 | Д153 |
| 31 | Пять типов sensor values | Д091–Д094 | Д153 |
| 32 | Ambient brightness correction | Д095 | Д153 |
| 33 | Sensors доступны automation | Д102, Д109 | Д153 |
| 34 | AND/OR Rule Builder | Д101–Д102, Д107 | Д153 |
| 35 | Исполнение rule | Д104, Д109–Д110 | Д153 |
| 36 | Delayed action | Д105, Д109 | Д153 |
| 37 | Priority | Д104, Д108–Д110 | Д153 |
| 38 | Manual override | Д032, Д108, Д119 | Д153 |
| 39 | Automation audit | Д106, Д109 | Д153 |
| 40 | Четыре presence source | Д096–Д100 | Д154 |
| 41 | Local-only habits | Д035, Д111–Д114, Д148 | Д154 |
| 42 | Recommendation после 30 дней | Д115, Д120 | Д154 |
| 43 | Confidence thresholds | Д115, Д117, Д119 | Д154 |
| 44 | Reject recommendation | Д116–Д117 | Д154 |
| 45 | Delete learning history | Д118, Д120 | Д154 |
| 46 | MQTT autodiscovery | Д121–Д122 | Д154 |
| 47 | Telegram commands | Д124–Д125 | Д154 |
| 48 | Работа без Telegram | Д125–Д126 | Д154 |
| 49 | HA-ready architecture | Д123 | Д154 |
| 50 | ML failure degradation | Д046, Д075, Д126 | Д155 |
| 51 | MQTT failure / GPIO alive | Д122, Д126 | Д155 |
| 52 | SQLite Safe Mode | Д128 | Д155 |
| 53 | LED last state restore | Д035, Д127 | Д155 |
| 54 | Music STOPPED after restart | Д045, Д127 | Д155 |
| 55 | Camera OFF after restart | Д081, Д090, Д127 | Д155 |
| 56 | Daily backup | Д129, Д150 | Д155 |
| 57 | Restore backup | Д129, Д140 | Д155 |
| 58 | Factory reset | Д130 | Д155 |
| 59 | Release tag → Actions | Д141–Д145 | Д155 |
| 60 | Self-hosted deploy | Д141–Д145 | Д155 |
| 61 | Readiness deploy gate | Д013, Д142, Д145 | Д155 |
| 62 | Automatic rollback | Д143–Д145 | Д155 |

## 8. Как измерять результат

### Performance

- Для каждого прогона сохранять hardware/OS/build, число LED, FPS настройки, сеть, camera resolution, активные модели и список сервисов. Числа без профиля нагрузки не сравнивать.
- REST p95 <100 мс проверять на согласованном наборе обычных операций и параллельных клиентов; долгие операции вроде restore измерять отдельно. До нагрузочного теста зафиксировать ожидаемые RPS, concurrency, warmup и длительность.
- WS latency <100 мс измерять от публикации события до получения клиентом, а не только временем записи в socket. Ограничения часов учитывать общей опорой/синхронизацией и погрешностью.
- LED rendering >=30 FPS и Ambilight 30 FPS проверять по реально применённым кадрам, а не только generated frames. Фиксировать p95/max frame interval, dropped frames и длительность прогона.
- Ambilight e2e <100 мс проверять от изменения тестового изображения до LED-реакции. Одного RTT недостаточно; использовать визуальную запись подходящего разрешения по времени либо общий измерительный стенд. Описать точность метода.
- Gesture reaction <500 мс измерять от начала устойчивого жеста до выполненного действия, включая confirmation. Отдельно сохранять detection/inference/queue/action timing и false actions/min.
- Scene inference — не чаще 1 раза в секунду, независимо от 30 FPS colors. Presence camera inference — раз в 2–3 секунды; измерять также пропуски при совместной нагрузке.
- Если цель недостижима на Pi 4 при всём обязательном наборе одновременно, документировать bottleneck и переоценить архитектуру/срок. Самовольное снижение обязательных требований не считается завершением.

### ML quality

ТЗ не задаёт численных порогов качества классов сцен, жестов и habit prediction. Поэтому **нельзя честно обещать заранее конкретную accuracy**. На Д005, Д072 и Д087 нужно определить пригодный набор, минимальное качество и процедуру приёмки до обучения/тюнинга.

Для сцен использовать per-class precision/recall/F1 и macro/micro показатели; для жестов — качество каждого класса, negative motion, false actions/min и end-to-end delay; для привычек — temporal holdout, baseline comparison, calibration и качество рекомендаций по targets. Confidence 85% — не доказательство 85% фактической точности без калибровки.

Тридцать дней — минимальный срок сбора, а не гарантия достаточного количества примеров. При редком использовании система должна показывать insufficient data и оставаться в SUGGEST/no-action. Синтетические 30-дневные fixtures нужны для тестов времени и retention, но не заменяют реальные наблюдения пользователя.

### Покрытие и тестовые артефакты

На Д001 определить границы core business logic. На Д146 собирать отдельно Go/Python/frontend coverage и общий показатель по заранее согласованному составу исполняемого кода. Не усреднять проценты трёх языков арифметически; при общей line metric использовать сумму покрытых строк / сумму учитываемых строк с описанными ограничениями инструментов. Generated/vendor/fixtures исключать явно и одинаково во всех прогонах.

Обязательные golden fixtures: Ambilight mapping, FFT, beats, LED effects, gesture mapping, scene preprocessing. Обязательные benchmarks: render loop, screen processing, WS, event bus, FFT, inference. HIL нельзя заменить высоким unit coverage.

## 9. Реальные календарные ожидания

### Habit learning

- Д035: запуск collector; сохранить фактический timestamp.
- После каждого нового источника проверять полноту данных. Ранние lighting-only наблюдения не гарантируют обучаемость volume/music targets.
- Д111: аудит качества и достаточности; при повреждённой или слишком неполной истории срок полезного наблюдения может увеличиться.
- Д115: программная проверка eligibility; Д120 и Д154: реальные демонстрации.
- После delete history новый 30-дневный период обязателен. Проверку удаления делать на тестовом пользователе/копии стенда либо после сохранения evidence исходного реального периода, чтобы не уничтожить единственное доказательство для следующей приёмки. Пользовательские данные не переносить в публичные fixtures.

### Длительные прогоны

48-часовой soak начинается на Д149 и идёт фоном. При стандартном рабочем графике два соседних рабочих дня не равны 48 часам: итог можно фиксировать не раньше фактического истечения срока, обычно на Д151. До изменения reference build сохранить результаты. Существенные исправления требуют нового длительного прогона и могут использовать резерв Д174. Daily backup должен быть подтверждён реальным расписанием, а не только ручным вызовом.

### Поставки и доступ к платформам

Если Pi/PSU/камеры/датчиков или Intel Mac нет к нужному дню, зафиксировать блокер сразу. Симуляция позволяет продолжать независимые работы, но M2/M4/M6 не закрываются без реального железа. Отсутствие обязательного стенда может добавить календарное ожидание сверх 180 рабочих дней.

## 10. Управление рисками и изменениями

| Риск | Ранний сигнал | Действие | Когда пересматривать оценку |
|---|---|---|---|
| LED/audio конфликт ресурсов | Не работают два канала при audio playback | Пересмотреть драйвер/пины/host graph в рамках подтверждённой архитектуры | Д008, до основного Lighting |
| Audio sources не дают общий PCM | Tap работает только для local files | Исправить единый audio graph до DSP | Д037, до M3 |
| Cross-platform capture/runtime | Пакет компилируется, но не запускается на ОС | Проверять native target и dependencies на реальной платформе | Д061–Д063, Д071 |
| Scene/gesture quality недостаточны | Ошибки на negative/holdout, высокая задержка | Данные/признаки/gating, затем повторная независимая оценка | Д078–Д079, Д089 |
| Перегрев/недостаток RAM | Throttling, OOM, frame drops при joint load | Ограничить queues/workers, профилировать и управлять приоритетами | Д055, Д100, Д135 |
| Habit data недостаточны | Мало real observations, неоткалиброванный confidence | No-action/SUGGEST, продолжать наблюдение; не включать фиктивный AUTO | Д111, Д120 |
| Конфликт автоматики с ручным действием | Свет включается после Master OFF | Общий command arbiter, scoped override, regression tests | Д032, Д108, Д119 |
| SQLite/WAL backup неконсистентен | Restore проходит, данные потеряны | Online backup, manifest и реальный restore test | Д129 |
| Rollback ломает схему | Старый binary не понимает новую БД | Pre-deploy snapshot и явная migration compatibility policy | Д143 |
| Public repo и self-hosted runner | PR может добраться до production host/secrets | Раздельные доверенные jobs и изоляция runner; deploy только из защищённого release path | Д015, Д137, Д141 |
| Нереалистичный темп одного разработчика | Два недельных рубежа подряд не закрыты | Пересчитать оставшийся объём по фактической скорости | Каждую пятницу |

Резерв 20 дней не покрывает произвольную замену архитектуры, обучение новому стеку, длительную поставку или отсутствие платформенных стендов. Если значимый риск реализовался, новый срок объявляется явно.

### Еженедельная ревизия

В конце пятого дня недели сравнить план/факт, закрытие критериев, defect backlog и остаток резерва. Не переносить задачи молча. В журнале фиксировать причину, зависимые дни и новый прогноз. Новая функция либо заменяет согласованный объём, либо добавляет оценку; она не помещается в резерв автоматически.

Срезы M1–M10 — внутренние демонстрации/pre-release. Продуктовая v1.0.0 возможна только после §50. Они не означают, что сокращённый набор возможностей уже удовлетворяет всему ТЗ.

## 11. Шаблоны рабочих записей

### Запись дня

```text
День / дата:
Запланированный результат:
Фактически выполнено:
Commit / artifact / config profile:
Проверки и ссылки на отчёты:
Аппаратный стенд / версии:
Затраты времени:
Отклонения и причина:
Открытые дефекты / зависимости:
Израсходованный резерв:
Следующий конкретный шаг:
```

### Строка приёмочного протокола

```text
Требование / пункт §50:
Build / commit / image digest / model version:
Предусловия и стенд:
Шаги воспроизведения:
Ожидаемый результат:
Фактический результат и измерения:
Status: PASS / FAIL / BLOCKED
Evidence: log / test report / локальная визуальная проверка
Дефект и повторная проверка:
```

Для доказательств приватности не создавать записи пользовательских камер/экранов, нарушающие само требование. Использовать synthetic fixtures, метрики, наблюдение сетевого payload и файловой активности, а также локальную визуальную проверку без сохранения чувствительного содержимого.
