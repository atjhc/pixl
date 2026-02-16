package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) renderCategoryPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	focusedBg := lipgloss.Color("#0891B2")
	unfocusedBg := lipgloss.Color("#3A3A3A")

	selectedBg := focusedBg
	if m.shapesFocusOnPanel {
		selectedBg = unfocusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(lipgloss.Color("#FFFFFF"))

	maxNameLen := 0
	for _, group := range characterGroups {
		if len(group.name) > maxNameLen {
			maxNameLen = len(group.name)
		}
	}
	lineWidth := maxNameLen + 2

	var content strings.Builder
	for i, group := range characterGroups {
		line := " " + group.name
		for len(line) < lineWidth {
			line += " "
		}

		if i == m.selectedCategory {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
		if i < len(characterGroups)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderShapesPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	focusedBg := lipgloss.Color("#0891B2")
	unfocusedBg := lipgloss.Color("#3A3A3A")

	selectedBg := unfocusedBg
	if m.shapesFocusOnPanel {
		selectedBg = focusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(lipgloss.Color("#FFFFFF"))

	var content strings.Builder
	if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
		group := characterGroups[m.selectedCategory]
		for i, char := range group.chars {
			line := " " + char + " "

			if char == m.selectedChar {
				content.WriteString(selectedStyle.Render(line))
			} else {
				content.WriteString(line)
			}
			if i < len(group.chars)-1 {
				content.WriteString("\n")
			}
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderToolPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")).
		Foreground(lipgloss.Color("#FFFFFF"))

	maxNameLen := 0
	for _, t := range toolRegistry {
		name := t.DisplayName(m)
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}
	lineWidth := maxNameLen + 2

	var content strings.Builder
	for i, t := range toolRegistry {
		displayName := t.DisplayName(m)

		line := " " + displayName
		for len(line) < lineWidth {
			line += " "
		}

		if t.Name() == m.selectedTool {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
		if i < len(toolRegistry)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderColorPicker(title string) string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")).
		Foreground(lipgloss.Color("#FFFFFF"))

	currentColor := m.foregroundColor
	if title == "Background" {
		currentColor = m.backgroundColor
	}

	maxNameLen := 0
	for _, c := range colors {
		if len(colorDisplayName(c.name)) > maxNameLen {
			maxNameLen = len(colorDisplayName(c.name))
		}
	}

	var content strings.Builder
	for i, color := range colors {
		var swatch string
		if color.name == "transparent" {
			swatch = "  "
		} else {
			swatch = color.style.Render("██")
		}

		displayName := colorDisplayName(color.name)

		for len(displayName) < maxNameLen {
			displayName += " "
		}

		name := displayName + " "
		if color.name == currentColor {
			name = selectedStyle.Render(name)
		}
		content.WriteString(fmt.Sprintf(" %s %s", swatch, name))
		if i < len(colors)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}
