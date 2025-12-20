package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework CoreFoundation
#include <CoreGraphics/CoreGraphics.h>
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

func main() {
	windowList := C.CGWindowListCopyWindowInfo(C.kCGWindowListOptionOnScreenOnly, C.kCGNullWindowID)
	if windowList == 0 {
		fmt.Println("Failed to get window list")
		return
	}
	defer C.CFRelease(C.CFTypeRef(windowList))

	count := C.CFArrayGetCount(windowList)
	for i := C.CFIndex(0); i < count; i++ {
		dict := C.CFDictionaryRef(C.CFArrayGetValueAtIndex(windowList, i))

		windowName := getStringValue(dict, C.kCGWindowOwnerName)
		ownerName := getStringValue(dict, C.kCGWindowOwnerName)
		x, y, w, h := getWindowBounds(dict)
		fmt.Printf("%s - %s (%.0f, %.0f) %.0fx%.0f\n", ownerName, windowName, x, y, w, h)

		fmt.Printf("Window Name: %s, Owner: %s\n", windowName, ownerName)
	}

	fmt.Printf("Found %d windows\n", count)
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
