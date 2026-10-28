# Д004. LED/audio hypotheses — результат

Дата: 25.10.2026. [Код и evidence](../../experiments/README.md), [результат replay](../../experiments/results/replay.json), [ADR-007](../architecture/adr/007-led-topology.md).

| Вариант | Вывод | Evidence / статус |
|---|---|---|
| Два PWM channels | Известный конфликт с встроенным analog audio выбранного driver; не принимать как full-TZ solution без нового доказательства | [Исследование Д002](../../hardware/research/SOURCES.md), DOC-ONLY |
| SPI chain, два logical slices | Candidate сохраняется. Logical independence и power-limiter math действительно проверены Go spike | PASS-SIM; физический SPI/audio BLOCKED |
| PCM transport | Резерв; конфликт I²S и один physical channel требуют проверки | DOC-ONLY |
| External LED controller | Возможная замена, не включён в обязательный MVP автоматически | Не выбран |

Go-проба сформировала monitor120/desk120 как **synthetic counts**, сохранила независимость второй зоны при OFF первой, отклонила overlap/invalid counts/duplicate IDs/nonfinite current и применяла сначала zone4A, затем global6A. Получилось ≤6A по модели, не по амперметру. 1000 random fixtures подтверждают upper bound и отсутствие повышения каналов.

Audio primary candidate: host PipeWire/WirePlumber+BlueZ, Core управляет через узкий adapter, общий monitor tap идёт в DSP. На доступном x86_64 host действительно создана isolated graph: synthetic1000Hz→null sink→monitor PCM, consumer прекращался/подключался заново, playback оставался жив. Это не физическое звучание, не A2DP и не доказательство source/output matrix.

В изоляции понадобился session manager: без него клиентские stream ports не становились готовы. После собственного D-Bus/WirePlumber и отключения device monitors граф заработал. Вначале также устранена ошибка cleanup BrokenPipe при завершении consumer; итоговые два replay не содержат FAIL.

На host нет `/dev/snd`, GPIO/I²C/SPI/camera devices; существующий user PipeWire/D-Bus недоступен из sandbox. Host services не менялись. Доступ к system Bluetooth не проверен: нет разрешённого target adapter/session. Pi пользователь подтвердил недоступной/ненастроенной, SSH target не передан.

**Следующий gate:** реальная Pi + короткие segments + analog/USB/Bluetooth outputs, PCM и privileges по [Д008](day008-hil-protocol.md). ADR-007 физическая часть остаётся provisional.
