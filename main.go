package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Layout constants
const (
	// Control bar
	controlBarHeight = 1

	// Picker layout
	pickerBorderWidth = 2 // Top and bottom borders (or left and right)

	// Picker item structure
	pickerItemPadding   = 1 // Leading space before content
	pickerSwatchWidth   = 2 // "██" for colors
	pickerItemSeparator = 1 // Space between swatch and name

	// Picker content offset (border + padding = where content starts)
	pickerContentOffset = 1 + pickerItemPadding // 2: position of swatch/text in picker

	// Toolbar button layout
	toolbarButtonPadding = 1 // Left/right padding inside each button
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
	selectedCategory   int  // Index of selected character category
	showingShapes      bool // Whether we're showing shapes (second level) or categories (first level)
	shapesFocusOnPanel bool // True if focus is on shapes panel, false if on categories
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
	lastMenu           int              // Most recently opened menu index (0=shape, 1=fg, 2=bg, 3=tool)
	// Toolbar button positions (calculated during render)
	toolbarShapeX      int
	toolbarForegroundX int
	toolbarBackgroundX int
	toolbarToolX       int
	// Toolbar selected item positions (for popup alignment)
	toolbarShapeItemX      int
	toolbarForegroundItemX int
	toolbarBackgroundItemX int
	toolbarToolItemX       int
}

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

func initialModel() *model {
	canvas := NewCanvas(100, 30)
	return &model{
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

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Resize canvas to fit new terminal size
		// Use constant instead of local variable
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
		// Handle picker hotkeys (1-9) when pickers are open
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1') // Convert '1'-'9' to 0-8

			if m.showCharPicker && !m.shapesFocusOnPanel {
				// Category picker hotkey
				if idx < len(characterGroups) {
					m.selectedCategory = idx
					return m, nil
				}
			} else if m.showCharPicker && m.shapesFocusOnPanel {
				// Shapes picker hotkey (only 1-9 supported)
				if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
					chars := characterGroups[m.selectedCategory].chars
					if idx < len(chars) {
						m.selectedChar = chars[idx]
						return m, nil
					}
				}
			} else if m.showFgPicker {
				// Foreground color picker hotkey (only 1-9 supported)
				if idx < len(colors) {
					m.foregroundColor = colors[idx].name
					return m, nil
				}
			} else if m.showBgPicker {
				// Background color picker hotkey (only 1-9 supported)
				if idx < len(colors) {
					m.backgroundColor = colors[idx].name
					return m, nil
				}
			} else if m.showToolPicker {
				// Tool picker hotkey
				if idx < len(tools) {
					m.selectedTool = tools[idx]
					return m, nil
				}
			}
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
			if m.showCharPicker {
				m.closeMenus()
			} else {
				m.openMenu(0)
			}
			return m, nil
		case "f":
			if m.showFgPicker {
				m.closeMenus()
			} else {
				m.openMenu(1)
			}
			return m, nil
		case "b":
			if m.showBgPicker {
				m.closeMenus()
			} else {
				m.openMenu(2)
			}
			return m, nil
		case "t":
			if m.showToolPicker {
				m.closeMenus()
			} else {
				m.openMenu(3)
			}
			return m, nil
		case "[":
			active := m.activeMenu()
			if active < 0 {
				m.openMenu(m.lastMenu)
			} else {
				m.openMenu((active - 1 + menuCount) % menuCount)
			}
			return m, nil
		case "]":
			active := m.activeMenu()
			if active < 0 {
				m.openMenu(m.lastMenu)
			} else {
				m.openMenu((active + 1) % menuCount)
			}
			return m, nil
		case "esc":
			if m.showCharPicker && m.shapesFocusOnPanel {
				// First esc moves focus back to categories
				m.shapesFocusOnPanel = false
				return m, nil
			}
			m.showCharPicker = false
			m.showFgPicker = false
			m.showBgPicker = false
			m.showToolPicker = false
			m.hasSelection = false // Clear selection
			m.shapesFocusOnPanel = false
			return m, nil
		case "left":
			if m.showCharPicker && m.shapesFocusOnPanel {
				// Move focus back to categories
				m.shapesFocusOnPanel = false
				return m, nil
			}
		case "right":
			if m.showCharPicker && !m.shapesFocusOnPanel {
				// Move focus to shapes panel
				m.shapesFocusOnPanel = true
				// If no shape is selected in this category, select the first one
				if m.selectedCategory >= 0 && m.selectedCategory < len(characterGroups) {
					currentIdx := m.findSelectedCharIndexInCategory(m.selectedCategory)
					// If selected char is not in current category, select first
					if currentIdx == 0 && m.selectedChar != characterGroups[m.selectedCategory].chars[0] {
						m.selectedChar = characterGroups[m.selectedCategory].chars[0]
					}
				}
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
				if m.shapesFocusOnPanel {
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
				if m.shapesFocusOnPanel {
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

		// Use constant instead of local variable
		canvasHeight := m.canvas.height

		// Handle popup and menu clicks (only on initial click, not during drag)
		if msg.Type == tea.MouseLeft && !m.mouseDown {
			// Handle popup clicks
			if m.showCharPicker {
				// Calculate category picker bounds - must match View() positioning
				categoryPickerLeft := m.toolbarShapeItemX - pickerContentOffset
				// Find longest category name to determine width
				maxCategoryWidth := 0
				for _, group := range characterGroups {
					nameWidth := len(group.name) + 2 // " name "
					if nameWidth > maxCategoryWidth {
						maxCategoryWidth = nameWidth
					}
				}
				categoryPickerWidth := maxCategoryWidth + pickerBorderWidth

				pickerHeight := len(characterGroups) + pickerBorderWidth
				pickerTop := controlBarHeight

				// Only handle clicks within the category picker bounds
				if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
					msg.X >= categoryPickerLeft && msg.X < categoryPickerLeft+categoryPickerWidth {
					row := msg.Y - pickerTop - 1 // -1 for border
					if row >= 0 && row < len(characterGroups) {
						// Check if clicking on a category (not just in the picker area)
						if msg.X >= categoryPickerLeft+1 && msg.X < categoryPickerLeft+categoryPickerWidth-1 {
							m.selectedCategory = row
							m.shapesFocusOnPanel = false // Clicking category means focus is on categories
							return m, nil
						}
					}
				}

				// Handle shapes picker (always visible when category picker is open)
				shapesPickerLeft := categoryPickerLeft + categoryPickerWidth
				// Content is " char " = 3 visual chars + border
				shapesPickerWidth := 3 + pickerBorderWidth

				shapesPickerHeight := len(characterGroups[m.selectedCategory].chars) + pickerBorderWidth
				// Match View() positioning: align first row with selected category, clamped to screen
				shapesCanvasY := m.selectedCategory
				if shapesCanvasY+shapesPickerHeight > m.canvas.height {
					shapesCanvasY = m.canvas.height - shapesPickerHeight
				}
				if shapesCanvasY < 0 {
					shapesCanvasY = 0
				}
				shapesPickerTop := controlBarHeight + shapesCanvasY

				if msg.Y >= shapesPickerTop && msg.Y < shapesPickerTop+shapesPickerHeight &&
					msg.X >= shapesPickerLeft && msg.X < shapesPickerLeft+shapesPickerWidth {
					shapeRow := msg.Y - shapesPickerTop - 1 // -1 for border
					if shapeRow >= 0 && shapeRow < len(characterGroups[m.selectedCategory].chars) {
						m.selectedChar = characterGroups[m.selectedCategory].chars[shapeRow]
						m.shapesFocusOnPanel = true // Clicking shape means focus is on shapes
						return m, nil
					}
				}
			} else if m.showFgPicker {
				pickerHeight := len(colors) + pickerBorderWidth
				pickerTop := controlBarHeight
				pickerLeft := m.toolbarForegroundItemX - pickerContentOffset

				// Calculate actual picker width: border + " ● ██ " + longest color name + border
				// Find longest color name
				maxNameLen := 0
				for _, c := range colors {
					displayName := strings.ReplaceAll(c.name, "_", " ")
					if len(displayName) > 0 {
						displayName = strings.ToUpper(displayName[:1]) + displayName[1:]
					}
					if len(displayName) > maxNameLen {
						maxNameLen = len(displayName)
					}
				}
				pickerWidth := pickerBorderWidth + pickerItemPadding + pickerSwatchWidth + pickerItemSeparator + maxNameLen + pickerBorderWidth

				if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
					colorIdx := msg.Y - pickerTop - 1
					if colorIdx >= 0 && colorIdx < len(colors) {
						m.foregroundColor = colors[colorIdx].name
						return m, nil
					}
				}
			} else if m.showBgPicker {
				pickerHeight := len(colors) + pickerBorderWidth
				pickerTop := controlBarHeight
				pickerLeft := m.toolbarBackgroundItemX - pickerContentOffset

				// Calculate actual picker width
				maxNameLen := 0
				for _, c := range colors {
					displayName := strings.ReplaceAll(c.name, "_", " ")
					if len(displayName) > 0 {
						displayName = strings.ToUpper(displayName[:1]) + displayName[1:]
					}
					if len(displayName) > maxNameLen {
						maxNameLen = len(displayName)
					}
				}
				pickerWidth := pickerBorderWidth + pickerItemPadding + pickerSwatchWidth + pickerItemSeparator + maxNameLen + pickerBorderWidth

				if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
					colorIdx := msg.Y - pickerTop - 1
					if colorIdx >= 0 && colorIdx < len(colors) {
						m.backgroundColor = colors[colorIdx].name
						return m, nil
					}
				}
			} else if m.showToolPicker {
				pickerHeight := len(tools) + pickerBorderWidth
				pickerTop := controlBarHeight
				pickerLeft := m.toolbarToolItemX - pickerContentOffset

				maxToolLen := 0
				for _, t := range tools {
					name := t
					if t == "Ellipse" && m.circleMode {
						name = "Circle"
					}
					if len(name) > maxToolLen {
						maxToolLen = len(name)
					}
				}
				pickerWidth := pickerBorderWidth + pickerItemPadding + maxToolLen + pickerItemPadding

				if msg.Y >= pickerTop && msg.Y < pickerTop+pickerHeight &&
					msg.X >= pickerLeft && msg.X < pickerLeft+pickerWidth {
					toolIdx := msg.Y - pickerTop - 1
					if toolIdx >= 0 && toolIdx < len(tools) {
						m.selectedTool = tools[toolIdx]
						return m, nil
					}
				}
			}

			// Check if clicking on control bar buttons
			if msg.Y < controlBarHeight {
				// Determine which button was clicked based on calculated positions
				// Check from right to left to handle overlapping ranges
				if m.toolbarToolX > 0 && msg.X >= m.toolbarToolX {
					// Tool button
					m.showToolPicker = !m.showToolPicker
					m.showCharPicker = false
					m.showFgPicker = false
					m.showBgPicker = false
					return m, nil
				} else if m.toolbarBackgroundX > 0 && msg.X >= m.toolbarBackgroundX && msg.X < m.toolbarToolX {
					// Background button
					m.showBgPicker = !m.showBgPicker
					m.showCharPicker = false
					m.showFgPicker = false
					m.showToolPicker = false
					return m, nil
				} else if m.toolbarForegroundX > 0 && msg.X >= m.toolbarForegroundX && msg.X < m.toolbarBackgroundX {
					// Foreground button
					m.showFgPicker = !m.showFgPicker
					m.showCharPicker = false
					m.showBgPicker = false
					m.showToolPicker = false
					return m, nil
				} else if m.toolbarShapeX > 0 && msg.X >= m.toolbarShapeX && msg.X < m.toolbarForegroundX {
					// Shapes button
					m.showCharPicker = !m.showCharPicker
					m.showFgPicker = false
					m.showBgPicker = false
					m.showToolPicker = false
					if m.showCharPicker {
						m.shapesFocusOnPanel = false // Start with focus on categories
					}
					return m, nil
				}
			}
		}

		// Handle mouse press (start of stroke)
		if msg.Type == tea.MouseLeft && !m.mouseDown && msg.Y >= controlBarHeight {
			m.mouseDown = true
			m.canvasBeforeStroke = m.copyCanvas()
			m.startX = msg.X
			m.startY = msg.Y - controlBarHeight
			if m.selectedTool == "Rectangle" || m.selectedTool == "Ellipse" || m.selectedTool == "Select" {
				m.showPreview = true
				m.previewEndX = msg.X
				m.previewEndY = msg.Y - controlBarHeight
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
				canvasY := msg.Y - controlBarHeight
				clampedY := canvasY
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
			} else if m.selectedTool == "Point" && msg.Y >= controlBarHeight {
				canvasY := msg.Y - controlBarHeight
				m.canvas.Set(canvasY, msg.X, m.selectedChar, m.foregroundColor, m.backgroundColor)
			}
		}

		// Handle mouse release (end of stroke)
		if msg.Type == tea.MouseRelease && m.mouseDown {
			m.mouseDown = false
			m.showPreview = false
			m.previewPoints = nil // Clear cached preview points

			// Clamp coordinates to canvas bounds
			canvasY := msg.Y - controlBarHeight
			clampedY := canvasY
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
			} else if m.selectedTool == "Fill" {
				m.floodFill(clampedY, clampedX)
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

func (m *model) View() string {
	if !m.ready {
		return "Initializing..."
	}

	// Use actual canvas height for rendering
	canvasHeight := m.canvas.height

	var b strings.Builder

	// Render control bar at the top
	b.WriteString(m.renderControlBar())

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
		// Position at top of canvas (canvas row 0)
		popupStartY = 0
		// Align popup dot (at offset 2) with selected character
		popupX = m.toolbarShapeItemX - pickerContentOffset

		// Always show shapes submenu when category picker is open
		popup2 = m.renderShapesPicker()
		popup2Lines = strings.Split(popup2, "\n")
		// Align first shapes row with the selected category row
		popup2StartY = popupStartY + m.selectedCategory
		// Clamp so the bottom doesn't go below the canvas
		if popup2StartY+len(popup2Lines) > canvasHeight {
			popup2StartY = canvasHeight - len(popup2Lines)
		}
		if popup2StartY < 0 {
			popup2StartY = 0
		}
		// Position shapes picker to the right of category picker
		// Calculate width by getting the first line's width (they should all be same)
		categoryWidth := 0
		if len(popupLines) > 0 {
			categoryWidth = lipgloss.Width(popupLines[0])
		}
		popup2X = popupX + categoryWidth
	} else if m.showFgPicker {
		popup = m.renderColorPicker("Foreground")
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		// Align popup swatch column with toolbar swatch
		popupX = m.toolbarForegroundItemX - pickerContentOffset
	} else if m.showBgPicker {
		popup = m.renderColorPicker("Background")
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		// Align popup swatch column with toolbar swatch
		popupX = m.toolbarBackgroundItemX - pickerContentOffset
	} else if m.showToolPicker {
		popup = m.renderToolPicker()
		popupLines = strings.Split(popup, "\n")
		popupStartY = 0
		// Align popup tool name column with toolbar tool name
		popupX = m.toolbarToolItemX - pickerContentOffset
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

					// Don't draw outline for 1-wide or 1-tall selections (no internal area)
					hasWidth := minX != maxX
					hasHeight := minY != maxY

					if hasWidth && hasHeight && i >= minY && i <= maxY && col >= minX && col <= maxX {
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

					// Only draw outline if selection has both width and height
					hasWidth := minX != maxX
					hasHeight := minY != maxY

					if hasWidth && hasHeight && i >= minY && i <= maxY && col >= minX && col <= maxX {
						if i == minY || i == maxY || col == minX || col == maxX {
							// Draw persistent selection border
							highlightStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
							var char string
							if i == minY && col == minX {
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

		// Don't add newline after the last row to prevent scrolling
		if i < canvasHeight-1 {
			b.WriteString(lineBuilder.String() + "\n")
		} else {
			b.WriteString(lineBuilder.String())
		}
	}

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

func (m *model) renderControlBar() string {
	// Muted cyan background for the entire toolbar
	bgColor := lipgloss.Color("#0E7490")        // Dark muted cyan
	baseColor := lipgloss.Color("#E0E0E0")      // Light gray text for base
	highlightColor := lipgloss.Color("#FFFFFF") // White text for highlighted

	// Base style for toolbar buttons
	baseStyle := lipgloss.NewStyle().
		Background(bgColor).
		Foreground(baseColor).
		Padding(0, 1)

	// Highlighted style for active menu
	highlightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")). // Lighter cyan for highlight
		Foreground(highlightColor).
		Padding(0, 1)

	// Build button text with underline using ANSI codes inline
	underlineOn := "\x1b[4m"
	underlineOff := "\x1b[24m"

	// Track X position as we build components
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
	// Position of the selected character: padding + "S" + "hapes: " = 1 + 1 + 7 = 9
	m.toolbarShapeItemX = currentX + 9
	currentX += lipgloss.Width(shapeButton)

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
	fgText := fmt.Sprintf("%sF%soreground: %s", underlineOn, underlineOff, fgSwatch)
	var fgButton string
	if m.showFgPicker {
		fgButton = highlightStyle.Render(fgText)
	} else {
		fgButton = baseStyle.Render(fgText)
	}
	m.toolbarForegroundX = currentX + toolbarButtonPadding
	// Position of the color swatch: padding + "F" + "oreground: " = 1 + 1 + 11 = 13
	m.toolbarForegroundItemX = currentX + 13
	currentX += lipgloss.Width(fgButton)

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
	bgText := fmt.Sprintf("%sB%sackground: %s", underlineOn, underlineOff, bgSwatch)
	var bgButton string
	if m.showBgPicker {
		bgButton = highlightStyle.Render(bgText)
	} else {
		bgButton = baseStyle.Render(bgText)
	}
	m.toolbarBackgroundX = currentX + toolbarButtonPadding
	// Position of the color swatch: padding + "B" + "ackground: " = 1 + 1 + 11 = 13
	m.toolbarBackgroundItemX = currentX + 13
	currentX += lipgloss.Width(bgButton)

	// Tool button - show "Circle" when in circle mode
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
	// Position of the tool name: padding + "T" + "ool: " = 1 + 1 + 5 = 7
	m.toolbarToolItemX = currentX + 7
	currentX += lipgloss.Width(toolButton)

	// Mode indicator - show if clipboard has content
	modeIndicator := ""
	if m.clipboard != nil && m.clipboardHeight > 0 && m.clipboardWidth > 0 {
		modeText := fmt.Sprintf("Mode: Yank (%dx%d)", m.clipboardWidth, m.clipboardHeight)
		modeIndicator = baseStyle.Render(modeText)
	}

	// Assemble control bar without separators and wrap entire line with background
	var barContent string
	if modeIndicator != "" {
		barContent = fmt.Sprintf("%s%s%s%s%s",
			shapeButton, fgButton, bgButton, toolButton, modeIndicator)
	} else {
		barContent = fmt.Sprintf("%s%s%s%s",
			shapeButton, fgButton, bgButton, toolButton)
	}

	// Wrap the entire bar with width and background
	barStyle := lipgloss.NewStyle().
		Background(bgColor).
		Width(m.width)

	return barStyle.Render(barContent) + "\n"
}

func (m model) getShapeX() int {
	return m.toolbarShapeX
}

func (m model) getForegroundX() int {
	return m.toolbarForegroundX
}

func (m model) getBackgroundX() int {
	return m.toolbarBackgroundX
}

func (m model) getToolX() int {
	return m.toolbarToolX
}

const menuCount = 4

func (m *model) activeMenu() int {
	if m.showCharPicker {
		return 0
	} else if m.showFgPicker {
		return 1
	} else if m.showBgPicker {
		return 2
	} else if m.showToolPicker {
		return 3
	}
	return -1
}

func (m *model) openMenu(idx int) {
	m.showCharPicker = idx == 0
	m.showFgPicker = idx == 1
	m.showBgPicker = idx == 2
	m.showToolPicker = idx == 3
	m.showingShapes = idx == 0
	m.shapesFocusOnPanel = false
	if idx == 0 {
		m.selectedCategory = m.findSelectedCharCategory()
	}
	if idx >= 0 {
		m.lastMenu = idx
	}
}

func (m *model) closeMenus() {
	m.showCharPicker = false
	m.showFgPicker = false
	m.showBgPicker = false
	m.showToolPicker = false
	m.showingShapes = false
	m.shapesFocusOnPanel = false
}

func colorDisplayName(name string) string {
	if name == "transparent" {
		return "None"
	}
	display := strings.ReplaceAll(name, "_", " ")
	return strings.ToUpper(display[:1]) + display[1:]
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

	// Copy only the internal region (excluding border)
	// The border is visual only and doesn't represent actual pixels
	internalMinY := minY + 1
	internalMaxY := maxY - 1
	internalMinX := minX + 1
	internalMaxX := maxX - 1

	// If selection is too small to have an internal area, don't copy anything
	if internalMaxY < internalMinY || internalMaxX < internalMinX {
		m.clipboard = nil
		m.clipboardWidth = 0
		m.clipboardHeight = 0
		return
	}

	m.clipboardHeight = internalMaxY - internalMinY + 1
	m.clipboardWidth = internalMaxX - internalMinX + 1
	m.clipboard = make([][]Cell, m.clipboardHeight)

	for y := 0; y < m.clipboardHeight; y++ {
		m.clipboard[y] = make([]Cell, m.clipboardWidth)
		for x := 0; x < m.clipboardWidth; x++ {
			cell := m.canvas.Get(internalMinY+y, internalMinX+x)
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

	// First copy (which only copies internal region)
	m.copySelection()

	// Then clear only the internal region (same as what was copied)
	minY, maxY := m.selectionStartY, m.selectionEndY
	if m.selectionStartY > m.selectionEndY {
		minY, maxY = m.selectionEndY, m.selectionStartY
	}
	minX, maxX := m.selectionStartX, m.selectionEndX
	if m.selectionStartX > m.selectionEndX {
		minX, maxX = m.selectionEndX, m.selectionStartX
	}

	// Clear only internal pixels (excluding border)
	internalMinY := minY + 1
	internalMaxY := maxY - 1
	internalMinX := minX + 1
	internalMaxX := maxX - 1

	// Only clear if there's an internal area
	if internalMaxY >= internalMinY && internalMaxX >= internalMinX {
		for y := internalMinY; y <= internalMaxY; y++ {
			for x := internalMinX; x <= internalMaxX; x++ {
				m.canvas.Set(y, x, " ", "transparent", "transparent")
			}
		}
		// Save to history
		m.saveToHistory()
	}
}

func (m *model) paste() {
	if m.clipboard == nil || m.clipboardHeight == 0 || m.clipboardWidth == 0 {
		return
	}

	// Use actual canvas height to ensure paste is on canvas
	// Use constant instead of local variable
	canvasHeight := m.canvas.height

	// Only paste if mouse is on canvas
	if m.mouseY < controlBarHeight {
		return
	}

	// Convert screen Y to canvas Y
	canvasMouseY := m.mouseY - controlBarHeight

	// Paste with top-left corner at current mouse position
	for y := 0; y < m.clipboardHeight; y++ {
		for x := 0; x < m.clipboardWidth; x++ {
			targetY := canvasMouseY + y
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

func (m *model) floodFill(row, col int) {
	target := m.canvas.Get(row, col)
	if target == nil {
		return
	}

	targetChar := target.char
	targetFg := target.foregroundColor
	targetBg := target.backgroundColor

	if targetChar == m.selectedChar && targetFg == m.foregroundColor && targetBg == m.backgroundColor {
		return
	}

	type point struct{ r, c int }
	queue := []point{{row, col}}
	visited := make(map[point]bool)
	visited[point{row, col}] = true

	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]

		cell := m.canvas.Get(p.r, p.c)
		if cell == nil || cell.char != targetChar || cell.foregroundColor != targetFg || cell.backgroundColor != targetBg {
			continue
		}

		m.canvas.Set(p.r, p.c, m.selectedChar, m.foregroundColor, m.backgroundColor)

		for _, d := range [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
			np := point{p.r + d[0], p.c + d[1]}
			if !visited[np] {
				visited[np] = true
				queue = append(queue, np)
			}
		}
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

	focusedBg := lipgloss.Color("#0891B2")
	unfocusedBg := lipgloss.Color("#3A3A3A")

	selectedBg := focusedBg
	if m.shapesFocusOnPanel {
		selectedBg = unfocusedBg
	}
	selectedStyle := lipgloss.NewStyle().Background(selectedBg).Foreground(lipgloss.Color("#FFFFFF"))

	// Find max name length for consistent line widths
	maxNameLen := 0
	for _, group := range characterGroups {
		if len(group.name) > maxNameLen {
			maxNameLen = len(group.name)
		}
	}
	lineWidth := maxNameLen + 2 // " name "

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

func (m model) renderShapesPicker() string {
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

func (m model) renderToolPicker() string {
	pickerStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("12"))

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")).
		Foreground(lipgloss.Color("#FFFFFF"))

	// Find max tool name width for consistent line widths
	maxNameLen := 0
	for _, tool := range tools {
		name := tool
		if tool == "Ellipse" && m.circleMode {
			name = "Circle"
		}
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
	}
	lineWidth := maxNameLen + 2 // " name "

	var content strings.Builder
	for i, tool := range tools {
		// Show "Circle" instead of "Ellipse" when in circle mode
		displayName := tool
		if tool == "Ellipse" && m.circleMode {
			displayName = "Circle"
		}

		line := " " + displayName
		for len(line) < lineWidth {
			line += " "
		}

		if tool == m.selectedTool {
			content.WriteString(selectedStyle.Render(line))
		} else {
			content.WriteString(line)
		}
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

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#0891B2")).
		Foreground(lipgloss.Color("#FFFFFF"))

	currentColor := m.foregroundColor
	if title == "Background" {
		currentColor = m.backgroundColor
	}

	// Find max color name width for consistent line widths
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

		// Pad name to consistent width
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
