# Mosaico

A scrollable column window manager for macOS, inspired by [niri](https://github.com/YaLTeR/niri).

Windows are arranged in columns on an infinite horizontal strip. Scroll left/right to navigate between columns.

```
                    +-- VISIBLE SCREEN --+
  +-------+ +-------| +-------+ +-------| +-------+
  | Col 0 | | Col 1 | | Col 2 | | Col 3 | | Col 4 |
  |Safari | | Term  | | VSCode| | Slack | | Notes |
  +-------+ +-------| +-------+ +-------| +-------+
       <--scroll--  +-----------------+  --scroll-->
```

## Features

- Scrollable column layout (like niri)
- Global hotkeys (Ctrl+Cmd+H/J/K/L for scroll, Ctrl+Cmd+Alt+H/J/K/L for movement)
- Fast window positioning (cached window references)
- Hide/unhide off-viewport windows
- Focus follows scroll
- Vertical stacking (multiple windows per column)
- Auto-detect new/closed windows
- Configurable hotkeys via TOML

## Requirements

- macOS 12+
- Accessibility permissions (System Settings → Privacy & Security → Accessibility)

## Installation

```bash
go build -o mosaico ./cmd/daemon
./mosaico
```

## Usage

Run the daemon:

```bash
./mosaico
```

Default hotkeys:

| Hotkey | Action |
|--------|--------|
| Ctrl+Cmd+H | Scroll left |
| Ctrl+Cmd+L | Scroll right |
| Ctrl+Cmd+K | Focus up (within column) |
| Ctrl+Cmd+J | Focus down (within column) |
| Ctrl+Cmd+Alt+H | Move window left |
| Ctrl+Cmd+Alt+L | Move window right |
| Ctrl+Cmd+Alt+K | Move window up |
| Ctrl+Cmd+Alt+J | Move window down |

## Configuration

Create `~/.config/mosaico/config.toml`:

```toml
[hotkeys]
modifier = "ctrl+cmd"
scroll_left = "h"
scroll_right = "l"
focus_up = "k"
focus_down = "j"
```

## How It Works

Mosaico uses macOS Accessibility API to:
1. List windows (`CGWindowListCopyWindowInfo`)
2. Position/resize windows (`AXUIElementSetAttributeValue`)
3. Hide/unhide apps (`NSRunningApplication.hide/unhide`)
4. Focus apps (`NSRunningApplication.activateWithOptions`)

Global hotkeys are captured via `CGEventTap`.

## Project Structure

```
mosaico/
├── cmd/
│   ├── daemon/main.go    # Main headless daemon
│   └── tui/main.go       # TUI for debugging
├── internal/
│   ├── strip/            # Column/window data model
│   ├── wm/               # macOS window manager
│   ├── hotkeys/          # Global hotkey handling
│   └── config/           # TOML config
└── go.mod
```

## License

MIT License - see [LICENSE](LICENSE)
