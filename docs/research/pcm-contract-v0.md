# Д004. PCM contract v0

**Проектный формат пробы:** 48000 frames/sec, stereo FL/FR, signed16 little-endian interleaved; 4 bytes/sample frame. Chunk candidate480 frames=10ms. Это design default для spike, не универсальная гарантия всех audio sources.

Производитель — общий playback-bus monitor host audio; consumer — DSP worker. Никакой microphone substitute. Packet содержит source_epoch, sequence, format version, monotonic capture timing и PCM только в RAM. Output смена не меняет логическую точку анализа без явного rebind.

Backpressure: bounded buffer, realtime consumer не тормозит playback; overflow отмечается счётчиком, старый analysis window отбрасывается, новый source_epoch сбрасывает normalization/BPM. Silence — реальные нули, no-signal — отдельный status, не придуманный beat. Таймаут/worker failure оставляет audio доступным.

**Выполнено:** FFmpeg synthetic tone/silence→raw PCM: 12000 sample frames за0.25sec, tone1000Hz, silence RMS0. Isolated PipeWire monitor передал тот же synthetic frequency с ожидаемой startup silence; active-window estimate около1000Hz. Все данные читались в RAM, scalar results в [run-1](../../experiments/results/run-1.json).

Overall crossing estimate за window с startup silence может быть ниже1000Hz; это не BPM и не измерение playback quality. Поэтому дополнительно проверена частота active portion, а не скрыта тишина нормализацией. Consumer reconnect действительно проверен, physical source switch/playback quality ещё BLOCKED.

**Открыто:** actual BlueZ/Spotify/local adapters, производственный chunk/window/latency budget, headless service owner, output switch и full DSP FFT/beat/BPM. Они не реализованы этим spike.
