package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/quiz"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// state enumerates the top-level TUI screens (docs/05 §TUI states).
type state int

const (
	stateHome state = iota
	stateReader
	stateQuiz
)

// Model is the root Bubble Tea model: a home dashboard plus a reader and an
// interactive quiz reachable from it.
type Model struct {
	state  state
	width  int
	height int

	engine string
	ws     workspace.Workspace

	courses []workspace.Course
	cursor  int

	// reader
	viewport viewport.Model
	vpReady  bool
	readTitle string

	// quiz
	pairs    []quiz.Pair
	qIndex   int
	revealed bool
	knew     int
	missed   int
	answer   textinput.Model
	quizDone bool
	quizName string

	// transient notice (e.g. "no guide yet")
	notice string
}

// New builds the root model from the workspace, active engine, and course list.
func New(engine string, ws workspace.Workspace, courses []workspace.Course) Model {
	ti := textinput.New()
	ti.Placeholder = "type your answer, then enter to reveal"
	ti.CharLimit = 0
	return Model{
		state:   stateHome,
		engine:  engine,
		ws:      ws,
		courses: courses,
		answer:  ti,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case stateHome:
			return m.updateHome(msg)
		case stateReader:
			return m.updateReader(msg)
		case stateQuiz:
			return m.updateQuiz(msg)
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.width > 0 && (m.width < 80 || m.height < 24) {
		return styleMuted.Render("resize terminal to at least 80×24")
	}

	var body string
	switch m.state {
	case stateReader:
		body = m.viewReader()
	case stateQuiz:
		body = m.viewQuiz()
	default:
		body = m.viewHome()
	}

	status := m.viewStatusBar()
	contentHeight := m.height - lipgloss.Height(status)
	if contentHeight < 0 {
		contentHeight = 0
	}
	content := lipgloss.NewStyle().Height(contentHeight).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, content, status)
}

// updateHome handles navigation and the open/generate/revise keys on the home
// dashboard.
func (m Model) updateHome(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.notice = ""
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.courses)-1 {
			m.cursor++
		}
	case "g", "enter":
		return m.openReader("guide")
	case "q":
		return m.openReader("qa")
	case "r":
		return m.startQuiz()
	}
	return m, nil
}

// viewStatusBar renders the persistent bottom anchor: root (left), context
// hints (center), engine (right). docs/07 §Status bar.
func (m Model) viewStatusBar() string {
	width := m.width
	if width <= 0 {
		width = 80
	}
	left := " " + m.ws.Root
	right := "engine:" + m.engine + " "

	hints := "g guide · q qa · r revise · ctrl+c quit"
	switch m.state {
	case stateReader:
		hints = "↑/↓ scroll · q back"
	case stateQuiz:
		hints = "quiz · q back"
	}

	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - lipgloss.Width(hints)
	if gap < 1 {
		gap = 1
	}
	leftPad := gap / 2
	rightPad := gap - leftPad

	row := left +
		lipgloss.NewStyle().Width(leftPad).Render("") +
		styleMuted.Render(hints) +
		lipgloss.NewStyle().Width(rightPad).Render("") +
		styleInfo.Render(right)
	return styleStatus.Width(width).Render(row)
}

func (m *Model) resizeViewport() {
	if m.width == 0 {
		return
	}
	h := m.height - 4
	if h < 3 {
		h = 3
	}
	w := m.width
	if w > 100 {
		w = 100
	}
	if !m.vpReady {
		m.viewport = viewport.New(w, h)
		m.vpReady = true
	} else {
		m.viewport.Width = w
		m.viewport.Height = h
	}
}
