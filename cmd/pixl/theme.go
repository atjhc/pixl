package main

import (
	"strconv"

	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	ToolbarBg          string
	ToolbarFg          string
	ToolbarHighlightBg string
	ToolbarHighlightFg string
	MenuBorder         string
	MenuSelectedBg     string
	MenuSelectedFg     string
	MenuUnfocusedBg    string
	CanvasBorder       string
	SelectionFg        string
	CursorFg           string
}

func defaultTheme() Theme {
	return Theme{
		ToolbarBg:          "cyan",
		ToolbarFg:          "bright-white",
		ToolbarHighlightBg: "bright-cyan",
		ToolbarHighlightFg: "bright-white",
		MenuBorder:         "bright-blue",
		MenuSelectedBg:     "bright-cyan",
		MenuSelectedFg:     "bright-white",
		MenuUnfocusedBg:    "bright-black",
		CanvasBorder:       "white",
		SelectionFg:        "bright-blue",
		CursorFg:           "bright-black",
	}
}

func (t *Theme) field(key string) *string {
	switch key {
	case "toolbar-bg":
		return &t.ToolbarBg
	case "toolbar-fg":
		return &t.ToolbarFg
	case "toolbar-highlight-bg":
		return &t.ToolbarHighlightBg
	case "toolbar-highlight-fg":
		return &t.ToolbarHighlightFg
	case "menu-border":
		return &t.MenuBorder
	case "menu-selected-bg":
		return &t.MenuSelectedBg
	case "menu-selected-fg":
		return &t.MenuSelectedFg
	case "menu-unfocused-bg":
		return &t.MenuUnfocusedBg
	case "canvas-border":
		return &t.CanvasBorder
	case "selection-fg":
		return &t.SelectionFg
	case "cursor-fg":
		return &t.CursorFg
	}
	return nil
}

var themeColorNames = map[string]string{
	"black":          "0",
	"red":            "1",
	"green":          "2",
	"yellow":         "3",
	"blue":           "4",
	"magenta":        "5",
	"cyan":           "6",
	"white":          "7",
	"bright-black":   "8",
	"bright-red":     "9",
	"bright-green":   "10",
	"bright-yellow":  "11",
	"bright-blue":    "12",
	"bright-magenta": "13",
	"bright-cyan":    "14",
	"bright-white":   "15",
}

// themeColor resolves a color name, ANSI number, or hex string to a lipgloss.Color.
// Named colors (e.g. "bright-cyan") map to the standard 16 ANSI colors.
// Raw ANSI numbers (e.g. "14") and hex values (e.g. "#0891B2") pass through as-is.
func themeColor(s string) lipgloss.Color {
	if code, ok := themeColorNames[s]; ok {
		return lipgloss.Color(code)
	}
	return lipgloss.Color(s)
}

// isValidThemeColor returns true if s is a recognized color name, ANSI number, or hex value.
func isValidThemeColor(s string) bool {
	if _, ok := themeColorNames[s]; ok {
		return true
	}
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	if len(s) == 7 && s[0] == '#' {
		for _, c := range s[1:] {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
		return true
	}
	return false
}
