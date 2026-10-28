# День 2. Предварительный BOM

Статус: **проектный вариант до замеров и стендовых испытаний**, 25 октября 2026 года.

Основание: [ТЗ](../../Smart_Home_Hub_TZ.md), [матрица требований](../../docs/requirements.md), [источники](../research/SOURCES.md). Связанные документы: [входные данные](design-inputs.json), [питание](power-budget.md), [схема](../wiring/wiring.md), [интерфейсы](../wiring/pin-map.md), [отчёт Д002](../../docs/DAY_002_RESULT.md).

**Подтверждена Raspberry Pi 4B с2GB RAM; наличие остальных компонентов неизвестно.** «Неизвестно» не означает «нужно купить». Точные цены, продавцы, заказ и локальная доступность не проверялись. Кандидат — вариант для проверки, а не утверждённая покупка. В ТЗ 120 LED приведены как пример, а не результат замера.

## 1. Контроллер, хранение, сеть, охлаждение

| ID | Компонент / кандидат | Количество | Статус наличия | Критерий выбора / ближайшая проверка |
|---|---|---:|---|---|
| B01 | Raspberry Pi 4 Model B, только2GB | 1 | Есть, подтверждено владельцем26.10.2026 | SSH/OS и фактический MemTotal ещё не проверены; замену платы не предполагаем |
| B02 | Отдельный USB-C PSU Pi, кандидат official 5.1V/3A | 1 | Неизвестно | S12; проверить суммарную Pi/USB/fan нагрузку и undervoltage |
| B03 | microSD для bootstrap | 1, если старт не с SSD | Неизвестно | Ёмкость/ресурс после выбора ОС; тест чтения/записи |
| B04 | USB SSD и совместимый USB enclosure/adapter | 1 комплект, целевой | Неизвестно | Boot compatibility и USB power; migration не считается уже выполненной |
| B05 | Радиатор и 5V fan 30–40мм с креплением | 1 комплект | Неизвестно | Не закрывает CSI/порты; ток/шум/температуры; S12 budget |
| B06 | Fan driver module с документированным 3.3V input | 1, проектный вариант | Неизвестно | Ток выше startup fan current, default state/защита; exact model после fan |
| B07 | Ethernet cable | 1 при wired LAN | Неизвестно | Предпочтительный эталон сети для latency; Wi-Fi проверить отдельно |
| B08 | Корпус/стойки, крепёж, разгрузка кабелей | 1 комплект | Неизвестно | Доступ к охлаждению/GPIO/CSI; отсутствие механических коротких замыканий |

## 2. LED, преобразование сигналов и питание

| ID | Компонент / кандидат | Количество | Статус наличия | Критерий выбора / ближайшая проверка |
|---|---|---:|---|---|
| B09 | WS2812B RGB, 5V, 60 LED/м для monitor.backlight | По четырём замерам | Неизвестно | Точный SKU/revision/datasheet, counts sides, порядок RGB/GRB |
| B10 | Такая же WS2812B для desk.wall | По замеру стола | Неизвестно | По возможности тот же SKU/партия; не RGBW/12V variant |
| B11 | Закрытый внешний LED PSU 5V DC | 1; ток после counts | Неизвестно | Full-white + auxiliary + reserve; OCP/OVP, DC terminals/rating; без привязки к 20A до замеров |
| B12 | Level shifter на SN74AHCT125, DIP/carrier с понятной схемой | 1 | Неизвестно | S03; 5V supply, 3.3V-compatible input, OE доступен; не generic MOSFET I²C converter |
| B13 | DATA resistor, кандидат 330Ω | 1 на первый DIN; ещё 1 на межзонный DIN при отдельном проводе | Неизвестно | Стартовое значение из диапазона S05; подтвердить waveform/edges |
| B14 | Bulk electrolytic, кандидат 1000µF ≥10V | По одному у ввода каждой зоны, далее по injection plan | Неизвестно | Полярность/voltage/ripple; S05; это не замена проводов нужного сечения |
| B15 | Ceramic decoupling 100nF | 1 у shifter + по фактическим discrete circuits | Неизвестно | Breakout могут уже иметь decoupling; не считать их bare IC |
| B16 | Main DC fuse/holder F0 | 1 | Неизвестно | Номинал/характеристика/отключающая способность после wiring current budget |
| B17 | Branch fuses/holders | По защищаемым feeds, минимум 2 зоны | Неизвестно | При нескольких injection leads каждый провод защищён в соответствии с topology |
| B18 | Физический LED power switch / disconnect | 1 | Неизвестно | DC rating под итоговый ток, доступен пользователю; не GPIO relay без расчёта |
| B19 | Distribution terminals, insulated connectors | По injection/wiring | Неизвестно | Реальный continuous current каждой ветви; не breadboard/Dupont для LED power |
| B20 | Пары силовых проводов +V/GND | Длины/сечения после замеров | Неизвестно | Voltage drop и ampacity/temperature; power-budget.md |
| B21 | DATA/GND signal cable и межзонное соединение | После замеров расстояний | Неизвестно | Для SPI chain нужен DOUT monitor→DIN desk; при длинном трассировании пересмотреть topology |
| B22 | OE enable stage 3.3V→open collector | 1 при выбранной enable-схеме | Неизвестно | Кандидат NPN stage; конкретный transistor pinout/BOM уточнить до монтажа |
| B23 | OE/input pull resistors и base resistor | По схеме: 10kΩ, 100kΩ, 4.7kΩ — проектные кандидаты | Неизвестно | Проверить startup/power-off; это проектные значения, не измеренные |

Защитные компоненты B13–B18 включены в проектный вариант. Неоднозначность «обязательные/рекомендуемые» §5 не закрыта изменением ТЗ: статус HW-010/HW-011 в матрице остаётся «уточнить обязательность».

## 3. Сенсоры, кнопка и privacy indicator

| ID | Компонент / кандидат | Количество | Статус наличия | Критерий выбора / ближайшая проверка |
|---|---|---:|---|---|
| B24 | VEML7700, готовый breakout с 3.3V logic | 1 | Неизвестно | Lux, I²C 0x10; схема конкретной платы и pull-ups; S06 |
| B25 | BME280, готовый breakout | 1 | Неизвестно | Temperature + humidity одним sensor, address 0x76/0x77; не BMP280; S07 |
| B26 | SGP40, готовый breakout | 1 | Неизвестно | VOC Index, I²C 0x59; raw signal→local algorithm, compensation; S08 |
| B27 | Panasonic EKMC1601111, PIR | 1 | Неизвестно | Кандидат 3.3V digital input, output load/warmup; корпус/дальность на стенде; S09 |
| B28 | Normally-open momentary button | 1 | Неизвестно | GPIO→GND, программный debounce; короткий test lead |
| B29 | Отдельный видимый CV LED | 1 | Неизвестно | Независим от LED-лент и power indicator камеры; виден из рабочей зоны |
| B30 | CV LED resistor, кандидат 1kΩ | 1 | Неизвестно | Проверить Vf и видимость при малом токе; увеличить сечение нагрузки через driver при необходимости |
| B31 | Sensor/button harness + prototype board | 1 комплект | Неизвестно | Низкотоковые сигналы; не использовать для передачи питания LED-зон |

Пять обязательных измерений дают **четыре sensor devices**: VEML7700 — свет, BME280 — температура и влажность, PIR — motion, SGP40 — air quality. Датчик температуры Pi не заменяет комнатный sensor. SPI sensor не требуется закупать лишь для заполнения всех bus types: выбранные датчики находятся на I²C/GPIO, SPI0 резервируется под LED.

## 4. Камеры и аудиостенд

| ID | Компонент / кандидат | Количество | Статус наличия | Критерий выбора / ближайшая проверка |
|---|---|---:|---|---|
| B32 | USB UVC webcam | 1 тип для обязательной проверки | Неизвестно | Exact model после инвентаризации; capture permissions/formats на выбранной Ubuntu |
| B33 | Pi Camera, кандидат Camera Module 2 либо 3 | 1 тип для отдельной проверки | Неизвестно | S10/S11; подходящий CSI cable Pi4, supported OS packages, один источник за раз |
| B34 | CSI cable соответствующих разъёмов | 1 для B33 | Неизвестно | Cable Pi4 не выбирать по комплекту Pi5; длина/ориентация по конкретной камере |
| B35 | USB DAC/USB audio output | 1 для проверки | Неизвестно | Linux compatibility/source routing и суммарный USB budget |
| B36 | Аналоговые колонки/наушники и 3.5mm cable | 1 комплект | Неизвестно | Проверка встроенного аналогового output обязательна отдельно от DAC |
| B37 | Bluetooth output device | 1 для проверки | Неизвестно | Проверить роли/concurrency совместно с phone input |
| B38 | Телефон с Bluetooth A2DP | 1 | Неизвестно | Pairing/metadata/controls, не только discovery |
| B39 | Powered USB hub | Условно | Неизвестно | Только если camera+SSD+DAC превышают подтверждённый power budget; исключить backfeed |

USB и Pi Camera нужны для подтверждения двух типов поддержки, не для обязательной одновременной работы двух камер. Финальный выбор production camera делается по качеству/latency и доступности.

## 5. Инструменты и окружения проверки

| ID | Что требуется | Количество / наличие | Для чего |
|---|---|---|---|
| B40 | Multimeter, измерение DC voltage/сопротивления и допустимый способ измерения тока | Неизвестно | Проверка polarity/drop/current; ток выше предела мультиметра — через подходящий shunt/meter |
| B41 | Current-limited low-voltage bench supply либо подходящий защищённый test PSU | Неизвестно | Первое включение коротких LED-участков |
| B42 | Logic analyzer/oscilloscope с достаточной полосой | Неизвестно, желательно | DATA timing/levels и поиск flicker на межзонном кабеле |
| B43 | Linux x86_64, Windows x86_64, macOS arm64, macOS x86_64 | Наличие всех неизвестно | Реальные обязательные screen-client acceptance targets |
| B44 | Внешний monitor и HDMI/USB-C adapter по ноутбуку | Неизвестно | Capture/mapping/monitor-off/latency стенд |

## 6. Порядок получения компонентов

1. **До Д006:** подтвердить B01–B08 и доступность измерительного инструмента. Для отсутствующей Pi simulation не заменяет M0 hardware evidence.
2. **До Д008:** короткие LED-отрезки, test PSU, shifter, resistor/capacitor, защищённые test leads и audio output. Большой PSU и полная лента для spike не обязательны.
3. **До Д027–Д034:** отмеренные B09/B10, рассчитанный PSU, вся power distribution/protection, mounting, кнопка/indicator.
4. **До Д036–Д045:** phone, DAC, analog/Bluetooth outputs, разрешённый Spotify test account.
5. **До Д061–Д068:** доступ ко всем четырём laptop targets; это организация стенда, не автоматическая закупка четырёх компьютеров.
6. **До Д081–Д095:** оба camera types для проверки, sensor breakouts и harness.

Сроки доставки не установлены без выбора продавцов/региона. Замены допустимы после сверки voltage, interface, physical pinout, address, driver и приёмочных требований; факт замены записывается в BOM.
