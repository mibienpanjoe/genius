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
	return m.startGenerate(kind, c.Name, nil)
}

// startGenerate resolves the course grounding and kicks off guide/qa generation
// asynchronously with a spinner, mirroring the solve flow. It refuses up front
// when no engine is wired or the course has no material (INV-05). When files is
// non-empty the grounding is scoped to those chapters (FR-055).
func (m Model) startGenerate(kind, course string, files []string) (tea.Model, tea.Cmd) {
	if m.eng == nil {
		m.notice = "no engine available to generate"
		m.noticeLvl = lvlWarn
		return m, nil
	}
	material, err := m.courseGrounding(course, files)
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
	m.genPath, m.genTitle = m.resolveGenTarget(kind, course, files)
	m.state = stateGenerating
	if m.reduceMotion {
		return m, genCmd(m.eng, kind, course, material)
	}
	return m, tea.Batch(m.spinner.Tick, genCmd(m.eng, kind, course, material))
}

// resolveGenTarget maps a generation scope to its artifact path and reader
// title, delegating placement to the workspace so the TUI and CLI agree: no
// files — or every chapter selected — targets the whole-course slot; any
// narrower selection targets a scoped artifact under the course subdir.
func (m Model) resolveGenTarget(kind, course string, files []string) (path, title string) {
	label := " · guide"
	var whole string
	if kind == "qa" {
		label = " · q&a"
		path, whole = m.ws.QATarget(course, files), m.ws.QAPath(course)
	} else {
		path, whole = m.ws.GuideTarget(course, files), m.ws.GuidePath(course)
	}
	if path == whole {
		return path, course + label
	}
	return path, course + "/" + workspace.ScopeName(files) + label
}

// courseGrounding returns the grounding blob for a generation: the whole course
// by default, or only the named chapter files when files is non-empty.
func (m Model) courseGrounding(course string, files []string) (string, error) {
	if len(files) > 0 {
		return m.ws.MaterialFromFiles(course, files)
	}
	return m.ws.CourseMaterial(course)
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

// genDone persists the artifact at the resolved scope path (force-overwriting so
// regenerate works), refreshes the dashboard chip counts, then opens the result
// in the reader.
func (m Model) genDone(msg genDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notice = "generate failed: " + msg.err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}

	if err := m.ws.WriteArtifact(m.genPath, []byte(msg.md+"\n"), true); err != nil {
		m.notice = "could not write " + msg.kind + ": " + err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}

	// Re-scan so a new scoped artifact shows up in the chip counts immediately.
	m.refreshCourses(m.genCourse)

	mm, cmd, err := m.openReaderPath(m.genPath, m.genTitle)
	if err != nil {
		m.notice = "generated, but could not open: " + err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}
	return mm, cmd
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
