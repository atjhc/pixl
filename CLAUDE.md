# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
go build -o pixl          # Build binary
go test ./...             # Run all tests
go test -run TestFoo      # Run a single test by name
go test -v ./...          # Verbose test output
```

No linter or formatter is configured beyond standard Go tooling (`go vet`, `gofmt`).

## Architecture

Pixl is a terminal-based pixel art editor built on **Bubble Tea** (Elm Architecture). The entire application is a single Go package (`package main`).

### Core Loop

Input events flow through `Update()` in **input.go**, which dispatches to tool methods and UI handlers. **view.go** renders the full screen each frame using column-by-column composition to prevent ANSI escape code bleeding between popups and the canvas.

### Key Types

- **`model`** (`main.go`): Central application state — canvas, tool selection, UI flags, history, config. Passed as `*model` to nearly everything.
- **`Canvas` / `Cell`** (`canvas.go`): 2D grid where each cell holds a character, foreground color, and background color.
- **`Tool` interface** (`tool_interface.go`): Defines `OnPress`/`OnDrag`/`OnRelease`/`RenderPreview` lifecycle. Seven implementations (Point, Rectangle, Box, Ellipse, Line, Fill, Select) are registered in `toolRegistry`.
- **`Config` / `Theme`** (`config.go`, `theme.go`): User settings loaded from `~/.config/pixl/config`. Theme colors are strings (hex or ANSI numbers) passed to `lipgloss.Color()` at usage sites.

### File Responsibilities

| File | Role |
|---|---|
| `main.go` | Model struct, `initialModel()`, program entry |
| `input.go` | All keyboard and mouse event handling (`Update`) |
| `view.go` | Screen rendering (`View`), cell rendering, canvas border |
| `tool_interface.go` | Tool interface + all tool implementations |
| `tools.go` | Drawing algorithms (Bresenham line, midpoint ellipse, flood fill) |
| `picker.go` | Picker panel rendering (glyphs, colors, tools, box styles) |
| `toolbar.go` | Control bar rendering |
| `menu.go` | Menu state management and navigation |
| `history.go` | Undo/redo stack (max 50), clipboard ops (yank/delete/paste) |
| `canvas.go` | Canvas data structure, file I/O |
| `palette.go` | Character groups (16 categories) and color definitions |
| `border_merge.go` | Box-drawing border merging with T-junctions |
| `config.go` | Config file parsing and application to model |
| `theme.go` | Theme struct and defaults |

### Rendering Model

The view renders **column-by-column** within each row. Popups (pickers) overlay the canvas by checking bounds per-column. Adjacent popup panels have their borders merged via `mergePopupBorders()`. All colors flow from `m.config.Theme` fields through `lipgloss.Color()`.

### Tool System

Tools follow a mouse lifecycle: `OnPress` → `OnDrag` (repeated) → `OnRelease`. Shape tools store preview points on `model` during drag; `RenderPreview` draws them without modifying the canvas. The canvas is only modified on release (or continuously for Point tool). History snapshots are saved per-brushstroke, not per-cell.

## Conventions

- Prefer guard clauses over nested conditionals.
- Use red/green TDD: write a failing test first, then make it pass.
- Keep comments minimal; prefer self-documenting code. When comments are used, explain *why*, not *what*.
- Config keys use kebab-case (e.g., `toolbar-highlight-bg`). Struct fields use PascalCase.
