# Перенос в Smart_home_hub

Дата: 25 октября 2026 года. Новый корень проекта — Git-репозиторий `Smart_home_hub` внутри рабочей области SmartHome.

Перенесены `.gitignore`, `Smart_Home_Hub_TZ.md`, `configs/`, `deploy/`, `docs/`, `experiments/`, `hardware/`, включая локальные игнорируемые caches. Служебные каталоги родительской рабочей области не переносились. Existing `.git/HEAD` и `.git/config` репозитория сохранены.

Сразу после перемещения проверены SHA256 всех93 source/docs/evidence файлов: содержимое совпало. Затем обновлено только указание project root в плане Д011–Д015 и добавлены README/этот отчёт. В родительской области оставлен README со ссылкой на репозиторий, дубликатов исходников нет.

Проверены относительные Markdown ссылки и ignore rules для cache/bytecode/private keys. `common.ROOT` и runners определяют корень относительно своего файла; изменений source paths для запуска не понадобилось.

Из нового расположения фактически выполнен `python3 experiments/run_all.py`:17 Python tests и3 Go tests с race detector прошли, probes зон/PCM/PipeWire/screen fixtures/inference/thermal/HTTPS успешны в своих host/simulation scope. Pi HIL остаётся BLOCKED, M0_NOT_READY не изменён. Исторические run1/run2/replay evidence не перезаписаны.

[Результат проверки после переноса](../experiments/results/relocation-check.json). Основные команды и навигация — в [README репозитория](../README.md).
