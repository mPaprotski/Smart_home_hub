# Dispatcher и permissions Д017

`internal/commands` — trusted internal API. С Д019 смонтирован HTTPS timezone gateway.
Principal/source/scopes/bindings передаются composition/gateway после authentication;
это не формат JSON body браузера. Session/token verifier и admission recheck реализованы на
Д019–Д020: [policy](sessions-tokens.md). Произвольный Go-код внутри Core не считается враждебным sandboxed клиентом.

Реально поддержана **SetUserSetting(timezone)**: user ownership, known scope,
validated value/expected revision/deadline, persistent idempotency и atomic outcome.
Остальные domain commands возвращают NOT_IMPLEMENTED до mutation. Devices/Lighting
Engine/ApplyPreset/automation scheduling этим dispatcher срезом не объявляются готовыми.

| Principal | Операция / resource | Проверка |
|---|---|---|
| Enabled user с settings:write в текущих DB roles | Свой timezone | Разрешена при valid CAS/deadline |
| Тот же user | Чужой user setting | FORBIDDEN, даже если роль admin |
| Viewer или user без scope/disabled | Mutation | FORBIDDEN |
| Trusted service fixture | Setting своего owner | scope settings:write + exact user binding + актуальные права owner |
| Trusted service fixture | Чужой binding/неизвестный scope | FORBIDDEN |
| Неизвестный principal kind/source/type | Любая mutation | Отказ, handler не вызван |
| Anonymous HTTP/поддельный header actor | Mutation route | Route отсутствует,404 |

Principal.ID различает user/service identities; UserID у service — owner, не dedupe ID.
User principal.ID должен совпадать с UserID. Реальные service credentials/revoke list
ещё не выпускаются; fixtures не являются network token bypass.

## Admission

Input/version/known operation/value/deadline → owner/binding → актуальные user.enabled/
roles внутри transaction → dedupe → CAS update + command record commit. Нет slow
adapter I/O под SQL lock. Supported setting handler синхронный; отдельной очереди,
pending worker или физических effects нет.

Dedupe key `(principal_id, operation, command_id)`, hash нормализованных semantic
полей; retry после revocation запрещён до чтения cached outcome. Другой payload с тем
же ID конфликтует. Concurrent дубликаты дают одну revision, concurrent CAS requests
не теряют update. В operation result APPLIED означает применение setting в SQLite,
не состояние LED/аудио. Ошибка commit возвращает error и пустой outcome, не APPLIED.

Resolved terminal metadata older24h удаляется при следующем admission; pending
statuses не удаляются. Periodic cleanup/durable scheduler/outcome reconciliation —
будущие модули. Нет автоматического replay неизвестного физического эффекта.
Event Bus notification не является источником истины command outcome; для текущего
setting slice отдельный command-event publisher ещё не подключён.

## Lighting arbitration hook

`AdmitLighting` — transaction policy port для будущих domain handlers. Получает
server-derived affected paths, trusted source и injected clock. Manual sources
создают lease900s; retry того же command ID lease не продлевает. Autonomous source
не может перекрыть пересекающийся lease; независимый path разрешён. Wildcard scope
поддерживает `lighting:*` и prefix `:*`. Новый manual intent может обновить lease.

Это policy fixture и persistence hook, а не команда реального Master OFF: safety,
capability, permission checks должны пройти **до** него; будущий caller оборачивает
admission единым transaction. Full scene protection/render epoch, physical state,
priority conflict batches, multi-action partial effects и delayed jobs — последующие
Lighting/Automation этапы. Unsupported batch сейчас отклоняется целиком.

Доказательства: две реальные user rows, revoked role/user, service identity/binding,
changed-payload retry, concurrent CAS/duplicate, rollback command-record failure,
fake-clock lease overlap/expiry/retry. [Отчёт](../DAY_017_RESULT.md).
