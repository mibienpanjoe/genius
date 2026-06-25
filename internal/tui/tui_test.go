package tui

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mibienpanjoe/genius/internal/workspace"
)

func keyRunes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestHomeRenders(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{Root: "/home/u/study"}, nil)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()

	for _, want := range []string{"getting started:", "No courses yet", "engine:claude", "/home/u/study"} {
		if !strings.Contains(out, want) {
			t.Errorf("home view missing %q", want)
		}
	}
}

func TestQuitKey(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{Root: "/x"}, nil)
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
	m := New("codex", nil, workspace.Workspace{Root: "/s"}, courses)
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

func TestSolveFlowNavigation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("exercises", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ws.Path("exercises", "algebra", "td1.md"),
		[]byte("Exercice 1\nFoo.\n\nExercice 2\nBar.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	courses := []workspace.Course{{Name: "algebra", ExerciseSets: 1}}

	m := New("claude", nil, ws, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// s → set picker, lists the set.
	s, _ := mm.(Model).Update(keyRunes("s"))
	if s.(Model).state != stateExSets {
		t.Fatalf("s should open set picker, state=%d", s.(Model).state)
	}
	if !strings.Contains(s.View(), "td1") {
		t.Errorf("set not listed: %q", s.View())
	}

	// enter → exercise list, enumerated.
	el, _ := s.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if el.(Model).state != stateExList {
		t.Fatalf("enter should open exercise list, state=%d", el.(Model).state)
	}
	out := el.View()
	if !strings.Contains(out, "Exercice 1") || !strings.Contains(out, "Exercice 2") {
		t.Errorf("exercises not listed: %q", out)
	}

	// space toggles selection on the current item.
	sel, _ := el.(Model).Update(keyRunes(" "))
	if !sel.(Model).exSelected[0] {
		t.Errorf("space should select the current exercise")
	}

	// enter with a nil engine bounces home with a notice (engine guard).
	done, _ := sel.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if done.(Model).state != stateHome {
		t.Errorf("nil engine should return home, state=%d", done.(Model).state)
	}
	if !strings.Contains(done.(Model).notice, "no engine") {
		t.Errorf("want no-engine notice, got %q", done.(Model).notice)
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
