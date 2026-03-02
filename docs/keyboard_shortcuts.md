# Keyboard Shortcuts

## Menus

| Key | Action |
|---|---|
| `f` | Toggle Foreground color picker |
| `b` | Toggle Background color picker |
| `g` | Toggle Glyph picker |
| `t` | Toggle Tool picker |
| `[` | Cycle to previous menu |
| `]` | Cycle to next menu |
| `1`-`9` | Select item by number in open picker |
| Arrow keys | Navigate within open picker |
| `Right` | Drill into submenu |
| `Left` | Back out of submenu |
| `Enter` | Confirm selection and close picker |
| `Esc` | Close picker (or back out one level) |

## Drawing

| Key | Action |
|---|---|
| `i` | Eyedropper — sample glyph and colors from canvas |
| `x` | Swap foreground and background colors |
| `Return` | Toggle Ellipse/Circle mode (Ellipse tool) |
| `Return` | Cycle box style (Box tool) |
| `Option/Alt` | Temporary circle mode while held (Ellipse tool) |

## Text Mode

When the Text tool is active and an insertion point is placed:

| Key | Action |
|---|---|
| Any character | Insert at cursor, advance right |
| `Space` | Insert space |
| `Return` | Move to next line at starting column |
| `Backspace` | Delete character to the left |
| `Option+Backspace` | Delete previous word |
| `Esc` | Exit text mode |

## Editing

| Key | Action |
|---|---|
| `u` | Undo (up to 50 levels) |
| `r` | Redo |
| `c` | Clear canvas (requires confirmation) |
| `y` | Yank (copy) selection |
| `d` | Delete (cut) selection |
| `p` | Paste at cursor |

## Command Palette

| Key | Action |
|---|---|
| `:` | Open command palette |
| Type | Filter commands by name |
| `Tab` | Autocomplete current word |
| `Up`/`Down` | Navigate results |
| `Enter` | Execute selected command |
| `Esc` or `:` | Close palette |
| `Backspace` | Delete character (closes if empty) |
| `Option+Backspace` | Delete word |

## Other

| Key | Action |
|---|---|
| `q` or `Ctrl+C` | Quit |

## Native Text Selection

To copy text from the canvas using your OS clipboard:

- **Terminal.app**: Hold **Option** and click-drag to select, then **Cmd+C** to copy
- **iTerm2**: Hold **Shift** and click-drag to select, then **Cmd+C** to copy

This bypasses the app's mouse capture and uses the terminal's native text selection.
