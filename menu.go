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
	m.toolPickerFocusOnStyle = false
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
	if isDrawingTool(tool) {
		m.drawingTool = tool
	}
}

func (m *model) closeMenus() {
	m.showCharPicker = false
	m.showFgPicker = false
	m.showBgPicker = false
	m.showToolPicker = false
	m.showingShapes = false
	m.shapesFocusOnPanel = false
	m.toolPickerFocusOnStyle = false
}

// Top-level tool picker has 4 items: drawing group, Box, Fill, Select
type toolPickerItem struct {
	icon     string
	name     string
	selected bool
}

var topLevelTools = []string{"Box", "Fill", "Select"}

func (m *model) toolPickerItems() []toolPickerItem {
	items := make([]toolPickerItem, 0, 4)

	// Drawing tools group
	drawingName := m.tool().DisplayName(m)
	if !isDrawingTool(m.selectedTool) {
		for _, opt := range drawingToolOptions {
			if opt.toolName == m.drawingTool && opt.circleMode == m.circleMode {
				drawingName = opt.name
				break
			}
		}
		if drawingName == "" {
			drawingName = "Points"
		}
	}
	items = append(items, toolPickerItem{
		name:     drawingName,
		selected: isDrawingTool(m.selectedTool),
	})

	// Box — icon is the depiction, name is "Style Box"
	s := boxStyles[m.boxStyle]
	items = append(items, toolPickerItem{
		icon:     s.tl + s.h + s.tr,
		name:     s.name + " Box",
		selected: m.selectedTool == "Box",
	})

	// Fill, Select
	for _, name := range []string{"Fill", "Select"} {
		for _, t := range toolRegistry {
			if t.Name() == name {
				items = append(items, toolPickerItem{
					name:     t.DisplayName(m),
					selected: m.selectedTool == name,
				})
				break
			}
		}
	}

	return items
}

func (m *model) toolPickerIndex() int {
	if isDrawingTool(m.selectedTool) {
		return 0
	}
	for i, name := range topLevelTools {
		if m.selectedTool == name {
			return i + 1
		}
	}
	return 0
}

func (m *model) setToolPickerIndex(idx int) {
	if idx == 0 {
		m.setTool(m.drawingTool)
		return
	}
	if idx-1 < len(topLevelTools) {
		m.setTool(topLevelTools[idx-1])
	}
}

func (m *model) toolHasSubmenu() bool {
	return isDrawingTool(m.selectedTool) || m.selectedTool == "Box"
}

func (m *model) toolSubmenuCount() int {
	if isDrawingTool(m.selectedTool) {
		return len(drawingToolOptions)
	}
	if m.selectedTool == "Box" {
		return len(boxStyles)
	}
	return 0
}

func (m *model) toolSubmenuIndex() int {
	if isDrawingTool(m.selectedTool) {
		return m.drawingToolOptionIndex()
	}
	if m.selectedTool == "Box" {
		return m.boxStyle
	}
	return 0
}

func (m *model) setToolSubmenuIndex(idx int) {
	if isDrawingTool(m.selectedTool) {
		if idx >= 0 && idx < len(drawingToolOptions) {
			opt := drawingToolOptions[idx]
			m.setTool(opt.toolName)
			m.circleMode = opt.circleMode
		}
		return
	}
	if m.selectedTool == "Box" {
		m.boxStyle = idx
	}
}

func (m *model) drawingToolOptionIndex() int {
	for i, opt := range drawingToolOptions {
		if opt.toolName == m.selectedTool {
			if opt.toolName == "Ellipse" && opt.circleMode != m.circleMode {
				continue
			}
			return i
		}
	}
	return 0
}
