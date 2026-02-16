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
