# pixl

A fast, lightweight terminal-based pixel art editor using Unicode characters.

Built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Features

- **9 drawing tools**: Point, Rectangle, Ellipse, Circle, Line, Fill, Box, Text, Select
- **8 box styles**: Single, Double, Rounded, Heavy, and 4 dashed variants with automatic border merging
- **Character palette**: 16 categories with hundreds of Unicode glyphs
- **Dual color support**: Foreground and background colors per cell
- **Command palette**: Fuzzy search for any tool or action with `:`
- **Eyedropper**: Sample glyph and colors from the canvas with `i`
- **Undo/redo**: Up to 50 levels, grouped by brushstroke
- **Live preview**: See shapes as you drag before committing
- **Configurable**: Theme colors, default tools, and keybindings via `~/.config/pixl/config`

## Installation

```bash
git clone <repo-url>
cd pixl
go build -o pixl ./cmd/pixl
./pixl
```

Or run directly:
```bash
go run ./cmd/pixl
```

## Usage

```bash
./pixl                    # Dynamic canvas, resizes with terminal
./pixl -w 40 -h 20       # Fixed 40x20 canvas
./pixl art.txt            # Open existing file
cat art.txt | ./pixl      # Read from stdin
```

On quit, the canvas is printed to stdout (or saved to the file if one was specified).

## Quick Start

- **Draw**: Click and drag on the canvas
- **Switch tools**: Press `t` to open the tool picker, or `:` for the command palette
- **Change glyph**: Press `g` to open the glyph picker
- **Change colors**: Press `f` for foreground, `b` for background
- **Undo/redo**: `u` / `r`
- **Quit**: `q`

## Documentation

- [Tools](docs/tools.md) — All tools and their behavior
- [Keyboard Shortcuts](docs/keyboard_shortcuts.md) — Complete keybinding reference
- [Command Palette](docs/command_palette.md) — Fuzzy search for tools and actions
- [Configuration](docs/configuration.md) — Config file options and theming
- [Architecture](docs/architecture.md) — How the codebase is structured
