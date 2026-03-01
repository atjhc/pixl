package main

import "math"

func normalizeRect(y1, x1, y2, x2 int) (minY, minX, maxY, maxX int) {
	minY, maxY = y1, y2
	if y1 > y2 {
		minY, maxY = y2, y1
	}
	minX, maxX = x1, x2
	if x1 > x2 {
		minX, maxX = x2, x1
	}
	return
}

func (m *model) drawRectangle(y1, x1, y2, x2 int) {
	minY, minX, maxY, maxX := normalizeRect(y1, x1, y2, x2)

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			if y == minY || y == maxY || x == minX || x == maxX {
				m.canvas.Set(y, x, m.selectedChar, m.foregroundColor, m.backgroundColor)
			}
		}
	}
}

func (m *model) drawBox(y1, x1, y2, x2 int) {
	minY, minX, maxY, maxX := normalizeRect(y1, x1, y2, x2)
	s := boxStyles[m.boxStyle]

	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			var ch string
			switch {
			case y == minY && x == minX:
				ch = s.tl
			case y == minY && x == maxX:
				ch = s.tr
			case y == maxY && x == minX:
				ch = s.bl
			case y == maxY && x == maxX:
				ch = s.br
			case y == minY || y == maxY:
				ch = s.h
			case x == minX || x == maxX:
				ch = s.v
			default:
				continue
			}
			if m.config.MergeBoxBorders && s.cross != "" {
				if existing := m.canvas.Get(y, x); existing != nil {
					eu, ed, el, er, ok := s.dirs(existing.char)
					if ok {
						nu, nd, nl, nr, _ := s.dirs(ch)
						ch = s.fromDirs(eu || nu, ed || nd, el || nl, er || nr)
					}
				}
			}
			m.canvas.Set(y, x, ch, m.foregroundColor, m.backgroundColor)
		}
	}
}

func (m *model) getCirclePoints(y1, x1, y2, x2 int, forceCircle bool) map[[2]int]bool {
	points := make(map[[2]int]bool)

	minY, minX, maxY, maxX := normalizeRect(y1, x1, y2, x2)

	centerY := (minY + maxY) / 2
	centerX := (minX + maxX) / 2

	if forceCircle {
		dy := y2 - y1
		dx := x2 - x1
		radius := int(0.5 + math.Sqrt(float64(dy*dy+dx*dx)))

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
	}

	// Ellipse based on bounding box
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

	for qi := 0; qi < len(queue); qi++ {
		p := queue[qi]

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

func getLinePoints(y1, x1, y2, x2 int) map[[2]int]bool {
	points := make(map[[2]int]bool)

	dy := y2 - y1
	dx := x2 - x1
	if dy < 0 {
		dy = -dy
	}
	if dx < 0 {
		dx = -dx
	}

	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}

	err := dx - dy
	for {
		points[[2]int{y1, x1}] = true
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}

	return points
}

func (m *model) drawLine(y1, x1, y2, x2 int) {
	for pt := range getLinePoints(y1, x1, y2, x2) {
		m.canvas.Set(pt[0], pt[1], m.selectedChar, m.foregroundColor, m.backgroundColor)
	}
}

func getEllipsePoints(centerY, centerX, ry, rx int) map[[2]int]bool {
	points := make(map[[2]int]bool)

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
	points[[2]int{centerY + y, centerX + x}] = true
	points[[2]int{centerY + y, centerX - x}] = true
	points[[2]int{centerY - y, centerX + x}] = true
	points[[2]int{centerY - y, centerX - x}] = true
}
