# Д017. Dispatcher и permissions — результат

Рабочий slice: SetUserSetting(timezone) через trusted internal command, DB-backed
rights/ownership, CAS и persistent dedupe. Lighting policy port проверен отдельно.
[Матрица и boundaries](security/permission-matrix.md).

| План | Результат и точный scope |
|---|---|
| D17-01 | PASS-HOST: metadata только internal; spoof headers не открывают HTTP command route |
| D17-02 | PASS-HOST: две user rows, own/foreign owner, exact service binding/scopes |
| D17-03 | PASS-HOST: unknown/unsupported type/source/scope отклоняются до mutation |
| D17-04 | PASS-HOST: invalid value/version/deadline/CAS не меняют state |
| D17-05 | PASS-HOST: concurrent CAS/duplicates; slow external adapter N/A — поддержанный handler не делает I/O |
| D17-06 | PASS-HOST: same request retry без повторной revision; different payload conflict; lease retry не продлевает TTL |
| D17-07 | PASS-HOST: manual scoped lease900s, autonomous overlap blocked, independent path/expiry |
| D17-08 | PASS-HOST для transaction rollback/no false APPLIED; physical UNKNOWN_OUTCOME/replay N/A — physical commands пока unsupported |
| D17-09 | PASS-HOST: роли проверяются перед cached result; unsupported batch отклонён целиком. Pending-job cancellation N/A — очереди/jobs нет |
| D17-10 | PASS-HOST/SIM: command HTTP gateway не смонтирован; anonymous controls404 |

Effect hardware/command event publisher/full priority engine/scene protection/queued
work/ApplyPreset не объявлены реализованными. Для текущей синхронной операции APPLIED
означает SQLite setting commit, не LED success. Это граница Д017 до domain modules,
а не завершение Automation/Lighting этапа.
