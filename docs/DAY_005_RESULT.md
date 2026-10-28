# День 5 — Screen/ML/measurement feasibility

Дата: 25 октября 2026 года. **Статус: PARTIAL.**

Основание: [план Д004–Д010](plans/DAYS_004_010_DETAILED.md), [архитектура](architecture/README.md). [Сводный результат](DAYS_004_010_RESULT.md).

## Фактически выполнено и границы

- Проверены synthetic four-edge weighted averages, corner exclusion и packet rejection cases.
- Попытка live capture в RAM завершилась BLOCKED/OSError; Windows и оба macOS targets отсутствуют.
- Separate local Python demo-worker дал2 успешных inference, затем intentional failure; synthetic color loop продолжился: 66 frames за~2.2sec.
- Это fixed logistic demo, не trained nine-label Scene model. ONNX runtime не установлен/не тестировался.
- Подготовлены protocol, dataset/quality/performance методики; качества реальных Scene/Gesture/Habit models не измерено.

## Проверки

| ID | Проверка | Статус | Evidence / ограничение |
|---|---|---|---|
| D05-01 | Four-target matrix | DOC-ONLY | Linux current display unavailable; Windows/macOS targets not supplied |
| D05-02 | Synthetic analyzer | PASS-SIM | Weighted edge fixtures and strict packet validator executed |
| D05-03 | Live capture / cleanup | BLOCKED | Pillow capture returns OSError; no screenshot file created |
| D05-04 | Inference separation | PASS-HOST | Tiny synthetic Python worker; bounded queue, <=1Hz and worker failure fallback |
| D05-05 | Performance method | DOC-ONLY | All8 target measurement methods and reference profiles specified |
| D05-06 | Dataset/quality policy | DOC-ONLY | Synthetic/licensed policy, split, pilot/freeze and provisional gates; actual model quality NOT_RUN |

## Артефакты

- [research/screen-platform-matrix.md](research/screen-platform-matrix.md)
- [protocols/screen-stream-v0.md](protocols/screen-stream-v0.md)
- [research/local-inference-options.md](research/local-inference-options.md)
- [research/performance-and-quality-plan.md](research/performance-and-quality-plan.md)
- [research/dataset-policy-v0.md](research/dataset-policy-v0.md)
- [Runnable experiments](../experiments/README.md)
- [Run1 JSON](../experiments/results/run-1.json) и [Run2 JSON](../experiments/results/run-2.json)
- [Replay summary](../experiments/results/replay.json)

## Следующий шаг

Для physical/target cases нужен настроенный Pi4 и перечисленный в HIL protocol стенд. Повторять только реально затронутые проверки; host PASS не превращать в target PASS. Отсутствующие размеры/компоненты остаются неизвестными.
