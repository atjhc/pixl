# Command Palette

The command palette provides fuzzy-matched access to all tools, actions, and settings from a single entry point.

## Usage

Press `:` to open the palette. Type to filter, press Enter to execute.

```
╭─────────────────────────╮
│ : rec_                  │
│ ─────────────────────── │
│ → Rectangle             │
│   Redo                  │
╰─────────────────────────╯
```

The selected item is highlighted. Unselected items are dimmed.

## Navigation

- **Type** to filter results (case-insensitive)
- **Up/Down** to move the selection
- **Tab** to autocomplete the current word from the selected item
- **Enter** to execute the selected command
- **Esc** or **:** to close without executing
- **Backspace** to delete a character (closes palette if already empty)
- **Option+Backspace** to delete a word
- **Space** to type spaces (for multi-word searches like "dashed heavy")

## Matching

Items whose name starts with your query appear first, followed by substring matches. For example, typing `box` shows all box styles first, then any other item containing "box".

## Available Commands

### Tools

Point, Rectangle, Ellipse, Circle, Line, Fill, Select, Text

### Box Styles

Single Box, Double Box, Rounded Box, Heavy Box, Dashed Box, Dashed Heavy Box, Dense Dashed Box, Dense Heavy Box

### Actions

Clear Canvas, Undo, Redo, Copy, Cut, Paste, Swap Colors, Eyedropper

## Tab Completion

Tab completes to the next word boundary of the selected item. For example:

- `das` + Tab → `Dashed `
- `Dashed H` + Tab → `Dashed Heavy `

This lets you narrow multi-word items quickly without typing every character.
