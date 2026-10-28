# Д005. Local inference feasibility

| Вариант | Статус | Trade-off |
|---|---|---|
| Отдельный локальный Python worker | PASS-HOST для tiny synthetic logistic inference | Изоляция отказов; нужно доставить Python/native deps для каждой платформы |
| Embedded ONNX/native inference в Go client | NOT_RUN: onnxruntime/dependencies отсутствуют | Возможен иной packaging, нужны four-target build/runtime/license checks |
| Remote cloud inference | Не рассматривается | Запрещён основным privacy requirement |

Worker принимает bounded JSON line с brightness0…1, считает фиксированную synthetic logistic function, возвращает demo label. Это **не обученная Scene model, не classifier всех девяти labels и не quality validation**. Публичного/личного dataset не скачивали, model/data созданы как synthetic contract fixture.

Actual probe: два успешных inference, затем worker намеренно завершён. Независимый synthetic color loop продолжил выдавать frames с target30Hz; requests не чаще1Hz, queue max1, unavailable worker дал fallback. Результаты/intervals/число frames в [replay runs](../../experiments/results/run-1.json). Это проверка lifecycle, не Ambilight FPS/e2e.

Решение для Д071: выбрать формат actual model после platform/runtime smoke, checksum/license/model manifest, versioned preprocessing/features, timeout/cancel, offline startup и inference memory budget. Кадры остаются на ноутбуке; Core получает technical labels/timing только при действующем stream binding.
