# Tools

pixl provides 9 tools for drawing and editing. All drawing tools use the selected foreground color, background color, and glyph.

Tools follow a mouse lifecycle: click to start, drag to shape, release to commit. The canvas is only modified on release (or continuously for Point). Each brushstroke is one undo operation.

## Drawing Tools

### Point

Freehand drawing. Click and drag to place the selected glyph at each cell the cursor passes through.

### Rectangle

Draw rectangle outlines. Click to set one corner, drag to size, release to commit. Shows a live preview while dragging.

### Ellipse

Draw ellipses with live preview. Uses the midpoint ellipse algorithm with terminal aspect ratio compensation.

- Press **Return** to toggle between Ellipse and Circle modes
- **Circle mode** constrains to a perfect circle
- Hold **Option/Alt** for temporary circle mode during a single drag

### Line

Draw straight lines using Bresenham's algorithm. Click to set the start point, drag to the end point, release to commit.

### Fill

Flood fill from the clicked cell. Replaces all connected cells that match the clicked cell's character and colors with the selected glyph and colors.

### Text

Type text directly onto the canvas. Click to place a blinking insertion point, then type.

- Characters are placed with the current foreground/background colors
- **Return** moves to the next line at the same starting column
- **Backspace** deletes the character to the left
- **Option+Backspace** deletes the previous word
- **Space** inserts a space character
- **Esc** exits text mode

### Box

Draw rectangles using box-drawing characters. Automatically uses the correct corner, horizontal, and vertical characters for the selected style.

When `merge-box-borders` is enabled (default), overlapping box edges are merged with T-junctions and crosses.

Press **Return** to cycle through box styles:

| Style | Horizontal | Vertical | Corners |
|---|---|---|---|
| Single | `─` | `│` | `┌┐└┘` |
| Double | `═` | `║` | `╔╗╚╝` |
| Rounded | `─` | `│` | `╭╮╰╯` |
| Heavy | `━` | `┃` | `┏┓┗┛` |
| Dashed | `╌` | `┆` | `┌┐└┘` |
| Dashed Heavy | `╍` | `┇` | `┏┓┗┛` |
| Dense Dashed | `┄` | `┊` | `┌┐└┘` |
| Dense Heavy | `┅` | `┋` | `┏┓┗┛` |

## Selection Tool

### Select

Drag to select a rectangular region. The selection is shown with a dashed border.

- `y` - Yank (copy) the selection
- `d` - Delete (cut) the selection
- `p` - Paste at the cursor location
- Transparent cells in the clipboard don't overwrite the destination when pasting

## Eyedropper

Press `i` to sample the glyph, foreground color, and background color from the cell under the cursor. This sets all three as the current drawing settings without opening any picker.

## Switching Tools

- Open the **Tool picker** with `t` to browse and select tools
- Use the **command palette** (`:`) to fuzzy-search for any tool by name
- Number keys `1`-`9` select items within open pickers
