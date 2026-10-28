# Storage Д016: SQLite и отдельный RAM cache

Target: **Raspberry Pi 4B, только2GB RAM**. [Ресурсный бюджет](pi4b-2gb.md).
Текущая implementation: `internal/storage`, SQL migrations1–3 embedded в binary.

Driver `modernc.org/sqlite v1.61.0` (BSD-style license upstream), SQLite **3.54.1**
подтверждена запросом в integration test. Driver pure-Go, поэтому прежняя сборка
`CGO_ENABLED=0` сохранена. WAL-reset patched versions проверяются при Open; host sqlite3
не используется как доказательство embedded version. Upstream:[WAL](https://sqlite.org/wal.html).

## Запуск и владение

```bash
make GO=/snap/go/current/bin/go build-go
./bin/hub --config configs/sqlite.dev.yaml
```

По умолчанию data directory `.runtime/data`, создаётся с0700, DB с0600. Существующие
каталоги/файлы с group/other permissions отвергаются, произвольный recursive chmod
не выполняется. `HUB_DB_PATH` — allowlisted infrastructure override; путь скрыт в
Config JSON/logs. Только Core открывает SQL. Python/bot/frontend не получают mount.

File-backed sqlite отдельно от memory_fake/not_implemented. При ошибке нет fallback,
listener не стартует. Memory observations cache остаётся RAM-only и не заменяется
SQL repository. Mode `:memory:` не может объявить persistent=true.

Один connection writer, MaxOpenConns/MaxIdleConns=1, `_txlock=immediate`. На каждом
новом connection DSN задаёт foreign_keys=ON, busy_timeout=500ms по умолчанию,
synchronous=FULL. Startup timeout10s. WAL/FK/FULL readback проверен. После migrations startup выполняет transaction с no-op UPDATE schema_migrations: readable-but-readonly WAL DB не объявляется writable/READY. WAL/SHM живут
рядом с DB, поэтому монтируется весь data directory. Справочник:[PRAGMA](https://sqlite.org/pragma.html).

## Схема и transactions

- V1: users/roles/user_roles/role_permissions, sessions schema, devices/zones/user_settings.
- V2: command_records и field-scoped manual_overrides.
- V3: bootstrap marker, bounded account attempts и private HMAC limiter key.

User не seed-ится. Password_hash только от AuthService, кроме synthetic repository
fixtures. Sessions таблица не означает session issuer: ни одной рабочей user session
на Д018 нет. Desired zone state отделено от live frame и applied hardware state.

Каждый migration step и его checksum/version commit вместе. Повторное применение
no-op; altered checksum/newer schema/order gap отвергаются. Failed step откатывает
только свой transaction; предыдущие успешно применённые версии сохраняются.
При startup error DB не стирается и не подменяется новой. Migrations запускаются
до AuthService/observer/HTTP; storage закрывается после пользователей БД.

Repositories проверяют known setting keys и timezone, zone mode/finite brightness,
FK/unique constraints и expected revision. Значение с expected=0 создаётся только
если отсутствует; update выполняется CAS. Hardware/hash/network I/O внутри transaction
не выполняются. SQL values — bound parameters.

Driver errors классифицируются в safe codes STORAGE_BUSY/READ_ONLY/CORRUPT/FULL/
CONSTRAINT/CANCELLED/UNAVAILABLE, без raw SQL/path/value. Public Auth API дополнительно
унифицирует storage failures в AUTH_UNAVAILABLE. Domain callbacks unit-of-work должны
возвращать safe domain error, а не произвольный текст с payload.

## Проверки

Integration tests используют настоящие files/connections: repeat/upgrade/failed
migration/checksum/newer-schema, FK на reopened connections, CAS/rollback, два writer
connections с BUSY, canceled context, corrupt/private/read-only DB и bounded SQLITE_FULL
через max_page_count. Последний не заполняет host disk.

Process helper делает commit и внезапный `os.Exit` с открытой/uncommitted transaction.
Другой процесс читает прежнее committed setting после recovery. Это не отключение
питания Pi/SD; physical storage reliability остаётся отдельной проверкой.

```bash
make GO=/snap/go/current/bin/go smoke-persistence
make compose-persistence-check
docker compose -p smarthome-d18-smoke -f deploy/docker/compose.dev.yaml -f deploy/docker/compose.sqlite.yaml build
.venv/bin/python scripts/persistence_compose_smoke.py
```

Compose smoke создаёт только свой новый private volume, admin, stop/recreate/verify,
readonly/unwritable failure и удаляет свой volume. Существующий test project/volume
он трогать отказывается. Для своей dev DB используйте другой project name и не
вызывайте `down --volumes`, если данные нужны.

Автоматический backup/restore/retention/privacy-purge здесь не реализован. Не копировать
один hub.db во время WAL writes. Recovery snapshot потребует supported backup API либо
проверенного stopped/checkpoint workflow; policy будущего этапа:[SQLite Backup](https://sqlite.org/backup.html).
