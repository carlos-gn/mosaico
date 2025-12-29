package strip

import (
	"fmt"
	"os"
	"sync"
)

type Strip struct {
	Columns       []*Column
	FocusedCol    int
	ViewportStart int
	VisibleCount  int
	Mutex         sync.Mutex
}

type Column struct {
	Windows []*Window // Should we have more than one window per column?
	Focused int
}

type Window struct {
	ID    uint32
	PID   uint32
	Title string
}

func New() *Strip {
	s := &Strip{
		Columns:       make([]*Column, 0),
		FocusedCol:    0,
		VisibleCount:  2,
		ViewportStart: 0,
	}
	return s
}

func (s *Strip) clampFocus() {
	if len(s.Columns) == 0 {
		s.FocusedCol = 0
		s.ViewportStart = 0
		return
	}
	if s.FocusedCol >= len(s.Columns) {
		s.FocusedCol = len(s.Columns) - 1
	}
	if s.ViewportStart >= len(s.Columns) {
		s.ViewportStart = len(s.Columns) - 1
	}

	if s.FocusedCol < s.ViewportStart {
		s.ViewportStart = s.FocusedCol
	}
	if s.FocusedCol >=
		s.ViewportStart+s.VisibleCount {
		s.ViewportStart = s.FocusedCol -
			s.VisibleCount + 1
	}

	// Clamp ViewportStart
	if s.ViewportStart < 0 {
		s.ViewportStart = 0
	}
}

func (s *Strip) AddWindow(w *Window) {
	if w == nil {
		fmt.Println("WARNING: AddWindow called with nil")
		return
	}
	nc := &Column{Windows: []*Window{w}}
	s.Columns = append(s.Columns, nc)
	fmt.Printf("AddWindow: now %d columns\n", len(s.Columns))
}

func (s *Strip) RemoveWindow() {
	if len(s.Columns) == 0 {
		return
	}

	col := s.Columns[s.FocusedCol]
	if len(col.Windows) == 1 {
		s.Columns = append(s.Columns[:s.FocusedCol], s.Columns[s.FocusedCol+1:]...)
		s.clampFocus()
		return
	}

	col.Windows = append(col.Windows[:col.Focused], col.Windows[col.Focused+1:]...)
}

func (s *Strip) RemoveWindowByID(id uint32) {
	for colIdx, col := range s.Columns {
		for winIdx, win := range col.Windows {
			if win.ID == id {
				// Remove window from column
				col.Windows = append(col.Windows[:winIdx], col.Windows[winIdx+1:]...)

				// If column is empty, remove it
				if len(col.Windows) == 0 {
					s.Columns = append(s.Columns[:colIdx], s.Columns[colIdx+1:]...)
				} else {
					col.clampFocus()
				}

				s.clampFocus()
				return
			}
		}
	}
}

func (s *Strip) RemoveWindowByPID(pid uint32) {
	for colIdx, col := range s.Columns {
		for winIdx, win := range col.Windows {
			if win.PID == pid {
				// Remove window from column
				col.Windows = append(col.Windows[:winIdx], col.Windows[winIdx+1:]...)

				// If column is empty, remove it
				if len(col.Windows) == 0 {
					s.Columns = append(s.Columns[:colIdx], s.Columns[colIdx+1:]...)
				} else {
					col.clampFocus()
				}

				s.clampFocus()
				return
			}
		}
	}
}
func (s *Strip) ScrollRight() {
	fmt.Fprintf(os.Stderr, "ScrollRight: FocusedCol=%d, ViewportStart=%d, Columns=%d\n",
		s.FocusedCol, s.ViewportStart, len(s.Columns))

	if s.FocusedCol >= len(s.Columns)-1 {
		return
	}
	s.FocusedCol++

	if s.FocusedCol >= s.ViewportStart+s.VisibleCount {
		s.ViewportStart++
	}

	fmt.Fprintf(os.Stderr, "After: FocusedCol=%d, ViewportStart=%d\n", s.FocusedCol, s.ViewportStart)
}

func (s *Strip) ScrollLeft() {
	if s.FocusedCol <= 0 {
		return
	}
	s.FocusedCol--

	if s.FocusedCol < s.ViewportStart {
		s.ViewportStart--
	}
}

func (s *Strip) ScrollUp() {
	if len(s.Columns) == 0 {
		return
	}
	col := s.Columns[s.FocusedCol]
	if len(col.Windows) == 1 {
		return
	}

	if col.Focused == 0 {
		return
	}

	col.Focused -= 1
}

func (s *Strip) ScrollDown() {
	if len(s.Columns) == 0 {
		return
	}

	col := s.Columns[s.FocusedCol]
	if len(col.Windows) == 1 {
		return
	}

	if col.Focused == len(col.Windows)-1 {
		return
	}

	col.Focused += 1
}

func (s *Strip) GetVisibleColumns() []*Column {
	if len(s.Columns) == 0 {
		return []*Column{}
	}

	end := min(s.ViewportStart+s.VisibleCount, len(s.Columns))

	return s.Columns[s.ViewportStart:end]
}

func (s *Strip) MoveWindowLeft() {
	if s.FocusedCol == 0 {
		return
	}

	col := s.Columns[s.FocusedCol]
	win := col.Windows[col.Focused]

	leftColumn := s.Columns[s.FocusedCol-1]

	col.Windows = append(col.Windows[:col.Focused], col.Windows[col.Focused+1:]...)
	col.clampFocus()

	if len(col.Windows) == 0 {
		s.Columns = append(s.Columns[:s.FocusedCol], s.Columns[s.FocusedCol+1:]...)
	}

	leftColumn.Windows = append(leftColumn.Windows, win)
	leftColumn.Focused = len(leftColumn.Windows) - 1

	s.FocusedCol--
	s.clampFocus()
}

func (s *Strip) MoveWindowRight() {
	col := s.Columns[s.FocusedCol]
	win := col.Windows[col.Focused]

	col.Windows = append(col.Windows[:col.Focused], col.Windows[col.Focused+1:]...)
	col.clampFocus()

	if s.FocusedCol >= len(s.Columns)-1 {
		newCol := &Column{Windows: []*Window{win}}
		s.Columns = append(s.Columns, newCol)
		newCol.Focused = 0
	} else {
		rightColumn := s.Columns[s.FocusedCol+1]
		rightColumn.Windows = append(rightColumn.Windows, win)
		rightColumn.Focused = len(rightColumn.Windows) - 1
	}

	deletedColumn := false
	if len(col.Windows) == 0 {
		s.Columns = append(s.Columns[:s.FocusedCol], s.Columns[s.FocusedCol+1:]...)
		deletedColumn = true
	}

	if !deletedColumn {
		s.FocusedCol++
	}
	s.clampFocus()
}

func (c *Column) MoveWindowUp() {
	if len(c.Windows) == 0 {
		return
	}

	if c.Focused > 0 {
		c.Windows[c.Focused], c.Windows[c.Focused-1] = c.Windows[c.Focused-1], c.Windows[c.Focused]
		c.Focused--
	}
}

func (c *Column) MoveWindowDown() {
	if len(c.Windows) == 0 {
		return
	}

	if c.Focused != len(c.Windows)-1 {
		c.Windows[c.Focused], c.Windows[c.Focused+1] = c.Windows[c.Focused+1], c.Windows[c.Focused]
		c.Focused++
	}
}

func (c *Column) clampFocus() {
	if len(c.Windows) == 0 {
		c.Focused = 0
		return
	}
	if c.Focused >= len(c.Windows) {
		c.Focused = len(c.Windows) - 1
	}
}

func (s *Strip) MoveWindowUp() {
	if len(s.Columns) == 0 {
		return
	}
	s.Columns[s.FocusedCol].MoveWindowUp()
}

func (s *Strip) MoveWindowDown() {
	if len(s.Columns) == 0 {
		return
	}

	s.Columns[s.FocusedCol].MoveWindowDown()
}

func (s *Strip) GetAllWindowIDs() map[uint32]bool {
	ids := make(map[uint32]bool)
	for _, col := range s.Columns {
		for _, win := range col.Windows {
			ids[win.ID] = true
		}
	}
	return ids
}

func (s *Strip) GetAllWindowPIDs() map[uint32]bool {
	pids := make(map[uint32]bool)
	for _, col := range s.Columns {
		for _, win := range col.Windows {
			pids[win.PID] = true
		}
	}
	return pids
}

// JumpToColumn moves focus to column n (1-indexed)
func (s *Strip) JumpToColumn(n int) {
	target := n - 1
	if target < 0 || target >= len(s.Columns) {
		return
	}
	s.FocusedCol = target
	s.clampFocus()
}

// MoveToColumn moves current window to column n (1-indexed)
func (s *Strip) MoveToColumn(n int) {
	target := n - 1
	if len(s.Columns) == 0 || target < 0 {
		return
	}

	// Get current window
	col := s.Columns[s.FocusedCol]
	if len(col.Windows) == 0 {
		return
	}
	win := col.Windows[col.Focused]

	// Remove from current column
	col.Windows = append(col.Windows[:col.Focused], col.Windows[col.Focused+1:]...)

	// If current column is now empty, remove it
	if len(col.Windows) == 0 {
		s.Columns = append(s.Columns[:s.FocusedCol], s.Columns[s.FocusedCol+1:]...)
	} else {
		col.clampFocus()
	}

	// Create columns up to target if needed
	for len(s.Columns) <= target {
		s.Columns = append(s.Columns, &Column{Windows: []*Window{}})
	}

	// Add to target column
	targetCol := s.Columns[target]
	targetCol.Windows = append(targetCol.Windows, win)
	targetCol.Focused = len(targetCol.Windows) - 1

	s.FocusedCol = target
	s.clampFocus()
}
