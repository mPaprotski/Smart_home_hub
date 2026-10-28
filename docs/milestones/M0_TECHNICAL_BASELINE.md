# M0 — техническая база

Дата: 25.10.2026. **M0_NOT_READY.** Ревизия выполнена, аппаратный рубеж не закрыт. User подтвердил отсутствие доступной/настроенной Pi. [JSON](../../experiments/results/m0.json), [cases](../../experiments/results/case-matrix.json), [сводка](../DAYS_004_010_RESULT.md).

| Gate | Область | Статус | Причина / результат | Evidence |
|---|---|---|---|---|
| G01 | Platform | BLOCKED | No Pi install/headless/reboot evidence | [DAY_006_RESULT.md](../DAY_006_RESULT.md) |
| G02 | Cooling/access | PARTIAL | Thermal policy PASS-SIM; physical cooling/bus tests BLOCKED | [DAY_007_RESULT.md](../DAY_007_RESULT.md) |
| G03 | LED independence | BLOCKED | Go mapping PASS-SIM; no physical WS2812B | [DAY_008_RESULT.md](../DAY_008_RESULT.md) |
| G04 | Audio/PCM on Pi | BLOCKED | Isolated host PCM passes; no physical analog/Bluetooth/USB target | [DAY_004_RESULT.md](../DAY_004_RESULT.md) |
| G05 | Target permissions | PARTIAL | Host isolation tested; target driver/container rights unknown | [research/host-container-boundaries.md](../research/host-container-boundaries.md) |
| G06 | LAN/mobile trust | BLOCKED | Loopback TLS passes; no real mDNS/phone/Pi LAN | [DAY_009_RESULT.md](../DAY_009_RESULT.md) |
| G07 | Capture/platform path | PARTIAL | Synthetic analyzer/local worker pass; live display and other OS unavailable | [DAY_005_RESULT.md](../DAY_005_RESULT.md) |
| G08 | Repeatability | PARTIAL | Two host suites same outcomes; physical clean-install/replay not run | [DAY_010_RESULT.md](../DAY_010_RESULT.md) |

## Повторяемость

Два прогона создавали отдельные временные process/socket/CA trees и завершали их. Это restart local probes, **не clean OS reinstall и не reboot Pi**. Host outcomes совпали, deterministic numerical fixtures совпали, все запускавшиеся code tests прошли. Durations включают build cache/начальную компиляцию и не являются сравнением performance production Hub.

## Решение о продолжении

Независимые задачи Д011 можно выполнять: repository scaffold, config/interfaces/storage/auth/simulation. Hardware-dependent claims, actual driver/container permissions, final LED topology и M2/M3 остаются открытыми до physical evidence. Нет основания списывать резерв как уже потраченные20 дней: фактическое ожидание hardware пока не известно.

После появления Pi: install/preflight→cooling/access→short LED/audio/PCM HIL→LAN/mobile trust→critical replay. Нужны фактические counts/PSU/cables/driver versions и закрытие partial-power вопросов. Это перенос blockers, не уменьшение объёма ТЗ.

## Уточнение hardware26.10.2026

Владелец подтвердил наличие Raspberry Pi 4B **только2GB RAM**. Настроенный SSH target/замеры не переданы, поэтому M0_NOT_READY сохраняется. [Ресурсный профиль](../development/pi4b-2gb.md). Старые JSON cases описывают предыдущую ревизию, а не отсутствие платы после этого уточнения.
