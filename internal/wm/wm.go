package wm

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ApplicationServices

#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <ApplicationServices/ApplicationServices.h>
*/
/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit

#import <AppKit/AppKit.h>

char* getBundleIDForPID(pid_t pid) {
	NSRunningApplication *app = [NSRunningApplication
runningApplicationWithProcessIdentifier:pid];
	if (app == nil) return NULL;
	const char *bundleID = [app.bundleIdentifier UTF8String];
	if (bundleID == NULL) return NULL;
	return strdup(bundleID);
}

void hideApp(pid_t pid) {
	NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
	[app hide];
}

void unhideApp(pid_t pid) {
	NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
	[app unhide];
}

void focusApp(pid_t pid) {
	NSRunningApplication *app = [NSRunningApplication runningApplicationWithProcessIdentifier:pid];
	[app activateWithOptions:NSApplicationActivateIgnoringOtherApps];
}
*/
import "C"
import (
	"fmt"
	"slices"
	"unsafe"
)

type WindowInfo struct {
	ID        uint32 // kCGWindowNumber
	PID       uint32 // kCGWindowOwnerPID
	BundleID  string
	Title     string
	OwnerName string
	X         float64
	Y         float64
	Width     float64
	Height    float64
}

var SkipApps = []string{
	"com.apple.WindowServer",
	"com.apple.dock",
	"com.apple.finder",
	"com.apple.Spotlight",
	"com.apple.notificationcenterui",
}

var windowCache = make(map[uint32]C.AXUIElementRef)

func GetWindowList() ([]WindowInfo, error) {
	windowList := C.CGWindowListCopyWindowInfo(C.kCGWindowListOptionOnScreenOnly, C.kCGNullWindowID)
	if windowList == 0 {
		return nil, fmt.Errorf("failed to get window list")
	}
	defer C.CFRelease(C.CFTypeRef(windowList))

	var windows []WindowInfo
	count := C.CFArrayGetCount(windowList)
	for i := range count {
		dict := C.CFDictionaryRef(C.CFArrayGetValueAtIndex(windowList, i))
		pid := getIntValue(dict, C.kCGWindowOwnerPID)
		bundleID := getBundleID(uint32(pid))
		if bundleID == "" {
			continue // skip system processes without bundle IDs
		}

		if slices.Contains(SkipApps, bundleID) {
			continue
		}

		windowName := getStringValue(dict, C.kCGWindowOwnerName)
		ownerName := getStringValue(dict, C.kCGWindowOwnerName)
		x, y, w, h := getWindowBounds(dict)
		id := getIntValue(dict, C.kCGWindowNumber)

		windows = append(windows, WindowInfo{
			ID:        uint32(id),
			PID:       uint32(pid),
			BundleID:  bundleID,
			Title:     windowName,
			OwnerName: ownerName,
			X:         x,
			Y:         y,
			Width:     w,
			Height:    h,
		})
	}

	return windows, nil
}

func GetWindow(pid uint32) (C.AXUIElementRef, error) {
	if win, ok := windowCache[pid]; ok {
		return win, nil
	}

	app := C.AXUIElementCreateApplication(C.pid_t(pid))
	windows, err := getAttribute(app, "AXWindows")
	if err != nil {
		C.CFRelease(C.CFTypeRef(app))
		return 0, err
	}

	windowArray := C.CFArrayRef(windows)
	if C.CFArrayGetCount(windowArray) == 0 {
		C.CFRelease(windows)
		C.CFRelease(C.CFTypeRef(app))
		return 0, fmt.Errorf("no windows for pid %d", pid)
	}

	window := C.AXUIElementRef(C.CFArrayGetValueAtIndex(windowArray, 0))
	C.CFRetain(C.CFTypeRef(window))
	windowCache[pid] = window

	C.CFRelease(windows)
	C.CFRelease(C.CFTypeRef(app))

	return window, nil
}

func SetPositionAndSize(pid uint32, x, y, w, h float64) error {
	window, err := GetWindow(pid)
	if err != nil {
		return err
	}

	var point C.CGPoint
	point.x = C.CGFloat(x)
	point.y = C.CGFloat(y)
	posValue := C.AXValueCreate(C.kAXValueTypeCGPoint, unsafe.Pointer(&point))
	posAttr := createCFString("AXPosition")
	C.AXUIElementSetAttributeValue(window, posAttr, C.CFTypeRef(unsafe.Pointer(posValue))) // THIS WAS MISSING
	C.CFRelease(C.CFTypeRef(posAttr))
	C.CFRelease(C.CFTypeRef(unsafe.Pointer(posValue)))

	var size C.CGSize
	size.width = C.CGFloat(w)
	size.height = C.CGFloat(h)
	sizeValue := C.AXValueCreate(C.kAXValueTypeCGSize, unsafe.Pointer(&size))
	sizeAttr := createCFString("AXSize")
	C.AXUIElementSetAttributeValue(window, sizeAttr, C.CFTypeRef(unsafe.Pointer(sizeValue)))
	C.CFRelease(C.CFTypeRef(sizeAttr))
	C.CFRelease(C.CFTypeRef(unsafe.Pointer(sizeValue)))

	return nil
}
func GetScreenBounds() (width, height float64, err error) {
	mainDisplayID := C.CGMainDisplayID()
	if mainDisplayID == 0 {
		return 0, 0, fmt.Errorf("failed to get main display ID")
	}

	rect := C.CGDisplayBounds(mainDisplayID)
	return float64(rect.size.width), float64(rect.size.height), nil
}

func HideApp(pid uint32) {
	fmt.Printf("HideApp PID=%d\n", pid)
	C.hideApp(C.pid_t(pid))
}

func UnhideApp(pid uint32) {
	fmt.Printf("UnhideApp PID=%d\n", pid)
	C.unhideApp(C.pid_t(pid))
}

func FocusApp(pid uint32) {
	C.focusApp(C.pid_t(pid))
}

func createCFString(s string) C.CFStringRef {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	return C.CFStringCreateWithCString(C.kCFAllocatorDefault, cstr, C.kCFStringEncodingUTF8)
}

func getAttribute(elem C.AXUIElementRef, attr string) (C.CFTypeRef, error) {
	cfAttr := createCFString(attr)
	defer C.CFRelease(C.CFTypeRef(cfAttr))

	var value C.CFTypeRef
	result := C.AXUIElementCopyAttributeValue(elem, cfAttr, &value)
	if result != 0 {
		return 0, fmt.Errorf("AXError: %d", result)
	}
	return value, nil
}

func getStringValue(dict C.CFDictionaryRef, key C.CFStringRef) string {
	v := C.CFDictionaryGetValue(dict, unsafe.Pointer(key))
	if v == nil {
		return ""
	}

	cfString := C.CFStringRef(v)

	// Get string length and buffer
	length := C.CFStringGetLength(cfString)
	maxSize := C.CFStringGetMaximumSizeForEncoding(length, C.kCFStringEncodingUTF8) + 1

	buffer := C.malloc(C.size_t(maxSize))
	defer C.free(buffer)

	C.CFStringGetCString(cfString, (*C.char)(buffer), maxSize, C.kCFStringEncodingUTF8)

	return C.GoString((*C.char)(buffer))
}

func getIntValue(dict C.CFDictionaryRef, key C.CFStringRef) int {
	v := C.CFDictionaryGetValue(dict, unsafe.Pointer(key))
	if v == nil {
		return 0
	}

	cfNumber := C.CFNumberRef(v)
	var intValue C.int
	C.CFNumberGetValue(cfNumber, C.kCFNumberIntType, unsafe.Pointer(&intValue))
	return int(intValue)
}

func getWindowBounds(dict C.CFDictionaryRef) (x, y, width, height float64) {
	boundsRef := C.CFDictionaryGetValue(dict, unsafe.Pointer(C.kCGWindowBounds))
	if boundsRef == nil {
		return
	}

	var rect C.CGRect
	C.CGRectMakeWithDictionaryRepresentation(C.CFDictionaryRef(boundsRef), &rect)

	return float64(rect.origin.x), float64(rect.origin.y),
		float64(rect.size.width), float64(rect.size.height)
}

func getBundleID(pid uint32) string {
	cstr := C.getBundleIDForPID(C.pid_t(pid))
	if cstr == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(cstr))
	return C.GoString(cstr)
}
