package tui

import (
	"os"
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

// startQuiz loads the selected course's Q&A file, parses it, and enters the
// quiz. A missing/unparseable file sets a home notice (ERR-061).
func (m Model) startQuiz() (tea.Model, tea.Cmd) {
	if len(m.courses) == 0 {
		return m, nil
	}
	course := m.courses[m.cursor].Name
	data, err := os.ReadFile(m.ws.QAPath(course))
	if err != nil {
		m.notice = "no q&a yet for " + course + " — generate it: genius qa " + course
		return m, nil
	}
	pairs, err := quiz.Parse(string(data))
	if err != nil {
		m.notice = "q&a for " + course + " is unparseable: " + err.Error()
		return m, nil
	}

	m.pairs = pairs
	m.qIndex = 0
	m.revealed = false
	m.knew, m.missed = 0, 0
	m.quizDone = false
	m.quizName = course
	m.answer.SetValue("")
	m.answer.Focus()
	m.state = stateQuiz
	return m, nil
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

	progress := styleMuted.Render("Q " + itoa(p.N) + "/" + itoa(len(m.pairs)))
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
	return card
}

func (m Model) viewQuizSummary() string {
	total := len(m.pairs)
	body := styleTitle.Render("session complete") + "\n\n" +
		styleBody.Render("questions: "+itoa(total)) + "\n" +
		lipgloss.NewStyle().Foreground(cSuccess).Render("knew it: "+itoa(m.knew)) + "\n" +
		lipgloss.NewStyle().Foreground(cError).Render("missed:  "+itoa(m.missed)) + "\n\n" +
		styleMuted.Render("q or enter to return home")
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cPrimary).
		Padding(1, 2).
		Render(body)
}

func chipKey(key, label string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Render("["+key+"] "+label)
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
