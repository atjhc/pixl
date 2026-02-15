package main

import "testing"

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
