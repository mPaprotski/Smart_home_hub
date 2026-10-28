# Подробный рабочий план: Д004–Д010

Подготовлен 25 октября 2026 года. Основание: [ТЗ v1.0](../../Smart_Home_Hub_TZ.md), [общий план](../DEVELOPMENT_PLAN_DAILY.md), [результат Д003](../DAY_003_RESULT.md), [архитектура v0](../architecture/README.md), [BOM](../../hardware/bom/BOM.md), [открытые вопросы](../open-questions.md).

**Это план будущей работы, а не отчёт о выполнении дней 4–10.** Установка ОС, подключение оборудования, эксперименты и выдача сертификатов в рамках подготовки этого документа не выполнялись. Известно только, что замеров пока нет; наличие Pi/периферии/платформенных ноутбуков не подтверждено.

**Последующее исполнение:** [фактический результат Д004–Д010](../DAYS_004_010_RESULT.md). Исходный план сохранён; local probes выполнены, physical gates остаются открытыми.

## Как использовать документ

Один день — до 8 часов: 6 часов предметной работы по пяти блокам, ещё 2 часа на подготовку/поиск актуальных инструкций/перерывы/разбор ошибок/оформление результата. Блоки ниже — бюджеты, не обещание уместить произвольную аппаратную проблему в таймер.

Единый дополнительный бюджет 120 минут: 20 минут входной проверки и планирования, 40 минут документации и версий, 40 минут разбора неожиданных ошибок, 20 минут записи результатов/следующего шага. Если troubleshooting превышает запас, остановиться на ясном воспроизводимом состоянии и записать перенос.

| День | Главный вопрос | Основной результат | Зависимость |
|---|---|---|---|
| Д004 | Как совместить независимые LED-зоны, audio и общий PCM? | Spike-протокол, минимальные host-пробы, shortlist adapters | Контракты Д003; Pi ещё не обязательно установлена |
| Д005 | Как захватывать экран/выполнять ML и измерять качество? | Capture/runtime матрица, protocol v0, measurement/dataset plan | Д003, доступные ноутбуки |
| Д006 | Работает ли целевая headless Pi с воспроизводимой Ubuntu? | Installed baseline, version manifest и install guide | Реальная Pi, питание/носитель/LAN |
| Д007 | Доступны ли cooling/GPIO/I²C/SPI с нужными правами? | Thermal/access spike, pinmux/permissions map | Д006, физическое охлаждение |
| Д008 | Подтверждён ли LED/audio подход на реальной Pi? | Hardware coexistence evidence и transport/audio decision | Д004 + Д006–Д007, короткий защищённый стенд |
| Д009 | Работает ли локальный HTTPS/mDNS с телефона? | Доверенный локальный HTTPS endpoint, bootstrap guide | Д006, LAN, доступ к test clients |
| Д010 | Достаточно ли доказательств для начала основного Core? | M0 review, обновлённые ADR/BOM/backlog/оценка | Результаты Д004–Д009 |

Д004 не требует заранее выполнить Д006: host audio и synthetic LED mapping можно проверить на developer Linux. Это лишь HOST/SIM evidence; Д008 повторяет критическое на Pi. Д005 не обучает полноценные модели, Д009 не реализует production Hub/auth. Сроки их реализации из общего плана сохраняются.

## Общие правила доказательств

У каждой проверки есть ID, окружение и результат:

- `PASS-HOST` — реально выполнено на developer host, не Pi.
- `PASS-SIM` — выполнено на synthetic inputs/fake adapters.
- `PASS-PI` — выполнено на целевой Pi с указанной сборкой/проводкой.
- `DOC-ONLY` — документально подтверждённый путь; запуска не было.
- `FAIL` — воспроизводимый отрицательный результат с причиной/следующим экспериментом.
- `BLOCKED` — нет необходимого hardware/access/data; это не PASS.
- `NOT_RUN` — проверка запланирована, но не выполнялась.

Неоднозначный результат — FAIL/BLOCKED либо новая проверка, а не «скорее работает». Сохранять raw technical logs локально с redaction; в публичный repo — методику, обезличенные summaries/config examples, versions/checksums. Runtime screen/camera/PCM не записывать ради evidence. Для визуальных проверок использовать synthetic images/test patterns и локальное наблюдение.

Будущие файлы указаны в `code formatting`: они ещё не созданы этим планом. Пути исходных материалов даны кликабельными ссылками. Production code начинается по общему графику; spike code держать в отдельной `experiments/` области, без автоматического включения в release.

---

## Д004. LED/audio spike и границы host/container

### Цель

Получить проверяемую гипотезу, как Go Core будет управлять двумя логическими LED-зонами, принимать Bluetooth/local/Spotify audio, выбирать выход и получать общий PCM для DSP. Исключить заведомо конфликтующий hardware path до основного кода.

**Вход:** ADR-001/002/007, аппаратные варианты Д002, Q-002/Q-007/Q-009/Q-026/Q-033. Для host-проб нужен Linux computer с доступным audio; Pi и итоговый PSU на этом этапе не обязательны. Bluetooth-проба требует реального адаптера/телефона.

### Блок 4.1 — зафиксировать experiment matrix, 60 минут

1. Прочитать принятый [ADR-007](../architecture/adr/007-led-topology.md) и [wiring draft](../../hardware/wiring/wiring.md); не начинать с GPIO18 только потому, что он в примере ТЗ.
2. Записать доступный host: OS/architecture, audio devices, Bluetooth adapter, installed audio daemons и пользователя services.
3. Проверить, доступна ли уже Pi с ОС. Если нет — пометить Pi cases BLOCKED до Д006/Д008, не выполнять случайную параллельную установку.
4. Разделить hypotheses: PWM two channels, SPI single chain/two slices, PCM alternative. У каждой указать resource conflicts, нужные permissions и вопрос независимости.
5. Сформировать output matrix: analog 3.5mm, USB DAC, Bluetooth; предусмотреть input/output Bluetooth одновременно как отдельный риск.
6. Определить минимальный эксперимент: local synthetic audio → выбранный output + общий monitor tap; fake zone A/B → общий frame без physical write.
7. Выбрать критерии остановки: несовместимый ресурс, нет нужного output, нужен недопустимый access, воспроизводимый crash или невозможный PCM tap.

**Сохранить:** `experiments/day004/README.md`, `docs/research/led-audio-hypotheses.md`.

### Блок 4.2 — исследовать audio graph, 90 минут

1. Описать source→mix/playback bus→output и точку PCM tap. Нужен анализ воспроизводимого Hub audio, а не случайный microphone capture.
2. Сравнить ограниченный shortlist host stacks. Начальный кандидат для исследования — PipeWire/WirePlumber + BlueZ; при неподтверждённости оставить альтернативу с обоснованием.
3. По документации **установленной версии** проверить headless service lifecycle: audio user, session activation, runtime directory, D-Bus и автозапуск без desktop login.
4. Убедиться, что не создаются два конкурирующих owners Bluetooth/audio для разных users/containers.
5. Проверить локальный synthetic WAV/tone через доступный output без привязки к Spotify account.
6. Установить, как Play/Pause/volume/mute и discovery доступны Go adapter: какие RPC/IPC/host APIs, какие права, какие events.
7. Составить capability table для sources: поддержка transport/metadata условна реальным source API; не скрывать gaps.
8. Если телефон/адаптер доступны — выполнить короткую A2DP input пробу. Если нет — записать точные шаги для Д008 и BLOCKED.

**Важно:** WirePlumber отдельно документирует logind/headless особенности Bluetooth. Не считать desktop playback доказательством headless startup. [Официальная документация](https://pipewire.pages.freedesktop.org/wireplumber/daemon/configuration/bluetooth.html).

### Блок 4.3 — получить PCM без хранения звука, 90 минут

1. Выбрать точку monitor capture на общем playback bus, чтобы local/Bluetooth/Spotify могли использовать один contract.
2. Зафиксировать **проектный формат** пробы: sample rate, channel count, sample representation, frame length и byte order. Числа пока не объявлять production defaults.
3. Подать synthetic tone известной частоты и silence; считать PCM только в bounded RAM buffer.
4. Проверить measured sample count/RMS/peak и отсутствие NaN/overflow; вместо записи PCM сохранять scalar summaries.
5. Проверить смену output: tap остаётся на выбранном logical bus либо должен переподключаться — это часть будущего adapter.
6. Проверить stop/start playback и no-signal поведение без вечного blocking read.
7. На паузе/тишине данные не должны подменяться фиктивным beat/BPM.
8. Согласовать технический flow Go audio manager→Python DSP, backpressure и source epoch. Полный FFT в этот день не писать.
9. Сформулировать, что повторить на Pi: format, silence, reconnect, output switch, A2DP source.

**Сохранить:** `docs/research/pcm-contract-v0.md`, source fixtures generator/metadata, raw scalar probe report. Никакого personal audio artifact.

### Блок 4.4 — проверить zone mapping и privilege boundaries, 60 минут

1. Создать минимальный synthetic frame builder для двух slices; counts брать из явно тестового сценария Д002.
2. Прогнать A=red/B=blue, затем A=OFF/B=green; проверить, что команды зоны A не изменяют B.
3. Проверить bounds/overlap, zero/negative/unknown counts и общий expected frame length.
4. Для каждой physical hypothesis выписать device nodes, IPC sockets, groups/capabilities, DMA/memory access и конфликтующие resources.
5. Разделить, что живёт на host, что в Core container, как ограничивается IPC. Не выдавать `--privileged` за окончательное решение.
6. Если driver требует root или широкий memory access, записать несовместимость с non-root constraint и варианты узкого adapter/helper boundary для проверки. Это новый design вопрос, не молчаливое расширение прав.
7. Прямой доступ bot/frontend/ML к LED/storage не добавлять даже в spike.

**Сохранить:** `docs/research/host-container-boundaries.md` и `experiments/day004/zone-frame-probe.*`.

### Блок 4.5 — выводы и подготовка Pi-пробы, 60 минут

1. Свести hypotheses в таблицу: подтверждено на host, известно из docs, что требует Pi.
2. Выбрать один primary candidate и один fallback по понятным причинам; не писать «driver окончательно выбран», если теста Pi не было.
3. Составить пошаговый Д008 protocol с inputs, safe initial brightness/current, expected output и cleanup.
4. Записать нужны ли DAC/hub/кабели/отдельный user service до Д008.
5. Обновить Q-002/Q-007/Q-009/Q-026/Q-033 только в проверенной части; согласовать с ADR-007.
6. Зафиксировать дефекты: symptom, environment, reproduction, hypothesis, next test. Не тратить остаток дня на неопределённое переключение всех Linux audio stacks.

### Проверки Д004

| ID | Проверка | Ожидаемый результат | Допустимое evidence Д004 |
|---|---|---|---|
| D04-01 | Two fake zone slices | Независимые colors/modes без overlap | PASS-SIM |
| D04-02 | Local source→output | Реальный слышимый test signal | PASS-HOST либо BLOCKED |
| D04-03 | Playback monitor→PCM | Корректные samples/RMS/timing | PASS-HOST либо BLOCKED |
| D04-04 | Pause/silence/restart | Чёткое no-signal и восстановление | PASS-HOST либо BLOCKED |
| D04-05 | Bluetooth input | Phone playback/tap, если есть hardware | PASS-HOST либо BLOCKED |
| D04-06 | Privilege inventory | Полный список required device/IPC permissions | DOC-ONLY, на Pi проверить Д008 |
| D04-07 | PWM/audio conflict decision | Не выбран заведомо конфликтующий path для полного ТЗ | DOC-ONLY + гипотеза |

**Готовность:** создан воспроизводимый host/sim эксперимент, понятен primary path и Pi gate. Если нет Linux/audio hardware, завершена только документальная ветка, Д004 отмечается частично выполненным. **Не готово:** список библиотек без probe, false audio tap с microphone или «работает в Docker» только под unlimited privileges.

**Выход дня:** `docs/DAY_004_RESULT.md`, hypothesis/PCM/boundary docs, spike code/commands, test matrix и unresolved blockers. Product Audio Manager, production Compose и все Spotify controls этим днём не закрываются.

---

## Д005. Screen capture / local ML spike и методика измерений

### Цель

Уменьшить риски четырёх платформ screen-client, определить разделение 30 FPS colors и медленного inference, заранее задать качество/latency/dataset policy. Полный Ambilight и обучение жестов/сцен ещё не делаются.

**Вход:** domain/events/camera/privacy contracts Д003, Q-006/Q-011/Q-013/Q-015/Q-018/Q-021/Q-022/Q-035. Нужен хотя бы один доступный development laptop; остальные targets можно исследовать документально, не помечая их запуск PASS.

### Блок 5.1 — platform и capture matrix, 60 минут

1. Записать Linux x86_64, Windows x86_64, macOS arm64, macOS x86_64 как четыре обязательных targets, а не «три ОС».
2. Для каждого указать доступную машину, OS version, display session, GPU/display topology, разрешения и кто сможет выполнить test.
3. Рассмотреть Windows native capture, macOS ScreenCaptureKit и Linux X11/Wayland-specific paths. Один generic Go wrapper не считать доказательством всей матрицы.
4. По каждому candidate записать minimum OS/API availability, native dependencies, Go binding/cgo implications, license и packaging.
5. Отдельно записать multi-monitor, DPI/Retina scaling, HDR/color space, cursor policy, rotation, lock screen, protected content и monitor-off limitations.
6. Выбрать одну платформу для реального минимального захвата сегодня. Недоступные hardware targets — DOC-ONLY/BLOCKED с дедлайном Д061–Д068.

**Сохранить:** `docs/research/screen-platform-matrix.md`. Начальные первичные источники приведены в конце документа; конкретные OS support conclusions требуют чтения доступных версий docs и запуска.

### Блок 5.2 — минимальный capture/analyzer, 90 минут

1. Создать демонстрационный screen pattern: разные сплошные цвета по четырём краям, движение внутри кадра, затем black bars.
2. Захватить текущий выбранный display в RAM. Запись screenshots/video не включать.
3. Проверить resolution/stride/pixel format/channel order и соответствие physical screen выбранному display.
4. Посчитать simple edge averages для synthetic pattern, затем зафиксировать weighted-average contract из ТЗ; полный configurable analyzer оставить Д059–Д064.
5. Измерить capture/analysis elapsed time и allocations/peak memory на коротком прогоне.
6. Проверить отказ разрешения и monitor disconnect: клиент выдаёт понятный error, не циклически пишет frame dumps.
7. Отпустить capture resources и убедиться, что процесс завершился; revoke permission проверить, если target это поддерживает.
8. Сохранить numeric expected/actual colors, не реальный screen content. Для недоступной платформы прогнать analyzer на generated in-memory fixtures.

**Сохранить:** `experiments/day005/capture-probe.*`, fixture generator, `docs/research/capture-probe-results.md`.

### Блок 5.3 — protocol и local inference delivery, 90 минут

1. Определить packet fields: version, client/session epoch, sequence, mapping version, zone binding, timestamps, RGB colors, technical timing. Frame image/pixel array всего экрана запрещён.
2. Установить validity checks: lengths/counts/finite values, maximum payload/rate, stale sequence/epoch, wrong binding и unknown version.
3. Описать latest-frame queue: render не ждёт старых пакетов; control commands имеют отдельный bounded path.
4. Разделить capture/colors 30 FPS и temporal sample buffer для Scene ML ≤1 inference/sec. ML worker failure не блокирует colors.
5. Сравнить два локальных inference paths: embedded/native runtime и отдельный packaged local worker; проверить platform/architecture matrix, dependencies/license/model format.
6. Если возможно, выполнить **tiny inference smoke** на synthetic public model/input на доступной машине. Это доказательство runtime loading, не качества Scene ML.
7. Не выбирать cloud API как замену отсутствующей arm64/Intel/macOS зависимости.
8. Зафиксировать mapping/token binding, revoke behavior и отсутствие authority stream на смену lighting mode согласно ADR-002/007.

**Сохранить:** `docs/protocols/screen-stream-v0.md`, `docs/research/local-inference-options.md`. Если решение runtime не подтверждено — выбрать shortlist и срок Д071, не выдумывать успешный packaging.

### Блок 5.4 — performance/quality методика, 60 минут

1. Для каждого §46 target определить начало/конец измерения, часы, aggregate и reference load. Использовать таблицу методик ниже.
2. Описать benchmark environment: Pi4B/2GB, actual LED count/resolution/network, enabled services, warmup, run duration и concurrency/RPS. Пока неизвестные параметры оставить явными.
3. Задать профиль **минимального spike** и будущий **полный joint load** отдельно. Д008 synthetic workload не объявляется финальным full-stack benchmark.
4. Gesture <500ms измерять с confirmation 300–500ms внутри бюджета, не только inference. При противоречии сохранить blocker, не менять точку отсчёта ради PASS.
5. Для Scene/gesture определить метрики, annotation rubric и кто утверждает численные thresholds до training. Если benchmark/data ещё нет, назвать thresholds provisional и назначить дату фиксации.
6. Для coverage описать core scope, исключения generated/vendor и способ суммирования покрытых строк, а не арифметическое среднее процентов языков.

### Блок 5.5 — fixtures/privacy и отчёт, 60 минут

1. Определить разрешённые материалы: synthetic patterns/audio/temporal sequences, public licensed datasets с attribution/manifest. Личные webcam/screen captures не становятся fixtures.
2. Описать splits по исходным роликам/времени, чтобы соседние кадры не оказались одновременно train/test.
3. Для Scene перечислить все девять labels, ambiguous/opposite combinations и expected adaptation boundaries.
4. Для gesture предусмотреть negative motion/partial hands/lighting/distance и false actions/min, а не только красивое positive demo.
5. Для habits определить temporal split/calibration/insufficient data; synthetic 30 days не заменяют реальные observations.
6. Проверить stdout/exception/crash/cache paths на отсутствие raw frames. Camera SDK sample с `save_frame` не переносить в проект.
7. Обновить Q-006/Q-011/Q-013/Q-015/Q-018/Q-021/Q-022/Q-035, составить preconditions Д071/Д087.
8. Сохранить results по каждому target с честным evidence level.

### Методики, которые должны появиться на Д005

| Target ТЗ | Измерение | Что нельзя засчитывать |
|---|---|---|
| REST p95 <100ms | Client request→полный response, определённые operations/RPS/concurrency | Время пустого handler без сети/нагрузки |
| WS event <100ms | Core publish→client receive с общей/оценённой опорой времени | Только socket write duration |
| LED >=30FPS | Фактически переданные/applied frames, interval/drop statistics | Скорость генератора без driver |
| Ambilight 30FPS | Capture→network→applied LED cadence | Capture FPS без LED |
| Ambilight e2e <100ms | Test screen change→physical LED response с указанной точностью | RTT, сумма некорректно синхронизированных timestamps |
| Gesture <500ms | Начало устойчивого gesture→реальное action, включая confirmation | Только landmark inference |
| Scene ≤1/sec | Actual inference count/interval и независимый color-loop timing | Timer configured=1sec без измерений |
| Presence 2–3sec | Actual inference intervals при совместной нагрузке | Interval без учёта starvation |

### Проверки Д005

- D05-01: 4-target matrix заполнена, unavailable target не назван протестированным.
- D05-02: хотя бы доступный capture path либо synthetic analyzer реально запущен.
- D05-03: edge colors/format/cleanup проверены, runtime frame files отсутствуют.
- D05-04: colors и inference очереди/частоты разделены, failure behavior описан.
- D05-05: каждый performance target имеет методику и будущую среду.
- D05-06: dataset licenses/split/negative examples/privacy и quality-gate deadlines определены.

**Выход дня:** `docs/DAY_005_RESULT.md`, matrix/protocol/measurement/dataset documents и minimal probe. **Готовность:** понятен технический путь и оставшиеся platform/quality риски. Full cross-platform packaging, gesture model и Scene ML quality остаются будущими gates M4/M5/M6.

---

## Д006. Воспроизводимая headless установка Raspberry Pi

### Цель

Получить управляемую по SSH Pi4B/2GB с выбранной поддерживаемой Ubuntu Server arm64, исправным storage/LAN/питанием и зафиксированным baseline. Это основа hardware evidence Д007/Д008, не deploy готового Hub.

**Обязательный вход:** физическая Pi, подходящий отдельный PSU, охлаждение, целевой носитель и компьютер для подготовки image, LAN access. Список наличия пока неизвестен. Если чего-то нет — реальная установка BLOCKED; допустимы только install guide/image manifest подготовка.

### Блок 6.1 — выбрать образ и проверить носитель, 60 минут

1. Проверить маркировку Pi/RAM и доступность корректного PSU, microSD/SSD, охлаждения, сетевого кабеля.
2. Выбрать поддерживаемую на дату установки Ubuntu Server **для Pi4/arm64**, сверив release lifecycle, board support и требуемые camera/audio/kernel packages.
3. Не устанавливать interim release только потому, что он был первым с нужной camera feature. Документ camera stack не заменяет lifecycle check.
4. Скачать официальный image, записать URL/version/checksum; проверить checksum/подпись способом производителя.
5. Точно идентифицировать целевой storage и сохранить нужные прежние данные до записи image. Device path не выводить из старого примера `/dev/sdX`.
6. Определить, какой носитель используется сейчас, а какой — поздняя цель microSD→SSD; не требовать сразу миграцию Д139.
7. Подготовить список rollback: исходный image, known-good boot media, сохранённые конфиги/данные.

**Сохранить:** `deploy/bootstrap/image-manifest.example.json`, `docs/operations/pi-install.md`. В version-controlled manifest не публиковать user SSH keys/пароли/private host identifiers.

### Блок 6.2 — подготовить headless доступ и загрузить ОС, 90 минут

1. Подготовить image штатным способом выбранной Ubuntu; проверить, поддерживает ли он нужные headless/cloud-init настройки.
2. Создать отдельного OS пользователя для администратора и SSH public-key access; приватный ключ остаётся на операторском компьютере.
3. Настроить hostname/timezone/network средствами выбранного image. Зафиксировать приватные значения только в local install notes.
4. Подключить охлаждение/сеть, загрузиться, дождаться завершения initial provisioning.
5. Найти Pi в LAN, проверить SSH host fingerprint локально/доверенным способом, подключиться.
6. Проверить failed login/неверный user key без блокировки собственного единственного доступа.
7. Проверить второй SSH сеанс до закрытия первого после смены auth/network config.
8. Описать recovery console/boot-media путь на случай потери SSH. Не публиковать administrative service напрямую в WAN.

**Ожидаемо:** headless доступ без подключённого монитора; первый запуск может требовать отдельного provisioning времени.

### Блок 6.3 — baseline inventory и обновление, 90 минут

1. Собрать OS/kernel/architecture/board/RAM/storage/mount/free-space сведения. Проверить `aarch64`, а не desktop x86 image.
2. Проверить актуальность system clock/timezone и механизм синхронизации; записать поведение без WAN.
3. Обновить необходимые OS packages штатным способом; сохранить итоговые versions, не запускать произвольный curl-to-shell.
4. Если требуется reboot — выполнить и снова проверить SSH/LAN/storage. Device names могут измениться: использовать устойчивые идентификаторы в local configs.
5. Проверить системные logs на storage I/O errors, undervoltage indications, read-only root и resource faults. Не считать отсутствие vendor utility доказательством отсутствия проблемы.
6. Прочитать `/sys/class/thermal/thermal_zone0/temp`, преобразовать единицы и сверить адекватность результата.
7. Снять idle CPU/RAM/temperature после стабилизации. Указать длительность наблюдения, ambient conditions и работающий fan.
8. Проверить USB inventory; оценить, помещаются ли будущие camera/SSD/DAC в power/ports budget. Не подключать отсутствующие devices фиктивно в отчёт.

**Примеры будущих read-only проверок, не выполненных этим планом:**

```bash
uname -a
cat /etc/os-release
uname -m
free -h
lsblk -o NAME,SIZE,TYPE,FSTYPE,MOUNTPOINTS
findmnt /
df -h
cat /sys/class/thermal/thermal_zone0/temp
timedatectl status
```

Выводы могут содержать private identifiers: перед публикацией redaction. Диагностика boot/voltage зависит от выбранной версии, точные команды записать после проверки их наличия.

### Блок 6.4 — runtime prerequisites и baseline backup, 60 минут

1. Проверить наличие systemd, нужных bus interfaces и package manager; не настраивать production Compose до его этапа.
2. Выбрать способ подготовки toolchains/containers для spike, записать versions и источники. Полный Go/React/Python monorepo на Д011 не подменять этим bootstrap.
3. Если нужен Docker для Д008, установить поддерживаемым способом, проверить architecture и простейший isolated container. Docker group не считать безобидной заменой least privilege.
4. Создать local directory для приватных install notes/секретов с ограниченными permissions, отдельно от public repo.
5. Сохранить baseline system config/порядок восстановления; actual full backup можно выполнить доступным методом, но не объявлять будущий daily app backup §35 уже реализованным.
6. Проверить, что bootstrap повторяем из image-manifest и инструкции; нет неописанного ручного toggle.

### Блок 6.5 — reboot acceptance и report, 60 минут

1. Выполнить штатный shutdown/reboot, подключиться заново и проверить storage/network/clock/temperature.
2. Сравнить baseline до/после, проверить stable mount/SSH permissions.
3. Зафиксировать точный software/firmware/package manifest и что поменялось update.
4. Обновить inventory `hardware/bom/design-inputs.json` **только реально проверенными данными**. Не затирать неизвестные null предположениями.
5. Записать local-only host identifiers вне public report; версия/модель/технический результат без personal data остаются в repo.
6. Приложить preconditions Д007 и неподтверждённые devices.

### Проверки Д006

| ID | Проверка | PASS-PI означает |
|---|---|---|
| D06-01 | Board/OS/architecture | Реальные Pi4B/2GB, supported Ubuntu Server/arm64 и version manifest |
| D06-02 | Headless SSH | Доступ после reboot без монитора, ключи/права проверены |
| D06-03 | Storage | Root writable, нет выявленных I/O ошибок, достаточно места |
| D06-04 | Power/cooling baseline | При baseline нет обнаруженных faults; температура читается, fan работает |
| D06-05 | Clock/LAN | Связь/время/hostname воспроизводимы; offline ограничения записаны |
| D06-06 | Reinstall guide | Все фактические шаги/versions/checksum/recovery описаны |

**Выход:** `docs/DAY_006_RESULT.md`, install guide, обезличенный `hardware/bom/pi-baseline.md`, image/package manifest. При отсутствии Pi/install access — report BLOCKED с конкретным списком недостающего, не «Ubuntu проверена в VM».

---

## Д007. Охлаждение, hardware access и permissions

### Цель

Проверить реальные GPIO/I²C/SPI interfaces и охлаждение до длинных audio/LED экспериментов. Определить доступ process/container к каждому ресурсу и thermal policy с гистерезисом.

**Вход:** Д006 PASS-PI, корректно собранное охлаждение; pin map Д002; fan driver/доступные датчики. Полная LED-лента не нужна. Если fan control hardware ещё нет, always-on cooling допускает часть опытов, но controlled fan criterion остаётся BLOCKED.

### Блок 7.1 — physical inspection и pinmux inventory, 60 минут

1. Сверить реальную board orientation и BCM/physical numbering по [pin map](../../hardware/wiring/pin-map.md).
2. Проверить fan mounting, electrical input/current и подключение через подходящий driver. GPIO не питает motor.
3. Убедиться, что силовые returns LED не проходят через Pi header; LED power пока отключено/не собрано.
4. Снять список enabled overlays/pin functions/device nodes на фактической Ubuntu. Не редактировать одновременно несколько boot config files «на всякий случай».
5. Найти конфликты SPI0, I²C1, UART/I²S/PWM с выбранной раскладкой. Зарезервированные линии не занимать button/fan автоматически.
6. Определить control defaults на boot/shutdown и факт появления устройств после config/reboot.
7. Сохранить before/after configuration и способ возврата без потери headless доступа.

### Блок 7.2 — GPIO/I²C/SPI access, 90 минут

1. Выбрать GPIO API, доступный текущему kernel/userland; фиксировать найденные gpiochips/lines, а не считать chip index универсальным.
2. Проверить безопасный low-current input/output loop на button/indicator либо test fixture. Motor/полную ленту для этой пробы не использовать.
3. Проверить known button press/release, input bias и GPIO output для indicator; это electrical smoke, не готовый press-recognition модуль Д033.
4. Для I²C подключать только проверенные 3.3V modules и адреса. Не выполнять массовое probing произвольных addresses как универсальную «безопасную диагностику».
5. Прочитать documented sensor ID/status конкретного module, если доступен. При отсутствии sensor проверить device node/permissions и оставить read evidence BLOCKED.
6. Сверить supply rails и pull-ups с фактическими breakout schematics; суммарное параллельное сопротивление не игнорировать.
7. Проверить SPI0 device accessibility/config с disconnected LED либо согласованным loop/test fixture. Данные MOSI на GPIO10 не должны попасть на случайно подключённое устройство.
8. Записать `root` и non-root результаты отдельно. Если чтение доступно только root — объяснить почему и следующий минимальный permission вариант.
9. Удалить/вернуть test config и убедиться, что baseline audio/storage/network не сломаны.

**Сохранить:** `docs/research/pi-hardware-access.md`, exact package/command/config list. Продуктовые sensor drivers ещё не писать.

### Блок 7.3 — температурные состояния и fan semantics, 90 минут

1. Уточнить Q-004: отдельно state thresholds NORMAL/WARM/HOT/CRITICAL, fan curve, warning threshold, hysteresis и sensor-failure behavior.
2. Ориентиры ТЗ <50/50–65/>65/>75°C не выдавать за полную уже заданную таблицу четырёх states.
3. Составить candidate table: вход в state, выход обратно, минимальная длительность стабильного значения, fan action, alert/optional workload reduction.
4. Определить shutdown/start defaults: при unavailable temperature sensor охлаждение не должно незаметно выключиться. Выбранный физический driver должен уметь реализовать заявленную fail policy.
5. Сделать tiny thermal policy probe с synthetic temperature sequence и fake clock: steady values, oscillation около порога, malformed/missing reads, restart.
6. Для deterministic tests использовать подставной input port, не писать fake temperature в системный thermal sysfs.
7. Если driver/fan поддерживают только ON/OFF, не заявлять физическую LOW speed. UI/policy должна соответствовать реальным capabilities.
8. Не нагревать реальную Pi до CRITICAL ради проверки transition: крайние ветки проверять synthetic input.

**Сохранить:** `docs/research/thermal-policy-v0.md`, `experiments/day007/thermal-policy-probe.*`, ожидаемые таблицы переходов.

### Блок 7.4 — ограниченная реальная thermal проба, 60 минут

1. До нагрузки записать idle temperature, ambient conditions, fan state, PSU/storage состояние.
2. Запустить ограниченный воспроизводимый CPU load с заданной длительностью, доступным memory cap и критерием остановки.
3. Наблюдать температуру/clock/throttling/error signals средствами фактической Ubuntu. Измерения сохранять как технические scalar records.
4. Проверить fan response при достижимом normal/warm диапазоне. Сравнить с commanded state и, если возможно, видимым вращением.
5. При необычном нагреве, fan failure, voltage/storage faults остановить нагрузку и записать FAIL; не продолжать ради красивого графика.
6. Остановить нагрузку, проверить cooldown и возврат fan по hysteresis. Документировать, что это короткий thermal spike, не soak.
7. Снова проверить SSH/storage/network после опыта.

### Блок 7.5 — least-privilege matrix и отчёт, 60 минут

1. Для `/dev/gpiochip*`, `/dev/i2c-*`, `/dev/spidev*`, audio IPC, temperature sysfs перечислить read/write, UID/group, device allowlist и readonly mounts.
2. Сопоставить host proof и возможный Core container access. `non-root` без нужных прав — не успех; полный `privileged` без ограничений — не итоговый design.
3. При необходимости узкого host adapter/helper описать API и authority; не менять silently контейнерную архитектуру ТЗ.
4. Проверить negative case: посторонний user/process не получает лишний device/IPC access.
5. Обновить fan/BOM/pin map только по фактическим изменениям. Сохранить candidate unresolved hardware safety пункты.
6. Подготовить готовый к запуску перечень Д008: package config, permissions, temperature watchdog и known-good reset.

### Проверки Д007

- D07-01: pinmux и real physical wiring согласованы, конфликтующие alternate functions не активны одновременно.
- D07-02: low-current GPIO input/output проверены на реальной Pi.
- D07-03: I²C documented read выполнен либо честно BLOCKED; voltage/pull-up review выполнен.
- D07-04: SPI device/permissions/timing prerequisites установлены; LED correctness ещё не заявлена.
- D07-05: thermal state/hysteresis/error ветки прошли deterministic probe.
- D07-06: реальная ограниченная нагрузка/охлаждение проверены без принудительного перегрева.
- D07-07: permissions matrix содержит positive и negative evidence.

**Выход:** `docs/DAY_007_RESULT.md`, hardware-access/thermal documents, limited-load report, permissions map и обновлённые вопросы. Отсутствие controlled fan или sensor evidence не закрывать synthetic tests; переносить конкретный пункт.

---

## Д008. LED/audio coexistence на реальной Raspberry Pi

### Цель

Подтвердить или опровергнуть основной технический риск: две независимо управляемые логические LED-зоны и обязательный аудиопуть одновременно работают на Pi4, общий PCM доступен, доступ к ресурсам воспроизводим.

**Вход:** Д004 protocol, Д006 installed baseline, Д007 thermal/access evidence. Нужны короткие реальные WS2812B segments, отдельный защищённый LED PSU/current limit, shifter/защита/проводка, phone/audio outputs и измерительные средства по BOM. Counts короткого стенда пересчитать по фактически используемым пикселям. Полный монитор/desk монтаж Д034 ещё не нужен.

### Блок 8.1 — собрать короткий test fixture, 60 минут

1. Сверить фактические LED SKU/revision/supply, arrow DIN/DOUT, RGB/GRB order и число пикселей каждого test segment.
2. Пересчитать test load существующим `hardware/bom/power_budget.py` с подтверждёнными или явно оценочными current coefficients.
3. Выбрать current limits, wires/connectors/fuses для фактического короткого стенда, а не для hypothetical 360LED assembly.
4. Проверить power-off continuity/polarity/common ground, DATA resistor/capacitors и отсутствие объединения +5V Pi/+5V LED.
5. Проверить shifter pinout/enable и минимальный physical LED disconnect. Открытые partial-power/backfeed вопросы нельзя считать решёнными одним OE.
6. До первого включения определить согласованный порядок питания и stop procedure. Произвольные power-order tests без isolation scheme не выполнять.
7. Снять baseline Pi temperature/voltage/status; при неисправном cooling или storage аппаратный опыт не начинать.

**Сохранить:** `hardware/wiring/test-fixture-day008.md`: actual parts/counts/cable lengths/pin assignments и фото/схема без чувствительных данных, если фото требуется.

### Блок 8.2 — transport и independent zones, 90 минут

1. Установить/собрать candidate driver exact version/commit; сохранить build flags/native dependencies и host configuration.
2. Проверить SPI/clock/DMA/permissions именно на выбранной Ubuntu. Не копировать boot config paths из другой ОС без сверки.
3. Начать с нулевого frame и малой brightness; проверить единичные primary colors, порядок каналов и границы цепочки.
4. Отправить A red/B blue, A OFF/B green, A dim static/B animated, затем поменять роли. Modes пробы могут быть простыми, но state separation должна быть настоящей.
5. Проверить, что A OFF — чёрный slice, а не обесточивание upstream сегмента с обрывом DATA к B.
6. Проверить bounds/неверный count: invalid config отклоняется до physical write.
7. Ввести согласованный ограничитель тестового frame/PSU; численная реализация spike не заменяет будущий tested production limiter.
8. Измерить actual transmit/frame cadence и memory use; source-generated 30FPS отдельно от фактически отправленных кадров.
9. Коротко проверить stop/start driver и cleanup: нет неконтролируемой вспышки/зависшего процесса.

**Сохранить:** transport config example, exact run commands и first HIL result. При flicker/clock/resource issues дать FAIL и минимальное reproduction.

### Блок 8.3 — audio и общий PCM на Pi, 90 минут

1. Повторить выбранный host audio setup Д004 с тем же service user/session strategy, адаптировав только реальные Pi devices.
2. Воспроизвести local synthetic audio через встроенный 3.5mm output при остановленном LED, затем при активном transport.
3. Получить playback monitor PCM на Pi, проверить sample format/RMS/silence и source epoch.
4. Подключить телефон A2DP в явном pairing window; проверить playback и тот же общий PCM tap. Pair outside window должен быть denied.
5. Если доступны outputs — повторить short smoke через USB DAC и Bluetooth output, без заявления полного UI/metadata/control acceptance Д045.
6. Проверить Bluetooth input/output concurrency отдельно. Если комбинация не поддерживается текущим graph/adapter, записать exact limitation и fallback, не скрыть заменой входа local file.
7. Spotify выполнить только при доступном совместимом local client/account; отсутствие доступа — BLOCKED этого smoke. Полная integration Д042 остаётся будущей.
8. Убедиться, что нет audio recording, full PCM dumps и listening history в logs.
9. Если LED ломает analog audio, вернуться к matrix Д004; не объявлять USB DAC equivalent закрытием встроенного output requirement.

### Блок 8.4 — совместная нагрузка и отказные случаи, 60 минут

1. На коротком стенде запустить одновременно animated zones, Bluetooth/local playback, PCM monitor, synthetic control requests и scalar metrics.
2. Заранее задать длительность/частоту/температурный stop criterion и сохранить условия. Это preliminary load, а не будущий full audio+Ambilight+CV+rules load Д135.
3. Измерить LED applied/transmitted FPS, audio dropouts, PCM continuity, CPU/RAM/temperature и доступность SSH/control loop.
4. Остановить PCM consumer: playback и базовый LED должны продолжать работу; проверить bounded queue и restart consumer.
5. Переключить output, phone disconnect/reconnect и driver restart по очереди, сохраняя cause/outcome.
6. Выполнить только согласованные power-off tests. Если hard-off может backfeed GPIO/DIN, сначала доработать isolation/interlock; pending hardware issue остаётся blocker постоянной установки.
7. Проверить работы process under intended non-root identity и отрицательный access из лишнего service. Containers тестировать с минимальным device/IPC list.
8. По окончании вернуть safe stopped state и проверить filesystem/log health.

### Блок 8.5 — решение и rollback, 60 минут

1. Сравнить результаты с гипотезами: transport, audio, PCM, permissions, thermal, power sequencing.
2. Принять primary topology/audio graph **только в границах выполненного стенда**; указать, что будет перепроверено с полной длиной/реальным монтажом.
3. Обновить ADR-007 статус доказательства и host boundary decision; не переписывать ТЗ под неудачный driver.
4. Зафиксировать known-good software versions/config/команды и способ отката предыдущих kernel/module/boot/audio настроек.
5. Обновить BOM/pin map только там, где actual wiring отличается от draft; отдельный test fixture не подменяет размеры комнаты.
6. Составить список blockers с назначенным резервом: driver privileges, analog output, PCM, signal integrity, hard-off, отсутствующий hardware.
7. Подготовить short reproduction, который можно повторить на Д010 без автора эксперимента.

### Матрица Д008

| ID | Случай | Требуемое наблюдение |
|---|---|---|
| D08-01 | Zone A/B independent state | Изменение одной не портит mode/color другой |
| D08-02 | Analog audio + LED | Реальный sound output и стабильная LED передача одновременно |
| D08-03 | Bluetooth input + PCM + LED | Phone playback, корректный monitor tap, без ресурсного конфликта |
| D08-04 | USB output | Доступный DAC работает; отсутствие устройства = BLOCKED |
| D08-05 | Bluetooth output/concurrency | Роли/ограничения реально измерены |
| D08-06 | PCM consumer failure | Playback/LED продолжаются, queue bounded |
| D08-07 | Frame cadence | Измеренные actual FPS, не только generator timer |
| D08-08 | Intended permissions | Не требуется неописанный unrestricted privileged deployment |
| D08-09 | Stop/restart/physical disconnect | Известный outcome и отсутствие неподтверждённого обещания power safety |
| D08-10 | Repeatability | Exact config/versions/cleanup повторяемы |

**Выход:** `docs/DAY_008_RESULT.md`, HIL matrix, validated test fixture, adapter shortlist/final-in-scope choice и updated ADR. **Готовность:** подтверждён основной Pi path, либо чётко объявлен technical blocker с новым планом. Test «светится лента при остановленном audio» недостаточен.

**Если нет Pi/LED/phone:** подготовить protocol и minimal spike code; весь HIL matrix остаётся BLOCKED/NOT_RUN, M0 hardware gate не закрывается. Продолжать независимую Д009 documentation можно, но actual LAN/Pi evidence также требует hardware.

---

## Д009. Локальное имя, HTTPS и процедура доверия

### Цель

С телефона/ноутбука открыть доверенный `https://smarthome.local` на реальной LAN Pi и воспроизвести установку/обновление сертификатов. Core/auth ещё не реализованы: endpoint для этой проверки — минимальный stub/status page, не готовый Hub Dashboard.

**Вход:** Д006 working Pi/LAN/SSH, test phone/laptop с правом установки trust, ADR-004 и network boundaries ТЗ. Не требуется готовая SQLite/auth/frontend implementation. Если изоляция LAN запрещает multicast или CA trust на телефоне недоступен, это диагностируется отдельно.

### Блок 9.1 — DNS/mDNS и proxy shortlist, 60 минут

1. Проверить, что `smarthome.local` не занят другим устройством; определить план поведения при конфликте, а не принимать чужой ответ за Pi.
2. Записать фактические LAN interfaces/subnets/multicast restrictions в private notes, в repo — обезличенный network scenario.
3. Выбрать mDNS publisher и reverse proxy по supported Ubuntu packages/документации, headless service lifecycle, local CA support, resource budget и logging.
4. Проверить, нужен ли отдельный proxy process/container и какие ports/mounts он получает; его наличие не меняет владельца Core SQLite/hardware.
5. Составить LAN exposure table: proxy HTTPS, internal loopback stub/Core port, SSH administration, будущие broker/API endpoints.
6. Решить, как поддерживается loopback-only/internal endpoint; не публиковать временный unauthenticated control API на 0.0.0.0.
7. Сохранить выбранные package versions/почему alternatives не нужны. Не устанавливать два proxy одновременно ради сравнения «на глаз».

### Блок 9.2 — mDNS и минимальный endpoint, 90 минут

1. Настроить publisher штатным способом ОС и проверить advertisement после restart.
2. С Pi и другого LAN клиента разрешить имя, сопоставить результат с нужным устройством.
3. Если доступны IPv4/IPv6, проверить оба пути либо явно задокументировать supported path/запрет лишнего exposure.
4. Поднять минимальный loopback status endpoint с заранее известным marker/build string. Не использовать production auth route до реализации auth.
5. Настроить proxy routing к этому endpoint; web root не содержит env/keys/backups/directory listing.
6. Проверить, что другой LAN клиент не ходит напрямую в stub port при принятой network policy.
7. Проверить phone Wi-Fi isolation/guest network/VPN routing как отдельные causes mDNS failures.
8. Зафиксировать resolver/cache поведение после изменения IP/перезапуска. Не «лечить» mismatch отключением TLS verification.

**Сохранить:** `configs/mdns.example.*`, `configs/proxy.example.*`, `experiments/day009/status-stub.*`, minimal endpoint instructions. Реальные сертификаты/keys в repo не писать.

### Блок 9.3 — local CA / сертификат и trust, 90 минут

1. Выбрать local CA/bootstrap процедуру, соответствующую proxy и actual clients; определить местонахождение/owner CA key и server key.
2. Выпустить server certificate с корректной проверкой имени `smarthome.local` и сроками по выбранной policy; SAN/chain проверить TLS-инструментом.
3. Root CA private key не копировать на телефоны/в browser assets; клиент получает только нужный публичный trust certificate.
4. Установить trust на тестовом ноутбуке и телефоне штатным способом их ОС/браузера. У некоторых clients trust stores различаются — записать фактические шаги.
5. Проверить valid chain/hostname/time, обычный browser access без bypass warning и CLI verification с trusted CA.
6. Отдельно проверить чистый client без trust: expected untrusted error, а не «оно вообще сломано». Это planned negative test, не повод выключать HTTPS.
7. Проверить ошибочный hostname и неполную chain; исправление должно быть в сертификате/config, не в `insecure skip verify` клиента.
8. Сохранить только public certificates/examples там, где это соответствует privacy policy; CA/server private keys остаются в защищённом local storage вне git/артефактов.

### Блок 9.4 — network boundaries и certificate lifecycle, 60 минут

1. Проверить plain HTTP behavior: redirect к HTTPS либо явный отказ по принятой policy; sensitive API requests не обслуживаются открыто.
2. Проверить локальные bind/firewall rules; Hub API не публикуется напрямую в WAN и случайно не остаётся открытым по IPv6.
3. Проверить logs/headers/proxy error pages на secrets/private paths. Client IP logging policy согласовать с Q-030.
4. Отрепетировать server certificate replacement/reload без ненужного downtime и описать, что меняется при CA replacement.
5. Различить renewal server cert и отзыв/утрату CA key: новый root требует повторного trust install; не обещать автоматическое удаление доверия на всех чужих устройствах.
6. Проверить secret file permissions и доступ proxy только к необходимым key/cert; bot/frontend не получают CA private key.
7. VPN remote access оставить отдельной будущей настройкой, если пользователь её не запросил; HTTPS/local CA не означает готовый remote access.

### Блок 9.5 — reboot и инструкция нового клиента, 60 минут

1. Перезапустить proxy/mDNS, затем Pi; проверить hostname resolution, TLS и expected marker снова.
2. Пройти инструкцию trust install/revoke с чистого клиента либо свежего browser profile/OS store согласно actual platform.
3. Проверить смену IP/истечение старой advertisement cache контролируемым способом, если позволяет LAN. Router config не менять вслепую.
4. Описать диагностику по слоям: IP reachability→name resolution→port→TLS chain/time/name→proxy upstream.
5. Сохранить certificate fingerprints/expiry без private secrets, config examples и cleanup steps для временного stub.
6. Обновить Q-010/Q-026/Q-029/Q-031 в подтверждённой части; auth cookies/CSRF будут отдельно проверяться Д019/Д022.

### Проверки Д009

| ID | Случай | PASS-PI / client означает |
|---|---|---|
| D09-01 | mDNS name | Имя разрешается в нужную Pi с LAN client |
| D09-02 | TLS trusted | Без bypass warnings открывается нужный marker с телефона/ноутбука |
| D09-03 | Untrusted/wrong name | Предсказуемый отказ, verification не отключается |
| D09-04 | Upstream isolation | Stub/internal Core port не выставлен как незащищённый public endpoint |
| D09-05 | Restart | Publisher/proxy/TLS сохраняют работоспособность после reboot |
| D09-06 | Cert replacement | Runbook замены/CA trust воспроизводим, keys не утекли |
| D09-07 | Repo hygiene | Нет private key/env/backup/private network data в tracked files |

**Выход:** `docs/DAY_009_RESULT.md`, `docs/operations/local-https-mdns.md`, proxy/mDNS examples и negative-test report. **Готовность:** реальный trusted endpoint и повторяемая процедура. Отключённая browser warning или только `curl -k` не закрывают проверку. Пункты §50 про admin/Dashboard ещё не выполнены — они зависят от Core/auth/UI будущих дней.

---

## Д010. Повторяемость экспериментов и контрольный рубеж M0

### Цель

Решить на основании evidence, можно ли начинать основной Infrastructure Core на Д011, и что остаётся риском следующих этапов. M0 — готовность технической базы, **не** приёмка готового Hub и не обещание, что полная нагрузка уже измерена.

**Вход:** reports Д004–Д009, actual software versions/configs, working Pi fixture, latest ADR/Q registry. Недостающие reports собрать как NOT_RUN/BLOCKED, а не заполнять положительными предположениями.

### Блок 10.1 — evidence inventory и список критериев, 60 минут

1. Собрать полный список experiments: ID, hypothesis, platform, exact build/config, outcome, evidence path, date и person/operator.
2. Проверить различие HOST/SIM/PI/DOC evidence и источники цифр. Ни один hardware PASS не должен опираться только на расчёт.
3. Сопоставить фактическую базу с requirements: Pi/headless/storage, cooling/interfaces, LED/audio/PCM, LAN/TLS, capture/ML path и security boundaries.
4. По каждому Q-002/Q-004/Q-006/Q-007/Q-010/Q-011/Q-013/Q-016/Q-021/Q-026 записать, что закрыто **на уровне M0**, а что проверится позже.
5. Убедиться, что реальный BOM/pin map соответствует испытанному стенду. Project room counts по-прежнему могут быть неизвестны: не заменить их test-strip counts.
6. Установить окончательный scope M0: минимальная Pi technical feasibility и reproducibility. Final four-platform clients/ML quality/full joint load не должны ошибочно попасть в «уже PASS».

### Блок 10.2 — восстановление из записанной процедуры, 90 минут

1. Подготовить disposable test media/чистую конфигурацию либо known baseline snapshot; сохранить рабочий носитель/данные перед переустановкой.
2. По install guide восстановить нужные OS/package/config шаги. Если полное imaging не укладывается/нет второго носителя, выбрать controlled clean-config replay и **явно указать**, что clean-image reinstall пока не проверен.
3. Запустить минимальные hardware/audio/TLS probes из записанных commands без ручных «ещё я забыл включить этот параметр».
4. Сверить dependency versions/checksums/config paths и service users; убрать неявную зависимость от интерактивной shell/session.
5. Повторить reboot и automatic host service startup там, где они уже нужны spike. Production Compose autostart Д139 не считать реализованным от одного audio daemon.
6. Записать время восстановления и каждый недокументированный шаг, затем исправить guide.
7. Сохранить known-good baseline/recovery point, который можно использовать при дальнейшей разработке.

### Блок 10.3 — короткий critical regression, 90 минут

1. Проверить headless SSH/storage/temp after reboot.
2. Повторить low-current GPIO, известный I²C access и SPI prerequisite test в пределах доступной периферии.
3. Повторить two logical zones при real LED + обязательный critical audio path (включая встроенный analog output) + общий PCM.
4. Проверить PCM consumer stop/start и очевидные driver errors. Проба должна подтвердить failure boundary, а не только stable playback.
5. Повторить mDNS/trusted HTTPS с другого LAN client и expected upstream marker.
6. На доступной laptop platform повторить capture/analyzer/runtime smoke либо отметить documentary-only coverage; не требовать за день полноценные четыре packages.
7. Выполнить bounded concurrent replay с прежним reference profile и сравнить FPS/resources/temperature, указав длительность.
8. Для проваленного critical case сохранить issue и остановить M0 PASS. Ошибка не исчезает от того, что отдельный модуль вчера работал.

### Блок 10.4 — решения, риски и новая оценка, 60 минут

1. Обновить ADR-007 до «подтверждён на таком-то стенде», если actual proof достаточен. Не писать «окончательная hardware topology для любой длины».
2. Зафиксировать audio graph/host-container rights и какие controls/outputs требуют будущего real integration.
3. По screen/ML выбрать подтверждённый shortlist/доказанный путь; нет platform runtime path — оставшийся значимый риск с дедлайном, не скрытый technical debt.
4. Обновить Q statuses без удаления старых вопросов/IDs. Закрытый design вопрос и закрытый test criterion различать.
5. Классифицировать остаток: блокирует Core Д011; блокирует hardware M2/M3; блокирует screen/ML M4/M5; допустимая поздняя настройка.
6. Пересчитать remaining workload по фактической сложности. Если driver требует большой архитектурной замены или расход нескольких дней резерва, записать новый прогноз сразу.
7. Спланировать один primary/fallback path на каждый critical risk и конкретный следующий эксперимент. Не тащить в Core одновременно три непроверенных audio stacks.

### Блок 10.5 — M0 report и handoff на Д011, 60 минут

1. Собрать `docs/milestones/M0_TECHNICAL_BASELINE.md` по gate-таблице ниже.
2. Для каждого gate дать status, evidence, limitations, owner и follow-up day.
3. Записать итог `M0_PASS` либо `M0_NOT_READY`. Документальный прогресс без Pi evidence не является полным PASS.
4. Если critical gate открыт — разрешить только независимые scaffold/config/docs задачи Д011, указав перенос зависимой hardware работы. Не переименовывать это в закрытый M0.
5. Подготовить список concrete Д011 tasks: directories/toolchains/commands/examples/CI skeleton на основе подтверждённых boundaries.
6. Проверить public artifacts на secret/env/key/DB/private identifiers и удалить из planned commit всё неподходящее. Никаких actual commits/push/deploy без необходимости этого этапа.
7. Сохранить `docs/DAY_010_RESULT.md`, обновлённый общий план/backlog/ADR и краткую памятку повторного запуска.

### Gate-таблица M0

| Gate | Что нужно доказать к Д010 | Тип evidence | Что остаётся на следующих этапах |
|---|---|---|---|
| G01 Platform | Real Pi4B/2GB, Ubuntu Server/arm64, headless доступ, stable baseline/reboot | PASS-PI Д006/Д010 | Clean production install/deploy/SSD migration |
| G02 Cooling/access | Cooling работает; thermal читается; нужные interfaces/permissions понятны | PASS-PI + synthetic policy | Production Thermal Manager, long soak, actual all-device load |
| G03 LED independence | Две logical zones реально адресуются независимо выбранным transport | PASS-PI Д008/Д010 | Full-length wiring/signal/current calibration, all effects |
| G04 Audio/PCM feasibility | Встроенный analog output и основной input/PCM path сосуществуют с LED | PASS-PI Д008/Д010 | Полная source×output×control matrix, Spotify integration, metadata/UI |
| G05 Permissions | Нет неописанной зависимости от unlimited privileges; выбран проверяемый минимальный путь | Actual identity/device checks + ADR | Финальные production images/host deployment hardening |
| G06 LAN trust | mDNS/HTTPS реально доступны хотя бы с согласованных phone/laptop clients | PASS-PI/client Д009/Д010 | Core auth/cookies/WS/PWA, все browser targets |
| G07 Capture/ML path | Есть рабочий доступный capture/analyzer smoke и documentary route/доступ к четырём targets; local inference feasibility не упирается в известный тупик | PASS-HOST/SIM + platform research | Реальные четыре packages/M4, Scene quality/M5, gesture quality/M6 |
| G08 Reproducibility | Сохранены versions/config/protocol и повторён critical scenario | Re-execution report | Полная install acceptance/rollback pipeline |

M0_PASS относится только к этому scope. G07 может включать явно подписанные DOC-ONLY строки остальных ОС, потому что cross-platform acceptance предусмотрена позже; **это не PASS четырёх платформ**. Если не найден технически допустимый путь обязательного target, gate не закрывается одной ссылкой на документацию.

**Критические blockers:** нет Pi evidence, LED/audio конфликт не устранён, analog requirement обойдён заменой DAC, общий PCM не работает, непонятны необходимые privileges, power/signal instability мешают короткому стенду, TLS доверие подменено отключением verification. Наличие такого blocker → M0_NOT_READY и пересмотр зависимостей.

**Выход:** M0 report, `DAY_010_RESULT.md`, version/config/evidence index, обновлённый risk register и оценка. Создание красивого отчёта без повторного эксперимента не считается достижением G08.

---

## Минимальная ветка без оборудования

Пользователь подтвердил отсутствие замеров, но не отсутствие Pi. До начала каждого дня проверять фактическую доступность, не делать вывод из null, что компонента точно нет.

| День | Что можно закончить без Pi/периферии | Что останется BLOCKED |
|---|---|---|
| Д004 | Hypotheses, fake zones, host audio/PCM при наличии host audio, permissions design | Pi resources/audio coexistence |
| Д005 | Protocol, dataset/measurement plan, synthetic analyzer, runtime smoke доступного host | Недоступные actual OS/architecture captures |
| Д006 | Выбор supported image/checksum/install/recovery guide | Реальный boot/SSH/storage/power baseline |
| Д007 | Thermal policy/fake-clock probes, permission/pinmux plan | Physical fan/GPIO/bus/temperature/load proof |
| Д008 | Reproduction protocol/minimal code/config examples | Вся real LED/audio/PCM/partial-power matrix |
| Д009 | CA/proxy/mDNS design и локальный developer sandbox | Actual Pi/LAN/mobile trust evidence |
| Д010 | Evidence review, risk/estimate updates и handoff независимых задач | M0_PASS без critical hardware gates |

Рабочее время не превращается автоматически в ожидание: допустимо переместить независимые задачи, сохранив исходные ID/статусы. Ожидание доставки/доступа к Mac не считается успешно отработанным HIL-днём и может увеличить календарный срок.

## Формат каждого experiment record

```text
Experiment ID / день:
Требование / Q / ADR:
Гипотеза:
Уровень: DOC / SIM / HOST / PI
Окружение: board/OS/kernel/architecture/packages/commit
Стенд: LED SKU/counts/cables/pins/PSU/outputs/available clients
Предусловия и исходная конфигурация:
Команды/шаги:
Ожидаемый результат:
Фактические scalar measurements / observations:
Outcome: PASS-* / FAIL / BLOCKED / NOT_RUN
Evidence: redacted log/config/fixture/checksum
Ограничения метода / погрешность:
Cleanup / rollback:
Следующий шаг и влияние на график:
```

Для FAIL записывать, что изменили между прогонами, и менять один фактор там, где возможно. Не сравнивать latency двух runs с разной resolution/network/LED count без описания различия. Не хранить user screen/camera/audio ради отчётности.

## Версии и environment manifest

К Д010 manifest должен содержать фактически известные: Pi model/RAM, OS image/source/hash/release, kernel/firmware, storage type, driver commit/build flags, audio daemon/session manager/BlueZ versions, system/container user model, device/IPC allowlist, pinmux, LED SKU/counts/reference power limits, camera/capture runtime versions при проверке, proxy/mDNS/TLS tooling и platform test matrix.

Поля «планируется» и «измерено» разделять. Credentials, private network identifiers, SSH/device tokens, CA/server private keys, database/user backups не входят в публичный manifest.

## Как этот этап связан с требованиями

| Область | Основные requirement IDs | День / evidence |
|---|---|---|
| Local architecture / adapters | ARCH-001…ARCH-012 | Д004/Д005/Д010: boundaries и feasibility |
| Pi/storage/power | HW-001…HW-012 | Д006–Д008: короткий физический стенд, final installation позже |
| Thermal | THERM-001…THERM-003 | Д007: policy + real baseline/load |
| Zones/current | LIGHT-001…LIGHT-004, POWER-001…POWER-005 | Д004 sim / Д008 HIL; production implementation Д026–Д035 |
| Audio/DSP | AUDIO-001…AUDIO-017, DSP-001/DSP-002 | Д004/Д008: feasibility, не полный functional acceptance |
| Capture/Scene | AMBI-001…AMBI-021, SCENE-001…SCENE-008 | Д005: routes/пробы/методика; full implementation Д056–Д080 |
| Camera/privacy | GEST-001/GEST-007, PRES-003/PRES-006, PRIV-001…PRIV-003 | Д005/Д006: policy/package feasibility; real CV позже |
| API/network/health | API-005…API-008, OBS-004/OBS-005 | Д009: HTTPS/LAN bootstrap; health contract из ADR-004 |
| Build/quality/performance | CI-001…CI-011, TEST-001…TEST-009, PERF-001…PERF-008 | Д005 methodology / Д010 reproducibility; полный CI/performance позже |

Ссылки на IDs задают область будущих работ; они не означают, что перечисленные требования все реализованы в эти семь дней. Полная [матрица](../requirements.md) и 62 пункта приёмки сохраняются.

## Первичные источники для выполнения плана

Проверка доступности/содержания при подготовке плана: 25 октября 2026 года. Повторно сверить versions/API prerequisites в день эксперимента.

- [Аппаратные источники Д002](../../hardware/research/SOURCES.md): Pi pinout, LED driver, sensor/shifter datasheets, Pi power/camera docs.
- [WirePlumber Bluetooth configuration](https://pipewire.pages.freedesktop.org/wireplumber/daemon/configuration/bluetooth.html): headless/logind/BlueZ integration. Версия страницы при проверке 0.5.18; это не автоматически выбранная package version проекта.
- [Microsoft Desktop Duplication API](https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api): первичный candidate path для Windows capture; native binding/minimum supported client определить в spike.
- [XDG ScreenCast portal](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.ScreenCast.html): candidate Linux Wayland flow; availability реального backend/permissions проверить на целевой session.
- [Apple ScreenCaptureKit](https://developer.apple.com/documentation/screencapturekit): официальный вход в macOS capture docs. В инструменте чтения страница потребовала JavaScript, Markdown вариант не открылся; availability/permissions/native packaging **не объявлены проверенными**, их нужно сверить через Apple docs/Xcode на Mac.
- [Ubuntu camera stack](https://documentation.ubuntu.com/hardware-support/boards/how-to/special_hardware/rpi-camera/): камера на Ubuntu и supported packages; страницу не использовать как recommendation устаревшей версии ОС.
- [Ubuntu release lifecycle](https://ubuntu.com/about/release-cycle): выбор реально поддерживаемого release для Д006. Точная версия в этом плане не закреплена без проверки board/audio/camera package compatibility.

## Итоговый комплект файлов к окончанию Д010

Ожидаются reports `docs/DAY_004_RESULT.md` … `docs/DAY_010_RESULT.md`, M0 baseline report, experiment README/scripts/fixtures, image/package manifests, install/TLS/recovery guides, actual test fixture/pin permissions, candidate protocols/quality plans и обновлённые ADR/backlog/questions.

Отчёты должны содержать actual outcomes, а не копию будущего плана с галочками. Hardware/OS/platform blockers и remaining model/quality work остаются видимыми. С этого комплекта начинается основная реализация Core на Д011 при выполненных зависимостях.
