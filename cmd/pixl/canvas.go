package main

import (
	"os"
	"strconv"
	"strings"
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

// LoadText populates the canvas from plain text, one character per cell.
func (c *Canvas) LoadText(text string) {
	lines := strings.Split(text, "\n")
	for row, line := range lines {
		if row >= c.height {
			break
		}

		fg := "white"
		bg := "transparent"
		col := 0
		i := 0
		runes := []rune(line)

		for i < len(runes) {
			if col >= c.width {
				break
			}

			if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
				// Parse ANSI escape: \x1b[...m
				j := i + 2
				for j < len(runes) && runes[j] != 'm' {
					j++
				}
				if j < len(runes) {
					params := string(runes[i+2 : j])
					fg, bg = applyANSIParams(params, fg, bg)
				}
				// Skip past 'm' if found, or past entire malformed sequence
				i = j + 1
				continue
			}

			r := runes[i]
			if r != ' ' {
				c.cells[row][col] = Cell{char: string(r), foregroundColor: fg, backgroundColor: bg}
			}
			col++
			i++
		}
	}
}

func applyANSIParams(params, fg, bg string) (string, string) {
	if params == "" || params == "0" {
		return "white", "transparent"
	}
	for _, p := range strings.Split(params, ";") {
		code, err := strconv.Atoi(p)
		if err != nil {
			continue
		}
		if name, ok := ansiFgToName[code]; ok {
			fg = name
		}
		if name, ok := ansiBgToName[code]; ok {
			bg = name
		}
		if code == 0 {
			fg = "white"
			bg = "transparent"
		}
	}
	return fg, bg
}

func visibleWidth(line string) int {
	width := 0
	runes := []rune(line)
	i := 0
	for i < len(runes) {
		if runes[i] == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && runes[j] != 'm' {
				j++
			}
			i = j + 1
			continue
		}
		width++
		i++
	}
	return width
}

var ansiFgToName = map[int]string{
	30: "black", 31: "red", 32: "green", 33: "yellow",
	34: "blue", 35: "magenta", 36: "cyan", 37: "white",
	90: "bright_black", 91: "bright_red", 92: "bright_green", 93: "bright_yellow",
	94: "bright_blue", 95: "bright_magenta", 96: "bright_cyan", 97: "bright_white",
}

var ansiBgToName = map[int]string{
	40: "black", 41: "red", 42: "green", 43: "yellow",
	44: "blue", 45: "magenta", 46: "cyan", 47: "white",
	100: "bright_black", 101: "bright_red", 102: "bright_green", 103: "bright_yellow",
	104: "bright_blue", 105: "bright_magenta", 106: "bright_cyan", 107: "bright_white",
}

func saveFile(path, content string) error {
	perm := os.FileMode(0666)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	}
	return os.WriteFile(path, []byte(content), perm)
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

// Copy returns a deep copy of the canvas.
func (c Canvas) Copy() Canvas {
	cells := make([][]Cell, c.height)
	for i := range cells {
		cells[i] = make([]Cell, c.width)
		copy(cells[i], c.cells[i])
	}
	return Canvas{width: c.width, height: c.height, cells: cells}
}

// Equals returns true if every cell in both canvases matches.
func (c Canvas) Equals(other Canvas) bool {
	if c.width != other.width || c.height != other.height {
		return false
	}
	for row := 0; row < c.height; row++ {
		for col := 0; col < c.width; col++ {
			cell1 := c.Get(row, col)
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
