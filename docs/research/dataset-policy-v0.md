# Д005. Fixture/dataset policy v0

Текущие fixtures полностью synthetic: generated tones/silence, numeric RGB borders, температуры/fake clock, bounded protocol messages и tiny logistic demo. Personal audio, реальные screens/webcam frames и third-party dataset не сохранялись/не скачивались.

## Manifest каждого будущего набора

`dataset_id`, version, task/labels, source URL/author, exact license/redistribution rights, checksums, synthetic generator/version/seed либо provenance, annotation rubric/version, sample counts/class distribution, train/validation/test split IDs, preprocessing version, negative cases, known limitations. Наличие набора в интернете не означает право коммитить его в public monorepo.

Split scene/video — по исходному ролику/sequence, не по соседним frames. Gesture split — по independent sessions/users, где возможно; negative motion обязателен. Habit split — по времени, пользовательские данные local-only, не public fixtures. Test holdout freeze до финального обучения; переразметка/новый holdout документируются.

Runtime camera frames только RAM, screen frames только laptop memory, PCM/DSP/scene history не записываются. Test failure diagnostics содержат scalar features/error codes и synthetic expected outputs, не raw personal frames. Exception/SDK auto-dumps/cache должны быть выключены/проверены до включения actual camera.

Для публичных model artifacts отдельно проверить license training data/model/runtime и отсутствие personal model state. Пользовательская habit model не release artifact. Delete history увеличивает privacy epoch и блокирует stale worker publish по ADR-006.

Численные startup quality gates находятся в [методике](performance-and-quality-plan.md); они не достигнуты этим документом. До actual Scene/Gesture training нужны пригодные annotated sets и независимые target-platform measurements.
