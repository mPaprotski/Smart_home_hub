# Д004/Д007. Host/container boundaries и доступ

Архитектура: [ADR-001](../architecture/adr/001-module-boundaries.md). Здесь реальный host evidence отдельно от design для Pi.

| Ресурс | Планируемый owner | Доступ Core/container | Фактический статус |
|---|---|---|---|
| GPIO LED/button/fan/indicator | Core hardware adapters | Только выбранные gpiochips/lines, без generic privileged | BLOCKED_NO_PI |
| I²C sensors | Sensor adapters Core | Только выбранный bus device, controlled addresses | BLOCKED_NO_PI |
| SPI0 LED | Lighting adapter | Только нужный spidev/binding; DMA requirements отдельно | BLOCKED_NO_PI |
| Host audio/BlueZ | Dedicated host service user | Узкий IPC/API; не весь host `/run`/system D-Bus | DOC-ONLY production; isolated PipeWire PASS-HOST |
| PCM | Audio→DSP bounded port | Memory stream, без filesystem recording | PASS-HOST synthetic monitor |
| Camera | Local Runtime single owner | Явный video/CSI доступ, consumer permissions от Core | BLOCKED_NO_CAMERA/PI |
| SQLite | Core Storage | Только persistent DB volume у Core | Design, Core не реализован |
| Bot/frontend | API only | Нет hardware/DB/CA-key mounts | Design |
| Thermal sysfs | Core Thermal | Read-only exact thermal source | На host есть thermal paths, но это не Pi readings |
| TLS CA/server key | Proxy/maintenance | Минимальные files/UIDs; CA private не передаётся clients | Host ephemeral TLS PASS, production DOC-ONLY |

Host: Ubuntu24.04.4 x86_64, PID1 не systemd. `systemctl --user`/host PipeWire/system Bluetooth недоступны из sandbox. Docker CLI обнаружен, daemon не подтверждён; container privileges/isolation и production Compose не тестировались. Прямой Go binary работает, snap launcher не работал; это toolchain environment issue, не defect Hub.

Probe изоляция: own temp runtime dir, own PipeWire remote, own D-Bus, ALSA/camera/BlueZ discovery выключены. Permission для временных sockets выдано отдельно; это не выдача Root/privileged production Core. Runtime probe processes non-root в текущем task environment, но **не** доказательство будущего Pi driver permissions.

На Pi: проверить actual pinmux/device ownership/ACLs после installation, negative access лишнему user, возможность non-root driver. Если DMA library требует broad memory access, оформить narrow helper/transport alternative в ADR и проверить threat boundary; не объявлять полный privileged контейнер «минимальными правами».
