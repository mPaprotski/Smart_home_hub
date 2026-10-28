# ADR-004. Readiness, degraded и release gate

Дата: 25.10.2026. Статус: **принята проектная интерпретация согласования §36 и §38; HTTP endpoints не реализованы**.

Источник: §36/38/44/47 ТЗ. Требования: OBS-004/OBS-005, REL-001…REL-007, DEPLOY-005/DEPLOY-006. Вопрос: Q-031.

## Контекст

ТЗ требует проверки SQLite/MQTT/Runtime/adapters в `/ready`, но допускает продолжение основных функций без ML/MQTT. Простая конъюнкция health всех сервисов делает допустимую degradation полным outage; игнорирование ML может пропустить broken release.

## Решение

`/health` сообщает liveness собственного control loop. `/ready` проверяет **все перечисленные зависимости** и возвращает diagnostics, но HTTP ready вычисляет по critical set текущего установленного profile. Critical: Core command/storage/schema и configured hardware adapters, необходимые базовому управлению этого profile. Simulation critical adapters — fake. ML/MQTT/Telegram/Spotify/camera не блокируют baseline manual ready, их отказ отражается degraded.

SQLite fail → SAFE_MODE, `/ready=503`, `/health=200`, если emergency loop работает. Critical hardware fail также normal ready false. Disabled/unconfigured optional module не называется healthy: state DISABLED/NOT_CONFIGURED и причина.

Production deploy дополнительно проверяет versions и smoke required services, перечисленных release manifest. В full profile первая версия включает Runtime/broker/bot/frontend; их неправильная установка не принимается по одному core 200. Временная недоступность внешнего Telegram/Spotify upstream не равна broken local service image: smoke различает local startup/credentials и upstream reachability.

Policy profile выбирается в trusted config/release manifest, не HTTP request. Public health response минимальный, подробности зависимостей доступны авторизованному System UI/deploy verifier. Никаких secrets/private addresses в unauthenticated diagnostics.

## Альтернативы

Все dependencies critical — противоречит §36. Всегда ready при живом HTTP — скрывает БД/driver failure. Удалить ML/MQTT из checks — не выполняет §38. Выбранная policy сохраняет checks и различает operational readiness/deployment success.

## Последствия и проверка

Health model имеет lifecycle, ready, degraded и reasons отдельно. [Примеры](../examples/readiness.json) включают normal/degraded/Safe Mode. Д013/Д126/Д128/Д142–Д145 проверяют отключение каждого сервиса, критический hardware failure, broken candidate и auto rollback. Если при реализации выяснится, что release system ожидает иной contract, изменить ADR и acceptance, а не молча менять HTTP semantics.
