# Roadmap

## Focus Detection
- [ ] Scroll to window when user clicks or alt-tabs to it
  - Use `NSWorkspace.didActivateApplicationNotification` for app-level focus
  - Find which column has that PID and scroll viewport to it
  - Later: AXObserver for `kAXFocusedWindowChangedNotification` for window-level granularity

## Window Placement
- [ ] New windows open in current viewport (not at end of strip)
  - Insert at `FocusedCol` or `FocusedCol + 1` instead of appending

## Mouse Support
- [ ] Handle user dragging windows
  - Options:
    - Re-layout on next hotkey (simple)
    - AXObserver for `kAXMovedNotification` (responsive)
    - Poll positions and snap back (fights user)
  - Need to decide: snap back vs reorder based on drop location
