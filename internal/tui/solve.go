package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/render"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// solveDoneMsg carries the engine's worked solution back to the UI thread.
type solveDoneMsg struct {
	md  string
	err error
}

// openExSets enters the solve flow for the highlighted course by listing its
// exercise sets. A course with no sets sets a home notice instead.
func (m Model) openExSets() (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	course := m.courses[m.cursor].Name
	sets, err := m.ws.ExerciseSets(course)
	if err != nil || len(sets) == 0 {
		m.notice = "no exercise sets for " + course +
			" — ingest one: genius ingest <file> --kind exercise --course " + course
		return m, nil
	}
	m.solveCourse = course
	m.exSets = sets
	m.exSetCursor = 0
	m.state = stateExSets
	return m, nil
}

func (m Model) updateExSets(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.state = stateHome
		return m, nil
	case "up", "k":
		if m.exSetCursor > 0 {
			m.exSetCursor--
		}
	case "down", "j":
		if m.exSetCursor < len(m.exSets)-1 {
			m.exSetCursor++
		}
	case "enter":
		return m.openExList(m.exSets[m.exSetCursor])
	}
	return m, nil
}

// openExList reads the chosen set, enumerates its exercises, and shows the
// selectable list. An un-enumerable set becomes a single whole-set item.
func (m Model) openExList(set string) (tea.Model, tea.Cmd) {
	md, err := m.ws.ReadExerciseSet(m.solveCourse, set)
	if err != nil {
		m.notice = "could not read set " + set + ": " + err.Error()
		m.state = stateHome
		return m, nil
	}
	exs, err := generate.Enumerate(md)
	if err != nil {
		if !errors.Is(err, generate.ErrNoExercises) {
			m.notice = "could not enumerate " + set + ": " + err.Error()
			m.state = stateHome
			return m, nil
		}
		exs = []generate.Exercise{{Label: "Whole set", Num: "all", Text: md}}
	}
	m.solveSet = set
	m.exItems = exs
	m.exCursor = 0
	m.exSelected = make(map[int]bool)
	m.state = stateExList
	return m, nil
}

func (m Model) updateExList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.state = stateExSets
		return m, nil
	case "up", "k":
		if m.exCursor > 0 {
			m.exCursor--
		}
	case "down", "j":
		if m.exCursor < len(m.exItems)-1 {
			m.exCursor++
		}
	case " ":
		m.exSelected[m.exCursor] = !m.exSelected[m.exCursor]
	case "enter":
		return m.startSolving()
	}
	return m, nil
}

// startSolving gathers the selected exercises (defaulting to the highlighted one
// when none are toggled), resolves the course grounding, and kicks off the
// engine asynchronously with a spinner.
func (m Model) startSolving() (tea.Model, tea.Cmd) {
	if m.eng == nil {
		m.notice = "no engine available to solve"
		m.state = stateHome
		return m, nil
	}

	var selected []generate.Exercise
	for i, ex := range m.exItems {
		if m.exSelected[i] {
			selected = append(selected, ex)
		}
	}
	if len(selected) == 0 {
		selected = []generate.Exercise{m.exItems[m.exCursor]}
	}

	material, err := m.ws.CourseMaterial(m.solveCourse)
	if err != nil {
		if errors.Is(err, workspace.ErrNoMaterial) {
			m.notice = "course " + m.solveCourse + " has no source material — ingest a document first"
		} else {
			m.notice = "grounding error: " + err.Error()
		}
		m.state = stateHome
		return m, nil
	}

	m.state = stateSolving
	return m, tea.Batch(m.spinner.Tick, solveCmd(m.eng, m.solveCourse, material, selected))
}

// solveCmd runs generate.Solve off the UI goroutine and reports the result.
func solveCmd(eng engine.Engine, course, material string, exs []generate.Exercise) tea.Cmd {
	return func() tea.Msg {
		md, err := generate.Solve(context.Background(), eng, course, material, exs)
		return solveDoneMsg{md: md, err: err}
	}
}

// solveDone renders the worked solution into the reader, or surfaces an error.
func (m Model) solveDone(msg solveDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notice = "solve failed: " + msg.err.Error()
		m.state = stateHome
		return m, nil
	}
	m.resizeViewport()
	out, rerr := render.Markdown(msg.md, m.viewport.Width)
	if rerr != nil {
		out = msg.md
	}
	m.viewport.SetContent(out)
	m.viewport.GotoTop()
	m.readTitle = m.solveCourse + " · " + m.solveSet + " · solution"
	m.state = stateReader
	return m, nil
}

func (m Model) viewExSets() string {
	title := styleTitle.Render(m.solveCourse + " · exercise sets")
	var b string
	b = title + "\n\n"
	for i, s := range m.exSets {
		bar := "  "
		nameStyle := styleBody
		if i == m.exSetCursor {
			bar = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(cText)
		}
		b += bar + nameStyle.Render(s) + "\n"
	}
	b += "\n" + styleMuted.Render("enter open · esc back")
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b)
}

func (m Model) viewExList() string {
	title := styleTitle.Render(m.solveCourse + " · " + m.solveSet)
	var b string
	b = title + "\n\n"
	for i, ex := range m.exItems {
		cursor := "  "
		if i == m.exCursor {
			cursor = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
		}
		box := "[ ]"
		if m.exSelected[i] {
			box = lipgloss.NewStyle().Foreground(cSuccess).Render("[x]")
		}
		label := styleBody.Render(ex.Label)
		if i == m.exCursor {
			label = lipgloss.NewStyle().Bold(true).Foreground(cText).Render(ex.Label)
		}
		b += cursor + box + " " + label + "\n"
	}
	b += "\n" + styleMuted.Render("space select · enter solve (current if none) · esc back")
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b)
}

func (m Model) viewSolving() string {
	body := m.spinner.View() + " solving " + m.solveCourse + " · " + m.solveSet +
		" with " + m.engine + "…\n\n" + styleMuted.Render("grounded only in the course material")
	return lipgloss.NewStyle().Padding(2, 0, 0, 3).Render(body)
}
