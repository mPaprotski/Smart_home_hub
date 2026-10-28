# День 7 — Thermal/hardware access

Дата: 25 октября 2026 года. **Статус: PARTIAL.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Реально выполнены synthetic thermal transitions/hysteresis/error cases и unit tests.
- 50/65/75°C и3°C hysteresis — project defaults, не уже откалиброванный hardware.
- Invalid sensor требует fan ON; физическая fail-on способность driver не доказана.
- GPIO/I2C/SPI/fan/load на Pi не запускались, sysfs не писался. Подготовлен exact access/permissions protocol.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D07-01 | Physical pinmux/wiring | BLOCKED | No target board; draft map preserved |
| D07-02 | GPIO input/output | BLOCKED | No gpiochip devices; no writes performed |
| D07-03 | Sensor I2C read | BLOCKED | No bus/module; no blind scans |
| D07-04 | SPI timing/rights | BLOCKED | No spidev target; host logic is synthetic only |
| D07-05 | Thermal policy | PASS-SIM | State/hysteresis/error tests executed without sysfs writes |
| D07-06 | Physical thermal load/fan | BLOCKED | No Pi/fan; deliberately no heating unrelated host |
| D07-07 | Permissions matrix | DOC-ONLY | Host/container design recorded; target negative-access tests NOT_RUN |

## Артефакты

- [research/pi-hardware-access.md](research/pi-hardware-access.md)
- [research/thermal-policy-v0.md](research/thermal-policy-v0.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
