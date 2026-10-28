# День 10 — Technical baseline M0

Дата: 25 октября 2026 года. **Статус: REVIEW_DONE_M0_NOT_READY.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Выполнены два независимых local suites: статусы совпали, deterministic edge/current fixtures совпали.
- 17 Python unit tests и3 Go tests с-race проходят; Go randomized test включает1000 случаев.
- Подготовлена evidence/gate matrix и обновлены риски; target-platform/HIL/phone checks не объявлены выполненными.
- M0_NOT_READY: отсутствие physical Pi evidence сохраняет hardware dependency. Независимый Core scaffold возможен с явными открытыми gates.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D10-01 | Evidence inventory | PASS-HOST | Case levels and limitations recorded for all planned days |
| D10-02 | Independent local replay | PASS-HOST | Two independent probe process trees, same outcomes/deterministic fixtures |
| D10-03 | Clean Pi install replay | BLOCKED | No Pi/media; not replaced with host process restart |
| D10-04 | Critical physical regression | BLOCKED | Physical gates still lack evidence |
| D10-05 | Risk/ADR/estimate review | DOC-ONLY | Physical parts provisional; independent Core scaffold may proceed, M0 not closed |

## Артефакты

- [milestones/M0_TECHNICAL_BASELINE.md](milestones/M0_TECHNICAL_BASELINE.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
