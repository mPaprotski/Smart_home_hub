# ADR-007. Логические зоны и физический LED transport

Дата: 25.10.2026. Статус: **абстракция принята; SPI-chain topology предварительная до Д004/Д008**.

Источник: §4/6/20/42 ТЗ. Требования: LIGHT-001…LIGHT-004, POWER-001…POWER-005, DEV-001/DEV-002. Вопросы: Q-002/Q-005/Q-007.

## Контекст

[День 2](../../DAY_002_RESULT.md) выявил документальный PWM/audio conflict и предложил SPI-chain. Замеров нет; качество DATA cable, privileges, timing и сохранение всех outputs на целевой Pi не проверены.

## Принятая часть

Zone — логический control resource, не GPIO. ZoneBinding связывает её с transport/channel и непересекающимся pixel slice. Desired mode/brightness/preset/override независимы; transport собирает общий frame и применяет zone/global current limits.

Это допускает несколько physical channels, single chain и future external adapter без изменения Lighting business logic. MATRIX/STRIP layout — mapping над slice. Нельзя получить zone independence разветвлением одного DATA на два DIN: нужны уникальные slices/потоки.

## Предварительная часть

Кандидат SPI0 chain: monitor slice затем desk slice. ТЗ явно требует независимые настройки, но не гарантирует физическую отказоустойчивость каждого LED segment. В проектной интерпретации независимость — control state, не separate failure domains. Если нужна вторая, single chain недостаточна и ADR должен быть заменён.

Не объявлять две физические линии обязательными только потому, что план Д002 ранее предполагал такую проверку. На Д004 сравнить реальные варианты из [wiring](../../../hardware/wiring/wiring.md), на Д008 подтвердить выбранный способ на Pi. Наличие поддержки GPIO18 в примере ТЗ не выбирает final pinout.

## Альтернативы и ограничения

Два PWM channels — отдельные DATA, но конфликт resource analog audio согласно исследованию Д002. USB DAC не заменяет requirement встроенного выхода без изменения scope. PCM/внешний controller — alternatives, требуют отдельной проверки/решения. SPI-chain требует monitor DOUT→desk DIN, общих lifecycle и питания, имеет общий failure path.

## Gate окончательного выбора

Реальные lengths/counts, independent mode/brightness demo, 3.5mm/USB/Bluetooth audio coexistence, driver non-root/device access, signal integrity, startup/hard-off без backfeed, >=30FPS joint load. До закрытия этого gate BOM/wiring остаются draft, а physical safety interlock/fuse/cable values — TBD. Новых аппаратных обещаний в этом ADR нет.

## Evidence после локальной части Д004–Д010

[Go slice/limiter probe](../../../experiments/day004/zone_frame_probe.go) и1000 randomized numeric cases прошли, [isolated host monitor PCM](../../research/led-audio-hypotheses.md) работает. Это PASS-SIM/HOST, не WS2812B/audio coexistence на Pi. Пользователь подтвердил отсутствие доступной/настроенной Pi; physical часть остаётся provisional, [M0_NOT_READY](../../milestones/M0_TECHNICAL_BASELINE.md).
