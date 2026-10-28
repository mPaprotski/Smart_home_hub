# ADR-006. Config, persistence и приватные данные

Дата: 25.10.2026. Статус: **модель принята; schemas/backup implementation ещё не выполнены**.

Источник: §17/26–27/31–35/38–40 ТЗ. Требования: STORE-001…STORE-008, HAB-001/HAB-007/HAB-011…HAB-014, REC-001…REC-005, PRIV-001…PRIV-003. Вопросы: Q-021/Q-022/Q-028/Q-030.

## Решение

Каждый configuration key имеет owner: env — secrets/explicit infrastructure overrides, YAML — hardware/infrastructure/safety/defaults, SQLite — validated runtime/user preferences. Whitelist precedence и restart-required status описаны в [data model](../data-model-v0.md); общий «env всегда побеждает всё» не используется.

Only Core repositories пишут SQLite. Runtime получает минимальную local training snapshot/контекст через ограниченный port, не доступ к auth DB. Model artifacts локальные, registry metadata в Core. Camera/screen/PCM/FFT/frame/scene history не сохраняются. Full generic event persistence не вводится.

Habit schema дополняется nullable volume, presence, ambient, monitor signal, origin/quality — проектное уточнение для targets/context §17.2/17.5. История треков не добавляется. Timezone фиксируется с наблюдением для воспроизводимости day features; timestamps UTC. До >=30 дней и достаточной data quality AUTO запрещён; минимальный срок не гарантирует полезность.

Observations 90 дней, automation audit 30 дней, backup последние 7 daily. Model update/invalidation не позволяет бесконечно удерживать старую private history через model artifact. Command dedupe metadata — отдельное проектное окно 24h без secrets/track payload; security audit точные retention числа Q-030.

Delete learning history повышает epoch, отменяет workers/jobs, удаляет DB/history/model/cache/recommendations и Hub-managed backups со старой history, затем создаёт новый clean recovery point. Tombstone epoch вне восстановимой БД защищает от возврата старой history через restore. Внешние экспорты вне контроля Hub не объявляются автоматически удалёнными. Подробная процедура/ошибки — [data model](../data-model-v0.md).

## Альтернативы и последствия

Хранить все events для удобства — privacy violation. Просто удалить observations без model/feedback/caches — неполное удаление. Restore старой БД поверх tombstone без проверки — обход privacy delete. Нужны purpose-specific repositories, purge workflow и maintenance access к backups; это дополнительная работа Д118/Д129–Д130, учтённая в backlog.

## Проверка

Schema/data-flow review, concurrent delete-vs-training fixture, retention boundary tests, DB/disk/network inspection и restore old epoch. Сегодня приняты semantics; exact habit thresholds/ML quality/OS paths и физическое secure erase storage не заявлены решёнными. Factory reset проверяет функциональное удаление управляемых данных, не недоказуемую перезапись всех flash blocks.
