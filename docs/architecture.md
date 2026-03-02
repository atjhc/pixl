# Architecture

pixl is built on [Bubble Tea](https://github.com/charmbracelet/bubbletea) using the Elm Architecture. The entire application is a single Go package (`package main`) under `cmd/pixl/`.

## Elm Architecture

- **Model** (`model` struct in `main.go`): Central application state — canvas, tool selection, UI flags, history, config
- **Update** (`Update()` in `input.go`): Handles all input events (keyboard, mouse, window resize, timers)
- **View** (`View()` in `view.go`): Renders the full screen each frame

## File Responsibilities

| File | Role |
|---|---|
| `main.go` | Model struct, `initialModel()`, program entry |
| `input.go` | All keyboard and mouse event handling |
| `view.go` | Screen rendering, cell rendering, canvas border |
| `tool_interface.go` | Tool interface + all tool implementations |
| `tools.go` | Drawing algorithms (Bresenham line, midpoint ellipse, flood fill) |
| `picker.go` | Picker panel rendering (glyphs, colors, tools, box styles) |
| `toolbar.go` | Toolbar rendering |
| `menu.go` | Menu state management, tool picker logic |
| `history.go` | Undo/redo stack, clipboard operations |
| `canvas.go` | Canvas data structure, file I/O |
| `palette.go` | Character groups (16 categories) and color definitions |
| `palette_cmd.go` | Command palette (fuzzy search, tab completion) |
| `border_merge.go` | Box-drawing border merging with T-junctions |
| `config.go` | Config file parsing, validation, application to model |
| `theme.go` | Theme struct, defaults, color resolution |

## Rendering Model

The view renders **column-by-column** within each row. For each cell, it checks (in order):

1. Is the cell inside a popup/picker overlay?
2. Is the cell inside a modal dialog (palette, confirm, alert)?
3. Is the cell the text insertion cursor?
4. Is the cell the hover cursor?
5. Render the actual canvas cell

Popups overlay the canvas by checking bounds per-column. Adjacent popup panels have their borders merged via `mergePopupBorders()`. This column-by-column approach prevents ANSI escape code bleeding between overlapping regions.

## Tool System

Tools implement the `Tool` interface:

```go
type Tool interface {
    Name() string
    DisplayName(m *model) string
    CursorChar(m *model) string
    OnPress(m *model, y, x int)
    OnDrag(m *model, y, x int)
    OnRelease(m *model, y, x int)
    OnKeyPress(m *model, key string) bool
    RenderPreview(m *model, row, col int) (string, bool)
    ModifiesCanvas() bool
}
```

Tools follow a mouse lifecycle: `OnPress` → `OnDrag` (repeated) → `OnRelease`. Shape tools store preview points during drag; `RenderPreview` draws them without modifying the canvas. The canvas is only modified on release. History snapshots are saved per-brushstroke, not per-cell.

## Menu System

Four toolbar menus defined via `iota` constants:

```go
const (
    menuForeground = iota
    menuBackground
    menuGlyph
    menuTool
    menuCount
)
```

The `menuKeys` array maps each index to its hotkey (`f`, `b`, `g`, `t`). This data-driven approach ensures the cycling order (`[`/`]`) always matches the toolbar visual order.

## Performance

- Shape preview points calculated once per mouse move, not per render
- Bresenham's line and midpoint ellipse algorithms for efficient shape drawing
- Column-by-column rendering minimizes ANSI escape processing
- Canvas snapshots only saved when actual changes occur
- Bounded history (50 undo levels max) to prevent unbounded memory growth
