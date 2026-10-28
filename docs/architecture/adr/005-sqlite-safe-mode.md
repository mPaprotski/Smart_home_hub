# ADR-005. Защищённое ручное управление при отказе SQLite

Дата: 25.10.2026. Статус: **принято как проектное решение; Unix socket/container permissions и recovery требуют проверки Д128**.

Источник: §26/27/35/36 ТЗ. Требования: REL-006/REL-007, AUTH-004…AUTH-011. Вопрос: Q-032, часть Q-028/Q-031.

## Контекст

SQLite содержит users/sessions/token hashes, но при её отказе нужны manual lighting и audio. Открыть anonymous LAN API или доверять любому старому token cache — значит потерять auth/revoke semantics.

## Решение

Safe Mode сохраняет Core process/adapters и RAM emergency state. Обычные authenticated REST/WS requests, требующие DB-backed credentials, закрываются с AUTH_BACKEND_UNAVAILABLE; существующие network sessions не считаются действующими по cache. Frontend может показать минимальный maintenance status, но не управлять без auth.

Emergency control — **локальный CLI через Unix domain socket**, не network listener. Socket принадлежит Core service и отдельной maintenance OS group, mode 0660; доступ проверяется peer credentials плюс allowlist разрешённых local UID/group mapping. Наличие OS root/admin доступа уже даёт владельцу возможность восстановления устройства; не создавать hardcoded emergency password.

Allowlist: manual zone state/Master OFF/presets уже известные в RAM, audio controls доступных adapters, минимальный status. При cold boot с недоступной БД доступны hardware defaults/STATIC/OFF, не обещание чтения неизвестных presets из сломанной SQLite. Create user/token/preset/rule, settings persistence, AUTO, training и collectors запрещены. Emergency override/outcomes живут в RAM, DTO явно `persistent=false`.

При DB failure после отправки hardware effect записать outcome в bounded RAM emergency ledger; uncertain effect пометить UNKNOWN_OUTCOME. Не пытаться бесконечно писать повреждённую БД и не использовать YAML как user database.

## Recovery

Переход RECOVERING только explicit local maintenance: остановить новые emergency effects на короткое согласованное окно, проверить/восстановить snapshot, schema/integrity и permissions. Старые auth sessions/tokens повторно проверить по recovered DB; при restore invalidate credentials по maintenance policy.

Emergency ручные состояния не должны теряться с автоматическим возвратом старого preset. Перед re-enable automation оператор выбирает сохранить текущие RAM settings в восстановленную БД либо вернуть backup state; Master OFF hold по умолчанию сохраняется до explicit release. После reconcile refresh snapshots/epochs, затем normal commands и automation. Flapping DB не вызывает самопроизвольных repeated transitions.

## Альтернативы и последствия

Cached session auth — риск stale revoke. Anonymous local subnet — не auth. Остановить весь Hub — нарушает manual availability. Emergency CLI выбран как минимальный local путь без переписывания auth. Телефонный UI в Safe Mode ограничен; ТЗ не требует именно LAN/Web control при DB failure, критерий §50.52 проверяется CLI.

До реализации подтвердить container socket placement/host UID mapping, restrictive ownership, отказ другим OS users и отсутствие TCP exposure. Не mount всего `/var/run` в bot/frontend. Проверки: missing/corrupt/read-only/disk-full SQLite, permitted/denied local peer, cold boot defaults, no persistent writes и controlled recovery на Д128/Д155.
