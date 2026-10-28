# Д020 — scoped tokens и durable security audit

Выполнен software slice 28.10.2026. Pi4B **2 GB RAM**, fake hardware.

- Migration005: hash-only api_tokens, role seeds, security_audit и индексы.
- Browser-session management API: secret один раз в201; subsequent list metadata-only.
  Cap16, expiry90d, narrow scopes/settings и own-user binding; без delegation.
- Bearer-only supported action, expiry/revoke/disabled owner/current role checks;
  cookie+bearer/duplicate headers запрещены. Bounded rate maps и connection caps.
- Mutation+success audit атомарны. Trigger injection подтверждает rollback при отказе
  audit; denials агрегируются, счётчики сигнализируют отказ записи/notification.
- SQL audit переживает закрытый Event Bus. Notification не содержит payload/secrets.
  Revocation-before-admission и cached retry после revoke не обходят проверку.
- Native/Compose TLS restart сохраняют сессию/токен; logout/revoke закрывают доступ.

D20-01–09: локальные checks пройдены. [Общий результат](DAYS_019_021_RESULT.md),
[policy/API boundaries](security/sessions-tokens.md), [OpenAPI](api/openapi.yaml).
Scopes пока только settings:read/write и binding user:owner. Token screen pairing,
MQTT/Telegram authorization и физические команды этим этапом не реализованы.
