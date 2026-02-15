package main

const menuCount = 4

func (m *model) activeMenu() int {
	if m.showCharPicker {
		return 0
	} else if m.showFgPicker {
		return 1
	} else if m.showBgPicker {
		return 2
	} else if m.showToolPicker {
		return 3
	}
	return -1
}

func (m *model) openMenu(idx int) {
	m.showCharPicker = idx == 0
	m.showFgPicker = idx == 1
	m.showBgPicker = idx == 2
	m.showToolPicker = idx == 3
	m.showingShapes = idx == 0
	m.shapesFocusOnPanel = false
	if idx == 0 {
		m.selectedCategory = m.findSelectedCharCategory()
	}
	if idx >= 0 {
		m.lastMenu = idx
	}
}

func (m *model) setTool(tool string) {
	m.selectedTool = tool
	m.hasSelection = false
}

func (m *model) closeMenus() {
	m.showCharPicker = false
	m.showFgPicker = false
	m.showBgPicker = false
	m.showToolPicker = false
	m.showingShapes = false
	m.shapesFocusOnPanel = false
}
