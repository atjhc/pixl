package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	MergeBoxBorders   bool
	DefaultGlyph      string
	DefaultForeground string
	DefaultBackground string
	DefaultTool       string
	DefaultBoxStyle   string
	Theme             Theme
	Warnings          []string
}

func loadConfig() Config {
	c := Config{
		MergeBoxBorders: true,
		Theme:           defaultTheme(),
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return c
	}

	f, err := os.Open(filepath.Join(home, ".config", "pixl", "config"))
	if err != nil {
		return c
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "merge-box-borders":
			c.MergeBoxBorders = val == "true"
		case "default-glyph":
			runes := []rune(val)
			if len(runes) != 1 {
				c.Warnings = append(c.Warnings, fmt.Sprintf("default-glyph must be a single character, got %q", val))
			} else {
				c.DefaultGlyph = val
			}
		case "default-foreground":
			if isValidCanvasColor(val) {
				c.DefaultForeground = val
			} else {
				c.Warnings = append(c.Warnings, fmt.Sprintf("invalid color %q for %s", val, key))
			}
		case "default-background":
			if isValidCanvasColor(val) {
				c.DefaultBackground = val
			} else {
				c.Warnings = append(c.Warnings, fmt.Sprintf("invalid color %q for %s", val, key))
			}
		case "default-tool":
			if isValidTool(val) {
				c.DefaultTool = val
			} else {
				c.Warnings = append(c.Warnings, fmt.Sprintf("invalid tool %q for %s", val, key))
			}
		case "default-box-style":
			if isValidBoxStyle(val) {
				c.DefaultBoxStyle = val
			} else {
				c.Warnings = append(c.Warnings, fmt.Sprintf("invalid box style %q for %s", val, key))
			}
		default:
			ptr := c.Theme.field(key)
			if ptr == nil {
				c.Warnings = append(c.Warnings, fmt.Sprintf("unknown config key %q", key))
			} else if !isValidThemeColor(val) {
				c.Warnings = append(c.Warnings, fmt.Sprintf("invalid color %q for %s", val, key))
			} else {
				*ptr = val
			}
		}
	}

	return c
}

func isValidTool(name string) bool {
	for _, t := range toolRegistry {
		if t.Name() == name {
			return true
		}
	}
	return false
}

func isValidBoxStyle(name string) bool {
	for _, s := range boxStyles {
		if s.name == name {
			return true
		}
	}
	return false
}

func (m *model) applyConfig() {
	if m.config.DefaultGlyph != "" {
		m.selectedChar = m.config.DefaultGlyph
	}
	if m.config.DefaultForeground != "" && isValidCanvasColor(m.config.DefaultForeground) {
		m.foregroundColor = m.config.DefaultForeground
	}
	if m.config.DefaultBackground != "" && isValidCanvasColor(m.config.DefaultBackground) {
		m.backgroundColor = m.config.DefaultBackground
	}
	if m.config.DefaultTool != "" {
		for _, t := range toolRegistry {
			if t.Name() != m.config.DefaultTool {
				continue
			}
			m.selectedTool = m.config.DefaultTool
			if isDrawingTool(m.config.DefaultTool) {
				m.drawingTool = m.config.DefaultTool
			}
			break
		}
	}
	if m.config.DefaultBoxStyle != "" {
		for i, s := range boxStyles {
			if s.name != m.config.DefaultBoxStyle {
				continue
			}
			m.boxStyle = i
			break
		}
	}
}
