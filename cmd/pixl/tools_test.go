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
		config:          Config{MergeBoxBorders: true},
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

func TestDrawBoxSingleStyle(t *testing.T) {
	m := newTestModel(6, 6)
	m.boxStyle = 0
	m.drawBox(1, 1, 3, 4)

	expect := map[[2]int]string{
		{1, 1}: "┌", {1, 2}: "─", {1, 3}: "─", {1, 4}: "┐",
		{2, 1}: "│", {2, 4}: "│",
		{3, 1}: "└", {3, 2}: "─", {3, 3}: "─", {3, 4}: "┘",
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}

	// Interior untouched
	if cell := m.canvas.Get(2, 2); cell.char != " " {
		t.Errorf("interior (2,2) = %q, want space", cell.char)
	}
}

func TestDrawBoxDoubleStyle(t *testing.T) {
	m := newTestModel(6, 6)
	m.boxStyle = 1
	m.drawBox(0, 0, 2, 3)

	expect := map[[2]int]string{
		{0, 0}: "╔", {0, 1}: "═", {0, 2}: "═", {0, 3}: "╗",
		{1, 0}: "║", {1, 3}: "║",
		{2, 0}: "╚", {2, 1}: "═", {2, 2}: "═", {2, 3}: "╝",
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}
}

func TestDrawBoxReversedCoords(t *testing.T) {
	m := newTestModel(6, 6)
	m.boxStyle = 0
	m.drawBox(3, 4, 1, 1)

	// Should produce same result as (1,1)-(3,4)
	if cell := m.canvas.Get(1, 1); cell == nil || cell.char != "┌" {
		t.Errorf("reversed coords: TL = %v, want ┌", cell)
	}
	if cell := m.canvas.Get(3, 4); cell == nil || cell.char != "┘" {
		t.Errorf("reversed coords: BR = %v, want ┘", cell)
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

func TestGetLinePointsHorizontal(t *testing.T) {
	points := getLinePoints(3, 1, 3, 5)

	for x := 1; x <= 5; x++ {
		if !points[[2]int{3, x}] {
			t.Errorf("missing point (3,%d) in horizontal line", x)
		}
	}
	if len(points) != 5 {
		t.Errorf("horizontal line: got %d points, want 5", len(points))
	}
}

func TestGetLinePointsVertical(t *testing.T) {
	points := getLinePoints(1, 3, 5, 3)

	for y := 1; y <= 5; y++ {
		if !points[[2]int{y, 3}] {
			t.Errorf("missing point (%d,3) in vertical line", y)
		}
	}
	if len(points) != 5 {
		t.Errorf("vertical line: got %d points, want 5", len(points))
	}
}

func TestGetLinePointsDiagonal(t *testing.T) {
	points := getLinePoints(0, 0, 4, 4)

	for i := 0; i <= 4; i++ {
		if !points[[2]int{i, i}] {
			t.Errorf("missing point (%d,%d) in diagonal line", i, i)
		}
	}
	if len(points) != 5 {
		t.Errorf("diagonal line: got %d points, want 5", len(points))
	}
}

func TestGetLinePointsSinglePoint(t *testing.T) {
	points := getLinePoints(2, 3, 2, 3)

	if len(points) != 1 {
		t.Errorf("single point line: got %d points, want 1", len(points))
	}
	if !points[[2]int{2, 3}] {
		t.Error("single point line should contain the point")
	}
}

func TestDrawLineSetsCanvas(t *testing.T) {
	m := newTestModel(10, 10)
	m.drawLine(1, 1, 1, 5)

	for x := 1; x <= 5; x++ {
		cell := m.canvas.Get(1, x)
		if cell == nil || cell.char != "#" {
			t.Errorf("drawLine: cell(1,%d) not set", x)
		}
	}
}

func TestDrawBoxMergesTJunctionsSingle(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 0

	// Draw first box
	m.drawBox(0, 0, 4, 4)
	// Draw adjacent box sharing right edge
	m.drawBox(0, 4, 4, 8)

	// Shared edge at x=4 should have T-junctions
	expect := map[[2]int]string{
		{0, 4}: "┬", // was ┐ + ┌
		{1, 4}: "│", // unchanged vertical
		{2, 4}: "│",
		{3, 4}: "│",
		{4, 4}: "┴", // was ┘ + └
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}
}

func TestDrawBoxMergesTJunctionsDouble(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 1

	m.drawBox(0, 0, 4, 4)
	m.drawBox(0, 4, 4, 8)

	expect := map[[2]int]string{
		{0, 4}: "╦",
		{4, 4}: "╩",
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}
}

func TestDrawBoxMergesTJunctionsHeavy(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 3

	m.drawBox(0, 0, 4, 4)
	m.drawBox(0, 4, 4, 8)

	expect := map[[2]int]string{
		{0, 4}: "┳",
		{4, 4}: "┻",
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}
}

func TestDrawBoxDashedNoMerge(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 4 // Dashed

	m.drawBox(0, 0, 4, 4)
	m.drawBox(0, 4, 4, 8)

	// Dashed has no T-junction chars, so second box overwrites
	cell := m.canvas.Get(0, 4)
	if cell == nil || cell.char != "┌" {
		got := ""
		if cell != nil {
			got = cell.char
		}
		t.Errorf("Dashed: cell(0,4) = %q, want ┌ (no merge)", got)
	}
}

func TestDrawBoxMergeDisabledByConfig(t *testing.T) {
	m := newTestModel(10, 10)
	m.config.MergeBoxBorders = false
	m.boxStyle = 0

	m.drawBox(0, 0, 4, 4)
	m.drawBox(0, 4, 4, 8)

	// With merge disabled, second box overwrites with ┌
	cell := m.canvas.Get(0, 4)
	if cell == nil || cell.char != "┌" {
		got := ""
		if cell != nil {
			got = cell.char
		}
		t.Errorf("merge disabled: cell(0,4) = %q, want ┌", got)
	}
}

func TestDrawBoxMergesVerticalAdjacent(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 0

	m.drawBox(0, 0, 3, 4)
	m.drawBox(3, 0, 6, 4)

	// Shared edge at y=3 should have T-junctions
	expect := map[[2]int]string{
		{3, 0}: "├", // was └ + ┌
		{3, 4}: "┤", // was ┘ + ┐
		{3, 1}: "─", // unchanged horizontal
		{3, 2}: "─",
		{3, 3}: "─",
	}
	for pos, want := range expect {
		cell := m.canvas.Get(pos[0], pos[1])
		if cell == nil || cell.char != want {
			got := ""
			if cell != nil {
				got = cell.char
			}
			t.Errorf("cell(%d,%d) = %q, want %q", pos[0], pos[1], got, want)
		}
	}
}

func TestDrawBoxMergesCross(t *testing.T) {
	m := newTestModel(10, 10)
	m.boxStyle = 0

	// 2x2 grid of boxes sharing a center point
	m.drawBox(0, 0, 3, 3)
	m.drawBox(0, 3, 3, 6)
	m.drawBox(3, 0, 6, 3)
	m.drawBox(3, 3, 6, 6)

	cell := m.canvas.Get(3, 3)
	if cell == nil || cell.char != "┼" {
		got := ""
		if cell != nil {
			got = cell.char
		}
		t.Errorf("center cross: cell(3,3) = %q, want ┼", got)
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
