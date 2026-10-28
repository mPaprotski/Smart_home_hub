# День 1. Матрица требований Smart Home Hub

Уточнение владельца26.10.2026: HW-001 требует только Raspberry Pi 4B с2GB RAM; [бюджет и memory gates](development/pi4b-2gb.md). Остальные функциональные требования не сняты.

Источник: [ТЗ версии 1.0](../Smart_Home_Hub_TZ.md). Дата разбора: 25 октября 2026 года.

Это декомпозиция исходного ТЗ, а не отчёт об уже разработанной системе. **Статус всех продуктовых требований: «Не реализовано / не проверено».** Проверки в таблицах — будущие способы доказательства. Отдельные пункты теста могут быть уточнены при решении связанных вопросов, но исходное требование нельзя молча ослабить.

Связанные документы: [backlog](backlog.md), [открытые вопросы](open-questions.md), [ежедневный план](DEVELOPMENT_PLAN_DAILY.md), [результат Дня 1](DAY_001_RESULT.md).

**Проектная проработка Д003:** [архитектурные контракты и ADR](architecture/README.md), [схема данных v0](architecture/data-model-v0.md), [сценарии](architecture/scenarios.md). IDs и обязательность требований сохранены; продуктовые статусы остаются «Не реализовано / не проверено».

**Feasibility Д004–Д010:** [execution report](DAYS_004_010_RESULT.md), [case matrix](../experiments/results/case-matrix.json), [M0 status](milestones/M0_TECHNICAL_BASELINE.md). Spike PASS-HOST/SIM не означает product requirement PASS; hardware и platform cases остаются BLOCKED.

## Правила чтения

- ID стабилен: после начала разработки строки не перенумеровывать; удалённые IDs помечать снятыми с обоснованием.
- «Обязательно» — условие первой версии. «Опционально» и «Рекомендация ТЗ» не блокируют обязательную приёмку. «Уточнить обязательность» сохраняет неоднозначность §5.
- «§50» связывает требование с номером финального сценария. «—» означает дополнительную обязательную/опциональную проверку вне этого сценария, а не отсутствие проверки.
- Этап — номер из §51. Дни — окно реализации/проверок из ежедневного плана. Сквозное требование может реализовываться в нескольких окнах.
- Q-xxx — вопрос из отдельного документа с текущим статусом. На Д003 часть semantics закреплена в ADR v0; незафиксированные предложения остаются вариантами. Проектное решение не означает выполненную продуктовую проверку.
- В ID ARCH/HW и подобных содержатся constraints и нефункциональные требования, а не только пользовательские функции. Перечни внутри одной строки описывают один contract/schema; проверять нужно каждый перечисленный элемент.
- Числа из примеров ТЗ (rate limits, confidence, fan policy) не объявлены новыми жёсткими нормативами. Обязательны соответствующие механизмы; точные настройки требуют решения.

Каталог содержит **367 требований/ограничений**: 357 — обязательно, 4 — опционально, 4 — рекомендация тз, 2 — уточнить обязательность.

## ARCH. Архитектура и автономность

Этап §51: **1**. Рабочие окна: **Д001–Д025**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| ARCH-001 | Основные функции и персональные данные работают локально без обязательного интернета | §1,2,54 | Обязательно | Отключить WAN; пройти manual LED, local audio, sensors и локальные rules | 48, 49 | Q-001 |
| ARCH-002 | Control plane реализован как модульный Go-монолит | §2,7 | Обязательно | Проверить структуру модулей и единственный Core process | — | — |
| ARCH-003 | ML/CV/DSP реализованы на Python в отдельном ML Runtime; Scene ML работает на ноутбуке | §2,7 | Обязательно | Проверить deployment/process map и размещение каждого inference/DSP worker | 18, 41, 50 | Q-006 |
| ARCH-004 | Компоненты взаимодействуют через REST/WebSocket/event bus | §2,7 | Обязательно | Проверить API/event contracts и отсутствие скрытых межсервисных вызовов hardware | — | — |
| ARCH-005 | Модель данных/API допускает несколько пользователей без переписывания ядра | §2,26 | Обязательно | Проверить User/Role/Session и привязку пользовательских сущностей в schema/API review | — | Q-024 |
| ARCH-006 | Проект хранится как публичный GitHub monorepo со структурой §41 | §2,41 | Обязательно | Сопоставить дерево репозитория с Go/ML/frontend/deploy/tests/docs и GitHub visibility | — | Q-026 |
| ARCH-007 | Hardware abstraction, control plane, realtime events, ML/CV, automation и UI разделены | §7,54 | Обязательно | Проверить зависимости модулей и возможность подмены adapters без изменения business logic | 49 | — |
| ARCH-008 | Core использует единый event interface вместо прямой зависимости там, где подходит событие | §8 | Обязательно | Проверить publish/subscribe между audio, lighting, sensors, automation и ML | — | Q-030 |
| ARCH-009 | LED, audio, sensor и device adapters выражены интерфейсами с mock implementations | §42 | Обязательно | Запустить одну business operation с real/fake adapter без изменения core logic | — | — |
| ARCH-010 | Первая версия содержит контейнеры smart-home-core, ml-runtime, frontend, telegram-bot, mqtt-broker | §47 | Обязательно | Проверить production Compose services и межсервисные boundaries | 2, 47 | — |
| ARCH-011 | SQLite находится в persistent volume | §47 | Обязательно | Пересоздать Core container и проверить сохранённые данные | 53 | — |
| ARCH-012 | Screen-client запускается отдельно на пользовательском ноутбуке | §47 | Обязательно | Запустить клиент без установки capture/inference в Hub | 11 | — |

## HW. Аппаратная база и питание

Этап §51: **1**. Рабочие окна: **Д002–Д010; Д027–Д034; Д139**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| HW-001 | Целевое устройство — Raspberry Pi 4 Model B с 2GB RAM и Ubuntu Server в headless mode | §3.1 | Обязательно | Загрузить Pi без дисплея и проверить model/2GB RAM/OS/SSH и совместный memory budget на единственной2GB плате | 1 | Q-002 |
| HW-002 | Допускается первоначальная microSD и поддерживается миграция на целевой USB SSD | §3.1 | Обязательно | Выполнить backup/migration/reboot с SSD и проверить данные, описать возврат | 1 | Q-003 |
| HW-003 | Pi имеет активное охлаждение: радиатор и вентилятор 30–40 мм | §3.2 | Обязательно | Осмотреть сборку и проверить работу вентилятора под нагрузкой | — | Q-004 |
| HW-004 | Вентилятор управляется GPIO через транзистор/MOSFET | §3.2 | Рекомендация ТЗ | Проверить схему и включение/выключение через выбранный driver | — | Q-004 |
| HW-005 | Все LED первой версии используют WS2812B 60 LED/м | §4.1 | Обязательно | Сверить BOM и установленную ленту; проверить driver output | 14 | Q-002 |
| HW-006 | LED питаются от отдельного внешнего PSU, а не от Raspberry Pi | §5 | Обязательно | Проверить wiring и независимое питание LED | — | Q-002 |
| HW-007 | Pi и LED PSU имеют общую землю | §5 | Обязательно | Проверить схему и физическую continuity test до подачи питания | — | Q-002 |
| HW-008 | DATA проходит преобразование уровней 3.3V → 5V | §5 | Обязательно | Проверить наличие и работу level shifter на собранном стенде | — | Q-002 |
| HW-009 | Для level shifting выбран 74AHCT125 либо 74HCT245 | §5 | Рекомендация ТЗ | Сопоставить BOM и фактический компонент либо записать обоснованную замену | — | Q-002 |
| HW-010 | На DATA предусмотрен последовательный резистор, на LED-питании — конденсатор | §5 | Уточнить обязательность | Осмотр и сверка схемы с фактической сборкой; статус обязательности уточнить по Q-002 | — | Q-002 |
| HW-011 | LED-питание имеет аппаратный предохранитель и физическое отключение | §5 | Уточнить обязательность | Проверить установку предохранителя и отключение питания LED независимо от ПО | — | Q-002 |
| HW-012 | Для длинных 5V линий предусмотрен power injection в нескольких точках | §5 | Обязательно | Проверить расчёт длины/падения напряжения и wiring под рабочей нагрузкой | — | Q-002 |

## THERM. Температура и охлаждение

Этап §51: **12**. Рабочие окна: **Д007; Д132**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| THERM-001 | Thermal Manager читает /sys/class/thermal/thermal_zone0/temp | §3.2 | Обязательно | Сравнить нормализованное значение с sysfs; смоделировать read failure | — | Q-004 |
| THERM-002 | Thermal Manager формирует NORMAL/WARM/HOT/CRITICAL | §3.2 | Обязательно | Подать значения вокруг согласованных порогов и проверить переходы | — | Q-004 |
| THERM-003 | Температурные пороги конфигурируются | §3.2 | Обязательно | Изменить config и проверить применение; ориентиры fan off <50, on 50–65/>65, warning >75°C не считать финальными границами всех состояний | — | Q-004 |

## LIGHT. Зоны, цвет, эффекты и presets

Этап §51: **2**. Рабочие окна: **Д026–Д035; Д051–Д055; Д060**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| LIGHT-001 | Есть независимая monitor.backlight по четырём сторонам монитора | §4.2 | Обязательно | Изменить mode/brightness зоны и проверить неизменность desk.wall | 13, 14 | Q-002 |
| LIGHT-002 | Есть независимая desk.wall вдоль стола со стороны стены | §4.2 | Обязательно | Изменить её состояние без изменения monitor.backlight | 9 | Q-002 |
| LIGHT-003 | Количество зон расширяется без изменения business logic | §4.2 | Обязательно | Добавить третью fake zone только конфигурацией/registry и выполнить commands | — | — |
| LIGHT-004 | Зона имеет unique ID, mode, brightness, master modifier, effect params, power limit, mapping, status, fallback preset | §4.3 | Обязательно | Проверить roundtrip модели/config/API и изоляцию настроек разных зон | — | Q-005 |
| LIGHT-005 | Поддерживаются RGB, HSV и ручной выбор цвета | §4.5 | Обязательно | Задать эквивалентные RGB/HSV цвета и сравнить ожидаемый frame; проверить UI color picker | — | Q-005 |
| LIGHT-006 | Поддерживается эквивалент color temperature и тёплый/холодный свет | §4.5 | Обязательно | Проверить монотонное изменение RGB по выбранной шкале и UI controls | — | Q-005 |
| LIGHT-007 | Переключение режимов плавное с configurable transition_duration_ms | §4.6 | Обязательно | Проверить промежуточные frames и длительность перехода с fake clock | 17 | Q-005 |
| LIGHT-008 | Preset хранит настройки нескольких зон и редактируется пользователем | §24.5 | Обязательно | Создать/сохранить/применить preset на две зоны и проверить roundtrip | 30 | — |
| LIGHT-009 | Night Shift даёт equivalent 2200–2700K, brightness 10–20%, reduced blue и отключённые animations | §14 | Обязательно | Проверить default RGB/brightness/static frames по согласованному tolerance | 24, 29 | Q-005 |
| LIGHT-010 | Night Shift использует fade 1–2 секунды, параметры режима настраиваются | §14 | Обязательно | Проверить default duration и изменение temperature/brightness/fade пользователем | 24, 29 | Q-005 |
| LIGHT-011 | Lighting Engine поддерживает режим OFF | §4.4 | Обязательно | Выбрать OFF через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | — | — |
| LIGHT-012 | Lighting Engine поддерживает режим STATIC | §4.4 | Обязательно | Выбрать STATIC через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | — | — |
| LIGHT-013 | Lighting Engine поддерживает режим BREATHING | §4.4 | Обязательно | Выбрать BREATHING через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | — | — |
| LIGHT-014 | Lighting Engine поддерживает режим RAINBOW | §4.4 | Обязательно | Выбрать RAINBOW через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | — | — |
| LIGHT-015 | Lighting Engine поддерживает режим MUSIC_PULSE | §4.4 | Обязательно | Выбрать MUSIC_PULSE через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 9 | — |
| LIGHT-016 | Lighting Engine поддерживает режим MUSIC_BEAT | §4.4 | Обязательно | Выбрать MUSIC_BEAT через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 9 | — |
| LIGHT-017 | Lighting Engine поддерживает режим MUSIC_SPECTRUM | §4.4 | Обязательно | Выбрать MUSIC_SPECTRUM через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 9 | — |
| LIGHT-018 | Lighting Engine поддерживает режим MUSIC_WAVE | §4.4 | Обязательно | Выбрать MUSIC_WAVE через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 9 | — |
| LIGHT-019 | Lighting Engine поддерживает режим MUSIC_CENTER | §4.4 | Обязательно | Выбрать MUSIC_CENTER через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 9 | — |
| LIGHT-020 | Lighting Engine поддерживает режим AMBILIGHT | §4.4 | Обязательно | Выбрать AMBILIGHT через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 13 | — |
| LIGHT-021 | Lighting Engine поддерживает режим NIGHT_SHIFT | §4.4 | Обязательно | Выбрать NIGHT_SHIFT через API/UI; сравнить frames и lifecycle с ожидаемым поведением режима | 24, 29 | — |

## POWER. Программный ограничитель тока

Этап §51: **2**. Рабочие окна: **Д028–Д029; Д034**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| POWER-001 | Для каждой зоны задаётся max_current_amp | §6 | Обязательно | Сконфигурировать разные лимиты и проверить независимое масштабирование кадров | — | Q-002 |
| POWER-002 | Задаётся global_max_current_amp всего PSU | §6 | Обязательно | Запросить кадры двух зон с суммой выше общего лимита и проверить итоговую сумму | — | Q-002 |
| POWER-003 | До отправки кадра ток оценивается по RGB | §6 | Обязательно | Проверить numeric fixtures black/white/primary colors и калибровку на стенде | — | Q-002 |
| POWER-004 | При превышении лимита brightness всего затронутого кадра пропорционально уменьшается | §6 | Обязательно | Проверить scale=limit/requested и сохранение пропорций RGB с учётом округления | — | Q-002 |
| POWER-005 | Zone/global limiter действует независимо от эффекта | §6 | Обязательно | Прогнать все 11 modes и transitions через один финальный limiter | — | — |

## AUDIO. Источники, выходы и управление

Этап §51: **3**. Рабочие окна: **Д036–Д045**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| AUDIO-001 | Поддерживается Bluetooth A2DP input | §9.1 | Обязательно | Подключить телефон и воспроизвести тестовый звук через Hub | 6, 7 | Q-007 |
| AUDIO-002 | Поддерживается Spotify Connect | §9.1 | Обязательно | Найти Hub в разрешённом Spotify сценарии и выполнить playback/control | 7 | Q-008 |
| AUDIO-003 | Поддерживаются локальные аудиофайлы | §9.1 | Обязательно | Воспроизвести fixture с диска без WAN | 7, 48 | Q-007 |
| AUDIO-004 | Bluetooth имя устройства — Smart Home Hub | §9.2 | Обязательно | Проверить discovery с телефона | 6 | — |
| AUDIO-005 | Новые Bluetooth pairing разрешены только в явно включённом PAIRING_MODE | §9.2 | Обязательно | Попытаться pair до/во время/после окна и проверить запрет вне режима | 6 | Q-009 |
| AUDIO-006 | Поддерживается 3.5 mm output | §9.3 | Обязательно | Воспроизвести звук на реальной Pi через аналоговый выход | 8 | Q-007 |
| AUDIO-007 | Поддерживается USB Audio/USB DAC output | §9.3 | Обязательно | Выбрать подключённый DAC и проверить звук | 8 | Q-007 |
| AUDIO-008 | Поддерживается Bluetooth output | §9.3 | Обязательно | Подключить Bluetooth output и воспроизвести звук | 8 | Q-007 |
| AUDIO-009 | Audio output выбирается в Web UI | §9.3 | Обязательно | Сменить output с телефона и сверить фактическую маршрутизацию | 8 | Q-007 |
| AUDIO-010 | При наличии source metadata отображаются track, artist, album, cover, progress | §9.5 | Обязательно | Проверить источник с полными/частичными/отсутствующими metadata | — | Q-007 |
| AUDIO-011 | История прослушивания не сохраняется | §9.5,52 | Обязательно | Сравнить БД/logs/files до/после playback; убедиться, что track metadata не попадают в habit records | — | Q-021 |
| AUDIO-012 | Поддерживается audio command Play | §9.4 | Обязательно | Вызвать Play на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |
| AUDIO-013 | Поддерживается audio command Pause | §9.4 | Обязательно | Вызвать Pause на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |
| AUDIO-014 | Поддерживается audio command Next | §9.4 | Обязательно | Вызвать Next на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |
| AUDIO-015 | Поддерживается audio command Previous | §9.4 | Обязательно | Вызвать Previous на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |
| AUDIO-016 | Поддерживается audio command Volume | §9.4 | Обязательно | Вызвать Volume на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |
| AUDIO-017 | Поддерживается audio command Mute | §9.4 | Обязательно | Вызвать Mute на каждом поддерживаемом source; проверить фактическое действие и API state, ограничения source фиксировать явно | 7 | Q-007 |

## DSP. Анализ аудио и музыкальные эффекты

Этап §51: **4**. Рабочие окна: **Д037; Д046–Д055**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| DSP-001 | DSP вычисляет overall volume, bass, mids, treble, beat и BPM | §10 | Обязательно | Прогнать synthetic tones/metronome/silence и сверить численные features | 9, 10 | Q-033 |
| DSP-002 | Анализ использует FFT и beat detection | §10 | Обязательно | Проверить algorithm fixtures и expected spectrum/onset results | 9 | Q-033 |
| DSP-003 | Работает автоматическая нормализация чувствительности | §10 | Обязательно | Сравнить один сигнал с разными gain и временем адаптации | 9 | Q-033 |
| DSP-004 | Пользователь настраивает sensitivity, smoothing, brightness и effect | §10 | Обязательно | Изменить каждый параметр и проверить ожидаемое изменение output | 9 | Q-033 |
| DSP-005 | MUSIC_PULSE преобразует громкость в общую brightness | §10 | Обязательно | Подать ступенчатый volume и сравнить frame brightness | 9 | — |
| DSP-006 | MUSIC_BEAT создаёт flash/pulse на каждый detected beat | §10 | Обязательно | Подать deterministic beats и проверить импульсы | 9 | Q-033 |
| DSP-007 | MUSIC_SPECTRUM отображает bass/mids/treble разными участками или цветами | §10 | Обязательно | Подать поочерёдно частоты полос и проверить mapping | 9 | Q-033 |
| DSP-008 | MUSIC_WAVE двигает музыкальные импульсы вдоль ленты | §10 | Обязательно | Сравнить golden frames во времени | 9 | — |
| DSP-009 | MUSIC_CENTER разводит импульс из центра в две стороны | §10 | Обязательно | Проверить symmetry на чётной/нечётной длине | 9 | — |
| DSP-010 | Web UI показывает realtime bass/mids/treble/BPM/beat | §10 | Обязательно | Сопоставить WebSocket payload и видимые indicators с DSP fixture | 10 | — |
| DSP-011 | DSP/spectrum/FFT/beat history не сохраняется | §10,31 | Обязательно | Проверить disk/DB/log activity при DSP и после restart | — | Q-021 |

## AMBI. Screen-client и Ambilight

Этап §51: **5**. Рабочие окна: **Д056–Д070**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| AMBI-001 | Screen-client реализован на Go и работает на пользовательском ноутбуке | §11.1 | Обязательно | Проверить source/build и запуск вне Pi | 11 | Q-006 |
| AMBI-002 | Поддерживается Linux x86_64 | §11.1 | Обязательно | Установить клиент на реальном Linux и пройти pairing/capture/stream | 11 | Q-011 |
| AMBI-003 | Поддерживается Windows x86_64 | §11.1 | Обязательно | Установить клиент на Windows и пройти pairing/capture/stream | 11 | Q-011 |
| AMBI-004 | Поддерживается macOS arm64 | §11.1 | Обязательно | Установить клиент на Apple Silicon и пройти permissions/pairing/stream | 11 | Q-011 |
| AMBI-005 | Поддерживается macOS x86_64 | §11.1 | Обязательно | Установить клиент на Intel Mac и пройти permissions/pairing/stream | 11 | Q-011 |
| AMBI-006 | Поддерживается внешний монитор через HDMI либо USB-C/HDMI | §11.1 | Обязательно | Проверить выбор внешнего display и соответствие LED его изображению | 14 | Q-011 |
| AMBI-007 | Кадры экрана анализируются только на ноутбуке; Hub не получает изображения | §11.2,40 | Обязательно | Проверить network payload и Hub API: только colors/technical metadata/scene results | 14 | Q-021 |
| AMBI-008 | Runtime screenshots и Ambilight frame history не сохраняются | §11.2,31 | Обязательно | Проверить filesystem/DB/log activity на клиенте и Hub | — | Q-021 |
| AMBI-009 | Screen-client pairing выполняется через Web UI, одноразовый code и выдачу device token | §11.3 | Обязательно | Пройти полный flow; повторное использование code отклоняется | 12 | Q-012 |
| AMBI-010 | Geometry задаёт counts top/right/bottom/left, start_corner и direction | §11.4 | Обязательно | Проверить все corners, clockwise/counter_clockwise и разные counts golden tests | 14 | Q-002 |
| AMBI-011 | Цвет вычисляется weighted average с большим весом у края экрана | §11.5 | Обязательно | Сравнить результат синтетической полосы с ручным расчётом Σcolor×weight/Σweight | 14 | Q-005 |
| AMBI-012 | Поддерживается black-bar detection | §11.6 | Обязательно | Подать fixtures с letterbox/pillarbox и проверить границы анализа | 14 | Q-005 |
| AMBI-013 | Поддерживается suppression почти чёрных пикселей | §11.6 | Обязательно | Проверить near-black fixtures и согласованный threshold | 14 | Q-005 |
| AMBI-014 | Настраиваются saturation, gamma и color boost | §11.6 | Обязательно | Проверить каждую коррекцию отдельно и порядок их композиции | 14 | Q-005 |
| AMBI-015 | Настраиваются smoothing и min/max brightness | §11.6 | Обязательно | Проверить temporal step и clamp на fixtures | 17 | Q-005 |
| AMBI-016 | Update rate конфигурируется | §11.6 | Обязательно | Изменить rate и измерить фактический поток цветов | 15 | Q-013 |
| AMBI-017 | Обязателен Cinema Mode с высоким smoothing, мягкими переходами и сниженным color boost | §11.8 | Обязательно | Сравнить параметры профиля и реакцию на тестовую смену сцены | 17 | Q-005 |
| AMBI-018 | Gaming Mode снижает smoothing/latency и ускоряет реакцию | §11.8 | Опционально | Сравнить latency и профиль с Cinema на одном стенде | — | Q-013 |
| AMBI-019 | При потере screen-client monitor.backlight переходит в заданный fallback preset | §11.9 | Обязательно | Оборвать stream и сверить итоговый preset после согласованного timeout | 20 | Q-014 |
| AMBI-020 | Клиент определяет выключенный монитор | §11.10 | Обязательно | Выключить/отключить monitor и проверить inactivity event | — | Q-014 |
| AMBI-021 | Клиент определяет длительно статичное изображение и при необходимости включает fallback/off | §11.10 | Обязательно | Подать frozen fixture с fake clock и проверить порог/возврат к активности | — | Q-014 |

## SCENE. Локальная классификация сцен

Этап §51: **6**. Рабочие окна: **Д071–Д080**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| SCENE-001 | Multi-label classifier определяет CALM/DYNAMIC/DARK/BRIGHT/HIGH_MOTION/LOW_MOTION/COLORFUL/LOW_CONTRAST/HIGH_CONTRAST | §12 | Обязательно | Оценить каждый класс и комбинации на независимом labelled holdout | 18 | Q-015 |
| SCENE-002 | Scene ML выполняется на ноутбуке в составе screen-client | §12.1 | Обязательно | Проверить process/runtime placement и отсутствие remote inference | 18 | Q-006 |
| SCENE-003 | Scene ML анализирует последовательность нескольких кадров | §12.2 | Обязательно | Подать одинаковый последний frame с разной motion history и проверить temporal features | 18 | Q-015 |
| SCENE-004 | ML-анализ выполняется примерно раз в секунду независимо от Ambilight 30 FPS | §12.2 | Обязательно | Посчитать inference и применённые LED frames за фиксированный интервал | 15, 18 | Q-013 |
| SCENE-005 | Scene ML не переключает lighting modes | §12.3 | Обязательно | Подать все labels и убедиться, что zone mode остаётся AMBILIGHT | 18 | — |
| SCENE-006 | Scene ML корректирует только brightness/saturation/smoothing/color_boost/transition_speed | §12.3 | Обязательно | Проверить allowed fields и label→parameter contract | 18 | Q-015 |
| SCENE-007 | Пользователь задаёт ml_influence 0–100%; 0% даёт обычный Ambilight | §12.3 | Обязательно | Сравнить output при disabled и 0%; проверить endpoints диапазона | 18, 19 | Q-015 |
| SCENE-008 | Scene classification history не хранится | §12.3 | Обязательно | Проверить DB/files/logs после потока labels и restart | — | Q-021 |

## GEST. Жесты и камера

Этап §51: **7**. Рабочие окна: **Д081–Д090**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| GEST-001 | Поддерживаются USB webcam и Raspberry Pi Camera Module | §13.1 | Обязательно | Пройти start/capture/reconnect на каждом типе камеры отдельно | 21 | Q-016 |
| GEST-002 | Gesture capture включается только при Gesture Mode | §13.2 | Обязательно | Выключить Gesture Mode и убедиться, что gesture consumer не захватывает кадры | 21 | Q-017 |
| GEST-003 | Жест требует confirmation 300–500 мс | §13.4 | Обязательно | Подать короткую/устойчивую последовательность с fake clock и проверить action | 22, 23, 24, 25 | Q-018 |
| GEST-004 | После жеста действует cooldown примерно 1000 мс | §13.4 | Обязательно | Повторять жест до/после threshold и проверить число actions | 22, 23, 24, 25 | Q-018 |
| GEST-005 | Activation gesture открывает ограниченное окно управления | §13.4 | Обязательно | Проверить action до активации, внутри окна и после expiry | 21 | Q-018 |
| GEST-006 | Recognition работает по всей области кадра | §13.4 | Обязательно | Повторить жест в центре/по углам/у краёв на согласованном стенде | — | Q-018 |
| GEST-007 | Кадры камеры обрабатываются только в RAM, не сохраняются и не отправляются внешним сервисам | §13.5,40 | Обязательно | Проверить filesystem/DB/logs/network при capture, exception и worker crash | 26 | Q-021 |
| GEST-008 | UI показывает recognition ON, detected gesture, confidence и resolved action | §13.6 | Обязательно | Сопоставить показанные значения с injected gesture events | 21 | — |
| GEST-009 | Пользователь редактирует gesture→action mapping | §13.6 | Обязательно | Переназначить gesture, перезапустить Core и проверить сохранение/исполнение | 22, 23, 24, 25 | Q-018 |
| GEST-010 | Default mapping OPEN_PALM → Play/Pause | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить OPEN_PALM на Pi; проверить ровно ожидаемое Play/Pause | 22 | Q-018 |
| GEST-011 | Default mapping SWIPE_RIGHT → Next Track | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить SWIPE_RIGHT на Pi; проверить ровно ожидаемое Next Track | 23 | Q-018 |
| GEST-012 | Default mapping SWIPE_LEFT → Previous Track | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить SWIPE_LEFT на Pi; проверить ровно ожидаемое Previous Track | 23 | Q-018 |
| GEST-013 | Default mapping THUMB_UP → Brightness + | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить THUMB_UP на Pi; проверить ровно ожидаемое Brightness + | — | Q-018 |
| GEST-014 | Default mapping THUMB_DOWN → Brightness - | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить THUMB_DOWN на Pi; проверить ровно ожидаемое Brightness - | — | Q-018 |
| GEST-015 | Default mapping FIST → Lights OFF | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить FIST на Pi; проверить ровно ожидаемое Lights OFF | — | Q-018 |
| GEST-016 | Default mapping TWO_FINGERS → Next Lighting Preset | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить TWO_FINGERS на Pi; проверить ровно ожидаемое Next Lighting Preset | — | Q-018 |
| GEST-017 | Default mapping PALM_FACE_DOWN + downward movement → Night Shift | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить PALM_FACE_DOWN + downward movement на Pi; проверить ровно ожидаемое Night Shift | 24 | Q-018 |
| GEST-018 | Default mapping ROTATE_CLOCKWISE → Volume + | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить ROTATE_CLOCKWISE на Pi; проверить ровно ожидаемое Volume + | 25 | Q-018 |
| GEST-019 | Default mapping ROTATE_COUNTER_CLOCKWISE → Volume - | §13.3 | Обязательно | Подать positive/negative temporal fixtures и выполнить ROTATE_COUNTER_CLOCKWISE на Pi; проверить ровно ожидаемое Volume - | 25 | Q-018 |

## PRES. Присутствие

Этап §51: **8**. Рабочие окна: **Д096–Д100**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| PRES-001 | Presence Engine объединяет camera person/Bluetooth phone/network/PIR | §15 | Обязательно | Прогнать timeline с каждым источником и конфликтующими комбинациями | 40 | Q-019 |
| PRES-002 | Итоговые состояния — PRESENT/PROBABLY_PRESENT/AWAY/UNKNOWN | §15 | Обязательно | Проверить переходы и различие unknown/stale/confirmed away | 40 | Q-019 |
| PRES-003 | Presence CV использует ту же камеру, что gesture recognition | §15.1 | Обязательно | Одновременно включить consumers и проверить единый lifecycle физической камеры | 40 | Q-017 |
| PRES-004 | CV определяет только person yes/no без идентификации личности | §15.1 | Обязательно | Проверить outputs/schema и отсутствие identity model/face matching | 40 | — |
| PRES-005 | Person inference выполняется каждые 2–3 секунды | §15.1 | Обязательно | Измерить intervals при совместном gesture/audio/LED load | 40 | Q-013 |
| PRES-006 | При любой активной CV обработке включён отдельный LED-индикатор на Pi | §15.2 | Обязательно | Проверить gesture-only, presence-only, оба режима, stop и crash | 27 | Q-017 |

## SENS. Сенсоры и ambient policy

Этап §51: **8**. Рабочие окна: **Д091–Д095**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| SENS-001 | Поддерживается ambient light sensor | §16 | Обязательно | Подключить выбранный sensor и сравнить readings при изменении света | 31 | Q-020 |
| SENS-002 | Поддерживается temperature sensor | §16 | Обязательно | Проверить реальные readings, units и invalid/stale case | 31 | Q-020 |
| SENS-003 | Поддерживается humidity sensor | §16 | Обязательно | Проверить реальные readings, units и invalid/stale case | 31 | Q-020 |
| SENS-004 | Поддерживается PIR motion sensor | §16 | Обязательно | Проверить motion/no-motion и disconnect | 31 | Q-020 |
| SENS-005 | Поддерживается air quality sensor | §16 | Обязательно | Проверить readings и физический смысл выбранных units/index | 31 | Q-020 |
| SENS-006 | Hardware adapters допускают GPIO/I²C/SPI согласно выбранным устройствам | §16,21 | Обязательно | Проверить abstraction и выбранные реальные bus drivers; не считать обязательной поддержку всех моделей | 31 | Q-020 |
| SENS-007 | Рост room brightness уменьшает LED brightness, снижение — увеличивает | §16.1 | Обязательно | Подать возрастающие/убывающие lux fixtures и проверить направление/modifier | 32 | Q-020 |
| SENS-008 | Temperature/humidity/air quality видны в UI и доступны Automation Engine | §16.2 | Обязательно | Сопоставить UI/API readings и выполнить rule по каждому типу значения | 31, 33 | Q-020 |
| SENS-009 | Automation не меняет активную музыкальную сцену без явной настройки правила | §16.2 | Обязательно | Запустить music mode, вызвать sensor rule с разрешением/без и сравнить actions | 35, 38 | Q-023 |

## HAB. Локальные привычки и рекомендации

Этап §51: **10**. Рабочие окна: **Д035; Д111–Д120**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| HAB-001 | Наблюдения содержат timestamp/day_of_week/light_zone/light_mode/brightness/selected_preset/music_active/tv_active/ambilight_active/gesture_used/manual_override/state_duration | §17.1 | Обязательно | Проверить schema и collector fixtures по каждому полю; отсутствие данных не выдумывать | 41, 42 | Q-022 |
| HAB-002 | MVP не моделирует последовательности действий между событиями | §17.1 | Обязательно | Проверить features/model inputs и отсутствие обязательного sequence pipeline | — | — |
| HAB-003 | Predictor предлагает preset, brightness, lighting mode, Ambilight, Night Shift, music start и volume | §17.2 | Обязательно | Для каждого target подготовить local training fixture и проверить допустимый candidate | 42 | Q-022 |
| HAB-004 | Предиктивная автоматизация включается после минимум 30 дней сбора | §17.3 | Обязательно | Проверить boundary fake clock и отдельно фактические real observation timestamps | 42 | Q-022 |
| HAB-005 | Поддерживаются SUGGEST и AUTO | §17.4 | Обязательно | Проверить suggestion без side effect и AUTO через command layer | 42, 43 | Q-023 |
| HAB-006 | Confidence policy имеет настраиваемые auto/suggest/no-action thresholds | §17.4 | Обязательно | Проверить below/equal/above пороги; 60/85% из ТЗ — пример до решения Q-022 | 43 | Q-022 |
| HAB-007 | Учитываются время/день/monitor/music/mode/brightness/gestures/presence/ambient light | §17.5 | Обязательно | Проверить feature extraction и значимость/доступность полей на local dataset | 41, 42 | Q-022 |
| HAB-008 | Пользователь может принять recommendation | §17.6 | Обязательно | Принять recommendation и проверить разрешённое action и feedback | 42 | Q-022 |
| HAB-009 | Пользователь может отклонить recommendation; отказ используется как negative feedback | §17.6 | Обязательно | Отклонить предложение и проверить feedback/training input без action | 44 | Q-022 |
| HAB-010 | Ручная отмена уже выполненной automation не считается feedback | §17.6 | Обязательно | Выполнить AUTO, вручную отменить и убедиться в отсутствии feedback record | 44 | — |
| HAB-011 | Habit observations, training и inference остаются локальными без third-party sync/cloud APIs | §17.7,40 | Обязательно | Проверить process placement, outbound traffic и запрет выгрузки истории | 41 | Q-021 |
| HAB-012 | История habits использует rolling window 90 дней | §17.8 | Обязательно | Проверить cleanup на границе retention и сохранение свежих records | — | Q-022 |
| HAB-013 | Delete learning history удаляет observations, feedback и обученное user state/model | §17.9 | Обязательно | Удалить историю при активном training и проверить storage/caches/jobs | 45 | Q-021 |
| HAB-014 | После удаления начинается новый 30-day learning period | §17.9 | Обязательно | Проверить reset timestamp, no-action до eligibility и повторный learning gate | 45 | Q-022 |

## AUTO. Правила, приоритеты и override

Этап §51: **9**. Рабочие окна: **Д032; Д101–Д110**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| AUTO-001 | Rules имеют IF conditions и THEN actions | §18 | Обязательно | Создать rule через API и выполнить соответствующее событие | 35 | Q-023 |
| AUTO-002 | Поддерживается AND | §18.2 | Обязательно | Проверить truth table двух conditions | 34 | — |
| AUTO-003 | Поддерживается OR | §18.2 | Обязательно | Проверить truth table двух conditions | 34 | — |
| AUTO-004 | Поддерживаются nested conditions | §18.2 | Обязательно | Проверить nested AND/OR с разными truth branches | 34 | Q-023 |
| AUTO-005 | Поддерживаются delayed actions | §18.3 | Обязательно | Пройти absent→wait 5min→OFF; проверить срок fake clock и реальное действие | 36 | Q-023 |
| AUTO-006 | Поддерживаются регулярные time/weekday schedules | §18.4 | Обязательно | Проверить weekday 07:30 и daily 23:30; timezone/clock policy решить отдельно | 35 | Q-023 |
| AUTO-007 | Сценарий может содержать несколько actions на разные зоны/audio | §18.5 | Обязательно | Выполнить Movie Mode с Ambilight, warm desk и music pause | 35 | Q-023 |
| AUTO-008 | Rule имеет priority; при конфликте побеждает более высокий | §18.6 | Обязательно | Активировать два rules на один resource и проверить победителя | 37 | Q-023 |
| AUTO-009 | Ручное действие создаёт override затронутого состояния; default 15 минут | §18.7 | Обязательно | Проверить scope, expiry и automation до/после 15min fake clock | 38 | Q-023 |
| AUTO-010 | Automation audit хранит rule/trigger/timestamp/actions/result/error | §18.8 | Обязательно | Пройти successful/failed run и проверить запись всех полей | 39 | Q-023 |
| AUTO-011 | Automation audit retention составляет 30 дней | §18.8 | Обязательно | Проверить cleanup старых и сохранение новых entries | 39 | Q-023 |
| AUTO-012 | Manual Lights OFF выключает все LED-зоны и создаёт override против немедленного включения automation | §49 | Обязательно | Выполнить Master OFF при активных rules и проверить все зоны/TTL | 38 | Q-023 |
| AUTO-013 | Rules поддерживают condition/trigger time | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные time; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-014 | Rules поддерживают condition/trigger day of week | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные day of week; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-015 | Rules поддерживают condition/trigger user presence | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные user presence; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-016 | Rules поддерживают condition/trigger monitor on/off | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные monitor on/off; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-017 | Rules поддерживают condition/trigger music playing | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные music playing; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-018 | Rules поддерживают condition/trigger Ambilight active | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные Ambilight active; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-019 | Rules поддерживают condition/trigger brightness | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные brightness; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-020 | Rules поддерживают condition/trigger ambient light | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные ambient light; проверить matched/skipped по согласованной semantics | 33, 35 | Q-023 |
| AUTO-021 | Rules поддерживают condition/trigger gesture event | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные gesture event; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-022 | Rules поддерживают condition/trigger ML event | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные ML event; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-023 | Rules поддерживают condition/trigger MQTT event | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные MQTT event; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-024 | Rules поддерживают condition/trigger device state | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные device state; проверить matched/skipped по согласованной semantics | 35 | Q-023 |
| AUTO-025 | Rules поддерживают condition/trigger sensor values | §18.1 | Обязательно | Подать подходящие/неподходящие и unavailable данные sensor values; проверить matched/skipped по согласованной semantics | 33, 35 | Q-023 |

## BTN. Физическая кнопка

Этап §51: **2**. Рабочие окна: **Д033–Д035**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| BTN-001 | В MVP есть минимум одна физическая кнопка | §19 | Обязательно | Проверить реальный GPIO input на Pi и simulation equivalent | 28, 29, 30 | Q-002 |
| BTN-002 | Single press по умолчанию переключает Lights ON/OFF | §19 | Обязательно | Проверить press sequence и отсутствие duplicate actions | 28 | Q-023 |
| BTN-003 | Double press по умолчанию включает Night Shift | §19 | Обязательно | Проверить double press без лишнего single action | 29 | — |
| BTN-004 | Long press по умолчанию выбирает next preset | §19 | Обязательно | Проверить long press и порядок presets | 30 | Q-005 |
| BTN-005 | Button mappings хранятся SQLite и редактируются Web UI | §19 | Обязательно | Переназначить три типа press, restart и повторно проверить | 28, 29, 30 | — |

## DEV. Устройства и матрица

Этап §51: **2**. Рабочие окна: **Д026; Д094; Д122**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| DEV-001 | LED Device поддерживает STRIP и MATRIX | §20 | Обязательно | Выполнить единый lighting API с virtual strip и virtual matrix | — | — |
| DEV-002 | MATRIX имеет width/height/orientation/serpentine_layout | §20 | Обязательно | Проверить table/golden mapping разных layouts и размеров | — | — |
| DEV-003 | Реальная LED-панель добавляется после логической поддержки | §20 | Опционально | Прогнать hardware mapping на купленной панели при наличии | — | Q-002 |
| DEV-004 | Device хранит id/type/status/capabilities/metadata | §21 | Обязательно | Проверить schema/API roundtrip и capability-based commands | 5, 46 | — |
| DEV-005 | Device статусы — ONLINE/OFFLINE/ERROR/DISABLED | §21 | Обязательно | Проверить переходы с real/fake disconnect/error/disable | 5, 46 | Q-030 |
| DEV-006 | Устройства добавляются вручную | §21 | Обязательно | Добавить device в UI/API и проверить registry persistence | 5 | — |
| DEV-007 | Поддерживается Bluetooth discovery и pairing | §21 | Обязательно | Найти устройство, явно разрешить pair и проверить registry | 6 | Q-009 |

## MQTT. MQTT и Home Assistant readiness

Этап §51: **11**. Рабочие окна: **Д121–Д123**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| MQTT-001 | Используется Mosquitto либо эквивалентный MQTT broker | §22 | Обязательно | Проверить запуск broker и подключения Core/test device | 46 | Q-025 |
| MQTT-002 | Anonymous MQTT access запрещён | §22 | Обязательно | Попытаться подключиться без credentials и получить отказ | — | — |
| MQTT-003 | MQTT использует username/password и ACL | §22 | Обязательно | Проверить valid/invalid login и запрет publish/subscribe на чужие topics | — | Q-025 |
| MQTT-004 | Broker доступен только LAN/Docker network | §22 | Обязательно | Проверить bind/firewall/container network и отсутствие WAN exposure | — | Q-025 |
| MQTT-005 | MQTT устройства обнаруживаются автоматически | §21,22 | Обязательно | Опубликовать discovery fixture, повторить её и проверить один Device | 46 | Q-025 |
| MQTT-006 | MQTT поддерживает device state/set, events и sensor state namespaces | §22 | Обязательно | Проверить согласованные topics внешним test client; конкретные строки §22 являются примерами | 46, 49 | Q-025 |
| MQTT-007 | HA может подключаться через MQTT/REST без изменения Core | §23,48 | Обязательно | Выполнить read/control внешним клиентом по опубликованным contracts | 49 | Q-025 |
| MQTT-008 | Core не зависит от установленного Home Assistant | §23,48 | Обязательно | Запустить и проверить Hub при отсутствии HA | 49 | — |
| MQTT-009 | Home Assistant MQTT Discovery добавляется в будущем | §48 | Опционально | При отдельном включении проверить discovery integration на HA-стенде | — | Q-025 |

## WEB. Web UI и PWA

Этап §51: **1**. Рабочие окна: **Д022–Д023; интерфейсы внутри этапов; Д156**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| WEB-001 | Frontend использует React/TypeScript/Vite | §24 | Обязательно | Проверить manifests, typecheck и production build | — | — |
| WEB-002 | UI responsive и ориентирован на телефон | §24 | Обязательно | Пройти основные flows на согласованной mobile/desktop viewport matrix | 5 | Q-029 |
| WEB-003 | Web UI поддерживает PWA | §24 | Обязательно | Проверить manifest, установку и согласованное offline/cache поведение по HTTPS | — | Q-029 |
| WEB-004 | Dashboard показывает system status, CPU/RAM, audio playing/volume, две зоны, automations enabled/last trigger | §24.1 | Обязательно | Сопоставить каждый indicator с injected state и реальной Pi telemetry | 5 | — |
| WEB-005 | Realtime события доступны по /ws/events | §24.2 | Обязательно | Проверить authenticated WS stream и восстановление соединения | 10 | Q-030 |
| WEB-006 | WS передаёт audio spectrum/beat/device state/lighting state/gesture/system metrics/automation event | §24.2 | Обязательно | Инжектировать каждый тип и проверить соответствующее обновление UI | 5, 10, 39 | Q-030 |
| WEB-007 | UI визуализирует текущее состояние LED | §24.3 | Обязательно | Сравнить virtual strip/matrix preview с применённым frame | 14 | Q-005 |
| WEB-008 | Ambilight preview показывает только color frame zones без screenshot | §24.4 | Обязательно | Проверить browser network/UI и отсутствие image payload | 14 | Q-021 |
| WEB-009 | Есть визуальный IF/AND/OR/THEN Rule Builder | §24.6 | Обязательно | Создать nested rule на телефоне и выполнить dry test/реальный trigger | 34 | Q-023 |
| WEB-010 | В UI доступна страница Dashboard | §24 | Обязательно | Открыть Dashboard, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | 5 | Q-029 |
| WEB-011 | В UI доступна страница Lighting | §24 | Обязательно | Открыть Lighting, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-012 | В UI доступна страница Audio | §24 | Обязательно | Открыть Audio, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-013 | В UI доступна страница Ambilight | §24 | Обязательно | Открыть Ambilight, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-014 | В UI доступна страница Gestures | §24 | Обязательно | Открыть Gestures, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-015 | В UI доступна страница Automations | §24 | Обязательно | Открыть Automations, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-016 | В UI доступна страница Predictions | §24 | Обязательно | Открыть Predictions, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-017 | В UI доступна страница Devices | §24 | Обязательно | Открыть Devices, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-018 | В UI доступна страница System | §24 | Обязательно | Открыть System, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |
| WEB-019 | В UI доступна страница Settings | §24 | Обязательно | Открыть Settings, проверить authorized navigation, loading/error и соответствующие функции её подсистемы | — | Q-029 |

## TG. Telegram Bot

Этап §51: **11**. Рабочие окна: **Д124–Д125**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| TG-001 | Bot входит в MVP как отдельный Go service/container | §25 | Обязательно | Проверить build/Compose и запуск отдельно от Core | 47 | Q-027 |
| TG-002 | Bot использует только Core API, без прямого hardware/storage доступа | §25 | Обязательно | Проверить imports/mounts/credentials и выполнить команды с API token | 47 | — |
| TG-003 | Bot поддерживает inline buttons | §25 | Обязательно | Нажать action buttons и проверить соответствующую scoped Core command | 47 | Q-027 |
| TG-004 | Telegram user ID явно привязан к local user | §25 | Обязательно | Проверить linked/unlinked пользователя и запрет управления без привязки | 47 | Q-027 |
| TG-005 | Через Telegram запрещена передача camera/screen frames, habit history, ML/CV data | §25 | Обязательно | Проверить responses/alerts/error handlers и network payload всех команд | — | Q-021 |
| TG-006 | Core функции работают без Telegram и WAN | §25 | Обязательно | Остановить bot/сеть и пройти local acceptance smoke | 48 | Q-001 |
| TG-007 | Bot поддерживает команду /status | §25 | Обязательно | Выполнить /status от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-008 | Bot поддерживает команду /lights | §25 | Обязательно | Выполнить /lights от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-009 | Bot поддерживает команду /music | §25 | Обязательно | Выполнить /music от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-010 | Bot поддерживает команду /presets | §25 | Обязательно | Выполнить /presets от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-011 | Bot поддерживает команду /ambilight | §25 | Обязательно | Выполнить /ambilight от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-012 | Bot поддерживает команду /night | §25 | Обязательно | Выполнить /night от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |
| TG-013 | Bot поддерживает команду /system | §25 | Обязательно | Выполнить /system от linked пользователя и проверить ожидаемую Core action/read, затем проверить отказ unlinked пользователю | 47 | Q-027 |

## AUTH. Пользователи, sessions и tokens

Этап §51: **1**. Рабочие окна: **Д017–Д020**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| AUTH-001 | Модель включает User/Role/Session; первый user имеет role admin | §26 | Обязательно | Проверить bootstrap и persistent schema; default credentials не предполагаются ТЗ | 4 | Q-024 |
| AUTH-002 | Login использует username/password | §26 | Обязательно | Проверить valid/invalid credentials и защищённый доступ | 4 | Q-024 |
| AUTH-003 | Пароли хранятся только как Argon2id hash | §26 | Обязательно | Проверить records и password verification без plaintext | 4 | Q-024 |
| AUTH-004 | Session cookies имеют HttpOnly/Secure/SameSite | §26 | Обязательно | Проверить Set-Cookie через HTTPS и browser behavior | 4 | Q-024 |
| AUTH-005 | Sessions имеют expiration | §26 | Обязательно | Проверить доступ до/после expiry fake clock | 4 | Q-024 |
| AUTH-006 | Поддерживается logout | §26 | Обязательно | Завершить session и проверить отказ повторного запроса с прежним cookie | 4 | — |
| AUTH-007 | Поддерживается revoke all sessions | §26 | Обязательно | Создать две sessions, отозвать и проверить обе включая WS | 4 | Q-024 |
| AUTH-008 | Machine API tokens доступны screen-client, bot, external apps, HA, CLI и IoT | §27 | Обязательно | Создать scoped tokens для каждого типа integration и проверить API access | 12, 47, 49 | Q-024 |
| AUTH-009 | Token показывается только при создании и хранится только как hash | §27 | Обязательно | Проверить create/list/detail/DB/logs без повторного вывода секрета | 12 | — |
| AUTH-010 | Tokens можно revoke | §27 | Обязательно | Отозвать token и проверить subsequent request/stream denial | 12 | Q-024 |
| AUTH-011 | Tokens имеют scopes | §27 | Обязательно | Проверить lights read/write, audio control, devices read, automation write, system read и denied scope | 47, 49 | Q-024 |

## API. API, HTTPS и сетевые границы

Этап §51: **1**. Рабочие окна: **Д009; Д013; Д020–Д023**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| API-001 | REST API имеет prefix /api/v1 | §28 | Обязательно | Проверить routing и versioned OpenAPI paths | — | — |
| API-002 | API имеет /auth,/users,/zones,/lights,/presets,/audio,/ambilight,/gestures,/automations,/predictions,/devices,/system,/config | §28 | Обязательно | Сверить schema/routers и выполнить smoke каждого ресурса; CRUD деталей уточняется | — | Q-024 |
| API-003 | API документировано OpenAPI/Swagger specification | §28 | Обязательно | Проверить schema validation и соответствие implemented operations | — | — |
| API-004 | Предусмотрен rate limiting auth/Web UI/external API | §28 | Обязательно | Проверить burst/limit/recovery для каждого класса; 10/120/60 req/min — примеры до решения | — | Q-024 |
| API-005 | Hub доступен как https://smarthome.local через mDNS | §29 | Обязательно | Открыть адрес с телефона/ноутбука после установки доверия | 3 | Q-010 |
| API-006 | HTTPS обязателен; local CA/self-signed bootstrap допускается с последующим trust install | §29 | Обязательно | Проверить TLS и инструкцию доверия без коммита private keys | 3 | Q-010 |
| API-007 | По умолчанию доступ только из LAN; API напрямую в интернет не публикуется | §30 | Обязательно | Проверить bind/proxy/firewall/router exposure | — | Q-010 |
| API-008 | Remote access при необходимости выполняется через Tailscale/WireGuard | §30 | Рекомендация ТЗ | Проверить выбранный VPN flow при отдельной настройке, без публичного Core API | — | Q-010 |

## STORE. Хранение, configuration и секреты

Этап §51: **1**. Рабочие окна: **Д011–Д012; Д016; Д021; Д137**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| STORE-001 | Основная БД — SQLite в WAL mode | §2,31 | Обязательно | Проверить PRAGMA journal_mode и concurrent read/write/restart | 53 | Q-028 |
| STORE-002 | Хранятся users/roles/sessions/zones/devices/presets/rules/audit/gesture mappings/settings/observations/feedback/token hashes | §31 | Обязательно | Проверить миграции и persistence roundtrip каждой сущности | 39, 41, 53 | Q-028 |
| STORE-003 | Не сохраняются spectrum/FFT/beat history, screen/camera frames и Ambilight frame history | §31 | Обязательно | Проверить schema и runtime file/log activity всех producers | 26 | Q-021 |
| STORE-004 | Env хранит secrets/infrastructure, YAML — hardware/infrastructure, SQLite — runtime/user settings | §32 | Обязательно | Проверить загрузку, валидацию и место записи каждого класса параметров | — | Q-028 |
| STORE-005 | Поддерживаются development/test/production profiles | §32 | Обязательно | Запустить три профиля с соответствующими examples и отсутствием production secrets в test | — | Q-028 |
| STORE-006 | В публичный repo не попадают реальные env/tokens/credentials/session secrets/keys/TLS keys/backups/personal/sensitive IP-domain данные | §33 | Обязательно | Просканировать tracked files/history/release artifacts и examples | — | Q-026 |
| STORE-007 | В repo хранятся .env.example и config.example.yaml без реальных секретов | §33 | Обязательно | Проверить clean checkout и запуск с заполненными тестовыми значениями | — | — |
| STORE-008 | CI/CD secrets находятся только в GitHub Secrets либо runner environment | §33 | Обязательно | Проверить workflow secret sources и отсутствие secret literals/artifacts | 59, 60 | Q-026 |

## SIM. Simulation без Raspberry Pi

Этап §51: **1**. Рабочие окна: **Д015; Д024–Д025; Д147**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| SIM-001 | Backend запускается без Pi/hardware libraries/devices | §2,34 | Обязательно | Запустить simulation из чистого checkout на обычном компьютере и в CI | — | — |
| SIM-002 | Есть virtual LED strip и virtual matrix | §34 | Обязательно | Применить colors/effects и сравнить expected frames обеих геометрий | — | — |
| SIM-003 | Есть fake sensors | §34 | Обязательно | Подать deterministic measurements/offline и проверить events/rules | — | — |
| SIM-004 | Есть fake audio events | §34 | Обязательно | Подать spectrum/beat/BPM и проверить reactive frames | — | — |
| SIM-005 | Есть fake MQTT devices | §34 | Обязательно | Пройти discovery/state/set/offline без реального IoT | 46 | Q-025 |
| SIM-006 | Есть fake presence | §34 | Обязательно | Проиграть state timeline и проверить automation | 40 | — |
| SIM-007 | Есть fake screen-client | §34 | Обязательно | Передать synthetic colors через тот же protocol в Hub | 14 | — |

## REC. Backup, restore и reset

Этап §51: **12**. Рабочие окна: **Д129–Д130**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| REC-001 | Автоматический backup создаётся один раз в сутки | §35 | Обязательно | Проверить расписание и фактический суточный запуск на работающем стенде | 56 | Q-028 |
| REC-002 | Хранятся последние 7 backup | §35 | Обязательно | Создать 8 snapshots и проверить удаление только самого старого | 56 | Q-028 |
| REC-003 | Backup включает SQLite/user config/rules/presets | §35 | Обязательно | Восстановить чистый стенд и сравнить все четыре категории | 56, 57 | Q-028 |
| REC-004 | Restore доступен через CLI либо Web UI | §35 | Обязательно | Выполнить документированный restore выбранным способом и пройти smoke | 57 | Q-028 |
| REC-005 | Factory reset удаляет user data/runtime config/learning history и оставляет ПО | §35 | Обязательно | Выполнить reset и проверить данные, model/cache, возможность нового bootstrap и установленную версию | 58 | Q-021 |

## REL. Отказы и lifecycle

Этап §51: **12**. Рабочие окна: **Д126–Д130**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| REL-001 | При ML Runtime failure lighting продолжает работать | §36 | Обязательно | Убить/заморозить Runtime и выполнить manual light/effect command | 50 | Q-031 |
| REL-002 | При ML Runtime failure audio продолжает работать | §36 | Обязательно | Убить Runtime при playback и выполнить audio controls | 50 | Q-031 |
| REL-003 | При ML Runtime failure Ambilight работает без ML | §36 | Обязательно | Выключить inference workers и проверить stream/colors/FPS | 19, 50 | Q-006 |
| REL-004 | При ML Runtime failure automation работает без ML | §36 | Обязательно | Выполнить sensor/time/manual rules при offline Runtime | 50 | Q-031 |
| REL-005 | При MQTT failure локальные GPIO/I²C/SPI устройства работают | §36 | Обязательно | Остановить broker и проверить LED/button/sensors через real adapters | 51 | Q-031 |
| REL-006 | SQLite failure включает Safe Mode с manual lighting/audio | §36 | Обязательно | Смоделировать недоступную/повреждённую БД и проверить разрешённые commands | 52 | Q-032 |
| REL-007 | В Safe Mode отключены automation и persistent changes | §36 | Обязательно | Попытаться выполнить rule/write config при failed SQLite | 52 | Q-032 |
| REL-008 | После restart LED zones восстанавливают last state | §36 | Обязательно | Настроить обе зоны, restart и сравнить mode/brightness/preset | 53 | Q-005 |
| REL-009 | После restart music остаётся STOPPED | §36 | Обязательно | Перезапустить во время playback и проверить отсутствие autoplay | 54 | — |
| REL-010 | После restart automation включена | §36 | Обязательно | Проверить enabled engine и допустимый regular rule после startup | — | Q-023 |
| REL-011 | После restart camera выключена до необходимости | §36 | Обязательно | Перезапустить с активным CV и проверить capture/indicator/explicit reactivation | 55 | Q-017 |
| REL-012 | Watchdog обнаруживает зависшие компоненты | §36 | Обязательно | Заморозить worker/adapter и проверить detection/recovery по согласованному timeout | 50 | Q-031 |
| REL-013 | Graceful shutdown прекращает commands, rules, fade OFF, flush SQLite, stop ML и завершает систему в этом порядке | §36 | Обязательно | Зафиксировать lifecycle events и проверить ordering/сохранность данных | 53, 54, 55 | Q-031 |
| REL-014 | Systemd запускает Compose, контейнеры имеют restart policies и healthchecks | §37 | Обязательно | Reboot Pi и crash services; проверить startup/restart/health; RestartSec=5 — пример настройки | 2 | Q-031 |

## OBS. Наблюдаемость и security events

Этап §51: **12**. Рабочие окна: **Д012–Д014; Д023; Д131–Д135**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| OBS-001 | Logs структурированы в JSON с DEBUG/INFO/WARN/ERROR levels | §38 | Обязательно | Проверить parse каждого уровня и filtering | — | Q-030 |
| OBS-002 | Logs используют correlation/event IDs где возможно | §38 | Рекомендация ТЗ | Проследить command→event→action по одному correlation ID | 39 | Q-030 |
| OBS-003 | Prometheus metrics доступны на /metrics | §38 | Обязательно | Проверить scrape/format/units и отсутствие sensitive labels | — | Q-034 |
| OBS-004 | GET /health показывает состояние процесса | §38 | Обязательно | Проверить normal/hung/dependency-down cases по согласованному health contract | — | Q-031 |
| OBS-005 | GET /ready проверяет SQLite/MQTT/ML Runtime/hardware adapters | §38 | Обязательно | Поочерёдно отключить dependencies и проверить diagnostics/status по принятой criticality policy | 61 | Q-031 |
| OBS-006 | Grafana/Prometheus разворачиваются optional Compose profile | §38,47 | Опционально | Включить/выключить profile и проверить dashboard/resource usage без влияния на базовую работу | — | Q-034 |
| OBS-007 | Alerts доставляются в Web UI, system logs и Telegram при доступности | §38 | Обязательно | Инжектировать alert, отключить bot и проверить оставшиеся каналы | — | Q-027 |
| OBS-008 | Экспортируется metric system_cpu_percent | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series system_cpu_percent, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-009 | Экспортируется metric system_memory_bytes | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series system_memory_bytes, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-010 | Экспортируется metric system_temperature_celsius | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series system_temperature_celsius, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-011 | Экспортируется metric led_fps | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series led_fps, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-012 | Экспортируется metric led_frame_latency_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series led_frame_latency_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-013 | Экспортируется metric led_estimated_current_amp | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series led_estimated_current_amp, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-014 | Экспортируется metric ambilight_fps | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series ambilight_fps, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-015 | Экспортируется metric ambilight_latency_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series ambilight_latency_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-016 | Экспортируется metric audio_fft_latency_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series audio_fft_latency_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-017 | Экспортируется metric audio_bpm | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series audio_bpm, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-018 | Экспортируется metric audio_beats_total | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series audio_beats_total, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-019 | Экспортируется metric gesture_inference_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series gesture_inference_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-020 | Экспортируется metric gesture_recognized_total | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series gesture_recognized_total, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-021 | Экспортируется metric presence_state | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series presence_state, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-022 | Экспортируется metric presence_inference_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series presence_inference_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-023 | Экспортируется metric scene_inference_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series scene_inference_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-024 | Экспортируется metric automation_trigger_total | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series automation_trigger_total, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-025 | Экспортируется metric automation_error_total | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series automation_error_total, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-026 | Экспортируется metric mqtt_connected_devices | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series mqtt_connected_devices, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-027 | Экспортируется metric http_requests_total | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series http_requests_total, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-028 | Экспортируется metric http_request_duration_ms | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series http_request_duration_ms, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-029 | Экспортируется metric ws_clients | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series ws_clients, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-030 | Экспортируется metric ml_runtime_up | §38 | Обязательно | Изменить соответствующее состояние/нагрузку и проверить series ml_runtime_up, тип и единицы; место laptop metric согласовать | — | Q-034 |
| OBS-031 | Формируется alert CPU temperature high | §38 | Обязательно | Инжектировать CPU temperature high, проверить создание/доставку/очистку согласно выбранной policy | — | Q-034 |
| OBS-032 | Формируется alert LED power limit reached | §38 | Обязательно | Инжектировать LED power limit reached, проверить создание/доставку/очистку согласно выбранной policy | — | Q-034 |
| OBS-033 | Формируется alert sensor offline | §38 | Обязательно | Инжектировать sensor offline, проверить создание/доставку/очистку согласно выбранной policy | — | Q-034 |
| OBS-034 | Формируется alert ML Runtime down | §38 | Обязательно | Инжектировать ML Runtime down, проверить создание/доставку/очистку согласно выбранной policy | — | Q-034 |
| OBS-035 | Формируется alert disk almost full | §38 | Обязательно | Инжектировать disk almost full, проверить создание/доставку/очистку согласно выбранной policy | — | Q-034 |
| OBS-036 | Логируется security event login success/failure | §39 | Обязательно | Выполнить login success/failure и проверить redacted event с результатом и временем без secrets | — | Q-030 |
| OBS-037 | Логируется security event API token create/revoke | §39 | Обязательно | Выполнить API token create/revoke и проверить redacted event с результатом и временем без secrets | — | Q-030 |
| OBS-038 | Логируется security event Bluetooth pairing | §39 | Обязательно | Выполнить Bluetooth pairing и проверить redacted event с результатом и временем без secrets | — | Q-030 |
| OBS-039 | Логируется security event unknown MQTT client | §39 | Обязательно | Выполнить unknown MQTT client и проверить redacted event с результатом и временем без secrets | — | Q-030 |
| OBS-040 | Логируется security event critical settings changes | §39 | Обязательно | Выполнить critical settings changes и проверить redacted event с результатом и временем без secrets | — | Q-030 |
| OBS-041 | Логируется security event factory reset | §39 | Обязательно | Выполнить factory reset и проверить redacted event с результатом и временем без secrets | — | Q-030 |

## PRIV. Сквозные ограничения privacy

Этап §51: **12**. Рабочие окна: **Все этапы; аудит Д148**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| PRIV-001 | Audio analysis выполняется локально | §40 | Обязательно | Проверить PCM/DSP flow и отсутствие внешних inference endpoints | — | Q-021 |
| PRIV-002 | Основная функциональность не использует облачные ML APIs | §40 | Обязательно | Проверить dependencies/config/network egress Core/Runtime/client | 18, 41 | Q-006 |
| PRIV-003 | Чувствительные ML/CV данные не покидают локальную систему | §40 | Обязательно | Проверить все external integrations, logs, diagnostics и release artifacts | 26, 41 | Q-021 |

## CI. Сборка и CI

Этап §51: **13**. Рабочие окна: **Д015; Д137–Д138**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| CI-001 | CI работает на GitHub Actions со стадиями lint/test/security/build/package | §43 | Обязательно | Проверить trigger, порядок stages и блокирование package при failed checks | 59 | Q-026 |
| CI-002 | Go checks включают gofmt/go vet/golangci-lint/govulncheck/gosec | §43 | Обязательно | Ввести по одному контролируемому нарушению и проверить соответствующий gate | 59 | Q-026 |
| CI-003 | CI выполняет go test -race ./... | §43,45 | Обязательно | Проверить workflow command и приложенный race/test report | 59 | Q-026 |
| CI-004 | Frontend CI выполняет ESLint/TypeScript check/unit tests/build | §43 | Обязательно | Проверить jobs и контролируемые negative cases | 59 | Q-026 |
| CI-005 | Python CI выполняет Ruff/mypy/pytest/dependency scan | §43 | Обязательно | Проверить jobs и reports; выбранный scanner/версии зафиксировать | 59 | Q-026 |
| CI-006 | Образы собираются multi-stage | §43 | Обязательно | Проверить Dockerfile stages и отсутствие build toolchain в final image | — | — |
| CI-007 | Контейнеры запускаются non-root | §43 | Обязательно | Проверить runtime UID каждого сервиса и документировать необходимые исключения как вопрос, не молча | — | Q-026 |
| CI-008 | Read-only filesystem используется где возможно | §43 | Обязательно | Проверить read-only rootfs и только необходимые writable volumes/tmpfs | — | Q-026 |
| CI-009 | У контейнеров есть healthchecks | §43 | Обязательно | Повредить readiness dependency и проверить container health status | 61 | Q-031 |
| CI-010 | Base images закреплены | §43 | Обязательно | Проверить pinned versions/digests и воспроизводимость build | — | Q-026 |
| CI-011 | Контейнеры поддерживают linux/arm64 и linux/amd64 | §43 | Обязательно | Собрать и запустить images обеих architectures, проверить native deps | — | Q-026 |

## DEPLOY. Release и rollback

Этап §51: **13**. Рабочие окна: **Д139–Д145; Д159–Д160**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| DEPLOY-001 | Production deploy запускается только по Git tag | §2,44 | Обязательно | Проверить отсутствие deploy на branch/PR и запуск доверенного tag | 59 | Q-026 |
| DEPLOY-002 | Версии релизов соответствуют SemVer | §44 | Обязательно | Проверить validation release tag и корректный release manifest | 59 | Q-026 |
| DEPLOY-003 | GitHub Actions разворачивает release через self-hosted runner на Pi | §2,44 | Обязательно | Выпустить pre-release и проверить фактическую установку exact artifact на Pi | 60 | Q-026 |
| DEPLOY-004 | Перед deploy проходят lint/test/security/build/package | §44 | Обязательно | Провалить один gate и проверить отсутствие обновления Pi | 59, 60 | Q-026 |
| DEPLOY-005 | После deploy проверяется /ready | §44 | Обязательно | Развернуть candidate и проверить readiness/smoke result в deployment log | 61 | Q-031 |
| DEPLOY-006 | Failed readiness автоматически откатывает предыдущую рабочую версию | §44 | Обязательно | Выпустить broken candidate и проверить images/schema/data/health предыдущего release | 62 | Q-028 |

## TEST. Методика и уровни тестирования

Этап §51: **13**. Рабочие окна: **Внутри этапов; Д146–Д155**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| TEST-001 | Обязательны unit/integration/E2E/HIL/simulation levels | §45 | Обязательно | Показать исполняемые suites/протоколы каждого уровня и запуск в подходящем окружении | — | Q-035 |
| TEST-002 | Core business logic coverage >=80% | §45 | Обязательно | Измерить покрытие согласованного core scope без generated/vendor | — | Q-035 |
| TEST-003 | Overall coverage >=70% | §45 | Обязательно | Собрать reports языков и вычислить согласованный overall metric | — | Q-035 |
| TEST-004 | Go core logic использует table-driven tests | §45 | Обязательно | Проверить tests важных business branches и edge cases | — | — |
| TEST-005 | Python tests используют pytest и deterministic image/video/audio fixtures | §45 | Обязательно | Повторить suite с фиксированными inputs и одинаковыми outputs | — | Q-021 |
| TEST-006 | Есть golden tests Ambilight mapping/FFT/beat detection/LED effects/gesture mapping/scene preprocessing | §45 | Обязательно | Для каждого семейства сверить expected outputs и sensitivity к нарушению алгоритма | — | Q-035 |
| TEST-007 | Simulation E2E проходит fake screen→analyzer→Core→virtual strip→expected colors | §45 | Обязательно | Запустить полный scenario в CI без Pi | 14 | — |
| TEST-008 | HIL проверяет WS2812B/button/sensors/Bluetooth/camera/external monitor/PSU/level shifter | §45 | Обязательно | Выполнить hardware checklist с evidence каждого компонента | — | Q-002 |
| TEST-009 | Benchmarks покрывают render/Ambilight/WS/event bus/FFT/ML inference | §45 | Обязательно | Сохранить воспроизводимые commands, environment и baseline каждого benchmark | — | Q-013 |

## PERF. Измеримые performance targets

Этап §51: **12**. Рабочие окна: **Д005; профилирование этапов; Д135; Д152–Д155**.

| ID | Требование | Источник ТЗ | Обязательность | Как проверить | §50 | Вопрос |
|---|---|---|---|---|---|---|
| PERF-001 | REST p95 <100 мс | §46 | Обязательно | Измерить end-to-end latency согласованного набора обычных API requests под заданной нагрузкой | — | Q-013 |
| PERF-002 | WebSocket event latency <100 мс | §46 | Обязательно | Измерить publish→client receive с учётом clock error | — | Q-013 |
| PERF-003 | LED rendering >=30 FPS | §46 | Обязательно | Посчитать фактически применённые кадры и интервалы при reference joint load | — | Q-013 |
| PERF-004 | Ambilight работает на 30 FPS | §11.7,46 | Обязательно | Измерить применённые frames, drops и стабильность на reference LAN | 15 | Q-013 |
| PERF-005 | Ambilight end-to-end latency <100 мс | §11.7,46 | Обязательно | Измерить screen change→physical LED change, не заменяя RTT | 16 | Q-013 |
| PERF-006 | Gesture reaction <500 мс | §46 | Обязательно | Измерить согласованное начало устойчивого жеста→выполненное действие, включая confirmation | 22, 23, 24, 25 | Q-018 |
| PERF-007 | Scene ML выполняется <=1 inference/sec | §46 | Обязательно | Посчитать вызовы inference по timestamps при 30 FPS capture | 18 | Q-013 |
| PERF-008 | Presence inference выполняется каждые 2–3 секунды | §46 | Обязательно | Измерить интервалы при совместной CV/audio/LED нагрузке | 40 | Q-013 |

## Связь всех 62 пунктов приёмки с требованиями

Здесь сохранены формулировки сценария §50. Номер пункта — не ID единственного требования: для проверки нужны все связанные строки и их зависимости. Статус ниже намеренно не PASS: выполнение Дня 1 не означает реализации функций.

| № §50 | Исходный критерий | Требования | Формальная проверка | Статус |
|---:|---|---|---|---|
| 1 | Raspberry Pi загружается с Ubuntu Server. | HW-001, HW-002 | Д152 | Не проверено |
| 2 | systemd автоматически запускает Docker Compose. | ARCH-010, REL-014 | Д152 | Не проверено |
| 3 | Пользователь открывает `https://smarthome.local`. | API-005, API-006 | Д152 | Не проверено |
| 4 | Пользователь авторизуется как admin. | AUTH-001, AUTH-002, AUTH-003, AUTH-004, AUTH-005, AUTH-006, AUTH-007 | Д152 | Не проверено |
| 5 | Dashboard показывает состояние Raspberry Pi и устройств. | DEV-004, DEV-005, DEV-006, WEB-002, WEB-004, WEB-006, WEB-010 | Д152 | Не проверено |
| 6 | Телефон подключается к `Smart Home Hub` по Bluetooth. | AUDIO-001, AUDIO-004, AUDIO-005, DEV-007 | Д152 | Не проверено |
| 7 | Пользователь запускает музыку. | AUDIO-001, AUDIO-002, AUDIO-003, AUDIO-012, AUDIO-013, AUDIO-014, AUDIO-015, AUDIO-016, AUDIO-017 | Д152 | Не проверено |
| 8 | Музыка воспроизводится через выбранный audio output. | AUDIO-006, AUDIO-007, AUDIO-008, AUDIO-009 | Д152 | Не проверено |
| 9 | `desk.wall` реагирует на bass/beat/FFT. | LIGHT-002, LIGHT-015, LIGHT-016, LIGHT-017, LIGHT-018, LIGHT-019, DSP-001, DSP-002, DSP-003, DSP-004, DSP-005, DSP-006, DSP-007, DSP-008, DSP-009 | Д152 | Не проверено |
| 10 | Web UI показывает realtime spectrum и BPM. | DSP-001, DSP-010, WEB-005, WEB-006 | Д152 | Не проверено |
| 11 | На ноутбуке запускается screen-client. | ARCH-012, AMBI-001, AMBI-002, AMBI-003, AMBI-004, AMBI-005 | Д152 | Не проверено |
| 12 | Screen-client проходит pairing. | AMBI-009, AUTH-008, AUTH-009, AUTH-010 | Д152 | Не проверено |
| 13 | `monitor.backlight` переключается в `AMBILIGHT`. | LIGHT-001, LIGHT-020 | Д152 | Не проверено |
| 14 | LED по четырём сторонам отображают цвета соответствующих краёв экрана. | HW-005, LIGHT-001, AMBI-006, AMBI-007, AMBI-010, AMBI-011, AMBI-012, AMBI-013, AMBI-014, WEB-007, WEB-008, SIM-007, TEST-007 | Д152 | Не проверено |
| 15 | Ambilight работает около 30 FPS. | AMBI-016, SCENE-004, PERF-004 | Д152 | Не проверено |
| 16 | End-to-end latency меньше 100 ms. | PERF-005 | Д152 | Не проверено |
| 17 | Cinema Mode обеспечивает плавную реакцию. | LIGHT-007, AMBI-015, AMBI-017 | Д152 | Не проверено |
| 18 | ML scene adaptation корректирует параметры Ambilight. | ARCH-003, SCENE-001, SCENE-002, SCENE-003, SCENE-004, SCENE-005, SCENE-006, SCENE-007, PRIV-002, PERF-007 | Д152 | Не проверено |
| 19 | При отключении ML обычный Ambilight продолжает работу. | SCENE-007, REL-003 | Д152 | Не проверено |
| 20 | При потере screen-client включается fallback preset. | AMBI-019 | Д152 | Не проверено |
| 21 | Пользователь включает Gesture Mode. | GEST-001, GEST-002, GEST-005, GEST-008 | Д153 | Не проверено |
| 22 | `OPEN_PALM` управляет Play/Pause. | GEST-003, GEST-004, GEST-009, GEST-010, PERF-006 | Д153 | Не проверено |
| 23 | Swipe переключает треки. | GEST-003, GEST-004, GEST-009, GEST-011, GEST-012, PERF-006 | Д153 | Не проверено |
| 24 | `PALM_FACE_DOWN` включает Night Shift. | LIGHT-009, LIGHT-010, LIGHT-021, GEST-003, GEST-004, GEST-009, GEST-017, PERF-006 | Д153 | Не проверено |
| 25 | Rotation управляет громкостью. | GEST-003, GEST-004, GEST-009, GEST-018, GEST-019, PERF-006 | Д153 | Не проверено |
| 26 | Кадры камеры нигде не сохраняются. | GEST-007, STORE-003, PRIV-003 | Д153 | Не проверено |
| 27 | LED-индикатор показывает активность камеры. | PRES-006 | Д153 | Не проверено |
| 28 | Single press физической кнопки переключает свет ON/OFF. | BTN-001, BTN-002, BTN-005 | Д153 | Не проверено |
| 29 | Double press включает Night Shift. | LIGHT-009, LIGHT-010, LIGHT-021, BTN-001, BTN-003, BTN-005 | Д153 | Не проверено |
| 30 | Long press переключает preset. | LIGHT-008, BTN-001, BTN-004, BTN-005 | Д153 | Не проверено |
| 31 | Система получает: освещённость; температуру; влажность; движение; качество воздуха. | SENS-001, SENS-002, SENS-003, SENS-004, SENS-005, SENS-006, SENS-008 | Д153 | Не проверено |
| 32 | Датчик освещённости корректирует brightness. | SENS-007 | Д153 | Не проверено |
| 33 | Значения доступны Automation Engine. | SENS-008, AUTO-020, AUTO-025 | Д153 | Не проверено |
| 34 | Rule Builder позволяет создать правило с AND/OR. | AUTO-002, AUTO-003, AUTO-004, WEB-009 | Д153 | Не проверено |
| 35 | Automation Engine выполняет правило. | SENS-009, AUTO-001, AUTO-006, AUTO-007, AUTO-013, AUTO-014, AUTO-015, AUTO-016, AUTO-017, AUTO-018, AUTO-019, AUTO-020, AUTO-021, AUTO-022, AUTO-023, AUTO-024, AUTO-025 | Д153 | Не проверено |
| 36 | Delayed action работает. | AUTO-005 | Д153 | Не проверено |
| 37 | Rule priority работает. | AUTO-008 | Д153 | Не проверено |
| 38 | Manual override временно блокирует автоматизацию. | SENS-009, AUTO-009, AUTO-012 | Д153 | Не проверено |
| 39 | Automation log содержит запись о срабатывании. | AUTO-010, AUTO-011, WEB-006, STORE-002 | Д153 | Не проверено |
| 40 | Presence Engine объединяет camera/PIR/Bluetooth/network. | PRES-001, PRES-002, PRES-003, PRES-004, PRES-005, SIM-006, PERF-008 | Д154 | Не проверено |
| 41 | Habit Predictor использует только локальные данные. | ARCH-003, HAB-001, HAB-007, HAB-011, STORE-002, PRIV-002, PRIV-003 | Д154 | Не проверено |
| 42 | После 30 дней обучения система может предложить действие. | HAB-001, HAB-003, HAB-004, HAB-005, HAB-007, HAB-008 | Д154 | Не проверено |
| 43 | Confidence thresholds работают. | HAB-005, HAB-006 | Д154 | Не проверено |
| 44 | Пользователь может отклонить recommendation. | HAB-009, HAB-010 | Д154 | Не проверено |
| 45 | Пользователь может удалить learning history. | HAB-013, HAB-014 | Д154 | Не проверено |
| 46 | MQTT-устройство обнаруживается автоматически. | DEV-004, DEV-005, MQTT-001, MQTT-005, MQTT-006, SIM-005 | Д154 | Не проверено |
| 47 | Telegram Bot выполняет базовые команды. | ARCH-010, TG-001, TG-002, TG-003, TG-004, TG-007, TG-008, TG-009, TG-010, TG-011, TG-012, TG-013, AUTH-008, AUTH-011 | Д154 | Не проверено |
| 48 | Smart Home Hub работает без Telegram. | ARCH-001, AUDIO-003, TG-006 | Д154 | Не проверено |
| 49 | Архитектура позволяет подключить Home Assistant. | ARCH-001, ARCH-007, MQTT-006, MQTT-007, MQTT-008, AUTH-008, AUTH-011 | Д154 | Не проверено |
| 50 | При падении ML Runtime: lighting работает; audio работает; Ambilight работает без ML; Automation Engine работает без ML. | ARCH-003, REL-001, REL-002, REL-003, REL-004, REL-012 | Д155 | Не проверено |
| 51 | При падении MQTT локальное GPIO-управление продолжает работу. | REL-005 | Д155 | Не проверено |
| 52 | При проблемах SQLite система переходит в Safe Mode. | REL-006, REL-007 | Д155 | Не проверено |
| 53 | После рестарта LED state восстанавливается. | ARCH-011, STORE-001, STORE-002, REL-008, REL-013 | Д155 | Не проверено |
| 54 | Музыка после рестарта остаётся STOPPED. | REL-009, REL-013 | Д155 | Не проверено |
| 55 | Camera после рестарта выключена. | REL-011, REL-013 | Д155 | Не проверено |
| 56 | Daily backup создаётся автоматически. | REC-001, REC-002, REC-003 | Д155 | Не проверено |
| 57 | Backup может быть восстановлен. | REC-003, REC-004 | Д155 | Не проверено |
| 58 | Factory reset работает. | REC-005 | Д155 | Не проверено |
| 59 | Release tag запускает GitHub Actions. | STORE-008, CI-001, CI-002, CI-003, CI-004, CI-005, DEPLOY-001, DEPLOY-002, DEPLOY-004 | Д155 | Не проверено |
| 60 | Self-hosted runner разворачивает новую версию. | STORE-008, DEPLOY-003, DEPLOY-004 | Д155 | Не проверено |
| 61 | `/ready` проверяет успешность deploy. | OBS-005, CI-009, DEPLOY-005 | Д155 | Не проверено |
| 62 | При неуспешном healthcheck выполняется rollback. | DEPLOY-006 | Д155 | Не проверено |

## Обязательные проверки вне 62 пунктов §50

Все строки «Обязательно» с «—» в колонке §50 остаются условиями первой версии. Среди них: все 11 lighting modes, arbitrary zones, MATRIX mapping, scopes/revoke/session protection, retention 90/30 дней, отсутствие DSP/frame history, PWA, API specification/rate limits, полный набор метрик, CI/coverage и контейнерные ограничения. Финальную приёмку нужно дополнить этими проверками из каталога; сценарий §50 сам по себе неполон.

## Границы объёма

### Обязательная первая версия

Весь каталог «Обязательно», включая Telegram как реализованный интерфейс, Scene ML, gestures, habit prediction и четыре платформы клиента. Telegram не обязателен для запуска базовой локальной системы, но его реализация обязательна для MVP по §25. Логическая MATRIX support обязательна, даже если физическую панель приобретут позже.

### Опционально и рекомендации

Gaming Mode (§11.8), физическая LED panel после логической поддержки (§20), Grafana/Prometheus collectors в optional profile (§38/47), будущий HA MQTT Discovery (§48). Prometheus-compatible endpoint /metrics обязателен независимо от установки collector. GPIO fan и конкретные level-shifter chips — рекомендации исходного ТЗ. VPN — рекомендуемый способ при наличии задачи удалённого доступа, а не обязательная установка двух VPN.

### Не входит в MVP по §52

- Полноценное распознавание личности и multi-user UX. Модель User/Role/Session и person yes/no при этом обязательны.
- Обязательная cloud integration, AirPlay, Kubernetes и несколько LED-протоколов.
- Обязательная установка Home Assistant. Совместимые REST/MQTT boundaries остаются обязательными.
- Хранение camera/screen frames и история музыки не просто отложены: они запрещены другими разделами первой версии.
- Облачные ML APIs для основной функциональности запрещены §40; их нельзя включать как optional обход.

### Будущее по §53, без обязательств первой версии

ESP32 controllers, Zigbee, Thread/Matter, smart sockets, relays, climate devices, несколько комнат, multi-user profiles, face recognition, mobile application, Home Assistant discovery, distributed k3s cluster. Наличие MQTT/расширяемой модели не означает реализацию этих устройств и протоколов сейчас.

## Проверка охвата разделов исходного ТЗ

Это навигационная сверка разделов; содержательная полнота обеспечивается строками каталога, а не одним наличием номера раздела.

| § ТЗ | Название | Где учтён |
|---:|---|---|
| 1 | Назначение | ARCH |
| 2 | Ключевые архитектурные решения | ARCH, STORE, SIM, DEPLOY |
| 3 | Аппаратная платформа | HW, THERM |
| 4 | LED-подсветка | HW, LIGHT |
| 5 | Электропитание LED | HW |
| 6 | Power Limiter | POWER |
| 7 | Архитектура системы | ARCH |
| 8 | Внутренний Event Bus | ARCH |
| 9 | Audio subsystem | AUDIO |
| 10 | DSP и музыкально-адаптивная подсветка | DSP |
| 11 | Ambilight | AMBI, PERF |
| 12 | ML Scene Adaptation | SCENE |
| 13 | Gesture Recognition | GEST |
| 14 | Night Shift | LIGHT |
| 15 | Presence Engine | PRES |
| 16 | Sensors | SENS |
| 17 | Habit Prediction | HAB |
| 18 | Automation Engine | AUTO |
| 19 | Physical Button | BTN |
| 20 | LED Panel | DEV |
| 21 | Device Manager | SENS, DEV, MQTT |
| 22 | MQTT | MQTT |
| 23 | Home Assistant | MQTT |
| 24 | Web UI | LIGHT, WEB |
| 25 | Telegram Bot | TG |
| 26 | Authentication и Multi-user readiness | ARCH, AUTH |
| 27 | Machine API Tokens | AUTH |
| 28 | REST API | API |
| 29 | mDNS и HTTPS | API |
| 30 | Remote Access | API |
| 31 | Storage | DSP, AMBI, STORE |
| 32 | Configuration | STORE |
| 33 | Public Repository Security | STORE |
| 34 | Simulation Mode | SIM |
| 35 | Backup и Recovery | REC |
| 36 | Reliability | REL |
| 37 | systemd / Docker Compose | REL |
| 38 | Observability | OBS |
| 39 | Security Events | OBS |
| 40 | Privacy Requirements | AMBI, GEST, HAB, PRIV |
| 41 | Repository Structure | ARCH |
| 42 | Hardware Abstractions | ARCH |
| 43 | CI/CD | CI |
| 44 | Deployment | DEPLOY |
| 45 | Testing | CI, TEST |
| 46 | Performance Targets | PERF |
| 47 | API/Service Separation | ARCH, OBS |
| 48 | Home Assistant Compatibility | MQTT |
| 49 | Master OFF | AUTO |
| 50 | Критерий приёмки | Таблица связи всех 62 пунктов выше |
| 51 | Этапы разработки | Этапы 1–13 в backlog.md и окна дней каталога |
| 52 | Не входит в MVP | AUDIO |
| 53 | Потенциальное развитие | Границы объёма: будущее |
| 54 | Итог | ARCH |
