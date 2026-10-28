# Д007. Thermal policy v0 и результаты

[Конфигурация-кандидат](../../configs/thermal-policy.example.json), [probe](../../experiments/day007/thermal_policy_probe.py), [unit tests](../../experiments/tests/test_experiments.py).

## Проектная таблица

| State | Вход на нагреве | Выход на охлаждении | Fan candidate |
|---|---|---|---|
| NORMAL | Ниже50°C при старте | WARM при>=50 | OFF |
| WARM | >=50°C | NORMAL при<47; HOT при>=65 | ON |
| HOT | >=65°C | WARM при<62; CRITICAL при>=75 | ON |
| CRITICAL | >=75°C | HOT при<72 | ON + warning; workload policy реализуется позже |

Hysteresis3°C, ON/OFF fan и обозначение75°C как CRITICAL — **проектные defaults**, не уже подтверждённые физические границы из ТЗ. ТЗ задаёт ориентиры fan/warning, но не полный четырёхсостояний mapping. Actual fan capability/default polarity, ambient conditions и thresholds откалибровать на Pi до production.

Missing/malformed/nonfinite/out-of-range sensor input сохраняет последний state с quality ERROR и требует fan ON. Это software demand, а не доказательство electrically fail-on driver: без подходящей схемы/проверки команда не гарантирует вращения. Для первого стенда разрешён always-on fan.

## Выполнено

Synthetic sequence проверила нагрев/охлаждение, точные границы hysteresis, скачки NORMAL→CRITICAL→NORMAL, отсутствие chatter около50°C и invalid samples. Реальный thermal sysfs Pi не писался, настоящая Pi не нагревалась. Порог CRITICAL не проверяется искусственным перегревом железа.

## Реальная load проба после появления Pi

Зафиксировать idle/ambient/fan, выбрать короткую ограниченную CPU нагрузку и stop criterion по фактическому охлаждению. Измерить temperature/clock/voltage/error signals, проверить visible fan response и cooldown. Missing fan/необычный нагрев/voltage-storage errors→остановка/FAIL. Extreme/error branches оставить synthetic. Exact load command и thresholds заполняются после inventory, не запускаются автоматически на чужом host.
