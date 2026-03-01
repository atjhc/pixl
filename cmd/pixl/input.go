package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleResize(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	}
	return m, nil
}

func (m *model) handleResize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.ready = true

	if m.hasFixedSize() {
		if !m.canvasInitialized {
			newCanvas := NewCanvas(m.fixedWidth, m.fixedHeight)
			for row := 0; row < min(m.canvas.height, m.fixedHeight); row++ {
				for col := 0; col < min(m.canvas.width, m.fixedWidth); col++ {
					cell := m.canvas.Get(row, col)
					if cell != nil {
						newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
					}
				}
			}
			m.canvas = newCanvas
			m.canvasInitialized = true
			if len(m.history) == 0 {
				m.history = []Canvas{m.canvas.Copy()}
				m.historyIndex = 0
			}
		}
	} else {
		canvasHeight := m.height - controlBarHeight
		if canvasHeight > 0 && (canvasHeight != m.canvas.height || m.width != m.canvas.width) {
			m.saveToHistory()
			newCanvas := NewCanvas(m.width, canvasHeight)
			for row := 0; row < min(m.canvas.height, canvasHeight); row++ {
				for col := 0; col < min(m.canvas.width, m.width); col++ {
					cell := m.canvas.Get(row, col)
					if cell != nil {
						newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
					}
				}
			}
			m.canvas = newCanvas
			if len(m.history) == 0 {
				m.history = []Canvas{m.canvas.Copy()}
				m.historyIndex = 0
			}
		}
	}
	return m, nil
}

func (m *model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.optionKeyHeld = msg.Alt

	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.String()[0] - '1')

		if m.showFgPicker {
			if idx < len(colors) {
				m.foregroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showBgPicker {
			if idx < len(colors) {
				m.backgroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showToolPicker && m.toolPickerFocusLevel == 3 {
			if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
				chars := characterGroups[m.selectedCategory].chars
				if idx < len(chars) {
					m.selectedChar = chars[idx]
					return m, nil
				}
			}
		} else if m.showToolPicker && m.toolPickerFocusLevel == 2 {
			if idx < len(characterGroups) {
				m.selectedCategory = idx
				return m, nil
			}
		} else if m.showToolPicker && m.toolPickerFocusLevel == 1 {
			if idx < m.toolSubmenuCount() {
				m.setToolSubmenuIndex(idx)
				return m, nil
			}
		} else if m.showToolPicker {
			items := m.toolPickerItems()
			if idx < len(items) {
				m.setToolPickerIndex(idx)
				return m, nil
			}
		}
	case "ctrl+c", "q":
		return m, tea.Quit
	case "c":
		m.canvas.Clear()
		m.saveToHistory()
		return m, nil
	case "u":
		m.undo()
		return m, nil
	case "r":
		m.redo()
		return m, nil
	case "y":
		m.copySelection()
		return m, nil
	case "d":
		m.cutSelection()
		return m, nil
	case "p":
		m.paste()
		return m, nil
	case "f":
		if m.showFgPicker {
			m.closeMenus()
		} else {
			m.openMenu(0)
		}
		return m, nil
	case "b":
		if m.showBgPicker {
			m.closeMenus()
		} else {
			m.openMenu(1)
		}
		return m, nil
	case "t":
		if m.showToolPicker {
			m.closeMenus()
		} else {
			m.openMenu(2)
		}
		return m, nil
	case "[":
		active := m.activeMenu()
		if active < 0 {
			m.openMenu(m.lastMenu)
		} else {
			m.openMenu((active - 1 + menuCount) % menuCount)
		}
		return m, nil
	case "]":
		active := m.activeMenu()
		if active < 0 {
			m.openMenu(m.lastMenu)
		} else {
			m.openMenu((active + 1) % menuCount)
		}
		return m, nil
	case "esc":
		if m.showToolPicker && m.toolPickerFocusLevel > 0 {
			m.toolPickerFocusLevel--
			if m.toolPickerFocusLevel == 1 && isDrawingTool(m.selectedTool) {
				m.onGlyphSelector = true
			} else if m.toolPickerFocusLevel == 0 {
				m.onGlyphSelector = false
			}
			return m, nil
		}
		m.showFgPicker = false
		m.showBgPicker = false
		m.showToolPicker = false
		m.selection.active = false
		m.toolPickerFocusLevel = 0
		return m, nil
	case "left":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusLevel > 0 {
			m.toolPickerFocusLevel--
			if m.toolPickerFocusLevel == 1 && isDrawingTool(m.selectedTool) {
				m.onGlyphSelector = true
			} else if m.toolPickerFocusLevel == 0 {
				m.onGlyphSelector = false
			}
			return m, nil
		}
		m.openMenu((m.activeMenu() - 1 + menuCount) % menuCount)
		return m, nil
	case "right":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusLevel == 0 && m.toolHasSubmenu() {
			m.toolPickerFocusLevel = 1
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusLevel == 1 && m.toolHasGlyphPicker() {
			m.toolPickerFocusLevel = 2
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusLevel == 2 {
			m.toolPickerFocusLevel = 3
			if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
				currentIdx := m.findSelectedCharIndexInCategory(m.selectedCategory)
				if currentIdx == 0 && m.selectedChar != characterGroups[m.selectedCategory].chars[0] {
					m.selectedChar = characterGroups[m.selectedCategory].chars[0]
				}
			}
			return m, nil
		}
		m.openMenu((m.activeMenu() + 1) % menuCount)
		return m, nil
	case "enter":
		if m.showToolPicker && m.toolPickerFocusLevel >= 1 {
			m.closeMenus()
			return m, nil
		}
		if m.tool().OnKeyPress(m, "enter") {
			return m, nil
		} else if m.showFgPicker {
			m.showFgPicker = false
			return m, nil
		} else if m.showBgPicker {
			m.showBgPicker = false
			return m, nil
		} else if m.showToolPicker {
			m.showToolPicker = false
			return m, nil
		}
	case "up":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showFgPicker {
			idx := m.findSelectedColorIndex(m.foregroundColor)
			if idx > 0 {
				m.foregroundColor = colors[idx-1].name
			}
			return m, nil
		} else if m.showBgPicker {
			idx := m.findSelectedColorIndex(m.backgroundColor)
			if idx > 0 {
				m.backgroundColor = colors[idx-1].name
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 3 {
			idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
			if idx > 0 {
				m.selectedChar = characterGroups[m.selectedCategory].chars[idx-1]
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 2 {
			if m.selectedCategory > 0 {
				m.selectedCategory--
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 1 {
			idx := m.toolSubmenuIndex()
			if idx > 0 {
				m.setToolSubmenuIndex(idx - 1)
			}
			return m, nil
		} else if m.showToolPicker {
			idx := m.toolPickerIndex()
			if idx > 0 {
				m.setToolPickerIndex(idx - 1)
			}
			return m, nil
		}
	case "down":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showFgPicker {
			idx := m.findSelectedColorIndex(m.foregroundColor)
			if idx < len(colors)-1 {
				m.foregroundColor = colors[idx+1].name
			}
			return m, nil
		} else if m.showBgPicker {
			idx := m.findSelectedColorIndex(m.backgroundColor)
			if idx < len(colors)-1 {
				m.backgroundColor = colors[idx+1].name
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 3 {
			idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
			if idx < len(characterGroups[m.selectedCategory].chars)-1 {
				m.selectedChar = characterGroups[m.selectedCategory].chars[idx+1]
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 2 {
			if m.selectedCategory < len(characterGroups)-1 {
				m.selectedCategory++
			}
			return m, nil
		} else if m.showToolPicker && m.toolPickerFocusLevel == 1 {
			idx := m.toolSubmenuIndex()
			if idx < m.toolSubmenuCount()-1 {
				m.setToolSubmenuIndex(idx + 1)
			}
			return m, nil
		} else if m.showToolPicker {
			items := m.toolPickerItems()
			idx := m.toolPickerIndex()
			if idx < len(items)-1 {
				m.setToolPickerIndex(idx + 1)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m *model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.mouseX = msg.X
	m.mouseY = msg.Y

	hoverX, hoverY := m.screenToCanvas(m.mouseX, m.mouseY)
	m.hoverRow = hoverY
	m.hoverCol = hoverX

	// Handle popup and menu clicks (only on initial click, not during drag)
	if msg.Type == tea.MouseLeft && !m.mouseDown {
		if m.showFgPicker {
			if idx := m.colorPickerClickIndex(msg, m.toolbar.foregroundItemX); idx >= 0 {
				m.foregroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showBgPicker {
			if idx := m.colorPickerClickIndex(msg, m.toolbar.backgroundItemX); idx >= 0 {
				m.backgroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showToolPicker {
			items := m.toolPickerItems()
			pickerHeight := len(items) + pickerBorderWidth
			pickerTop := controlBarHeight
			pickerLeft := m.toolbar.toolItemX - pickerContentOffset

			iconCol := 0
			for _, item := range items {
				if w := lipgloss.Width(item.icon); w > iconCol {
					iconCol = w
				}
			}
			maxToolLen := 0
			for _, item := range items {
				if w := lipgloss.Width(item.name); w > maxToolLen {
					maxToolLen = w
				}
			}
			lineWidth := 1 + maxToolLen + 1
			if iconCol > 0 {
				lineWidth = 1 + iconCol + 1 + maxToolLen + 1
			}
			pickerWidth := pickerBorderWidth + lineWidth

			if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
				msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
				itemIdx := msg.Y - pickerTop - 1
				if itemIdx >= 0 && itemIdx < len(items) {
					m.setToolPickerIndex(itemIdx)
					m.toolPickerFocusLevel = 0
					return m, nil
				}
			}

			if m.toolHasSubmenu() {
				submenuLeft := pickerLeft + pickerWidth - 1
				popup2 := m.renderToolSubmenuPicker()
				popup2Lines := strings.Split(popup2, "\n")
				submenuWidth := 0
				if len(popup2Lines) > 0 {
					submenuWidth = lipgloss.Width(popup2Lines[0])
				}

				screenRows := m.height - controlBarHeight
				if !m.hasFixedSize() {
					screenRows = m.canvas.height
				}
				pickerIdx := m.toolPickerIndex()
				submenuCanvasY := pickerIdx
				if submenuCanvasY+len(popup2Lines) > screenRows {
					submenuCanvasY = screenRows - len(popup2Lines)
				}
				if submenuCanvasY < 0 {
					submenuCanvasY = 0
				}
				submenuTop := controlBarHeight + submenuCanvasY

				if msg.Y >= submenuTop && msg.Y < submenuTop+len(popup2Lines) &&
					msg.X >= submenuLeft && msg.X < submenuLeft+submenuWidth {
					itemIdx := msg.Y - submenuTop - 1
					if itemIdx >= 0 && itemIdx < m.toolSubmenuCount() {
						m.setToolSubmenuIndex(itemIdx)
						m.toolPickerFocusLevel = 1
						return m, nil
					}
				}

				// popup3 and popup4: glyph category and glyph pickers
				if m.toolHasGlyphPicker() {
					popup3 := m.renderCategoryPicker()
					popup3Lines := strings.Split(popup3, "\n")
					popup3Width := 0
					if len(popup3Lines) > 0 {
						popup3Width = lipgloss.Width(popup3Lines[0])
					}

					// popup3 aligns with the Glyphs row (index 0) in the submenu
					popup3CanvasY := submenuCanvasY
					if popup3CanvasY+len(popup3Lines) > screenRows {
						popup3CanvasY = screenRows - len(popup3Lines)
					}
					if popup3CanvasY < 0 {
						popup3CanvasY = 0
					}
					popup3Top := controlBarHeight + popup3CanvasY
					popup3Left := submenuLeft + submenuWidth - 1

					if msg.Y >= popup3Top && msg.Y < popup3Top+len(popup3Lines) &&
						msg.X >= popup3Left && msg.X < popup3Left+popup3Width {
						row := msg.Y - popup3Top - 1
						if row >= 0 && row < len(characterGroups) {
							m.selectedCategory = row
							m.toolPickerFocusLevel = 2
							return m, nil
						}
					}

					// popup4: glyphs picker
					popup4 := m.renderGlyphsPicker()
					popup4Lines := strings.Split(popup4, "\n")
					popup4Width := 0
					if len(popup4Lines) > 0 {
						popup4Width = lipgloss.Width(popup4Lines[0])
					}

					popup4CanvasY := popup3CanvasY + m.selectedCategory
					if popup4CanvasY+len(popup4Lines) > screenRows {
						popup4CanvasY = screenRows - len(popup4Lines)
					}
					if popup4CanvasY < 0 {
						popup4CanvasY = 0
					}
					popup4Top := controlBarHeight + popup4CanvasY
					popup4Left := popup3Left + popup3Width - 1

					if msg.Y >= popup4Top && msg.Y < popup4Top+len(popup4Lines) &&
						msg.X >= popup4Left && msg.X < popup4Left+popup4Width {
						glyphRow := msg.Y - popup4Top - 1
						if glyphRow >= 0 && glyphRow < len(characterGroups[m.selectedCategory].chars) {
							m.selectedChar = characterGroups[m.selectedCategory].chars[glyphRow]
							m.toolPickerFocusLevel = 3
							return m, nil
						}
					}
				}
			}
		}

		// Check if clicking on control bar buttons
		if msg.Y < controlBarHeight {
			if m.toolbar.toolX > 0 && msg.X >= m.toolbar.toolX {
				m.showToolPicker = !m.showToolPicker
				m.showFgPicker = false
				m.showBgPicker = false
				return m, nil
			} else if m.toolbar.backgroundX > 0 && msg.X >= m.toolbar.backgroundX && msg.X < m.toolbar.toolX {
				m.showBgPicker = !m.showBgPicker
				m.showFgPicker = false
				m.showToolPicker = false
				return m, nil
			} else if msg.X >= m.toolbar.foregroundX && msg.X < m.toolbar.backgroundX {
				m.showFgPicker = !m.showFgPicker
				m.showBgPicker = false
				m.showToolPicker = false
				return m, nil
			}
		}
	}

	// Handle mouse press (start of stroke)
	if msg.Type == tea.MouseLeft && !m.mouseDown && msg.Y >= controlBarHeight {
		cx, cy := m.screenToCanvas(msg.X, msg.Y)
		if m.hasFixedSize() && (cy < 0 || cy >= m.canvas.height || cx < 0 || cx >= m.canvas.width) {
			return m, nil
		}
		m.mouseDown = true
		m.canvasBeforeStroke = m.canvas.Copy()
		m.startX = cx
		m.startY = cy
		m.tool().OnPress(m, cy, cx)
	}

	// Handle drag events
	if (msg.Type == tea.MouseLeft || msg.Type == tea.MouseMotion) && m.mouseDown {
		canvasX, canvasY := m.screenToCanvas(msg.X, msg.Y)

		m.tool().OnDrag(m, canvasY, canvasX)
	}

	// Handle mouse release (end of stroke)
	if msg.Type == tea.MouseRelease && m.mouseDown {
		m.mouseDown = false
		m.showPreview = false
		m.previewPoints = nil

		canvasX, canvasY := m.screenToCanvas(msg.X, msg.Y)
		clampedY, clampedX := m.clampToCanvas(canvasY, canvasX)

		tool := m.tool()
		tool.OnRelease(m, clampedY, clampedX)

		m.optionKeyHeld = false

		if tool.ModifiesCanvas() && !m.canvas.Equals(m.canvasBeforeStroke) {
			m.saveToHistory()
		}
	}

	return m, nil
}

func (m *model) screenToCanvas(screenX, screenY int) (canvasX, canvasY int) {
	if m.hasFixedSize() {
		offY, offX := m.canvasOffset()
		canvasY = screenY - controlBarHeight - offY - 2
		canvasX = screenX - offX - 1
	} else {
		canvasY = screenY - controlBarHeight
		canvasX = screenX
	}
	return
}

func (m *model) colorPickerClickIndex(msg tea.MouseMsg, itemX int) int {
	pickerHeight := len(colors) + pickerBorderWidth
	pickerTop := controlBarHeight
	pickerLeft := itemX - pickerContentOffset

	maxNameLen := 0
	for _, c := range colors {
		displayName := strings.ReplaceAll(c.name, "_", " ")
		if len(displayName) > 0 {
			displayName = strings.ToUpper(displayName[:1]) + displayName[1:]
		}
		if len(displayName) > maxNameLen {
			maxNameLen = len(displayName)
		}
	}
	pickerWidth := pickerBorderWidth + pickerItemPadding + pickerSwatchWidth + pickerItemSeparator + maxNameLen + pickerBorderWidth

	if msg.Y < pickerTop || msg.Y >= pickerTop+pickerHeight || msg.X < pickerLeft || msg.X >= pickerLeft+pickerWidth {
		return -1
	}
	colorIdx := msg.Y - pickerTop - 1
	if colorIdx < 0 || colorIdx >= len(colors) {
		return -1
	}
	return colorIdx
}

func (m *model) clampToCanvas(y, x int) (int, int) {
	if y < 0 {
		y = 0
	} else if y >= m.canvas.height {
		y = m.canvas.height - 1
	}
	if x < 0 {
		x = 0
	} else if x >= m.canvas.width {
		x = m.canvas.width - 1
	}
	return y, x
}
