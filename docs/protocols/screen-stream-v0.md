# Д005. Screen color packet v0

Это project spike contract; pairing/auth/transport на сеть ещё не реализованы. Источник authority — device token binding, не содержимое packet. [Валидатор](../../experiments/day005/screen_probe.py) проверен positive/negative fixtures.

```json
{
  "schema_version": 1,
  "stream_epoch": "synthetic-stream-a",
  "sequence": 1,
  "zone_id": "monitor.backlight",
  "mapping_version": 1,
  "colors": [[255, 0, 0], [255, 0, 0], [0, 255, 0], [0, 255, 0], [0, 0, 255], [0, 0, 255], [255, 255, 0], [255, 255, 0]]
}
```

Probe bound: 8 synthetic LED, максимум1024 colors/32KiB, protocol/mapping version1. Это ограничение эксперимента; production negotiated bounds по реальной geometry не выбраны. Zone и stream epoch проверяются против доверенного server context. Sequence строго возрастает, bool не принимается за int. RGB только целые0…255, ровно3 channels/pixel; extra fields вроде screenshot отклоняются.

Timing, transport heartbeat и Scene metadata будут отдельными typed messages/versioned extension, а не невалидированным arbitrary payload. Reconnect создаёт новый epoch; старые frames отбрасываются. Color queue latest-value bounded, render не ждёт stale frames. Control commands идут отдельно и проходят dispatcher.

Colors не меняют zone.mode, не создают manual override и не включают свет после Master OFF. Renderer принимает stream только при действующем binding, разрешённом render epoch и desired AMBILIGHT. Frames/camera pixels/full screenshot в packet отсутствуют; логирование packet целиком не включать.

На Д005 выполнены synthetic packet/golden tests, не WS/client-server pairing. Rate enforcement, revoked-token close и network loss/fallback станут реализацией Д056–Д070.
