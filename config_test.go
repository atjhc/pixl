package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T, content string) {
	t.Helper()
	dir := t.TempDir()
	configDir := filepath.Join(dir, ".config", "pixl")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config"), []byte(content), 0644)
	t.Setenv("HOME", dir)
}

func TestLoadConfigDefaults(t *testing.T) {
	c := loadConfig()
	if !c.MergeBoxBorders {
		t.Error("MergeBoxBorders should default to true")
	}
	if c.DefaultGlyph != "" {
		t.Errorf("DefaultGlyph should default to empty, got %q", c.DefaultGlyph)
	}
	if c.DefaultForeground != "" {
		t.Errorf("DefaultForeground should default to empty, got %q", c.DefaultForeground)
	}
	if c.DefaultBackground != "" {
		t.Errorf("DefaultBackground should default to empty, got %q", c.DefaultBackground)
	}
	if c.DefaultTool != "" {
		t.Errorf("DefaultTool should default to empty, got %q", c.DefaultTool)
	}
	if c.DefaultBoxStyle != "" {
		t.Errorf("DefaultBoxStyle should default to empty, got %q", c.DefaultBoxStyle)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	writeTestConfig(t, "merge-box-borders = false\n")

	c := loadConfig()
	if c.MergeBoxBorders {
		t.Error("MergeBoxBorders should be false when config says false")
	}
}

func TestLoadConfigIgnoresComments(t *testing.T) {
	writeTestConfig(t, "# a comment\nmerge-box-borders = true\n")

	c := loadConfig()
	if !c.MergeBoxBorders {
		t.Error("MergeBoxBorders should be true")
	}
}

func TestLoadConfigDefaultGlyph(t *testing.T) {
	writeTestConfig(t, "default-glyph = #\n")

	c := loadConfig()
	if c.DefaultGlyph != "#" {
		t.Errorf("DefaultGlyph = %q, want #", c.DefaultGlyph)
	}
}

func TestLoadConfigDefaultForeground(t *testing.T) {
	writeTestConfig(t, "default-foreground = red\n")

	c := loadConfig()
	if c.DefaultForeground != "red" {
		t.Errorf("DefaultForeground = %q, want red", c.DefaultForeground)
	}
}

func TestLoadConfigDefaultBackground(t *testing.T) {
	writeTestConfig(t, "default-background = blue\n")

	c := loadConfig()
	if c.DefaultBackground != "blue" {
		t.Errorf("DefaultBackground = %q, want blue", c.DefaultBackground)
	}
}

func TestLoadConfigDefaultTool(t *testing.T) {
	writeTestConfig(t, "default-tool = Line\n")

	c := loadConfig()
	if c.DefaultTool != "Line" {
		t.Errorf("DefaultTool = %q, want Line", c.DefaultTool)
	}
}

func TestLoadConfigDefaultBoxStyle(t *testing.T) {
	writeTestConfig(t, "default-box-style = Double\n")

	c := loadConfig()
	if c.DefaultBoxStyle != "Double" {
		t.Errorf("DefaultBoxStyle = %q, want Double", c.DefaultBoxStyle)
	}
}

func TestConfigAppliedToModel(t *testing.T) {
	m := initialModel()
	m.config = Config{
		DefaultGlyph:      "X",
		DefaultForeground: "red",
		DefaultBackground: "blue",
		DefaultTool:       "Line",
		DefaultBoxStyle:   "Heavy",
		MergeBoxBorders:   true,
	}
	m.applyConfig()

	if m.selectedChar != "X" {
		t.Errorf("selectedChar = %q, want X", m.selectedChar)
	}
	if m.foregroundColor != "red" {
		t.Errorf("foregroundColor = %q, want red", m.foregroundColor)
	}
	if m.backgroundColor != "blue" {
		t.Errorf("backgroundColor = %q, want blue", m.backgroundColor)
	}
	if m.selectedTool != "Line" {
		t.Errorf("selectedTool = %q, want Line", m.selectedTool)
	}
	if m.drawingTool != "Line" {
		t.Errorf("drawingTool = %q, want Line", m.drawingTool)
	}
	if m.boxStyle != 3 {
		t.Errorf("boxStyle = %d, want 3 (Heavy)", m.boxStyle)
	}
}

func TestConfigAppliedToModelDefaults(t *testing.T) {
	m := initialModel()
	m.config = Config{MergeBoxBorders: true}
	m.applyConfig()

	if m.selectedChar != "●" {
		t.Errorf("selectedChar = %q, want ● (unchanged)", m.selectedChar)
	}
	if m.foregroundColor != "white" {
		t.Errorf("foregroundColor = %q, want white (unchanged)", m.foregroundColor)
	}
	if m.selectedTool != "Point" {
		t.Errorf("selectedTool = %q, want Point (unchanged)", m.selectedTool)
	}
}

func TestConfigInvalidToolIgnored(t *testing.T) {
	m := initialModel()
	m.config = Config{DefaultTool: "NonExistent"}
	m.applyConfig()

	if m.selectedTool != "Point" {
		t.Errorf("selectedTool = %q, want Point (invalid tool ignored)", m.selectedTool)
	}
}

func TestConfigInvalidBoxStyleIgnored(t *testing.T) {
	m := initialModel()
	m.config = Config{DefaultBoxStyle: "NonExistent"}
	m.applyConfig()

	if m.boxStyle != 0 {
		t.Errorf("boxStyle = %d, want 0 (invalid style ignored)", m.boxStyle)
	}
}

func TestLoadConfigThemeOverrides(t *testing.T) {
	writeTestConfig(t, `toolbar-bg = red
toolbar-fg = green
toolbar-highlight-bg = blue
toolbar-highlight-fg = black
menu-border = magenta
menu-selected-bg = yellow
menu-selected-fg = white
menu-unfocused-bg = bright-red
canvas-border = bright-green
selection-fg = bright-yellow
cursor-fg = bright-cyan
`)

	c := loadConfig()

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"ToolbarBg", c.Theme.ToolbarBg, "red"},
		{"ToolbarFg", c.Theme.ToolbarFg, "green"},
		{"ToolbarHighlightBg", c.Theme.ToolbarHighlightBg, "blue"},
		{"ToolbarHighlightFg", c.Theme.ToolbarHighlightFg, "black"},
		{"MenuBorder", c.Theme.MenuBorder, "magenta"},
		{"MenuSelectedBg", c.Theme.MenuSelectedBg, "yellow"},
		{"MenuSelectedFg", c.Theme.MenuSelectedFg, "white"},
		{"MenuUnfocusedBg", c.Theme.MenuUnfocusedBg, "bright-red"},
		{"CanvasBorder", c.Theme.CanvasBorder, "bright-green"},
		{"SelectionFg", c.Theme.SelectionFg, "bright-yellow"},
		{"CursorFg", c.Theme.CursorFg, "bright-cyan"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestLoadConfigThemePartialOverride(t *testing.T) {
	writeTestConfig(t, "toolbar-bg = red\n")

	c := loadConfig()

	if c.Theme.ToolbarBg != "red" {
		t.Errorf("ToolbarBg = %q, want red", c.Theme.ToolbarBg)
	}
	if c.Theme.ToolbarFg != "bright-white" {
		t.Errorf("ToolbarFg = %q, want bright-white (default)", c.Theme.ToolbarFg)
	}
	if c.Theme.MenuBorder != "bright-blue" {
		t.Errorf("MenuBorder = %q, want bright-blue (default)", c.Theme.MenuBorder)
	}
}
