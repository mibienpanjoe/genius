package tui

import (
	"context"
	"errors"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// genDoneMsg carries a freshly generated guide/qa artifact back to the UI thread.
type genDoneMsg struct {
	kind  string // "guide" or "qa"
	md    string
	err   error
	epoch int // the work epoch this result belongs to (stale results ignored)
}

// generateOrOpen is the home action for the g/q keys: open the existing artifact
// when one is present, otherwise generate it in-app. force (G/Q) always
// regenerates, replacing any existing artifact.
func (m Model) generateOrOpen(kind string, force bool) (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	m.back = stateHome
	c := m.courses[m.cursor]
	has := c.HasGuide
	if kind == "qa" {
		has = c.HasQA
	}
	if has && !force {
		return m.openReader(kind)
	}
	if kind == "qa" {
		return m.promptQACount(c.Name, nil)
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
	m.genScope = strings.TrimSuffix(m.genTitle, genLabel(kind))
	m.state = stateGenerating
	ctx, epoch := m.beginWork()
	if m.reduceMotion {
		return m, genCmd(ctx, m.eng, kind, course, material, m.genCount, epoch)
	}
	return m, tea.Batch(m.spinner.Tick, genCmd(ctx, m.eng, kind, course, material, m.genCount, epoch))
}

// promptQACount defers Q&A generation behind a count entry (FR-054): it stashes
// the pending scope, seeds the input with the default, and shows the prompt. The
// engine is checked up front so a missing one never reaches the input.
func (m Model) promptQACount(course string, files []string) (tea.Model, tea.Cmd) {
	if m.eng == nil {
		m.notice = "no engine available to generate"
		m.noticeLvl = lvlWarn
		return m, nil
	}
	m.qaCourse = course
	m.qaFiles = files
	m.qaInput.SetValue("")
	m.qaInput.Focus()
	m.state = stateQACount
	return m, nil
}

// updateQACount drives the count entry: enter starts generation with the typed
// count (blank = default), esc returns to the originating screen.
func (m Model) updateQACount(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = m.back
		return m, nil
	case "enter":
		m.genCount = 0 // blank/invalid falls back to DefaultQACount
		if n, err := strconv.Atoi(strings.TrimSpace(m.qaInput.Value())); err == nil && n > 0 {
			m.genCount = n
		}
		return m.startGenerate("qa", m.qaCourse, m.qaFiles)
	}
	var cmd tea.Cmd
	m.qaInput, cmd = m.qaInput.Update(msg)
	return m, cmd
}

// viewQACount renders the pair-count prompt before Q&A generation.
func (m Model) viewQACount() string {
	scope := m.qaCourse
	if len(m.qaFiles) > 0 {
		scope = m.qaCourse + "/" + workspace.ScopeName(m.qaFiles)
	}
	body := styleTitle.Render("Q&A · "+scope) + "\n\n" +
		styleBody.Render("how many Q&A pairs? ") + m.qaInput.View() + "\n\n" +
		styleMuted.Render("default "+strconv.Itoa(generate.DefaultQACount)+" · enter generate · esc back")
	return lipgloss.NewStyle().Padding(2, 0, 0, 3).Render(body)
}

// digitsOnly rejects any non-digit keystroke in the count input.
func digitsOnly(s string) error {
	for _, r := range s {
		if r < '0' || r > '9' {
			return errors.New("digits only")
		}
	}
	return nil
}

// resolveGenTarget maps a generation scope to its artifact path and reader
// title, delegating placement to the workspace so the TUI and CLI agree: no
// files — or every chapter selected — targets the whole-course slot; any
// narrower selection targets a scoped artifact under the course subdir.
func (m Model) resolveGenTarget(kind, course string, files []string) (path, title string) {
	label := genLabel(kind)
	var whole string
	if kind == "qa" {
		path, whole = m.ws.QATarget(course, files), m.ws.QAPath(course)
	} else {
		path, whole = m.ws.GuideTarget(course, files), m.ws.GuidePath(course)
	}
	if path == whole {
		return path, course + label
	}
	return path, course + "/" + workspace.ScopeName(files) + label
}

// genLabel is the reader-title suffix for a generation kind.
func genLabel(kind string) string {
	if kind == "qa" {
		return " · q&a"
	}
	return " · guide"
}

// courseGrounding returns the grounding blob for a generation: the whole course
// by default, or only the named chapter files when files is non-empty.
func (m Model) courseGrounding(course string, files []string) (string, error) {
	if len(files) > 0 {
		return m.ws.MaterialFromFiles(course, files)
	}
	return m.ws.CourseMaterial(course)
}

// genCmd runs the generator off the UI goroutine and reports the result. count
// sets the Q&A pair target (<=0 = DefaultQACount); it is ignored for guides.
func genCmd(ctx context.Context, eng engine.Engine, kind, course, material string, count, epoch int) tea.Cmd {
	return func() tea.Msg {
		var (
			md  string
			err error
		)
		switch kind {
		case "guide":
			md, err = generate.Guide(ctx, eng, course, material)
		case "qa":
			md, err = generate.QA(ctx, eng, course, material, generate.QAOpts{Count: count})
		}
		return genDoneMsg{kind: kind, md: md, err: err, epoch: epoch}
	}
}

// genDone persists the artifact at the resolved scope path (force-overwriting so
// regenerate works), refreshes the dashboard chip counts, then opens the result
// in the reader.
func (m Model) genDone(msg genDoneMsg) (tea.Model, tea.Cmd) {
	if msg.epoch != m.workEpoch {
		return m, nil // superseded or cancelled — ignore this result
	}
	m.cancel = nil
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
	// Refresh the chapter badges too, so the ✓ is current when the reader is
	// closed back into the chapter hub.
	if m.back == stateChapters {
		m.loadChapScopes()
	}

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
	scope := m.genScope
	if scope == "" {
		scope = m.genCourse
	}
	body := m.spinnerHead() + "generating " + what + " for " + scope +
		" with " + m.engine + "…\n\n" + styleMuted.Render("grounded only in the course material")
	return lipgloss.NewStyle().Padding(2, 0, 0, 3).Render(body)
}
