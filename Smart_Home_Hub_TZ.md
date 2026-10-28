# ТЗ: Smart Home Hub на базе Raspberry Pi 4

Версия: 1.1 (26.10.2026: подтверждена Raspberry Pi 4B с2GB RAM)  
Статус: к разработке  
Тип проекта: pet-project с возможностью развития в курсовую/дипломную работу

---

## 1. Назначение

**Smart Home Hub** — локальная интеллектуальная система управления мультимедийным окружением комнаты и устройствами умного дома на базе Raspberry Pi 4.

Система объединяет:

- управление адресной LED-подсветкой;
- музыкально-адаптивное освещение;
- Ambilight-подсветку внешнего монитора;
- воспроизведение и управление музыкой;
- жестовое управление через камеру;
- локальную ML-классификацию характера видеосцен;
- предиктивную автоматизацию на основе привычек пользователя;
- определение присутствия пользователя;
- работу с физическими датчиками;
- пользовательские automation rules;
- управление через Web UI и Telegram-бота;
- MQTT-интеграцию для будущих IoT-устройств;
- потенциальную интеграцию с Home Assistant.

Ключевой принцип проекта — **local-first**: основные функции, персональные данные, CV/ML-обработка и история пользовательского поведения должны работать и храниться локально.

Интернет не является обязательным для базовой работы Smart Home Hub. Исключение — Telegram-бот и внешние сервисы, явно включённые пользователем.

---

## 2. Ключевые архитектурные решения

1. Центральный контроллер — **Raspberry Pi 4 Model B (4B), только 2 GB RAM**.
2. ОС — **Ubuntu Server**.
3. Система работает headless.
4. Основной backend/control plane реализуется на **Go**.
5. ML/CV/DSP-функциональность реализуется на **Python**.
6. Архитектура — **модульный Go-монолит + отдельный Python ML Runtime**.
7. Контейнеризация — **Docker Compose**.
8. Автозапуск Docker Compose выполняется через **systemd**.
9. Для разработки и CI поддерживается simulation mode без Raspberry Pi.
10. Внутреннее взаимодействие компонентов — через REST/WebSocket/event bus.
11. Для IoT-расширения используется **MQTT**.
12. Основная БД — **SQLite в WAL mode**.
13. Архитектура данных и API проектируется так, чтобы переход от одного пользователя к нескольким не требовал переписывания ядра.
14. Репозиторий — публичный monorepo на **GitHub**.
15. Продакшн-деплой выполняется автоматически по Git tag через GitHub Actions и self-hosted runner на Raspberry Pi.
16. Все изображения камеры, кадры экрана и чувствительные ML-данные не покидают локальную систему.

---

## 3. Аппаратная платформа

### 3.1 Raspberry Pi

Основное устройство:

- Raspberry Pi 4 Model B (4B);
- RAM: **только 2 GB** (единственное доступное устройство, подтверждено владельцем26.10.2026);
- Ubuntu Server;
- headless mode;
- первоначально допустима microSD;
- целевой вариант хранения — USB SSD;
- поддерживается процедура миграции microSD → USB SSD.

### 3.1.1 Обязательное ограничение памяти: 2 GB

Целевой и единственный доступный контроллер — **Raspberry Pi 4B с 2 GB RAM**.
Разработка, BOM, deployment и совместные нагрузочные испытания не должны
предполагать наличие версии с 4/8 GB либо скрытую замену платы. После boot доступная
ОС память меньше номинальной; фактический MemTotal/MemAvailable фиксируется на плате.

- Начальный плановый envelope всей системы вместе с ОС — не более1536MiB,
  оставшийся номинальный запас не считается заранее доступной памятью приложения.
- Для Core текущий dev container limit128MiB, Go soft memory limit96MiB,
  одна Argon2id job одновременно. Лимиты будущих сервисов подбираются измерениями;
  предварительный [бюджет](configs/pi4b-2gb-budget.example.json) не доказывает их достаточность.
- ML/CV/DSP используют bounded queues/latest-frame, compact/quantized модели,
  ограниченные resolution/FPS и один тяжёлый worker. Совместимость всех функций
  и performance targets подтверждается на2GB плате под совместной нагрузкой.
- Обучение планируется небольшими batch без одновременной тяжёлой inference-нагрузки;
  если требуемая модель не укладывается, сначала исследовать оптимизации и зафиксировать
  blocker, не исключать функциональные требования молча.
- Full Grafana/Prometheus — optional и по умолчанию выключены на Pi. Контейнер
  frontend в production должен раздавать готовый bundle, а не запускать Node/Vite dev server.
- CI builds и тяжёлые dependency installs выполняются на development/hosted CI,
  Pi runner — deployment/smoke; swap не считается дополнительной физической RAM.
- При memory pressure ограничивается optional ML/CV работа; local control plane
  должен сохранять работу. OOM и непрерывный swap не считаются успешной приёмкой.

### 3.1.2 Разделение нагрузки и порядок деградации

Pi4B/2GB — control plane и hardware hub. Screen capture, Ambilight analysis и
Scene ML работают на ноутбуке; на Pi остаются приём зональных цветов и LED output.
Один лёгкий Python Runtime на Pi обслуживает DSP/gesture/presence по потребности:
модели загружаются лениво, количество workers и очередей ограничено, не требуется
постоянно держать все модели в памяти. Тяжёлые Python workers, PostgreSQL и Kubernetes
исключены из текущей архитектуры; SQLite, Docker Compose и minimal images сохраняются.

Приоритет сохранения функций при memory pressure (сверху — важнее):

1. Smart Home Core: auth/safety, persistence и ручное управление.
2. Lighting.
3. Audio.
4. Automation.
5. Ambilight.
6. Sensors / MQTT.
7. Presence.
8. Gesture ML.
9. Остальные ML-функции.

Деградация начинается снизу. Safety/thermal monitoring и emergency OFF относятся
к первому уровню независимо от названия модуля. При остановке CV presence становится
UNKNOWN/stale, не AWAY; отсутствие модели не разрешает automation отключать свет.
Отключение Scene ML не останавливает laptop color analysis/Ambilight output.

Ориентир обычной загрузки — около70–75% фактического usable MemTotal по оценке
`1 - MemAvailable/MemTotal`, то есть примерно25–30% available headroom. Суммарный
budget1.4–1.5GB — ориентир всей системы вместе с ОС, не лимит приложений сверх ОС.
Reclaimable page cache не считается безвозвратно занятой памятью; измеряются также
process/cgroup peaks, swap activity, CPU/thermal и latency. Только RAM не считается
единственным возможным bottleneck.

Сначала уменьшаются optional workload/inference rate, затем освобождаются модели;
возврат RSS проверяется измерением и не предполагается автоматически от Python del.
При необходимости допустим контролируемый restart Python Runtime без остановки Core.
Восстановление — после устойчивого снятия pressure с hysteresis, с действующими
permissions/leases; camera не получает новое разрешение автоматически. Планируемый
controller должен пройти fake-clock/fault tests; наличие этого текста не означает,
что автоматическая деградация уже реализована.

Host swap на SSD допускается как optional страховка и не считается RAM или способом
пройти realtime acceptance. Сейчас swap настройки host не изменяются; действующий
Core dev container по-прежнему не использует swap. Grafana/Prometheus — optional/off.

Все исходные performance targets сохраняются как цели; гарантия их одновременного
достижения на2GB до физических измерений не даётся. Подробности:[профиль2GB](docs/development/pi4b-2gb.md).

### 3.2 Охлаждение

Raspberry Pi должен использовать активное охлаждение:

- радиатор на SoC;
- вентилятор 30–40 мм;
- предпочтительно GPIO-управляемый вентилятор через транзистор/MOSFET.

Thermal Manager должен считывать:

```text
/sys/class/thermal/thermal_zone0/temp
```

и формировать состояния:

```text
NORMAL
WARM
HOT
CRITICAL
```

Ориентировочная политика:

```text
< 50°C      FAN OFF
50–65°C     FAN LOW / ON
> 65°C      FAN ON
> 75°C      WARNING
```

Пороговые значения должны быть конфигурируемыми.

---

## 4. LED-подсветка

### 4.1 Стандарт LED

Для проекта используется единый стандарт адресной LED-подсветки:

**WS2812B, 60 LED/м.**

Поддержка нескольких типов LED-протоколов в MVP не требуется.

### 4.2 Зоны

Первая версия содержит минимум две независимые зоны:

```text
monitor.backlight
desk.wall
```

#### `monitor.backlight`

Назначение:

- лента по четырём сторонам внешнего монитора;
- основной режим — Ambilight;
- независимые настройки яркости и эффектов.

#### `desk.wall`

Назначение:

- LED-лента вдоль стола со стороны стены;
- музыкально-адаптивные эффекты;
- декоративные эффекты;
- Night Shift.

Архитектура должна поддерживать произвольное количество зон без изменения бизнес-логики.

Пример:

```text
room.monitor
room.desk
room.wall
room.bed
```

### 4.3 Модель LED-зоны

Пример конфигурации:

```yaml
id: monitor.backlight
driver: ws281x
gpio_pin: 18
led_count: 120
brightness_limit: 0.6
max_current_amp: 4.0
start_corner: bottom_left
direction: clockwise
geometry: perimeter
```

Каждая зона имеет:

- уникальный ID;
- mode;
- brightness;
- master brightness modifier;
- effect parameters;
- power limit;
- LED mapping;
- статус;
- fallback preset.

### 4.4 Режимы

Обязательные режимы:

```text
OFF
STATIC
BREATHING
RAINBOW
MUSIC_PULSE
MUSIC_BEAT
MUSIC_SPECTRUM
MUSIC_WAVE
MUSIC_CENTER
AMBILIGHT
NIGHT_SHIFT
```

### 4.5 Цвет

Поддержать:

- RGB;
- HSV;
- ручной выбор цвета;
- эквивалент цветовой температуры;
- тёплую/холодную подсветку.

### 4.6 Fade

Переключение режимов должно происходить плавно.

Параметр:

```text
transition_duration_ms
```

должен быть настраиваемым.

---

## 5. Электропитание LED

LED-лента не питается от Raspberry Pi.

Используется отдельный мощный внешний PSU.

Принцип:

```text
External PSU
   ├── LED +V
   └── GND ────────────┐
                       │
Raspberry Pi GND ──────┘

Raspberry Pi GPIO
       │
       ▼
3.3V → 5V Level Shifter
       │
       ▼
LED DATA
```

Обязательные/рекомендуемые элементы:

- level shifter 3.3V → 5V;
- рекомендуемый вариант: 74AHCT125 / 74HCT245;
- последовательный резистор на data line;
- конденсатор на питании LED;
- аппаратный предохранитель;
- физическая возможность отключить питание LED;
- общая земля Raspberry Pi и LED PSU.

Для длинных 5V-линий предусмотреть power injection в нескольких точках.

---

## 6. Power Limiter

Lighting Engine должен иметь программный ограничитель энергопотребления.

Для каждой зоны задаётся:

```text
max_current_amp
```

Также задаётся общий предел PSU:

```text
global_max_current_amp
```

Перед отправкой LED-кадра система оценивает ток по RGB-значениям.

Если:

```text
estimated_current > configured_limit
```

яркость всего кадра должна быть пропорционально уменьшена.

Пример:

```text
requested_current = 7.2 A
limit = 5.0 A

scale = 5.0 / 7.2
```

Power Limiter должен работать:

- по каждой зоне;
- глобально;
- независимо от выбранного эффекта.

---

## 7. Архитектура системы

```text
                    ┌─────────────────────────┐
                    │      Web UI / PWA       │
                    └────────────┬────────────┘
                                 │ HTTPS / WS
                                 ▼
┌─────────────────────────────────────────────────────┐
│               Smart Home Core (Go)                  │
│                                                     │
│  API / Auth / Sessions                              │
│  Event Bus                                          │
│  Device Manager                                     │
│  Lighting Engine                                    │
│  Audio Manager                                      │
│  Ambilight Controller                               │
│  Automation Engine                                  │
│  Presence Engine                                    │
│  Sensor Manager                                     │
│  MQTT Integration                                   │
│  Storage                                            │
│  Thermal Manager                                    │
│  System Manager                                     │
└─────────┬─────────────┬─────────────┬───────────────┘
          │             │             │
          │             │             └── GPIO / I²C / SPI
          │             │
          │             └── MQTT → external IoT
          │
          ▼
┌────────────────────────────┐
│     ML Runtime (Python)    │
│                            │
│ Gesture Recognition        │
│ Habit Prediction           │
│ Presence CV                │
└────────────────────────────┘
```

ML-анализ видеосцены для Ambilight выполняется на screen-client ноутбука.

---

## 8. Внутренний Event Bus

Smart Home Core должен использовать единый событийный интерфейс.

Примеры событий:

```text
BEAT_DETECTED
GESTURE_RECOGNIZED
SCENE_CHANGED
TV_SIGNAL_LOST
LIGHT_MODE_CHANGED
USER_ACTION
DEVICE_CONNECTED
DEVICE_DISCONNECTED
SENSOR_UPDATED
USER_PRESENT
USER_AWAY
AUTOMATION_TRIGGERED
SYSTEM_WARNING
```

Компоненты не должны напрямую зависеть друг от друга там, где взаимодействие может быть выражено через событие.

---

## 9. Audio subsystem

### 9.1 Источники

Поддержать:

- Bluetooth A2DP;
- Spotify Connect;
- локальные аудиофайлы.

AirPlay в MVP не входит.

### 9.2 Bluetooth

Raspberry Pi должен отображаться как Bluetooth-аудиоустройство:

```text
Smart Home Hub
```

Pairing новых устройств разрешается только после явной активации `PAIRING_MODE`.

### 9.3 Audio output

Поддержать:

- 3.5 mm;
- USB Audio / USB DAC;
- Bluetooth output.

Пользователь должен иметь возможность выбирать output device через Web UI.

### 9.4 Управление

Поддержать:

```text
Play
Pause
Next
Previous
Volume
Mute
```

### 9.5 Метаданные

Если источник предоставляет данные, отображать:

- track;
- artist;
- album;
- cover;
- progress.

История прослушивания не сохраняется.

---

## 10. DSP и музыкально-адаптивная подсветка

Обрабатывать:

```text
overall volume
bass
mids
treble
beat
BPM
```

Использовать:

- FFT;
- beat detection;
- автоматическую нормализацию чувствительности;
- ручной sensitivity;
- smoothing.

Параметры пользователя:

```text
sensitivity
smoothing
brightness
effect
```

Эффекты:

### MUSIC_PULSE

Громкость управляет общей яркостью.

### MUSIC_BEAT

Каждый beat создаёт вспышку или импульс.

### MUSIC_SPECTRUM

Bass/mids/treble отображаются разными участками или цветами.

### MUSIC_WAVE

Музыкальные импульсы движутся вдоль LED-ленты.

### MUSIC_CENTER

Импульс появляется в центре и расходится в две стороны.

Web UI должен показывать realtime:

```text
Bass
Mids
Treble
BPM
Beat
```

История DSP-данных не сохраняется.

---

## 11. Ambilight

### 11.1 Источник изображения

Ambilight работает через отдельный **Go screen-client**, установленный на ноутбук.

Поддерживаемые платформы:

```text
Linux x86_64
Windows x86_64
macOS arm64
macOS x86_64
```

Ноутбук может быть подключён к монитору через HDMI или USB-C/HDMI-переходник.

### 11.2 Privacy

Кадры экрана анализируются только на ноутбуке.

Raspberry Pi не получает полноразмерное изображение.

В Hub передаются:

- рассчитанные цвета зон;
- технические metadata;
- результаты scene classification.

Снимки экрана не сохраняются.

### 11.3 Pairing screen-client

Flow:

```text
Web UI
  ↓
Add screen client
  ↓
One-time pairing code
  ↓
Laptop client
  ↓
Hub issues device token
```

### 11.4 Геометрия

LED размещаются по всем четырём сторонам монитора.

Для корректного mapping указываются:

```text
top_led_count
right_led_count
bottom_led_count
left_led_count

start_corner
direction
```

Направление:

```text
clockwise
counter_clockwise
```

### 11.5 Алгоритм цвета

Используется **weighted average**.

Каждой LED-зоне экрана соответствует полоса пикселей.

Пиксели у физического края экрана имеют больший вес.

Формула:

```text
R = Σ(Ri × wi) / Σwi
G = Σ(Gi × wi) / Σwi
B = Σ(Bi × wi) / Σwi
```

Пример весов:

```text
edge row 1   weight 1.0
edge row 2   weight 0.8
edge row 3   weight 0.5
edge row 4   weight 0.2
```

### 11.6 Обработка изображения

Поддержать:

- black-bar detection;
- suppression почти чёрных пикселей;
- saturation adjustment;
- gamma correction;
- color boost;
- smoothing;
- min/max brightness;
- configurable update rate.

### 11.7 Производительность

Цели:

```text
Ambilight FPS: 30
end-to-end latency: < 100 ms
```

### 11.8 Профили

Обязательный:

```text
Cinema Mode
```

Опциональный:

```text
Gaming Mode
```

Cinema Mode:

- более высокий smoothing;
- мягкие переходы;
- пониженная агрессивность color boost.

Gaming Mode:

- минимальный smoothing;
- минимальная задержка;
- быстрая реакция.

### 11.9 Fallback

При потере screen-client зона `monitor.backlight` переходит в заданный fallback preset.

### 11.10 Screen inactivity

Screen-client должен определять:

- выключенный монитор;
- длительно статичное изображение.

При необходимости Ambilight должен переходить в fallback/off-состояние.

---

## 12. ML Scene Adaptation

Классифицируется не жанр фильма, а **характер текущей сцены**.

Поддерживается multi-label классификация.

Классы:

```text
CALM
DYNAMIC
DARK
BRIGHT
HIGH_MOTION
LOW_MOTION
COLORFUL
LOW_CONTRAST
HIGH_CONTRAST
```

Одна сцена может иметь несколько меток.

Пример:

```text
DARK + HIGH_MOTION + DYNAMIC
```

### 12.1 Выполнение

Scene ML выполняется на ноутбуке в составе screen-client.

Причины:

- меньше нагрузка на Raspberry Pi;
- кадры не передаются по сети;
- лучше privacy;
- выше доступная вычислительная мощность.

### 12.2 Частота

ML-анализ:

```text
1 inference / second
```

При этом Ambilight остаётся на 30 FPS.

Для анализа динамики используется последовательность нескольких кадров.

### 12.3 Воздействие ML

ML не переключает режимы.

Он только корректирует:

```text
brightness
saturation
smoothing
color_boost
transition_speed
```

Пример:

```text
DARK
→ brightness ↓

DYNAMIC / HIGH_MOTION
→ smoothing ↓
→ transition_speed ↑

CALM / LOW_MOTION
→ smoothing ↑

COLORFUL
→ color_boost ↓ при необходимости
```

Пользователь настраивает:

```text
ml_influence = 0..100%
```

При `0%` используется обычный Ambilight.

История scene classification не хранится.

---

## 13. Gesture Recognition

### 13.1 Камера

Поддерживается:

- USB webcam;
- Raspberry Pi Camera Module.

Конкретная модель выбирается позже исходя из цены и совместимости.

Камера располагается рядом с монитором/Raspberry Pi.

### 13.2 Активация

Камера для gesture recognition активируется только при включённом Gesture Mode.

### 13.3 Жесты

```text
OPEN_PALM
→ Play/Pause

SWIPE_RIGHT
→ Next Track

SWIPE_LEFT
→ Previous Track

THUMB_UP
→ Brightness +

THUMB_DOWN
→ Brightness -

FIST
→ Lights OFF

TWO_FINGERS
→ Next Lighting Preset

PALM_FACE_DOWN + downward movement
→ Night Shift

ROTATE_CLOCKWISE
→ Volume +

ROTATE_COUNTER_CLOCKWISE
→ Volume -
```

### 13.4 Защита от ложных срабатываний

Поддержать:

```text
gesture_confirmation_ms = 300..500
gesture_cooldown_ms ≈ 1000
```

Также используется activation gesture, открывающий окно жестового управления на ограниченное время.

Распознавание выполняется по всей области кадра.

### 13.5 Privacy

Кадры:

- обрабатываются только в RAM;
- не сохраняются;
- не отправляются во внешние сервисы;
- не передаются в облако.

### 13.6 UI

Web UI показывает:

```text
Gesture recognition: ON
Detected gesture
Confidence
Resolved action
```

Gesture → Action mapping должен редактироваться пользователем.

---

## 14. Night Shift

Night Shift:

```text
color_temperature ≈ 2200–2700K equivalent
brightness = 10–20%
blue channel strongly reduced
animations disabled
transition = 1–2 sec fade
```

Значения должны быть настраиваемыми.

---

## 15. Presence Engine

Presence Engine объединяет несколько источников:

```text
Camera person detection
Bluetooth phone presence
Network presence
PIR sensor
```

Результирующие состояния:

```text
PRESENT
PROBABLY_PRESENT
AWAY
UNKNOWN
```

### 15.1 Camera presence

Используется та же камера, что для gesture recognition.

Для presence detection распознаётся:

```text
person detected = yes/no
```

Идентификация личности в MVP не выполняется.

Потенциальное face recognition предусматривается как future extension.

Интервал CV presence detection:

```text
2–3 sec
```

### 15.2 Privacy indicator

При активной CV-обработке должен включаться отдельный LED-индикатор на Raspberry Pi.

---

## 16. Sensors

MVP поддерживает:

- ambient light sensor;
- temperature sensor;
- humidity sensor;
- PIR motion sensor;
- air quality sensor.

Подключение возможно через:

```text
GPIO
I²C
SPI
```

### 16.1 Ambient light

Используется для автоматической коррекции яркости:

```text
room brighter
→ LED brightness lower

room darker
→ LED brightness higher
```

### 16.2 Temperature/Humidity/Air Quality

Показываются в Web UI и доступны Automation Engine.

Автоматизация не должна вмешиваться в активный музыкальный lighting mode, если её действие меняет используемую музыкальную сцену, если только правило явно не настроено пользователем.

---

## 17. Habit Prediction

### 17.1 Сбор данных

Сохраняются:

```text
timestamp
day_of_week
light_zone
light_mode
brightness
selected_preset
music_active
tv_active
ambilight_active
gesture_used
manual_override
state_duration
```

Последовательности действий между событиями не моделируются в MVP.

### 17.2 Предсказываемые действия

- preset;
- brightness;
- lighting mode;
- включение Ambilight;
- Night Shift;
- запуск музыки;
- volume.

### 17.3 Learning period

До включения предиктивной автоматизации система должна собирать данные не менее:

```text
30 days
```

### 17.4 Режимы

```text
SUGGEST
AUTO
```

Пример confidence policy:

```text
confidence >= 85%
→ auto

60% <= confidence < 85%
→ suggestion

confidence < 60%
→ no action
```

Пороговые значения настраиваемые.

### 17.5 Контекст

Учитывать:

- time of day;
- day of week;
- monitor activity;
- music state;
- selected lighting mode;
- brightness;
- gestures;
- user presence;
- ambient light.

### 17.6 Feedback

Пользователь может:

- принять recommendation;
- отклонить recommendation.

Отказ используется как отрицательный feedback.

Ручная отмена уже выполненной автоматизации не считается feedback-сигналом в MVP.

### 17.7 Privacy

**ОБЯЗАТЕЛЬНО:** все данные о привычках пользователя хранятся и обрабатываются исключительно локально.

Запрещено:

- отправлять habit data в облачные API;
- использовать внешние ML API;
- синхронизировать историю с third-party сервисами.

### 17.8 Retention

Используется rolling window:

```text
90 days
```

### 17.9 Delete learning history

При удалении истории должны быть удалены:

- observations;
- feedback;
- обученное пользовательское состояние/модель.

После удаления запускается новый 30-дневный learning period.

---

## 18. Automation Engine

Поддерживает пользовательские правила:

```text
IF conditions
THEN actions
```

### 18.1 Conditions

Поддержать:

- time;
- day of week;
- user presence;
- monitor on/off;
- music playing;
- Ambilight active;
- brightness;
- ambient light;
- gesture event;
- ML event;
- MQTT event;
- device state;
- sensor values.

### 18.2 Logic

Поддержать:

```text
AND
OR
nested conditions
```

### 18.3 Delayed actions

Пример:

```text
IF user_present = false
THEN wait 5 min
THEN lights = OFF
```

### 18.4 Schedules

Пример:

```text
Mon–Fri 07:30 → Work preset
Every day 23:30 → Night Shift
```

Одноразовые сценарии в MVP не требуются.

### 18.5 Multi-action scenarios

Пример:

```text
Movie Mode

monitor.backlight = AMBILIGHT
desk.wall = STATIC
desk.wall.temperature = WARM
desk.wall.brightness = 20%
music = PAUSE
```

### 18.6 Rule priority

Каждое правило имеет:

```text
priority
```

При конфликте выполняется правило с более высоким приоритетом.

### 18.7 Manual Override

Ручное действие пользователя создаёт временный manual override.

Default:

```text
15 min
```

Automation Engine не должен менять затронутое состояние до завершения override.

### 18.8 Audit

Хранить журнал:

- rule;
- trigger;
- timestamp;
- actions;
- result;
- error.

Retention:

```text
30 days
```

---

## 19. Physical Button

В MVP используется минимум одна физическая кнопка.

Default mapping:

```text
single press
→ lights ON/OFF

double press
→ Night Shift

long press
→ Next Preset
```

Mapping хранится в SQLite и редактируется через Web UI.

---

## 20. LED Panel

Система должна поддерживать LED matrix/panel на базе WS2812B.

Тип устройства:

```text
LED Device
├── STRIP
└── MATRIX
```

Для MATRIX:

```text
width
height
orientation
serpentine_layout
```

Конкретная панель может быть добавлена позже.

---

## 21. Device Manager

Модель:

```text
Device
├── id
├── type
├── status
├── capabilities
└── metadata
```

Статусы:

```text
ONLINE
OFFLINE
ERROR
DISABLED
```

Пример capabilities:

```text
rgb
per_pixel
brightness
temperature
humidity
illuminance
motion
air_quality
audio_output
```

Поддержать:

- ручное добавление устройств;
- MQTT autodiscovery;
- Bluetooth discovery;
- Bluetooth pairing;
- hardware adapters для GPIO/I²C/SPI.

---

## 22. MQTT

MQTT закладывается для:

- ESP32;
- внешних датчиков;
- будущих реле;
- Home Assistant;
- других IoT-устройств.

Broker: Mosquitto либо эквивалент.

Требования:

- anonymous access запрещён;
- username/password;
- ACL;
- доступ только LAN/Docker network.

Примеры topics:

```text
smarthome/devices/+/state
smarthome/devices/+/set
smarthome/events/#
smarthome/sensors/+/state
```

---

## 23. Home Assistant

Home Assistant не является обязательной частью Smart Home Hub.

Smart Home Hub должен оставаться автономным.

Интеграция реализуется в будущем через:

- MQTT;
- REST API.

Основное ядро не должно зависеть от Home Assistant.

---

## 24. Web UI

Стек:

```text
React
TypeScript
Vite
```

Интерфейс:

- responsive;
- PWA;
- ориентирован на управление с телефона.

Страницы:

```text
Dashboard
Lighting
Audio
Ambilight
Gestures
Automations
Predictions
Devices
System
Settings
```

### 24.1 Dashboard

Показывать:

```text
System status

Raspberry Pi
CPU
RAM

Audio
Playing
Volume

Lighting
monitor.backlight
desk.wall

Automations
Enabled/Disabled
Last trigger
```

### 24.2 Realtime

Используется WebSocket:

```text
/ws/events
```

События:

- audio spectrum;
- beat;
- device state;
- lighting state;
- gesture event;
- system metrics;
- automation event.

### 24.3 LED visualizer

Web UI должен визуализировать текущее состояние LED.

### 24.4 Ambilight preview

Показывается только цветовая рамка зон.

Полноценный screenshot в Hub не передаётся.

### 24.5 Preset editor

Редактор должен позволять настраивать несколько зон в одном preset.

### 24.6 Automation editor

Визуальный rule builder:

```text
IF
[ condition ]
AND/OR
[ condition ]

THEN
[ action ]
```

---

## 25. Telegram Bot

Telegram Bot входит в MVP.

Реализуется отдельным Go-сервисом/контейнером.

Telegram Bot не имеет прямого доступа к hardware/storage.

Он взаимодействует только с Smart Home Core API.

Команды:

```text
/status
/lights
/music
/presets
/ambilight
/night
/system
```

Поддержать inline buttons.

Telegram является optional internet-dependent interface.

Основные функции Smart Home Hub должны работать без Telegram.

Запрещено передавать через Telegram:

- camera frames;
- screen frames;
- habit history;
- ML data;
- CV data.

Telegram user ID должен быть явно привязан к локальному пользователю.

---

## 26. Authentication и Multi-user readiness

Хотя MVP рассчитан на одного пользователя, модель проектируется сразу с:

```text
User
Role
Session
```

Первый пользователь:

```text
role = admin
```

Authentication:

```text
username + password
```

Пароли хранятся только как:

```text
Argon2id hash
```

Sessions:

- HttpOnly cookies;
- Secure;
- SameSite;
- expiration;
- logout;
- revoke all sessions.

---

## 27. Machine API Tokens

Поддержать API tokens для:

- screen-client;
- Telegram service;
- external apps;
- Home Assistant;
- CLI;
- IoT integrations.

Токен:

- показывается только при создании;
- хранится в БД только как hash;
- может быть revoked;
- имеет scopes.

Пример scopes:

```text
lights:read
lights:write
audio:control
devices:read
automation:write
system:read
```

---

## 28. REST API

Префикс:

```text
/api/v1
```

Ресурсы:

```text
/auth
/users

/zones
/lights
/presets

/audio

/ambilight

/gestures

/automations

/predictions

/devices

/system
/config
```

API должен иметь OpenAPI/Swagger specification.

Rate limiting должен быть предусмотрен.

Пример:

```text
Web UI       120 req/min
External API 60 req/min
Auth         10 req/min
```

---

## 29. mDNS и HTTPS

Hub должен быть доступен по:

```text
https://smarthome.local
```

Использовать mDNS.

HTTPS обязателен.

Для локальной установки допускается собственный локальный CA / self-signed bootstrap с последующей установкой доверия.

---

## 30. Remote Access

По умолчанию Smart Home Hub доступен только из LAN.

API не должен напрямую публиковаться в интернет.

Для удалённого доступа рекомендуются:

- Tailscale;
- WireGuard.

---

## 31. Storage

Основная БД:

```text
SQLite
```

Режим:

```text
WAL
```

Хранить:

- users;
- roles;
- sessions;
- LED zones;
- devices;
- presets;
- automation rules;
- automation log;
- gesture mappings;
- system settings;
- habit observations;
- prediction feedback;
- API tokens.

Не хранить:

- audio spectrum history;
- FFT history;
- beat history;
- screen frames;
- camera frames;
- Ambilight frame history.

---

## 32. Configuration

Три уровня:

### `.env`

Секреты и инфраструктура:

```text
SESSION_SECRET
TELEGRAM_BOT_TOKEN
DATABASE_PATH
MQTT_USERNAME
MQTT_PASSWORD
```

### `config.yaml`

Hardware/infrastructure:

```yaml
server:
  host: 0.0.0.0
  port: 8080

lighting:
  gpio_pin: 18
  max_current_amp: 5.0

ml:
  endpoint: http://ml-runtime:8090

mqtt:
  enabled: true
  broker: mqtt://mosquitto:1883
```

### SQLite

Runtime/user settings:

- brightness;
- presets;
- automations;
- gesture mappings;
- ML preferences.

Поддержать:

```text
development
test
production
```

---

## 33. Public Repository Security

Репозиторий публичный.

Запрещено коммитить:

- `.env`;
- реальные Telegram tokens;
- MQTT credentials;
- session secrets;
- private keys;
- TLS private keys;
- реальные device tokens;
- backup БД;
- персональные данные;
- приватные IP/domain данные, если они чувствительны.

Хранить:

```text
.env.example
config.example.yaml
```

Secrets для CI/CD — только через GitHub Secrets / self-hosted runner environment.

---

## 34. Simulation Mode

Backend должен запускаться без Raspberry Pi.

Hardware interfaces должны иметь mock implementations.

Simulation mode должен поддерживать:

- virtual LED strip;
- virtual matrix;
- fake sensors;
- fake audio events;
- fake MQTT devices;
- fake presence;
- fake screen-client.

Это обязательно для локальной разработки и CI.

---

## 35. Backup и Recovery

Автоматический backup:

```text
1 раз в сутки
```

Хранить:

```text
последние 7 backup
```

Backup включает:

- SQLite;
- user config;
- automation rules;
- presets.

Поддержать restore через CLI или Web UI.

Factory reset должен:

- удалить user data;
- удалить runtime config;
- удалить learning history;
- оставить установленное ПО.

---

## 36. Reliability

### ML Runtime failure

При падении ML Runtime:

```text
Lighting → работает
Audio → работает
Ambilight → работает без ML
Automation → работает без ML
```

### MQTT failure

GPIO/I²C/SPI устройства продолжают работать.

### SQLite failure

Safe Mode:

```text
Lighting manual control available
Audio available
Automation disabled
Persistent changes disabled
```

### Restart state

После рестарта:

```text
LED zones → restore last state
Music → STOPPED
Automation → enabled
Camera → disabled until required
```

### Watchdog

Система должна контролировать зависшие компоненты.

### Graceful shutdown

Последовательность:

```text
stop accepting commands
stop automations
fade LEDs to OFF
flush SQLite
stop ML runtime
shutdown
```

---

## 37. systemd / Docker Compose

systemd запускает Docker Compose.

Пример policy:

```text
Restart=always
RestartSec=5
```

Контейнеры также должны иметь restart policies и healthchecks.

---

## 38. Observability

### Logging

Structured JSON logs.

Levels:

```text
DEBUG
INFO
WARN
ERROR
```

По возможности использовать correlation/event IDs.

### Metrics

Prometheus:

```text
/metrics
```

Метрики:

```text
system_cpu_percent
system_memory_bytes
system_temperature_celsius

led_fps
led_frame_latency_ms
led_estimated_current_amp

ambilight_fps
ambilight_latency_ms

audio_fft_latency_ms
audio_bpm
audio_beats_total

gesture_inference_ms
gesture_recognized_total

presence_state
presence_inference_ms

scene_inference_ms

automation_trigger_total
automation_error_total

mqtt_connected_devices

http_requests_total
http_request_duration_ms

ws_clients

ml_runtime_up
```

### Grafana

Grafana — optional Docker Compose profile.

### Health

```text
GET /health
GET /ready
```

`/ready` проверяет:

- SQLite;
- MQTT;
- ML Runtime;
- hardware adapters.

### Alerts

Alerts:

- CPU temperature high;
- LED power limit reached;
- sensor offline;
- ML Runtime down;
- disk almost full.

Доставка:

- Web UI;
- Telegram, если доступен;
- system logs.

---

## 39. Security Events

Логировать:

- login success/failure;
- API token create/revoke;
- Bluetooth pairing;
- unknown MQTT client;
- critical settings changes;
- factory reset.

---

## 40. Privacy Requirements

**ОБЯЗАТЕЛЬНО:**

```text
camera frames
→ RAM only

screen frames
→ laptop only

habit data
→ local only

ML inference
→ local only, кроме технически необходимого screen-client на локальном ноутбуке

audio analysis
→ local only
```

Облачные ML API запрещены для основной функциональности.

---

## 41. Repository Structure

```text
smart-home-hub/
├── cmd/
│   ├── hub/
│   │   └── main.go
│   └── screen-client/
│       └── main.go
│
├── internal/
│   ├── api/
│   ├── auth/
│   ├── audio/
│   ├── lighting/
│   ├── ambilight/
│   ├── automation/
│   ├── devices/
│   ├── events/
│   ├── mqtt/
│   ├── presence/
│   ├── sensors/
│   ├── storage/
│   ├── system/
│   └── config/
│
├── pkg/
│
├── ml/
│   ├── gesture/
│   ├── habits/
│   ├── presence/
│   └── common/
│
├── frontend/
│
├── telegram/
│
├── migrations/
│
├── configs/
│
├── deploy/
│   ├── docker/
│   ├── systemd/
│   └── scripts/
│
├── hardware/
│   ├── wiring/
│   └── bom/
│
├── tests/
│   ├── integration/
│   ├── e2e/
│   ├── hardware/
│   └── fixtures/
│
├── docs/
│
├── .github/
│   └── workflows/
│
├── docker-compose.yml
├── docker-compose.dev.yml
├── .env.example
├── config.example.yaml
└── README.md
```

---

## 42. Hardware Abstractions

Использовать интерфейсы.

Пример:

```go
type LEDController interface {
    SetPixel(index int, color Color) error
    Show() error
}

type AudioSource interface {
    ReadFrame() ([]byte, error)
}

type Sensor interface {
    Read(ctx context.Context) (Measurement, error)
}

type DeviceAdapter interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Status() DeviceStatus
}
```

Цель:

- mockability;
- simulation mode;
- замена GPIO на ESP32/MQTT без переписывания ядра.

---

## 43. CI/CD

Платформа:

```text
GitHub Actions
```

Deployment:

```text
GitHub Actions
+
self-hosted runner on Raspberry Pi
```

### CI stages

```text
lint
test
security
build
package
```

### Go

Использовать:

```text
gofmt
go vet
golangci-lint
govulncheck
gosec
go test -race ./...
```

### Frontend

Использовать:

```text
ESLint
TypeScript check
unit tests
build
```

### Python

Использовать:

```text
Ruff
mypy
pytest
dependency scan
```

### Containers

Требования:

- multi-stage builds;
- non-root;
- read-only FS, где возможно;
- healthchecks;
- pinned base images;
- multi-arch:

```text
linux/arm64
linux/amd64
```

---

## 44. Deployment

Production deploy запускается только по Git tag.

Пример:

```text
v0.1.0
v0.2.0
v1.0.0
```

Использовать SemVer.

Flow:

```text
tag
  ↓
GitHub Actions
  ↓
lint/test/security/build
  ↓
package images
  ↓
self-hosted runner
  ↓
deploy
  ↓
/ready healthcheck
  ↓
success
```

Если `/ready` не проходит:

```text
automatic rollback
```

на предыдущую рабочую версию.

---

## 45. Testing

Обязательные уровни:

- unit;
- integration;
- e2e;
- hardware-in-the-loop;
- simulation.

### Coverage

Цели:

```text
core business logic >= 80%
overall >= 70%
```

### Go

Table-driven tests обязательны для core logic.

В CI:

```text
go test -race ./...
```

### Python

Использовать:

- pytest;
- deterministic fixtures;
- test images/videos/audio.

### Golden tests

Для:

- Ambilight color mapping;
- FFT;
- beat detection;
- LED effects;
- gesture mapping;
- scene preprocessing.

### Simulation E2E

Сценарий:

```text
fake screen stream
→ screen analyzer
→ Smart Home Core
→ virtual LED strip
→ expected colors
```

### Hardware acceptance

Проверить на реальном железе:

- WS2812B;
- physical button;
- sensors;
- Bluetooth audio;
- camera;
- external monitor;
- PSU;
- level shifter.

### Benchmarks

Для:

- LED render loop;
- Ambilight processing;
- WebSocket;
- event bus;
- FFT;
- ML inference.

---

## 46. Performance Targets

```text
REST p95                < 100 ms
WebSocket event latency < 100 ms
LED rendering           >= 30 FPS
Ambilight               30 FPS
Ambilight e2e latency   < 100 ms
Gesture reaction        < 500 ms
Scene ML                <= 1 inference/sec
Presence detection      every 2–3 sec
```

---

## 47. API/Service Separation

Контейнеры первой версии:

```text
smart-home-core
ml-runtime
frontend
telegram-bot
mqtt-broker
```

SQLite хранится в persistent volume.

Screen client запускается на пользовательском ноутбуке отдельно.

Grafana/Prometheus могут запускаться optional profile.

---

## 48. Home Assistant Compatibility

Архитектура должна позволять подключить Home Assistant без изменения core.

Основные способы:

```text
MQTT
REST API
```

В будущем возможно добавить Home Assistant MQTT Discovery.

---

## 49. Master OFF

При ручном `Lights OFF`:

- выключаются все LED-зоны;
- создаётся manual override;
- Automation Engine не должен немедленно включать свет обратно.

Default override:

```text
15 min
```

---

## 50. Критерий приёмки

Первая версия считается готовой, если выполняется следующий сценарий.

1. Raspberry Pi загружается с Ubuntu Server.
2. systemd автоматически запускает Docker Compose.
3. Пользователь открывает `https://smarthome.local`.
4. Пользователь авторизуется как admin.
5. Dashboard показывает состояние Raspberry Pi и устройств.

### Audio

6. Телефон подключается к `Smart Home Hub` по Bluetooth.
7. Пользователь запускает музыку.
8. Музыка воспроизводится через выбранный audio output.
9. `desk.wall` реагирует на bass/beat/FFT.
10. Web UI показывает realtime spectrum и BPM.

### Ambilight

11. На ноутбуке запускается screen-client.
12. Screen-client проходит pairing.
13. `monitor.backlight` переключается в `AMBILIGHT`.
14. LED по четырём сторонам отображают цвета соответствующих краёв экрана.
15. Ambilight работает около 30 FPS.
16. End-to-end latency меньше 100 ms.
17. Cinema Mode обеспечивает плавную реакцию.
18. ML scene adaptation корректирует параметры Ambilight.
19. При отключении ML обычный Ambilight продолжает работу.
20. При потере screen-client включается fallback preset.

### Gestures

21. Пользователь включает Gesture Mode.
22. `OPEN_PALM` управляет Play/Pause.
23. Swipe переключает треки.
24. `PALM_FACE_DOWN` включает Night Shift.
25. Rotation управляет громкостью.
26. Кадры камеры нигде не сохраняются.
27. LED-индикатор показывает активность камеры.

### Physical Control

28. Single press физической кнопки переключает свет ON/OFF.
29. Double press включает Night Shift.
30. Long press переключает preset.

### Sensors

31. Система получает:
    - освещённость;
    - температуру;
    - влажность;
    - движение;
    - качество воздуха.
32. Датчик освещённости корректирует brightness.
33. Значения доступны Automation Engine.

### Automation

34. Rule Builder позволяет создать правило с AND/OR.
35. Automation Engine выполняет правило.
36. Delayed action работает.
37. Rule priority работает.
38. Manual override временно блокирует автоматизацию.
39. Automation log содержит запись о срабатывании.

### Presence / ML

40. Presence Engine объединяет camera/PIR/Bluetooth/network.
41. Habit Predictor использует только локальные данные.
42. После 30 дней обучения система может предложить действие.
43. Confidence thresholds работают.
44. Пользователь может отклонить recommendation.
45. Пользователь может удалить learning history.

### MQTT / Integrations

46. MQTT-устройство обнаруживается автоматически.
47. Telegram Bot выполняет базовые команды.
48. Smart Home Hub работает без Telegram.
49. Архитектура позволяет подключить Home Assistant.

### Fault Tolerance

50. При падении ML Runtime:
    - lighting работает;
    - audio работает;
    - Ambilight работает без ML;
    - Automation Engine работает без ML.
51. При падении MQTT локальное GPIO-управление продолжает работу.
52. При проблемах SQLite система переходит в Safe Mode.
53. После рестарта LED state восстанавливается.
54. Музыка после рестарта остаётся STOPPED.
55. Camera после рестарта выключена.

### Backup / Deployment

56. Daily backup создаётся автоматически.
57. Backup может быть восстановлен.
58. Factory reset работает.
59. Release tag запускает GitHub Actions.
60. Self-hosted runner разворачивает новую версию.
61. `/ready` проверяет успешность deploy.
62. При неуспешном healthcheck выполняется rollback.

---

## 51. Этапы разработки

### Этап 1. Infrastructure Core

- monorepo;
- Go core;
- SQLite;
- config;
- Docker Compose;
- systemd;
- auth;
- health;
- simulation mode.

### Этап 2. Lighting Engine

- WS2812B driver;
- zones;
- presets;
- Power Limiter;
- physical button;
- LED visualizer.

### Этап 3. Audio

- Bluetooth A2DP;
- Spotify Connect;
- local files;
- output device selection;
- transport controls.

### Этап 4. Music Reactive Lighting

- audio capture;
- FFT;
- beat detection;
- BPM;
- smoothing;
- effects;
- WebSocket visualization.

### Этап 5. Screen Client / Ambilight

- cross-platform Go client;
- pairing;
- screen capture;
- weighted average;
- mapping;
- 30 FPS;
- fallback.

### Этап 6. Scene ML

- multi-frame features;
- multi-label classification;
- ML influence;
- Cinema/Gaming profiles.

### Этап 7. Gesture Control

- camera integration;
- gesture model;
- cooldown;
- activation gesture;
- remapping.

### Этап 8. Sensors / Presence

- ambient light;
- temperature;
- humidity;
- PIR;
- air quality;
- Presence Engine.

### Этап 9. Automation Engine

- rule DSL/model;
- rule builder;
- priority;
- delayed actions;
- schedules;
- manual override;
- logs.

### Этап 10. Habit Prediction

- event collection;
- 90-day retention;
- learning period;
- prediction;
- confidence policy;
- feedback.

### Этап 11. MQTT / Telegram / Home Assistant readiness

- Mosquitto;
- autodiscovery;
- Telegram service;
- external API tokens;
- HA-compatible topics/API.

### Этап 12. Reliability / Observability

- Prometheus;
- Grafana optional profile;
- alerts;
- backup;
- recovery;
- watchdog;
- graceful shutdown.

### Этап 13. CI/CD

- GitHub Actions;
- multi-arch builds;
- security checks;
- self-hosted runner;
- tag-based deploy;
- rollback.

---

## 52. Не входит в MVP

- полноценное распознавание личности;
- обязательная cloud-интеграция;
- облачные ML API;
- AirPlay;
- полноценный multi-user UX;
- Kubernetes;
- несколько различных стандартов LED-протоколов;
- обязательный Home Assistant;
- хранение camera/screen frames;
- история прослушивания музыки.

---

## 53. Потенциальное развитие

В дальнейшем система может быть расширена:

- ESP32-контроллерами;
- Zigbee;
- Thread/Matter;
- умными розетками;
- реле;
- климатическими устройствами;
- несколькими комнатами;
- multi-user profiles;
- face recognition;
- mobile application;
- Home Assistant discovery;
- распределённым k3s-кластером при появлении нескольких вычислительных узлов.

---

## 54. Итог

Smart Home Hub должен представлять собой автономную, расширяемую, local-first edge-платформу, где Raspberry Pi 4 выступает центральным управляющим узлом.

Ключевая архитектурная цель — разделить:

- hardware abstraction;
- control plane;
- realtime events;
- ML/CV;
- automation;
- user interfaces.

Это позволит начать с одной Raspberry Pi и двух LED-зон, а в дальнейшем расширить систему до полноценной многокомнатной платформы умного дома без переписывания основной архитектуры.
