package main

func (m *model) saveToHistory() {
	if len(m.history) == 0 {
		m.history = []Canvas{m.canvas.Copy()}
		m.historyIndex = 0
		return
	}

	// Remove any redo history when making a new change
	if m.historyIndex < len(m.history)-1 {
		m.history = m.history[:m.historyIndex+1]
	}

	m.history = append(m.history, m.canvas.Copy())
	m.historyIndex++

	// Limit history size to 50 states
	if len(m.history) > 50 {
		m.history = m.history[1:]
		m.historyIndex--
	}
}

func (m *model) undo() {
	if m.historyIndex > 0 {
		m.historyIndex--
		m.canvas = m.history[m.historyIndex].Copy()
	}
}

func (m *model) redo() {
	if m.historyIndex < len(m.history)-1 {
		m.historyIndex++
		m.canvas = m.history[m.historyIndex].Copy()
	}
}

func (m *model) copySelection() {
	if !m.hasSelection {
		return
	}

	minY, minX, maxY, maxX := normalizeRect(m.selectionStartY, m.selectionStartX, m.selectionEndY, m.selectionEndX)

	// Internal region excludes the visual border
	internalMinY := minY + 1
	internalMaxY := maxY - 1
	internalMinX := minX + 1
	internalMaxX := maxX - 1

	if internalMaxY < internalMinY || internalMaxX < internalMinX {
		m.clipboard = nil
		m.clipboardWidth = 0
		m.clipboardHeight = 0
		return
	}

	m.clipboardHeight = internalMaxY - internalMinY + 1
	m.clipboardWidth = internalMaxX - internalMinX + 1
	m.clipboard = make([][]Cell, m.clipboardHeight)

	for y := 0; y < m.clipboardHeight; y++ {
		m.clipboard[y] = make([]Cell, m.clipboardWidth)
		for x := 0; x < m.clipboardWidth; x++ {
			cell := m.canvas.Get(internalMinY+y, internalMinX+x)
			if cell != nil {
				m.clipboard[y][x] = *cell
			} else {
				m.clipboard[y][x] = Cell{char: " ", foregroundColor: "transparent", backgroundColor: "transparent"}
			}
		}
	}
}

func (m *model) cutSelection() {
	if !m.hasSelection {
		return
	}

	m.copySelection()

	minY, minX, maxY, maxX := normalizeRect(m.selectionStartY, m.selectionStartX, m.selectionEndY, m.selectionEndX)

	internalMinY := minY + 1
	internalMaxY := maxY - 1
	internalMinX := minX + 1
	internalMaxX := maxX - 1

	if internalMaxY >= internalMinY && internalMaxX >= internalMinX {
		for y := internalMinY; y <= internalMaxY; y++ {
			for x := internalMinX; x <= internalMaxX; x++ {
				m.canvas.Set(y, x, " ", "transparent", "transparent")
			}
		}
		m.saveToHistory()
	}
}

func (m *model) paste() {
	if m.clipboard == nil || m.clipboardHeight == 0 || m.clipboardWidth == 0 {
		return
	}

	var originY, originX int
	if m.hasSelection {
		originY, originX, _, _ = normalizeRect(m.selectionStartY, m.selectionStartX, m.selectionEndY, m.selectionEndX)
		// Selection border is visual-only; content starts 1 cell inside
		originY++
		originX++
	} else {
		if m.mouseY < controlBarHeight {
			return
		}
		originX, originY = m.screenToCanvas(m.mouseX, m.mouseY)
	}

	for y := 0; y < m.clipboardHeight; y++ {
		for x := 0; x < m.clipboardWidth; x++ {
			targetY := originY + y
			targetX := originX + x
			if targetY < 0 || targetY >= m.canvas.height || targetX < 0 || targetX >= m.canvas.width {
				continue
			}

			cell := m.clipboard[y][x]
			existingCell := m.canvas.Get(targetY, targetX)

			// Skip fully transparent cells
			if cell.foregroundColor == "transparent" && cell.backgroundColor == "transparent" {
				continue
			}

			newChar := cell.char
			newFg := cell.foregroundColor
			newBg := cell.backgroundColor

			if cell.foregroundColor == "transparent" && existingCell != nil {
				newChar = existingCell.char
				newFg = existingCell.foregroundColor
			}

			if cell.backgroundColor == "transparent" && existingCell != nil {
				newBg = existingCell.backgroundColor
			}

			m.canvas.Set(targetY, targetX, newChar, newFg, newBg)
		}
	}

	m.saveToHistory()
	m.hasSelection = false
}
