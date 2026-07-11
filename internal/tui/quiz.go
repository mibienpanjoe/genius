package tui

import (
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/quiz"
	"github.com/mibienpanjoe/genius/internal/render"
)

// renderAnswer styles a revealed answer's markdown body, trimming the trailing
// blank line Glamour adds.
func renderAnswer(md string, width int) (string, error) {
	out, err := render.Markdown(md, width)
	if err != nil {
		return "", err
	}
	return strings.TrimRight(out, "\n"), nil
}

// quizSource is one revisable Q&A in the picker: the whole-course Q&A, a single
// chapter, a span, or the synthetic "all chapters merged" entry.
type quizSource struct {
	label  string // shown in the picker
	path   string // file to read (empty for merged)
	merged bool   // concatenate every non-merged source
}

// startQuiz gathers the course's Q&A sources (whole + per-chapter) and enters
// the quiz. With a single source it loads straight in; with several it opens the
// source picker (whole / chapter / all merged). No Q&A at all sets a home notice
// (ERR-061).
func (m Model) startQuiz() (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	course := m.courses[m.cursor].Name

	sources := m.quizSources(course)
	switch len(sources) {
	case 0:
		m.notice = "no q&a yet for " + course + " — press q to build it"
		m.noticeLvl = lvlWarn
		return m, nil
	case 1:
		return m.loadQuizSource(course, sources[0])
	}

	m.quizPick = sources
	m.quizPickCursor = 0
	m.quizName = course
	m.state = stateQuizPick
	return m, nil
}

// quizSources lists the available Q&A for a course: the whole-course file (if
// present) plus each scoped artifact, and an "all merged" entry when more than
// one exists.
func (m Model) quizSources(course string) []quizSource {
	var sources []quizSource
	if _, err := os.Stat(m.ws.QAPath(course)); err == nil {
		sources = append(sources, quizSource{label: "whole course", path: m.ws.QAPath(course)})
	}
	scopes, _ := m.ws.QAScopes(course)
	for _, s := range scopes {
		sources = append(sources, quizSource{label: s, path: m.ws.ChapterQAPath(course, s)})
	}
	if len(sources) > 1 {
		sources = append(sources, quizSource{label: "all chapters merged", merged: true})
	}
	return sources
}

// updateQuizPick drives the Q&A source picker.
func (m Model) updateQuizPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.state = stateHome
		return m, nil
	case "up", "k":
		if m.quizPickCursor > 0 {
			m.quizPickCursor--
		}
	case "down", "j":
		if m.quizPickCursor < len(m.quizPick)-1 {
			m.quizPickCursor++
		}
	case "enter", " ":
		return m.loadQuizSource(m.quizName, m.quizPick[m.quizPickCursor])
	}
	return m, nil
}

// loadQuizSource reads the chosen source (or merges all of them), parses the
// Q&A, and enters the quiz; parse/read errors set a home notice.
func (m Model) loadQuizSource(course string, src quizSource) (tea.Model, tea.Cmd) {
	var md string
	if src.merged {
		var b strings.Builder
		for _, s := range m.quizPick {
			if s.merged {
				continue
			}
			data, err := os.ReadFile(s.path)
			if err != nil {
				continue
			}
			b.Write(data)
			b.WriteString("\n\n")
		}
		md = b.String()
	} else {
		data, err := os.ReadFile(src.path)
		if err != nil {
			m.notice = "no q&a yet for " + course + " — press q to build it"
			m.noticeLvl = lvlWarn
			m.state = stateHome
			return m, nil
		}
		md = string(data)
	}

	pairs, err := quiz.Parse(md)
	if err != nil {
		m.notice = "q&a for " + course + " is unparseable: " + err.Error()
		m.noticeLvl = lvlErr
		m.state = stateHome
		return m, nil
	}
	return m.enterQuiz(course, pairs), nil
}

// enterQuiz resets the quiz state for a fresh run over pairs.
func (m Model) enterQuiz(course string, pairs []quiz.Pair) Model {
	m.pairs = pairs
	m.qIndex = 0
	m.revealed = false
	m.knew, m.missed = 0, 0
	m.quizDone = false
	m.quizName = course
	m.answer.SetValue("")
	m.answer.Focus()
	m.state = stateQuiz
	return m
}

// viewQuizPick lists the Q&A sources to revise.
func (m Model) viewQuizPick() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render(m.quizName + " · revise"))
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render("which q&a?"))
	b.WriteString("\n\n")

	for i, s := range m.quizPick {
		cursor := "  "
		label := styleBody.Render(s.label)
		if i == m.quizPickCursor {
			cursor = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
			label = lipgloss.NewStyle().Bold(true).Foreground(cText).Render(s.label)
		}
		b.WriteString(cursor + label + "\n")
	}

	b.WriteString("\n" + styleMuted.Render("↑/↓ pick · enter revise · esc back"))
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}

func (m Model) updateQuiz(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "ctrl+c":
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		m.state = stateHome
		return m, nil
	}

	if m.quizDone {
		if msg.String() == "q" || msg.String() == "enter" {
			m.state = stateHome
		}
		return m, nil
	}

	if !m.revealed {
		switch msg.String() {
		case "enter":
			m.revealed = true
			m.answer.Blur()
			return m, nil
		}
		var cmd tea.Cmd
		m.answer, cmd = m.answer.Update(msg)
		return m, cmd
	}

	// Revealed: self-grade and advance.
	switch msg.String() {
	case "y":
		m.knew++
		return m.nextQuestion(), nil
	case "n":
		m.missed++
		return m.nextQuestion(), nil
	case " ", "enter":
		return m.nextQuestion(), nil
	case "q":
		m.state = stateHome
		return m, nil
	}
	return m, nil
}

func (m Model) nextQuestion() Model {
	m.qIndex++
	if m.qIndex >= len(m.pairs) {
		m.quizDone = true
		return m
	}
	m.revealed = false
	m.answer.SetValue("")
	m.answer.Focus()
	return m
}

func (m Model) viewQuiz() string {
	if m.quizDone {
		return m.viewQuizSummary()
	}
	p := m.pairs[m.qIndex]

	// Position in this session, not the pair's own N: merged sources restart
	// their numbering per file, so p.N would repeat and never reach the total.
	progress := styleMuted.Render("Q " + strconv.Itoa(m.qIndex+1) + "/" + strconv.Itoa(len(m.pairs)))
	q := styleBody.Bold(true).Render(p.Question)

	var b string
	b = progress + "\n\n" + q + "\n\n"
	if !m.revealed {
		b += m.answer.View() + "\n\n" + styleMuted.Render("enter to reveal")
	} else {
		divider := lipgloss.NewStyle().Foreground(cLine).Render(repeat("─", 40))
		ans, err := renderAnswer(p.Answer, cardWidth(m.width))
		if err != nil {
			ans = p.Answer
		}
		grade := chipKey("y", "knew it", cSuccess) + "  " +
			chipKey("n", "missed", cError) + "  " +
			chipKey("space", "next", cTextMuted)
		b += divider + "\n" + ans + "\n" + grade
	}

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cPrimary).
		Padding(1, 2).
		Width(cardWidth(m.width)).
		Render(b)
	return center(m.width, m.contentHeight(), card)
}

func (m Model) viewQuizSummary() string {
	total := len(m.pairs)
	body := styleTitle.Render("session complete") + "\n\n" +
		styleBody.Render("questions: "+strconv.Itoa(total)) + "\n" +
		lipgloss.NewStyle().Foreground(cSuccess).Render("knew it: "+strconv.Itoa(m.knew)) + "\n" +
		lipgloss.NewStyle().Foreground(cError).Render("missed:  "+strconv.Itoa(m.missed)) + "\n\n" +
		styleMuted.Render("q or enter to return home")
	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cPrimary).
		Padding(1, 2).
		Render(body)
	return center(m.width, m.contentHeight(), card)
}

func chipKey(key, label string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Render("[" + key + "] " + label)
}

func cardWidth(termWidth int) int {
	w := termWidth - 6
	if w > 96 {
		w = 96
	}
	if w < 20 {
		w = 20
	}
	return w
}

func repeat(s string, n int) string {
	out := make([]byte, 0, len(s)*n)
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}
