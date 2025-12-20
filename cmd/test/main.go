package main

import (
	"fmt"

	"github.com/machina/mosaico/internal/wm"
)

func main() {
	windows, _ := wm.GetWindowList()
	for _, w := range windows {
		fmt.Printf("%s (PID: %d, BundleID: %s) at %.0f, %.0f\n", w.OwnerName, w.PID, w.BundleID, w.X, w.Y)
	}

	// Move first window
	if len(windows) > 0 {
		wm.SetPosition(windows[0].PID, 0, 0)
	}
}
