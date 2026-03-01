package main

import (
	"bufio"
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
			c.DefaultGlyph = val
		case "default-foreground":
			c.DefaultForeground = val
		case "default-background":
			c.DefaultBackground = val
		case "default-tool":
			c.DefaultTool = val
		case "default-box-style":
			c.DefaultBoxStyle = val
		default:
			if ptr := c.Theme.field(key); ptr != nil && isValidThemeColor(val) {
				*ptr = val
			}
		}
	}

	return c
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
