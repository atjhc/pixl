package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	controlBarHeight = 1

	pickerBorderWidth   = 2 // Top and bottom borders (or left and right)
	pickerItemPadding   = 1 // Leading space before content
	pickerSwatchWidth   = 2 // "██" for colors
	pickerItemSeparator = 1 // Space between swatch and name

	// Picker content offset (border + padding = where content starts)
	pickerContentOffset = 1 + pickerItemPadding

	toolbarButtonPadding = 1 // Left/right padding inside each button
)

func (m *model) renderControlBar() string {
	bgColor := themeColor(m.config.Theme.ToolbarBg)
	baseColor := themeColor(m.config.Theme.ToolbarFg)
	highlightColor := themeColor(m.config.Theme.ToolbarHighlightFg)

	baseStyle := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(baseColor).
		Padding(0, 1)

	highlightStyle := lipgloss.NewStyle().
		Background(themeColor(m.config.Theme.ToolbarHighlightBg)).
		Foreground(highlightColor).
		Padding(0, 1)

	underlineOn := "\x1b[4m"
	underlineOff := "\x1b[24m"

	sep := " "

	currentX := 0

	// Foreground color button
	var fgSwatch string
	if m.foregroundColor == "transparent" {
		fgSwatch = "  "
	} else {
		fgSwatch = colorStyleByName(m.foregroundColor).Render("██")
	}
	fgText := fmt.Sprintf("%sF%soreground: %s", underlineOn, underlineOff, fgSwatch)
	var fgButton string
	if m.showFgPicker {
		fgButton = highlightStyle.Render(fgText)
	} else {
		fgButton = baseStyle.Render(fgText)
	}
	m.toolbar.foregroundX = currentX + toolbarButtonPadding
	m.toolbar.foregroundItemX = currentX + 13
	currentX += lipgloss.Width(fgButton)
	currentX += 1 // separator

	// Background color button
	var bgSwatch string
	if m.backgroundColor == "transparent" {
		bgSwatch = "  "
	} else {
		bgSwatch = colorStyleByName(m.backgroundColor).Render("██")
	}
	bgText := fmt.Sprintf("%sB%sackground: %s", underlineOn, underlineOff, bgSwatch)
	var bgButton string
	if m.showBgPicker {
		bgButton = highlightStyle.Render(bgText)
	} else {
		bgButton = baseStyle.Render(bgText)
	}
	m.toolbar.backgroundX = currentX + toolbarButtonPadding
	m.toolbar.backgroundItemX = currentX + 13
	currentX += lipgloss.Width(bgButton)
	currentX += 1 // separator

	// Tool button
	toolName := m.tool().DisplayName(m)
	toolText := fmt.Sprintf("%sT%sool: %s", underlineOn, underlineOff, toolName)
	var toolButton string
	if m.showToolPicker {
		toolButton = highlightStyle.Render(toolText)
	} else {
		toolButton = baseStyle.Render(toolText)
	}
	m.toolbar.toolX = currentX + toolbarButtonPadding
	m.toolbar.toolItemX = currentX + 7
	currentX += lipgloss.Width(toolButton)

	// Mode indicator
	modeIndicator := ""
	if m.clipboard.cells != nil && m.clipboard.height > 0 && m.clipboard.width > 0 {
		modeText := fmt.Sprintf("Mode: Yank (%dx%d)", m.clipboard.width, m.clipboard.height)
		modeIndicator = baseStyle.Render(modeText)
	}

	barContent := fgButton + sep + bgButton + sep + toolButton + sep + modeIndicator

	fileIndicator := ""
	if m.filePath != "" {
		displayName := shortPath(m.filePath)
		contentWidth := lipgloss.Width(barContent)
		nameWidth := lipgloss.Width(displayName)
		gap := m.width - contentWidth - nameWidth - 1
		if gap > 0 {
			fileIndicator = baseStyle.Copy().Padding(0, 0).Render(
				strings.Repeat(" ", gap) + displayName,
			)
		} else {
			fileIndicator = baseStyle.Copy().Padding(0, 0, 0, 1).Render(displayName)
		}
	}

	barStyle := lipgloss.NewStyle().
		Background(bgColor).
		Width(m.width)

	return barStyle.Render(barContent + fileIndicator) + "\n"
}

func shortPath(path string) string {
	dir, file := filepath.Split(path)
	dir = filepath.Clean(dir)
	parent := filepath.Base(dir)
	if parent == "." || parent == "/" {
		return file
	}
	return parent + "/" + file
}
