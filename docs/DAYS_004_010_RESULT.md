# Дни 4–10 — фактический результат

Дата: 25 октября 2026 года. **Выполнена доступная host/simulation/documentation часть всех семи дней. Аппаратные дни целиком закрыть нельзя: Pi пока нет либо она не настроена, доступа по SSH нет. M0_NOT_READY.**

Это execution report, а не очередной план: добавлен исполняемый spike code, проведены локальные эксперименты и сохранены технические результаты.

## По дням

| День | Статус | Отчёт |
|---|---|---|
| Д004 | PARTIAL | [Д004 — LED/audio feasibility](DAY_004_RESULT.md) |
| Д005 | PARTIAL | [Д005 — Screen/ML/measurement feasibility](DAY_005_RESULT.md) |
| Д006 | BLOCKED_TARGET_PREPARATION_DONE | [Д006 — Headless Pi setup](DAY_006_RESULT.md) |
| Д007 | PARTIAL | [Д007 — Thermal/hardware access](DAY_007_RESULT.md) |
| Д008 | BLOCKED_TARGET | [Д008 — Physical LED/audio Pi spike](DAY_008_RESULT.md) |
| Д009 | PARTIAL | [Д009 — Local HTTPS/mDNS](DAY_009_RESULT.md) |
| Д010 | REVIEW_DONE_M0_NOT_READY | [Д010 — Technical baseline M0](DAY_010_RESULT.md) |

## Что действительно работает в экспериментах

- Go: независимые slices двух зон, invalid mapping rejection, zone/global power limiter с idle floor; 1000 randomized bound cases.
- Audio: FFmpeg synthetic tone/silence и отдельный PipeWire null-sink graph с monitor PCM, consumer reconnect без остановки playback.
- Screen/ML: weighted synthetic edge analyzer, packet validation, local Python demo inference с bounded queue/1Hz и failure, независимый synthetic30Hz loop.
- Thermal: four-state transitions, hysteresis и invalid input→fan ON demand на synthetic данных.
- HTTPS: ephemeral CA/SAN, loopback reverse proxy/backend, отказ недоверенному CA/неверному имени, закрытые control routes, замена сертификата и cleanup.
- Два независимых replay, 17 Python tests и3 Go tests с race detector проходят. Это не coverage будущего Core и не physical acceptance.

## Что подготовлено для оборудования

Install/image metadata/verification/recovery guide, private cloud-init/network templates, read-only Pi preflight, exact HIL checklist, thermal/access/permissions plan, Caddy/Avahi examples, platform/protocol/ML-quality/dataset/performance documents. Все examples явно не применены к системе.

[Ubuntu image candidate](../deploy/bootstrap/image-manifest.candidate.json): published hash fetched; image bytes и signature не проверены, носитель не записан. WirePlumber host0.4 recipe не переносится на target0.5 автоматически. Production Caddy/Docker/systemd deployment не выполнен.

## Evidence и воспроизведение

- [Experiments README](../experiments/README.md), [runner](../experiments/run_all.py), [replay](../experiments/day010/replay.py).
- [Run1](../experiments/results/run-1.json), [Run2](../experiments/results/run-2.json), [replay summary](../experiments/results/replay.json).
- [Case matrix](../experiments/results/case-matrix.json), [M0 gates](milestones/M0_TECHNICAL_BASELINE.md), [источники](research/SOURCES_D004_D010.md).

Для isolated PipeWire/TLS нужны разрешённые local sockets; в restricted sandbox возможен BLOCKED. Инструменты не меняют host services/trust и не читают microphone. Generated personal media/keys/DB не добавлены в repo; technical summaries redacted.

## Невыполненные обязательные проверки

Реальные Pi headless install/reboot/storage/power, fan/GPIO/I²C/SPI и physical LED/audio/PCM, actual A2DP/USB/Bluetooth outputs, signal/power sequencing, mDNS/mobile trust, реальные четыре capture targets. Generic Linux screen capture attempt вернула OSError; это отдельный открытый пункт, не synthetic PASS.

Нет raw camera/screen/audio recordings, actual model quality scores и выдуманного FPS/latency для физической ленты. Зафиксированные решения Д003 остаются проектными; hardware часть ADR-007 не утверждена как проверенная.

## Статусы cases

| Статус | Количество |
|---|---:|
| BLOCKED | 28 |
| DOC-ONLY | 8 |
| PASS-HOST | 9 |
| PASS-SIM | 3 |

Cases с PASS-HOST/PASS-SIM относятся только к явно указанному scope. Ни одного PASS-PI нет. Source ТЗ и исторические reports Д001–Д003 сохранены.
