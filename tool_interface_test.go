package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestToolRegistryOrderAndNames(t *testing.T) {
	expected := []string{"Point", "Rectangle", "Box", "Ellipse", "Line", "Fill", "Select"}

	if len(toolRegistry) != len(expected) {
		t.Fatalf("toolRegistry has %d tools, want %d", len(toolRegistry), len(expected))
	}

	for i, name := range expected {
		if toolRegistry[i].Name() != name {
			t.Errorf("toolRegistry[%d].Name() = %q, want %q", i, toolRegistry[i].Name(), name)
		}
	}
}

func TestToolDisplayName(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Point"
	if got := m.tool().DisplayName(m); got != "Points" {
		t.Errorf("Point DisplayName = %q, want Points", got)
	}

	m.selectedTool = "Ellipse"
	m.circleMode = false
	if got := m.tool().DisplayName(m); got != "Ellipse" {
		t.Errorf("DisplayName with circleMode=false = %q, want Ellipse", got)
	}

	m.circleMode = true
	if got := m.tool().DisplayName(m); got != "Oval" {
		t.Errorf("DisplayName with circleMode=true = %q, want Oval", got)
	}
}

func TestToolCursorChar(t *testing.T) {
	m := newTestModel(10, 10)

	m.selectedTool = "Point"
	if got := m.tool().CursorChar(m); got != "" {
		t.Errorf("Point CursorChar = %q, want empty", got)
	}

	for _, name := range []string{"Rectangle", "Ellipse", "Line", "Fill"} {
		m.selectedTool = name
		if got := m.tool().CursorChar(m); got != "" {
			t.Errorf("%s CursorChar = %q, want empty (use selected char)", name, got)
		}
	}

	m.selectedTool = "Select"
	if got := m.tool().CursorChar(m); got != "┼" {
		t.Errorf("Select CursorChar = %q, want ┼", got)
	}
}

func TestBoxToolCursorChar(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Box"

	if got := m.tool().CursorChar(m); got != "┌" {
		t.Errorf("Box CursorChar with style 0 = %q, want ┌", got)
	}

	m.boxStyle = 1
	if got := m.tool().CursorChar(m); got != "╔" {
		t.Errorf("Box CursorChar with style 1 = %q, want ╔", got)
	}
}

func TestBoxToolCyclesStyles(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Box"

	if m.boxStyle != 0 {
		t.Fatalf("initial boxStyle = %d, want 0", m.boxStyle)
	}

	m.tool().OnKeyPress(m, "enter")
	if m.boxStyle != 1 {
		t.Errorf("after 1 enter, boxStyle = %d, want 1", m.boxStyle)
	}

	m.tool().OnKeyPress(m, "enter")
	if m.boxStyle != 2 {
		t.Errorf("after 2 enters, boxStyle = %d, want 2", m.boxStyle)
	}

	// Cycle through all 5 styles back to 0
	m.tool().OnKeyPress(m, "enter") // 3
	m.tool().OnKeyPress(m, "enter") // 4
	m.tool().OnKeyPress(m, "enter") // 0
	if m.boxStyle != 0 {
		t.Errorf("after wrapping, boxStyle = %d, want 0", m.boxStyle)
	}
}

func TestBoxToolDisplayName(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Box"

	m.boxStyle = 0
	if got := m.tool().DisplayName(m); got != "┌─┐ Single Box" {
		t.Errorf("DisplayName style 0 = %q, want %q", got, "┌─┐ Single Box")
	}

	m.boxStyle = 1
	if got := m.tool().DisplayName(m); got != "╔═╗ Double Box" {
		t.Errorf("DisplayName style 1 = %q, want %q", got, "╔═╗ Double Box")
	}
}

func TestToolModifiesCanvas(t *testing.T) {
	m := newTestModel(10, 10)

	for _, name := range []string{"Point", "Rectangle", "Box", "Ellipse", "Line", "Fill"} {
		m.selectedTool = name
		if !m.tool().ModifiesCanvas() {
			t.Errorf("%s.ModifiesCanvas() = false, want true", name)
		}
	}

	m.selectedTool = "Select"
	if m.tool().ModifiesCanvas() {
		t.Error("Select.ModifiesCanvas() = true, want false")
	}
}

func TestEllipseToolOnKeyPress(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Ellipse"
	m.circleMode = false

	if !m.tool().OnKeyPress(m, "enter") {
		t.Error("Ellipse OnKeyPress(enter) should return true")
	}
	if !m.circleMode {
		t.Error("Ellipse OnKeyPress(enter) should toggle circleMode")
	}

	m.selectedTool = "Point"
	if m.tool().OnKeyPress(m, "enter") {
		t.Error("Point OnKeyPress(enter) should return false")
	}
}

func TestToolLookupFallback(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "NonExistent"

	tool := m.tool()
	if tool.Name() != "Point" {
		t.Errorf("fallback tool = %q, want Point", tool.Name())
	}
}

// --- Tool picker grouping tests ---

func TestToolPickerItems(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Point"

	items := m.toolPickerItems()
	if len(items) != 4 {
		t.Fatalf("toolPickerItems count = %d, want 4", len(items))
	}

	// First item should show current drawing tool name
	if items[0].name != "Points" {
		t.Errorf("item 0 name = %q, want Points", items[0].name)
	}
	if !items[0].selected {
		t.Error("item 0 should be selected when tool is Point")
	}

	if items[1].icon != "┌─┐" {
		t.Errorf("item 1 icon = %q, want ┌─┐", items[1].icon)
	}
	if items[1].name != "Single Box" {
		t.Errorf("item 1 name = %q, want Single Box", items[1].name)
	}
	if items[2].name != "Fill" {
		t.Errorf("item 2 name = %q, want Fill", items[2].name)
	}
	if items[3].name != "Select" {
		t.Errorf("item 3 name = %q, want Select", items[3].name)
	}
}

func TestToolPickerItemsEllipseSelected(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Ellipse"
	m.circleMode = true

	items := m.toolPickerItems()
	if items[0].name != "Oval" {
		t.Errorf("drawing group name = %q, want Oval", items[0].name)
	}
	if !items[0].selected {
		t.Error("drawing group should be selected when Ellipse is active")
	}
}

func TestToolPickerItemsBoxSelected(t *testing.T) {
	m := newTestModel(10, 10)
	m.selectedTool = "Box"

	items := m.toolPickerItems()
	if items[0].selected {
		t.Error("drawing group should not be selected when Box is active")
	}
	if !items[1].selected {
		t.Error("Box item should be selected")
	}
}

func TestToolPickerUpDownNavigation(t *testing.T) {
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

	// Down from drawing group (index 0) should go to Box (index 1)
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Box" {
		t.Errorf("down from drawing group: selectedTool = %q, want Box", m.selectedTool)
	}

	// Down from Box (index 1) should go to Fill (index 2)
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Fill" {
		t.Errorf("down from Box: selectedTool = %q, want Fill", m.selectedTool)
	}

	// Down from Fill (index 2) should go to Select (index 3)
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Select" {
		t.Errorf("down from Fill: selectedTool = %q, want Select", m.selectedTool)
	}

	// Down from Select (index 3) should stay at Select
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Select" {
		t.Errorf("down at bottom: selectedTool = %q, want Select", m.selectedTool)
	}

	// Up from Select should go to Fill
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.selectedTool != "Fill" {
		t.Errorf("up from Select: selectedTool = %q, want Fill", m.selectedTool)
	}

	// Up all the way back to drawing group should restore Point
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.selectedTool != "Point" {
		t.Errorf("up to drawing group: selectedTool = %q, want Point", m.selectedTool)
	}
}

func TestToolPickerRemembersDrawingTool(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(80, 30),
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Rectangle",
		drawingTool:     "Rectangle",
		width:           80,
		height:          31,
		ready:           true,
		showToolPicker:  true,
	}

	// Navigate away from drawing group
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown}) // Box
	if m.selectedTool != "Box" {
		t.Fatalf("expected Box, got %q", m.selectedTool)
	}

	// Navigate back - should restore Rectangle
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.selectedTool != "Rectangle" {
		t.Errorf("should restore drawing tool Rectangle, got %q", m.selectedTool)
	}
}

func TestDrawingToolsSubmenuNavigation(t *testing.T) {
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

	// Right arrow on drawing group should open submenu
	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if !m.toolPickerFocusOnStyle {
		t.Error("right arrow on drawing group should set toolPickerFocusOnStyle")
	}

	// Down should go to Rectangle
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Rectangle" {
		t.Errorf("down in drawing submenu: selectedTool = %q, want Rectangle", m.selectedTool)
	}

	// Down to Ellipse
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Ellipse" || m.circleMode {
		t.Errorf("expected Ellipse (circleMode=false), got %q circleMode=%v", m.selectedTool, m.circleMode)
	}

	// Down to Oval
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Ellipse" || !m.circleMode {
		t.Errorf("expected Ellipse (circleMode=true/Oval), got %q circleMode=%v", m.selectedTool, m.circleMode)
	}

	// Down to Line
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Line" {
		t.Errorf("expected Line, got %q", m.selectedTool)
	}

	// Down at bottom stays
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedTool != "Line" {
		t.Errorf("down at bottom: expected Line, got %q", m.selectedTool)
	}

	// Left exits submenu
	m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if m.toolPickerFocusOnStyle {
		t.Error("left should exit drawing submenu")
	}
	if !m.showToolPicker {
		t.Error("tool picker should remain open")
	}
}

func TestDrawingToolsSubmenuEnterCloses(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Line",
		drawingTool:           "Line",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.showToolPicker {
		t.Error("enter should close tool picker")
	}
	if m.selectedTool != "Line" {
		t.Errorf("should preserve tool, got %q", m.selectedTool)
	}
}

func TestDrawingToolsSubmenuNumberKeys(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Point",
		drawingTool:           "Point",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
	}

	// Press 3 to select Ellipse
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.selectedTool != "Ellipse" || m.circleMode {
		t.Errorf("key 3: expected Ellipse, got %q circleMode=%v", m.selectedTool, m.circleMode)
	}

	// Press 4 to select Oval
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	if m.selectedTool != "Ellipse" || !m.circleMode {
		t.Errorf("key 4: expected Oval, got %q circleMode=%v", m.selectedTool, m.circleMode)
	}

	// Press 1 to select Points
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.selectedTool != "Point" {
		t.Errorf("key 1: expected Point, got %q", m.selectedTool)
	}
}

func TestToolPickerNumberKeysTopLevel(t *testing.T) {
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

	// Press 2 to select Box
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	if m.selectedTool != "Box" {
		t.Errorf("key 2 top-level: expected Box, got %q", m.selectedTool)
	}

	// Press 3 to select Fill
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.selectedTool != "Fill" {
		t.Errorf("key 3 top-level: expected Fill, got %q", m.selectedTool)
	}

	// Press 1 to select drawing group (should restore last drawing tool)
	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	if m.selectedTool != "Point" {
		t.Errorf("key 1 top-level: expected Point (drawing tool), got %q", m.selectedTool)
	}
}

// --- Right arrow on non-submenu tool should switch menus ---

func TestRightArrowOnFillSwitchesMenu(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(80, 30),
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Fill",
		drawingTool:     "Point",
		width:           80,
		height:          31,
		ready:           true,
		showToolPicker:  true,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if m.toolPickerFocusOnStyle {
		t.Error("right on Fill should not open submenu")
	}
	if m.showToolPicker {
		t.Error("should have switched to next menu")
	}
}

// --- Box style submenu tests (still valid) ---

func TestBoxStyleSubmenuKeyboardNavigation(t *testing.T) {
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

	// Right arrow when Box is selected should focus on style submenu
	m.handleKey(tea.KeyMsg{Type: tea.KeyRight})
	if !m.toolPickerFocusOnStyle {
		t.Error("right arrow on Box tool should set toolPickerFocusOnStyle")
	}
	if !m.showToolPicker {
		t.Error("tool picker should remain open")
	}

	// Down arrow in style submenu should change boxStyle
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.boxStyle != 1 {
		t.Errorf("down in style submenu: boxStyle = %d, want 1", m.boxStyle)
	}

	// Up arrow should go back
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.boxStyle != 0 {
		t.Errorf("up in style submenu: boxStyle = %d, want 0", m.boxStyle)
	}

	// Left arrow should exit style submenu
	m.handleKey(tea.KeyMsg{Type: tea.KeyLeft})
	if m.toolPickerFocusOnStyle {
		t.Error("left arrow should clear toolPickerFocusOnStyle")
	}
	if !m.showToolPicker {
		t.Error("tool picker should remain open after left")
	}
}

func TestBoxStyleSubmenuEnterCloses(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		drawingTool:           "Point",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
		boxStyle:              2,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyEnter})
	if m.showToolPicker {
		t.Error("enter in style submenu should close tool picker")
	}
	if m.boxStyle != 2 {
		t.Errorf("enter should preserve boxStyle, got %d", m.boxStyle)
	}
}

func TestBoxStyleSubmenuEscGoesBack(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		drawingTool:           "Point",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if m.toolPickerFocusOnStyle {
		t.Error("esc should clear toolPickerFocusOnStyle")
	}
	if !m.showToolPicker {
		t.Error("tool picker should remain open after esc from style submenu")
	}
}

func TestBoxStyleSubmenuNumberKeys(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		drawingTool:           "Point",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
		boxStyle:              0,
	}

	m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	if m.boxStyle != 2 {
		t.Errorf("number key 3 should set boxStyle to 2, got %d", m.boxStyle)
	}
}

func TestBoxStyleSubmenuUpDownBounds(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		drawingTool:           "Point",
		width:                 80,
		height:                31,
		ready:                 true,
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
		boxStyle:              0,
	}

	// Up at top should not go negative
	m.handleKey(tea.KeyMsg{Type: tea.KeyUp})
	if m.boxStyle != 0 {
		t.Errorf("up at top: boxStyle = %d, want 0", m.boxStyle)
	}

	// Down to last
	m.boxStyle = len(boxStyles) - 1
	m.handleKey(tea.KeyMsg{Type: tea.KeyDown})
	if m.boxStyle != len(boxStyles)-1 {
		t.Errorf("down at bottom: boxStyle = %d, want %d", m.boxStyle, len(boxStyles)-1)
	}
}

func TestCloseMenusResetsStyleFocus(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		showToolPicker:        true,
		toolPickerFocusOnStyle: true,
	}

	m.closeMenus()
	if m.toolPickerFocusOnStyle {
		t.Error("closeMenus should reset toolPickerFocusOnStyle")
	}
}

func TestOpenMenuResetsStyleFocus(t *testing.T) {
	m := &model{
		canvas:                NewCanvas(80, 30),
		selectedChar:          "●",
		foregroundColor:       "white",
		backgroundColor:       "transparent",
		selectedTool:          "Box",
		toolPickerFocusOnStyle: true,
	}

	m.openMenu(3)
	if m.toolPickerFocusOnStyle {
		t.Error("openMenu should reset toolPickerFocusOnStyle")
	}
}

func TestDrawingToolOptions(t *testing.T) {
	expected := []struct {
		name     string
		toolName string
	}{
		{"Points", "Point"},
		{"Rectangle", "Rectangle"},
		{"Ellipse", "Ellipse"},
		{"Oval", "Ellipse"},
		{"Line", "Line"},
	}

	if len(drawingToolOptions) != len(expected) {
		t.Fatalf("drawingToolOptions count = %d, want %d", len(drawingToolOptions), len(expected))
	}

	for i, e := range expected {
		if drawingToolOptions[i].name != e.name {
			t.Errorf("drawingToolOptions[%d].name = %q, want %q", i, drawingToolOptions[i].name, e.name)
		}
		if drawingToolOptions[i].toolName != e.toolName {
			t.Errorf("drawingToolOptions[%d].toolName = %q, want %q", i, drawingToolOptions[i].toolName, e.toolName)
		}
	}
}

func TestIsDrawingTool(t *testing.T) {
	for _, name := range []string{"Point", "Rectangle", "Ellipse", "Line"} {
		if !isDrawingTool(name) {
			t.Errorf("isDrawingTool(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"Box", "Fill", "Select"} {
		if isDrawingTool(name) {
			t.Errorf("isDrawingTool(%q) = true, want false", name)
		}
	}
}
