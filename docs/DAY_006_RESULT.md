# День 6 — Headless Pi setup

Дата: 25 октября 2026 года. **Статус: BLOCKED_TARGET_PREPARATION_DONE.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Пользователь подтвердил отсутствие доступной/настроенной Pi, SSH target не передан.
- Снят read-only redacted inventory текущего Ubuntu24.04.4 x86_64 host; это не Pi baseline.
- Выбран documentary candidate Ubuntu Server26.04.1 LTS arm64+raspi; published SHA256 получен из official release metadata.
- Image не скачивался, signature/image bytes не проверялись, storage не записывался; Pi/SSH/reboot не выполнялись.
- Готовы install/recovery guide, cloud-init/network templates и guarded read-only preflight.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D06-01 | Pi/OS/architecture | BLOCKED | Current host x86_64, no Pi access; published image metadata fetched only |
| D06-02 | Headless SSH | BLOCKED | No SSH target; no board installation performed |
| D06-03 | Target storage | BLOCKED | No Pi storage selected or written |
| D06-04 | Pi power/cooling | BLOCKED | No physical PSU/fan/thermal evidence |
| D06-05 | Pi LAN/clock | BLOCKED | Current host metadata is not Pi evidence |
| D06-06 | Reinstall guide | DOC-ONLY | Private seed templates, install/recovery guide, read-only preflight and candidate image manifest ready |

## Артефакты

- [operations/pi-install.md](operations/pi-install.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
