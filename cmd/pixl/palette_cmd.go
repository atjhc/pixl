package main

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type paletteItem struct {
	name   string
	action func(m *model)
}

func (m *model) paletteItems() []paletteItem {
	return []paletteItem{
		{"Point", func(m *model) { m.setTool("Point") }},
		{"Rectangle", func(m *model) { m.setTool("Rectangle") }},
		{"Ellipse", func(m *model) { m.setTool("Ellipse") }},
		{"Circle", func(m *model) { m.setTool("Ellipse"); m.circleMode = true }},
		{"Line", func(m *model) { m.setTool("Line") }},
		{"Fill", func(m *model) { m.setTool("Fill") }},
		{"Box", func(m *model) { m.setTool("Box") }},
		{"Select", func(m *model) { m.setTool("Select") }},
		{"Clear Canvas", func(m *model) { m.confirmClear = true }},
		{"Undo", func(m *model) { m.undo() }},
		{"Redo", func(m *model) { m.redo() }},
		{"Copy", func(m *model) { m.copySelection() }},
		{"Cut", func(m *model) { m.cutSelection() }},
		{"Paste", func(m *model) { m.paste() }},
		{"Swap Colors", func(m *model) {
			m.foregroundColor, m.backgroundColor = m.backgroundColor, m.foregroundColor
		}},
		{"Eyedropper", func(m *model) {
			if cell := m.canvas.Get(m.hoverRow, m.hoverCol); cell != nil {
				m.selectedChar = cell.char
				m.foregroundColor = cell.foregroundColor
				m.backgroundColor = cell.backgroundColor
			}
		}},
	}
}

func filterPalette(items []paletteItem, query string) []paletteItem {
	if query == "" {
		return items
	}
	q := strings.ToLower(query)
	var prefix, substring []paletteItem
	for _, item := range items {
		lower := strings.ToLower(item.name)
		if strings.HasPrefix(lower, q) {
			prefix = append(prefix, item)
		} else if strings.Contains(lower, q) {
			substring = append(substring, item)
		}
	}
	return append(prefix, substring...)
}

const paletteMaxVisible = 10

func (m *model) handlePaletteKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.closePalette()
		return m, nil
	case tea.KeyEnter:
		items := filterPalette(m.paletteItems(), m.paletteQuery)
		if m.paletteIndex < len(items) {
			items[m.paletteIndex].action(m)
		}
		m.closePalette()
		return m, nil
	case tea.KeyBackspace:
		if len(m.paletteQuery) == 0 {
			m.closePalette()
			return m, nil
		}
		if msg.Alt {
			m.paletteQuery = deleteWord(m.paletteQuery)
		} else {
			m.paletteQuery = m.paletteQuery[:len(m.paletteQuery)-1]
		}
		m.paletteIndex = 0
		return m, nil
	case tea.KeyUp:
		if m.paletteIndex > 0 {
			m.paletteIndex--
		}
		return m, nil
	case tea.KeyDown:
		items := filterPalette(m.paletteItems(), m.paletteQuery)
		max := len(items) - 1
		if max > paletteMaxVisible-1 {
			max = paletteMaxVisible - 1
		}
		if m.paletteIndex < max {
			m.paletteIndex++
		}
		return m, nil
	case tea.KeySpace:
		m.paletteQuery += " "
		m.paletteIndex = 0
		return m, nil
	case tea.KeyRunes:
		s := string(msg.Runes)
		if s == ":" {
			m.closePalette()
			return m, nil
		}
		m.paletteQuery += s
		m.paletteIndex = 0
		return m, nil
	}
	return m, nil
}

func deleteWord(s string) string {
	i := len(s)
	// Skip trailing spaces
	for i > 0 && s[i-1] == ' ' {
		i--
	}
	// Delete the word
	for i > 0 && s[i-1] != ' ' {
		i--
	}
	return s[:i]
}

func (m *model) closePalette() {
	m.showPalette = false
	m.paletteQuery = ""
	m.paletteIndex = 0
}

func (m *model) renderPalette() string {
	borderColor := themeColor(m.config.Theme.MenuBorder)
	selectedFg := themeColor(m.config.Theme.MenuSelectedFg)

	selectedStyle := lipgloss.NewStyle().
		Foreground(selectedFg)

	items := filterPalette(m.paletteItems(), m.paletteQuery)

	visible := items
	if len(visible) > paletteMaxVisible {
		visible = visible[:paletteMaxVisible]
	}

	// Find max width for consistent padding
	maxWidth := 20
	for _, item := range visible {
		w := lipgloss.Width(item.name) + 4 // "▸ " prefix + padding
		if w > maxWidth {
			maxWidth = w
		}
	}
	inputLine := ": " + m.paletteQuery + "_"
	if w := lipgloss.Width(inputLine); w > maxWidth {
		maxWidth = w
	}

	var content strings.Builder

	// Input line
	content.WriteString(inputLine)
	if pad := maxWidth - lipgloss.Width(inputLine); pad > 0 {
		content.WriteString(strings.Repeat(" ", pad))
	}
	content.WriteString("\n")

	// Separator
	content.WriteString(strings.Repeat("─", maxWidth))
	content.WriteString("\n")

	// Results
	for i, item := range visible {
		var line string
		if i == m.paletteIndex {
			line = "▸ " + item.name
		} else {
			line = "  " + item.name
		}
		if pad := maxWidth - lipgloss.Width(line); pad > 0 {
			line += strings.Repeat(" ", pad)
		}
		if i == m.paletteIndex {
			line = selectedStyle.Render(line)
		}
		content.WriteString(line)
		if i < len(visible)-1 {
			content.WriteString("\n")
		}
	}

	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1)

	return dialogStyle.Render(content.String())
}
