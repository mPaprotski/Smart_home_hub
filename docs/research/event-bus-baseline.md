# Event Bus baseline Д014

Go1.27.3, linux/amd64, Intel Core i5-8250U, developer host, 25.10.2026 UTC. Не Raspberry Pi и не доказательство realtime SLA. [Сырые результаты](../results/days011-015/event-bus-benchmark.txt).

Команда: `make bench` (на этом host `make GO=/snap/go/current/bin/go bench`). Benchmark включает envelope/JSON validation, payload copy и неблокирующие slow subscribers, latest slot1. Никаких subscriber-draining workers/сетевых клиентов; это намеренно worst-case backpressure для bounded latest, а не e2e delivery latency.

| Subscribers | ns/op | B/op | allocs/op |
|---|---:|---:|---:|
| 0 | 4547 | 456 | 6 |
| 1 slow latest | 5907 | 488 | 8 |
| 10 slow latest | 14025 | 776 | 26 |
| 100 slow latest | 93067 | 3662 | 205 |

Default observer control queue64, max subscriptions128, max payload16KiB. Эти caps ограничивают память, а не делают bus hard realtime: в худшем разрешённом config объём существенно выше default. `New`/Subscribe выполняются при trusted wiring; per-resource latest и max subscriptions не дают неограниченный resource map. Memory store ограничен5 observations, LED fake — latest frame/counter. После restart нет replay.

Политики: Control overflow invalidates subscription с ErrOverflow и требует future snapshot resync; Trigger drop increments counter; Latest вытесняет старое значение только exact `kind:id`, slot1. Publisher не вызывает consumer callback и не создаёт goroutine на message; mutex охватывает bounded validation/copies/fanout. Каждый producer обязан сериализовать sequence/revision сам. Bus не dedupe/auth gateway, не durable audit.

Tests подтверждают isolation caller→consumer и consumer→consumer, finite queues, exact filtering, per-publisher order, independent fast subscriber при slow consumer, concurrent publish/unsubscribe/close, repeated close и observable overflow. Race detector запускался на продуктовых packages. External validation/schema evolution/WS resync и нагрузка полного Lighting Engine появятся с соответствующими модулями.
