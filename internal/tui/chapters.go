package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// openChapters enters the chapter-scope picker for the highlighted course, so a
// guide or Q&A can be grounded on specific chapters instead of the whole course
// (FR-055). A course with no chapters sets a notice instead.
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
	m.state = stateChapters
	return m, nil
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
	case "g":
		return m.startGenerate("guide", m.chapCourse, m.scopedChapters())
	case "q":
		return m.startGenerate("qa", m.chapCourse, m.scopedChapters())
	}
	return m, nil
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

func (m Model) viewChapters() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render(m.chapCourse + " · chapters"))
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render("select chapters to ground on, then "))
	b.WriteString(styleKey.Render("g") + styleMuted.Render(" guide / "))
	b.WriteString(styleKey.Render("q") + styleMuted.Render(" q&a  (none = whole course)"))
	b.WriteString("\n\n")

	start, end := window(len(m.chapFiles), m.chapCursor, m.contentHeight()-7)
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
		label := styleBody.Render(m.chapFiles[i])
		if i == m.chapCursor {
			label = lipgloss.NewStyle().Bold(true).Foreground(cText).Render(m.chapFiles[i])
		}
		b.WriteString(cursor + box + " " + label + "\n")
	}
	if end < len(m.chapFiles) {
		b.WriteString(styleMuted.Render("   ↓ more") + "\n")
	}

	b.WriteString("\n" + styleMuted.Render("space select · g guide · q q&a · esc back"))
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}
