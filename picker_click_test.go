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
				showToolPicker:  true,
				selectedCategory: catIdx,
				onGlyphSelector: true,
				fixedWidth:      fixedW,
				fixedHeight:     fixedH,
			}

			// Render the toolbar to set X positions
			m.renderControlBar()

			// Calculate popup positions matching view.go logic
			popup := m.renderToolPicker()
			popupLines := strings.Split(popup, "\n")
			popupX := m.toolbarToolItemX - pickerContentOffset

			popup2 := m.renderToolSubmenuPicker()
			popup2Lines := strings.Split(popup2, "\n")
			pickerIdx := m.toolPickerIndex()

			screenRows := m.height - controlBarHeight
			if !m.hasFixedSize() {
				screenRows = m.canvas.height
			}

			popup2StartY := pickerIdx
			if popup2StartY+len(popup2Lines) > screenRows {
				popup2StartY = screenRows - len(popup2Lines)
			}
			if popup2StartY < 0 {
				popup2StartY = 0
			}
			toolPickerWidth := 0
			if len(popupLines) > 0 {
				toolPickerWidth = lipgloss.Width(popupLines[0])
			}
			popup2X := popupX + toolPickerWidth - 1

			popup2Width := 0
			if len(popup2Lines) > 0 {
				popup2Width = lipgloss.Width(popup2Lines[0])
			}

			// popup3: category picker
			popup3 := m.renderCategoryPicker()
			popup3Lines := strings.Split(popup3, "\n")

			popup3StartY := popup2StartY
			if popup3StartY+len(popup3Lines) > screenRows {
				popup3StartY = screenRows - len(popup3Lines)
			}
			if popup3StartY < 0 {
				popup3StartY = 0
			}
			popup3X := popup2X + popup2Width - 1

			popup3Width := 0
			if len(popup3Lines) > 0 {
				popup3Width = lipgloss.Width(popup3Lines[0])
			}

			// popup4: glyphs picker
			popup4 := m.renderGlyphsPicker()
			popup4Lines := strings.Split(popup4, "\n")

			popup4StartY := popup3StartY + catIdx
			if popup4StartY+len(popup4Lines) > screenRows {
				popup4StartY = screenRows - len(popup4Lines)
			}
			if popup4StartY < 0 {
				popup4StartY = 0
			}
			popup4X := popup3X + popup3Width - 1

			clickX := popup4X + 2

			for glyphIdx, expectedChar := range group.chars {
				clickY := controlBarHeight + popup4StartY + 1 + glyphIdx

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
					showToolPicker:  true,
					selectedCategory: catIdx,
					onGlyphSelector: true,
					fixedWidth:      fixedW,
					fixedHeight:     fixedH,
				}
				// Set toolbar positions
				m2.renderControlBar()

				msg := tea.MouseMsg{
					X:    clickX,
					Y:    clickY,
					Type: tea.MouseLeft,
				}

				m2.handleMouse(msg)

				if m2.selectedChar != expectedChar {
					t.Errorf("glyph[%d] at screen (%d,%d): got %q, want %q (popup4StartY=%d)",
						glyphIdx, clickX, clickY, m2.selectedChar, expectedChar, popup4StartY)
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

	submenuLeft := toolPopupX + toolPickerWidth - 1
	submenuTop := controlBarHeight + popup2StartY

	for optIdx, opt := range drawingToolOptions {
		clickX := submenuLeft + 2
		clickY := submenuTop + 1 + 1 + optIdx // +1 for Glyph selector entry at index 0

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

	stylePickerLeft := toolPopupX + toolPickerWidth - 1
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
		m2.renderControlBar()

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

func TestGlyphsPickerClickTargets(t *testing.T) {
	t.Run("non-fixed 80x30", func(t *testing.T) {
		testGlyphsPickerClickTargets(t, 80, 30, 80, 31, 0, 0)
	})
	t.Run("fixed 20x20 in 80x40", func(t *testing.T) {
		testGlyphsPickerClickTargets(t, 20, 20, 80, 40, 20, 20)
	})
	t.Run("fixed 10x10 in 80x30", func(t *testing.T) {
		testGlyphsPickerClickTargets(t, 10, 10, 80, 30, 10, 10)
	})
	t.Run("small non-fixed 80x20", func(t *testing.T) {
		testGlyphsPickerClickTargets(t, 80, 20, 80, 21, 0, 0)
	})
}
