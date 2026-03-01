package main

import (
	"strings"
	"testing"
)

func TestMergeBoxChars(t *testing.T) {
	tests := []struct {
		a, b rune
		want rune
	}{
		{'┐', '┌', '┬'},
		{'┘', '└', '┴'},
		{'│', '┌', '├'},
		{'│', '└', '├'},
		{'┐', '│', '┤'},
		{'┘', '│', '┤'},
		{'│', '│', '│'},
		{'─', '─', '─'},
		{'┤', '├', '┼'},
	}
	for _, tt := range tests {
		got := mergeBoxChars(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("mergeBoxChars(%c, %c) = %c, want %c", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestAnsiTrimLastChar(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "plain text",
			input: "hello┐",
			want:  "hello",
		},
		{
			name:  "with ANSI suffix",
			input: "hello\x1b[34m┐\x1b[0m",
			want:  "hello\x1b[34m\x1b[0m",
		},
		{
			name:  "only char",
			input: "┐",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansiTrimLastChar(tt.input)
			if got != tt.want {
				t.Errorf("ansiTrimLastChar(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAnsiReplaceFirstChar(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		replacement rune
		want        string
	}{
		{
			name:        "plain text",
			input:       "┌hello",
			replacement: '├',
			want:        "├hello",
		},
		{
			name:        "with ANSI prefix",
			input:       "\x1b[34m┌\x1b[0mhello",
			replacement: '├',
			want:        "\x1b[34m├\x1b[0mhello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ansiReplaceFirstChar(tt.input, tt.replacement)
			if got != tt.want {
				t.Errorf("ansiReplaceFirstChar(%q, %c) = %q, want %q", tt.input, tt.replacement, got, tt.want)
			}
		})
	}
}

func TestMergePopupBorders(t *testing.T) {
	// Simulate two simple bordered panels side by side
	popup1 := []string{
		"\x1b[34m┌──┐\x1b[0m",
		"\x1b[34m│\x1b[0mAB\x1b[34m│\x1b[0m",
		"\x1b[34m│\x1b[0mCD\x1b[34m│\x1b[0m",
		"\x1b[34m└──┘\x1b[0m",
	}
	popup2 := []string{
		"\x1b[34m┌──┐\x1b[0m",
		"\x1b[34m│\x1b[0mEF\x1b[34m│\x1b[0m",
		"\x1b[34m└──┘\x1b[0m",
	}

	merged1, merged2 := mergePopupBorders(popup1, popup2, 1)

	// popup1 line 0 should be unchanged (no overlap)
	if merged1[0] != popup1[0] {
		t.Errorf("merged1[0] = %q, want %q", merged1[0], popup1[0])
	}

	// popup1 lines 1-3 should have last char trimmed
	for i := 1; i <= 3; i++ {
		if strings.Contains(merged1[i], "│\x1b[0m") && i >= 1 && i <= 3 {
			// Last visible char should be removed
		}
		trimmed := ansiTrimLastChar(popup1[i])
		if merged1[i] != trimmed {
			t.Errorf("merged1[%d] = %q, want %q", i, merged1[i], trimmed)
		}
	}

	// popup2 first chars should be merged junctions
	// popup2[0] starts at popup1 row 1: popup1 has │, popup2 has ┌ → ├
	firstChar0 := firstVisibleChar(merged2[0])
	if firstChar0 != '├' {
		t.Errorf("merged2[0] first char = %c, want ├", firstChar0)
	}

	// popup2[1] starts at popup1 row 2: popup1 has │, popup2 has │ → │
	firstChar1 := firstVisibleChar(merged2[1])
	if firstChar1 != '│' {
		t.Errorf("merged2[1] first char = %c, want │", firstChar1)
	}

	// popup2[2] starts at popup1 row 3: popup1 has ┘, popup2 has └ → ┴
	firstChar2 := firstVisibleChar(merged2[2])
	if firstChar2 != '┴' {
		t.Errorf("merged2[2] first char = %c, want ┴", firstChar2)
	}
}

func TestMergePopupBordersOffset(t *testing.T) {
	popup1 := []string{
		"\x1b[34m┌──┐\x1b[0m",
		"\x1b[34m│\x1b[0mAB\x1b[34m│\x1b[0m",
		"\x1b[34m│\x1b[0mCD\x1b[34m│\x1b[0m",
		"\x1b[34m│\x1b[0mEF\x1b[34m│\x1b[0m",
		"\x1b[34m│\x1b[0mGH\x1b[34m│\x1b[0m",
		"\x1b[34m└──┘\x1b[0m",
	}
	popup2 := []string{
		"\x1b[34m┌──┐\x1b[0m",
		"\x1b[34m│\x1b[0mXY\x1b[34m│\x1b[0m",
		"\x1b[34m└──┘\x1b[0m",
	}

	// popup2 starts at row 2 of popup1
	merged1, merged2 := mergePopupBorders(popup1, popup2, 2)

	// Rows 0-1 of popup1 should be unchanged
	if merged1[0] != popup1[0] {
		t.Errorf("merged1[0] should be unchanged")
	}
	if merged1[1] != popup1[1] {
		t.Errorf("merged1[1] should be unchanged")
	}

	// Rows 2-4 of popup1 should have last char trimmed
	for i := 2; i <= 4; i++ {
		trimmed := ansiTrimLastChar(popup1[i])
		if merged1[i] != trimmed {
			t.Errorf("merged1[%d] = %q, want %q", i, merged1[i], trimmed)
		}
	}

	// popup2 row 0 at popup1 row 2: popup1 has │, popup2 has ┌ → ├
	if fc := firstVisibleChar(merged2[0]); fc != '├' {
		t.Errorf("merged2[0] first char = %c, want ├", fc)
	}
	// popup2 row 2 at popup1 row 4: popup1 has │, popup2 has └ → ├
	if fc := firstVisibleChar(merged2[2]); fc != '├' {
		t.Errorf("merged2[2] first char = %c, want ├", fc)
	}
}

// helper to extract first visible rune from an ANSI string
func firstVisibleChar(s string) rune {
	i := 0
	runes := []rune(s)
	for i < len(runes) {
		if runes[i] == '\x1b' {
			for i < len(runes) && runes[i] != 'm' {
				i++
			}
			i++ // skip 'm'
			continue
		}
		return runes[i]
	}
	return 0
}
