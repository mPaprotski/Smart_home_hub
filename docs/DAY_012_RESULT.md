# Д012. Config/lifecycle/logging — результат

Созданы strict loader с env whitelist, typed Secret/redaction, semantic LED mapping checks, sequential startup/reverse cleanup и bounded stop. [Владение конфигурацией и exit codes](development/configuration.md).

| Проверка | Результат / evidence |
|---|---|
| D12-01 Ownership/precedence | PASS-HOST; YAML/defaults + only infrastructure env; disabled secrets optional; user preferences не захватываются |
| D12-02 Syntax/semantic errors | PASS-HOST; unknown/duplicate/multiple-document/type/ranges/NaN/overlap/matrix/side-count checks |
| D12-03 Partial startup | PASS-HOST; component failure очищает успешно начатые components в reverse order |
| D12-04 Cancellation | PASS-HOST/SIM; canceled context, repeated stop, shutdown deadline и настоящий SIGTERM exit0/closed listener |
| D12-05 Redaction | PASS-HOST; JSON/fmt/GoString/safe config errors и actual process sentinel tests |
| D12-06 No hidden hardware | PASS-SIM; product wiring только synthetic; production/hardware/public bind отклоняются до listener |

Код: `internal/config`, `app`, `logging`, `cmd/hub`. [Unit/race log](results/days011-015/quality.log), [black-box failure evidence](results/days011-015/config-smoke.json), [process evidence](results/days011-015/native-smoke.json). Не загружается `.env`, нет реальных tokens. ML/MQTT endpoint/transport integration, production auth и SQLite — будущие; enabled integrations с token всё ещё explicit NOT_IMPLEMENTED/degraded.
