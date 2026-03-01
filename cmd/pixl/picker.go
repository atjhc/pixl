package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m *model) renderCategoryPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	focusedBg := themeColor(m.config.Theme.MenuSelectedBg)
	unfocusedBg := themeColor(m.config.Theme.MenuUnfocusedBg)

	selectedBg := focusedBg
	if m.toolPickerFocusLevel != 2 {
		selectedBg = unfocusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(themeColor(m.config.Theme.MenuSelectedFg))

	maxNameWidth := 0
	for _, group := range characterGroups {
		if w := lipgloss.Width(group.name); w > maxNameWidth {
			maxNameWidth = w
		}
	}
	lineWidth := maxNameWidth + 2

	var content strings.Builder
	for i, group := range characterGroups {
		line := " " + group.name
		for lipgloss.Width(line) < lineWidth {
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

func (m *model) renderGlyphsPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	focusedBg := themeColor(m.config.Theme.MenuSelectedBg)
	unfocusedBg := themeColor(m.config.Theme.MenuUnfocusedBg)

	selectedBg := unfocusedBg
	if m.toolPickerFocusLevel == 3 {
		selectedBg = focusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(themeColor(m.config.Theme.MenuSelectedFg))

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

func (m *model) renderDrawingToolPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	focusedBg := themeColor(m.config.Theme.MenuSelectedBg)
	unfocusedBg := themeColor(m.config.Theme.MenuUnfocusedBg)

	selectedBg := unfocusedBg
	if m.toolPickerFocusLevel == 1 {
		selectedBg = focusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(themeColor(m.config.Theme.MenuSelectedFg))

	highlightIdx := m.toolSubmenuIndex()

	iconCol := lipgloss.Width(m.selectedChar)

	glyphLabel := "Glyph"
	maxNameWidth := lipgloss.Width(glyphLabel)
	for _, opt := range drawingToolOptions {
		if w := lipgloss.Width(opt.name); w > maxNameWidth {
			maxNameWidth = w
		}
	}

	totalEntries := len(drawingToolOptions) + 1
	lineWidth := 1 + iconCol + 1 + maxNameWidth + 1

	var content strings.Builder

	// Entry 0: Glyph selector
	icon := m.selectedChar
	for lipgloss.Width(icon) < iconCol {
		icon += " "
	}
	line := " " + icon + " " + glyphLabel
	for lipgloss.Width(line) < lineWidth {
		line += " "
	}
	if highlightIdx == 0 {
		content.WriteString(selectedStyle.Render(line))
	} else {
		content.WriteString(line)
	}
	content.WriteString("\n")

	// Entries 1+: drawing tool options
	for i, opt := range drawingToolOptions {
		padIcon := ""
		for lipgloss.Width(padIcon) < iconCol {
			padIcon += " "
		}
		line := " " + padIcon + " " + opt.name
		for lipgloss.Width(line) < lineWidth {
			line += " "
		}

		if i+1 == highlightIdx {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
		if i < totalEntries-2 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderToolSubmenuPicker() string {
	if isDrawingTool(m.selectedTool) {
		return m.renderDrawingToolPicker()
	}
	if m.selectedTool == "Box" {
		return m.renderBoxStylePicker()
	}
	return ""
}

func (m *model) renderToolPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	focusedBg := themeColor(m.config.Theme.MenuSelectedBg)
	unfocusedBg := themeColor(m.config.Theme.MenuUnfocusedBg)

	selectedBg := focusedBg
	if m.toolPickerFocusLevel >= 1 {
		selectedBg = unfocusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(themeColor(m.config.Theme.MenuSelectedFg))

	items := m.toolPickerItems()

	iconCol := 0
	for _, item := range items {
		if w := lipgloss.Width(item.icon); w > iconCol {
			iconCol = w
		}
	}

	maxNameWidth := 0
	for _, item := range items {
		if w := lipgloss.Width(item.name); w > maxNameWidth {
			maxNameWidth = w
		}
	}

	var content strings.Builder
	for i, item := range items {
		var line string
		if iconCol > 0 {
			icon := item.icon
			for lipgloss.Width(icon) < iconCol {
				icon = " " + icon
			}
			line = " " + icon + " " + item.name
		} else {
			line = " " + item.name
		}
		lineWidth := 1 + iconCol + 1 + maxNameWidth + 1
		if iconCol == 0 {
			lineWidth = 1 + maxNameWidth + 1
		}
		for lipgloss.Width(line) < lineWidth {
			line += " "
		}

		if item.selected {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
		if i < len(items)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderBoxStylePicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	focusedBg := themeColor(m.config.Theme.MenuSelectedBg)
	unfocusedBg := themeColor(m.config.Theme.MenuUnfocusedBg)

	selectedBg := unfocusedBg
	if m.toolPickerFocusLevel == 1 {
		selectedBg = focusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(themeColor(m.config.Theme.MenuSelectedFg))

	var content strings.Builder
	for i, s := range boxStyles {
		line := fmt.Sprintf(" %s%s %s ", s.tl, s.tr, s.name)

		if i == m.boxStyle {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
		if i < len(boxStyles)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m *model) renderColorPicker(title string) string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(themeColor(m.config.Theme.MenuBorder))

	selectedStyle := lipgloss.NewStyle().
		Background(themeColor(m.config.Theme.MenuSelectedBg)).
		Foreground(themeColor(m.config.Theme.MenuSelectedFg))

	currentColor := m.foregroundColor
	if title == "Background" {
		currentColor = m.backgroundColor
	}

	maxNameWidth := 0
	for _, c := range colors {
		if w := lipgloss.Width(colorDisplayName(c.name)); w > maxNameWidth {
			maxNameWidth = w
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

		for lipgloss.Width(displayName) < maxNameWidth {
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
