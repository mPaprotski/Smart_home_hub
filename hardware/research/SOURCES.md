# День 2. Источники аппаратной проработки

Проверено 25 октября 2026 года. Использованы документация производителей и исходный репозиторий драйвера. Это документальная проверка, не проверка экземпляров на стенде. Версии ОС/kernel/firmware, ревизии плат и реальные схемы breakout необходимо зафиксировать на Д006–Д008.

| ID | Первичный источник | Для чего использован |
|---|---|---|
| S01 | [Raspberry Pi hardware / GPIO](https://www.raspberrypi.com/documentation/computers/raspberry-pi.html#gpio-and-the-40-pin-header) | BCM/physical mapping, 3.3V logic, отсутствие 5V tolerance, I²C/SPI |
| S02 | [rpi_ws281x README](https://github.com/jgarff/rpi_ws281x) | Сравнение PWM/PCM/SPI, audio conflicts, SPI0 timing/permissions |
| S03 | [TI SN74AHCT125 datasheet](https://www.ti.com/lit/ds/symlink/sn74ahct125.pdf) | TTL input, питание, OE, pinout; проверены страницы 3–5 |
| S04 | [Adafruit NeoPixel power](https://learn.adafruit.com/adafruit-neopixel-uberguide/powering-neopixels) | Оценочный full-white ток, запас по питанию |
| S05 | [Adafruit NeoPixel best practices](https://learn.adafruit.com/adafruit-neopixel-uberguide/best-practices) | DATA resistor и bulk capacitor |
| S06 | [SparkFun VEML7700 hardware](https://github.com/sparkfun/SparkFun_Ambient_Light_Sensor-VEML7700/blob/main/docs/hardware_overview.md) | 3.3V breakout, I²C 0x10, pull-up jumpers |
| S07 | [Bosch BME280 datasheet](https://www.bosch-sensortec.com/media/boschsensortec/downloads/datasheets/bst-bme280-ds002.pdf) | Temperature/humidity, supply, 0x76/0x77, I²C mode |
| S08 | [Sensirion SGP40 datasheet](https://sensirion.com/media/documents/296373BB/6203C5DF/Sensirion_Gas_Sensors_Datasheet_SGP40.pdf) | 3.3V, I²C 0x59, VOC Index pipeline, compensation |
| S09 | [Panasonic EKMC1601111 specification](https://mediap.industry.panasonic.eu/assets/imported/industrial.panasonic.com/cdbs/www-data/pdf/EWA0000/bltn_eng_ekmc160611_ast-ind-247327.pdf) | Digital PIR, supply/output/warmup; документ ревизии 7.1 при проверке |
| S10 | [Raspberry Pi camera documentation](https://www.raspberrypi.com/documentation/accessories/camera.html) | CSI cameras/connectors, Camera Module 2/3 |
| S11 | [Ubuntu camera support](https://documentation.ubuntu.com/hardware-support/boards/how-to/special_hardware/rpi-camera/) | CSI/libcamera/rpicam и ограничения Python camera stack |
| S12 | [Raspberry Pi 15W USB-C supply](https://www.raspberrypi.com/products/type-c-power-supply/) | Отдельное питание Pi: кандидат 5.1V/3A |

## Что установлено по документации

**LED/audio:** S02 описывает два канала PWM, один PCM и один SPI. PWM конфликтует со встроенным аналоговым audio; PCM занимает ресурс I²S. SPI допускает аналоговый audio, но LED занимают выбранную SPI bus. Для Pi 4 README указывает ограничения SPI clock и отдельную настройку; конкретные boot paths/параметры надо подтвердить на выбранной Ubuntu. Не переносить инструкции Raspberry Pi OS в Ubuntu без проверки. DMA и permissions требуют отдельного spike.

**Уровни:** S01 подтверждает 3.3V GPIO. S03: SN74AHCT125 питается 4.5–5.5V, принимает high от 2V; поэтому 3.3V→5V — обоснованный кандидат. Это не эквивалент произвольного HC125 либо двунаправленного I²C level converter. Подключение OE/power sequencing остаётся частью схемы.

**Сенсоры:** S06/S07/S08 позволяют предварительно разместить VEML7700, BME280 и SGP40 на одной 3.3V I²C bus без совпадения адресов. Конкретные breakout boards могут иметь собственные regulators/pull-ups и другую маркировку VIN/VDD: их схемы нужно проверить отдельно. BME280 нужен именно с humidity, не BMP280.

SGP40 даёт сигнал для расчёта VOC Index, а не измеренный CO₂/PM2.5. Алгоритму требуется регулярный поток и environmental compensation; нельзя трактовать индекс как концентрацию конкретного газа. S09 допускает питание PIR 3.3V, указывает до 100µA output load и до 30 секунд startup instability.

**Камера:** S11 описывает camera stack начиная с Ubuntu 25.04 и отдельные условия Python integration. Это не выбор версии ОС для проекта и не рекомендация устанавливать устаревший interim release. На Д006 выбрать поддерживаемую Ubuntu и проверить именно её packages/kernel/CSI. USB UVC camera остаётся первым кандидатом для быстрого стенда; поддержка Pi Camera проверяется отдельно.

## Степень доказательства

- Документально: pinout, номинальные интерфейсы и конфликт PWM/audio.
- Расчётно: примерные токи, serial-chain mapping и время передачи.
- Не подтверждено: фактические компоненты, их наличие/цены, конкретная Ubuntu, driver privileges, audio coexistence, сигнал на реальных проводах, температурный режим и 30 FPS на Pi.

## Неудачный поиск и замена кандидата

Прямая страница BH1750FVI на ROHM при проверке вернула 404. Это не доказывает снятие датчика с производства. Для текущего BOM выбран лучше документированный кандидат VEML7700; BH1750 остаётся возможной заменой только после проверки первичного datasheet и фактического breakout.
