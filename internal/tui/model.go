package tui

import (
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/generate"
	"github.com/mibienpanjoe/genius/internal/quiz"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// state enumerates the top-level TUI screens (docs/05 §TUI states).
type state int

const (
	stateHome state = iota
	stateReader
	stateQuiz
	stateExSets     // pick an exercise set for the solve flow
	stateExList     // pick exercises within a set
	stateSolving    // engine is producing the solution (spinner)
	stateGenerating // engine is producing a guide or Q&A (spinner)
	stateHelp       // keybinding reference overlay
	stateIngestPick // browse the filesystem to choose document(s)
	stateIngestOpts // choose course/kind/flags before ingesting
	stateIngesting  // converter is running (spinner)
	stateChapters   // pick chapters to scope a guide/Q&A
)

// Notice severity levels — drive the glyph and color so meaning is never carried
// by color alone (docs/07 §Notices).
const (
	lvlWarn = iota // !  pending/empty/refusal
	lvlErr         // ✗  failures
	lvlInfo        // i  neutral
)

// Model is the root Bubble Tea model: a home dashboard plus a reader and an
// interactive quiz reachable from it.
type Model struct {
	state  state
	width  int
	height int

	engine string
	eng    engine.Engine
	ws     workspace.Workspace

	courses []workspace.Course
	cursor  int

	// solve flow
	exSets      []string            // set names for the active course
	exSetCursor int                 // cursor in the set picker
	exItems     []generate.Exercise // enumerated exercises of the chosen set
	exCursor    int                 // cursor in the exercise list
	exSelected  map[int]bool        // toggled exercises (space)
	solveCourse string              // course being solved
	solveSet    string              // set being solved
	spinner     spinner.Model       // shown while the engine runs

	// generate flow (guide / qa)
	genKind   string // "guide" or "qa" being generated
	genCourse string // course being generated for

	// ingest flow
	ingDir      string          // directory currently browsed
	ingEntries  []dirEntry      // dirs + ingestable files in ingDir
	ingCursor   int             // cursor in the browser
	ingSelected map[string]bool // chosen file paths (batch)
	ingInput    textinput.Model // target course/set name
	ingKind     string          // "course" or "exercise"
	ingDescribe bool            // vision-caption figures
	ingOverlay  bool            // overwrite existing targets
	ingField    int             // active field in the options form

	// chapter-scope flow (per-chapter guide/Q&A)
	chapCourse   string       // course whose chapters are listed
	chapFiles    []string     // chapter filenames
	chapCursor   int          // cursor in the chapter list
	chapSelected map[int]bool // toggled chapters

	// reader
	viewport  viewport.Model
	vpReady   bool
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

	// transient notice (e.g. "no guide yet") and its severity (lvlWarn/Err/Info)
	notice    string
	noticeLvl int

	// reduceMotion replaces spinners with a static label (GENIUS_NO_ANIM/NO_COLOR).
	reduceMotion bool
}

// New builds the root model from the workspace, active engine, and course list.
// eng drives the interactive solve flow; it may be nil in contexts that never
// solve (e.g. tests), in which case the solve action reports it.
func New(engineName string, eng engine.Engine, ws workspace.Workspace, courses []workspace.Course) Model {
	ti := textinput.New()
	ti.Placeholder = "type your answer, then enter to reveal"
	ti.CharLimit = 0

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(cPrimary)

	ing := textinput.New()
	ing.Placeholder = "course name"
	ing.CharLimit = 64

	return Model{
		state:        stateHome,
		engine:       engineName,
		eng:          eng,
		ws:           ws,
		courses:      courses,
		answer:       ti,
		spinner:      sp,
		ingInput:     ing,
		ingDescribe:  true,
		ingKind:      "course",
		reduceMotion: os.Getenv("GENIUS_NO_ANIM") != "" || os.Getenv("NO_COLOR") != "",
	}
}

// noticeLine renders a transient notice with a leading glyph + semantic color so
// meaning survives color stripping (docs/07 §Notices).
func noticeLine(lvl int, msg string) string {
	glyph, color := "!", cWarning
	switch lvl {
	case lvlErr:
		glyph, color = "✗", cError
	case lvlInfo:
		glyph, color = "i", cInfo
	}
	return lipgloss.NewStyle().Foreground(color).Render(glyph + " " + msg)
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeViewport()
		return m, nil

	case solveDoneMsg:
		return m.solveDone(msg)

	case genDoneMsg:
		return m.genDone(msg)

	case ingestDoneMsg:
		return m.ingestDone(msg)

	case spinner.TickMsg:
		if m.state != stateSolving && m.state != stateGenerating && m.state != stateIngesting {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch m.state {
		case stateHome:
			return m.updateHome(msg)
		case stateReader:
			return m.updateReader(msg)
		case stateQuiz:
			return m.updateQuiz(msg)
		case stateExSets:
			return m.updateExSets(msg)
		case stateExList:
			return m.updateExList(msg)
		case stateHelp:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			m.state = stateHome
			return m, nil
		case stateIngestPick:
			return m.updateIngestPick(msg)
		case stateIngestOpts:
			return m.updateIngestOpts(msg)
		case stateChapters:
			return m.updateChapters(msg)
		case stateSolving, stateGenerating, stateIngesting:
			if msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
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
	case stateExSets:
		body = m.viewExSets()
	case stateExList:
		body = m.viewExList()
	case stateSolving:
		body = m.viewSolving()
	case stateGenerating:
		body = m.viewGenerating()
	case stateHelp:
		body = m.viewHelp()
	case stateIngestPick:
		body = m.viewIngestPick()
	case stateIngestOpts:
		body = m.viewIngestOpts()
	case stateIngesting:
		body = m.viewIngesting()
	case stateChapters:
		body = m.viewChapters()
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
	m.noticeLvl = lvlWarn
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "?":
		m.state = stateHelp
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.courses)-1 {
			m.cursor++
		}
	case "g", "enter":
		return m.generateOrOpen("guide", false)
	case "q":
		return m.generateOrOpen("qa", false)
	case "G":
		return m.generateOrOpen("guide", true)
	case "Q":
		return m.generateOrOpen("qa", true)
	case "r":
		return m.startQuiz()
	case "s":
		return m.openExSets()
	case "i":
		return m.openIngest()
	case "f":
		return m.openChapters()
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

	hints := "↑/↓ move · enter/g guide · q qa · r revise · s solve · i ingest · f chapters · ? help"
	switch m.state {
	case stateReader:
		hints = "↑/↓ scroll · q back"
	case stateQuiz:
		hints = "quiz · q back"
	case stateExSets:
		hints = "↑/↓ pick set · enter open · esc back"
	case stateExList:
		hints = "↑/↓ move · space select · enter solve · esc back"
	case stateSolving:
		hints = "solving… · ctrl+c quit"
	case stateGenerating:
		hints = "generating… · ctrl+c quit"
	case stateHelp:
		hints = "any key to close"
	case stateIngestPick:
		hints = "↑/↓ move · enter open/pick · space select · ⌫ up · esc back"
	case stateIngestOpts:
		hints = "↑/↓ field · tab kind · enter ingest · esc back"
	case stateIngesting:
		hints = "ingesting… · ctrl+c quit"
	case stateChapters:
		hints = "↑/↓ move · space select · g guide · q q&a · esc back"
	}

	// Keep the bar to a single row: truncate hints that can't fit between the
	// root (left) and engine (right) anchors.
	if avail := width - lipgloss.Width(left) - lipgloss.Width(right) - 2; avail >= 1 &&
		lipgloss.Width(hints) > avail {
		hints = fitCell(hints, avail)
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

// spinnerHead returns the animated spinner glyph, or a static label under
// reduced motion (docs/07: GENIUS_NO_ANIM/NO_COLOR).
func (m Model) spinnerHead() string {
	if m.reduceMotion {
		return "working… "
	}
	return m.spinner.View() + " "
}

// contentHeight is the space above the one-row status bar.
func (m Model) contentHeight() int {
	if h := m.height - 1; h > 0 {
		return h
	}
	return 0
}

// center places s in the middle of a width×height area (no-op before the first
// WindowSizeMsg).
func center(width, height int, s string) string {
	if width <= 0 || height <= 0 {
		return s
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, s)
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
