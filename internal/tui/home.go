package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/workspace"
)

// Course-list column widths (cells). The header and every row share them so the
// g/q/e labels sit directly above their chips regardless of course-name length.
const (
	listNameCol = 24
	listChipCol = 6
	listGutter  = 1 // min gap between a full-width name and the first chip
)

// viewHome renders the home dashboard: the wordmark banner, then either the
// empty / first-run state (tips + notice) or the populated course list followed
// by a detail panel for the selected course. docs/07 §Home dashboard.
func (m Model) viewHome() string {
	var b strings.Builder

	b.WriteString(renderWordmark(m.width))
	b.WriteString("\n\n")

	if len(m.courses) > 0 {
		b.WriteString(m.viewCourseList())
		b.WriteString("\n")
		b.WriteString(m.viewCourseDetail())
		if m.notice != "" {
			b.WriteString("\n\n")
			b.WriteString(noticeLine(m.noticeLvl, m.notice))
		}
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

// viewCourseList renders the header row plus one fixed-column row per course,
// the selected row marked with a left bar and a surface background (docs/07
// §Course list).
func (m Model) viewCourseList() string {
	var b strings.Builder

	header := lipgloss.NewStyle().Bold(true).Foreground(cText).
		Render(fitCell("COURSES", 2+listNameCol+listGutter)) +
		styleMuted.Render(fitCell("g", listChipCol)+fitCell("q", listChipCol)+fitCell("e", listChipCol))
	b.WriteString(header)
	b.WriteString("\n")

	for i, c := range m.courses {
		b.WriteString(m.courseRow(i, c))
		b.WriteString("\n")
	}
	return b.String()
}

// courseRow renders one course as bar + fixed-width name + g/q/e chips. The
// selected row carries a surface background across every segment so the
// highlight is a solid block (not just the bar).
func (m Model) courseRow(i int, c workspace.Course) string {
	selected := i == m.cursor

	barStyle := lipgloss.NewStyle().Foreground(cPrimary)
	nameStyle := styleBody
	if selected {
		nameStyle = lipgloss.NewStyle().Bold(true).Foreground(cText)
		barStyle = barStyle.Background(cSurface)
		nameStyle = nameStyle.Background(cSurface)
	}

	bar := "  "
	if selected {
		bar = "▌ "
	}

	return barStyle.Render(bar) +
		nameStyle.Render(fitCell(c.Name, listNameCol)+strings.Repeat(" ", listGutter)) +
		chipCell("g", c.GuideCount(), listChipCol, selected) +
		chipCell("q", c.QACount(), listChipCol, selected) +
		chipCell("e", c.ExerciseSets, listChipCol, selected)
}

// viewCourseDetail fills the space below the list with a card for the selected
// course: which artifacts exist and the contextual keys that act on it.
func (m Model) viewCourseDetail() string {
	c := m.courses[m.cursor]

	title := styleTitle.Render(strings.ToUpper(c.Name))
	rows := []string{
		styleMuted.Render(fitCell("guide", 12)) + artifactState(c.HasGuide, "press g to build"),
		styleMuted.Render(fitCell("q&a", 12)) + artifactState(c.HasQA, "press q to build"),
		styleMuted.Render(fitCell("exercises", 12)) + exerciseState(c.ExerciseSets),
	}
	body := title + "\n\n" + strings.Join(rows, "\n") + "\n\n" + m.detailHints(c)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cLine).
		Padding(1, 2).
		Render(body)
}

// artifactState shows a green ready marker or a muted call to action.
func artifactState(ok bool, build string) string {
	if ok {
		return lipgloss.NewStyle().Foreground(cSuccess).Render("✓ ready")
	}
	return styleMuted.Render("— " + build)
}

func exerciseState(sets int) string {
	if sets > 0 {
		noun := "set"
		if sets > 1 {
			noun += "s"
		}
		return lipgloss.NewStyle().Foreground(cSuccess).
			Render(fmt.Sprintf("✓ %d %s · press s to solve", sets, noun))
	}
	return styleMuted.Render("— none ingested")
}

// detailHints lists only the keys that do something for this course, so the
// legend matches the course's actual state (open vs build).
func (m Model) detailHints(c workspace.Course) string {
	var parts []string
	if c.HasGuide {
		parts = append(parts, styleKey.Render("g")+styleMuted.Render(" open guide"))
	} else {
		parts = append(parts, styleKey.Render("g")+styleMuted.Render(" build guide"))
	}
	if c.HasQA {
		parts = append(parts,
			styleKey.Render("q")+styleMuted.Render(" open q&a"),
			styleKey.Render("r")+styleMuted.Render(" revise"))
	} else {
		parts = append(parts, styleKey.Render("q")+styleMuted.Render(" build q&a"))
	}
	if c.ExerciseSets > 0 {
		parts = append(parts, styleKey.Render("s")+styleMuted.Render(" solve"))
	}
	parts = append(parts, styleKey.Render("G/Q")+styleMuted.Render(" regen"))
	return strings.Join(parts, styleMuted.Render(" · "))
}

// chipCell renders a single count chip padded to a fixed column so chips align
// under the header labels; >0 in success color, 0 muted (FR-022). When the row
// is selected it carries the surface background.
func chipCell(label string, n, w int, selected bool) string {
	txt := fmt.Sprintf("%s·%d", label, n)
	style := styleMuted
	if n > 0 {
		style = lipgloss.NewStyle().Foreground(cSuccess)
	}
	if selected {
		style = style.Background(cSurface)
	}
	if pad := w - len([]rune(txt)); pad > 0 {
		txt += strings.Repeat(" ", pad)
	}
	return style.Render(txt)
}

// fitCell pads s with spaces to exactly w cells, or truncates with an ellipsis
// when longer (rune-aware).
func fitCell(s string, w int) string {
	r := []rune(s)
	switch {
	case len(r) == w:
		return s
	case len(r) < w:
		return s + strings.Repeat(" ", w-len(r))
	case w <= 1:
		return string(r[:w])
	default:
		return string(r[:w-1]) + "…"
	}
}
