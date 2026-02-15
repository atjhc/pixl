package main

import "testing"

func TestNormalizeRect(t *testing.T) {
	tests := []struct {
		name                           string
		y1, x1, y2, x2                int
		wantMinY, wantMinX, wantMaxY, wantMaxX int
	}{
		{"already normalized", 1, 2, 3, 4, 1, 2, 3, 4},
		{"swapped Y", 3, 2, 1, 4, 1, 2, 3, 4},
		{"swapped X", 1, 4, 3, 2, 1, 2, 3, 4},
		{"swapped both", 3, 4, 1, 2, 1, 2, 3, 4},
		{"equal coords", 2, 2, 2, 2, 2, 2, 2, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minY, minX, maxY, maxX := normalizeRect(tt.y1, tt.x1, tt.y2, tt.x2)
			if minY != tt.wantMinY || minX != tt.wantMinX || maxY != tt.wantMaxY || maxX != tt.wantMaxX {
				t.Errorf("normalizeRect(%d,%d,%d,%d) = (%d,%d,%d,%d), want (%d,%d,%d,%d)",
					tt.y1, tt.x1, tt.y2, tt.x2,
					minY, minX, maxY, maxX,
					tt.wantMinY, tt.wantMinX, tt.wantMaxY, tt.wantMaxX)
			}
		})
	}
}

func newTestModel(width, height int) *model {
	return &model{
		canvas:          NewCanvas(width, height),
		selectedChar:    "#",
		foregroundColor: "white",
		backgroundColor: "transparent",
	}
}

func TestDrawRectangle(t *testing.T) {
	m := newTestModel(5, 5)
	m.drawRectangle(1, 1, 3, 3)

	// Border cells should be set
	borderCells := [][2]int{
		{1, 1}, {1, 2}, {1, 3},
		{2, 1}, {2, 3},
		{3, 1}, {3, 2}, {3, 3},
	}
	for _, pos := range borderCells {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != "#" {
			t.Errorf("border cell(%d,%d) = %v, want #", pos[0], pos[1], cell)
		}
	}

	// Interior should be untouched
	interior := m.canvas.Get(2, 2)
	if interior.char != " " {
		t.Errorf("interior cell(2,2).char = %q, want space", interior.char)
	}

	// Outside should be untouched
	outside := m.canvas.Get(0, 0)
	if outside.char != " " {
		t.Errorf("outside cell(0,0).char = %q, want space", outside.char)
	}
}

func TestDrawRectangleSingleCell(t *testing.T) {
	m := newTestModel(5, 5)
	m.drawRectangle(2, 2, 2, 2)

	cell := m.canvas.Get(2, 2)
	if cell.char != "#" {
		t.Errorf("single-cell rect: char = %q, want #", cell.char)
	}

	// Adjacent cells untouched
	if adj := m.canvas.Get(2, 1); adj.char != " " {
		t.Errorf("adjacent cell should be space, got %q", adj.char)
	}
}

func TestFloodFillRegion(t *testing.T) {
	m := newTestModel(5, 5)

	// Create a boundary with a different character
	m.canvas.Set(0, 2, "X", "red", "transparent")
	m.canvas.Set(1, 2, "X", "red", "transparent")
	m.canvas.Set(2, 2, "X", "red", "transparent")
	m.canvas.Set(2, 0, "X", "red", "transparent")
	m.canvas.Set(2, 1, "X", "red", "transparent")

	m.floodFill(0, 0)

	// Cells in the bounded region (0,0)-(1,1) should be filled
	for _, pos := range [][2]int{{0, 0}, {0, 1}, {1, 0}, {1, 1}} {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell.char != "#" {
			t.Errorf("fill region cell(%d,%d).char = %q, want #", pos[0], pos[1], cell.char)
		}
	}

	// Boundary cells should be unchanged
	if cell := m.canvas.Get(0, 2); cell.char != "X" {
		t.Errorf("boundary cell(0,2).char = %q, want X", cell.char)
	}

	// Cells beyond boundary should be untouched (still default space)
	if cell := m.canvas.Get(0, 3); cell.char != " " {
		t.Errorf("cell beyond boundary (0,3).char = %q, want space", cell.char)
	}
}

func TestFloodFillNoOpWhenSameColor(t *testing.T) {
	m := newTestModel(3, 3)
	m.selectedChar = " "
	m.foregroundColor = "white"
	m.backgroundColor = "transparent"

	before := m.canvas.Copy()
	m.floodFill(1, 1)

	if !m.canvas.Equals(before) {
		t.Error("flood fill with matching color should be a no-op")
	}
}

func TestFloodFillOutOfBounds(t *testing.T) {
	m := newTestModel(3, 3)
	// Should not panic
	m.floodFill(-1, 0)
	m.floodFill(0, -1)
	m.floodFill(3, 0)
	m.floodFill(0, 3)
}

func TestGetCirclePointsSymmetry(t *testing.T) {
	m := newTestModel(20, 20)

	// Ellipse centered at (5,5) with bounding box (2,2)-(8,8)
	points := m.getCirclePoints(2, 2, 8, 8, false)

	if len(points) == 0 {
		t.Fatal("expected some circle points")
	}

	centerY, centerX := 5, 5

	// Every point should have its horizontal and vertical mirror
	for pt := range points {
		dy := pt[0] - centerY
		dx := pt[1] - centerX

		mirrorH := [2]int{centerY + dy, centerX - dx}
		mirrorV := [2]int{centerY - dy, centerX + dx}

		if !points[mirrorH] {
			t.Errorf("missing horizontal mirror for %v: expected %v", pt, mirrorH)
		}
		if !points[mirrorV] {
			t.Errorf("missing vertical mirror for %v: expected %v", pt, mirrorV)
		}
	}
}

func TestGetCirclePointsSinglePoint(t *testing.T) {
	m := newTestModel(10, 10)
	points := m.getCirclePoints(3, 3, 3, 3, false)

	if len(points) != 1 {
		t.Errorf("degenerate ellipse: got %d points, want 1", len(points))
	}
	if !points[[2]int{3, 3}] {
		t.Error("degenerate ellipse should contain center point")
	}
}

func TestDrawCircleSetsCanvas(t *testing.T) {
	m := newTestModel(10, 10)
	m.drawCircle(2, 2, 6, 6, false)

	points := m.getCirclePoints(2, 2, 6, 6, false)
	for pt := range points {
		cell := m.canvas.Get(pt[0], pt[1])
		if cell == nil || cell.char != "#" {
			t.Errorf("drawCircle: cell(%d,%d) not set", pt[0], pt[1])
		}
	}
}

func TestGetEllipsePointsVerticalLine(t *testing.T) {
	m := newTestModel(10, 10)
	// rx=0 triggers the vertical line branch
	points := m.getCirclePoints(2, 3, 6, 3, false)

	if len(points) == 0 {
		t.Fatal("expected vertical line points")
	}

	for pt := range points {
		if pt[1] != 3 {
			t.Errorf("vertical line point %v has unexpected x", pt)
		}
	}
}
