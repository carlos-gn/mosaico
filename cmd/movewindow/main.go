package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation -framework ApplicationServices

#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <ApplicationServices/ApplicationServices.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	trusted := C.AXIsProcessTrusted()
	if trusted == 0 {
		fmt.Println("Accessibility permissions not granted!")
		fmt.Println("Go to System Settings → Privacy & Security → Accessibility")
		fmt.Println("Add your terminal (or this binary) to the list")
		return
	}
	fmt.Println("Accessibility permissions granted.")

	// Get window list
	windowList := C.CGWindowListCopyWindowInfo(C.kCGWindowListOptionOnScreenOnly, C.kCGNullWindowID)
	if windowList == 0 {
		fmt.Println("Failed to get window list")
		return
	}
	defer C.CFRelease(C.CFTypeRef(windowList))

	// List windows pids
	count := C.CFArrayGetCount(windowList)
	var windowPIDs []int
	for i := range count {
		dict := C.CFDictionaryRef(C.CFArrayGetValueAtIndex(windowList, i))

		windowName := getStringValue(dict, C.kCGWindowOwnerName)
		ownerName := getStringValue(dict, C.kCGWindowOwnerName)
		windowPID := getIntValue(dict, C.kCGWindowOwnerPID)
		windowPIDs = append(windowPIDs, windowPID)
		x, y, w, h := getWindowBounds(dict)
		fmt.Printf("%s - %s (%.0f, %.0f) %.0fx%.0f\n", ownerName, windowName, x, y, w, h)

		fmt.Printf("Window Name: %s, Owner: %s, PID: %d\n", windowName, ownerName, windowPID)
	}

	appElement := C.AXUIElementCreateApplication(C.pid_t(windowPIDs[0]))
	defer C.CFRelease(C.CFTypeRef(appElement))

	attr, err := getAttribute(appElement, "AXWindows")
	if attr == 0 {
		fmt.Println("Failed to get AXWindows attribute")
		return
	}

	if err != nil {
		fmt.Println(err)
		return
	}
	defer C.CFRelease(attr)

	fmt.Println("AXWindows attribute obtained.")

	windowArray := C.CFArrayRef(attr)
	count = C.CFArrayGetCount(windowArray)
	fmt.Printf("Window count: %d\n", count)

	if count > 0 {
		window := C.AXUIElementRef(C.CFArrayGetValueAtIndex(windowArray, 0))
		pos, err := getAttribute(window, "AXPosition")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer C.CFRelease(pos)
		var point C.CGPoint
		axValue := *(*C.AXValueRef)(unsafe.Pointer(&pos))
		C.AXValueGetValue(axValue, C.kAXValueTypeCGPoint, unsafe.Pointer(&point))
		fmt.Printf("Position: %.0f, %.0f\n", point.x, point.y)
		fmt.Println("Window position attribute obtained.")

		var newCGPoint C.CGPoint
		newCGPoint.x = point.x + 100
		newCGPoint.y = point.y + 100

		newPosValue := C.AXValueCreate(C.kAXValueTypeCGPoint, unsafe.Pointer(&newCGPoint))

		errNew := C.AXUIElementSetAttributeValue(window, createCFString("AXPosition"), C.CFTypeRef(unsafe.Pointer(newPosValue)))
		if errNew != 0 {
			fmt.Printf("Failed to set new position: AXError %d\n", errNew)
			return
		}
		C.CFRelease(C.CFTypeRef(unsafe.Pointer(newPosValue)))
		fmt.Println("Window position updated.")
	}

	// return

	// workspace := appkit.Workspace_SharedWorkspace()
	// frontApp := workspace.FrontmostApplication()
	// pid := frontApp.ProcessIdentifier()

	// fmt.Printf("Frontmost app: %s (PID: %d)\n", frontApp.LocalizedName(), pid)

	// app := C.AXUIElementCreateApplication(C.pid_t(pid))
	// defer C.CFRelease(C.CFTypeRef(app))

	// _, err := getAttribute(app, "AXWindows")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
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
