package main

import "testing"

func TestNewCanvas(t *testing.T) {
	c := NewCanvas(3, 2)

	if c.width != 3 {
		t.Errorf("width = %d, want 3", c.width)
	}
	if c.height != 2 {
		t.Errorf("height = %d, want 2", c.height)
	}

	for row := 0; row < c.height; row++ {
		for col := 0; col < c.width; col++ {
			cell := c.Get(row, col)
			if cell == nil {
				t.Fatalf("Get(%d,%d) = nil", row, col)
			}
			if cell.char != " " {
				t.Errorf("cell(%d,%d).char = %q, want %q", row, col, cell.char, " ")
			}
			if cell.foregroundColor != "white" {
				t.Errorf("cell(%d,%d).foregroundColor = %q, want %q", row, col, cell.foregroundColor, "white")
			}
			if cell.backgroundColor != "transparent" {
				t.Errorf("cell(%d,%d).backgroundColor = %q, want %q", row, col, cell.backgroundColor, "transparent")
			}
		}
	}
}

func TestSetGet(t *testing.T) {
	c := NewCanvas(3, 3)

	c.Set(1, 2, "X", "red", "blue")
	cell := c.Get(1, 2)
	if cell == nil {
		t.Fatal("Get(1,2) = nil after Set")
	}
	if cell.char != "X" || cell.foregroundColor != "red" || cell.backgroundColor != "blue" {
		t.Errorf("cell = %+v, want {X red blue}", *cell)
	}
}

func TestSetOutOfBoundsIsNoop(t *testing.T) {
	c := NewCanvas(3, 3)
	// Should not panic
	c.Set(-1, 0, "X", "red", "blue")
	c.Set(0, -1, "X", "red", "blue")
	c.Set(3, 0, "X", "red", "blue")
	c.Set(0, 3, "X", "red", "blue")
}

func TestGetOutOfBoundsReturnsNil(t *testing.T) {
	c := NewCanvas(3, 3)

	tests := []struct {
		row, col int
	}{
		{-1, 0},
		{0, -1},
		{3, 0},
		{0, 3},
	}
	for _, tt := range tests {
		if got := c.Get(tt.row, tt.col); got != nil {
			t.Errorf("Get(%d,%d) = %+v, want nil", tt.row, tt.col, got)
		}
	}
}

func TestClear(t *testing.T) {
	c := NewCanvas(3, 3)
	c.Set(0, 0, "X", "red", "blue")
	c.Set(2, 2, "Y", "green", "yellow")

	c.Clear()

	for row := 0; row < c.height; row++ {
		for col := 0; col < c.width; col++ {
			cell := c.Get(row, col)
			if cell.char != " " || cell.foregroundColor != "white" || cell.backgroundColor != "transparent" {
				t.Errorf("cell(%d,%d) after Clear = %+v", row, col, *cell)
			}
		}
	}
}

func TestLoadText(t *testing.T) {
	c := NewCanvas(5, 3)
	c.LoadText("AB\nCD")

	tests := []struct {
		row, col int
		wantChar string
	}{
		{0, 0, "A"},
		{0, 1, "B"},
		{0, 2, " "}, // untouched
		{1, 0, "C"},
		{1, 1, "D"},
		{2, 0, " "}, // row beyond text
	}
	for _, tt := range tests {
		cell := c.Get(tt.row, tt.col)
		if cell.char != tt.wantChar {
			t.Errorf("cell(%d,%d).char = %q, want %q", tt.row, tt.col, cell.char, tt.wantChar)
		}
	}
}

func TestLoadTextStopsAtBoundaries(t *testing.T) {
	c := NewCanvas(2, 2)
	c.LoadText("ABCDEF\nGHIJKL\nMNOPQR")

	if cell := c.Get(0, 0); cell.char != "A" {
		t.Errorf("(0,0) = %q, want A", cell.char)
	}
	if cell := c.Get(0, 1); cell.char != "B" {
		t.Errorf("(0,1) = %q, want B", cell.char)
	}
	if cell := c.Get(1, 0); cell.char != "G" {
		t.Errorf("(1,0) = %q, want G", cell.char)
	}
}

func TestLoadTextSkipsSpaces(t *testing.T) {
	c := NewCanvas(3, 1)
	c.Set(0, 1, "X", "red", "blue")
	c.LoadText("A B")

	if cell := c.Get(0, 0); cell.char != "A" {
		t.Errorf("(0,0) = %q, want A", cell.char)
	}
	// Space in input should not overwrite existing cell
	if cell := c.Get(0, 1); cell.char != "X" {
		t.Errorf("(0,1) = %q, want X (space should not overwrite)", cell.char)
	}
	if cell := c.Get(0, 2); cell.char != "B" {
		t.Errorf("(0,2) = %q, want B", cell.char)
	}
}

func TestLoadTextANSIForeground(t *testing.T) {
	c := NewCanvas(3, 1)
	c.LoadText("\x1b[31mA\x1b[0mB")

	cellA := c.Get(0, 0)
	if cellA.char != "A" {
		t.Errorf("(0,0).char = %q, want A", cellA.char)
	}
	if cellA.foregroundColor != "red" {
		t.Errorf("(0,0).fg = %q, want red", cellA.foregroundColor)
	}

	cellB := c.Get(0, 1)
	if cellB.char != "B" {
		t.Errorf("(0,1).char = %q, want B", cellB.char)
	}
	if cellB.foregroundColor != "white" {
		t.Errorf("(0,1).fg = %q, want white (reset)", cellB.foregroundColor)
	}
}

func TestLoadTextANSIBackground(t *testing.T) {
	c := NewCanvas(2, 1)
	c.LoadText("\x1b[31;44mX\x1b[0m")

	cell := c.Get(0, 0)
	if cell.char != "X" {
		t.Errorf("(0,0).char = %q, want X", cell.char)
	}
	if cell.foregroundColor != "red" {
		t.Errorf("(0,0).fg = %q, want red", cell.foregroundColor)
	}
	if cell.backgroundColor != "blue" {
		t.Errorf("(0,0).bg = %q, want blue", cell.backgroundColor)
	}
}

func TestLoadTextANSIDoesNotCountAsColumns(t *testing.T) {
	c := NewCanvas(3, 1)
	c.LoadText("\x1b[31mA\x1b[0m \x1b[34mB\x1b[0m")

	if cell := c.Get(0, 0); cell.char != "A" || cell.foregroundColor != "red" {
		t.Errorf("(0,0) = %+v, want A/red", cell)
	}
	if cell := c.Get(0, 2); cell.char != "B" || cell.foregroundColor != "blue" {
		t.Errorf("(0,2) = %+v, want B/blue", cell)
	}
}

func TestCopyIsDeep(t *testing.T) {
	orig := NewCanvas(3, 3)
	orig.Set(1, 1, "X", "red", "blue")

	cp := orig.Copy()

	cp.Set(1, 1, "Y", "green", "yellow")

	origCell := orig.Get(1, 1)
	if origCell.char != "X" {
		t.Errorf("original mutated after copy modification: char = %q, want X", origCell.char)
	}

	cpCell := cp.Get(1, 1)
	if cpCell.char != "Y" {
		t.Errorf("copy cell = %q, want Y", cpCell.char)
	}
}

func TestCopyPreservesDimensions(t *testing.T) {
	orig := NewCanvas(5, 3)
	cp := orig.Copy()

	if cp.width != orig.width || cp.height != orig.height {
		t.Errorf("copy dimensions %dx%d, want %dx%d", cp.width, cp.height, orig.width, orig.height)
	}
}

func TestEqualsIdentical(t *testing.T) {
	a := NewCanvas(3, 3)
	b := NewCanvas(3, 3)
	a.Set(0, 0, "X", "red", "blue")
	b.Set(0, 0, "X", "red", "blue")

	if !a.Equals(b) {
		t.Error("identical canvases should be equal")
	}
}

func TestEqualsDifferentChar(t *testing.T) {
	a := NewCanvas(3, 3)
	b := NewCanvas(3, 3)
	a.Set(0, 0, "X", "red", "blue")
	b.Set(0, 0, "Y", "red", "blue")

	if a.Equals(b) {
		t.Error("canvases with different chars should not be equal")
	}
}

func TestEqualsDifferentFg(t *testing.T) {
	a := NewCanvas(3, 3)
	b := NewCanvas(3, 3)
	a.Set(0, 0, "X", "red", "blue")
	b.Set(0, 0, "X", "green", "blue")

	if a.Equals(b) {
		t.Error("canvases with different foreground should not be equal")
	}
}

func TestEqualsDifferentBg(t *testing.T) {
	a := NewCanvas(3, 3)
	b := NewCanvas(3, 3)
	a.Set(0, 0, "X", "red", "blue")
	b.Set(0, 0, "X", "red", "green")

	if a.Equals(b) {
		t.Error("canvases with different background should not be equal")
	}
}

func TestEqualsDifferentDimensions(t *testing.T) {
	a := NewCanvas(3, 3)
	b := NewCanvas(4, 3)

	if a.Equals(b) {
		t.Error("canvases with different widths should not be equal")
	}

	c := NewCanvas(3, 4)
	if a.Equals(c) {
		t.Error("canvases with different heights should not be equal")
	}
}
