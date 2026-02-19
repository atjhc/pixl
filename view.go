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
	var popup3 string
	var popup3Lines []string
	var popup3StartY int
	var popup3X int
	var popup4 string
	var popup4Lines []string
	var popup4StartY int
	var popup4X int

	if m.showFgPicker {
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

		if m.toolHasSubmenu() {
			popup2 = m.renderToolSubmenuPicker()
			popup2Lines = strings.Split(popup2, "\n")
			pickerIdx := m.toolPickerIndex()
			popup2StartY = popupStartY + pickerIdx
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
			popupLines, popup2Lines = mergePopupBorders(popupLines, popup2Lines, popup2StartY-popupStartY)
			popup2X = popupX + toolPickerWidth - 1

			if m.toolHasGlyphPicker() {
				popup3 = m.renderCategoryPicker()
				popup3Lines = strings.Split(popup3, "\n")
				// Align with the Glyphs row (index 0 in submenu)
				popup3StartY = popup2StartY
				if popup3StartY+len(popup3Lines) > screenRows {
					popup3StartY = screenRows - len(popup3Lines)
				}
				if popup3StartY < 0 {
					popup3StartY = 0
				}
				popup2Width := 0
				if len(popup2Lines) > 0 {
					popup2Width = lipgloss.Width(popup2Lines[0])
				}
				popup2Lines, popup3Lines = mergePopupBorders(popup2Lines, popup3Lines, popup3StartY-popup2StartY)
				popup3X = popup2X + popup2Width - 1

				popup4 = m.renderGlyphsPicker()
				popup4Lines = strings.Split(popup4, "\n")
				popup4StartY = popup3StartY + m.selectedCategory
				if popup4StartY+len(popup4Lines) > screenRows {
					popup4StartY = screenRows - len(popup4Lines)
				}
				if popup4StartY < 0 {
					popup4StartY = 0
				}
				popup3Width := 0
				if len(popup3Lines) > 0 {
					popup3Width = lipgloss.Width(popup3Lines[0])
				}
				popup3Lines, popup4Lines = mergePopupBorders(popup3Lines, popup4Lines, popup4StartY-popup3StartY)
				popup4X = popup3X + popup3Width - 1
			}
		}
	}

	// Render screen rows
	offY, offX := m.canvasOffset()
	borderStyle := lipgloss.NewStyle().Foreground(themeColor(m.config.Theme.CanvasBorder))
	labelStyle := lipgloss.NewStyle().Foreground(themeColor(m.config.Theme.CanvasBorder))
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

			if !inPopup && popup3Lines != nil && i >= popup3StartY && i < popup3StartY+len(popup3Lines) {
				popup3LineIdx := i - popup3StartY
				popup3Line := popup3Lines[popup3LineIdx]
				popup3Width := lipgloss.Width(popup3Line)

				if col >= popup3X && col < popup3X+popup3Width {
					if col == popup3X {
						lineBuilder.WriteString("\x1b[0m" + popup3Line)
					}
					inPopup = true
				}
			}

			if !inPopup && popup4Lines != nil && i >= popup4StartY && i < popup4StartY+len(popup4Lines) {
				popup4LineIdx := i - popup4StartY
				popup4Line := popup4Lines[popup4LineIdx]
				popup4Width := lipgloss.Width(popup4Line)

				if col >= popup4X && col < popup4X+popup4Width {
					if col == popup4X {
						lineBuilder.WriteString("\x1b[0m" + popup4Line)
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
	if m.showPreview {
		if rendered, ok := m.tool().RenderPreview(m, row, col); ok {
			return rendered
		}
	}

	if m.hasSelection {
		minY, minX, maxY, maxX := normalizeRect(m.selectionStartY, m.selectionStartX, m.selectionEndY, m.selectionEndX)
		hasWidth := minX != maxX
		hasHeight := minY != maxY
		if hasWidth && hasHeight && row >= minY && row <= maxY && col >= minX && col <= maxX {
			if row == minY || row == maxY || col == minX || col == maxX {
				highlightStyle := lipgloss.NewStyle().Foreground(themeColor(m.config.Theme.SelectionFg))
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
		ghostStyle := lipgloss.NewStyle().Foreground(themeColor(m.config.Theme.CursorFg))
		cursorChar := m.selectedChar
		if c := m.tool().CursorChar(m); c != "" {
			cursorChar = c
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
