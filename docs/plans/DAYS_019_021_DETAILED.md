# Подробный рабочий план: Д019–Д021

Дата:26.10.2026, Europe/Minsk. Основание: [общий план](../DEVELOPMENT_PLAN_DAILY.md),
[результат Д016–Д018](../DAYS_016_018_RESULT.md), [ТЗv1.1](../../Smart_Home_Hub_TZ.md),
[domain/data model](../architecture/data-model-v0.md), [commands](../architecture/commands.md),
[password policy](../security/admin-bootstrap.md), [resource priorities](../development/pi4b-2gb.md).

**Software slice выполнен и локально проверен 28.10.2026.** [Результат и границы](../DAYS_019_021_RESULT.md). Ниже сохранён исходный рабочий план.
Target — **только Raspberry Pi 4B с2GB RAM**. Scene ML/screen capture/analysis остаются
на laptop. Автоматический memory-pressure controller отдельно запланирован, не реализован.
M0 физически открыт, remote GitHub CI не запускался.

## 1. Цель и расписание

| День | Результат | Демонстрация к концу дня |
|---|---|---|
| 19 | Настоящие browser sessions, CSRF/origin policy, revoke и защищённый WS lifecycle | Login200 по HTTPS, `/me`, разрешённое изменение своего setting; после logout/expiry REST/WS недоступны |
| 20 | Scoped machine tokens и durable security audit | Token secret показан один раз, ограничен owner/binding/scope, revoke запрещает новые действия, audit не содержит secrets |
| 21 | Domain migrations/data dictionary, минимизация и retention | Схема покрывает presets/rules/mappings/habits, upgrade с Д018 работает, запрещённой media/track/DSP history нет |

Каждый день: **360мин предметной работы +120мин сопровождения =8ч**.
Сопровождение:20мин вход/разбор блокеров,40мин diagnostics/tooling,40мин evidence/docs,
20мин review/передача. Трудоёмкость — оценка; скрывать невыполненные проверки ради
календарной отметки нельзя. Номера дней — рабочие этапы, не даты выполнения.

## 2. Входная база и необходимые изменения

1. AuthService уже проверяет Argon2id credentials, bootstrap одноразовый. Но valid
   `POST /api/v1/auth/login` пока503 `AUTH_SESSION_NOT_IMPLEMENTED`, cookies не выдаются.
2. Sessions сейчас только schema `id_hash,user_id,expires_at,revoked_at`. Для idle/absolute
   expiry/CSRF/metadata нужны **новые migrations**, не правка опубликованных шагов1–3.
3. Dispatcher реально исполняет SetUserSetting(timezone), проверяет current roles/CAS/
   dedupe. Он пока принимает trusted principal напрямую: нужен gateway с credential refs
   и повторной проверкой перед cached outcome/admission.
4. `AuthorizeTx` и scope validation имеют короткие allowlists. Добавление token/audit
   прав требует согласованного изменения allowlists, seed roles и tests, не wildcard admin.
5. HTTP сейчас буферизует response. Этот writer не поддерживает websocket Hijacker;
   upgrade нельзя включить просто добавлением route в существующий JSON handler.
6. SQLite startup подтверждает writable access, даже когда readable WAL DB доступна
   read-only. Сохранить regression на фактический RO mount; Docker inspect важнее одного
   `-v ...:ro`, поскольку Compose run уже терял такой override.
7. Dev Core ограничен128MiB/no swap, Go soft96MiB, один hashing job. Последний x86
   container peak~65.66MiB — текущий baseline, не прогноз полного проекта или Pi.
8. Production/hardware fail-closed сохраняется. Для auth smoke нужны HTTPS и synthetic
   secrets, а не отключение Secure cookie или certificate verification.

---

## Д019. Sessions, CSRF и credential-aware gateway

### Цель

Заменить stage verifier реальным login/session flow. Browser cookie подтверждает
identity, но permission/owner/capability checks по-прежнему выполняет Core. Защищённый
WebSocket transport подготовить к Д023 без выдачи fake Dashboard/events acceptance.

### Блок19.1. Session model и migration —60мин

1. Спроектировать opaque session ID из crypto/rand:32random bytes; DB хранит digest,
   не raw cookie. Session ID не user ID/JWT с неподтверждённой revocation semantics.
2. Расширить session metadata: created_at,last_seen_at,absolute expiry,revoked_at,
   user FK и state для CSRF. Использовать UTC/injected clock.
3. Начальные **проектные** defaults: idle30мин, absolute12ч, максимум5active sessions
   на user. При достижении cap — documented policy, например revoke oldest; не unlimited rows.
4. Выбрать synchronizer CSRF token, привязанный к session. State/retrieval нужны для
   reload/multiple tabs; hash-only требование raw session cookie не означает, что
   случайный CSRF state нужно пытаться восстановить из hash. Secret server state —
   только private auth storage, не generic user_settings/logs/audit.
5. Задать indexes/expiry cleanup и bounded query limits. На каждом protected request
   проверять expiry/revocation/user.enabled; positive auth cache пока не вводить.
6. Last-seen writes не переносить в каждый WS frame. Если вводится coalescing,
   определить его точность и тестами доказать, что absolute expiry не продлевается.
7. Login выдаёт новый ID, не принимает pre-auth cookie как готовую session. Renewal/
   rotation не сохраняет возможность fixation старым ID.
8. Прогнать upgrade от настоящей v3 fixture, repeat/checksum и failed-step rollback.

Источник security decisions:[OWASP Session Management](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html).

### Блок19.2. Login/session API и dispatcher gateway —75мин

1. Подключить issuer после успешного AuthService: session commit завершается до
   отправки cookie. Issuer/storage error оставляет503 без Set-Cookie/fake success.
2. Contract: POST `/api/v1/auth/login`→200, GET `/api/v1/auth/me`, получение CSRF state,
   POST logout, POST revoke-all-current-user-sessions. Точные route names закрепить
   в OpenAPI перед реализацией; logout/revoke-all не GET.
3. Cookie candidate `__Host-shh_session`: Secure,HttpOnly,Path=/,no Domain,
   SameSite=Lax. Expiry/Max-Age согласованы с server lifetime; SameSite не заменяет CSRF.
4. `/me` возвращает минимальный current-user DTO, без password_hash/cookie digest/
   internal keys. Ответы no-store. Нет неявного доступа к чужим users по ID в body.
5. Logout сначала persist revocation, затем очищает cookie с теми же attributes.
   Revoke-all означает **browser sessions данного user**, включая текущую; machine
   tokens Д020 не смешиваются с этой операцией без отдельного explicit API.
6. Добавить минимальный authenticated setting endpoint поверх готового dispatcher,
   например `PUT /api/v1/users/me/settings/timezone`. Body — typed value/revision/
   command ID; principal/source/bindings/priority достраивает gateway.
7. Ввести credential reference в trusted principal. На mutation проверять session
   действительность внутри того же admission transaction **до dedupe/CAS**. Не
   ограничиться middleware, которое проверило cookie раньше и сохранило stale principal.
8. Определить race semantics: revoke commit до admission запрещает command; уже
   admitted/in-flight effect не обещается откатить. Для будущего queued work требуется
   повторная проверка перед исполнением. Пока physical/queued handlers не создавать.

### Блок19.3. CSRF, Origin и HTTPS/proxy boundary —90мин

1. Для browser mutations проверять session-bound CSRF token в отдельном header,
   сравнение безопасное. Token не URL query, не metric label, не raw log.
2. Проверять exact allowed origins: scheme+host+port. Не substring match и не
   wildcard CORS. SameSite/cookie flags — дополнительные уровни, не единственная защита.
3. Login CSRF защитить отдельно, когда authenticated session ещё отсутствует:
   same-origin browser policy + strict JSON/custom-header gate либо explicit pre-auth
   CSRF flow. Выбранный вариант описать и проверить forged cross-site request.
4. Определить handling отсутствующего/`null` Origin. Browser flow не должен получить
   bypass только потому, что attacker убрал header; non-browser clients позже используют
   отдельный bearer mechanism. Referer fallback, если выбран, также exact-origin.
5. Проверять Content-Type/body size, unknown fields, malformed/multiple credentials.
   Политика пустых/дублированных auth headers/cookies должна быть deterministic.
6. Secure transport — actual TLS либо явно доверенный proxy с ограниченным backend
   reachability. X-Forwarded-Proto/Forwarded/XFF от произвольного peer игнорировать;
   не объявлять доверенными все private CIDRs. Настроить expected public origin.
7. Добавить локальный HTTPS test/proxy profile с synthetic CA/cert, client доверяет
   только test CA. Не использовать insecure skip verify и не менять системное trust store.
8. Проверить cookie jar реальным HTTPS client: session отправляется по TLS, отсутствует
   по HTTP, prefix flags корректны. Proxy тестировать фактически, не только r.TLS double.
9. Не публиковать backend наружу ради теста; native loopback и existing explicit
   container binding сохраняются. LAN/mDNS/mobile trust и production rollout — отдельные gates.

Источник:[OWASP CSRF](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html).

### Блок19.4. WebSocket security lifecycle v0 —75мин

1. Выбрать maintained websocket library после compatibility/security review, pin
   exact version. Не писать собственный RFC6455 parser ради одного probe.
2. Разделить JSON buffering и upgrade path. Authenticate/Origin/limits выполняются
   до передачи исходного ResponseWriter в websocket upgrade. Denied handshake —
   safe JSON/error, без частично отправленного101.
3. Реализовать минимальный защищённый WS lifecycle на согласованном route:
   upgrade, bounded read/write, heartbeat, session expiry/revoke closure. Business
   event stream/snapshot/filtering/reconnect и Dashboard остаются Д023.
4. Candidate caps для2GB:16connections total,2per session, message<=4KiB для auth
   transport probe; compression выключена. Это stage defaults, не лимиты будущих
   large snapshot messages без отдельного review.
5. Bounded registry по session/user нужен для closure после logout/revoke/disable.
   Commit revocation → notify registry; предусмотреть periodic DB recheck как fallback,
   с явно заданным maximum revoke latency. Нельзя полагаться только на lossy Event Bus.
6. Истёкшая/отозванная session закрывает существующий socket и запрещает reconnect.
   Проверить registry/goroutine cleanup при timeout/normal close/server shutdown.
7. Не передавать raw tokens через URL query. Browser использует secure cookie и
   validated Origin; чувствительные business events до Д023 не публикуются.
8. Выполнить actual network upgrade/revoke/expiry tests, включая fake clock-triggered
   recheck. Если доступны только handler tests — WS lifecycle ещё не PASS.

Источник:[OWASP WebSocket Security](https://cheatsheetseries.owasp.org/cheatsheets/WebSocket_Security_Cheat_Sheet.html).

### Блок19.5. Integration, readiness и отчёт —60мин

1. TLS login→cookie→`/me`→CSRF setting mutation→logout→401; last setting сохраняется.
2. Wrong/missing CSRF/origin, expired/revoked session, disabled user и spoofed metadata
   не изменяют setting/revision/command record. Retry cached command после revoke запрещён.
3. Fake clock проверяет idle/absolute expiry, cap sessions, rotation/fixation и cleanup.
   Race tests используют barriers, не timing luck, для revoke-vs-admission.
4. WS accepted connection закрывается на revoke/expiry; registry очищается. Handshake
   failure не создаёт subscription/worker. Full streaming тестами не подменять этот scope.
5. `auth.required=true` ready200 только когда storage/schema/issuer/validation готовы;
   session table или один200 `/me` не включают production profile автоматически.
6. Перепроверить old memory_fake и stage-disabled profiles, metrics label bounds,
   Core128MiB/no-swap responsiveness и secret sentinel redaction.
7. Обновить OpenAPI, security docs, smoke/CI, `docs/DAY_019_RESULT.md`. Старый stage503
   остаётся только при реальной недоступности issuer, не после успешного session flow.

### Проверки Д019

| ID | Сценарий | Условие завершения |
|---|---|---|
| D19-01 | Migration/hash-only session | Upgrade/repeat/rollback, raw cookie отсутствует в DB/logs |
| D19-02 | TLS login/cookies | Реальный200 login, flags/jar проверены по HTTPS и proxy |
| D19-03 | Expiry/rotation/caps | Fake-clock idle/absolute/rotation и bounded active sessions |
| D19-04 | Logout/revoke-all | Persist revoke, clear cookie, последующие REST401 |
| D19-05 | CSRF/origin/login CSRF | Cross-site/missing/null/invalid cases не вызывают mutation |
| D19-06 | Proxy trust | Spoofed forwarded headers не создают secure/trusted identity |
| D19-07 | Command admission | Current credential check перед dedupe/CAS; deterministic revoke race |
| D19-08 | WS lifecycle | Actual upgrade, revoke/expiry closure и no leaked connections |
| D19-09 | Readiness/memory/failures | Issuer failure fail closed,128MiB fixture без secret leaks/OOM |

**Артефакты:** new migration/session repository/service, credential-aware middleware/
gateway, CSRF/origin/proxy policy, WS lifecycle v0, TLS smoke, OpenAPI и report.

---

## Д020. Machine tokens и security audit

### Цель

Дать service/API клиенту отдельный ограниченный доступ, не browser session и не admin
роль по умолчанию. Обязательные security outcomes сохранять в SQL, независимо от
Event Bus delivery; заодно закрыть audit login/settings, начатых на прошлых днях.

### Блок20.1. Token model/schema —60мин

1. Новая migration для service principals/token records/scopes/resource bindings,
   owner/user refs, created/expiry/revocation и bounded metadata.
2. Token — public selector ID + crypto/rand secret32bytes. Формат/length/prefix
   валидировать до lookup. Raw secret не хранить, только fast digest высокого энтропийного
   секрета; Argon2 остаётся для человеческих passwords, не каждого API request.
3. Token ID и owner UserID различать: существующий dispatcher уже предотвращает
   service identity collision в dedupe. Private hash не возвращать в list DTO.
4. Проектные defaults: expiry90дней, максимум16active tokens на user; scope/binding
   arrays ограничены. Значения подтвердить tests/resource review, не считать нормативом.
5. Нет implicit wildcard, admin token или unbounded delegation. Неизвестный scope
   отвергается, а не тихо удаляется из requested grant.
6. Добавить `tokens:manage`/`audit:read` и прочие нужные **реализованные** permissions
   согласованно в registry, DB seed migration и checker. Role admin не разрешает
   несуществующую capability.

### Блок20.2. Management API и one-time reveal —75мин

1. Authenticated session+CSRF создаёт/list/revoke собственные tokens. Machine token
   не получает право выпускать новые tokens через тот же endpoint.
2. Creation проверяет requested scopes⊆grantor rights и bindings⊆разрешённых resources.
   Чужая zone/user/device, disabled owner и ungrantable scope отвергают весь запрос.
3. На текущем slice доказать доступ к собственному timezone setting. Lights/audio
   handlers пока unsupported: наличие соответствующего scope не делает feature готовой.
4. Полный token показать только creation response, no-store. List возвращает public
   ID/name/scopes/bindings/timestamps/status; не hash и не recoverable secret.
5. Если response оборвался после commit, secret не «показывается повторно»: владелец
   видит public metadata и может revoke/create replacement. Retry contract описать явно.
6. Revoke idempotent, с owner/permission проверкой и transaction. Нельзя revoke чужой
   token только по угадываемому public ID.
7. Поля source/principal/priority не принимать как authority. Generic token gateway
   назначает EXTERNAL_API; Telegram/device-specific mapping придёт с интеграциями.
8. Optional name metadata ограничить по размеру/charset, не логировать raw label
   как arbitrary audit payload; pagination/list limits обязательны.

### Блок20.3. Bearer verifier и общий dispatcher —75мин

1. Принимать token только Authorization Bearer по защищённому transport. Cookies+
   bearer одновременно — явный отказ, без ambiguous precedence. Token не URL/query.
2. Проверить secret digest, expiry/revoked_at, principal enabled, current owner rights,
   requested operation scope и exact binding. Grant не переживает disable owner.
3. Не использовать positive cache с бесконечным TTL; current credential проверять
   перед cached command result и внутри mutation admission transaction.
4. Browser-cookie branch сохраняет CSRF; bearer-only без ambient cookie — отдельный
   auth mechanism. Это не повод отключать CSRF сразу для всех API requests.
5. Candidate limit60requests/мин per token + bounded aggregate peer-IP/admission;
   limiter maps/TTL/capacity задать явно. Last-used writes coalesce, не SQL write на frame.
6. Revoke запрещает новые requests и закрывает активные token-bound channels. Ввести
   тот же bounded registry/invalidation port, что Д019; real screen pairing/stream — позже.
7. Проверить role/scope change между verification и admission, source forgery, wrong
   zone binding, reused command ID после revoke, concurrent revoke/write.
8. Existing user/service scope allowlists согласовать с persistence; header principal
   не принимается от client даже в simulation.

### Блок20.4. Durable audit и privacy —90мин

1. Новая security_audit schema: stable ID/server UTC/action/result/principal ref/
   resource refs/safe reason code и versioned allowlisted metadata, необходимые indexes.
2. События: login success/failure, logout/revoke sessions, token create/revoke/use-denied,
   sensitive setting changes. Не generic request/event-body archive.
3. Password/PHC/session cookie/CSRF/bearer/Authorization/raw request body/query/media/
   track/scene history исключить из records и stderr/metrics. Для failed login без
   principal — reason/counter, не plaintext guessed usernames.
4. Session/token creation/revoke и supported setting mutation commit вместе с
   обязательной audit row. Token/cookie выдаётся после commit. Audit insert failure
   отменяет sensitive mutation; нет success при потерянной обязательной записи.
5. Failed authentication уже denied даже если audit storage недоступен: не создавать
   unbounded fallback queue/file с raw inputs. Задать bounded technical failure counter
   и status policy. Повторный guessing не должен неограниченно раздувать audit.
6. Начальный security retention30дней — проектная policy; batch cleanup с indexes,
   bounded chunk size, fake-clock boundary tests. Read access — отдельный permission,
   current-user/resource ownership и safe pagination; не общий broadcast.
7. Event Bus уведомляет после commit. Subscriber overflow не теряет mandatory audit
   и не откатывает уже committed effect. Publish failure — safe technical observation.
8. Audit record не хранит reversible token secret ради troubleshooting. Verify constant
   metadata size, no log injection и concurrent audit writes/cleanup.

Источник:[OWASP Logging](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html).

### Блок20.5. Сквозные проверки/отчёт —60мин

1. Browser login/CSRF→create token→one-time copy→bearer setting mutation→revoke→401;
   setting/revision и command/audit records сверить в private temp DB.
2. Из двух users/нескольких bindings проверить cross-owner denial, insufficient scope,
   expired/disabled/revoked credentials и отсутствие скрытого admin bypass.
3. Audit failure injection до commit не выдаёт cookie/token/setting success. Overflow
   bus после commit не удаляет audit. Secret sentinel отсутствует в export/list/logs.
4. Повторить token validation/revoke после process restart, quotas/expiry/retention
   через fake clock и concurrency через barriers/race detector.
5. Обновить OpenAPI auth schemes и route-level requirements, safe examples без real
   secret, CI/smoke и `docs/DAY_020_RESULT.md`. Remote CI только после фактического run.
6. Новый credential middleware проверить в Core128MiB fixture. Никакого Redis/нового
   Python worker для sessions/token/audit не добавлять.

### Проверки Д020

| ID | Сценарий | Условие завершения |
|---|---|---|
| D20-01 | Token storage/reveal | Raw secret только creation response, digest-only DB, safe list |
| D20-02 | Owner/grant policy | Cross-user/ungrantable/unknown scopes/bindings rejected |
| D20-03 | Bearer auth | Valid supported action проходит; revoked/expired/disabled401 |
| D20-04 | Scope/resource restrictions | Недостаточные права403 до handler/dedupe disclosure |
| D20-05 | Race/reauthorization | Revocation-before-admission отменяет command, old result не bypass |
| D20-06 | Auth branches/limits | Cookie+bearer ambiguous denied; CSRF для cookie сохранён; maps bounded |
| D20-07 | Durable audit | Atomic sensitive mutation+audit; injected failure rollback |
| D20-08 | Bus/privacy | Audit переживает dropped notification; secrets/raw media отсутствуют |
| D20-09 | Retention/restart/resources | Bounded cleanup/quotas/expiry и реальные process/resource smoke |

**Артефакты:** token/audit migrations и repositories, issuer/verifier/management API,
credential registry hooks, rate/retention policies, OpenAPI/security docs/report.

---

## Д021. Domain schemas и data minimization

### Цель

Подготовить реальную версионируемую схему для следующих domain этапов, не объявляя
готовыми Rule Builder, predictor, camera или learning collector. Generic observations
из bus остаются RAM; habit observations — отдельный минимальный purpose-specific model.

### Блок21.1. Data dictionary и ownership —60мин

1. Сопоставить schema v3+Д019/20 с logical model, requirements и будущими handlers.
   Не создавать вторую users/settings/token schema в другом package.
2. Для каждой сущности записать owner, purpose, source, fields/types, nullable/units,
   revision/schema version, FK/delete policy, indexes, retention, read/write permission.
3. Разделить persisted desired state, operation/audit outcome и runtime applied state.
   Last frame/PCM/features/presence evidence/camera leases в SQLite не переселяются.
4. Проверить model targets habit prediction: время/weekday/zone/mode/color/brightness/
   volume плюс необходимый context presence/ambient/monitor signal/origin/quality.
5. Missing evidence — NULL/UNKNOWN с freshness, не ноль/AWAY. `tv_active` не угадывает
   наличие телевизора по отсутствующему screen-client.
6. Даты UTC; user timezone/context timezone и timezone changes определены для features.
   Schema support не является запуском30-дневного real learning window.

### Блок21.2. Presets/rules/mappings/audit —75мин

1. Новые migrations для presets/preset_entries с owner/revision/zone refs и bounded
   typed fields, без generic arbitrary effect payload по умолчанию.
2. Automation rules: owner/enabled/revision и versioned validated DSL envelope.
   Позднейший Rule Builder использует этот contract; неизвестные future operations
   не считаются исполняемыми от наличия JSON column.
3. Gesture/button mappings: owner/device+gesture/press kind, typed command template,
   confirmation/cooldown fields. Не разрешать произвольные shell/SQL/field paths.
4. Automation audit — отдельный от security audit журнал outcome/action summaries;
   defaults30дней и per-owner access. Delete rule не каскадно стирает audit смысл.
5. Выбрать current minimal delayed-job schema только если нужны реальные invariants;
   иначе задокументировать future migration вместо пустого обещания scheduler/replay.
6. Проверить role/service binding и resource deletion policies. Shared device/zone
   не удаляется автоматически с user; mapping deletion не открывает camera permission.

### Блок21.3. Habits/recommendations/feedback —90мин

1. Спроектировать habit_learning_state: started_at,privacy_epoch,current model refs
   и quality/eligibility metadata. Никакого AUTO eligibility только от наличия rows.
2. Habit observations — allowlisted minimal fields из block21.1, owner+time+purpose.
   Не whole event payload и не unlimited JSON dump; bounded values/units/quality.
3. Recommendation/model metadata — owner/model version/privacy epoch/confidence/
   status/expiry/minimal proposed action. Model artifact отдельно local file, не BLOB
   сериализованного framework в основной auth DB.
4. Feedback — owner+recommendation+decision/time/privacy epoch с uniqueness;
   repeated same decision idempotent, conflicting decision явно обрабатывается.
5. Privacy epoch constraint для future training publication: stale generation не может
   стать current. SQL/repository contract проверяется fixture; реальное training ещё нет.
6. Camera/screen frames, PCM, DSP/beat history, scene labels history, track titles/
   listening history, raw presence CV evidence — запрещены как persisted history.
7. Проверить Q-022 coverage и update dictionary; не закрывать quality/model evaluation
   вопрос лишь из-за полной таблицы.

### Блок21.4. Constraints, indexes и retention —75мин

1. FK/NOT NULL/UNIQUE/CHECK и explicit revisions; finite/range validation в repositories
   и доступных SQL constraints. Не ожидать, что SQLite dynamic typing заменяет validation.
2. Основные indexes: owner+timestamp observations, owner+status+expiry recommendations,
   rule+timestamp automation audit, recommendation feedback uniqueness. Показать query plans.
3. Retention habit observations90календарных дней, automation audit30дней; security
   policy из Д020 отдельно. Feedback/model references не удерживают deleted/private inputs
   бесконечно. Cutoff UTC, timezone применяется к features, не к произвольному rolling cutoff.
4. Cleanup batches bounded, indexed и cancelable. Temp fixture проверяет rows до/на/после
   cutoff и FK effects; не делить numeric durations на «месяц=30дней» без contract.
5. Privacy delete/backup/tombstone generation protocol описать как future workflow,
   не притворяться, что один DELETE уже чистит artifacts/backups/in-flight training.
6. Capacity planning для2GB: disk/index growth, bounded query pagination/exports,
   отсутствие full-table materialization в RAM и100k synthetic rows как measured fixture,
   если помещается в test budget. Число rows — test size, не заявленный product maximum.
7. Сохранить one Core database owner; Runtime получает только authorised minimal data
   через будущий port, а не DB mount/raw auth database.

### Блок21.5. Upgrade, fixtures и handoff —60мин

1. Fresh DB и реальная копия Д018→Д019→Д020→Д021: существующие admin/settings/session/
   token data сохраняются; миграции старых шагов/checksums не переписаны.
2. Invalid FK/ranges/unknown fields/cross-owner/duplicate feedback и failed migration
   не оставляют partial state. Привязать negative cases к data dictionary.
3. Repo fixtures создают/читают минимальные preset/rule/mapping/habit/recommendation/
   feedback rows. Это schema/repository acceptance, не работающий predictor/UI.
4. Проверить private temp DB/logs/exports на запрещённые fields и sentinels. Public
   artifacts — summaries, не private SQLite file даже с synthetic auth hash.
5. Повторить quality/OpenAPI/native/persistent Compose, DB writable/readonly regression,
   resource caps и старые auth cases. Full-profile physical Pi tests остаются BLOCKED.
6. `docs/DAY_021_RESULT.md`, dictionary/ER/update Q-022/Q-030 и список готовых repositories
   передать Д022: React login/logout/client; Д023: реальные events/Dashboard.

### Проверки Д021

| ID | Сценарий | Условие завершения |
|---|---|---|
| D21-01 | Schema/upgrade | Fresh/latest/Д018 upgrade, данные сохранены, failure rollback |
| D21-02 | Presets/rules/mappings | Typed/owned rows и FK/revision cases; unsupported execution не маскируется |
| D21-03 | Habit target coverage | Volume/presence/ambient/timezone/origin/quality/NULL semantics отражены |
| D21-04 | Privacy | Нет generic media/track/DSP/scene history и bus persistence |
| D21-05 | Feedback/model epochs | Uniqueness/owner/stale-generation fixtures исполняются |
| D21-06 | Retention |90/30-day UTC boundaries, bounded cleanup, integrity refs |
| D21-07 | RAM/query planning | Indexed/paginated queries без полного dataset в RAM; measured synthetic fixture |
| D21-08 | Handoff | Data dictionary/ER/permissions/remaining model/privacy work перечислены |

**Артефакты:** new domain migrations/repositories/tests, `docs/storage/data-dictionary.md`,
updated logical ER/retention rules, `docs/DAY_021_RESULT.md`, redacted evidence.

---

## 3. Предлагаемые файлы и воспроизводимость

```text
internal/auth/                 # Sessions, CSRF, token verifier, credential refs
internal/storage/migrations/   # Новые версии поверх уже реализованных1–3
internal/api/                  # Login/me/logout/token/setting gateway, upgrade path
internal/commands/             # Current-credential admission check
internal/security/            # При необходимости audit policy; без второго auth owner
scripts/                      # TLS/session/token/schema/persistence smoke
configs/                      # Explicit HTTPS/auth profiles, no secrets
deploy/docker/               # Test proxy profile и private backend topology
docs/security/               # Sessions/CSRF/tokens/audit contracts
docs/storage/data-dictionary.md
docs/DAY_019_RESULT.md
docs/DAY_020_RESULT.md
docs/DAY_021_RESULT.md
docs/results/days019-021/     # Redacted summaries, не cookies/tokens/DB
```

Имена directories — предложение: сначала review actual code, не создавать дублирующие
modules. Existing middleware/Make/CI/OpenAPI/config/readiness/resources обновлять вместе
с новым кодом. Frontend auth page ещё Д022, не делать fake login UI раньше backend.

Проверки каждого дня: targeted tests→Go race/vet→TypeScript/Python unchanged gates→
OpenAPI/actionlint/build→native/TLS/persistent Compose. На этой машине Go запускается
через `make GO=/snap/go/current/bin/go …`. Новые Make targets добавлять с настоящей
реализацией; сейчас будущие session/token smoke команды не объявляются существующими.

PASS-HOST/SIM/COMPOSE — только исполненные cases; PASS-CI требует remote run/commit.
Pi2GB memory/CPU/temp/FPS и LAN/mobile trust — отдельная evidence, не x86 simulation.
При блокере exact scope/BLOCKED записать честно, не ослаблять cookie/origin/auth ради
одного200. Публикация GitHub/изменение host swap/production deploy этим планом не выполняются.

## 4. Выход к Д022

К началу frontend auth нужны реальные HTTPS login/logout/`/me`/CSRF, truthful readiness,
bearer service permissions и safe audit, compatible schemas и dictionary. Д023 завершит
business WS snapshot/filtering/reconnect, Д024 — полноценные virtual fixtures, Д025 — M1.
Сами по себе новые таблицы/101 upgrade/один login200 не закрывают M1 или physical M0.
