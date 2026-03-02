# Configuration

pixl reads settings from `~/.config/pixl/config`. Lines starting with `#` are comments. Keys use kebab-case.

## Format

```
key = value
# This is a comment
```

## General Options

| Key | Default | Description |
|---|---|---|
| `merge-box-borders` | `true` | Merge overlapping box-drawing borders with T-junctions and crosses |
| `default-glyph` | `●` | Starting glyph character (must be a single character) |
| `default-foreground` | `white` | Starting foreground color |
| `default-background` | `transparent` | Starting background color |
| `default-tool` | `Point` | Starting tool (Point, Rectangle, Ellipse, Line, Fill, Box, Text, Select) |
| `default-box-style` | `Single` | Starting box style (Single, Double, Rounded, Heavy, Dashed, Dashed Heavy, Dense Dashed, Dense Heavy) |

## Theme Options

Theme colors control the UI appearance. An empty value (or omitting the key) uses the terminal's default colors.

| Key | Default | Description |
|---|---|---|
| `toolbar-bg` | *(terminal default)* | Toolbar background |
| `toolbar-fg` | *(terminal default)* | Toolbar text |
| `toolbar-highlight-bg` | *(terminal default)* | Active toolbar button background |
| `toolbar-highlight-fg` | *(terminal default)* | Active toolbar button text |
| `menu-border` | `bright-blue` | Picker/menu border color |
| `menu-selected-bg` | `bright-cyan` | Selected item background (focused picker) |
| `menu-selected-fg` | `bright-white` | Selected item text (focused picker) |
| `menu-unfocused-bg` | `bright-black` | Selected item background (unfocused picker) |
| `canvas-border` | `white` | Canvas border color (fixed-size mode) |
| `selection-fg` | `bright-blue` | Selection box border color |
| `cursor-fg` | `bright-black` | Hover cursor color |

## Color Formats

Theme colors accept:

- **Named colors** (16 standard terminal colors): `black`, `red`, `green`, `yellow`, `blue`, `magenta`, `cyan`, `white`, `bright-black`, `bright-red`, `bright-green`, `bright-yellow`, `bright-blue`, `bright-magenta`, `bright-cyan`, `bright-white`
- **ANSI numbers**: `0`-`255` (e.g. `42`)
- **Hex values**: `#RRGGBB` (e.g. `#0E7490`)

Canvas colors (foreground/background) use the named color set plus `transparent`.

ANSI numbers and hex values may not display correctly on all terminals.

## Validation

Invalid config values are rejected with a warning dialog shown on startup. The warning auto-dismisses after 5 seconds or on any keypress. Invalid values fall back to defaults.

## Example

```
# General
merge-box-borders = true
default-glyph = ●
default-foreground = white
default-background = transparent
default-tool = Point
default-box-style = Single

# Theme
menu-border = bright-blue
menu-selected-bg = bright-cyan
menu-selected-fg = bright-white
menu-unfocused-bg = bright-black
canvas-border = white
selection-fg = bright-blue
cursor-fg = bright-black
```
