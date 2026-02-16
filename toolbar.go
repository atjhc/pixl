package main

import (
	"fmt"

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
	bgColor := lipgloss.Color("#0E7490")
	baseColor := lipgloss.Color("#E0E0E0")
	highlightColor := lipgloss.Color("#FFFFFF")

	baseStyle := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(baseColor).
		Padding(0, 1)

	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")).
		Foreground(highlightColor).
		Padding(0, 1)

	underlineOn := "\x1b[4m"
	underlineOff := "\x1b[24m"

	sep := " "

	currentX := 0

	// Shape button
	shapeText := fmt.Sprintf("%sS%shapes: %s", underlineOn, underlineOff, m.selectedChar)
	var shapeButton string
	if m.showCharPicker {
		shapeButton = highlightStyle.Render(shapeText)
	} else {
		shapeButton = baseStyle.Render(shapeText)
	}
	m.toolbarShapeX = currentX + toolbarButtonPadding
	m.toolbarShapeItemX = currentX + 9
	currentX += lipgloss.Width(shapeButton)
	currentX += 1 // separator

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
		fgButton = highlightStyle.Copy().Padding(0, 0, 0, 1).Render(fgText)
	} else {
		fgButton = baseStyle.Copy().Padding(0, 0, 0, 1).Render(fgText)
	}
	m.toolbarForegroundX = currentX + toolbarButtonPadding
	m.toolbarForegroundItemX = currentX + 13
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
		bgButton = highlightStyle.Copy().Padding(0, 0, 0, 1).Render(bgText)
	} else {
		bgButton = baseStyle.Copy().Padding(0, 0, 0, 1).Render(bgText)
	}
	m.toolbarBackgroundX = currentX + toolbarButtonPadding
	m.toolbarBackgroundItemX = currentX + 13
	currentX += lipgloss.Width(bgButton)
	currentX += 1 // separator

	// Tool button
	toolName := m.selectedTool
	if m.selectedTool == "Ellipse" && m.circleMode {
		toolName = "Circle"
	}
	toolText := fmt.Sprintf("%sT%sool: %s", underlineOn, underlineOff, toolName)
	var toolButton string
	if m.showToolPicker {
		toolButton = highlightStyle.Render(toolText)
	} else {
		toolButton = baseStyle.Render(toolText)
	}
	m.toolbarToolX = currentX + toolbarButtonPadding
	m.toolbarToolItemX = currentX + 7
	currentX += lipgloss.Width(toolButton)

	// Mode indicator
	modeIndicator := ""
	if m.clipboard != nil && m.clipboardHeight > 0 && m.clipboardWidth > 0 {
		modeText := fmt.Sprintf("Mode: Yank (%dx%d)", m.clipboardWidth, m.clipboardHeight)
		modeIndicator = baseStyle.Render(modeText)
	}

	barContent := shapeButton + sep + fgButton + sep + bgButton + sep + toolButton + sep + modeIndicator

	barStyle := lipgloss.NewStyle().
		Background(bgColor).
		Width(m.width)

	return barStyle.Render(barContent) + "\n"
}
