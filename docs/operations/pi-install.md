# Д006. Headless Ubuntu Pi4 — инструкция установки

**Hardware baseline от26.10.2026: только Raspberry Pi 4B с2GB RAM.** [Ресурсный профиль](../development/pi4b-2gb.md). Pi совместные memory/performance gates ещё не выполнены.

Статус: **подготовлено, на Pi не выполнено**. Плата Pi4B/2GB есть по подтверждению26.10.2026; настроенного SSH target пока нет. Доступное окружение — x86_64 Ubuntu24.04.4, не целевая плата.

## Выбранный кандидат образа

Ubuntu Server26.04.1 LTS, **preinstalled-server-arm64+raspi**, а не generic arm64 ISO/desktop/x86 image. На дату25.10.2026 официальный [Ubuntu Pi download](https://ubuntu.com/download/raspberry-pi) и [board matrix](https://ubuntu.com/hardware/docs/boards/how-to/ubuntu_supported/raspberry-pi/) указывают поддержку Pi4B. Это документальный выбор; exact EEPROM/kernel/audio/camera compatibility проверяется на реальной плате.

[Manifest кандидата](../../deploy/bootstrap/image-manifest.candidate.json) содержит опубликованный SHA256, реально полученный по HTTPS из официального SHA256SUMS. **Образ1.4GB не скачивался, его bytes/checksum/signature не проверялись и носитель не записывался.** Published hash не является проверкой отсутствующего файла. Если release metadata изменится, обновить manifest с provenance до установки.

## Перед записью

1. Подтвердить Raspberry Pi 4B с 2GB, PSU/cooling, выбранный microSD/SSD, Ethernet и компьютер оператора. Заполнить только известные поля inventory.
2. Скопировать нужные прежние данные целевого носителя; точно определить его в системном disk inventory. Не угадывать `/dev/sdX` и не записывать image в рабочий SSD компьютера.
3. Скачать конкретный image, SHA256SUMS и detached signature из того же официального release directory.
4. Проверить подпись manifest доверенным Ubuntu signing key по [официальной инструкции](https://ubuntu.com/tutorials/how-to-verify-ubuntu), затем SHA256 image. Ключ доверия не выбирается только по совпавшему имени отправителя.
5. Сверить image filename/architecture/hash с candidate manifest. Несовпадение — выяснить версию, не отключать проверку.
6. Записать image штатным Imager/проверенным инструментом; эта инструкция не выполняет destructive write автоматически.

## Private headless seed

Templates: [user-data](../../deploy/bootstrap/user-data.example), [network-config](../../deploy/bootstrap/network-config.example). Скопировать в gitignored `deploy/bootstrap/private/`, заполнить публичный SSH key, локально сгенерированный password hash для console/sudo и actual interface name. Реальные seeds не коммитить: hash и network identifiers тоже sensitive.

Шаблон отключает SSH password login, не включает passwordless sudo, не устанавливает packages и не перезагружает Pi автоматически. Password hash placeholder нельзя оставить буквально. При выбранном image проверить cloud-init seed location/поддержку customisation; default user/password не предполагать. Hostname smarthome сначала проверить на конфликт в LAN.

## Первый запуск

1. Подключить охлаждение/сеть, загрузить Pi и дождаться first-boot provisioning.
2. Найти target в собственной LAN, сверить SSH host key через доверенный локальный путь, зайти отдельным admin user.
3. Подтвердить board/OS/architecture/clock/storage; `python3 deploy/bootstrap/pi_preflight.py --require-pi` — read-only helper, не полная приёмка.
4. Обновить нужные packages штатным способом выбранной Ubuntu, при необходимости reboot; сохранить exact versions. Рецепт host WirePlumber0.4 не переносить на target0.5 без миграции.
5. Проверить root/mount/free space, storage/voltage errors, Ethernet/Wi-Fi и temperature sysfs. Не использовать vendor utility, которой нет, как единственное доказательство отсутствия faults.
6. Снять idle CPU/RAM/temp/fan и USB budget. Camera/SSD/DAC вместе могут требовать отдельного питания/hub, но это надо измерить.
7. Перезагрузить, снова проверить headless SSH и state. Сохранить приватные host/IP/MAC/key сведения вне public reports.

## Recovery и следующие gates

Сохранить исходный verified image/seed/config и рабочий носитель до рискованных boot/kernel changes. При потере SSH использовать локальную консоль/known-good boot media, не менять firewall вслепую. Migrate microSD→SSD отдельно на Д139 после backup и проверки USB boot/power.

Production Core/Compose/auth ещё не установлены. Д006 закрывается только после actual headless boot/reboot/storage/power checks; заранее написанный guide остаётся DOC-ONLY. Д007/Д008 выполняются затем по hardware-access/HIL protocol. Не делать prototype camera snapshot на диск: policy проекта RAM-only сохраняется и при настройке камеры.
