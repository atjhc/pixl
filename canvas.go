package main

import "strings"

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
		col := 0
		for _, r := range line {
			if col >= c.width {
				break
			}
			if r != ' ' {
				c.cells[row][col] = Cell{char: string(r), foregroundColor: "white", backgroundColor: "transparent"}
			}
			col++
		}
	}
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
	newCanvas := NewCanvas(c.width, c.height)
	for row := 0; row < c.height; row++ {
		for col := 0; col < c.width; col++ {
			cell := c.Get(row, col)
			if cell != nil {
				newCanvas.Set(row, col, cell.char, cell.foregroundColor, cell.backgroundColor)
			}
		}
	}
	return newCanvas
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
