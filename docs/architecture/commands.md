# Контракт команд v0

Основание: §18/19/25–28/49 ТЗ. Связи: [domain](domain-model.md), [ADR-001](adr/001-module-boundaries.md), [ADR-002](adr/002-command-arbitration.md), [сценарии](scenarios.md).

## 1. Доверенная точка входа

```text
Transport/gateway authentication
  → derive principal + source + resource bindings
  → parse/validate command
  → check permission/capability/expected_revision
  → dispatcher admission + conflict arbitration
  → commit desired state/override/idempotency metadata (если persistent operation)
  → domain handler → adapter
  → record applied revision/outcome → publish redacted events
```

Любой UI/Telegram/CLI/external API mutation проходит этот путь. Кнопка и gesture mapping создают команды через внутренний доверенный gateway, ограниченный конкретными operations. Automation/Prediction отправляют proposals, не вызывают Lighting/Audio hardware напрямую. Sensor/Runtime/broker payload не становится trusted action без валидации и mapping/rule.

В Safe Mode локальный emergency gateway заменяет DB-backed auth проверкой OS peer identity по ADR-005; anonymous bypass отсутствует. Auth выполняется **до** получения cached idempotent response и повторно перед фактическим исполнением queued/delayed job: отозванные credentials не действуют. Уже совершённый физический эффект отозвать токеном нельзя, но pending work отменяется. Пользователь не задаёт `source`, actor/scopes, priority или `override_bypass` через body. Для trusted scheduler priority берётся из сохранённого rule, не из внешнего event.

## 2. Internal Command envelope

| Поле | Семантика |
|---|---|
| command_id | Уникальный ID операции; retry использует тот же ID |
| schema_version | Версия контракта, начальная 1 |
| type | Значение из command catalog |
| source | WEB_UI/BUTTON/GESTURE/TELEGRAM/EXTERNAL_API/LOCAL_CLI/AUTOMATION/PREDICTION/SYSTEM; назначает gateway |
| principal | kind/id/user_id и проверенные credential refs; secrets не включать |
| target | kind/id либо явный серверный selector Master OFF |
| payload | Typed fields конкретной команды, никаких arbitrary field paths/shell |
| expected_revision | Обязательна для редактирования persisted entities; для управляющего slider может быть отсутствующей |
| created_at / expires_at | Server UTC timestamps, queue deadline; клиентское время не считается авторитетным |
| correlation_id / causation_id | Связь workflow и trigger; не auth |
| dedupe_key | Scope `(principal, operation, command_id)`; hash normalized request проверяет совпадение payload |

Примеры: [commands.json](examples/commands.json). Это **внутренние** envelopes; обычный REST request передаёт business payload и idempotency key, а trusted metadata достраивает Core. Device-bound token не может выбрать другую зону даже с правильным `lights:write`.

## 3. Catalog и разрешения

| Команда | Target / payload | Проверка прав | Затронутые paths |
|---|---|---|---|
| SetZoneState | zone; mode/color/brightness/effect params/transition | lights:write + zone binding | Поля, явно задаваемые intent; SetColor также STATIC mode, если это указано contract |
| ApplyPreset | preset ID, validated zone entries | lights:write на все entries | Union изменяемых paths всех зон |
| MasterLightsOff | server selector всех зон | lights:write на весь lighting scope | `lighting:*`; снимает live render activation всех зон |
| SetAudioPlayback | PLAYING/PAUSED/STOPPED | audio:control | audio:playback |
| NextTrack / PreviousTrack | active source | audio:control + source capability | audio:playback; non-idempotent effect учитывается отдельно |
| SetVolume / SetMute | absolute volume 0…1 / bool | audio:control | audio:volume / audio:mute |
| SetAudioOutput | available output ID | audio:control + capability | audio:output |
| SetCameraPermissions | explicit gesture/presence permission | admin + privacy capability | Camera permission revision; только direct user context |
| SetPairingMode | kind, enabled, TTL | admin/device pairing permission | Pairing window; не ambient automation |
| SaveRule / SetRuleEnabled | typed rule + expected revision | automation:write + owner | Rule/settings, не немедленный hardware effect |
| AcceptRecommendation / RejectRecommendation | recommendation ID | authenticated owner | Feedback; accept action заново арбитрируется |
| ClearManualOverride | override ID/scope | authenticated owner/admin | Только явная отмена lease; не действие automation |
| DeleteLearningHistory | explicit confirmation + user ID | owner/admin direct interactive context | privacy epoch/observations/feedback/model; не background rule |
| FactoryReset / RestoreBackup | explicit local/admin maintenance flow | admin + separate maintenance gate | STOP writers/jobs до destructive workflow; не обычная rule action |

Имена дополнительных permissions — проектные внутренние capability names, не новые публичные scopes ТЗ. Их точное кодирование/привязка к API tokens — Д017–Д020; broad `system:read` не даёт privacy/reset/write прав. Button/gesture templates по умолчанию допускают только light/audio controls, не reset/restore/auth/config.

## 4. Concurrency и arbitration

Control-plane mutations упорядочиваются одним Core dispatcher с короткой admission section. Внешний I/O/ML не выполняется под глобальной блокировкой. Handlers применяют hardware commands последовательно на соответствующий adapter; глобальный power limiter объединяет latest desired slices перед Show().

Порядок проверок: safety/capability/auth → active override → music protection → rule priority → revision и очередь. Safety limits/thermal critical handling никогда не bypass. Manual commands обрабатываются по server admission order; last accepted absolute value выигрывает на одном path. Два независимых ресурса не конфликтуют.

Для rules одного conflict batch приоритет выше выигрывает. Равный priority: стабильный порядок по `rule_id` по возрастанию. После batch применённое automation intent удерживает автоматическое владение path, пока rule enabled и его guard истинный; более низкий priority не может немедленно перезаписать победителя следующим event. Владение — RAM index, на boot rules переоцениваются; это не manual override.

Rule без state guard, trigger-only action не получает бессрочное владение: его contention window ограничен текущим dispatch batch. Для event-only rules порядок между разными batches определяется admission order. Приоритет не означает вечную блокировку всех последующих действий. Детали burst coalescing/численные queue limits фиксируются до evaluator на Д101.

`allow_music_scene_change=false` по умолчанию: automation, меняющая mode/preset/effect активной music zone, SKIPPED с MUSIC_SCENE_PROTECTED. Явный true снимает только это препятствие, не override/auth/limiter.

## 5. Manual override и live effects

Принятый manual intent создаёт override затронутых fields на **900 секунд** по умолчанию. Repeated manual action на том же scope продлевает TTL от нового admission; invalid/denied request lease не создаёт. Новый scope не продлевает несвязанные leases. Master OFF получает `lighting:*`, чтобы добавленная зона также не включилась автоматически.

Lease фиксируется с принятым intent, до hardware I/O: automation не может обогнать pending manual action. При hardware failure override сохраняет ручное намерение до expiry/explicit clear; UI показывает FAILED/desired≠applied. Retry того же command ID не продлевает TTL. Самостоятельно повторённая команда с новым ID — новое намерение.

Live frames и effect ticks **не являются новыми autonomous control commands**: в режиме, явно выбранном пользователем, они продолжают render. Ручной выбор AMBILIGHT не замораживает цвета на 15 минут. Ручная brightness защищает brightness от ambient/Scene ML corrections в соответствующем поле; остальные разрешённые effect параметры продолжают обновляться. Master OFF устанавливает mode OFF и invalidates render epoch — stream не может включить свет.

Automation/prediction/ambient не создают manual lease. Telegram/external API mutation от привязанного user/token создаёт его для light/audio controls; read, pairing, rule edit, feedback reject — нет. System restore/fallback/thermal — служебные причины с ограниченными полномочиями, не маскируются под manual action.

## 6. Результат и HTTP semantics

| Статус | Значение |
|---|---|
| ACCEPTED | Валидирована/арбитрирована, записана либо принята в RAM Safe Mode; физический эффект ещё не подтверждён |
| RUNNING | Handler начал применение |
| APPLIED | Adapter успешно принял effect; applied_revision обновлён; аппаратная обратная связь может отсутствовать |
| PARTIAL | Multi-action часть выполнена, часть failed/skipped; результат по каждому action |
| FAILED | Известно, что действие не выполнено/отклонено adapter |
| CANCELLED | Rule/permission/override/guard сделали pending действие недопустимым |
| EXPIRED | Deadline истёк до запуска |
| UNKNOWN_OUTCOME | Был crash/timeout после отправки non-idempotent effect, подтвердить результат нельзя |

Невалидная/неавторизованная команда не получает ACCEPTED. HTTP: 200 — completed known result; 202 — accepted operation с operation ID; 400 — malformed, 401/403 — auth/scope, 404 — неизвестный разрешённый resource, 409 — конфликт/dedupe reuse, 412 — expected revision mismatch, 422 — semantic validation, 429 — limit, 503 — dependency/queue/Safe Mode restriction. В operation detail terminal FAILED/PARTIAL не маскируется HTTP 200 успехом hardware.

Response содержит command_id/status/correlation_id/result/error и per-action results, desired/applied revisions при применимости. Errors — code, безопасное message, retryable; без tokens/паролей/raw frames/traceback. Для rejected automation intent писать audit SKIPPED с reason, не ложный APPLIED.

## 7. Idempotency, crash и несколько действий

Dedupe metadata для durable commands сохраняется в SQLite. Проектное окно completed command metadata — **24 часа**, без секретных payload и track history; in-flight/delayed work не удаляется лишь из-за TTL. Окно документировать клиентам; после него non-idempotent request автоматически не повторять. Same ID + иной payload → IDEMPOTENCY_CONFLICT. Retry сначала повторно проверяет доступ, затем возвращает известный outcome без повторного эффекта.

Гарантия exactly-once для внешнего hardware не заявляется. Absolute SetVolume/SetZoneState можно reconcile по desired state; NextTrack/PreviousTrack после uncertain crash → UNKNOWN_OUTCOME без слепого replay. Toggle на входе разрешается в absolute intent под admission lock, но при uncertain NextTrack результат неизвестен. Hardware state restore использует desired logical settings, не replay журнала.

Multi-action: до admission проверить весь payload/права/ссылки. Known validation/auth errors отклоняют весь сценарий без effects. Runtime hardware отказ может дать PARTIAL, физической общей transaction нет. В sequential action list после hard failure прекращать оставшиеся actions по default stop-on-error; уже выполненные не компенсировать автоматически. Rule policy/skipped conflict отражается по action; для ApplyPreset, если target paths пересекают active override, весь preset SKIPPED, чтобы не применить неожиданную половину. Общая priority/override проверка происходит снова перед effect.

## 8. Delayed work

DelayedJob хранит rule revision, trigger, due time и guard. Перед исполнением: rule enabled/unchanged, caller/owner разрешён, deadline не истёк, guard TRUE, нет override/music conflict. AWAY→wait→OFF обязательно revalidates AWAY; UNKNOWN тоже не TRUE.

Для absence delay guard включает presence, а для event-only action guard явно задан template. Изменение/удаление rule отменяет старые jobs. После restart будущее work восстанавливается без дубликата по job ID; просроченные jobs по проектному default **EXPIRED, без catch-up**. Регулярный schedule также не воспроизводит пропущенные runs; следующий run вычисляется по timezone. При CLOCK_UNCERTAIN time schedules/habit AUTO не запускаются, override сохраняется консервативно. Эти политики проверяются fake clock и реальным restart в Д103–Д105.
