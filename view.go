package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	canvasHeight := m.canvas.height

	screenRows := m.height - controlBarHeight
	if !m.hasFixedSize() {
		screenRows = canvasHeight
	}

	var b strings.Builder

	b.WriteString(m.renderControlBar())

	// Determine popup info
	var popup string
	var popupLines []string
	var popupStartY int
	var popupX int
	var popup2 string
	var popup2Lines []string
	var popup2StartY int
	var popup2X int

	if m.showCharPicker {
		popup = m.renderCategoryPicker()
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		popupX = m.toolbarShapeItemX - pickerContentOffset

		popup2 = m.renderShapesPicker()
		popup2Lines = strings.Split(popup2, "\n")
		popup2StartY = popupStartY + m.selectedCategory
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
		popup2X = popupX + categoryWidth
	} else if m.showFgPicker {
		popup = m.renderColorPicker("Foreground")
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		popupX = m.toolbarForegroundItemX - pickerContentOffset
	} else if m.showBgPicker {
		popup = m.renderColorPicker("Background")
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		popupX = m.toolbarBackgroundItemX - pickerContentOffset
	} else if m.showToolPicker {
		popup = m.renderToolPicker()
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		popupX = m.toolbarToolItemX - pickerContentOffset
	}

	// Render screen rows
	offY, offX := m.canvasOffset()
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	label := fmt.Sprintf("%dx%d", m.canvas.width, m.canvas.height)

	labelRow := offY
	topBorderRow := offY + 1
	canvasStartRow := offY + 2
	canvasEndRow := offY + 1 + m.canvas.height
	bottomBorderRow := offY + 2 + m.canvas.height

	for i := 0; i < screenRows; i++ {
		var lineBuilder strings.Builder

		for col := 0; col < m.width; col++ {
			inPopup := false

			if popupLines != nil && i >= popupStartY && i < popupStartY+len(popupLines) {
				popupLineIdx := i - popupStartY
				popupLine := popupLines[popupLineIdx]
				popupWidth := lipgloss.Width(popupLine)

				if col >= popupX && col < popupX+popupWidth {
					if col == popupX {
						lineBuilder.WriteString("\x1b[0m" + popupLine)
					}
					inPopup = true
				}
			}

			if !inPopup && popup2Lines != nil && i >= popup2StartY && i < popup2StartY+len(popup2Lines) {
				popup2LineIdx := i - popup2StartY
				popup2Line := popup2Lines[popup2LineIdx]
				popup2Width := lipgloss.Width(popup2Line)

				if col >= popup2X && col < popup2X+popup2Width {
					if col == popup2X {
						lineBuilder.WriteString("\x1b[0m" + popup2Line)
					}
					inPopup = true
				}
			}

			if inPopup {
				continue
			}

			if m.hasFixedSize() {
				if i == labelRow {
					labelCol := col - offX
					if labelCol >= 0 && labelCol < len(label) {
						lineBuilder.WriteString(labelStyle.Render(string(label[labelCol])))
						continue
					}
					lineBuilder.WriteString(" ")
					continue
				}

				if i == topBorderRow {
					if col == offX {
						lineBuilder.WriteString(borderStyle.Render("┌"))
						continue
					} else if col == offX+m.canvas.width+1 {
						lineBuilder.WriteString(borderStyle.Render("┐"))
						continue
					} else if col > offX && col < offX+m.canvas.width+1 {
						lineBuilder.WriteString(borderStyle.Render("─"))
						continue
					}
					lineBuilder.WriteString(" ")
					continue
				}

				if i == bottomBorderRow {
					if col == offX {
						lineBuilder.WriteString(borderStyle.Render("└"))
						continue
					} else if col == offX+m.canvas.width+1 {
						lineBuilder.WriteString(borderStyle.Render("┘"))
						continue
					} else if col > offX && col < offX+m.canvas.width+1 {
						lineBuilder.WriteString(borderStyle.Render("─"))
						continue
					}
					lineBuilder.WriteString(" ")
					continue
				}

				if i >= canvasStartRow && i <= canvasEndRow {
					if col == offX {
						lineBuilder.WriteString(borderStyle.Render("│"))
						continue
					} else if col == offX+m.canvas.width+1 {
						lineBuilder.WriteString(borderStyle.Render("│"))
						continue
					} else if col > offX && col < offX+m.canvas.width+1 {
						canvasRow := i - canvasStartRow
						canvasCol := col - offX - 1
						lineBuilder.WriteString(m.renderCellAt(canvasRow, canvasCol))
						continue
					}
					lineBuilder.WriteString(" ")
					continue
				}

				lineBuilder.WriteString(" ")
				continue
			}

			lineBuilder.WriteString(m.renderCellAt(i, col))
		}

		if i < screenRows-1 {
			b.WriteString(lineBuilder.String() + "\n")
		} else {
			b.WriteString(lineBuilder.String())
		}
	}

	return b.String()
}

func (m *model) renderCellAt(row, col int) string {
	if m.showPreview && m.selectedTool == "Rectangle" {
		minY, minX, maxY, maxX := normalizeRect(m.startY, m.startX, m.previewEndY, m.previewEndX)
		if row >= minY && row <= maxY && col >= minX && col <= maxX {
			if row == minY || row == maxY || col == minX || col == maxX {
				return m.styledChar()
			}
		}
	} else if m.showPreview && m.selectedTool == "Ellipse" {
		if m.previewPoints[[2]int{row, col}] {
			return m.styledChar()
		}
	} else if m.showPreview && m.selectedTool == "Select" {
		minY, minX, maxY, maxX := normalizeRect(m.startY, m.startX, m.previewEndY, m.previewEndX)
		hasWidth := minX != maxX
		hasHeight := minY != maxY
		if hasWidth && hasHeight && row >= minY && row <= maxY && col >= minX && col <= maxX {
			if row == minY || row == maxY || col == minX || col == maxX {
				dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
				var char string
				if minY == maxY && minX == maxX {
					char = "□"
				} else if row == minY && col == minX {
					char = "┌"
				} else if row == minY && col == maxX {
					char = "┐"
				} else if row == maxY && col == minX {
					char = "└"
				} else if row == maxY && col == maxX {
					char = "┘"
				} else if row == minY || row == maxY {
					char = "┈"
				} else {
					char = "┊"
				}
				return dimStyle.Render(char)
			}
		}
	}

	if m.hasSelection {
		minY, minX, maxY, maxX := normalizeRect(m.selectionStartY, m.selectionStartX, m.selectionEndY, m.selectionEndX)
		hasWidth := minX != maxX
		hasHeight := minY != maxY
		if hasWidth && hasHeight && row >= minY && row <= maxY && col >= minX && col <= maxX {
			if row == minY || row == maxY || col == minX || col == maxX {
				highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
				var char string
				if row == minY && col == minX {
					char = "┌"
				} else if row == minY && col == maxX {
					char = "┐"
				} else if row == maxY && col == minX {
					char = "└"
				} else if row == maxY && col == maxX {
					char = "┘"
				} else if row == minY || row == maxY {
					char = "┈"
				} else {
					char = "┊"
				}
				return highlightStyle.Render(char)
			}
		}
	}

	if !m.mouseDown &&
		row == m.hoverRow && col == m.hoverCol &&
		row >= 0 && row < m.canvas.height && col >= 0 && col < m.canvas.width {
		ghostStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		cursorChar := m.selectedChar
		if m.selectedTool != "Point" {
			cursorChar = "┼"
		}
		return ghostStyle.Render(cursorChar)
	}

	cell := m.canvas.Get(row, col)
	if cell == nil {
		return " "
	}
	if cell.foregroundColor == "transparent" {
		return " "
	}
	style := colorStyleByName(cell.foregroundColor)
	if cell.backgroundColor != "transparent" {
		style = style.Background(colorStyleByName(cell.backgroundColor).GetForeground())
	}
	return style.Render(cell.char)
}

func (m *model) styledChar() string {
	style := colorStyleByName(m.foregroundColor)
	if m.backgroundColor != "transparent" {
		style = style.Background(colorStyleByName(m.backgroundColor).GetForeground())
	}
	return style.Render(m.selectedChar)
}

func (m *model) renderCanvas() string {
	var b strings.Builder

	for row := 0; row < m.canvas.height; row++ {
		for col := 0; col < m.canvas.width; col++ {
			cell := m.canvas.Get(row, col)
			if cell != nil {
				if cell.foregroundColor == "transparent" {
					b.WriteString(" ")
				} else {
					style := colorStyleByName(cell.foregroundColor)
					if cell.backgroundColor != "transparent" {
						style = style.Background(colorStyleByName(cell.backgroundColor).GetForeground())
					}
					b.WriteString(style.Render(cell.char))
				}
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *model) renderCanvasPlain() string {
	var b strings.Builder

	for row := 0; row < m.canvas.height; row++ {
		for col := 0; col < m.canvas.width; col++ {
			cell := m.canvas.Get(row, col)
			if cell == nil || cell.foregroundColor == "transparent" {
				b.WriteString(" ")
				continue
			}

			fg := colorToANSI(cell.foregroundColor)
			bg := colorToANSIBg(cell.backgroundColor)

			if fg == "" && bg == "" {
				b.WriteString(cell.char)
				continue
			}

			b.WriteString("\x1b[")
			if fg != "" && bg != "" {
				b.WriteString(fg + ";" + bg)
			} else if fg != "" {
				b.WriteString(fg)
			} else {
				b.WriteString(bg)
			}
			b.WriteString("m")
			b.WriteString(cell.char)
			b.WriteString("\x1b[0m")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m *model) hasFixedSize() bool {
	return m.fixedWidth > 0 && m.fixedHeight > 0
}

func (m *model) canvasOffset() (offsetY, offsetX int) {
	if !m.hasFixedSize() {
		return 0, 0
	}
	offsetY = (m.height - controlBarHeight - m.canvas.height - 3) / 2
	offsetX = (m.width - m.canvas.width - 2) / 2
	if offsetY < 0 {
		offsetY = 0
	}
	if offsetX < 0 {
		offsetX = 0
	}
	return
}
