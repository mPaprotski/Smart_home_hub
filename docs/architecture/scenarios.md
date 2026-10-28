# Проверка архитектурных контрактов сценариями

Дата: 25 октября 2026 года. **Метод: последовательный разбор контрактов на бумаге**, не выполнение продукта/симулятора/HIL. Результат каждой строки — ожидаемое поведение и будущая проверка. Все product-test statuses остаются NOT_RUN.

Связи: [commands](commands.md), [events](events.md), [state machines](state-machines.md), [data model](data-model-v0.md), [ADR index](README.md).

## Сценарии конфликтов

### S01. Master OFF, затем правило включения

**До:** обе зоны включены, enabled rule пытается включить свет при PRESENT. **Шаги:** user MasterLightsOff; rule proposal приходит после admission. **Ожидание:** desired OFF всех зон, invalidate render epochs, wildcard lease 900 секунд; automation SKIPPED/OVERRIDE_ACTIVE, audit содержит причину. На expiry свет сам не включается без нового допустимого trigger/re-evaluation. Новая зона в течение lease тоже не включается автоматически. Проверка Д032/Д108; ADR-002.

### S02. Новый manual OFF и retry

**Шаги:** OFF ID=A в t0; retry A в t+30; новое OFF ID=B в t+60. **Ожидание:** retry A возвращает outcome, не продлевает lease и не повторяет effect. B — новое намерение и expiry t+960. Auth/revoke проверяется даже для cached response. Проверка Д020/Д032; ADR-002.

### S03. Gesture и UI одновременно меняют громкость

**До:** volume 0.3. **Шаги:** trusted gesture mapper получает rotation, формирует intent, UI присылает absolute 0.5. **Ожидание:** dispatcher упорядочивает изменение, relative action вычисляется под control serialization, adapter применяет в порядке admission. Нет lost update из предварительного чтения старого volume. Последний accepted intent задаёт итог, ID обоих commands различны; scopes/cooldown проверены. Проверка Д086/Д088; ADR-001/002.

### S04. Ручная brightness одной зоны

**Шаги:** user ставит monitor brightness=0.3; ambient и scene proposals пытаются изменить brightness; desk rule меняет desk brightness. **Ожидание:** monitor brightness protected, rendered colors AMBILIGHT продолжаются, разрешённый smoothing может меняться, desk независима. Safety limiter остаётся активным и может снизить effective brightness ниже desired. Проверка Д095/Д108; ADR-002/007.

### S05. Возврат пользователя до delayed OFF

**До:** rule AWAY→wait 5min→OFF. **Шаги:** на t+4min приходит PRESENT или UNKNOWN; срок timer наступает. **Ожидание:** guard не TRUE, job CANCELLED, свет не выключается. При PRESENT→AWAY заново создаётся новый trigger/job с новым сроком, старый не resurrect. Проверка Д105/Д109; ADR-002.

### S06. Конфликт правил и равный priority

**Шаги:** state-guard rules A priority20/B priority10 одновременно true на zone mode; затем новое событие B при всё ещё true A. **Ожидание:** A остаётся владельцем, B не перезаписывает его. При equal priority выбирается stable rule_id; если A false/disabled, ownership освобождается. Event-only rules не получают вечную блокировку и конкурируют в своём batch. Проверка Д104/Д110; ADR-002.

### S07. Multi-zone preset и partial hardware failure

**Вариант 1:** одна entry нарушает capability/revision/auth — весь запрос rejected до эффектов. **Вариант 2:** active override одной зоны — preset целиком SKIPPED. **Вариант 3:** после admission первая зона applied, вторая driver failed — PARTIAL с per-action outcome, default stop-on-error; нет ложного global success и автоматического физического rollback. Проверка Д031/Д104; ADR-001/002.

## Камера, realtime и перезапуск

### S08. Presence уже использует камеру, включается Gesture Mode

**До:** current-boot presence consent и lease, одна capture session/indicator ON. **Шаги:** user включает gesture permission/mode. **Ожидание:** добавляется consumer той же session; второй hardware open не выполняется. Gesture confirmation/activation window независимы от person inference каждые 2–3sec. Отключение gesture не выключает presence capture, indicator остаётся ON. Проверка Д081/Д099; ADR-003.

### S09. Indicator/capture failure и restart

**Шаги:** known indicator write failure перед start; затем crash/restart Runtime/Core с прежними saved preferences. **Ожидание:** capture не начинается при known indicator failure; cleanup/lease epochs invalidated. После нового Core boot camera OFF, current-boot consent отсутствует, restart preferences не запускают её. Перегоревший LED без feedback не может быть автоматически обнаружен — для этого нужна физическая HIL проверка, а не выдуманная telemetry. Проверка Д082/Д090; ADR-003.

### S10. Screen stream возвращается после Master OFF

**До:** desired AMBILIGHT, затем stream loss→fallback. **Шаги:** user Master OFF, клиент reconnect и присылает fresh/old frames. **Ожидание:** desired OFF; render epoch/lease invalidate, frames не меняют mode. После 15min тоже нет autonomous frame-driven включения. Новый user Select AMBILIGHT создаёт разрешённый render epoch. Проверка Д065/Д069/Д108; ADR-002/007.

### S11. Shutdown fade и восстановление state

**До:** desk RAINBOW/brightness0.3, monitor desired AMBILIGHT. **Шаги:** graceful shutdown, boot без screen-client. **Ожидание:** shutdown fade не пишет desired OFF; desk восстанавливает RAINBOW0.3, monitor временно fallback с desired AMBILIGHT, music STOPPED, camera OFF, engine enabled с сохранением disabled rules/override. Проверка Д127; ADR-002/003/007.

## Отказы и безопасность

### S12. ML Runtime упал во время Ambilight

**Шаги:** сначала остановить Pi Runtime, затем отдельно laptop Scene ML. **Ожидание:** static/manual lighting, playback, обычный screen colors и non-ML rules доступны; gestures/CV/DSP degraded. Laptop Scene ML fail отключает adaptation без остановки colors. `/ready` baseline 200 при исправных critical dependencies, diagnostics degraded. При проверке нового release broken required Runtime smoke делает deploy failed, несмотря на core 200. Проверка Д075/Д126/Д142; ADR-001/004.

### S13. SQLite недоступна, пользователь меняет свет

**Шаги:** DB failure при живых adapters; запрос старой LAN session и local maintenance CLI. **Ожидание:** `/ready=503`, network DB-backed auth не работает по stale cache, automation/persistence выключены. Разрешённый local peer выполняет manual lighting/audio в RAM, `persistent=false`; другой OS user denied. Cold boot не читает presets из broken DB. Проверка Д128; ADR-005.

### S14. Recovery после emergency Master OFF

**До:** safe CLI выключил свет, backup содержит включённый preset. **Шаги:** operator восстанавливает БД. **Ожидание:** explicit RECOVERING/reconcile, старый preset не включается сам; выбор сохранить emergency state или вернуть backup явный, default Master OFF hold остаётся. Credentials revalidated, generation/epoch обновлены до automation re-enable. Проверка Д128/Д129; ADR-005/006.

### S15. Token revoke, source spoof и replay

**Шаги:** external token присылает `source=button`, просит другую zone; затем genuine bound token revoked при открытом WS/stream; затем повторяет старый command ID. **Ожидание:** body source не доверяется, binding/scopes проверены; active stream/WS закрываются, epoch invalidates; повтор denied даже при существующем cached outcome. Password/token secrets не появляются в events/audit. Проверка Д020/Д056/Д069; ADR-001/002.

### S16. Crash после NextTrack

**Шаги:** adapter отправил NextTrack, процесс упал до известного outcome. **Ожидание:** после boot UNKNOWN_OUTCOME, автоматического повторного NextTrack нет. Для absolute brightness/volume допускается reconcile desired state; это не exactly-once гарантия. Проверка Д039/Д127; ADR-002.

## Данные и транспорт

### S17. Две зоны, одна SPI chain

**До:** virtual layout monitor120+desk120, counts чисто синтетические. **Шаги:** monitor OFF, desk MUSIC_BEAT; сформировать frame, затем обе зоны full white. **Ожидание:** monitor indices0–119 чёрные, desk120–239 имеет effect; global limiter учитывает обе зоны и idle. Изменение monitor не переключает desk mode. Обрыв/обесточивание monitor может нарушить desk DATA — physical independence не обещана, HIL spike остаётся открытым. Проверка Д026/Д028/Д008; ADR-007.

### S18. Delete habits одновременно с training/restore

**До:** training captured privacy_epoch=7, managed backup epoch7. **Шаги:** user Delete→epoch8, worker позже завершает epoch7; затем restore старого backup. **Ожидание:** old model publish отвергнут, DB/artifacts/caches/managed old backup очищены, новый started_at. Local tombstone8 переживает normal restore, старая history удаляется до normal mode; внешние copies без доступа не объявлены удалёнными. Проверка Д118/Д129; ADR-006.

### S19. WS consumer отстал и потерял state update

**Шаги:** slow phone browser вызывает bounded buffer overflow между initial snapshot и событиями. **Ожидание:** server RESYNC_REQUIRED, новый snapshot+watermark, old subscription epoch discard. Render не ждёт браузер, command outcome доступен query, обязательный audit уже записан владельцем. Проверка Д014/Д023/Д147; ADR-001.

### S20. Неверное время после boot и missed schedule

**До:** active override и delayed job сохранялись до restart. **Шаги:** boot с недостоверным UTC, затем корректное время позже due. **Ожидание:** manual control доступен, affected scopes conservatively protected; time schedules/habit AUTO paused. После надёжного времени lease remaining оценивается по deadline; overdue job EXPIRED без catch-up, schedule ждёт следующий occurrence. Проверка Д103/Д105; ADR-002.

## Итог разбора

Во всех 20 сценариях определены инициатор, policy, state/outcome и следующий уровень доказательства. При разборе уточнены: data-plane delegation для Ambilight, защитный scope Master OFF, сохранение desired state при shutdown, stale UNKNOWN guard, граница exactly-once и OS-auth Safe Mode. Программный/hardware PASS ни одному сценарию не присвоен.

Следующие реализации должны превратить соответствующие сценарии в meaningful unit/integration/simulation/HIL checks. Нельзя считать эти текстовые сценарии уже выполненными тестами и включать их в coverage.
