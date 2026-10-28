# Д005. Capture/platform matrix

| Target | Candidate API/runtime | Запуск | Remaining gate |
|---|---|---|---|
| Linux x86_64 / X11 | Native X11 path; host smoke через Pillow ImageGrab | Реальная capture attempt завершилась BLOCKED/OSError; synthetic analyzer PASS-SIM | Real session/display permissions, DPI/multi-display/capture cleanup |
| Linux x86_64 / Wayland | XDG ScreenCast portal + PipeWire frames | DOC-ONLY; portal session не создавалась | Compositor backend, consent, Go binding, frame formats |
| Windows x86_64 | Desktop Duplication/DXGI candidate | BLOCKED_NO_TARGET | Native build, BGRA/stride, display rotation, lock/revoke/protected content |
| macOS arm64 | ScreenCaptureKit candidate | BLOCKED_NO_TARGET | Actual minimum OS/permissions/native dependencies/package |
| macOS x86_64 | ScreenCaptureKit candidate | BLOCKED_NO_TARGET | Intel Mac native/runtime availability и отдельная установка |

Synthetic fixture имеет четыре известные edges; weighted average с depth weights проверен numeric golden. Corner exclusion проверен отдельно: угловой цвет соседней стороны не портит край. Probe не является готовым Go screen-client, не передаёт screenshots и не хранит runtime frames.

Для каждой платформы фиксировать OS/API version, architecture, permissions, multiple displays, scale/HDR/color space, cursor policy, stride/channel order, monitor-off и unsupported protected content. Go wrapper/cross-compile не закрывает capture test. Full packaging targets остаются Д061–Д068.

Источники: [Microsoft Desktop Duplication](https://learn.microsoft.com/en-us/windows/win32/direct3ddxgi/desktop-dup-api), [XDG ScreenCast](https://flatpak.github.io/xdg-desktop-portal/docs/doc-org.freedesktop.portal.ScreenCast.html), [Apple ScreenCaptureKit](https://developer.apple.com/documentation/screencapturekit). Apple API page требует JS; версия/availability на Mac здесь не выдана за проверенную.
