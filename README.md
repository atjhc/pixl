# pixl

A fast, lightweight terminal-based pixel art editor using Unicode characters.

Built with Go and Bubble Tea.

## Features

### Drawing Tools
- **Point Tool**: Freehand drawing by clicking and dragging
- **Rectangle Tool**: Draw rectangle outlines with live preview
- **Ellipse Tool**: Draw ellipses/circles with live preview
  - Press **Return** to toggle between Ellipse and Circle modes
  - Circle mode draws perfect circles (compensated for terminal aspect ratio)
  - Hold **Option/Alt** + any key for temporary circle mode
- **Select Tool**: Drag to select regions, then copy/cut/paste
  - `y` - Yank (copy) selection
  - `d` - Delete (cut) selection
  - `p` - Paste at cursor location
  - Transparent cells don't overwrite destination when pasting

### Canvas & Display
- **Dynamic Canvas**: Automatically resizes to fit terminal window
- **Live Preview**: See shapes as you drag before committing
- **Dual Color Support**: Both foreground and background colors per cell

### Character Palette
16 categories with hundreds of Unicode characters:
- Circles, Squares, Triangles, Diamonds, Stars
- Blocks, Shading, Dots
- Box Drawing (Single, Double, Diagonal)
- Curves, Arrows, Hearts, Weather, Symbols

### Color Palette
Extended color support in three groups:
- **Basic**: Red, Green, Blue, Yellow, Magenta, Cyan, White
- **Extended**: Bright versions of all basic colors
- **Grayscale**: Black, Grey, Dark Grey
- **Transparent**: For background-only or foreground-only drawing

### History & Editing
- **Undo**: Press `u` to undo (up to 50 operations)
- **Redo**: Press `r` to redo
- **Brushstroke Grouping**: Each continuous drag is one undo operation
- **Clear Canvas**: Press `c` to clear everything

### Keyboard Shortcuts

**Pickers:**
- `f` - Toggle Foreground color picker
- `b` - Toggle Background color picker
- `t` - Toggle Tool picker (glyphs are accessed via the Tool picker)

**Editing:**
- `u` - Undo
- `r` - Redo
- `c` - Clear canvas
- `y` - Yank (copy) selection
- `d` - Delete (cut) selection
- `p` - Paste at cursor

**Other:**
- `Return` - Toggle Circle/Ellipse mode (when Ellipse tool selected)
- `Esc` - Close open pickers / Clear selection
- `q` or `Ctrl+C` - Quit

### Configuration

pixl reads settings from `~/.config/pixl/config`. Lines starting with `#` are comments.

```
# Defaults
merge-box-borders = true
default-glyph = ●
default-foreground = white
default-tool = Point

# Theme (defaults shown)
toolbar-bg = cyan
toolbar-fg = bright-white
toolbar-highlight-bg = bright-cyan
toolbar-highlight-fg = bright-white
menu-border = bright-blue
menu-selected-bg = bright-cyan
menu-selected-fg = bright-white
menu-unfocused-bg = bright-black
canvas-border = white
selection-fg = bright-blue
cursor-fg = bright-black
```

Theme colors accept any of the 16 standard terminal color names:
`black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`,
`bright-black`, `bright-red`, `bright-green`, `bright-yellow`, `bright-blue`,
`bright-magenta`, `bright-cyan`, `bright-white`.
ANSI numbers (e.g. `42`) and hex values (e.g. `#0E7490`) also work but may
not display correctly on all terminals.

### Mouse Support
- Click and drag to draw with Point tool
- Click and drag to create shapes with Rectangle/Ellipse tools
- Click pickers to select characters, colors, and tools
- Live preview shows shape outline while dragging
- Pickers stay open while drawing for quick switching

## Installation

```bash
# Clone the repository
git clone <repo-url>
cd pixl

# Build
go build -o pixl

# Run
./pixl
```

Or run directly without building:
```bash
go run main.go
```

## Architecture

Built using the Elm Architecture via Bubble Tea:

- **Model**: Application state (canvas, selected tool/char/colors, history, UI state)
- **Update**: Handle messages (mouse events, key presses, window resizes)
- **View**: Render current state using Lip Gloss with column-by-column rendering

### Key Files

- `main.go` - Complete application (~1600 lines)
  - Canvas data structure with dual-color cells
  - Tool system (Point, Rectangle, Ellipse, Select)
  - Shape drawing algorithms (Bresenham's rectangle, midpoint ellipse)
  - Selection and clipboard system with transparency support
  - Undo/redo history management (up to 50 levels)
  - Picker rendering and interaction
  - Column-by-column rendering to prevent ANSI bleed

### How It Works

1. **Canvas**: 2D array of cells, each storing character, foreground color, and background color
2. **Mouse Events**: Bubble Tea provides cell-accurate mouse coordinates
3. **Shape Algorithms**:
   - Rectangle: Simple edge detection
   - Ellipse: Midpoint ellipse algorithm with aspect ratio compensation
   - Preview: Pre-calculated and cached for performance
4. **History**: Undo stack with canvas snapshots, grouped by brushstroke
5. **Rendering**: Column-by-column rendering prevents ANSI escape code bleeding between popups and canvas
6. **Styling**: Lip Gloss for declarative terminal styling

## Performance Optimizations

- **Preview Caching**: Shape preview points calculated once per mouse move, not per render
- **Efficient Algorithms**: Bresenham's and midpoint algorithms for optimal shape drawing
- **Column-by-Column Rendering**: Minimizes ANSI escape code processing
- **Smart History**: Canvas snapshots only saved when changes occur
- **Bounded History**: Max 50 undo levels to prevent memory issues

## Why Bubble Tea?

- **Explicit State**: All state in one model struct, easy to reason about and debug
- **Pure Functions**: View is deterministic based on model state
- **Type Safety**: Go's type system catches errors at compile time
- **Event Loop**: Clean separation of input handling and rendering
- **Mouse Support**: Built-in cell-accurate mouse event handling
- **No Dependencies**: Single binary with no runtime dependencies
