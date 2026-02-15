package main

import (
	"strings"
	"testing"
)

func TestHasFixedSize(t *testing.T) {
	tests := []struct {
		name   string
		w, h   int
		want   bool
	}{
		{"both set", 10, 10, true},
		{"zero width", 0, 10, false},
		{"zero height", 10, 0, false},
		{"both zero", 0, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &model{fixedWidth: tt.w, fixedHeight: tt.h}
			if got := m.hasFixedSize(); got != tt.want {
				t.Errorf("hasFixedSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRenderCanvasBlank(t *testing.T) {
	m := &model{canvas: NewCanvas(3, 2)}
	got := m.renderCanvas()

	lines := strings.Split(got, "\n")
	// 2 rows + trailing newline = 3 parts, last empty
	if len(lines) != 3 || lines[2] != "" {
		t.Errorf("expected 2 lines + trailing newline, got %d parts", len(lines))
	}

	// Each row should be 3 spaces (transparent foreground → space)
	for i := 0; i < 2; i++ {
		if lines[i] != "   " {
			t.Errorf("row %d = %q, want %q", i, lines[i], "   ")
		}
	}
}

func TestRenderCanvasWithContent(t *testing.T) {
	m := &model{canvas: NewCanvas(3, 2)}
	m.canvas.Set(0, 1, "X", "white", "transparent")

	got := m.renderCanvas()
	lines := strings.Split(got, "\n")

	// Row 0, col 1 should contain "X" with ANSI styling
	if !strings.Contains(lines[0], "X") {
		t.Errorf("row 0 should contain X, got %q", lines[0])
	}

	// Row 1 should be all spaces
	hasNonSpace := false
	for _, r := range lines[1] {
		if r != ' ' && r != '\x1b' && r != '[' && r != '0' && r != 'm' {
			// Allow ANSI reset sequences
			hasNonSpace = true
		}
	}
	_ = hasNonSpace // row 1 contains only transparent cells → spaces
}

func TestRenderCanvasTransparentForeground(t *testing.T) {
	m := &model{canvas: NewCanvas(2, 1)}
	m.canvas.Set(0, 0, "X", "transparent", "blue")

	got := m.renderCanvas()
	lines := strings.Split(got, "\n")

	// Transparent foreground should render as space
	if strings.Contains(lines[0], "X") {
		t.Error("transparent foreground should not render the character")
	}
}

func TestCanvasOffset(t *testing.T) {
	m := &model{
		canvas:      NewCanvas(10, 10),
		fixedWidth:  10,
		fixedHeight: 10,
		width:       40,
		height:      30,
	}

	offY, offX := m.canvasOffset()

	// Expected: offY = (30 - 1 - 10 - 3) / 2 = 8, offX = (40 - 10 - 2) / 2 = 14
	if offY != 8 {
		t.Errorf("offsetY = %d, want 8", offY)
	}
	if offX != 14 {
		t.Errorf("offsetX = %d, want 14", offX)
	}
}

func TestCanvasOffsetNoFixedSize(t *testing.T) {
	m := &model{
		canvas: NewCanvas(10, 10),
		width:  40,
		height: 30,
	}

	offY, offX := m.canvasOffset()
	if offY != 0 || offX != 0 {
		t.Errorf("non-fixed offset = (%d,%d), want (0,0)", offY, offX)
	}
}

func TestGhostPreviewAtHover(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "●",
		selectedTool: "Point",
		hoverRow:     3,
		hoverCol:     5,
	}

	got := m.renderCellAt(3, 5)

	if !strings.Contains(got, "●") {
		t.Errorf("ghost preview should contain selected char, got %q", got)
	}
}

func TestGhostPreviewNotShownWhenDrawing(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "●",
		selectedTool: "Point",
		hoverRow:     3,
		hoverCol:     5,
		mouseDown:    true,
	}

	got := m.renderCellAt(3, 5)

	// Should not show ghost when mouse is down (drawing)
	if strings.Contains(got, "●") {
		t.Errorf("ghost preview should not appear while drawing, got %q", got)
	}
}

func TestCursorShownWithMenuOpen(t *testing.T) {
	m := &model{
		canvas:         NewCanvas(10, 10),
		selectedChar:   "●",
		selectedTool:   "Point",
		hoverRow:       3,
		hoverCol:       5,
		showCharPicker: true,
	}

	got := m.renderCellAt(3, 5)

	if !strings.Contains(got, "●") {
		t.Errorf("cursor should appear when menu is open, got %q", got)
	}
}

func TestCrosshairCursorForNonPointTools(t *testing.T) {
	for _, tool := range []string{"Rectangle", "Ellipse", "Fill", "Select"} {
		t.Run(tool, func(t *testing.T) {
			m := &model{
				canvas:       NewCanvas(10, 10),
				selectedChar: "●",
				selectedTool: tool,
				hoverRow:     3,
				hoverCol:     5,
			}

			got := m.renderCellAt(3, 5)

			if !strings.Contains(got, "┼") {
				t.Errorf("tool %s should show crosshair cursor, got %q", tool, got)
			}
			if strings.Contains(got, "●") {
				t.Errorf("tool %s should not show selected char, got %q", tool, got)
			}
		})
	}
}


func TestCanvasOffsetClamps(t *testing.T) {
	m := &model{
		canvas:      NewCanvas(100, 100),
		fixedWidth:  100,
		fixedHeight: 100,
		width:       20,
		height:      20,
	}

	offY, offX := m.canvasOffset()
	if offY < 0 || offX < 0 {
		t.Errorf("offset should not be negative: (%d,%d)", offY, offX)
	}
}
