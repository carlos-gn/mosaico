package hotkeys

/*
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>

void hotkeyCallback(int keyCode, int modifiers);

static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
	if(type == kCGEventKeyDown) {
		CGKeyCode keyCode = (CGKeyCode)CGEventGetIntegerValueField(event, kCGKeyboardEventKeycode);
		CGEventFlags flags = CGEventGetFlags(event);
		hotkeyCallback((int)keyCode, (int)flags);
	}
	return event;
}

static CFMachPortRef createEventTap() {
    return CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        (CGEventMask)(1 << kCGEventKeyDown),
        eventTapCallback,
        NULL
    );
}
*/
import "C"

import (
	"strings"

	"github.com/machina/mosaico/internal/config"
)

type Command int

const (
	CmdScrollLeft Command = iota
	CmdScrollRight
	CmdFocusUp
	CmdFocusDown
)

var Commands = make(chan Command, 10)

var (
	modifierMask     int
	moveModifierMask int
	keyScrollLeft    int
	keyScrollRight   int
	keyFocusUp       int
	keyFocusDown     int
)

var handlers Handlers

type Handlers struct {
	ScrollLeft      func()
	ScrollRight     func()
	FocusUp         func()
	FocusDown       func()
	MoveWindowRight func()
	MoveWindowLeft  func()
	MoveWindowUp    func()
	MoveWindowDown  func()
}

func Configure(cfg config.HotkeyConfig) {
	modifierMask = ParseModifier(cfg.Modifier)
	moveModifierMask = ParseModifier(cfg.MoveModifier)
	keyScrollLeft = ParseKey(cfg.ScrollLeft)
	keyScrollRight = ParseKey(cfg.ScrollRight)
	keyFocusUp = ParseKey(cfg.FocusUp)
	keyFocusDown = ParseKey(cfg.FocusDown)
}

func SetHandlers(h Handlers) {
	handlers = h
}

//export hotkeyCallback
func hotkeyCallback(keyCode C.int, modifiers C.int) {
	if int(modifiers)&moveModifierMask == moveModifierMask {
		switch int(keyCode) {
		case keyScrollLeft:
			if handlers.MoveWindowLeft != nil {
				handlers.MoveWindowLeft()
			}
		case keyScrollRight:
			if handlers.MoveWindowRight != nil {
				handlers.MoveWindowRight()
			}
		case keyFocusUp:
			if handlers.MoveWindowUp != nil {
				handlers.MoveWindowUp()
			}
		case keyFocusDown:
			if handlers.MoveWindowDown != nil {
				handlers.MoveWindowDown()
			}
		}

	} else if int(modifiers)&modifierMask == modifierMask {
		switch int(keyCode) {
		case keyScrollLeft:
			if handlers.ScrollLeft != nil {
				handlers.ScrollLeft()
			}
		case keyScrollRight:
			if handlers.ScrollRight != nil {
				handlers.ScrollRight()
			}
		case keyFocusUp:
			if handlers.FocusUp != nil {
				handlers.FocusUp()
			}
		case keyFocusDown:
			if handlers.FocusDown != nil {
				handlers.FocusDown()
			}
		}
	}

}

func StartEventTap() {
	tap := C.createEventTap()
	runLoopSource := C.CFMachPortCreateRunLoopSource(C.kCFAllocatorDefault, tap, 0)
	C.CFRunLoopAddSource(C.CFRunLoopGetCurrent(), runLoopSource, C.kCFRunLoopCommonModes)
	C.CGEventTapEnable(tap, C.bool(true))
	C.CFRunLoopRun()
}

func ParseModifier(s string) int {
	mask := 0
	if strings.Contains(s, "ctrl") {
		mask |= 0x40000
	}
	if strings.Contains(s, "cmd") {
		mask |= 0x100000
	}
	if strings.Contains(s, "alt") || strings.Contains(s, "opt") {
		mask |= 0x80000
	}
	if strings.Contains(s, "shift") {
		mask |= 0x20000
	}
	return mask
}

// Key name to keycode
func ParseKey(s string) int {
	keys := map[string]int{
		"h": 4, "j": 38, "k": 40, "l": 37,
		"a": 0, "s": 1, "d": 2, "f": 3,
		// add more as needed
	}
	return keys[strings.ToLower(s)]
}
