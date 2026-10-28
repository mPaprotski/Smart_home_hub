# Д009. Локальный HTTPS/mDNS bootstrap

**Выполнено на host:** isolated loopback HTTPS reverse proxy→loopback status backend, ephemeral CA/SAN, valid trust, negative CA/hostname tests, запрет control routes/POST и замена leaf certificate. [Код](../../experiments/day009/https_probe.py), [evidence](../../experiments/results/run-1.json).

**Не выполнено:** Pi/LAN mDNS, Caddy install/config validation, реальный phone/laptop trust store, порт443 production proxy, reboot services. Current host trust store и system services не изменялись; private keys удалены с temporary directory.

## Целевая конфигурация-кандидат

[Caddyfile](../../configs/Caddyfile.example), [Avahi daemon](../../configs/avahi-daemon.conf.example), [HTTPS service advertisement](../../configs/smarthome.service.example). Это examples для проверки на target, не уже применённые host files. Caddy выбран кандидатом по встроенному local HTTPS; Nginx+external CA остаётся альтернативой при actual incompatibility.

Caddy docs описывают local CA, но [доверие на других устройствах](https://caddyserver.com/docs/automatic-https#local-https) надо настроить отдельно. Установка доверия Pi не делает телефон автоматически доверенным. Template reverse-proxies internal port8080; до Core там может быть только harmless loopback status stub без user data/actions.

## Порядок на реальной Pi

1. Убедиться, что smarthome.local не принадлежит другому устройству; проверить маршрутизацию/guest isolation/multicast LAN. Не подменять ошибку identity отключением TLS verification.
2. Установить выбранные поддерживаемые proxy/mDNS packages, зафиксировать versions/config paths/service users. Не копировать template поверх чужого рабочего config без diff/rollback.
3. Настроить publisher hostname и `_https._tcp` service443. DNS-SD advertisement сам по себе не создаёт произвольный hostname alias; hostname/publisher должны действительно разрешаться.
4. Запустить loopback-only backend, затем proxy. Проверить routing marker, отсутствие env/key/DB/backups в web root и недоступность backend из LAN.
5. Выпустить local certificate с SAN smarthome.local и валидной chain/time. Proxy получает только свои keys/storage; bot/frontend не получают CA key.
6. Сверить root certificate fingerprint доверенным способом. На phone/laptop установить только public root trust штатным механизмом конкретной ОС/браузера; записать шаги и права. Root private key остаётся у CA owner.
7. Открыть обычным браузером без bypass warnings и TLS-клиентом с правильным trust. Проверить clean untrusted client и wrong hostname ожидаемым отказом.
8. Проверить IPv4/IPv6/firewall exposure и HTTP redirect/refusal. Локальный API не публиковать напрямую в WAN. Remote VPN в эту установку автоматически не входит.
9. Заменить server certificate, reload, проверить новый serial/chain без переустановки неизменного root. Замена/компрометация CA — другой runbook с новым trust и удалением старого доверия на управляемых clients.
10. Restart proxy/mDNS/Pi и повторить разрешение имени/TLS marker. Полный auth/cookies/WS проверяется в следующих этапах, а не этим status stub.

## Диагностика

IP reachability→mDNS resolution→правильный адрес/порт→TLS time/name/chain→proxy upstream→Core/auth. Не начинать с `insecure_skip_verify`/`curl -k`. Сырые private keys/host addresses не помещать в public report.

D09-01/02/05 target cases остаются BLOCKED до Pi/LAN/mobile. Host D09-03/04/06 crypto/isolation/replacement checks выполнены только в ephemeral environment. Работа Caddy не выводится из успешности Python TLS proxy.
