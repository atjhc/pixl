package main

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Available characters grouped by type
var characterGroups = []struct {
	name  string
	chars []string
}{
	{"Circles", []string{"○", "◌", "◍", "◎", "●", "◐", "◑", "◒", "◓", "◔", "◕", "◖", "◗"}},
	{"Squares", []string{"■", "□", "▪", "▫", "▮"}},
	{"Triangles", []string{"▲", "△", "▼", "▽", "◀", "◁", "▶", "▷", "◢", "◣", "◤", "◥"}},
	{"Diamonds", []string{"◆", "◇", "◈", "⬥", "⬦"}},
	{"Stars", []string{"★", "☆", "✦", "✧", "✪", "✫", "✬", "✭", "✮", "✯", "✰"}},
	{"Blocks", []string{"▀", "▄", "▌", "▐", "▖", "▗", "▘", "▝", "▞", "▟", "▙", "▚", "▛", "▜"}},
	{"Shading", []string{"█", "▓", "▒", "░"}},
	{"Dots", []string{"•", "∙", "․", "⋅", "▪", "▫"}},
	{"Box Single", []string{"─", "│", "┌", "┐", "└", "┘", "├", "┤", "┬", "┴", "┼"}},
	{"Box Double", []string{"═", "║", "╔", "╗", "╚", "╝", "╠", "╣", "╦", "╩", "╬"}},
	{"Box Diag", []string{"╱", "╲", "╳", "⁄"}},
	{"Curves", []string{"◜", "◝", "◞", "◟", "╭", "╮", "╰", "╯"}},
	{"Arrows", []string{"←", "→", "↑", "↓", "↖", "↗", "↘", "↙", "⬆", "⬇", "⬅", "➡"}},
	{"Hearts", []string{"♥", "♡", "♠", "♣", "♦"}},
	{"Weather", []string{"☀", "☁", "☂", "☃", "❄", "⛈"}},
	{"Symbols", []string{"☺", "☻", "✓", "✗", "⚙", "⚠", "☢"}},
}

// Available colors
var colors = []struct {
	name  string
	style lipgloss.Style
}{
	{"transparent", lipgloss.NewStyle()},
	{"black", lipgloss.NewStyle().Foreground(lipgloss.Color("0"))},
	{"red", lipgloss.NewStyle().Foreground(lipgloss.Color("1"))},
	{"green", lipgloss.NewStyle().Foreground(lipgloss.Color("2"))},
	{"yellow", lipgloss.NewStyle().Foreground(lipgloss.Color("3"))},
	{"blue", lipgloss.NewStyle().Foreground(lipgloss.Color("4"))},
	{"magenta", lipgloss.NewStyle().Foreground(lipgloss.Color("5"))},
	{"cyan", lipgloss.NewStyle().Foreground(lipgloss.Color("6"))},
	{"white", lipgloss.NewStyle().Foreground(lipgloss.Color("7"))},
	{"bright_black", lipgloss.NewStyle().Foreground(lipgloss.Color("8"))},
	{"bright_red", lipgloss.NewStyle().Foreground(lipgloss.Color("9"))},
	{"bright_green", lipgloss.NewStyle().Foreground(lipgloss.Color("10"))},
	{"bright_yellow", lipgloss.NewStyle().Foreground(lipgloss.Color("11"))},
	{"bright_blue", lipgloss.NewStyle().Foreground(lipgloss.Color("12"))},
	{"bright_magenta", lipgloss.NewStyle().Foreground(lipgloss.Color("13"))},
	{"bright_cyan", lipgloss.NewStyle().Foreground(lipgloss.Color("14"))},
	{"bright_white", lipgloss.NewStyle().Foreground(lipgloss.Color("15"))},
}

var ansiColorCodes = map[string]string{
	"black":          "30",
	"red":            "31",
	"green":          "32",
	"yellow":         "33",
	"blue":           "34",
	"magenta":        "35",
	"cyan":           "36",
	"white":          "",
	"bright_black":   "90",
	"bright_red":     "91",
	"bright_green":   "92",
	"bright_yellow":  "93",
	"bright_blue":    "94",
	"bright_magenta": "95",
	"bright_cyan":    "96",
	"bright_white":   "97",
}

var ansiBgColorCodes = map[string]string{
	"black":          "40",
	"red":            "41",
	"green":          "42",
	"yellow":         "43",
	"blue":           "44",
	"magenta":        "45",
	"cyan":           "46",
	"white":          "47",
	"bright_black":   "100",
	"bright_red":     "101",
	"bright_green":   "102",
	"bright_yellow":  "103",
	"bright_blue":    "104",
	"bright_magenta": "105",
	"bright_cyan":    "106",
	"bright_white":   "107",
}

func normalizeColorName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func colorToANSI(name string) string {
	return ansiColorCodes[normalizeColorName(name)]
}

func colorToANSIBg(name string) string {
	return ansiBgColorCodes[normalizeColorName(name)]
}

func colorDisplayName(name string) string {
	if name == "transparent" {
		return "None"
	}
	display := strings.ReplaceAll(name, "_", " ")
	return strings.ToUpper(display[:1]) + display[1:]
}

func colorStyleByName(name string) lipgloss.Style {
	normalized := normalizeColorName(name)
	for _, c := range colors {
		if c.name == normalized {
			return c.style
		}
	}
	return lipgloss.NewStyle()
}

func (m *model) findSelectedCharCategory() int {
	for i, group := range characterGroups {
		for _, char := range group.chars {
			if char == m.selectedChar {
				return i
			}
		}
	}
	return 0
}

func (m *model) findSelectedCharIndexInCategory(categoryIdx int) int {
	if categoryIdx < 0 || categoryIdx >= len(characterGroups) {
		return 0
	}
	for i, char := range characterGroups[categoryIdx].chars {
		if char == m.selectedChar {
			return i
		}
	}
	return 0
}

func (m *model) findSelectedColorIndex(colorName string) int {
	for i, color := range colors {
		if color.name == colorName {
			return i
		}
	}
	return 0
}
