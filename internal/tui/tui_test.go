package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"genius/internal/workspace"
)

func TestHomeRenders(t *testing.T) {
	m := New("claude", workspace.Workspace{Root: "/home/u/study"}, nil)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()

	for _, want := range []string{"getting started:", "No courses yet", "engine:claude", "/home/u/study"} {
		if !strings.Contains(out, want) {
			t.Errorf("home view missing %q", want)
		}
	}
}

func TestQuitKey(t *testing.T) {
	m := New("claude", workspace.Workspace{Root: "/x"}, nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should quit from home")
	}
}

func TestPopulatedHomeAndNav(t *testing.T) {
	courses := []workspace.Course{
		{Name: "algebra", HasGuide: true, HasQA: true, ExerciseSets: 1},
		{Name: "history", ExerciseSets: 0},
	}
	m := New("codex", workspace.Workspace{Root: "/s"}, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()
	for _, want := range []string{"COURSES", "algebra", "history", "g·1", "e·0"} {
		if !strings.Contains(out, want) {
			t.Errorf("populated home missing %q", want)
		}
	}

	// cursor starts at 0; down moves to 1, clamps there.
	model := mm.(Model)
	if model.cursor != 0 {
		t.Fatalf("cursor should start 0, got %d", model.cursor)
	}
	down := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	m2, _ := model.Update(down)
	if m2.(Model).cursor != 1 {
		t.Errorf("down should move cursor to 1")
	}
	m3, _ := m2.(Model).Update(down)
	if m3.(Model).cursor != 1 {
		t.Errorf("cursor should clamp at last index")
	}
}

func TestGradientColorAtBounds(t *testing.T) {
	if got := gradientColorAt(0); string(got) != "#6BA8F5" {
		t.Errorf("t=0 want #6BA8F5 got %s", got)
	}
	if got := gradientColorAt(1); string(got) != "#E66FB0" {
		t.Errorf("t=1 want #E66FB0 got %s", got)
	}
}
