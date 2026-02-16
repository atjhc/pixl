package main

import "testing"

func TestToolRegistryOrderAndNames(t *testing.T) {
	expected := []string{"Point", "Rectangle", "Ellipse", "Line", "Fill", "Select"}

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
	m.selectedTool = "Ellipse"

	m.circleMode = false
	if got := m.tool().DisplayName(m); got != "Ellipse" {
		t.Errorf("DisplayName with circleMode=false = %q, want Ellipse", got)
	}

	m.circleMode = true
	if got := m.tool().DisplayName(m); got != "Circle" {
		t.Errorf("DisplayName with circleMode=true = %q, want Circle", got)
	}
}

func TestToolCursorChar(t *testing.T) {
	m := newTestModel(10, 10)

	m.selectedTool = "Point"
	if got := m.tool().CursorChar(); got != "" {
		t.Errorf("Point CursorChar = %q, want empty", got)
	}

	for _, name := range []string{"Rectangle", "Ellipse", "Line", "Fill"} {
		m.selectedTool = name
		if got := m.tool().CursorChar(); got != "" {
			t.Errorf("%s CursorChar = %q, want empty (use selected char)", name, got)
		}
	}

	m.selectedTool = "Select"
	if got := m.tool().CursorChar(); got != "┼" {
		t.Errorf("Select CursorChar = %q, want ┼", got)
	}
}

func TestToolModifiesCanvas(t *testing.T) {
	m := newTestModel(10, 10)

	for _, name := range []string{"Point", "Rectangle", "Ellipse", "Line", "Fill"} {
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
