# Mosaico

A scrollable column window manager for macOS, inspired by [niri](https://github.com/YaLTeR/niri).

## Requirements

- macOS 12+
- Accessibility permissions (System Settings → Privacy & Security → Accessibility)

## Installation

```bash
go build -o mosaico ./cmd/daemon
./mosaico
```

## Hotkeys

| Hotkey | Action |
|--------|--------|
| Ctrl+Cmd+Alt+H | Scroll left |
| Ctrl+Cmd+Alt+L | Scroll right |
| Ctrl+Cmd+Alt+K | Focus up |
| Ctrl+Cmd+Alt+J | Focus down |
| Ctrl+Cmd+Alt+1-9 | Jump to column |
| Shift+Ctrl+Cmd+Alt+H/L | Move window left/right |
| Shift+Ctrl+Cmd+Alt+K/J | Move window up/down |
| Shift+Ctrl+Cmd+Alt+1-9 | Move window to column |

## License

MIT
