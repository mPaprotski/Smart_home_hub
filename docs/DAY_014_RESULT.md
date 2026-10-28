# Д014. Bounded Event Bus — результат

`internal/events`: envelope v1 по architecture, RAM-only queue/fanout, exact `kind:id` filtering, nonblocking overflow classes, owned payload copies. [Реализованный контракт](architecture/events.md), [host benchmark](research/event-bus-baseline.md).

| Проверка | Результат / evidence |
|---|---|
| D14-01 Envelope/API | PASS-HOST; known types/version/IDs/UTC/nonzero seq/revision/JSON/size, no known raw-media fields |
| D14-02 Delivery/filter/order | PASS-HOST;100 ordered events exact filter, irrelevant resource не доставляется |
| D14-03 Slow consumer | PASS-HOST; slow control invalidated, fast subscriber получает весь batch; latest queue всегда1 |
| D14-04 Overflow | PASS-HOST; control ErrOverflow/Done, trigger drop counter, latest replacement/drop counter |
| D14-05 Cancellation/close | PASS-HOST; concurrent publish/unsubscribe/close и idempotent Close без race/panic |
| D14-06 Isolation | PASS-HOST; мутации caller/одного subscriber не меняют чужие данные |
| D14-07 Race/benchmark | PASS-HOST; product race tests, actual0/1/10/100 slow subscriber benchmark |

Числа:4.547µs/no subscribers,93.067µs/100 slow subscribers на x86 host; не latency Pi/WS SLA. [Сырые benchmark данные](results/days011-015/event-bus-benchmark.txt), [race log](results/days011-015/quality.log). Нет external auth/dedupe, durable audit, replay/exactly-once и browser resync; эти обязанности остаются у будущих gateway/storage/WS модулей. `Close` отменяет subscription явно; lifecycle владеет worker context/cleanup.
