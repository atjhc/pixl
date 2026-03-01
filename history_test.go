package main

import "testing"

func newHistoryModel() *model {
	return &model{
		canvas:          NewCanvas(5, 5),
		selectedChar:    "#",
		foregroundColor: "white",
		backgroundColor: "transparent",
		history:         []Canvas{},
		historyIndex:    -1,
	}
}

func TestSaveToHistoryFirstEntry(t *testing.T) {
	m := newHistoryModel()
	m.saveToHistory()

	if len(m.history) != 1 {
		t.Fatalf("history length = %d, want 1", len(m.history))
	}
	if m.historyIndex != 0 {
		t.Errorf("historyIndex = %d, want 0", m.historyIndex)
	}
}

func TestSaveToHistoryIsIndependent(t *testing.T) {
	m := newHistoryModel()
	m.canvas.Set(0, 0, "A", "white", "transparent")
	m.saveToHistory()

	m.canvas.Set(0, 0, "B", "white", "transparent")

	saved := m.history[0].Get(0, 0)
	if saved.char != "A" {
		t.Errorf("history snapshot mutated: char = %q, want A", saved.char)
	}
}

func TestUndoRedo(t *testing.T) {
	m := newHistoryModel()

	m.saveToHistory() // state 0: blank

	m.canvas.Set(0, 0, "A", "white", "transparent")
	m.saveToHistory() // state 1: A

	m.canvas.Set(0, 0, "B", "white", "transparent")
	m.saveToHistory() // state 2: B

	m.undo()
	if cell := m.canvas.Get(0, 0); cell.char != "A" {
		t.Errorf("after undo: char = %q, want A", cell.char)
	}

	m.undo()
	if cell := m.canvas.Get(0, 0); cell.char != " " {
		t.Errorf("after second undo: char = %q, want space", cell.char)
	}

	m.redo()
	if cell := m.canvas.Get(0, 0); cell.char != "A" {
		t.Errorf("after redo: char = %q, want A", cell.char)
	}

	m.redo()
	if cell := m.canvas.Get(0, 0); cell.char != "B" {
		t.Errorf("after second redo: char = %q, want B", cell.char)
	}
}

func TestUndoAtBoundaryIsNoop(t *testing.T) {
	m := newHistoryModel()
	m.saveToHistory()

	m.undo() // already at 0
	if m.historyIndex != 0 {
		t.Errorf("historyIndex = %d after undo at boundary, want 0", m.historyIndex)
	}
}

func TestRedoAtBoundaryIsNoop(t *testing.T) {
	m := newHistoryModel()
	m.saveToHistory()

	m.redo() // already at end
	if m.historyIndex != 0 {
		t.Errorf("historyIndex = %d after redo at boundary, want 0", m.historyIndex)
	}
}

func TestHistoryTruncatesRedoBranch(t *testing.T) {
	m := newHistoryModel()
	m.saveToHistory() // 0

	m.canvas.Set(0, 0, "A", "white", "transparent")
	m.saveToHistory() // 1

	m.canvas.Set(0, 0, "B", "white", "transparent")
	m.saveToHistory() // 2

	m.undo() // back to 1
	m.undo() // back to 0

	m.canvas.Set(0, 0, "C", "white", "transparent")
	m.saveToHistory() // should truncate redo branch

	if len(m.history) != 2 {
		t.Errorf("history length = %d, want 2 (original + new)", len(m.history))
	}

	m.redo() // should be no-op since redo branch was truncated
	if cell := m.canvas.Get(0, 0); cell.char != "C" {
		t.Errorf("after redo on truncated branch: char = %q, want C", cell.char)
	}
}

func TestHistoryCapsAt50(t *testing.T) {
	m := newHistoryModel()

	for i := 0; i < 55; i++ {
		m.canvas.Set(0, 0, "X", "white", "transparent")
		m.saveToHistory()
	}

	if len(m.history) > 50 {
		t.Errorf("history length = %d, want <= 50", len(m.history))
	}
}

func TestCopyPasteRoundTrip(t *testing.T) {
	m := newHistoryModel()
	m.canvas.Set(1, 1, "A", "red", "blue")
	m.canvas.Set(1, 2, "B", "green", "yellow")
	m.canvas.Set(2, 1, "C", "cyan", "magenta")
	m.canvas.Set(2, 2, "D", "white", "black")

	// Selection border wraps the content, so select (0,0)-(3,3) to capture (1,1)-(2,2) internally
	m.selection.active = true
	m.selection.startY = 0
	m.selection.startX = 0
	m.selection.endY = 3
	m.selection.endX = 3

	m.copySelection()

	if m.clipboard.width != 2 || m.clipboard.height != 2 {
		t.Fatalf("clipboard size = %dx%d, want 2x2", m.clipboard.width, m.clipboard.height)
	}

	// Paste at a different location via selection
	m.selection.active = true
	m.selection.startY = 0
	m.selection.startX = 0
	m.selection.endY = 3
	m.selection.endX = 3

	// Move paste target: select region starting at (0,3) so internal is (1,4)
	m.selection.startY = 0
	m.selection.startX = 3
	m.selection.endY = 3
	m.selection.endX = 3

	// Instead, let's just verify clipboard content directly
	if m.clipboard.cells[0][0].char != "A" || m.clipboard.cells[0][1].char != "B" {
		t.Errorf("clipboard row 0 = [%q, %q], want [A, B]", m.clipboard.cells[0][0].char, m.clipboard.cells[0][1].char)
	}
	if m.clipboard.cells[1][0].char != "C" || m.clipboard.cells[1][1].char != "D" {
		t.Errorf("clipboard row 1 = [%q, %q], want [C, D]", m.clipboard.cells[1][0].char, m.clipboard.cells[1][1].char)
	}

	// Paste at (0,0) selection to place content at (1,1) internal
	m.selection.startY = 0
	m.selection.startX = 0
	m.selection.endY = 3
	m.selection.endX = 3

	// Clear target area first
	m.canvas.Clear()
	m.paste()

	if cell := m.canvas.Get(1, 1); cell.char != "A" {
		t.Errorf("paste at (1,1): char = %q, want A", cell.char)
	}
	if cell := m.canvas.Get(2, 2); cell.char != "D" {
		t.Errorf("paste at (2,2): char = %q, want D", cell.char)
	}
}

func TestCutSelectionClearsRegion(t *testing.T) {
	m := newHistoryModel()
	m.canvas.Set(1, 1, "A", "red", "blue")
	m.canvas.Set(2, 2, "B", "green", "yellow")
	m.saveToHistory()

	m.selection.active = true
	m.selection.startY = 0
	m.selection.startX = 0
	m.selection.endY = 3
	m.selection.endX = 3

	m.cutSelection()

	// Clipboard should have content
	if m.clipboard.cells == nil {
		t.Fatal("clipboard is nil after cut")
	}

	// Internal region (1,1)-(2,2) should be cleared
	if cell := m.canvas.Get(1, 1); cell.char != " " || cell.foregroundColor != "transparent" {
		t.Errorf("cut region cell(1,1) not cleared: %+v", *cell)
	}
	if cell := m.canvas.Get(2, 2); cell.char != " " || cell.foregroundColor != "transparent" {
		t.Errorf("cut region cell(2,2) not cleared: %+v", *cell)
	}
}

func TestPasteWithNoClipboardIsNoop(t *testing.T) {
	m := newHistoryModel()
	before := m.canvas.Copy()
	m.paste()
	if !m.canvas.Equals(before) {
		t.Error("paste with empty clipboard should not modify canvas")
	}
}

func TestCopyWithNoSelectionIsNoop(t *testing.T) {
	m := newHistoryModel()
	m.selection.active = false
	m.copySelection()

	if m.clipboard.cells != nil {
		t.Error("copy with no selection should not set clipboard")
	}
}

func TestPasteSkipsTransparentCells(t *testing.T) {
	m := newHistoryModel()
	m.canvas.Set(1, 1, "Z", "red", "blue")

	// Manually set clipboard with a transparent cell
	m.clipboard.cells = [][]Cell{
		{{char: " ", foregroundColor: "transparent", backgroundColor: "transparent"}},
	}
	m.clipboard.width = 1
	m.clipboard.height = 1
	m.selection.active = true
	m.selection.startY = 0
	m.selection.startX = 0
	m.selection.endY = 2
	m.selection.endX = 2

	m.paste()

	// The existing cell should be preserved since clipboard cell is fully transparent
	if cell := m.canvas.Get(1, 1); cell.char != "Z" {
		t.Errorf("transparent paste should preserve existing cell, got char = %q", cell.char)
	}
}
