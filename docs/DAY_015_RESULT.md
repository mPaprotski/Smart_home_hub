# Д015. Fakes/native/Compose/CI — результат

Вертикальный срез работает: seeded sensor → typed envelope → настоящий Core Event Bus → observer → bounded observation store. Дополнены LED frame ownership/counter, rotated/serpentine matrix fixtures, synthetic audio spectrum/beat/BPM, четыре presence states, screen epoch/seq/error и MQTT discovery/state/disconnect. Quality/offline injection и lifecycle протестированы.

| Проверка | Результат / evidence |
|---|---|
| D15-01 Adapter boundaries | PASS-HOST; lifecycle/LED/storage ports, без hardware-specific dependencies |
| D15-02 Sensor→Core | PASS-SIM; deterministic seed/clock fixture и actual asynchronous bus observer |
| D15-03 Fake operations/lifecycle | PASS-SIM; finite values, all4 presence states, stale/offline/error, frame ownership/mapping, stop |
| D15-04 Mode separation | PASS-SIM; explicit simulation, memory_fake/persistent=false; production/hardware fail closed |
| D15-05 Native smoke | PASS-SIM; real process start/ready/event/stop, отдельно от handler tests |
| D15-06 Compose | PASS-COMPOSE; actual build/healthy/run/events/restart/SIGTERM/down, nonroot/read-only/loopback inspect |
| D15-07 Local quality | PASS-HOST;23 Go tests+race,14 Python cases, lint/types/format/build, OpenAPI |
| D15-08 Workflow protection | PASS-HOST; actionlint OK, pinned official action SHA, minimal permissions/no production secrets/hosted runner |
| D15-09 Remote CI | BLOCKED; source uncommitted/not pushed, remote workflow не запускался; plan не разрешает публикацию этим этапом |
| D15-10 Handoff | DOC-ONLY; [Д016 inputs/limits](DAYS_011_015_RESULT.md#передача-на-д016), quickstart, contracts и evidence готовы |

[Native JSON](results/days011-015/native-smoke.json), [Compose JSON](results/days011-015/compose-smoke.json), [workflow](../.github/workflows/ci.yml), [commands](development/README.md). Непостоянные fake storage/LED fixtures не заменяют SQLite/Lighting Engine. Полный simulation UI, реальные ML/MQTT и Pi HIL — будущие этапы. M0_NOT_READY сохранён, PASS-CI не заявлен.
