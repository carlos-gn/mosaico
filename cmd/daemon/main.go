package main

import (
	"fmt"
	"syscall"
	"time"

	"github.com/machina/mosaico/internal/config"
	"github.com/machina/mosaico/internal/hotkeys"
	"github.com/machina/mosaico/internal/strip"
	"github.com/machina/mosaico/internal/wm"
)

var globalStrip *strip.Strip

// func applyLayout() {
// 	gap := 10.0
// 	screenWidth, screenHeight, _ := wm.GetScreenBounds()
// 	colWidth := screenWidth / float64(globalStrip.VisibleCount)

// 	for i, col := range globalStrip.Columns {
// 		x := (float64(i)-float64(globalStrip.ViewportStart))*colWidth + gap/2
// 		w := colWidth - gap

// 		visible := x >= 0 && x < screenWidth

// 		for j, win := range col.Windows {
// 			if visible {
// 				wm.UnhideApp(win.PID)

// 				// Vertical stacking
// 				winCount := len(col.Windows)
// 				winH := (screenHeight - gap - gap*float64(winCount-1)) / float64(winCount)
// 				winY := gap/2 + float64(j)*(winH+gap)

// 				wm.SetPositionAndSize(win.PID, x, winY, w, winH)
// 			} else {
// 				wm.HideApp(win.PID)
// 			}
// 		}
// 	}
// }

func applyLayout() {
	gap := float64(10)

	screenWidth, screenHeight, _ := wm.GetScreenBounds()
	colWidth := screenWidth / float64(globalStrip.VisibleCount)
	for i, col := range globalStrip.Columns {
		x := (float64(i)-float64(globalStrip.ViewportStart))*colWidth + gap/2
		w := colWidth - gap

		// Calculate height per window

		visible := x >= 0 && x+w <= screenWidth

		for j, win := range col.Windows {
			if visible {
				winCount := len(col.Windows)
				totalGaps := gap * float64(winCount-1)
				winHeight := (screenHeight - gap - totalGaps) / float64(winCount)

				winY := gap/2 + float64(j)*(winHeight+gap)
				err := wm.SetPositionAndSize(win.PID, float64(x), float64(winY), float64(w), float64(winHeight))
				if err != nil {
					fmt.Printf("ERROR SetPositionAndSize %s: %v\n", win.Title, err)
				} else {
					fmt.Printf("Positioned %s at x=%.0f\n", win.Title, x)
				}
				wm.UnhideApp(win.PID)
			} else {
				wm.HideApp(win.PID)
			}
			fmt.Printf("Col %d: x=%.0f visible=%v, win: %s (PID=%d)\n", i, x, visible, win.Title, win.PID)
		}
	}
}

func focusCurrentWindow() {
	if len(globalStrip.Columns) == 0 {
		return
	}
	col := globalStrip.Columns[globalStrip.FocusedCol]
	if len(col.Windows) == 0 {
		return
	}
	win := col.Windows[col.Focused]
	wm.FocusApp(win.PID)
}

func watchWindows() {
	ticker := time.NewTicker(2 * time.Second)
	for range ticker.C {
		globalStrip.Mutex.Lock()

		windows, _ := wm.GetWindowList()
		currentPIDs := globalStrip.GetAllWindowPIDs()

		// Add new windows
		for _, w := range windows {
			if !currentPIDs[w.PID] {
				globalStrip.AddWindow(&strip.Window{
					ID: w.ID, PID: w.PID, Title: w.OwnerName,
				})
				wm.GetWindow(w.PID) // warm cache
				fmt.Printf("New window: %s\n", w.OwnerName)
			}
		}

		// Remove closed windows (check if process still running)
		for pid := range currentPIDs {
			if !isProcessRunning(pid) {
				globalStrip.RemoveWindowByPID(pid)
				fmt.Printf("Removed window PID=%d\n", pid)
			}
		}

		globalStrip.Mutex.Unlock()
		applyLayout()
	}
}

func isProcessRunning(pid uint32) bool {
	err := syscall.Kill(int(pid), 0)
	return err == nil
}

func main() {
	// Load config
	cfg, _ := config.Load("~/.config/mosaico/config.toml")
	hotkeys.Configure(cfg.Hotkeys)

	// Initialize strip with current windows
	globalStrip = strip.New()
	windows, _ := wm.GetWindowList()
	fmt.Printf("Found %d windows\n", len(windows))
	for _, w := range windows {
		fmt.Printf("Window: ID=%d PID=%d Title=%s\n", w.ID, w.PID, w.OwnerName)
		globalStrip.AddWindow(&strip.Window{
			ID: w.ID, PID: w.PID, Title: w.OwnerName,
		})
	}
	fmt.Printf("Strip has %d columns\n", len(globalStrip.Columns))

	// Warm the cache
	for _, col := range globalStrip.Columns {
		for _, win := range col.Windows {
			wm.GetWindow(win.PID)
		}
	}

	for i, col := range globalStrip.Columns {
		if len(col.Windows) == 0 {
			fmt.Printf("EMPTY COL %d before layout!\n", i)
		}
	}

	applyLayout()

	// Set callback handlers
	hotkeys.SetHandlers(hotkeys.Handlers{
		ScrollLeft: func() {
			globalStrip.Mutex.Lock()
			defer globalStrip.Mutex.Unlock()
			globalStrip.ScrollLeft()
			applyLayout()
			focusCurrentWindow()
		},
		ScrollRight: func() {
			globalStrip.Mutex.Lock()
			defer globalStrip.Mutex.Unlock()
			globalStrip.ScrollRight()
			applyLayout()
			focusCurrentWindow()
		},
		FocusUp: func() {
			globalStrip.Mutex.Lock()
			defer globalStrip.Mutex.Unlock()
			globalStrip.ScrollUp()
			applyLayout()
			focusCurrentWindow()
		},
		FocusDown: func() {
			globalStrip.Mutex.Lock()
			defer globalStrip.Mutex.Unlock()
			globalStrip.ScrollDown()
			applyLayout()
			focusCurrentWindow()
		},
	})

	go watchWindows()

	// Start event tap (blocks forever)
	hotkeys.StartEventTap()
}
