# Evidence Д019–Д021

Локальный прогон 28.10.2026, Europe/Minsk; x86 host, synthetic hardware,
целевой Pi4B2GB отсутствует в тестовом контуре. Git commit/push не выполнялись.
Исторические результаты Д016–Д018 сохранены отдельно и не перезаписаны.

| Файл | Получение / смысл |
|---|---|
| quality.log | `make check GO=/snap/go/current/bin/go`; format/vet/race/types/tests/OpenAPI/build |
| security-race.log | Дополнительный `go test -race ./internal/api ./internal/auth ./internal/storage` |
| model-aging.log | Focused race test старения model metadata и FK feedback cascade |
| native-smokes.log | `make smoke-auth smoke smoke-persistence`; disposable TLS CA, local process restart, прежние dev profiles |
| native-auth.json | Первый JSON result из native-smokes.log; certificate verification включена |
| compose-auth.json | `scripts/auth_compose_smoke.py`; schema7,128MiB hard/no swap, UID10001, read-only rootfs |
| compose-persistence.json | Прежний persistence smoke на новом образе; bootstrap/recreate/flood/actual read-only data mount |
| dataset.log | `go test -v ./internal/storage -run '^TestBoundedIndexedDatasetAndAuditCutoff$'`;100000 synthetic rows |
| govulncheck.log | `.tools/bin/govulncheck ./cmd/... ./internal/... ./telegram/...`; x/crypto/openpgp advisory не reachable/imported |
| actionlint.log | `.tools/bin/actionlint`; пустой вывод означает success |
| summary.json | Проверенные границы, target,26 plan checks, открытые gates |
| source-sha256.json | SHA256 файлов source/config/build/tests/API на момент прогона; не Git commit |

Go cache/module cache локальны в .cache. Tool versions pinned в go.mod/locks/.tool-versions.
Проверки, которым нужен listener или Docker, выполнялись с разрешённым доступом к
локальному network/daemon. Secrets synthetic; raw cookies/tokens/keys не записаны
в evidence. Smoke отказывается использовать существующий test project/volume;
созданные disposable данные удалены. Host services, swap и global CA trust не менялись.

После полного quality gate добавлен отдельный model-aging regression; затем весь
quality gate повторён на окончательной append-only schema7. Дополнительный race log
зафиксировал security cases до добавления time-index migration; полный quality log
покрывает итоговый код. Измерения памяти — cgroup memory.peak контейнера, не RSS
всего проекта и не оценка реального Pi. Нет remote CI/HIL/performance acceptance.
