# Д005. Performance и quality gates v0

Это принятая методика будущих измерений; host feasibility numbers не являются приёмкой §46. [Контракты](../architecture/README.md), [реальные host результаты](../../experiments/results/run-1.json).

## Reference profile перед каждым benchmark

Записать board/OS/kernel/build, adapter/model versions, LED counts/layout, capture resolution/rate/pixel format, LAN path/jitter, active services/workers, temperature/fan, sample size и warmup. Неизвестные параметры не подменять вариантом120LED из ТЗ. Separate profiles: isolated microbenchmark, короткий LED+audio spike, финальный full joint load.

| Target ТЗ | Точка измерения / метод | Gate |
|---|---|---|
| REST p95<100ms | Client send→полный response, warmup30sec + измерение120sec; проектный initial profile5 concurrent clients/20req/sec обычных reads/writes | p95<100ms, errors отдельно; restore/backup long ops отдельно |
| WS event<100ms | Publish→client receive; одна опора времени/оценённый clock offset, event IDs | latency<100ms на согласованном profile, drops/overflow отдельно |
| LED>=30FPS | Фактически отправленные/applied frames, не генерация; run≥60sec | Средний rate>=30, interval p95/max/drops публикуются |
| Ambilight30FPS | Capture→network→driver applied cadence при target30 | Rate/drops на выбранной сети; synthetic loop не evidence |
| Ambilight e2e<100ms | Generated screen transition→physical LED transition; визуальная/high-speed или общая измерительная опора | Все валидные контрольные transitions<100ms с указанной погрешностью; RTT недостаточно |
| Gesture reaction<500ms | Начало устойчивого gesture→effect; confirmation включён, camera/detection/queue отдельно | Все проверяемые reference gestures<500ms; распределение/провалы отдельно |
| Scene<=1 inference/sec | Actual request/start timestamps и worker counters | Не чаще1Hz, colors не блокируются |
| Presence2–3sec | Actual person inference intervals при shared camera/CPU | Intervals2–3sec, starvation/error явно учитываются |

Нагрузочные числа выше — проектный baseline методики, не новые требования ТЗ. Перед final acceptance уточнить с реальным room/reference profile; изменение зафиксировать до прогона, не после неудачного результата. Перцентиль не подменяет ограничение там, где ТЗ не задаёт percentile.

## Dataset и ML thresholds до обучения

Приняты **проектные стартовые gates**, их пригодность надо проверить на pilot annotation до freeze holdout. Нельзя сообщать достигнутое качество до реальной независимой оценки.

- Scene: все9 labels, per-label precision/recall/F1, macro/micro; стартовая цель macro-F1>=0.80 и каждый класс F1>=0.70. Сравнить с эвристическим baseline на том же holdout; label rubric/thresholds freeze до training/tuning. При слишком малом/перекошенном наборе результат INCONCLUSIVE, не PASS.
- Gesture: все10 actions плюс no-hand/negative motion, varying light/distance/position. Стартовый gate precision>=0.95/recall>=0.85 по каждому action, false actions<=1 на5 минут negative sequence при принятой activation policy. Timing<500ms считается отдельно. Activation gesture/window и population coverage ещё требуют pilot.
- Habits: temporal split, target-wise precision/calibration, frequency baseline, eligibility>=30 реальных дней и достаточная data quality. Confidence85% не равен accuracy85%; AUTO до калибровки/explicit opt-in не разрешается. Численные data-sufficiency thresholds нужны до Д115.

Если pilot показывает, что threshold/definition непригоден, изменить **версию методики и holdout policy до финального обучения**, сохранив причину. Не тюнить по test и не снижать threshold задним числом ради PASS.

## Coverage

Core scope: dispatcher/arbitration, lighting mapping/limiter, rules/schedules/delays, presence fusion, auth business policy, storage transaction/retention, prediction eligibility/feedback. Generated/vendor/native SDK исключены явно. Go/Python/TS reports раздельны; overall при line-based aggregation=sum covered lines/sum eligible lines, а не средний процент языков. Цели ТЗ80% core/70% overall проверяются Д146, не текущими spike tests.

## Выполнено сейчас

Synthetic worker requests<=1Hz и independent66 frames за~2.2sec действительно проверены. Packet/golden/thermal/limiter checks выполнены. Реальные REST/WS/Core/Ambilight/gesture/presence targets и model quality не измерены: продукта/hardware для этого ещё нет. Статусы в reports остаются NOT_RUN/BLOCKED.
