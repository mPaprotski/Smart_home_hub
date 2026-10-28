# День 4 — LED/audio feasibility

Дата: 25 октября 2026 года. **Статус: PARTIAL.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Реально выполнен Go zone/frame/limiter spike, включая1000 deterministic random cases; current result5.959529A при synthetic limit6A.
- FFmpeg synthetic tone1000Hz и silence декодированы в48kHz stereo S16LE; raw PCM не записывался.
- Собственный PipeWire/WirePlumber/D-Bus graph дал playback→null sink→monitor PCM; consumer restart не остановил playback.
- Host PipeWire1.0.5/WirePlumber0.4.17; конфигурация0.4 не предназначена для target0.5 без миграции.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D04-01 | Two logical slices | PASS-SIM | Go fixture240LED, independent slices; physical driver not used |
| D04-02 | Physical local playback | BLOCKED | No soundcard exposed; virtual sink is separate host evidence |
| D04-03 | Monitor PCM | PASS-HOST | Isolated PipeWire null sink→monitor; 48kHz stereo S16LE, no physical source |
| D04-04 | Silence/consumer restart | PASS-HOST | FFmpeg silence and monitor consumer reconnect; physical output restart pending |
| D04-05 | Bluetooth input | BLOCKED | No target adapter/phone session; system D-Bus not available |
| D04-06 | Permissions inventory | DOC-ONLY | Own temp sockets tested; Pi/container device rights unknown |
| D04-07 | Transport/audio hypothesis | DOC-ONLY | SPI candidate preserved; physical proof deferred |

## Артефакты

- [research/led-audio-hypotheses.md](research/led-audio-hypotheses.md)
- [research/pcm-contract-v0.md](research/pcm-contract-v0.md)
- [research/host-container-boundaries.md](research/host-container-boundaries.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
