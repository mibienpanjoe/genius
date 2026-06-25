package tui

import (
	"context"
	"errors"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// genDoneMsg carries a freshly generated guide/qa artifact back to the UI thread.
type genDoneMsg struct {
	kind string // "guide" or "qa"
	md   string
	err  error
}

// generateOrOpen is the home action for the g/q keys: open the existing artifact
// when one is present, otherwise generate it in-app. force (G/Q) always
// regenerates, replacing any existing artifact.
func (m Model) generateOrOpen(kind string, force bool) (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	c := m.courses[m.cursor]
	has := c.HasGuide
	if kind == "qa" {
		has = c.HasQA
	}
	if has && !force {
		return m.openReader(kind)
	}
	return m.startGenerate(kind, c.Name)
}

// startGenerate resolves the course grounding and kicks off guide/qa generation
// asynchronously with a spinner, mirroring the solve flow. It refuses up front
// when no engine is wired or the course has no material (INV-05).
func (m Model) startGenerate(kind, course string) (tea.Model, tea.Cmd) {
	if m.eng == nil {
		m.notice = "no engine available to generate"
		m.noticeLvl = lvlWarn
		return m, nil
	}
	material, err := m.ws.CourseMaterial(course)
	if err != nil {
		if errors.Is(err, workspace.ErrNoMaterial) {
			m.notice = "course " + course + " has no source material — ingest a document first"
			m.noticeLvl = lvlWarn
		} else {
			m.notice = "grounding error: " + err.Error()
			m.noticeLvl = lvlErr
		}
		return m, nil
	}
	m.genKind = kind
	m.genCourse = course
	m.state = stateGenerating
	if m.reduceMotion {
		return m, genCmd(m.eng, kind, course, material)
	}
	return m, tea.Batch(m.spinner.Tick, genCmd(m.eng, kind, course, material))
}

// genCmd runs the generator off the UI goroutine and reports the result. Q&A
// uses the default count/scope; narrower control stays a CLI concern.
func genCmd(eng engine.Engine, kind, course, material string) tea.Cmd {
	return func() tea.Msg {
		var (
			md  string
			err error
		)
		switch kind {
		case "guide":
			md, err = generate.Guide(context.Background(), eng, course, material)
		case "qa":
			md, err = generate.QA(context.Background(), eng, course, material, generate.QAOpts{})
		}
		return genDoneMsg{kind: kind, md: md, err: err}
	}
}

// genDone persists the artifact (force-overwriting so regenerate works), updates
// the course's chip count, then opens the result in the reader.
func (m Model) genDone(msg genDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notice = "generate failed: " + msg.err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}

	var path string
	switch msg.kind {
	case "guide":
		path = m.ws.GuidePath(m.genCourse)
	case "qa":
		path = m.ws.QAPath(m.genCourse)
	}
	if err := m.ws.WriteArtifact(path, []byte(msg.md+"\n"), true); err != nil {
		m.notice = "could not write " + msg.kind + ": " + err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}

	// Reflect the new artifact in the dashboard chip counts.
	for i := range m.courses {
		if m.courses[i].Name == m.genCourse {
			if msg.kind == "guide" {
				m.courses[i].HasGuide = true
			} else {
				m.courses[i].HasQA = true
			}
		}
	}

	return m.openReader(msg.kind)
}

// viewGenerating mirrors viewSolving: a spinner with the grounding reassurance.
func (m Model) viewGenerating() string {
	what := "study guide"
	if m.genKind == "qa" {
		what = "Q&A"
	}
	body := m.spinnerHead() + "generating " + what + " for " + m.genCourse +
		" with " + m.engine + "…\n\n" + styleMuted.Render("grounded only in the course material")
	return lipgloss.NewStyle().Padding(2, 0, 0, 3).Render(body)
}
