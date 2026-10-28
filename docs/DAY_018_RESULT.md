# Д018. Bootstrap и credentials — результат

Локальный `hub admin bootstrap` с no-echo TTY/explicit bounded stdin, random-salted
Argon2id, atomic persistent marker, CLI verify, AuthService и stage login API.
[Инструкция и limits](security/admin-bootstrap.md).

| План | Результат |
|---|---|
| D18-01 | PASS-HOST: random salts, valid/invalid verification, no plaintext DB/logs |
| D18-02 | PASS-HOST: PHC algorithm/version/base64/params/input-size bounds до hashing |
| D18-03 | PASS-HOST/COMPOSE: bootstrap once, concurrent repository bootstrap, deletion не сбрасывает marker |
| D18-04 | PASS-HOST: unknown/disabled/malformed hash/invalid username unified INVALID_CREDENTIALS |
| D18-05 | PASS-HOST/SIM: account/IP/global bounds/expiry, persisted cooldown после restart, XFF игнорируется |
| D18-06 | PASS-HOST/COMPOSE: Argon benchmark,128MiB no-swap container/flood/peak; Pi latency BLOCKED |
| D18-07 | PASS-HOST/SIM: valid503/no cookie, invalid401, exhausted429; no fake session |
| D18-08 | PASS-HOST/COMPOSE: admin persistence/recreate, failed bootstrap rollback |
| D18-09 | PASS-HOST/SIM: synthetic secret absent в logical DB/log/evidence, PHC не публикуется |
| D18-10 | PASS-HOST/SIM: auth-required ready503; production/anonymous control routes закрыты |

Sessions/CSRF/logout/revoke-all и positive login200 — Д019. Real passwords вводятся
через local TTY, HTTP tests используют synthetic data. Security audit/tokens — Д020.
[Общий отчёт и evidence](DAYS_016_018_RESULT.md). Target — толькоPi4B/2GB; модель2GB
сама по себе не подтверждает производительность будущего Python runtime.
