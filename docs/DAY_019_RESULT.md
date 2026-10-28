# Д019 — browser sessions, CSRF, HTTPS и WS lifecycle

Выполнен software slice 28.10.2026, simulation на x86. Target — Pi4B **2 GB RAM**.

- Migration004 расширяет sessions, не меняет migrations1–3. Raw cookie не сохраняется;
  random256-bit cookie/hash, CSRF, idle30min, absolute12h, cap5, persistent revoke.
- В явном HTTPS profile login200 выдаёт __Host cookie. `/me`, `/csrf`, logout,
  revoke-all работают. Same-origin/CSRF обязательны для cookie mutations.
- TLS listener и healthcheck с отдельным CA; точечный trusted proxy /32,/128.
  Untrusted XFP/ambiguous credentials не повышают доверие.
- Timezone gateway получает identity/source от сервера, повторно проверяет credential
  внутри immediate transaction до dedupe/CAS. Missing revision отвергается.
- Настоящий WSS upgrade, caps16/2, no query credentials, periodic expiry/revoke close,
  shutdown cleanup. Только auth-only greeting; Dashboard events ещё отсутствуют.

D19-01–09: локальные checks пройдены. Evidence: [общий отчёт](DAYS_019_021_RESULT.md),
[security policy](security/sessions-tokens.md),
[HTTPS/WS tests](../internal/api/session_test.go),
[credential tests](../internal/auth/access_test.go).
HTTP(S) proxy проверен настоящим TLS frontend → HTTP backend. Проверка Secure
cookie jar выполняется также с обычным hostname: Go1.27 loopback считает secure
context даже по HTTP, поэтому localhost-only assertion был бы некорректен.

Pi/production HTTPS, мобильные клиенты и remote CI не проверены. Readiness issuer
не означает, что администратор уже создан: bootstrap остаётся отдельным local CLI.
