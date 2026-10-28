# День 9 — Local HTTPS/mDNS

Дата: 25 октября 2026 года. **Статус: PARTIAL.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Реально подняты ephemeral loopback HTTP backend и HTTPS proxy, проверен marker через trusted CA/SAN smarthome.local.
- Untrusted CA/wrong hostname отвергнуты; control routes отсутствуют, POST405. Leaf certificate заменён с сохранением CA trust.
- Private files0600 в temporary directory удалены; system trust store/host services/LAN не менялись.
- mDNS/Pi/phone/laptop trust и production Caddy не тестировались. Подготовлены Caddy/Avahi examples и runbook.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D09-01 | LAN mDNS resolution | BLOCKED | No target publisher/mobile LAN; templates not applied |
| D09-02 | Pi phone/laptop trusted TLS | BLOCKED | Only loopback CA/SAN/proxy verified; no mobile/system trust installation |
| D09-03 | Negative TLS verification | PASS-HOST | Untrusted CA and wrong hostname rejected |
| D09-04 | Internal upstream isolation | PASS-HOST | Both listeners loopback-only, no control routes, POST405; Pi firewall pending |
| D09-05 | Target reboot/services | BLOCKED | No Pi/systemd target |
| D09-06 | Certificate replacement | PASS-HOST | Two leaf certificates signed by ephemeral CA; private temp files removed |
| D09-07 | Public artifact hygiene | PASS-HOST | Examples/redacted summaries; no generated keys/screenshots/PCM/DB persisted |

## Артефакты

- [operations/local-https-mdns.md](operations/local-https-mdns.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
