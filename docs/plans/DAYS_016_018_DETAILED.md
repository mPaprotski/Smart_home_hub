# Подробный рабочий план: Д016–Д018

**Hardware baseline от26.10.2026: только Raspberry Pi 4B с2GB RAM.** [Ресурсный профиль](../development/pi4b-2gb.md). Pi совместные memory/performance gates ещё не выполнены.

Подготовлен **26 октября 2026 года, Europe/Minsk**. Основание: [ТЗ](../../Smart_Home_Hub_TZ.md), [общий план](../DEVELOPMENT_PLAN_DAILY.md), [результат Д011–Д015](../DAYS_011_015_RESULT.md), [модель данных](../architecture/data-model-v0.md), [команды](../architecture/commands.md), [ADR-002](../architecture/adr/002-command-arbitration.md), [ADR-006](../architecture/adr/006-config-data-ownership.md).

**Программный срез Д016–Д018 реализован:** [фактический отчёт](../DAYS_016_018_RESULT.md). Ниже сохранён рабочий план; отчёт отдельно отмечает supported command subset, будущие physical/queued handlers, sessions и внешние gates. M0/M1/remote CI не закрыты.

## 1. Результат трёх дней

| День | Основная задача | Что должно работать к вечеру |
|---|---|---|
| 16 | SQLite WAL, repositories, migrations | Пользовательские настройки переживают настоящий restart; ошибка миграции не оставляет частично применённый шаг |
| 17 | Command dispatcher, identity, scopes, ownership | Один проверенный путь изменения состояния; недостаточные права, конфликт revisions и неизвестные команды не вызывают handler |
| 18 | Локальный admin bootstrap, Argon2id, credential verification | Первый admin создаётся один раз без default password; пароль хранится только как salted hash; перебор ограничен |

План рассчитан на одного разработчика: каждый день **360мин предметной работы +120мин сопровождения =480мин**. Сопровождение:20мин вход и разбор блокеров,40мин диагностика/tooling,40мин отчёт и evidence,20мин review/передача. Это оценка, а не обещание исправить любую несовместимость за один день. При блокере сохранять фактический незавершённый scope.

Д019 завершает sessions/logout/revocation/cookies/CSRF, Д020 — machine tokens и полноценный security audit. До них не открывать управляющие REST/WS routes и production/LAN режим. Pi для этих трёх дней не требуется; физический M0 остаётся NOT_READY.

## 2. От чего начинаем

Уже существуют `internal/config`, `app`, `api`, `events`, `devices`, `storage`, `simulation`, root Go module, Make/CI, React shell и Python contracts. Проверены23 Go tests с race detector,14 Python cases, native/Compose smoke и чистый source quickstart; это исходная база, не будущий гарантированный test count.

Важные ограничения текущего кода:

1. `internal/storage.Memory` — bounded latest-observation cache. Его interface `Observations` не предназначен для users/settings/sessions. **Оставить его в RAM**, добавить отдельные persistent repositories. Не записывать sensor/audio/screen observations в SQLite только потому, что появился SQL driver.
2. `config.Storage` сейчас строка `memory_fake/not_implemented`; добавить `sqlite` и отдельные typed SQLite options с strict YAML/env validation. Не менять все старые simulation fixtures на реальную БД неявно.
3. `system.Health` и OpenAPI пока описывают четыре dependencies и persistent=false. SQLite wiring требует пересмотра readiness/schema/tests; нельзя оставить hardcoded metadata после подключения persistence.
4. В dev Docker rootfs read-only, UID10001, data volume отсутствует. Для SQLite нужен отдельный writable data directory с правильным owner, включая WAL/SHM sidecars. Read-only остального rootfs сохранить.
5. Нативный startup сейчас observer→fake suite→HTTP. Добавить storage/migrations до repositories и HTTP, а shutdown закрывает storage после пользователей БД.
6. HTTP до сих пор read-only, production/hardware отклоняются. Д017 проверяет commands внутри Core и tests, а не добавляет anonymous mutation API.
7. Старые отчёты/результаты не перезаписывать. Новые evidence складывать отдельно, реальные DB/password/hash/session artifacts в публичный отчёт не включать.

---

## Д016. SQLite WAL, миграции и первая persistence

### Цель

Получить настоящий file-backed Storage с контролируемым lifecycle, версионируемой схемой и integration tests. Показательная операция — сохранить user setting/desired zone fixture, остановить процесс, запустить заново и прочитать то же значение. Это проверка persisted settings; доставка physical LED frame после boot относится к Lighting Engine.

### Блок16.1. Driver и контракт подключения —60мин

1. Сверить Go1.27.3, native Linux/amd64 и текущий Docker `CGO_ENABLED=0`. Выбрать maintained SQLite driver, совместимый с этим build, либо обоснованно изменить build chain. Pure-Go driver — первый кандидат, не автоматически принятый dependency.
2. Проверить upstream/support/license, exact module version, transitive dependencies и **реальную embedded SQLite version** через `SELECT sqlite_version()`. Host `sqlite3 --version` не показывает версию библиотеки внутри Go driver.
3. Проверить актуальные upstream advisories. В частности, SQLite описывает WAL-reset fix в3.51.3 и отдельных backports; выбрать поддерживаемую исправленную сборку, а не старый пакет только ради успешной компиляции. Источник:[SQLite WAL](https://sqlite.org/wal.html#walresetbug).
4. Зафиксировать выбор/версии в `docs/development/storage.md`, dependency versions — в go.mod/go.sum. Не переписывать CI Docker на CGO случайно.
5. Спроектировать typed options: DB path, busy timeout, startup/migration timeout, bounded pool, synchronous/checkpoint policy. DB path — инфраструктура; runtime user preferences остаются в repositories.
6. Первоначально один writer connection; при separate reader pool задать явные limits и применить connection-local PRAGMA на **каждом** connection. Тест конкурентной блокировки всё равно должен использовать два настоящих connections.

**Выход:** driver выбран и реально открывает test DB; настройки и ограничения зафиксированы.

### Блок16.2. Schema v1 и migrations —90мин

1. Создать versioned embedded SQL migrations и `schema_migrations(version, checksum, applied_at)`. Каждую applied migration считать неизменяемой.
2. В минимальную v1 включить users, roles, user_roles, role_permissions, sessions schema, devices, zones, user_settings. Детали zone_bindings/transports добавить только в объёме необходимого repository slice; остальную логическую модель не превращать сразу в десятки пустых таблиц.
3. Для users задать stable ID, normalized unique username, password_hash, enabled, timezone, created_at. В пустой БД пользователей нет; запрещён seed `admin/admin` или hash общего пароля.
4. Roles/scopes валидировать через allowlist; `user_roles` — composite uniqueness и foreign keys. Одного admin bootstrap достаточно для начала; multi-user UI не требуется.
5. Sessions schema подготовить к hash-only opaque identifiers, UTC expiry/revocation/user FK. На Д016 строки sessions не создаются, никакие cookies не выпускаются.
6. Settings: user+key uniqueness, schema version, validated typed value, revision. Zones: desired fields и revision; live RGB frames/status/evidence не сохранять.
7. Определить NOT NULL/UNIQUE/CHECK/FK/delete policy. Не использовать каскадное удаление shared hardware как побочный эффект удаления user.
8. Migration runner: exclusive startup admission, проверка checksums/schema version, transaction на каждый поддерживаемый migration step. Schema update и запись его version должны commit вместе.
9. Ошибка шага откатывает **этот шаг**; ранее успешно применённые версии остаются. Не обещать откат всей цепочки migrations и не создавать фиктивную «чистую БД» поверх повреждённой.
10. Повторный запуск no-op для актуальной схемы. DB с более новой схемой — понятный отказ, без auto downgrade. Для старой схемы использовать заранее созданный fixture v1→v2, не произвольный неизвестный SQLite файл.

**Выход:** новая БД и controlled upgrade работают; полусхема и checksum mismatch отвергаются.

### Блок16.3. SQLite lifecycle и repositories —90мин

1. Открывать только explicit local path; приватная data directory, owner permissions и безопасные новые file modes. Не исправлять существующие чужие каталоги рекурсивным chmod.
2. Включить WAL и проверить returned mode; включить foreign_keys и проверить readback. Опечатка PRAGMA не должна остаться незамеченной. Правила connection settings сверять с [SQLite PRAGMA](https://sqlite.org/pragma.html).
3. Выбрать synchronous policy с объяснением durability trade-off; для persisted settings/auth начать с FULL и измерить нагрузку. Не считать быстрый host test доказательством устойчивости SD-карты к потере питания.
4. Задать bounded busy timeout, context deadlines и классификацию BUSY/LOCKED/constraint/read-only/corruption. Не retry бесконечно и не повторять non-idempotent операции вслепую.
5. Ввести узкие User/Role/Settings/Zone repositories и transaction/unit-of-work boundary. SQL только с parameter binding, без конкатенации пользовательских values/identifiers.
6. Implement get/set одного validated setting, optimistic update `WHERE revision=expected`, короткие transactions. Нельзя держать SQL transaction во время Argon2, hardware I/O или ожидания Event Bus consumer.
7. Закрывать rows/statements/transactions во всех error paths; Close учитывает app shutdown. Migration failure не запускает dependent components.
8. Observation cache/frames оставить отдельно. Не добавлять `events(payload)` и универсальную сериализацию bus в SQL.

**Выход:** setting CRUD/CAS имеет реальные SQLite integration tests и безопасные ошибки.

### Блок16.4. Wiring, readiness, Docker и restart —60мин

1. Сохранить отдельные modes: memory_fake, not_implemented, sqlite. Никакого fallback sqlite→memory при open/migration error.
2. В sqlite profile `storage.kind=sqlite`, persistent=true только для успешно подключённой file-backed БД. При fake adapters simulated=true сохраняется: persistence не превращает fake hardware в реальное.
3. Ready только после migrations и доступности required repository operation; подробности hash/path/SQL не выдавать наружу. Определить startup failure policy: exit с безопасной ошибкой допустим, anonymous degraded startup — нет.
4. В version/ready schemas заменить hardcoded fields, обновить существующие tests/fixtures. `memory_fake` обязан остаться persistent=false, not_implemented — ready503.
5. Создать отдельный persistent dev Compose profile/override, монтирующий **каталог** data, не один DB file. Проверить UID10001 и запись WAL/SHM; не давать Python/bot/frontend mount БД.
6. Smoke: clean volume→migrate→write synthetic setting→stop→start→read. Использовать local repository/maintenance harness, не временный anonymous HTTP setter.
7. Раздельно проверить clean stop и внезапное завершение process с committed/uncommitted transaction. SIGKILL process recovery — не физическое отключение питания Pi.
8. Cleanup удаляет только test-owned temporary DB/volume; development DB пользователя не трогать.

### Блок16.5. Ошибки, recovery contract и отчёт —60мин

1. Lock contention воспроизвести отдельным connection, проверить deadline и освобождение после rollback.
2. Inject failure в середине migration: ранее созданные объекты этого шага и schema version не должны сохраниться.
3. Проверить недоступный каталог/read-only/невалидную DB/новую schema/checksum mismatch. Disk-full — через bounded fixture/VFS fault injection, либо явно BLOCKED, не заполнять диск host.
4. Проверить FK на каждом pool connection, revision conflict и rollback multi-write operation.
5. Для этого этапа описать backup boundary: consistent snapshot через поддержанный backup API либо остановленную БД с проверенным завершением/checkpoint. Не `cp hub.db` во время записи WAL. Автоматический daily backup/restore/privacy-purge остаётся последующим этапом. Источник:[SQLite Backup](https://sqlite.org/backup.html).
6. Записать driver/SQLite versions, PRAGMA readback, scenarios, фактические результаты в `docs/DAY_016_RESULT.md`. Только synthetic DB в tests; реальные файлы БД и auth hashes не публиковать.

### Проверки Д016

| ID | Сценарий | Условие завершения |
|---|---|---|
| D16-01 | Новый file-backed DB | Schema актуальна, WAL/FK подтверждены реальными connections |
| D16-02 | Upgrade/repeat/checksum/newer schema | Upgrade корректен, repeat no-op, incompatible cases fail closed |
| D16-03 | Failed migration | Шаг и version откатились вместе; прежняя схема читается |
| D16-04 | Persistence | Значение/revision пережили настоящий restart процесса |
| D16-05 | BUSY/cancel | Два connections, bounded timeout/cancel, после release запись возможна |
| D16-06 | Constraints/CAS | FK/unique/revision/transaction rollback реально исполняются |
| D16-07 | Startup failure | Нет fallback fake, ложного ready или listener с неготовым required storage |
| D16-08 | Persistent Compose | После stop/start данные сохранены; owner/WAL/SHM/rootfs policy проверены |
| D16-09 | Privacy | Observation/media histories и credentials отсутствуют в generic SQL/log evidence |

**Артефакты:** storage implementation/repositories/migrations/tests, typed config, persistent dev Compose profile, `docs/development/storage.md`, `docs/DAY_016_RESULT.md`, отдельный evidence directory.

---

## Д017. Единый Command Dispatcher и permissions

### Цель

Реализовать внутреннюю границу команд, через которую позднее пойдут UI/CLI/button/Telegram/automation/ML. Этот день не создаёт реальные Telegram/ML transports и полноценный Lighting Engine.

### Блок17.1. Identity и command types —60мин

1. Разделить внешний business request и внутренний command envelope. Principal/source/scopes/priority/bindings назначаются trusted gateway, не JSON body.
2. Ввести User/Service/System principal types с stable refs, source enum, typed target и affected paths. Временные test principals существуют только в dependency injection; не добавить HTTP header «считать admin».
3. Создать explicit catalog/registry известных commands и typed validators. Unknown command/schema/scope/capability — отказ.
4. Envelope: command ID/version/type/principal/source/target/typed payload/server timestamps/deadline/correlation/expected revision. Ссылаться на архитектурный контракт, не вводить второй несовместимый формат.
5. Задать safe error/result model: denied/invalid не ACCEPTED; adapter failure не APPLIED. Не выдавать пустой handler `return nil` за поддержку operation.

### Блок17.2. Permission matrix и ownership —75мин

1. Описать allowlist scopes и permission checks для выбранных commands; зарегистрированный handler и доступная capability нужны даже admin.
2. Загружать актуальные user.enabled/roles и resource ownership из repositories, не доверять caller-supplied копии старых прав.
3. Проверить user-owned setting/preset/rule boundaries; shared/system zone policy описать отдельно. Owner nullable не означает anonymous write.
4. Device-bound principal ограничен назначенным device/zone; один scope не даёт доступ ко всем targets. Real token issuance/verifier — Д020, здесь trusted fixture principal и общий checker.
5. Source не даёт привилегии сам по себе. PREDICTION/AUTOMATION не могут представиться manual UI/SYSTEM.
6. Для batch проверить права и inputs всех entries до первой mutation. Неразрешённый entry не приводит к частично применённому preset.
7. При denied operation не выдавать private resource details; определить consistent forbidden/not-found policy для будущего HTTP gateway.

**Артефакт:** `docs/security/permission-matrix.md`, table-driven cases с двумя пользователями/несколькими resources.

### Блок17.3. Admission и arbitration boundary —90мин

1. Pipeline: validate→resolve identity/resources→permission/capability→revision/deadline→arbitration→transactional admission→handler→outcome/event.
2. Первый end-to-end persistent handler — изменение разрешённого user setting, например timezone с настоящей валидацией. Он доказывает ownership/revision/transaction; не требует hardware.
3. Lighting command pipeline проверить через явно synthetic handler/LED port. Реализованный subset перечислить; остальные registered contracts возвращают NOT_IMPLEMENTED без ACCEPTED/effect.
4. Affected paths вычисляет сервер по typed intent. Arbitrary field-path patch, shell command или SQL selector не допускаются.
5. Arbitration port получает validated principal/source/paths и injected clock. Зафиксировать порядок safety/auth/capability→manual lease→domain guards→automation priority.
6. Для lighting fixture проверить default manual lease900s и пересечение paths; invalid/denied request lease не создаёт. Полный Master OFF/render epoch/scene protection/effects — Д032+, не обещать их готовность от одного hook.
7. Manual retry одного command ID не продлевает lease. Новый допустимый manual intent может продлить его; autonomous intent не может снять lease собственным priority.
8. Per-resource ordering/CAS защищают concurrent mutation. Медленный adapter не держит global dispatcher/DB lock.

### Блок17.4. Idempotency, commit и outcomes —75мин

1. Добавить migration для command_records и минимального override persistence, если они реально нужны выбранному slice. Не создавать SQLite command store, который пишет arbitrary raw payload/secrets.
2. Dedupe key scope: principal+operation+command_id; normalized semantic payload hash. Тот же ID/тот же request возвращает прежний outcome после **повторной** проверки действующих прав.
3. Тот же ID с другим payload — conflict; второй handler не запускается. Дедупликация concurrent одинаковых requests должна быть атомарной.
4. Persistent state/revision/command admission commit согласованно; Event Bus публикует факт только после commit. Не держать transaction при внешнем effect.
5. Handler failure хранит честный outcome. Crash между effect и outcome для non-idempotent operation — UNKNOWN_OUTCOME, не automatic replay.
6. Если очередь нужна сейчас: capacity/deadline/cancel/overload error задать явно. Если исполнение синхронное, не обещать future durable scheduler и recovery delayed jobs.
7. Pending execution повторно проверяет permission/revocation/expiry. Реальные session/token revocations появятся Д019–Д020; тест использует mutable credential/authorization provider.
8. Определить bounded retention command metadata, pending не удалять как completed. Cleanup scheduling/full audit — последующие этапы.

### Блок17.5. Integration и отчёт —60мин

1. Тест с двумя пользователями и persisted settings: разрешённая mutation проходит, чужая — нет, DB/revision/handler counter после отказа не изменились.
2. Client metadata forgery, неизвестный command/scope/capability, expired request и stale revision не вызывают handler.
3. Две concurrent команды с одной expected revision: только одна выигрывает, вторая получает conflict, update не теряется.
4. Retry/concurrent duplicate/different-payload retry проверять по реальному handler call count и persisted records.
5. Fake clock проверяет lease expiry/overlap и retry TTL; реальные sleeps на15мин не нужны.
6. Adapter fail/commit fail/event subscriber overflow не превращаются в ложный success. Состояние операции доступно независимо от Event Bus notification.
7. HTTP mutation routes пока не монтировать; internal test invocation не делает anonymous local CLI универсальным privileged gateway.
8. Обновить command contract, permission matrix и `docs/DAY_017_RESULT.md`: перечислить реально поддержанный subset и deferred modules.

### Проверки Д017

| ID | Сценарий | Условие завершения |
|---|---|---|
| D17-01 | Trusted metadata | Подмена source/principal/scopes извне не меняет effective identity |
| D17-02 | Ownership/scopes | UserA не меняет UserB; binding исключает чужую зону |
| D17-03 | Catalog/capability | Unknown/unsupported operation отклоняется до effects |
| D17-04 | Validation/revision/deadline | Invalid/expired/stale request не меняет DB/lease |
| D17-05 | Concurrent ordering | CAS предотвращает lost update, медленный fake не блокирует независимый resource |
| D17-06 | Idempotency | Retry не повторяет handler/lease; changed payload конфликтует |
| D17-07 | Arbitration subset | Manual lease/affected paths/fake-clock expiry соблюдены |
| D17-08 | Transaction/outcome | Commit failure не вызывает effect; unknown physical outcome не replay |
| D17-09 | Reauthorization/batch | Смена прав отменяет pending execution; batch prevalidation без частичной mutation |
| D17-10 | Exposure | Anonymous write routes/privileged spoof headers отсутствуют |

**Артефакты:** `internal/commands`/authorization ports, handlers/repositories/migrations/tests, permission matrix, updated contract, `docs/DAY_017_RESULT.md`.

---

## Д018. Admin bootstrap, Argon2id и проверка credentials

### Цель

Создать первого администратора безопасным локальным способом и проверить пароль через AuthService. Подготовить login transport и ограничения попыток без выдачи неготовых sessions. Реальные user passwords не использовать в публичных smoke/CI fixtures.

### Блок18.1. Password policy и hasher —75мин

1. Выбрать maintained Argon2id implementation; для Go проверить [golang.org/x/crypto/argon2](https://pkg.go.dev/golang.org/x/crypto/argon2), закрепить совместимую версию в go.mod/go.sum.
2. Определить username normalization и уникальность до bootstrap; password не trim/lowercase и не нормализовать скрыто. Поддержать password manager/paste/Unicode.
3. Проектный начальный policy: минимум15 Unicode characters без обязательной смеси символов; maximum1024 UTF-8 bytes, отдельный bounded HTTP body. Проверить неоднозначности counting и зафиксировать их в contract; значения — проектные defaults.
4. Начальная Argon2id cost policy: m=19456KiB, t=2, p=1; random salt16bytes, output32bytes. Это базовая выбранная конфигурация по [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html), не результат benchmark Pi.
5. Хранить encoded hash с algorithm/version/m/t/p/salt/hash, например PHC-format. Salt через crypto/rand, одинаковые пароли дают разные hashes.
6. Verify парсит формат строго, ограничивает размеры/params **до** выделения памяти и дорогого вычисления; malformed/unsupported hash даёт безопасный отказ. Не позволять повреждённой БД запросить гигабайты памяти.
7. Digest сравнивать constant-time; определить needs-rehash policy. Дорогой hash выполнять вне SQL transaction и без общей блокировки Core.
8. Ограничить concurrent hash jobs отдельным semaphore; initial cap1 для128MiB dev container. Отмена запроса не должна снимать admission slot до завершения реально работающего Argon2 job. No unbounded queue/goroutine per login.

### Блок18.2. Одноразовый локальный bootstrap —75мин

1. Добавить отдельный explicit CLI workflow, например `hub admin bootstrap --config … --username …`; это будущий интерфейс, команда ещё не реализована.
2. На интерактивном запуске читать password с TTY без echo и confirmation. Не принимать password как CLI flag, URL query или фиксированное env значение.
3. Для CI предусмотреть explicit stdin/pipe mode только с synthetic secret, без echo и без command-history literal. При отсутствии TTY/явного stdin режима отказать, не висеть в ожидании ввода.
4. Bootstrap открывает ту же file-backed SQLite через Storage/migrations; memory_fake bootstrap не должен сообщать о создании постоянного admin.
5. Проверить initialized marker внутри transaction, атомарно создать user+admin role binding+bootstrap marker. Защитить unique constraints; два concurrent bootstrap процесса должны дать ровно одного первого администратора.
6. Не ограничиваться `COUNT(users)=0`: после bootstrap нельзя автоматически разблокировать процедуру удалением/disable user. Recovery/reset первого admin — отдельный явный будущий workflow.
7. На отказе/ошибке rollback без half-created user/role/marker. Пароль в памяти не класть в command_records/audit/errors; immutable strings и GC не дают обещания гарантированного secure erase.
8. Вывод — safe ID/результат, без password/hash/salt. Повторный bootstrap — bounded error и ненулевой exit code, без замены существующего hash.

### Блок18.3. AuthService и stage login API —75мин

1. Реализовать Authenticate(username,password) через repository+hasher+attempt limiter. Правильный пароль возвращает trusted principal внутри Core, не копию caller metadata.
2. Неизвестный username, disabled user и неверный пароль получают один public error `INVALID_CREDENTIALS`, без разглашения существования account. Нормализованный dummy-hash path уменьшает timing discrepancy; не обещать абсолютно одинаковое время всей системы. Основание:[OWASP Authentication](https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html).
3. Проверить input size/content-type/strict body fields до дорогого hash. Пароль никогда не появляется в request logger/trace/metric label.
4. Добавить contract `POST /api/v1/auth/login`:400 malformed,401 invalid credentials,429 limit,503 storage/session service unavailable; request ID безопасный.
5. **Граница Д018:** valid credentials подтверждены service tests, но session issuer из Д019 ещё отсутствует. HTTP для valid credentials возвращает503 `AUTH_SESSION_NOT_IMPLEMENTED`, без Set-Cookie/access token и без сообщения «вход выполнен». Production/LAN/write API остаются закрытыми.
6. Позитивный HTTP login200 и защищённый `/me` включить на Д019 одновременно с настоящими sessions/cookie/CSRF/revocation. Не выдавать временный bearer token или signed user ID ради demo.
7. Открытый local HTTP endpoint проверять только synthetic credentials; реальную password ceremony проводить через local TTY. HTTPS/proxy trust integration — отдельная готовность, не подразумевается от localhost.
8. OpenAPI/health/version обновить честно: credential verifier available не означает authenticated service ready. При required auth до Д019 ready503; explicit development core-only readiness может быть отдельно допустимым профилем.

### Блок18.4. Limits, failure paths и benchmark —75мин

1. Проектные начальные limits:5 failed attempts/15мин на normalized account key,20 requests/мин на source IP, global hash concurrency1. Проверить как defaults, не называть отраслевыми нормативами.
2. Ограничить limiter key count/TTL/eviction и общий admission. Account key не включать в метрики/публичный лог; хранить в RAM безопасный keyed digest, без plaintext username-history.
3. Брать client IP из actual peer; forwarded headers доверять только явно настроенному reverse proxy. Header spoof не должен сбрасывать rate limit.
4. Не sleep внутри handler на15мин;429 с bounded Retry-After/cooldown, injected clock в tests. Account lockout ограничен по времени, без permanent lock от attacker.
5. Ошибки после рестарта и распределённый перебор учитывать в policy: минимально зафиксировать RAM limiter reset, а persisted account cooldown/counters — хранить атомарно, если заявляется защита через restart. Не обещать такой защиты только RAM тестами.
6. Synthetic benchmark: hash/verify latency, peak/concurrent memory, response на saturation, сохранение отзывчивости `/health`/Core observer при login flood. Оценить128MiB container budget; не уменьшать cost ниже выбранного baseline ради зелёного smoke.
7. На host измерить фактически, на Pi performance gate оставить BLOCKED. Ожидаемое число ms заранее не выдумывать.
8. Проверить DB unavailable, malformed hash, user disabled, concurrent bootstrap/login, random source error через injection. External response sanitization без SQL/PHC/password contents.

### Блок18.5. Сквозная проверка и отчёт —60мин

1. Новый private temp DB→bootstrap synthetic admin→verify valid/invalid→process restart→verify снова.
2. Повторный/concurrent bootstrap запрещён; прежний hash и marker сохранены. Проверить одинаковый password с разными salts на независимых fixtures.
3. В stage HTTP valid credentials дают честный503 без session; invalid401; throttle429; anonymous control routes остаются недоступны.
4. Пароль-sentinel отсутствует в DB logical values вне salt/hash format, stdout/stderr/error/audit/public artifacts. Hash не публиковать в отчёте; test-owned DB уничтожить.
5. Проверить parameter bounds PHC, long/Unicode passwords, malformed payload/content-type и no fake session fallback.
6. Прогнать затронутые Go unit/integration/race tests, existing Python/frontend checks/build, OpenAPI validation, native и persistent Compose smoke. Не расширять frontend до login page до рабочего session backend.
7. Записать выбранные policies, benchmark scope, unresolved Pi gate и honest session status в `docs/DAY_018_RESULT.md`; передать Д019 готовый AuthService и issuer port.

### Проверки Д018

| ID | Сценарий | Условие завершения |
|---|---|---|
| D18-01 | Hash/verify | Разные random salts, valid/invalid comparison, plaintext не сохраняется |
| D18-02 | Hash/input bounds | Malformed/oversized/опасные PHC params отклоняются до большой allocation |
| D18-03 | Bootstrap | Ровно один admin+marker, нет default password, повтор/concurrent запуск не заменяет hash |
| D18-04 | Credentials | Unknown/disabled/wrong password дают один public error без account enumeration details |
| D18-05 | Throttling | Account/IP/global bounds, fake-clock expiry, spoofed proxy headers не обходят limits |
| D18-06 | Runtime budget | Измерены host/container hash latency/memory и Core responsiveness; Pi отдельно |
| D18-07 | Stage API | Correct password не создаёт fake session;503 до issuer Д019,401/429 по контракту |
| D18-08 | Restart/rollback | Admin переживает restart; failed bootstrap не оставляет partial account |
| D18-09 | Secret handling | Sentinel/PHC не попали в logs/evidence/command records; DB остаётся private |
| D18-10 | Exposure/readiness | Required auth не READY без sessions; production/LAN mutations не открыты |

**Артефакты:** `internal/auth`, password/limiter policy, bootstrap CLI, stage login contract/tests, OpenAPI, `docs/security/admin-bootstrap.md`, `docs/DAY_018_RESULT.md` и synthetic benchmark summaries.

---

## 3. Проверки по завершении каждого дня

- Сначала targeted unit/integration tests изменённого кода и negative cases; затем существующий `make check` с актуальным scope. На этой машине: `make GO=/snap/go/current/bin/go check`.
- Обновить smoke runners для нового profile, сохранив прежние memory_fake/not_implemented semantics. Новые команды Make добавлять только вместе с настоящей реализацией.
- SQLite tests — настоящие temporary files/connections, не только mock repository и не только `:memory:`. Auth/permission tests — synthetic credentials и injected clock, без Pi.
- Проверять fresh DB и upgrade fixture, clean source install/build. Для Docker различать syntax check и actual build/start/restart/cleanup.
- Evidence: `PASS-HOST`, `PASS-SIM`, `PASS-COMPOSE` только по исполненному сценарию; `PASS-CI` требует remote run/commit. Недоступные platform/fault injection checks — BLOCKED с причиной.
- ТЗ/старые reports/results сохраняются; global acceptance, M0 и M1 не закрываются от наличия новых unit tests.

## 4. Ожидаемая структура новых артефактов

```text
internal/storage/              # Existing RAM observations + separate SQL repositories
  migrations/                  # Embedded versioned SQL
internal/commands/             # Envelope, admission, registry, outcomes
internal/auth/                 # Principal/permission ports, passwords, bootstrap, limiter
cmd/hub/                       # Storage wiring + explicit local bootstrap subcommand
scripts/                       # Persistent/command/auth synthetic smoke
configs/                       # Explicit sqlite development example
deploy/docker/                # Persistent dev override/profile
docs/development/storage.md
docs/security/permission-matrix.md
docs/security/admin-bootstrap.md
docs/DAY_016_RESULT.md
docs/DAY_017_RESULT.md
docs/DAY_018_RESULT.md
docs/results/days016-018/      # Redacted summaries, no DB/credentials
```

Имена — предлагаемые; сначала проверить реальные packages и не создавать дублирующие public APIs. Config/health/API/Make/CI/ignore rules обновляются вместе с реализацией. DB directories/sidecars/real hash fixtures не попадают в build context/public artifacts.

## 5. Что передаём в дни19–20

| К Д018 должно быть | Д019–Д020 ещё должны сделать |
|---|---|
| SQLite users/roles/session schema и миграции | Реальное хранение sessions, expiry, logout, revoke-all, cleanup |
| Проверенный AuthService + Argon2id | Session issuer/middleware, cookie flags, CSRF/origin policy, защищённый `/me` |
| Command identity/permission boundary | Подключение authenticated HTTP/WS/device gateways и credential revocation |
| Hash-only token storage contract | Machine token create/list/revoke/scopes/bindings, secret показ один раз |
| Минимальные safe error/log summaries | Полный durable security audit и retention policy |
| Host/container measurements | Pi Argon2/storage performance и физический M0 |

**Условие завершения трёхдневного среза:** persisted setting/admin переживают restart; unauthorized command не меняет состояние; bootstrap одноразовый; password verifier и limits проверены; отсутствие sessions отражено явным отказом и не открывает управление. Это software gate перед Д019, не production acceptance.
