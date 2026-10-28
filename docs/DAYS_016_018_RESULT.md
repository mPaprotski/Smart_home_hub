# Дни16–18: фактический результат

Дата:26.10.2026, Europe/Minsk. [План](plans/DAYS_016_018_DETAILED.md).
**Программный срез SQLite/commands/credentials реализован.** Sessions/CSRF/token
issuance остаются на Д019–Д020, физический M0 и remote CI не закрыты.

Одновременно выполнено уточнение владельца: **единственная плата — Raspberry Pi 4B,
только2GB RAM**. ТЗ обновлено доv1.1, исправлены requirements/BOM/plans/inventory/
release profile/preflight, добавлен [ресурсный бюджет](development/pi4b-2gb.md).
Прежние требования4GB больше не используются. Старые measured evidence/hash сохранены
как исторические; новый SHA ТЗ фиксируется в summary этого этапа.

## Выполнено по дням

| День | Реализация | Отчёт |
|---|---|---|
| 16 | Pure-Go SQLite3.54.1/WAL/FK/FULL, file privacy,3 migrations/checksums, users/roles/settings/zones repositories, CAS, errors, storage lifecycle/readiness | [Д016](DAY_016_RESULT.md) |
| 17 | Trusted command identity/source, ownership/scopes/bindings, persisted setting handler, CAS/idempotency/reauthorization/rollback, manual lease policy hook | [Д017](DAY_017_RESULT.md) |
| 18 | Local one-time bootstrap, Argon2id, strict PHC bounds, persistent account cooldown, IP/global limits, CLI verify и stage login API | [Д018](DAY_018_RESULT.md) |

Инструкции: [Storage](development/storage.md), [permissions](security/permission-matrix.md),
[admin/password policy](security/admin-bootstrap.md), [OpenAPI](api/openapi.yaml).

## Что действительно проверялось

- Product Go unit/integration/race tests, включая настоящие file-backed DB и два
  connections, failed migration/upgrade/checksum/newer-schema, FK/CAS/rollback,
  BUSY/cancel/readonly/corruption и bounded SQLITE_FULL без заполнения host disk.
- Отдельные процессы: committed setting переживает внезапный exit; uncommitted
  update не сохраняется. Это process crash recovery, не отключение питания Pi.
- Commands: две реальные user rows, revoked roles/disabled user, service binding/
  identity isolation, concurrent duplicate/CAS, changed-payload conflict, record
  insert failure rollback, lease overlap/expiry и отсутствие продления retry.
- Passwords: random salts, hash verification/PHC bounds, random source failure,
  Unicode/length policy, bounded explicit stdin/no implicit pipe, dummy path,
  account/IP/global bounds и cooldown после restart.
- Native smoke: bootstrap once, admin verify/restart, persistent=true при
  simulated=true, ready200 core-only и ready503 required-auth, login401/429/503,
  отсутствие Set-Cookie/plaintext/hash logs, observer/health при synthetic flood.
- Persistent Compose: actual build/start/bootstrap/stop/recreate/verify, private
  data volume, read-only rootfs/UID10001,128MiB hard cap без swap, login flood,
  cgroup memory peak, readonly/unwritable storage fail-closed, test volume cleanup.
- Старые memory_fake/not_implemented native smoke сохранены. Frontend/Python
  format/lint/types/build и14 Python cases прошли, OpenAPI/actionlint проверены.
- Govulncheckv1.9.0:0 reachable vulnerabilities и0 affected imported packages;
  module-level GO-2026-5932 относится к **неимпортируемому openpgp** внутри
  x/crypto, fix N/A. Это не утверждение «в модулях вообще нет advisories».

[Quality log](results/days016-018/quality.log), [native](results/days016-018/native-smoke.json),
[Compose](results/days016-018/compose-smoke.json), [benchmark](results/days016-018/argon-benchmark.txt),
[dependency scan](results/days016-018/govulncheck.log), [summary](results/days016-018/summary.json).

## Важные результаты проверки ошибок

1. Первый native restart smoke ожидал401 на уже exhausted account; реальный persistent
   cooldown дал429. Harness исправлен и теперь отдельно подтверждает сохранение cooldown.
2. Readonly regression сначала завис: Compose run потерял RO override, и фактический
   mount оказался RW. Harness исправлен на отдельный isolated Engine container с
   inspect RW=false до запуска. Дополнительно добавлены startup writable transaction
   probe без изменения metadata и readable mode=ro regression test. Ни один зависший
   или неверно сконфигурированный запуск не объявлен PASS.
3. Существующие public data directory/file отвергаются; tests создают отдельный0700
   data subdirectory. Правила не ослаблены ради зелёного теста.

## Границы результата

Working persistent command — SetUserSetting(timezone). Lighting arbitration — tested
policy port, не готовый Lighting Engine/физический Master OFF. Batch/automation/delayed
jobs/adapter commands пока отклоняются; slow physical adapter/UNKNOWN_OUTCOME recovery
не имитируются успехом. Command-event publisher пока не подключён; SQL outcome самодостаточен.

Credential verification — не session login. Valid HTTP credentials намеренно503
AUTH_SESSION_NOT_IMPLEMENTED, никаких временных bearer tokens/cookies. Required auth
readiness остаётся503, production/LAN controls не открыты. Machine credentials/revoke
и durable security audit завершаются позднее.

RAM observations/media history не записываются в SQLite. Backup API/scheduler/restore,
SD power-loss/cooling/real drivers и общий ML/CV/DSP memory/performance gate на Pi2GB
не реализованы этими днями. Текущий Core128MiB тест на x86 не доказывает fit всего проекта.

Измеренный x86 Argon2id baseline:~32.1ms/verify,~19.9MB allocations/op,5 repetitions;
это не peak RSS и не Pi latency. Реальный container peak — в Compose JSON. ML budgets
в отдельном профиле — кандидаты, а не измеренные reservations.

## Передача на Д019

AuthService/SQLite users/roles/session schema готовы для opaque session issuer,
hash-only storage, expiry/logout/revoke-all, cookie/CSRF/origin policy и `/me`.
Positive HTTP login200 подключать вместе с этими защитами. Trusted proxy IP policy,
real machine token verifier, dispatcher transport gateway и security audit пока не готовы.

Изменения локальные, commit/push/remote workflow не выполнялись. CI workflow расширен
persistent native/Compose и pinned govulncheck; PASS-CI не заявлен.
