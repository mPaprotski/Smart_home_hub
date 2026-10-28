# Sessions, machine tokens и security audit — Д019–Д021

Реализовано 28.10.2026. Target: Raspberry Pi 4B **2 GB RAM**; физическая проверка
не выполнена. Production/hardware profiles закрыты. Это локальный development slice.

## Включение HTTPS

По умолчанию diagnostics/credentials-only профили сохранены: валидный пароль без
session issuer возвращает503, cookie не выдаётся. Для реальных сессий включается
`configs/auth.dev.yaml`: `auth.enabled: true`, SQLite, `auth.public_origin` с HTTPS
и TLS certificate/key. Пара cert/key обязательна. Secrets не помещать в Git.

Локальный исполняемый пример с одноразовыми сертификатами и проверкой цепочки:
`make smoke-auth`. Для Compose добавить к dev+sqlite файл
`deploy/docker/compose.auth.yaml`; в private volume нужны `tls/server.pem`,
`tls/server.key`, `tls/ca.pem`, принадлежащие UID10001. `HUB_PUBLIC_ORIGIN`
должен точно соответствовать HTTPS origin клиента, включая порт.
`--healthcheck-ca` добавляет только явно указанный CA к trust конкретного probe;
проверка сертификата не отключается. Smoke не меняет системное хранилище доверия.

Proxy termination разрешается только для явно перечисленных host CIDR /32 или
/128. Только от такого peer принимается единственный `X-Forwarded-Proto: https`.
X-Forwarded-For не определяет пользователя/права. Proxy должен перезаписывать XFP,
сохранять публичный Host и закрывать прямой доступ к backend; этот dev smoke не
является production reverse-proxy deployment.

## Browser session

Login: `POST /api/v1/auth/login`, JSON username/password <=4096 bytes,
`Origin` точно равен public_origin, `X-Hub-Request: 1`. Пароль проверяет Argon2id;
одна hashing job, существующие account/peer/global limits сохранены.
Каждый успешный login выдаёт новый random256-bit cookie. SQLite содержит SHA256
cookie, owner, created/last-seen/expires/revoked и CSRF state. Raw cookie не хранится.
CSRF state хранится в закрытой БД: утверждение «в БД нет никаких secrets» неверно.

Cookie: `__Host-shh_session`, Secure, HttpOnly, SameSite=Lax, Path=/, без Domain,
12h absolute и30min idle. Максимум5 активных sessions/user; старейшая вытесняется.
HTTP authentication обновляет last-seen. WS heartbeat не продлевает idle.
После restart срок сохраняется. Legacy sessions v3 без CSRF не становятся валидными.

`GET /api/v1/auth/me` возвращает только user_id/credential_kind;
`GET /api/v1/auth/csrf` возвращает CSRF текущей сессии. Для mutations обязательны
точный Origin и `X-CSRF-Token`. Missing/null/cross-origin запрещены.
Logout/revoke-all — POST с CSRF; изменение revoked и audit атомарно, cookie очищается.
Revoke-all относится к browser sessions, machine tokens отзываются отдельно.
Нет CORS wildcard. Browser UI/login page будут реализованы в последующие дни.

## Machine tokens

Создание/список/отзыв доступны только browser session с актуальным `tokens:manage`,
только для своего owner. Secret `selector.secret` показывается один раз в201;
SQLite хранит selector и SHA256 random256-bit secret, не сам secret.
Выдача ограничена16 активными tokens/owner,90-day expiry, names1..64 printable ASCII.
Scopes сейчас только `settings:read`, `settings:write`; binding только `user:<owner>`.
Нельзя выдавать права выше текущих owner roles, unknown scope/wildcard/foreign binding
отклоняются. Это ещё не протокол pairing screen-client/MQTT/Telegram.

Bearer-only requests требуют HTTPS. Cookie+Bearer, несколько Authorization или
дубли session cookie отвергаются. Cookie не служит fallback для неверного bearer.
Token не может выдавать другие tokens или читать session CSRF. Admission проверяет
живой credential, owner enabled, current roles/scopes/bindings внутри той же SQL
transaction **до** cached command result и CAS. Отзыв, завершившийся до admission,
закрывает действие. Уже committed command отзыв не отменяет.
Rate limits:60/min token и120/min peer, bounded map4096; переизбыток fail closed.

## Command и WebSocket boundaries

Единственный HTTP mutation gateway — `PUT /api/v1/users/me/settings/timezone`:
value, expected_revision и command_id обязательны; unknown fields запрещены.
Identity/source выводятся сервером: WEB_UI или EXTERNAL_API. Клиент не задаёт
principal/source/priority. CAS conflict412; command-id/payload conflict409.
Это не универсальный gateway оборудования.

`/ws/events` принимает настоящий WSS upgrade с live credential и точным Origin,
без query credentials. До16 connections total,2/credential, compression off,
read limit4096. Первое сообщение `{"transport":"auth_only","events":false}`.
Revoke закрывает зарегистрированные connections; periodic15s check закрывает
expired/disabled credentials; ping timeout3s. Race verify/register закрывается
повторной проверкой. Shutdown отменяет registry и ждёт handlers.
Business events, snapshot, replay/reconnect protocol относятся к Д023.

## Audit и отказоустойчивость

SQL security_audit — источник истины: login/logout/revoke/token management/setting
mutation и allowlisted denials. Mutation и success audit в одной transaction;
ошибка audit отменяет mutation. Denials агрегируются по action глобально1/min,
не пишут введённые usernames, IP, payload или произвольные exception strings.
`hub_audit_write_failures_total` показывает неудачный denial audit.
Успешный commit публикует bounded non-sensitive notification; при потере bus
notification SQL остаётся сохранённым, растёт `hub_audit_notifications_failed_total`.
Notification содержит только вид изменения, без secret/пользователя/setting value.

Audit retention30 UTC days, habit90 UTC days, bounded maintenance batches100/min.
Истёкшие/отозванные credentials очищаются; expiry/revoke проверяются и до удаления.
Список audit только собственный principal и audit:read, cursor/limit100.
Глобально агрегированные denials не доступны в own-principal списке; отдельный
operator API не реализован. Backup hygiene/secure erase/privacy deletion controller
остаются отдельными задачами. Логи smoke синтетические и secrets не содержат.
