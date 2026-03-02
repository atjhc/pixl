package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestScreenToCanvasVariableSize(t *testing.T) {
	m := &model{
		canvas: NewCanvas(10, 10),
	}

	tests := []struct {
		name            string
		screenX, screenY int
		wantX, wantY    int
	}{
		{"origin", 0, controlBarHeight, 0, 0},
		{"offset", 3, controlBarHeight + 5, 3, 5},
		{"above control bar", 0, 0, 0, -controlBarHeight},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y := m.screenToCanvas(tt.screenX, tt.screenY)
			if x != tt.wantX || y != tt.wantY {
				t.Errorf("screenToCanvas(%d,%d) = (%d,%d), want (%d,%d)",
					tt.screenX, tt.screenY, x, y, tt.wantX, tt.wantY)
			}
		})
	}
}

func TestScreenToCanvasFixedSize(t *testing.T) {
	m := &model{
		canvas:     NewCanvas(10, 10),
		fixedWidth: 10,
		fixedHeight: 10,
		width:      40,
		height:     30,
	}

	offY, offX := m.canvasOffset()

	// Canvas starts at (offX + 1, controlBarHeight + offY + 2) in screen coords
	screenX := offX + 1
	screenY := controlBarHeight + offY + 2

	x, y := m.screenToCanvas(screenX, screenY)
	if x != 0 || y != 0 {
		t.Errorf("screenToCanvas for canvas origin = (%d,%d), want (0,0)", x, y)
	}

	// One cell to the right and down
	x, y = m.screenToCanvas(screenX+3, screenY+2)
	if x != 3 || y != 2 {
		t.Errorf("screenToCanvas offset = (%d,%d), want (3,2)", x, y)
	}
}

func TestOptionKeyHeldClearsOnNonAltKey(t *testing.T) {
	m := &model{
		canvas: NewCanvas(5, 5),
	}

	// Simulate Alt key press
	altMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: true}
	m.handleKey(altMsg)
	if !m.optionKeyHeld {
		t.Fatal("optionKeyHeld should be true after Alt key")
	}

	// Simulate non-Alt key press — should clear
	plainMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}, Alt: false}
	m.handleKey(plainMsg)
	if m.optionKeyHeld {
		t.Error("optionKeyHeld should be false after non-Alt key")
	}
}

func TestResizeSavesHistory(t *testing.T) {
	m := &model{
		canvas: NewCanvas(10, 10),
		width:  10,
		height: 11, // 10 canvas rows + 1 control bar
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()

	// Draw more content that will be lost on shrink
	m.canvas.Set(9, 9, "Y", "green", "yellow")
	m.saveToHistory()

	// Shrink terminal so canvas loses bottom row
	m.handleResize(tea.WindowSizeMsg{Width: 10, Height: 6}) // 5 canvas rows

	// Cell at (9,9) is outside new canvas
	if cell := m.canvas.Get(9, 9); cell != nil {
		t.Error("cell (9,9) should be nil after shrink")
	}

	// Undo should restore pre-resize canvas with the lost cell
	m.undo()
	if cell := m.canvas.Get(9, 9); cell == nil || cell.char != "Y" {
		t.Errorf("undo after resize should restore cell (9,9), got %+v", cell)
	}
}

func TestMousePressStartsStroke(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "blue",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:        10,
		height:       11,
	}
	m.saveToHistory()

	msg := tea.MouseMsg{X: 3, Y: controlBarHeight + 2, Type: tea.MouseLeft}
	m.handleMouse(msg)

	if !m.mouseDown {
		t.Error("mouseDown should be true after left click")
	}
	if m.startX != 3 || m.startY != 2 {
		t.Errorf("start = (%d,%d), want (3,2)", m.startX, m.startY)
	}
}

func TestMouseDragDrawsPoints(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "blue",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:        10,
		height:       11,
	}
	m.saveToHistory()

	// Press
	m.handleMouse(tea.MouseMsg{X: 1, Y: controlBarHeight, Type: tea.MouseLeft})
	// Drag
	m.handleMouse(tea.MouseMsg{X: 2, Y: controlBarHeight, Type: tea.MouseLeft})
	m.handleMouse(tea.MouseMsg{X: 3, Y: controlBarHeight, Type: tea.MouseLeft})

	for _, col := range []int{1, 2, 3} {
		cell := m.canvas.Get(0, col)
		if cell == nil || cell.char != "X" {
			t.Errorf("cell(0,%d) = %+v, want X", col, cell)
		}
	}
}

func TestMouseReleaseSavesHistory(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "blue",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:        10,
		height:       11,
	}
	m.saveToHistory()
	histLen := len(m.history)

	// Press + drag + release
	m.handleMouse(tea.MouseMsg{X: 0, Y: controlBarHeight, Type: tea.MouseLeft})
	m.handleMouse(tea.MouseMsg{X: 1, Y: controlBarHeight, Type: tea.MouseLeft})
	m.handleMouse(tea.MouseMsg{X: 1, Y: controlBarHeight, Type: tea.MouseRelease})

	if m.mouseDown {
		t.Error("mouseDown should be false after release")
	}
	if len(m.history) <= histLen {
		t.Error("history should have a new entry after stroke")
	}
}

func TestMousePressOutsideFixedCanvasIgnored(t *testing.T) {
	m := &model{
		canvas:      NewCanvas(5, 5),
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "blue",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:       20,
		height:      20,
		fixedWidth:  5,
		fixedHeight: 5,
	}
	m.saveToHistory()

	// Click far outside the canvas area
	msg := tea.MouseMsg{X: 0, Y: controlBarHeight, Type: tea.MouseLeft}
	m.handleMouse(msg)

	if m.mouseDown {
		t.Error("click outside fixed canvas should not start stroke")
	}
}

func TestToolbarToolClickToggles(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		width:   80,
		height:  30,
		toolbar: toolbarLayout{toolX: 60},
	}

	// Click on the tool button area
	msg := tea.MouseMsg{X: 65, Y: 0, Type: tea.MouseLeft}
	m.handleMouse(msg)

	if !m.showToolPicker {
		t.Error("clicking tool button should open tool picker")
	}

	// Click again to close
	m.handleMouse(msg)
	if m.showToolPicker {
		t.Error("clicking tool button again should close tool picker")
	}
}

func TestToolbarFgClickToggles(t *testing.T) {
	m := &model{
		canvas:            NewCanvas(10, 10),
		selectedTool:      "Point",
		drawingTool:       "Point",
		foregroundColor: "white",
		width:          80,
		height:         30,
		toolbar:        toolbarLayout{foregroundX: 10, backgroundX: 30, toolX: 60},
	}

	msg := tea.MouseMsg{X: 15, Y: 0, Type: tea.MouseLeft}
	m.handleMouse(msg)

	if !m.showFgPicker {
		t.Error("clicking fg button should open fg picker")
	}
	if m.showBgPicker || m.showToolPicker {
		t.Error("other pickers should be closed")
	}
}

func TestMouseReleaseWithShapeTool(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "transparent",
		selectedTool: "Rectangle",
		drawingTool:  "Rectangle",
		width:        10,
		height:       11,
	}
	m.saveToHistory()

	// Press at (0,0)
	m.handleMouse(tea.MouseMsg{X: 0, Y: controlBarHeight, Type: tea.MouseLeft})
	// Drag to (4,4)
	m.handleMouse(tea.MouseMsg{X: 4, Y: controlBarHeight + 4, Type: tea.MouseLeft})
	// Release
	m.handleMouse(tea.MouseMsg{X: 4, Y: controlBarHeight + 4, Type: tea.MouseRelease})

	// Rectangle should be drawn — corners should have char
	for _, pos := range [][2]int{{0, 0}, {0, 4}, {4, 0}, {4, 4}} {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != "X" {
			t.Errorf("rectangle corner (%d,%d) = %+v, want X", pos[0], pos[1], cell)
		}
	}
}

func TestColorPickerClickIndex(t *testing.T) {
	m := &model{
		canvas:  NewCanvas(80, 30),
		width:   80,
		height:  31,
		toolbar: toolbarLayout{foregroundItemX: 10},
	}

	tests := []struct {
		name string
		x, y int
		want int
	}{
		{"first color", 10, controlBarHeight + 1, 0},
		{"second color", 10, controlBarHeight + 2, 1},
		{"above picker", 10, controlBarHeight, -1},
		{"left of picker", 0, controlBarHeight + 1, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.MouseMsg{X: tt.x, Y: tt.y, Type: tea.MouseLeft}
			got := m.colorPickerClickIndex(msg, m.toolbar.foregroundItemX)
			if got != tt.want {
				t.Errorf("colorPickerClickIndex at (%d,%d) = %d, want %d", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestClearCanvasRequiresConfirmation(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()

	c := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}

	// First c: should NOT clear, should show confirmation
	m.handleKey(c)
	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Error("first c should not clear the canvas")
	}
	if !m.confirmClear {
		t.Error("first c should set confirmClear")
	}

	// Second c: should clear the canvas
	m.handleKey(c)
	cell = m.canvas.Get(0, 0)
	if cell == nil || cell.char != " " || cell.foregroundColor != "white" {
		t.Errorf("second c should clear the canvas, got %+v", cell)
	}
	if m.confirmClear {
		t.Error("confirmClear should be reset after clearing")
	}
}

func TestClearCanvasConfirmationCancelledByOtherKey(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()

	c := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	m.handleKey(c)
	if !m.confirmClear {
		t.Fatal("first c should set confirmClear")
	}

	// Press a different key — should cancel
	other := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	m.handleKey(other)
	if m.confirmClear {
		t.Error("non-c key should cancel confirmClear")
	}

	// Canvas should be intact
	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Error("canvas should not be cleared after cancellation")
	}
}

func TestClearCanvasConfirmationCancelledByEsc(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()

	c := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	m.handleKey(c)
	if !m.confirmClear {
		t.Fatal("first c should set confirmClear")
	}

	esc := tea.KeyMsg{Type: tea.KeyEscape}
	m.handleKey(esc)
	if m.confirmClear {
		t.Error("esc should cancel confirmClear")
	}

	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Error("canvas should not be cleared after esc cancellation")
	}
}

func TestSwapForegroundBackground(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		foregroundColor: "red",
		backgroundColor: "blue",
	}

	x := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	m.handleKey(x)

	if m.foregroundColor != "blue" {
		t.Errorf("foreground should be blue after swap, got %s", m.foregroundColor)
	}
	if m.backgroundColor != "red" {
		t.Errorf("background should be red after swap, got %s", m.backgroundColor)
	}
}

func TestSwapForegroundBackgroundWithTransparent(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		foregroundColor: "white",
		backgroundColor: "transparent",
	}

	x := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	m.handleKey(x)

	if m.foregroundColor != "transparent" {
		t.Errorf("foreground should be transparent after swap, got %s", m.foregroundColor)
	}
	if m.backgroundColor != "white" {
		t.Errorf("background should be white after swap, got %s", m.backgroundColor)
	}
}

func TestClearCanvasConfirmationDismissedByTimeout(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()
	m.confirmClear = true

	// Simulate timeout message
	m.Update(clearConfirmTimeout{})
	if m.confirmClear {
		t.Error("timeout should dismiss confirmClear")
	}

	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Error("canvas should not be cleared by timeout")
	}
}

func TestClearCanvasConfirmationDismissedByMousePress(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		selectedChar: "X",
		foregroundColor: "red",
		backgroundColor: "blue",
		width:        10,
		height:       11,
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()
	m.confirmClear = true

	msg := tea.MouseMsg{X: 3, Y: controlBarHeight + 2, Type: tea.MouseLeft}
	m.handleMouse(msg)
	if m.confirmClear {
		t.Error("mouse press should dismiss confirmClear")
	}
}

func TestClearCanvasFirstCReturnsTickCmd(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.saveToHistory()

	c := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}
	_, cmd := m.handleKey(c)
	if cmd == nil {
		t.Error("first c should return a tick command for auto-dismiss")
	}
}

func TestEyedropperSamplesCell(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		hoverRow:        3,
		hoverCol:        5,
	}
	m.canvas.Set(3, 5, "X", "red", "blue")

	i := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	m.handleKey(i)

	if m.selectedChar != "X" {
		t.Errorf("selectedChar = %q, want X", m.selectedChar)
	}
	if m.foregroundColor != "red" {
		t.Errorf("foregroundColor = %q, want red", m.foregroundColor)
	}
	if m.backgroundColor != "blue" {
		t.Errorf("backgroundColor = %q, want blue", m.backgroundColor)
	}
}

func TestEyedropperDefaultCell(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		selectedChar:    "●",
		foregroundColor: "red",
		backgroundColor: "blue",
		hoverRow:        3,
		hoverCol:        5,
	}

	// Sampling a default cell picks up its defaults
	i := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	m.handleKey(i)

	if m.selectedChar != " " {
		t.Errorf("selectedChar = %q, want space", m.selectedChar)
	}
	if m.foregroundColor != "white" {
		t.Errorf("foregroundColor = %q, want white", m.foregroundColor)
	}
	if m.backgroundColor != "transparent" {
		t.Errorf("backgroundColor = %q, want transparent", m.backgroundColor)
	}
}

func TestEyedropperOutOfBounds(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		selectedChar:    "●",
		foregroundColor: "red",
		backgroundColor: "blue",
		hoverRow:        -1,
		hoverCol:        -1,
	}

	i := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}}
	m.handleKey(i)

	if m.selectedChar != "●" {
		t.Errorf("selectedChar should be unchanged for out of bounds, got %q", m.selectedChar)
	}
}

func TestEscClosesMenuBeforeClearingSelection(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Select",
		drawingTool:  "Point",
	}
	m.showFgPicker = true
	m.selection.active = true

	esc := tea.KeyMsg{Type: tea.KeyEscape}

	// First esc: should close the menu but keep the selection
	m.handleKey(esc)
	if m.showFgPicker {
		t.Error("first esc should close the menu")
	}
	if !m.selection.active {
		t.Error("first esc should NOT clear the selection")
	}

	// Second esc: should clear the selection
	m.handleKey(esc)
	if m.selection.active {
		t.Error("second esc should clear the selection")
	}
}

func TestEscClosesToolPickerBeforeClearingSelection(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Select",
		drawingTool:  "Point",
	}
	m.showToolPicker = true
	m.toolPickerFocusLevel = 0
	m.selection.active = true

	esc := tea.KeyMsg{Type: tea.KeyEscape}

	// First esc: should close tool picker but keep selection
	m.handleKey(esc)
	if m.showToolPicker {
		t.Error("first esc should close the tool picker")
	}
	if !m.selection.active {
		t.Error("first esc should NOT clear the selection")
	}

	// Second esc: should clear the selection
	m.handleKey(esc)
	if m.selection.active {
		t.Error("second esc should clear the selection")
	}
}

func TestEscClearsSelectionWhenNoMenuOpen(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Select",
		drawingTool:  "Point",
	}
	m.selection.active = true

	esc := tea.KeyMsg{Type: tea.KeyEscape}
	m.handleKey(esc)

	if m.selection.active {
		t.Error("esc with no menu open should clear the selection")
	}
}

func TestClampToCanvas(t *testing.T) {
	m := &model{
		canvas: NewCanvas(5, 5),
	}

	tests := []struct {
		name       string
		y, x       int
		wantY, wantX int
	}{
		{"inside", 2, 3, 2, 3},
		{"top-left corner", 0, 0, 0, 0},
		{"bottom-right corner", 4, 4, 4, 4},
		{"above canvas", -1, 2, 0, 2},
		{"below canvas", 5, 2, 4, 2},
		{"left of canvas", 2, -1, 2, 0},
		{"right of canvas", 2, 5, 2, 4},
		{"both negative", -3, -5, 0, 0},
		{"both over", 10, 10, 4, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotY, gotX := m.clampToCanvas(tt.y, tt.x)
			if gotY != tt.wantY || gotX != tt.wantX {
				t.Errorf("clampToCanvas(%d,%d) = (%d,%d), want (%d,%d)",
					tt.y, tt.x, gotY, gotX, tt.wantY, tt.wantX)
			}
		})
	}
}
