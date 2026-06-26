package tui

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/workspace"
)

// openChapters enters the chapter hub for the highlighted course: per-chapter
// guides and Q&A can be built and opened without touching the whole-course
// artifacts (FR-055). A course with no chapters sets a notice instead.
func (m Model) openChapters() (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	course := m.courses[m.cursor].Name
	files, err := m.ws.CourseFiles(course)
	if err != nil {
		m.notice = "could not read chapters for " + course + ": " + err.Error()
		m.noticeLvl = lvlErr
		return m, nil
	}
	if len(files) == 0 {
		m.notice = "no chapters to scope for " + course + " — ingest a document first"
		m.noticeLvl = lvlWarn
		return m, nil
	}
	m.chapCourse = course
	m.chapFiles = files
	m.chapCursor = 0
	m.chapSelected = make(map[int]bool)
	m.loadChapScopes()
	m.state = stateChapters
	return m, nil
}

// loadChapScopes refreshes the lists of existing scoped artifacts so the hub can
// badge each chapter and footer the combined spans.
func (m *Model) loadChapScopes() {
	m.chapGuideScopes, _ = m.ws.GuideScopes(m.chapCourse)
	m.chapQAScopes, _ = m.ws.QAScopes(m.chapCourse)
}

func (m Model) updateChapters(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.state = stateHome
		return m, nil
	case "up", "k":
		if m.chapCursor > 0 {
			m.chapCursor--
		}
	case "down", "j":
		if m.chapCursor < len(m.chapFiles)-1 {
			m.chapCursor++
		}
	case " ":
		m.chapSelected[m.chapCursor] = !m.chapSelected[m.chapCursor]
	case "g", "enter":
		return m.generateOrOpenScoped("guide", false)
	case "q":
		return m.generateOrOpenScoped("qa", false)
	case "G":
		return m.generateOrOpenScoped("guide", true)
	case "Q":
		return m.generateOrOpenScoped("qa", true)
	}
	return m, nil
}

// generateOrOpenScoped opens the scoped artifact for the current selection when
// it already exists, otherwise builds it. force (G/Q) always rebuilds. With no
// chapters selected the scope is the whole course (resolveGenTarget).
func (m Model) generateOrOpenScoped(kind string, force bool) (tea.Model, tea.Cmd) {
	files := m.scopedChapters()
	path, title := m.resolveGenTarget(kind, m.chapCourse, files)
	if !force {
		if _, err := os.Stat(path); err == nil {
			if mm, cmd, oerr := m.openReaderPath(path, title); oerr == nil {
				return mm, cmd
			}
		}
	}
	return m.startGenerate(kind, m.chapCourse, files)
}

// scopedChapters returns the toggled chapter filenames, or nil (whole course)
// when none are selected.
func (m Model) scopedChapters() []string {
	var files []string
	for i, f := range m.chapFiles {
		if m.chapSelected[i] {
			files = append(files, f)
		}
	}
	return files
}

// hasScope reports whether name is among the existing scoped artifacts.
func hasScope(scopes []string, name string) bool {
	for _, s := range scopes {
		if s == name {
			return true
		}
	}
	return false
}

// combinedScopes returns the scope names that span more than one chapter
// (contain a "+"), for the footer listing.
func combinedScopes(scopes []string) []string {
	var out []string
	for _, s := range scopes {
		if strings.Contains(s, "+") {
			out = append(out, s)
		}
	}
	return out
}

func (m Model) viewChapters() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render(m.chapCourse + " · chapters"))
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render("select chapters, then "))
	b.WriteString(styleKey.Render("g") + styleMuted.Render(" guide / "))
	b.WriteString(styleKey.Render("q") + styleMuted.Render(" q&a · "))
	b.WriteString(styleKey.Render("G/Q") + styleMuted.Render(" rebuild  (none = whole course)"))
	b.WriteString("\n\n")

	start, end := window(len(m.chapFiles), m.chapCursor, m.contentHeight()-9)
	if start > 0 {
		b.WriteString(styleMuted.Render("   ↑ more") + "\n")
	}
	for i := start; i < end; i++ {
		cursor := "  "
		if i == m.chapCursor {
			cursor = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
		}
		box := "[ ]"
		if m.chapSelected[i] {
			box = lipgloss.NewStyle().Foreground(cSuccess).Render("[x]")
		}
		name := fitCell(m.chapFiles[i], 22)
		label := styleBody.Render(name)
		if i == m.chapCursor {
			label = lipgloss.NewStyle().Bold(true).Foreground(cText).Render(name)
		}
		scope := workspace.Slug(m.chapFiles[i])
		badges := scopeBadge("guide", hasScope(m.chapGuideScopes, scope)) + "  " +
			scopeBadge("qa", hasScope(m.chapQAScopes, scope))
		b.WriteString(cursor + box + " " + label + "  " + badges + "\n")
	}
	if end < len(m.chapFiles) {
		b.WriteString(styleMuted.Render("   ↓ more") + "\n")
	}

	if span := append(combinedScopes(m.chapGuideScopes), combinedScopes(m.chapQAScopes)...); len(span) > 0 {
		b.WriteString("\n" + styleMuted.Render("spans: "+strings.Join(span, ", ")))
	}

	b.WriteString("\n\n" + styleMuted.Render("space select · g guide · q q&a · G/Q rebuild · esc back"))
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}

// scopeBadge renders a per-chapter artifact marker: green ✓ when it exists, a
// muted — otherwise, so meaning survives color stripping.
func scopeBadge(label string, ok bool) string {
	if ok {
		return lipgloss.NewStyle().Foreground(cSuccess).Render(label + " ✓")
	}
	return styleMuted.Render(label + " —")
}
