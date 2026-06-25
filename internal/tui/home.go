package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewHome renders the home dashboard: the wordmark banner, then either the
// empty / first-run state (tips + notice) or the populated course list.
// docs/07 §Home dashboard.
func (m Model) viewHome() string {
	var b strings.Builder

	banner := renderWordmark(m.width)
	b.WriteString(banner)
	b.WriteString("\n\n")

	if len(m.courses) > 0 {
		b.WriteString(m.viewCourseList())
		return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
	}

	// Empty-state getting-started block.
	tips := []string{
		styleMuted.Render("getting started:"),
		styleMuted.Render("1. ingest a course:  ") + styleKey.Render("genius ingest lecture.pdf"),
		styleMuted.Render("2. generate a guide: press ") + styleKey.Render("g") +
			styleMuted.Render("   3. revise: press ") + styleKey.Render("r"),
		styleMuted.Render("4. ") + styleKey.Render("?") + styleMuted.Render(" for help"),
	}
	b.WriteString(strings.Join(tips, "\n"))
	b.WriteString("\n\n")

	notice := styleNotice.Render(
		"No courses yet — ingest a PDF or PPT to begin.\n" +
			"Workspace: " + m.ws.Root)
	b.WriteString(notice)
	b.WriteString("\n\n")

	b.WriteString(styleMuted.Render("using: ") +
		styleInfo.Render("engine:"+m.engine) +
		styleMuted.Render(" · "+m.ws.Root))

	// Indent the whole home body two cells.
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}

// viewCourseList renders the populated home: a header row plus one row per
// course with g·N q·N e·N count chips, the selected row marked with a left bar.
// docs/07 §Course list.
func (m Model) viewCourseList() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(cText).
		Render("COURSES") + styleMuted.Render("              g    q    e")
	b.WriteString(header)
	b.WriteString("\n")

	for i, c := range m.courses {
		selected := i == m.cursor
		bar := "  "
		nameStyle := styleBody
		if selected {
			bar = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(cText)
		}
		name := nameStyle.Render(padRight(c.Name, 18))
		chips := chip("g", c.GuideCount()) + "  " +
			chip("q", c.QACount()) + "  " +
			chip("e", c.ExerciseSets)
		b.WriteString(bar + name + chips + "\n")
	}

	b.WriteString("\n")
	b.WriteString(styleMuted.Render("enter/g guide · ") + styleKey.Render("q") +
		styleMuted.Render(" qa · ") + styleKey.Render("r") + styleMuted.Render(" revise · ") +
		styleKey.Render("s") + styleMuted.Render(" solve"))

	if m.notice != "" {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(cWarning).Render("! " + m.notice))
	}
	return b.String()
}

// chip renders a single count chip; >0 in success color, 0 muted (FR-022).
func chip(label string, n int) string {
	style := styleMuted
	if n > 0 {
		style = lipgloss.NewStyle().Foreground(cSuccess)
	}
	return style.Render(fmt.Sprintf("%s·%d", label, n))
}

func padRight(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}
