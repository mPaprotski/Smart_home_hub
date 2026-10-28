# Структура и зависимости

- `cmd/hub`: composition root, flags, signals; `cmd/screen-client`, `telegram/cmd/bot`: buildable NOT_IMPLEMENTED entrypoints.
- `internal/config`: strict YAML/env; `app`: lifecycle; `logging`: JSON; `buildinfo`: version/commit.
- `internal/api`: stdlib net/http read-only routes/middleware/metrics; `system`: cached dependency health evaluator.
- `internal/events`: bounded RAM event bus/envelope.
- `internal/devices`: ports/DTOs, bounded LED frame fake, matrix mapping.
- `internal/storage`: explicit bounded nonpersistent observation store.
- `internal/simulation`: seeded synthetic sources and Core observer, actual bus delivery.
- `frontend`: React/TS/Vite static development shell; `ml`: installable Python scalar-observation contract.
- `deploy/docker`, `.github/workflows`, `scripts`, `Makefile`: dev build/check/start orchestration.
- `docs`, `hardware`, `configs`, `experiments`: существующие аналитика/спайки и новые отчёты, preserved.

Один root Go module `github.com/mPaprotski/Smart_home_hub`. Product `internal` не импортирует experiments; CI/Make перечисляют product packages, старые module-off probes остаются отдельными. Camera/GPIO/BlueZ/PipeWire drivers не импортируются продуктом. Нет import-time workers/devices. Владелец lifecycle собирается явно в main: observer → fake adapter suite → HTTP; stop в обратном порядке, затем bus.Close.

Будущие lighting/audio/automation/auth модули добавляются при реализации, без пустых public API. `bin`, local `.venv/.tools/.cache`, frontend artifacts, Python build metadata, SQLite/env/secrets игнорируются. ТЗ сохранено без изменений. Root `.git` workspace не переинициализирован; целевой nested repo имеет прежний Git remote, исходники пока untracked.
