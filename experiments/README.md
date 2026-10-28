# Эксперименты Д004–Д010

Реальные изолированные host/simulation пробы. Это не production Smart Home Core и не эмуляция доказательств Raspberry Pi.

## Запуск

Из корня проекта:

```bash
python3 experiments/run_all.py
python3 experiments/day010/replay.py --live-capture
python3 deploy/bootstrap/pi_preflight.py --require-pi
```

Первый запуск пишет redacted JSON в `experiments/results/latest.json`. Replay запускает две независимые suites и пишет `run-1.json`, `run-2.json`, `replay.json`. `--live-capture` пробует один кадр в RAM, публикует только размер/формат; при отсутствии доступа возвращает BLOCKED без screenshot. Flag не обязателен для synthetic analyzer.

Для PipeWire/HTTPS нужны разрешённые Unix/loopback sockets. В ограниченном sandbox эти пробы могут быть BLOCKED; это не доказательство дефекта host audio/TLS. Разрешение требуется только локальным ephemeral servers, не production network и не системным службам. В этой сессии два повторных прогона выполнены с разрешением таких sockets.

`run_all.py` возвращает 1 при FAIL, 0 при отсутствии FAIL даже если есть BLOCKED; `--require-m0` возвращает 2, пока нет target acceptance. Код 0 не означает готовность M0. `pi_preflight --require-pi` возвращает 2 на другом компьютере, ничего не устанавливает и не пишет GPIO.

## Проверки

| Файл | Проверяет | Не доказывает |
|---|---|---|
| day004/zone_frame_probe.go | Independent slices, mapping validation, idle-aware zone/global limit | WS2812B/DMA/power wiring на Pi |
| day004/pcm_probe.py | Synthetic FFmpeg→48kHz stereo S16LE, tone/silence | Физический playback/BlueZ |
| day004/pipewire_probe.py | Собственный null sink, playback→monitor PCM, consumer reconnect | Аналоговый/USB/Bluetooth output или production headless session |
| day005/screen_probe.py | Weighted edge fixtures, строгий packet validator, optional RAM capture | Cross-platform готовый client или full Ambilight |
| day005/inference_probe.py | Отдельный Python worker, bounded queue, независимый 30Hz synthetic loop и failure | Качество девяти классов Scene ML или ONNX runtime |
| day006/environment_probe.py | Read-only redacted versions/devices | Installed Pi |
| day007/thermal_policy_probe.py | Состояния, hysteresis, invalid sensor→fan ON | Реальное вращение/нагрев/fail-safe circuitry |
| day008/hil_preflight.py | Явно открытые physical cases и target detection | Любой hardware PASS |
| day009/https_probe.py | Ephemeral CA/SAN, negative verification, loopback proxy/backend, cert replacement | mDNS/LAN/mobile trust или Caddy installation |
| tests + Go test | Edge cases и 1000 deterministic random limiter cases | Coverage будущего продукта |

## Изоляция и зависимости

Python 3.12 и стандартная библиотека для большинства проб; Pillow нужен только optional capture, cryptography — TLS. FFmpeg и PipeWire tools используются, если установлены. Ничего автоматически не устанавливается.

PipeWire spike написан/проверен с host PipeWire1.0.5 и WirePlumber0.4.17. Копирует только **системные публичные defaults** `/usr/share/wireplumber` во временную директорию, отключает ALSA/V4L2/libcamera/BlueZ monitor, создаёт собственный session D-Bus и null sink. System D-Bus указывает на несуществующий temp socket. Нет подключения к host PipeWire, колонкам или микрофону. Host audio и system/user services не меняются.

Конфигурация WirePlumber0.4 **не переносится** на 0.5 без миграции. Для кандидата Ubuntu26.04.1 требуется отдельная проверка installed versions/0.5 config. Этот spike — воспроизводимое host evidence, не production installer.

Go запускается обычным бинарником; на этом компьютере snap launcher не работал, прямой `/snap/go/current/bin/go` работал. Cache в `experiments/.cache`, без загрузки сторонних Go dependencies. Если toolchain иначе установлен, runner использует доступный `go`.

Временные PCM/screen buffers остаются в RAM, private TLS keys — mode0600 во временной директории, удаляемой после пробы. JSON результатов содержит technical scalar summaries, не user audio/screenshots. Синтетические RGB/temperature fixtures можно сохранять: они не получены с камеры/захвата пользователя.

Все процессы создаются probe и завершаются им; host service configs/trust stores не меняются. На целевой Pi аппаратные действия выполняются отдельно по [протоколу](../docs/research/day008-hil-protocol.md), только после реальных замеров/проверки схемы.
