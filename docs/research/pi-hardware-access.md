# Д007. Hardware access — подготовка и фактический статус

Текущий host: x86_64 Ubuntu, не Pi; `/dev/gpiochip*`, `/dev/i2c-*`, `/dev/spidev*`, `/dev/video*`, `/dev/snd/*` отсутствуют в доступном окружении. Никакие GPIO outputs, fan или power rails не переключались. Наличие host thermal paths не делает их датчиком Pi.

Read-only gate: `python3 deploy/bootstrap/pi_preflight.py --require-pi`. На текущем host ожидаемый exit2/BLOCKED_NO_PI. Детектор не устанавливает packages и не пробует bus addresses. Если Pi будет обнаружена, это только TARGET_DETECTED_NOT_ACCEPTED: далее нужен HIL, не автоматический PASS.

## Когда появится Pi

1. Проверить board/model/Ubuntu/kernel/EEPROM/storage/power/fan и [pin map](../../hardware/wiring/pin-map.md).
2. Снять фактический pinmux/overlays; не считать все зарезервированные alternate functions свободными.
3. Включить только выбранные interfaces средствами конкретной Ubuntu, сохранить before/after и recovery path.
4. Проверить низкотоковый button/indicator input/output, затем known sensor register reads. Массовый blind I²C probing не использовать как единственное evidence.
5. Сверить exact breakout supply/pull-ups/ID/addresses, не подать5V на GPIO. Реальные GPIO library/chip indices подобрать по обнаружению, не hardcode gpiochip0.
6. Для SPI0 сначала test fixture/права, затем LED transport вместе с audio. Other slaves на LED bus не добавлять.
7. Проверить ordinary service user positive access и посторонний user denial; container device allowlist/IPC scopes описать по actual calls.
8. Сравнить systemd/headless audio session и actual D-Bus/BlueZ roles. Текущий isolated null sink не заменяет этот шаг.

## Permissions record template

Для каждого ресурса: exact path/line, owner/group/ACL, требуемые operations, service UID, container mount/device/capability, positive check, negative check, cleanup/revoke, residual risk. Если DMA driver требует широких memory privileges — unresolved design issue по ADR-007/вопросу Q-026, не «обоснование» unrestricted privileged всех services.

Physical fan/thermal/load/bus tests пока BLOCKED. Достаточная симуляция thermal policy описана отдельно в [thermal-policy-v0](thermal-policy-v0.md).
