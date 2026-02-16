package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func testShapesPickerClickTargets(t *testing.T, canvasW, canvasH, termW, termH, fixedW, fixedH int) {
	t.Helper()
	for catIdx, group := range characterGroups {
		t.Run(fmt.Sprintf("category_%d_%s", catIdx, group.name), func(t *testing.T) {
			m := &model{
				canvas:           NewCanvas(canvasW, canvasH),
				selectedChar:     "●",
				foregroundColor:  "white",
				backgroundColor:  "transparent",
				selectedTool:     "Point",
				drawingTool:      "Point",
				width:            termW,
				height:           termH,
				ready:            true,
				showCharPicker:   true,
				selectedCategory: catIdx,
				fixedWidth:       fixedW,
				fixedHeight:      fixedH,
			}

			// Calculate where the view renders the shapes panel
			popup := m.renderCategoryPicker()
			popupLines := strings.Split(popup, "\n")
			popupX := m.toolbarShapeItemX - pickerContentOffset

			popup2 := m.renderShapesPicker()
			popup2Lines := strings.Split(popup2, "\n")

			screenRows := m.height - controlBarHeight
			if !m.hasFixedSize() {
				screenRows = m.canvas.height
			}

			popup2StartY := catIdx
			if popup2StartY+len(popup2Lines) > screenRows {
				popup2StartY = screenRows - len(popup2Lines)
			}
			if popup2StartY < 0 {
				popup2StartY = 0
			}

			categoryWidth := 0
			if len(popupLines) > 0 {
				categoryWidth = lipgloss.Width(popupLines[0])
			}
			popup2X := popupX + categoryWidth

			clickX := popup2X + 2

			for shapeIdx, expectedChar := range group.chars {
				clickY := controlBarHeight + popup2StartY + 1 + shapeIdx

				m2 := &model{
					canvas:           NewCanvas(canvasW, canvasH),
					selectedChar:     "●",
					foregroundColor:  "white",
					backgroundColor:  "transparent",
					selectedTool:     "Point",
					drawingTool:      "Point",
					width:            termW,
					height:           termH,
					ready:            true,
					showCharPicker:   true,
					selectedCategory: catIdx,
					fixedWidth:       fixedW,
					fixedHeight:      fixedH,
				}

				msg := tea.MouseMsg{
					X:    clickX,
					Y:    clickY,
					Type: tea.MouseLeft,
				}

				m2.handleMouse(msg)

				if m2.selectedChar != expectedChar {
					t.Errorf("shape[%d] at screen (%d,%d): got %q, want %q (popup2StartY=%d)",
						shapeIdx, clickX, clickY, m2.selectedChar, expectedChar, popup2StartY)
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

	toolPopup := m.renderToolPicker()
	toolPopupLines := strings.Split(toolPopup, "\n")
	toolPopupX := m.toolbarToolItemX - pickerContentOffset
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

	submenuLeft := toolPopupX + toolPickerWidth
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

	toolPopup := m.renderToolPicker()
	toolPopupLines := strings.Split(toolPopup, "\n")
	toolPopupX := m.toolbarToolItemX - pickerContentOffset
	toolPickerWidth := 0
	if len(toolPopupLines) > 0 {
		toolPickerWidth = lipgloss.Width(toolPopupLines[0])
	}

	stylePopup := m.renderBoxStylePicker()
	stylePopupLines := strings.Split(stylePopup, "\n")

	screenRows := m.height - controlBarHeight
	if !m.hasFixedSize() {
		screenRows = m.canvas.height
	}
	pickerIdx := m.toolPickerIndex()
	popup2StartY := pickerIdx
	if popup2StartY+len(stylePopupLines) > screenRows {
		popup2StartY = screenRows - len(stylePopupLines)
	}
	if popup2StartY < 0 {
		popup2StartY = 0
	}

	stylePickerLeft := toolPopupX + toolPickerWidth
	stylePickerTop := controlBarHeight + popup2StartY

	for styleIdx, bs := range boxStyles {
		clickX := stylePickerLeft + 2
		clickY := stylePickerTop + 1 + styleIdx

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

		msg := tea.MouseMsg{
			X:    clickX,
			Y:    clickY,
			Type: tea.MouseLeft,
		}
		m2.handleMouse(msg)

		if m2.boxStyle != styleIdx {
			t.Errorf("clicking style %q at (%d,%d): got boxStyle=%d, want %d",
				bs.name, clickX, clickY, m2.boxStyle, styleIdx)
		}
	}
}

func TestShapesPickerClickTargets(t *testing.T) {
	t.Run("non-fixed 80x30", func(t *testing.T) {
		testShapesPickerClickTargets(t, 80, 30, 80, 31, 0, 0)
	})
	t.Run("fixed 20x20 in 80x40", func(t *testing.T) {
		testShapesPickerClickTargets(t, 20, 20, 80, 40, 20, 20)
	})
	t.Run("fixed 10x10 in 80x30", func(t *testing.T) {
		testShapesPickerClickTargets(t, 10, 10, 80, 30, 10, 10)
	})
	t.Run("small non-fixed 80x20", func(t *testing.T) {
		testShapesPickerClickTargets(t, 80, 20, 80, 21, 0, 0)
	})
}
