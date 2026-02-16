package main

import "unicode/utf8"

func boxDirs(ch rune) (up, down, left, right bool) {
	switch ch {
	// Single / Rounded
	case '─':
		left, right = true, true
	case '│':
		up, down = true, true
	case '┌', '╭':
		down, right = true, true
	case '┐', '╮':
		down, left = true, true
	case '└', '╰':
		up, right = true, true
	case '┘', '╯':
		up, left = true, true
	case '├':
		up, down, right = true, true, true
	case '┤':
		up, down, left = true, true, true
	case '┬':
		down, left, right = true, true, true
	case '┴':
		up, left, right = true, true, true
	case '┼':
		up, down, left, right = true, true, true, true
	// Double
	case '═':
		left, right = true, true
	case '║':
		up, down = true, true
	case '╔':
		down, right = true, true
	case '╗':
		down, left = true, true
	case '╚':
		up, right = true, true
	case '╝':
		up, left = true, true
	case '╠':
		up, down, right = true, true, true
	case '╣':
		up, down, left = true, true, true
	case '╦':
		down, left, right = true, true, true
	case '╩':
		up, left, right = true, true, true
	case '╬':
		up, down, left, right = true, true, true, true
	// Heavy
	case '━':
		left, right = true, true
	case '┃':
		up, down = true, true
	case '┏':
		down, right = true, true
	case '┓':
		down, left = true, true
	case '┗':
		up, right = true, true
	case '┛':
		up, left = true, true
	case '┣':
		up, down, right = true, true, true
	case '┫':
		up, down, left = true, true, true
	case '┳':
		down, left, right = true, true, true
	case '┻':
		up, left, right = true, true, true
	case '╋':
		up, down, left, right = true, true, true, true
	}
	return
}

func dirsToBox(up, down, left, right bool) rune {
	switch {
	case up && down && left && right:
		return '┼'
	case up && down && right:
		return '├'
	case up && down && left:
		return '┤'
	case down && left && right:
		return '┬'
	case up && left && right:
		return '┴'
	case up && down:
		return '│'
	case left && right:
		return '─'
	case down && right:
		return '┌'
	case down && left:
		return '┐'
	case up && right:
		return '└'
	case up && left:
		return '┘'
	}
	return ch(up, down, left, right)
}

func ch(up, down, left, right bool) rune {
	if up {
		return '│'
	}
	if down {
		return '│'
	}
	if left {
		return '─'
	}
	if right {
		return '─'
	}
	return ' '
}

func mergeBoxChars(a, b rune) rune {
	au, ad, al, ar := boxDirs(a)
	bu, bd, bl, br := boxDirs(b)
	return dirsToBox(au || bu, ad || bd, al || bl, ar || br)
}

func lastVisibleCharPos(s string) (byteStart, byteEnd int, r rune) {
	byteStart = -1
	i := 0
	bytes := []byte(s)
	for i < len(bytes) {
		if bytes[i] == '\x1b' {
			for i < len(bytes) && bytes[i] != 'm' {
				i++
			}
			if i < len(bytes) {
				i++ // skip 'm'
			}
			continue
		}
		ru, size := utf8.DecodeRune(bytes[i:])
		byteStart = i
		byteEnd = i + size
		r = ru
		i += size
	}
	return
}

func firstVisibleCharPos(s string) (byteStart, byteEnd int, r rune) {
	i := 0
	bytes := []byte(s)
	for i < len(bytes) {
		if bytes[i] == '\x1b' {
			for i < len(bytes) && bytes[i] != 'm' {
				i++
			}
			if i < len(bytes) {
				i++ // skip 'm'
			}
			continue
		}
		ru, size := utf8.DecodeRune(bytes[i:])
		return i, i + size, ru
	}
	return -1, -1, 0
}

func ansiTrimLastChar(line string) string {
	start, end, _ := lastVisibleCharPos(line)
	if start < 0 {
		return line
	}
	return line[:start] + line[end:]
}

func ansiReplaceFirstChar(line string, replacement rune) string {
	start, end, _ := firstVisibleCharPos(line)
	if start < 0 {
		return line
	}
	return line[:start] + string(replacement) + line[end:]
}

func mergePopupBorders(popupLines, popup2Lines []string, popup2RowOffset int) ([]string, []string) {
	out1 := make([]string, len(popupLines))
	copy(out1, popupLines)
	out2 := make([]string, len(popup2Lines))
	copy(out2, popup2Lines)

	for i := range popup2Lines {
		p1Row := popup2RowOffset + i
		if p1Row < 0 || p1Row >= len(popupLines) {
			continue
		}

		_, _, lastChar := lastVisibleCharPos(popupLines[p1Row])
		_, _, firstChar := firstVisibleCharPos(popup2Lines[i])

		merged := mergeBoxChars(lastChar, firstChar)

		out1[p1Row] = ansiTrimLastChar(popupLines[p1Row])
		out2[i] = ansiReplaceFirstChar(popup2Lines[i], merged)
	}

	return out1, out2
}
