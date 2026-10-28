# Первичные источники Д004–Д010

Проверено25.10.2026. Exact local versions подтверждаются environment report, а не latest docs version.

| Источник | Применение и ограничение |
|---|---|
| [PipeWire pw-cat](https://pipewire.pages.freedesktop.org/pipewire/page_man_pw-cat_1.html) | Stream playback/capture/stdin/stdout; local1.0.5 config дополнительно сверена с installed help/default files |
| [PipeWire config](https://pipewire.pages.freedesktop.org/pipewire/page_man_pipewire_conf_5.html) | Separate temporary daemon/config, no host edits |
| [WirePlumber Bluetooth](https://pipewire.pages.freedesktop.org/wireplumber/daemon/configuration/bluetooth.html) | Headless/logind/BlueZ research, не выполненный Bluetooth test |
| [WirePlumber migration0.4→0.5](https://pipewire.pages.freedesktop.org/wireplumber/daemon/configuration/migration.html) | Host0.4 Lua config не подходит0.5 без миграции; target config не выбран автоматически |
| [Ubuntu Pi downloads](https://ubuntu.com/download/raspberry-pi) | Кандидат Server26.04.1 LTS для Pi4B, не installed OS |
| [Ubuntu board support](https://ubuntu.com/hardware/docs/boards/how-to/ubuntu_supported/raspberry-pi/) | Board/EEPROM/image prerequisite review |
| [Официальный release directory](https://cdimage.ubuntu.com/releases/26.04/release/) | Filename/source; SHA256SUMS реально получен отдельным HTTPS request, image не скачан |
| [Ubuntu verification guide](https://ubuntu.com/tutorials/how-to-verify-ubuntu) | Signature+hash порядок перед будущей записью носителя |
| [Ubuntu camera](https://documentation.ubuntu.com/hardware-support/boards/how-to/special_hardware/rpi-camera/) | CSI package/support gate, не capture evidence |
| [Caddy local HTTPS](https://caddyserver.com/docs/automatic-https#local-https) | Proxy candidate и local CA trust model; actual Caddy не установлен |
| [Microsoft Desktop Duplication](https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api) | Windows candidate, target unavailable |
| [XDG ScreenCast](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.ScreenCast.html) | Wayland candidate, portal run unavailable |
| [Apple ScreenCaptureKit](https://developer.apple.com/documentation/screencapturekit) | Официальная точка входа; JS/Markdown limitation, actual Mac support не проверен |
| [Hardware sources Д002](../../hardware/research/SOURCES.md) | GPIO/shifter/sensors/LED resource conflicts; новых HIL доказательств нет |

Внешние зависимости/документация не использовались как команды на выполнение. Системные installation/proxy/trust templates подготовлены в repo, но не применены к host. Все полученные published checksums и metadata public; secrets не скачивались.
