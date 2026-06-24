package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

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

	data, err := os.ReadFile(path)
	if err != nil {
		m.notice = "no " + kind + " yet for " + course + " — generate it from the CLI: genius " + kind + " " + course
		return m, nil
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
	return m, nil
}

func (m Model) updateReader(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.state = stateHome
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
	return title + "\n\n" + m.viewport.View() + "\n" + footer
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
