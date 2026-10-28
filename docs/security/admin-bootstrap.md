# Admin bootstrap и credentials: Д018

Target hardware: **Pi4B, только2GB RAM**. Bootstrap работает локально с explicit SQLite
profile; anonymous HTTP bootstrap и default password отсутствуют.

```bash
make GO=/snap/go/current/bin/go build-go
./bin/hub admin bootstrap --config configs/sqlite.dev.yaml --username admin
./bin/hub admin verify --config configs/sqlite.dev.yaml --username admin
```

CLI читает password с TTY без echo; bootstrap просит confirmation. Password не flag,
не environment value. `--password-stdin` — explicit mode для pipe/CI с **synthetic**
secret; real password не писать literal в shell history. Нет TTY и explicit stdin —
быстрый отказ. Line protocol снимает только terminal LF/CRLF, не trim пробелы/пароль.
Input ограничен1024bytes; buffers очищаются best effort, Go/JSON immutable strings
не позволяют обещать полное secure erase.

Вся DB directory0700, DB0600. Bootstrap атомарно создаёт normalized unique username,
user/admin binding и persistent singleton marker. Повтор/конкурентный запуск не
перезаписывает hash. Даже удаление/disable users не снимает marker. Password hash
вычисляется до transaction; failed transaction не оставляет partial account.
Reset/recovery administrator — отдельный будущий явный workflow.

## Пароли и cost

Policy новых passwords:>=15 Unicode characters,<=1024 UTF-8 bytes; без composition
rules и скрытой normalization/trim. Username — lowercase trimmed ASCII letters,
digits, dot/underscore/hyphen,1..64bytes; policy проверяется до account creation.

Argon2id через `golang.org/x/crypto v0.58.0`: m=19456KiB,t=2,p=1, random salt16bytes,
output32bytes, encoded PHC/v19. Базовая конфигурация сверена с
[OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html).
PHC parser проверяет algorithm/version/size/base64/порядок параметров и bounds
m19456..65536KiB,t2..4,p1..4 до allocation. Digest compare constant-time.
Один hash job одновременно, без unbounded wait queue; slot освобождается только
после фактического завершения Argon2. Отмена ctx не «убивает» вычисление преждевременно.
Новые hashes используют baseline; policy более дорогих imported hashes потребует
отдельного memory benchmark. Automatic rehash/import не реализованы.

Неизвестный/disabled/invalid username и malformed stored hash проходят dummy path
и возвращают INVALID_CREDENTIALS. Это уменьшение timing discrepancy, не обещание
идентичных end-to-end timings. Dummy hash не является действующей учётной записью.

## Limits

- Account: максимум5 attempts/15мин до успешной verification; при success счётчик
  сбрасывается. После exhaustion account blocked до окончания окна, включая valid password.
- Account keys: HMAC digest, persisted private random key,<=4096 rows; expiry cleanup
  при admission. Cooldown реально переживает процессный restart.
- Peer IP:20requests/мин, RAM map<=4096 keys с TTL. После restart IP window сбрасывается.
- Global: один credential operation и один Argon2 job; при saturation429 без queue.

Это проектные defaults. Full limiter возвращает отказ для новых keys, не unlimited
allocation. Rate-limit key names/passwords не попадают в metrics/logs. X-Forwarded-For
игнорируется для peer identity. С Д019 поддерживается точечный trusted proxy только
для определения TLS transport: [policy](sessions-tokens.md). Account lockout
ограничен временем и может временно затронуть пользователя при targeted guessing;
recovery/production tuning — отдельный этап.

## Legacy stage HTTP contract (auth.enabled=false)

POST `/api/v1/auth/login`: JSON only,<=4096bytes, known fields username/password,
400 malformed,401 unified invalid credentials,429 Retry-After60s,503 auth unavailable.
**Valid credentials также дают503 AUTH_SESSION_NOT_IMPLEMENTED** до подключения
session issuer на Д019. Нет Set-Cookie/access token, no `/me`, command API не открывается.
Credential verification доступна через service/CLI; это не завершённый login flow.

С `auth.required: true` readiness503 даже после SQLite/admin setup, потому что sessions
не включены в legacy profile. В configs/auth.dev.yaml issuer уже реализован;
[новый HTTPS contract](sessions-tokens.md). Explicit core-only development profile может быть ready200 при
persistent=true/simulated=true. Production/hardware modes fail closed.
Реальные passwords — local TTY; HTTP smoke использует только synthetic data.

[OpenAPI](../api/openapi.yaml), [native/Compose evidence](../DAYS_016_018_RESULT.md).
Baseline verify на x86 host ~32.1ms/op и~19.9MB allocations/op,5 repetitions; это не
peak RSS и не результат Pi. Смотрите raw benchmark и container cgroup peak отдельно.
Повторять на Pi4B/2GB до принятия auth performance gate.
