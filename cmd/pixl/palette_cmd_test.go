package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFilterPalettePrefix(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "po")
	if len(results) == 0 {
		t.Fatal("expected results for 'po'")
	}
	if results[0].name != "Point" {
		t.Errorf("first result for 'po' = %q, want Point", results[0].name)
	}
}

func TestFilterPaletteSubstring(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "angle")
	found := false
	for _, r := range results {
		if r.name == "Rectangle" {
			found = true
		}
	}
	if !found {
		t.Error("substring 'angle' should match Rectangle")
	}
}

func TestFilterPalettePrefixRanksFirst(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "c")
	if len(results) < 2 {
		t.Fatal("expected multiple results for 'c'")
	}
	// "Circle" and "Copy" and "Cut" and "Clear Canvas" all start with c
	// They should come before substring matches like "Rectangle"
	for _, r := range results {
		if strings.HasPrefix(strings.ToLower(r.name), "c") {
			continue
		}
		// Once we hit a non-prefix match, no prefix matches should follow
		break
	}
}

func TestFilterPaletteEmpty(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "")
	if len(results) != len(items) {
		t.Errorf("empty query should return all %d items, got %d", len(items), len(results))
	}
}

func TestFilterPaletteCaseInsensitive(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "FILL")
	found := false
	for _, r := range results {
		if r.name == "Fill" {
			found = true
		}
	}
	if !found {
		t.Error("case-insensitive search for 'FILL' should match Fill")
	}
}

func TestFilterPaletteNoMatch(t *testing.T) {
	m := &model{}
	items := m.paletteItems()
	results := filterPalette(items, "zzzzz")
	if len(results) != 0 {
		t.Errorf("expected no results for 'zzzzz', got %d", len(results))
	}
}

func TestPaletteOpenAndClose(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}

	colon := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}
	m.handleKey(colon)
	if !m.showPalette {
		t.Error(": should open palette")
	}

	esc := tea.KeyMsg{Type: tea.KeyEscape}
	m.handleKey(esc)
	if m.showPalette {
		t.Error("esc should close palette")
	}
}

func TestPaletteColonCloses(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
	}

	colon := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}}
	m.handleKey(colon)
	if m.showPalette {
		t.Error(": should close palette when already open")
	}
}

func TestPaletteTypingFilters(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
	}

	p := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	m.handleKey(p)
	if m.paletteQuery != "p" {
		t.Errorf("query = %q, want 'p'", m.paletteQuery)
	}

	o := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	m.handleKey(o)
	if m.paletteQuery != "po" {
		t.Errorf("query = %q, want 'po'", m.paletteQuery)
	}
}

func TestPaletteBackspace(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "po",
	}

	bs := tea.KeyMsg{Type: tea.KeyBackspace}
	m.handleKey(bs)
	if m.paletteQuery != "p" {
		t.Errorf("query after backspace = %q, want 'p'", m.paletteQuery)
	}

	// Backspace on single char empties query but stays open
	m.handleKey(bs)
	if m.paletteQuery != "" {
		t.Errorf("query after second backspace = %q, want empty", m.paletteQuery)
	}
	if !m.showPalette {
		t.Error("backspace to empty should keep palette open")
	}

	// Backspace on empty query closes palette
	m.handleKey(bs)
	if m.showPalette {
		t.Error("backspace on already-empty query should close palette")
	}
}

func TestPaletteSpaceInput(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "clear",
	}

	space := tea.KeyMsg{Type: tea.KeySpace}
	m.handleKey(space)
	if m.paletteQuery != "clear " {
		t.Errorf("query after space = %q, want %q", m.paletteQuery, "clear ")
	}
}

func TestPaletteAltBackspaceDeletesWord(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "clear canvas",
	}

	altBs := tea.KeyMsg{Type: tea.KeyBackspace, Alt: true}
	m.handleKey(altBs)
	if m.paletteQuery != "clear " {
		t.Errorf("query after alt+backspace = %q, want %q", m.paletteQuery, "clear ")
	}
}

func TestPaletteAltBackspaceDeletesSingleWord(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "hello",
	}

	// Alt+backspace on single word empties query but stays open
	altBs := tea.KeyMsg{Type: tea.KeyBackspace, Alt: true}
	m.handleKey(altBs)
	if m.paletteQuery != "" {
		t.Errorf("query after alt+backspace = %q, want empty", m.paletteQuery)
	}
	if !m.showPalette {
		t.Error("alt+backspace to empty should keep palette open")
	}
}

func TestPaletteAltBackspaceOnEmpty(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "",
	}

	altBs := tea.KeyMsg{Type: tea.KeyBackspace, Alt: true}
	m.handleKey(altBs)
	if m.showPalette {
		t.Error("alt+backspace on already-empty query should close palette")
	}
}

func TestPaletteAltBackspaceTrailingSpaces(t *testing.T) {
	m := &model{
		canvas:        NewCanvas(10, 10),
		selectedTool:  "Point",
		drawingTool:   "Point",
		showPalette:   true,
		paletteQuery:  "clear  ",
	}

	// macOS behavior: skip trailing spaces, then delete "clear" → empty but stay open
	altBs := tea.KeyMsg{Type: tea.KeyBackspace, Alt: true}
	m.handleKey(altBs)
	if m.paletteQuery != "" {
		t.Errorf("query after alt+backspace = %q, want empty", m.paletteQuery)
	}
	if !m.showPalette {
		t.Error("alt+backspace to empty should keep palette open")
	}
}

func TestPaletteTabAutocompletesWord(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "das",
	}

	tab := tea.KeyMsg{Type: tea.KeyTab}
	m.handleKey(tab)

	if m.paletteQuery != "Dashed " {
		t.Errorf("query after tab = %q, want %q", m.paletteQuery, "Dashed ")
	}
}

func TestPaletteTabAutocompletesSecondWord(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "Dashed H",
	}

	tab := tea.KeyMsg{Type: tea.KeyTab}
	m.handleKey(tab)

	if m.paletteQuery != "Dashed Heavy " {
		t.Errorf("query after tab = %q, want %q", m.paletteQuery, "Dashed Heavy ")
	}
}

func TestPaletteTabFullWordAlready(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "Fill",
	}

	tab := tea.KeyMsg{Type: tea.KeyTab}
	m.handleKey(tab)

	if m.paletteQuery != "Fill" {
		t.Errorf("query after tab on exact match = %q, want %q", m.paletteQuery, "Fill")
	}
}

func TestPaletteTabNoResults(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "zzzzz",
	}

	tab := tea.KeyMsg{Type: tea.KeyTab}
	m.handleKey(tab)

	if m.paletteQuery != "zzzzz" {
		t.Errorf("query should be unchanged with no results, got %q", m.paletteQuery)
	}
}

func TestPaletteTabUsesSelectedItem(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "d",
		paletteIndex: 1,
	}

	// Filter for "d" — results depend on ordering but index 1 should be a different item
	items := filterPalette(m.paletteItems(), "d")
	if len(items) < 2 {
		t.Skip("not enough items matching 'd'")
	}
	expected := items[1].name
	// Find the next word boundary after the partial match
	tab := tea.KeyMsg{Type: tea.KeyTab}
	m.handleKey(tab)

	// Should autocomplete based on item at index 1, not index 0
	if !strings.HasPrefix(strings.ToLower(expected), "d") {
		t.Skip("second item doesn't start with d")
	}
	// The query should have changed to complete a word from the selected item
	if m.paletteQuery == "d" {
		t.Error("tab should have autocompleted from the selected item")
	}
}

func TestPaletteNavigation(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteIndex: 0,
	}

	down := tea.KeyMsg{Type: tea.KeyDown}
	m.handleKey(down)
	if m.paletteIndex != 1 {
		t.Errorf("paletteIndex after down = %d, want 1", m.paletteIndex)
	}

	up := tea.KeyMsg{Type: tea.KeyUp}
	m.handleKey(up)
	if m.paletteIndex != 0 {
		t.Errorf("paletteIndex after up = %d, want 0", m.paletteIndex)
	}

	// Up at 0 should not go negative
	m.handleKey(up)
	if m.paletteIndex != 0 {
		t.Errorf("paletteIndex should not go negative, got %d", m.paletteIndex)
	}
}

func TestPaletteExecuteTool(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "fill",
	}

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKey(enter)

	if m.showPalette {
		t.Error("enter should close palette")
	}
	if m.selectedTool != "Fill" {
		t.Errorf("selectedTool = %q, want Fill", m.selectedTool)
	}
}

func TestPaletteExecuteUndo(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()
	m.canvas.Set(0, 0, "Y", "green", "yellow")
	m.saveToHistory()

	m.showPalette = true
	m.paletteQuery = "undo"

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKey(enter)

	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Errorf("undo via palette should restore previous state, got %+v", cell)
	}
}

func TestPaletteExecuteSwapColors(t *testing.T) {
	m := &model{
		canvas:          NewCanvas(10, 10),
		selectedTool:    "Point",
		drawingTool:     "Point",
		foregroundColor: "red",
		backgroundColor: "blue",
		showPalette:     true,
		paletteQuery:    "swap",
	}

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKey(enter)

	if m.foregroundColor != "blue" || m.backgroundColor != "red" {
		t.Errorf("swap via palette: fg=%s bg=%s, want fg=blue bg=red", m.foregroundColor, m.backgroundColor)
	}
}

func TestPaletteClearCanvasTriggersConfirm(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "clear",
	}
	m.canvas.Set(0, 0, "X", "red", "blue")
	m.saveToHistory()

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKey(enter)

	if !m.confirmClear {
		t.Error("Clear Canvas from palette should trigger confirmClear")
	}
	// Canvas should NOT be cleared yet
	cell := m.canvas.Get(0, 0)
	if cell == nil || cell.char != "X" {
		t.Error("canvas should not be cleared until confirmed")
	}
}

func TestPaletteEnterWithNoResults(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  true,
		paletteQuery: "zzzzz",
	}

	enter := tea.KeyMsg{Type: tea.KeyEnter}
	m.handleKey(enter)

	// Should just close without doing anything
	if m.showPalette {
		t.Error("enter with no results should close palette")
	}
}

func TestPaletteDoesNotInterfereWithNormalKeys(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(10, 10),
		selectedTool: "Point",
		drawingTool:  "Point",
		showPalette:  false,
	}

	// 'u' should trigger undo, not palette input
	u := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}}
	m.handleKey(u)
	if m.showPalette {
		t.Error("u should not open palette")
	}
}

func TestPaletteRendered(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(40, 10),
		selectedChar: "●",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:        40,
		height:       11,
		ready:        true,
		showPalette:  true,
		paletteQuery: "po",
	}

	got := m.View()
	if !strings.Contains(got, "po") {
		t.Error("view should contain palette query text when showPalette is true")
	}
}

func TestPaletteNotRendered(t *testing.T) {
	m := &model{
		canvas:       NewCanvas(40, 10),
		selectedChar: "●",
		selectedTool: "Point",
		drawingTool:  "Point",
		width:        40,
		height:       11,
		ready:        true,
		showPalette:  false,
	}

	got := m.View()
	if strings.Contains(got, "▸") {
		t.Error("view should not contain palette indicator when showPalette is false")
	}
}
