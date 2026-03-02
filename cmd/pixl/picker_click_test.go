package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func testGlyphsPickerClickTargets(t *testing.T, canvasW, canvasH, termW, termH, fixedW, fixedH int) {
	t.Helper()
	for catIdx, group := range characterGroups {
		t.Run(fmt.Sprintf("category_%d_%s", catIdx, group.name), func(t *testing.T) {
			m := &model{
				canvas:          NewCanvas(canvasW, canvasH),
				selectedChar:    "●",
				foregroundColor: "white",
				backgroundColor: "transparent",
				selectedTool:    "Point",
				drawingTool:     "Point",
				width:           termW,
				height:          termH,
				ready:           true,
				showGlyphPicker: true,
				selectedCategory: catIdx,
				fixedWidth:      fixedW,
				fixedHeight:     fixedH,
			}

			m.renderControlBar()

			screenRows := m.height - controlBarHeight
			if !m.hasFixedSize() {
				screenRows = m.canvas.height
			}

			// Category picker (popup)
			catPopup := m.renderCategoryPicker()
			catLines := strings.Split(catPopup, "\n")
			catLeft := m.toolbar.glyphItemX - pickerContentOffset

			catWidth := 0
			if len(catLines) > 0 {
				catWidth = lipgloss.Width(catLines[0])
			}

			// Glyph picker (popup2)
			glyphPopup := m.renderGlyphsPicker()
			glyphLines := strings.Split(glyphPopup, "\n")

			glyphStartY := catIdx
			if glyphStartY+len(glyphLines) > screenRows {
				glyphStartY = screenRows - len(glyphLines)
			}
			if glyphStartY < 0 {
				glyphStartY = 0
			}
			glyphLeft := catLeft + catWidth - 1

			clickX := glyphLeft + 2

			for glyphIdx, expectedChar := range group.chars {
				clickY := controlBarHeight + glyphStartY + 1 + glyphIdx

				m2 := &model{
					canvas:          NewCanvas(canvasW, canvasH),
					selectedChar:    "●",
					foregroundColor: "white",
					backgroundColor: "transparent",
					selectedTool:    "Point",
					drawingTool:     "Point",
					width:           termW,
					height:          termH,
					ready:           true,
					showGlyphPicker: true,
					selectedCategory: catIdx,
					fixedWidth:      fixedW,
					fixedHeight:     fixedH,
				}
				m2.renderControlBar()

				msg := tea.MouseMsg{
					X:    clickX,
					Y:    clickY,
					Type: tea.MouseLeft,
				}

				m2.handleMouse(msg)

				if m2.selectedChar != expectedChar {
					t.Errorf("glyph[%d] at screen (%d,%d): got %q, want %q (glyphStartY=%d)",
						glyphIdx, clickX, clickY, m2.selectedChar, expectedChar, glyphStartY)
				}
			}
		})
	}
}

func TestDrawingToolPickerClickTargets(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(80, 30),
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Point",
		drawingTool:     "Point",
		width:           80,
		height:          31,
		ready:           true,
		showToolPicker:  true,
	}
	m.renderControlBar()

	toolPopup := m.renderToolPicker()
	toolPopupLines := strings.Split(toolPopup, "\n")
	toolPopupX := m.toolbar.toolItemX - pickerContentOffset
	toolPickerWidth := 0
	if len(toolPopupLines) > 0 {
		toolPickerWidth = lipgloss.Width(toolPopupLines[0])
	}

	submenuPopup := m.renderDrawingToolPicker()
	submenuLines := strings.Split(submenuPopup, "\n")

	screenRows := m.height - controlBarHeight
	if !m.hasFixedSize() {
		screenRows = m.canvas.height
	}
	pickerIdx := m.toolPickerIndex()
	popup2StartY := pickerIdx
	if popup2StartY+len(submenuLines) > screenRows {
		popup2StartY = screenRows - len(submenuLines)
	}
	if popup2StartY < 0 {
		popup2StartY = 0
	}

	submenuLeft := toolPopupX + toolPickerWidth - 1
	submenuTop := controlBarHeight + popup2StartY

	for optIdx, opt := range drawingToolOptions {
		clickX := submenuLeft + 2
		clickY := submenuTop + 1 + optIdx

		m2 := &model{
			canvas:          NewCanvas(80, 30),
			selectedChar:    "●",
			foregroundColor: "white",
			backgroundColor: "transparent",
			selectedTool:    "Point",
			drawingTool:     "Point",
			width:           80,
			height:          31,
			ready:           true,
			showToolPicker:  true,
		}
		m2.renderControlBar()

		msg := tea.MouseMsg{
			X:    clickX,
			Y:    clickY,
			Type: tea.MouseLeft,
		}
		m2.handleMouse(msg)

		if m2.selectedTool != opt.toolName {
			t.Errorf("clicking option %q at (%d,%d): got selectedTool=%q, want %q",
				opt.name, clickX, clickY, m2.selectedTool, opt.toolName)
		}
		if opt.toolName == "Ellipse" && m2.circleMode != opt.circleMode {
			t.Errorf("clicking option %q at (%d,%d): got circleMode=%v, want %v",
				opt.name, clickX, clickY, m2.circleMode, opt.circleMode)
		}
	}
}

func TestBoxStylePickerClickTargets(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(80, 30),
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Box",
		drawingTool:     "Point",
		width:           80,
		height:          31,
		ready:           true,
		showToolPicker:  true,
		boxStyle:        0,
	}
	m.renderControlBar()

	toolPopup := m.renderToolPicker()
	toolPopupLines := strings.Split(toolPopup, "\n")
	toolPopupX := m.toolbar.toolItemX - pickerContentOffset
	toolPickerWidth := 0
	if len(toolPopupLines) > 0 {
		toolPickerWidth = lipgloss.Width(toolPopupLines[0])
	}

	boxPopup := m.renderBoxStylePicker()
	boxLines := strings.Split(boxPopup, "\n")

	screenRows := m.height - controlBarHeight
	if !m.hasFixedSize() {
		screenRows = m.canvas.height
	}

	pickerIdx := m.toolPickerIndex()
	popup2StartY := pickerIdx
	if popup2StartY+len(boxLines) > screenRows {
		popup2StartY = screenRows - len(boxLines)
	}
	if popup2StartY < 0 {
		popup2StartY = 0
	}

	submenuLeft := toolPopupX + toolPickerWidth - 1
	submenuTop := controlBarHeight + popup2StartY

	for styleIdx, style := range boxStyles {
		clickX := submenuLeft + 2
		clickY := submenuTop + 1 + styleIdx

		m2 := &model{
			canvas:          NewCanvas(80, 30),
			selectedChar:    "●",
			foregroundColor: "white",
			backgroundColor: "transparent",
			selectedTool:    "Box",
			drawingTool:     "Point",
			width:           80,
			height:          31,
			ready:           true,
			showToolPicker:  true,
			boxStyle:        0,
		}
		m2.renderControlBar()

		msg := tea.MouseMsg{
			X:    clickX,
			Y:    clickY,
			Type: tea.MouseLeft,
		}
		m2.handleMouse(msg)

		if m2.boxStyle != styleIdx {
			t.Errorf("clicking box style %q at (%d,%d): got boxStyle=%d, want %d",
				style.name, clickX, clickY, m2.boxStyle, styleIdx)
		}
	}
}

func TestGlyphsPickerClickTargetsVariableSize(t *testing.T) {
	testGlyphsPickerClickTargets(t, 80, 30, 80, 31, 0, 0)
}

func TestGlyphsPickerClickTargetsFixedSize(t *testing.T) {
	testGlyphsPickerClickTargets(t, 40, 20, 80, 50, 40, 20)
}
