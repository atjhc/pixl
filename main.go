package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Canvas represents the drawing area
type Canvas struct {
	width  int
	height int
	cells  [][]Cell
}

// Cell represents a single cell in the canvas
type Cell struct {
	char            string
	foregroundColor string
	backgroundColor string
}

// NewCanvas creates a new canvas
func NewCanvas(width, height int) Canvas {
	cells := make([][]Cell, height)
	for i := range cells {
		cells[i] = make([]Cell, width)
		for j := range cells[i] {
			cells[i][j] = Cell{char: " ", foregroundColor: "white", backgroundColor: "transparent"}
		}
	}
	return Canvas{width: width, height: height, cells: cells}
}

// Set sets a character and colors at the given position
func (c *Canvas) Set(row, col int, char, fgColor, bgColor string) {
	if row >= 0 && row < c.height && col >= 0 && col < c.width {
		c.cells[row][col] = Cell{char: char, foregroundColor: fgColor, backgroundColor: bgColor}
	}
}

// Get gets the cell at the given position
func (c *Canvas) Get(row, col int) *Cell {
	if row >= 0 && row < c.height && col >= 0 && col < c.width {
		return &c.cells[row][col]
	}
	return nil
}

// Clear clears the canvas
func (c *Canvas) Clear() {
	for i := range c.cells {
		for j := range c.cells[i] {
			c.cells[i][j] = Cell{char: " ", foregroundColor: "white", backgroundColor: "transparent"}
		}
	}
}

// Model represents the application state
type model struct {
	canvas             Canvas
	selectedChar       string
	foregroundColor    string
	backgroundColor    string
	mouseX             int
	mouseY             int
	width              int
	height             int
	ready              bool
	showCharPicker     bool
	showFgPicker       bool
	showBgPicker       bool
	showToolPicker     bool
	selectedTool       string
	selectedCategory   int      // Index of selected character category
	showingShapes      bool     // Whether we're showing shapes (second level) or categories (first level)
	history            []Canvas // Undo history
	historyIndex       int      // Current position in history (-1 means at current state)
	mouseDown          bool     // Whether mouse button is currently pressed
	canvasBeforeStroke Canvas   // Canvas state at start of current stroke
	startX             int      // Start X position for shape tools
	startY             int      // Start Y position for shape tools
	previewEndX        int              // Current end X for preview
	previewEndY        int              // Current end Y for preview
	showPreview        bool             // Whether to show shape preview
	optionKeyHeld      bool             // Whether Option/Alt key is held (for constraining shapes)
	circleMode         bool             // Whether to draw circles instead of ellipses (toggle with Return)
	previewPoints      map[[2]int]bool  // Cached preview points for performance
	hasSelection       bool             // Whether there's an active selection
	selectionStartY    int              // Selection box start Y
	selectionStartX    int              // Selection box start X
	selectionEndY      int              // Selection box end Y
	selectionEndX      int              // Selection box end X
	clipboard          [][]Cell         // Copied cells
	clipboardWidth     int              // Width of clipboard contents
	clipboardHeight    int              // Height of clipboard contents
}

// Available characters grouped by type
var characterGroups = []struct {
	name  string
	chars []string
}{
	{"Circles", []string{"○", "◌", "◍", "◎", "●", "◐", "◑", "◒", "◓", "◔", "◕", "◖", "◗"}},
	{"Squares", []string{"■", "□", "▪", "▫", "◾", "◽", "▮"}},
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
	{"Symbols", []string{"☺", "☻", "✓", "✗", "⚡", "⚙", "⚠", "☢"}},
}

// Available tools
var tools = []string{
	"Point",
	"Rectangle",
	"Ellipse",
	"Select",
}

// Available colors
var colors = []struct {
	name  string
	style lipgloss.Style
}{
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
	{"transparent", lipgloss.NewStyle()},
}

func initialModel() model {
	canvas := NewCanvas(100, 30)
	return model{
		canvas:          canvas,
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Point",
		ready:           false,
		history:         []Canvas{},
		historyIndex:    -1,
		mouseDown:       false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Resize canvas to fit new terminal size
		controlBarHeight := 2
		canvasHeight := m.height - controlBarHeight
		if canvasHeight > 0 && (canvasHeight != m.canvas.height || m.width != m.canvas.width) {
			// Create new canvas with new dimensions, copying old content
			newCanvas := NewCanvas(m.width, canvasHeight)
			// Copy existing content
			for row := 0; row < min(m.canvas.height, canvasHeight); row++ {
				for col := 0; col < min(m.canvas.width, m.width); col++ {
					cell := m.canvas.Get(row, col)
					if cell != nil {
						newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
					}
				}
			}
			m.canvas = newCanvas

			// Save initial state to history if this is the first resize
			if len(m.history) == 0 {
				m.history = []Canvas{m.copyCanvas()}
				m.historyIndex = 0
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Track Option/Alt key state
		// When Alt is pressed with any key, set the flag
		// It will be cleared on mouse release
		if msg.Alt {
			m.optionKeyHeld = true
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "c":
			m.canvas.Clear()
			m.saveToHistory()
			return m, nil
		case "u":
			m.undo()
			return m, nil
		case "r":
			m.redo()
			return m, nil
		case "y":
			// Yank (copy) selection
			m.copySelection()
			return m, nil
		case "d":
			// Delete (cut) selection
			m.cutSelection()
			return m, nil
		case "p":
			// Paste at cursor position
			m.paste()
			return m, nil
		case "s":
			m.showCharPicker = !m.showCharPicker
			m.showFgPicker = false
			m.showBgPicker = false
			m.showToolPicker = false
			if m.showCharPicker {
				m.selectedCategory = m.findSelectedCharCategory()
				m.showingShapes = true
			} else {
				m.showingShapes = false
			}
			return m, nil
		case "f":
			m.showFgPicker = !m.showFgPicker
			m.showCharPicker = false
			m.showBgPicker = false
			m.showToolPicker = false
			return m, nil
		case "b":
			m.showBgPicker = !m.showBgPicker
			m.showCharPicker = false
			m.showFgPicker = false
			m.showToolPicker = false
			return m, nil
		case "t":
			m.showToolPicker = !m.showToolPicker
			m.showCharPicker = false
			m.showFgPicker = false
			m.showBgPicker = false
			return m, nil
		case "esc":
			if m.showCharPicker && m.showingShapes {
				m.showingShapes = false
				return m, nil
			}
			m.showCharPicker = false
			m.showFgPicker = false
			m.showBgPicker = false
			m.showToolPicker = false
			m.hasSelection = false // Clear selection
			return m, nil
		case "left":
			if m.showCharPicker && m.showingShapes {
				m.showingShapes = false
				return m, nil
			}
		case "right":
			if m.showCharPicker && !m.showingShapes {
				m.showingShapes = true
				return m, nil
			}
		case "enter":
			// Check for Ellipse toggle first, before dismissing pickers
			if m.selectedTool == "Ellipse" {
				// Toggle between ellipse and circle mode
				m.circleMode = !m.circleMode
				return m, nil
			} else if m.showCharPicker {
				m.showCharPicker = false
				m.showingShapes = false
				return m, nil
			} else if m.showFgPicker {
				m.showFgPicker = false
				return m, nil
			} else if m.showBgPicker {
				m.showBgPicker = false
				return m, nil
			} else if m.showToolPicker {
				m.showToolPicker = false
				return m, nil
			}
		case "up":
			if m.showCharPicker {
				if m.showingShapes {
					// Navigate within shapes of current category
					idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
					if idx > 0 {
						m.selectedChar = characterGroups[m.selectedCategory].chars[idx-1]
					}
				} else {
					// Navigate categories
					if m.selectedCategory > 0 {
						m.selectedCategory--
					}
				}
				return m, nil
			} else if m.showFgPicker {
				idx := m.findSelectedColorIndex(m.foregroundColor)
				if idx > 0 {
					m.foregroundColor = colors[idx-1].name
				}
				return m, nil
			} else if m.showBgPicker {
				idx := m.findSelectedColorIndex(m.backgroundColor)
				if idx > 0 {
					m.backgroundColor = colors[idx-1].name
				}
				return m, nil
			} else if m.showToolPicker {
				idx := m.findSelectedToolIndex()
				if idx > 0 {
					m.selectedTool = tools[idx-1]
				}
				return m, nil
			}
		case "down":
			if m.showCharPicker {
				if m.showingShapes {
					// Navigate within shapes of current category
					idx := m.findSelectedCharIndexInCategory(m.selectedCategory)
					if idx < len(characterGroups[m.selectedCategory].chars)-1 {
						m.selectedChar = characterGroups[m.selectedCategory].chars[idx+1]
					}
				} else {
					// Navigate categories
					if m.selectedCategory < len(characterGroups)-1 {
						m.selectedCategory++
					}
				}
				return m, nil
			} else if m.showFgPicker {
				idx := m.findSelectedColorIndex(m.foregroundColor)
				if idx < len(colors)-1 {
					m.foregroundColor = colors[idx+1].name
				}
				return m, nil
			} else if m.showBgPicker {
				idx := m.findSelectedColorIndex(m.backgroundColor)
				if idx < len(colors)-1 {
					m.backgroundColor = colors[idx+1].name
				}
				return m, nil
			} else if m.showToolPicker {
				idx := m.findSelectedToolIndex()
				if idx < len(tools)-1 {
					m.selectedTool = tools[idx+1]
				}
				return m, nil
			}
		}

	case tea.MouseMsg:
		m.mouseX = msg.X
		m.mouseY = msg.Y

		controlBarHeight := 2
		canvasHeight := m.height - controlBarHeight

		// Handle popup and menu clicks (only on initial click, not during drag)
		if msg.Type == tea.MouseLeft && !m.mouseDown {
			// Handle popup clicks
			if m.showCharPicker {
				// Calculate category picker bounds
				categoryPickerLeft := m.getShapeX() - 2
				// Find longest category name to determine width
				maxCategoryWidth := 0
				for _, group := range characterGroups {
					nameWidth := len(" ● ") + len(group.name) + len(" ")
					if nameWidth > maxCategoryWidth {
						maxCategoryWidth = nameWidth
					}
				}
				categoryPickerWidth := maxCategoryWidth + 2 // +2 for borders

				pickerHeight := len(characterGroups) + 2
				pickerTop := canvasHeight - pickerHeight

				// Only handle clicks within the category picker bounds
				if msg.Y >= pickerTop && msg.Y < canvasHeight &&
					msg.X >= categoryPickerLeft && msg.X < categoryPickerLeft+categoryPickerWidth {
					row := msg.Y - pickerTop - 1 // -1 for border
					if row >= 0 && row < len(characterGroups) {
						// Check if clicking on a category (not just in the picker area)
						if msg.X >= categoryPickerLeft+1 && msg.X < categoryPickerLeft+categoryPickerWidth-1 {
							m.selectedCategory = row
							return m, nil
						}
					}
				}

				// Handle shapes picker if showing
				if m.showingShapes {
					shapesPickerLeft := categoryPickerLeft + categoryPickerWidth
					shapesPickerWidth := 4 // Tighter bounds for shape picker

					shapesPickerTop := canvasHeight - len(characterGroups[m.selectedCategory].chars) - 2
					if shapesPickerTop < pickerTop {
						shapesPickerTop = pickerTop
					}

					if msg.Y >= shapesPickerTop && msg.Y < canvasHeight &&
						msg.X >= shapesPickerLeft && msg.X < shapesPickerLeft+shapesPickerWidth {
						shapeRow := msg.Y - shapesPickerTop - 1 // -1 for border
						if shapeRow >= 0 && shapeRow < len(characterGroups[m.selectedCategory].chars) {
							m.selectedChar = characterGroups[m.selectedCategory].chars[shapeRow]
							return m, nil
						}
					}
				}
			} else if m.showFgPicker {
				pickerHeight := len(colors) + 2
				pickerTop := canvasHeight - pickerHeight
				pickerLeft := m.getForegroundX() - 2

				// Color picker is about 7 chars wide ( + border + " ● ██")
				if msg.Y >= pickerTop && msg.Y < canvasHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+7 {
					colorIdx := msg.Y - pickerTop - 1
					if colorIdx >= 0 && colorIdx < len(colors) {
						m.foregroundColor = colors[colorIdx].name
						return m, nil
					}
				}
			} else if m.showBgPicker {
				pickerHeight := len(colors) + 2
				pickerTop := canvasHeight - pickerHeight
				pickerLeft := m.getBackgroundX() - 2

				// Color picker is about 7 chars wide
				if msg.Y >= pickerTop && msg.Y < canvasHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+7 {
					colorIdx := msg.Y - pickerTop - 1
					if colorIdx >= 0 && colorIdx < len(colors) {
						m.backgroundColor = colors[colorIdx].name
						return m, nil
					}
				}
			} else if m.showToolPicker {
				pickerHeight := len(tools) + 2
				pickerTop := canvasHeight - pickerHeight
				pickerLeft := m.getToolX() - 2

				// Tool picker width (longest tool name + padding + border)
				if msg.Y >= pickerTop && msg.Y < canvasHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+15 {
					toolIdx := msg.Y - pickerTop - 1
					if toolIdx >= 0 && toolIdx < len(tools) {
						m.selectedTool = tools[toolIdx]
						return m, nil
					}
				}
			}

			// Check if clicking on control bar buttons
			if msg.Y >= canvasHeight {
				// Shapes button (approximately x: 0-15)
				if msg.X < 15 {
					m.showCharPicker = !m.showCharPicker
					m.showFgPicker = false
					m.showBgPicker = false
					m.showToolPicker = false
					return m, nil
				}
				// Foreground button (approximately x: 20-45)
				if msg.X >= 20 && msg.X < 45 {
					m.showFgPicker = !m.showFgPicker
					m.showCharPicker = false
					m.showBgPicker = false
					m.showToolPicker = false
					return m, nil
				}
				// Background button (approximately x: 50-75)
				if msg.X >= 50 && msg.X < 75 {
					m.showBgPicker = !m.showBgPicker
					m.showCharPicker = false
					m.showFgPicker = false
					m.showToolPicker = false
					return m, nil
				}
				// Tool button
				if msg.X >= 58 && msg.X < 75 {
					m.showToolPicker = !m.showToolPicker
					m.showCharPicker = false
					m.showFgPicker = false
					m.showBgPicker = false
					return m, nil
				}
			}
		}

		// Handle mouse press (start of stroke)
		if msg.Type == tea.MouseLeft && !m.mouseDown && msg.Y < canvasHeight {
			m.mouseDown = true
			m.canvasBeforeStroke = m.copyCanvas()
			m.startX = msg.X
			m.startY = msg.Y
			if m.selectedTool == "Rectangle" || m.selectedTool == "Ellipse" || m.selectedTool == "Select" {
				m.showPreview = true
				m.previewEndX = msg.X
				m.previewEndY = msg.Y
				// Clear previous selection when starting a new one
				if m.selectedTool == "Select" {
					m.hasSelection = false
				}
				// Pre-calculate initial preview points for Ellipse
				if m.selectedTool == "Ellipse" {
					m.previewPoints = m.getCirclePoints(m.startY, m.startX, m.previewEndY, m.previewEndX, m.circleMode || m.optionKeyHeld)
				}
			}
		}

		// Handle drag events (MouseLeft with mouseDown=true OR MouseMotion)
		if (msg.Type == tea.MouseLeft || msg.Type == tea.MouseMotion) && m.mouseDown {
			if m.selectedTool == "Rectangle" || m.selectedTool == "Ellipse" || m.selectedTool == "Select" {
				// Clamp coordinates to canvas bounds
				clampedY := msg.Y
				if clampedY < 0 {
					clampedY = 0
				} else if clampedY >= canvasHeight {
					clampedY = canvasHeight - 1
				}
				clampedX := msg.X
				if clampedX < 0 {
					clampedX = 0
				} else if clampedX >= m.canvas.width {
					clampedX = m.canvas.width - 1
				}

				m.previewEndX = clampedX
				m.previewEndY = clampedY
				// Pre-calculate preview points for Ellipse to improve performance
				if m.selectedTool == "Ellipse" {
					m.previewPoints = m.getCirclePoints(m.startY, m.startX, m.previewEndY, m.previewEndX, m.circleMode || m.optionKeyHeld)
				}
			} else if m.selectedTool == "Point" && msg.Y < canvasHeight {
				m.canvas.Set(msg.Y, msg.X, m.selectedChar, m.foregroundColor, m.backgroundColor)
			}
		}

		// Handle mouse release (end of stroke)
		if msg.Type == tea.MouseRelease && m.mouseDown {
			m.mouseDown = false
			m.showPreview = false
			m.previewPoints = nil // Clear cached preview points

			// Clamp coordinates to canvas bounds
			clampedY := msg.Y
			if clampedY < 0 {
				clampedY = 0
			} else if clampedY >= canvasHeight {
				clampedY = canvasHeight - 1
			}
			clampedX := msg.X
			if clampedX < 0 {
				clampedX = 0
			} else if clampedX >= m.canvas.width {
				clampedX = m.canvas.width - 1
			}

			// For shape tools, draw the shape on release
			if m.selectedTool == "Rectangle" {
				m.drawRectangle(m.startY, m.startX, clampedY, clampedX)
			} else if m.selectedTool == "Ellipse" {
				// Use circle mode if toggle is on OR option key is held
				m.drawCircle(m.startY, m.startX, clampedY, clampedX, m.circleMode || m.optionKeyHeld)
			} else if m.selectedTool == "Select" {
				// Finalize selection
				m.hasSelection = true
				m.selectionStartY = m.startY
				m.selectionStartX = m.startX
				m.selectionEndY = clampedY
				m.selectionEndX = clampedX
			}

			// Clear option key state after drawing
			m.optionKeyHeld = false

			// Save to history if canvas changed (but not for selections)
			if m.selectedTool != "Select" && !m.canvasEquals(m.canvasBeforeStroke) {
				m.saveToHistory()
			}
		}
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Calculate canvas height (terminal height - control bar - 1 line margin)
	controlBarHeight := 2
	canvasHeight := m.height - controlBarHeight

	var b strings.Builder

	// Determine popup info
	var popup string
	var popupLines []string
	var popupStartY int
	var popupX int
	var popup2 string
	var popup2Lines []string
	var popup2StartY int
	var popup2X int

	if m.showCharPicker {
		popup = m.renderCategoryPicker()
		popupLines = strings.Split(popup, "\n")
		popupStartY = canvasHeight - len(popupLines)
		// Position above the shape in control bar, with dot aligned
		// Shape is at position 8, dot in popup is at offset 2 (border + space)
		popupX = m.getShapeX() - 2

		if m.showingShapes {
			popup2 = m.renderShapesPicker()
			popup2Lines = strings.Split(popup2, "\n")
			// Bottom-align both popups
			popup2StartY = canvasHeight - len(popup2Lines)
			// Position shapes picker to the right of category picker
			// Calculate width by getting the first line's width (they should all be same)
			categoryWidth := 0
			if len(popupLines) > 0 {
				categoryWidth = lipgloss.Width(popupLines[0])
			}
			popup2X = popupX + categoryWidth
		}
	} else if m.showFgPicker {
		popup = m.renderColorPicker("Foreground")
		popupLines = strings.Split(popup, "\n")
		popupStartY = canvasHeight - len(popupLines)
		popupX = m.getForegroundX() - 2
	} else if m.showBgPicker {
		popup = m.renderColorPicker("Background")
		popupLines = strings.Split(popup, "\n")
		popupStartY = canvasHeight - len(popupLines)
		popupX = m.getBackgroundX() - 2
	} else if m.showToolPicker {
		popup = m.renderToolPicker()
		popupLines = strings.Split(popup, "\n")
		popupStartY = canvasHeight - len(popupLines)
		popupX = m.getToolX() - 2
	}

	// Render canvas with popup overlay
	for i := 0; i < canvasHeight; i++ {
		var lineBuilder strings.Builder

		// Render each column
		for col := 0; col < m.width; col++ {
			// Check if this position is covered by a popup
			inPopup := false

			// Check first popup
			if popupLines != nil && i >= popupStartY && i < popupStartY+len(popupLines) {
				popupLineIdx := i - popupStartY
				popupLine := popupLines[popupLineIdx]
				popupWidth := lipgloss.Width(popupLine)

				if col >= popupX && col < popupX+popupWidth {
					// This column is covered by popup - render popup content
					if col == popupX {
						lineBuilder.WriteString("\x1b[0m" + popupLine)
					}
					inPopup = true
				}
			}

			// Check second popup
			if !inPopup && popup2Lines != nil && i >= popup2StartY && i < popup2StartY+len(popup2Lines) {
				popup2LineIdx := i - popup2StartY
				popup2Line := popup2Lines[popup2LineIdx]
				popup2Width := lipgloss.Width(popup2Line)

				if col >= popup2X && col < popup2X+popup2Width {
					// This column is covered by popup2 - render popup content
					if col == popup2X {
						lineBuilder.WriteString("\x1b[0m" + popup2Line)
					}
					inPopup = true
				}
			}

			// If not in popup, render canvas cell or preview
			if !inPopup {
				// Check if this position is part of the preview shape
				inPreview := false
				if m.showPreview && m.selectedTool == "Rectangle" {
					minY, maxY := m.startY, m.previewEndY
					if m.startY > m.previewEndY {
						minY, maxY = m.previewEndY, m.startY
					}
					minX, maxX := m.startX, m.previewEndX
					if m.startX > m.previewEndX {
						minX, maxX = m.previewEndX, m.startX
					}

					if i >= minY && i <= maxY && col >= minX && col <= maxX {
						if i == minY || i == maxY || col == minX || col == maxX {
							// Draw preview with current character and colors
							style := lipgloss.NewStyle()
							for _, c := range colors {
								if c.name == m.foregroundColor {
									style = c.style
									break
								}
							}
							if m.backgroundColor != "transparent" {
								for _, c := range colors {
									if c.name == m.backgroundColor {
										style = style.Background(c.style.GetForeground())
										break
									}
								}
							}
							lineBuilder.WriteString(style.Render(m.selectedChar))
							inPreview = true
						}
					}
				} else if m.showPreview && m.selectedTool == "Ellipse" {
					// Use pre-calculated points for performance
					if m.previewPoints[[2]int{i, col}] {
						style := lipgloss.NewStyle()
						for _, c := range colors {
							if c.name == m.foregroundColor {
								style = c.style
								break
							}
						}
						if m.backgroundColor != "transparent" {
							for _, c := range colors {
								if c.name == m.backgroundColor {
									style = style.Background(c.style.GetForeground())
									break
								}
							}
						}
						lineBuilder.WriteString(style.Render(m.selectedChar))
						inPreview = true
					}
				} else if m.showPreview && m.selectedTool == "Select" {
					// Show selection box preview while dragging
					minY, maxY := m.startY, m.previewEndY
					if m.startY > m.previewEndY {
						minY, maxY = m.previewEndY, m.startY
					}
					minX, maxX := m.startX, m.previewEndX
					if m.startX > m.previewEndX {
						minX, maxX = m.previewEndX, m.startX
					}

					if i >= minY && i <= maxY && col >= minX && col <= maxX {
						if i == minY || i == maxY || col == minX || col == maxX {
							// Draw selection preview with dashed border
							dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
							var char string
							// Single pixel selection
							if minY == maxY && minX == maxX {
								char = "□"
							} else if i == minY && col == minX {
								// Top-left corner
								char = "┌"
							} else if i == minY && col == maxX {
								// Top-right corner
								char = "┐"
							} else if i == maxY && col == minX {
								// Bottom-left corner
								char = "└"
							} else if i == maxY && col == maxX {
								// Bottom-right corner
								char = "┘"
							} else if i == minY || i == maxY {
								// Top/bottom edges
								char = "┈"
							} else {
								// Left/right edges
								char = "┊"
							}
							lineBuilder.WriteString(dimStyle.Render(char))
							inPreview = true
						}
					}
				}

				// Check for active selection border (persistent after release)
				inSelection := false
				if !inPreview && m.hasSelection {
					minY, maxY := m.selectionStartY, m.selectionEndY
					if m.selectionStartY > m.selectionEndY {
						minY, maxY = m.selectionEndY, m.selectionStartY
					}
					minX, maxX := m.selectionStartX, m.selectionEndX
					if m.selectionStartX > m.selectionEndX {
						minX, maxX = m.selectionEndX, m.selectionStartX
					}

					if i >= minY && i <= maxY && col >= minX && col <= maxX {
						if i == minY || i == maxY || col == minX || col == maxX {
							// Draw persistent selection border
							highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
							var char string
							// Single pixel selection
							if minY == maxY && minX == maxX {
								char = "□"
							} else if i == minY && col == minX {
								// Top-left corner
								char = "┌"
							} else if i == minY && col == maxX {
								// Top-right corner
								char = "┐"
							} else if i == maxY && col == minX {
								// Bottom-left corner
								char = "└"
							} else if i == maxY && col == maxX {
								// Bottom-right corner
								char = "┘"
							} else if i == minY || i == maxY {
								// Top/bottom edges
								char = "┈"
							} else {
								// Left/right edges
								char = "┊"
							}
							lineBuilder.WriteString(highlightStyle.Render(char))
							inSelection = true
						}
					}
				}

				// Render canvas cell if not in preview or selection border
				if !inPreview && !inSelection {
					cell := m.canvas.Get(i, col)
					if cell != nil {
						if cell.foregroundColor == "transparent" {
							lineBuilder.WriteString(" ")
						} else {
							style := lipgloss.NewStyle()
							for _, c := range colors {
								if c.name == cell.foregroundColor {
									style = c.style
									break
								}
							}
							if cell.backgroundColor != "transparent" {
								for _, c := range colors {
									if c.name == cell.backgroundColor {
										style = style.Background(c.style.GetForeground())
										break
									}
								}
							}
							lineBuilder.WriteString(style.Render(cell.char))
						}
					}
				}
			}
		}

		b.WriteString(lineBuilder.String() + "\n")
	}

	// Render control bar
	b.WriteString(m.renderControlBar())

	return b.String()
}

func (m model) renderCanvas() string {
	var b strings.Builder

	for row := 0; row < m.canvas.height; row++ {
		for col := 0; col < m.canvas.width; col++ {
			cell := m.canvas.Get(row, col)
			if cell != nil {
				// Handle transparent foreground
				if cell.foregroundColor == "transparent" {
					b.WriteString(" ")
				} else {
					// Build style with foreground and background colors
					style := lipgloss.NewStyle()

					// Apply foreground color
					for _, c := range colors {
						if c.name == cell.foregroundColor {
							style = c.style
							break
						}
					}

					// Apply background color if not transparent
					if cell.backgroundColor != "transparent" {
						for _, c := range colors {
							if c.name == cell.backgroundColor {
								style = style.Background(c.style.GetForeground())
								break
							}
						}
					}

					b.WriteString(style.Render(cell.char))
				}
			}
		}
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) renderControlBar() string {
	underlineStyle := lipgloss.NewStyle().Underline(true)

	// Shape button - underline 'S' for hotkey
	shapeButton := underlineStyle.Render("S") + fmt.Sprintf("hapes: %s", m.selectedChar)

	// Foreground color button
	var fgSwatch string
	if m.foregroundColor == "transparent" {
		fgSwatch = "  "
	} else {
		var fgStyle lipgloss.Style
		for _, c := range colors {
			if c.name == m.foregroundColor {
				fgStyle = c.style
				break
			}
		}
		fgSwatch = fgStyle.Render("██")
	}
	fgButton := underlineStyle.Render("F") + fmt.Sprintf("oreground: %s", fgSwatch)

	// Background color button
	var bgSwatch string
	if m.backgroundColor == "transparent" {
		bgSwatch = "  "
	} else {
		var bgStyle lipgloss.Style
		for _, c := range colors {
			if c.name == m.backgroundColor {
				bgStyle = c.style
				break
			}
		}
		bgSwatch = bgStyle.Render("██")
	}
	bgButton := underlineStyle.Render("B") + fmt.Sprintf("ackground: %s", bgSwatch)

	// Tool button - show "Circle" when in circle mode
	toolName := m.selectedTool
	if m.selectedTool == "Ellipse" && m.circleMode {
		toolName = "Circle"
	}
	toolButton := underlineStyle.Render("T") + fmt.Sprintf("ool: %s", toolName)

	// Mode indicator - show if clipboard has content
	modeIndicator := ""
	if m.clipboard != nil && m.clipboardHeight > 0 && m.clipboardWidth > 0 {
		modeIndicator = fmt.Sprintf("  |  Mode: Yank (%dx%d)", m.clipboardWidth, m.clipboardHeight)
	}

	controlBar := fmt.Sprintf(" %s  |  %s  |  %s  |  %s%s", shapeButton, fgButton, bgButton, toolButton, modeIndicator)

	return controlBar
}

func (m model) getShapeX() int {
	return 9 // " Shapes: " is 9 chars (with leading space)
}

func (m model) getForegroundX() int {
	// " Shapes: " (9) + "●" (2 cells) + "  |  " (5) + "Foreground: " (12) = 28
	return 28
}

func (m model) getBackgroundX() int {
	// 28 + "██" (2 cells) + "  |  " (5) + "Background: " (12) = 47
	return 47
}

func (m model) getToolX() int {
	// 47 + "██" (2 cells) + "  |  " (5) + "Tool: " (6) = 60, -1 for left alignment
	return 59
}

func (m model) findSelectedToolIndex() int {
	for i, tool := range tools {
		if tool == m.selectedTool {
			return i
		}
	}
	return 0
}

func (m model) findSelectedCharCategory() int {
	for i, group := range characterGroups {
		for _, char := range group.chars {
			if char == m.selectedChar {
				return i
			}
		}
	}
	return 0
}

func (m model) findSelectedCharIndexInCategory(categoryIdx int) int {
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

func (m model) findSelectedColorIndex(colorName string) int {
	for i, color := range colors {
		if color.name == colorName {
			return i
		}
	}
	return 0
}

func (m *model) saveToHistory() {
	// Initialize history if empty
	if len(m.history) == 0 {
		m.history = []Canvas{m.copyCanvas()}
		m.historyIndex = 0
		return
	}

	// Remove any redo history when making a new change
	if m.historyIndex < len(m.history)-1 {
		m.history = m.history[:m.historyIndex+1]
	}

	// Add current canvas to history
	m.history = append(m.history, m.copyCanvas())
	m.historyIndex++

	// Limit history size to 50 states
	if len(m.history) > 50 {
		m.history = m.history[1:]
		m.historyIndex--
	}
}

func (m *model) undo() {
	// Can only undo if we're not at the first state
	if m.historyIndex > 0 {
		m.historyIndex--
		m.canvas = m.copyFromCanvas(m.history[m.historyIndex])
	}
}

func (m *model) redo() {
	// Can only redo if there are states ahead
	if m.historyIndex < len(m.history)-1 {
		m.historyIndex++
		m.canvas = m.copyFromCanvas(m.history[m.historyIndex])
	}
}

func (m *model) copySelection() {
	if !m.hasSelection {
		return
	}

	// Normalize selection bounds
	minY, maxY := m.selectionStartY, m.selectionEndY
	if m.selectionStartY > m.selectionEndY {
		minY, maxY = m.selectionEndY, m.selectionStartY
	}
	minX, maxX := m.selectionStartX, m.selectionEndX
	if m.selectionStartX > m.selectionEndX {
		minX, maxX = m.selectionEndX, m.selectionStartX
	}

	// Copy the selected region
	m.clipboardHeight = maxY - minY + 1
	m.clipboardWidth = maxX - minX + 1
	m.clipboard = make([][]Cell, m.clipboardHeight)

	for y := 0; y < m.clipboardHeight; y++ {
		m.clipboard[y] = make([]Cell, m.clipboardWidth)
		for x := 0; x < m.clipboardWidth; x++ {
			cell := m.canvas.Get(minY+y, minX+x)
			if cell != nil {
				m.clipboard[y][x] = *cell
			} else {
				m.clipboard[y][x] = Cell{char: " ", foregroundColor: "transparent", backgroundColor: "transparent"}
			}
		}
	}
}

func (m *model) cutSelection() {
	if !m.hasSelection {
		return
	}

	// First copy
	m.copySelection()

	// Then clear the selected region
	minY, maxY := m.selectionStartY, m.selectionEndY
	if m.selectionStartY > m.selectionEndY {
		minY, maxY = m.selectionEndY, m.selectionStartY
	}
	minX, maxX := m.selectionStartX, m.selectionEndX
	if m.selectionStartX > m.selectionEndX {
		minX, maxX = m.selectionEndX, m.selectionStartX
	}

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			m.canvas.Set(y, x, " ", "transparent", "transparent")
		}
	}

	// Save to history
	m.saveToHistory()
}

func (m *model) paste() {
	if m.clipboard == nil || m.clipboardHeight == 0 || m.clipboardWidth == 0 {
		return
	}

	// Calculate canvas height to ensure paste is on canvas
	controlBarHeight := 2
	canvasHeight := m.height - controlBarHeight

	// Only paste if mouse is on canvas
	if m.mouseY >= canvasHeight {
		return
	}

	// Paste with top-left corner at current mouse position
	for y := 0; y < m.clipboardHeight; y++ {
		for x := 0; x < m.clipboardWidth; x++ {
			targetY := m.mouseY + y
			targetX := m.mouseX + x
			// Only paste if within canvas bounds
			if targetY >= 0 && targetY < canvasHeight && targetX >= 0 && targetX < m.canvas.width {
				cell := m.clipboard[y][x]
				existingCell := m.canvas.Get(targetY, targetX)

				// Skip fully transparent cells (don't overwrite destination)
				if cell.foregroundColor == "transparent" && cell.backgroundColor == "transparent" {
					continue
				}

				// Handle partial transparency by preserving destination values
				newChar := cell.char
				newFg := cell.foregroundColor
				newBg := cell.backgroundColor

				// If foreground is transparent, preserve destination's character and foreground
				if cell.foregroundColor == "transparent" && existingCell != nil {
					newChar = existingCell.char
					newFg = existingCell.foregroundColor
				}

				// If background is transparent, preserve destination's background
				if cell.backgroundColor == "transparent" && existingCell != nil {
					newBg = existingCell.backgroundColor
				}

				m.canvas.Set(targetY, targetX, newChar, newFg, newBg)
			}
		}
	}

	// Save to history and clear selection
	m.saveToHistory()
	m.hasSelection = false
}

func (m model) copyCanvas() Canvas {
	newCanvas := NewCanvas(m.canvas.width, m.canvas.height)
	for row := 0; row < m.canvas.height; row++ {
		for col := 0; col < m.canvas.width; col++ {
			cell := m.canvas.Get(row, col)
			if cell != nil {
				newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
			}
		}
	}
	return newCanvas
}

func (m model) copyFromCanvas(source Canvas) Canvas {
	newCanvas := NewCanvas(source.width, source.height)
	for row := 0; row < source.height; row++ {
		for col := 0; col < source.width; col++ {
			cell := source.Get(row, col)
			if cell != nil {
				newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
			}
		}
	}
	return newCanvas
}

func (m *model) drawRectangle(y1, x1, y2, x2 int) {
	// Normalize coordinates
	minY, maxY := y1, y2
	if y1 > y2 {
		minY, maxY = y2, y1
	}
	minX, maxX := x1, x2
	if x1 > x2 {
		minX, maxX = x2, x1
	}

	// Draw rectangle outline
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			// Draw only on the border
			if y == minY || y == maxY || x == minX || x == maxX {
				m.canvas.Set(y, x, m.selectedChar, m.foregroundColor, m.backgroundColor)
			}
		}
	}
}

func (m *model) getCirclePoints(y1, x1, y2, x2 int, forceCircle bool) map[[2]int]bool {
	points := make(map[[2]int]bool)

	// Calculate bounding box
	minY, maxY := y1, y2
	if y1 > y2 {
		minY, maxY = y2, y1
	}
	minX, maxX := x1, x2
	if x1 > x2 {
		minX, maxX = x2, x1
	}

	centerY := (minY + maxY) / 2
	centerX := (minX + maxX) / 2

	if forceCircle {
		// Perfect circle based on distance from start to end
		dy := y2 - y1
		dx := x2 - x1
		radius := int(0.5 + sqrt(float64(dy*dy+dx*dx)))

		if radius == 0 {
			points[[2]int{y1, x1}] = true
			return points
		}

		rx := radius
		ry := radius / 2
		if ry == 0 {
			ry = 1
		}

		return getEllipsePoints(y1, x1, ry, rx)
	} else {
		// Ellipse based on bounding box
		// Use full bounding box - user is explicitly defining the shape
		rx := (maxX - minX) / 2
		ry := (maxY - minY) / 2
		if ry == 0 {
			ry = 1
		}

		if rx == 0 && ry == 0 {
			points[[2]int{centerY, centerX}] = true
			return points
		}

		if rx == 0 {
			for y := minY; y <= maxY; y++ {
				points[[2]int{y, centerX}] = true
			}
			return points
		}

		return getEllipsePoints(centerY, centerX, ry, rx)
	}
}

func (m *model) drawCircle(y1, x1, y2, x2 int, forceCircle bool) {
	points := m.getCirclePoints(y1, x1, y2, x2, forceCircle)
	for point := range points {
		m.canvas.Set(point[0], point[1], m.selectedChar, m.foregroundColor, m.backgroundColor)
	}
}

func getEllipsePoints(centerY, centerX, ry, rx int) map[[2]int]bool {
	points := make(map[[2]int]bool)

	// Region 1
	x := 0
	y := ry
	rx2 := rx * rx
	ry2 := ry * ry
	twoRx2 := 2 * rx2
	twoRy2 := 2 * ry2
	px := 0
	py := twoRx2 * y

	addEllipsePoints(points, centerY, centerX, x, y)

	// Region 1
	p := int(float64(ry2) - float64(rx2*ry) + 0.25*float64(rx2))
	for px < py {
		x++
		px += twoRy2
		if p < 0 {
			p += ry2 + px
		} else {
			y--
			py -= twoRx2
			p += ry2 + px - py
		}
		addEllipsePoints(points, centerY, centerX, x, y)
	}

	// Region 2
	p = int(float64(ry2*(x+1)*(x+1)) + float64(rx2*(y-1)*(y-1)) - float64(rx2*ry2))
	for y > 0 {
		y--
		py -= twoRx2
		if p > 0 {
			p += rx2 - py
		} else {
			x++
			px += twoRy2
			p += rx2 - py + px
		}
		addEllipsePoints(points, centerY, centerX, x, y)
	}

	return points
}

func addEllipsePoints(points map[[2]int]bool, centerY, centerX, x, y int) {
	// Add 4 symmetric points
	points[[2]int{centerY + y, centerX + x}] = true
	points[[2]int{centerY + y, centerX - x}] = true
	points[[2]int{centerY - y, centerX + x}] = true
	points[[2]int{centerY - y, centerX - x}] = true
}

func sqrt(x float64) float64 {
	if x == 0 {
		return 0
	}
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func (m model) canvasEquals(other Canvas) bool {
	if m.canvas.width != other.width || m.canvas.height != other.height {
		return false
	}
	for row := 0; row < m.canvas.height; row++ {
		for col := 0; col < m.canvas.width; col++ {
			cell1 := m.canvas.Get(row, col)
			cell2 := other.Get(row, col)
			if cell1 == nil || cell2 == nil {
				return false
			}
			if cell1.char != cell2.char || cell1.foregroundColor != cell2.foregroundColor || cell1.backgroundColor != cell2.backgroundColor {
				return false
			}
		}
	}
	return true
}

func (m model) renderCategoryPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	var content strings.Builder
	for i, group := range characterGroups {
		// Add dot for selected category
		prefix := "   "
		if i == m.selectedCategory {
			prefix = " ● "
		}

		content.WriteString(prefix + group.name + " ")
		if i < len(characterGroups)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m model) renderShapesPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	var content strings.Builder
	if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
		group := characterGroups[m.selectedCategory]
		for i, char := range group.chars {
			// Add dot for selected character, shape aligns with dot column
			var line string
			if char == m.selectedChar {
				line = " ● " + char
			} else {
				line = "   " + char
			}

			content.WriteString(line)
			if i < len(group.chars)-1 {
				content.WriteString("\n")
			}
		}
	}

	return pickerStyle.Render(content.String())
}

func (m model) renderToolPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	var content strings.Builder
	for i, tool := range tools {
		// Add dot for selected tool
		prefix := "   "
		if tool == m.selectedTool {
			prefix = " ● "
		}

		// Show "Circle" instead of "Ellipse" when in circle mode
		displayName := tool
		if tool == "Ellipse" && m.circleMode {
			displayName = "Circle"
		}

		content.WriteString(prefix + displayName)
		if i < len(tools)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m model) renderColorPicker(title string) string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	currentColor := m.foregroundColor
	if title == "Background" {
		currentColor = m.backgroundColor
	}

	var content strings.Builder
	for i, color := range colors {
		// Add dot for selected color
		prefix := "   "
		if color.name == currentColor {
			prefix = " ● "
		}

		var swatch string
		if color.name == "transparent" {
			swatch = "  "
		} else {
			swatch = color.style.Render("██")
		}
		line := prefix + swatch

		content.WriteString(line)
		// Don't add newline after last color
		if i < len(colors)-1 {
			content.WriteString("\n")
		}
	}

	return pickerStyle.Render(content.String())
}

func (m model) overlayPopupAt(base, popup string, topOffset, leftOffset int) string {
	baseLines := strings.Split(base, "\n")
	popupLines := strings.Split(popup, "\n")

	for i, popupLine := range popupLines {
		lineIdx := topOffset + i
		if lineIdx >= 0 && lineIdx < len(baseLines) {
			baseLine := baseLines[lineIdx]

			// Replace part of base line with popup line
			before := ""
			if leftOffset < len(baseLine) {
				before = baseLine[:leftOffset]
			}

			after := ""
			popupWidth := lipgloss.Width(popupLine)
			afterStart := leftOffset + popupWidth
			if afterStart < len(baseLine) {
				after = baseLine[afterStart:]
			}

			baseLines[lineIdx] = before + popupLine + after
		}
	}

	return strings.Join(baseLines, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// truncateToVisualWidth truncates or pads a string to a specific visual width,
// properly handling ANSI escape sequences
func truncateToVisualWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}

	var result strings.Builder
	visualWidth := 0
	inEscape := false

	for _, r := range s {
		// Handle ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
		}

		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		// Count visual width
		if visualWidth >= width {
			break
		}

		result.WriteRune(r)
		visualWidth++
	}

	// Pad with spaces if needed
	for visualWidth < width {
		result.WriteRune(' ')
		visualWidth++
	}

	return result.String()
}

// skipVisualWidth skips the first N visual characters and returns the rest,
// properly handling ANSI escape sequences
func skipVisualWidth(s string, skip int) string {
	if skip <= 0 {
		return s
	}

	var result strings.Builder
	visualWidth := 0
	inEscape := false
	skipping := true

	for _, r := range s {
		// Handle ANSI escape sequences
		if r == '\x1b' {
			inEscape = true
		}

		if inEscape {
			if !skipping {
				result.WriteRune(r)
			}
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		// Count visual width
		if skipping {
			visualWidth++
			if visualWidth > skip {
				skipping = false
				result.WriteRune(r)
			}
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
