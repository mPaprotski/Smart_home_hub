# Результат Д019–Д021

28.10.2026: **software slice реализован и проверен локально**. Основание —
[подробный план](plans/DAYS_019_021_DETAILED.md). Аппаратная цель неизменна:
**Raspberry Pi 4B, только2 GB RAM**. Проверки выполнены на x86 с simulated adapters.

| День | Реализовано | Проверки |
|---|---|---|
| [19](DAY_019_RESULT.md) | HTTPS session issuer, CSRF/origin, revoke, защищённый timezone gateway, auth-only WS | D19-01–09: PASS-HOST/SIM |
| [20](DAY_020_RESULT.md) | Scoped hash-only tokens, owner/current-role checks, atomic audit, quotas | D20-01–09: PASS-HOST/SIM |
| [21](DAY_021_RESULT.md) | Schema6, typed repositories, privacy epochs,90/30-day retention | D21-01–08: PASS-HOST/SIM |

## Проверка

`make check`: Go format/vet/race, frontend format/lint/typecheck/build,
Python format/lint/mypy/14 tests/wheel, OpenAPI — проходят. Дополнительные tests:
WS idle-expiry, verify→revoke→admission, role removal/token expiry, closed Event Bus
при committed setting/audit, model aging/FK cascade. Native HTTPS, прежние native
simulation/persistence smokes и Compose persistence/auth smoke проходят.
Actionlint проходит. Govulncheck:0 reachable/imported vulnerabilities; остаётся
1 advisory в required module для неимпортируемого x/crypto/openpgp (GO-2026-5932),
поэтому не заявляется отсутствие всех module advisories.

[Машинная сводка и логи](results/days019-021/summary.json),
[источники/команды](results/days019-021/README.md).

Compose HTTPS: Core hard limit128MiB, memory+swap128MiB (swap отсутствует),
GOMEMLIMIT96MiB, UID10001, read-only rootfs, loopback publication, проверенный
тестовый CA. Измеренный peak46.0MiB, после restart26.2MiB; OOM=false.
Persistence flood regression peak66.6MiB, health200, flood401/429. Данные и контейнеры
одноразовых tests удалены. Фактический RO data mount проверен Docker inspect;
попытка startup с ним закрывается. Ни host swap, ни глобальный CA trust не менялись.

100000-row synthetic SQLite fixture: limit100, covering time index, retention;
примерно2.86s insert/read/plan/cleanup на host без race. Это не Pi benchmark и
не оценка общего RAM всех сервисов. [Словарь схемы](storage/data-dictionary.md)
и [security policy](security/sessions-tokens.md) описывают точные границы.

## Что остаётся

- Д022–Д025: UI/login, Dashboard, business WS/snapshot/reconnect, остальные fakes,
  полноценная simulation и M1 review. Текущий WS отвечает auth_only/events:false.
- Реальные Pi SSH, GPIO/LED/audio/camera и RAM/thermal/HIL measurements; M0 открыт.
- Collector/predictor/training, rule execution, machine pairing, backup/secure erase
  и privacy deletion controller. Schema не объявляет эти функции готовыми.
- Полный resource degradation controller и совокупный memory budget на2GB.
- Remote GitHub CI, commit/push/deployment не выполнялись.

Legacy dev profile остаётся credentials-only; настоящие sessions включаются явным
`configs/auth.dev.yaml`. Production/hardware profiles по-прежнему fail closed.
