package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/trace"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/machina/mosaico/internal/config"
	"github.com/machina/mosaico/internal/hotkeys"
	"github.com/machina/mosaico/internal/strip"
	"github.com/machina/mosaico/internal/wm"
)

var (
	windowStyle        = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Width(20).Height(5)
	focusedWindowStyle = windowStyle.BorderForeground(lipgloss.Color("10"))
	columnStyle        = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder()).
				Padding(0, 1)

	focusedColumnStyle = columnStyle.
				BorderForeground(lipgloss.Color("12")) // blue
)

type WindowsChanged struct{}

type model struct {
	strip        *strip.Strip
	screenWidth  float64
	screenHeight float64
	colWidth     float64
	gap          float64
	debug        string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "l":
			fmt.Fprintf(os.Stderr, "BEFORE: FocusedCol=%d, ViewportStart=%d\n",
				m.strip.FocusedCol, m.strip.ViewportStart)
			m.strip.ScrollRight()
			fmt.Fprintf(os.Stderr, "AFTER: FocusedCol=%d, ViewportStart=%d\n",
				m.strip.FocusedCol, m.strip.ViewportStart)
			m.applyLayout()
		case "h":
			m.strip.ScrollLeft()
			m.applyLayout()
			focusedCol := m.strip.Columns[m.strip.FocusedCol]
			if len(focusedCol.Windows) > 0 {
				wm.FocusApp(focusedCol.Windows[focusedCol.Focused].PID)
			}
		case "a":
			window := &strip.Window{ID: rand.Uint32(), Title: "wooo"}
			m.strip.AddWindow(window)
		case "k":
			m.strip.ScrollUp()
			m.applyLayout()
			focusedCol := m.strip.Columns[m.strip.FocusedCol]
			if len(focusedCol.Windows) > 0 {
				wm.FocusApp(focusedCol.Windows[focusedCol.Focused].PID)
			}
		case "j":
			m.strip.ScrollDown()
			m.applyLayout()
			focusedCol := m.strip.Columns[m.strip.FocusedCol]
			if len(focusedCol.Windows) > 0 {
				wm.FocusApp(focusedCol.Windows[focusedCol.Focused].PID)
			}
		case "d":
			m.strip.RemoveWindow()
			m.applyLayout()
		case "H":
			m.strip.MoveWindowLeft()
			m.applyLayout()
		case "L":
			m.strip.MoveWindowRight()
			m.applyLayout()
		case "K":
			m.strip.MoveWindowUp()
			m.applyLayout()
		case "J":
			m.strip.MoveWindowDown()
			m.applyLayout()

		}
	case hotkeys.Command:
		switch msg {
		case hotkeys.CmdScrollLeft:
			m.strip.ScrollLeft()
			m.applyLayout()
			focusedCol := m.strip.Columns[m.strip.FocusedCol]
			if len(focusedCol.Windows) > 0 {
				wm.FocusApp(focusedCol.Windows[focusedCol.Focused].PID)
			}
		case hotkeys.CmdScrollRight:
			m.strip.ScrollRight()
			m.applyLayout()
			focusedCol := m.strip.Columns[m.strip.FocusedCol]
			if len(focusedCol.Windows) > 0 {
				wm.FocusApp(focusedCol.Windows[focusedCol.Focused].PID)
			}
		case hotkeys.CmdFocusDown:
			m.strip.ScrollDown()
			m.applyLayout()
		case hotkeys.CmdFocusUp:
			m.strip.ScrollUp()
			m.applyLayout()
		}

	case WindowsChanged:
		fmt.Fprintln(os.Stderr, "WindowsChanged received") // stderr won't mess up TUI
		m.applyLayout()

	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {

	if len(m.strip.Columns) == 0 {
		return "No windows. Press 'a' to add."
	}

	visible := m.strip.GetVisibleColumns()
	var columnBoxes []string
	for i, col := range visible {
		coldIndex := m.strip.ViewportStart + i
		var windowBoxes []string
		for o, win := range col.Windows {
			if o == col.Focused && coldIndex == m.strip.FocusedCol {
				windowBoxes = append(windowBoxes, focusedWindowStyle.Render(win.Title))
			} else {
				windowBoxes = append(windowBoxes, windowStyle.Render(win.Title))
			}
		}
		stack := lipgloss.JoinVertical(lipgloss.Left, windowBoxes...)
		if coldIndex == m.strip.FocusedCol {
			columnBoxes = append(columnBoxes, focusedColumnStyle.Render(stack))
		} else {
			columnBoxes = append(columnBoxes, columnStyle.Render(stack))
		}
	}

	debug := fmt.Sprintf("\nscreen: %v x %v | colWidth: %v | gap: %v\n",
		m.screenWidth, m.screenHeight, m.colWidth, m.gap)

	return lipgloss.JoinHorizontal(lipgloss.Top, columnBoxes...) + "\n" + debug + "\n" + m.debug
}

func (m *model) applyLayout() {
	gap := float64(10)

	screenWidth, screenHeight, err := wm.GetScreenBounds()
	if err != nil {
		// handle error, for now just hardcode
		screenWidth = 2560
		screenHeight = 1440
	}

	colWidth := screenWidth / float64(m.strip.VisibleCount)
	for i, col := range m.strip.Columns {
		x := (float64(i)-float64(m.strip.ViewportStart))*colWidth + gap/2
		w := colWidth - gap

		// Calculate height per window
		winCount := len(col.Windows)
		totalGaps := gap * float64(winCount-1)
		winHeight := (screenHeight - gap - totalGaps) / float64(winCount)

		if x < 0 || x+w > screenWidth {
			wm.HideApp(col.Windows[0].PID)
		} else {
			wm.UnhideApp(col.Windows[0].PID)
		}

		for j, win := range col.Windows {
			winY := gap/2 + float64(j)*(winHeight+gap)
			wm.SetPositionAndSize(win.PID, float64(x), float64(winY), float64(w), float64(winHeight))
		}
	}
	m.screenWidth = screenWidth
	m.screenHeight = screenHeight
	m.colWidth = colWidth
	m.gap = gap
}

func isProcessRunning(pid uint32) bool {
	// kill -0 checks if process exists without killing it
	err := syscall.Kill(int(pid), 0)
	return err == nil
}

func watchWindows(p *tea.Program, s *strip.Strip) {
	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		start := time.Now()
		windows, _ := wm.GetWindowList()
		fmt.Fprintf(os.Stderr, "GetWindowList took: %v\n", time.Since(start))
		currentPIDs := s.GetAllWindowPIDs()
		changed := false

		newPIDs := make(map[uint32]bool)
		for _, w := range windows {
			// fmt.Printf("ID: %d, PID: %d, Name: %s\n", w.ID, w.PID, w.OwnerName)
			newPIDs[w.PID] = true
			if !currentPIDs[w.PID] {
				s.AddWindow(&strip.Window{
					ID:    w.ID,
					PID:   w.PID,
					Title: w.OwnerName,
				})
				changed = true
			}
		}

		// for pid := range currentPIDs {
		// 	if !isProcessRunning(pid) {
		// 		s.RemoveWindowByPID(pid)
		// 		changed = true
		// 	}
		// }

		if changed {
			p.Send(WindowsChanged{})
		}
	}
}

func main() {
	f, _ := os.Create("trace.out")
	defer f.Close()
	trace.Start(f)
	defer trace.Stop()
	s := strip.New()
	p := tea.NewProgram(model{strip: s})

	go watchWindows(p, s)

	go func() {
		for cmd := range hotkeys.Commands {
			p.Send(cmd)
		}
	}()

	config, _ := config.Load("~/.config/mosaico/config.toml")
	hotkeys.Configure(config.Hotkeys)
	go hotkeys.StartEventTap()

	if _, err := p.Run(); err != nil {
		fmt.Printf("!!! there's been an error: %v", err)
		os.Exit(1)
	}

}
