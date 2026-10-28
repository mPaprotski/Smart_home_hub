# Д011. Monorepo/toolchain — результат

Реализованы root Go module и entrypoints Hub/screen-client/bot, React/TypeScript/Vite shell, Python scalar-observation package, Make/npm wrapper и pinned dependencies. Source root сохранён внутри целевого repo; нет imports из experiments. Полный контекст: [общий отчёт](DAYS_011_015_RESULT.md).

| Проверка | Результат / evidence |
|---|---|
| D11-01 Единственный root | PASS-HOST; прежние docs/hardware/experiments сохранены, ТЗ SHA неизменен |
| D11-02 Go module/build | PASS-HOST; три binary собраны, module-aware build; screen-client/bot честно NOT_IMPLEMENTED |
| D11-03 Frontend | PASS-HOST; actual lint/typecheck/format/build на project-local Node22.23.4 |
| D11-04 Python | PASS-HOST;14 pytest cases, Ruff/mypy и wheel/sdist, без camera/model imports |
| D11-05 Commands | PASS-HOST; fmt/check разделены, errors завершают Make; negative CLI exit2 проверены |
| D11-06 Clean quickstart | PASS-HOST; fresh source copy, новая venv/npm ci, check/build/native smoke; не опубликованный checkout |

[Quality log](results/days011-015/quality.log), [clean copy log](results/days011-015/clean-source.log), [quickstart](development/README.md). Node18 сохранён, локальная22 установлена явно. Статический UI пока без API, поэтому frontend unit tests N/A. Python runtime — только контракт, inference не реализован.
