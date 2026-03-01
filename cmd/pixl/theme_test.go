package main

import "testing"

func TestDefaultTheme(t *testing.T) {
	th := defaultTheme()

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"ToolbarBg", th.ToolbarBg, "cyan"},
		{"ToolbarFg", th.ToolbarFg, "bright-white"},
		{"ToolbarHighlightBg", th.ToolbarHighlightBg, "bright-cyan"},
		{"ToolbarHighlightFg", th.ToolbarHighlightFg, "bright-white"},
		{"MenuBorder", th.MenuBorder, "bright-blue"},
		{"MenuSelectedBg", th.MenuSelectedBg, "bright-cyan"},
		{"MenuSelectedFg", th.MenuSelectedFg, "bright-white"},
		{"MenuUnfocusedBg", th.MenuUnfocusedBg, "bright-black"},
		{"CanvasBorder", th.CanvasBorder, "white"},
		{"SelectionFg", th.SelectionFg, "bright-blue"},
		{"CursorFg", th.CursorFg, "bright-black"},
	}

	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %q, want %q", c.name, c.got, c.want)
		}
	}
}

func TestThemeColorResolvesNames(t *testing.T) {
	checks := []struct {
		input string
		want  string
	}{
		{"red", "1"},
		{"bright-blue", "12"},
		{"cyan", "6"},
		{"black", "0"},
		{"bright-white", "15"},
	}

	for _, c := range checks {
		got := themeColor(c.input)
		if string(got) != c.want {
			t.Errorf("themeColor(%q) = %q, want %q", c.input, got, c.want)
		}
	}
}

func TestIsValidThemeColorRejectsInvalidHex(t *testing.T) {
	valid := []string{"#FF0000", "#aabbcc", "#123456", "red", "bright-blue", "42"}
	for _, v := range valid {
		if !isValidThemeColor(v) {
			t.Errorf("isValidThemeColor(%q) = false, want true", v)
		}
	}

	invalid := []string{"#zzzzzz", "#GGGGGG", "#12345g", "#nothex"}
	for _, v := range invalid {
		if isValidThemeColor(v) {
			t.Errorf("isValidThemeColor(%q) = true, want false", v)
		}
	}
}

func TestThemeColorFallsThrough(t *testing.T) {
	got := themeColor("#FF0000")
	if string(got) != "#FF0000" {
		t.Errorf("themeColor(#FF0000) = %q, want #FF0000", got)
	}

	got = themeColor("42")
	if string(got) != "42" {
		t.Errorf("themeColor(42) = %q, want 42", got)
	}
}
