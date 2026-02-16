package main

import "github.com/charmbracelet/lipgloss"

type Tool interface {
	Name() string
	DisplayName(m *model) string
	CursorChar(m *model) string

	OnPress(m *model, y, x int)
	OnDrag(m *model, y, x int)
	OnRelease(m *model, y, x int)
	OnKeyPress(m *model, key string) bool

	RenderPreview(m *model, row, col int) (string, bool)
	ModifiesCanvas() bool
}

var toolRegistry = []Tool{
	PointTool{},
	RectangleTool{},
	BoxTool{},
	EllipseTool{},
	LineTool{},
	FillTool{},
	SelectTool{},
}

func (m *model) tool() Tool {
	for _, t := range toolRegistry {
		if t.Name() == m.selectedTool {
			return t
		}
	}
	return toolRegistry[0]
}

// PointTool draws individual characters on drag.
type PointTool struct{}

func (t PointTool) Name() string                { return "Point" }
func (t PointTool) DisplayName(_ *model) string { return "Points" }
func (t PointTool) CursorChar(_ *model) string    { return "" }
func (t PointTool) ModifiesCanvas() bool         { return true }

func (t PointTool) OnPress(_ *model, _, _ int)              {}
func (t PointTool) OnKeyPress(_ *model, _ string) bool      { return false }
func (t PointTool) RenderPreview(_ *model, _, _ int) (string, bool) { return "", false }

func (t PointTool) OnDrag(m *model, y, x int) {
	if y >= 0 && y < m.canvas.height && x >= 0 && x < m.canvas.width {
		m.canvas.Set(y, x, m.selectedChar, m.foregroundColor, m.backgroundColor)
	}
}

func (t PointTool) OnRelease(_ *model, _, _ int) {}

// RectangleTool draws rectangle outlines.
type RectangleTool struct{}

func (t RectangleTool) Name() string                { return "Rectangle" }
func (t RectangleTool) DisplayName(_ *model) string { return "Rectangle" }
func (t RectangleTool) CursorChar(_ *model) string    { return "" }
func (t RectangleTool) ModifiesCanvas() bool         { return true }
func (t RectangleTool) OnKeyPress(_ *model, _ string) bool { return false }

func (t RectangleTool) OnPress(m *model, y, x int) {
	m.showPreview = true
	m.previewEndX = x
	m.previewEndY = y
}

func (t RectangleTool) OnDrag(m *model, y, x int) {
	clampedY, clampedX := m.clampToCanvas(y, x)
	m.previewEndX = clampedX
	m.previewEndY = clampedY
}

func (t RectangleTool) OnRelease(m *model, y, x int) {
	m.drawRectangle(m.startY, m.startX, y, x)
}

func (t RectangleTool) RenderPreview(m *model, row, col int) (string, bool) {
	minY, minX, maxY, maxX := normalizeRect(m.startY, m.startX, m.previewEndY, m.previewEndX)
	if row >= minY && row <= maxY && col >= minX && col <= maxX {
		if row == minY || row == maxY || col == minX || col == maxX {
			return m.styledChar(), true
		}
	}
	return "", false
}

// BoxTool draws box-drawing rectangles with distinct corner and edge characters.
type BoxTool struct{}

type boxStyle struct {
	name                   string
	h, v, tl, tr, bl, br  string
}

var boxStyles = []boxStyle{
	{"Single", "─", "│", "┌", "┐", "└", "┘"},
	{"Double", "═", "║", "╔", "╗", "╚", "╝"},
	{"Rounded", "─", "│", "╭", "╮", "╰", "╯"},
	{"Heavy", "━", "┃", "┏", "┓", "┗", "┛"},
	{"Dashed", "┄", "┆", "┌", "┐", "└", "┘"},
}

func (t BoxTool) Name() string { return "Box" }
func (t BoxTool) DisplayName(m *model) string {
	s := boxStyles[m.boxStyle]
	return s.tl + s.h + s.tr + " Box"
}
func (t BoxTool) CursorChar(m *model) string  { return boxStyles[m.boxStyle].tl }
func (t BoxTool) ModifiesCanvas() bool        { return true }

func (t BoxTool) OnKeyPress(m *model, key string) bool {
	if key != "enter" {
		return false
	}
	m.boxStyle = (m.boxStyle + 1) % len(boxStyles)
	return true
}

func (t BoxTool) OnPress(m *model, y, x int) {
	m.showPreview = true
	m.previewEndX = x
	m.previewEndY = y
}

func (t BoxTool) OnDrag(m *model, y, x int) {
	clampedY, clampedX := m.clampToCanvas(y, x)
	m.previewEndX = clampedX
	m.previewEndY = clampedY
}

func (t BoxTool) OnRelease(m *model, y, x int) {
	m.drawBox(m.startY, m.startX, y, x)
}

func (t BoxTool) RenderPreview(m *model, row, col int) (string, bool) {
	minY, minX, maxY, maxX := normalizeRect(m.startY, m.startX, m.previewEndY, m.previewEndX)
	if row < minY || row > maxY || col < minX || col > maxX {
		return "", false
	}
	if row != minY && row != maxY && col != minX && col != maxX {
		return "", false
	}

	s := boxStyles[m.boxStyle]
	var ch string
	switch {
	case row == minY && col == minX:
		ch = s.tl
	case row == minY && col == maxX:
		ch = s.tr
	case row == maxY && col == minX:
		ch = s.bl
	case row == maxY && col == maxX:
		ch = s.br
	case row == minY || row == maxY:
		ch = s.h
	default:
		ch = s.v
	}

	style := colorStyleByName(m.foregroundColor)
	if m.backgroundColor != "transparent" {
		style = style.Background(colorStyleByName(m.backgroundColor).GetForeground())
	}
	return style.Render(ch), true
}

// EllipseTool draws ellipses/circles.
type EllipseTool struct{}

type ellipseMode struct {
	name       string
	isCircle   bool
}

var ellipseModes = []ellipseMode{
	{"Ellipse", false},
	{"Oval", true},
}

type drawingToolOption struct {
	name       string
	toolName   string
	circleMode bool
}

var drawingToolOptions = []drawingToolOption{
	{"Points", "Point", false},
	{"Rectangle", "Rectangle", false},
	{"Ellipse", "Ellipse", false},
	{"Oval", "Ellipse", true},
	{"Line", "Line", false},
}

func isDrawingTool(name string) bool {
	for _, opt := range drawingToolOptions {
		if opt.toolName == name {
			return true
		}
	}
	return false
}

func (m *model) ellipseModeIndex() int {
	for i, mode := range ellipseModes {
		if mode.isCircle == m.circleMode {
			return i
		}
	}
	return 0
}

func (t EllipseTool) Name() string { return "Ellipse" }
func (t EllipseTool) DisplayName(m *model) string {
	return ellipseModes[m.ellipseModeIndex()].name
}
func (t EllipseTool) CursorChar(_ *model) string { return "" }
func (t EllipseTool) ModifiesCanvas() bool { return true }

func (t EllipseTool) OnKeyPress(m *model, key string) bool {
	if key == "enter" {
		m.circleMode = !m.circleMode
		return true
	}
	return false
}

func (t EllipseTool) OnPress(m *model, y, x int) {
	m.showPreview = true
	m.previewEndX = x
	m.previewEndY = y
	m.previewPoints = m.getCirclePoints(m.startY, m.startX, m.previewEndY, m.previewEndX, m.circleMode || m.optionKeyHeld)
}

func (t EllipseTool) OnDrag(m *model, y, x int) {
	clampedY, clampedX := m.clampToCanvas(y, x)
	m.previewEndX = clampedX
	m.previewEndY = clampedY
	m.previewPoints = m.getCirclePoints(m.startY, m.startX, m.previewEndY, m.previewEndX, m.circleMode || m.optionKeyHeld)
}

func (t EllipseTool) OnRelease(m *model, y, x int) {
	m.drawCircle(m.startY, m.startX, y, x, m.circleMode || m.optionKeyHeld)
}

func (t EllipseTool) RenderPreview(m *model, row, col int) (string, bool) {
	if m.previewPoints[[2]int{row, col}] {
		return m.styledChar(), true
	}
	return "", false
}

// LineTool draws lines using Bresenham's algorithm.
type LineTool struct{}

func (t LineTool) Name() string                { return "Line" }
func (t LineTool) DisplayName(_ *model) string { return "Line" }
func (t LineTool) CursorChar(_ *model) string    { return "" }
func (t LineTool) ModifiesCanvas() bool         { return true }
func (t LineTool) OnKeyPress(_ *model, _ string) bool { return false }

func (t LineTool) OnPress(m *model, y, x int) {
	m.showPreview = true
	m.previewEndX = x
	m.previewEndY = y
	m.previewPoints = getLinePoints(m.startY, m.startX, m.previewEndY, m.previewEndX)
}

func (t LineTool) OnDrag(m *model, y, x int) {
	clampedY, clampedX := m.clampToCanvas(y, x)
	m.previewEndX = clampedX
	m.previewEndY = clampedY
	m.previewPoints = getLinePoints(m.startY, m.startX, m.previewEndY, m.previewEndX)
}

func (t LineTool) OnRelease(m *model, y, x int) {
	m.drawLine(m.startY, m.startX, y, x)
}

func (t LineTool) RenderPreview(m *model, row, col int) (string, bool) {
	if m.previewPoints[[2]int{row, col}] {
		return m.styledChar(), true
	}
	return "", false
}

// FillTool performs flood fill.
type FillTool struct{}

func (t FillTool) Name() string                { return "Fill" }
func (t FillTool) DisplayName(_ *model) string { return "Fill" }
func (t FillTool) CursorChar(_ *model) string    { return "" }
func (t FillTool) ModifiesCanvas() bool         { return true }
func (t FillTool) OnKeyPress(_ *model, _ string) bool { return false }
func (t FillTool) OnPress(_ *model, _, _ int)          {}
func (t FillTool) OnDrag(_ *model, _, _ int)           {}
func (t FillTool) RenderPreview(_ *model, _, _ int) (string, bool) { return "", false }

func (t FillTool) OnRelease(m *model, y, x int) {
	m.floodFill(y, x)
}

// SelectTool creates selection rectangles.
type SelectTool struct{}

func (t SelectTool) Name() string                { return "Select" }
func (t SelectTool) DisplayName(_ *model) string { return "Select" }
func (t SelectTool) CursorChar(_ *model) string    { return "┼" }
func (t SelectTool) ModifiesCanvas() bool         { return false }
func (t SelectTool) OnKeyPress(_ *model, _ string) bool { return false }

func (t SelectTool) OnPress(m *model, y, x int) {
	m.showPreview = true
	m.previewEndX = x
	m.previewEndY = y
	m.hasSelection = false
}

func (t SelectTool) OnDrag(m *model, y, x int) {
	clampedY, clampedX := m.clampToCanvas(y, x)
	m.previewEndX = clampedX
	m.previewEndY = clampedY
}

func (t SelectTool) OnRelease(m *model, y, x int) {
	dy := m.startY - y
	dx := m.startX - x
	if dy < 0 {
		dy = -dy
	}
	if dx < 0 {
		dx = -dx
	}
	if dy > 1 && dx > 1 {
		m.hasSelection = true
		m.selectionStartY = m.startY
		m.selectionStartX = m.startX
		m.selectionEndY = y
		m.selectionEndX = x
	}
}

func (t SelectTool) RenderPreview(m *model, row, col int) (string, bool) {
	minY, minX, maxY, maxX := normalizeRect(m.startY, m.startX, m.previewEndY, m.previewEndX)
	hasWidth := minX != maxX
	hasHeight := minY != maxY
	if !hasWidth || !hasHeight || row < minY || row > maxY || col < minX || col > maxX {
		return "", false
	}
	if row != minY && row != maxY && col != minX && col != maxX {
		return "", false
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	var char string
	switch {
	case minY == maxY && minX == maxX:
		char = "□"
	case row == minY && col == minX:
		char = "┌"
	case row == minY && col == maxX:
		char = "┐"
	case row == maxY && col == minX:
		char = "└"
	case row == maxY && col == maxX:
		char = "┘"
	case row == minY || row == maxY:
		char = "┈"
	default:
		char = "┊"
	}
	return dimStyle.Render(char), true
}
