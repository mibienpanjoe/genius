package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/render"
)

// openReader loads a course artifact ("guide" or "qa") for the selected course,
// renders it with Glamour, and switches to the reader. A missing artifact sets
// a home notice instead of navigating.
func (m Model) openReader(kind string) (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	course := m.courses[m.cursor].Name

	var path, title string
	switch kind {
	case "guide":
		path = m.ws.GuidePath(course)
		title = course + " · guide"
	case "qa":
		path = m.ws.QAPath(course)
		title = course + " · q&a"
	}

	mm, cmd, err := m.openReaderPath(path, title)
	if err != nil {
		m.notice = "no " + kind + " yet for " + course
		m.noticeLvl = lvlWarn
		return m, nil
	}
	return mm, cmd
}

// openReaderPath loads any markdown artifact by path, renders it with Glamour,
// and switches to the reader. It returns an error (without mutating state) when
// the file can't be read, so callers can set their own notice.
func (m Model) openReaderPath(path, title string) (tea.Model, tea.Cmd, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return m, nil, err
	}

	m.resizeViewport()
	out, rerr := render.Markdown(string(data), m.viewport.Width)
	if rerr != nil {
		out = string(data) // fall back to raw markdown
	}
	m.viewport.SetContent(out)
	m.viewport.GotoTop()
	m.readTitle = title
	m.state = stateReader
	return m, nil, nil
}

func (m Model) updateReader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.state = m.back
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m Model) viewReader() string {
	title := styleTitle.Render(m.readTitle)
	footer := styleMuted.Render(scrollLabel(m.viewport.ScrollPercent()) + " · q back")
	block := title + "\n\n" + m.viewport.View() + "\n" + footer
	if m.width <= 0 {
		return block
	}
	return lipgloss.Place(m.width, m.contentHeight(), lipgloss.Center, lipgloss.Top, block)
}

func scrollLabel(pct float64) string {
	p := int(pct * 100)
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return itoa(p) + "%"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
