# ADR-003. Общая камера и privacy lifecycle

Дата: 25.10.2026. Статус: **принято как проектное решение; hardware indicator/camera stack не проверены**.

Источник: §13.2/13.5/15/36/40 ТЗ. Требования: GEST-001…GEST-009, PRES-003…PRES-006, REL-011, PRIV-003. Вопрос: Q-017, часть Q-016/Q-018/Q-021.

## Контекст

Gesture Mode должен управлять gesture capture, presence использует ту же камеру. Раздельное открытие устройства конфликтует, а persisted ON может незаметно включить камеру после restart. Простой GPIO indicator не имеет физической обратной связи.

## Решение

Camera Manager — один capture owner внутри локального Runtime с command-authority Core. Consumers gesture/presence получают bounded leases текущего process/boot epoch. Raw frames остаются в Runtime RAM, Core получает только validated features/person yes-no/gestures.

`gesture_allowed` и `cv_presence_allowed` независимы. CV presence может использовать камеру при выключенном Gesture Mode **только после отдельного явного разрешения presence**. После boot capture OFF и нет armed consumers, независимо от прошлых preferences. Пользователь повторно разрешает текущий boot; тогда presence sampling является законной периодической потребностью. Revocation немедленно отзывает leases/epoch и pending mapped actions.

Gesture Mode разрешает capture, activation gesture затем открывает ограниченное окно действий. Эти два шага не смешиваются: камера не способна обнаруживать activation gesture, пока capture выключен.

Indicator commanded ON перед capture; OFF только после закрытия последнего consumer и освобождения buffers. Known GPIO/indicator I/O error блокирует start/прерывает capture. Без physical feedback нельзя обещать обнаружение перегоревшего LED; visibility проверяется HIL. В состоянии хранить commanded/unknown, не фиктивный verified physical ON. Crash/lease expiry/watchdog поведение проверяется отдельно; finally не считается гарантией после SIGKILL.

## Альтернативы

Камера постоянно ON ради presence — не подходит privacy/restart policy. Два camera handles — риск busy/дублированных ресурсов. Передача raw frames в Core/UI для удобства — нарушает выбранную границу privacy. Автоматическое восстановление persisted camera ON — отклонено в пользу §36 camera OFF after restart.

## Последствия и проверка

Нужны current-boot arm и UI индикация двух permissions; второй consumer не включает вторую физическую камеру. Q-017 решён на уровне semantics, но остаётся implementation evidence. Q-016 и exact capture runtime открыты. Д081–Д083/Д090/Д099 проверяют оба camera types, indicator order, revoke, crash и отсутствие файлов/сетевых frames. Privacy backup/fixtures остаются сквозным Q-021.
