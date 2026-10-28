# ADR-002. Команды, ручной override и конфликты

Дата: 25.10.2026. Статус: **принято как проектное решение; численные limits очередей уточняются при benchmark**.

Источник: §16.2/18/19/27/49 ТЗ. Требования: AUTO-001…AUTO-025, BTN-002…BTN-005, HAB-005, HAB-006, SENS-009, AUTH-011. Вопрос: Q-023, часть Q-022/Q-024/Q-030.

## Контекст

Lights OFF может немедленно отменяться правилом, delayed absence — сработать после возвращения, а две команды громкости — потерять изменения. Один priority недостаточен: manual, safety и continuous live rendering имеют разный смысл.

## Решение

Все mutations проходят [command contract](../commands.md). Core назначает identity/source и сериализует admission. Per-resource revision защищает edits, adapter queue упорядочивает effects. External I/O не держит общий lock.

Порядок: непреодолимые safety/auth/capability ограничения → manual override → music-scene protection → automation priority. Manual intent создаёт field-scoped lease default 900 секунд, Master OFF — `lighting:*` и OFF всех зон. Повторное новое ручное действие продлевает lease, idempotent retry нет. Denied/invalid commands lease не создают; hardware failure после admission оставляет lease, чтобы защитить ещё не выполненное ручное намерение.

Source mapping: UI/button/gesture/linked Telegram/authorized external user controls/local CLI — manual. Rule/habit AUTO/ambient — autonomous. Screen colors/effect ticks — data plane уже выбранного режима: не manual и не новые команды смены mode. При ручной brightness scene/ambient не меняют её до expiry; остальные параметры могут продолжать разрешённое движение. Master OFF отсекает live stream по mode/render epoch.

Правила конфликтуют по затронутым paths. Выше priority выигрывает, равные — стабильный rule_id. Истинный state-guard rule удерживает automated ownership до false/disable; event-only action только в своём conflict batch. Источники и актуальные scopes проверяются до execution, не только при создании job.

Delayed actions сохраняют rule revision/guard/due time. Presence AWAY должна оставаться TRUE в момент effect; UNKNOWN не AWAY. Rule edit/delete/disable отменяет старое work. After restart future jobs восстанавливаются, overdue default EXPIRED без catch-up. Timezone-aware schedules не дублируются по local wall clock key; time jumps проверяются fake clock.

Multi-action имеет all-input validation, но физической атомарности нет. Runtime failure → per-action PARTIAL/FAILED, default stop-on-error; no automatic compensation неизвестных hardware effects. ApplyPreset при override одного entry отклоняется целиком, чтобы сохранить согласованный preset.

## Альтернативы

Global override на любую кнопку — излишне блокирует независимые зоны/audio. Priority поверх manual — нарушает §18.7/49. Exactly-once hardware replay — неподтверждаемая гарантия для NextTrack при crash. Live stream как новая manual command каждого frame — ломает TTL/нагрузку.

## Последствия и проверка

Нужны typed affected paths, saved overrides, dedupe metadata, explicit outcomes. Decision semantics приняты; код/нагрузка/таймеры ещё не проверены. [Сценарии S01–S07/S11–S15](../scenarios.md) задают будущие regression tests; Д032, Д101–Д110, Д119 подтверждают реализацию. Конкретные confidence defaults остаются Q-022, labels/thresholds gestures — Q-018.
