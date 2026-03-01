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
	if !m.selection.active {
		return
	}

	minY, minX, maxY, maxX := normalizeRect(m.selection.startY, m.selection.startX, m.selection.endY, m.selection.endX)

	// Internal region excludes the visual border
	internalMinY := minY + 1
	internalMaxY := maxY - 1
	internalMinX := minX + 1
	internalMaxX := maxX - 1

	if internalMaxY < internalMinY || internalMaxX < internalMinX {
		m.clipboard.cells = nil
		m.clipboard.width = 0
		m.clipboard.height = 0
		return
	}

	m.clipboard.height = internalMaxY - internalMinY + 1
	m.clipboard.width = internalMaxX - internalMinX + 1
	m.clipboard.cells = make([][]Cell, m.clipboard.height)

	for y := 0; y < m.clipboard.height; y++ {
		m.clipboard.cells[y] = make([]Cell, m.clipboard.width)
		for x := 0; x < m.clipboard.width; x++ {
			cell := m.canvas.Get(internalMinY+y, internalMinX+x)
			if cell != nil {
				m.clipboard.cells[y][x] = *cell
			} else {
				m.clipboard.cells[y][x] = Cell{char: " ", foregroundColor: "transparent", backgroundColor: "transparent"}
			}
		}
	}
}

func (m *model) cutSelection() {
	if !m.selection.active {
		return
	}

	m.copySelection()

	minY, minX, maxY, maxX := normalizeRect(m.selection.startY, m.selection.startX, m.selection.endY, m.selection.endX)

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
	if m.clipboard.cells == nil || m.clipboard.height == 0 || m.clipboard.width == 0 {
		return
	}

	var originY, originX int
	if m.selection.active {
		originY, originX, _, _ = normalizeRect(m.selection.startY, m.selection.startX, m.selection.endY, m.selection.endX)
		// Selection border is visual-only; content starts 1 cell inside
		originY++
		originX++
	} else {
		if m.mouseY < controlBarHeight {
			return
		}
		originX, originY = m.screenToCanvas(m.mouseX, m.mouseY)
	}

	for y := 0; y < m.clipboard.height; y++ {
		for x := 0; x < m.clipboard.width; x++ {
			targetY := originY + y
			targetX := originX + x
			if targetY < 0 || targetY >= m.canvas.height || targetX < 0 || targetX >= m.canvas.width {
				continue
			}

			cell := m.clipboard.cells[y][x]
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
	m.selection.active = false
}
