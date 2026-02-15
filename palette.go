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

// Available tools
var tools = []string{
	"Point",
	"Rectangle",
	"Ellipse",
	"Fill",
	"Select",
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

func colorDisplayName(name string) string {
	if name == "transparent" {
		return "None"
	}
	display := strings.ReplaceAll(name, "_", " ")
	return strings.ToUpper(display[:1]) + display[1:]
}

func colorStyleByName(name string) lipgloss.Style {
	for _, c := range colors {
		if c.name == name {
			return c.style
		}
	}
	return lipgloss.NewStyle()
}

func (m *model) findSelectedToolIndex() int {
	for i, tool := range tools {
		if tool == m.selectedTool {
			return i
		}
	}
	return 0
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
