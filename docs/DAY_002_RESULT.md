# День 2 — результат аппаратной проработки

Дата: 25 октября 2026 года. Основание: [ТЗ](../Smart_Home_Hub_TZ.md), Д002 [ежедневного плана](DEVELOPMENT_PLAN_DAILY.md), [День 1](DAY_001_RESULT.md).

**Статус: предварительный вариант Дня 2 выполнен.** Пользователь сообщил, что замеров пока нет, и поручил подготовить варианты. BOM/расчёты/схема готовы в этой границе; фактический подбор номиналов и монтаж остаются открыты до замеров и аппаратных проверок.

## 1. Созданные материалы

| Файл | Содержимое |
|---|---|
| [BOM.md](../hardware/bom/BOM.md) | 44 позиции/категории: контроллер/хранение/cooling, LED/power/protection, четыре sensors, кнопка/indicator, камеры/audio и test environments; наличие неизвестно |
| [design-inputs.json](../hardware/bom/design-inputs.json) | Что задано ТЗ, что неизвестно (`null`) и какие допущения использованы |
| [power-budget.md](../hardware/bom/power-budget.md) | Geometry/counts, четыре сценария, PSU reserve, limiter/idle semantics, wire drop, injection/fuses и отдельный Pi budget |
| [power_budget.py](../hardware/bom/power_budget.py) | Воспроизводимый расчёт на стандартной библиотеке Python, CLI inputs и input validation |
| [power-scenarios.json](../hardware/bom/power-scenarios.json) | Численные результаты четырёх иллюстративных сценариев |
| [wiring.md](../hardware/wiring/wiring.md) | DC/data topology, alternatives, level shifter/OE, sensors/button/indicator/fan, список обязательных spike checks |
| [pin-map.md](../hardware/wiring/pin-map.md) | BCM/physical mapping, reserved resources, I²C addresses и non-GPIO interfaces |
| [SOURCES.md](../hardware/research/SOURCES.md) | 12 первичных источников и границы документального подтверждения |

## 2. Основные результаты

1. Выявлен документированный **PWM LED / встроенный analog audio conflict** у rpi_ws281x. Две PWM DATA линии не считаются готовой базой полного ТЗ.
2. Проработан кандидат **SPI0 chain с двумя независимо управляемыми logical zones**. Это проектный вывод; требует проверки значения независимости зон, межзонного кабеля, driver/clock/permissions и joint audio load. Не объявлен окончательным ADR.
3. Sensor candidate set: VEML7700 + BME280 + SGP40 + PIR. BME280 даёт две величины, поэтому четыре devices закрывают пять типов измерений. VOC Index не называется измеренным CO₂.
4. Подготовлены S0/S1/S2/S3: 60/240/300/360 LED суммарно. Это не измерения пользовательской комнаты. Подбор PSU классов учитывает auxiliary 0.1A и условный запас 25%.
5. GPIO assignments не дублируются в основном варианте; I²C sharing намеренное, SPI0 зарезервирована под LED. Physical/BCM numbering разделены.
6. Выделены unresolved power sequencing/hard-off/partial-power вопросы. Draft-схема не выдана за ready-to-wire installation.

## 3. Что проверено в рамках Дня 2

- [x] Прочитаны ТЗ, вопросы и ограничения предыдущего дня; исходное ТЗ не изменено.
- [x] BOM покрывает все аппаратные классы MVP и необходимые acceptance environments, без фиктивных claims наличия/покупки.
- [x] Расчёты четырёх сценариев воспроизводятся CLI и совпадают с таблицей.
- [x] Проверены invalid/missing counts, negative/nonfinite parameters и неверный reserve; неизвестные значения не подменяются нулём.
- [x] Проверены уникальные B01–B44, BCM/physical mapping основного варианта, address uniqueness и внутренние Markdown links.
- [x] Обновлены Q-002/Q-003/Q-004/Q-007/Q-016/Q-020 с проектным прогрессом; вопросы не объявлены закрытыми.
- [x] Backlog отражает выполненный Д001 и предварительный результат Д002; аппаратные spikes остаются невыполненными.

Проверки здесь — арифметика/структура документов/сверка документации, **не hardware-in-the-loop**. Монтаж, подача питания, закупки, driver installation, actual GPIO reads, audio/camera/PSU load tests не выполнялись.

## 4. Что требуется после замеров

| Данные / решение | Что блокирует | Следующее окно |
|---|---|---|
| Реальные четыре counts монитора и desk count | Exact LED quantity/PSU/injection/mapping | До финального BOM и Д027–Д034 |
| Наличие Pi/PSU/storage/cooling | Реальная подготовка стенда | До Д006 |
| Distances и межзонная DATA трасса | Выбор single-chain topology, wire sections, signal integrity | Д003–Д008 |
| Exact LED/SKU/PSU/fan/sensor breakout | Datasheet values, protection/permissions/driver compatibility | BOM уточнение до подключения |
| Что означает независимость зон | Возможность принять две logical zones на одном канале | Д003–Д004 |
| LED power sequencing/physical hard-off interlock | Безопасная постоянная работа двух источников | Д004/Д008 до полной установки |
| Выбранная поддерживаемая Ubuntu и camera models | Реальная CSI/UVC compatibility | Д006 и Д081–Д083 |
| Actual cable ratings/fuse curves/connectors | Закрытие wiring/power protection | До монтажа Д034 |

Отсутствие замеров по согласованному варианту не мешает Д003: можно определить состояния/команды/events и записать условные аппаратные boundaries. Оно мешает считать финальную электрическую конфигурацию и M0 уже подтверждёнными.

## 5. Привязка к требованиям и следующим дням

Подготовлена доказательная база для HW-001…HW-012, THERM-001…THERM-003, LIGHT-001…LIGHT-004, POWER-001…POWER-005, GEST-001, SENS-001…SENS-009, BTN-001, DEV-001…DEV-003, PRES-006 и TEST-008 из [матрицы](requirements.md). Их продуктовый статус не меняется на PASS: соответствие подтвердится аппаратными и программными проверками будущих этапов.

Следующий плановый день — **Д003: доменные контракты и решения**. Выбор final hardware topology и противоречия readiness/camera/override оформляются там; experiments Д004/Д008 проверят предположения.
