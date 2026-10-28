# Д008. Протокол реального LED/audio HIL

**Текущий статус: все10 hardware cases BLOCKED.** Пользователь сообщил: Pi пока нет либо не настроена, SSH target отсутствует. Выполненный Go/PipeWire host spike не меняет этот статус.

## Preconditions до подключения

Actual Pi4B/2GB/Ubuntu baseline Д006, working cooling Д007, короткие known-count segments, подходящий отдельный LED PSU/ограничение тока, shifter/resistor/capacitors/common ground, выбранные wire/fuse/connector ratings и проверенный power sequencing. Exact SKU/counts/cable lengths заполнить в local fixture record; unknown room geometry не брать из120LED example.

## Последовательность

1. Снять питание, проверить continuity/polarity/GND, отсутствие связи +5V двух PSU и upstream/downstream DATA direction.
2. Проверить шифтер/OE/partial-power behavior по реальной схеме. Known backfeed risk сначала устранить isolation/interlock; не экспериментировать с опасным порядком включения ради галочки.
3. Запустить initialized zero frame на малой brightness и пределах короткого test PSU.
4. A red/B blue → A OFF/B green → разные brightness/простая animation → поменять роли. Записать logical offsets и RGB order.
5. Проверить analog3.5mm playback сначала отдельно, затем вместе с LED. Не закрывать analog требование одним USB DAC.
6. В явном pairing window подключить телефон, проверить input→playback→общий PCM. Outside-window pairing denied.
7. По наличию повторить USB/Bluetooth outputs и Bluetooth role concurrency; Spotify требует разрешённого account/client. Каждый недоступный источник/output остаётся отдельным BLOCKED.
8. Остановить/восстановить DSP consumer, затем output/client/driver по одному; playback и basic lighting должны сохраняться в оговорённом failure scope.
9. Измерить actual transmitted frames, CPU/RAM/temperature и audio/PCM continuity под short joint load. Full CV/Ambilight/rules benchmark ещё позже.
10. Проверить intended service user/device access и cleanup, согласованный физический LED disconnect и restart.

Заполнить D08-01…D08-10 по [подробному плану](../plans/DAYS_004_010_DETAILED.md); каждому нужны outcome/evidence/versions/known limitations. Только тогда менять physical status ADR-007. Длительная desk/monitor wiring и финальная limiter calibration остаются Д034.

Read-only подготовка: `python3 experiments/day008/hil_preflight.py`. Она намеренно не выдаёт аппаратный PASS даже при обнаруженной Pi и не управляет линиями без реальной схемы.
