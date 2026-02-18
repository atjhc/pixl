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
	if msg.Alt {
		m.optionKeyHeld = true
	}

	switch msg.String() {
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		idx := int(msg.String()[0] - '1')

		if m.showCharPicker && !m.glyphsFocusOnPanel {
			if idx < len(characterGroups) {
				m.selectedCategory = idx
				return m, nil
			}
		} else if m.showCharPicker && m.glyphsFocusOnPanel {
			if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
				chars := characterGroups[m.selectedCategory].chars
				if idx < len(chars) {
					m.selectedChar = chars[idx]
					return m, nil
				}
			}
		} else if m.showFgPicker {
			if idx < len(colors) {
				m.foregroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showBgPicker {
			if idx < len(colors) {
				m.backgroundColor = colors[idx].name
				return m, nil
			}
		} else if m.showToolPicker && m.toolPickerFocusOnStyle {
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
	case "g":
		if m.showCharPicker {
			m.closeMenus()
		} else {
			m.openMenu(0)
		}
		return m, nil
	case "f":
		if m.showFgPicker {
			m.closeMenus()
		} else {
			m.openMenu(1)
		}
		return m, nil
	case "b":
		if m.showBgPicker {
			m.closeMenus()
		} else {
			m.openMenu(2)
		}
		return m, nil
	case "t":
		if m.showToolPicker {
			m.closeMenus()
		} else {
			m.openMenu(3)
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
		if m.showCharPicker && m.glyphsFocusOnPanel {
			m.glyphsFocusOnPanel = false
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusOnStyle {
			m.toolPickerFocusOnStyle = false
			return m, nil
		}
		m.showCharPicker = false
		m.showFgPicker = false
		m.showBgPicker = false
		m.showToolPicker = false
		m.hasSelection = false
		m.glyphsFocusOnPanel = false
		m.toolPickerFocusOnStyle = false
		return m, nil
	case "left":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showCharPicker && m.glyphsFocusOnPanel {
			m.glyphsFocusOnPanel = false
			return m, nil
		}
		if m.showToolPicker && m.toolPickerFocusOnStyle {
			m.toolPickerFocusOnStyle = false
			return m, nil
		}
		m.openMenu((m.activeMenu() - 1 + menuCount) % menuCount)
		return m, nil
	case "right":
		if m.activeMenu() < 0 {
			m.openMenu(m.lastMenu)
			return m, nil
		}
		if m.showCharPicker && !m.glyphsFocusOnPanel {
			m.glyphsFocusOnPanel = true
			if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
				currentIdx := m.findSelectedCharIndexInCategory(m.selectedCategory)
				if currentIdx == 0 && m.selectedChar != characterGroups[m.selectedCategory].chars[0] {
					m.selectedChar = characterGroups[m.selectedCategory].chars[0]
				}
			}
			return m, nil
		}
		if m.showToolPicker && !m.toolPickerFocusOnStyle && m.toolHasSubmenu() {
			m.toolPickerFocusOnStyle = true
			return m, nil
		}
		m.openMenu((m.activeMenu() + 1) % menuCount)
		return m, nil
	case "enter":
		if m.showToolPicker && m.toolPickerFocusOnStyle {
			m.closeMenus()
			return m, nil
		}
		if m.tool().OnKeyPress(m, "enter") {
			return m, nil
		} else if m.showCharPicker {
			m.showCharPicker = false
			m.showingGlyphs = false
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
		if m.showCharPicker {
			if m.glyphsFocusOnPanel {
				idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
				if idx > 0 {
					m.selectedChar = characterGroups[m.selectedCategory].chars[idx-1]
				}
			} else {
				if m.selectedCategory > 0 {
					m.selectedCategory--
				}
			}
			return m, nil
		} else if m.showFgPicker {
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
		} else if m.showToolPicker && m.toolPickerFocusOnStyle {
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
		if m.showCharPicker {
			if m.glyphsFocusOnPanel {
				idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
				if idx < len(characterGroups[m.selectedCategory].chars)-1 {
					m.selectedChar = characterGroups[m.selectedCategory].chars[idx+1]
				}
			} else {
				if m.selectedCategory < len(characterGroups)-1 {
					m.selectedCategory++
				}
			}
			return m, nil
		} else if m.showFgPicker {
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
		} else if m.showToolPicker && m.toolPickerFocusOnStyle {
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
		if m.showCharPicker {
			categoryPickerLeft := m.toolbarGlyphItemX - pickerContentOffset
			maxCategoryWidth := 0
			for _, group := range characterGroups {
				nameWidth := len(group.name) + 2
				if nameWidth > maxCategoryWidth {
					maxCategoryWidth = nameWidth
				}
			}
			categoryPickerWidth := maxCategoryWidth + pickerBorderWidth

			pickerHeight := len(characterGroups) + pickerBorderWidth
			pickerTop := controlBarHeight

			if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
				msg.X >= categoryPickerLeft && msg.X < categoryPickerLeft+categoryPickerWidth {
				row := msg.Y - pickerTop - 1
				if row >= 0 && row < len(characterGroups) {
					if msg.X >= categoryPickerLeft+1 && msg.X < categoryPickerLeft+categoryPickerWidth-1 {
						m.selectedCategory = row
						m.glyphsFocusOnPanel = false
						return m, nil
					}
				}
			}

			glyphsPickerLeft := categoryPickerLeft + categoryPickerWidth - 1
			glyphsPickerWidth := 3 + pickerBorderWidth

			glyphsPickerHeight := len(characterGroups[m.selectedCategory].chars) + pickerBorderWidth
			screenRows := m.height - controlBarHeight
			if !m.hasFixedSize() {
				screenRows = m.canvas.height
			}
			glyphsCanvasY := m.selectedCategory
			if glyphsCanvasY+glyphsPickerHeight > screenRows {
				glyphsCanvasY = screenRows - glyphsPickerHeight
			}
			if glyphsCanvasY < 0 {
				glyphsCanvasY = 0
			}
			glyphsPickerTop := controlBarHeight + glyphsCanvasY

			if msg.Y >= glyphsPickerTop && msg.Y < glyphsPickerTop+glyphsPickerHeight &&
				msg.X >= glyphsPickerLeft && msg.X < glyphsPickerLeft+glyphsPickerWidth {
				glyphRow := msg.Y - glyphsPickerTop - 1
				if glyphRow >= 0 && glyphRow < len(characterGroups[m.selectedCategory].chars) {
					m.selectedChar = characterGroups[m.selectedCategory].chars[glyphRow]
					m.glyphsFocusOnPanel = true
					return m, nil
				}
			}
		} else if m.showFgPicker {
			pickerHeight := len(colors) + pickerBorderWidth
			pickerTop := controlBarHeight
			pickerLeft := m.toolbarForegroundItemX - pickerContentOffset

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

			if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
				msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
				colorIdx := msg.Y - pickerTop - 1
				if colorIdx >= 0 && colorIdx < len(colors) {
					m.foregroundColor = colors[colorIdx].name
					return m, nil
				}
			}
		} else if m.showBgPicker {
			pickerHeight := len(colors) + pickerBorderWidth
			pickerTop := controlBarHeight
			pickerLeft := m.toolbarBackgroundItemX - pickerContentOffset

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

			if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
				msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
				colorIdx := msg.Y - pickerTop - 1
				if colorIdx >= 0 && colorIdx < len(colors) {
					m.backgroundColor = colors[colorIdx].name
					return m, nil
				}
			}
		} else if m.showToolPicker {
			items := m.toolPickerItems()
			pickerHeight := len(items) + pickerBorderWidth
			pickerTop := controlBarHeight
			pickerLeft := m.toolbarToolItemX - pickerContentOffset

			maxToolLen := 0
			for _, item := range items {
				if len(item.name) > maxToolLen {
					maxToolLen = len(item.name)
				}
			}
			pickerWidth := pickerBorderWidth + pickerItemPadding + maxToolLen + pickerItemPadding

			if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
				msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
				itemIdx := msg.Y - pickerTop - 1
				if itemIdx >= 0 && itemIdx < len(items) {
					m.setToolPickerIndex(itemIdx)
					m.toolPickerFocusOnStyle = false
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
						m.toolPickerFocusOnStyle = true
						return m, nil
					}
				}
			}
		}

		// Check if clicking on control bar buttons
		if msg.Y < controlBarHeight {
			if m.toolbarToolX > 0 && msg.X >= m.toolbarToolX {
				m.showToolPicker = !m.showToolPicker
				m.showCharPicker = false
				m.showFgPicker = false
				m.showBgPicker = false
				return m, nil
			} else if m.toolbarBackgroundX > 0 && msg.X >= m.toolbarBackgroundX && msg.X < m.toolbarToolX {
				m.showBgPicker = !m.showBgPicker
				m.showCharPicker = false
				m.showFgPicker = false
				m.showToolPicker = false
				return m, nil
			} else if m.toolbarForegroundX > 0 && msg.X >= m.toolbarForegroundX && msg.X < m.toolbarBackgroundX {
				m.showFgPicker = !m.showFgPicker
				m.showCharPicker = false
				m.showBgPicker = false
				m.showToolPicker = false
				return m, nil
			} else if m.toolbarGlyphX > 0 && msg.X >= m.toolbarGlyphX && msg.X < m.toolbarForegroundX {
				m.showCharPicker = !m.showCharPicker
				m.showFgPicker = false
				m.showBgPicker = false
				m.showToolPicker = false
				if m.showCharPicker {
					m.glyphsFocusOnPanel = false
				}
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
