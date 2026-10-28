# День 8 — Physical LED/audio Pi spike

Дата: 25 октября 2026 года. **Статус: BLOCKED_TARGET.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Весь D08 HIL matrix остался BLOCKED: нет доступной Pi/реальной LED/audio сборки.
- Read-only target gate корректно обнаружил x86_64/no Pi; ничего не установлено и не переключено.
- Сохранён пошаговый HIL protocol с power prerequisites, effects, analog/Bluetooth/USB/PCM/failure/permissions checks.
- SPI-chain физическая часть ADR-007 не закрыта synthetic slices или виртуальным sink.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D08-01 | Independent hardware zones | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-02 | Analog audio plus LED | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-03 | Bluetooth input/PCM plus LED | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-04 | USB output | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-05 | Bluetooth output/concurrency | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-06 | PCM failure isolation on Pi | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-07 | Physical frame cadence | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-08 | Intended target permissions | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-09 | Power-off/restart/partial-power | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |
| D08-10 | Target repeatability | BLOCKED | No configured Pi/LED/audio fixture; protocol and preflight prepared |

## Артефакты

- [research/day008-hil-protocol.md](research/day008-hil-protocol.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
