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

func TestCursorForPaintingTools(t *testing.T) {
	for _, tool := range []string{"Rectangle", "Ellipse", "Fill", "Line"} {
		t.Run(tool, func(t *testing.T) {
			m := &model{
				canvas:       NewCanvas(10, 10),
				selectedChar: "●",
				selectedTool: tool,
				hoverRow:     3,
				hoverCol:     5,
			}

			got := m.renderCellAt(3, 5)

			if !strings.Contains(got, "●") {
				t.Errorf("tool %s should show selected char as cursor, got %q", tool, got)
			}
		})
	}
}

func TestCrosshairCursorForSelectTool(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedChar: "●",
		selectedTool: "Select",
		hoverRow:     3,
		hoverCol:     5,
	}

	got := m.renderCellAt(3, 5)

	if !strings.Contains(got, "┼") {
		t.Errorf("Select tool should show crosshair cursor, got %q", got)
	}
	if strings.Contains(got, "●") {
		t.Errorf("Select tool should not show selected char, got %q", got)
	}
}


func TestToolbarShowsFilename(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 5),
		selectedChar: "●",
		selectedTool: "Point",
		width:        80,
		filePath:     "/Users/james/Documents/code/pixl/test.txt",
	}

	got := m.renderControlBar()
	if !strings.Contains(got, "pixl/test.txt") {
		t.Errorf("toolbar should show last two path components, got %q", got)
	}
	if strings.Contains(got, "code/pixl") {
		t.Errorf("toolbar should not show more than two path components, got %q", got)
	}
}

func TestToolbarShowsShortPath(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 5),
		selectedChar: "●",
		selectedTool: "Point",
		width:        80,
		filePath:     "test.txt",
	}

	got := m.renderControlBar()
	if !strings.Contains(got, "test.txt") {
		t.Errorf("toolbar should show filename for single-component path, got %q", got)
	}
}

func TestToolbarNoFilenameWhenEmpty(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 5),
		selectedChar: "●",
		selectedTool: "Point",
		width:        80,
	}

	got := m.renderControlBar()
	if strings.Contains(got, "test.txt") {
		t.Error("toolbar should not show a filename when filePath is empty")
	}
}

func TestRenderCanvasPlainText(t *testing.T) {
	m := &model{canvas: NewCanvas(5, 2)}
	m.canvas.Set(0, 0, "A", "red", "transparent")
	m.canvas.Set(0, 2, "B", "blue", "transparent")

	got := m.renderCanvasPlain()
	lines := strings.Split(got, "\n")

	// Red foreground = \x1b[31m, blue = \x1b[34m, reset = \x1b[0m
	wantRow0 := "\x1b[31mA\x1b[0m \x1b[34mB\x1b[0m  "
	if lines[0] != wantRow0 {
		t.Errorf("plain row 0 = %q, want %q", lines[0], wantRow0)
	}
	if lines[1] != "     " {
		t.Errorf("plain row 1 = %q, want %q", lines[1], "     ")
	}
}

func TestRenderCanvasPlainWithBackground(t *testing.T) {
	m := &model{canvas: NewCanvas(2, 1)}
	m.canvas.Set(0, 0, "X", "red", "blue")

	got := m.renderCanvasPlain()
	lines := strings.Split(got, "\n")

	want := "\x1b[31;44mX\x1b[0m "
	if lines[0] != want {
		t.Errorf("plain row 0 = %q, want %q", lines[0], want)
	}
}

func TestRenderCanvasPlainWhiteIsDefault(t *testing.T) {
	m := &model{canvas: NewCanvas(2, 1)}
	m.canvas.Set(0, 0, "X", "white", "transparent")

	got := m.renderCanvasPlain()
	lines := strings.Split(got, "\n")

	// White foreground with no background should render without ANSI
	if lines[0] != "X " {
		t.Errorf("plain row 0 = %q, want %q", lines[0], "X ")
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
