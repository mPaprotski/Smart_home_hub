# Д021 — domain storage, privacy и retention

Выполнен software slice 28.10.2026. Target Pi4B **2 GB RAM**; host x86 verification.

- Migration006: presets/entries, disabled typed rules, gesture/button mappings,
  automation audit, learning state/epoch, habit observations, model metadata,
  typed recommendations и immutable feedback. Foreign keys и owner/revision checks.
- Habit schema содержит volume/presence/ambient/monitor/timezone/origin/quality,
  nullable measurements сохраняют NULL. Generic events/media/track/DSP history нет.
- Current-epoch guards отклоняют stale worker publication. Same feedback retry
  idempotent, другая decision conflict, чужой owner запрещён.
- Retention90d habits/30d audits UTC, bounded cleanup; exact cutoff сохраняется.
  Read-side cutoff скрывает expired habits до удаления. Model aging удаляет старые
  recs/feedback и возвращает insufficient; collector/training ещё не включены.
- Настоящая DB v3 обновляется до7: user/settings rev7/bootstrap сохраняются,
  legacy session без CSRF отклоняется. Migration integrity/failure rollback сохранены.
- Fixture100000 rows возвращает только100, unbounded limit запрещён; time query
  использует covering habit_owner_time. Измерение host — в evidence/dataset.log.

D21-01–08: локальные checks пройдены. [Словарь схемы и ER](storage/data-dictionary.md),
[общий результат](DAYS_019_021_RESULT.md).
Нет predictor/collector/automation execution/model files/полного privacy delete.
Epoch guards не заменяют backup erasure. Физическая RAM полного проекта на Pi
и контроль memory-pressure остаются будущими проверками.
