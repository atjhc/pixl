package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	canvas             Canvas
	selectedChar       string
	foregroundColor    string
	backgroundColor    string
	mouseX             int
	mouseY             int
	width              int
	height             int
	ready              bool
	showCharPicker     bool
	showFgPicker       bool
	showBgPicker       bool
	showToolPicker     bool
	selectedTool       string
	selectedCategory   int
	showingShapes      bool
	shapesFocusOnPanel bool
	history            []Canvas
	historyIndex       int
	mouseDown          bool
	canvasBeforeStroke Canvas
	startX             int
	startY             int
	previewEndX        int
	previewEndY        int
	showPreview        bool
	optionKeyHeld      bool
	circleMode         bool
	previewPoints      map[[2]int]bool
	hasSelection       bool
	selectionStartY    int
	selectionStartX    int
	selectionEndY      int
	selectionEndX      int
	clipboard          [][]Cell
	clipboardWidth     int
	clipboardHeight    int
	lastMenu           int
	fixedWidth         int
	fixedHeight        int
	canvasInitialized  bool
	// Toolbar button positions (calculated during render)
	toolbarShapeX      int
	toolbarForegroundX int
	toolbarBackgroundX int
	toolbarToolX       int
	// Toolbar selected item positions (for popup alignment)
	toolbarShapeItemX      int
	toolbarForegroundItemX int
	toolbarBackgroundItemX int
	toolbarToolItemX       int
}

func initialModel() *model {
	canvas := NewCanvas(100, 30)
	return &model{
		canvas:          canvas,
		selectedChar:    "●",
		foregroundColor: "white",
		backgroundColor: "transparent",
		selectedTool:    "Point",
		ready:           false,
		history:         []Canvas{},
		historyIndex:    -1,
		mouseDown:       false,
	}
}

func (m *model) Init() tea.Cmd {
	return nil
}

func main() {
	flagW := flag.Int("w", 0, "fixed canvas width")
	flagH := flag.Int("h", 0, "fixed canvas height")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: pixl [--help] [-w width] [-h height]\n\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	m := initialModel()

	opts := []tea.ProgramOption{
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	}

	var stdinText string

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err == nil && len(data) > 0 {
			stdinText = string(data)
			m.canvas.LoadText(stdinText)
		}
		tty, err := os.Open("/dev/tty")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening terminal: %v\n", err)
			os.Exit(1)
		}
		defer tty.Close()
		opts = append(opts, tea.WithInput(tty))
	}

	if *flagW > 0 && *flagH > 0 {
		m.fixedWidth = *flagW
		m.fixedHeight = *flagH
	} else if stdinText != "" && *flagW == 0 && *flagH == 0 {
		lines := strings.Split(strings.TrimRight(stdinText, "\n"), "\n")
		maxWidth := 0
		for _, line := range lines {
			if len(line) > maxWidth {
				maxWidth = len(line)
			}
		}
		if maxWidth > 0 && len(lines) > 0 {
			m.fixedWidth = maxWidth
			m.fixedHeight = len(lines)
		}
	}

	p := tea.NewProgram(m, opts...)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if fm, ok := finalModel.(*model); ok {
		fmt.Print(fm.renderCanvas())
	}
}
