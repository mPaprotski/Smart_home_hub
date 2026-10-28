# День 2. Предварительная карта интерфейсов

Статус: **candidate mapping для Raspberry Pi 4 40-pin header**, 25 октября 2026 года. GPIO ниже — BCM, physical — номер контакта разъёма. Не являются командами конфигурации ОС.

Основание mapping и 3.3V logic — [Raspberry Pi S01](https://www.raspberrypi.com/documentation/computers/raspberry-pi.html#gpio-and-the-40-pin-header). Проектные назначения — наш выбор для spike; ни одно не проверено на конкретной Pi. [Схема](wiring.md), [BOM](../bom/BOM.md), [источники](../research/SOURCES.md).

## 1. Основной кандидат B: SPI chain

| Назначение | BCM GPIO | Physical pin | Направление | Уровень / питание | Статус и примечание |
|---|---:|---:|---|---|---|
| I²C1 SDA sensors | 2 | 3 | Bidirectional | 3.3V | Общий SDA; не отдельный GPIO на каждый sensor |
| I²C1 SCL sensors | 3 | 5 | Output clock | 3.3V | Candidate 100kHz; pull-ups/ёмкость проверить |
| SPI0 MOSI → AHCT → LED DIN | 10 | 19 | Output | 3.3V до buffer / 5V после | Один поток, два logical slices |
| Momentary button | 17 | 11 | Input | Pull-up 3.3V, замыкание на GND | Debounce; не назначать одновременно SPI1 CE1 |
| PIR motion OUT | 27 | 13 | Input | 3.3V, high impedance | Конкретный sensor output подтвердить |
| Visible CV LED | 22 | 15 | Output | 3.3V через resistor | Проверить Vf/current/visibility |
| Fan module control | 16 | 36 | Output | 3.3V logic | Не motor power; actual module/default state TBD |
| LED DATA enable stage | 25 | 22 | Output | 3.3V → transistor/module | Не прямое подключение к 5V OE |
| Sensors low-current supply | — | 1 или 17 | Power | 3.3V_PI | После учёта общего budget |
| Fan power | — | 2 или 4 | Power | 5V_PI | Через подходящий driver; LED PSU отдельно |
| Signal reference GND | — | 6, другие GND при необходимости | Reference | 0V common | Не силовой return ленты |

Первый GPIO10 не значит physical pin10: physical pin10 — другой сигнал. На этапе монтажа проверять ориентацию header по board marking.

## 2. Резервы и несовместимые назначения

| Ресурс | BCM / physical | Правило проекта |
|---|---|---|
| SPI0 MISO | GPIO9 / 21 | Не нужен WS2812B, но bus зарезервирована целиком |
| SPI0 SCLK | GPIO11 / 23 | Не выводить как WS2812B clock: у ленты только DATA; оставить bus reservation |
| SPI0 CE0/CE1 | GPIO8 / 24, GPIO7 / 26 | Не использовать для второго generic SPI sensor на LED bus |
| PWM0 alternative | GPIO18 / 12; GPIO12 / 32 | Только альтернативный вариант A, не второй transport по умолчанию |
| PWM1 alternative | GPIO13 / 33; GPIO19 / 35 | PWM0/PWM1 вариант конфликтует с аналоговым audio согласно S02 |
| PCM/I²S | GPIO18 / 12, GPIO19 / 35, GPIO20 / 38, GPIO21 / 40 | Резерв для анализа alternative C/I²S; не назначать fan/button здесь |
| HAT EEPROM | GPIO0 / 27, GPIO1 / 28 | Не занимать как обычные sensors GPIO |
| UART | GPIO14 / 8, GPIO15 / 10 | Сохранить возможную консоль диагностики; применение не обязательно |
| Свободные кандидаты | GPIO4 / 7, GPIO5 / 29, GPIO6 / 31, GPIO12 / 32, GPIO13 / 33, GPIO18 / 12, GPIO19 / 35, GPIO20 / 38, GPIO21 / 40, GPIO23 / 16, GPIO24 / 18, GPIO26 / 37 | Свободны только при отсутствии включённых alternate functions; не все одновременно доступны при вариантах A/C |

**Вариант A:** monitor GPIO18/PWM0 и desk GPIO13/PWM1 — только alternative candidate. Не включать его одновременно с вариантом B и не считать, что USB DAC автоматически закрывает требование встроенного 3.5mm output. Решение по Q-002/Q-007 до реализации adapters.

## 3. I²C address map

| Устройство-кандидат | Address 7-bit | GPIO / physical | Supply / logic proposal | Измерение |
|---|---|---|---|---|
| VEML7700 breakout | 0x10 | SDA2/3, SCL3/5 | 3.3V / 3.3V | illuminance, lux |
| BME280 breakout | 0x76 либо 0x77 | Общие SDA/SCL | 3.3V / 3.3V | °C, %RH |
| SGP40 breakout | 0x59 | Общие SDA/SCL | 3.3V / 3.3V | raw VOC, local VOC Index |

Адреса — по S06/S07/S08, не получены сканированием реальной bus. BME280 strap выбрать явно, SDO не оставлять неопределённым на bare IC; готовый module может иметь jumper. Выбранные three addresses различны, второй sensor той же фиксированной address потребовал бы другого bus/multiplexer.

Проверить точные платы: input supply pin, onboard pull-ups/rails, regulators и параллельное сопротивление. Не переносить bare-chip supply range на вход VIN carrier без schematic.

## 4. Вне GPIO header

| Интерфейс | Подключение | Что подтвердить |
|---|---|---|
| Pi power | USB-C | PSU и low-voltage detection under load |
| USB webcam | USB host | UVC modes/power/permissions; exact model TBD |
| Pi Camera | CSI connector | Pi4-compatible cable/camera/orientation/OS stack |
| USB DAC | USB host | ALSA/audio graph, power, source switching |
| USB SSD | USB host | Enclosure/boot/power с webcam/DAC |
| Analog audio | Pi 3.5mm jack | Совместная работа LED/audio выбранного transport |
| Bluetooth phone/output | Встроенный radio | Pairing mode, роли, одновременность input/output |
| LAN | Ethernet/Wi-Fi | Reference latency на Ethernet; Wi-Fi отдельно |

## 5. Проверка конфликтов карты

На уровне документа основная карта не использует один GPIO для двух несвязанных функций. Три sensors совместно используют одну I²C bus намеренно. SPI0 целиком зарезервирована. GPIO16 не назначается SPI1 CE2, GPIO17 не назначается SPI1 CE1; SPI1 не включается в этой схеме.

Проверка уникальности таблицы не доказывает, что kernel overlays/audio/driver не занимают ресурс. На Д006/Д008 сверить фактические pinmux/overlays, driver requirements, device permissions и осциллограмму при joint load. Bus label в API не заменяет physical wiring review.
