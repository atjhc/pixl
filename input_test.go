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
